package scan

import "testing"

func mediaPathRun(t *testing.T, kv map[string]string) Section {
	t.Helper()
	r := probes(t, kv)
	return sectionMediaPaths(r, Inputs{}, claim{})
}

func TestMediaPathsAbsentIsInconclusive(t *testing.T) {
	s := mediaPathRun(t, map[string]string{})
	if s.Determination != Inconclusive {
		t.Fatalf("determination = %q, want inconclusive on an empty payload", s.Determination)
	}
}

func TestMediaPathsUnsupportedIsInconclusive(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.stylesheet": `{"status":"unsupported","value":{"reason":"no document"}}`,
		"media.complement": `{"status":"unsupported","value":{"reason":"no matchMedia"}}`,
	})
	if s.Determination != Inconclusive {
		t.Fatalf("determination = %q, want inconclusive when both probes are unsupported", s.Determination)
	}
}

func TestMediaPathsErrorStatusIsInconclusive(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.stylesheet": `{"status":"error","value":{"reason":"threw"}}`,
	})
	if s.Determination == Contradiction || s.Determination == Instrumented {
		t.Fatalf("determination = %q on an error status; an error is not evidence", s.Determination)
	}
}

func TestMediaPathsControlNotOkAbstainsEvenOnDisagreement(t *testing.T) {

	s := mediaPathRun(t, map[string]string{
		"media.stylesheet": ok(`{
			"controlOk": false, "widthValid": true, "heightValid": true,
			"numeric": [{"feature":"width","op":"min","px":320,"jsMatches":true,"cssMatches":false}],
			"discrete": []
		}`),
	})
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction, but the read-back control itself did not read back correctly")
	}
}

func TestMediaPathsFeatureTautologyNotEstablishedAbstainsThatFeature(t *testing.T) {

	s := mediaPathRun(t, map[string]string{
		"media.stylesheet": ok(`{
			"controlOk": true, "widthValid": false, "heightValid": true,
			"numeric": [{"feature":"width","op":"min","px":320,"jsMatches":true,"cssMatches":false}],
			"discrete": []
		}`),
	})
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction, but width's own tautology was not established")
	}
	if s.Determination != Inconclusive {
		t.Fatalf("determination = %q, want inconclusive: nothing else was reported", s.Determination)
	}
}

func TestMediaPathsNumericAgreementIsConsistent(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.stylesheet": ok(`{
			"controlOk": true, "widthValid": true, "heightValid": true,
			"numeric": [
				{"feature":"width","op":"min","px":320,"jsMatches":true,"cssMatches":true},
				{"feature":"width","op":"max","px":320,"jsMatches":false,"cssMatches":false},
				{"feature":"height","op":"min","px":400,"jsMatches":true,"cssMatches":true}
			],
			"discrete": []
		}`),
	})
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent: every entry agrees and validity was established", s.Determination)
	}
}

func TestMediaPathsNumericDisagreementIsContradiction(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.stylesheet": ok(`{
			"controlOk": true, "widthValid": true, "heightValid": true,
			"numeric": [
				{"feature":"width","op":"min","px":320,"jsMatches":true,"cssMatches":false}
			],
			"discrete": []
		}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: matchMedia and the cascade disagree on an established feature", s.Determination)
	}
}

func TestMediaPathsNumericMissingSideAbstainsThatEntry(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.stylesheet": ok(`{
			"controlOk": true, "widthValid": true, "heightValid": true,
			"numeric": [
				{"feature":"width","op":"min","px":320,"jsMatches":true,"cssMatches":null}
			],
			"discrete": []
		}`),
	})
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction, but the CSS side of that one entry was not reported")
	}
}

func TestMediaPathsDiscreteAgreementIsConsistent(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.stylesheet": ok(`{
			"controlOk": true, "widthValid": true, "heightValid": true,
			"numeric": [],
			"discrete": [
				{"feature":"orientation","value":"portrait","jsMatches":true,"cssMatches":true},
				{"feature":"orientation","value":"landscape","jsMatches":false,"cssMatches":false}
			]
		}`),
	})
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent", s.Determination)
	}
}

func TestMediaPathsDiscreteDisagreementIsContradiction(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.stylesheet": ok(`{
			"controlOk": true, "widthValid": true, "heightValid": true,
			"numeric": [],
			"discrete": [
				{"feature":"orientation","value":"portrait","jsMatches":true,"cssMatches":false},
				{"feature":"orientation","value":"landscape","jsMatches":false,"cssMatches":true}
			]
		}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: matchMedia and the cascade select different enumerated values", s.Determination)
	}
}

