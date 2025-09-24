package presentation

import (
	"strings"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/types/glmath"
	"paleotronic.com/fmt"
	"paleotronic.com/octalyzer/video/font"
	"paleotronic.com/runestring"

	ogdl "gopkg.in/rveen/ogdl.v1"
)

func main() {

	s := `
# camera settings
init
	video
		hgrmode 1
		dhgrmode 2
		tintmode 3
`

	g := ogdl.FromString(s)

	v, _ := g.GetFloat64("init.camera.position[0]")

	g.Get("init").Add(mgl64.Vec3{1, 2, 5})

	fmt.Printf("camera.position[0] %f\n", v)

	fmt.Println(g.Text())

}

var configSections = []string{"boot", "video", "palette", "control", "camera", "audio", "hardware", "memory"}

type Presentation struct {
	Filepath string
	G        map[string]*ogdl.Graph
}

func (p *Presentation) LoadString(section string, s string) {
	p.G[section] = ogdl.FromString(s)
}

func (p *Presentation) LoadBytes(section string, data []byte) {
	p.G[section] = ogdl.FromBytes(data)
	if section == "boot" {
		fmt.RPrintf("Loading: \n%s\n\n", p.G[section].Text())
	}
}

func (p *Presentation) ReadVec3(section string, path string) (*glmath.Vector3, error) {
	out := &glmath.Vector3{}

	v1, e := p.graph(section).GetFloat64(path + "[0]")
	v2, e := p.graph(section).GetFloat64(path + "[1]")
	v3, e := p.graph(section).GetFloat64(path + "[2]")

	out[0], out[1], out[2] = v1, v2, v3

	return out, e

}

func (p *Presentation) WriteVec3(section string, path string, v *glmath.Vector3) error {

	p.graph(section).Set(path+"[0]", v[0])
	p.graph(section).Set(path+"[1]", v[1])
	p.graph(section).Set(path+"[2]", v[2])

	return nil

}

func (p *Presentation) WriteVideoPalette(section string, path string, pal *types.VideoPalette) error {

	p.graph(section).Set(path+".size", pal.Size())
	for i := 0; i < pal.Size(); i++ {
		cpath := fmt.Sprintf("%s.color%d", path, i)
		p.WriteVideoColor(section, cpath, pal.Get(i))
	}

	return nil

}

func (p *Presentation) ReadVideoPalette(section string, path string, pal *types.VideoPalette) error {

	size, e := p.graph(section).GetInt64(path + ".size")
	for pal.Size() < int(size) {
		pal.Add(&types.VideoColor{})
	}
	for i := 0; i < int(size); i++ {
		c := pal.Get(i)
		cpath := fmt.Sprintf("%s.color%d", path, i)
		p.ReadVideoColor(section, cpath, c)
	}
	return e

}

func (p *Presentation) WriteVideoColor(section string, path string, c *types.VideoColor) error {

	rgbapath := path + ".rgba"

	p.graph(section).Set(rgbapath+"[0]", c.Red)
	p.graph(section).Set(rgbapath+"[1]", c.Green)
	p.graph(section).Set(rgbapath+"[2]", c.Blue)
	p.graph(section).Set(rgbapath+"[3]", c.Alpha)
	p.graph(section).Set(path+".offset", c.Offset)
	p.graph(section).Set(path+".depth", c.Depth)

	return nil

}

func (p *Presentation) ReadVideoColor(section string, path string, c *types.VideoColor) error {

	rgbapath := path + ".rgba"

	r, e := p.graph(section).GetInt64(rgbapath + "[0]")
	g, e := p.graph(section).GetInt64(rgbapath + "[1]")
	b, e := p.graph(section).GetInt64(rgbapath + "[2]")
	a, e := p.graph(section).GetInt64(rgbapath + "[3]")
	o, e := p.graph(section).GetInt64(path + ".offset")
	d, e := p.graph(section).GetInt64(path + ".depth")

	c.Red = uint8(r)
	c.Green = uint8(g)
	c.Blue = uint8(b)
	c.Alpha = uint8(a)
	c.Offset = int8(o)
	c.Depth = uint8(d)

	return e

}

func (p *Presentation) WriteInt(section, path string, i int) error {

	p.graph(section).Set(path, i)

	return nil

}

func (p *Presentation) ReadInt(section, path string) (int, error) {

	i, e := p.graph(section).GetInt64(path)

	return int(i), e

}

func (p *Presentation) WriteFloat(section, path string, i float64) error {

	p.graph(section).Set(path, i)

	return nil

}

func (p *Presentation) ReadFloat(section, path string) (float64, error) {

	i, e := p.graph(section).GetFloat64(path)

	return i, e

}

func (p *Presentation) WriteString(section, path string, s string) error {

	p.graph(section).Set(path, s)

	return nil

}

func (p *Presentation) ReadString(section, path string) (string, error) {

	i, e := p.graph(section).GetString(path)

	return i, e

}

func (p *Presentation) WriteStringlist(section, path string, s []string) error {

	p.graph(section).Set(path, ogdl.New(""))

	for _, str := range s {
		p.graph(section).Get(path).Add(str)
	}

	return nil

}

