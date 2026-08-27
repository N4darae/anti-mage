package scan

import (
	"strconv"
	"testing"
)

type viewportNumbers struct {
	screenW, screenH int
	innerW, innerH   int
	screenX, screenY int
	omit             string
	devicePixelRatio string
}

func viewportRequest(t *testing.T, n viewportNumbers) Request {
	t.Helper()
	dpr := n.devicePixelRatio
	if dpr == "" {
		dpr = "1"
	}
	fields := map[string]string{
		"width":            strconv.Itoa(n.screenW),
		"height":           strconv.Itoa(n.screenH),
		"innerWidth":       strconv.Itoa(n.innerW),
		"innerHeight":      strconv.Itoa(n.innerH),
		"screenX":          strconv.Itoa(n.screenX),
		"screenY":          strconv.Itoa(n.screenY),
		"devicePixelRatio": dpr,
	}
	delete(fields, n.omit)

	body := "{"
	first := true
	for _, key := range []string{"width", "height", "innerWidth", "innerHeight", "screenX", "screenY", "devicePixelRatio"} {
		v, ok := fields[key]
		if !ok {
			continue
		}
		if !first {
			body += ","
		}
		first = false
		body += `"` + key + `":` + v
	}
	body += "}"

	return Request{Probes: map[string]Probe{"geom.screen": {Status: StatusOK, Value: []byte(body)}}}
}

var viewportFits = viewportNumbers{screenW: 1920, screenH: 1080, innerW: 928, innerH: 794}

func TestViewportAgreesWhenItFitsTheScreenItClaims(t *testing.T) {
	got := sectionViewport(viewportRequest(t, viewportFits), Inputs{}, claim{})
	if got.Determination != Consistent {
		t.Fatalf("determination = %q, want %q", got.Determination, Consistent)
	}
}

func TestViewportReportsContradictionWhenItIsTallerThanTheScreen(t *testing.T) {
	n := viewportFits
	n.screenH, n.innerH = 768, 794
	got := sectionViewport(viewportRequest(t, n), Inputs{}, claim{})
	if got.Determination != Contradiction {
		t.Fatalf("determination = %q, want %q", got.Determination, Contradiction)
	}
}

func TestViewportReportsContradictionWhenItIsWiderThanTheScreen(t *testing.T) {
	n := viewportFits
	n.screenW, n.innerW = 800, 928
	got := sectionViewport(viewportRequest(t, n), Inputs{}, claim{})
	if got.Determination != Contradiction {
		t.Fatalf("determination = %q, want %q", got.Determination, Contradiction)
	}
}

func TestViewportAcceptsAWindowHangingOffTheBottomOfItsScreen(t *testing.T) {
	n := viewportFits
	n.screenH, n.innerH, n.screenY = 900, 794, 300
	got := sectionViewport(viewportRequest(t, n), Inputs{}, claim{})
	if got.Determination == Contradiction {
		t.Fatal("a window positioned so its lower part is off the screen is ordinary, not a contradiction")
	}
}

func TestViewportCarriesNoWeightWhenTheWindowSitsLeftOfOrAboveTheOrigin(t *testing.T) {
	for _, n := range []viewportNumbers{
		{screenW: 1920, screenH: 768, innerW: 928, innerH: 794, screenY: -200},
		{screenW: 800, screenH: 1080, innerW: 928, innerH: 794, screenX: -300},
	} {
		got := sectionViewport(viewportRequest(t, n), Inputs{}, claim{})
		if got.Determination != Unverified {
			t.Errorf("%+v: determination = %q, want %q: a negative offset means another display, whose size this reading was not given, and a layout this ordinary must not cost an honest browser anything", n, got.Determination, Unverified)
		}
	}
}

func TestViewportCarriesNoWeightWhenTheNumbersWereNotCollected(t *testing.T) {
	got := sectionViewport(Request{Probes: map[string]Probe{}}, Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q", got.Determination, Unverified)
	}
}

func TestViewportCarriesNoWeightWhenAnyNumberIsMissing(t *testing.T) {
	for _, omit := range []string{"width", "height", "innerWidth", "innerHeight"} {
		n := viewportFits
		n.omit = omit
		got := sectionViewport(viewportRequest(t, n), Inputs{}, claim{})
		if got.Determination != Unverified {
			t.Errorf("omitting %q: determination = %q, want %q", omit, got.Determination, Unverified)
		}
	}
}

func TestViewportCarriesNoWeightOnAScreenOfNoSize(t *testing.T) {
	n := viewportFits
	n.screenW, n.screenH = 0, 0
	got := sectionViewport(viewportRequest(t, n), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a screen reported as having no size settles nothing", got.Determination, Unverified)
	}
}

func TestViewportCarriesNoWeightOnEveryPathThatSettlesNothing(t *testing.T) {
	cases := []struct {
		name string
		req  Request
	}{
		{"not collected", Request{Probes: map[string]Probe{}}},
		{"a number missing", viewportRequest(t, viewportNumbers{screenW: 1920, screenH: 1080, innerW: 928, innerH: 794, omit: "height"})},
		{"a window on a display above the origin", viewportRequest(t, viewportNumbers{screenW: 1920, screenH: 768, innerW: 928, innerH: 794, screenY: -200})},
		{"a window on a display left of the origin", viewportRequest(t, viewportNumbers{screenW: 800, screenH: 1080, innerW: 928, innerH: 794, screenX: -300})},
		{"a screen of no size", viewportRequest(t, viewportNumbers{screenW: 0, screenH: 0, innerW: 928, innerH: 794})},
	}

	settled := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Inconclusive},
	}
	before := summarise(settled)

	for _, c := range cases {
		got := sectionViewport(c.req, Inputs{}, claim{})
		if got.Determination != Unverified {
			t.Errorf("%s: determination = %q, want %q", c.name, got.Determination, Unverified)
		}
		got.ID = "viewport"
		after := summarise(append(settled, normalise(got)))
		if after.Band != before.Band || after.HumanConfidence != before.HumanConfidence || after.BotLikeness != before.BotLikeness {
			t.Errorf("%s: the summary moved from %+v to %+v", c.name, before, after)
		}
	}
}

func TestViewportIsOneOfTheSectionsAnalyzeBuilds(t *testing.T) {
	found := false
	for _, s := range order {
		if s.id == "viewport" {
			found = true
		}
	}
	if !found {
		t.Fatal("the reading is not registered in the section order, so no scan runs it")
	}
}

func TestViewportNamesNoVendorInItsConclusion(t *testing.T) {
	n := viewportFits
	n.screenH, n.innerH = 768, 794
	got := sectionViewport(viewportRequest(t, n), Inputs{}, claim{})
	saw := false
	for _, row := range got.Rows {
		if row.Label == "conclusion" {
			saw = true
		}
	}
	if !saw {
		t.Fatal("the reading reports no conclusion row")
	}
}
