package scan

import "testing"

func sectionByID(rep Report, id string) Section {
	for _, s := range rep.Sections {
		if s.ID == id {
			return s
		}
	}
	return Section{}
}

func TestAReadingThatNeverWeighsAnythingIsNeverCountedAgainstTheEnvironment(t *testing.T) {
	reported := Request{Probes: map[string]Probe{
		"media.matrix": {Status: StatusOK, Value: []byte(`{"video/mp4; codecs=avc1":"probably"}`)},
	}}
	absent := Request{Probes: map[string]Probe{}}
	unsup := Request{Probes: map[string]Probe{"media.matrix": {Status: StatusUnsupported}}}
	broken := Request{Probes: map[string]Probe{"media.matrix": {Status: StatusOK, Value: []byte(`"not a map"`)}}}

	for name, r := range map[string]Request{"reported": reported, "not collected": absent, "unsupported": unsup, "unreadable": broken} {
		got := sectionCapabilities(r, Inputs{}, claim{}).Determination
		if got != Unverified {
			t.Errorf("%s: determination = %q, want %q; a reading that draws no conclusion from any answer must not turn the absence of an answer into one",
				name, got, Unverified)
		}
	}

	withMatrix := AnalyzeWith(reported, Inputs{}, nil).Summary
	withoutMatrix := AnalyzeWith(absent, Inputs{}, nil).Summary
	if withMatrix.HumanConfidence != withoutMatrix.HumanConfidence || withMatrix.BotLikeness != withoutMatrix.BotLikeness {
		t.Errorf("reporting the codec matrix moved the summary from %+v to %+v; the reading is weighed either way or neither way",
			withoutMatrix, withMatrix)
	}
}

func TestAProbeReportingItselfUnsupportedCannotRaiseTheScore(t *testing.T) {
	base := map[string]Probe{
		"scope.main": {Status: StatusOK, Value: []byte(
			`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/126","platform":"Win32"}`)},
	}
	clean := Request{Probes: map[string]Probe{}}
	for k, v := range base {
		clean.Probes[k] = v
	}
	suppressed := Request{Probes: map[string]Probe{}}
	for k, v := range base {
		suppressed.Probes[k] = v
	}
	suppressed.Probes["scope.worker"] = Probe{Status: StatusUnsupported, Value: []byte(`{"reason":"no worker here"}`)}

	cleanRep := AnalyzeWith(clean, Inputs{ElapsedMS: 1000}, nil)
	suppRep := AnalyzeWith(suppressed, Inputs{ElapsedMS: 1000}, nil)

	if got := sectionByID(suppRep, "claims").Determination; got == Consistent {
		t.Errorf("the claims reading answered %q for a probe it could neither confirm nor contradict", got)
	}
	if got, want := determinedSections(suppRep), determinedSections(cleanRep); got > want {
		t.Errorf("suppressing a probe raised the count of sections that reached a determination from %d to %d; a probe that reports itself unsupported must not be worth more than one that answers",
			want, got)
	}
}

func determinedSections(rep Report) int {
	n := 0
	for _, s := range rep.Sections {
		switch s.Determination {
		case Consistent, Contradiction, Instrumented:
			n++
		}
	}
	return n
}
