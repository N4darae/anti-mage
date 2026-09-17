package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/N4darae/anti-mage/assess"
)

func newTestServer(t *testing.T, files fstest.MapFS) *Server {
	t.Helper()
	return New(Options{Addr: "127.0.0.1:0", Web: files, Log: log.New(io.Discard, "", 0)})
}

func TestBootstrapIssuesUnpredictableInputs(t *testing.T) {
	s := newTestServer(t, nil)
	get := func() Bootstrap {
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/bootstrap", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
		var b Bootstrap
		if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return b
	}
	a, b := get(), get()
	if a.Nonce == "" || a.Nonce == b.Nonce {
		t.Errorf("nonces %q and %q must both be present and differ", a.Nonce, b.Nonce)
	}
	if len(a.Font.Controls) != controlFamilyCount {
		t.Fatalf("got %d control families, want %d", len(a.Font.Controls), controlFamilyCount)
	}
	same := 0
	for i := range a.Font.Controls {
		if a.Font.Controls[i] == b.Font.Controls[i] {
			same++
		}
	}
	if same == len(a.Font.Controls) {
		t.Errorf("two scans were issued identical control families; a page could hold a list of the names to withhold")
	}
	if len(a.Time.Offsets) != offsetSampleCount {
		t.Fatalf("got %d instants, want %d", len(a.Time.Offsets), offsetSampleCount)
	}
	identical := true
	for i := range a.Time.Offsets {
		if a.Time.Offsets[i] != b.Time.Offsets[i] {
			identical = false
		}
	}
	if identical {
		t.Errorf("two scans were issued identical instants; a page could answer them from a table")
	}
	for _, o := range a.Time.Offsets {
		if o.Date == "" || o.EpochMs == 0 {
			t.Errorf("instant %+v must carry both a date and epoch milliseconds so the moment is unambiguous", o)
		}
	}
}

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

func TestScanReturnsOneAssessmentAndNoBreakdown(t *testing.T) {
	s := newTestServer(t, nil)
	body := `{"v":1,"mode":"public","probes":{"auto.residue":{"status":"ok","value":{"webdriver":true}}}}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/scan", strings.NewReader(body))
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("reply is not JSON: %v", err)
	}
	for _, k := range []string{"v", "determination", "score", "statement", "supplied"} {
		if _, present := out[k]; !present {
			t.Errorf("the reply carries no %q: %v", k, out)
		}
	}
	if out["determination"] != string(assess.Instrumented) {
		t.Errorf("determination = %v; a browser declaring itself under remote control is instrumented", out["determination"])
	}

	for _, k := range []string{"sections", "summary", "band", "rows", "determinations"} {
		if _, present := out[k]; present {
			t.Errorf("the reply carries %q; a caller must not be able to see which reading moved the score", k)
		}
	}
	if len(out) != 5 {
		t.Errorf("the reply carries %d fields (%v), want the five of an assessment", len(out), out)
	}
}

func TestScanRejectsAnOversizeBody(t *testing.T) {
	s := newTestServer(t, nil)
	big := `{"v":1,"probes":{"x":{"status":"ok","value":"` + strings.Repeat("A", maxScanBody+1024) + `"}}}`
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/scan", strings.NewReader(big)))
	if rec.Code == http.StatusOK {
		t.Errorf("an oversize body was accepted")
	}
}

func TestScanRejectsGarbage(t *testing.T) {
	s := newTestServer(t, nil)
	for _, body := range []string{``, `not json`, `[]`, `[1,2]`, `"a string"`, `42`, `true`, `{`, `{"probes":`} {
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/scan", strings.NewReader(body)))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %q status %d, want %d", body, rec.Code, http.StatusBadRequest)
		}
		if got := rec.Body.String(); got != "request body is not the expected JSON object\n" {
			t.Errorf("body %q answered %q; every rejection carries the one fixed message and nothing the decoder saw", body, got)
		}
	}
	for _, body := range []string{`{}`, `null`, `{"probes":[]}`, `{"probes":{}}`, `{"observations":"not an object"}`, `{"nonce":7}`} {
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/scan", strings.NewReader(body)))
		if rec.Code != http.StatusOK {
			t.Fatalf("body %q status %d, want %d", body, rec.Code, http.StatusOK)
		}
		var a assess.Assessment
		if err := json.Unmarshal(rec.Body.Bytes(), &a); err != nil {
			t.Fatalf("body %q produced a reply that is not an assessment: %v", body, err)
		}
		if a.Determination != assess.NotEvaluated {
			t.Errorf("body %q reached %q; a payload carrying no observation must reach no determination", body, a.Determination)
		}
		if a.Score != 0 {
			t.Errorf("body %q scored %d; a payload carrying no observation must score nothing", body, a.Score)
		}
	}
}

func TestScanRejectsMethodsOtherThanPost(t *testing.T) {
	s := newTestServer(t, nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/scan", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status %d, want 405", rec.Code)
	}
}

func TestStaticFilesAreServedAndTraversalIsRefused(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "secret.txt"), []byte("KEEP-THIS-OFF-THE-WIRE"), 0o600); err != nil {
		t.Fatal(err)
	}
	pages := filepath.Join(root, "web")
	if err := os.Mkdir(pages, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"index.html": "<html><head></head><body></body></html>",
		"style.css":  "body{color:red}",
	} {
		if err := os.WriteFile(filepath.Join(pages, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	s := New(Options{Addr: "127.0.0.1:0", Web: os.DirFS(pages), Log: log.New(io.Discard, "", 0)})

	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/style.css", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "color:red") {
		t.Errorf("style.css was not served: %d %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/css") {
		t.Errorf("content type %q", ct)
	}

	escapes := []string{
		"/../secret.txt", "/..%2fsecret.txt", "/%2e%2e/secret.txt",
		"/./../secret.txt", "/../../secret.txt", "/web/../../secret.txt",
		"/..\\secret.txt", "/a/../../secret.txt", "//../secret.txt",
	}
	for _, p := range escapes {
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		if bytes.Contains(rec.Body.Bytes(), []byte("KEEP-THIS-OFF-THE-WIRE")) {
			t.Errorf("%s escaped the page filesystem", p)
		}
		if rec.Code == http.StatusOK {
			t.Errorf("%s was answered %d; a path outside the page filesystem has no answer", p, rec.Code)
		}
	}
}

func TestNonLoopbackAddressIsRefused(t *testing.T) {

	for _, addr := range []string{"0.0.0.0:8787", "198.51.100.10:8787", "[::]:8787", ":8787", ":0", "[::1%25lo]:0"} {
		s := New(Options{Addr: addr, Log: log.New(io.Discard, "", 0)})
		if err := s.ListenAndServe(context.Background()); err == nil {
			t.Errorf("address %q was accepted; this server examines the browser on this machine only", addr)
		}
	}
}

func TestBootstrapJSONAlias(t *testing.T) {
	s := newTestServer(t, nil)
	for _, p := range []string{"/api/bootstrap", "/bootstrap.json"} {
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status %d", p, rec.Code)
			continue
		}
		var b Bootstrap
		if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil || b.Nonce == "" {
			t.Errorf("%s: %v", p, err)
		}
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
