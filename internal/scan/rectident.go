package scan

import (
	_ "time/tzdata"
)

const rectTolerance = 1e-4

type rectPoint struct {
	x, y, left, top, right, bottom, width, height                                 float64
	haveX, haveY, haveLeft, haveTop, haveRight, haveBottom, haveWidth, haveHeight bool
}

type rectIdentities struct {
	base, twin, shifted, restored                 rectPoint
	haveBase, haveTwin, haveShifted, haveRestored bool

	shiftPx     float64
	haveShiftPx bool
}

func readRectPoint(v any, path ...string) (rectPoint, bool) {
	obj, ok := object(v, path...)
	if !ok {
		return rectPoint{}, false
	}
	var p rectPoint
	p.x, p.haveX = num(obj, "x")
	p.y, p.haveY = num(obj, "y")
	p.left, p.haveLeft = num(obj, "left")
	p.top, p.haveTop = num(obj, "top")
	p.right, p.haveRight = num(obj, "right")
	p.bottom, p.haveBottom = num(obj, "bottom")
	p.width, p.haveWidth = num(obj, "width")
	p.height, p.haveHeight = num(obj, "height")
	return p, true
}

func readRectIdentities(v any) rectIdentities {
	var ri rectIdentities
	ri.base, ri.haveBase = readRectPoint(v, "base")
	ri.twin, ri.haveTwin = readRectPoint(v, "twin")
	ri.shifted, ri.haveShifted = readRectPoint(v, "shifted")
	ri.restored, ri.haveRestored = readRectPoint(v, "restored")
	ri.shiftPx, ri.haveShiftPx = num(v, "shiftPx")
	return ri
}

func (ri rectIdentities) identities() [][]identity {
	var groups [][]identity
	for _, rp := range []struct {
		label string
		p     rectPoint
		have  bool
	}{
		{"base rect", ri.base, ri.haveBase},
		{"twin rect", ri.twin, ri.haveTwin},
		{"shifted rect", ri.shifted, ri.haveShifted},
		{"restored rect", ri.restored, ri.haveRestored},
	} {
		if !rp.have {
			continue
		}
		groups = append(groups, selfConsistency(rp.label, rp.p))
	}
	if ri.haveBase && ri.haveTwin {
		groups = append(groups, equalElements(ri.base, ri.twin))
	}
	if ri.haveBase && ri.haveShifted && ri.haveRestored && ri.haveShiftPx {
		groups = append(groups, translation(ri.base, ri.shifted, ri.restored, ri.shiftPx))
	}
	return groups
}

func selfConsistency(label string, p rectPoint) []identity {
	if p.haveWidth && p.width < 0 {
		return nil
	}
	if p.haveHeight && p.height < 0 {
		return nil
	}
	var ids []identity
	if p.haveRight && p.haveLeft && p.haveWidth {
		ids = append(ids, identity{label + ": right minus left equals width", p.right - p.left, p.width})
	}
	if p.haveBottom && p.haveTop && p.haveHeight {
		ids = append(ids, identity{label + ": bottom minus top equals height", p.bottom - p.top, p.height})
	}
	if p.haveX && p.haveLeft && p.haveWidth {
		ids = append(ids, identity{label + ": x equals left", p.x, p.left})
	}
	if p.haveY && p.haveTop && p.haveHeight {
		ids = append(ids, identity{label + ": y equals top", p.y, p.top})
	}
	return ids
}

func equalElements(base, twin rectPoint) []identity {
	var ids []identity
	if base.haveWidth && twin.haveWidth {
		ids = append(ids, identity{"two identically styled elements report equal width", base.width, twin.width})
	}
	if base.haveHeight && twin.haveHeight {
		ids = append(ids, identity{"two identically styled elements report equal height", base.height, twin.height})
	}
	return ids
}

func translation(base, shifted, restored rectPoint, shiftPx float64) []identity {
	var ids []identity
	if base.haveLeft && shifted.haveLeft {
		ids = append(ids, identity{"translating right by an exact amount moves left by that amount", shifted.left - base.left, shiftPx})
	}
	if base.haveRight && shifted.haveRight {
		ids = append(ids, identity{"translating right by an exact amount moves right by that amount", shifted.right - base.right, shiftPx})
	}
	if base.haveTop && shifted.haveTop {
		ids = append(ids, identity{"a horizontal translation leaves top unchanged", shifted.top, base.top})
	}
	if base.haveBottom && shifted.haveBottom {
		ids = append(ids, identity{"a horizontal translation leaves bottom unchanged", shifted.bottom, base.bottom})
	}
	if base.haveWidth && shifted.haveWidth {
		ids = append(ids, identity{"a horizontal translation leaves width unchanged", shifted.width, base.width})
	}
	if base.haveHeight && shifted.haveHeight {
		ids = append(ids, identity{"a horizontal translation leaves height unchanged", shifted.height, base.height})
	}
	if base.haveLeft && restored.haveLeft {
		ids = append(ids, identity{"translating back returns left to its original value", restored.left, base.left})
	}
	if base.haveTop && restored.haveTop {
		ids = append(ids, identity{"translating back returns top to its original value", restored.top, base.top})
	}
	if base.haveWidth && restored.haveWidth {
		ids = append(ids, identity{"translating back returns width to its original value", restored.width, base.width})
	}
	if base.haveHeight && restored.haveHeight {
		ids = append(ids, identity{"translating back returns height to its original value", restored.height, base.height})
	}
	return ids
}
