package scan

func sectionAudioBuf(r Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}

	gcd := explainedBy(c, keyGetChannelData)
	var t tally

	if v, outcome := r.read("audio.views"); outcome != probeReadable {
		s.add("offline render", audioRenderNote(outcome))
	} else {
		t = s.addRenderReading(readAudioRender(v), gcd)
	}

	if v, outcome := r.read("audio.repeat"); outcome != probeReadable {
		s.add("repeat render", audioRenderNote(outcome))
	} else {
		s.addRepeatReading(v)
	}

	s.settle(t, conclusions{
		inconclusive:  "nothing was compared",
		contradiction: "an offline audio render disagreed with itself",
		instrumented:  "an offline audio render disagreed with itself where a modification this environment carries accounts for it",
		consistent:    "an offline audio render agreed with itself on every invariant examined",
	})
	return s
}

func audioRenderNote(outcome probeOutcome) string {
	switch outcome {
	case probeAbsent:
		return "not collected"
	case probeUnsupported:
		return "unsupported"
	case probeErrored:
		return "the render attempt did not finish cleanly"
	}
	return "reported ok but not readable"
}

func (s *Section) addRenderReading(a audioRender, gcd explanation) (t tally) {
	if a.haveRequested() {
		s.add("requested render", itoa(a.reqRate)+" Hz, "+itoa(a.reqChannels)+" channel(s), "+itoa(a.reqLength)+" frame(s)")
	}
	if a.haveRendered() {
		s.add("rendered buffer", itoa(a.rendRate)+" Hz, "+itoa(a.rendChannels)+" channel(s), "+itoa(a.rendLength)+" frame(s), duration "+itoa(a.rendDuration)+" s")
		if a.haveReqRate && a.rendRate != a.reqRate {
			s.add("requested sample rate against rendered", "the buffer's sample rate differs from the one requested")
		}
	}

	if want, ok := a.durationFromLength(); ok {
		bad := 0
		if !closeEnough(want, a.rendDuration, audioDurationEpsilonSeconds) {
			bad = 1
			s.add("duration against length and sample rate", "disagree")
		}
		t.foldPlain(1, bad)
	}

	if a.haveChannelsServed && a.haveRendChannels {
		bad := 0
		if a.channelsServed != a.rendChannels {
			bad = 1
			s.add("channels served against numberOfChannels", "disagree")
		}
		t.fold(1, bad, gcd)
	}

	if a.haveReqLength && a.haveRendLength {
		bad := 0
		if a.reqLength != a.rendLength {
			bad = 1
			s.add("rendered length against construction", "disagree")
		}
		t.foldPlain(1, bad)
	}

	agree, compared := a.viewsAgreement()
	switch {
	case !compared:
		s.add("copyFromChannel against getChannelData", "not compared")
	case agree:
		s.add("copyFromChannel against getChannelData", "agree")
		t.fold(1, 0, gcd)
	default:
		s.add("copyFromChannel against getChannelData", "disagree")
		views := gcd
		if scaled, shape, characterised := audioScaleShape(a.views); characterised {
			s.add("the shape of the disagreement", shape)
			if scaled.downgrades() {
				views = scaled
			}
		}
		t.fold(1, 1, views)
	}
	return t
}

func (s *Section) addRepeatReading(v any) {
	if completed, have := boolean(v, "secondRenderCompleted"); !have || !completed {
		s.add("repeat render", "second render did not complete")
		return
	}
	agree, compared := agreementAt(v, "repeat")
	if !compared {
		s.add("repeat render", "not compared")
		return
	}
	if agree {
		s.add("repeat render agreement", "two renders of one identical graph produced identical output")
		return
	}
	s.add("repeat render agreement", "two renders of one identical graph produced different output")
}
