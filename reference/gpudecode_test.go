package reference

import "testing"

func TestEveryGeForceDecodeEntryCarriesItsSource(t *testing.T) {
	if len(NvidiaGeForceDecode) == 0 {
		t.Fatal("the table is empty")
	}
	for series, entry := range NvidiaGeForceDecode {
		if entry.Family == "" {
			t.Errorf("series %q names no architecture", series)
		}
		if entry.Source.Origin == "" {
			t.Errorf("series %q carries no source origin", series)
		}
		if entry.Source.Checked == "" {
			t.Errorf("series %q carries no date it was checked", series)
		}
	}
}

func TestOnlyEntriesObservedOnRealHardwareAreVerified(t *testing.T) {
	verified := []string{}
	for series, entry := range NvidiaGeForceDecode {
		if entry.Verified {
			verified = append(verified, series)
		}
	}
	if len(verified) == 0 {
		t.Fatal("no entry is verified, so this table can never be read as evidence")
	}
	for _, series := range verified {
		if NvidiaGeForceDecode[series].Observed == "" {
			t.Errorf("series %q is verified but names no system it was observed on", series)
		}
	}
}

func TestAmpereIsTheObservedGenerationAndCarriesAV1Decode(t *testing.T) {
	entry, ok := NvidiaGeForceDecode["30"]
	if !ok {
		t.Fatal("the 30 series is absent from the table")
	}
	if !entry.Verified {
		t.Error("the 30 series was observed on real hardware and should be verified")
	}
	if !entry.AV1Decode {
		t.Error("Ampere carries an AV1 decoder")
	}
}

func TestTuringCarriesNoAV1DecoderAndIsNotVerified(t *testing.T) {
	entry, ok := NvidiaGeForceDecode["20"]
	if !ok {
		t.Fatal("the 20 series is absent from the table")
	}
	if entry.AV1Decode {
		t.Error("Turing carries no AV1 decoder")
	}
	if entry.Verified {
		t.Error("the 20 series has not been observed here, so it must not be verified")
	}
}

func TestGenerationsNotObservedHereAreNotVerified(t *testing.T) {
	for _, series := range []string{"40", "50"} {
		entry, ok := NvidiaGeForceDecode[series]
		if !ok {
			t.Fatalf("the %s series is absent from the table", series)
		}
		if entry.Verified {
			t.Errorf("the %s series has not been observed here, so it must not be verified", series)
		}
	}
}
