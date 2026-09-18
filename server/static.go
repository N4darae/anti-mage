package server

import (
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"
)

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
