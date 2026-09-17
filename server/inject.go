package server

import (
	"bytes"
	"encoding/json"
)

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
