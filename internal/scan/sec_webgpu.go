package scan

import "strings"

var softwareRasteriserMarkers = []string{
	"swiftshader",
	"llvmpipe",
	"basic render",
	"software adapter",
	"software rasterizer",
	"microsoft basic",
}

var webgpuHardwarePaths = []string{"default", "highPerformance", "lowPower"}

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

	renderer, named := hwDecodeRenderer(r)
	if !named {
		return webgpuNoWeight(s, "no graphics device was named", "with no device named there is nothing for an empty adapter to disagree with")
	}
	s.Rows = append(s.Rows, Row{Label: "graphics device, as reported", Value: clip(renderer, 90), Note: anomalyNote})

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
		s.Rows = append(s.Rows, Row{Label: "adapter for the " + webgpuPathLabel(path) + " request", Value: answerOrAbsent(got, true), Note: anomalyNote})
	}

	if granted > 0 {
		s.Determination = Consistent
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "the newer graphics interface reaches a device too", Note: anomalyNote})
		return s
	}

	if fallback, known := boolean(adapter, "variants", "fallback", "adapter"); known && fallback {
		s.Rows = append(s.Rows, Row{Label: "adapter for the software fallback request", Value: answerOrAbsent(true, true), Note: anomalyNote})
		s.Determination, s.weight = Contradiction, weightOnlyDeliberate
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the newer graphics interface refuses every device it has and then grants one",
			Note:  anomalyNote,
		})
		return s
	}

	s.Determination = Instrumented
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "one graphics interface names a hardware device and the other reaches no device at all",
		Note:  anomalyNote,
	})
	return s
}

func webgpuNoWeight(s Section, value, note string) Section {
	s.Determination = Unverified
	s.Rows = append(s.Rows, Row{Label: "conclusion", Value: value, Note: anomalyNote})
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

func isSoftwareRasteriser(renderer string) bool {
	lower := strings.ToLower(renderer)
	for _, marker := range softwareRasteriserMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
