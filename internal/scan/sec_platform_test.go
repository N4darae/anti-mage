package scan

import "testing"

func platformClaim(t *testing.T, ua, plat, uaData string) (Request, claim) {
	t.Helper()
	v := `{"userAgent":"` + ua + `","platform":"` + plat + `","userAgentData":{"platform":"` + uaData + `"}}`
	r := Request{Probes: map[string]Probe{"scope.main": {Status: StatusOK, Value: []byte(v)}}}
	return r, readClaim(r)
}

func TestKernelNameOnASurfaceThatReportsTheKernelIsNotADisagreement(t *testing.T) {
	honest := []struct {
		name, ua, plat, uaData string
		want                   osFamily
	}{
		{
			"Android, where navigator.platform names the kernel",
			"Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36",
			"Linux armv8l", "Android", osAndroid,
		},
		{
			"Android without client hints",
			"Mozilla/5.0 (Android 14; Mobile; rv:127.0) Gecko/127.0 Firefox/127.0",
			"Linux aarch64", "", osAndroid,
		},
		{
			"ChromeOS, where navigator.platform names the kernel",
			"Mozilla/5.0 (X11; CrOS x86_64 14541.0.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
			"Linux x86_64", "Chrome OS", osChrome,
		},
		{
			"a tablet whose platform names the desktop family it shares",
			"Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/126 Mobile/15E148 Safari/604.1",
			"MacIntel", "", osIOS,
		},
		{
			"a phone that names one family everywhere",
			"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			"iPhone", "", osIOS,
		},
		{
			"a desktop that names one family everywhere",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
			"Win32", "Windows", osWindows,
		},
	}
	for _, c := range honest {
		r, cl := platformClaim(t, c.ua, c.plat, c.uaData)
		if !cl.Agreed {
			t.Errorf("%s: the surfaces were read as disagreeing", c.name)
		}
		if cl.Family != c.want {
			t.Errorf("%s: family = %q, want %q; the more specific of two compatible names is the claim",
				c.name, cl.Family, c.want)
		}
		if got := sectionPlatform(r, Inputs{}, cl).Determination; got != Consistent {
			t.Errorf("%s: determination = %q, want %q", c.name, got, Consistent)
		}
	}
}

func TestSurfacesNamingFamiliesThatShareNoKernelStillDisagree(t *testing.T) {
	spoofs := []struct{ name, ua, plat, uaData string }{
		{"a desktop claim over a foreign kernel", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/126", "Linux x86_64", "Windows"},
		{"a desktop claim over another desktop", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/126", "MacIntel", "Windows"},
		{"a phone claim over a desktop kernel", "Mozilla/5.0 (Linux; Android 14) Chrome/126 Mobile", "Win32", "Android"},
		{"client hints naming a different family", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Safari/605", "MacIntel", "Windows"},
		{"two of three naming families that share nothing", "Mozilla/5.0 (X11; Linux x86_64) Chrome/126", "Win32", "Android"},
	}
	for _, c := range spoofs {
		r, cl := platformClaim(t, c.ua, c.plat, c.uaData)
		if cl.Agreed {
			t.Errorf("%s: read as agreeing", c.name)
		}
		if got := sectionPlatform(r, Inputs{}, cl).Determination; got != Contradiction {
			t.Errorf("%s: determination = %q, want %q", c.name, got, Contradiction)
		}
	}
}

func TestCompatibilityIsSymmetricAndOnlyEverKernelDeep(t *testing.T) {
	all := []osFamily{osWindows, osMac, osLinux, osAndroid, osIOS, osChrome}
	for _, a := range all {
		if !compatibleFamilies(a, a) {
			t.Errorf("%q is not compatible with itself", a)
		}
		for _, b := range all {
			if compatibleFamilies(a, b) != compatibleFamilies(b, a) {
				t.Errorf("compatibility of %q and %q is not symmetric", a, b)
			}
		}
	}
	for _, pair := range [][2]osFamily{
		{osWindows, osLinux}, {osWindows, osMac}, {osWindows, osAndroid},
		{osWindows, osIOS}, {osWindows, osChrome}, {osMac, osLinux},
		{osMac, osAndroid}, {osMac, osChrome}, {osLinux, osIOS},
		{osAndroid, osIOS}, {osAndroid, osChrome}, {osIOS, osChrome},
	} {
		if compatibleFamilies(pair[0], pair[1]) {
			t.Errorf("%q and %q were read as compatible; neither is the other's kernel", pair[0], pair[1])
		}
	}
}
