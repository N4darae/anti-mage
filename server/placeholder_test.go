package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestPlaceholderWhenTheIndexPageIsAbsent(t *testing.T) {
	s := newTestServer(t, fstest.MapFS{"style.css": &fstest.MapFile{Data: []byte("body{}")}})
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "/api/scan") {
		t.Errorf("the placeholder must say what is still reachable: %s", body)
	}
	if !strings.Contains(body, "not present") {
		t.Errorf("the placeholder must say what is missing rather than pretend to be the page")
	}
}
