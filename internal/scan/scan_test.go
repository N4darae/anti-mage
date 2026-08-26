package scan

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func probes(t *testing.T, kv map[string]string) Request {
	t.Helper()
	var sb strings.Builder
	sb.WriteString(`{"v":1,"mode":"public","probes":{`)
	first := true
	for id, val := range kv {
		if !first {
			sb.WriteString(",")
		}
		first = false
		sb.WriteString(`"` + id + `":`)
		sb.WriteString(val)
	}
	sb.WriteString(`}}`)
	r, err := DecodeRequest([]byte(sb.String()))
	if err != nil {
		t.Fatalf("DecodeRequest(%s): %v", sb.String(), err)
	}
	return r
}

func ok(v string) string { return `{"status":"ok","value":` + v + `}` }

func section(t *testing.T, rep Report, id string) Section {
	t.Helper()
	for _, s := range rep.Sections {
		if s.ID == id {
			return s
		}
	}
	t.Fatalf("report has no section %q", id)
	return Section{}
}

func TestEmptyRequestConcludesNothing(t *testing.T) {
	rep := Analyze(Request{}, Inputs{})
	if rep.V != 1 {
		t.Errorf("v = %d, want 1", rep.V)
	}
	if len(rep.Sections) != len(order) {
		t.Fatalf("got %d sections, want %d", len(rep.Sections), len(order))
	}
	for _, s := range rep.Sections {
		switch s.Determination {
		case Inconclusive, Unverified:
		default:
			t.Errorf("section %s: determination %q on an empty request; absence must never be evidence", s.ID, s.Determination)
		}
		if s.Rows == nil {
			t.Errorf("section %s: nil rows must be an empty slice so the wire form is an array", s.ID)
		}
		if s.Title == "" {
			t.Errorf("section %s: empty title", s.ID)
		}
	}
	if rep.Summary.Band != BandNotEvaluated {
		t.Errorf("band = %q, want %q", rep.Summary.Band, BandNotEvaluated)
	}
	if rep.Summary.HumanConfidence != 0 || rep.Summary.BotLikeness != 0 {
		t.Errorf("scores = %d/%d, want 0/0", rep.Summary.HumanConfidence, rep.Summary.BotLikeness)
	}
}

func TestUnsupportedIsNeverEvidence(t *testing.T) {
	var sb strings.Builder
	sb.WriteString(`{"v":1,"mode":"public","probes":{`)
	ids := []string{
		"font.resolved", "font.coverage", "font.controls",
		"native.tostring", "native.ownkeys", "native.descriptor", "native.receiver",
		"scope.main", "scope.worker", "scope.iframe", "scope.availability",
		"geom.screen", "geom.css", "time.zone", "time.offsets",
		"media.matrix", "auto.residue", "perm.state",
	}
	for i, id := range ids {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(`"` + id + `":{"status":"unsupported","value":{"reason":"no"}}`)
	}
	sb.WriteString(`}}`)
	r, err := DecodeRequest([]byte(sb.String()))
	if err != nil {
		t.Fatal(err)
	}
	rep := Analyze(r, Inputs{})
	for _, s := range rep.Sections {
		if s.Determination == Contradiction || s.Determination == Instrumented {
			t.Errorf("section %s: %q from a payload where every probe was unsupported", s.ID, s.Determination)
		}
	}
	if rep.Summary.BotLikeness != 0 {
		t.Errorf("botLikeness = %d on an all-unsupported scan, want 0", rep.Summary.BotLikeness)
	}
}

func TestErrorStatusIsNeverEvidence(t *testing.T) {
	r := probes(t, map[string]string{
		"auto.residue": `{"status":"error","value":{"webdriver":true}}`,
		"scope.main":   `{"status":"error","value":{"platform":"Win32","userAgent":"X11; Linux"}}`,
	})
	rep := Analyze(r, Inputs{})
	if d := section(t, rep, "automation").Determination; d != Inconclusive {
		t.Errorf("automation = %q from a probe the collector reported an error for, want inconclusive", d)
	}
	if d := section(t, rep, "platform").Determination; d != Inconclusive {
		t.Errorf("platform = %q from an errored probe, want inconclusive", d)
	}
}

func TestUnknownProbeIdsIgnored(t *testing.T) {
	r := probes(t, map[string]string{
		"totally.new.probe": ok(`{"anything":1}`),
		"another":           ok(`"x"`),
	})
	rep := Analyze(r, Inputs{})
	if rep.Summary.Band != BandNotEvaluated {
		t.Errorf("band = %q; unknown ids must be ignored, not read", rep.Summary.Band)
	}
}

func TestFutureVersionDegrades(t *testing.T) {
	r, err := DecodeRequest([]byte(`{"v":99,"mode":"public","probes":{"font.resolved":{"status":"ok","value":["Arial"],"extra":{"deep":[1,2]}},"brandnew":{"status":"ok","value":1}},"unknownTop":true}`))
	if err != nil {
		t.Fatalf("a payload from a later collector must still decode: %v", err)
	}
	rep := Analyze(r, Inputs{})
	for _, s := range rep.Sections {
		if s.Determination == Contradiction || s.Determination == Instrumented {
			t.Errorf("section %s: %q from a future payload", s.ID, s.Determination)
		}
	}
}

