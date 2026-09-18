// Package server serves the scanner over loopback.

package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"path"
	"strings"
	"time"
)

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
