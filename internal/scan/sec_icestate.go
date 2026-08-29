package scan

func sectionICEState(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	raw, ok := r.value("webrtc.ice")
	if !ok {
		s.Rows = append(s.Rows, Row{
			Label: "gathering against its own state machine",
			Value: "not collected",
			Note:  anomalyNote,
		})
		return s
	}

	described, haveDescribed := boolean(raw, "localDescriptionSet")
	finalState, haveState := str(raw, "finalState")
	sdpLines, haveSDP := num(raw, "sdpCandidateLines")
	statsCount, haveStats := num(raw, "statsLocalCandidates")
	gatheringEvents, haveEvents := num(raw, "gatheringEvents")

	if !haveDescribed || !haveState || !haveEvents || (!haveSDP && !haveStats) {
		s.Rows = append(s.Rows, Row{
			Label: "gathering against its own state machine",
			Value: "not reported",
			Note:  anomalyNote,
		})
		return s
	}

	s.Rows = append(s.Rows, Row{
		Label: "a local description was accepted",
		Value: answerOrAbsent(described, true),
		Note:  anomalyNote,
	})
	s.Rows = append(s.Rows, Row{
		Label: "gathering state, at the end of this reading's watch",
		Value: valueOrAbsent(finalState),
		Note:  anomalyNote,
	})
	s.Rows = append(s.Rows, Row{
		Label: "state changes the connection announced",
		Value: itoa(gatheringEvents),
		Note:  anomalyNote,
	})
	if haveSDP {
		s.Rows = append(s.Rows, Row{
			Label: "candidate lines in the connection's own local description",
			Value: itoa(sdpLines),
			Note:  anomalyNote,
		})
	}
	if haveStats {
		s.Rows = append(s.Rows, Row{
			Label: "local candidates in the connection's own statistics",
			Value: itoa(statsCount),
			Note:  anomalyNote,
		})
	}

	gathered := (haveSDP && sdpLines > 0) || (haveStats && statsCount > 0)
	neverBegan := described && finalState == "new" && gatheringEvents == 0
	heldButNeverDescribed := described && finalState == "complete" &&
		haveStats && statsCount > 0 && haveSDP && sdpLines == 0

	if heldButNeverDescribed {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "this connection holds candidates it reports it has finished gathering, and describes itself as carrying none",
			Note:  anomalyNote,
		})
		return s
	}

	if !neverBegan || !gathered {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the state this connection reports and the candidates it holds do not disagree",
			Note:  anomalyNote,
		})
		return s
	}

	s.Determination, s.weight = Contradiction, weightOnlyDeliberate
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "this connection holds candidates it can only have gathered, and reports that gathering never began",
		Note:  anomalyNote,
	})
	return s
}
