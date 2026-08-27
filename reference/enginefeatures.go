package reference

type EngineFeature struct {
	ID string

	Description string

	ShipsInMajor int

	Source Source

	Observed string

	Verified bool
}

var EngineFeatures = []EngineFeature{
	{
		ID:           "textFit",
		Description:  "the text-fit CSS property",
		ShipsInMajor: 150,
		Source: Source{
			Origin:  "developer.chrome.com/release-notes/150, 'text-fit scales the font size of text nodes to fit its containing box'",
			Checked: "2026-08-27",
		},
		Verified: false,
	},
	{
		ID:           "bgClipBorderArea",
		Description:  "background-clip: border-area",
		ShipsInMajor: 150,
		Source: Source{
			Origin:  "developer.chrome.com/release-notes/150, 'background-clip: border-area clips a background to the area painted by its border strokes'",
			Checked: "2026-08-27",
		},
		Verified: false,
	},
	{
		ID:           "rubyOverhang",
		Description:  "the ruby-overhang CSS property",
		ShipsInMajor: 151,
		Source: Source{
			Origin:  "developer.chrome.com/release-notes/151, 'ruby-overhang accepts auto, spaces, or none'",
			Checked: "2026-08-27",
		},
		Observed: "Chrome 151.0.7922.174, Windows 10: CSS.supports('ruby-overhang','auto') answered true",
		Verified: true,
	},
	{
		ID:           "animationEventAnimation",
		Description:  "AnimationEvent.prototype.animation",
		ShipsInMajor: 151,
		Source: Source{
			Origin:  "developer.chrome.com/release-notes/151, read-only animation property added to AnimationEvent",
			Checked: "2026-08-27",
		},
		Observed: "Chrome 151.0.7922.174, Windows 10: 'animation' in AnimationEvent.prototype answered true",
		Verified: true,
	},
	{
		ID:           "transitionEventAnimation",
		Description:  "TransitionEvent.prototype.animation",
		ShipsInMajor: 151,
		Source: Source{
			Origin:  "developer.chrome.com/release-notes/151, read-only animation property added to TransitionEvent",
			Checked: "2026-08-27",
		},
		Observed: "Chrome 151.0.7922.174, Windows 10: 'animation' in TransitionEvent.prototype answered true",
		Verified: true,
	},
	{
		ID:           "userMediaElement",
		Description:  "the usermedia element global constructor surface",
		ShipsInMajor: 151,
		Source: Source{
			Origin:  "developer.chrome.com/release-notes/151, '<usermedia> element (MVP)'",
			Checked: "2026-08-27",
		},
		Observed: "Chrome 151.0.7922.174, Windows 10: 'HTMLUserMediaElement' in window answered true",
		Verified: true,
	},
	{
		ID:           "softNavigations",
		Description:  "the soft-navigation performance entry type",
		ShipsInMajor: 151,
		Source: Source{
			Origin:  "developer.chrome.com/release-notes/151, Soft Navigations API, 'soft-navigation' and 'interaction-contentful-paint' entries",
			Checked: "2026-08-27",
		},
		Observed: "Chrome 151.0.7922.174, Windows 10: PerformanceObserver.supportedEntryTypes included 'soft-navigation'",
		Verified: true,
	},
}
