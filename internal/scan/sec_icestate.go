package scan

func sectionICEState(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	raw, ok := r.value("webrtc.ice")
	if !ok {
		s.Rows = append(s.Rows, Row{
			Label: "gathering against its own state machine",
			Value: "not collected",
			Note:  "this reading has nothing to read a gathering process against, so it carries no weight either way",
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
			Note:  "this payload predates the fields this reading needs, so it carries no weight either way",
		})
		return s
	}

	s.Rows = append(s.Rows, Row{
		Label: "a local description was accepted",
		Value: answerOrAbsent(described, true),
		Note:  "setting one is what asks a connection to start gathering; a connection never asked was never expected to gather",
	})
	s.Rows = append(s.Rows, Row{
		Label: "gathering state, at the end of this reading's watch",
		Value: valueOrAbsent(finalState),
		Note:  "the state the connection itself reported",
	})
	s.Rows = append(s.Rows, Row{
		Label: "state changes the connection announced",
		Value: itoa(gatheringEvents),
		Note:  "",
	})
	if haveSDP {
		s.Rows = append(s.Rows, Row{
			Label: "candidate lines in the connection's own local description",
			Value: itoa(sdpLines),
			Note:  "a candidate reaches the local description only once it has been gathered",
		})
	}
	if haveStats {
		s.Rows = append(s.Rows, Row{
			Label: "local candidates in the connection's own statistics",
			Value: itoa(statsCount),
			Note:  "the connection's own accounting of what it holds, read through an interface separate from the event that announces one",
		})
	}

	gathered := (haveSDP && sdpLines > 0) || (haveStats && statsCount > 0)
	neverBegan := described && finalState == "new" && gatheringEvents == 0

	if !neverBegan || !gathered {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the state this connection reports and the candidates it holds do not disagree",
			Note:  "this reading settles nothing else about candidate gathering; how many candidates an honest environment produces is not something this project has established",
		})
		return s
	}

	s.Determination = Contradiction
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "this connection holds candidates it can only have gathered, and reports that gathering never began",
		Note:  "a connection that accepted a local description and now carries local candidates has gathered them; a gathering state of new, announced by no state change, says the process never started. Both are this connection's own statements about itself, and they cannot both be true. Nothing here reads the number of candidates or what they contain.",
	})
	return s
}
