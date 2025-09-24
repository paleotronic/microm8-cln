package types

import (
	"errors"

	"paleotronic.com/core/hires"
	"paleotronic.com/core/memory"
	"paleotronic.com/fmt"
)

const (
	MAXLAYER           int  = 16
	APPLE_TEXT_PAGE_0  byte = 0
	APPLE_TEXT_PAGE_1  byte = 1
	APPLE_LORES_PAGE_0 byte = 0
	APPLE_HIRES_PAGE_0 byte = 1
	APPLE_HIRES_PAGE_1 byte = 2
)

type Layerable interface {
	ReadLayersFromMemory()
	WriteLayersToMemory()
	GetGFXLayerByID(name string) (*LayerSpec, bool)
	GetHUDLayerByID(name string) (*LayerSpec, bool)
}

type LayerRect struct {
	X0, Y0, X1, Y1 uint16
}

func (l LayerRect) Contains(x, y uint16) bool {

	return (x >= l.X0 && x <= l.X1 && y >= l.Y0 && y <= l.Y1)
}

func (l LayerRect) Equals(o LayerRect) bool {

	return (l.X0 == o.X0 && l.X1 == o.X1 && l.Y0 == o.Y0 && l.Y1 == o.Y1)

}

type LayerPos struct {
	X, Y, Z int16
}

type LayerPosMod struct {
	XPercent float64
	YPercent float64
}

func (l *LayerPos) MarshalBinary() ([]byte, error) {

	return []byte{
		byte(l.X & 0xff), byte((l.X >> 8) & 0xff),
		byte(l.Y & 0xff), byte((l.Y >> 8) & 0xff),
		byte(l.Z & 0xff), byte((l.Z >> 8) & 0xff),
	}, nil

}

func (l *LayerPos) UnmarshalBinary(data []byte) error {

	if len(data) < 6 {
		return errors.New("not enough data for LayerPos")
	}

	l.X = int16(data[0]) + (int16(data[1]) << 8)
	l.Y = int16(data[2]) + (int16(data[3]) << 8)
	l.Z = int16(data[4]) + (int16(data[5]) << 8)

	return nil

}

type LayerFormat byte

const (
	// Apple Modes
	LF_TEXT_LINEAR   LayerFormat = 0
	LF_TEXT_WOZ      LayerFormat = 1
	LF_LOWRES_LINEAR LayerFormat = 2
	LF_LOWRES_WOZ    LayerFormat = 3
	LF_HGR_LINEAR    LayerFormat = 4
	LF_HGR_WOZ       LayerFormat = 5
	LF_HGR_X         LayerFormat = 6
	LF_VECTOR        LayerFormat = 7
	LF_DHGR_WOZ      LayerFormat = 8
	LF_CUBE_PACKED   LayerFormat = 9
	// BBC Modes
	LF_BBC_2COL_LINEAR LayerFormat = 10
	LF_BBC_4COL_LINEAR LayerFormat = 11
	LF_BBC_8COL_LINEAR LayerFormat = 12
	LF_BBC_TELE_TEXT   LayerFormat = 13
	// GS
	LF_SUPER_HIRES LayerFormat = 16
	// Speccy
	LF_SPECTRUM_0 LayerFormat = 17
)

func (lf LayerFormat) String() string {
	switch lf {
	case 0:
		return "Linear TextBuffer"
	case 1:
		return "Wozniak TextBuffer"
	case 2:
		return "Linear Lowres buffer"
	case 3:
		return "Wozniak Lowres buffer"
	case 4:
		return "Linear Hires buffer"
	case 5:
		return "Wozniak Hires buffer"
	case 6:
		return "Linear Xtended Hires buffer"
	case 7:
		return "Vector Graphics buffer"
	case 8:
		return "Wozniak Double-Hires buffer"
	case 9:
		return "3D cube buffer"
	case 10:
		return "BBC Micro 2 color framebuffer"
	case 11:
		return "BBC Micro 4 color framebuffer"
	case 12:
		return "BBC Micro 8 color framebuffer"
	case 13:
		return "BBC Micro Teletext buffer"
	case 16:
		return "Apple Super HiRES"
	case 17:
		return "ZX Spectrum 256x192 15 color"
	default:
		return "Custom buffer"
	}
}

// LayerSubFormat is a set of formats for text display...
// It allows more rigid char movement than other modes.
type LayerSubFormat byte

