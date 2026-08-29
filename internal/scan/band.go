package scan

import "math"

const (
	BandNotEvaluated = "not-evaluated"

	BandInsufficient = "insufficient"

	BandCoherent = "coherent"

	BandDiscrepant = "discrepant"

	BandInstrumented = "instrumented"
)

const humanConfidenceCap = 90

const botLikenessCap = 90

type findingWeight int

const (
	weightNone findingWeight = 0

	weightDeclaredModification findingWeight = 10

	weightDisagreement findingWeight = 40

	weightOnlyDeliberate findingWeight = 70
)

func (s Section) weighs() findingWeight {
	switch s.Determination {
	case Contradiction, Instrumented:
	default:
		return weightNone
	}
	if s.weight != weightNone {
		return s.weight
	}
	if s.Determination == Contradiction {
		return weightDisagreement
	}
	return weightDeclaredModification
}

func combineWeights(ws []findingWeight) int {
	if len(ws) == 0 {
		return 0
	}
	unaccounted := 1.0
	for _, w := range ws {
		unaccounted *= 1 - float64(w)/100
	}
	n := round10(int(math.Round(100 * (1 - unaccounted))))
	if n > botLikenessCap {
		return botLikenessCap
	}
	return n
}

func summarise(sections []Section) Summary {
	candidates, consistent, contradictions, instrumented, deliberate := 0, 0, 0, 0, 0
	var weights []findingWeight
	for _, s := range sections {
		switch s.Determination {
		case Unverified:
			continue
		case Consistent:
			consistent++
		case Contradiction:
			contradictions++
		case Instrumented:
			instrumented++
		}
		candidates++
		if w := s.weighs(); w != weightNone {
			weights = append(weights, w)
			if w == weightOnlyDeliberate {
				deliberate++
			}
		}
	}
	flagged := contradictions + instrumented
	determined := consistent + flagged

	sum := Summary{}
	switch {
	case determined == 0:
		sum.Band = BandNotEvaluated
		sum.Headline = "No section reached a determination, so this scan says nothing about this environment."
	case deliberate > 0 && contradictions > 0:
		sum.Band = BandInstrumented
		sum.Headline = "This environment appears instrumented: parts of it report facts that cannot all be true at once, and at least one of those facts is one nothing but a deliberate change produces."
	case deliberate > 0:
		sum.Band = BandInstrumented
		sum.Headline = "This environment appears instrumented: it reports something nothing but a deliberate change to it produces."
	case contradictions == 0 && instrumented > 0:
		sum.Band = BandInstrumented
		sum.Headline = "Parts of this environment report that they have been modified. Privacy, accessibility and content-blocking tools modify the same parts, so this describes the environment and not the person using it."
	case instrumented > 0 || flagged >= 2:
		sum.Band = BandInstrumented
		sum.Headline = "This environment appears instrumented: parts of it report facts that cannot all be true at once."
	case flagged == 1:
		sum.Band = BandDiscrepant
		sum.Headline = "One section reports two facts that cannot both be true. The rest agree."
	case determined*2 >= candidates:
		sum.Band = BandCoherent
		sum.Headline = "Every measurement that completed agrees with the platform this browser claims."
	default:
		sum.Band = BandInsufficient
		sum.Headline = "Too few sections reached a determination to characterise this environment either way."
	}

	if candidates > 0 {
		if contradictions == 0 && deliberate == 0 {
			sum.HumanConfidence = round10(100 * consistent / candidates)
			if sum.HumanConfidence > humanConfidenceCap {
				sum.HumanConfidence = humanConfidenceCap
			}
		}
		sum.BotLikeness = combineWeights(weights)
	}
	return sum
}

func round10(n int) int {
	if n < 0 {
		return 0
	}
	if n > 100 {
		n = 100
	}
	return ((n + 5) / 10) * 10
}
