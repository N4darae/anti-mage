package scan

import (
	"sort"
	"strconv"
	_ "time/tzdata"
)

func sectionCapabilities(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	raw, ok := r.value("media.matrix")
	if !ok {
		s.add("codec answers", "not collected")
		return s
	}
	m, isMap := raw.(map[string]any)
	if !isMap {
		s.add("codec answers", "not readable")
		return s
	}

	type entry struct{ label, value string }
	var entries []entry
	for _, k := range keys(m) {
		entries = append(entries, entry{k, renderCodec(m[k])})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].label < entries[j].label })
	const maxRows = 40
	for i, e := range entries {
		if i >= maxRows {
			s.add("further answers", strconv.Itoa(len(entries)-maxRows)+" more not shown")
			break
		}
		s.add(clip(e.label, 120), e.value)
	}
	s.conclude("reported, not weighed")
	return s
}

func renderCodec(v any) string {
	switch t := v.(type) {
	case string:
		if t == "" {
			return `"" (no answer)`
		}
		return clip(t, 60)
	case bool:
		if t {
			return "supported"
		}
		return "not supported"
	case map[string]any:
		out := ""
		for _, k := range keys(t) {
			part := ""
			switch x := t[k].(type) {
			case string:
				part = k + " " + clip(x, 40)
			case bool:
				part = k + " " + boolWord(x)
			case float64:
				part = k + " " + itoa(x)
			default:
				continue
			}
			if out != "" {
				out += ", "
			}
			out += part
		}
		if out == "" {
			return "no answer"
		}
		return clip(out, 200)
	}
	return "no answer"
}
