package scan

import "strings"

type osFamily string

const (
	osUnknown osFamily = ""
	osWindows osFamily = "windows"
	osMac     osFamily = "macos"
	osLinux   osFamily = "linux"
	osAndroid osFamily = "android"
	osIOS     osFamily = "ios"
	osChrome  osFamily = "chromeos"
)

type claim struct {
	Family osFamily

	Agreed bool

	UserAgent     string
	NavPlatform   string
	UADataPlat    string
	uaFamily      osFamily
	platFamily    osFamily
	uaDataFamily  osFamily
	surfacesKnown int

	natives nativeReading
}

func familyFromUA(ua string) osFamily {
	s := strings.ToLower(ua)
	switch {
	case s == "":
		return osUnknown
	case strings.Contains(s, "android"):
		return osAndroid
	case strings.Contains(s, "iphone"), strings.Contains(s, "ipad"), strings.Contains(s, "ipod"):
		return osIOS
	case strings.Contains(s, "cros"):
		return osChrome
	case strings.Contains(s, "windows"), strings.Contains(s, "win64"), strings.Contains(s, "win32"):
		return osWindows
	case strings.Contains(s, "mac os x"), strings.Contains(s, "macintosh"):
		return osMac
	case strings.Contains(s, "x11"), strings.Contains(s, "linux"):
		return osLinux
	}
	return osUnknown
}

func familyFromPlatform(p string) osFamily {
	s := strings.ToLower(p)
	switch {
	case s == "":
		return osUnknown
	case strings.HasPrefix(s, "win"):
		return osWindows
	case strings.HasPrefix(s, "mac"), strings.HasPrefix(s, "iphone"), strings.HasPrefix(s, "ipad"):

		if strings.HasPrefix(s, "iphone") || strings.HasPrefix(s, "ipad") {
			return osIOS
		}
		return osMac
	case strings.Contains(s, "android"):
		return osAndroid
	case strings.Contains(s, "cros"):
		return osChrome
	case strings.Contains(s, "linux"), strings.Contains(s, "x11"), strings.Contains(s, "bsd"):
		return osLinux
	}
	return osUnknown
}

func familyFromUAData(p string) osFamily {
	switch strings.ToLower(p) {
	case "":
		return osUnknown
	case "windows":
		return osWindows
	case "macos":
		return osMac
	case "linux":
		return osLinux
	case "android":
		return osAndroid
	case "chrome os", "chromeos":
		return osChrome
	}
	return osUnknown
}

func readClaim(r Request) claim {
	var c claim
	c.natives = readNativeAccessors(r)
	v, ok := r.value("scope.main")
	if !ok {
		return c
	}
	c.UserAgent, _ = str(v, "userAgent")
	c.NavPlatform, _ = str(v, "platform")

	if s, ok := str(v, "userAgentData", "platform"); ok {
		c.UADataPlat = s
	} else if s, ok := str(v, "uaDataPlatform"); ok {
		c.UADataPlat = s
	}

	c.uaFamily = familyFromUA(c.UserAgent)
	c.platFamily = familyFromPlatform(c.NavPlatform)
	c.uaDataFamily = familyFromUAData(c.UADataPlat)

	var known []osFamily
	for _, f := range []osFamily{c.uaFamily, c.platFamily, c.uaDataFamily} {
		if f != osUnknown {
			c.surfacesKnown++
			known = append(known, f)
		}
	}
	switch c.surfacesKnown {
	case 0:
		return c
	case 1:
		c.Family = known[0]
		return c
	}
	for i := 0; i < len(known); i++ {
		for j := i + 1; j < len(known); j++ {
			if !compatibleFamilies(known[i], known[j]) {
				return c
			}
		}
	}
	c.Family, c.Agreed = mostSpecificFamily(known), true
	return c
}

func compatibleFamilies(a, b osFamily) bool {
	if a == b {
		return true
	}
	for _, p := range [][2]osFamily{{osAndroid, osLinux}, {osChrome, osLinux}, {osIOS, osMac}} {
		if (a == p[0] && b == p[1]) || (a == p[1] && b == p[0]) {
			return true
		}
	}
	return false
}

func familySpecificity(f osFamily) int {
	switch f {
	case osAndroid, osChrome, osIOS:
		return 2
	case osWindows, osLinux, osMac:
		return 1
	}
	return 0
}

func mostSpecificFamily(fs []osFamily) osFamily {
	best := osUnknown
	for _, f := range fs {
		if familySpecificity(f) > familySpecificity(best) {
			best = f
		}
	}
	return best
}

func (f osFamily) String() string {
	switch f {
	case osWindows:
		return "Windows"
	case osMac:
		return "macOS"
	case osLinux:
		return "Linux"
	case osAndroid:
		return "Android"
	case osIOS:
		return "iOS"
	case osChrome:
		return "ChromeOS"
	}
	return "not determined"
}
