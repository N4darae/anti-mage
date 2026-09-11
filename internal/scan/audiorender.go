package scan

import (
	"math"
	_ "time/tzdata"
)

const audioDurationEpsilonSeconds = 1e-6

const (
	audioScaleMaxRelativeResidual = 2.384185791015625e-07

	audioScaleMinSamples = 32
)

const audioUniformScaleNote = "same signal at two scales; read as an environment reporting itself modified"

type audioRender struct {
	reqRate, reqChannels, reqLength             float64
	haveReqRate, haveReqChannels, haveReqLength bool

	rendRate, rendChannels, rendLength, rendDuration                 float64
	haveRendRate, haveRendChannels, haveRendLength, haveRendDuration bool

	channelsServed     float64
	haveChannelsServed bool

	copyAvailable, haveCopyAvailable bool

	views any
}

func readAudioRender(v any) audioRender {
	var a audioRender
	a.reqRate, a.haveReqRate = num(v, "requested", "sampleRateHz")
	a.reqChannels, a.haveReqChannels = num(v, "requested", "numberOfChannels")
	a.reqLength, a.haveReqLength = num(v, "requested", "lengthFrames")

	a.rendRate, a.haveRendRate = num(v, "rendered", "sampleRateHz")
	a.rendChannels, a.haveRendChannels = num(v, "rendered", "numberOfChannels")
	a.rendLength, a.haveRendLength = num(v, "rendered", "lengthFrames")
	a.rendDuration, a.haveRendDuration = num(v, "rendered", "durationSeconds")

	a.channelsServed, a.haveChannelsServed = num(v, "channelsServed")
	a.copyAvailable, a.haveCopyAvailable = boolean(v, "copyFromChannelAvailable")
	a.views = v
	return a
}

func (a audioRender) haveRequested() bool {
	return a.haveReqRate && a.haveReqChannels && a.haveReqLength
}

func (a audioRender) haveRendered() bool {
	return a.haveRendRate && a.haveRendChannels && a.haveRendLength && a.haveRendDuration
}

func (a audioRender) durationFromLength() (float64, bool) {
	if !a.haveRendRate || !a.haveRendLength || !a.haveRendDuration || a.rendRate <= 0 {
		return 0, false
	}
	return a.rendLength / a.rendRate, true
}

func (a audioRender) viewsAgreement() (agree, compared bool) {
	if !a.haveCopyAvailable || !a.copyAvailable {
		return false, false
	}
	return agreementAt(a.views, "views")
}

func agreementAt(v any, path ...string) (agree, compared bool) {
	obj, ok := object(v, path...)
	if !ok {
		return false, false
	}
	if done, _ := boolean(obj, "compared"); !done {
		return false, false
	}
	agree, compared = boolean(obj, "agree")
	return agree, compared
}

func closeEnough(a, b, epsilon float64) bool {
	return math.Abs(a-b) <= epsilon
}

func audioScaleShape(v any) (explanation, string, bool) {
	scale, ok := object(v, "views", "scale")
	if !ok {
		return explanation{}, "", false
	}
	fitted, haveFitted := boolean(scale, "fitted")
	if !haveFitted {
		return explanation{}, "", false
	}
	if !fitted {
		return explanation{}, "no single factor was fitted", true
	}

	_, haveFactor := num(scale, "factor")
	compared, haveCompared := num(scale, "comparedSamples")
	residual, haveResidual := num(scale, "maxRelativeResidual")
	zerosAltered, haveZeros := num(scale, "zerosAltered")
	if !haveFactor || !haveCompared || !haveResidual || !haveZeros {
		return explanation{}, "not characterised", true
	}

	switch {
	case compared < audioScaleMinSamples:
		return explanation{}, "too few samples to characterise", true
	case zerosAltered > 0, residual > audioScaleMaxRelativeResidual:
		return explanation{}, "not one uniform scaling", true
	}
	return explainedStructurally(audioUniformScaleNote), "one uniform scaling of the whole channel", true
}
