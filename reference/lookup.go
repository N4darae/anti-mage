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

func nativeForms(name, body string) []string {
	return []string{
		"function " + name + "() " + body,
		"function get " + name + "() " + body,
		"function () " + body,
	}
}
