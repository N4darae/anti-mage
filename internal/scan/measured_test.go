package scan

import "testing"

const measuredWindowsScope = `{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36",` +
	`"platform":"Win32","hardwareConcurrency":8,"language":"vi-VN","languages":["vi-VN","vi","en-US","en"],` +
	`"timeZone":"Asia/Ho_Chi_Minh","locale":"vi"}`

const measuredControlScope = `{"userAgent":"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/151.0.0.0 Safari/537.36",` +
	`"platform":"Linux x86_64","userAgentData":{"platform":"Linux"},"hardwareConcurrency":8,"language":"en-US","languages":["en-US"],` +
	`"timeZone":"UTC","locale":"en-US"}`

var measuredNatives = map[string]string{
	"native.tostring":   ok(`{"navigator.platform":"function get platform() { [native code] }","screen.width":"function get width() { [native code] }"}`),
	"native.ownkeys":    ok(`{"navigator.platform":{"ownKeys":["length","name"],"getOwnPropertyNames":["length","name"],"descriptors":["length","name"]}}`),
	"native.descriptor": ok(`{"navigator.platform":{"onPrototype":true},"screen.width":{"onPrototype":true}}`),
	"native.receiver":   ok(`{"navigator.platform":{"threw":true,"name":"TypeError"},"screen.width":{"threw":true,"name":"TypeError"}}`),
}

var measuredInputs = Inputs{
	Nonce:        "issued-for-this-scan",
	OffsetDates:  []string{"2025-01-11", "2025-07-08", "2026-01-14", "2026-07-19"},
	FontControls: []string{"Zzqx a1 Absent", "Segoe UI b2 Phantom"},
}

func withNatives(kv map[string]string) map[string]string {
	for k, v := range measuredNatives {
		kv[k] = v
	}
	return kv
}

func TestMeasuredWindowsClaimIsNotAccused(t *testing.T) {
	r := probes(t, withNatives(map[string]string{
		"scope.main":    ok(measuredWindowsScope),
		"scope.worker":  ok(measuredWindowsScope),
		"scope.iframe":  ok(measuredWindowsScope),
		"font.resolved": ok(orderingViolationFonts),
		"font.controls": ok(`{"Zzqx a1 Absent":false,"Segoe UI b2 Phantom":false}`),
		"geom.screen":   ok(`{"width":1536,"height":864,"availWidth":1536,"availHeight":816,"devicePixelRatio":1}`),
		"geom.css":      ok(`{"dppx":1.0000000595493472}`),
		"time.zone":     ok(`{"timeZone":"Asia/Ho_Chi_Minh","locale":"vi"}`),
		"time.offsets":  ok(`{"2025-01-11":-420,"2025-07-08":-420,"2026-01-14":-420,"2026-07-19":-420}`),
		"auto.residue":  ok(`{"webdriver":false,"driverProperties":[]}`),
		"perm.state":    ok(`{"notifications":{"query":"prompt","actual":"default"}}`),
		"media.matrix":  ok(`{"video/mp4; codecs=avc1.42E01E":"probably","video/webm; codecs=vp9":"probably"}`),
	}))
	rep := Analyze(r, measuredInputs)

	want := map[string]Determination{
		"platform":     Consistent,
		"fonts":        Consistent,
		"scopes":       Consistent,
		"natives":      Consistent,
		"geometry":     Consistent,
		"time":         Consistent,
		"automation":   Consistent,
		"permissions":  Consistent,
		"capabilities": Unverified,
	}
	for id, w := range want {
		if got := section(t, rep, id).Determination; got != w {
			t.Errorf("section %s = %q, want %q", id, got, w)
		}
	}

	if rep.Summary.Band != BandCoherent {
		t.Errorf("band = %q, want %q", rep.Summary.Band, BandCoherent)
	}
	if rep.Summary.BotLikeness != 0 {
		t.Errorf("botLikeness = %d on a tier gap alone, want 0", rep.Summary.BotLikeness)
	}
}

func TestMeasuredHonestControlIsNeverAccused(t *testing.T) {
	r := probes(t, withNatives(map[string]string{
		"scope.main":    ok(measuredControlScope),
		"scope.worker":  ok(measuredControlScope),
		"scope.iframe":  ok(measuredControlScope),
		"font.resolved": ok(substitutedFonts),
		"font.controls": ok(`{"Zzqx a1 Absent":false,"Segoe UI b2 Phantom":false}`),
		"geom.screen":   ok(`{"width":1920,"height":1080,"availWidth":1920,"availHeight":1080,"devicePixelRatio":1}`),
		"geom.css":      ok(`{"dppx":1.0000000595}`),
		"time.zone":     ok(`{"timeZone":"UTC","locale":"en-US"}`),
		"time.offsets":  ok(`{"2025-01-11":0,"2025-07-08":0,"2026-01-14":0,"2026-07-19":0}`),
		"auto.residue":  ok(`{"webdriver":false,"driverProperties":[]}`),
		"perm.state":    ok(`{"notifications":{"query":"prompt","actual":"default"}}`),
		"media.matrix":  ok(`{"video/mp4; codecs=avc1.42E01E":"probably","video/webm; codecs=vp9":"probably"}`),
	}))
	rep := Analyze(r, measuredInputs)

	for _, s := range rep.Sections {
		if s.Determination == Contradiction || s.Determination == Instrumented {
			t.Errorf("section %s = %q on a conforming browser; every such check must be dropped or made to abstain", s.ID, s.Determination)
		}
	}
	if rep.Summary.Band != BandCoherent {
		t.Errorf("band = %q, want %q", rep.Summary.Band, BandCoherent)
	}
	if rep.Summary.BotLikeness != 0 {
		t.Errorf("botLikeness = %d on a conforming browser, want 0", rep.Summary.BotLikeness)
	}
	if d := section(t, rep, "fonts").Determination; d != Inconclusive {
		t.Errorf("fonts = %q on a Linux machine claiming Linux; the verified release tables describe Windows, so the section must abstain", d)
	}
}

func TestMeasuredCarefulOverrideOnEitherBrowser(t *testing.T) {
	patched := `{"userAgent":"Mozilla/5.0 (X11; Linux x86_64) HeadlessChrome/151.0.0.0","platform":"PatchedPlatform/9000","hardwareConcurrency":1337}`
	truth := `{"userAgent":"Mozilla/5.0 (X11; Linux x86_64) HeadlessChrome/151.0.0.0","platform":"Linux x86_64","hardwareConcurrency":8}`
	r := probes(t, map[string]string{
		"scope.main":   ok(patched),
		"scope.worker": ok(truth),
		"scope.iframe": ok(truth),
	})
	rep := Analyze(r, measuredInputs)
	if d := section(t, rep, "scopes").Determination; d != Instrumented {
		t.Errorf("scopes = %q, want instrumented", d)
	}
	if rep.Summary.Band != BandInstrumented {
		t.Errorf("band = %q, want %q", rep.Summary.Band, BandInstrumented)
	}
}
