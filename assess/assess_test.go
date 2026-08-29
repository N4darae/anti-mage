package assess

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/N4darae/anti-mage/internal/scan"
)

func ok(v string) Observation {
	return Observation{Status: StatusOK, Value: json.RawMessage(v)}
}

func unsup(reason string) Observation {
	return Observation{Status: StatusUnsupported, Value: json.RawMessage(`{"reason":` + quote(reason) + `}`)}
}

func quote(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return string(b)
}

func honest() map[string]Observation {
	const scope = `{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36",` +
		`"platform":"Win32","hardwareConcurrency":8,"language":"en-GB","languages":["en-GB","en"],` +
		`"timeZone":"Europe/London","locale":"en-GB"}`
	return map[string]Observation{
		"scope.main":   ok(scope),
		"scope.worker": ok(scope),
		"scope.iframe": ok(scope),
		"native.tostring": ok(`{"navigator.platform":"function get platform() { [native code] }",` +
			`"screen.width":"function get width() { [native code] }"}`),
		"native.ownkeys": ok(`{"navigator.platform":{"ownKeys":["length","name"],` +
			`"getOwnPropertyNames":["length","name"],"descriptors":["length","name"]}}`),
		"native.descriptor": ok(`{"navigator.platform":{"onPrototype":true}}`),
		"native.receiver":   ok(`{"navigator.platform":{"threw":true,"name":"TypeError"}}`),
		"geom.screen":       ok(`{"width":1536,"height":864,"availWidth":1536,"availHeight":816,"devicePixelRatio":1.33333}`),
		"geom.css":          ok(`{"dppx":1.333330094819539}`),
		"auto.residue":      ok(`{"webdriver":false,"driverNames":[]}`),
		"perm.state":        ok(`{"notifications":{"query":"prompt","actual":"default"}}`),

		"math.exact": ok(`{
			"abs.nan":"NaN","abs.negZero":"0","abs.negInf":"Infinity","abs.neg":"7","abs.pos":"7",
			"sign.nan":"NaN","sign.posZero":"0","sign.negZero":"-0","sign.pos":"1","sign.neg":"-1",
			"floor.nan":"NaN","floor.posInf":"Infinity","floor.negInf":"-Infinity","floor.negZero":"-0","floor.fracNeg":"-1","floor.fracPos":"2",
			"ceil.nan":"NaN","ceil.posInf":"Infinity","ceil.negInf":"-Infinity","ceil.posZero":"0","ceil.fracNeg":"-0","ceil.fracPos":"3",
			"trunc.nan":"NaN","trunc.fracNeg":"-0","trunc.fracPos":"0","trunc.negInt":"-3",
			"round.nan":"NaN","round.negZero":"-0","round.halfNeg":"-0","round.halfPos":"1","round.negHalfInt":"-2","round.posInf":"Infinity","round.negInf":"-Infinity",
			"min.nan":"NaN","min.zero":"-0","min.basic":"1","min.empty":"Infinity",
			"max.nan":"NaN","max.zero":"0","max.basic":"3","max.empty":"-Infinity",
			"fround.nan":"NaN","fround.negZero":"-0","fround.overflow":"Infinity","fround.tieToEven":"16777216","fround.exact":"0.5",
			"clz32.zero":"32","clz32.one":"31","clz32.negOne":"0","clz32.nan":"32",
			"imul.basic":"12","imul.overflow":"-5","imul.bigxbig":"1","imul.nan":"0",
			"sqrt.nan":"NaN","sqrt.negative":"NaN","sqrt.negZero":"-0","sqrt.posInf":"Infinity","sqrt.perfect":"2","sqrt.exactFraction":"2.5"
		}`),
		"math.repeat": ok(`{
			"sin":   {"a":"0.8414709848078965", "b":"0.8414709848078965"},
			"sqrt":  {"a":"1.4142135623730951", "b":"1.4142135623730951"},
			"round": {"a":"3", "b":"3"}
		}`),

		"throw.mandated": ok(`{
			"atob.invalidChars": {
				"available": true, "threw": true, "name": "InvalidCharacterError", "ctor": "DOMException",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]", "hasStack": true,
				"attempt2": {"threw": true, "name": "InvalidCharacterError", "ctor": "DOMException", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]"}
			},
			"createElement.invalidName": {
				"available": true, "threw": true, "name": "InvalidCharacterError", "ctor": "DOMException",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]", "hasStack": true,
				"attempt2": {"threw": true, "name": "InvalidCharacterError", "ctor": "DOMException", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]"}
			},
			"json.malformed": {
				"available": true, "threw": true, "name": "SyntaxError", "ctor": "SyntaxError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "SyntaxError", "ctor": "SyntaxError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"property.accessOnNull": {
				"available": true, "threw": true, "name": "TypeError", "ctor": "TypeError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "TypeError", "ctor": "TypeError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"call.nonCallable": {
				"available": true, "threw": true, "name": "TypeError", "ctor": "TypeError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "TypeError", "ctor": "TypeError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"array.negativeLength": {
				"available": true, "threw": true, "name": "RangeError", "ctor": "RangeError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "RangeError", "ctor": "RangeError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"decodeURIComponent.malformed": {
				"available": true, "threw": true, "name": "URIError", "ctor": "URIError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "URIError", "ctor": "URIError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"structuredClone.function": {
				"available": true, "threw": true, "name": "DataCloneError", "ctor": "DOMException",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]", "hasStack": true,
				"attempt2": {"threw": true, "name": "DataCloneError", "ctor": "DOMException", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]"}
			}
		}`),

		"rect.identities": ok(`{
			"shiftPx": 41,
			"base":     {"x":37,"y":53,"left":37,"top":53,"right":218,"bottom":150,"width":181,"height":97},
			"twin":     {"x":37,"y":53,"left":37,"top":53,"right":218,"bottom":150,"width":181,"height":97},
			"shifted":  {"x":78,"y":53,"left":78,"top":53,"right":259,"bottom":150,"width":181,"height":97},
			"restored": {"x":37,"y":53,"left":37,"top":53,"right":218,"bottom":150,"width":181,"height":97}
		}`),
		"text.metrics": ok(`{
			"empty": {"width": 0, "box": {
				"actualBoundingBoxLeft": 0, "actualBoundingBoxRight": 0,
				"actualBoundingBoxAscent": 0, "actualBoundingBoxDescent": 0,
				"fontBoundingBoxAscent": 29, "fontBoundingBoxDescent": 8,
				"emHeightAscent": 24, "emHeightDescent": 6,
				"hangingBaseline": 21.6, "alphabeticBaseline": 0, "ideographicBaseline": -8
			}},
			"full": {"width": 123.4, "box": {
				"actualBoundingBoxLeft": 0, "actualBoundingBoxRight": 123.4,
				"actualBoundingBoxAscent": 21, "actualBoundingBoxDescent": 5,
				"fontBoundingBoxAscent": 29, "fontBoundingBoxDescent": 8,
				"emHeightAscent": 24, "emHeightDescent": 6,
				"hangingBaseline": 21.6, "alphabeticBaseline": 0, "ideographicBaseline": -8
			}},
			"repeat": {"width": 123.4},
			"prefixWidths": [0, 10, 20, 35, 60, 90, 123.4]
		}`),

		"media.stylesheet": ok(`{
			"controlOk": true, "widthValid": true, "heightValid": true,
			"numeric": [
				{"feature":"width","op":"min","px":1536,"jsMatches":true,"cssMatches":true},
				{"feature":"width","op":"max","px":1535,"jsMatches":false,"cssMatches":false},
				{"feature":"height","op":"min","px":738,"jsMatches":true,"cssMatches":true},
				{"feature":"height","op":"max","px":737,"jsMatches":false,"cssMatches":false}
			],
			"discrete": [
				{"feature":"orientation","value":"landscape","jsMatches":true,"cssMatches":true},
				{"feature":"orientation","value":"portrait","jsMatches":false,"cssMatches":false},
				{"feature":"hover","value":"hover","jsMatches":true,"cssMatches":true},
				{"feature":"hover","value":"none","jsMatches":false,"cssMatches":false},
				{"feature":"pointer","value":"fine","jsMatches":true,"cssMatches":true},
				{"feature":"pointer","value":"coarse","jsMatches":false,"cssMatches":false},
				{"feature":"pointer","value":"none","jsMatches":false,"cssMatches":false},
				{"feature":"prefers-color-scheme","value":"light","jsMatches":true,"cssMatches":true},
				{"feature":"prefers-color-scheme","value":"dark","jsMatches":false,"cssMatches":false},
				{"feature":"prefers-reduced-motion","value":"no-preference","jsMatches":true,"cssMatches":true},
				{"feature":"prefers-reduced-motion","value":"reduce","jsMatches":false,"cssMatches":false}
			]
		}`),
		"media.complement": ok(`{
			"innerWidth": 1536, "innerHeight": 738,
			"complements": [
				{"query":"(min-width: 1px)","matches":true,"negationMatches":false},
				{"query":"(min-width: 999999px)","matches":false,"negationMatches":true}
			],
			"brackets": [
				{"feature":"width","value":1536,"insideBelowPx":1535,"insideAbovePx":1537,"minInside":true,"maxInside":true,"minOutside":false,"maxOutside":false},
				{"feature":"height","value":738,"insideBelowPx":737,"insideAbovePx":739,"minInside":true,"maxInside":true,"minOutside":false,"maxOutside":false}
			]
		}`),

		"audio.views": ok(`{
			"requested": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160},
			"rendered": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160, "durationSeconds": 0.02},
			"channelsServed": 2,
			"copyFromChannelAvailable": true,
			"views": {"compared": true, "agree": true, "sampleCount": 320, "differingSampleCount": 0, "maxAbsoluteDifference": 0}
		}`),
		"audio.repeat": ok(`{
			"secondRenderCompleted": true,
			"repeat": {"compared": true, "agree": true, "sampleCount": 320, "differingSampleCount": 0, "maxAbsoluteDifference": 0}
		}`),
	}
}

