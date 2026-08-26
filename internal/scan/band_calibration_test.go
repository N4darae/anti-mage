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