func TestMediaPathsDiscreteAmbiguousCSSSideAbstains(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.stylesheet": ok(`{
			"controlOk": true, "widthValid": true, "heightValid": true,
			"numeric": [],
			"discrete": [
				{"feature":"hover","value":"hover","jsMatches":true,"cssMatches":true},
				{"feature":"hover","value":"none","jsMatches":false,"cssMatches":true}
			]
		}`),
	})
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction, but the CSS side never settled on exactly one value")
	}
}

func TestMediaPathsDiscreteNoSideTrueAbstains(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.stylesheet": ok(`{
			"controlOk": true, "widthValid": true, "heightValid": true,
			"numeric": [],
			"discrete": [
				{"feature":"pointer","value":"fine","jsMatches":false,"cssMatches":false},
				{"feature":"pointer","value":"coarse","jsMatches":false,"cssMatches":false},
				{"feature":"pointer","value":"none","jsMatches":false,"cssMatches":false}
			]
		}`),
	})
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction, but neither side settled on any matching value")
	}
}

func TestMediaPathsComplementAgreementIsConsistent(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.complement": ok(`{
			"innerWidth": null, "innerHeight": null,
			"complements": [
				{"query":"(min-width: 1px)","matches":true,"negationMatches":false},
				{"query":"(min-width: 999999px)","matches":false,"negationMatches":true}
			],
			"brackets": []
		}`),
	})
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent", s.Determination)
	}
}

func TestMediaPathsComplementBothMatchIsContradiction(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.complement": ok(`{
			"innerWidth": null, "innerHeight": null,
			"complements": [
				{"query":"(min-width: 1px)","matches":true,"negationMatches":true}
			],
			"brackets": []
		}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: a query and its negation cannot both match", s.Determination)
	}
}

func TestMediaPathsComplementBothFailIsContradiction(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.complement": ok(`{
			"innerWidth": null, "innerHeight": null,
			"complements": [
				{"query":"(min-width: 1px)","matches":false,"negationMatches":false}
			],
			"brackets": []
		}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: a query and its negation cannot both fail", s.Determination)
	}
}

func TestMediaPathsComplementMissingSideAbstains(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.complement": ok(`{
			"innerWidth": null, "innerHeight": null,
			"complements": [
				{"query":"(min-width: 1px)","matches":true,"negationMatches":null}
			],
			"brackets": []
		}`),
	})
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction, but the negation side was never reported")
	}
}

func TestMediaPathsBracketHoldingIsConsistent(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.complement": ok(`{
			"innerWidth": 1024, "innerHeight": 700,
			"complements": [],
			"brackets": [
				{"feature":"width","value":1024,"insideBelowPx":1023,"insideAbovePx":1025,
				 "minInside":true,"maxInside":true,"minOutside":false,"maxOutside":false}
			]
		}`),
	})
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent", s.Determination)
	}
}

func TestMediaPathsBracketExcludingTheViewportIsContradiction(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.complement": ok(`{
			"innerWidth": 1024, "innerHeight": 700,
			"complements": [],
			"brackets": [
				{"feature":"width","value":1024,"insideBelowPx":1023,"insideAbovePx":1025,
				 "minInside":false,"maxInside":true,"minOutside":false,"maxOutside":false}
			]
		}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: a bracket a pixel below the value read must hold", s.Determination)
	}
}

func TestMediaPathsBracketIncludingAValueItMustExcludeIsContradiction(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.complement": ok(`{
			"innerWidth": 1024, "innerHeight": 700,
			"complements": [],
			"brackets": [
				{"feature":"width","value":1024,"insideBelowPx":1023,"insideAbovePx":1025,
				 "minInside":true,"maxInside":true,"minOutside":true,"maxOutside":false}
			]
		}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: two pixels past the value read must not match", s.Determination)
	}
}

