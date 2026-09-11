package scan

import (
	_ "time/tzdata"
)

const mediaPathMaxRows = 8

func sectionMediaPaths(r Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}
	var t tally

	mm := explainedBy(c, keyMatchMedia)
	cascade := explainedBy(c, keyMatchMedia, keyGetComputedStyle)

	applied, failed := s.addStylesheetReading(r)
	t.fold(applied, failed, cascade)
	applied, failed = s.addComplementReading(r)
	t.fold(applied, failed, mm)

	s.settle(t, conclusions{
		inconclusive:  "nothing could be compared",
		contradiction: outOf(t.unexplained, t.applied) + " requirements did not hold",
		instrumented:  outOf(t.explained, t.applied) + " requirements did not hold",
		consistent:    "matchMedia, the cascade and the JavaScript-read viewport agree",
	})
	return s
}

func (s *Section) addStylesheetReading(r Request) (applied, failed int) {
	raw, ok := r.value("media.stylesheet")
	if !ok {
		s.add("stylesheet probe", "not collected")
		return 0, 0
	}
	sheet := readMediaStylesheet(raw)
	if !sheet.controlOk {
		s.add("stylesheet read-back", "not established")
		return 0, 0
	}
	if !sheet.widthValid {
		s.add("width feature, stylesheet path", "not established")
	}
	if !sheet.heightValid {
		s.add("height feature, stylesheet path", "not established")
	}

	rowsUsed := 0
	for _, e := range sheet.numeric {
		if !sheet.validFor(e.feature) || !e.answered {
			continue
		}
		applied++
		if e.js == e.css {
			continue
		}
		failed++
		if rowsUsed < mediaPathMaxRows {
			rowsUsed++
			s.add("matchMedia vs the cascade",
				"("+e.op+"-"+e.feature+": "+itoa(e.px)+"px): matchMedia said "+mediaPathBoolWord(e.js)+", the cascade said "+mediaPathBoolWord(e.css))
		}
	}

	for _, d := range sheet.discreteReadings() {
		applied++
		if !d.disagree {
			continue
		}
		failed++
		if rowsUsed < mediaPathMaxRows {
			rowsUsed++
			s.add("matchMedia vs the cascade",
				clip(d.feature, 40)+": matchMedia said "+clip(d.js, 40)+", the cascade said "+clip(d.css, 40))
		}
	}
	return applied, failed
}

func (s *Section) addComplementReading(r Request) (applied, failed int) {
	raw, ok := r.value("media.complement")
	if !ok {
		s.add("complement probe", "not collected")
		return 0, 0
	}
	mc := readMediaComplement(raw)

	rowsUsed := 0
	for _, e := range mc.complements {
		applied++
		if e.matches != e.negation {
			continue
		}
		failed++
		if rowsUsed < mediaPathMaxRows {
			rowsUsed++
			word := "both matched"
			if !e.matches {
				word = "both failed"
			}
			s.add("query and its negation", clip(e.query, 80)+" and its negation "+word)
		}
	}

	for _, b := range mc.brackets {
		for _, f := range b.flags {
			applied++
			if f.matched == f.wantTrue {
				continue
			}
			failed++
			if rowsUsed < mediaPathMaxRows {
				rowsUsed++
				s.add("bracket vs the JavaScript-read viewport", clip(b.feature, 40)+" "+f.what)
			}
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
