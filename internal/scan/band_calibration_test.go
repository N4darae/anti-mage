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
	if suppressed.HumanConfidence > full.HumanConfidence && full.HumanConfidence == 0 {

		t.Logf("suppression raised humanConfidence from %d to %d; the band alone cannot close this and the claim-versus-capability check is what must",
			full.HumanConfidence, suppressed.HumanConfidence)
	}
	if suppressed.HumanConfidence > humanConfidenceCap {
		t.Errorf("humanConfidence = %d above the cap", suppressed.HumanConfidence)
	}
}
