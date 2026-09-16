package scan

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/N4darae/anti-mage/reference"
)

var softwareRasteriserMarkers = []string{
	"swiftshader",
	"llvmpipe",
	"basic render",
	"software adapter",
	"software rasterizer",
	"microsoft basic",
}

var vendorsNamingNoSilicon = []string{"google", "mesa", "microsoft", "swiftshader", "warp"}

var webgpuHardwarePaths = []string{"default", "highPerformance", "lowPower"}

var angleDeviceIdentifier = regexp.MustCompile(`\(0x([0-9A-Fa-f]{8})\)`)

const swiftshaderDeviceIdentifier = 0xC0DE

func readGPURenderer(r Request) (string, bool) {
	raw, ok := r.value("gpu.renderer")
	if !ok {
		return "", false
	}
	name := firstString(raw, []string{"unmaskedRenderer"}, []string{"renderer"})
	return name, name != ""
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

func namesASubstituteDevice(renderer string) bool {
	return strings.Contains(strings.ToLower(renderer), "or similar")
}

func adapterAcrossEveryRequest(adapter any) (string, string, bool) {
	var vendor, architecture string
	for i, path := range webgpuHardwarePaths {
		granted, known := boolean(adapter, "variants", path, "adapter")
		if !known || !granted {
			return "", "", false
		}
		v, haveVendor := str(adapter, "variants", path, "vendor")
		a, haveArchitecture := str(adapter, "variants", path, "architecture")
		if !haveVendor || !haveArchitecture {
			return "", "", false
		}
		v, a = strings.ToLower(strings.TrimSpace(v)), strings.ToLower(strings.TrimSpace(a))
		if v == "" || a == "" {
			return "", "", false
		}
		if i == 0 {
			vendor, architecture = v, a
			continue
		}
		if v != vendor || a != architecture {
			return "", "", false
		}
	}
	return vendor, architecture, true
}

type gpuArchReading struct {
	renderer string

	device uint32

	vendor string

	reported string

	tabulated string

	agrees bool
}

func readGPUArch(r Request) (gpuArchReading, string, bool) {
	var reading gpuArchReading

	adapter, ok := r.value("gpu.adapter")
	if !ok {
		return reading, "the adapter interface was not collected", false
	}
	if secure, known := boolean(adapter, "secureContext"); known && !secure {
		return reading, "the page was not delivered to a secure context, where the interface is gated", false
	}
	if present, known := boolean(adapter, "present"); !known || !present {
		return reading, "this browser does not expose the interface", false
	}

	renderer, named := readGPURenderer(r)
	if !named {
		return reading, "no graphics device was named", false
	}
	reading.renderer = renderer

	if isSoftwareRasteriser(renderer) {
		return reading, "the device named draws in software, which names no silicon", false
	}
	m := angleDeviceIdentifier.FindStringSubmatch(renderer)
	if m == nil {
		if namesASubstituteDevice(renderer) {
			return reading, "this browser said the device it named stands in for the one present", false
		}
		return reading, "the device named carries no identifier for this reading to place", false
	}
	identifier, err := strconv.ParseUint(m[1], 16, 32)
	if err != nil {
		return reading, "the identifier the device named carries could not be read", false
	}
	reading.device = uint32(identifier)
	if reading.device == swiftshaderDeviceIdentifier {
		return reading, "the identifier the device named carries is the one a software renderer uses", false
	}

	vendor, architecture, agreed := adapterAcrossEveryRequest(adapter)
	if !agreed {
		return reading, "the requests this reading compares did not all reach one device", false
	}
	reading.vendor, reading.reported = vendor, architecture

	if contains(vendorsNamingNoSilicon, vendor) {
		return reading, "the interface names a renderer rather than a manufacturer of silicon", false
	}

	table, tabulated := reference.GPUArchitectures[vendor]
	if !tabulated {
		return reading, "this project's table does not carry the manufacturer reported", false
	}
	if !table.Verified(architecture) {
		return reading, "the table entry for the generation reported has not been confirmed on a real system of that configuration", false
	}

	placed, ok := table.Architecture(reading.device)
	if !ok {
		return reading, "this project's table does not place the identifier the device named carries", false
	}
	reading.tabulated = placed
	reading.agrees = placed == architecture
	return reading, "", true
}

func gpuArchContradicted(r Request) bool {
	reading, _, settled := readGPUArch(r)
	return settled && !reading.agrees
}
