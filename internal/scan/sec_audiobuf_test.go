package scan

import "testing"

func audioViewsJSON(overrides string) string {
	base := `{
		"requested": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160},
		"rendered": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160, "durationSeconds": 0.02},
		"channelsServed": 2,
		"copyFromChannelAvailable": true,
		"views": {"compared": true, "agree": true, "sampleCount": 320, "differingSampleCount": 0, "maxAbsoluteDifference": 0}
	}`
	if overrides == "" {
		return base
	}
	return overrides
}

func audioRepeatJSON(overrides string) string {
	base := `{
		"secondRenderCompleted": true,
		"repeat": {"compared": true, "agree": true, "sampleCount": 320, "differingSampleCount": 0, "maxAbsoluteDifference": 0}
	}`
	if overrides == "" {
		return base
	}
	return overrides
}

func audioSection(t *testing.T, kv map[string]string) Section {
	t.Helper()
	r := probes(t, kv)
	return sectionAudioBuf(r, Inputs{}, claim{})
}

func TestAudioBufAllAgreeIsConsistent(t *testing.T) {
	s := audioSection(t, map[string]string{
		"audio.views":  ok(audioViewsJSON("")),
		"audio.repeat": ok(audioRepeatJSON("")),
	})
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent", s.Determination)
	}
	for _, row := range s.Rows {
		if row.Value == "disagree" {
			t.Errorf("row %q reported disagree in an all-agreeing payload", row.Label)
		}
	}
}

func TestAudioBufMissingChannelsServedAbstains(t *testing.T) {
	s := audioSection(t, map[string]string{
		"audio.views": ok(`{
			"requested": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160},
			"rendered": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160, "durationSeconds": 0.02},
			"copyFromChannelAvailable": true
		}`),
	})
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction with no channel count reported; absence is never evidence. rows: %+v", s.Rows)
	}
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent: the arithmetic invariants still held", s.Determination)
	}
}

