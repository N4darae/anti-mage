package scan

import (
	"strings"
	"testing"
)

func rectsOf(t *testing.T, kv map[string]string) Section {
	t.Helper()
	r := probes(t, kv)
	return sectionRects(r, Inputs{}, claim{})
}

const consistentRectPayload = `{
	"shiftPx": 41,
	"base":     {"x":37,"y":53,"left":37,"top":53,"right":218,"bottom":150,"width":181,"height":97},
	"twin":     {"x":37,"y":53,"left":37,"top":53,"right":218,"bottom":150,"width":181,"height":97},
	"shifted":  {"x":78,"y":53,"left":78,"top":53,"right":259,"bottom":150,"width":181,"height":97},
	"restored": {"x":37,"y":53,"left":37,"top":53,"right":218,"bottom":150,"width":181,"height":97}
}`

const consistentTextPayload = `{
	"empty": {"width": 0, "box": {
		"actualBoundingBoxLeft": 0, "actualBoundingBoxRight": 0,
		"actualBoundingBoxAscent": 0, "actualBoundingBoxDescent": 0,
		"fontBoundingBoxAscent": 29, "fontBoundingBoxDescent": 8,
		"emHeightAscent": 24, "emHeightDescent": 6,
		"hangingBaseline": 21.6, "alphabeticBaseline": 0, "ideographicBaseline": -8
	}},
	"full": {"width": 123.4, "box": {
		"actualBoundingBoxLeft": 0, "actualBoundingBoxRight": 123.4,
		"actualBoundingBoxAscent": 21, "actualBoundingBoxDescent": 5,
		"fontBoundingBoxAscent": 29, "fontBoundingBoxDescent": 8,
		"emHeightAscent": 24, "emHeightDescent": 6,
		"hangingBaseline": 21.6, "alphabeticBaseline": 0, "ideographicBaseline": -8
	}},
	"repeat": {"width": 123.4},
	"prefixWidths": [0, 10, 20, 35, 60, 90, 123.4]
}`

func TestRectsConsistentPayloadIsConsistent(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"rect.identities": ok(consistentRectPayload),
		"text.metrics":    ok(consistentTextPayload),
	})
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent; rows: %+v", s.Determination, s.Rows)
	}
}

func TestRectsNeitherProbeIsInconclusive(t *testing.T) {
	s := rectsOf(t, map[string]string{})
	if s.Determination != Inconclusive {
		t.Fatalf("determination = %q, want inconclusive on an empty request", s.Determination)
	}
}

func TestRectsUnsupportedProbesAreInconclusive(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"rect.identities": `{"status":"unsupported","value":{"reason":"no layout"}}`,
		"text.metrics":    `{"status":"unsupported","value":{"reason":"no canvas"}}`,
	})
	if s.Determination != Inconclusive {
		t.Fatalf("determination = %q, want inconclusive on two unsupported probes", s.Determination)
	}
}

func TestRectsErrorStatusIsInconclusive(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"rect.identities": `{"status":"error","value":{"reason":"threw"}}`,
	})
	if s.Determination != Inconclusive {
		t.Fatalf("determination = %q, want inconclusive on an error status", s.Determination)
	}
}

func TestRectsTooFewFieldsIsInconclusive(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"rect.identities": ok(`{"base": {"width": 181}}`),
	})
	if s.Determination != Inconclusive {
		t.Fatalf("determination = %q, want inconclusive with only one field reported", s.Determination)
	}
}

func TestRectSelfConsistencyViolationIsContradiction(t *testing.T) {

	s := rectsOf(t, map[string]string{
		"rect.identities": ok(`{"base": {"x":0,"y":0,"left":0,"top":0,"right":179,"bottom":10,"width":181,"height":10}}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: right-left disagrees with width", s.Determination)
	}
}

func TestRectXNotEqualLeftIsContradiction(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"rect.identities": ok(`{"base": {"x":5,"y":0,"left":0,"top":0,"right":181,"bottom":10,"width":181,"height":10}}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: x disagrees with left although width is non-negative", s.Determination)
	}
}

func TestNegativeWidthRectIsExcludedNotConvicted(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"rect.identities": ok(`{"base": {"x":50,"y":0,"left":10,"top":0,"right":50,"bottom":10,"width":-40,"height":10}}`),
	})
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction on a negative-width rect, which the specification permits to diverge")
	}
}

