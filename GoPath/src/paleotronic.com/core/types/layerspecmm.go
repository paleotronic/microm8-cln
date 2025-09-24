package types

import (
	"paleotronic.com/fmt"

	"paleotronic.com/core/hires"
	"paleotronic.com/core/memory"
)

type LayerSpecMapped struct {
	Index       int
	Base        int // Base Address
	Mm          *memory.MemoryMap
	Control     *TextBuffer   // pointer to text buffer controller
	VControl    *VectorBuffer // pointer to vector buffer controller
	HControl    hires.HGRControllable
	CubeControl *CubeScreen
}

const (
	LO_MEMCHG        = 0
	LO_INDEX         = 1
	LO_ID            = 2
	LO_ACTIVE        = 3
	LO_KIND          = 4
	LO_FORMAT        = 5
	LO_SUB_FORMAT    = 6
	LO_NUMBLOCKS     = 7
	LO_BLOCKBASE     = 8
	LO_B0_BASE       = 8
	LO_B0_SIZE       = 9
	LO_B1_BASE       = 10
	LO_B1_SIZE       = 11
	LO_B2_BASE       = 12
	LO_B2_SIZE       = 13
	LO_B3_BASE       = 14
	LO_B3_SIZE       = 15
	LO_WIDTH         = 16
	LO_HEIGHT        = 17
	LO_BOUNDS_X0     = 18
	LO_BOUNDS_Y0     = 19
	LO_BOUNDS_X1     = 20
	LO_BOUNDS_Y1     = 21
	LO_POS_X         = 22
	LO_POS_Y         = 23
	LO_POS_Z         = 24
	LO_MONO          = 25
	LO_PAL_SIZE      = 26
	LO_PAL_ITEM_SIZE = 2
	LO_PAL_RGBA      = 0
	LO_PAL_OFFS      = 1
	LO_PAL_BASE      = 27
)

func (ls *LayerSpecMapped) GetIndex() byte {
	return byte(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_INDEX))
}

func (ls *LayerSpecMapped) SetIndex(n byte) {
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_INDEX, uint64(n))
}

// ID returns the layer id as a string direct from memory
func (ls *LayerSpecMapped) GetID() string {
	v := ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_ID)
	s := ""
	for i := 0; i < 4; i++ {
		s = string(rune(v&0xff)) + s
		v = v >> 8
	}

	return s
}

// SetID sets the id value
func (ls *LayerSpecMapped) SetID(s string) {

	if len(s) > 4 {
		s = s[0:4]
	}
	for len(s) < 4 {
		s = s + " "
	}
	var v uint64
	for i := 0; i < 4; i++ {
		v = (v << 8) | uint64(s[i]&0xff)
	}

	// log.Printf("setting id for layer to %s", s)
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_ID, v)
}

// IsActive says whether a layer is active or not
func (ls *LayerSpecMapped) GetActive() bool {
	v := ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_ACTIVE)
	return (v != 0)
}

// SetActive sets a layers Active status
func (ls *LayerSpecMapped) SetActive(b bool) {

	c := ls.GetActive()
	if (b && !c) || (!b && c) {
		// update
		var v uint64
		if b {
			v = 1
		}
		ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_ACTIVE, v)
	}
}

// IsActive says whether a layer is active or not
func (ls *LayerSpecMapped) GetMono() bool {
	v := ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_MONO)
	return (v != 0)
}

// SetActive sets a layers Active status
func (ls *LayerSpecMapped) SetMono(b bool) {

	c := ls.GetMono()
	if (b && !c) || (!b && c) {
		// update
		var v uint64
		if b {
			v = 1
		}
		ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_MONO, v)
	}
}

const MEMCHG_REBUILD = 0x01
const MEMCHG_REFRESH = 0x02

// IsDirty says whether a layer is dirty memcfg wise or not
func (ls *LayerSpecMapped) GetDirty() bool {
	v := ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_MEMCHG)
	return (v&MEMCHG_REBUILD != 0)
}

// SetActive sets a layers Active status
func (ls *LayerSpecMapped) SetDirty(b bool) {
	c := ls.GetDirty()
	if (b && !c) || (!b && c) {
		// update
		v := ls.Mm.ReadInterpreterMemory(ls.Index, ls.Base+LO_MEMCHG) & (0xFFFF ^ MEMCHG_REBUILD)
		if b {
			v = v | MEMCHG_REBUILD
		}
		ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_MEMCHG, v)
	}
}