func TestMediaPathsRoundedViewportIsNotAContradiction(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.complement": ok(`{
			"innerWidth": 1003, "innerHeight": 716,
			"complements": [],
			"brackets": [
				{"feature":"width","value":1003,"insideBelowPx":1002,"insideAbovePx":1004,
				 "minInside":true,"maxInside":true,"minOutside":false,"maxOutside":false}
			]
		}`),
	})
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent: rounding is not deception", s.Determination)
	}
}

func TestMediaPathsBracketNullSideAbstainsThatSide(t *testing.T) {
	s := mediaPathRun(t, map[string]string{
		"media.complement": ok(`{
			"innerWidth": 1, "innerHeight": 1,
			"complements": [],
			"brackets": [
				{"feature":"width","value":1,"insideBelowPx":0,"insideAbovePx":2,
				 "minInside":true,"maxInside":true,"minOutside":null,"maxOutside":null}
			]
		}`),
	})
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent: an unasked side is not a disagreement", s.Determination)
	}
}

func TestMediaPathsHostileShapesNeverPanicOrAccuse(t *testing.T) {
	hostile := []string{
		`{}`,
		`{"probes":null}`,
		`{"probes":{}}`,
		`{"probes":{"media.stylesheet":{"status":"ok"}}}`,
		`{"probes":{"media.stylesheet":{"status":"ok","value":null}}}`,
		`{"probes":{"media.stylesheet":{"status":"ok","value":0}}}`,
		`{"probes":{"media.stylesheet":{"status":"ok","value":true}}}`,
		`{"probes":{"media.stylesheet":{"status":"ok","value":"not an object"}}}`,
		`{"probes":{"media.stylesheet":{"status":"ok","value":[1,2,3]}}}`,
		`{"probes":{"media.stylesheet":{"status":"ok","value":{"numeric":"nope","discrete":123}}}}`,
		`{"probes":{"media.stylesheet":{"status":"ok","value":{"controlOk":true,"widthValid":true,"heightValid":true,"numeric":[null,1,"x",true,[],{}]}}}}`,
		`{"probes":{"media.stylesheet":{"status":"ok","value":{"controlOk":true,"widthValid":true,"heightValid":true,"numeric":[{"feature":123,"op":null,"px":"x","jsMatches":"yes","cssMatches":[]}]}}}}`,
		`{"probes":{"media.stylesheet":{"status":"ok","value":{"controlOk":"yes","widthValid":1,"heightValid":null}}}}`,
		`{"probes":{"media.stylesheet":{"status":"ok","value":{"discrete":[{"feature":"orientation","value":123,"jsMatches":true,"cssMatches":true}]}}}}`,
		`{"probes":{"media.complement":{"status":"ok","value":"nope"}}}`,
		`{"probes":{"media.complement":{"status":"ok","value":{"complements":"nope","brackets":null}}}}`,
		`{"probes":{"media.complement":{"status":"ok","value":{"complements":[null,1,"x",{}],"brackets":[null,1,"x",{}]}}}}`,
		`{"probes":{"media.complement":{"status":"ok","value":{"brackets":[{"feature":"width","insideBelowPx":"yes","minInside":"true"}]}}}}`,
		`{"probes":{"media.stylesheet":{"status":"ok","value":{"a":{"b":{"c":{"d":{"e":{"f":1}}}}}}}}}`,
		`{"probes":{"media.stylesheet":{"status":"ok","value":{"controlOk":true,"widthValid":true,"heightValid":true,"numeric":[{"feature":"width","op":"` + string(make([]byte, 5000)) + `","px":1,"jsMatches":true,"cssMatches":false}]}}}}`,
	}
	for _, in := range hostile {
		r, err := DecodeRequest([]byte(in))
		if err != nil {
			continue
		}
		got := sectionMediaPaths(r, Inputs{}, claim{})
		if !validDetermination(got.Determination) {
			t.Errorf("input %.60s: invalid determination %q", in, got.Determination)
		}
	}
}

func TestMediaPathsHugeArrayIsBounded(t *testing.T) {
	entries := "["
	for i := 0; i < 5000; i++ {
		if i > 0 {
			entries += ","
		}
		entries += `{"feature":"width","op":"min","px":1,"jsMatches":true,"cssMatches":true}`
	}
	entries += "]"
	s := mediaPathRun(t, map[string]string{
		"media.stylesheet": ok(`{"controlOk":true,"widthValid":true,"heightValid":true,"numeric":` + entries + `,"discrete":[]}`),
	})
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent", s.Determination)
	}
}
