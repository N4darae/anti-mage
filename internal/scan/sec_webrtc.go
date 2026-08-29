package scan

var knownIceCandidateTypes = map[string]bool{
	"host":  true,
	"srflx": true,
	"prflx": true,
	"relay": true,
}

func sectionWebRTC(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	if r.unsupported("webrtc.ice") {
		s.Rows = append(s.Rows, Row{
			Label: "ICE gathering",
			Value: "reported unsupported",
			Note:  anomalyNote,
		})
		return s
	}

	raw, ok := r.value("webrtc.ice")
	if !ok {
		s.Rows = append(s.Rows, Row{
			Label: "ICE gathering",
			Value: "not collected",
			Note:  anomalyNote,
		})
		return s
	}

	finalState, haveState := str(raw, "finalState")
	timedOut, haveTimedOut := boolean(raw, "timedOut")
	sawStates, _ := strList(raw, "sawStates")
	candidateTypes, _ := strList(raw, "candidateTypes")

	if !haveState || !haveTimedOut {
		s.Rows = append(s.Rows, Row{
			Label: "ICE gathering",
			Value: "reported incompletely",
			Note:  anomalyNote,
		})
		return s
	}

	s.Rows = append(s.Rows, Row{
		Label: "gathering state, at the end of this reading's watch",
		Value: valueOrAbsent(finalState),
		Note:  anomalyNote,
	})
	s.Rows = append(s.Rows, Row{
		Label: "states observed in order",
		Value: joinLimit(sawStates, 10),
		Note:  anomalyNote,
	})
	s.Rows = append(s.Rows, Row{
		Label: "gathering reached its own deadline before finishing",
		Value: answerOrAbsent(timedOut, true),
		Note:  anomalyNote,
	})

	counts := countIceCandidateTypes(candidateTypes)
	for _, t := range []string{"host", "srflx", "prflx", "relay"} {
		s.Rows = append(s.Rows, Row{
			Label: "candidates of type " + t,
			Value: itoa(float64(counts[t])),
			Note:  anomalyNote,
		})
	}
	unknown := len(candidateTypes) - counts["host"] - counts["srflx"] - counts["prflx"] - counts["relay"]
	if unknown > 0 {
		s.Rows = append(s.Rows, Row{
			Label: "candidates of an unrecognised type",
			Value: itoa(float64(unknown)),
			Note:  anomalyNote,
		})
	}

	coherent := iceCoherent(finalState, timedOut, len(candidateTypes) > 0)
	conclusion := "the candidate count and the gathering state this session reported are consistent with each other"
	if !coherent {
		conclusion = "the candidate count and the gathering state this session reported do not fit the shape this reading expects from one gathering process"
	}
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: conclusion,
		Note:  anomalyNote,
	})
	return s
}

func countIceCandidateTypes(types []string) map[string]int {
	out := map[string]int{"host": 0, "srflx": 0, "prflx": 0, "relay": 0}
	for _, t := range types {
		if knownIceCandidateTypes[t] {
			out[t]++
		}
	}
	return out
}

func iceCoherent(finalState string, timedOut bool, haveCandidates bool) bool {
	if timedOut {
		return finalState != "complete"
	}
	if finalState == "complete" {
		return true
	}
	return !haveCandidates
}
