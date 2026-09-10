package scan

import (
	_ "time/tzdata"
)

func sectionMath(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Inconclusive}

	exactSeen, exactWrong := mathRunExact(r, &s)
	repeatSeen, repeatWrong := mathRunRepeat(r, &s)

	if exactSeen == 0 && repeatSeen == 0 {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "nothing to examine",
			Note:  anomalyNote,
		})
		return s
	}

	if exactWrong > 0 || repeatWrong > 0 {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "a numeric built-in disagreed with what the specification requires, or with its own second call",
			Note:  anomalyNote,
		})
		return s
	}

	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "every exactly-specified result matched the specification, and every repeated call agreed with its own earlier reading",
		Note:  anomalyNote,
	})
	return s
}

func mathRunExact(r Request, s *Section) (seen, wrong int) {
	raw, ok := r.value("math.exact")
	if !ok {
		return 0, 0
	}
	for _, g := range mathExactGroups {
		for _, c := range g.cases {
			got, present := str(raw, c.key)
			if !present {

				continue
			}
			seen++
			if got != c.expected {
				wrong++
				if wrong <= mathMaxRows {
					s.Rows = append(s.Rows, Row{
						Label: c.expr,
						Value: "reported " + clip(got, 40),
						Note:  anomalyNote,
					})
				}
			}
		}
	}
	return seen, wrong
}

const mathMaxRows = 12

func mathRunRepeat(r Request, s *Section) (seen, wrong int) {
	raw, ok := r.value("math.repeat")
	if !ok {
		return 0, 0
	}
	m, isMap := object(raw)
	if !isMap {
		return 0, 0
	}
	for _, k := range keys(m) {
		a, aok := str(raw, k, "a")
		b, bok := str(raw, k, "b")
		if !aok || !bok {

			continue
		}
		seen++
		if a != b {
			wrong++
			if wrong <= mathMaxRows {
				s.Rows = append(s.Rows, Row{
					Label: "Math." + clip(k, 40) + ", called twice with the same argument",
					Value: "reported " + clip(a, 40) + " then " + clip(b, 40),
					Note:  anomalyNote,
				})
			}
		}
	}
	return seen, wrong
}
