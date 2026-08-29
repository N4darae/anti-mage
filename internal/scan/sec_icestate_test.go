package scan

import "testing"

func iceStateRequest(body string) Request {
	return Request{Probes: map[string]Probe{
		"webrtc.ice": {Status: StatusOK, Value: []byte(body)},
	}}
}

const gatheringNeverBegan = `{"finalState":"new","timedOut":true,"sawStates":["new"],"candidateTypes":[],` +
	`"localDescriptionSet":true,"signalingState":"have-local-offer","gatheringEvents":0,"candidateEvents":0,` +
	`"sdpCandidateLines":1,"statsLocalCandidates":4,"statsRead":true}`

func TestICEStateContradictsWhenCandidatesExistAndGatheringNeverBegan(t *testing.T) {
	got := sectionICEState(iceStateRequest(gatheringNeverBegan), Inputs{}, claim{})
	if got.Determination != Contradiction {
		t.Fatalf("determination = %q, want %q: a connection carrying its own gathered candidates cannot also report that gathering never began", got.Determination, Contradiction)
	}
}

func TestICEStateIsUnverifiedWhenGatheringRanNormally(t *testing.T) {
	body := `{"finalState":"complete","timedOut":false,"sawStates":["new","gathering","complete"],"candidateTypes":["host"],` +
		`"localDescriptionSet":true,"signalingState":"have-local-offer","gatheringEvents":2,"candidateEvents":2,` +
		`"sdpCandidateLines":1,"statsLocalCandidates":4,"statsRead":true}`
	got := sectionICEState(iceStateRequest(body), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q", got.Determination, Unverified)
	}
}

func TestICEStateIsUnverifiedWhenGatheringBeganButDidNotFinish(t *testing.T) {
	body := `{"finalState":"gathering","timedOut":true,"sawStates":["new","gathering"],"candidateTypes":[],` +
		`"localDescriptionSet":true,"signalingState":"have-local-offer","gatheringEvents":1,"candidateEvents":0,` +
		`"sdpCandidateLines":2,"statsLocalCandidates":2,"statsRead":true}`
	got := sectionICEState(iceStateRequest(body), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a gathering still in progress is an ordinary state", got.Determination, Unverified)
	}
}

func TestICEStateIsUnverifiedWhenNoLocalDescriptionWasSet(t *testing.T) {
	body := `{"finalState":"new","timedOut":true,"sawStates":["new"],"candidateTypes":[],` +
		`"localDescriptionSet":false,"describeError":"blocked","signalingState":"stable","gatheringEvents":0,"candidateEvents":0,` +
		`"sdpCandidateLines":0,"statsLocalCandidates":0,"statsRead":true}`
	got := sectionICEState(iceStateRequest(body), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a connection never given a local description was never asked to gather", got.Determination, Unverified)
	}
}

func TestICEStateIsUnverifiedWhenGatheringNeverBeganAndNoCandidatesExist(t *testing.T) {
	body := `{"finalState":"new","timedOut":true,"sawStates":["new"],"candidateTypes":[],` +
		`"localDescriptionSet":true,"signalingState":"have-local-offer","gatheringEvents":0,"candidateEvents":0,` +
		`"sdpCandidateLines":0,"statsLocalCandidates":0,"statsRead":true}`
	got := sectionICEState(iceStateRequest(body), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a connection that gathered nothing contradicts nothing", got.Determination, Unverified)
	}
}

func TestICEStateIsUnverifiedOnAPayloadThatPredatesTheReading(t *testing.T) {
	body := `{"finalState":"new","timedOut":true,"sawStates":["new"],"candidateTypes":["host"]}`
	got := sectionICEState(iceStateRequest(body), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a payload without the fields this reading needs settles nothing", got.Determination, Unverified)
	}
}

func TestICEStateIsUnverifiedWhenNothingWasCollected(t *testing.T) {
	for _, r := range []Request{
		{Probes: map[string]Probe{}},
		{Probes: map[string]Probe{"webrtc.ice": {Status: StatusUnsupported, Value: []byte(`{"reason":"none"}`)}}},
	} {
		got := sectionICEState(r, Inputs{}, claim{})
		if got.Determination != Unverified {
			t.Fatalf("determination = %q, want %q", got.Determination, Unverified)
		}
	}
}

