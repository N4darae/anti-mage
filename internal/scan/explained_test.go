package scan

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

var measuredNativeTargetKeys = []string{
	"AudioBuffer.prototype.getChannelData",
	"CanvasRenderingContext2D.prototype.measureText",
	"Date.prototype.getTimezoneOffset",
	"Function.prototype.toString",
	"HTMLCanvasElement.prototype.toDataURL",
	"HTMLMediaElement.prototype.canPlayType",
	"Intl.DateTimeFormat.prototype.resolvedOptions",
	"Performance.prototype.now",
	"WebGLRenderingContext.prototype.getParameter",
	"navigator.deviceMemory",
	"navigator.hardwareConcurrency",
	"navigator.language",
	"navigator.languages",
	"navigator.platform",
	"navigator.plugins",
	"navigator.userAgent",
	"navigator.webdriver",
	"screen.availHeight",
	"screen.availWidth",
	"screen.colorDepth",
	"screen.height",
	"screen.pixelDepth",
	"screen.width",
	"window.matchMedia",
}

func nativesSaying(modified, intact []string) map[string]string {
	entries := map[string]string{}
	for _, k := range intact {
		entries[k] = `"function ` + propertyName(k) + `() { [native code] }"`
	}
	for _, k := range modified {
		entries[k] = `"function (q) { return this; }"`
	}
	parts := make([]string, 0, len(entries))
	for _, k := range sortedStrings(entries) {
		parts = append(parts, `"`+k+`":`+entries[k])
	}
	return map[string]string{"native.tostring": ok("{" + strings.Join(parts, ",") + "}")}
}

