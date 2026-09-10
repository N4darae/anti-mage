package scan

import (
	_ "time/tzdata"
)

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