func with(changes map[string]Observation) map[string]Observation {
	obs := honest()
	for id, o := range changes {
		if o.Status == "" && o.Value == nil {
			delete(obs, id)
			continue
		}
		obs[id] = o
	}
	return obs
}

func TestNothingFoundIsCoherentAndScoresNothing(t *testing.T) {
	a := Evaluate(Environment{Observations: honest()})
	if a.Determination != Coherent {
		t.Fatalf("determination = %q, want %q; statement was %q", a.Determination, Coherent, a.Statement)
	}
	if a.Score != 0 {
		t.Errorf("score = %d, want 0; nothing disagreed, so nothing may raise the score", a.Score)
	}
	if !a.Determination.Established() {
		t.Errorf("%q must count as established: enough was read to say that nothing disagreed", a.Determination)
	}
}

func TestHonestFixtureIsCompleteForEveryNewSection(t *testing.T) {
	env := Environment{Observations: honest()}
	a := Evaluate(env)
	if a.Determination != Coherent {
		t.Fatalf("determination = %q, want %q; statement was %q", a.Determination, Coherent, a.Statement)
	}
	if a.Score != 0 {
		t.Errorf("score = %d, want 0; a browser that agrees with itself must not score", a.Score)
	}

	newIDs := []string{
		"math.exact", "math.repeat",
		"throw.mandated",
		"rect.identities", "text.metrics",
		"media.stylesheet", "media.complement",
		"audio.views", "audio.repeat",
	}
	for _, id := range newIDs {
		i := sort.SearchStrings(a.Supplied, id)
		if i >= len(a.Supplied) || a.Supplied[i] != id {
			t.Errorf("supplied = %v; missing %q, which the fixture must hand over for the five newest sections to be exercised at all", a.Supplied, id)
		}
	}

	rep := scan.AnalyzeWith(scan.Request{V: 1, Mode: "public", Probes: probesOf(env)}, scan.Inputs{}, nil)
	want := map[string]bool{"numerics": true, "throws": true, "rects": true, "mediapaths": true, "audio": true}
	seen := map[string]bool{}
	for _, sec := range rep.Sections {
		if !want[sec.ID] {
			continue
		}
		seen[sec.ID] = true
		if sec.Determination != scan.Consistent {
			t.Errorf("section %q = %q on the honest fixture, want %q; the fixture no longer supplies what this section reads", sec.ID, sec.Determination, scan.Consistent)
		}
	}
	for id := range want {
		if !seen[id] {
			t.Errorf("section %q was not found in the report at all", id)
		}
	}
}