func sortedStrings(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func with(base map[string]string, extra map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func run(t *testing.T, kv map[string]string, build sectionFunc) Section {
	t.Helper()
	r := probes(t, kv)
	return build(r, Inputs{}, readClaim(r))
}

func TestMappedAccessorKeysAreKeysTheCollectorReports(t *testing.T) {
	reported := map[string]bool{}
	for _, k := range measuredNativeTargetKeys {
		reported[k] = true
	}
	for _, k := range mappedAccessorKeys {
		if !reported[k] {
			t.Errorf("explained.go maps a requirement onto %q, which no measured payload reported on; a key this package invents can never be reported modified and so downgrades nothing", k)
		}
	}
}

func TestPrePassAndNativesSectionAgreeOnWhatModifiedMeans(t *testing.T) {
	cases := []struct {
		name  string
		kv    map[string]string
		wantN int
	}{
		{"nothing reported", map[string]string{}, 0},
		{"all intact", nativesSaying(nil, measuredNativeTargetKeys), 0},
		{"one modified", nativesSaying([]string{keyMeasureText}, measuredNativeTargetKeys), 1},
		{"three modified", nativesSaying([]string{keyMeasureText, keyGetChannelData, keyMatchMedia}, measuredNativeTargetKeys), 3},
		{"unsupported probe", map[string]string{"native.tostring": unsup("no accessor was reachable")}, 0},
		{"unparseable probe", map[string]string{"native.tostring": ok(`"not an object"`)}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := probes(t, tc.kv)
			c := readClaim(r)
			sec := sectionNatives(r, Inputs{}, c)

			if len(c.natives.modified) != tc.wantN {
				t.Fatalf("the shared reading calls %d accessors modified, want %d", len(c.natives.modified), tc.wantN)
			}
			rowed := map[string]bool{}
			for _, row := range sec.Rows {
				if row.Value == "does not meet a requirement on a built-in" {
					rowed[row.Label] = true
				}
			}
			for k := range c.natives.modified {
				if !rowed[k] {
					t.Errorf("the shared reading calls %q modified but the natives section shows no row for it", k)
				}
			}
			for k := range rowed {
				if !c.natives.reportsModified(k) {
					t.Errorf("the natives section shows a row for %q but the shared reading does not call it modified", k)
				}
			}
			if (tc.wantN > 0) != (sec.Determination == Instrumented) {
				t.Errorf("determination = %q with %d modified accessors; the two must move together", sec.Determination, tc.wantN)
			}
		})
	}
}

func TestAbsentNativesReadingDowngradesNothing(t *testing.T) {
	natives := map[string]map[string]string{
		"no natives probe at all": {},
		"unsupported":             {"native.tostring": unsup("no accessor was reachable")},
		"error status":            {"native.tostring": `{"status":"error","value":{"reason":"threw"}}`},
		"unparseable value":       {"native.tostring": ok(`"not an object"`)},
		"empty object":            {"native.tostring": ok(`{}`)},
	}
	for name, nat := range natives {
		t.Run(name, func(t *testing.T) {
			sec := run(t, with(textMetricsWidthDisagrees(), nat), sectionRects)
			if sec.Determination != Contradiction {
				t.Fatalf("determination = %q, want contradiction: nothing in this payload says any accessor is not a built-in, so nothing explains the disagreement; rows: %+v", sec.Determination, sec.Rows)
			}
		})
	}
}

func TestAnotherAccessorReportedModifiedDowngradesNothing(t *testing.T) {
	others := []string{
		"navigator.platform", "screen.width", "Function.prototype.toString",
		"AudioBuffer.prototype.getChannelData", "window.matchMedia",
	}
	for _, other := range others {
		t.Run(other, func(t *testing.T) {
			sec := run(t, with(textMetricsWidthDisagrees(), nativesSaying([]string{other}, measuredNativeTargetKeys)), sectionRects)
			if sec.Determination != Contradiction {
				t.Fatalf("determination = %q, want contradiction: a modified %s explains nothing about a text metric; rows: %+v", sec.Determination, other, sec.Rows)
			}
		})
	}
}

func textMetricsWidthDisagrees() map[string]string {
	return map[string]string{
		"rect.identities": ok(consistentRectPayload),
		"text.metrics":    ok(strings.Replace(consistentTextPayload, `"empty": {"width": 0,`, `"empty": {"width": 0.004,`, 1)),
	}
}

func rectIdentityDisagrees() map[string]string {
	return map[string]string{
		"rect.identities": ok(strings.Replace(consistentRectPayload, `"right":218,"bottom":150,"width":181,"height":97},
	"twin"`, `"right":218,"bottom":150,"width":175,"height":97},
	"twin"`, 1)),
		"text.metrics": ok(consistentTextPayload),
	}
}

func TestRectsTextMetricDisagreementIsDowngradedWhenMeasureTextIsReportedModified(t *testing.T) {
	kv := with(textMetricsWidthDisagrees(), nativesSaying([]string{keyMeasureText}, measuredNativeTargetKeys))
	sec := run(t, kv, sectionRects)
	if sec.Determination != Instrumented {
		t.Fatalf("determination = %q, want instrumented: the payload reports measureText as not a built-in and every text metric read arrived through it; rows: %+v", sec.Determination, sec.Rows)
	}
	if !rowsMentionTheExplanation(sec, keyMeasureText) {
		t.Errorf("no row says which accessor explains the disagreement; rows: %+v", sec.Rows)
	}
}

func TestRectsUnexplainedDisagreementAlongsideAnExplainedOneStillConvicts(t *testing.T) {
	both := map[string]string{
		"rect.identities": rectIdentityDisagrees()["rect.identities"],
		"text.metrics":    textMetricsWidthDisagrees()["text.metrics"],
	}
	sec := run(t, with(both, nativesSaying([]string{keyMeasureText}, measuredNativeTargetKeys)), sectionRects)
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: the rect identity that failed reads no accessor the environment reports as modified; rows: %+v", sec.Determination, sec.Rows)
	}

	if !rowsMentionTheExplanation(sec, keyMeasureText) {
		t.Errorf("the explained disagreement disappeared from the rows; rows: %+v", sec.Rows)
	}
}

func TestRectsRectIdentityConvictsEvenWithMeasureTextReportedModified(t *testing.T) {
	sec := run(t, with(rectIdentityDisagrees(), nativesSaying([]string{keyMeasureText}, measuredNativeTargetKeys)), sectionRects)
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction; rows: %+v", sec.Determination, sec.Rows)
	}
}

