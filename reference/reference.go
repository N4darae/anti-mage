package reference

type Source struct {
	Origin  string
	Checked string
}

type Table struct {
	Values   []string
	Source   Source
	Verified bool
}
