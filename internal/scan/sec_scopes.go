package scan

import "strconv"

type scopeFact struct {
	label string
	paths [][]string
}

func comparableFacts() []scopeFact {
	return []scopeFact{
		{"navigator.userAgent", [][]string{{"userAgent"}}},
		{"navigator.platform", [][]string{{"platform"}}},
		{"navigator.hardwareConcurrency", [][]string{{"hardwareConcurrency"}}},
		{"navigator.language", [][]string{{"language"}}},
		{"navigator.languages", [][]string{{"languages"}}},
		{"resolved time zone", [][]string{{"timeZone"}, {"tz"}}},
		{"resolved locale", [][]string{{"locale"}}},
	}
}

func sectionScopes(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Inconclusive}

	type scopeRead struct {
		id    string
		label string
	}
	scopes := []scopeRead{
		{"scope.main", "main thread"},
		{"scope.worker", "dedicated worker"},
		{"scope.iframe", "same-origin frame"},
	}

	values := map[string]any{}
	var live []scopeRead
	for _, sc := range scopes {
		v, ok := r.value(sc.id)
		if !ok {
			reason := "not collected"
			if r.unsupported(sc.id) {
				reason = "the browser would not create this scope"
			} else if r.ran(sc.id) {
				reason = "the collector reported an error for this scope"
			}
			s.Rows = append(s.Rows, Row{Label: sc.label, Value: reason, Note: "not read as evidence"})
			continue
		}
		values[sc.id] = v
		live = append(live, sc)
	}
	if av, ok := r.value("scope.availability"); ok {
		if created, ok := nameSet(av, "created", "available", "ok"); ok {
			s.Rows = append(s.Rows, Row{
				Label: "scopes the browser created",
				Value: joinLimit(created, 8),
				Note:  "a scope that could not be created is not evidence of anything",
			})
		}
	}
	s.Rows = append(s.Rows, Row{
		Label: "scopes compared",
		Value: strconv.Itoa(len(live)),
		Note:  "facts are compared only within this one page load",
	})
	if len(live) < 2 {
		return s
	}

	disagreed := 0
	compared := 0
	for _, f := range comparableFacts() {
		seen := map[string][]string{}
		for _, sc := range live {
			raw, ok := factValue(values[sc.id], f)
			if !ok {
				continue
			}
			seen[raw] = append(seen[raw], sc.label)
		}
		if len(seen) == 0 {
			continue
		}
		if len(seen) == 1 {
			for v, where := range seen {
				if len(where) < 2 {
					s.Rows = append(s.Rows, Row{Label: f.label, Value: clip(v, 200), Note: "read in one scope only, so not compared"})
					continue
				}
				compared++
				s.Rows = append(s.Rows, Row{Label: f.label, Value: clip(v, 200), Note: "identical in " + strconv.Itoa(len(where)) + " scopes"})
			}
			continue
		}
		compared++
		disagreed++
		note := ""
		for _, v := range keys(toAnyMap(seen)) {
			if note != "" {
				note += "; "
			}
			note += joinLimit(seen[v], 3) + ": " + clip(v, 120)
		}
		s.Rows = append(s.Rows, Row{Label: f.label, Value: "the scopes disagree", Note: note})
	}

	if compared == 0 {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "no fact was reported by two scopes", Note: "nothing was compared"})
		return s
	}
	if disagreed > 0 {
		s.Determination = Instrumented
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: strconv.Itoa(disagreed) + " of " + strconv.Itoa(compared) + " facts differ between scopes of the same page",
			Note:  "one definition in the specification cannot produce two values; something in this page's JavaScript environment has been redefined",
		})
		return s
	}
	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "every fact reported by two or more scopes matched",
		Note:  strconv.Itoa(compared) + " facts compared across " + strconv.Itoa(len(live)) + " scopes",
	})
	return s
}

func factValue(scope any, f scopeFact) (string, bool) {
	for _, p := range f.paths {
		x, ok := field(scope, p...)
		if !ok {
			continue
		}
		switch t := x.(type) {
		case string:
			return t, true
		case float64:
			return itoa(t), true
		case bool:
			if t {
				return "true", true
			}
			return "false", true
		case []any:
			var parts []string
			for _, e := range t {
				if s, ok := e.(string); ok {
					parts = append(parts, s)
				}
			}
			if len(parts) == 0 {
				continue
			}
			return joinLimit(parts, 8), true
		}
	}
	return "", false
}

func toAnyMap[T any](m map[string]T) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
