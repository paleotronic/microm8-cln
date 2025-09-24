package hardware

import (
	"errors"
	"os"
	"strings"

	"paleotronic.com/core/hardware/common"
	"paleotronic.com/core/hardware/restalgia"
	"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
	"paleotronic.com/octalyzer/assets"
	"paleotronic.com/octalyzer/video/font"

	yaml "gopkg.in/yaml.v2"
)

type VoicePort struct {
	Name string
	Inst string
	Port int
}

type MachineCPUSpec struct {
	Model                 string
	Clocks                int
	VerticalRetraceCycles int
	VBlankCycles          int
	FPS                   int
	LinesPerFrame         int
	ScanCycles            int
	VBlankStartScan       int
}

var DefaultMachineSpec = MachineCPUSpec{
	Model:                 "6502",
	Clocks:                1020484,
	VerticalRetraceCycles: 12480,
	VBlankCycles:          4550,
	FPS:                   60,
	LinesPerFrame:         262,
	ScanCycles:            65,
}

type MachineSpec struct {
	Name              string
	ID                string
	SortOrder         int
	Font              string
	AuxFonts          []string
	RAM               []RAMArea
	ROMS              []RomSpec
	Components        []DeviceMapping
	Layers            []*types.LayerSpec
	Restalgia         []*VoicePort
	CPU               MachineCPUSpec
	CapsOnly          bool
	AltPassThrough    bool
	AllowDisklessBoot bool
}

type RAMArea struct {
	Name    string
	Base    int
	Size    int
	Type    int
	Regions []RAMBlock
}

type RAMBlock struct {
	Name      string
	Base      int
	End       int
	Src       string
	SrcBegin  int
	SrcLength int
	Active    bool
	Mux       int
	ForceMux  bool
	Mode      string
	Blank     bool
	Fill      uint64
}

type RomSpec struct {
	Name string
	Addr int
}

type DeviceMapping struct {
	Base     int
	Size     int
	Type     string
	LayerID  string
	Options  string
	Misc     map[string]map[interface{}]interface{}
	Switches []DeviceIOSwitch
	Rules    []DeviceIORule
}

// DeviceIOSwitch
type DeviceIOSwitch struct {
	Name     string
	Status   int
	IsOn     int
	IsOff    int
	Enable   int
	Disable  int
	RW       bool
	Triggers []string
}

// DeviceIORule
type DeviceIORule struct {
	Triggers   []string
	Conditions []string
}

func LoadSpec(filename string) (MachineSpec, error) {
	var ms MachineSpec
	//data, err := ioutil.ReadFile(filename)

	data, err := assets.Asset("profile/" + filename)

	if err != nil {
		return ms, err
	}
	err = yaml.Unmarshal(data, &ms)
	//fmt.Println(ms)
	return ms, err
}

func InjectROM(RAM *memory.MemoryMap, name string, index int, base int) error {

	data, e := assets.Asset(name)
	if e != nil {
		log.Printf("Error loading rom '%s': %s", name, e.Error())
		return e
	}

	rawdata := make([]uint64, len(data))
	for i, v := range data {
		rawdata[i] = uint64(v)
	}

	RAM.BlockWrite(index, RAM.MEMBASE(index)+base, rawdata)

	log.Printf("Spec: Loaded rom '%s' to address 0x%x", name, base)

	return nil

}