func (p *Presentation) ReadStringlist(section, path string) ([]string, error) {

	var z []string

	i := 0
	for p.graph(section).Get(path).GetAt(i) != nil {
		z = append(z, p.graph(section).Get(path).GetAt(i).ThisString())
		i++
	}

	return z, nil

}

func (p *Presentation) Apply(context string, ent interfaces.Interpretable) error {

	p.applyVideoEffects(context, ent)
	time.Sleep(10 * time.Millisecond)
	p.applyCameraEffects(context, ent)
	p.applyInputState(context, ent)
	p.applyAudioState(context, ent)
	p.applyHardwareState(context, ent)
	p.applyOverlayState(context, ent)
	p.applyBackdropState(context, ent)
	p.applyBackdropPosState(context, ent)

	settings.FirstBoot[ent.GetMemIndex()] = false
	settings.IsPakBoot = false

	return nil

}

func (p *Presentation) applyVideoEffects(context string, ent interfaces.Interpretable) error {

	// hgr palette
	p.applyHGRPaletteData(context, ent)
	p.applyDHGRPaletteData(context, ent)
	p.applyXGRPaletteData(context, ent)
	p.applyLORESPaletteData(context, ent)
	p.applyDLGRPaletteData(context, ent)
	p.applyTEXTPaletteData(context, ent)
	p.applySHRPaletteData(context, ent)

	time.Sleep(5 * time.Millisecond)

	v, e := p.ReadVec3(context, ".video.layerpos")
	if e == nil {
		ent.GetProducer().SetMasterLayerPos(ent.GetMemIndex(), v[0], v[1])
	}

	// video.hgrmode
	mode, err := p.ReadInt("video", context+".video.hgrmode")
	if err == nil {
		fmt.Printf("HGR mode = %d\n", mode)
		settings.LastRenderModeHGR[ent.GetMemIndex()] = settings.VideoMode(mode ^ 1)
		if settings.FirstBoot[0] && ent.GetMemIndex() == 0 {
			settings.DefaultRenderModeHGR = settings.VideoMode(mode)
		}
		ent.GetMemoryMap().IntSetHGRRender(ent.GetMemIndex(), settings.VideoMode(mode))
		//time.Sleep(1 * time.Millisecond)
	}

	//if strings.HasPrefix(settings.SpecName[ent.GetMemIndex()], "apple2") {
	ntsc, err := p.ReadInt("video", context+".video.hgrntsc")
	if err == nil {
		fmt.Printf("HGR mode = %d\n", ntsc)
		settings.UseDHGRForHGR[ent.GetMemIndex()] = (ntsc != 0)
		h1, _ := ent.GetGFXLayerByID("HGR1")
		if h1 != nil {
			h1.SetDirty(true)
		}
		h2, _ := ent.GetGFXLayerByID("HGR2")
		if h2 != nil {
			h2.SetDirty(true)
		}
	} else {
		settings.UseDHGRForHGR[ent.GetMemIndex()] = false
		h1, _ := ent.GetGFXLayerByID("HGR1")
		if h1 != nil {
			h1.SetDirty(true)
		}
		h2, _ := ent.GetGFXLayerByID("HGR2")
		if h2 != nil {
			h2.SetDirty(true)
		}
	}

	blend, err := p.ReadInt("video", context+".video.vertblend")
	if err == nil {
		fmt.Printf("HGR mode = %d\n", blend)
		settings.UseVerticalBlend[ent.GetMemIndex()] = (blend != 0)
		h1, _ := ent.GetGFXLayerByID("HGR1")
		if h1 != nil {
			h1.SetDirty(true)
		}
		h2, _ := ent.GetGFXLayerByID("HGR2")
		if h2 != nil {
			h2.SetDirty(true)
		}
	} else {
		settings.UseVerticalBlend[ent.GetMemIndex()] = false
		h1, _ := ent.GetGFXLayerByID("HGR1")
		if h1 != nil {
			h1.SetDirty(true)
		}
		h2, _ := ent.GetGFXLayerByID("HGR2")
		if h2 != nil {
			h2.SetDirty(true)
		}
	}

	unified, err := p.ReadInt("video", context+".unified")
	if err == nil {
		settings.UnifiedRender[ent.GetMemIndex()] = (unified != 0)
		if settings.FirstBoot[0] && ent.GetMemIndex() == 0 {
			settings.UnifiedRenderGlobal = (unified != 0)
		}
	} else {
		settings.UnifiedRender[ent.GetMemIndex()] = false
		if settings.FirstBoot[0] && ent.GetMemIndex() == 0 {
			settings.UnifiedRenderGlobal = false
		}
	}
	//}

	// video.hgrmode
	mode, err = p.ReadInt("video", context+".video.shrmode")
	if err == nil {
		fmt.Printf("SHR mode = %d\n", mode)
		settings.LastRenderModeSHR[ent.GetMemIndex()] = settings.VideoMode(mode ^ 1)
		if settings.FirstBoot[0] && ent.GetMemIndex() == 0 {
			settings.DefaultRenderModeSHR = settings.VideoMode(mode)
		}
		ent.GetMemoryMap().IntSetSHRRender(ent.GetMemIndex(), settings.VideoMode(mode))
		//time.Sleep(1 * time.Millisecond)
	}

	mode, err = p.ReadInt("video", context+".video.zxmode")
	if err == nil {
		fmt.Printf("ZX mode = %d\n", mode)
		settings.LastRenderModeSpectrum[ent.GetMemIndex()] = settings.VideoMode(mode ^ 1)
		if settings.FirstBoot[0] && ent.GetMemIndex() == 0 {
			settings.DefaultRenderModeSpectrum = settings.VideoMode(mode)
		}
		ent.GetMemoryMap().IntSetSpectrumRender(ent.GetMemIndex(), settings.VideoMode(mode))
		//time.Sleep(1 * time.Millisecond)
	}

	mode, err = p.ReadInt("video", context+".video.grmode")
	if err == nil {
		fmt.Printf("GR mode = %d\n", mode)
		settings.LastRenderModeGR[ent.GetMemIndex()] = settings.VideoMode(mode ^ 1)
		if settings.FirstBoot[0] && ent.GetMemIndex() == 0 {
			settings.DefaultRenderModeGR = settings.VideoMode(mode)
		}
		ent.GetMemoryMap().IntSetGRRender(ent.GetMemIndex(), settings.VideoMode(mode))
		//time.Sleep(1 * time.Millisecond)
	}

	mode, err = p.ReadInt("video", context+".video.dhgrmode")
	if err == nil {
		ent.GetMemoryMap().IntSetDHGRRender(ent.GetMemIndex(), settings.VideoMode(mode))
	}

	mode, err = p.ReadInt("video", context+".video.dhgrhighbit")
	if err == nil {
		settings.DHGRHighBit[ent.GetMemIndex()] = settings.DHGRHighBitMode(mode)
	}

	mode, err = p.ReadInt("video", context+".video.voxeldepth")
	if err == nil {
		settings.LastVoxelDepth[ent.GetMemIndex()] = settings.VoxelDepth(mode ^ 1)
		ent.GetMemoryMap().IntSetVoxelDepth(ent.GetMemIndex(), settings.VoxelDepth(mode))
	}

	mode, err = p.ReadInt("video", context+".video.tintmode")
	if err == nil {
		settings.LastTintMode[ent.GetMemIndex()] = settings.VideoPaletteTint(mode ^ 1)
		ent.GetMemoryMap().IntSetVideoTint(ent.GetMemIndex(), settings.VideoPaletteTint(mode))
		if settings.FirstBoot[0] && ent.GetMemIndex() == 0 {
			settings.DefaultTintMode = settings.VideoPaletteTint(mode)
		}
	}

	if f, err := p.ReadFloat("video", context+".video.light.ambient"); err == nil {
		ent.GetMemoryMap().IntSetAmbientLevel(ent.GetMemIndex(), float32(f))
	}

	if f, err := p.ReadFloat("video", context+".video.light.diffuse"); err == nil {
		ent.GetMemoryMap().IntSetDiffuseLevel(ent.GetMemIndex(), float32(f))
	}

	if f, err := p.ReadFloat("video", context+".video.scanline"); err == nil {
		settings.ScanLineIntensity = float32(f)
	}

	if dsl, err := p.ReadInt("video", context+".video.scanlinedisable"); err == nil {
		settings.DisableScanlines = (dsl != 0)
	}

	if fs, err := p.ReadInt("video", context+".video.fullscreen"); err == nil && !settings.WindowApplied {
		if !settings.UseFullScreen {
			settings.Windowed = (fs == 0)
		}
		settings.WindowApplied = true
	}

	var bgcolor types.VideoColor
	if err := p.ReadVideoColor("video", context+".bgcolor", &bgcolor); err == nil {
		ent.GetMemoryMap().SetBGColor(
			ent.GetMemIndex(),
			bgcolor.Red,
			bgcolor.Green,
			bgcolor.Blue,
			bgcolor.Alpha,
		)
	}

	if fs, err := p.ReadInt("video", context+".video.font"); err == nil {
		fn := int(fs)
		index := ent.GetMemIndex()
		fonts := settings.AuxFonts[index]
		if fn >= 0 && fn < len(fonts) {
			fname := fonts[fn]
			f, err := font.LoadFromFile(fname)
			if err == nil {
				settings.DefaultFont[index] = f
			}
		}

	}

	mode, err = p.ReadInt("video", context+".highcontrast")
	if err == nil {
		settings.HighContrastUI = (mode != 0)
	}

	hover, err := p.ReadInt("video", context+".menuhover")
	if err == nil {
		settings.HamburgerOnHover = (hover != 0)
	}

	return nil
}

