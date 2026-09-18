package server

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
