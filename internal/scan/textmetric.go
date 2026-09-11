package scan

import (
	_ "time/tzdata"
)

const textMetricTolerance = 1e-6

var textMetricFontLevelKeys = []string{
	"fontBoundingBoxAscent", "fontBoundingBoxDescent",
	"emHeightAscent", "emHeightDescent",
	"hangingBaseline", "alphabeticBaseline", "ideographicBaseline",
}

type textMetrics struct {
	haveEmpty, haveFull bool

	emptyWidth, fullWidth, repeatWidth             float64
	haveEmptyWidth, haveFullWidth, haveRepeatWidth bool

	emptyBox, fullBox         map[string]any
	haveEmptyBox, haveFullBox bool

	prefixWidths     []float64
	havePrefixWidths bool
}

func readTextMetrics(v any) textMetrics {
	var m textMetrics
	var emptyObj, fullObj, repeatObj map[string]any
	var haveRepeat bool
	emptyObj, m.haveEmpty = object(v, "empty")
	fullObj, m.haveFull = object(v, "full")
	repeatObj, haveRepeat = object(v, "repeat")

	if m.haveEmpty {
		m.emptyWidth, m.haveEmptyWidth = num(emptyObj, "width")
		m.emptyBox, m.haveEmptyBox = object(emptyObj, "box")
	}
	if m.haveFull {
		m.fullWidth, m.haveFullWidth = num(fullObj, "width")
		m.fullBox, m.haveFullBox = object(fullObj, "box")
	}
	if haveRepeat {
		m.repeatWidth, m.haveRepeatWidth = num(repeatObj, "width")
	}
	m.prefixWidths, m.havePrefixWidths = numList(v, "prefixWidths")
	return m
}

func (m textMetrics) identities() [][]identity {
	var groups [][]identity
	if m.haveEmptyWidth {
		groups = append(groups, []identity{{"measureText of the empty string reports zero width", m.emptyWidth, 0}})
	}
	if m.haveFullWidth && m.haveRepeatWidth {
		groups = append(groups, []identity{{"the same string measured twice in the same context reports the same width", m.fullWidth, m.repeatWidth}})
	}
	if m.haveEmptyBox {
		groups = append(groups, baselines(m.emptyBox,
			"empty-string measurement: alphabeticBaseline is exactly zero",
			"empty-string measurement: ideographicBaseline sits exactly at the font's descent")...)
	}
	if m.haveFullBox {
		groups = append(groups, baselines(m.fullBox,
			"probe-string measurement: alphabeticBaseline is exactly zero",
			"probe-string measurement: ideographicBaseline sits exactly at the font's descent")...)
	}
	if m.haveEmptyBox && m.haveFullBox {
		groups = append(groups, fontLevelInvariants(m.emptyBox, m.fullBox))
	}
	return groups
}

func baselines(box map[string]any, alphabeticLabel, ideographicLabel string) [][]identity {
	var alphabetic, ideographic []identity
	if v, ok := num(box, "alphabeticBaseline"); ok {
		alphabetic = []identity{{alphabeticLabel, v, 0}}
	}
	if v, ok := num(box, "ideographicBaseline"); ok {
		if descent, haveDescent := num(box, "fontBoundingBoxDescent"); haveDescent {
			ideographic = []identity{{ideographicLabel, v, -descent}}
		}
	}
	return [][]identity{alphabetic, ideographic}
}

func fontLevelInvariants(a, b map[string]any) []identity {
	var ids []identity
	for _, k := range textMetricFontLevelKeys {
		av, aok := num(a, k)
		bv, bok := num(b, k)
		if !aok || !bok {
			continue
		}
		ids = append(ids, identity{"text metrics: " + k + " does not depend on which string was measured", av, bv})
	}
	return ids
}

func (s *Section) addPrefixWidthReading(widths []float64) {
	if len(widths) < 2 {
		return
	}
	decreases := 0
	for i := 1; i < len(widths); i++ {
		if widths[i] < widths[i-1]-textMetricTolerance {
			decreases++
		}
	}
	value := "non-decreasing across every prefix measured"
	if decreases > 0 {
		value = outOf(decreases, len(widths)-1) + " prefix extension(s) measured narrower"
	}
	s.add("text metrics: width across prefixes", value)
}
