package scan

import (
	"strings"
	"testing"
)

func canvasSerialRequest(t *testing.T, body string) Request {
	t.Helper()
	if body == "" {
		return Request{Probes: map[string]Probe{}}
	}
	return Request{Probes: map[string]Probe{
		"canvas.serial": {Status: StatusOK, Value: []byte(body)},
	}}
}

const canvasSerialAgree = `{"dataUrlAvailable":true,"blobAvailable":true,"rawHash":"aaaa","dataUrlHash":"bbbb","dataUrlLength":42350,"dataUrlPixelsMatchRaw":true,"blobHash":"bbbb","blobLength":42350,"blobPixelsMatchRaw":true}`

const canvasSerialDiverge = `{"dataUrlAvailable":true,"blobAvailable":true,"rawHash":"aaaa","dataUrlHash":"bbbb","dataUrlLength":54113,"dataUrlPixelsMatchRaw":true,"blobHash":"cccc","blobLength":31150,"blobPixelsMatchRaw":true}`

const canvasSerialPixelMismatch = `{"dataUrlAvailable":true,"blobAvailable":true,"rawHash":"aaaa","dataUrlHash":"bbbb","dataUrlLength":54113,"dataUrlPixelsMatchRaw":false,"blobHash":"cccc","blobLength":31150,"blobPixelsMatchRaw":true}`

const canvasSerialBlobUnavailable = `{"dataUrlAvailable":true,"blobAvailable":false,"rawHash":"aaaa","dataUrlHash":"bbbb","dataUrlLength":54113,"dataUrlPixelsMatchRaw":true}`

func TestCanvasSerialAlwaysReturnsUnverified(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"not collected", ""},
		{"agreement", canvasSerialAgree},
		{"divergence", canvasSerialDiverge},
		{"pixel mismatch", canvasSerialPixelMismatch},
		{"blob unavailable", canvasSerialBlobUnavailable},
	}
	for _, c := range cases {
		got := sectionCanvasSerial(canvasSerialRequest(t, c.body), Inputs{}, claim{})
		if got.Determination != Unverified {
			t.Errorf("%s: determination = %q, want %q: this reading has not established a normative baseline across builds and must never score", c.name, got.Determination, Unverified)
		}
	}
}

func TestCanvasSerialDoesNotMoveTheScoreOnAnyInputShape(t *testing.T) {
	base := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Inconclusive},
	}
	before := summarise(base)

	cases := []struct {
		name string
		body string
	}{
		{"agreement", canvasSerialAgree},
		{"divergence", canvasSerialDiverge},
		{"missing data", ""},
		{"pixels differ from raw", canvasSerialPixelMismatch},
	}
	for _, c := range cases {
		sec := sectionCanvasSerial(canvasSerialRequest(t, c.body), Inputs{}, claim{})
		sec.ID = "canvasserial"
		after := summarise(append(append([]Section{}, base...), normalise(sec)))
		if before.Band != after.Band {
			t.Errorf("%s: band moved from %q to %q merely by adding this reading", c.name, before.Band, after.Band)
		}
		if before.HumanConfidence != after.HumanConfidence {
			t.Errorf("%s: human confidence moved from %d to %d", c.name, before.HumanConfidence, after.HumanConfidence)
		}
		if before.BotLikeness != after.BotLikeness {
			t.Errorf("%s: bot likeness moved from %d to %d", c.name, before.BotLikeness, after.BotLikeness)
		}
	}
}

func canvasSerialWrongSection(r Request, in Inputs, c claim) Section {
	sec := sectionCanvasSerial(r, in, c)
	for _, row := range sec.Rows {
		if row.Label == "conclusion" && strings.Contains(row.Value, "different bytes") {
			sec.Determination = Contradiction
		}
	}
	return sec
}

func TestCanvasSerialMovementTestHasTeeth(t *testing.T) {
	base := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Inconclusive},
	}
	before := summarise(base)

	wrong := canvasSerialWrongSection(canvasSerialRequest(t, canvasSerialDiverge), Inputs{}, claim{})
	wrong.ID = "canvasserial"
	after := summarise(append(append([]Section{}, base...), normalise(wrong)))

	if before.Band == after.Band && before.HumanConfidence == after.HumanConfidence && before.BotLikeness == after.BotLikeness {
		t.Fatal("a deliberately wrong section that scores divergence as a contradiction left the summary unchanged; the movement test above would not have caught a scoring reading, so it has no teeth")
	}
}

func TestCanvasSerialNamesNoVendorInItsConclusion(t *testing.T) {
	got := sectionCanvasSerial(canvasSerialRequest(t, canvasSerialDiverge), Inputs{}, claim{})
	for _, row := range got.Rows {
		if row.Label != "conclusion" {
			continue
		}
		for _, forbidden := range []string{"Chrome", "Chromium", "NVIDIA"} {
			if strings.Contains(row.Value+" "+row.Note, forbidden) {
				t.Errorf("the conclusion names %q: %q / %q", forbidden, row.Value, row.Note)
			}
		}
	}
}

func TestCanvasSerialAbstainsWhenBlobIsUnavailable(t *testing.T) {
	got := sectionCanvasSerial(canvasSerialRequest(t, canvasSerialBlobUnavailable), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q", got.Determination, Unverified)
	}
	found := false
	for _, row := range got.Rows {
		if strings.Contains(row.Value, "unavailable") {
			found = true
		}
	}
	if !found {
		t.Error("expected a row noting one of the paths was unavailable")
	}
}

func TestCanvasSerialAbstainsWhenDecodedPixelsDoNotMatchRaw(t *testing.T) {
	got := sectionCanvasSerial(canvasSerialRequest(t, canvasSerialPixelMismatch), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q", got.Determination, Unverified)
	}
	for _, row := range got.Rows {
		if row.Label == "conclusion" && !strings.Contains(row.Value, "did not decode back") {
			t.Errorf("conclusion did not mention the pixel mismatch: %q", row.Value)
		}
	}
}

func TestCanvasSerialRecordsAgreementAndDivergenceDifferently(t *testing.T) {
	agree := sectionCanvasSerial(canvasSerialRequest(t, canvasSerialAgree), Inputs{}, claim{})
	diverge := sectionCanvasSerial(canvasSerialRequest(t, canvasSerialDiverge), Inputs{}, claim{})

	if agree.Determination != Unverified || diverge.Determination != Unverified {
		t.Fatal("both shapes must remain unverified")
	}

	agreeConclusion, divergeConclusion := "", ""
	for _, row := range agree.Rows {
		if row.Label == "conclusion" {
			agreeConclusion = row.Value
		}
	}
	for _, row := range diverge.Rows {
		if row.Label == "conclusion" {
			divergeConclusion = row.Value
		}
	}
	if agreeConclusion == divergeConclusion {
		t.Error("agreement and divergence produced the same conclusion text; the reading is not recording what it saw")
	}
	if !strings.Contains(agreeConclusion, "identical bytes") {
		t.Errorf("agreement conclusion = %q", agreeConclusion)
	}
	if !strings.Contains(divergeConclusion, "different bytes") {
		t.Errorf("divergence conclusion = %q", divergeConclusion)
	}
}
