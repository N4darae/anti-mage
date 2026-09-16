package assess

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDeterminationsAreOrderedAndTotal(t *testing.T) {
	scale := []Determination{NotEvaluated, Insufficient, Coherent, Discrepant, Instrumented}
	for i, lower := range scale {
		for j, higher := range scale {
			if got, want := higher.AtLeast(lower), j >= i; got != want {
				t.Errorf("%q.AtLeast(%q) = %v, want %v", higher, lower, got, want)
			}
		}
	}

	unknown := Determination("something else")
	if unknown.AtLeast(Insufficient) {
		t.Errorf("an unrecognised determination satisfied a threshold")
	}
	if unknown.Established() {
		t.Errorf("an unrecognised determination claimed to be established")
	}
	if unknown.String() != "something else" {
		t.Errorf("String() = %q", unknown.String())
	}
}

func TestTheAssessmentCarriesNoBreakdown(t *testing.T) {
	a := Evaluate(Environment{
		Observations: with(map[string]Observation{"geom.css": ok(`{"dppx":2}`)}),
		Findings:     []Finding{{Name: "a reading of my own", Verdict: Contradiction}},
	})
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"v": true, "determination": true, "score": true, "statement": true, "supplied": true}
	for k := range out {
		if !want[k] {
			t.Errorf("the assessment carries %q; a caller must not be able to see which reading moved the score", k)
		}
	}
	for k := range want {
		if _, present := out[k]; !present {
			t.Errorf("the assessment carries no %q", k)
		}
	}

	for _, leak := range []string{"geom", "dppx", "a reading of my own", "natives", "fonts", "section"} {
		if strings.Contains(strings.ToLower(a.Statement), leak) {
			t.Errorf("statement %q names %q; it must say what the determination means and not which reading produced it", a.Statement, leak)
		}
	}
}