func TestRectsCleanPayloadIsConsistentWithAModifiedAccessorReported(t *testing.T) {
	clean := map[string]string{
		"rect.identities": ok(consistentRectPayload),
		"text.metrics":    ok(consistentTextPayload),
	}
	sec := run(t, with(clean, nativesSaying([]string{keyMeasureText}, measuredNativeTargetKeys)), sectionRects)
	if sec.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent: nothing disagreed; rows: %+v", sec.Determination, sec.Rows)
	}
}

func audioPayload(views, duration string) map[string]string {
	return map[string]string{
		"audio.views": ok(`{
			"requested": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160},
			"rendered": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160, "durationSeconds": ` + duration + `},
			"channelsServed": 2,
			"copyFromChannelAvailable": true,
			"views": ` + views + `
		}`),
	}
}

const audioViewsAgree = `{"compared": true, "agree": true, "sampleCount": 320}`
const audioViewsDisagree = `{"compared": true, "agree": false, "sampleCount": 320, "differingSampleCount": 2, "maxAbsoluteDifference": 0.001}`

func TestAudioViewsDisagreementIsDowngradedWhenGetChannelDataIsReportedModified(t *testing.T) {
	kv := with(audioPayload(audioViewsDisagree, "0.02"), nativesSaying([]string{keyGetChannelData}, measuredNativeTargetKeys))
	sec := run(t, kv, sectionAudioBuf)
	if sec.Determination != Instrumented {
		t.Fatalf("determination = %q, want instrumented; rows: %+v", sec.Determination, sec.Rows)
	}
	if !rowsMentionTheExplanation(sec, keyGetChannelData) {
		t.Errorf("no row names the accessor that explains the disagreement; rows: %+v", sec.Rows)
	}
}

func TestAudioDurationArithmeticConvictsEvenWithGetChannelDataReportedModified(t *testing.T) {
	kv := with(audioPayload(audioViewsAgree, "0.5"), nativesSaying([]string{keyGetChannelData}, measuredNativeTargetKeys))
	sec := run(t, kv, sectionAudioBuf)
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: duration, length and sampleRate are the buffer's own attributes and do not pass through getChannelData; rows: %+v", sec.Determination, sec.Rows)
	}
}

func TestAudioExplainedViewsPlusUnexplainedArithmeticStillConvicts(t *testing.T) {
	kv := with(audioPayload(audioViewsDisagree, "0.5"), nativesSaying([]string{keyGetChannelData}, measuredNativeTargetKeys))
	sec := run(t, kv, sectionAudioBuf)
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction; rows: %+v", sec.Determination, sec.Rows)
	}
}

func TestAudioViewsDisagreementConvictsWithNoNativesReading(t *testing.T) {
	sec := run(t, audioPayload(audioViewsDisagree, "0.02"), sectionAudioBuf)
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction; rows: %+v", sec.Determination, sec.Rows)
	}
}

func mediaPathsColourSchemeDisagrees() map[string]string {
	return map[string]string{
		"media.stylesheet": ok(`{
			"controlOk": true, "widthValid": true, "heightValid": true,
			"numeric": [],
			"discrete": [
				{"feature":"prefers-color-scheme","value":"dark","jsMatches":true,"cssMatches":false},
				{"feature":"prefers-color-scheme","value":"light","jsMatches":false,"cssMatches":true}
			]
		}`),
	}
}

func TestMediaPathsDisagreementIsDowngradedWhenMatchMediaIsReportedModified(t *testing.T) {
	kv := with(mediaPathsColourSchemeDisagrees(), nativesSaying([]string{keyMatchMedia}, measuredNativeTargetKeys))
	sec := run(t, kv, sectionMediaPaths)
	if sec.Determination != Instrumented {
		t.Fatalf("determination = %q, want instrumented; rows: %+v", sec.Determination, sec.Rows)
	}
	if !rowsMentionTheExplanation(sec, keyMatchMedia) {
		t.Errorf("no row names the accessor that explains the disagreement; rows: %+v", sec.Rows)
	}
}