func TestOneContradictionIsDiscrepantAndClearsTheFloor(t *testing.T) {

	a := Evaluate(Environment{Observations: with(map[string]Observation{
		"geom.css": ok(`{"dppx":2}`),
	})})
	if a.Determination != Discrepant {
		t.Fatalf("determination = %q, want %q; statement was %q", a.Determination, Discrepant, a.Statement)
	}
	if a.Score < 30 {
		t.Errorf("score = %d; one finding among many must not quantise down to nothing found", a.Score)
	}
	if a.Score > 100 {
		t.Errorf("score = %d, above the scale", a.Score)
	}
}

func TestOneModifiedAccessorIsInstrumentedAndScoresBelowAContradiction(t *testing.T) {
	blocked := Evaluate(Environment{Observations: with(map[string]Observation{
		"native.tostring": ok(`{"canvas.toDataURL":"function () {\n      return orig.apply(this, arguments);\n    }"}`),
		"native.ownkeys": ok(`{"canvas.toDataURL":{"ownKeys":["length","name","prototype"],` +
			`"getOwnPropertyNames":["length","name","prototype"],"descriptors":["length","name","prototype"]}}`),
		"native.descriptor": ok(`{"navigator.platform":{"onPrototype":false}}`),
		"native.receiver":   {},
	})})
	caught := Evaluate(Environment{Observations: with(map[string]Observation{
		"geom.css": ok(`{"dppx":2}`),
	})})

	if blocked.Determination != Instrumented {
		t.Fatalf("determination = %q, want %q; statement was %q", blocked.Determination, Instrumented, blocked.Statement)
	}
	if blocked.Score < 10 {
		t.Errorf("score = %d; a single modification must not read as nothing found", blocked.Score)
	}
	if blocked.Score >= caught.Score {
		t.Errorf("score: modified accessor %d, contradiction %d; the honest population that modifies a browser surface is large and the one that contradicts itself is not, so the first must score strictly lower",
			blocked.Score, caught.Score)
	}
	if !strings.Contains(strings.ToLower(blocked.Statement), "modif") {
		t.Errorf("statement = %q; the top determination has to say what it means", blocked.Statement)
	}
}

