package reference

func nativeForms(name, body string) []string {
	return []string{
		"function " + name + "() " + body,
		"function get " + name + "() " + body,
		"function () " + body,
	}
}
