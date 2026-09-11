package scan

import (
	"math"
	_ "time/tzdata"
)

type identity struct {
	label     string
	got, want float64
}

func (s *Section) checkIdentities(ids []identity, tolerance float64) (applied, failed int) {
	for _, id := range ids {
		applied++
		if math.Abs(id.got-id.want) > tolerance {
			failed++
			s.add(id.label, "does not hold")
		}
	}
	return applied, failed
}

func sectionRects(r Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}
	var t tally

	textm := explainedBy(c, keyMeasureText)
	rects := explainedBy(c, keyBoundingRect)

	if v, ok := r.value("rect.identities"); ok {
		for _, g := range readRectIdentities(v).identities() {
			a, f := s.checkIdentities(g, rectTolerance)
			t.fold(a, f, rects)
		}
	} else {
		s.add("rect geometry", "not collected")
	}

	if v, ok := r.value("text.metrics"); ok {
		m := readTextMetrics(v)
		for _, g := range m.identities() {
			a, f := s.checkIdentities(g, textMetricTolerance)
			t.fold(a, f, textm)
		}
		if m.havePrefixWidths {
			s.addPrefixWidthReading(m.prefixWidths)
		}
		if !m.haveEmpty && !m.haveFull {
			s.add("text metrics", "not collected")
		}
	} else {
		s.add("text metrics", "not collected")
	}

	s.settle(t, conclusions{
		inconclusive:  "nothing could be compared",
		contradiction: outOf(t.unexplained, t.applied) + " identities did not hold",
		instrumented:  outOf(t.explained, t.applied) + " identities did not hold",
		consistent:    "every identity checked held",
	})
	return s
}
