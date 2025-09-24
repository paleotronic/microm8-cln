package apple2helpers

import (
	"paleotronic.com/log"

	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
)

var SlotZoneFormat [settings.NUMSLOTS]types.LayerFormat
var SlotZoneMode [settings.NUMSLOTS]string

func SetZoneFormat(e interfaces.Interpretable, f types.LayerFormat, name string) {
	SlotZoneFormat[e.GetMemIndex()] = f
	SlotZoneMode[e.GetMemIndex()] = name
}

func GetZoneFormat(e interfaces.Interpretable) types.LayerFormat {
	return SlotZoneFormat[e.GetMemIndex()]
}

func GetZoneSpec(e interfaces.Interpretable) *types.LayerSpecMapped {
	if SlotZoneMode[e.GetMemIndex()] != "" {
		return GETGFX(e, SlotZoneMode[e.GetMemIndex()])
	}
	return GETGFX(e, "HGR1")
}

func CreateZoneConfig(e interfaces.Interpretable) *settings.VideoPaletteZoneConfig {
	var l = GetZoneSpec(e)
	settings.ColorZone[e.GetMemIndex()][SlotZoneFormat[e.GetMemIndex()]] = settings.NewVideoPaletteZoneConfig(
		int(l.GetWidth()),
		int(l.GetHeight()),
	)
	return settings.ColorZone[e.GetMemIndex()][SlotZoneFormat[e.GetMemIndex()]]
}

func GetZoneConfig(e interfaces.Interpretable) *settings.VideoPaletteZoneConfig {
	log.Printf("SlotZoneFormat is %d", SlotZoneFormat[e.GetMemIndex()])
	return settings.ColorZone[e.GetMemIndex()][SlotZoneFormat[e.GetMemIndex()]]
}

// GetPalette converts a std VideoPalette to a Zone Palette
func GetPalette(p types.VideoPalette) *settings.VideoPalette {
	var out = settings.NewVideoPalette()
	for i := 0; i < p.Size(); i++ {
		c := p.Get(i)
		cc := &settings.VideoColor{
			R:      c.Red,
			G:      c.Green,
			B:      c.Blue,
			A:      c.Alpha,
			Depth:  c.Depth,
			Offset: c.Offset,
		}
		out.Add(cc)
	}
	return out
}

// AddZoneToConfig adds a new zone, copying the default mode palette to it
func AddZoneToConfig(e interfaces.Interpretable, x0, y0, x1, y1 int) (*settings.VideoPaletteZone, int) {
	zc := GetZoneConfig(e)
	if zc == nil {
		zc = CreateZoneConfig(e)
	}
	var l = GetZoneSpec(e)
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 >= int(l.GetWidth()) {
		x1 = int(l.GetWidth()) - 1
	}
	if y1 >= int(l.GetHeight()) {
		y1 = int(l.GetHeight()) - 1
	}
	z := &settings.VideoPaletteZone{
		X0:      x0,
		Y0:      y0,
		X1:      x1,
		Y1:      y1,
		Palette: GetPalette(l.GetPalette()),
	}
	zc.AddZone(z)
	return z, len(zc.Zones) - 1
}

func SetZone(e interfaces.Interpretable, zone int, x0, y0, x1, y1 int) *settings.VideoPaletteZone {
	zc := GetZoneConfig(e)
	if zc == nil {
		zc = CreateZoneConfig(e)
	}
	log.Printf("request to create zone %d, x0=%d, y0=%d, x1=%d, y1=%d", x0, y0, x1, y1)
	var l = GetZoneSpec(e)
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 >= int(l.GetWidth()) {
		x1 = int(l.GetWidth()) - 1
	}
	if y1 >= int(l.GetHeight()) {
		y1 = int(l.GetHeight()) - 1
	}
	z := &settings.VideoPaletteZone{
		X0:      x0,
		Y0:      y0,
		X1:      x1,
		Y1:      y1,
		Palette: GetPalette(l.GetPalette()),
	}
	log.Printf("adding zone config %+v", *z)
	zc.SetZone(zone, z)
	return z
}

func DeleteZoneFromConfig(e interfaces.Interpretable, z int) {
	zc := GetZoneConfig(e)
	if zc == nil {
		return
	}
	zc.DeleteZone(z)
}

// RemoveZoneFromConfig removes a palette zone from a zone config
func RemoveZoneFromConfig(e interfaces.Interpretable, z int) {
	zc := GetZoneConfig(e)
	if zc == nil {
		return
	}
	zc.RemoveZone(z)
}

// ResetZonePalette resets a zone palette
func ResetZonePalette(e interfaces.Interpretable, z int) {
	zc := GetZoneConfig(e)
	if zc == nil {
		return
	}
	if z < 0 || z >= len(zc.Zones) || zc.Zones[z] == nil {
		return
	}
	var l = GetZoneSpec(e)
	zc.Zones[z].Palette = GetPalette(l.GetPalette())
}

func InitZones(e interfaces.Interpretable) {
	var l = GetZoneSpec(e)
	settings.ColorZone[e.GetMemIndex()][SlotZoneFormat[e.GetMemIndex()]] = settings.NewVideoPaletteZoneConfig(int(l.GetWidth()), int(l.GetHeight()))
}

func GetZonePaletteValue(e interfaces.Interpretable, z int, c int) (*settings.VideoColor, bool) {
	zc := GetZoneConfig(e)
	if zc == nil || z < 0 || z >= len(zc.Zones) || zc.Zones[z] == nil {
		return &settings.VideoColor{0, 0, 0, 0, 0, 0}, false
	}
	return zc.Zones[z].Palette.Get(c), (c > 0 && c < zc.Zones[z].Palette.Size())
}

func SetZonePaletteValue(e interfaces.Interpretable, z int, c int, cc *settings.VideoColor) bool {
	zc := GetZoneConfig(e)
	if zc == nil || z < 0 || z >= len(zc.Zones) || zc.Zones[z] == nil {
		return false
	}
	if c > 0 && c < zc.Zones[z].Palette.Size() {
		log.Printf("updating zone %d color %d to %+v", z, c, *cc)
		//zc.Zones[z].Palette.SetColor(c, cc)
		zc.SetZoneColor(z, c, cc) // call forces updates
	}
	return false
}

func SetZonePaletteRGBA(e interfaces.Interpretable, z int, c int, R, G, B, A uint8) bool {
	if cc, ok := GetZonePaletteValue(e, z, c); ok {
		cc.R, cc.G, cc.B, cc.A = R, G, B, A
		return SetZonePaletteValue(e, z, c, cc)
	}
	return false
}

func SetZonePaletteDepth(e interfaces.Interpretable, z int, c int, D uint8) bool {
	if cc, ok := GetZonePaletteValue(e, z, c); ok {
		cc.Depth = D
		return SetZonePaletteValue(e, z, c, cc)
	}
	return false
}

func SetZonePaletteOffset(e interfaces.Interpretable, z int, c int, Offset int8) bool {
	if cc, ok := GetZonePaletteValue(e, z, c); ok {
		cc.Offset = Offset
		return SetZonePaletteValue(e, z, c, cc)
	}
	return false
}