func TestHostilePayloadsNeverPanicOrAccuse(t *testing.T) {
	hostile := []string{
		`{}`,
		`{"v":1}`,
		`{"probes":null}`,
		`{"probes":{}}`,
		`{"probes":{"font.resolved":{"status":"ok"}}}`,
		`{"probes":{"font.resolved":{"status":"ok","value":null}}}`,
		`{"probes":{"font.resolved":{"status":"ok","value":0}}}`,
		`{"probes":{"font.resolved":{"status":"ok","value":true}}}`,
		`{"probes":{"font.resolved":{"status":"ok","value":"Arial"}}}`,
		`{"probes":{"font.resolved":{"status":"ok","value":[null,1,true,{},[]]}}}`,
		`{"probes":{"font.resolved":{"status":"ok","value":{"Arial":{"ascii":"not a bool"}}}}}`,
		`{"probes":{"font.controls":{"status":"ok","value":[]}}}`,
		`{"probes":{"font.coverage":{"status":"ok","value":[1,2,3]}}}`,
		`{"probes":{"scope.main":{"status":"ok","value":[]}}}`,
		`{"probes":{"scope.main":{"status":"ok","value":{"userAgent":42,"platform":[],"userAgentData":"str"}}}}`,
		`{"probes":{"scope.worker":{"status":"ok","value":{"userAgent":{"nested":"deep"}}}}}`,
		`{"probes":{"native.tostring":{"status":"ok","value":"not an object"}}}`,
		`{"probes":{"native.tostring":{"status":"ok","value":{"a":{"toString":null}}}}}`,
		`{"probes":{"native.ownkeys":{"status":"ok","value":{"a":{"ownKeys":"nope"}}}}}`,
		`{"probes":{"native.descriptor":{"status":"ok","value":{"a":null}}}}`,
		`{"probes":{"native.receiver":{"status":"ok","value":{"a":{"threw":"yes"}}}}}`,
		`{"probes":{"geom.screen":{"status":"ok","value":{"width":-1,"height":0,"availWidth":"x"}}}}`,
		`{"probes":{"geom.screen":{"status":"ok","value":{"devicePixelRatio":0}},"geom.css":{"status":"ok","value":{"dppx":0}}}}`,
		`{"probes":{"time.zone":{"status":"ok","value":{"timeZone":"Not/AZone"}},"time.offsets":{"status":"ok","value":{"2026-01-15":0}}}}`,
		`{"probes":{"time.zone":{"status":"ok","value":123},"time.offsets":{"status":"ok","value":"nope"}}}`,
		`{"probes":{"time.offsets":{"status":"ok","value":{"not-a-date":5,"":1,"99999999999999999999":2}}}}`,
		`{"probes":{"media.matrix":{"status":"ok","value":{"a":{"b":{"c":"deep"}}}}}}`,
		`{"probes":{"perm.state":{"status":"ok","value":{"notifications":{"query":123}}}}}`,
		`{"probes":{"auto.residue":{"status":"ok","value":{"webdriver":"true","driverProperties":"x"}}}}`,

		`{"probes":{"":{"status":"ok","value":1},"native.tostring":{"status":"ok","value":{"":" bad"}}}}`,

		`{"probes":{"scope.main":{"status":"ok","value":` + deepJSON(200) + `}}}`,

		`{"probes":{"scope.main":{"status":"ok","value":{"userAgent":"` + strings.Repeat("A", 20000) + `"}}}}`,

		`{"probes":{"font.resolved":{"status":"error","value":1},"font.resolved":{"status":"ok","value":["Arial"]}}}`,
	}
	for _, in := range hostile {
		r, err := DecodeRequest([]byte(in))
		if err != nil {
			continue
		}
		rep := Analyze(r, Inputs{})
		if len(rep.Sections) != len(order) {
			t.Fatalf("input %.60s: got %d sections", in, len(rep.Sections))
		}
		for _, s := range rep.Sections {
			if !validDetermination(s.Determination) {
				t.Fatalf("input %.60s: section %s has determination %q", in, s.ID, s.Determination)
			}
			if s.Determination == Contradiction || s.Determination == Instrumented {
				t.Errorf("input %.60s: section %s answered %q; a value of the wrong shape is not a browser disagreeing with itself", in, s.ID, s.Determination)
			}
			for _, row := range s.Rows {
				for _, r := range row.Value + row.Label + row.Note {
					if r < 0x20 && r != '\n' && r != '\t' {
						t.Fatalf("input %.60s: a control character survived into a row", in)
					}
				}
			}
		}
		if _, err := json.Marshal(rep); err != nil {
			t.Fatalf("input %.60s: report will not encode: %v", in, err)
		}
	}
}

func deepJSON(n int) string {
	return strings.Repeat(`{"a":`, n) + `1` + strings.Repeat(`}`, n)
}

func TestTruncatedPayloadsRejectedNotMisread(t *testing.T) {
	full := `{"v":1,"mode":"public","probes":{"scope.main":{"status":"ok","value":{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)","platform":"Win32"}}}}`
	for i := 1; i < len(full); i++ {
		r, err := DecodeRequest([]byte(full[:i]))
		if err != nil {
			continue
		}

		rep := Analyze(r, Inputs{})
		for _, s := range rep.Sections {
			if !validDetermination(s.Determination) {
				t.Fatalf("prefix of length %d produced determination %q", i, s.Determination)
			}
			if s.Determination == Contradiction || s.Determination == Instrumented {
				t.Errorf("prefix of length %d made section %s answer %q; half a payload is not evidence", i, s.ID, s.Determination)
			}
		}
		if rep.Summary.Band != BandNotEvaluated && rep.Summary.Band != BandInsufficient && rep.Summary.Band != BandCoherent {
			t.Errorf("prefix of length %d reached the band %q", i, rep.Summary.Band)
		}
	}
}