const (
	LSF_FREEFORM     LayerSubFormat = 0
	LSF_FIXED_80_48  LayerSubFormat = 1
	LSF_FIXED_80_24  LayerSubFormat = 2
	LSF_FIXED_40_24  LayerSubFormat = 3
	LSF_FIXED_40_48  LayerSubFormat = 4
	LSF_COLOR_LAYER  LayerSubFormat = 10
	LSF_SINGLE_LAYER LayerSubFormat = 11
	LSF_GREY_LAYER   LayerSubFormat = 12
	LSF_GREEN_LAYER  LayerSubFormat = 13
	LSF_AMBER_LAYER  LayerSubFormat = 14
	LSF_VOXELS       LayerSubFormat = 15
)

func (sf LayerSubFormat) String() string {
	switch sf {
	case LSF_FREEFORM:
		return "freeform"
	case LSF_FIXED_80_24:
		return "80x24 Apple"
	case LSF_FIXED_40_24:
		return "40x24 Apple"
	case LSF_FIXED_80_48:
		return "80x48 Apple"
	case LSF_FIXED_40_48:
		return "40x48 Apple"
	case LSF_VOXELS:
		return "Voxels"
	case LSF_SINGLE_LAYER:
		return "Raster"
	}
	return "Unknown"
}

func (l *LayerRect) MarshalBinary() ([]byte, error) {

	return []byte{
		byte(l.X0 % 256), byte(l.X0 / 256),
		byte(l.Y0 % 256), byte(l.Y0 / 256),
		byte(l.X1 % 256), byte(l.X1 / 256),
		byte(l.Y1 % 256), byte(l.Y1 / 256),
	}, nil

}

func (l *LayerRect) UnmarshalBinary(data []byte) error {

	if len(data) < 8 {
		return errors.New("not enough data for LayerRect")
	}

	l.X0 = uint16(data[0]) + 256*uint16(data[1])
	l.Y0 = uint16(data[2]) + 256*uint16(data[3])
	l.X1 = uint16(data[4]) + 256*uint16(data[5])
	l.Y1 = uint16(data[6]) + 256*uint16(data[7])

	return nil

}

type LayerSpec struct {
	Format        LayerFormat
	Index         byte   // Index within class (see type)
	Type          byte   // 0 = HUD, 1 = GFX
	ID            string // must be 4 chars
	Active        byte   // 0 = hidden, else shown
	Base          uint64
	Blocks        []memory.MemoryRange
	Width         uint16
	Height        uint16
	VirtualWidth  uint16
	VirtualHeight uint16
	Bounds        LayerRect
	Pos           LayerPos
	Palette       VideoPalette
	Control       *TextBuffer   // pointer to text buffer controller
	VControl      *VectorBuffer // pointer to vector buffer controller
	HControl      hires.HGRControllable
	CubeControl   *CubeScreen
}

func (ls *LayerSpec) String() string {

	return fmt.Sprintf(
		"Index: %d, Id: %s, Format: %s (@ $%05x), Type: %d, Active: %d, Width: %d, Height: %d, Palette: %d colors [%s], Bounds(Top-Left: %d, %d, Bot-Right: %d, %d)",
		ls.Index,
		ls.ID,
		ls.Format.String(),
		ls.Base,
		ls.Type,
		ls.Active,
		ls.Width,
		ls.Height,
		ls.Palette.Size(),
		ls.Palette.String(),
		ls.Bounds.X0,
		ls.Bounds.Y0,
		ls.Bounds.X1,
		ls.Bounds.Y1,
	)

}

func (ls *LayerSpec) StringWithoutActive() string {

	return fmt.Sprintf(
		"Index: %d, Id: %s, Format: %s (@ $%05x), Type: %d, Width: %d, Height: %d, Palette: %d colors [%s], Bounds(Top-Left: %d, %d, Bot-Right: %d, %d)",
		ls.Index,
		ls.ID,
		ls.Format.String(),
		ls.Base,
		ls.Type,
		ls.Width,
		ls.Height,
		ls.Palette.Size(),
		ls.Palette.String(),
		ls.Bounds.X0,
		ls.Bounds.Y0,
		ls.Bounds.X1,
		ls.Bounds.Y1,
	)

}

