package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/N4darae/anti-mage/assess"
)

func TestScanUsesTheIssuedInputsWhenTheNonceComesBack(t *testing.T) {
	s := newTestServer(t, nil)
	b, err := s.iss.issue(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	got := s.iss.resolve(b.Nonce, time.Now())
	if got.Nonce != b.Nonce || len(got.FontControls) != controlFamilyCount || len(got.OffsetDates) != offsetSampleCount {
		t.Errorf("resolve returned %+v, which does not match what was issued", got)
	}
	if unknown := s.iss.resolve("not-a-nonce", time.Now()); unknown.Nonce != "" {
		t.Errorf("an unknown nonce resolved to %+v; it must resolve to nothing so the report says the inputs were not vouched for", unknown)
	}
	if expired := s.iss.resolve(b.Nonce, time.Now().Add(issueTTL+time.Minute)); expired.Nonce != "" {
		t.Errorf("an expired nonce resolved to %+v", expired)
	}
}

func TestInputsRecoveredFromTheEchoedControlFamiliesWithoutANonce(t *testing.T) {
	s := newTestServer(t, nil)
	now := time.Now()
	b, err := s.iss.issue(now)
	if err != nil {
		t.Fatal(err)
	}

	obj := map[string]bool{}
	for _, n := range b.Font.Controls {
		obj[n] = false
	}
	controls, _ := json.Marshal(obj)
	body := `{"v":1,"mode":"public","probes":{` +
		`"font.resolved":{"status":"ok","value":["Arial"]},` +
		`"font.controls":{"status":"ok","value":` + string(controls) + `}}}`

	env, err := assess.Decode([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if env.Nonce != "" {
		t.Fatalf("the payload carried a nonce it was not given: %q", env.Nonce)
	}
	got := s.iss.resolveByControls(env.NamesReported("font.controls"), now)
	if got.Nonce != b.Nonce {
		t.Fatalf("the issued inputs were not recovered from the echoed control families: %+v", got)
	}
	if len(got.FontControls) != controlFamilyCount || len(got.OffsetDates) != offsetSampleCount {
		t.Errorf("recovery returned %+v, which does not match what was issued", got)
	}

	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/scan", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestInputsAreNotRecoveredFromNamesThatWereNeverIssued(t *testing.T) {
	s := newTestServer(t, nil)
	if _, err := s.iss.issue(time.Now()); err != nil {
		t.Fatal(err)
	}
	got := s.iss.resolveByControls([]string{"Arial Nonexistentica", "made up"}, time.Now())
	if got.Nonce != "" {
		t.Errorf("names that were never issued recovered %+v", got)
	}
	if got := s.iss.resolveByControls(nil, time.Now()); got.Nonce != "" {
		t.Errorf("an empty set recovered %+v", got)
	}
}