func TestRandomBytesNeverPanic(t *testing.T) {
	rng := rand.New(rand.NewSource(20260826))
	alphabet := []byte(`{}[]",:0123456789truefalsnl \/ok.statusvaluepobes-`)
	for i := 0; i < 4000; i++ {
		n := 1 + rng.Intn(220)
		b := make([]byte, n)
		for j := range b {
			b[j] = alphabet[rng.Intn(len(alphabet))]
		}
		r, err := DecodeRequest(b)
		if err != nil {
			continue
		}
		rep := Analyze(r, Inputs{})
		if len(rep.Sections) == 0 {
			t.Fatalf("no sections for %q", b)
		}
	}
}

func TestAnalyzeIsAFunctionOfItsArguments(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":      ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)","platform":"Win32","hardwareConcurrency":8,"timeZone":"Asia/Ho_Chi_Minh"}`),
		"scope.worker":    ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)","platform":"Win32","hardwareConcurrency":8,"timeZone":"Asia/Ho_Chi_Minh"}`),
		"font.resolved":   ok(`["Arial","Cambria Math","Segoe UI Symbol"]`),
		"font.controls":   ok(`{"Zzqx 0000 Absent":false}`),
		"time.zone":       ok(`{"timeZone":"Asia/Ho_Chi_Minh","locale":"vi"}`),
		"time.offsets":    ok(`{"2026-01-15":-420,"2026-07-15":-420}`),
		"geom.screen":     ok(`{"width":1536,"height":864,"availWidth":1536,"availHeight":816,"devicePixelRatio":1}`),
		"geom.css":        ok(`{"dppx":1.0000000595493472}`),
		"auto.residue":    ok(`{"webdriver":false,"driverProperties":[]}`),
		"perm.state":      ok(`{"notifications":{"query":"prompt","actual":"default"}}`),
		"media.matrix":    ok(`{"video/mp4; codecs=\"avc1.42E01E\"":"probably"}`),
		"native.receiver": ok(`{"navigator.platform":{"threw":true,"name":"TypeError"}}`),
	})
	in := Inputs{Nonce: "n", OffsetDates: []string{"2026-01-15", "2026-07-15"}, FontControls: []string{"Zzqx 0000 Absent"}}
	first := Analyze(r, in)
	for i := 0; i < 20; i++ {
		if got := Analyze(r, in); !reflect.DeepEqual(got, first) {
			t.Fatalf("run %d differs from the first; Analyze is not a pure function", i)
		}
	}
}

const orderingViolationFonts = `["Bahnschrift","HoloLens MDL2 Assets","Ink Free","Segoe MDL2 Assets","Segoe UI Historic",` +
	`"Cambria Math","Javanese Text","Leelawadee UI","Lucida Console",` +
	`"Marlett","Myanmar Text","Segoe UI Emoji","Segoe UI Symbol"]`

const substitutedFonts = `["Cambria Math","Lucida Console","Marlett","Segoe UI Emoji","Segoe UI Symbol",` +
	`"Arial","Segoe UI","Verdana","DejaVu Sans","Liberation Sans"]`

func TestFontsTierGapIsNotAContradiction(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":    ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/149.0.0.0","platform":"Win32"}`),
		"font.resolved": ok(orderingViolationFonts),
		"font.controls": ok(`{"Zzqx a1 Absent":false}`),
	})
	rep := Analyze(r, Inputs{})
	if d := section(t, rep, "fonts").Determination; d == Contradiction {
		t.Fatalf("determination = contradiction on a tier gap; the install is cumulative but the probe is not")
	}
	if rep.Summary.BotLikeness != 0 {
		t.Errorf("botLikeness = %d on a tier gap alone, want 0", rep.Summary.BotLikeness)
	}
}

func TestFontsHonestControlAbstainsOnItsOwnPlatform(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":    ok(`{"userAgent":"Mozilla/5.0 (X11; Linux x86_64) HeadlessChrome/151.0.0.0","platform":"Linux x86_64"}`),
		"font.resolved": ok(substitutedFonts),
		"font.controls": ok(`{"Zzqx a1 Absent":false}`),
	})
	s := section(t, Analyze(r, Inputs{}), "fonts")
	if s.Determination == Contradiction || s.Determination == Instrumented {
		t.Fatalf("determination = %q on a platform that resolves Windows names through font substitution is not evidence of anything", s.Determination)
	}
}

func TestFontsHonestControlSetUnderAWindowsClaimStillDoesNotAccuse(t *testing.T) {

	r := probes(t, map[string]string{
		"scope.main":    ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)","platform":"Win32"}`),
		"font.resolved": ok(substitutedFonts),
		"font.controls": ok(`{"Zzqx a1 Absent":false}`),
	})
	s := section(t, Analyze(r, Inputs{}), "fonts")
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction; a partial match is the ordinary state of a width probe and must not be one")
	}
}

