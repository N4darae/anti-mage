package scan

import "strings"

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

func (e explanation) note() string {
	if e.structural != "" {
		return e.structural
	}
	if len(e.modified) == 0 {
		return ""
	}
	return "this reading arrived through " + strings.Join(e.modified, " and ") +
		", which this environment itself reports is not a built-in, so the disagreement is explained by a modification the environment declares rather than counted against it a second time"
}

func (e explanation) annotate(note string) string {
	n := e.note()
	switch {
	case n == "":
		return note
	case note == "":
		return n
	}
	return note + "; " + n
}

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

const explainedConclusion = "every disagreement found is accounted for by a modification this environment carries openly, either an accessor it reports as not a built-in or a transform that left the structure of the reading intact, so this section reports the environment as modified rather than counting a disagreement already accounted for"

func partlyExplainedNote(explained int) string {
	if explained == 0 {
		return ""
	}
	return " A further " + itoa(float64(explained)) +
		" did not hold but read their evidence through an accessor this environment reports as not a built-in, and are not part of this conclusion; the rest are explained by nothing this environment declared."
}
