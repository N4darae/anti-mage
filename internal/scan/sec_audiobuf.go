package scan

import "math"

func sectionAudioBuf(r Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}

	gcd := explainedBy(c, keyGetChannelData)
	var t tally

	if !r.ran("audio.views") {
		s.Rows = append(s.Rows, Row{Label: "offline render", Value: "not collected", Note: anomalyNote})
	} else if r.unsupported("audio.views") {
		s.Rows = append(s.Rows, Row{Label: "offline render", Value: "unsupported", Note: anomalyNote})
	} else if status, _ := audioProbeStatus(r, "audio.views"); status == StatusError {
		s.Rows = append(s.Rows, Row{Label: "offline render", Value: "the render attempt did not finish cleanly", Note: anomalyNote})
	} else if v, ok := r.value("audio.views"); !ok {
		s.Rows = append(s.Rows, Row{Label: "offline render", Value: "reported ok but not readable", Note: anomalyNote})
	} else {
		t = audioViewsAndArithmetic(v, &s.Rows, gcd)
	}

	if !r.ran("audio.repeat") {
		s.Rows = append(s.Rows, Row{Label: "repeat render", Value: "not collected", Note: anomalyNote})
	} else if r.unsupported("audio.repeat") {
		s.Rows = append(s.Rows, Row{Label: "repeat render", Value: "unsupported", Note: anomalyNote})
	} else if status, _ := audioProbeStatus(r, "audio.repeat"); status == StatusError {
		s.Rows = append(s.Rows, Row{Label: "repeat render", Value: "the render attempt did not finish cleanly", Note: anomalyNote})
	} else if v, ok := r.value("audio.repeat"); !ok {
		s.Rows = append(s.Rows, Row{Label: "repeat render", Value: "reported ok but not readable", Note: anomalyNote})
	} else {
		audioRepeatInformational(v, &s.Rows)
	}

	s.Determination = t.determination()
	switch s.Determination {
	case Inconclusive:
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "nothing was compared", Note: anomalyNote})
	case Contradiction:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "an offline audio render disagreed with itself",
			Note:  anomalyNote,
		})
	case Instrumented:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "an offline audio render disagreed with itself where a modification this environment carries accounts for it",
			Note:  anomalyNote,
		})
	default:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "an offline audio render agreed with itself on every invariant examined",
			Note:  anomalyNote,
		})
	}
	return s
}

func audioProbeStatus(r Request, id string) (string, bool) {
	p, ok := r.Probes[id]
	if !ok {
		return "", false
	}
	return p.Status, true
}

const audioDurationEpsilonSeconds = 1e-6

func audioViewsAndArithmetic(v any, rows *[]Row, gcd explanation) (t tally) {
	reqRate, haveReqRate := num(v, "requested", "sampleRateHz")
	reqChannels, haveReqChannels := num(v, "requested", "numberOfChannels")
	reqLength, haveReqLength := num(v, "requested", "lengthFrames")
	if haveReqRate && haveReqChannels && haveReqLength {
		*rows = append(*rows, Row{
			Label: "requested render",
			Value: itoa(reqRate) + " Hz, " + itoa(reqChannels) + " channel(s), " + itoa(reqLength) + " frame(s)",
			Note:  anomalyNote,
		})
	}

	rendRate, haveRendRate := num(v, "rendered", "sampleRateHz")
	rendChannels, haveRendChannels := num(v, "rendered", "numberOfChannels")
	rendLength, haveRendLength := num(v, "rendered", "lengthFrames")
	rendDuration, haveRendDuration := num(v, "rendered", "durationSeconds")
	if haveRendRate && haveRendChannels && haveRendLength && haveRendDuration {
		*rows = append(*rows, Row{
			Label: "rendered buffer",
			Value: itoa(rendRate) + " Hz, " + itoa(rendChannels) + " channel(s), " + itoa(rendLength) + " frame(s), duration " + itoa(rendDuration) + " s",
			Note:  anomalyNote,
		})
		if haveReqRate && rendRate != reqRate {
			*rows = append(*rows, Row{
				Label: "requested sample rate against rendered",
				Value: "the buffer's sample rate differs from the one requested",
				Note:  anomalyNote,
			})
		}
	}

	if haveRendRate && haveRendLength && haveRendDuration && rendRate > 0 {
		want := rendLength / rendRate
		bad := 0
		if math.Abs(want-rendDuration) > audioDurationEpsilonSeconds {
			bad = 1
			*rows = append(*rows, Row{
				Label: "duration against length and sample rate",
				Value: "disagree",
				Note:  anomalyNote,
			})
		}
		t.foldPlain(1, bad)
	}

	channelsServed, haveChannelsServed := num(v, "channelsServed")
	if haveChannelsServed && haveRendChannels {
		bad := 0
		if channelsServed != rendChannels {
			bad = 1
			*rows = append(*rows, Row{
				Label: "channels served against numberOfChannels",
				Value: "disagree",
				Note:  anomalyNote,
			})
		}
		t.fold(1, bad, gcd)
	}

	if haveReqLength && haveRendLength {
		bad := 0
		if reqLength != rendLength {
			bad = 1
			*rows = append(*rows, Row{
				Label: "rendered length against construction",
				Value: "disagree",
				Note:  anomalyNote,
			})
		}
		t.foldPlain(1, bad)
	}

	copyAvailable, haveCopyAvailable := boolean(v, "copyFromChannelAvailable")
	if !haveCopyAvailable {
		*rows = append(*rows, Row{Label: "copyFromChannel against getChannelData", Value: "not compared", Note: anomalyNote})
	} else if !copyAvailable {
		*rows = append(*rows, Row{Label: "copyFromChannel against getChannelData", Value: "not compared", Note: anomalyNote})
	} else if views, ok := object(v, "views"); ok {
		compared, _ := boolean(views, "compared")
		if compared {
			agree, haveAgree := boolean(views, "agree")
			if haveAgree {
				viewsExplanation := gcd
				bad := 0
				if !agree {
					bad = 1
					*rows = append(*rows, Row{
						Label: "copyFromChannel against getChannelData",
						Value: "disagree",
						Note:  anomalyNote,
					})
					if scaled, row, reported := audioUniformScale(views); reported {
						*rows = append(*rows, row)
						if scaled.downgrades() {
							viewsExplanation = scaled
						}
					}
				} else {
					*rows = append(*rows, Row{
						Label: "copyFromChannel against getChannelData",
						Value: "agree",
						Note:  anomalyNote,
					})
				}
				t.fold(1, bad, viewsExplanation)
			} else {
				*rows = append(*rows, Row{Label: "copyFromChannel against getChannelData", Value: "not compared", Note: anomalyNote})
			}
		} else {
			*rows = append(*rows, Row{Label: "copyFromChannel against getChannelData", Value: "not compared", Note: anomalyNote})
		}
	} else {
		*rows = append(*rows, Row{Label: "copyFromChannel against getChannelData", Value: "not compared", Note: anomalyNote})
	}

	return t
}