func TestFontsInventedControlResolvingGatesTheSection(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":    ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)","platform":"Win32"}`),
		"font.resolved": ok(orderingViolationFonts),
		"font.controls": ok(`{"Zzqx a1 Absent":true,"Segoe UI b2 Phantom":false}`),
	})
	s := section(t, Analyze(r, Inputs{FontControls: []string{"Zzqx a1 Absent", "Segoe UI b2 Phantom"}}), "fonts")
	if s.Determination != Inconclusive {
		t.Fatalf("determination = %q; a width probe that resolves a family that does not exist is not measuring font presence", s.Determination)
	}
}

func TestFontsWithoutControlsIsUnverified(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":    ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)","platform":"Win32"}`),
		"font.resolved": ok(orderingViolationFonts),
	})
	s := section(t, Analyze(r, Inputs{}), "fonts")
	if s.Determination != Unverified {
		t.Fatalf("determination = %q, want unverified: without invented-name controls the width probe cannot be trusted", s.Determination)
	}
}

func TestFontsCoverageAloneCannotChangeTheDetermination(t *testing.T) {

	base := map[string]string{
		"scope.main":    ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)","platform":"Win32"}`),
		"font.resolved": ok(orderingViolationFonts),
		"font.controls": ok(`{"Zzqx a1 Absent":false}`),
	}
	without := section(t, Analyze(probes(t, base), Inputs{}), "fonts").Determination

	for name, cov := range map[string]string{
		"every family fails coverage": `{"Bahnschrift":{"covers":false},"Ink Free":{"covers":false},` +
			`"Segoe MDL2 Assets":{"covers":false},"Cambria Math":{"covers":false},` +
			`"Javanese Text":{"covers":false},"Myanmar Text":{"covers":false}}`,
		"the markers fail coverage": `{"Bahnschrift":{"covers":false},"Ink Free":{"covers":false},` +
			`"HoloLens MDL2 Assets":{"covers":false},"Segoe MDL2 Assets":{"covers":false},` +
			`"Segoe UI Historic":{"covers":false}}`,
		"everything covers": `{"Bahnschrift":{"covers":true},"Cambria Math":{"covers":true}}`,
	} {
		with := map[string]string{}
		for k, v := range base {
			with[k] = v
		}
		with["font.coverage"] = ok(cov)
		got := section(t, Analyze(probes(t, with), Inputs{}), "fonts").Determination
		if got != without {
			t.Errorf("%s: determination = %q with coverage, %q without; a coverage value alone must not move it", name, got, without)
		}
	}
}

func TestFontsFullyInstalledWindowsIsNotAContradiction(t *testing.T) {
	all := []string{}
	for _, t2 := range windowsTiers() {
		all = append(all, t2.table.Values...)
	}
	enc, err := json.Marshal(all)
	if err != nil {
		t.Fatal(err)
	}
	cov := map[string]map[string]bool{}
	for _, f := range all {
		cov[f] = map[string]bool{"covers": false}
	}
	covEnc, err := json.Marshal(cov)
	if err != nil {
		t.Fatal(err)
	}
	r := probes(t, map[string]string{
		"scope.main":    ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)","platform":"Win32"}`),
		"font.resolved": ok(string(enc)),
		"font.controls": ok(`{"Zzqx a1 Absent":false}`),
		"font.coverage": ok(string(covEnc)),
	})
	rep := Analyze(r, Inputs{})
	s := section(t, rep, "fonts")
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction on a machine with every tiered family installed")
	}
	if rep.Summary.BotLikeness != 0 {
		t.Errorf("botLikeness = %d on a fully installed Windows font set, want 0", rep.Summary.BotLikeness)
	}
}

