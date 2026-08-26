package scan

const (
	BandNotEvaluated = "not-evaluated"

	BandInsufficient = "insufficient"

	BandCoherent = "coherent"

	BandDiscrepant = "discrepant"

	BandInstrumented = "instrumented"
)

const humanConfidenceCap = 90

const (
	contradictionFloor = 30
	instrumentedFloor  = 10
)

func summarise(sections []Section) Summary {
	candidates, consistent, contradictions, instrumented := 0, 0, 0, 0
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
	}
	flagged := contradictions + instrumented
	determined := consistent + flagged

	sum := Summary{}
	switch {
	case determined == 0:
		sum.Band = BandNotEvaluated
		sum.Headline = "No section reached a determination, so this scan says nothing about this environment."
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

		if contradictions == 0 {
			sum.HumanConfidence = round10(100 * consistent / candidates)
			if sum.HumanConfidence > humanConfidenceCap {
				sum.HumanConfidence = humanConfidenceCap
			}
		}
		sum.BotLikeness = round10(100 * (2*contradictions + instrumented) / (2 * candidates))
		switch {
		case contradictions > 0 && sum.BotLikeness < contradictionFloor:
			sum.BotLikeness = contradictionFloor
		case contradictions == 0 && instrumented > 0 && sum.BotLikeness < instrumentedFloor:
			sum.BotLikeness = instrumentedFloor
		}
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
