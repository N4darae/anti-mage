package scan

func sectionWebGPU(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Inconclusive}

	adapter, ok := r.value("gpu.adapter")
	if !ok {
		return webgpuNoWeight(s, "the adapter interface was not collected", "this reading compares a named device against the adapters granted for it; without the adapters it carries no weight either way")
	}

	if secure, known := boolean(adapter, "secureContext"); known && !secure {
		return webgpuNoWeight(s, "the page was not delivered to a secure context", "the interface is gated there, so nothing follows from its answer")
	}
	if present, known := boolean(adapter, "present"); !known || !present {
		return webgpuNoWeight(s, "this browser does not expose the interface", "a browser is never read as lacking a feature")
	}

	renderer, named := readGPURenderer(r)
	if !named {
		return webgpuNoWeight(s, "no graphics device was named", "with no device named there is nothing for an empty adapter to disagree with")
	}
	s.add("graphics device, as reported", clip(renderer, 90))

	if isSoftwareRasteriser(renderer) {
		return webgpuNoWeight(s, "the device named draws in software", "an environment already drawing in software is not contradicted by having no hardware adapter")
	}

	granted, missing := 0, 0
	for _, path := range webgpuHardwarePaths {
		got, known := boolean(adapter, "variants", path, "adapter")
		if !known {
			return webgpuNoWeight(s, "one of the adapter requests was not reported", "nothing was compared")
		}
		if got {
			granted++
		} else {
			missing++
		}
		s.add("adapter for the "+webgpuPathLabel(path)+" request", answerOrAbsent(got, true))
	}

	if granted > 0 {
		s.Determination = Consistent
		s.conclude("the newer graphics interface reaches a device too")
		return s
	}

	if fallback, known := boolean(adapter, "variants", "fallback", "adapter"); known && fallback {
		s.add("adapter for the software fallback request", answerOrAbsent(true, true))
		s.Determination, s.weight = Contradiction, weightOnlyDeliberate
		s.conclude("the newer graphics interface refuses every device it has and then grants one")
		return s
	}

	s.Determination = Instrumented
	s.conclude("one graphics interface names a hardware device and the other reaches no device at all")
	return s
}

func webgpuNoWeight(s Section, value, note string) Section {
	s.Determination = Unverified
	s.conclude(value)
	return s
}

func webgpuPathLabel(path string) string {
	switch path {
	case "highPerformance":
		return "high performance"
	case "lowPower":
		return "low power"
	}
	return "unqualified"
}
