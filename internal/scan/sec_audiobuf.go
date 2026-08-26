package scan

import "math"

func sectionAudioBuf(r Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}

	gcd := explainedBy(c, keyGetChannelData)
	var t tally

	if !r.ran("audio.views") {
		s.Rows = append(s.Rows, Row{Label: "offline render", Value: "not collected", Note: "the collector did not report audio.views"})
	} else if r.unsupported("audio.views") {
		reason, _ := unsupportedReason(r, "audio.views")
		s.Rows = append(s.Rows, Row{Label: "offline render", Value: "unsupported", Note: clip(reason, 160)})
	} else if status, _ := audioProbeStatus(r, "audio.views"); status == StatusError {
		reason, _ := unsupportedReason(r, "audio.views")
		s.Rows = append(s.Rows, Row{Label: "offline render", Value: "the render attempt did not finish cleanly", Note: "an error is not a claim about what this browser can do: " + clip(reason, 140)})
	} else if v, ok := r.value("audio.views"); !ok {
		s.Rows = append(s.Rows, Row{Label: "offline render", Value: "reported ok but not readable", Note: "the value did not parse as an object this engine understands"})
	} else {
		t = audioViewsAndArithmetic(v, &s.Rows, gcd)
	}

	if !r.ran("audio.repeat") {
		s.Rows = append(s.Rows, Row{Label: "repeat render", Value: "not collected", Note: "the collector did not report audio.repeat"})
	} else if r.unsupported("audio.repeat") {
		reason, _ := unsupportedReason(r, "audio.repeat")
		s.Rows = append(s.Rows, Row{Label: "repeat render", Value: "unsupported", Note: clip(reason, 160)})
	} else if status, _ := audioProbeStatus(r, "audio.repeat"); status == StatusError {
		reason, _ := unsupportedReason(r, "audio.repeat")
		s.Rows = append(s.Rows, Row{Label: "repeat render", Value: "the render attempt did not finish cleanly", Note: "an error is not a claim about what this browser can do: " + clip(reason, 140)})
	} else if v, ok := r.value("audio.repeat"); !ok {
		s.Rows = append(s.Rows, Row{Label: "repeat render", Value: "reported ok but not readable", Note: "the value did not parse as an object this engine understands"})
	} else {
		audioRepeatInformational(v, &s.Rows)
	}

	s.Determination = t.determination()
	switch s.Determination {
	case Inconclusive:
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "nothing was compared", Note: "no exact-agreement invariant was measured"})
	case Contradiction:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "an offline audio render disagreed with itself",
			Note: itoa(float64(t.unexplained)) + " of " + itoa(float64(t.applied)) + " invariant(s) examined did not hold." +
				partlyExplainedNote(t.explained),
		})
	case Instrumented:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "an offline audio render disagreed with itself where a modified accessor accounts for it",
			Note:  itoa(float64(t.explained)) + " of " + itoa(float64(t.applied)) + " invariant(s) examined did not hold. " + explainedConclusion,
		})
	default:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "an offline audio render agreed with itself on every invariant examined",
			Note:  itoa(float64(t.applied)) + " invariant(s) examined",
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
			Note:  "chosen by the probe, short enough to finish inside one collection cycle",
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
			Note:  "read from the buffer the render produced, not from what was requested",
		})
		if haveReqRate && rendRate != reqRate {
			*rows = append(*rows, Row{
				Label: "requested sample rate against rendered",
				Value: "the buffer's sample rate differs from the one requested",
				Note:  "not a finding: only the rendered buffer's own figures feed the checks below",
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
				Note:  "the buffer reports duration " + itoa(rendDuration) + " s but length/sampleRate computes " + itoa(want) + " s",
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
				Note:  gcd.annotate("getChannelData served " + itoa(channelsServed) + " channel(s) but numberOfChannels reports " + itoa(rendChannels)),
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
				Note:  "the context was constructed for " + itoa(reqLength) + " frame(s) but the buffer holds " + itoa(rendLength),
			})
		}
		t.foldPlain(1, bad)
	}

	copyAvailable, haveCopyAvailable := boolean(v, "copyFromChannelAvailable")
	if !haveCopyAvailable {
		*rows = append(*rows, Row{Label: "copyFromChannel against getChannelData", Value: "not compared", Note: "the collector did not report whether copyFromChannel is available"})
	} else if !copyAvailable {
		*rows = append(*rows, Row{Label: "copyFromChannel against getChannelData", Value: "not compared", Note: "the collector reported copyFromChannel as not available in this browser"})
	} else if views, ok := object(v, "views"); ok {
		compared, _ := boolean(views, "compared")
		if compared {
			agree, haveAgree := boolean(views, "agree")
			if haveAgree {
				sampleCount, _ := num(views, "sampleCount")
				bad := 0
				if !agree {
					bad = 1
					differing, _ := num(views, "differingSampleCount")
					maxDiff, _ := num(views, "maxAbsoluteDifference")
					*rows = append(*rows, Row{
						Label: "copyFromChannel against getChannelData",
						Value: "disagree",
						Note: gcd.annotate("of " + itoa(sampleCount) + " sample(s) compared, " + itoa(differing) +
							" differed; largest absolute difference " + itoa(maxDiff)),
					})
				} else {
					*rows = append(*rows, Row{
						Label: "copyFromChannel against getChannelData",
						Value: "agree",
						Note:  itoa(sampleCount) + " sample(s) compared, byte for byte",
					})
				}
				t.fold(1, bad, gcd)
			} else {
				*rows = append(*rows, Row{Label: "copyFromChannel against getChannelData", Value: "not compared", Note: "the comparison result did not parse"})
			}
		} else {
			*rows = append(*rows, Row{Label: "copyFromChannel against getChannelData", Value: "not compared", Note: "the render did not reach a comparison"})
		}
	} else {
		*rows = append(*rows, Row{Label: "copyFromChannel against getChannelData", Value: "not compared", Note: "no comparison was reported"})
	}

	return t
}

