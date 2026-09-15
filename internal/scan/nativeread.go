package scan

import (
	"sort"
	_ "time/tzdata"
)

type nativeReading struct {
	targets []string

	applied int

	violations []nativeViolation

	modified map[string]bool
}

type nativeViolation struct {
	target string
	what   string
}

func (n nativeReading) reportsModified(key string) bool {
	return n.modified[key]
}

func (n *nativeReading) note(id, what string) {
	n.violations = append(n.violations, nativeViolation{id, what})
	n.modified[id] = true
}

func eachAccessor(r Request, probe string, targets map[string]bool, visit func(id string, raw any)) {
	v, ok := r.value(probe)
	if !ok {
		return
	}
	m, isMap := v.(map[string]any)
	if !isMap {
		return
	}
	for _, id := range keys(m) {
		targets[id] = true
		visit(id, m[id])
	}
}

func readNativeAccessors(r Request) nativeReading {
	n := nativeReading{modified: map[string]bool{}}
	targets := map[string]bool{}

	eachAccessor(r, "native.tostring", targets, func(id string, raw any) {
		src, have := readToString(raw)
		if !have {
			return
		}
		n.applied++
		switch classifyToString(src, propertyName(id)) {
		case toStringForeign:
			n.note(id, "its serialisation is not that of a built-in function")
		case toStringWrongName:
			n.note(id, "its serialisation names a different function than the one probed")
		}
	})

	eachAccessor(r, "native.ownkeys", targets, func(id string, raw any) {
		sets, isCtor, have := readOwnKeys(raw)
		if !have || len(sets) == 0 {
			return
		}
		n.applied++
		first := sets[0]
		for _, other := range sets[1:] {
			if !sameSet(first, other) {
				n.note(id, "three enumerators of its own keys do not agree")
				break
			}
		}
		if !isCtor && contains(first, "prototype") {
			n.note(id, "it carries an own prototype property, which a built-in that is not a constructor does not have")
		}
	})

	eachAccessor(r, "native.descriptor", targets, func(id string, raw any) {
		onProto, unforgeable, have := readDescriptor(raw)
		if !have || unforgeable {
			return
		}
		n.applied++
		if !onProto {
			n.note(id, "the property does not sit on the interface prototype object where the interface definition puts it")
		}
	})

	eachAccessor(r, "native.receiver", targets, func(id string, raw any) {
		threw, skipped, have := readReceiver(raw)
		if !have || skipped {
			return
		}
		n.applied++
		if !threw {
			n.note(id, "called with a receiver that does not implement the interface it answered instead of throwing a TypeError")
		}
	})

	n.targets = make([]string, 0, len(targets))
	for id := range targets {
		n.targets = append(n.targets, id)
	}
	sort.Strings(n.targets)
	sort.Slice(n.violations, func(i, j int) bool {
		if n.violations[i].target != n.violations[j].target {
			return n.violations[i].target < n.violations[j].target
		}
		return n.violations[i].what < n.violations[j].what
	})
	return n
}
