package scan

import (
	_ "time/tzdata"
)

type throwsMandatedCase struct {
	id    string
	label string

	mandate      string
	clause       string
	ctor         string
	name         string
	domException bool
}

var throwsMandatedCases = []throwsMandatedCase{
	{
		id:           "atob.invalidChars",
		mandate:      "the HTML atob() method",
		label:        "atob() on input the forgiving-base64 grammar rejects",
		clause:       "atob() throws InvalidCharacterError on input whose length is 1 mod 4 (html.spec.whatwg.org/multipage/webappapis.html#dom-windowbase64-atob)",
		ctor:         "DOMException",
		name:         "InvalidCharacterError",
		domException: true,
	},
	{
		id:           "createElement.invalidName",
		mandate:      "DOM's Document.createElement()",
		label:        "document.createElement('') with an empty local name",
		clause:       "createElement throws InvalidCharacterError when localName is not a valid XML Name (dom.spec.whatwg.org/#dom-document-createelement)",
		ctor:         "DOMException",
		name:         "InvalidCharacterError",
		domException: true,
	},
	{
		id:      "json.malformed",
		mandate: "ECMA-262's JSON.parse",
		label:   "JSON.parse on text that is not valid JSON",
		clause:  "JSON.parse throws SyntaxError on invalid JSON text (tc39.es/ecma262/#sec-json.parse)",
		ctor:    "SyntaxError",
		name:    "SyntaxError",
	},
	{
		id:      "property.accessOnNull",
		mandate: "ECMA-262's property-accessor semantics",
		label:   "reading a property of null",
		clause:  "property access throws TypeError when the base value is null or undefined (tc39.es/ecma262/#sec-property-accessors-runtime-semantics-evaluation)",
		ctor:    "TypeError",
		name:    "TypeError",
	},
	{
		id:      "call.nonCallable",
		mandate: "ECMA-262's Call abstract operation",
		label:   "calling a plain object as a function",
		clause:  "calling a non-callable value throws TypeError (tc39.es/ecma262/#sec-call)",
		ctor:    "TypeError",
		name:    "TypeError",
	},
	{
		id:      "array.negativeLength",
		mandate: "ECMA-262's Array constructor",
		label:   "new Array(-1)",
		clause:  "Array constructor throws RangeError on an invalid length argument (tc39.es/ecma262/#sec-array-len)",
		ctor:    "RangeError",
		name:    "RangeError",
	},
	{
		id:      "decodeURIComponent.malformed",
		mandate: "ECMA-262's decodeURIComponent",
		label:   "decodeURIComponent('%')",
		clause:  "decodeURIComponent throws URIError on a malformed percent-encoded sequence (tc39.es/ecma262/#sec-decodeuricomponent-uri)",
		ctor:    "URIError",
		name:    "URIError",
	},
	{
		id:           "structuredClone.function",
		mandate:      "HTML's StructuredSerialize",
		label:        "structuredClone of a Function value",
		clause:       "structuredClone throws DataCloneError on a value with no defined serialization, such as a Function (html.spec.whatwg.org/multipage/structured-data.html#structuredserialize)",
		ctor:         "DOMException",
		name:         "DataCloneError",
		domException: true,
	},
}
