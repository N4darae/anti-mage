// Package assess reads whatever a caller observed about a browser environment

package assess

import (
	"github.com/N4darae/anti-mage/internal/scan"
)

func Evaluate(env Environment) Assessment {
	req := scan.Request{
		V:      1,
		Mode:   "public",
		Nonce:  env.Nonce,
		Probes: probesOf(env),
	}
	in := scan.Inputs{
		Nonce:        env.Nonce,
		OffsetDates:  env.OffsetDates,
		FontControls: env.FontControls,
		ElapsedMS:    elapsedOf(env),
	}
	rep := scan.AnalyzeWith(req, in, suppliedSections(env.Findings))

	a := Assessment{
		V:        1,
		Score:    clampScore(rep.Summary.BotLikeness),
		Supplied: suppliedIDs(env.Observations),
	}
	a.Determination = determinationOf(rep.Summary.Band)
	if a.Determination == NotEvaluated {

		a.Score = 0
	}
	a.Statement = statements[a.Determination]
	return a
}

func (e Environment) NamesReported(id string) []string {
	return scan.Request{Probes: probesOf(e)}.ReportedNames(id)
}

var statements = map[Determination]string{
	NotEvaluated: "Nothing that was supplied reached a reading, so this says nothing about this environment.",
	Insufficient: "Nothing disagreed, and too little was established to characterise this environment either way.",
	Coherent:     "Everything that could be read agrees with the platform this environment claims.",
	Discrepant:   "One body of evidence reports two facts that cannot both be true. Everything else that could be read agrees.",
	Instrumented: "This environment appears modified. Privacy, accessibility and content-blocking tools modify the same surfaces, in large numbers, so this describes the environment and not the person using it.",
}
