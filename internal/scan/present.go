package scan

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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

func itoa(f float64) string {
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func clip(s string, n int) string {
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return ' '
		}
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func joinLimit(xs []string, n int) string {
	if len(xs) == 0 {
		return "none"
	}
	if len(xs) <= n {
		return strings.Join(xs, ", ")
	}
	return strings.Join(xs[:n], ", ") + fmt.Sprintf(", and %d more", len(xs)-n)
}

func boolWord(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func answerOrAbsent(value, present bool) string {
	if !present {
		return "not reported"
	}
	if value {
		return "yes"
	}
	return "no"
}

func valueOrAbsent(s string) string {
	if s == "" {
		return "not reported"
	}
	return clip(s, 40)
}
