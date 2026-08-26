package scan

import "math"

const rectTolerance = 1e-4

const textmTolerance = 1e-6

type rectPoint struct {
	x, y, left, top, right, bottom, width, height                                 float64
	haveX, haveY, haveLeft, haveTop, haveRight, haveBottom, haveWidth, haveHeight bool
}

func rectRead(v any, path ...string) (rectPoint, bool) {
	obj, ok := object(v, path...)
	if !ok {
		return rectPoint{}, false
	}
	var p rectPoint
	p.x, p.haveX = num(obj, "x")
	p.y, p.haveY = num(obj, "y")
	p.left, p.haveLeft = num(obj, "left")
	p.top, p.haveTop = num(obj, "top")
	p.right, p.haveRight = num(obj, "right")
	p.bottom, p.haveBottom = num(obj, "bottom")
	p.width, p.haveWidth = num(obj, "width")
	p.height, p.haveHeight = num(obj, "height")
	return p, true
}

func rectAssertEqual(s *Section, label string, got, want float64) (applied, failed int) {
	applied = 1
	if math.Abs(got-want) > rectTolerance {
		failed = 1
		s.Rows = append(s.Rows, Row{
			Label: label,
			Value: "does not hold",
			Note:  "computed " + itoa(got) + ", expected " + itoa(want) + " from this rect's own other fields",
		})
	}
	return
}

func rectCheckSelfConsistency(s *Section, label string, p rectPoint) (applied, failed int) {
	if p.haveWidth && p.width < 0 {
		return 0, 0
	}
	if p.haveHeight && p.height < 0 {
		return 0, 0
	}
	if p.haveRight && p.haveLeft && p.haveWidth {
		a, f := rectAssertEqual(s, label+": right minus left equals width", p.right-p.left, p.width)
		applied += a
		failed += f
	}
	if p.haveBottom && p.haveTop && p.haveHeight {
		a, f := rectAssertEqual(s, label+": bottom minus top equals height", p.bottom-p.top, p.height)
		applied += a
		failed += f
	}
	if p.haveX && p.haveLeft && p.haveWidth {
		a, f := rectAssertEqual(s, label+": x equals left", p.x, p.left)
		applied += a
		failed += f
	}
	if p.haveY && p.haveTop && p.haveHeight {
		a, f := rectAssertEqual(s, label+": y equals top", p.y, p.top)
		applied += a
		failed += f
	}
	return
}

func rectCheckEqualElements(s *Section, base, twin rectPoint) (applied, failed int) {
	if base.haveWidth && twin.haveWidth {
		a, f := rectAssertEqual(s, "two identically styled elements report equal width", base.width, twin.width)
		applied += a
		failed += f
	}
	if base.haveHeight && twin.haveHeight {
		a, f := rectAssertEqual(s, "two identically styled elements report equal height", base.height, twin.height)
		applied += a
		failed += f
	}
	return
}

func rectCheckShift(s *Section, base, shifted, restored rectPoint, shiftPx float64) (applied, failed int) {
	add := func(a, f int) {
		applied += a
		failed += f
	}
	if base.haveLeft && shifted.haveLeft {
		add(rectAssertEqual(s, "translating right by an exact amount moves left by that amount", shifted.left-base.left, shiftPx))
	}
	if base.haveRight && shifted.haveRight {
		add(rectAssertEqual(s, "translating right by an exact amount moves right by that amount", shifted.right-base.right, shiftPx))
	}
	if base.haveTop && shifted.haveTop {
		add(rectAssertEqual(s, "a horizontal translation leaves top unchanged", shifted.top, base.top))
	}
	if base.haveBottom && shifted.haveBottom {
		add(rectAssertEqual(s, "a horizontal translation leaves bottom unchanged", shifted.bottom, base.bottom))
	}
	if base.haveWidth && shifted.haveWidth {
		add(rectAssertEqual(s, "a horizontal translation leaves width unchanged", shifted.width, base.width))
	}
	if base.haveHeight && shifted.haveHeight {
		add(rectAssertEqual(s, "a horizontal translation leaves height unchanged", shifted.height, base.height))
	}
	if base.haveLeft && restored.haveLeft {
		add(rectAssertEqual(s, "translating back returns left to its original value", restored.left, base.left))
	}
	if base.haveTop && restored.haveTop {
		add(rectAssertEqual(s, "translating back returns top to its original value", restored.top, base.top))
	}
	if base.haveWidth && restored.haveWidth {
		add(rectAssertEqual(s, "translating back returns width to its original value", restored.width, base.width))
	}
	if base.haveHeight && restored.haveHeight {
		add(rectAssertEqual(s, "translating back returns height to its original value", restored.height, base.height))
	}
	return
}

