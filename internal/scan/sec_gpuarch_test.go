package scan

import "testing"

const blackwellRenderer = "ANGLE (NVIDIA, NVIDIA GeForce RTX 5070 Ti Laptop GPU (0x00002F18) Direct3D11 vs_5_0 ps_5_0, D3D11)"

const rendererWithoutIdentifier = "ANGLE (NVIDIA Corporation, NVIDIA GeForce RTX 3060 Ti/PCIe/SSE2, OpenGL 4.5.0)"

const renoirRenderer = "ANGLE (AMD, AMD Radeon(TM) R3 Graphics (0x00001636) Direct3D11 vs_5_0 ps_5_0, D3D11)"

func adapterEverywhere(vendor, architecture string) string {
	one := `{"adapter":true,"vendor":"` + vendor + `","architecture":"` + architecture + `"}`
	return `{"present":true,"secureContext":true,"variants":{
		"default":` + one + `,
		"highPerformance":` + one + `,
		"lowPower":` + one + `,
		"fallback":{"adapter":false}}}`
}

func TestGPUArchAgreesWhenTheDeviceNamedIsTheGenerationReported(t *testing.T) {
	got := sectionGPUArch(webgpuRequest(t, adapterEverywhere("nvidia", "ampere"), hardwareRenderer), Inputs{}, claim{})
	if got.Determination != Consistent {
		t.Fatalf("determination = %q, want %q", got.Determination, Consistent)
	}
}

func TestGPUArchReadsADeviceFromOneGenerationReportedAsAnother(t *testing.T) {
	got := sectionGPUArch(webgpuRequest(t, adapterEverywhere("nvidia", "ampere"), blackwellRenderer), Inputs{}, claim{})
	if got.Determination != Contradiction {
		t.Fatalf("determination = %q, want %q", got.Determination, Contradiction)
	}
	if got.weighs() != weightOnlyDeliberate {
		t.Fatalf("weight = %d, want %d", got.weighs(), weightOnlyDeliberate)
	}
}

func TestGPUArchCarriesNoWeightWhenTheDeviceNamedCarriesNoIdentifier(t *testing.T) {
	got := sectionGPUArch(webgpuRequest(t, adapterEverywhere("nvidia", "ampere"), rendererWithoutIdentifier), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a backend that names no identifier settles nothing", got.Determination, Unverified)
	}
}

func TestGPUArchCarriesNoWeightWhenThePowerPreferencesReachDifferentDevices(t *testing.T) {
	body := `{"present":true,"secureContext":true,"variants":{
		"default":{"adapter":true,"vendor":"nvidia","architecture":"ampere"},
		"highPerformance":{"adapter":true,"vendor":"nvidia","architecture":"ampere"},
		"lowPower":{"adapter":true,"vendor":"intel","architecture":"gen-12-lp"},
		"fallback":{"adapter":false}}}`
	got := sectionGPUArch(webgpuRequest(t, body, hardwareRenderer), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a machine with two graphics devices may answer each request with a different one", got.Determination, Unverified)
	}
}

func TestGPUArchCarriesNoWeightWhenOneRequestReachedNoDevice(t *testing.T) {
	body := `{"present":true,"secureContext":true,"variants":{
		"default":{"adapter":true,"vendor":"nvidia","architecture":"ampere"},
		"highPerformance":{"adapter":true,"vendor":"nvidia","architecture":"ampere"},
		"lowPower":{"adapter":false},
		"fallback":{"adapter":false}}}`
	got := sectionGPUArch(webgpuRequest(t, body, hardwareRenderer), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q", got.Determination, Unverified)
	}
}

func TestGPUArchCarriesNoWeightWhenNoGenerationIsReported(t *testing.T) {
	got := sectionGPUArch(webgpuRequest(t, adapterEverywhere("", ""), hardwareRenderer), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: an engine that names no generation is never read as lacking one", got.Determination, Unverified)
	}
}

func TestGPUArchCarriesNoWeightWhenTheGenerationReportedIsNotConfirmedOnRealHardware(t *testing.T) {
	got := sectionGPUArch(webgpuRequest(t, adapterEverywhere("amd", "rdna-2"), renoirRenderer), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: an unconfirmed table entry is never read as evidence", got.Determination, Unverified)
	}
}

func TestGPUArchCarriesNoWeightForASoftwareRenderer(t *testing.T) {
	renderer := "ANGLE (Google, Vulkan 1.3.0 (SwiftShader Device (Subzero) (0x0000C0DE)), SwiftShader driver)"
	got := sectionGPUArch(webgpuRequest(t, adapterEverywhere("google", "swiftshader"), renderer), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q", got.Determination, Unverified)
	}
}

func TestGPUArchCarriesNoWeightWhenTheBrowserSaysTheNameStandsInForAnother(t *testing.T) {
	renderer := "ANGLE (NVIDIA, NVIDIA GeForce GTX 980 Direct3D11 vs_5_0 ps_5_0), or similar"
	got := sectionGPUArch(webgpuRequest(t, adapterEverywhere("nvidia", "ampere"), renderer), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: an engine that substitutes a device name says so, and is not read as hiding one", got.Determination, Unverified)
	}
}

func TestGPUArchIsNotSwitchedOffByTextAppendedToTheDeviceName(t *testing.T) {
	renderer := blackwellRenderer + ", or similar"
	got := sectionGPUArch(webgpuRequest(t, adapterEverywhere("nvidia", "ampere"), renderer), Inputs{}, claim{})
	if got.Determination != Contradiction {
		t.Fatalf("determination = %q, want %q: free text an environment writes about itself must not decide this reading", got.Determination, Contradiction)
	}
}

func TestGPUArchCarriesNoWeightForAnIdentifierTheTableDoesNotPlace(t *testing.T) {
	renderer := "ANGLE (NVIDIA, NVIDIA GeForce ZZZ (0x0000AAAA) Direct3D11 vs_5_0 ps_5_0, D3D11)"
	got := sectionGPUArch(webgpuRequest(t, adapterEverywhere("nvidia", "ampere"), renderer), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: silicon newer than this project's table settles nothing", got.Determination, Unverified)
	}
}

func TestHWDecodeCarriesNoWeightOnceTheDeviceNamedIsShownNotToBePresent(t *testing.T) {
	r := webgpuRequest(t, adapterEverywhere("nvidia", "ampere"), blackwellRenderer)
	r.Probes["media.matrix"] = Probe{Status: StatusOK, Value: []byte(`{
		"H.264 high, MP4":{"powerEfficient":true},
		"AV1, WebM":{"decodingInfoSupported":true,"powerEfficient":true}}`)}
	got := sectionHWDecode(r, Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: one rewritten device name must not be read as two bodies of evidence", got.Determination, Unverified)
	}
}
