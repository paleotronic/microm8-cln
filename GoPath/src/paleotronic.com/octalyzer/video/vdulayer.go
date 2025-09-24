package video

import (
	//"image"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
)

type VDULayer interface {
	Render()
	GetPosition() (float32, float32, float32)
	SetPosition(x, y, z float32)
}

type BaseLayer struct {
	X, Y, Z float32 // bottom left corner -- its a GL thing, OK? ;)
	//Pitch, Yaw float32
	ID string
	// Storage
	Buffer *memory.MemoryControlBlock
	// Pixel / Decal arrangement
	Width  int
	Height int
	// Px / Decal dimensions
	UnitWidth  float32
	UnitHeight float32
	UnitDepth  float32
	//
	BoxWidth  float32
	BoxHeight float32
	BoxDepth  float32
	//
	//Palette types.VideoPalette
	//Bounds  Rectangle
	//
	//Hidden bool
	// Transparent
	TransparentIndex []int
	// Format
	Format types.LayerFormat
	// PosChanged
	PosChanged bool
	//
	MPos types.LayerPos
	//
	Spec *types.LayerSpecMapped
	// Tint color
	Tint        *types.VideoColor
	TintChanged bool
	// Controller
	Controller        *types.OrbitCameraData
	glWidth, glHeight float32
	lastAspect        float64
}

func (b *BaseLayer) IsTransparent(c int, p *types.VideoPalette) bool {
	if b.Format == types.LF_SUPER_HIRES {
		return c == 0
	}
	return p.Get(c%p.Size()).Alpha == 0
}

func (b *BaseLayer) GetPosition() (float32, float32, float32) {
	x, y, z := b.Spec.GetPos()
	return float32(x), float32(y), float32(z)
}

func (b *BaseLayer) SetPosition(x, y, z float32) {

	b.Spec.SetPos(float64(x), float64(y), float64(z))

	b.PosChanged = true
}

func (b *BaseLayer) Render() {
	// nothing here -- we override this one in a descendant
}

func (b *BaseLayer) GetUnitWidth() float32 {
	return b.UnitWidth
}

func (b *BaseLayer) GetUnitHeight() float32 {
	return b.UnitHeight
}

func (b *BaseLayer) GetUnitDepth() float32 {
	return b.UnitDepth
}

func (b *BaseLayer) GetWidth() int {
	return b.Width
}

func (b *BaseLayer) GetHeight() int {
	return b.Height
}
