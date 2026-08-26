package scan

import "testing"

func unsup(reason string) string {
	return `{"status":"unsupported","value":` + jsonQuote(reason) + `}`
}

func jsonQuote(s string) string { return `"` + s + `"` }

func TestSuppressedProbeIsAContradictionWhenTheFacilityIsProven(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":      ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/149.0.0.0","platform":"Win32"}`),
		"font.resolved":   unsup("no canvas"),
		"native.tostring": ok(`{"CanvasRenderingContext2D.measureText":"function measureText() { [native code] }"}`),
	})
	rep := Analyze(r, Inputs{})
	if d := section(t, rep, "claims").Determination; d != Contradiction {
		t.Fatalf("claims = %q, want contradiction: the width probe was called unsupported while the interface it needs was reported as a working native", d)
	}
	if rep.Summary.BotLikeness == 0 {
		t.Errorf("botLikeness = 0 while a payload contradicts itself")
	}
}

func TestSuppressedProbeIsSilentWhenNothingProvesTheFacility(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":    ok(`{"userAgent":"Mozilla/5.0 (X11; Linux x86_64) Chrome/151.0.0.0","platform":"Linux x86_64"}`),
		"font.resolved": unsup("no canvas"),
	})
	if d := section(t, Analyze(r, Inputs{}), "claims").Determination; d == Contradiction {
		t.Fatalf("claims = contradiction with nothing to corroborate it")
	}
}

func TestErrorStatusIsNotASuppressionClaim(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":      ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/149.0.0.0","platform":"Win32"}`),
		"native.tostring": ok(`{"CanvasRenderingContext2D.measureText":"function measureText() { [native code] }"}`),
	})
	r.Probes["font.resolved"] = Probe{Status: StatusError, Value: []byte(`"threw"`)}
	if d := section(t, Analyze(r, Inputs{}), "claims").Determination; d == Contradiction {
		t.Fatalf("claims = contradiction on an error status; an error is not a claim of inability")
	}
}

func TestNoSuppressionIsNotAConclusion(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":    ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/149.0.0.0","platform":"Win32"}`),
		"font.resolved": ok(`["Cambria Math"]`),
	})
	if d := section(t, Analyze(r, Inputs{}), "claims").Determination; d == Contradiction || d == Instrumented {
		t.Fatalf("claims = %q with no probe reporting itself unsupported", d)
	}
}

func TestTimeoutIsNotAClaimOfInability(t *testing.T) {
	for _, reason := range []string{
		"timed out after 9000 ms: worker",
		"aborted",
		"cancelled by the page",
		"took too long",
	} {
		r := probes(t, map[string]string{
			"scope.main":      ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/149.0.0.0","platform":"Win32"}`),
			"native.tostring": ok(`{"CanvasRenderingContext2D.measureText":"function measureText() { [native code] }"}`),
		})
		r.Probes["font.resolved"] = Probe{Status: StatusUnsupported, Value: []byte(`{"reason":` + jsonQuote(reason) + `}`)}
		rep := Analyze(r, Inputs{})
		if d := section(t, rep, "claims").Determination; d == Contradiction {
			t.Errorf("reason %q produced a contradiction; a probe that did not finish claims nothing", reason)
		}
		if rep.Summary.BotLikeness != 0 {
			t.Errorf("reason %q gave botLikeness %d, want 0", reason, rep.Summary.BotLikeness)
		}
	}
}

func TestAbsenceReasonStillContradicts(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":      ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/149.0.0.0","platform":"Win32"}`),
		"native.tostring": ok(`{"CanvasRenderingContext2D.measureText":"function measureText() { [native code] }"}`),
	})
	r.Probes["font.resolved"] = Probe{Status: StatusUnsupported, Value: []byte(`{"reason":"facility not available"}`)}
	if d := section(t, Analyze(r, Inputs{}), "claims").Determination; d != Contradiction {
		t.Fatalf("claims = %q, want contradiction", d)
	}
}

func TestClaimedWaitLongerThanTheScanIsAContradiction(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":      ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/149.0.0.0","platform":"Win32"}`),
		"native.tostring": ok(`{"CanvasRenderingContext2D.measureText":"function measureText() { [native code] }"}`),
	})
	r.Probes["font.resolved"] = Probe{Status: StatusUnsupported,
		Value: []byte(`{"reason":"timed out after 9000 ms: font.resolved"}`)}

	rep := Analyze(r, Inputs{Nonce: "n", ElapsedMS: 400})
	if d := section(t, rep, "claims").Determination; d != Contradiction {
		t.Fatalf("claims = %q; a probe cannot wait 9000 ms inside a 400 ms scan", d)
	}
	if rep.Summary.BotLikeness == 0 {
		t.Errorf("botLikeness = 0 on an impossible duration")
	}
}

func TestClaimedWaitInsideTheScanIsNotAContradiction(t *testing.T) {
	for _, elapsed := range []int{9000, 9200, 12000, 60000} {
		r := probes(t, map[string]string{
			"scope.main":      ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/149.0.0.0","platform":"Win32"}`),
			"native.tostring": ok(`{"CanvasRenderingContext2D.measureText":"function measureText() { [native code] }"}`),
		})
		r.Probes["font.resolved"] = Probe{Status: StatusUnsupported,
			Value: []byte(`{"reason":"timed out after 9000 ms: font.resolved"}`)}
		rep := Analyze(r, Inputs{Nonce: "n", ElapsedMS: elapsed})
		if d := section(t, rep, "claims").Determination; d == Contradiction {
			t.Errorf("elapsed %d ms: claims = contradiction; the scan was long enough to contain the wait", elapsed)
		}
	}
}

func TestNoServerMeasurementMeansNoClockFinding(t *testing.T) {
	r := probes(t, map[string]string{
		"scope.main":      ok(`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/149.0.0.0","platform":"Win32"}`),
		"native.tostring": ok(`{"CanvasRenderingContext2D.measureText":"function measureText() { [native code] }"}`),
	})
	r.Probes["font.resolved"] = Probe{Status: StatusUnsupported,
		Value: []byte(`{"reason":"timed out after 9000 ms: font.resolved"}`)}
	if d := section(t, Analyze(r, Inputs{}), "claims").Determination; d == Contradiction {
		t.Fatalf("claims = contradiction with no server measurement to compare against")
	}
}

func TestClaimedWaitParsing(t *testing.T) {
	for _, c := range []struct {
		in    string
		want  int
		named bool
	}{
		{"timed out after 9000 ms: worker", 9000, true},
		{"timed out after 9000ms", 9000, true},
		{"aborted after 3 s", 3000, true},
		{"waited 12 seconds", 12000, true},
		{"cancelled", 0, false},
		{"facility not available", 0, false},
		{"timed out after 500 ms, then 9000 ms", 9000, true},
	} {
		got, named := claimedWaitMS(c.in)
		if named != c.named || got != c.want {
			t.Errorf("claimedWaitMS(%q) = (%d, %v), want (%d, %v)", c.in, got, named, c.want, c.named)
		}
	}
}
