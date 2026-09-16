package assess

import (
	"encoding/json"
)

func ok(v string) Observation {
	return Observation{Status: StatusOK, Value: json.RawMessage(v)}
}

func unsup(reason string) Observation {
	return Observation{Status: StatusUnsupported, Value: json.RawMessage(`{"reason":` + quote(reason) + `}`)}
}

func quote(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return string(b)
}

func honest() map[string]Observation {
	const scope = `{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36",` +
		`"platform":"Win32","hardwareConcurrency":8,"language":"en-GB","languages":["en-GB","en"],` +
		`"timeZone":"Europe/London","locale":"en-GB"}`
	return map[string]Observation{
		"scope.main":   ok(scope),
		"scope.worker": ok(scope),
		"scope.iframe": ok(scope),
		"native.tostring": ok(`{"navigator.platform":"function get platform() { [native code] }",` +
			`"screen.width":"function get width() { [native code] }"}`),
		"native.ownkeys": ok(`{"navigator.platform":{"ownKeys":["length","name"],` +
			`"getOwnPropertyNames":["length","name"],"descriptors":["length","name"]}}`),
		"native.descriptor": ok(`{"navigator.platform":{"onPrototype":true}}`),
		"native.receiver":   ok(`{"navigator.platform":{"threw":true,"name":"TypeError"}}`),
		"geom.screen":       ok(`{"width":1536,"height":864,"availWidth":1536,"availHeight":816,"devicePixelRatio":1.33333}`),
		"geom.css":          ok(`{"dppx":1.333330094819539}`),
		"auto.residue":      ok(`{"webdriver":false,"driverNames":[]}`),
		"perm.state":        ok(`{"notifications":{"query":"prompt","actual":"default"}}`),

		"math.exact": ok(`{
			"abs.nan":"NaN","abs.negZero":"0","abs.negInf":"Infinity","abs.neg":"7","abs.pos":"7",
			"sign.nan":"NaN","sign.posZero":"0","sign.negZero":"-0","sign.pos":"1","sign.neg":"-1",
			"floor.nan":"NaN","floor.posInf":"Infinity","floor.negInf":"-Infinity","floor.negZero":"-0","floor.fracNeg":"-1","floor.fracPos":"2",
			"ceil.nan":"NaN","ceil.posInf":"Infinity","ceil.negInf":"-Infinity","ceil.posZero":"0","ceil.fracNeg":"-0","ceil.fracPos":"3",
			"trunc.nan":"NaN","trunc.fracNeg":"-0","trunc.fracPos":"0","trunc.negInt":"-3",
			"round.nan":"NaN","round.negZero":"-0","round.halfNeg":"-0","round.halfPos":"1","round.negHalfInt":"-2","round.posInf":"Infinity","round.negInf":"-Infinity",
			"min.nan":"NaN","min.zero":"-0","min.basic":"1","min.empty":"Infinity",
			"max.nan":"NaN","max.zero":"0","max.basic":"3","max.empty":"-Infinity",
			"fround.nan":"NaN","fround.negZero":"-0","fround.overflow":"Infinity","fround.tieToEven":"16777216","fround.exact":"0.5",
			"clz32.zero":"32","clz32.one":"31","clz32.negOne":"0","clz32.nan":"32",
			"imul.basic":"12","imul.overflow":"-5","imul.bigxbig":"1","imul.nan":"0",
			"sqrt.nan":"NaN","sqrt.negative":"NaN","sqrt.negZero":"-0","sqrt.posInf":"Infinity","sqrt.perfect":"2","sqrt.exactFraction":"2.5"
		}`),
		"math.repeat": ok(`{
			"sin":   {"a":"0.8414709848078965", "b":"0.8414709848078965"},
			"sqrt":  {"a":"1.4142135623730951", "b":"1.4142135623730951"},
			"round": {"a":"3", "b":"3"}
		}`),

		"throw.mandated": ok(`{
			"atob.invalidChars": {
				"available": true, "threw": true, "name": "InvalidCharacterError", "ctor": "DOMException",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]", "hasStack": true,
				"attempt2": {"threw": true, "name": "InvalidCharacterError", "ctor": "DOMException", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]"}
			},
			"createElement.invalidName": {
				"available": true, "threw": true, "name": "InvalidCharacterError", "ctor": "DOMException",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]", "hasStack": true,
				"attempt2": {"threw": true, "name": "InvalidCharacterError", "ctor": "DOMException", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]"}
			},
			"json.malformed": {
				"available": true, "threw": true, "name": "SyntaxError", "ctor": "SyntaxError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "SyntaxError", "ctor": "SyntaxError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"property.accessOnNull": {
				"available": true, "threw": true, "name": "TypeError", "ctor": "TypeError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "TypeError", "ctor": "TypeError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"call.nonCallable": {
				"available": true, "threw": true, "name": "TypeError", "ctor": "TypeError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "TypeError", "ctor": "TypeError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"array.negativeLength": {
				"available": true, "threw": true, "name": "RangeError", "ctor": "RangeError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "RangeError", "ctor": "RangeError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"decodeURIComponent.malformed": {
				"available": true, "threw": true, "name": "URIError", "ctor": "URIError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "URIError", "ctor": "URIError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"structuredClone.function": {
				"available": true, "threw": true, "name": "DataCloneError", "ctor": "DOMException",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]", "hasStack": true,
				"attempt2": {"threw": true, "name": "DataCloneError", "ctor": "DOMException", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]"}
			}
		}`),

		"rect.identities": ok(`{
			"shiftPx": 41,
			"base":     {"x":37,"y":53,"left":37,"top":53,"right":218,"bottom":150,"width":181,"height":97},
			"twin":     {"x":37,"y":53,"left":37,"top":53,"right":218,"bottom":150,"width":181,"height":97},
			"shifted":  {"x":78,"y":53,"left":78,"top":53,"right":259,"bottom":150,"width":181,"height":97},
			"restored": {"x":37,"y":53,"left":37,"top":53,"right":218,"bottom":150,"width":181,"height":97}
		}`),
		"text.metrics": ok(`{
			"empty": {"width": 0, "box": {
				"actualBoundingBoxLeft": 0, "actualBoundingBoxRight": 0,
				"actualBoundingBoxAscent": 0, "actualBoundingBoxDescent": 0,
				"fontBoundingBoxAscent": 29, "fontBoundingBoxDescent": 8,
				"emHeightAscent": 24, "emHeightDescent": 6,
				"hangingBaseline": 21.6, "alphabeticBaseline": 0, "ideographicBaseline": -8
			}},
			"full": {"width": 123.4, "box": {
				"actualBoundingBoxLeft": 0, "actualBoundingBoxRight": 123.4,
				"actualBoundingBoxAscent": 21, "actualBoundingBoxDescent": 5,
				"fontBoundingBoxAscent": 29, "fontBoundingBoxDescent": 8,
				"emHeightAscent": 24, "emHeightDescent": 6,
				"hangingBaseline": 21.6, "alphabeticBaseline": 0, "ideographicBaseline": -8
			}},
			"repeat": {"width": 123.4},
			"prefixWidths": [0, 10, 20, 35, 60, 90, 123.4]
		}`),

		"media.stylesheet": ok(`{
			"controlOk": true, "widthValid": true, "heightValid": true,
			"numeric": [
				{"feature":"width","op":"min","px":1536,"jsMatches":true,"cssMatches":true},
				{"feature":"width","op":"max","px":1535,"jsMatches":false,"cssMatches":false},
				{"feature":"height","op":"min","px":738,"jsMatches":true,"cssMatches":true},
				{"feature":"height","op":"max","px":737,"jsMatches":false,"cssMatches":false}
			],
			"discrete": [
				{"feature":"orientation","value":"landscape","jsMatches":true,"cssMatches":true},
				{"feature":"orientation","value":"portrait","jsMatches":false,"cssMatches":false},
				{"feature":"hover","value":"hover","jsMatches":true,"cssMatches":true},
				{"feature":"hover","value":"none","jsMatches":false,"cssMatches":false},
				{"feature":"pointer","value":"fine","jsMatches":true,"cssMatches":true},
				{"feature":"pointer","value":"coarse","jsMatches":false,"cssMatches":false},
				{"feature":"pointer","value":"none","jsMatches":false,"cssMatches":false},
				{"feature":"prefers-color-scheme","value":"light","jsMatches":true,"cssMatches":true},
				{"feature":"prefers-color-scheme","value":"dark","jsMatches":false,"cssMatches":false},
				{"feature":"prefers-reduced-motion","value":"no-preference","jsMatches":true,"cssMatches":true},
				{"feature":"prefers-reduced-motion","value":"reduce","jsMatches":false,"cssMatches":false}
			]
		}`),
		"media.complement": ok(`{
			"innerWidth": 1536, "innerHeight": 738,
			"complements": [
				{"query":"(min-width: 1px)","matches":true,"negationMatches":false},
				{"query":"(min-width: 999999px)","matches":false,"negationMatches":true}
			],
			"brackets": [
				{"feature":"width","value":1536,"insideBelowPx":1535,"insideAbovePx":1537,"minInside":true,"maxInside":true,"minOutside":false,"maxOutside":false},
				{"feature":"height","value":738,"insideBelowPx":737,"insideAbovePx":739,"minInside":true,"maxInside":true,"minOutside":false,"maxOutside":false}
			]
		}`),

		"audio.views": ok(`{
			"requested": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160},
			"rendered": {"sampleRateHz": 8000, "numberOfChannels": 2, "lengthFrames": 160, "durationSeconds": 0.02},
			"channelsServed": 2,
			"copyFromChannelAvailable": true,
			"views": {"compared": true, "agree": true, "sampleCount": 320, "differingSampleCount": 0, "maxAbsoluteDifference": 0}
		}`),
		"audio.repeat": ok(`{
			"secondRenderCompleted": true,
			"repeat": {"compared": true, "agree": true, "sampleCount": 320, "differingSampleCount": 0, "maxAbsoluteDifference": 0}
		}`),
	}
}

func with(changes map[string]Observation) map[string]Observation {
	obs := honest()
	for id, o := range changes {
		if o.Status == "" && o.Value == nil {
			delete(obs, id)
			continue
		}
		obs[id] = o
	}
	return obs
}