func (p *Presentation) applyCameraEffects(context string, ent interfaces.Interpretable) error {

	index := ent.GetMemIndex()
	mm := ent.GetMemoryMap()
	control := types.NewOrbitController(mm, index, 0)

	control.ResetALL()

	pos, err := p.ReadVec3("camera", context+".camera.position")
	if err == nil {
		control.SetPosition(pos)
	}

	lookat, err := p.ReadVec3("camera", context+".camera.lookat")
	if err == nil {
		control.SetTarget(lookat)
	}

	angle, err := p.ReadVec3("camera", context+".camera.angle")
	if err == nil {
		control.SetRotation(angle)
	}

	zoom, err := p.ReadFloat("camera", context+".camera.zoom")
	if err == nil {
		control.SetZoom(zoom)
	}

	fov, err := p.ReadFloat("camera", context+".camera.fov")
	if err == nil {
		control.SetFOV(fov)
	}

	near, err := p.ReadFloat("camera", context+".camera.near")
	if err == nil {
		control.SetNear(near)
	}

	far, err := p.ReadFloat("camera", context+".camera.far")
	if err == nil {
		control.SetFar(far)
	}

	panx, err := p.ReadFloat("camera", context+".camera.panx")
	if err == nil {
		control.SetPanX(panx)
	}

	pany, err := p.ReadFloat("camera", context+".camera.pany")
	if err == nil {
		control.SetPanY(pany)
	}

	//	if settings.FirstBoot {
	aspect, err := p.ReadFloat("video", context+".aspect")
	if aspect <= 0 {
		aspect = 1.46
	}
	if err == nil {
		for i := 0; i < 9; i++ {
			c := types.NewOrbitController(mm, index, i-1)
			c.SetAspect(aspect)
		}
	}
	//	}

	return nil
}

