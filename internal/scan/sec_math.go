package scan

const mathMaxRows = 12

type mathExactResult struct {
	c   mathExactCase
	got string
}

type mathExactReading struct {
	seen  int
	wrong []mathExactResult
}

type mathRepeatResult struct {
	name, first, second string
}

type mathRepeatReading struct {
	seen  int
	wrong []mathRepeatResult
}

func readMathExact(r Request) mathExactReading {
	var reading mathExactReading
	raw, ok := r.value("math.exact")
	if !ok {
		return reading
	}
	for _, g := range mathExactGroups {
		for _, c := range g.cases {
			got, present := str(raw, c.key)
			if !present {
				continue
			}
			reading.seen++
			if got != c.expected {
				reading.wrong = append(reading.wrong, mathExactResult{c, got})
			}
		}
	}
	return reading
}

func readMathRepeat(r Request) mathRepeatReading {
	var reading mathRepeatReading
	raw, ok := r.value("math.repeat")
	if !ok {
		return reading
	}
	m, isMap := object(raw)
	if !isMap {
		return reading
	}
	for _, k := range keys(m) {
		first, haveFirst := str(raw, k, "a")
		second, haveSecond := str(raw, k, "b")
		if !haveFirst || !haveSecond {
			continue
		}
		reading.seen++
		if first != second {
			reading.wrong = append(reading.wrong, mathRepeatResult{k, first, second})
		}
	}
	return reading
}

func sectionMath(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Inconclusive}

	exact := readMathExact(r)
	repeat := readMathRepeat(r)

	for i, w := range exact.wrong {
		if i >= mathMaxRows {
			break
		}
		s.add(w.c.expr, "reported "+clip(w.got, 40))
	}
	for i, w := range repeat.wrong {
		if i >= mathMaxRows {
			break
		}
		s.add("Math."+clip(w.name, 40)+", called twice with the same argument",
			"reported "+clip(w.first, 40)+" then "+clip(w.second, 40))
	}

	if exact.seen == 0 && repeat.seen == 0 {
		s.conclude("nothing to examine")
		return s
	}
	if len(exact.wrong) > 0 || len(repeat.wrong) > 0 {
		s.Determination = Contradiction
		s.conclude("a numeric built-in disagreed with what the specification requires, or with its own second call")
		return s
	}
	s.Determination = Consistent
	s.conclude("every exactly-specified result matched the specification, and every repeated call agreed with its own earlier reading")
	return s
}