func BuildHardwareIO(i interfaces.Interpretable, ms MachineSpec, skipreset bool) {

	mm := i.GetMemoryMap()
	index := i.GetMemIndex()

	fmt.Println("NOT SKIPPING MACHINE")

	//i.DeleteCycleCounters()

	// // RAM mappings
	// mbm := memory.NewMemoryManagementUnit()
	// mm.BlockMapper[index] = mbm

	// for _, rb := range ms.RAM {
	// 	log.Printf("Processing memory allocation %s", rb.Name)
	// 	baseaddr := mm.MEMBASE(index) + rb.Base
	// 	for _, r := range rb.Regions {
	// 		rname := rb.Name + "." + r.Name
	// 		log.Printf("~> Handling region %s @ 0x%.4x - 0x%.4x", rname, r.Base, r.End)

	// 		size := r.End - r.Base + 1

	// 		if rb.Type == 0 || rb.Type == 1 {
	// 			mb := memory.NewMemoryBlockRAM(mm, baseaddr, r.Base, size, r.Active, rname, r.Mux)
	// 			if r.Mode != "" {
	// 				mb.SetState(r.Mode)
	// 			}
	// 			mbm.Register(mb)
	// 		} else {
	// 			if r.Blank {
	// 				data := make([]uint64, size)
	// 				for i, _ := range data {
	// 					data[i] = r.Fill
	// 				}
	// 				mb := memory.NewMemoryBlockROM(mm, baseaddr, r.Base, size, r.Active, rname, data)
	// 				mbm.Register(mb)
	// 			} else {
	// 				data, e := LoadData(r.Src, r.SrcBegin, r.SrcLength)
	// 				if e != nil {
	// 					panic(e)
	// 				}
	// 				mb := memory.NewMemoryBlockROM(mm, baseaddr, r.Base, size, r.Active, rname, data)
	// 				mbm.Register(mb)
	// 			}
	// 		}
	// 	}
	// }

	i.ClearAudioPorts()
	for _, vc := range ms.Restalgia {
		restalgia.CreateVoice(i, vc.Port, vc.Name, vc.Inst)
		i.SetAudioPort(vc.Name, vc.Port)
	}

	mbm := mm.BlockMapper[index]

	mm.InterpreterMappings[index] = make(memory.MapList)

	//fmt.Printf("Creating machine [%s]\n", ms.Name)
	switch {
	case strings.HasPrefix(ms.ID, "apple2"):
		for _, dm := range ms.Components {
			device := FactoryProduce(mm, index*memory.OCTALYZER_INTERPRETER_SIZE, dm.Base, dm.Type, i, dm.Misc, dm.Options, ms)
			if device == nil {
				continue
			}
			mm.MapInterpreterRegion(index, memory.MemoryRange{
				Base: device.GetBase(),
				Size: device.GetSize()},
				device)

			mb := memory.NewMemoryBlockIO(mm, index, mm.MEMBASE(index), device.GetBase(), device.GetSize(), true, strings.ToLower(device.GetLabel()), device)
			mbm.Register(mb)
		}
	}

	// save default
	mbm.Reset(skipreset)

	// log.Println("READ CONTEXT")
	// act := mbm.GetActiveBlocks(memory.MA_READ)
	// for i, b := range act {
	// 	log.Println(i, b.String())
	// }
	// log.Println("WRITE CONTEXT")
	// act = mbm.GetActiveBlocks(memory.MA_WRITE)
	// for i, b := range act {
	// 	log.Println(i, b.String())
	// }
}

func BuildHardware(i interfaces.Interpretable, ms MachineSpec, skipreset bool) {

	mm := i.GetMemoryMap()
	index := i.GetMemIndex()

	fmt.Println("NOT SKIPPING MACHINE")

	//i.DeleteCycleCounters()

	// RAM mappings
	mbm := memory.NewMemoryManagementUnit()

	if mm.BlockMapper[index] != nil {
		mm.BlockMapper[index].Done()
	}

	mm.BlockMapper[index] = mbm

	for _, rb := range ms.RAM {
		log.Printf("Processing memory allocation %s", rb.Name)
		baseaddr := mm.MEMBASE(index) + rb.Base
		for _, r := range rb.Regions {
			rname := rb.Name + "." + r.Name
			log.Printf("~> Handling region %s @ 0x%.4x - 0x%.4x", rname, r.Base, r.End)

			size := r.End - r.Base + 1

			if rb.Type == 0 || rb.Type == 1 {
				mb := memory.NewMemoryBlockRAM(mm, index, baseaddr, r.Base, size, r.Active, rname, r.Mux, r.ForceMux, rb.Type)
				if r.Mode != "" {
					mb.SetState(r.Mode)
				}
				mbm.Register(mb)
			} else {
				if r.Blank {
					data := make([]uint64, size)
					for i, _ := range data {
						data[i] = r.Fill
					}
					mb := memory.NewMemoryBlockROM(mm, index, baseaddr, r.Base, size, r.Active, rname, data)
					mbm.Register(mb)
				} else {
					data, e := common.LoadData(r.Src, r.SrcBegin, r.SrcLength)
					if e != nil {
						panic(e)
					}
					mb := memory.NewMemoryBlockROM(mm, index, baseaddr, r.Base, size, r.Active, rname, data)
					mbm.Register(mb)
				}
			}
		}
	}

	mm.InterpreterMappings[index] = make(memory.MapList)

	i.ClearAudioPorts()
	for _, vc := range ms.Restalgia {
		restalgia.CreateVoice(i, vc.Port, vc.Name, vc.Inst)
		i.SetAudioPort(vc.Name, vc.Port)
	}

	//fmt.Printf("Creating machine [%s]\n", ms.Name)
	for _, dm := range ms.Components {
		device := FactoryProduce(mm, index*memory.OCTALYZER_INTERPRETER_SIZE, dm.Base, dm.Type, i, dm.Misc, dm.Options, ms)
		if device == nil {
			continue
		}
		mm.MapInterpreterRegion(index, memory.MemoryRange{
			Base: device.GetBase(),
			Size: device.GetSize()},
			device)

		mb := memory.NewMemoryBlockIO(mm, index, mm.MEMBASE(index), device.GetBase(), device.GetSize(), true, strings.ToLower(device.GetLabel()), device)
		mbm.Register(mb)
	}

	// mm.

	// save default
	mbm.Reset(skipreset)

	log.Println("READ CONTEXT")
	act := mbm.GetActiveBlocks(memory.MA_READ)
	for i, b := range act {
		log.Println(i, b.String())
	}
	log.Println("WRITE CONTEXT")
	act = mbm.GetActiveBlocks(memory.MA_WRITE)
	for i, b := range act {
		log.Println(i, b.String())
	}
}

