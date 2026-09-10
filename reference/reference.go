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

type GPUArchitectureBlock struct {
	Mask uint32

	Devices map[uint32]string
}

type GPUArchitectureVendor struct {
	PCIVendor uint32

	Blocks []GPUArchitectureBlock

	Source Source

	Observed map[string]string
}

type GPUDecode struct {
	Family string

	AV1Decode bool

	Source Source

	Observed string

	Verified bool
}

type EngineFeature struct {
	ID string

	Description string

	ShipsInMajor int

	Source Source

	Observed string

	Verified bool
}

type NativeToStringResult struct {
	Measured    Table
	OtherEngine Table
}