func TestFontsNoWindowsFamilyAtAllUnderAWindowsClaim(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":    ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)","platform":"Win32"}`),
		"font.resolved": ok(`["DejaVu Sans","Liberation Sans","Noto Sans"]`),
		"font.controls": ok(`{"Zzqx a1 Absent":false}`),
	})
	s := section(t, Analyze(r, Inputs{}), "fonts")
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction: Windows is claimed and no family the vendor publishes for Windows resolved", s.Determination)
	}
}

func TestScopesDisagreementIsInstrumented(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":   ok(`{"userAgent":"UA","platform":"PatchedPlatform/9000","hardwareConcurrency":1337}`),
		"scope.worker": ok(`{"userAgent":"UA","platform":"Linux x86_64","hardwareConcurrency":8}`),
		"scope.iframe": ok(`{"userAgent":"UA","platform":"Linux x86_64","hardwareConcurrency":8}`),
	})
	s := section(t, Analyze(r, Inputs{}), "scopes")
	if s.Determination != Instrumented {
		t.Fatalf("determination = %q, want instrumented", s.Determination)
	}
}

func TestScopesAgreementIsConsistent(t *testing.T) {
	one := `{"userAgent":"UA","platform":"Win32","hardwareConcurrency":8,"languages":["vi-VN","vi"],"timeZone":"Asia/Ho_Chi_Minh","locale":"vi"}`
	r := probes(t, map[string]string{
		"scope.main":   ok(one),
		"scope.worker": ok(one),
		"scope.iframe": ok(one),
	})
	s := section(t, Analyze(r, Inputs{}), "scopes")
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent", s.Determination)
	}
}

func TestScopesIgnoresFactsHonestBrowsersDisagreeOn(t *testing.T) {

	r := probes(t, map[string]string{
		"scope.main":   ok(`{"userAgent":"UA","platform":"Win32","origin":"http://127.0.0.1:8787","scope":"Window","hasWindow":true,"deviceMemory":32,"userAgentData":{"platform":"Windows"},"isSecureContext":true}`),
		"scope.iframe": ok(`{"userAgent":"UA","platform":"Win32","origin":"null","scope":"Window","hasWindow":true,"isSecureContext":false}`),
		"scope.worker": ok(`{"userAgent":"UA","platform":"Win32","origin":"http://127.0.0.1:8787","scope":"DedicatedWorkerGlobalScope","hasWindow":false}`),
	})
	s := section(t, Analyze(r, Inputs{}), "scopes")
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent: those facts differ on an honest browser by design", s.Determination)
	}
}

func TestScopesOneScopeCannotBeCompared(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":         ok(`{"userAgent":"UA","platform":"Win32"}`),
		"scope.worker":       `{"status":"unsupported","value":{"reason":"worker-src none"}}`,
		"scope.iframe":       `{"status":"unsupported","value":{"reason":"frame-src none"}}`,
		"scope.availability": ok(`{"main":true,"worker":false,"iframe":false}`),
	})
	s := section(t, Analyze(r, Inputs{}), "scopes")
	if s.Determination != Inconclusive {
		t.Fatalf("determination = %q; a scope the browser would not create is not evidence", s.Determination)
	}
}

func TestNativesGenuineAccessorsAreConsistent(t *testing.T) {
	r := probes(t, map[string]string{
		"native.tostring":   ok(`{"navigator.platform":"function get platform() { [native code] }","screen.width":"function get width() { [native code] }"}`),
		"native.ownkeys":    ok(`{"navigator.platform":{"ownKeys":["length","name"],"getOwnPropertyNames":["length","name"],"descriptors":["length","name"]}}`),
		"native.descriptor": ok(`{"navigator.platform":{"onPrototype":true}}`),
		"native.receiver":   ok(`{"navigator.platform":{"threw":true,"name":"TypeError"}}`),
	})
	s := section(t, Analyze(r, Inputs{}), "natives")
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent; got rows %+v", s.Determination, s.Rows)
	}
}

func TestNativesCrudeOverrideIsInstrumented(t *testing.T) {
	r := probes(t, map[string]string{
		"native.tostring":   ok(`{"canvas.toDataURL":"function () {\n      return orig.apply(this, arguments);\n    }"}`),
		"native.ownkeys":    ok(`{"canvas.toDataURL":{"ownKeys":["length","name","prototype"],"getOwnPropertyNames":["length","name","prototype"],"descriptors":["length","name","prototype"]}}`),
		"native.descriptor": ok(`{"navigator.platform":{"onPrototype":false}}`),
	})
	s := section(t, Analyze(r, Inputs{}), "natives")
	if s.Determination != Instrumented {
		t.Fatalf("determination = %q, want instrumented", s.Determination)
	}
}

func TestNativesCarefulOverrideCaughtByTheBrandCheck(t *testing.T) {
	r := probes(t, map[string]string{
		"native.tostring":   ok(`{"navigator.platform":"function platform() { [native code] }"}`),
		"native.ownkeys":    ok(`{"navigator.platform":{"ownKeys":["length","name"],"getOwnPropertyNames":["length","name"],"descriptors":["length","name"]}}`),
		"native.descriptor": ok(`{"navigator.platform":{"onPrototype":true}}`),
		"native.receiver":   ok(`{"navigator.platform":{"threw":false,"resultType":"string"}}`),
	})
	s := section(t, Analyze(r, Inputs{}), "natives")
	if s.Determination != Instrumented {
		t.Fatalf("determination = %q, want instrumented: an accessor that answers an alien receiver has no brand to check", s.Determination)
	}
}

func TestNativesUnnamedSerialisationAbstains(t *testing.T) {

	r := probes(t, map[string]string{
		"native.tostring": ok(`{"Function.prototype.toString":"function () { [native code] }"}`),
		"native.receiver": ok(`{"Function.prototype.toString":{"threw":true,"name":"TypeError"}}`),
	})
	s := section(t, Analyze(r, Inputs{}), "natives")
	if s.Determination == Instrumented {
		t.Fatalf("determination = instrumented; an unnamed native serialisation is conforming")
	}
}

func TestNativesConstructorPrototypeKeyAbstains(t *testing.T) {
	r := probes(t, map[string]string{
		"native.ownkeys": ok(`{"Intl.DateTimeFormat":{"kind":"ctor","ownKeys":["length","name","prototype","supportedLocalesOf"],"getOwnPropertyNames":["length","name","prototype","supportedLocalesOf"],"descriptors":["length","name","prototype","supportedLocalesOf"]}}`),
	})
	s := section(t, Analyze(r, Inputs{}), "natives")
	if s.Determination == Instrumented {
		t.Fatalf("determination = instrumented; a constructor has an own prototype property")
	}
}

func TestNativesEnumeratorDisagreement(t *testing.T) {
	r := probes(t, map[string]string{
		"native.ownkeys": ok(`{"navigator.platform":{"ownKeys":["length","name"],"getOwnPropertyNames":["length"],"descriptors":["length","name"]}}`),
	})
	s := section(t, Analyze(r, Inputs{}), "natives")
	if s.Determination != Instrumented {
		t.Fatalf("determination = %q, want instrumented", s.Determination)
	}
}

func TestGeometryAgreementIsConsistent(t *testing.T) {
	r := probes(t, map[string]string{
		"geom.screen": ok(`{"width":1536,"height":864,"availWidth":1536,"availHeight":816,"devicePixelRatio":1.33333}`),
		"geom.css":    ok(`{"dppx":1.333330094819539}`),
	})
	s := section(t, Analyze(r, Inputs{}), "geometry")
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent", s.Determination)
	}
}

func TestGeometryRatioMismatchIsAContradiction(t *testing.T) {
	r := probes(t, map[string]string{
		"geom.screen": ok(`{"width":1920,"height":1080,"availWidth":1920,"availHeight":1080,"devicePixelRatio":1}`),
		"geom.css":    ok(`{"dppx":2}`),
	})
	s := section(t, Analyze(r, Inputs{}), "geometry")
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction", s.Determination)
	}
}

func TestGeometryAvailableAreaLargerThanScreen(t *testing.T) {
	r := probes(t, map[string]string{
		"geom.screen": ok(`{"width":800,"height":600,"availWidth":1200,"availHeight":600}`),
	})
	s := section(t, Analyze(r, Inputs{}), "geometry")
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction", s.Determination)
	}
}

func TestTimeCorrectOffsetsAreConsistent(t *testing.T) {

	r := probes(t, map[string]string{
		"time.zone":    ok(`{"timeZone":"Asia/Ho_Chi_Minh","locale":"vi"}`),
		"time.offsets": ok(`{"2025-01-11":-420,"2025-07-08":-420,"2026-01-14":-420,"2026-07-19":-420}`),
	})
	in := Inputs{OffsetDates: []string{"2025-01-11", "2025-07-08", "2026-01-14", "2026-07-19"}}
	s := section(t, Analyze(r, in), "time")
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent; rows %+v", s.Determination, s.Rows)
	}
}

func TestTimeDaylightZoneHandledCorrectly(t *testing.T) {

	r := probes(t, map[string]string{
		"time.zone":    ok(`{"timeZone":"America/New_York"}`),
		"time.offsets": ok(`{"2026-01-14":300,"2026-07-19":240,"2025-01-11":300,"2025-07-08":240}`),
	})
	s := section(t, Analyze(r, Inputs{}), "time")
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent; rows %+v", s.Determination, s.Rows)
	}
}

func TestTimeConstantOffsetAgainstADaylightZoneIsAContradiction(t *testing.T) {

	r := probes(t, map[string]string{
		"time.zone":    ok(`{"timeZone":"America/New_York"}`),
		"time.offsets": ok(`{"2026-01-14":300,"2026-07-19":300,"2025-07-08":300,"2024-07-10":300}`),
	})
	s := section(t, Analyze(r, Inputs{}), "time")
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction", s.Determination)
	}
}

func TestTimeOneMismatchIsNotEnough(t *testing.T) {
	r := probes(t, map[string]string{
		"time.zone":    ok(`{"timeZone":"America/New_York"}`),
		"time.offsets": ok(`{"2026-01-14":300,"2026-07-19":240,"2025-01-11":300,"2025-07-08":300}`),
	})
	s := section(t, Analyze(r, Inputs{}), "time")
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction on a single mismatch; one stale zone rule must not be enough")
	}
}

func TestTimeUnknownZoneIsUnverified(t *testing.T) {
	r := probes(t, map[string]string{
		"time.zone":    ok(`{"timeZone":"Mars/Olympus_Mons"}`),
		"time.offsets": ok(`{"2026-01-14":0,"2026-07-19":0}`),
	})
	s := section(t, Analyze(r, Inputs{}), "time")
	if s.Determination != Unverified {
		t.Fatalf("determination = %q, want unverified: this build may be older than the browser", s.Determination)
	}
}

func TestTimeIgnoringTheIssuedInstantsIsUnverified(t *testing.T) {
	r := probes(t, map[string]string{
		"time.zone":    ok(`{"timeZone":"America/New_York"}`),
		"time.offsets": ok(`{"2001-01-01":300,"2001-07-01":240}`),
	})
	in := Inputs{OffsetDates: []string{"2026-01-14", "2026-07-19"}}
	s := section(t, Analyze(r, in), "time")
	if s.Determination != Unverified {
		t.Fatalf("determination = %q, want unverified: the collector answered about instants it chose", s.Determination)
	}
}

func TestTimeSamplesNearATransitionAreDiscarded(t *testing.T) {

	r := probes(t, map[string]string{
		"time.zone":    ok(`{"timeZone":"America/New_York"}`),
		"time.offsets": ok(`{"2026-03-08":999,"2026-01-14":300,"2026-07-19":240}`),
	})
	s := section(t, Analyze(r, Inputs{}), "time")
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent: the guarded sample must be discarded; rows %+v", s.Determination, s.Rows)
	}
}

func TestAutomationSelfDeclarationIsInstrumented(t *testing.T) {
	r := probes(t, map[string]string{"auto.residue": ok(`{"webdriver":true,"driverProperties":[]}`)})
	s := section(t, Analyze(r, Inputs{}), "automation")
	if s.Determination != Instrumented {
		t.Fatalf("determination = %q, want instrumented", s.Determination)
	}
}

func TestAutomationCleanIsConsistent(t *testing.T) {
	r := probes(t, map[string]string{"auto.residue": ok(`{"webdriver":false,"driverProperties":[]}`)})
	s := section(t, Analyze(r, Inputs{}), "automation")
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent", s.Determination)
	}
}

func TestPermissionsMismatch(t *testing.T) {
	r := probes(t, map[string]string{"perm.state": ok(`{"notifications":{"query":"prompt","actual":"denied"}}`)})
	s := section(t, Analyze(r, Inputs{}), "permissions")
	if s.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction", s.Determination)
	}
}

func TestPermissionsGeolocationIsNotChecked(t *testing.T) {

	r := probes(t, map[string]string{"perm.state": ok(`{"geolocation":{"query":"prompt","actual":"ERROR:1:User denied Geolocation"}}`)})
	s := section(t, Analyze(r, Inputs{}), "permissions")
	if s.Determination == Contradiction {
		t.Fatalf("determination = contradiction; that pairing was measured on Chromium 151")
	}
}

func TestCapabilitiesNeverReachesAVerdict(t *testing.T) {
	r := probes(t, map[string]string{
		"media.matrix": ok(`{"video/mp4":"probably","video/webm; codecs=\"vp9\"":{"canPlayType":"probably","isTypeSupported":true,"decodingInfo":{"supported":false}}}`),
	})
	s := section(t, Analyze(r, Inputs{}), "capabilities")
	if s.Determination != Unverified {
		t.Fatalf("determination = %q, want unverified: the codec matrix separated neither browser this project measured", s.Determination)
	}
}

func bandRank(b string) int {
	return map[string]int{
		BandNotEvaluated: 0,
		BandInsufficient: 1,
		BandCoherent:     2,
		BandDiscrepant:   3,
		BandInstrumented: 4,
	}[b]
}

func mix(flagged, total int, d Determination) []Section {
	var secs []Section
	for i := 0; i < flagged; i++ {
		secs = append(secs, Section{Determination: d})
	}
	for i := flagged; i < total; i++ {
		secs = append(secs, Section{Determination: Consistent})
	}
	return secs
}

func TestBandIsMonotoneInFlaggedSections(t *testing.T) {
	prev := -1
	for flagged := 0; flagged <= 9; flagged++ {
		got := bandRank(summarise(mix(flagged, 9, Contradiction)).Band)
		if got < prev {
			t.Fatalf("band rank fell from %d to %d with %d flagged; the band must be monotone", prev, got, flagged)
		}
		prev = got
	}
}

func TestBandHumanConfidenceNeverRisesWithMoreFlagged(t *testing.T) {
	last := 101
	for flagged := 0; flagged <= 9; flagged++ {
		got := summarise(mix(flagged, 9, Contradiction)).HumanConfidence
		if got > last {
			t.Fatalf("humanConfidence rose from %d to %d as flagged went to %d", last, got, flagged)
		}
		last = got
	}
}

func TestBandBotLikenessNeverFallsWithMoreFlagged(t *testing.T) {
	last := -1
	for flagged := 0; flagged <= 9; flagged++ {
		got := summarise(mix(flagged, 9, Contradiction)).BotLikeness
		if got < last {
			t.Fatalf("botLikeness fell from %d to %d as flagged went to %d", last, got, flagged)
		}
		last = got
	}
}

func TestBandDegradesInBothDirections(t *testing.T) {

	secs := []Section{
		{Determination: Consistent}, {Determination: Consistent},
		{Determination: Consistent}, {Determination: Consistent},
		{Determination: Inconclusive}, {Determination: Inconclusive},
		{Determination: Inconclusive}, {Determination: Inconclusive},
		{Determination: Inconclusive},
	}
	sum := summarise(secs)
	if sum.HumanConfidence > 50 {
		t.Errorf("humanConfidence = %d with only four of nine sections determined; want a low number", sum.HumanConfidence)
	}
	if sum.BotLikeness != 0 {
		t.Errorf("botLikeness = %d with nothing flagged, want 0", sum.BotLikeness)
	}
	if sum.Band == BandCoherent {
		t.Errorf("band = %q with fewer than half the sections determined", sum.Band)
	}
}

func TestBandNeverClaimsCertainty(t *testing.T) {
	sum := summarise(mix(0, 9, Contradiction))
	if sum.HumanConfidence > humanConfidenceCap {
		t.Errorf("humanConfidence = %d, want at most %d", sum.HumanConfidence, humanConfidenceCap)
	}
	if sum.Band != BandCoherent {
		t.Errorf("band = %q, want %q", sum.Band, BandCoherent)
	}
}

func TestBandUnverifiedSectionsAreNotCounted(t *testing.T) {
	withUnverified := summarise([]Section{
		{Determination: Consistent}, {Determination: Consistent},
		{Determination: Unverified}, {Determination: Unverified},
	})
	without := summarise([]Section{
		{Determination: Consistent}, {Determination: Consistent},
	})
	if withUnverified != without {
		t.Errorf("unverified sections changed the summary: %+v against %+v", withUnverified, without)
	}
}

func TestBandScoresAreCoarse(t *testing.T) {
	for flagged := 0; flagged <= 9; flagged++ {
		sum := summarise(mix(flagged, 9, Instrumented))
		if sum.HumanConfidence%10 != 0 || sum.BotLikeness%10 != 0 {
			t.Errorf("scores %d/%d are finer than the coarse steps the report allows", sum.HumanConfidence, sum.BotLikeness)
		}
	}
}

func TestBandSingleInstrumentedSectionReachesTheTopBand(t *testing.T) {
	if got := summarise(mix(1, 9, Instrumented)).Band; got != BandInstrumented {
		t.Errorf("band = %q, want %q", got, BandInstrumented)
	}
}

func TestReportNamesNoProductAndAssertsNoPerson(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":    ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)","platform":"Win32"}`),
		"font.resolved": ok(orderingViolationFonts),
		"font.controls": ok(`{"Zzqx a1 Absent":false}`),
		"auto.residue":  ok(`{"webdriver":true}`),
	})
	rep := Analyze(r, Inputs{})
	blob, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(blob))
	for _, forbidden := range []string{"is a bot", "you are a bot", "definitely", "certainly", "guaranteed"} {
		if strings.Contains(text, forbidden) {
			t.Errorf("report contains %q", forbidden)
		}
	}
	if !strings.Contains(text, "instrumented") {
		t.Errorf("a scan with a self-declared remote-control flag should reach the instrumented band")
	}
}

