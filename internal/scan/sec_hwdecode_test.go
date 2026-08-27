package scan

import (
	"strings"
	"testing"
)

func hwDecodeRequest(t *testing.T, renderer string, controlEfficient, av1Efficient any, av1Supported any) Request {
	t.Helper()
	r := Request{Probes: map[string]Probe{}}
	if renderer != "" {
		r.Probes["gpu.renderer"] = Probe{
			Status: StatusOK,
			Value:  []byte(`{"unmaskedRenderer":"` + renderer + `"}`),
		}
	}
	matrix := `{"H.264 high, MP4":{"contentType":"video/mp4","decodingInfoSupported":true,"powerEfficient":` +
		jsonBool(controlEfficient) + `},` +
		`"AV1, WebM":{"contentType":"video/webm","decodingInfoSupported":` + jsonBool(av1Supported) +
		`,"powerEfficient":` + jsonBool(av1Efficient) + `}}`
	r.Probes["media.matrix"] = Probe{Status: StatusOK, Value: []byte(matrix)}
	return r
}

func jsonBool(v any) string {
	switch v {
	case true:
		return "true"
	case false:
		return "false"
	}
	return "null"
}

const ampereRenderer = "ANGLE (NVIDIA, NVIDIA GeForce RTX 3060 Ti (0x00002489) Direct3D11 vs_5_0 ps_5_0, D3D11)"
const adaRenderer = "ANGLE (NVIDIA, NVIDIA GeForce RTX 4070 Ti (0x00002782) Direct3D11 vs_5_0 ps_5_0, D3D11)"

func TestHWDecodeCarriesNoWeightWhenNoRendererWasCollected(t *testing.T) {
	got := sectionHWDecode(hwDecodeRequest(t, "", true, false, true), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a payload that names no device leaves this reading nothing to apply to, and it must not enter the count", got.Determination, Unverified)
	}
}

func TestHWDecodeDoesNotDiluteAPayloadThatPredatesIt(t *testing.T) {
	sections := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Inconclusive},
	}
	before := summarise(sections)

	absent := sectionHWDecode(Request{Probes: map[string]Probe{}}, Inputs{}, claim{})
	absent.ID = "hwdecode"
	after := summarise(append(sections, normalise(absent)))

	if before.Band != after.Band {
		t.Errorf("band moved from %q to %q merely by adding a reading the payload has no data for", before.Band, after.Band)
	}
	if before.HumanConfidence != after.HumanConfidence {
		t.Errorf("human confidence moved from %d to %d", before.HumanConfidence, after.HumanConfidence)
	}
}

