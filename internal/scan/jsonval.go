package scan

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

func field(v any, path ...string) (any, bool) {
	cur := v
	for _, k := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[k]
		if !ok || cur == nil {
			return nil, false
		}
	}
	return cur, true
}

func str(v any, path ...string) (string, bool) {
	x, ok := field(v, path...)
	if !ok {
		return "", false
	}
	s, ok := x.(string)
	return s, ok
}

func num(v any, path ...string) (float64, bool) {
	x, ok := field(v, path...)
	if !ok {
		return 0, false
	}
	f, ok := x.(float64)
	if !ok || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	return f, true
}

func boolean(v any, path ...string) (bool, bool) {
	x, ok := field(v, path...)
	if !ok {
		return false, false
	}
	b, ok := x.(bool)
	return b, ok
}

func object(v any, path ...string) (map[string]any, bool) {
	x, ok := field(v, path...)
	if !ok {
		return nil, false
	}
	m, ok := x.(map[string]any)
	return m, ok
}

func strList(v any, path ...string) ([]string, bool) {
	x, ok := field(v, path...)
	if !ok {
		return nil, false
	}
	a, ok := x.([]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(a))
	for _, e := range a {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}
	return out, true
}

func nameSet(v any, truthKeys ...string) ([]string, bool) {
	if v == nil {
		return nil, false
	}
	if a, ok := v.([]any); ok {
		out := make([]string, 0, len(a))
		for _, e := range a {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		sort.Strings(out)
		return out, true
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(m))
	for k, e := range m {
		switch t := e.(type) {
		case bool:
			if t {
				out = append(out, k)
			}
		case map[string]any:
			for _, tk := range truthKeys {
				if b, ok := t[tk].(bool); ok && b {
					out = append(out, k)
					break
				}
			}
		}
	}
	sort.Strings(out)
	return out, true
}

func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func joinLimit(xs []string, n int) string {
	if len(xs) == 0 {
		return "none"
	}
	if len(xs) <= n {
		return strings.Join(xs, ", ")
	}
	return strings.Join(xs[:n], ", ") + fmt.Sprintf(", and %d more", len(xs)-n)
}

func clip(s string, n int) string {
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return ' '
		}
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func itoa(f float64) string {
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}