func (p *Presentation) saveCameraEffects(context string, ent interfaces.Interpretable) error {

	if settings.SkipCameraOnSave {
		return nil
	}

	index := ent.GetMemIndex()
	mm := ent.GetMemoryMap()
	cindex := mm.GetCameraConfigure(index)
	control := types.NewOrbitController(mm, index, cindex)

	pos := control.GetPosition()
	lookat := control.GetTarget()
	angle := control.GetAngle()

	p.WriteVec3("camera", context+".camera.position", pos)
	p.WriteVec3("camera", context+".camera.lookat", lookat)
	p.WriteVec3("camera", context+".camera.angle", angle)
	p.WriteFloat("camera", context+".camera.zoom", control.GetZoom())
	p.WriteFloat("camera", context+".camera.fov", control.GetFOV())
	p.WriteFloat("camera", context+".camera.near", control.GetNear())
	p.WriteFloat("camera", context+".camera.far", control.GetFar())
	p.WriteFloat("camera", context+".camera.panx", control.GetPanX())
	p.WriteFloat("camera", context+".camera.pany", control.GetPanY())
	p.WriteFloat("video", context+".aspect", control.GetAspect())

	return nil
}

func (p *Presentation) saveVideoEffects(context string, ent interfaces.Interpretable) error {

	p.WriteInt("video", context+".video.grmode", int(ent.GetMemoryMap().IntGetGRRender(ent.GetMemIndex())))
	p.WriteInt("video", context+".video.hgrmode", int(ent.GetMemoryMap().IntGetHGRRender(ent.GetMemIndex())))
	p.WriteInt("video", context+".video.dhgrmode", int(ent.GetMemoryMap().IntGetDHGRRender(ent.GetMemIndex())))
	p.WriteInt("video", context+".video.tintmode", int(ent.GetMemoryMap().IntGetVideoTint(ent.GetMemIndex())))
	p.WriteInt("video", context+".video.voxeldepth", int(ent.GetMemoryMap().IntGetVoxelDepth(ent.GetMemIndex())))
	p.WriteInt("video", context+".video.dhgrenhanced", int(settings.DHGRHighBit[ent.GetMemIndex()]))

	p.WriteFloat("video", context+".video.light.ambient", float64(ent.GetMemoryMap().IntGetAmbientLevel(ent.GetMemIndex())))
	p.WriteFloat("video", context+".video.light.diffuse", float64(ent.GetMemoryMap().IntGetDiffuseLevel(ent.GetMemIndex())))

	x, y := ent.GetProducer().GetMasterLayerPos(ent.GetMemIndex())
	p.WriteVec3("video", context+".video.layerpos", &glmath.Vector3{
		x, y, 0,
	})

	p.WriteFloat("video", context+".video.scanline", float64(settings.ScanLineIntensity))

	if settings.DisableScanlines {
		p.WriteInt("video", context+".video.scanlinedisable", 1)
	} else {
		p.WriteInt("video", context+".video.scanlinedisable", 0)
	}

	if settings.Windowed {
		p.WriteInt("video", context+".video.fullscreen", 0)
	} else {
		p.WriteInt("video", context+".video.fullscreen", 1)
	}

	var bgcolor types.VideoColor
	bgcolor.Red, bgcolor.Green, bgcolor.Blue, bgcolor.Alpha = ent.GetMemoryMap().GetBGColor(ent.GetMemIndex())
	p.WriteVideoColor("video", context+".bgcolor", &bgcolor)

	p.saveHGRPaletteData(context, ent)
	p.saveDHGRPaletteData(context, ent)
	p.saveXGRPaletteData(context, ent)
	p.saveLORESPaletteData(context, ent)
	p.saveDLGRPaletteData(context, ent)
	p.saveTEXTPaletteData(context, ent)
	p.saveSHRPaletteData(context, ent)

	return nil

}

