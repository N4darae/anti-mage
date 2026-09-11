package scan

import (
	"strconv"
	_ "time/tzdata"
)

func noted(label, value string) Row {
	return Row{Label: label, Value: value, Note: anomalyNote}
}

func (s *Section) add(label, value string) {
	s.Rows = append(s.Rows, noted(label, value))
}

func (s *Section) conclude(value string) {
	s.add("conclusion", value)
}

type conclusions struct {
	inconclusive  string
	contradiction string
	instrumented  string
	consistent    string
}

func (s *Section) settle(t tally, c conclusions) {
	s.Determination = t.determination()
	switch s.Determination {
	case Inconclusive:
		s.conclude(c.inconclusive)
	case Contradiction:
		s.conclude(c.contradiction)
	case Instrumented:
		s.conclude(c.instrumented)
	default:
		s.conclude(c.consistent)
	}
}

func countOf(n int, unit string) string {
	return strconv.Itoa(n) + " " + unit
}

func outOf(part, whole int) string {
	return strconv.Itoa(part) + " of " + strconv.Itoa(whole)
}