func TestNegativeHeightRectIsExcludedNotConvicted(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"rect.identities": ok(`{"base": {"x":0,"y":50,"left":0,"top":10,"right":10,"bottom":50,"width":10,"height":-40}}`),
	})
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction on a negative-height rect, which the specification permits to diverge")
	}
}

func TestRectUnequalTwinWidthIsContradiction(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"rect.identities": ok(`{
			"base": {"width": 181, "height": 97},
			"twin": {"width": 190, "height": 97}
		}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: two identically styled elements reported different widths", s.Determination)
	}
}

func TestRectShiftByWrongAmountIsContradiction(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"rect.identities": ok(`{
			"shiftPx": 41,
			"base":     {"left":0,"right":181,"top":0,"bottom":10,"width":181,"height":10},
			"shifted":  {"left":30,"right":211,"top":0,"bottom":10,"width":181,"height":10},
			"restored": {"left":0,"right":181,"top":0,"bottom":10,"width":181,"height":10}
		}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: left moved by 30, not the reported shiftPx of 41", s.Determination)
	}
}

func TestRectShiftChangingWidthIsContradiction(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"rect.identities": ok(`{
			"shiftPx": 41,
			"base":     {"left":0,"right":181,"top":0,"bottom":10,"width":181,"height":10},
			"shifted":  {"left":41,"right":230,"top":0,"bottom":10,"width":189,"height":10},
			"restored": {"left":0,"right":181,"top":0,"bottom":10,"width":181,"height":10}
		}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: a pure horizontal translation changed width", s.Determination)
	}
}

func TestRectFailedRestoreIsContradiction(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"rect.identities": ok(`{
			"shiftPx": 41,
			"base":     {"left":0,"right":181,"top":0,"bottom":10,"width":181,"height":10},
			"shifted":  {"left":41,"right":222,"top":0,"bottom":10,"width":181,"height":10},
			"restored": {"left":5,"right":186,"top":0,"bottom":10,"width":181,"height":10}
		}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: translating back did not restore the original left", s.Determination)
	}
}

func TestTextNonZeroEmptyWidthIsContradiction(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"text.metrics": ok(`{"empty": {"width": 3.2}}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: measureText('') did not report zero width", s.Determination)
	}
}

func TestTextRepeatedMeasurementDisagreementIsContradiction(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"text.metrics": ok(`{"full": {"width": 100}, "repeat": {"width": 100.5}}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: the same string measured twice disagreed", s.Determination)
	}
}

func TestTextNonZeroAlphabeticBaselineIsContradiction(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"text.metrics": ok(`{"full": {"width": 10, "box": {"alphabeticBaseline": 2.5}}}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: alphabeticBaseline was not zero under textBaseline alphabetic", s.Determination)
	}
}

func TestTextFontLevelMetricDivergenceIsContradiction(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"text.metrics": ok(`{
			"empty": {"width": 0, "box": {"fontBoundingBoxAscent": 29, "fontBoundingBoxDescent": 8}},
			"full":  {"width": 50, "box": {"fontBoundingBoxAscent": 40, "fontBoundingBoxDescent": 8}}
		}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: fontBoundingBoxAscent depended on which string was measured", s.Determination)
	}
}

func TestTextPrefixNarrowingNeverConvicts(t *testing.T) {
	s := rectsOf(t, map[string]string{
		"text.metrics": ok(`{
			"empty":  {"width": 0, "box": {"alphabeticBaseline": 0}},
			"full":   {"width": 50, "box": {"alphabeticBaseline": 0}},
			"repeat": {"width": 50},
			"prefixWidths": [0, 20, 45, 40, 50]
		}`),
	})
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent: only the prefix widths narrowed, which is not a requirement", s.Determination)
	}
	found := false
	for _, row := range s.Rows {
		if row.Label == "text metrics: width across prefixes" {
			found = true
			if row.Value == "non-decreasing across every prefix measured" {
				t.Errorf("expected the narrowing to be reported, got %q", row.Value)
			}
		}
	}
	if !found {
		t.Errorf("expected a prefix-width row to be reported")
	}
}