func (p *Presentation) saveTEXTPaletteData(context string, ent interfaces.Interpretable) error {

	base := context + ".video.textpalette"

	ls, ok := ent.GetHUDLayerByID("TEXT")
	if ok {
		pal := ls.GetPalette()
		p.WriteVideoPalette("palette", base, &pal)
	}

	return nil
}

func (p *Presentation) applyTEXTPaletteData(context string, ent interfaces.Interpretable) error {
	base := context + ".video.textpalette"

	ls, ok := ent.GetHUDLayerByID("TEXT")
	if ok {
		pal := ls.GetPalette()
		if p.ReadVideoPalette("palette", base, &pal) == nil {
			ls.SetPalette(pal)
			//ls.SetRefresh(true)
			fmt.Println(pal.String())
		}
	}

	ls2, ok := ent.GetHUDLayerByID("TXT2")
	if ok {
		pal := ls2.GetPalette()
		if p.ReadVideoPalette("palette", base, &pal) == nil {
			ls.SetPalette(pal)
			//ls.SetRefresh(true)
			fmt.Println(pal.String())
		}
	}
	return nil
}

func (p *Presentation) saveHGRPaletteData(context string, ent interfaces.Interpretable) error {

	base := context + ".video.hgrpalette"

	ls, ok := ent.GetGFXLayerByID("HGR1")
	if ok {
		pal := ls.GetPalette()
		p.WriteVideoPalette("palette", base, &pal)
	}

	return nil
}

func (p *Presentation) applyHGRPaletteData(context string, ent interfaces.Interpretable) error {
	base := context + ".video.hgrpalette"

	ls, ok := ent.GetGFXLayerByID("HGR1")
	if ok {
		pal := ls.GetPalette()
		if p.ReadVideoPalette("palette", base, &pal) == nil {
			ls.SetPalette(pal)
			fmt.Println(pal.String())
		}
	}

	ls2, ok := ent.GetGFXLayerByID("HGR2")
	if ok {
		pal := ls2.GetPalette()
		if p.ReadVideoPalette("palette", base, &pal) == nil {
			ls2.SetPalette(pal)
			fmt.Println(pal.String())
		}
	}
	return nil
}

func (p *Presentation) saveDHGRPaletteData(context string, ent interfaces.Interpretable) error {

	base := context + ".video.dhgrpalette"

	ls, ok := ent.GetGFXLayerByID("DHR1")
	if ok {
		pal := ls.GetPalette()
		p.WriteVideoPalette("palette", base, &pal)
	}

	return nil
}

func (p *Presentation) applyDHGRPaletteData(context string, ent interfaces.Interpretable) error {
	base := context + ".video.dhgrpalette"

	ls, ok := ent.GetGFXLayerByID("DHR1")
	if ok {
		pal := ls.GetPalette()
		if p.ReadVideoPalette("palette", base, &pal) == nil {
			ls.SetPalette(pal)
			fmt.Println(pal.String())
		}
	}

	ls2, ok := ent.GetGFXLayerByID("DHR2")
	if ok {
		pal := ls2.GetPalette()
		if p.ReadVideoPalette("palette", base, &pal) == nil {
			ls2.SetPalette(pal)
			fmt.Println(pal.String())
		}
	}
	return nil
}

func (p *Presentation) saveXGRPaletteData(context string, ent interfaces.Interpretable) error {

	base := context + ".video.xgrpalette"

	ls, ok := ent.GetGFXLayerByID("XGR1")
	if ok {
		pal := ls.GetPalette()
		p.WriteVideoPalette("palette", base, &pal)
	}

	return nil
}

func (p *Presentation) applyXGRPaletteData(context string, ent interfaces.Interpretable) error {
	base := context + ".video.xgrpalette"

	ls, ok := ent.GetGFXLayerByID("XGR1")
	if ok {
		pal := ls.GetPalette()
		if p.ReadVideoPalette("palette", base, &pal) == nil {
			ls.SetPalette(pal)
			fmt.Println(pal.String())
		}
	}

	ls2, ok := ent.GetGFXLayerByID("XGR2")
	if ok {
		pal := ls2.GetPalette()
		if p.ReadVideoPalette("palette", base, &pal) == nil {
			ls2.SetPalette(pal)
			fmt.Println(pal.String())
		}
	}
	return nil
}

func (p *Presentation) saveLORESPaletteData(context string, ent interfaces.Interpretable) error {

	base := context + ".video.lorespalette"

	ls, ok := ent.GetGFXLayerByID("LOGR")
	if ok {
		pal := ls.GetPalette()
		p.WriteVideoPalette("palette", base, &pal)
	}

	return nil
}

