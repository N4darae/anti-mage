package assess

import (
	"encoding/json"
	"math/rand"
	"strings"
	"testing"
)

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
