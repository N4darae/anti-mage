package reference

import (
	"strings"
	"testing"
)

func TestEveryGPUArchitectureVendorCarriesItsSource(t *testing.T) {
	if len(GPUArchitectures) == 0 {
		t.Fatal("the table is empty")
	}
	for vendor, entry := range GPUArchitectures {
		if entry.PCIVendor == 0 {
			t.Errorf("vendor %q carries no PCI identifier", vendor)
		}
		if len(entry.Blocks) == 0 {
			t.Errorf("vendor %q places no device", vendor)
		}
		if entry.Source.Origin == "" {
			t.Errorf("vendor %q carries no source origin", vendor)
		}
		if entry.Source.Checked == "" {
			t.Errorf("vendor %q carries no date it was checked", vendor)
		}
	}
}

func TestOnlyArchitecturesObservedOnRealHardwareAreVerified(t *testing.T) {
	verified := 0
	for vendor, entry := range GPUArchitectures {
		for architecture, observed := range entry.Observed {
			verified++
			if strings.TrimSpace(observed) == "" {
				t.Errorf("%s %s is verified but names no system it was observed on", vendor, architecture)
			}
			if !entry.Verified(architecture) {
				t.Errorf("%s %s is observed but does not read as verified", vendor, architecture)
			}
		}
	}
	if verified == 0 {
		t.Fatal("no architecture is verified, so this table can never be read as evidence")
	}
}

func TestEveryArchitectureNameIsSpeltTheWayTheInterfaceReportsIt(t *testing.T) {
	for vendor, entry := range GPUArchitectures {
		for _, block := range entry.Blocks {
			for device, architecture := range block.Devices {
				if architecture != strings.ToLower(architecture) || strings.Contains(architecture, " ") {
					t.Errorf("%s 0x%08X names architecture %q, which is not spelt the way the interface reports it", vendor, device, architecture)
				}
			}
		}
	}
}

func TestTheObservedDeviceIsPlacedInTheGenerationItWasObservedIn(t *testing.T) {
	nvidia, ok := GPUArchitectures["nvidia"]
	if !ok {
		t.Fatal("the table carries no entry for the manufacturer observed")
	}
	if !nvidia.Verified("ampere") {
		t.Fatal("the generation observed on real hardware does not read as verified")
	}
	got, placed := nvidia.Architecture(0x2489)
	if !placed || got != "ampere" {
		t.Fatalf("Architecture(0x2489) = %q, %v, want %q, true", got, placed, "ampere")
	}
}

func TestDevicesArePlacedThroughTheMaskTheirManufacturerUses(t *testing.T) {
	for _, c := range []struct {
		vendor string
		device uint32
		want   string
	}{
		{"nvidia", 0x2F18, "blackwell"},
		{"nvidia", 0x2F58, "blackwell"},
		{"nvidia", 0x2704, "lovelace"},
		{"nvidia", 0x2489, "ampere"},
		{"amd", 0x1636, "gcn-5"},
		{"amd", 0x163F, "rdna-2"},
		{"amd", 0x67DF, "gcn-4"},
		{"intel", 0x3E9B, "gen-9"},
	} {
		entry, ok := GPUArchitectures[c.vendor]
		if !ok {
			t.Fatalf("the table carries no entry for %q", c.vendor)
		}
		got, placed := entry.Architecture(c.device)
		if !placed || got != c.want {
			t.Errorf("%s Architecture(0x%04X) = %q, %v, want %q, true", c.vendor, c.device, got, placed, c.want)
		}
	}
}

func TestSiliconTheTableDoesNotCarryIsNotPlaced(t *testing.T) {
	nvidia := GPUArchitectures["nvidia"]
	if got, placed := nvidia.Architecture(0xAAAA); placed {
		t.Fatalf("Architecture(0xAAAA) = %q, true, want a device the table declines to place", got)
	}
}