func (ls *LayerSpecMapped) GetRefresh() bool {
	v := ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_MEMCHG)
	b := (v&MEMCHG_REFRESH != 0)
	return b
}

// SetActive sets a layers Active status
func (ls *LayerSpecMapped) SetRefresh(b bool) {
	c := ls.GetRefresh()
	if (b && !c) || (!b && c) {
		// update
		v := ls.Mm.ReadInterpreterMemory(ls.Index, ls.Base+LO_MEMCHG) & (0xFFFF ^ MEMCHG_REFRESH)
		if b {
			v = v | MEMCHG_REFRESH
		}
		ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_MEMCHG, v)
		//fmt.Printf("Marked layer %s for refresh of visual state = %v\n", ls.GetID(), b)
	}
}

func (ls *LayerSpecMapped) GetKind() uint64 {
	return ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_KIND)
}

func (ls *LayerSpecMapped) SetKind(k uint64) {
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_KIND, k)
}

func (ls *LayerSpecMapped) GetFormat() LayerFormat {
	return LayerFormat(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_FORMAT))
}

func (ls *LayerSpecMapped) SetFormat(f LayerFormat) {
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_FORMAT, uint64(f))
}

func (ls *LayerSpecMapped) GetSubFormat() LayerSubFormat {
	return LayerSubFormat(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_SUB_FORMAT))
}

func (ls *LayerSpecMapped) SetSubFormat(f LayerSubFormat) {
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_SUB_FORMAT, uint64(f))
}

func (ls *LayerSpecMapped) GetNumBlocks() int {
	return int(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_NUMBLOCKS))
}

func (ls *LayerSpecMapped) SetNumBlocks(n int) {
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_NUMBLOCKS, uint64(n))
}

func (ls *LayerSpecMapped) GetBlock(block int) (int, int) {

	if block < 0 || block > 3 {
		return 0, 0
	}

	base := int(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_BLOCKBASE+block*2))
	size := int(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_BLOCKBASE+block*2+1))

	return base, size

}

func (ls *LayerSpecMapped) SetBlock(block int, base, size int) {

	if block < 0 || block > 3 {
		return
	}

	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_BLOCKBASE+block*2, uint64(base))
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_BLOCKBASE+block*2+1, uint64(size))

}

func (ls *LayerSpecMapped) GetWidth() uint16 {
	return uint16(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_WIDTH))
}

func (ls *LayerSpecMapped) SetWidth(n uint16) {
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_WIDTH, uint64(n))
}

func (ls *LayerSpecMapped) GetHeight() uint16 {
	return uint16(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_HEIGHT))
}

func (ls *LayerSpecMapped) SetHeight(n uint16) {
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_HEIGHT, uint64(n))
}

func (ls *LayerSpecMapped) GetBounds() (uint16, uint16, uint16, uint16) {

	x0 := uint16(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_BOUNDS_X0))
	y0 := uint16(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_BOUNDS_Y0))
	x1 := uint16(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_BOUNDS_X1))
	y1 := uint16(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_BOUNDS_Y1))

	return x0, y0, x1, y1

}

func (ls *LayerSpecMapped) SetBounds(x0, y0, x1, y1 uint16) {

	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_BOUNDS_X0, uint64(x0))
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_BOUNDS_Y0, uint64(y0))
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_BOUNDS_X1, uint64(x1))
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_BOUNDS_Y1, uint64(y1))

}

func (ls *LayerSpecMapped) GetPos() (float64, float64, float64) {

	x := Uint2Float64(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_POS_X))
	y := Uint2Float64(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_POS_Y))
	z := Uint2Float64(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_POS_Z))

	return x, y, z

}

func (ls *LayerSpecMapped) SetPos(x, y, z float64) {

	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_POS_X, Float642uint(x))
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_POS_Y, Float642uint(y))
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_POS_Z, Float642uint(z))

}

func (ls *LayerSpecMapped) GetPaletteSize() int {
	return int(ls.Mm.ReadInterpreterMemorySilent(ls.Index, ls.Base+LO_PAL_SIZE))
}

func (ls *LayerSpecMapped) SetPaletteSize(n int) {
	ls.Mm.WriteInterpreterMemory(ls.Index, ls.Base+LO_PAL_SIZE, uint64(n))
}

