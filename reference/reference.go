package reference

// Source is where a value came from and when that was last checked.
type Source struct {
	Origin  string // spec section, vendor page URL, or engine build measured
	Checked string // YYYY-MM-DD
}

// Table is a set of values with its provenance. Verified reports whether the
// values have been observed on a system of the configuration described.
type Table struct {
	Values   []string
	Source   Source
	Verified bool
}