func (p *Presentation) applyLORESPaletteData(context string, ent interfaces.Interpretable) error {
	base := context + ".video.lorespalette"
	ls, ok := ent.GetGFXLayerByID("LOGR")
	if ok {
		pal := ls.GetPalette()
		if p.ReadVideoPalette("palette", base, &pal) == nil {
			ls.SetPalette(pal)
			fmt.Println(pal.String())
		}
	}

	ls2, ok := ent.GetGFXLayerByID("LGR2")
	if ok {
		pal := ls2.GetPalette()
		if p.ReadVideoPalette("palette", base, &pal) == nil {
			ls2.SetPalette(pal)
			fmt.Println(pal.String())
		}
	}

	return nil
}

func (p *Presentation) applySHRPaletteData(context string, ent interfaces.Interpretable) error {
	base := context + ".video.superhirespalette"
	pal := types.NewVideoPalette()
	if p.ReadVideoPalette("palette", base, pal) == nil {
		data := pal.ToRGB12()
		for i, v := range data {
			settings.DefaultSHR320Palette[i] = v
		}
	}
	return nil
}

func (p *Presentation) saveSHRPaletteData(context string, ent interfaces.Interpretable) error {

	base := context + ".video.superhirespalette"

	pal := types.NewVideoPalette()
	pal.FromRGB12(settings.DefaultSHR320Palette[:])
	p.WriteVideoPalette("palette", base, pal)

	return nil
}

func (p *Presentation) saveDLGRPaletteData(context string, ent interfaces.Interpretable) error {

	base := context + ".video.dlorespalette"

	ls, ok := ent.GetGFXLayerByID("DLGR")
	if ok {
		pal := ls.GetPalette()
		p.WriteVideoPalette("palette", base, &pal)
	}

	return nil
}

func (p *Presentation) applyDLGRPaletteData(context string, ent interfaces.Interpretable) error {
	base := context + ".video.dlorespalette"
	ls, ok := ent.GetGFXLayerByID("DLGR")
	if ok {
		pal := ls.GetPalette()
		if p.ReadVideoPalette("palette", base, &pal) == nil {
			ls.SetPalette(pal)
			fmt.Println(pal.String())
		}
	}
	ls, ok = ent.GetGFXLayerByID("DLG2")
	if ok {
		pal := ls.GetPalette()
		if p.ReadVideoPalette("palette", base, &pal) == nil {
			ls.SetPalette(pal)
			fmt.Println(pal.String())
		}
	}
	return nil
}

func (p *Presentation) graph(section string) *ogdl.Graph {
	g, ok := p.G[section]
	if !ok {
		g = ogdl.NewPath("init")
		p.G[section] = g
	}
	return g
}

func NewPresentationState(ent interfaces.Interpretable, path string) (*Presentation, error) {

	g := make(map[string]*ogdl.Graph)

	p := &Presentation{G: g, Filepath: path}
	p.saveVideoEffects("init", ent)
	p.saveCameraEffects("init", ent)
	p.saveAudioState("init", ent)
	p.saveBackdropState("init", ent)
	p.saveBackdropPosState("init", ent)
	p.saveOverlayState("init", ent)
	p.saveInputState("init", ent)
	p.saveHardwareState("init", ent)

	return p, nil

}

func (p *Presentation) applyAudioState(context string, ent interfaces.Interpretable) {
	section := "audio"
	if v, e := p.ReadFloat(section, context+".speaker.volume"); e == nil {
		settings.SpeakerVolume[ent.GetMemIndex()] = v
	}
	if v, e := p.ReadFloat(section, context+".mockingboard.psg0balance"); e == nil {
		settings.MockingBoardPSG0Bal = v
	}
	if v, e := p.ReadFloat(section, context+".mockingboard.psg1balance"); e == nil {
		settings.MockingBoardPSG1Bal = v
	}
	// RecordIgnoreAudio
	if v, e := p.ReadInt(section, context+".recording.noaudio"); e == nil {
		settings.RecordIgnoreAudio[ent.GetMemIndex()] = (v != 0)
	}

}

func (p *Presentation) saveAudioState(context string, ent interfaces.Interpretable) {

	section := "audio"
	p.WriteString(section, context+".music.source", "")
	p.WriteInt(section, context+".music.leadin", 0)
	p.WriteInt(section, context+".music.fadein", 0)
	p.WriteFloat(section, context+".speaker.volume", settings.SpeakerVolume[ent.GetMemIndex()])
	p.WriteFloat(section, context+".mockingboard.psg0balance", settings.MockingBoardPSG0Bal)
	p.WriteFloat(section, context+".mockingboard.psg1balance", settings.MockingBoardPSG1Bal)
	// _, path, loop := ent.GetMemoryMap().IntGetRestalgiaPath(ent.GetMemIndex())
	// p.WriteString(section, context+".restalgia.source", path)
	// if loop {
	// 	p.WriteInt(section, context+".restalgia.loop", 1)
	// } else {
	// 	p.WriteInt(section, context+".restalgia.loop", 0)
	// }
}