func TestMostlyUnsupportedIsUncertainInBothDirections(t *testing.T) {
	obs := map[string]Observation{}
	for id := range honest() {
		obs[id] = unsup("this browser has no such interface")
	}
	for _, id := range []string{"time.zone", "time.offsets", "media.matrix", "font.resolved", "font.controls", "scope.availability"} {
		obs[id] = unsup("this browser has no such interface")
	}
	a := Evaluate(Environment{Observations: obs})

	if a.Score != 0 {
		t.Errorf("score = %d; an observation the browser could not make must never be scored against it", a.Score)
	}
	if a.Determination == Coherent {
		t.Errorf("determination = %q; too little was established to report this environment as fine", a.Determination)
	}
	if a.Determination.Established() {
		t.Errorf("determination = %q reports itself established on a browser that answered almost nothing", a.Determination)
	}
	if a.Determination.AtLeast(Discrepant) {
		t.Errorf("determination = %q; absence is not evidence and must not reach a flagged reading", a.Determination)
	}
}

func TestNoEvidenceAtAllReachesNoReading(t *testing.T) {
	a := Evaluate(Environment{})
	if a.Determination != NotEvaluated {
		t.Errorf("determination = %q, want %q", a.Determination, NotEvaluated)
	}
	if a.Score != 0 {
		t.Errorf("score = %d, want 0", a.Score)
	}
	if a.Statement == "" {
		t.Errorf("an assessment with no reading still has to say so")
	}
	if a.Supplied == nil {
		t.Errorf("supplied must be an empty list rather than a null, so the wire form is an array")
	}
	if a.Determination.Established() {
		t.Errorf("%q cannot be established", a.Determination)
	}
}

func TestNoStatusOtherThanOKIsEverEvidence(t *testing.T) {

	for _, status := range []string{StatusUnsupported, StatusError, "", "weird", "OK "} {
		obs := with(map[string]Observation{
			"geom.screen": {Status: status, Value: json.RawMessage(`{"width":800,"height":600,"availWidth":1200,"availHeight":600}`)},
			"geom.css":    {Status: status, Value: json.RawMessage(`{"dppx":2}`)},
		})
		a := Evaluate(Environment{Observations: obs})
		if status == "OK " {

			if !a.Determination.AtLeast(Discrepant) {
				t.Errorf("status %q: determination %q; the three statuses are accepted with surrounding space", status, a.Determination)
			}
			continue
		}
		if a.Determination.AtLeast(Discrepant) {
			t.Errorf("status %q produced %q; only an observation reported as observed is evidence", status, a.Determination)
		}
		if a.Score != 0 {
			t.Errorf("status %q scored %d, want 0", status, a.Score)
		}
	}
}