func TestAudioBufZeroChannelsServedStillContradicts(t *testing.T) {
	s := audioSection(t, map[string]string{
		"audio.views": ok(`{
			"requested": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160},
			"rendered": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160, "durationSeconds": 0.02},
			"channelsServed": 0,
			"copyFromChannelAvailable": true
		}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: a buffer reporting two channels served none", s.Determination)
	}
}

func TestAudioBufEmptyRequestIsInconclusive(t *testing.T) {
	s := sectionAudioBuf(Request{Probes: map[string]Probe{}}, Inputs{}, claim{})
	if s.Determination != Inconclusive {
		t.Fatalf("determination = %q on an empty request, want inconclusive", s.Determination)
	}
	if s.Rows == nil {
		t.Errorf("rows is nil; wire contract requires an array")
	}
}

func TestAudioBufUnsupportedIsInconclusiveNeverContradiction(t *testing.T) {
	r := probes(t, map[string]string{})
	r.Probes["audio.views"] = Probe{Status: StatusUnsupported, Value: []byte(`{"reason":"OfflineAudioContext is not available"}`)}
	r.Probes["audio.repeat"] = Probe{Status: StatusUnsupported, Value: []byte(`{"reason":"OfflineAudioContext is not available"}`)}
	s := sectionAudioBuf(r, Inputs{}, claim{})
	if s.Determination == Contradiction || s.Determination == Instrumented {
		t.Fatalf("determination = %q on an unsupported probe; absence must never be evidence", s.Determination)
	}
	if s.Determination != Inconclusive {
		t.Fatalf("determination = %q, want inconclusive", s.Determination)
	}
}

func TestAudioBufErrorStatusIsNeverAFinding(t *testing.T) {
	r := probes(t, map[string]string{})
	r.Probes["audio.views"] = Probe{Status: StatusError, Value: []byte(`{"reason":"threw"}`)}
	r.Probes["audio.repeat"] = Probe{Status: StatusError, Value: []byte(`{"reason":"threw"}`)}
	s := sectionAudioBuf(r, Inputs{}, claim{})
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction on an error status; an error is not a claim")
	}
}

func TestAudioBufNoCopyFromChannelAbstainsOnViewsOnly(t *testing.T) {
	views := `{
		"requested": {"sampleRateHz": 8000, "numberOfChannels": 1, "lengthFrames": 160},
		"rendered": {"sampleRateHz": 8000, "numberOfChannels": 1, "lengthFrames": 160, "durationSeconds": 0.02},
		"channelsServed": 1,
		"copyFromChannelAvailable": false
	}`
	s := audioSection(t, map[string]string{"audio.views": ok(views)})
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent: arithmetic still holds even without copyFromChannel", s.Determination)
	}
	found := false
	for _, row := range s.Rows {
		if row.Label == "copyFromChannel against getChannelData" {
			found = true
			if row.Value != "not compared" {
				t.Errorf("value = %q, want \"not compared\" when copyFromChannel is unavailable", row.Value)
			}
		}
	}
	if !found {
		t.Errorf("no row reported the views comparison as abstained")
	}
}

func TestAudioBufViewsDisagreementIsContradiction(t *testing.T) {
	views := `{
		"requested": {"sampleRateHz": 8000, "numberOfChannels": 1, "lengthFrames": 160},
		"rendered": {"sampleRateHz": 8000, "numberOfChannels": 1, "lengthFrames": 160, "durationSeconds": 0.02},
		"channelsServed": 1,
		"copyFromChannelAvailable": true,
		"views": {"compared": true, "agree": false, "sampleCount": 160, "differingSampleCount": 3, "maxAbsoluteDifference": 0.0004}
	}`
	s := audioSection(t, map[string]string{"audio.views": ok(views)})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: getChannelData and copyFromChannel read the same channel data", s.Determination)
	}
}

func TestAudioBufDurationArithmeticMismatchIsContradiction(t *testing.T) {
	views := `{
		"requested": {"sampleRateHz": 8000, "numberOfChannels": 1, "lengthFrames": 160},
		"rendered": {"sampleRateHz": 8000, "numberOfChannels": 1, "lengthFrames": 160, "durationSeconds": 9},
		"channelsServed": 1,
		"copyFromChannelAvailable": false
	}`
	s := audioSection(t, map[string]string{"audio.views": ok(views)})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: duration must equal length/sampleRate", s.Determination)
	}
}

func TestAudioBufChannelsServedMismatchIsContradiction(t *testing.T) {
	views := `{
		"requested": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160},
		"rendered": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160, "durationSeconds": 0.02},
		"channelsServed": 1,
		"copyFromChannelAvailable": false
	}`
	s := audioSection(t, map[string]string{"audio.views": ok(views)})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: numberOfChannels must match channels served", s.Determination)
	}
}

func TestAudioBufLengthMismatchIsContradiction(t *testing.T) {
	views := `{
		"requested": {"sampleRateHz": 8000, "numberOfChannels": 1, "lengthFrames": 160},
		"rendered": {"sampleRateHz": 8000, "numberOfChannels": 1, "lengthFrames": 159, "durationSeconds": 0.019875},
		"channelsServed": 1,
		"copyFromChannelAvailable": false
	}`
	s := audioSection(t, map[string]string{"audio.views": ok(views)})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: rendered length must equal what was constructed", s.Determination)
	}
}

func TestAudioBufDurationEpsilonToleratesFloatRounding(t *testing.T) {
	views := `{
		"requested": {"sampleRateHz": 8000, "numberOfChannels": 1, "lengthFrames": 160},
		"rendered": {"sampleRateHz": 8000, "numberOfChannels": 1, "lengthFrames": 160, "durationSeconds": 0.020000000499},
		"channelsServed": 1,
		"copyFromChannelAvailable": false
	}`
	s := audioSection(t, map[string]string{"audio.views": ok(views)})
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent: a sub-microsecond rounding difference must not convict", s.Determination)
	}
}

func TestAudioBufRepeatMismatchNeverConvicts(t *testing.T) {
	repeat := `{
		"secondRenderCompleted": true,
		"repeat": {"compared": true, "agree": false, "sampleCount": 320, "differingSampleCount": 2, "maxAbsoluteDifference": 0.00001}
	}`
	s := audioSection(t, map[string]string{
		"audio.views":  ok(audioViewsJSON("")),
		"audio.repeat": ok(repeat),
	})
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction on a repeat-render mismatch; the specification does not guarantee offline rendering is deterministic, so this must never convict")
	}
	found := false
	for _, row := range s.Rows {
		if row.Label == "repeat render agreement" {
			found = true
			if row.Value != "two renders of one identical graph produced different output" {
				t.Errorf("value = %q", row.Value)
			}
		}
	}
	if !found {
		t.Errorf("no row reported the repeat comparison")
	}
}

func TestAudioBufRepeatSecondRenderIncompleteIsSilent(t *testing.T) {
	repeat := `{"secondRenderCompleted": false}`
	s := audioSection(t, map[string]string{
		"audio.views":  ok(audioViewsJSON("")),
		"audio.repeat": ok(repeat),
	})
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction when the second render never completed")
	}
}

func TestAudioBufNeverPanicsOnHostileInput(t *testing.T) {
	hostile := []string{
		`null`,
		`true`,
		`42`,
		`"just a string"`,
		`[]`,
		`{}`,
		`{"requested": null, "rendered": null}`,
		`{"requested": "nope", "rendered": 5}`,
		`{"rendered": {"sampleRateHz": "eight thousand"}}`,
		`{"rendered": {"sampleRateHz": 0, "lengthFrames": 1, "durationSeconds": 1}}`,
		`{"views": {"compared": true, "agree": "yes"}}`,
		`{"views": "not an object"}`,
		`{"copyFromChannelAvailable": "not a bool"}`,
		`{"channelsServed": "NaN"}`,
		`{"channelsServed": 1e999}`,
		`{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":"deep"}}}}}}}}}}}`,
		`{"requested": {"lengthFrames": ` + longNumberLiteral() + `}}`,
		`{"repeat": {"compared": true, "agree": null}}`,
		`{"secondRenderCompleted": "true"}`,
		`{"secondRenderCompleted": true, "repeat": null}`,
		`{"secondRenderCompleted": true, "repeat": {"compared": "yes"}}`,
	}
	for _, h := range hostile {
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					t.Errorf("input %s panicked: %v", h, rec)
				}
			}()
			r := probes(t, map[string]string{
				"audio.views":  ok(h),
				"audio.repeat": ok(h),
			})
			s := sectionAudioBuf(r, Inputs{}, claim{})
			if s.Determination == Contradiction {

				t.Errorf("input %s produced a contradiction from malformed or absent data", h)
			}
		}()
	}
}