func TestMediaPathsDisagreementConvictsWithNoNativesReading(t *testing.T) {
	sec := run(t, mediaPathsColourSchemeDisagrees(), sectionMediaPaths)
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction; rows: %+v", sec.Determination, sec.Rows)
	}
}

func TestMediaPathsNegationDisagreementIsDowngradedWhenMatchMediaIsReportedModified(t *testing.T) {
	kv := map[string]string{
		"media.complement": ok(`{"complements":[{"query":"(min-width: 1px)","matches":true,"negationMatches":true}]}`),
	}
	sec := run(t, with(kv, nativesSaying([]string{keyMatchMedia}, measuredNativeTargetKeys)), sectionMediaPaths)
	if sec.Determination != Instrumented {
		t.Fatalf("determination = %q, want instrumented; rows: %+v", sec.Determination, sec.Rows)
	}
	plain := run(t, kv, sectionMediaPaths)
	if plain.Determination != Contradiction {
		t.Fatalf("determination = %q with no natives reading, want contradiction", plain.Determination)
	}
}

func TestGeometryAvailableSpaceIsDowngradedWhenTheScreenGettersAreReportedModified(t *testing.T) {
	kv := map[string]string{
		"geom.screen": ok(`{"width":1536,"height":864,"availWidth":1600,"availHeight":816,"devicePixelRatio":1}`),
		"geom.css":    ok(`{"dppx":1}`),
	}
	sec := run(t, with(kv, nativesSaying([]string{keyScreenAvailWidth}, measuredNativeTargetKeys)), sectionGeometry)
	if sec.Determination != Instrumented {
		t.Fatalf("determination = %q, want instrumented: both sides of that comparison are read through the screen attribute getters; rows: %+v", sec.Determination, sec.Rows)
	}
	plain := run(t, kv, sectionGeometry)
	if plain.Determination != Contradiction {
		t.Fatalf("determination = %q with no natives reading, want contradiction", plain.Determination)
	}
}

func TestGeometryDevicePixelRatioIsDowngradedWhenMatchMediaIsReportedModified(t *testing.T) {
	kv := map[string]string{
		"geom.screen": ok(`{"width":1536,"height":864,"availWidth":1536,"availHeight":816,"devicePixelRatio":1}`),
		"geom.css":    ok(`{"dppx":2}`),
	}
	sec := run(t, with(kv, nativesSaying([]string{keyMatchMedia}, measuredNativeTargetKeys)), sectionGeometry)
	if sec.Determination != Instrumented {
		t.Fatalf("determination = %q, want instrumented: the CSS side of that ratio is recovered by bisecting a media query; rows: %+v", sec.Determination, sec.Rows)
	}
	plain := run(t, kv, sectionGeometry)
	if plain.Determination != Contradiction {
		t.Fatalf("determination = %q with no natives reading, want contradiction", plain.Determination)
	}
}

func TestGeometryAvailableSpaceConvictsWhenOnlyMatchMediaIsReportedModified(t *testing.T) {
	kv := map[string]string{
		"geom.screen": ok(`{"width":1536,"height":864,"availWidth":1600,"availHeight":816,"devicePixelRatio":1}`),
		"geom.css":    ok(`{"dppx":1}`),
	}
	sec := run(t, with(kv, nativesSaying([]string{keyMatchMedia}, measuredNativeTargetKeys)), sectionGeometry)
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction; rows: %+v", sec.Determination, sec.Rows)
	}
}

func TestNumericsAndThrowsConvictWithEveryProbedAccessorReportedModified(t *testing.T) {
	all := nativesSaying(measuredNativeTargetKeys, nil)

	numerics := with(map[string]string{
		"math.exact": ok(mathExactProbe(map[string]string{"round.halfNeg": "0"})),
	}, all)
	if sec := run(t, numerics, sectionMath); sec.Determination != Contradiction {
		t.Errorf("numerics = %q, want contradiction: no accessor the sweep probes carries a numeric built-in, so nothing there can be explained away; rows: %+v", sec.Determination, sec.Rows)
	}

	throws := throwsAllGoodPayload()
	throws["throw.mandated"] = ok(`{
		"atob.invalidChars": {
			"available": true, "threw": true, "name": "Error", "ctor": "Error",
			"ctorGlobalAvailable": true, "instanceOf": false, "toStringTag": "[object Error]", "hasStack": true,
			"attempt2": {"threw": true, "name": "Error", "ctor": "Error", "ctorGlobalAvailable": true, "instanceOf": false, "toStringTag": "[object Error]"}
		}
	}`)
	if sec := run(t, with(throws, all), sectionThrows); sec.Determination != Contradiction {
		t.Errorf("throws = %q, want contradiction: the sweep probes none of the built-ins that section reads; rows: %+v", sec.Determination, sec.Rows)
	}
}