func textmAssertEqual(s *Section, label string, got, want float64, e explanation) (applied, failed int) {
	applied = 1
	if math.Abs(got-want) > textmTolerance {
		failed = 1
		s.Rows = append(s.Rows, Row{
			Label: label,
			Value: "does not hold",
			Note:  e.annotate("computed " + itoa(got) + ", expected " + itoa(want)),
		})
	}
	return
}

var textmFontLevelKeys = []string{
	"fontBoundingBoxAscent", "fontBoundingBoxDescent",
	"emHeightAscent", "emHeightDescent",
	"hangingBaseline", "alphabeticBaseline", "ideographicBaseline",
}

func textmCheckFontLevelInvariant(s *Section, a, b map[string]any, e explanation) (applied, failed int) {
	for _, k := range textmFontLevelKeys {
		av, aok := num(a, k)
		bv, bok := num(b, k)
		if !aok || !bok {
			continue
		}
		ap, fp := textmAssertEqual(s, "text metrics: "+k+" does not depend on which string was measured", av, bv, e)
		applied += ap
		failed += fp
	}
	return
}

func textmCheckAlphabeticBaseline(s *Section, label string, box map[string]any, e explanation) (applied, failed int) {
	v, ok := num(box, "alphabeticBaseline")
	if !ok {
		return 0, 0
	}
	return textmAssertEqual(s, label, v, 0, e)
}

func textmReadWidths(v any, path ...string) ([]float64, bool) {
	x, ok := field(v, path...)
	if !ok {
		return nil, false
	}
	arr, ok := x.([]any)
	if !ok {
		return nil, false
	}
	out := make([]float64, 0, len(arr))
	for _, e := range arr {
		f, isNum := e.(float64)
		if !isNum || math.IsNaN(f) || math.IsInf(f, 0) {
			continue
		}
		out = append(out, f)
	}
	return out, true
}

func textmReportMonotonicity(s *Section, widths []float64) {
	if len(widths) < 2 {
		return
	}
	decreases := 0
	for i := 1; i < len(widths); i++ {
		if widths[i] < widths[i-1]-textmTolerance {
			decreases++
		}
	}
	value := "non-decreasing across every prefix measured"
	if decreases > 0 {
		value = itoa(float64(decreases)) + " of " + itoa(float64(len(widths)-1)) + " prefix extension(s) measured narrower"
	}
	s.Rows = append(s.Rows, Row{
		Label: "text metrics: width across prefixes",
		Value: value,
		Note:  "not a requirement: kerning and ligature substitution can legitimately narrow a longer prefix, so this is reported only and never affects the determination",
	})
}