func LoadIOToInterpreter(i interfaces.Interpretable, specfile string) {

	skipmachine := false

	fmt.Printf("Existing spec %s, new spec %s\n", i.GetSpec(), specfile)

	if i.GetSpec() == specfile {
		// skip
		//fmt.Println("SKIPPING SPEC", i.GetSpec(), specfile)
		//return i.GetHUDLayerSet(), i.GetGFXLayerSet(), nil
		skipmachine = true
	}

	i.SetSpec(specfile)

	ms, err := LoadSpec(specfile)
	if err != nil {
		panic(err)
		//return []*types.LayerSpecMapped(nil), []*types.LayerSpecMapped(nil), err
	}

	if !skipmachine {

		BuildHardwareIO(i, ms, false)

	}

}

func LoadSpecPaletteData(i interfaces.Interpretable, specfile string, layer string) (*types.VideoPalette, error) {
	ms, err := LoadSpec(specfile)
	if err != nil {
		return nil, err
	}
	for _, l := range ms.Layers {
		if l.ID == layer {
			return &l.Palette, nil
		}
	}
	return nil, errors.New("Palette " + layer + " not found")
}

func LoadSpecToInterpreter(i interfaces.Interpretable, specfile string) ([]*types.LayerSpecMapped, []*types.LayerSpecMapped, error) {

	skipmachine := false

	fmt.Printf("Existing spec %s, new spec %s\n", i.GetSpec(), specfile)

	if i.GetSpec() == specfile {
		// skip
		//fmt.Println("SKIPPING SPEC", i.GetSpec(), specfile)
		//return i.GetHUDLayerSet(), i.GetGFXLayerSet(), nil
		skipmachine = true
	}

	i.SetSpec(specfile)

	mm := i.GetMemoryMap()

	ms, err := LoadSpec(specfile)
	if err != nil {
		specfile = settings.DefaultProfile
		i.SetSpec(specfile)
		ms, _ = LoadSpec(specfile)
	}

	index := i.GetMemIndex()

	settings.SystemID[index] = ms.ID

	if ms.Font != "" && settings.FirstBoot[index] {
		f, err := font.LoadFromFile(ms.Font)
		if err != nil {
			panic(err)
		}
		settings.DefaultFont[index] = f
		settings.Font[index] = ms.Font
	}

	settings.AuxFonts[index] = append([]string{ms.Font}, ms.AuxFonts...)

	if !skipmachine {

		BuildHardware(i, ms, false)

	}

	settings.CPUModel[i.GetMemIndex()] = ms.CPU.Model
	settings.CPUClock[i.GetMemIndex()] = ms.CPU.Clocks

	settings.SpecName[i.GetMemIndex()] = ms.Name
	settings.SetSubtitle(settings.SpecName[i.GetMemIndex()])

	//fmt.Printf("Mappable structure:\n%v\n", mm.InterpreterMappings)

	// Make structures
	hudlayers := make([]*types.LayerSpecMapped, memory.OCTALYZER_MAX_HUD_LAYERS)
	gfxlayers := make([]*types.LayerSpecMapped, memory.OCTALYZER_MAX_GFX_LAYERS)

	var rsh [256]memory.ReadSubscriptionHandler
	var esh [256]memory.ExecSubscriptionHandler
	var wsh [256]memory.WriteSubscriptionHandler

	for _, l := range ms.Layers {

		if l == nil {
			l = &types.LayerSpec{}
		}

		//fmt.RPrintf("[spec] vm#%d, layer: %d, id: %s\n", i.GetMemIndex(), lno, l.ID)

		offset := memory.OCTALYZER_LAYERSPEC_SIZE * int(l.Index)
		gbase := mm.MEMBASE(index) + memory.OCTALYZER_GFX_BASE + offset
		hbase := mm.MEMBASE(index) + memory.OCTALYZER_HUD_BASE + offset

		if len(l.Blocks) > 0 {
			mm.CreateMemoryHint(i.GetMemIndex(), l.ID, l.Blocks)
		}
		switch l.Type {
		case 0: // hud
			var tb *types.TextBuffer
			if l.Format == types.LF_TEXT_WOZ {
				memory.WarmStart = true
				tb = types.NewTextBufferMapped(
					false,
					int(l.Base),
					types.W_NORMAL_H_NORMAL,
					memory.NewMappedRegionFromHint(
						mm,
						index*memory.OCTALYZER_INTERPRETER_SIZE,
						int(l.Base),
						4096,
						l.Format.String(),
						l.ID,
						rsh,
						esh,
						wsh,
					),
					i,
				)
				memory.WarmStart = false
			} else {
				tb = types.NewTextBufferMapped(
					true,
					int(l.Base),
					types.W_NORMAL_H_NORMAL,
					memory.NewMappedRegionFromHint(
						mm,
						index*memory.OCTALYZER_INTERPRETER_SIZE,
						int(l.Base),
						4096,
						l.Format.String(),
						l.ID,
						rsh,
						esh,
						wsh,
					),
					i,
				)
			}
			l.Control = tb

			hudlayers[l.Index%memory.OCTALYZER_MAX_HUD_LAYERS] = types.NewLayerSpecMapped(mm, l, i.GetMemIndex(), hbase)
		case 1: // gfx
			switch l.Format {
			case types.LF_DHGR_WOZ:
				mcb := mm.GetHintedMemorySlice(index, l.ID)
				mcb.UseMM = true
				var woz *hires.DHGRScreen = hires.NewDHGRScreen(
					mcb,
				)
				l.HControl = woz
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(mm, l, i.GetMemIndex(), gbase)
			case types.LF_HGR_WOZ:
				mcb := mm.GetHintedMemorySlice(index, l.ID)
				mcb.UseMM = true
				var woz *hires.HGRScreen = hires.NewHGRScreen(
					mcb,
				)
				l.HControl = woz
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(mm, l, i.GetMemIndex(), gbase)
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS].SetDirty(true)
			case types.LF_HGR_X:
				var woz *hires.IndexedVideoBuffer = hires.NewIndexedVideoBuffer(
					280,
					192,
					mm.GetHintedMemorySlice(index, l.ID),
				)
				l.HControl = woz
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(mm, l, i.GetMemIndex(), gbase)
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS].SetDirty(true)
			case types.LF_LOWRES_WOZ:
				var tb *types.TextBuffer
				memory.WarmStart = true
				tb = types.NewTextBufferMapped(
					false,
					int(l.Base),
					types.W_NORMAL_H_NORMAL,
					memory.NewMappedRegionFromHint(
						mm,
						index*memory.OCTALYZER_INTERPRETER_SIZE,
						int(l.Base),
						4096,
						l.Format.String(),
						l.ID,
						rsh,
						esh,
						wsh,
					),
					i,
				)
				memory.WarmStart = false
				l.Control = tb
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(mm, l, i.GetMemIndex(), gbase)
			case types.LF_LOWRES_LINEAR:
				var tb *types.TextBuffer
				tb = types.NewTextBufferMapped(
					true,
					int(l.Base),
					types.W_NORMAL_H_NORMAL,
					memory.NewMappedRegionFromHint(
						mm,
						index*memory.OCTALYZER_INTERPRETER_SIZE,
						int(l.Base),
						4096,
						l.Format.String(),
						l.ID,
						rsh,
						esh,
						wsh,
					),
					i,
				)
				l.Control = tb
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(mm, l, i.GetMemIndex(), gbase)
			case types.LF_VECTOR:
				var vb *types.VectorBuffer
				vb = types.NewVectorBufferMapped(
					int(l.Base),
					0x10000,
					memory.NewMappedRegionFromHint(
						mm,
						index*memory.OCTALYZER_INTERPRETER_SIZE,
						int(l.Base),
						0x10000,
						l.Format.String(),
						l.ID,
						rsh,
						esh,
						wsh,
					),
				)
				l.VControl = vb
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(mm, l, i.GetMemIndex(), gbase)
			case types.LF_CUBE_PACKED:
				cb := types.NewCubeScreen(
					int(l.Base),
					0x10000,
					mm.GetHintedMemorySlice(index, l.ID),
				)
				l.CubeControl = cb
				log.Printf("Added Layer %s with CubeControl = %v", l.ID, l.CubeControl)
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(mm, l, i.GetMemIndex(), gbase)
			default:
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(mm, l, i.GetMemIndex(), gbase)
			}
		}
	}

	return hudlayers, gfxlayers, nil
}

func SaveSpec(filename string, ms MachineSpec) error {
	data, err := yaml.Marshal(&ms)
	if err != nil {
		return err
	}
	//fmt.Println(string(data))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	f.Write(data)
	return nil
}
