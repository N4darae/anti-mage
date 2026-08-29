// Package server serves the scanner over loopback.

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
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

func (s *Server) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				s.logger.Printf("recovered from panic serving %s: %v", r.URL.Path, v)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
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

const bootstrapElementID = "am-bootstrap"

func injectBootstrap(page []byte, b Bootstrap) []byte {
	blob, err := json.Marshal(b)
	if err != nil {
		return page
	}
	if out, ok := replaceElementContents(page, bootstrapElementID, blob); ok {
		return appendGlobalsScript(out)
	}
	tag := []byte(`<script id="` + bootstrapElementID + `" type="application/json">`)
	tag = append(tag, blob...)
	tag = append(tag, []byte(`</script>`+"\n")...)
	tag = append(tag, globalsScript()...)

	type anchor struct {
		text  string
		after bool
	}
	for _, a := range []anchor{{"</head>", false}, {"</HEAD>", false}, {"<body>", true}, {"<BODY>", true}} {
		i := bytes.Index(page, []byte(a.text))
		if i < 0 {
			continue
		}
		if a.after {
			i += len(a.text)
		}
		out := make([]byte, 0, len(page)+len(tag))
		out = append(out, page[:i]...)
		out = append(out, tag...)
		return append(out, page[i:]...)
	}
	return append(tag, page...)
}

func replaceElementContents(page []byte, name string, body []byte) ([]byte, bool) {
	idAttr := []byte(`id="` + name + `"`)
	at := bytes.Index(page, idAttr)
	if at < 0 {
		return page, false
	}

	open := bytes.LastIndex(page[:at], []byte("<script"))
	if open < 0 {
		return page, false
	}
	closeAngle := bytes.IndexByte(page[at:], '>')
	if closeAngle < 0 {
		return page, false
	}
	contentStart := at + closeAngle + 1

	if bytes.IndexByte(page[open:at], '>') >= 0 {
		return page, false
	}
	rel := bytes.Index(page[contentStart:], []byte("</script"))
	if rel < 0 {
		return page, false
	}
	contentEnd := contentStart + rel

	out := make([]byte, 0, len(page)+len(body))
	out = append(out, page[:contentStart]...)
	out = append(out, body...)
	return append(out, page[contentEnd:]...), true
}

func globalsScript() []byte {
	return []byte(`<script>window.AM_BOOTSTRAP=(function(){try{` +
		`return JSON.parse(document.getElementById("` + bootstrapElementID + `").textContent);` +
		`}catch(e){return null}})();</script>` + "\n")
}

func appendGlobalsScript(page []byte) []byte {
	script := globalsScript()
	for _, a := range []string{"</head>", "</HEAD>"} {
		if i := bytes.Index(page, []byte(a)); i >= 0 {
			out := make([]byte, 0, len(page)+len(script))
			out = append(out, page[:i]...)
			out = append(out, script...)
			return append(out, page[i:]...)
		}
	}
	return page
}

func (s *Server) servePlaceholder(w http.ResponseWriter, r *http.Request, why string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	b, err := s.iss.issue(time.Now())
	blob := "null"
	if err == nil {
		if j, err := json.MarshalIndent(b, "", "  "); err == nil {
			blob = string(j)
		}
	}
	fmt.Fprintf(w, `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>anti-mage scanner</title>
<meta name="viewport" content="width=device-width,initial-scale=1">
<style>
:root{color-scheme:light dark}
body{font:14px/1.5 ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;margin:0 auto;padding:2rem;max-width:44rem}
h1{font-size:1.1rem}
pre{overflow-x:auto;padding:1rem;border:1px solid currentColor;border-radius:4px}
code{white-space:nowrap}
</style></head><body>
<h1>anti-mage scanner</h1>
<p>The server is running. The page is not being served, because %s.</p>
<p>The API is up regardless:</p>
<ul>
<li><code>GET /api/bootstrap</code> &mdash; this scan's server-chosen probe inputs</li>
<li><code>POST /api/scan</code> &mdash; the observation payload, returning one assessment</li>
</ul>
<p>The bootstrap for a scan started now:</p>
<pre>%s</pre>
</body></html>
`, html.EscapeString(why), html.EscapeString(blob))
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(w, "could not encode the reply", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_, _ = w.Write(buf.Bytes())
}
