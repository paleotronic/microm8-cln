package settings

import "image/color"
import "paleotronic.com/log"

type VideoColor struct {
	R, G, B, A uint8
	Depth      uint8
	Offset     int8
}

func (this VideoColor) ToColorRGBA() color.RGBA {
	return color.RGBA{this.R, this.G, this.B, this.A}
}

type VideoPalette struct {
	Content []*VideoColor
}

func NewVideoPalette() *VideoPalette {
	return &VideoPalette{
		Content: []*VideoColor{},
	}
}

func (vp *VideoPalette) Size() int {
	return len(vp.Content)
}

func (vp *VideoPalette) Add(vc *VideoColor) {
	vp.Content = append(vp.Content, vc)
}

func (vp *VideoPalette) Get(c int) *VideoColor {
	return vp.Content[c%len(vp.Content)]
}

func (vp *VideoPalette) SetColor(c int, vc *VideoColor) {
	vp.Content[c] = vc
}

type VideoPaletteZone struct {
	X0, Y0  int
	X1, Y1  int
	Palette *VideoPalette
}

type VideoPaletteZoneConfig struct {
	Palettes      []*VideoPalette
	PixelMapping  []byte
	Zones         []*VideoPaletteZone
	Width, Height int
	Updated       bool
}

func NewVideoPaletteZoneConfig(width, height int) *VideoPaletteZoneConfig {
	return &VideoPaletteZoneConfig{
		PixelMapping: make([]byte, width*height),
		Width:        width,
		Height:       height,
		Zones:        []*VideoPaletteZone{},
		Updated:      true,
	}
}

func (zc *VideoPaletteZoneConfig) AddZone(z *VideoPaletteZone) {
	zc.Zones = append(zc.Zones, z)
	zc.updatePixelMap()
}

func (zc *VideoPaletteZoneConfig) SetZone(zone int, z *VideoPaletteZone) {
	for len(zc.Zones)-1 < zone {
		log.Println("adding zone record")
		zc.Zones = append(zc.Zones, nil)
	}
	zc.Zones[zone] = z
	zc.updatePixelMap()
}

func (zc *VideoPaletteZoneConfig) RemoveZone(z int) {
	if z < 0 || z >= len(zc.Zones) {
		return
	}
	zc.Zones = append(zc.Zones[:z], zc.Zones[z+1:]...)
	zc.updatePixelMap()
}

func (zc *VideoPaletteZoneConfig) DeleteZone(z int) {
	if z < 0 || z >= len(zc.Zones) {
		return
	}
	zc.Zones[z] = nil
	zc.updatePixelMap()
}

func (zc *VideoPaletteZoneConfig) updatePixelMap() {
	zc.PixelMapping = make([]byte, zc.Width*zc.Height) // defaults to zero
	for i, z := range zc.Zones {
		if z == nil {
			continue
		}
		log.Printf("updating zone %d -> %+v", i, *z)
		for y := z.Y0; y <= z.Y1; y++ {
			for x := z.X0; x <= z.X1; x++ {
				zc.PixelMapping[y*zc.Width+x] = byte(i + 1)
			}
		}
	}
	log.Println("updated config")
	zc.Updated = true
}

func (zc *VideoPaletteZoneConfig) IsUpdated() bool {
	b := zc.Updated
	zc.Updated = false
	return b
}

func (zc *VideoPaletteZoneConfig) SetUpdate(b bool) {
	zc.Updated = b
}

func (zc *VideoPaletteZoneConfig) GetZoneAt(x, y int) int {
	idx := y*zc.Width + x
	if idx >= len(zc.PixelMapping) {
		return 0
	}
	return int(zc.PixelMapping[idx])
}

func (zc *VideoPaletteZoneConfig) GetColorAt(x, y, c int) *VideoColor {
	idx := y*zc.Width + x
	if idx >= len(zc.PixelMapping) {
		return nil
	}
	var i = int(zc.PixelMapping[idx])
	if i == 0 || zc.Zones[i-1] == nil {
		return nil
	}
	var p = zc.Zones[i-1].Palette
	return p.Get(c % p.Size())
}

func (zc *VideoPaletteZoneConfig) SetZoneColor(z int, c int, col *VideoColor) {
	if z < 1 || z > len(zc.Zones) || zc.Zones[z] == nil {
		return
	}
	var p = zc.Zones[z].Palette
	p.SetColor(c%p.Size(), col)
	zc.Updated = true
}

func (zc *VideoPaletteZoneConfig) GetZoneColor(z int, c int) *VideoColor {
	if z < 1 || z > len(zc.Zones) || zc.Zones[z] == nil {
		return nil
	}
	var p = zc.Zones[z].Palette
	return p.Get(c % p.Size())
}

// ColorZone mappings
var ColorZone [NUMSLOTS][32]*VideoPaletteZoneConfig

func InitSlotZones(slotid int) {
	for i := 0; i < len(ColorZone[slotid]); i++ {
		ColorZone[slotid][i] = nil
	}
}

func init() {
	// p := NewVideoPalette()
	// p.Add(&VideoColor{R: 0, G: 0, B: 0, A: 0, Depth: 20, Offset: 0})
	// p.Add(&VideoColor{R: 96, G: 0, B: 0, A: 255, Depth: 20, Offset: 0})
	// p.Add(&VideoColor{R: 128, G: 0, B: 0, A: 255, Depth: 20, Offset: 0})
	// p.Add(&VideoColor{R: 255, G: 0, B: 0, A: 255, Depth: 20, Offset: 0})
	// p.Add(&VideoColor{R: 0, G: 0, B: 0, A: 0, Depth: 20, Offset: 0})
	// p.Add(&VideoColor{R: 96, G: 0, B: 0, A: 255, Depth: 20, Offset: 0})
	// p.Add(&VideoColor{R: 128, G: 0, B: 0, A: 255, Depth: 20, Offset: 0})
	// p.Add(&VideoColor{R: 255, G: 0, B: 0, A: 255, Depth: 20, Offset: 0})
	// ColorZone[0][5] = NewVideoPaletteZoneConfig(280, 192)
	// ColorZone[0][5].AddZone(
	// 	&VideoPaletteZone{
	// 		Palette: p,
	// 		X0:      50, Y0: 50,
	// 		X1: 139, Y1: 150,
	// 	},
	// )
}