func audioRepeatInformational(v any, rows *[]Row) {
	secondCompleted, haveSecond := boolean(v, "secondRenderCompleted")
	if !haveSecond || !secondCompleted {
		*rows = append(*rows, Row{Label: "repeat render", Value: "second render did not complete", Note: anomalyNote})
		return
	}
	repeat, ok := object(v, "repeat")
	if !ok {
		*rows = append(*rows, Row{Label: "repeat render", Value: "not compared", Note: anomalyNote})
		return
	}
	compared, _ := boolean(repeat, "compared")
	if !compared {
		*rows = append(*rows, Row{Label: "repeat render", Value: "not compared", Note: anomalyNote})
		return
	}
	agree, haveAgree := boolean(repeat, "agree")
	if !haveAgree {
		*rows = append(*rows, Row{Label: "repeat render", Value: "not compared", Note: anomalyNote})
		return
	}
	if agree {
		*rows = append(*rows, Row{
			Label: "repeat render agreement",
			Value: "two renders of one identical graph produced identical output",
			Note:  anomalyNote,
		})
		return
	}
	*rows = append(*rows, Row{
		Label: "repeat render agreement",
		Value: "two renders of one identical graph produced different output",
		Note:  anomalyNote,
	})
}

const (
	audioScaleMaxRelativeResidual = 2.384185791015625e-07

	audioScaleMinSamples = 32
)

func audioUniformScale(views any) (explanation, Row, bool) {
	scale, ok := object(views, "scale")
	if !ok {
		return explanation{}, Row{}, false
	}
	fitted, haveFitted := boolean(scale, "fitted")
	if !haveFitted {
		return explanation{}, Row{}, false
	}
	if !fitted {
		return explanation{}, Row{
			Label: "the shape of the disagreement",
			Value: "no single factor was fitted",
			Note:  anomalyNote,
		}, true
	}

	_, haveFactor := num(scale, "factor")
	compared, haveCompared := num(scale, "comparedSamples")
	residual, haveResidual := num(scale, "maxRelativeResidual")
	zerosAltered, haveZeros := num(scale, "zerosAltered")
	if !haveFactor || !haveCompared || !haveResidual || !haveZeros {
		return explanation{}, Row{
			Label: "the shape of the disagreement",
			Value: "not characterised",
			Note:  anomalyNote,
		}, true
	}

	switch {
	case compared < audioScaleMinSamples:
		return explanation{}, Row{
			Label: "the shape of the disagreement",
			Value: "too few samples to characterise",
			Note:  anomalyNote,
		}, true
	case zerosAltered > 0:
		return explanation{}, Row{
			Label: "the shape of the disagreement",
			Value: "not one uniform scaling",
			Note:  anomalyNote,
		}, true
	case residual > audioScaleMaxRelativeResidual:
		return explanation{}, Row{
			Label: "the shape of the disagreement",
			Value: "not one uniform scaling",
			Note:  anomalyNote,
		}, true
	}

	return explainedStructurally(audioUniformScaleNote), Row{
		Label: "the shape of the disagreement",
		Value: "one uniform scaling of the whole channel",
		Note:  anomalyNote,
	}, true
}

const audioUniformScaleNote = "same signal at two scales; read as an environment reporting itself modified"
