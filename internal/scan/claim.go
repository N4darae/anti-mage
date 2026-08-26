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

	tally := map[osFamily]int{}
	for _, f := range []osFamily{c.uaFamily, c.platFamily, c.uaDataFamily} {
		if f != osUnknown {
			c.surfacesKnown++
			tally[f]++
		}
	}
	switch c.surfacesKnown {
	case 0:
		return c
	case 1:

		for f := range tally {
			c.Family = f
		}
		return c
	}
	for f, n := range tally {
		if n == c.surfacesKnown {
			c.Family, c.Agreed = f, true
			return c
		}
	}

	return c
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