func TestOneModifiedAccessorNoLongerScoresAsAContradiction(t *testing.T) {
	kv := with(textMetricsWidthDisagrees(), nativesSaying([]string{keyMeasureText}, measuredNativeTargetKeys))
	rep := Analyze(probes(t, kv), Inputs{})

	for _, sec := range rep.Sections {
		if sec.Determination == Contradiction {
			t.Fatalf("section %q reports a contradiction; the only disagreement in this payload is explained by an accessor the environment reports as modified; rows: %+v", sec.ID, sec.Rows)
		}
	}
	if section(t, rep, "natives").Determination != Instrumented {
		t.Errorf("natives = %q, want instrumented: nothing is being excused, the environment is still described as modified", section(t, rep, "natives").Determination)
	}
	if section(t, rep, "rects").Determination != Instrumented {
		t.Errorf("rects = %q, want instrumented", section(t, rep, "rects").Determination)
	}
	if rep.Summary.Band != BandInstrumented {
		t.Errorf("band = %q, want %q", rep.Summary.Band, BandInstrumented)
	}

	unexplained := with(textMetricsWidthDisagrees(), nativesSaying(nil, measuredNativeTargetKeys))
	other := Analyze(probes(t, unexplained), Inputs{})
	if section(t, other, "rects").Determination != Contradiction {
		t.Fatalf("rects = %q on a payload where nothing explains the disagreement, want contradiction", section(t, other, "rects").Determination)
	}
	if rep.Summary.BotLikeness >= other.Summary.BotLikeness {
		t.Errorf("botLikeness: explained %d, unexplained %d; a disagreement the environment's own report accounts for must weigh strictly less than one it does not",
			rep.Summary.BotLikeness, other.Summary.BotLikeness)
	}
}

func TestDowngradedRowsCarryNoWeight(t *testing.T) {
	sections := []struct {
		name  string
		kv    map[string]string
		build sectionFunc
	}{
		{"rects", with(textMetricsWidthDisagrees(), nativesSaying([]string{keyMeasureText}, measuredNativeTargetKeys)), sectionRects},
		{"audio", with(audioPayload(audioViewsDisagree, "0.02"), nativesSaying([]string{keyGetChannelData}, measuredNativeTargetKeys)), sectionAudioBuf},
		{"mediapaths", with(mediaPathsColourSchemeDisagrees(), nativesSaying([]string{keyMatchMedia}, measuredNativeTargetKeys)), sectionMediaPaths},
	}
	banned := []string{"weight", "score", "point", "worth", "penalt"}
	for _, s := range sections {
		t.Run(s.name, func(t *testing.T) {
			sec := run(t, s.kv, s.build)
			if sec.Determination != Instrumented {
				t.Fatalf("determination = %q, want instrumented", sec.Determination)
			}
			for _, row := range sec.Rows {
				text := strings.ToLower(row.Label + " " + row.Value + " " + row.Note)
				for _, b := range banned {
					if strings.Contains(text, b) {
						t.Errorf("row %+v says %q; a row may not say what a reading was worth", row, b)
					}
				}
			}

			if _, err := json.Marshal(sec); err != nil {
				t.Fatalf("the section does not encode: %v", err)
			}
		})
	}
}

func rowsMentionTheExplanation(sec Section, key string) bool {
	for _, row := range sec.Rows {
		if strings.Contains(row.Note, key) && strings.Contains(row.Note, "not a built-in") {
			return true
		}
	}
	return false
}
