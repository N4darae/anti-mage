package scan

import "testing"

const engine149UA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36"
const engine151UA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36"

func engineFeaturesRequest(t *testing.T, features string) Request {
	t.Helper()
	r := Request{Probes: map[string]Probe{}}
	if features != "" {
		r.Probes["engine.features"] = Probe{Status: StatusOK, Value: []byte(features)}
	}
	return r
}

func TestEngineVersionAbstainsWhenNoVersionCanBeParsed(t *testing.T) {
	got := sectionEngineVersion(engineFeaturesRequest(t, `{"rubyOverhang":false}`), Inputs{}, claim{UserAgent: "some browser, no version here"})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a claim with no parseable version leaves this reading nothing to compare", got.Determination, Unverified)
	}
}

func TestEngineVersionAbstainsWhenNoFeaturesWereCollected(t *testing.T) {
	got := sectionEngineVersion(Request{Probes: map[string]Probe{}}, Inputs{}, claim{UserAgent: engine149UA})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q", got.Determination, Unverified)
	}
}

func TestEngineVersionAbstainsWhenClaimAlreadyCoversEveryVerifiedFeature(t *testing.T) {
	got := sectionEngineVersion(engineFeaturesRequest(t, `{"rubyOverhang":true,"animationEventAnimation":true,"transitionEventAnimation":true,"userMediaElement":true,"softNavigations":true}`), Inputs{}, claim{UserAgent: engine151UA})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a claimed version at or beyond every verified row leaves nothing later to test", got.Determination, Unverified)
	}
}

func TestEngineVersionIsConsistentWhenNoLaterCapabilityIsPresent(t *testing.T) {
	got := sectionEngineVersion(engineFeaturesRequest(t, `{"rubyOverhang":false,"animationEventAnimation":false,"transitionEventAnimation":false,"userMediaElement":false,"softNavigations":false}`), Inputs{}, claim{UserAgent: engine149UA})
	if got.Determination != Consistent {
		t.Fatalf("determination = %q, want %q: this is exactly what the paladin session in this round reported, and its engine genuinely was 149", got.Determination, Consistent)
	}
}

func TestEngineVersionIsContradictionWhenALaterCapabilityIsPresent(t *testing.T) {
	got := sectionEngineVersion(engineFeaturesRequest(t, `{"rubyOverhang":true,"animationEventAnimation":false,"transitionEventAnimation":false,"userMediaElement":false,"softNavigations":false}`), Inputs{}, claim{UserAgent: engine149UA})
	if got.Determination != Contradiction {
		t.Fatalf("determination = %q, want %q: a capability tabulated as shipping at 151 cannot exist in a genuine 149 engine", got.Determination, Contradiction)
	}
}

func TestEngineVersionDoesNotScoreAMissingCapabilityAsEvidence(t *testing.T) {
	got := sectionEngineVersion(engineFeaturesRequest(t, `{}`), Inputs{}, claim{UserAgent: engine149UA})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: with no feature keys reported, no row is applicable, and absence must never be scored", got.Determination, Unverified)
	}
}

func TestEngineVersionIgnoresUnverifiedRows(t *testing.T) {

	got := sectionEngineVersion(engineFeaturesRequest(t, `{"textFit":true,"bgClipBorderArea":true}`), Inputs{}, claim{UserAgent: engine149UA})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: textFit and bgClipBorderArea are tabulated but not Verified, so their presence must not drive a verdict", got.Determination, Unverified)
	}
}

func TestEngineVersionDoesNotDiluteAPayloadThatPredatesIt(t *testing.T) {
	sections := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Inconclusive},
	}
	before := summarise(sections)

	absent := sectionEngineVersion(Request{Probes: map[string]Probe{}}, Inputs{}, claim{})
	absent.ID = "engineversion"
	after := summarise(append(sections, normalise(absent)))

	if before.Band != after.Band {
		t.Errorf("band moved from %q to %q merely by adding a reading the payload has no data for", before.Band, after.Band)
	}
	if before.HumanConfidence != after.HumanConfidence {
		t.Errorf("human confidence moved from %d to %d", before.HumanConfidence, after.HumanConfidence)
	}
	if before.BotLikeness != after.BotLikeness {
		t.Errorf("bot likeness moved from %d to %d", before.BotLikeness, after.BotLikeness)
	}
}
