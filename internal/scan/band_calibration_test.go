package scan

import "testing"

func TestModifiedAccessorScoresBelowAContradiction(t *testing.T) {
	blocked := summarise([]Section{
		{ID: "natives", Determination: Instrumented},
		{ID: "fonts", Determination: Consistent},
		{ID: "scopes", Determination: Consistent},
		{ID: "geometry", Determination: Consistent},
		{ID: "time", Determination: Consistent},
		{ID: "automation", Determination: Consistent},
		{ID: "permissions", Determination: Consistent},
		{ID: "platform", Determination: Consistent},
	})
	caught := summarise([]Section{
		{ID: "fonts", Determination: Contradiction},
		{ID: "natives", Determination: Consistent},
		{ID: "scopes", Determination: Consistent},
		{ID: "geometry", Determination: Consistent},
		{ID: "time", Determination: Consistent},
		{ID: "automation", Determination: Consistent},
		{ID: "permissions", Determination: Consistent},
		{ID: "platform", Determination: Consistent},
	})

	if blocked.BotLikeness >= caught.BotLikeness {
		t.Errorf("botLikeness: modified accessor %d, contradiction %d; a modified accessor must score strictly lower",
			blocked.BotLikeness, caught.BotLikeness)
	}
	if blocked.HumanConfidence == 0 {
		t.Errorf("humanConfidence = 0 for a browser whose only finding is a modified accessor; modification does not argue against a person")
	}
	if caught.HumanConfidence != 0 {
		t.Errorf("humanConfidence = %d alongside a contradiction, want 0", caught.HumanConfidence)
	}
}

func TestSuppressionCannotBuyConfidence(t *testing.T) {
	full := summarise([]Section{
		{Determination: Consistent}, {Determination: Consistent}, {Determination: Consistent},
		{Determination: Consistent}, {Determination: Consistent}, {Determination: Consistent},
		{Determination: Consistent}, {Determination: Contradiction},
	})
	suppressed := summarise([]Section{
		{Determination: Consistent}, {Determination: Consistent}, {Determination: Inconclusive},
		{Determination: Inconclusive}, {Determination: Inconclusive}, {Determination: Inconclusive},
		{Determination: Inconclusive}, {Determination: Inconclusive},
	})
	if suppressed.Band == BandCoherent {
		t.Errorf("two sections of eight reached the band %q; a scan that established almost nothing must not reach the band that says the environment agrees", suppressed.Band)
	}
	if full.BotLikeness <= suppressed.BotLikeness {
		t.Errorf("botLikeness: a scan carrying a contradiction scored %d and a scan that established almost nothing scored %d; the score is what must separate them",
			full.BotLikeness, suppressed.BotLikeness)
	}
	if suppressed.HumanConfidence > humanConfidenceCap {
		t.Errorf("humanConfidence = %d above the cap", suppressed.HumanConfidence)
	}
}

func sectionsOf(consistent, contradiction, instrumented, unverified int) []Section {
	var out []Section
	for i := 0; i < consistent; i++ {
		out = append(out, Section{Determination: Consistent})
	}
	for i := 0; i < contradiction; i++ {
		out = append(out, Section{Determination: Contradiction})
	}
	for i := 0; i < instrumented; i++ {
		out = append(out, Section{Determination: Instrumented})
	}
	for i := 0; i < unverified; i++ {
		out = append(out, Section{Determination: Unverified})
	}
	return out
}

func TestEachFurtherDisagreementRaisesTheScoreUntilTheScaleRunsOut(t *testing.T) {
	last := -1
	for c := 0; c <= 4; c++ {
		got := summarise(sectionsOf(15-c, c, 0, 7)).BotLikeness
		if got <= last {
			t.Fatalf("%d disagreements scored %d, and %d scored %d; a reading that failed more of its own tests must not read the same as one that failed fewer",
				c, got, c-1, last)
		}
		last = got
	}
	if top := summarise(sectionsOf(9, 6, 0, 7)).BotLikeness; top != botLikenessCap {
		t.Errorf("six disagreements scored %d, want the cap %d; no finite reading establishes certainty", top, botLikenessCap)
	}
}

func TestTheScoreDoesNotDependOnHowManySectionsAgreed(t *testing.T) {
	small := summarise(sectionsOf(3, 1, 0, 0)).BotLikeness
	large := summarise(sectionsOf(39, 1, 0, 0)).BotLikeness
	if small != large {
		t.Errorf("one disagreement scored %d among four sections and %d among forty; evidence of a modification is not diluted by the sections that found none, and adding a section to this library must not lower the score of an environment it already caught",
			small, large)
	}
}

func TestAFindingOnlyADeliberateChangeProducesOutweighsAPlainDisagreement(t *testing.T) {
	deliberate := summarise(append(sectionsOf(14, 0, 0, 7),
		Section{Determination: Instrumented, weight: weightOnlyDeliberate}))
	plain := summarise(sectionsOf(14, 1, 0, 7))
	if deliberate.BotLikeness <= plain.BotLikeness {
		t.Errorf("a finding nothing but a deliberate change produces scored %d and a plain disagreement scored %d; the first has to weigh more",
			deliberate.BotLikeness, plain.BotLikeness)
	}
	if deliberate.Band != BandInstrumented {
		t.Errorf("band = %q, want %q; a reading only a deliberate change produces is a reading that the environment was changed", deliberate.Band, BandInstrumented)
	}
	if deliberate.HumanConfidence != 0 {
		t.Errorf("humanConfidence = %d beside a finding only a deliberate change produces, want 0", deliberate.HumanConfidence)
	}
}

func TestWithdrawingASectionNeverRaisesTheScore(t *testing.T) {
	full := sectionsOf(10, 2, 1, 3)
	base := summarise(full).BotLikeness
	for i := range full {
		withdrawn := make([]Section, 0, len(full))
		withdrawn = append(withdrawn, full[:i]...)
		withdrawn = append(withdrawn, full[i+1:]...)
		if got := summarise(withdrawn).BotLikeness; got > base {
			t.Errorf("withdrawing section %d raised the score from %d to %d; an environment must not buy a better reading by answering less", i, base, got)
		}
	}
}