func TestRemovingObservationsNeverRaisesTheScore(t *testing.T) {
	full := honest()
	ids := make([]string, 0, len(full))
	for id := range full {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	base := Evaluate(Environment{Observations: full})
	for i := range ids {
		obs := honest()
		for _, id := range ids[:i+1] {
			delete(obs, id)
		}
		a := Evaluate(Environment{Observations: obs})
		if a.Score > base.Score {
			t.Errorf("removing %v raised the score from %d to %d", ids[:i+1], base.Score, a.Score)
		}
	}
}

func TestScoreIsCoarse(t *testing.T) {
	for _, env := range []Environment{
		{Observations: honest()},
		{Observations: with(map[string]Observation{"geom.css": ok(`{"dppx":2}`)})},
		{Observations: with(map[string]Observation{"scope.worker": ok(`{"userAgent":"other","platform":"Linux x86_64","hardwareConcurrency":1}`)})},
		{Findings: []Finding{{Name: "a", Verdict: Contradiction}, {Name: "b", Verdict: Modified}}},
	} {
		a := Evaluate(env)
		if a.Score%10 != 0 {
			t.Errorf("score = %d; a score with more precision than a step of ten would let one reading be separated from another by watching the last digit", a.Score)
		}
		if a.Score < 0 || a.Score > 100 {
			t.Errorf("score = %d, off the scale", a.Score)
		}
	}
}

func TestDeterminationsAreOrderedAndTotal(t *testing.T) {
	scale := []Determination{NotEvaluated, Insufficient, Coherent, Discrepant, Instrumented}
	for i, lower := range scale {
		for j, higher := range scale {
			if got, want := higher.AtLeast(lower), j >= i; got != want {
				t.Errorf("%q.AtLeast(%q) = %v, want %v", higher, lower, got, want)
			}
		}
	}

	unknown := Determination("something else")
	if unknown.AtLeast(Insufficient) {
		t.Errorf("an unrecognised determination satisfied a threshold")
	}
	if unknown.Established() {
		t.Errorf("an unrecognised determination claimed to be established")
	}
	if unknown.String() != "something else" {
		t.Errorf("String() = %q", unknown.String())
	}
}

func TestTheAssessmentCarriesNoBreakdown(t *testing.T) {
	a := Evaluate(Environment{
		Observations: with(map[string]Observation{"geom.css": ok(`{"dppx":2}`)}),
		Findings:     []Finding{{Name: "a reading of my own", Verdict: Contradiction}},
	})
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"v": true, "determination": true, "score": true, "statement": true, "supplied": true}
	for k := range out {
		if !want[k] {
			t.Errorf("the assessment carries %q; a caller must not be able to see which reading moved the score", k)
		}
	}
	for k := range want {
		if _, present := out[k]; !present {
			t.Errorf("the assessment carries no %q", k)
		}
	}

	for _, leak := range []string{"geom", "dppx", "a reading of my own", "natives", "fonts", "section"} {
		if strings.Contains(strings.ToLower(a.Statement), leak) {
			t.Errorf("statement %q names %q; it must say what the determination means and not which reading produced it", a.Statement, leak)
		}
	}
}

func TestSuppliedIsAnEchoAndNothingMore(t *testing.T) {
	obs := map[string]Observation{
		"perm.state":            ok(`{"notifications":{"query":"prompt","actual":"default"}}`),
		"auto.residue":          unsup("no such member"),
		"an.id.nothing.reads":   ok(`{"anything":1}`),
		"":                      ok(`{"dropped":1}`),
		"zzz.last.alphabetical": ok(`1`),
	}
	a := Evaluate(Environment{Observations: obs})
	want := []string{"an.id.nothing.reads", "auto.residue", "perm.state", "zzz.last.alphabetical"}
	if !reflect.DeepEqual(a.Supplied, want) {
		t.Errorf("supplied = %v, want %v: every id the caller handed over, sorted, whether or not anything read it", a.Supplied, want)
	}
}

func TestTheStatementNamesNoProductAndAssertsNoPerson(t *testing.T) {
	envs := []Environment{
		{},
		{Observations: honest()},
		{Observations: with(map[string]Observation{"geom.css": ok(`{"dppx":2}`)})},
		{Observations: with(map[string]Observation{"auto.residue": ok(`{"webdriver":true}`)})},
		{Findings: []Finding{{Name: "mine", Verdict: Modified}}},
	}

	banned := regexp.MustCompile(`(?i)\b(bot|bots|human|humans|robot|crawler|scraper|fraud|fake|spoofed)\b`)
	for _, env := range envs {
		a := Evaluate(env)
		if m := banned.FindString(a.Statement); m != "" {
			t.Errorf("statement %q contains %q; the ceiling is that an environment appears modified", a.Statement, m)
		}
	}
}

func TestASuppliedFindingIsFoldedIntoTheOneScore(t *testing.T) {
	clean := Evaluate(Environment{Observations: honest()})
	if clean.Determination != Coherent || clean.Score != 0 {
		t.Fatalf("the baseline moved: %+v", clean)
	}
	withOne := Evaluate(Environment{
		Observations: honest(),
		Findings:     []Finding{{Name: "a reading of my own", Verdict: Contradiction}},
	})
	if !withOne.Determination.AtLeast(Discrepant) {
		t.Errorf("determination = %q; a contradiction the caller found must reach the same reading as one this library found", withOne.Determination)
	}
	if withOne.Score <= clean.Score {
		t.Errorf("score went from %d to %d; evidence that disagrees must raise it", clean.Score, withOne.Score)
	}

	agreeing := Evaluate(Environment{
		Observations: honest(),
		Findings:     []Finding{{Name: "a reading of my own", Verdict: Consistent}},
	})
	if agreeing.Score != 0 || !agreeing.Determination.Established() {
		t.Errorf("a finding that agrees produced %+v; agreement cannot raise the score", agreeing)
	}
}

