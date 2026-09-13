package scan

import (
	"strings"
	_ "time/tzdata"

	"github.com/N4darae/anti-mage/reference"
)

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
