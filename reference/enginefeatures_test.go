package reference

import "testing"

func TestEngineFeaturesCarrySourceAndIDs(t *testing.T) {
	seen := map[string]bool{}
	for _, fe := range EngineFeatures {
		if fe.ID == "" {
			t.Fatalf("engine feature with empty ID: %+v", fe)
		}
		if seen[fe.ID] {
			t.Fatalf("duplicate engine feature ID %q", fe.ID)
		}
		seen[fe.ID] = true
		if fe.Source.Origin == "" {
			t.Fatalf("engine feature %q has no source", fe.ID)
		}
		if fe.ShipsInMajor <= 0 {
			t.Fatalf("engine feature %q has no tabulated version", fe.ID)
		}
		if fe.Verified && fe.Observed == "" {
			t.Fatalf("engine feature %q is marked Verified but records no observation", fe.ID)
		}
	}
}

func TestEngineFeaturesOnlyMarkVerifiedWhatWasObservedOnTheExactVersionClaimed(t *testing.T) {
	for _, fe := range EngineFeatures {
		if fe.Verified && fe.ShipsInMajor != 151 {
			t.Fatalf("engine feature %q is marked Verified for a tabulated version this project has not observed on a real system of that exact configuration", fe.ID)
		}
	}
}