func TestRectsNeverPanicsOnHostileShapes(t *testing.T) {
	hostile := []string{
		`null`,
		`true`,
		`42`,
		`"a string, not an object"`,
		`[]`,
		`[1,2,3]`,
		`{}`,
		`{"base": null}`,
		`{"base": "not an object"}`,
		`{"base": []}`,
		`{"base": {"width": "181"}}`,
		`{"base": {"width": null}}`,
		`{"base": {"width": true}}`,
		`{"base": {"width": {"nested": {"nested": {"nested": "deep"}}}}}`,
		`{"base": {"width": 1e400}}`,
		`{"shiftPx": "forty-one"}`,
		`{"shiftPx": null}`,
		`{"prefixWidths": "not an array"}`,
		`{"prefixWidths": [null, "x", {}, [], 1, 2]}`,
		`{"prefixWidths": [1e400, -1e400]}`,
		`{"empty": {"box": "not an object"}}`,
		`{"empty": {"box": []}}`,
		`{"empty": {"box": {"alphabeticBaseline": "zero"}}}`,
		`{"full": {"box": null}}`,
		`{"base": {"x":0,"y":0,"left":0,"top":0,"right":0,"bottom":0,"width":0,"height":0}, "twin": {"width": 1e300, "height": -1e300}}`,
		`"` + strings.Repeat("a", 5000) + `"`,
		`{"prefixWidths": [` + strings.TrimSuffix(strings.Repeat("0,", 5000), ",") + `]}`,
	}
	for _, h := range hostile {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("sectionRects panicked on %s: %v", h, r)
				}
			}()
			r := probes(t, map[string]string{
				"rect.identities": ok(h),
				"text.metrics":    ok(h),
			})
			sectionRects(r, Inputs{}, claim{})
		}()
	}
}

func TestRectDeeplyNestedHostilePayloadNeverPanics(t *testing.T) {
	nested := `{"base":`
	for i := 0; i < 500; i++ {
		nested += `{"width":`
	}
	nested += `1`
	for i := 0; i < 500; i++ {
		nested += `}`
	}
	nested += `}`
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("sectionRects panicked on deeply nested input: %v", r)
		}
	}()
	r := probes(t, map[string]string{"rect.identities": ok(nested)})
	sectionRects(r, Inputs{}, claim{})
}

func TestRectsEmptyRowsAreNeverNil(t *testing.T) {
	s := rectsOf(t, map[string]string{})
	if s.Rows == nil {
		t.Errorf("rows is nil; the wire contract expects an array")
	}
}

func TestRectsContradictsWhenTheIdeographicBaselineIsNotTheFontDescent(t *testing.T) {
	text := `{
		"empty": {"width": 0, "box": {
			"actualBoundingBoxLeft": 0, "actualBoundingBoxRight": 0,
			"actualBoundingBoxAscent": 0, "actualBoundingBoxDescent": 0,
			"fontBoundingBoxAscent": 29, "fontBoundingBoxDescent": 8,
			"emHeightAscent": 24, "emHeightDescent": 6,
			"hangingBaseline": 21.6, "alphabeticBaseline": 0, "ideographicBaseline": -8.234375
		}},
		"full": {"width": 123.4, "box": {
			"actualBoundingBoxLeft": 0, "actualBoundingBoxRight": 123.4,
			"actualBoundingBoxAscent": 21, "actualBoundingBoxDescent": 5,
			"fontBoundingBoxAscent": 29, "fontBoundingBoxDescent": 8,
			"emHeightAscent": 24, "emHeightDescent": 6,
			"hangingBaseline": 21.6, "alphabeticBaseline": 0, "ideographicBaseline": -8.234375
		}},
		"repeat": {"width": 123.4},
		"prefixWidths": [0, 10, 20, 35, 60, 90, 123.4]
	}`
	s := rectsOf(t, map[string]string{
		"rect.identities": ok(consistentRectPayload),
		"text.metrics":    ok(text),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want %q; rows: %+v", s.Determination, Contradiction, s.Rows)
	}
}

func TestRectsReadsNothingIntoAnAbsentFontDescent(t *testing.T) {
	text := `{
		"empty": {"width": 0, "box": {
			"actualBoundingBoxLeft": 0, "actualBoundingBoxRight": 0,
			"actualBoundingBoxAscent": 0, "actualBoundingBoxDescent": 0,
			"emHeightAscent": 24, "emHeightDescent": 6,
			"hangingBaseline": 21.6, "alphabeticBaseline": 0, "ideographicBaseline": -8.234375
		}},
		"repeat": {"width": 0},
		"prefixWidths": [0]
	}`
	s := rectsOf(t, map[string]string{
		"rect.identities": ok(consistentRectPayload),
		"text.metrics":    ok(text),
	})
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want %q: with no descent reported there is nothing to compare the baseline against; rows: %+v", s.Determination, Consistent, s.Rows)
	}
}