func (ls *LayerSpec) MarshalBinary() ([]byte, error) {

	data := []byte{
		byte(MtLayerSpec),
		byte(ls.Format),
		ls.Index,
		ls.Type,
		ls.Active,
		byte(ls.Width % 256), byte(ls.Width / 256),
		byte(ls.Height % 256), byte(ls.Height / 256),
		byte(ls.VirtualWidth % 256), byte(ls.VirtualWidth / 256),
		byte(ls.VirtualHeight % 256), byte(ls.VirtualHeight / 256),
		byte(ls.Base % 256),
		byte((ls.Base / 256) % 256),
		byte((ls.Base / 65536) % 256),
	}
	p, _ := ls.Pos.MarshalBinary()
	data = append(data, p...)
	b, _ := ls.Bounds.MarshalBinary()
	data = append(data, b...)
	for len(ls.ID) < 4 {
		ls.ID = ls.ID + " "
	}
	data = append(data, []byte(ls.ID[0:4])...)
	p, _ = ls.Palette.MarshalBinary()
	data = append(data, p...)

	return data, nil

}

func (ls *LayerSpec) UnmarshalBinary(data []byte) error {

	if data[0] != MtLayerSpec {
		return errors.New("not a LayerSpec")
	}

	if len(data) < 27 {
		return errors.New("not enough data")
	}

	ls.Format = LayerFormat(data[1])
	ls.Index = data[2]
	ls.Type = data[3]
	ls.Active = data[4]
	ls.Width = uint16(data[5]) + 256*uint16(data[6])
	ls.Height = uint16(data[7]) + 256*uint16(data[8])
	ls.VirtualWidth = uint16(data[9]) + 256*uint16(data[10])
	ls.VirtualHeight = uint16(data[11]) + 256*uint16(data[12])

	ls.Base = uint64(data[13]) + 256*uint64(data[14]) + 65536*uint64(data[15])

	//
	pchunk := data[16:22]
	_ = ls.Pos.UnmarshalBinary(pchunk)

	// add 6
	bchunk := data[22:30]
	ls.ID = string(data[30:34])
	////fmt.Printf("[%s]\n", ls.ID)
	pchunk = data[34:]

	ls.Bounds.UnmarshalBinary(bchunk)
	_ = ls.Palette.UnmarshalBinary(pchunk)

	return nil

}

func ModeAsLayers(vm *VideoMode) []LayerSpec {

	layers := make([]LayerSpec, 0)

	xtf := 80 / vm.Columns
	ytf := 48 / vm.Rows

	// Text mode x-translation
	layers = append(layers,
		LayerSpec{
			Index:         0,
			Type:          0, // Text
			Width:         80,
			Height:        48,
			VirtualWidth:  uint16(vm.Columns),
			VirtualHeight: uint16(vm.Rows),
			Bounds: LayerRect{
				X0: uint16(vm.DefaultWindow.Left * xtf),
				Y0: uint16(vm.DefaultWindow.Top * ytf),
				X1: uint16(vm.DefaultWindow.Right * xtf),
				Y1: uint16(vm.DefaultWindow.Bottom * ytf),
			},
			Palette: vm.Palette,
		},
	)

	return layers
}

func LayerConfigGFX(index byte, active bool, width, height int, palette VideoPalette, bounds LayerRect) LayerSpec {

	var af byte = 0
	if active {
		af = 1
	}

	return LayerSpec{
		Index:         index,
		Type:          1,
		Active:        af,
		Width:         uint16(width),
		Height:        uint16(height),
		VirtualWidth:  uint16(width),
		VirtualHeight: uint16(height),
		Palette:       palette,
		Bounds:        bounds,
	}

}

func LayerConfigText(index byte, active bool, width, height int, vwidth, vheight int, palette VideoPalette, bounds LayerRect) LayerSpec {

	var af byte = 0
	if active {
		af = 1
	}

	return LayerSpec{
		Index:         index,
		Type:          0,
		Active:        af,
		Width:         uint16(width),
		Height:        uint16(height),
		VirtualWidth:  uint16(vwidth),
		VirtualHeight: uint16(vheight),
		Palette:       palette,
		Bounds:        bounds,
	}

}

// Need a combined structure to represent this in order to maintain video sync
type LayerBundle struct {
	HUDLayers []LayerSpec
	GFXLayers []LayerSpec
}