func TestFindingsAreIdentifiedByNameAndOrderFree(t *testing.T) {
	once := Evaluate(Environment{
		Observations: honest(),
		Findings:     []Finding{{Name: "mine", Verdict: Contradiction}},
	})
	repeated := Evaluate(Environment{
		Observations: honest(),
		Findings: []Finding{
			{Name: "mine", Verdict: Contradiction},
			{Name: " mine ", Verdict: Contradiction},
			{Name: "mine", Verdict: Contradiction},
		},
	})
	if !reflect.DeepEqual(once, repeated) {
		t.Errorf("naming one finding three times changed the assessment from %+v to %+v", once, repeated)
	}

	forward := Evaluate(Environment{Findings: []Finding{
		{Name: "a", Verdict: Contradiction}, {Name: "b", Verdict: Consistent}, {Name: "c", Verdict: Modified},
	}})
	backward := Evaluate(Environment{Findings: []Finding{
		{Name: "c", Verdict: Modified}, {Name: "b", Verdict: Consistent}, {Name: "a", Verdict: Contradiction},
	}})
	if !reflect.DeepEqual(forward, backward) {
		t.Errorf("the order of findings changed the assessment: %+v then %+v", forward, backward)
	}
}

func TestAFindingWithNoNameOrAnUnknownVerdictIsNotEvidence(t *testing.T) {
	base := Evaluate(Environment{Observations: honest()})
	for _, f := range []Finding{
		{Name: "", Verdict: Contradiction},
		{Name: "   ", Verdict: Modified},
		{Name: "mine", Verdict: Verdict("very bad")},
		{Name: "mine", Verdict: ""},
	} {
		a := Evaluate(Environment{Observations: honest(), Findings: []Finding{f}})
		if a.Score > base.Score {
			t.Errorf("finding %+v raised the score from %d to %d; an unnamed or unrecognised reading is not evidence", f, base.Score, a.Score)
		}
		if a.Determination.AtLeast(Discrepant) {
			t.Errorf("finding %+v reached %q", f, a.Determination)
		}
	}
}

func TestAnUnverifiedFindingIsLeftOutOfTheDenominator(t *testing.T) {
	base := Evaluate(Environment{Observations: honest()})
	unverified := Evaluate(Environment{
		Observations: honest(),
		Findings:     []Finding{{Name: "a table nobody has confirmed", Verdict: Unverified}},
	})
	if !reflect.DeepEqual(base, unverified) {
		t.Errorf("an unverified finding changed the assessment from %+v to %+v; a gap in reference data is a fact about the data, not the browser",
			base, unverified)
	}
}

func TestAnInconclusiveFindingLowersConfidenceRatherThanRaisingTheScore(t *testing.T) {
	base := Evaluate(Environment{Observations: honest()})
	many := make([]Finding, 0, 20)
	for i := 0; i < 20; i++ {
		many = append(many, Finding{Name: string(rune('a'+i)) + " reading", Verdict: Inconclusive})
	}
	a := Evaluate(Environment{Observations: honest(), Findings: many})
	if a.Score > base.Score {
		t.Errorf("twenty readings that concluded nothing raised the score from %d to %d", base.Score, a.Score)
	}
	if a.Determination.Established() {
		t.Errorf("determination = %q; once most of what was looked at concluded nothing, the assessment must stop claiming to be established", a.Determination)
	}
}

func TestEvaluateIsAFunctionOfItsArgument(t *testing.T) {
	env := Environment{
		Observations: with(map[string]Observation{"geom.css": ok(`{"dppx":2}`)}),
		Findings:     []Finding{{Name: "b", Verdict: Modified}, {Name: "a", Verdict: Consistent}},
		Nonce:        "abc",
		OffsetDates:  []string{"2026-01-12", "2026-07-04"},
		FontControls: []string{"Zzqx 0000 Absent"},
		ElapsedMS:    412,
	}
	first := Evaluate(env)
	for i := 0; i < 50; i++ {
		if got := Evaluate(env); !reflect.DeepEqual(got, first) {
			t.Fatalf("run %d returned %+v, first returned %+v", i, got, first)
		}
	}
}

func TestANegativeElapsedIsNotAMeasurement(t *testing.T) {
	obs := with(map[string]Observation{
		"font.resolved":   unsup("timed out after 9000 ms"),
		"native.tostring": ok(`{"CanvasRenderingContext2D.measureText":"function measureText() { [native code] }"}`),
	})

	priced := Evaluate(Environment{Observations: obs, ElapsedMS: 400})
	if !priced.Determination.AtLeast(Discrepant) {
		t.Fatalf("determination = %q; a claimed wait longer than the caller's own measurement of the scan is a contradiction", priced.Determination)
	}
	for _, ms := range []int{0, -1, -1000} {
		a := Evaluate(Environment{Observations: obs, ElapsedMS: ms})
		if a.Determination.AtLeast(Discrepant) {
			t.Errorf("elapsed %d ms produced %q; with no measurement there is nothing to price a claim against", ms, a.Determination)
		}
	}
}

