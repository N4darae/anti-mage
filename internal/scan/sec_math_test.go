package scan

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func mathSection(r Request) Section {
	return sectionMath(r, Inputs{}, claim{})
}

func mathAnalyze(r Request) Report {
	sec := sectionMath(r, Inputs{}, claim{})
	sec.ID, sec.Title = "math", "Numeric built-ins against the specification"
	return AnalyzeWith(r, Inputs{}, []Section{sec})
}

func mathGoodExactJSON() map[string]string {
	out := map[string]string{}
	for _, g := range mathExactGroups {
		for _, c := range g.cases {
			out[c.key] = c.expected
		}
	}
	return out
}

func mathExactProbe(overrides map[string]string) string {
	m := mathGoodExactJSON()
	for k, v := range overrides {
		m[k] = v
	}
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func mathGoodRepeatJSON(keys []string, overrides map[string][2]string) string {
	out := map[string]map[string]string{}
	for _, k := range keys {
		out[k] = map[string]string{"a": "1", "b": "1"}
	}
	for k, pair := range overrides {
		out[k] = map[string]string{"a": pair[0], "b": pair[1]}
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestMathAllCorrectIsConsistent(t *testing.T) {
	r := probes(t, map[string]string{
		"math.exact":  ok(mathExactProbe(nil)),
		"math.repeat": ok(mathGoodRepeatJSON([]string{"sin", "sqrt", "round"}, nil)),
	})
	if d := mathSection(r).Determination; d != Consistent {
		t.Fatalf("math = %q, want consistent", d)
	}
}

func TestMathWrongExactResultIsContradiction(t *testing.T) {
	r := probes(t, map[string]string{
		"math.exact": ok(mathExactProbe(map[string]string{"round.halfNeg": "0"})),
	})
	if d := mathSection(r).Determination; d != Contradiction {
		t.Fatalf("math = %q, want contradiction: Math.round(-0.5) must be -0, not 0", d)
	}
	if rep := mathAnalyze(r); rep.Summary.BotLikeness == 0 {
		t.Errorf("botLikeness = 0 on a replaced Math.round")
	}
}

func TestMathEveryExactCaseIsChecked(t *testing.T) {
	for _, g := range mathExactGroups {
		for _, c := range g.cases {
			bad := "not-" + c.expected
			r := probes(t, map[string]string{
				"math.exact": ok(mathExactProbe(map[string]string{c.key: bad})),
			})
			if d := mathSection(r).Determination; d != Contradiction {
				t.Errorf("case %s: math = %q with a wrong value, want contradiction", c.key, d)
			}
		}
	}
}

func TestMathRepeatDisagreementIsContradiction(t *testing.T) {
	r := probes(t, map[string]string{
		"math.repeat": ok(mathGoodRepeatJSON([]string{"sqrt"}, map[string][2]string{
			"sin": {"0.644217687237691", "0.6442176872376909"},
		})),
	})
	if d := mathSection(r).Determination; d != Contradiction {
		t.Fatalf("math = %q, want contradiction: Math.sin disagreed with its own repeat", d)
	}
}

func TestMathAbsentProbesAreInconclusive(t *testing.T) {
	r := probes(t, map[string]string{})
	if d := mathSection(r).Determination; d != Inconclusive {
		t.Fatalf("math = %q, want inconclusive with no probes at all", d)
	}
}

func TestMathUnsupportedProbesAreInconclusive(t *testing.T) {
	r := probes(t, map[string]string{
		"math.exact":  `{"status":"unsupported","value":{"reason":"no"}}`,
		"math.repeat": `{"status":"unsupported","value":{"reason":"no"}}`,
	})
	if d := mathSection(r).Determination; d != Inconclusive {
		t.Fatalf("math = %q, want inconclusive on unsupported", d)
	}
}

func TestMathMissingCaseIsSilent(t *testing.T) {
	r := probes(t, map[string]string{
		"math.exact": ok(`{"abs.nan":"NaN"}`),
	})
	if d := mathSection(r).Determination; d == Contradiction {
		t.Fatalf("math = contradiction with only one correct case reported")
	}
}

func TestMathHostileInputNeverPanics(t *testing.T) {
	hostile := []string{
		`null`,
		`true`,
		`42`,
		`"just a string"`,
		`[]`,
		`{}`,
		`{"abs.nan":null}`,
		`{"abs.nan":42}`,
		`{"abs.nan":true}`,
		`{"abs.nan":["NaN"]}`,
		`{"abs.nan":{"nested":{"nested":{"nested":"NaN"}}}}`,
		`{"` + strings.Repeat("x", 5000) + `":"NaN"}`,
	}
	for _, v := range hostile {
		func() {
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("math.exact = %s panicked: %v", clip(v, 60), p)
				}
			}()
			r := probes(t, map[string]string{"math.exact": ok(v)})
			if d := mathSection(r).Determination; d == Contradiction {
				t.Errorf("math.exact = %s produced contradiction on hostile input", clip(v, 60))
			}
		}()
	}
	for _, v := range hostile {
		func() {
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("math.repeat = %s panicked: %v", clip(v, 60), p)
				}
			}()
			r := probes(t, map[string]string{"math.repeat": ok(v)})
			if d := mathSection(r).Determination; d == Contradiction {
				t.Errorf("math.repeat = %s produced contradiction on hostile input", clip(v, 60))
			}
		}()
	}
}

