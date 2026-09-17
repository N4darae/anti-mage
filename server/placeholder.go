package server

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"time"
)

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
