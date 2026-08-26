package scan

import (
	"sort"
	"strconv"
	"strings"

	"github.com/N4darae/anti-mage/reference"
)

func sectionNatives(_ Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}
	n := c.natives

	s.Rows = append(s.Rows, Row{
		Label: "accessors probed",
		Value: strconv.Itoa(len(n.targets)),
		Note:  "requirements applied: " + strconv.Itoa(n.applied),
	})
	if n.applied == 0 {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "no requirement could be applied", Note: "the collector reported nothing this engine could test"})
		return s
	}
	if len(n.violations) == 0 {
		s.Determination = Consistent
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "every requirement held on every accessor tested",
			Note:  "serialisation, enumerator agreement, where the member lives, and the receiver brand check",
		})
		return s
	}
	shown := n.violations
	if len(shown) > 8 {
		shown = shown[:8]
	}
	for _, v := range shown {
		s.Rows = append(s.Rows, Row{Label: clip(v.target, 120), Value: "does not meet a requirement on a built-in", Note: v.what})
	}
	s.Determination = Instrumented
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: strconv.Itoa(len(n.violations)) + " of " + strconv.Itoa(n.applied) + " requirements did not hold",
		Note:  "an accessor that claims to be a built-in is not behaving as one; something in this page's JavaScript environment has been redefined",
	})
	return s
}

type toStringVerdict int

const (
	toStringNative toStringVerdict = iota
	toStringForeign
	toStringWrongName
	toStringAbstain
)

func classifyToString(src, name string) toStringVerdict {
	if src == "" {
		return toStringAbstain
	}
	forms := reference.NativeToStringForms(name)
	for _, f := range forms.Measured.Values {
		if src == f {
			return toStringNative
		}
	}
	if forms.OtherEngine.Verified {
		for _, f := range forms.OtherEngine.Values {
			if src == f {
				return toStringNative
			}
		}
	}
	if !strings.Contains(src, "[native code]") {
		return toStringForeign
	}

	declared, ok := declaredName(src)
	if !ok || declared == "" {
		return toStringAbstain
	}
	if declared == name || declared == "get "+name || declared == "set "+name {
		return toStringNative
	}
	return toStringWrongName
}

func declaredName(src string) (string, bool) {
	const kw = "function "
	i := strings.Index(src, kw)
	if i < 0 {
		return "", false
	}
	rest := src[i+len(kw):]
	j := strings.IndexByte(rest, '(')
	if j < 0 {
		return "", false
	}
	return strings.TrimSpace(rest[:j]), true
}

func propertyName(id string) string {
	if i := strings.LastIndexByte(id, '.'); i >= 0 && i+1 < len(id) {
		return id[i+1:]
	}
	return id
}

func readToString(v any) (string, bool) {
	if s, ok := v.(string); ok {
		return s, true
	}
	for _, k := range []string{"toString", "toStr", "source", "src", "value"} {
		if s, ok := str(v, k); ok {
			return s, true
		}
	}
	return "", false
}

func readOwnKeys(v any) (sets [][]string, isCtor, ok bool) {
	if a, isArr := v.([]any); isArr {
		one := make([]string, 0, len(a))
		for _, e := range a {
			if s, isStr := e.(string); isStr {
				one = append(one, s)
			}
		}
		return [][]string{one}, false, true
	}
	m, isMap := v.(map[string]any)
	if !isMap {
		return nil, false, false
	}
	for _, k := range []string{"ownKeys", "reflectOwnKeys", "getOwnPropertyNames", "gopn", "descriptors", "gopd", "propertyDescriptors"} {
		if l, have := strList(m, k); have {
			sets = append(sets, l)
		}
	}
	for _, k := range []string{"constructor", "isConstructor", "ctor"} {
		if b, have := boolean(m, k); have && b {
			isCtor = true
		}
	}
	if s, have := str(m, "kind"); have && s == "ctor" {
		isCtor = true
	}
	return sets, isCtor, len(sets) > 0
}

func readDescriptor(v any) (onProto, unforgeable, ok bool) {
	if b, isBool := v.(bool); isBool {
		return b, false, true
	}
	for _, k := range []string{"unforgeable", "legacyUnforgeable"} {
		if b, have := boolean(v, k); have && b {
			unforgeable = true
		}
	}

	if m, isMap := v.(map[string]any); isMap {
		for _, k := range []string{"shadowedOnInstance", "shadowed", "instanceShadows"} {
			raw, present := m[k]
			if !present || raw == nil {
				continue
			}
			if b, isBool := raw.(bool); isBool {
				return !b, unforgeable, true
			}
		}
	}
	for _, k := range []string{"onPrototype", "onProto", "prototypeOwns", "descExists"} {
		if b, have := boolean(v, k); have {
			return b, unforgeable, true
		}
	}

	m, isMap := v.(map[string]any)
	if !isMap {
		return false, unforgeable, false
	}
	for _, k := range []string{"instanceOwnDescriptor", "instanceOwn", "ownOnInstance"} {
		raw, present := m[k]
		if !present {
			continue
		}
		return raw == nil, unforgeable, true
	}
	return false, unforgeable, false
}

func readReceiver(v any) (threw, skipped, ok bool) {
	if b, isBool := v.(bool); isBool {
		return b, false, true
	}
	for _, k := range []string{"skipped", "skip"} {
		if b, have := boolean(v, k); have && b {
			return false, true, true
		}
	}
	if b, have := boolean(v, "isTypeError"); have {
		return b, false, true
	}
	if b, have := boolean(v, "threw"); have {

		if !b {
			return false, false, true
		}
		if n, haveName := str(v, "name"); haveName {
			return n == "TypeError", false, true
		}
		return false, true, true
	}
	return false, false, false
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x := append([]string(nil), a...)
	y := append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}