func sectionRects(r Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}
	var t tally

	textm := explainedBy(c, keyMeasureText)

	if v, ok := r.value("rect.identities"); ok {
		base, haveBase := rectRead(v, "base")
		twin, haveTwin := rectRead(v, "twin")
		shifted, haveShifted := rectRead(v, "shifted")
		restored, haveRestored := rectRead(v, "restored")
		shiftPx, haveShiftPx := num(v, "shiftPx")

		for _, rp := range []struct {
			label string
			p     rectPoint
			have  bool
		}{
			{"base rect", base, haveBase},
			{"twin rect", twin, haveTwin},
			{"shifted rect", shifted, haveShifted},
			{"restored rect", restored, haveRestored},
		} {
			if !rp.have {
				continue
			}
			t.foldPlain(rectCheckSelfConsistency(&s, rp.label, rp.p))
		}
		if haveBase && haveTwin {
			t.foldPlain(rectCheckEqualElements(&s, base, twin))
		}
		if haveBase && haveShifted && haveRestored && haveShiftPx {
			t.foldPlain(rectCheckShift(&s, base, shifted, restored, shiftPx))
		}
	} else {
		s.Rows = append(s.Rows, Row{Label: "rect geometry", Value: "not collected", Note: "the collector did not report a usable rect.identities probe"})
	}

	if v, ok := r.value("text.metrics"); ok {
		emptyObj, haveEmptyObj := object(v, "empty")
		fullObj, haveFullObj := object(v, "full")
		repeatObj, haveRepeatObj := object(v, "repeat")

		var emptyWidth, fullWidth, repeatWidth float64
		var haveEmptyWidth, haveFullWidth, haveRepeatWidth bool
		var emptyBox, fullBox map[string]any
		var haveEmptyBox, haveFullBox bool

		if haveEmptyObj {
			emptyWidth, haveEmptyWidth = num(emptyObj, "width")
			emptyBox, haveEmptyBox = object(emptyObj, "box")
		}
		if haveFullObj {
			fullWidth, haveFullWidth = num(fullObj, "width")
			fullBox, haveFullBox = object(fullObj, "box")
		}
		if haveRepeatObj {
			repeatWidth, haveRepeatWidth = num(repeatObj, "width")
		}

		if haveEmptyWidth {
			a, f := textmAssertEqual(&s, "measureText of the empty string reports zero width", emptyWidth, 0, textm)
			t.fold(a, f, textm)
		}
		if haveFullWidth && haveRepeatWidth {
			a, f := textmAssertEqual(&s, "the same string measured twice in the same context reports the same width", fullWidth, repeatWidth, textm)
			t.fold(a, f, textm)
		}
		if haveEmptyBox {
			a, f := textmCheckAlphabeticBaseline(&s, "empty-string measurement: alphabeticBaseline is exactly zero", emptyBox, textm)
			t.fold(a, f, textm)
		}
		if haveFullBox {
			a, f := textmCheckAlphabeticBaseline(&s, "probe-string measurement: alphabeticBaseline is exactly zero", fullBox, textm)
			t.fold(a, f, textm)
		}
		if haveEmptyBox && haveFullBox {
			a, f := textmCheckFontLevelInvariant(&s, emptyBox, fullBox, textm)
			t.fold(a, f, textm)
		}
		if widths, ok := textmReadWidths(v, "prefixWidths"); ok {
			textmReportMonotonicity(&s, widths)
		}
		if !haveEmptyObj && !haveFullObj {
			s.Rows = append(s.Rows, Row{Label: "text metrics", Value: "not collected", Note: "the collector reported text.metrics without a usable empty or full measurement"})
		}
	} else {
		s.Rows = append(s.Rows, Row{Label: "text metrics", Value: "not collected", Note: "the collector did not report a usable text.metrics probe"})
	}

	s.Determination = t.determination()
	switch s.Determination {
	case Inconclusive:
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "nothing could be compared", Note: "too few of these probes' fields were reported"})
	case Contradiction:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: itoa(float64(t.unexplained)) + " of " + itoa(float64(t.applied)) + " identities did not hold",
			Note: "the numbers this browser reported for its own layout or text measurement are not consistent with each other." +
				partlyExplainedNote(t.explained),
		})
	case Instrumented:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: itoa(float64(t.explained)) + " of " + itoa(float64(t.applied)) + " identities did not hold",
			Note:  explainedConclusion,
		})
	default:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "every identity checked held",
			Note:  itoa(float64(t.applied)) + " identities applied",
		})
	}
	return s
}