func audioRepeatInformational(v any, rows *[]Row) {
	secondCompleted, haveSecond := boolean(v, "secondRenderCompleted")
	if !haveSecond || !secondCompleted {
		*rows = append(*rows, Row{Label: "repeat render", Value: "second render did not complete", Note: "nothing to compare"})
		return
	}
	repeat, ok := object(v, "repeat")
	if !ok {
		*rows = append(*rows, Row{Label: "repeat render", Value: "not compared", Note: "no comparison was reported"})
		return
	}
	compared, _ := boolean(repeat, "compared")
	if !compared {
		*rows = append(*rows, Row{Label: "repeat render", Value: "not compared", Note: "the render did not reach a comparison"})
		return
	}
	agree, haveAgree := boolean(repeat, "agree")
	if !haveAgree {
		*rows = append(*rows, Row{Label: "repeat render", Value: "not compared", Note: "the comparison result did not parse"})
		return
	}
	sampleCount, _ := num(repeat, "sampleCount")
	if agree {
		*rows = append(*rows, Row{
			Label: "repeat render agreement",
			Value: "two renders of one identical graph produced identical output",
			Note:  itoa(sampleCount) + " sample(s) compared. Not scored either way: this project found no specification text requiring offline rendering to be deterministic.",
		})
		return
	}
	differing, _ := num(repeat, "differingSampleCount")
	maxDiff, _ := num(repeat, "maxAbsoluteDifference")
	*rows = append(*rows, Row{
		Label: "repeat render agreement",
		Value: "two renders of one identical graph produced different output",
		Note: "of " + itoa(sampleCount) + " sample(s) compared, " + itoa(differing) + " differed, largest absolute difference " +
			itoa(maxDiff) + ". Not scored either way: this project found no specification text requiring offline rendering to be deterministic.",
	})
}
