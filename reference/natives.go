package reference

type NativeToStringResult struct {
	Measured    Table
	OtherEngine Table
}

func NativeToStringForms(name string) NativeToStringResult {
	return NativeToStringResult{
		Measured: Table{
			Values: []string{
				"function " + name + "() { [native code] }",
				"function get " + name + "() { [native code] }",
				"function () { [native code] }",
			},
			Source:   Source{Origin: "ECMA-262, Function.prototype.toString, NativeFunction grammar (tc39.es/ecma262/#sec-function.prototype.tostring); measured on V8 12.4.254.21-node.56", Checked: "2026-08-25"},
			Verified: true,
		},
		OtherEngine: Table{
			Values: []string{
				"function " + name + "() {\n    [native code]\n}",
				"function get " + name + "() {\n    [native code]\n}",
				"function () {\n    [native code]\n}",
			},
			Source:   Source{Origin: "ECMA-262, Function.prototype.toString, NativeFunction grammar permits implementation-defined whitespace (tc39.es/ecma262/#sec-function.prototype.tostring)", Checked: "2026-08-25"},
			Verified: false,
		},
	}
}

var NativeFunctionOwnKeys = Table{
	Values:   []string{"length", "name"},
	Source:   Source{Origin: "ECMA-262, Standard Built-in ECMAScript Objects (tc39.es/ecma262/multipage/ecmascript-standard-built-in-objects.html); measured on V8 12.4.254.21-node.56 via Object.getOwnPropertyNames, Reflect.ownKeys and Object.getOwnPropertyDescriptors", Checked: "2026-08-25"},
	Verified: true,
}

var TrustedErrorNames = Table{
	Values: []string{
		"AggregateError", "Error", "EvalError", "InternalError", "RangeError",
		"ReferenceError", "SyntaxError", "TypeError", "URIError",
	},
	Source:   Source{Origin: "ECMA-262, Error Objects (tc39.es/ecma262/multipage/fundamental-objects.html#sec-error-objects): Error, AggregateError, and the NativeError types EvalError/RangeError/ReferenceError/SyntaxError/TypeError/URIError; developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Error for the Firefox-only InternalError; measured on V8 12.4.254.21-node.56, where InternalError is correctly absent", Checked: "2026-08-25"},
	Verified: true,
}

var TrustedDOMExceptionNames = Table{
	Values:   []string{"InvalidStateError", "SecurityError"},
	Source:   Source{Origin: "WebIDL, DOMException error names table (webidl.spec.whatwg.org/#dfn-error-names-table)", Checked: "2026-08-25"},
	Verified: true,
}

var CSSSystemFontKeywords = Table{
	Values:   []string{"caption", "icon", "menu", "message-box", "small-caption", "status-bar"},
	Source:   Source{Origin: "CSS Fonts Module Level 4, system-family-name value (w3.org/TR/css-fonts-4/#system-family-name-value)", Checked: "2026-08-25"},
	Verified: true,
}

var BraveNativeToString = Table{
	Values:   []string{"function isBrave() { [native code] }"},
	Source:   Source{Origin: "derived from the measured V8 native-function toString form (NativeToStringForms) applied to the property name isBrave", Checked: "2026-08-25"},
	Verified: false,
}
