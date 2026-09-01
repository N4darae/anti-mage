package reference

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

var GPUArchitectures = map[string]GPUArchitectureVendor{
	"nvidia": {
		PCIVendor: 0x10DE,
		Blocks: []GPUArchitectureBlock{
			{
				Mask: 0xFFFFFF00,
				Devices: map[uint32]string{
					0x00000D00: "fermi",
					0x00000F00: "kepler",
					0x00001000: "kepler",
					0x00001100: "kepler",
					0x00001200: "kepler",
					0x00001300: "maxwell",
					0x00001400: "maxwell",
					0x00001500: "pascal",
					0x00001600: "maxwell",
					0x00001700: "maxwell",
					0x00001B00: "pascal",
					0x00001C00: "pascal",
					0x00001D00: "pascal",
					0x00001E00: "turing",
					0x00001F00: "turing",
					0x00002000: "ampere",
					0x00002100: "turing",
					0x00002200: "ampere",
					0x00002400: "ampere",
					0x00002500: "ampere",
					0x00002600: "lovelace",
					0x00002700: "lovelace",
					0x00002800: "lovelace",
					0x00002B00: "blackwell",
					0x00002C00: "blackwell",
					0x00002D00: "blackwell",
					0x00002F00: "blackwell",
				},
			},
			{
				Mask: 0xFF000000,
				Devices: map[uint32]string{
					0x1E000000: "kepler",
					0x92000000: "maxwell",
					0x93000000: "pascal",
					0x97000000: "ampere",
					0xA5000000: "volta",
				},
			},
		},
		Source: Source{
			Origin:  "dawn.googlesource.com/dawn/+/refs/heads/main/src/dawn/gpu_info.json, the generator input Chromium derives GPUAdapterInfo.architecture from, read whole, sha256 b032f18d982f8e370f32af1e58e44611be2052c40e680cc554e90e8e5ea02a1c",
			Checked: "2026-08-31",
		},
		Observed: map[string]string{
			"ampere": "GeForce RTX 3060 Ti (0x00002489), Windows 10 22H2, Chrome 151 and Edge 151: WEBGL_debug_renderer_info named the device and GPUAdapterInfo reported architecture ampere on the default, high performance and low power requests",
		},
	},
	"amd": {
		PCIVendor: 0x1002,
		Blocks: []GPUArchitectureBlock{
			{
				Mask: 0x0000FFF0,
				Devices: map[uint32]string{
					0x00001110: "rdna-3",
					0x00001300: "gcn-1",
					0x00001310: "gcn-1",
					0x000013C0: "rdna-2",
					0x000013F0: "rdna-2",
					0x00001430: "rdna-2",
					0x00001500: "rdna-3",
					0x00001580: "rdna-3",
					0x000015B0: "rdna-3",
					0x000015C0: "rdna-3",
					0x000015D0: "gcn-5",
					0x000015E0: "rdna-2",
					0x00001640: "rdna-2",
					0x00001680: "rdna-2",
					0x00001900: "rdna-3",
					0x00006600: "gcn-1",
					0x00006610: "gcn-1",
					0x00006640: "gcn-2",
					0x00006650: "gcn-2",
					0x00006660: "gcn-1",
					0x000066A0: "gcn-5",
					0x00006730: "terascale-2",
					0x00006740: "terascale-2",
					0x00006750: "terascale-2",
					0x00006780: "terascale-2",
					0x00006790: "gcn-1",
					0x000067A0: "gcn-2",
					0x000067B0: "gcn-2",
					0x000067C0: "gcn-4",
					0x000067D0: "gcn-4",
					0x000067E0: "gcn-4",
					0x000067F0: "gcn-4",
					0x00006800: "gcn-1",
					0x00006810: "gcn-1",
					0x00006820: "gcn-1",
					0x00006830: "gcn-1",
					0x00006860: "gcn-5",
					0x00006870: "gcn-5",
					0x00006900: "gcn-3",
					0x00006920: "gcn-3",
					0x00006930: "gcn-3",
					0x00006940: "gcn-4",
					0x00006980: "gcn-4",
					0x00006990: "gcn-4",
					0x000069A0: "gcn-5",
					0x00006FD0: "gcn-4",
					0x00007300: "gcn-3",
					0x00007310: "rdna-1",
					0x00007340: "rdna-1",
					0x00007360: "rdna-1",
					0x00007380: "cdna-1",
					0x000073A0: "rdna-2",
					0x000073B0: "rdna-2",
					0x000073D0: "rdna-2",
					0x000073E0: "rdna-2",
					0x000073F0: "rdna-2",
					0x00007400: "rdna-2",
					0x00007420: "rdna-2",
					0x00007430: "rdna-2",
					0x00007440: "rdna-3",
					0x00007450: "rdna-3",
					0x00007470: "rdna-3",
					0x00007480: "rdna-3",
					0x00007550: "rdna-4",
					0x00007590: "rdna-4",
					0x00009830: "gcn-2",
					0x00009850: "gcn-2",
					0x00009870: "gcn-3",
					0x000098E0: "gcn-3",
					0x00009920: "gcn-4",
				},
			},
			{
				Mask: 0x0000FFFF,
				Devices: map[uint32]string{
					0x00001636: "gcn-5",
					0x00001638: "gcn-5",
					0x0000163F: "rdna-2",
				},
			},
		},
		Source: Source{
			Origin:  "dawn.googlesource.com/dawn/+/refs/heads/main/src/dawn/gpu_info.json, the generator input Chromium derives GPUAdapterInfo.architecture from, read whole, sha256 b032f18d982f8e370f32af1e58e44611be2052c40e680cc554e90e8e5ea02a1c",
			Checked: "2026-08-31",
		},
	},
	"intel": {
		PCIVendor: 0x8086,
		Blocks: []GPUArchitectureBlock{
			{
				Mask: 0x0000FF00,
				Devices: map[uint32]string{
					0x00000100: "gen-7",
					0x00000400: "gen-7",
					0x00000A00: "gen-7",
					0x00000C00: "gen-7",
					0x00000D00: "gen-7",
					0x00000F00: "gen-7",
					0x00001600: "gen-8",
					0x00001900: "gen-9",
					0x00002200: "gen-8",
					0x00003100: "gen-9",
					0x00003E00: "gen-9",
					0x00004500: "gen-11",
					0x00004600: "gen-12-lp",
					0x00004900: "gen-12-lp",
					0x00004C00: "gen-12-lp",
					0x00004E00: "gen-11",
					0x00004F00: "gen-12-hp",
					0x00005600: "gen-12-hp",
					0x00005900: "gen-9",
					0x00005A00: "gen-9",
					0x00006400: "xe-2-lpg",
					0x00007D00: "xe-lpg",
					0x00008700: "gen-9",
					0x00008A00: "gen-11",
					0x00009800: "gen-11",
					0x00009A00: "gen-12-lp",
					0x00009B00: "gen-9",
					0x0000A700: "gen-12-lp",
					0x0000B000: "xe-3-lpg",
					0x0000B600: "xe-lpg",
					0x0000E200: "xe-2-hpg",
					0x0000FD00: "xe-3-lpg-xs",
				},
			},
		},
		Source: Source{
			Origin:  "dawn.googlesource.com/dawn/+/refs/heads/main/src/dawn/gpu_info.json, the generator input Chromium derives GPUAdapterInfo.architecture from, read whole, sha256 b032f18d982f8e370f32af1e58e44611be2052c40e680cc554e90e8e5ea02a1c",
			Checked: "2026-08-31",
		},
	},
}