func (ls *LayerSpecMapped) SetPaletteColor(index int, vc *VideoColor) {
	maxcols := ls.GetPaletteSize()
	if index < 0 || index >= maxcols {
		return
	}
	offset := ls.Base + LO_PAL_BASE + (index * LO_PAL_ITEM_SIZE)

	var rgba uint64
	var zoff uint64

	zoff = uint64(32768+int(vc.Offset)) | (uint64(vc.Depth) << 16)
	rgba = vc.ToUintRGBA()

	ls.Mm.WriteInterpreterMemory(ls.Index, offset, rgba)
	ls.Mm.WriteInterpreterMemory(ls.Index, offset+1, zoff)
}

func (ls *LayerSpecMapped) GetPaletteColor(index int) *VideoColor {

	vc := &VideoColor{}

	maxcols := ls.GetPaletteSize()
	if index < 0 || index >= maxcols {
		return nil
	}
	offset := ls.Base + LO_PAL_BASE + (index * LO_PAL_ITEM_SIZE)

	var rgba uint64 = ls.Mm.ReadInterpreterMemorySilent(ls.Index, offset)
	var zoff uint64 = ls.Mm.ReadInterpreterMemorySilent(ls.Index, offset+1)

	vc.FromUintRGBA(rgba)
	vc.Offset = int8(int(zoff&65535) - 32768)
	vc.Depth = uint8(int(zoff) >> 16)

	return vc
}

func (ls *LayerSpecMapped) GetBase() int {
	base, _ := ls.GetBlock(0)
	return base
}

func (ls *LayerSpecMapped) PopulateFromSpec(spec *LayerSpec) {

	var reconfig bool
	if spec.ID == ls.GetID() && spec.Format == ls.GetFormat() {
		reconfig = true
	}

	if !reconfig {

		ls.SetID(spec.ID)
		ls.SetFormat(spec.Format)
		ls.SetActive((spec.Active == 1))
		ls.SetIndex(spec.Index)
		ls.SetWidth(spec.Width)
		ls.SetHeight(spec.Height)
		ls.SetNumBlocks(len(spec.Blocks))
		ls.SetBounds(
			spec.Bounds.X0,
			spec.Bounds.Y0,
			spec.Bounds.X1,
			spec.Bounds.Y1,
		)

		// map blocks
		for i, b := range spec.Blocks {
			ls.SetBlock(i, b.Base, b.Size)
		}

	}

	// map in palette
	ls.SetPaletteSize(spec.Palette.Size())
	for i := 0; i < spec.Palette.Size(); i++ {
		ls.SetPaletteColor(i, spec.Palette.Get(i))
	}

	// Add tail end memory mappers
	ls.HControl = spec.HControl
	ls.VControl = spec.VControl
	ls.Control = spec.Control

	// mark dirty for a full spec load
	ls.SetDirty(true)
}

// NewLayerSpecMapped creates a layerspec based off the old non memory mapped version
// The other is still useful for YAML files, as it has named fields
func NewLayerSpecMapped(mm *memory.MemoryMap, spec *LayerSpec, index int, base int) *LayerSpecMapped {

	this := &LayerSpecMapped{
		Mm:    mm,
		Index: index,
		Base:  (base % memory.OCTALYZER_INTERPRETER_SIZE),
	}

	this.PopulateFromSpec(spec)

	return this

}

func (ls *LayerSpecMapped) String() string {

	x0, y0, x1, y1 := ls.GetBounds()

	return fmt.Sprintf(
		"Index: %d, Id: %s, Format: %s (@ $%05x), SubFormat: %s, Type: %d, Active: %v, Width: %d, Height: %d, Palette: %d colors, Bounds(Top-Left: %d, %d, Bot-Right: %d, %d)",
		ls.GetIndex(),
		ls.GetID(),
		ls.GetFormat().String(),
		ls.GetBase(),
		ls.GetSubFormat().String(),
		ls.GetKind(),
		ls.GetActive(),
		ls.GetWidth(),
		ls.GetHeight(),
		ls.GetPaletteSize(),
		x0,
		y0,
		x1,
		y1,
	)

}

func (ls *LayerSpecMapped) GetBoundsRect() LayerRect {

	x0, y0, x1, y1 := ls.GetBounds()

	return LayerRect{x0, y0, x1, y1}

}

func (ls *LayerSpecMapped) GetPalette() VideoPalette {

	p := *NewVideoPalette()
	for i := 0; i < ls.GetPaletteSize(); i++ {
		p.Add(ls.GetPaletteColor(i))
	}

	return p

}

func (ls *LayerSpecMapped) SetPalette(p VideoPalette) {

	for i := 0; i < p.Size(); i++ {
		ls.SetPaletteColor(i, p.Get(i))
	}
	ls.SetRefresh(true)

}