func TestDecodeReadsTheObservationsUnderEitherSpelling(t *testing.T) {
	for _, field := range []string{"probes", "observations"} {
		body := `{"v":1,"nonce":"n1","` + field + `":{"perm.state":{"status":"ok","value":{"notifications":{"query":"prompt","actual":"default"}}}}}`
		env, err := Decode([]byte(body))
		if err != nil {
			t.Fatalf("%s: %v", field, err)
		}
		if env.Nonce != "n1" {
			t.Errorf("%s: nonce = %q", field, env.Nonce)
		}
		if len(env.Observations) != 1 {
			t.Fatalf("%s: got %d observations", field, len(env.Observations))
		}
		if got := env.Observations["perm.state"].Status; got != StatusOK {
			t.Errorf("%s: status = %q", field, got)
		}
	}
}

func TestDecodeTakesNothingFromThePayloadThatTheCallerMustChoose(t *testing.T) {
	body := `{"v":1,"probes":{"perm.state":{"status":"ok","value":1}},
	  "elapsedMs":999999,"elapsed":999999,"ElapsedMS":999999,
	  "offsetDates":["2020-01-01"],"dates":["2020-01-01"],
	  "fontControls":["Zzqx 0000 Absent"],"controls":["Zzqx 0000 Absent"],
	  "findings":[{"name":"trust me","verdict":"consistent"}],
	  "observations":{"auto.residue":{"status":"ok","value":{"webdriver":false}}}}`
	env, err := Decode([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if env.ElapsedMS != 0 {
		t.Errorf("elapsed = %d; a duration the examined environment supplied is not a measurement", env.ElapsedMS)
	}
	if env.OffsetDates != nil || env.FontControls != nil {
		t.Errorf("issued inputs came out of the payload: %v %v; the point of choosing them per scan is that the environment did not", env.OffsetDates, env.FontControls)
	}
	if env.Findings != nil {
		t.Errorf("findings came out of the payload: %v; an environment must not hand itself a reading", env.Findings)
	}
	if len(env.Observations) != 2 {
		t.Errorf("got %d observations, want the two spellings read as one field", len(env.Observations))
	}
}

func TestDecodeRejectsOnlyWhatIsNotAJSONObject(t *testing.T) {
	for _, body := range []string{`[]`, `[1,2,3]`, `"a string"`, `42`, `true`} {
		if _, err := Decode([]byte(body)); err != ErrNotAnObject {
			t.Errorf("Decode(%s) error = %v, want ErrNotAnObject", body, err)
		}
	}
	for _, body := range []string{``, `{`, `{"probes":`, `not json`, `{"probes":{"a":}}`, "\x00\x01"} {
		if _, err := Decode([]byte(body)); err == nil {
			t.Errorf("Decode(%q) returned no error", body)
		}
	}

	env, err := Decode([]byte(`null`))
	if err != nil {
		t.Errorf("Decode(null): %v", err)
	}
	if len(env.Observations) != 0 || env.Nonce != "" {
		t.Errorf("Decode(null) = %+v, want nothing", env)
	}
}

func TestDecodeKeepsWhatItCanOfAGarbledPayload(t *testing.T) {

	env, err := Decode([]byte(`{"nonce":42,"probes":{"good":{"status":"ok","value":1},"bad":7,"":{"status":"ok"}},"observations":"not an object"}`))
	if err != nil {
		t.Fatalf("a garbled payload must decode to what can be read of it: %v", err)
	}
	if env.Nonce != "" {
		t.Errorf("nonce = %q; a nonce of the wrong type is no nonce", env.Nonce)
	}
	if len(env.Observations) != 1 {
		t.Fatalf("got %d observations, want the one that was readable: %+v", len(env.Observations), env.Observations)
	}
	if _, present := env.Observations["good"]; !present {
		t.Errorf("the readable observation was dropped: %+v", env.Observations)
	}
}

func TestNamesReportedReadsNoConclusion(t *testing.T) {
	env := Environment{Observations: map[string]Observation{
		"keyed":       ok(`{"Zzqx 0000 Absent":false,"Zzqx 0001 Absent":true}`),
		"listed":      ok(`["one","two"]`),
		"unsupported": unsup("no canvas"),
		"garbage":     {Status: StatusOK, Value: json.RawMessage(`{"a":`)},
	}}
	got := env.NamesReported("keyed")
	sort.Strings(got)
	if !reflect.DeepEqual(got, []string{"Zzqx 0000 Absent", "Zzqx 0001 Absent"}) {
		t.Errorf("keyed = %v", got)
	}
	if got := env.NamesReported("listed"); len(got) != 2 {
		t.Errorf("listed = %v", got)
	}
	for _, id := range []string{"unsupported", "garbage", "absent"} {
		if got := env.NamesReported(id); len(got) != 0 {
			t.Errorf("%s = %v, want nothing", id, got)
		}
	}
}

func deepJSON(n int) string {
	return strings.Repeat(`{"a":`, n) + `1` + strings.Repeat(`}`, n)
}

func TestHostileInputNeverPanics(t *testing.T) {
	values := []string{
		`null`, `true`, `false`, `0`, `-1`, `1e999`, `-1e999`, `1e-999`,
		`""`, `" "`, `[]`, `{}`, `[[[[[[[[[[]]]]]]]]]]`,
		`{"width":"not a number","height":[],"availWidth":{}}`,
		`{"userAgent":[],"platform":{},"userAgentData":"not an object"}`,
		`{"notifications":{"query":[],"actual":{}}}`,
		`{"webdriver":"yes","driverNames":{"a":{"found":"maybe"}}}`,
		`{"reason":` + quote(strings.Repeat("x", 5000)) + `}`,
		`{"reason":"timed out after 99999999999999999999 ms"}`,
		deepJSON(2000),
		`[` + strings.Repeat(`"n",`, 5000) + `"n"]`,
	}
	ids := []string{
		"scope.main", "scope.worker", "scope.iframe", "scope.availability",
		"font.resolved", "font.controls", "font.coverage",
		"native.tostring", "native.ownkeys", "native.descriptor", "native.receiver",
		"geom.screen", "geom.css", "time.zone", "time.offsets",
		"auto.residue", "perm.state", "media.matrix",
		"", "  ", strings.Repeat("q", 4096), "\x00\x01",
	}
	statuses := []string{StatusOK, StatusUnsupported, StatusError, "", "OK", "unknown"}

	for _, v := range values {
		for _, st := range statuses {
			obs := map[string]Observation{}
			for _, id := range ids {
				obs[id] = Observation{Status: st, Value: json.RawMessage(v)}
			}
			env := Environment{
				Observations: obs,
				Findings: []Finding{
					{Name: "", Verdict: Verdict(v)},
					{Name: strings.Repeat("z", 4096), Verdict: Modified},
				},
				Nonce:        v,
				OffsetDates:  []string{"", v, "not-a-date"},
				FontControls: []string{"", v},
				ElapsedMS:    -1 << 62,
			}
			a := Evaluate(env)
			if a.Score < 0 || a.Score > 100 {
				t.Fatalf("value %.40s status %q: score %d off the scale", v, st, a.Score)
			}
			if a.Statement == "" {
				t.Fatalf("value %.40s status %q: no statement", v, st)
			}
		}
	}
}

func TestTruncatedAndRandomBytesNeverPanic(t *testing.T) {
	full := `{"v":1,"nonce":"n","probes":{"scope.main":{"status":"ok","value":{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)","platform":"Win32"}},` +
		`"geom.screen":{"status":"ok","value":{"width":800,"height":600,"availWidth":1200,"availHeight":600}}}}`
	for i := 0; i <= len(full); i++ {
		env, err := Decode([]byte(full[:i]))
		if err != nil {
			continue
		}
		a := Evaluate(env)
		if a.Score < 0 || a.Score > 100 {
			t.Fatalf("prefix of length %d scored %d", i, a.Score)
		}
	}

	r := rand.New(rand.NewSource(20260826))
	buf := make([]byte, 512)
	for i := 0; i < 2000; i++ {
		n := r.Intn(len(buf))
		for j := 0; j < n; j++ {
			buf[j] = byte(r.Intn(256))
		}
		env, err := Decode(buf[:n])
		if err != nil {
			continue
		}
		if a := Evaluate(env); a.Score < 0 || a.Score > 100 {
			t.Fatalf("random bytes scored %d", a.Score)
		}
	}
}

func TestAnUnreadableObservationValueIsNotEvidence(t *testing.T) {

	obs := with(map[string]Observation{
		"geom.screen": {Status: StatusOK, Value: json.RawMessage(`{"width":800,`)},
		"geom.css":    {Status: StatusOK, Value: nil},
	})
	a := Evaluate(Environment{Observations: obs})
	if a.Determination.AtLeast(Discrepant) {
		t.Errorf("determination = %q; a value that does not parse says nothing", a.Determination)
	}
	if a.Score != 0 {
		t.Errorf("score = %d, want 0", a.Score)
	}
}