func TestHWDecodeCarriesNoWeightOnARendererItCannotPlace(t *testing.T) {
	for _, renderer := range []string{
		"ANGLE (Intel, Intel(R) UHD Graphics 620 Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"ANGLE (AMD, AMD Radeon RX 6800 XT Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"Apple GPU",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 1080 Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"ANGLE (NVIDIA, NVIDIA Quadro P2000 Direct3D11 vs_5_0 ps_5_0, D3D11)",
	} {
		got := sectionHWDecode(hwDecodeRequest(t, renderer, true, false, true), Inputs{}, claim{})
		if got.Determination != Unverified {
			t.Errorf("renderer %q: determination = %q, want %q: a device outside this project's table leaves the reading nothing to apply", renderer, got.Determination, Unverified)
		}
	}
}

func TestHWDecodeYieldsUnverifiedForAGenerationNotObservedOnRealHardware(t *testing.T) {
	got := sectionHWDecode(hwDecodeRequest(t, adaRenderer, true, false, true), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: the 40 series table entry is not verified", got.Determination, Unverified)
	}
}

func TestHWDecodeReportsContradictionWhenAVerifiedGenerationLacksItsOwnDecoder(t *testing.T) {
	got := sectionHWDecode(hwDecodeRequest(t, ampereRenderer, true, false, true), Inputs{}, claim{})
	if got.Determination != Contradiction {
		t.Fatalf("determination = %q, want %q", got.Determination, Contradiction)
	}
}

func TestHWDecodeNamesNoVendorInItsConclusion(t *testing.T) {
	got := sectionHWDecode(hwDecodeRequest(t, ampereRenderer, true, false, true), Inputs{}, claim{})
	for _, row := range got.Rows {
		if row.Label != "conclusion" {
			continue
		}
		for _, forbidden := range []string{"NVIDIA", "GeForce", "RTX", "3060"} {
			if strings.Contains(row.Value+" "+row.Note, forbidden) {
				t.Errorf("the conclusion names %q: %q / %q", forbidden, row.Value, row.Note)
			}
		}
	}
}

func TestHWDecodeAgreesWhenTheDecoderIsReportedAsPresent(t *testing.T) {
	got := sectionHWDecode(hwDecodeRequest(t, ampereRenderer, true, true, true), Inputs{}, claim{})
	if got.Determination != Consistent {
		t.Fatalf("determination = %q, want %q", got.Determination, Consistent)
	}
}

func TestHWDecodeCarriesNoWeightWhenNoHardwareDecoderWasDemonstrated(t *testing.T) {
	got := sectionHWDecode(hwDecodeRequest(t, ampereRenderer, false, false, true), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: without a control codec decoded in hardware, nothing was shown to be missing", got.Determination, Unverified)
	}
}

func TestHWDecodeCarriesNoWeightWhenTheCodecIsNotSupportedAtAll(t *testing.T) {
	got := sectionHWDecode(hwDecodeRequest(t, ampereRenderer, true, false, false), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a codec this build cannot decode at all is an absence", got.Determination, Unverified)
	}
}

func TestHWDecodeCarriesNoWeightWhenAnAnswerIsMissing(t *testing.T) {
	cases := []struct {
		name                    string
		control, av1, supported any
	}{
		{"no control answer", nil, false, true},
		{"no codec answer", true, nil, true},
		{"no support answer", true, false, nil},
	}
	for _, c := range cases {
		got := sectionHWDecode(hwDecodeRequest(t, ampereRenderer, c.control, c.av1, c.supported), Inputs{}, claim{})
		if got.Determination != Unverified {
			t.Errorf("%s: determination = %q, want %q", c.name, got.Determination, Unverified)
		}
	}
}

func TestHWDecodeCarriesNoWeightWhenTheMatrixWasNotCollected(t *testing.T) {
	r := Request{Probes: map[string]Probe{
		"gpu.renderer": {Status: StatusOK, Value: []byte(`{"unmaskedRenderer":"` + ampereRenderer + `"}`)},
	}}
	got := sectionHWDecode(r, Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q", got.Determination, Unverified)
	}
}

func TestHWDecodeIsOneOfTheSectionsAnalyzeBuilds(t *testing.T) {
	found := false
	for _, s := range order {
		if s.id == "hwdecode" {
			found = true
		}
	}
	if !found {
		t.Fatal("the reading is not registered in the section order, so no scan runs it")
	}
}

func TestAnUnverifiedSectionDoesNotMoveTheScore(t *testing.T) {
	withoutReading := summarise([]Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
	})
	withUnverified := summarise([]Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "hwdecode", Determination: Unverified},
	})
	if withoutReading.BotLikeness != withUnverified.BotLikeness {
		t.Errorf("bot likeness moved from %d to %d", withoutReading.BotLikeness, withUnverified.BotLikeness)
	}
	if withoutReading.HumanConfidence != withUnverified.HumanConfidence {
		t.Errorf("human confidence moved from %d to %d", withoutReading.HumanConfidence, withUnverified.HumanConfidence)
	}
	if withoutReading.Band != withUnverified.Band {
		t.Errorf("band moved from %q to %q", withoutReading.Band, withUnverified.Band)
	}
}

func TestHWDecodeCarriesNoWeightOnEveryPathThatSettlesNothing(t *testing.T) {
	cases := []struct {
		name string
		req  Request
	}{
		{"no renderer", hwDecodeRequest(t, "", true, false, true)},
		{"renderer this project cannot place", hwDecodeRequest(t, "ANGLE (Intel, Intel(R) UHD Graphics 620 Direct3D11 vs_5_0 ps_5_0, D3D11)", true, false, true)},
		{"generation not observed on real hardware", hwDecodeRequest(t, adaRenderer, true, false, true)},
		{"no control answer", hwDecodeRequest(t, ampereRenderer, nil, false, true)},
		{"no codec answer", hwDecodeRequest(t, ampereRenderer, true, nil, true)},
		{"no support answer", hwDecodeRequest(t, ampereRenderer, true, false, nil)},
		{"no hardware decoder demonstrated", hwDecodeRequest(t, ampereRenderer, false, false, true)},
		{"codec this build cannot decode at all", hwDecodeRequest(t, ampereRenderer, true, false, false)},
		{"matrix not collected", Request{Probes: map[string]Probe{
			"gpu.renderer": {Status: StatusOK, Value: []byte(`{"unmaskedRenderer":"` + ampereRenderer + `"}`)},
		}}},
	}

	settled := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Inconclusive},
	}
	before := summarise(settled)

	for _, c := range cases {
		got := sectionHWDecode(c.req, Inputs{}, claim{})
		if got.Determination != Unverified {
			t.Errorf("%s: determination = %q, want %q", c.name, got.Determination, Unverified)
		}
		got.ID = "hwdecode"
		after := summarise(append(settled, normalise(got)))
		if after.Band != before.Band || after.HumanConfidence != before.HumanConfidence || after.BotLikeness != before.BotLikeness {
			t.Errorf("%s: the summary moved from %+v to %+v", c.name, before, after)
		}
	}
}