func TestMathHugeWrongValueContradictsButStaysBounded(t *testing.T) {
	huge := strings.Repeat("N", 100000)
	r := probes(t, map[string]string{
		"math.exact": ok(`{"abs.nan":"` + huge + `"}`),
	})
	var sec Section
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("panicked on a huge wrong value: %v", p)
			}
		}()
		sec = mathSection(r)
	}()
	if sec.Determination != Contradiction {
		t.Fatalf("math = %q, want contradiction: a present, wrong value is not absence", sec.Determination)
	}
	for _, row := range sec.Rows {
		if len(row.Value) > 100 || len(row.Note) > 200 {
			t.Errorf("row %+v is not bounded despite a 100000-byte input", row)
		}
	}
}

func TestMathHostileRepeatPairsAbstain(t *testing.T) {
	hostile := []string{
		`{"sin":{"a":"1"}}`,
		`{"sin":{"a":1,"b":1}}`,
		`{"sin":{"a":null,"b":"1"}}`,
		`{"sin":"not an object"}`,
		`{"sin":{"a":{"nested":true},"b":"1"}}`,
	}
	for _, v := range hostile {
		r := probes(t, map[string]string{"math.repeat": ok(v)})
		if d := mathSection(r).Determination; d == Contradiction {
			t.Errorf("math.repeat = %s produced contradiction on a hostile pair", clip(v, 60))
		}
	}
}

func TestMathErrorStatusIsInconclusive(t *testing.T) {
	r := probes(t, map[string]string{})
	r.Probes["math.exact"] = Probe{Status: StatusError, Value: []byte(`"threw"`)}
	r.Probes["math.repeat"] = Probe{Status: StatusError, Value: []byte(`"threw"`)}
	if d := mathSection(r).Determination; d != Inconclusive {
		t.Fatalf("math = %q, want inconclusive on error status", d)
	}
}

func TestMathRepeatRowsAreBounded(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{`)
	for i := 0; i < 500; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `"k%d":{"a":"1","b":"2"}`, i)
	}
	b.WriteString(`}`)
	sec := sectionMath(probes(t, map[string]string{"math.repeat": ok(b.String())}), Inputs{}, claim{})
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: 500 pairs disagreed with themselves", sec.Determination)
	}
	if len(sec.Rows) > mathMaxRows+2 {
		t.Errorf("%d rows from 500 disagreements; the cap is %d plus the conclusion", len(sec.Rows), mathMaxRows)
	}
	for _, row := range sec.Rows {
		if row.Label == "conclusion" && !strings.Contains(row.Note, "500 repeat disagreement(s)") {
			t.Errorf("the conclusion stopped counting past the row cap: %q", row.Note)
		}
	}
}
