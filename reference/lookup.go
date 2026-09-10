package reference

func (v GPUArchitectureVendor) Architecture(device uint32) (string, bool) {
	for _, block := range v.Blocks {
		if name, ok := block.Devices[device&block.Mask]; ok {
			return name, true
		}
	}
	return "", false
}

func (v GPUArchitectureVendor) Verified(architecture string) bool {
	_, ok := v.Observed[architecture]
	return ok
}

func NativeToStringForms(name string) NativeToStringResult {
	return NativeToStringResult{
		Measured: Table{
			Values:   nativeForms(name, "{ [native code] }"),
			Source:   Source{Origin: "ECMA-262, Function.prototype.toString, NativeFunction grammar (tc39.es/ecma262/#sec-function.prototype.tostring); measured on V8 12.4.254.21-node.56", Checked: "2026-08-25"},
			Verified: true,
		},
		OtherEngine: Table{
			Values:   nativeForms(name, "{\n    [native code]\n}"),
			Source:   Source{Origin: "ECMA-262, Function.prototype.toString, NativeFunction grammar permits implementation-defined whitespace (tc39.es/ecma262/#sec-function.prototype.tostring)", Checked: "2026-08-25"},
			Verified: false,
		},
	}
}

func nativeForms(name, body string) []string {
	return []string{
		"function " + name + "() " + body,
		"function get " + name + "() " + body,
		"function () " + body,
	}
}
