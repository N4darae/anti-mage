package scan

import "strconv"

func sectionMath(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Inconclusive}

	exactSeen, exactWrong := mathRunExact(r, &s)
	repeatSeen, repeatWrong := mathRunRepeat(r, &s)

	if exactSeen == 0 && repeatSeen == 0 {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "nothing to examine",
			Note:  "neither the exact-result probe nor the repeated-call probe reported a readable case",
		})
		return s
	}

	if exactWrong > 0 || repeatWrong > 0 {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "a numeric built-in disagreed with what the specification requires, or with its own second call",
			Note:  strconv.Itoa(exactWrong) + " exact mismatch(es), " + strconv.Itoa(repeatWrong) + " repeat disagreement(s)",
		})
		return s
	}

	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "every exactly-specified result matched the specification, and every repeated call agreed with its own earlier reading",
		Note:  strconv.Itoa(exactSeen) + " exact case(s) and " + strconv.Itoa(repeatSeen) + " repeated call(s) examined",
	})
	return s
}

type mathExactCase struct {
	key      string
	expr     string
	expected string
}

type mathExactGroup struct {
	clause string
	cases  []mathExactCase
}

var mathExactGroups = []mathExactGroup{
	{

		clause: "tc39.es/ecma262/#sec-math.abs",
		cases: []mathExactCase{
			{"abs.nan", "Math.abs(NaN)", "NaN"},
			{"abs.negZero", "Math.abs(-0)", "0"},
			{"abs.negInf", "Math.abs(-Infinity)", "Infinity"},
			{"abs.neg", "Math.abs(-7)", "7"},
			{"abs.pos", "Math.abs(7)", "7"},
		},
	},
	{

		clause: "tc39.es/ecma262/#sec-math.sign",
		cases: []mathExactCase{
			{"sign.nan", "Math.sign(NaN)", "NaN"},
			{"sign.posZero", "Math.sign(0)", "0"},
			{"sign.negZero", "Math.sign(-0)", "-0"},
			{"sign.pos", "Math.sign(5)", "1"},
			{"sign.neg", "Math.sign(-5)", "-1"},
		},
	},
	{

		clause: "tc39.es/ecma262/#sec-math.floor",
		cases: []mathExactCase{
			{"floor.nan", "Math.floor(NaN)", "NaN"},
			{"floor.posInf", "Math.floor(Infinity)", "Infinity"},
			{"floor.negInf", "Math.floor(-Infinity)", "-Infinity"},
			{"floor.negZero", "Math.floor(-0)", "-0"},
			{"floor.fracNeg", "Math.floor(-0.5)", "-1"},
			{"floor.fracPos", "Math.floor(2.7)", "2"},
		},
	},
	{

		clause: "tc39.es/ecma262/#sec-math.ceil",
		cases: []mathExactCase{
			{"ceil.nan", "Math.ceil(NaN)", "NaN"},
			{"ceil.posInf", "Math.ceil(Infinity)", "Infinity"},
			{"ceil.negInf", "Math.ceil(-Infinity)", "-Infinity"},
			{"ceil.posZero", "Math.ceil(0)", "0"},
			{"ceil.fracNeg", "Math.ceil(-0.5)", "-0"},
			{"ceil.fracPos", "Math.ceil(2.3)", "3"},
		},
	},
	{

		clause: "tc39.es/ecma262/#sec-math.trunc",
		cases: []mathExactCase{
			{"trunc.nan", "Math.trunc(NaN)", "NaN"},
			{"trunc.fracNeg", "Math.trunc(-0.9)", "-0"},
			{"trunc.fracPos", "Math.trunc(0.9)", "0"},
			{"trunc.negInt", "Math.trunc(-3.9)", "-3"},
		},
	},
	{

		clause: "tc39.es/ecma262/#sec-math.round",
		cases: []mathExactCase{
			{"round.nan", "Math.round(NaN)", "NaN"},
			{"round.negZero", "Math.round(-0)", "-0"},
			{"round.halfNeg", "Math.round(-0.5)", "-0"},
			{"round.halfPos", "Math.round(0.5)", "1"},
			{"round.negHalfInt", "Math.round(-2.5)", "-2"},
			{"round.posInf", "Math.round(Infinity)", "Infinity"},
			{"round.negInf", "Math.round(-Infinity)", "-Infinity"},
		},
	},
	{

		clause: "tc39.es/ecma262/#sec-math.min",
		cases: []mathExactCase{
			{"min.nan", "Math.min(1, NaN)", "NaN"},
			{"min.zero", "Math.min(-0, 0)", "-0"},
			{"min.basic", "Math.min(3, 1, 2)", "1"},
			{"min.empty", "Math.min()", "Infinity"},
		},
	},
	{

		clause: "tc39.es/ecma262/#sec-math.max",
		cases: []mathExactCase{
			{"max.nan", "Math.max(1, NaN)", "NaN"},
			{"max.zero", "Math.max(-0, 0)", "0"},
			{"max.basic", "Math.max(3, 1, 2)", "3"},
			{"max.empty", "Math.max()", "-Infinity"},
		},
	},
	{

		clause: "tc39.es/ecma262/#sec-math.fround",
		cases: []mathExactCase{
			{"fround.nan", "Math.fround(NaN)", "NaN"},
			{"fround.negZero", "Math.fround(-0)", "-0"},
			{"fround.overflow", "Math.fround(2**150)", "Infinity"},
			{"fround.tieToEven", "Math.fround(16777217)", "16777216"},
			{"fround.exact", "Math.fround(0.5)", "0.5"},
		},
	},
	{

		clause: "tc39.es/ecma262/#sec-math.clz32",
		cases: []mathExactCase{
			{"clz32.zero", "Math.clz32(0)", "32"},
			{"clz32.one", "Math.clz32(1)", "31"},
			{"clz32.negOne", "Math.clz32(-1)", "0"},
			{"clz32.nan", "Math.clz32(NaN)", "32"},
		},
	},
	{

		clause: "tc39.es/ecma262/#sec-math.imul",
		cases: []mathExactCase{
			{"imul.basic", "Math.imul(3, 4)", "12"},
			{"imul.overflow", "Math.imul(0xffffffff, 5)", "-5"},
			{"imul.bigxbig", "Math.imul(0x7fffffff, 0x7fffffff)", "1"},
			{"imul.nan", "Math.imul(NaN, 5)", "0"},
		},
	},
	{

		clause: "tc39.es/ecma262/#sec-math.sqrt",
		cases: []mathExactCase{
			{"sqrt.nan", "Math.sqrt(NaN)", "NaN"},
			{"sqrt.negative", "Math.sqrt(-1)", "NaN"},
			{"sqrt.negZero", "Math.sqrt(-0)", "-0"},
			{"sqrt.posInf", "Math.sqrt(Infinity)", "Infinity"},
			{"sqrt.perfect", "Math.sqrt(4)", "2"},
			{"sqrt.exactFraction", "Math.sqrt(6.25)", "2.5"},
		},
	},
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
						Note:  g.clause + " requires " + c.expected,
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
					Note:  "a pure function must return the same result for the same argument in the same call",
				})
			}
		}
	}
	return seen, wrong
}
