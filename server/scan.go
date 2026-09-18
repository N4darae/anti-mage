package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/N4darae/anti-mage/assess"
)

const maxScanBody = 1 << 20

func (s *Server) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	b, err := s.iss.issue(time.Now())
	if err != nil {
		s.logger.Printf("issuing bootstrap: %v", err)
		http.Error(w, "could not issue scan inputs", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxScanBody))
	if err != nil {
		http.Error(w, "request body too large or unreadable", http.StatusRequestEntityTooLarge)
		return
	}
	env, err := assess.Decode(body)
	if err != nil {
		http.Error(w, "request body is not the expected JSON object", http.StatusBadRequest)
		return
	}

	nonce := env.Nonce
	if nonce == "" {
		nonce = r.Header.Get("X-Anti-Mage-Nonce")
	}
	if nonce == "" {
		nonce = r.URL.Query().Get("nonce")
	}
	now := time.Now()
	inputs := s.iss.resolve(nonce, now)
	if inputs.Nonce == "" {

		inputs = s.iss.resolveByControls(env.NamesReported("font.controls"), now)
	}

	env.Nonce = inputs.Nonce
	env.OffsetDates = inputs.OffsetDates
	env.FontControls = inputs.FontControls
	env.ElapsedMS = inputs.ElapsedMS

	rep := assess.Evaluate(env)
	s.dumpPayload(body, rep)

	writeJSON(w, http.StatusOK, rep)
}

func (s *Server) dumpPayload(body []byte, rep assess.Assessment) {
	if s.opts.DumpDir == "" {
		return
	}
	if err := os.MkdirAll(s.opts.DumpDir, 0o755); err != nil {
		s.logger.Printf("dump directory: %v", err)
		return
	}
	name := fmt.Sprintf("scan-%s-%s-%d.json", time.Now().Format("20060102-150405"), rep.Determination, rep.Score)
	dest := filepath.Join(s.opts.DumpDir, name)
	if err := os.WriteFile(dest, body, 0o644); err != nil {
		s.logger.Printf("writing %s: %v", dest, err)
		return
	}
	s.logger.Printf("saved this payload to %s", dest)
}
