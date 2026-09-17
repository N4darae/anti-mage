package server

type Bootstrap struct {
	V        int           `json:"v"`
	Nonce    string        `json:"nonce"`
	IssuedAt string        `json:"issuedAt"`
	Font     BootstrapFont `json:"font"`
	Time     BootstrapTime `json:"time"`

	Fonts BootstrapFont `json:"fonts"`
	Dates []string      `json:"dates"`
}

type BootstrapFont struct {
	Controls      []string `json:"controls"`
	MeasureString string   `json:"measureString"`
	Bases         []string `json:"bases"`
}

type BootstrapTime struct {
	Offsets []OffsetSample `json:"offsets"`
}

type OffsetSample struct {
	Date    string `json:"date"`
	EpochMs int64  `json:"epochMs"`
	ISO     string `json:"iso"`
}
