package scan

import (
	_ "time/tzdata"
)

var knownIceCandidateTypes = []string{"host", "srflx", "prflx", "relay"}

type iceReading struct {
	finalState string
	haveState  bool

	timedOut     bool
	haveTimedOut bool

	sawStates      []string
	candidateTypes []string

	described     bool
	haveDescribed bool

	sdpLines float64
	haveSDP  bool

	statsCandidates float64
	haveStats       bool

	gatheringEvents float64
	haveEvents      bool
}

func readICE(r Request) (iceReading, bool) {
	raw, ok := r.value("webrtc.ice")
	if !ok {
		return iceReading{}, false
	}
	var ice iceReading
	ice.finalState, ice.haveState = str(raw, "finalState")
	ice.timedOut, ice.haveTimedOut = boolean(raw, "timedOut")
	ice.sawStates, _ = strList(raw, "sawStates")
	ice.candidateTypes, _ = strList(raw, "candidateTypes")
	ice.described, ice.haveDescribed = boolean(raw, "localDescriptionSet")
	ice.sdpLines, ice.haveSDP = num(raw, "sdpCandidateLines")
	ice.statsCandidates, ice.haveStats = num(raw, "statsLocalCandidates")
	ice.gatheringEvents, ice.haveEvents = num(raw, "gatheringEvents")
	return ice, true
}

func (ice iceReading) candidateCounts() map[string]int {
	out := map[string]int{}
	for _, t := range knownIceCandidateTypes {
		out[t] = 0
	}
	for _, t := range ice.candidateTypes {
		if _, known := out[t]; known {
			out[t]++
		}
	}
	return out
}

func (ice iceReading) unrecognisedCandidates() int {
	counts := ice.candidateCounts()
	unknown := len(ice.candidateTypes)
	for _, t := range knownIceCandidateTypes {
		unknown -= counts[t]
	}
	return unknown
}

func (ice iceReading) coherent() bool {
	if ice.timedOut {
		return ice.finalState != "complete"
	}
	if ice.finalState == "complete" {
		return true
	}
	return len(ice.candidateTypes) == 0
}

func (ice iceReading) reportsItsStateMachine() bool {
	return ice.haveDescribed && ice.haveState && ice.haveEvents && (ice.haveSDP || ice.haveStats)
}

func (ice iceReading) heldWhatItDescribesAsNone() bool {
	return ice.described && ice.finalState == "complete" &&
		ice.haveStats && ice.statsCandidates > 0 && ice.haveSDP && ice.sdpLines == 0
}

func (ice iceReading) gatheredBeforeItBegan() bool {
	gathered := (ice.haveSDP && ice.sdpLines > 0) || (ice.haveStats && ice.statsCandidates > 0)
	neverBegan := ice.described && ice.finalState == "new" && ice.gatheringEvents == 0
	return gathered && neverBegan
}
