package scan

import "sort"

func sectionMediaPaths(r Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}
	var t tally

	mm := explainedBy(c, keyMatchMedia)
	cascade := explainedBy(c, keyMatchMedia, keyGetComputedStyle)

	applied, failed := mediaPathReadStylesheet(r, &s, cascade)
	t.fold(applied, failed, cascade)
	applied, failed = mediaPathReadComplement(r, &s, mm)
	t.fold(applied, failed, mm)

	s.Determination = t.determination()
	switch s.Determination {
	case Inconclusive:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "nothing could be compared",
			Note:  "too little of what this section needs was collected or established as valid",
		})
	case Contradiction:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: mediaPathCount(t.unexplained) + " of " + mediaPathCount(t.applied) + " requirements did not hold",
			Note: "matchMedia, the cascade and the viewport this browser reports through JavaScript are not consistent with each other." +
				partlyExplainedNote(t.explained),
		})
	case Instrumented:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: mediaPathCount(t.explained) + " of " + mediaPathCount(t.applied) + " requirements did not hold",
			Note:  explainedConclusion,
		})
	default:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "matchMedia, the cascade and the JavaScript-read viewport agree",
			Note:  mediaPathCount(t.applied) + " requirements applied",
		})
	}
	return s
}

const mediaPathMaxEntries = 500

const mediaPathMaxRows = 8

func mediaPathReadStylesheet(r Request, s *Section, cascade explanation) (applied, failed int) {
	raw, ok := r.value("media.stylesheet")
	if !ok {
		s.Rows = append(s.Rows, Row{Label: "stylesheet probe", Value: "not collected", Note: "the collector did not report it"})
		return 0, 0
	}
	controlOk, _ := boolean(raw, "controlOk")
	if !controlOk {
		s.Rows = append(s.Rows, Row{
			Label: "stylesheet read-back",
			Value: "not established",
			Note:  "an unconditional declaration on the probe's own selector did not read back through getComputedStyle, so nothing else this probe reported is trusted this run",
		})
		return 0, 0
	}
	widthValid, _ := boolean(raw, "widthValid")
	heightValid, _ := boolean(raw, "heightValid")
	if !widthValid {
		s.Rows = append(s.Rows, Row{Label: "width feature, stylesheet path", Value: "not established", Note: "a tautological min-width/max-width pair that must always match did not read back as matching"})
	}
	if !heightValid {
		s.Rows = append(s.Rows, Row{Label: "height feature, stylesheet path", Value: "not established", Note: "a tautological min-height/max-height pair that must always match did not read back as matching"})
	}

	rowsUsed := 0
	numeric, _ := field(raw, "numeric")
	if arr, isArr := numeric.([]any); isArr {
		if len(arr) > mediaPathMaxEntries {
			arr = arr[:mediaPathMaxEntries]
		}
		for _, e := range arr {
			feature, haveFeature := str(e, "feature")
			if !haveFeature {
				continue
			}
			validForFeature := (feature == "width" && widthValid) || (feature == "height" && heightValid)
			if !validForFeature {
				continue
			}
			js, haveJS := boolean(e, "jsMatches")
			css, haveCSS := boolean(e, "cssMatches")
			if !haveJS || !haveCSS {
				continue
			}
			applied++
			if js == css {
				continue
			}
			failed++
			if rowsUsed < mediaPathMaxRows {
				rowsUsed++
				op, _ := str(e, "op")
				px, _ := num(e, "px")
				s.Rows = append(s.Rows, Row{
					Label: "matchMedia vs the cascade",
					Value: "(" + op + "-" + feature + ": " + itoa(px) + "px): matchMedia said " + mediaPathBoolWord(js) + ", the cascade said " + mediaPathBoolWord(css),
					Note:  cascade.annotate("the same query, evaluated by two paths of the same engine, must agree"),
				})
			}
		}
	}

	discrete, _ := field(raw, "discrete")
	if arr, isArr := discrete.([]any); isArr {
		if len(arr) > mediaPathMaxEntries {
			arr = arr[:mediaPathMaxEntries]
		}
		type bucket struct {
			jsTrue, cssTrue []string
		}
		buckets := map[string]*bucket{}
		for _, e := range arr {
			feature, haveFeature := str(e, "feature")
			value, haveValue := str(e, "value")
			if !haveFeature || !haveValue {
				continue
			}
			b := buckets[feature]
			if b == nil {
				b = &bucket{}
				buckets[feature] = b
			}
			if js, have := boolean(e, "jsMatches"); have && js {
				b.jsTrue = append(b.jsTrue, value)
			}
			if css, have := boolean(e, "cssMatches"); have && css {
				b.cssTrue = append(b.cssTrue, value)
			}
		}
		features := make([]string, 0, len(buckets))
		for f := range buckets {
			features = append(features, f)
		}
		sort.Strings(features)
		for _, feature := range features {
			b := buckets[feature]
			if len(b.jsTrue) != 1 || len(b.cssTrue) != 1 {

				continue
			}
			applied++
			if b.jsTrue[0] == b.cssTrue[0] {
				continue
			}
			failed++
			if rowsUsed < mediaPathMaxRows {
				rowsUsed++
				s.Rows = append(s.Rows, Row{
					Label: "matchMedia vs the cascade",
					Value: clip(feature, 40) + ": matchMedia said " + clip(b.jsTrue[0], 40) + ", the cascade said " + clip(b.cssTrue[0], 40),
					Note:  cascade.annotate("this feature's values are mutually exclusive, and the same query evaluated by two paths of the same engine must agree"),
				})
			}
		}
	}
	return applied, failed
}

