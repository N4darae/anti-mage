package reference

type GPUDecode struct {
	Family string

	AV1Decode bool

	Source Source

	Observed string

	Verified bool
}

var NvidiaGeForceDecode = map[string]GPUDecode{
	"20": {
		Family:    "Turing",
		AV1Decode: false,
		Source: Source{
			Origin:  "NVIDIA Video Codec SDK, NVDEC Application Note, decode capability table (docs.nvidia.com/video-technologies/video-codec-sdk); nvidia.com/en-us/geforce/news/gfecnt/202009/rtx-30-series-av1-decoding, which announces AV1 decode as new to the 30 series",
			Checked: "2026-08-27",
		},
		Verified: false,
	},
	"30": {
		Family:    "Ampere",
		AV1Decode: true,
		Source: Source{
			Origin:  "nvidia.com/en-us/geforce/news/gfecnt/202009/rtx-30-series-av1-decoding; NVIDIA Video Codec SDK, NVDEC Application Note, decode capability table",
			Checked: "2026-08-27",
		},
		Observed: "GeForce RTX 3060 Ti (0x00002489), Windows 10 22H2, Chrome 151.0.7922.174: MediaCapabilities.decodingInfo reported powerEfficient for AV1 at 1920x1080 30fps, alongside H.264, VP9 and HEVC",
		Verified: true,
	},
	"40": {
		Family:    "Ada Lovelace",
		AV1Decode: true,
		Source: Source{
			Origin:  "NVIDIA Video Codec SDK, NVDEC Application Note, decode capability table",
			Checked: "2026-08-27",
		},
		Verified: false,
	},
	"50": {
		Family:    "Blackwell",
		AV1Decode: true,
		Source: Source{
			Origin:  "NVIDIA Video Codec SDK, NVDEC Application Note, decode capability table",
			Checked: "2026-08-27",
		},
		Verified: false,
	},
}
