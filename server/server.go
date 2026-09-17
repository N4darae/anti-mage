// Package server serves the scanner over loopback.

package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/N4darae/anti-mage/assess"
)

const maxScanBody = 1 << 20

const scanTimeout = 30 * time.Second

type Options struct {
	Addr string

	Web fs.FS

	Log *log.Logger

	DumpDir string
}

type Server struct {
	opts   Options
	iss    *issuer
	mux    *http.ServeMux
	logger *log.Logger
}

func New(o Options) *Server {
	if o.Log == nil {
		o.Log = log.Default()
	}
	s := &Server{opts: o, iss: newIssuer(), mux: http.NewServeMux(), logger: o.Log}
	s.mux.HandleFunc("/api/bootstrap", s.handleBootstrap)
	s.mux.HandleFunc("/bootstrap.json", s.handleBootstrap)
	s.mux.HandleFunc("/api/scan", s.handleScan)
	s.mux.HandleFunc("/", s.handleStatic)
	return s
}

func (s *Server) Handler() http.Handler { return s.recoverer(s.headers(s.mux)) }

func (s *Server) ListenAndServe(ctx context.Context) error {
	host, _, err := net.SplitHostPort(s.opts.Addr)
	if err != nil {
		return fmt.Errorf("address %q: %w", s.opts.Addr, err)
	}
	if err := requireLoopback(host); err != nil {
		return err
	}
	ln, err := net.Listen("tcp", s.opts.Addr)
	if err != nil {
		return err
	}
	if err := requireLoopbackAddr(ln.Addr()); err != nil {
		_ = ln.Close()
		return err
	}
	srv := &http.Server{
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       scanTimeout,
		WriteTimeout:      scanTimeout,
		IdleTimeout:       2 * time.Minute,
	}
	s.logger.Printf("open http://%s/ in the browser you want to examine", ln.Addr().String())
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func requireLoopback(host string) error {
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("address %q is not loopback; this server serves the browser on this machine only", host)
	}
	return nil
}

func requireLoopbackAddr(a net.Addr) error {
	t, ok := a.(*net.TCPAddr)
	if !ok || t.IP == nil || !t.IP.IsLoopback() {
		return fmt.Errorf("address %q is not loopback; this server serves the browser on this machine only", a.String())
	}
	return nil
}

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

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if name == "" || strings.HasSuffix(r.URL.Path, "/") {
		name = path.Join(name, "index.html")
	}
	if s.opts.Web == nil {
		s.servePlaceholder(w, r, "no page filesystem was configured")
		return
	}
	data, err := fs.ReadFile(s.opts.Web, name)
	if err != nil {
		if name == "index.html" {
			s.servePlaceholder(w, r, "web/index.html is not present in this build")
			return
		}
		http.NotFound(w, r)
		return
	}
	if name == "index.html" {
		b, err := s.iss.issue(time.Now())
		if err != nil {
			s.logger.Printf("issuing bootstrap: %v", err)
		} else {
			data = injectBootstrap(data, b)
		}
	}
	ctype := mime.TypeByExtension(path.Ext(name))
	if ctype == "" {
		ctype = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ctype)
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	_, _ = w.Write(data)
}
