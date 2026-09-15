package scan

import (
	"strconv"
	_ "time/tzdata"
)

const nativesMaxRows = 8

func sectionNatives(_ Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}
	n := c.natives

	s.add("accessors probed", strconv.Itoa(len(n.targets)))
	if n.applied == 0 {
		s.conclude("no requirement could be applied")
		return s
	}
	if len(n.violations) == 0 {
		s.Determination = Consistent
		s.conclude("every requirement held on every accessor tested")
		return s
	}

	shown := n.violations
	if len(shown) > nativesMaxRows {
		shown = shown[:nativesMaxRows]
	}
	for _, v := range shown {
		s.add(clip(v.target, 120), "does not meet a requirement on a built-in")
	}
	s.Determination = Instrumented
	s.conclude(outOf(len(n.violations), n.applied) + " requirements did not hold")
	return s
}