func (p *Presentation) saveBackdropState(context string, ent interfaces.Interpretable) {

	section := "video"
	p.WriteString(section, context+".backdrop.source", "")
	p.WriteFloat(section, context+".backdrop.opacity", 1.0)
	p.WriteFloat(section, context+".backdrop.zoom", 16.0)
	p.WriteFloat(section, context+".backdrop.zrat", 0)
	p.WriteInt(section, context+".backdrop.camtrack", 0)
}

func (p *Presentation) applyBackdropState(context string, ent interfaces.Interpretable) {

	section := "video"
	var source string
	var opacity float64 = 1
	var zoom float64 = 16
	var zrat float64
	var camtrack bool

	if v, e := p.ReadString(section, context+".backdrop.source"); e == nil {
		source = v
	}
	if v, e := p.ReadFloat(section, context+".backdrop.opacity"); e == nil {
		opacity = v
	}
	if v, e := p.ReadFloat(section, context+".backdrop.zoom"); e == nil {
		zoom = v
	}
	if v, e := p.ReadFloat(section, context+".backdrop.zrat"); e == nil {
		zrat = v
	}
	if v, e := p.ReadInt(section, context+".backdrop.camtrack"); e == nil {
		camtrack = (v != 0)
	}

	if source != "" && !strings.HasPrefix(source, "/") && ent.GetWorkDir() != "" {
		source = "/" + strings.Trim(ent.GetWorkDir(), "/") + "/" + source
	}

	vm := ent.VM()

	_, err := vm.ExecuteRequest(
		"vm.gfx.setbackdrop",
		source,
		7,
		float32(opacity),
		float32(zoom),
		float32(zrat),
		camtrack)

	if err != nil {
		vm.Logf("Failed to set backdrop from presentation state")
		return
	}

	time.Sleep(5 * time.Millisecond)
}

func (p *Presentation) applyOverlayState(context string, ent interfaces.Interpretable) {

	section := "video"
	var source string

	vm := ent.VM()

	if v, e := p.ReadString(section, context+".overlay.source"); e == nil {
		source = v
	}

	if source != "" && !strings.HasPrefix(source, "/") && ent.GetWorkDir() != "" {
		source = "/" + strings.Trim(ent.GetWorkDir(), "/") + "/" + source
	}

	_, err := vm.ExecuteRequest("vm.gfx.setoverlay", source)
	if err != nil {
		vm.Logf("Failed to set overlay from presentation state")
		return
	}

}

func (p *Presentation) saveOverlayState(context string, ent interfaces.Interpretable) {

	section := "video"
	p.WriteString(section, context+".overlay.source", "")
}

func (p *Presentation) applyInputState(context string, ent interfaces.Interpretable) {

	section := "input"

	vm := ent.VM()

	if v, e := p.ReadString(section, context+".keyboard.paste"); e == nil {
		var delay time.Duration
		if ms, e := p.ReadInt(section, context+".keyboard.delayms"); e == nil {
			delay = time.Duration(ms) * time.Millisecond
		}
		time.AfterFunc(delay, func() {
			ent.SetPasteBuffer(runestring.Cast(v + "\r"))
		})
	}

	if v, e := p.ReadInt(section, context+".mouse"); e == nil {
		_, err := vm.ExecuteRequest("vm.input.setmousemode", settings.MouseMode(v))
		if err != nil {
			return
		}
	}

	if v, e := p.ReadInt(section, context+".joystick.axis0"); e == nil {
		_, err := vm.ExecuteRequest("vm.input.setjoystickaxis0", int(v))
		if err != nil {
			return
		}
	}

	if v, e := p.ReadInt(section, context+".joystick.axis1"); e == nil {
		_, err := vm.ExecuteRequest("vm.input.setjoystickaxis1", int(v))
		if err != nil {
			return
		}
	}

	if v, e := p.ReadInt(section, context+".uppercase"); e == nil {
		_, err := vm.ExecuteRequest("vm.input.setuppercaseonly", (v != 0))
		if err != nil {
			return
		}
	}

	if v, e := p.ReadInt(section, context+".joystick.reversex"); e == nil {
		settings.JoystickReverseX[ent.GetMemIndex()] = (v != 0)
	}

	if v, e := p.ReadInt(section, context+".joystick.reversey"); e == nil {
		settings.JoystickReverseY[ent.GetMemIndex()] = (v != 0)
	}

	// paste stuff
	if v, e := p.ReadInt(section, context+".paste.cps"); e == nil {
		settings.PasteCPS = v
	}

	if v, e := p.ReadInt(section, context+".paste.warp"); e == nil {
		settings.PasteWarp = v != 0
	}

}

func (p *Presentation) saveInputState(context string, ent interfaces.Interpretable) {

	section := "input"
	p.WriteInt(section, context+".mouse", int(settings.GetMouseMode()))

	p.WriteInt(section, context+".joystick.axis0", ent.GetMemoryMap().PaddleMap[ent.GetMemIndex()][0])
	p.WriteInt(section, context+".joystick.axis1", ent.GetMemoryMap().PaddleMap[ent.GetMemIndex()][1])

	v := 0
	if ent.GetMemoryMap().IntGetUppercaseOnly(ent.GetMemIndex()) {
		v = 1
	}
	p.WriteInt(section, context+".uppercase", v)

	//p.WriteStringlist(section, ".init.gerbils", []string{"a", "b", "c"})
}

