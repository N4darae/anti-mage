package scan

import "strconv"

func sectionWebRTC(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	if r.unsupported("webrtc.ice") {
		s.add("ICE gathering", "reported unsupported")
		return s
	}

	ice, ok := readICE(r)
	if !ok {
		s.add("ICE gathering", "not collected")
		return s
	}
	if !ice.haveState || !ice.haveTimedOut {
		s.add("ICE gathering", "reported incompletely")
		return s
	}

	s.add("gathering state, at the end of this reading's watch", valueOrAbsent(ice.finalState))
	s.add("states observed in order", joinLimit(ice.sawStates, 10))
	s.add("gathering reached its own deadline before finishing", answerOrAbsent(ice.timedOut, true))

	counts := ice.candidateCounts()
	for _, t := range knownIceCandidateTypes {
		s.add("candidates of type "+t, strconv.Itoa(counts[t]))
	}
	if unknown := ice.unrecognisedCandidates(); unknown > 0 {
		s.add("candidates of an unrecognised type", strconv.Itoa(unknown))
	}

	if ice.coherent() {
		s.conclude("the candidate count and the gathering state this session reported are consistent with each other")
		return s
	}
	s.conclude("the candidate count and the gathering state this session reported do not fit the shape this reading expects from one gathering process")
	return s
}