func TestNativesInstanceShadowingViaTheDescriptorShape(t *testing.T) {

	shadowed := probes(t, map[string]string{
		"native.descriptor": ok(`{"navigator.platform":{"instanceOwnDescriptor":{"kind":"accessor","enumerable":false}}}`),
	})
	if d := section(t, Analyze(shadowed, Inputs{}), "natives").Determination; d != Instrumented {
		t.Errorf("determination = %q, want instrumented", d)
	}
	clean := probes(t, map[string]string{
		"native.descriptor": ok(`{"navigator.platform":{"instanceOwnDescriptor":null}}`),
	})
	if d := section(t, Analyze(clean, Inputs{}), "natives").Determination; d != Consistent {
		t.Errorf("determination = %q, want consistent: a null instance descriptor is the correct answer", d)
	}
	absent := probes(t, map[string]string{
		"native.descriptor": ok(`{"navigator.platform":{}}`),
	})
	if d := section(t, Analyze(absent, Inputs{}), "natives").Determination; d != Inconclusive {
		t.Errorf("determination = %q, want inconclusive: the collector reported nothing to test", d)
	}
}

func TestNativesUnforgeableMemberAbstains(t *testing.T) {
	r := probes(t, map[string]string{
		"native.descriptor": ok(`{"window.location":{"unforgeable":true,"onPrototype":false}}`),
	})
	if d := section(t, Analyze(r, Inputs{}), "natives").Determination; d == Instrumented {
		t.Errorf("determination = instrumented; a member the interface definition marks unforgeable does live on the instance")
	}
}