func longNumberLiteral() string {
	digits := make([]byte, 400)
	for i := range digits {
		digits[i] = '9'
	}
	return string(digits)
}

func TestAudioBufHugeStringDoesNotBreakRows(t *testing.T) {
	huge := make([]byte, 5000)
	for i := range huge {
		huge[i] = 'x'
	}
	r := probes(t, map[string]string{})
	r.Probes["audio.views"] = Probe{Status: StatusUnsupported, Value: []byte(`{"reason":"` + string(huge) + `"}`)}
	s := sectionAudioBuf(r, Inputs{}, claim{})
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction on an unsupported report")
	}
	for _, row := range s.Rows {
		if len(row.Note) > 400 {
			t.Errorf("row %q note is %d bytes; a hostile payload must not make a row unbounded", row.Label, len(row.Note))
		}
	}
}

func audioViewsDisagreeJSON(scale string) string {
	return `{
		"requested": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160},
		"rendered": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160, "durationSeconds": 0.02},
		"channelsServed": 2,
		"copyFromChannelAvailable": true,
		"views": {"compared": true, "agree": false, "sampleCount": 320, "differingSampleCount": 318,
			"maxAbsoluteDifference": 0.00004363059997558594` + scale + `}
	}`
}

func TestAudioBufUniformScaleReadsAsInstrumented(t *testing.T) {
	s := audioSection(t, map[string]string{
		"audio.views": ok(audioViewsDisagreeJSON(`,
			"scale": {"fitted": true, "factor": 0.9999563507759233, "comparedSamples": 318,
				"maxRelativeResidual": 4.3081e-8, "ratioMin": 0.9999563076946744,
				"ratioMax": 0.9999563938571723, "zerosAltered": 0}`)),
		"audio.repeat": ok(audioRepeatJSON("")),
	})
	if s.Determination != Instrumented {
		t.Fatalf("determination = %q, want instrumented for a disagreement one factor explains. rows: %+v", s.Determination, s.Rows)
	}
}