func (lb *LayerBundle) MarshalBinary() ([]byte, error) {

	// Header
	data := make([]byte, 5)
	data[0] = MtLayerBundle
	data[3] = byte(len(lb.HUDLayers))
	data[4] = byte(len(lb.GFXLayers))

	// data[1,2] = total size of packet - used for validation - see later

	// Each HUDLayer
	for _, l := range lb.HUDLayers {

		bb, _ := l.MarshalBinary()
		data = append(data, byte(len(bb)%256), byte(len(bb)/256))
		data = append(data, bb...)

	}

	// Each GFXLayer
	for _, l := range lb.GFXLayers {
		bb, _ := l.MarshalBinary()
		data = append(data, byte(len(bb)%256), byte(len(bb)/256))
		data = append(data, bb...)
	}

	// Update length
	data[1] = byte(len(data) % 256)
	data[2] = byte(len(data) / 256)

	return data, nil

}

func (lb *LayerBundle) UnmarshalBinary(data []byte) error {

	if len(data) < 6 {
		return errors.New("Invalid layer bundle (not enough data)")
	}

	if data[0] != MtLayerBundle {
		return errors.New("Not a layer bundle")
	}

	size := int(data[1]) + 256*int(data[2])

	if len(data) != size {
		return errors.New(fmt.Sprintf("layerbundle should be %d bytes, got %d bytes\n", size, len(data)))
	}

	// Ok should be safe here to start the unpacking
	numHUD := int(data[3])
	numGFX := int(data[4])

	ptr := 5 // start of data
	countHUD := 0
	countGFX := 0

	lb.HUDLayers = make([]LayerSpec, 0)
	lb.GFXLayers = make([]LayerSpec, 0)

	for ptr < size {

		// get size of chunk
		csize := int(data[ptr]) + 256*int(data[ptr+1])
		ptr += 2

		//fmt.Printf("---> Unpack layer length %d bytes @ %d (%d)\n", csize, ptr, size)

		// now ptr is at the start of chunk and we have the csize,  so
		tmp := data[ptr : ptr+csize]
		ls := LayerSpec{}
		als := &ls

		e := als.UnmarshalBinary(tmp)
		if e != nil {
			return e
		}

		if ls.Type == 0 {
			lb.HUDLayers = append(lb.HUDLayers, ls)
			countHUD++
		} else if ls.Type == 1 {
			lb.GFXLayers = append(lb.GFXLayers, ls)
			countGFX++
		} else {
			return errors.New("Invalid layer type")
		}

		// bump ptr to start of next layer
		ptr += csize

	}

	if countHUD != numHUD {
		errors.New(fmt.Sprintf("Expected %d HUD type, got %d\n", numHUD, countHUD))
	}

	if countGFX != numGFX {
		errors.New(fmt.Sprintf("Expected %d GFX type, got %d\n", numHUD, countHUD))
	}

	return nil
}

func (lb *LayerBundle) String() string {
	out := "HUD Layers:\r\n"
	for _, ll := range lb.HUDLayers {
		out = out + ll.String() + "\r\n"
	}
	out = out + "GFX Layers:\r\n"
	for _, ll := range lb.GFXLayers {
		out = out + ll.String() + "\r\n"
	}
	return out
}

func (ls *LayerSpec) WriteToMemory(mm *memory.MemoryMap, index int, globalbase int) error {
	b, e := ls.MarshalBinary()
	if e != nil {
		return e
	}
	if len(b) > memory.OCTALYZER_LAYERSPEC_SIZE {
		return errors.New("LayerSpec seems bigger than limit")
	}

	chunk := make([]uint64, len(b))

	for i, v := range b {
		//mm.WriteGlobal( globalbase+i, uint64(v) )
		chunk[i] = uint64(v)
	}

	mm.BlockWrite(index, globalbase, chunk)

	return nil
}

func (ls *LayerSpec) ReadFromMemory(mm *memory.MemoryMap, index int, globalbase int) error {
	uc := mm.Data[index][globalbase : globalbase+memory.OCTALYZER_LAYERSPEC_SIZE]

	b := make([]byte, len(uc))
	for i, v := range uc {
		b[i] = byte(v)
	}

	e := ls.UnmarshalBinary(b)

	return e
}

func (ls *LayerSpec) Equals(ns *LayerSpec) bool {
	return (ns.String() == ls.String())
}

func (ls *LayerSpec) EqualsExceptActive(ns *LayerSpec) bool {
	return (ns.StringWithoutActive() == ls.StringWithoutActive())
}
