package scan

func sectionICEState(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	ice, ok := readICE(r)
	if !ok {
		s.add("gathering against its own state machine", "not collected")
		return s
	}
	if !ice.reportsItsStateMachine() {
		s.add("gathering against its own state machine", "not reported")
		return s
	}

	s.add("a local description was accepted", answerOrAbsent(ice.described, true))
	s.add("gathering state, at the end of this reading's watch", valueOrAbsent(ice.finalState))
	s.add("state changes the connection announced", itoa(ice.gatheringEvents))
	if ice.haveSDP {
		s.add("candidate lines in the connection's own local description", itoa(ice.sdpLines))
	}
	if ice.haveStats {
		s.add("local candidates in the connection's own statistics", itoa(ice.statsCandidates))
	}

	if ice.heldWhatItDescribesAsNone() {
		s.Determination = Contradiction
		s.conclude("this connection holds candidates it reports it has finished gathering, and describes itself as carrying none")
		return s
	}
	if !ice.gatheredBeforeItBegan() {
		s.conclude("the state this connection reports and the candidates it holds do not disagree")
		return s
	}

	s.Determination, s.weight = Contradiction, weightOnlyDeliberate
	s.conclude("this connection holds candidates it can only have gathered, and reports that gathering never began")
	return s
}