func TestAudioBufPerSampleNoiseStaysContradiction(t *testing.T) {
	s := audioSection(t, map[string]string{
		"audio.views": ok(audioViewsDisagreeJSON(`,
			"scale": {"fitted": true, "factor": 1.0000012, "comparedSamples": 318,
				"maxRelativeResidual": 0.0004, "ratioMin": 0.9996, "ratioMax": 1.0004, "zerosAltered": 0}`)),
		"audio.repeat": ok(audioRepeatJSON("")),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction for noise no single factor explains. rows: %+v", s.Determination, s.Rows)
	}
}

func TestAudioBufAlteredZerosStaysContradiction(t *testing.T) {
	s := audioSection(t, map[string]string{
		"audio.views": ok(audioViewsDisagreeJSON(`,
			"scale": {"fitted": true, "factor": 0.9999563507759233, "comparedSamples": 318,
				"maxRelativeResidual": 4.3081e-8, "ratioMin": 0.9999563076946744,
				"ratioMax": 0.9999563938571723, "zerosAltered": 2}`)),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction where a zero on one path is not a zero on the other. rows: %+v", s.Determination, s.Rows)
	}
}

func TestAudioBufTooFewSamplesStaysContradiction(t *testing.T) {
	s := audioSection(t, map[string]string{
		"audio.views": ok(audioViewsDisagreeJSON(`,
			"scale": {"fitted": true, "factor": 0.99995, "comparedSamples": 8,
				"maxRelativeResidual": 1e-9, "ratioMin": 0.99995, "ratioMax": 0.99995, "zerosAltered": 0}`)),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction where too few samples were fitted. rows: %+v", s.Determination, s.Rows)
	}
}

func TestAudioBufOlderPayloadWithoutScaleStaysContradiction(t *testing.T) {
	s := audioSection(t, map[string]string{
		"audio.views":  ok(audioViewsDisagreeJSON("")),
		"audio.repeat": ok(audioRepeatJSON("")),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction for a payload that carries no fit at all. rows: %+v", s.Determination, s.Rows)
	}
}

func TestAudioBufUniformScaleDoesNotExcuseOtherInvariants(t *testing.T) {
	s := audioSection(t, map[string]string{
		"audio.views": ok(`{
			"requested": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160},
			"rendered": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160, "durationSeconds": 0.05},
			"channelsServed": 2,
			"copyFromChannelAvailable": true,
			"views": {"compared": true, "agree": false, "sampleCount": 320, "differingSampleCount": 318,
				"maxAbsoluteDifference": 0.00004363059997558594,
				"scale": {"fitted": true, "factor": 0.9999563507759233, "comparedSamples": 318,
					"maxRelativeResidual": 4.3081e-8, "ratioMin": 0.9999563076946744,
					"ratioMax": 0.9999563938571723, "zerosAltered": 0}}
		}`),
	})
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: a duration that does not match length over sample rate is not a scaling. rows: %+v", s.Determination, s.Rows)
	}
}
