package scan

import "sort"

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

func readNativeAccessors(r Request) nativeReading {
	n := nativeReading{modified: map[string]bool{}}
	targets := map[string]bool{}

	note := func(id, what string) {
		n.violations = append(n.violations, nativeViolation{id, what})
		n.modified[id] = true
	}

	if v, ok := r.value("native.tostring"); ok {
		if m, ok := v.(map[string]any); ok {
			for _, id := range keys(m) {
				targets[id] = true
				src, have := readToString(m[id])
				if !have {
					continue
				}
				n.applied++
				switch classifyToString(src, propertyName(id)) {
				case toStringForeign:
					note(id, "its serialisation is not that of a built-in function")
				case toStringWrongName:
					note(id, "its serialisation names a different function than the one probed")
				}
			}
		}
	}

	if v, ok := r.value("native.ownkeys"); ok {
		if m, ok := v.(map[string]any); ok {
			for _, id := range keys(m) {
				targets[id] = true
				sets, isCtor, have := readOwnKeys(m[id])
				if !have || len(sets) == 0 {
					continue
				}
				n.applied++
				first := sets[0]
				for _, other := range sets[1:] {
					if !sameSet(first, other) {
						note(id, "three enumerators of its own keys do not agree")
						break
					}
				}
				if !isCtor && contains(first, "prototype") {
					note(id, "it carries an own prototype property, which a built-in that is not a constructor does not have")
				}
			}
		}
	}

	if v, ok := r.value("native.descriptor"); ok {
		if m, ok := v.(map[string]any); ok {
			for _, id := range keys(m) {
				targets[id] = true
				onProto, unforgeable, have := readDescriptor(m[id])
				if !have || unforgeable {
					continue
				}
				n.applied++
				if !onProto {
					note(id, "the property does not sit on the interface prototype object where the interface definition puts it")
				}
			}
		}
	}

	if v, ok := r.value("native.receiver"); ok {
		if m, ok := v.(map[string]any); ok {
			for _, id := range keys(m) {
				targets[id] = true
				threw, skipped, have := readReceiver(m[id])
				if !have || skipped {
					continue
				}
				n.applied++
				if !threw {
					note(id, "called with a receiver that does not implement the interface it answered instead of throwing a TypeError")
				}
			}
		}
	}

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
