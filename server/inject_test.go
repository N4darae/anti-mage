package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestBootstrapIsInjectedIntoTheIndexPage(t *testing.T) {
	files := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!doctype html><html><head><title>t</title></head><body>hello</body></html>")},
		"style.css":  &fstest.MapFile{Data: []byte("body{}")},
	}
	s := newTestServer(t, files)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="`+bootstrapElementID+`"`) {
		t.Fatalf("the bootstrap element is not in the page: %s", body)
	}
	if !strings.Contains(body, "hello") {
		t.Errorf("the page's own content was lost")
	}
	if strings.Index(body, bootstrapElementID) > strings.Index(body, "</head>") {
		t.Errorf("the bootstrap must be injected before the collector's own scripts can run")
	}

	start := strings.Index(body, `type="application/json">`) + len(`type="application/json">`)
	end := strings.Index(body[start:], "</script>")
	var b Bootstrap
	if err := json.Unmarshal([]byte(body[start:start+end]), &b); err != nil {
		t.Fatalf("the injected JSON does not parse: %v", err)
	}
	if b.Nonce == "" {
		t.Errorf("the injected bootstrap carries no nonce")
	}
}

func TestBootstrapReplacesThePagesOwnElementInPlace(t *testing.T) {
	page := `<!doctype html><html><head><title>t</title>` +
		`<script type="application/json" id="am-bootstrap">{}</script>` +
		`</head><body>hello</body></html>`
	files := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte(page)}}
	s := newTestServer(t, files)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	body := rec.Body.String()
	if strings.Count(body, `id="am-bootstrap"`) != 1 {
		t.Fatalf("the page's own element must be filled in, not duplicated:\n%s", body)
	}
	if strings.Contains(body, `id="am-bootstrap">{}<`) {
		t.Errorf("the element was left empty")
	}
	if !strings.Contains(body, "hello") || !strings.Contains(body, "<title>t</title>") {
		t.Errorf("the page's own content was damaged:\n%s", body)
	}
	var b Bootstrap
	start := strings.Index(body, `id="am-bootstrap">`) + len(`id="am-bootstrap">`)
	end := start + strings.Index(body[start:], "</script>")
	if err := json.Unmarshal([]byte(body[start:end]), &b); err != nil {
		t.Fatalf("injected JSON does not parse: %v", err)
	}
	if b.Nonce == "" || len(b.Fonts.Controls) == 0 || len(b.Dates) == 0 {
		t.Errorf("the bootstrap must carry the nonce and both inputs under both spellings: %+v", b)
	}
	if !strings.Contains(body, "window.AM_BOOTSTRAP") {
		t.Errorf("the global a collector may read was not set")
	}

	if n := strings.Count(body, b.Nonce); n != 1 {
		t.Errorf("the nonce appears %d times; it must appear only inside the JSON element, never in a script a browser executes", n)
	}
}

func TestBootstrapLeavesAPageItCannotModifyAlone(t *testing.T) {

	page := `<html><head><script type="application/json" id="am-bootstrap">{}`
	files := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte(page)}}
	s := newTestServer(t, files)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(rec.Body.String(), page) {
		t.Errorf("the page was altered even though its structure was not the expected one:\n%s", rec.Body.String())
	}
}
