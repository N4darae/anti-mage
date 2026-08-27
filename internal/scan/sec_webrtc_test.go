package scan

import "testing"

func webrtcRequest(t *testing.T, finalState string, timedOut bool, sawStates, candidateTypes []string) Request {
	t.Helper()
	body := `{"finalState":"` + finalState + `","timedOut":` + jsonBool(timedOut) +
		`,"sawStates":` + jsonStringArray(sawStates) +
		`,"candidateTypes":` + jsonStringArray(candidateTypes) + `}`
	return Request{Probes: map[string]Probe{
		"webrtc.ice": {Status: StatusOK, Value: []byte(body)},
	}}
}

func jsonStringArray(xs []string) string {
	out := "["
	for i, x := range xs {
		if i > 0 {
			out += ","
		}
		out += `"` + x + `"`
	}
	return out + "]"
}

func TestWebRTCIsUnverifiedWhenNothingWasCollected(t *testing.T) {
	got := sectionWebRTC(Request{Probes: map[string]Probe{}}, Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a payload with nothing to read leaves this reading nothing to apply to", got.Determination, Unverified)
	}
}

func TestWebRTCIsUnverifiedWhenTheInterfaceIsReportedUnsupported(t *testing.T) {
	r := Request{Probes: map[string]Probe{
		"webrtc.ice": {Status: StatusUnsupported, Value: []byte(`{"reason":"no RTCPeerConnection"}`)},
	}}
	got := sectionWebRTC(r, Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a browser is never read as lacking a feature", got.Determination, Unverified)
	}
}

func TestWebRTCIsUnverifiedWhenTheGatheringStateIsMissing(t *testing.T) {
	r := Request{Probes: map[string]Probe{
		"webrtc.ice": {Status: StatusOK, Value: []byte(`{"timedOut":false,"sawStates":[],"candidateTypes":[]}`)},
	}}
	got := sectionWebRTC(r, Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: an incomplete payload settles nothing", got.Determination, Unverified)
	}
}

func TestWebRTCIsUnverifiedWithNoCandidatesAtAll(t *testing.T) {
	got := sectionWebRTC(webrtcRequest(t, "complete", false, []string{"gathering", "complete"}, nil), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: this reading never carries a verdict, including when a real environment produced no candidates at all", got.Determination, Unverified)
	}
}

func TestWebRTCIsUnverifiedWithHostCandidatesReported(t *testing.T) {
	got := sectionWebRTC(webrtcRequest(t, "complete", false, []string{"new", "gathering", "complete"}, []string{"host"}), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q", got.Determination, Unverified)
	}
}

func TestWebRTCIsUnverifiedOnTimeout(t *testing.T) {
	got := sectionWebRTC(webrtcRequest(t, "gathering", true, []string{"new", "gathering"}, nil), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: gathering that never finished settles nothing", got.Determination, Unverified)
	}
}

func TestWebRTCReportsCandidateCountsByType(t *testing.T) {
	got := sectionWebRTC(webrtcRequest(t, "complete", false, []string{"complete"}, []string{"host", "host", "srflx"}), Inputs{}, claim{})
	counts := map[string]string{}
	for _, row := range got.Rows {
		counts[row.Label] = row.Value
	}
	if counts["candidates of type host"] != "2" {
		t.Errorf("host count = %q, want %q", counts["candidates of type host"], "2")
	}
	if counts["candidates of type srflx"] != "1" {
		t.Errorf("srflx count = %q, want %q", counts["candidates of type srflx"], "1")
	}
	if counts["candidates of type relay"] != "0" {
		t.Errorf("relay count = %q, want %q", counts["candidates of type relay"], "0")
	}
}

func TestWebRTCNeverReachesContradictionOrInstrumented(t *testing.T) {
	cases := []Request{
		{Probes: map[string]Probe{}},
		{Probes: map[string]Probe{"webrtc.ice": {Status: StatusUnsupported, Value: []byte(`{}`)}}},
		webrtcRequest(t, "complete", false, []string{"gathering", "complete"}, nil),
		webrtcRequest(t, "complete", false, []string{"gathering", "complete"}, []string{"host"}),
		webrtcRequest(t, "gathering", true, []string{"new", "gathering"}, nil),
		webrtcRequest(t, "new", false, []string{}, []string{"host", "srflx", "relay"}),
	}
	for i, r := range cases {
		got := sectionWebRTC(r, Inputs{}, claim{})
		if got.Determination == Contradiction || got.Determination == Instrumented {
			t.Errorf("case %d: determination = %q, this reading must never reach a verdict", i, got.Determination)
		}
	}
}

func TestWebRTCDoesNotDiluteAPayloadThatPredatesIt(t *testing.T) {
	sections := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Inconclusive},
	}
	before := summarise(sections)

	absent := sectionWebRTC(Request{Probes: map[string]Probe{}}, Inputs{}, claim{})
	absent.ID = "webrtc"
	after := summarise(append(sections, normalise(absent)))

	if before.Band != after.Band {
		t.Errorf("band moved from %q to %q merely by adding a reading the payload has no data for", before.Band, after.Band)
	}
	if before.HumanConfidence != after.HumanConfidence {
		t.Errorf("human confidence moved from %d to %d", before.HumanConfidence, after.HumanConfidence)
	}
	if before.BotLikeness != after.BotLikeness {
		t.Errorf("bot likeness moved from %d to %d", before.BotLikeness, after.BotLikeness)
	}
}

func TestWebRTCDoesNotDiluteAPayloadEvenWhenNoCandidatesWereProduced(t *testing.T) {
	sections := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Contradiction},
	}
	before := summarise(sections)

	zeroCandidates := sectionWebRTC(webrtcRequest(t, "complete", false, []string{"gathering", "complete"}, nil), Inputs{}, claim{})
	zeroCandidates.ID = "webrtc"
	after := summarise(append(sections, normalise(zeroCandidates)))

	if before.Band != after.Band {
		t.Errorf("band moved from %q to %q by adding the zero-candidate reading, which must stay out of the count", before.Band, after.Band)
	}
	if before.HumanConfidence != after.HumanConfidence {
		t.Errorf("human confidence moved from %d to %d", before.HumanConfidence, after.HumanConfidence)
	}
	if before.BotLikeness != after.BotLikeness {
		t.Errorf("bot likeness moved from %d to %d", before.BotLikeness, after.BotLikeness)
	}
}