func TestSuppliedSectionsKeepTheirIdentityAndReachTheSummary(t *testing.T) {
	rep := AnalyzeWith(Request{}, Inputs{}, []Section{
		{ID: "mine", Title: "a reading of my own", Determination: Contradiction},
		{ID: "also-mine", Title: "another", Determination: Consistent},
	})
	if len(rep.Sections) != len(order)+2 {
		t.Fatalf("got %d sections, want %d", len(rep.Sections), len(order)+2)
	}
	got := section(t, rep, "mine")
	if got.Title != "a reading of my own" {
		t.Errorf("title = %q; nothing in this package owns the identity of a section it did not build", got.Title)
	}
	if got.Rows == nil {
		t.Errorf("rows must be an empty slice rather than a null")
	}
	if rep.Summary.Band != BandDiscrepant {
		t.Errorf("band = %q, want %q: a supplied section reaches the summary the same way a built one does", rep.Summary.Band, BandDiscrepant)
	}
}

func TestASuppliedSectionCannotWidenTheVocabulary(t *testing.T) {
	rep := AnalyzeWith(Request{}, Inputs{}, []Section{
		{ID: "mine", Determination: Determination("catastrophic")},
	})
	if got := section(t, rep, "mine").Determination; got != Inconclusive {
		t.Errorf("determination = %q, want %q: an invented determination is not evidence", got, Inconclusive)
	}
	if rep.Summary.Band != BandNotEvaluated {
		t.Errorf("band = %q, want %q", rep.Summary.Band, BandNotEvaluated)
	}
}

func TestAnalyzeIsAnalyzeWithNothingSupplied(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main": ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)","platform":"Win32"}`),
	})
	if !reflect.DeepEqual(Analyze(r, Inputs{}), AnalyzeWith(r, Inputs{}, nil)) {
		t.Errorf("Analyze and AnalyzeWith disagree on the same payload")
	}
}

func TestReportedNamesReadsNoConclusion(t *testing.T) {
	r := probes(t, map[string]string{
		"keyed":  ok(`{"a":false,"b":true}`),
		"listed": ok(`["a","b","c"]`),
		"scalar": ok(`42`),
		"broken": `{"status":"unsupported","value":{"reason":"none"}}`,
	})
	if got := r.ReportedNames("keyed"); len(got) != 2 {
		t.Errorf("keyed = %v", got)
	}
	if got := r.ReportedNames("listed"); len(got) != 3 {
		t.Errorf("listed = %v", got)
	}
	for _, id := range []string{"scalar", "broken", "absent"} {
		if got := r.ReportedNames(id); len(got) != 0 {
			t.Errorf("%s = %v, want nothing", id, got)
		}
	}
}
