package scan

import (
	"strconv"
	_ "time/tzdata"
)

const nativesMaxRows = 8

func sectionNatives(_ Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}
	n := c.natives

	s.Rows = append(s.Rows, Row{
		Label: "accessors probed",
		Value: strconv.Itoa(len(n.targets)),
		Note:  anomalyNote,
	})
	if n.applied == 0 {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "no requirement could be applied", Note: anomalyNote})
		return s
	}
	if len(n.violations) == 0 {
		s.Determination = Consistent
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "every requirement held on every accessor tested",
			Note:  anomalyNote,
		})
		return s
	}
	shown := n.violations
	if len(shown) > 8 {
		shown = shown[:8]
	}
	for _, v := range shown {
		s.Rows = append(s.Rows, Row{Label: clip(v.target, 120), Value: "does not meet a requirement on a built-in", Note: anomalyNote})
	}
	s.Determination = Instrumented
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: strconv.Itoa(len(n.violations)) + " of " + strconv.Itoa(n.applied) + " requirements did not hold",
		Note:  anomalyNote,
	})
	return s
}