func mediaPathReadComplement(r Request, s *Section, mm explanation) (applied, failed int) {
	raw, ok := r.value("media.complement")
	if !ok {
		s.Rows = append(s.Rows, Row{Label: "complement probe", Value: "not collected", Note: "the collector did not report it"})
		return 0, 0
	}

	rowsUsed := 0
	complements, _ := field(raw, "complements")
	if arr, isArr := complements.([]any); isArr {
		if len(arr) > mediaPathMaxEntries {
			arr = arr[:mediaPathMaxEntries]
		}
		for _, e := range arr {
			m, haveM := boolean(e, "matches")
			n, haveN := boolean(e, "negationMatches")
			if !haveM || !haveN {
				continue
			}
			applied++
			if m != n {
				continue
			}
			failed++
			if rowsUsed < mediaPathMaxRows {
				rowsUsed++
				q, _ := str(e, "query")
				word := "both matched"
				if !m {
					word = "both failed"
				}
				s.Rows = append(s.Rows, Row{
					Label: "query and its negation",
					Value: clip(q, 80) + " and its negation " + word,
					Note:  mm.annotate("a query and its own negation cannot agree"),
				})
			}
		}
	}

	brackets, _ := field(raw, "brackets")
	if arr, isArr := brackets.([]any); isArr {
		if len(arr) > mediaPathMaxEntries {
			arr = arr[:mediaPathMaxEntries]
		}
		for _, e := range arr {
			feature, haveFeature := str(e, "feature")
			if !haveFeature {
				continue
			}
			check := func(label, key string, wantTrue bool) {
				v, have := boolean(e, key)
				if !have {
					return
				}
				applied++
				if v == wantTrue {
					return
				}
				failed++
				if rowsUsed < mediaPathMaxRows {
					rowsUsed++
					s.Rows = append(s.Rows, Row{
						Label: "bracket vs the JavaScript-read viewport",
						Value: clip(feature, 40) + " " + label,
						Note:  mm.annotate("the width/height media feature is defined as the same viewport quantity JavaScript reads directly, so a bracket around it cannot exclude it and a query just past it cannot include it"),
					})
				}
			}

			check("brackets the viewport a pixel below the value it reads, but did not match", "minInside", true)
			check("brackets the viewport a pixel above the value it reads, but did not match", "maxInside", true)
			check("excludes a value two pixels past the one it reads, but matched anyway", "minOutside", false)
			check("excludes a value two pixels before the one it reads, but matched anyway", "maxOutside", false)
		}
	}
	return applied, failed
}

func mediaPathBoolWord(b bool) string {
	if b {
		return "matches"
	}
	return "does not match"
}

func mediaPathCount(n int) string {
	return itoa(float64(n))
}