func (p *Presentation) saveHardwareState(context string, ent interfaces.Interpretable) {

	section := "hardware"
	v := 0
	if settings.NoDiskWarp[ent.GetMemIndex()] {
		v = 1
	}
	p.WriteInt(section, context+".apple2.disk.nowarp", v)
	v = 0
	if settings.PreserveDSK {
		v = 1
	}
	p.WriteInt(section, context+".apple2.disk.nodskwoz", v)

	p.WriteInt(section, context+".printer.timeout", settings.PrintToPDFTimeoutSec)
}

func (p *Presentation) applyHardwareState(context string, ent interfaces.Interpretable) {

	section := "hardware"

	vm := ent.VM()

	//log.Printf("checking value for %s.cpu.speed", context)
	if v, e := p.ReadFloat(section, context+".cpu.speed"); e == nil {
		//log.Printf("Got value for %s.cpu.speed = %f", context, v)
		if v != 0 {
			time.AfterFunc(3*time.Second, func() {
				vm := ent.GetProducer().GetInterpreter(ent.GetMemIndex()).VM()
				_, err := vm.ExecuteRequest("vm.hardware.setcpuwarp", v)
				if err != nil {
					return
				}
			})
		}
	}

	if v, e := p.ReadInt(section, context+".apple2.disk.nowarp"); e == nil {
		_, err := vm.ExecuteRequest("vm.hardware.setdisablewarp", (v != 0))
		if err != nil {
			return
		}
	}

	if v, e := p.ReadInt(section, context+".serial.mode"); e == nil {
		_, err := vm.ExecuteRequest("vm.hardware.setserialmode", v)
		if err != nil {
			return
		}
	}

	if v, e := p.ReadInt(section, context+".serial.dipsw1"); e == nil {
		_, err := vm.ExecuteRequest("vm.hardware.setserialdipsw1", v)
		if err != nil {
			return
		}
	}

	if v, e := p.ReadInt(section, context+".serial.dipsw2"); e == nil {
		_, err := vm.ExecuteRequest("vm.hardware.setserialdipsw2", v)
		if err != nil {
			return
		}
	}

	if v, e := p.ReadInt(section, context+".apple2.disk.nodskwoz"); e == nil {
		_, err := vm.ExecuteRequest("vm.hardware.setpreservedsk", (v != 0))
		if err != nil {
			return
		}
	}

	if v, e := p.ReadInt(section, context+".liverecording"); e == nil {
		_, err := vm.ExecuteRequest("vm.hardware.setliverecording", (v != 0))
		if err != nil {
			return
		}
	}

	if v, e := p.ReadInt(section, context+".disablefractionalrewind"); e == nil {
		//settings.DisableFractionalRewindSpeeds = (v != 0)
		_, err := vm.ExecuteRequest("vm.hardware.setdisablefractionalrewind", (v != 0))
		if err != nil {
			return
		}
	}

	if v, e := p.ReadInt(section, context+".printer.timeout"); e == nil {
		// settings.PrintToPDFTimeoutSec = int(v)
		_, err := vm.ExecuteRequest("vm.hardware.setpdftimeoutsec", (v != 0))
		if err != nil {
			return
		}
	}

	index := ent.GetMemIndex()
	settings.MemLocks[index] = make(map[int]uint64)
	idx := 0
	a, e := p.ReadInt(section, fmt.Sprintf("%s.memory.locks.lock%d.address", context, idx))
	if e == nil {
		for e == nil && a != 0 {
			// lock exists
			v, e := p.ReadInt(section, fmt.Sprintf("%s.memory.locks.lock%d.value", context, idx))
			if e == nil {
				settings.MemLocks[index][a] = uint64(v)
			}
			idx++
			a, e = p.ReadInt(section, fmt.Sprintf("%s.memory.locks.lock%d.address", context, idx))
		}
	}

}

func (p *Presentation) applyBackdropPosState(context string, ent interfaces.Interpretable) {

	section := "video"
	var pos *glmath.Vector3

	if v, e := p.ReadVec3(section, context+".backdrop.position"); e == nil {
		pos = v
	}

	if pos != nil {
		ent.GetMemoryMap().IntSetBackdropPos(
			ent.GetMemIndex(),
			pos.X(),
			pos.Y(),
			pos.Z(),
		)
	}
}

func (p *Presentation) saveBackdropPosState(context string, ent interfaces.Interpretable) {

	section := "video"
	x, y, z := ent.GetMemoryMap().IntGetBackdropPos(ent.GetMemIndex())
	p.WriteVec3(section, context+".backdrop.position", &glmath.Vector3{x, y, z})
}

func (p *Presentation) GetBytes(section string) []byte {
	return []byte(p.graph(section).Text())
}
