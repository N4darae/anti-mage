package scan

import "testing"

const hardwareRenderer = "ANGLE (NVIDIA, NVIDIA GeForce RTX 3060 Ti (0x00002489) Direct3D11 vs_5_0 ps_5_0, D3D11)"

func webgpuRequest(t *testing.T, adapter string, renderer string) Request {
	t.Helper()
	r := Request{Probes: map[string]Probe{}}
	if adapter != "" {
		r.Probes["gpu.adapter"] = Probe{Status: StatusOK, Value: []byte(adapter)}
	}
	if renderer != "" {
		r.Probes["gpu.renderer"] = Probe{Status: StatusOK, Value: []byte(`{"unmaskedRenderer":"` + renderer + `"}`)}
	}
	return r
}

const adapterNoneAnywhere = `{"present":true,"secureContext":true,"variants":{
	"default":{"adapter":false},
	"highPerformance":{"adapter":false},
	"lowPower":{"adapter":false},
	"fallback":{"adapter":false}}}`

const adapterOnlyFallbackMissing = `{"present":true,"secureContext":true,"variants":{
	"default":{"adapter":true,"vendor":"nvidia","architecture":"ampere","maxTextureDimension2D":16384},
	"highPerformance":{"adapter":true,"vendor":"nvidia","architecture":"ampere","maxTextureDimension2D":16384},
	"lowPower":{"adapter":true,"vendor":"nvidia","architecture":"ampere","maxTextureDimension2D":16384},
	"fallback":{"adapter":false}}}`

func TestWebGPUReportsModificationWhenHardwareIsNamedAndNoAdapterIsGranted(t *testing.T) {
	got := sectionWebGPU(webgpuRequest(t, adapterNoneAnywhere, hardwareRenderer), Inputs{}, claim{})
	if got.Determination != Instrumented {
		t.Fatalf("determination = %q, want %q", got.Determination, Instrumented)
	}
}

func TestWebGPUAgreesWhenOnlyTheFallbackPathIsEmpty(t *testing.T) {
	got := sectionWebGPU(webgpuRequest(t, adapterOnlyFallbackMissing, hardwareRenderer), Inputs{}, claim{})
	if got.Determination != Consistent {
		t.Fatalf("determination = %q, want %q: no software backend on a machine with real hardware is ordinary", got.Determination, Consistent)
	}
}

func TestWebGPUCarriesNoWeightWhenItWasNotCollected(t *testing.T) {
	got := sectionWebGPU(webgpuRequest(t, "", hardwareRenderer), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q", got.Determination, Unverified)
	}
}

func TestWebGPUCarriesNoWeightWhenTheInterfaceIsAbsent(t *testing.T) {
	body := `{"present":false,"secureContext":true,"variants":{}}`
	got := sectionWebGPU(webgpuRequest(t, body, hardwareRenderer), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a browser without the interface is never scored for lacking it", got.Determination, Unverified)
	}
}

func TestWebGPUCarriesNoWeightOutsideASecureContext(t *testing.T) {
	body := `{"present":false,"secureContext":false,"variants":{}}`
	got := sectionWebGPU(webgpuRequest(t, body, hardwareRenderer), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: the interface is gated here, so its absence says nothing", got.Determination, Unverified)
	}
}

func TestWebGPUCarriesNoWeightWhenNoDeviceWasNamed(t *testing.T) {
	got := sectionWebGPU(webgpuRequest(t, adapterNoneAnywhere, ""), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: with no device named there is nothing for the empty adapter to disagree with", got.Determination, Unverified)
	}
}

func TestWebGPUCarriesNoWeightWhenTheDeviceNamedIsASoftwareRasteriser(t *testing.T) {
	for _, renderer := range []string{
		"Google SwiftShader",
		"ANGLE (Google, Vulkan 1.3.0 (SwiftShader Device (Subzero)), SwiftShader driver)",
		"Microsoft Basic Render Driver",
		"llvmpipe (LLVM 15.0.7, 256 bits)",
		"ANGLE (Software Adapter, Direct3D11 vs_5_0 ps_5_0, D3D11)",
	} {
		got := sectionWebGPU(webgpuRequest(t, adapterNoneAnywhere, renderer), Inputs{}, claim{})
		if got.Determination != Unverified {
			t.Errorf("renderer %q: determination = %q, want %q: an environment already drawing in software is not contradicted by having no hardware adapter", renderer, got.Determination, Unverified)
		}
	}
}

func TestWebGPUCarriesNoWeightWhenAPathWasNotReported(t *testing.T) {
	body := `{"present":true,"secureContext":true,"variants":{"fallback":{"adapter":false}}}`
	got := sectionWebGPU(webgpuRequest(t, body, hardwareRenderer), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q", got.Determination, Unverified)
	}
}

func TestWebGPURaisesTenAndNoMore(t *testing.T) {
	flagged := summarise([]Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Consistent},
		{ID: "d", Determination: Consistent},
		{ID: "webgpu", Determination: Instrumented},
	})
	if flagged.Band != BandInstrumented {
		t.Fatalf("band = %q, want %q", flagged.Band, BandInstrumented)
	}
	if flagged.BotLikeness != 10 {
		t.Fatalf("botLikeness = %d, want 10", flagged.BotLikeness)
	}
}

func TestWebGPUDoesNotDiluteAPayloadThatPredatesIt(t *testing.T) {
	sections := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Inconclusive},
	}
	before := summarise(sections)
	absent := sectionWebGPU(Request{Probes: map[string]Probe{}}, Inputs{}, claim{})
	absent.ID = "webgpu"
	after := summarise(append(sections, normalise(absent)))

	if before.Band != after.Band || before.HumanConfidence != after.HumanConfidence {
		t.Errorf("adding a reading the payload has no data for moved the summary from %+v to %+v", before, after)
	}
}

func TestWebGPUIsOneOfTheSectionsAnalyzeBuilds(t *testing.T) {
	found := false
	for _, s := range order {
		if s.id == "webgpu" {
			found = true
		}
	}
	if !found {
		t.Fatal("the reading is not registered in the section order, so no scan runs it")
	}
}

const adapterOnlyFallbackGranted = `{"present":true,"secureContext":true,"variants":{
	"default":{"adapter":false},
	"highPerformance":{"adapter":false},
	"lowPower":{"adapter":false},
	"fallback":{"adapter":true,"vendor":"","architecture":"","maxTextureDimension2D":16384}}}`

func TestWebGPUContradictsWhenEveryHardwarePathIsRefusedAndTheFallbackIsGranted(t *testing.T) {
	got := sectionWebGPU(webgpuRequest(t, adapterOnlyFallbackGranted, hardwareRenderer), Inputs{}, claim{})
	if got.Determination != Contradiction {
		t.Fatalf("determination = %q, want %q: an interface that grants its own software backend was never short of a device to grant", got.Determination, Contradiction)
	}
}

func TestWebGPUStaysAModificationWhenTheFallbackPathWasNotReported(t *testing.T) {
	body := `{"present":true,"secureContext":true,"variants":{
		"default":{"adapter":false},
		"highPerformance":{"adapter":false},
		"lowPower":{"adapter":false}}}`
	got := sectionWebGPU(webgpuRequest(t, body, hardwareRenderer), Inputs{}, claim{})
	if got.Determination != Instrumented {
		t.Fatalf("determination = %q, want %q: without the fallback answer there is no second fact to contradict the first", got.Determination, Instrumented)
	}
}
