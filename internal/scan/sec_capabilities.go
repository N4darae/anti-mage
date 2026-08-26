package scan

import (
	"sort"
	"strconv"
)

func sectionCapabilities(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	raw, ok := r.value("media.matrix")
	if !ok {
		s.Determination = Inconclusive
		s.Rows = append(s.Rows, Row{Label: "codec answers", Value: "not collected", Note: "the collector did not report them"})
		return s
	}
	m, isMap := raw.(map[string]any)
	if !isMap {
		s.Rows = append(s.Rows, Row{Label: "codec answers", Value: "not readable", Note: "reported in a shape this engine does not render"})
		return s
	}

	type entry struct{ label, value string }
	var entries []entry
	for _, k := range keys(m) {
		entries = append(entries, entry{k, renderCodec(m[k])})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].label < entries[j].label })
	const maxRows = 40
	for i, e := range entries {
		if i >= maxRows {
			s.Rows = append(s.Rows, Row{
				Label: "further answers",
				Value: strconv.Itoa(len(entries)-maxRows) + " more not shown",
				Note:  "",
			})
			break
		}
		s.Rows = append(s.Rows, Row{Label: clip(e.label, 120), Value: e.value, Note: ""})
	}
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "reported, not weighed",
		Note:  "this project measured the codec matrix on two browsers and it separated neither from the other, so no determination is drawn from it",
	})
	return s
}

func renderCodec(v any) string {
	switch t := v.(type) {
	case string:
		if t == "" {
			return `"" (no answer)`
		}
		return clip(t, 60)
	case bool:
		if t {
			return "supported"
		}
		return "not supported"
	case map[string]any:
		out := ""
		for _, k := range keys(t) {
			part := ""
			switch x := t[k].(type) {
			case string:
				part = k + " " + clip(x, 40)
			case bool:
				part = k + " " + boolWord(x)
			case float64:
				part = k + " " + itoa(x)
			default:
				continue
			}
			if out != "" {
				out += ", "
			}
			out += part
		}
		if out == "" {
			return "no answer"
		}
		return clip(out, 200)
	}
	return "no answer"
}

func boolWord(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