func TestICEStateReadsStatsAloneAsCandidatesGathered(t *testing.T) {
	body := `{"finalState":"new","timedOut":true,"sawStates":["new"],"candidateTypes":[],` +
		`"localDescriptionSet":true,"signalingState":"have-local-offer","gatheringEvents":0,"candidateEvents":0,` +
		`"sdpCandidateLines":0,"statsLocalCandidates":4,"statsRead":true}`
	got := sectionICEState(iceStateRequest(body), Inputs{}, claim{})
	if got.Determination != Contradiction {
		t.Fatalf("determination = %q, want %q", got.Determination, Contradiction)
	}
}

func TestICEStateDoesNotDiluteAPayloadThatPredatesIt(t *testing.T) {
	sections := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Inconclusive},
	}
	before := summarise(sections)
	absent := sectionICEState(Request{Probes: map[string]Probe{}}, Inputs{}, claim{})
	absent.ID = "icestate"
	after := summarise(append(sections, normalise(absent)))
	if before.Band != after.Band || before.HumanConfidence != after.HumanConfidence || before.BotLikeness != after.BotLikeness {
		t.Fatalf("summary moved from %+v to %+v merely by adding a reading the payload has no data for", before, after)
	}
}

func TestICEStateContradictsWhenGatheringFinishedWithCandidatesHeldAndNoneDescribed(t *testing.T) {
	body := `{"localDescriptionSet":true,"finalState":"complete","gatheringEvents":2,
		"sdpCandidateLines":0,"statsLocalCandidates":2,"candidateEvents":1,"candidateTypes":[]}`
	got := sectionICEState(iceStateRequest(body), Inputs{}, claim{})
	if got.Determination != Contradiction {
		t.Fatalf("determination = %q, want %q; rows: %+v", got.Determination, Contradiction, got.Rows)
	}
}

func TestICEStateReadsNothingIntoFinishedGatheringThatHoldsNothing(t *testing.T) {
	body := `{"localDescriptionSet":true,"finalState":"complete","gatheringEvents":2,
		"sdpCandidateLines":0,"statsLocalCandidates":0,"candidateEvents":1,"candidateTypes":[]}`
	got := sectionICEState(iceStateRequest(body), Inputs{}, claim{})
	if got.Determination == Contradiction {
		t.Fatalf("determination = %q; a connection that gathered nothing and describes nothing agrees with itself, and an environment that reaches no network is never scored for it", got.Determination)
	}
}

func TestICEStateReadsNothingWhenTheDescriptionCouldNotBeCounted(t *testing.T) {
	body := `{"localDescriptionSet":true,"finalState":"complete","gatheringEvents":2,
		"statsLocalCandidates":2,"candidateEvents":1,"candidateTypes":[]}`
	got := sectionICEState(iceStateRequest(body), Inputs{}, claim{})
	if got.Determination == Contradiction {
		t.Fatalf("determination = %q; with no count of the description there is nothing for the statistics to disagree with", got.Determination)
	}
}

func TestICEStateWeighsSilencedAnnouncementAboveAPlainDisagreement(t *testing.T) {
	silenced := sectionICEState(iceStateRequest(gatheringNeverBegan), Inputs{}, claim{})
	if silenced.weighs() != weightOnlyDeliberate {
		t.Fatalf("weight = %d, want %d: turning this facility off leaves nothing to hold, so results held behind a state machine that never announced them is not a state anything but a deliberate change reaches",
			silenced.weighs(), weightOnlyDeliberate)
	}
	body := `{"localDescriptionSet":true,"finalState":"complete","gatheringEvents":2,
		"sdpCandidateLines":0,"statsLocalCandidates":2,"candidateEvents":1,"candidateTypes":[]}`
	described := sectionICEState(iceStateRequest(body), Inputs{}, claim{})
	if described.weighs() != weightDisagreement {
		t.Fatalf("weight = %d, want %d: an environment that rewrites its own description is what the tools that shield this facility do, so this reading is not weighed as one nothing else produces",
			described.weighs(), weightDisagreement)
	}
}
