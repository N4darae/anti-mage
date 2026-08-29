package scan

const (
	keyMeasureText       = "CanvasRenderingContext2D.prototype.measureText"
	keyGetChannelData    = "AudioBuffer.prototype.getChannelData"
	keyMatchMedia        = "window.matchMedia"
	keyGetComputedStyle  = "window.getComputedStyle"
	keyDevicePixelRatio  = "window.devicePixelRatio"
	keyBoundingRect      = "Element.prototype.getBoundingClientRect"
	keyScreenWidth       = "screen.width"
	keyScreenHeight      = "screen.height"
	keyScreenAvailWidth  = "screen.availWidth"
	keyScreenAvailHeight = "screen.availHeight"
)

var mappedAccessorKeys = []string{
	keyMeasureText, keyGetChannelData, keyMatchMedia,
	keyGetComputedStyle, keyDevicePixelRatio, keyBoundingRect,
	keyScreenWidth, keyScreenHeight, keyScreenAvailWidth, keyScreenAvailHeight,
}

type explanation struct {
	modified []string

	structural string
}

func explainedBy(c claim, keys ...string) explanation {
	var e explanation
	for _, k := range keys {
		if c.natives.reportsModified(k) {
			e.modified = append(e.modified, k)
		}
	}
	return e
}

func explainedStructurally(note string) explanation {
	return explanation{structural: note}
}

func (e explanation) downgrades() bool { return len(e.modified) > 0 || e.structural != "" }

type tally struct {
	applied int

	unexplained int

	explained int
}

func (t *tally) fold(applied, failed int, e explanation) {
	t.applied += applied
	if failed == 0 {
		return
	}
	if e.downgrades() {
		t.explained += failed
		return
	}
	t.unexplained += failed
}

func (t *tally) foldPlain(applied, failed int) { t.fold(applied, failed, explanation{}) }

func (t tally) failed() int { return t.unexplained + t.explained }

func (t tally) determination() Determination {
	switch {
	case t.applied == 0:
		return Inconclusive
	case t.unexplained > 0:
		return Contradiction
	case t.explained > 0:
		return Instrumented
	}
	return Consistent
}
