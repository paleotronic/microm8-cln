package video

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/nfnt/resize"
	"paleotronic.com/accelimage"
	"paleotronic.com/core/hires"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
	"paleotronic.com/gl"
	"paleotronic.com/glumby"
	"paleotronic.com/log"
	"paleotronic.com/octalyzer/video/font"
)

const pixelBorder = 2

var psTexture *glumby.Texture
var psTextureNormal *glumby.Texture
var psTextureNarrow *glumby.Texture

var auxlorescolor = [16]int{0, 8, 1, 9, 2, 10, 3, 11, 4, 12, 5, 13, 6, 14, 7, 15}

var HGRPixelSize float32 = 4

const flSegments = 8
const maxStripeLength = 20
const maxColors = 24
const maxZones = 16

var GFXMBO = glumby.NewMBOManager()
var GFXMBOLastLayerID string

type GraphicsLayer struct {
	sync.Mutex
	BaseLayer
	PixelData             []int
	MaskIndex             int // Palette index to be treated as transparent
	PxMesh                *glumby.Mesh
	Points                bool
	Changed, DepthChanged bool
	PreviousBuffer        []uint64
	HControl              hires.HGRControllable
	VControl              *types.VectorBuffer
	CubeControl           *types.CubeScreen
	LastVectorSize        int
	LastVectorMap         map[types.VectorType][]*types.Vector
	DoubleRes             bool
	UsePaletteOffsets     bool
	lastFullFrame         time.Time
	//MBO                                *glumby.MBOManager //[flSegments]*glumby.MeshBufferObject
	usePointSprites                    bool
	d                                  []*Decal
	NumTextures                        int
	BitmapLayers                       []*image.RGBA
	BitmapDirty                        []bool
	ScreenTextures                     []*glumby.Texture
	ScreenTexW, ScreenTexH             int
	Pixel                              *image.RGBA
	Palette                            types.VideoPalette
	BitmapPosX, BitmapPosY, BitmapPosZ float32
	scandata                           [200][]int
	glWidth, glHeight                  float32
	framedata                          []uint64
	scanchanged                        [200]bool
	fscanchanged                       [200]bool
	sscanchanged                       [200]bool
	lastBounds                         types.LayerRect
	meshcache                          [maxZones * maxColors * maxStripeLength]*glumby.Mesh
	lastVoxelDepths                    [maxZones * maxColors]uint8
	lastVoxelDepthf                    [maxZones * maxColors]float32
	stdVoxelDepth                      uint8
	Force                              bool
	mixed                              bool
	cubepoints                         []types.CubePointColor
	DisableZones                       bool
	sc                                 *types.SpriteController
	subformat                          types.LayerSubFormat
	MinRefresh, MaxRefresh             int
}

func ror4bit(c int) int {
	return (c >> 1) | ((c & 1) << 3)
}

func rol4bit(c int) int {
	return ((c << 1) | ((c & 8) >> 3)) & 0xf
}

func NewGraphicsLayer(width, height int, glWidth, glHeight float32, format types.LayerFormat, data *memory.MemoryControlBlock, spec *types.LayerSpecMapped) *GraphicsLayer {
	// var ms runtime.MemStats
	// runtime.ReadMemStats(&ms)
	// log.Printf("Called NewGraphicsLayer for %s with mem at %d\n", spec.GetID(), ms.HeapInuse)
	index := data.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE

	// This is for DHGR style rendering of HGR
	if format == types.LF_HGR_WOZ && settings.UseDHGRForHGR[index] {
		width = 560 // 560 pixels
	}

	this := &GraphicsLayer{}
	this.Width = width
	this.Height = height
	this.glHeight = glHeight
	this.glWidth = glWidth
	this.Controller = types.NewOrbitController(data.GetMM(), index, 0)
	this.BitmapPosX = glWidth / 2
	this.BitmapPosY = glHeight / 2
	this.CalcDims()
	this.BoxDepth = this.UnitHeight * 2
	this.Spec = spec
	//this.MBO = glumby.NewMBOManager()
	this.usePointSprites = true
	this.ScreenTexH = SH
	this.ScreenTexW = SW
	this.glWidth = glWidth
	this.glHeight = glHeight
	this.lastBounds = types.LayerRect{0, 0, 0, 0}

	//this.Palette = avp
	this.Clear()
	this.PixelData = make([]int, width*height)

	this.Buffer = data
	this.PreviousBuffer = make([]uint64, data.Size)
	this.Format = format

	// added for sprite control
	this.sc = types.NewSpriteController(index, this.Buffer.GetMM(), memory.MICROM8_SPRITE_CONTROL_BASE)
	//this.sc.TestMode()

	if this.Format == types.LF_LOWRES_WOZ {
		if width == 80 {
			this.DoubleRes = true
		}
		this.UsePaletteOffsets = true
	}

	if this.Format == types.LF_HGR_WOZ {
		index := data.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
		if settings.UseDHGRForHGR[index] {
			this.HControl = hires.NewDHGRScreen(data)
		} else {
			this.HControl = hires.NewHGRScreen(data)
		}
		this.UsePaletteOffsets = true
		//		if settings.PureBoot(index) {
		//			this.Spec.SetSubFormat(types.LSF_COLOR_LAYER)
		//		}

		//if !settings.UnifiedRender {
		mm := this.Buffer.GetMM()
		mmu := mm.BlockMapper[index]
		l := &memory.MemoryListener{}
		l.Start = data.GStart[0] % memory.OCTALYZER_INTERPRETER_SIZE
		l.End = l.Start + 8192
		l.Label = spec.GetID()
		l.Target = this
		l.Type = memory.MA_WRITE
		mmu.RegisterListener(l)
		//}

		this.BoxDepth = this.UnitHeight * 4
		//fmt.Printf("Register listener: %v\n", l)

	}

	if this.Format == types.LF_DHGR_WOZ && spec.GetID() != "COMB" {
		index := data.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
		this.HControl = hires.NewDHGRScreen(data)
		this.UsePaletteOffsets = true
		//		if settings.PureBoot(index) {
		//			this.Spec.SetSubFormat(types.LSF_SINGLE_LAYER)
		//		}
		mm := this.Buffer.GetMM()
		mmu := mm.BlockMapper[index]
		l1 := &memory.MemoryListener{}
		l1.Start = data.GStart[0] % memory.OCTALYZER_INTERPRETER_SIZE
		l1.End = l1.Start + 8192
		l1.Label = spec.GetID()
		l1.Target = this
		l1.Type = memory.MA_WRITE
		mmu.RegisterListener(l1)
		this.BoxDepth = this.UnitHeight * 4
		//fmt.Printf("Register listener: %v\n", l1)
	}

	if this.Format == types.LF_SUPER_HIRES {
		this.BoxDepth = this.UnitHeight * 4
		this.HControl = hires.NewSuperHiResBuffer(data)
		this.UsePaletteOffsets = true
		this.ScreenTexH = 400
		this.ScreenTexW = 640

		index := data.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
		mm := this.Buffer.GetMM()
		// WriteListeners are tagged to a 4K block
		mm.RegisterWriteListener(index, 0x12000, this.VideoRAMUpdateSHR)
		mm.RegisterWriteListener(index, 0x13000, this.VideoRAMUpdateSHR)
		mm.RegisterWriteListener(index, 0x14000, this.VideoRAMUpdateSHR)
		mm.RegisterWriteListener(index, 0x15000, this.VideoRAMUpdateSHR)
		mm.RegisterWriteListener(index, 0x16000, this.VideoRAMUpdateSHR)
		mm.RegisterWriteListener(index, 0x17000, this.VideoRAMUpdateSHR)
		mm.RegisterWriteListener(index, 0x18000, this.VideoRAMUpdateSHR)
		mm.RegisterWriteListener(index, 0x19000, this.VideoRAMUpdateSHR)
	}

	if this.Format == types.LF_HGR_X {
		this.BoxDepth = this.UnitHeight * 4
		this.HControl = hires.NewIndexedVideoBuffer(width, height, data)
		this.UsePaletteOffsets = true
	}

	if this.Format == types.LF_VECTOR {

		var rsh [256]memory.ReadSubscriptionHandler
		var esh [256]memory.ExecSubscriptionHandler
		var wsh [256]memory.WriteSubscriptionHandler

		index := data.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
		globalbase := index * memory.OCTALYZER_INTERPRETER_SIZE
		base := data.GStart[0] - globalbase

		this.VControl = types.NewVectorBufferMapped(
			base,
			0x10000,
			memory.NewMappedRegionFromHint(
				data.GetMM(),
				globalbase,
				base,
				data.Size,
				this.Format.String(),
				"VCTR",
				rsh,
				esh,
				wsh,
			),
		)
	}

	if this.Format == types.LF_CUBE_PACKED {
		index := data.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
		globalbase := index * memory.OCTALYZER_INTERPRETER_SIZE
		base := data.GStart[0] - globalbase

		this.BoxDepth = this.UnitHeight
		this.PxMesh = nil

		this.CubeControl = types.NewCubeScreen(
			base,
			0x10000,
			data.GetMM().GetHintedMemorySlice(index, spec.GetID()),
		)
		if this.Spec.GetActive() {
			this.MakeUpdatesCUBES()
		}
	}

	// Setup pixel texture for plotting
	pxtmp, err := font.LoadPNG(fontpixel)
	if err != nil {
		panic(err)
	}
	bounds := this.GetPixelBox(0, 0)
	w, h := uint(bounds.Dx()), uint(bounds.Dy())
	this.Pixel = accelimage.ImageRGBA(resize.Resize(w, h, pxtmp, resize.Lanczos3))

	this.SetSubFormat(spec.GetSubFormat(), true)

	this.Palette = spec.GetPalette()

	// Initial fetch so we have scandata
	this.Fetch()

	return this
}

func (this *GraphicsLayer) VideoRAMUpdateSHR(addr int, value *uint64) {

	this.Lock()
	defer this.Unlock()

	offset := addr - this.Buffer.GStart[0]

	// var checkMinMax = func(y int) {
	// 	if y < this.MinRefresh {
	// 		this.MinRefresh = y
	// 		if this.MaxRefresh < 0 {
	// 			this.MaxRefresh = y
	// 		}
	// 	}
	// 	if y > this.MaxRefresh {
	// 		this.MaxRefresh = y
	// 		if this.MinRefresh >= this.Height {
	// 			this.MinRefresh = y
	// 		}
	// 	}
	// }

	// log2.Printf("memory update : offset = %.5x", offset)

	if offset < 0x7d00 {
		y := offset / 160
		if y >= 0 && y < 200 {
			this.scanchanged[y] = true
			// checkMinMax(y)
		}
		this.Force = false
	} else if offset < 0x7e00 {
		y := offset - 0x7d00
		if y >= 0 && y < 200 {
			this.scanchanged[y] = true
			// checkMinMax(y)
		}
		this.Force = false
	} else {
		this.Force = true // palette changed
	}

}

func (this *GraphicsLayer) ProcessEvent(label string, addr int, value *uint64, action memory.MemoryAction) (bool, bool) {

	if this.Format == types.LF_HGR_WOZ {
		//fmt.Printf("Update to layer %s\n", this.Spec.GetID())
		offset := addr - (this.Buffer.GStart[0] % memory.OCTALYZER_INTERPRETER_SIZE)
		if offset%128 >= 120 {
			return true, true
		}
		y := this.HControl.OffsetToScanline(offset)
		this.scanchanged[y] = true
	} else if this.Format == types.LF_DHGR_WOZ {
		offset := addr - (this.Buffer.GStart[0] % memory.OCTALYZER_INTERPRETER_SIZE)
		if offset%128 >= 120 {
			return true, true
		}
		y := this.HControl.OffsetToScanline(offset)
		this.scanchanged[y] = true
	}

	return true, true
}

func (this *GraphicsLayer) Done() {
	if this.NumTextures > 0 {
		// Need to clean these up..
		var textures = make([]uint32, this.NumTextures)
		for i, tex := range this.ScreenTextures {
			textures[i] = tex.Handle()
		}
		gl.DeleteTextures(int32(this.NumTextures), &textures[0])
	}
}

func (this *GraphicsLayer) SetSubFormat(lsf types.LayerSubFormat, force bool) {

	if lsf == this.Spec.GetSubFormat() && !force {
		return
	}

	this.Force = true

	//fmt.Println("Sf change")

	if this.NumTextures > 0 {
		// Need to clean these up..
		var textures = make([]uint32, this.NumTextures)
		for i, tex := range this.ScreenTextures {
			textures[i] = tex.Handle()
		}
		gl.DeleteTextures(int32(this.NumTextures), &textures[0])
	}

	//	this.MBO = glumby.NewMBOManager()

	if lsf == types.LSF_SINGLE_LAYER {

		numcols := this.Spec.GetPaletteSize()

		if lsf == types.LSF_SINGLE_LAYER {
			numcols = 1
		}

		this.NumTextures = numcols

		this.d = make([]*Decal, numcols)
		this.BitmapLayers = make([]*image.RGBA, numcols)
		this.ScreenTextures = make([]*glumby.Texture, numcols)
		this.BitmapDirty = make([]bool, numcols)

		for i := 0; i < numcols; i++ {
			this.ScreenTextures[i] = glumby.NewTextureBlank(this.ScreenTexW+pixelBorder, this.ScreenTexH, color.RGBA{0, 0, 0, 0})
			this.d[i] = NewDecal(float32(this.ScreenTexW), float32(this.ScreenTexH))
			this.d[i].Texture = this.ScreenTextures[i]
			this.d[i].Mesh = GetPlaneAsTrianglesInv(this.glWidth, this.glHeight)

			this.BitmapLayers[i] = image.NewRGBA(image.Rect(0, 0, this.ScreenTexW+pixelBorder, this.ScreenTexH))
			draw.Draw(this.BitmapLayers[i], this.BitmapLayers[i].Bounds(), image.Transparent, image.ZP, draw.Src)

			this.d[i].Texture.SetSourceSame(this.BitmapLayers[i]) // force update
		}

	}

	if lsf != types.LSF_SINGLE_LAYER {
		this.Palette = this.Spec.GetPalette()
	}

	// force refresh
	for y := 0; y < this.Height; y++ {
		this.scanchanged[y] = true
	}

	//fmt.Printf("Confirm subformat %d\n", lsf)
	this.Spec.SetSubFormat(lsf)

	v := this.Spec.Mm.IntGetVoxelDepth(this.Spec.Base / memory.OCTALYZER_INTERPRETER_SIZE)
	this.SetVoxelDepth(v)

	this.PxMesh = nil

	for i, _ := range this.meshcache {
		this.meshcache[i] = nil
	}

	this.Changed = true
	for i, _ := range this.PixelData {
		this.PixelData[i] = 0x00
	}

	this.Force = true

	//fmt.Printf("Subformat is %d\n", lsf)
}

func (this *GraphicsLayer) SetVoxelDepth(v settings.VoxelDepth) {

	lsf := this.Spec.GetSubFormat()

	mult := float32(v) + 1

	this.BoxDepth = 1 * this.BoxHeight * mult
	if lsf == types.LSF_VOXELS && this.Spec.GetFormat() != types.LF_LOWRES_WOZ {
		this.BoxDepth = 2 * this.BoxHeight * mult
	}

	this.PxMesh = nil

	for i, _ := range this.meshcache {
		this.meshcache[i] = nil
	}

	if this.Format != types.LF_LOWRES_WOZ {
		for i, _ := range this.PixelData {
			this.PixelData[i] = 0
		}
	}

	for y, _ := range this.fscanchanged {
		this.fscanchanged[y] = true
	}

	for y, _ := range this.scanchanged {
		this.scanchanged[y] = true
	}

	this.Changed = true
	this.DepthChanged = true
	this.stdVoxelDepth = uint8(1+v) * 10

}

func (d *GraphicsLayer) GetPixelBox(x, y int) image.Rectangle {
	pw := float32(d.ScreenTexW) / float32(d.Width)
	ph := float32(d.ScreenTexH) / float32(d.Height)
	x0 := int(float32(x) * pw)
	y0 := int(float32(y) * ph)
	x1 := int(float32(x+1) * pw)
	y1 := int(float32(y+1) * ph)
	return image.Rect(x0, y0, x1, y1)
}

func (d *GraphicsLayer) GetScanLineBox(x, y int) image.Rectangle {
	ph := float32(d.ScreenTexH) / float32(d.Height)
	y0 := int(float32(y) * ph)
	y1 := int(float32(y+1) * ph)
	return image.Rect(0, y0, d.ScreenTexW, y1)
}

func (d *GraphicsLayer) Clear() {
	d.PixelData = make([]int, d.Width*d.Height)
	d.Changed = false
}

// GetHashKey returns a unique key for a given 3D point
func (d *GraphicsLayer) GetHashKey(x, y, z int) int {
	return x + (y * d.Width) + (z * d.Width * d.Height)
}

// Unpack the key
func (d *GraphicsLayer) GetPointFromHashKey(hashKey int) (int, int, int) {
	return (hashKey % d.Width), (hashKey / d.Width) % d.Height, 0
}

func (d *GraphicsLayer) PlotScanLinePoints(y, c int) {
	for x := 0; x < d.Width; x++ {
		d.Plot(x, y, c)
	}
}

func (d *GraphicsLayer) PlotScanLine(y, c int) {
	target := d.GetScanLineBox(0, y)
	accelimage.FillRGBA(d.BitmapLayers[0], target, color.RGBA{0, 0, 0, 0})
}

func RGB12ToRGBA(c int) color.RGBA {
	var r, g, b, a uint8
	r = uint8(c>>4) | 0x0f
	g = uint8(c&0xf0) | 0x0f
	b = (uint8(c&0x0f) << 4) | 0x0f
	a = uint8(0xff)
	if r == 0xf && g == 0xf && b == 0xf {
		a = 0x00
	}
	return color.RGBA{
		r, g, b, a,
	}
}

func (d *GraphicsLayer) getCol(memindex int, x, y, c int) (color.RGBA, bool) {
	var fg color.RGBA
	var col *settings.VideoColor

	var isSolid bool
	if d.Format == types.LF_SUPER_HIRES {
		fg = RGB12ToRGBA(c)
		isSolid = (c != 0)
	} else {
		isSolid = !d.IsTransparent(c, &d.Palette)
		if zc := settings.ColorZone[memindex][d.Format]; zc != nil && !d.DisableZones {
			col = zc.GetColorAt(x, y, c)
			if col != nil {
				fg = col.ToColorRGBA()
			} else {
				fg = d.Palette.Get(c).ToColorRGBA()
			}
		} else {
			fg = d.Palette.Get(c).ToColorRGBA()
		}
	}
	return fg, isSolid
}

func (d *GraphicsLayer) PlotPixelBlended(x, y, c, ct, cb int, b1, b2, w1, w2 int) {

	sf := d.subformat

	if sf != types.LSF_SINGLE_LAYER {
		return
	}

	var memindex = d.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE

	var index int

	var fg, fgt, fgb color.RGBA
	var isSolid bool

	fg, isSolid = d.getCol(memindex, x, y, c)

	if c != b1 && c != b2 && c != w1 && c != w2 {
		// non black/white main color
		var r, g, b int = int(fg.R), int(fg.G), int(fg.B)
		var count = 1
		if y < d.Height-1 && cb != b1 && cb != b2 && cb != w1 && cb != w2 {
			// non black/white top color
			fgb, _ = d.getCol(memindex, x, y+1, cb)
			r += int(fgb.R)
			g += int(fgb.G)
			b += int(fgb.B)
			count++
		} else {
			if y > 0 && ct != b1 && ct != b2 && ct != w1 && ct != w2 {
				// non black/white top color
				fgt, _ = d.getCol(memindex, x, y-1, ct)
				r += int(fgt.R)
				g += int(fgt.G)
				b += int(fgt.B)
				count++
			}
		}
		if count > 1 {
			fg.R = uint8(r / count)
			fg.G = uint8(g / count)
			fg.B = uint8(b / count)
		}
	}

	target := d.GetPixelBox(x, y)

	// Mark new pixel or clear it
	if isSolid {

		accelimage.FillRGBAWithFilter(
			d.BitmapLayers[index],
			target,
			fg,
			scanFunc,
		)

	} else {
		accelimage.FillRGBA(d.BitmapLayers[index], target, color.RGBA{0, 0, 0, 0})
	}

	d.BitmapDirty[index] = true

}

func (d *GraphicsLayer) PlotPixel(x, y, c int) {

	sf := d.subformat

	if sf != types.LSF_SINGLE_LAYER {
		return
	}

	var memindex = d.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE

	var index int
	var col *settings.VideoColor

	var fg color.RGBA

	var isSolid bool
	if d.Format == types.LF_SUPER_HIRES {
		fg = RGB12ToRGBA(c)
		isSolid = (c != 0)
	} else {
		isSolid = !d.IsTransparent(c, &d.Palette)

		if zc := settings.ColorZone[memindex][d.Format]; zc != nil && !d.DisableZones {
			col = zc.GetColorAt(x, y, c)
			if col != nil {
				fg = col.ToColorRGBA()
			} else {
				fg = d.Palette.Get(c).ToColorRGBA()
			}
		} else {
			fg = d.Palette.Get(c).ToColorRGBA()
		}
	}

	target := d.GetPixelBox(x, y)

	// Mark new pixel or clear it
	if isSolid {

		accelimage.FillRGBAWithFilter(
			d.BitmapLayers[index],
			target,
			fg,
			scanFunc,
		)

	} else {
		accelimage.FillRGBA(d.BitmapLayers[index], target, color.RGBA{0, 0, 0, 0})
	}

	d.BitmapDirty[index] = true

}

func (d *GraphicsLayer) Plot(x, y, c int) {
	var idx int

	idx = y*d.Width + x
	if idx >= d.Width*d.Height {
		return
	}

	oc := d.PixelData[idx]

	// Only process if different
	if oc == c {
		return
	}

	d.PixelData[idx] = c
	d.Changed = true

}

func (d *GraphicsLayer) Fill(c int, p *types.VideoPalette) {

	if d.IsTransparent(c, p) {
		d.Clear()
		return
	}

	for y := 0; y < d.Height; y++ {
		for x := 0; x < d.Width; x++ {
			d.Plot(x, y, c)
		}
	}

}

func (d *GraphicsLayer) Fetch() {

	if settings.UnifiedRender[d.Buffer.Index] && d.Format != types.LF_SUPER_HIRES {
		return
	}

	if d.Format == types.LF_DHGR_WOZ {
		d.FetchUpdatesWozDHGR()
	} else if d.Format == types.LF_HGR_WOZ {
		d.FetchUpdatesWozHGR()
	} else if d.Format == types.LF_HGR_X {
		d.FetchUpdatesXHGR()
	} else if d.Format == types.LF_LOWRES_WOZ || d.Format == types.LF_LOWRES_LINEAR {
		d.FetchUpdatesLORES()
	} else if d.Format == types.LF_VECTOR {
		d.FetchUpdatesVECTOR()
	} else if d.Format == types.LF_SUPER_HIRES {
		d.FetchUpdatesSHR()
	} else if d.Format == types.LF_SPECTRUM_0 {
		d.FetchUpdatesZXVideo()
	}
}

func (d *GraphicsLayer) FetchUpdatesXHGR() {
	var scanstart int
	var raw []uint64

	for y := 0; y < d.Height; y++ {
		scanstart = y * 70
		raw = d.Buffer.ReadSlice(scanstart, scanstart+70)
		// check each scanline for updates
		d.scandata[y] = d.HControl.ColorsForScanLine(raw, d.Spec.GetMono())
	}
}

func (this *GraphicsLayer) FetchUpdatesSHR() {
	// Capture framedata and palette at Fetch cycle
	// this.Lock()
	// defer this.Unlock()

	// shr, ok := this.HControl.(*hires.SuperHiResBuffer)
	// if ok {
	// 	shr.LoadPaletteCache()
	// }
	// this.framedata = this.Buffer.ReadSliceCopy(0, this.Buffer.Size)
	//this.fscanchanged = this.scanchanged
	// for i, _ := range this.scanchanged {
	// 	this.scanchanged[i] = false
	// }
}

func (d *GraphicsLayer) FetchUpdatesLORES() {
	tmp := d.Buffer.ReadSlice(0, 4096)
	d.framedata = make([]uint64, len(tmp))
	for i, v := range tmp {
		d.framedata[i] = v
	}
}

func (d *GraphicsLayer) FetchUpdatesVECTOR() {

}

func (d *GraphicsLayer) Update() {

	if settings.UnifiedRender[d.Buffer.Index] && d.Format != types.LF_SUPER_HIRES {
		return
	}

	var tintstate bool = d.TintChanged

	d.subformat = d.Spec.GetSubFormat()

	// force updates on zone changes
	memindex := d.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
	zc := settings.ColorZone[memindex][d.Format]
	if zc != nil && zc.IsUpdated() {
		d.Force = true
		zc.SetUpdate(false)
	}

	if d.TintChanged {
		d.TintChanged = false
		//log2.Println("TINT HAS CHANGED", d.TintChanged)
		d.Palette = d.Spec.GetPalette()
		if d.Tint != nil {
			if d.Format == types.LF_SUPER_HIRES {
				shr, ok := d.HControl.(*hires.SuperHiResBuffer)
				if ok {
					shr.SetTint(
						float32(d.Tint.Red)/255,
						float32(d.Tint.Green)/255,
						float32(d.Tint.Blue)/255,
						true,
					)
				}
			} else {
				pa := d.Palette.Desaturate().Tint(d.Tint.Red, d.Tint.Green, d.Tint.Blue)
				d.Palette = *pa
			}
			d.Tint = nil
		} else {
			if d.Format == types.LF_SUPER_HIRES {
				shr, ok := d.HControl.(*hires.SuperHiResBuffer)
				if ok {
					shr.SetTint(
						1,
						1,
						1,
						false,
					)
				}
			}
		}
		d.TintChanged = false
		if d.Format == types.LF_DHGR_WOZ || d.Format == types.LF_HGR_WOZ || d.Format == types.LF_SUPER_HIRES {
			for i, _ := range d.scanchanged {
				d.scanchanged[i] = true
			}
		} else if d.Format == types.LF_LOWRES_WOZ {
			d.Force = true
		}
	}

	if d.Format == types.LF_SUPER_HIRES {
		d.MakeUpdatesSHR()
	} else if d.Format == types.LF_SPECTRUM_0 {
		d.MakeUpdatesZXVideo()
	} else if d.Format == types.LF_CUBE_PACKED {
		d.MakeUpdatesCUBES()
	} else if d.Format == types.LF_DHGR_WOZ {
		d.MakeUpdatesWozDHGR()
	} else if d.Format == types.LF_HGR_WOZ {
		if settings.UseDHGRForHGR[memindex] {
			d.MakeUpdatesWozHGRViaDHGR()
		} else {
			d.MakeUpdatesWozHGR()
		}
	} else if d.Format == types.LF_HGR_X {
		d.MakeUpdatesXHGR()
	} else if d.Format == types.LF_LOWRES_WOZ || d.Format == types.LF_LOWRES_LINEAR {
		if d.subformat == types.LSF_SINGLE_LAYER {
			d.MakeUpdatesWozLORES()
		} else {
			d.MakeUpdatesLORES()
		}
	} else if d.Format == types.LF_VECTOR {
		d.MakeUpdatesVECTOR()
	}

	if tintstate && !d.Changed {
		d.Changed = true
	}

	d.Force = false

}

func (v *GraphicsLayer) MakeUpdatesCUBES() {
	size := v.CubeControl.Size()

	if size != v.LastVectorSize || v.CubeControl.Changed() {
		//log.Printf("Cube count changed: %d -> %d", v.LastVectorSize, size)
		v.cubepoints = v.CubeControl.GetMap()
	}

	v.LastVectorSize = size
}

func lsfIn(l types.LayerSubFormat, list []types.LayerSubFormat) bool {
	for _, v := range list {
		if v == l {
			return true
		}
	}
	return false
}

func (this *GraphicsLayer) RedoMeshes() {

	if this.Spec == nil {
		return
	}

	numcols := this.Spec.GetPaletteSize()
	for i := 0; i < numcols; i++ {
		if this.d[i] != nil {
			this.d[i].Mesh = GetPlaneAsTrianglesInv(this.glWidth, this.glHeight)
		}
	}
}

func (this *GraphicsLayer) CalcDims() {

	// as := this.Controller.GetAspect()
	// fmt.Printf("Aspect now %.2f\n", as)

	// this.glWidth = float32(float64(types.CHEIGHT) * as)

	this.UnitWidth = this.glWidth / float32(this.Width)
	this.UnitHeight = this.glHeight / float32(this.Height)
	this.UnitDepth = this.UnitHeight
	// this.BitmapPosX = types.CWIDTH / 2
	// this.BitmapPosY = this.glHeight / 2
	this.UnitWidth = this.glWidth / float32(this.Width)
	this.UnitHeight = this.glHeight / float32(this.Height)
	this.BoxWidth = this.UnitWidth
	this.BoxHeight = this.UnitHeight

	//this.RedoMeshes()
	for i, _ := range this.d {
		this.d[i].Mesh = GetPlaneAsTrianglesInv(this.glWidth, this.glHeight)
	}

	//this.Controller.SetPos(float64(this.glWidth/2), float64(this.glHeight/2), types.CDIST*types.GFXMULT)
	//this.Controller.SetLookAt(mgl64.Vec3{float64(this.glWidth / 2), float64(this.glHeight / 2), 0})

	this.PxMesh = nil

	for i, _ := range this.meshcache {
		this.meshcache[i] = nil
	}

	this.Changed = true
}

func (d *GraphicsLayer) Render() {

	d.Lock()
	defer d.Unlock()

	//fmt.Printf("%s, lsf=%d\n", d.Spec.GetID(), d.Spec.GetSubFormat())
	d.DisableZones = d.Spec.GetMono() || d.Tint != nil

	gl.Enable(gl.TEXTURE_2D)

	if d.lastAspect != d.Controller.GetAspect() {
		d.lastAspect = d.Controller.GetAspect()
		d.glWidth = d.glHeight * float32(d.lastAspect)
		d.CalcDims()
	}

	if d.Format == types.LF_SUPER_HIRES {
		if lsfIn(d.subformat, []types.LayerSubFormat{types.LSF_VOXELS}) {
			d.RenderCubes()
		} else if lsfIn(d.subformat, []types.LayerSubFormat{types.LSF_GREY_LAYER, types.LSF_COLOR_LAYER, types.LSF_AMBER_LAYER, types.LSF_GREEN_LAYER, types.LSF_SINGLE_LAYER}) {
			d.RenderTextureLayers()
		} else {
			d.RenderPoints(d.usePointSprites)
		}
	} else if d.Format == types.LF_DHGR_WOZ {
		//d.MakeUpdatesWozDHGR()
		if lsfIn(d.subformat, []types.LayerSubFormat{types.LSF_VOXELS}) {
			d.RenderCubes()
		} else if lsfIn(d.subformat, []types.LayerSubFormat{types.LSF_GREY_LAYER, types.LSF_COLOR_LAYER, types.LSF_AMBER_LAYER, types.LSF_GREEN_LAYER, types.LSF_SINGLE_LAYER}) {
			d.RenderTextureLayers()
		} else {
			d.RenderPoints(d.usePointSprites)
		}
	} else if d.Format == types.LF_SPECTRUM_0 {
		//d.MakeUpdatesWozDHGR()
		if lsfIn(d.subformat, []types.LayerSubFormat{types.LSF_VOXELS}) {
			d.RenderCubes()
		} else if lsfIn(d.subformat, []types.LayerSubFormat{types.LSF_GREY_LAYER, types.LSF_COLOR_LAYER, types.LSF_AMBER_LAYER, types.LSF_GREEN_LAYER, types.LSF_SINGLE_LAYER}) {
			d.RenderTextureLayers()
		} else {
			d.RenderPoints(d.usePointSprites)
		}
	} else if d.Format == types.LF_HGR_WOZ {
		//		//d.MakeUpdatesWozHGR()
		//fmt.Printf("lsf=%d\n", d.Spec.GetSubFormat())
		if d.subformat == types.LSF_VOXELS {
			d.RenderCubes()
		} else if lsfIn(d.subformat, []types.LayerSubFormat{types.LSF_GREY_LAYER, types.LSF_COLOR_LAYER, types.LSF_AMBER_LAYER, types.LSF_GREEN_LAYER, types.LSF_SINGLE_LAYER}) {
			d.RenderTextureLayers()
		} else {
			d.RenderPoints(d.usePointSprites)
			//d.RenderCubes()
		}
	} else if d.Format == types.LF_HGR_X {
		//d.MakeUpdatesXHGR()
		if lsfIn(d.subformat, []types.LayerSubFormat{types.LSF_VOXELS}) {
			d.RenderCubes()
		} else if lsfIn(d.subformat, []types.LayerSubFormat{types.LSF_GREY_LAYER, types.LSF_COLOR_LAYER, types.LSF_AMBER_LAYER, types.LSF_GREEN_LAYER, types.LSF_SINGLE_LAYER}) {
			d.RenderTextureLayers()
		} else {
			d.RenderPoints(d.usePointSprites)
		}
	} else if d.Format == types.LF_LOWRES_WOZ || d.Format == types.LF_LOWRES_LINEAR {
		//d.MakeUpdatesLORES()
		log.Println("mode =", d.subformat.String())
		if d.subformat == types.LSF_SINGLE_LAYER {
			log.Println("LORES: texture")
			d.RenderTextureLayers()
		} else {
			log.Println("LORES: cube")
			d.RenderCubes()
		}
	} else if d.Format == types.LF_VECTOR {
		//d.MakeUpdatesVECTOR()
		d.RenderVectors()
	} else if d.Format == types.LF_CUBE_PACKED {
		d.RenderCubeScreen()
	}

	d.PosChanged = false

	gl.Disable(gl.TEXTURE_2D)
}

func (v *GraphicsLayer) RenderCubeScreen() {

	var offx, offy, offz float32
	offx = float32(v.Width) / 2
	offy = float32(v.Height) / 2
	offz = 0

	lx, ly, lz := v.Spec.GetPos()
	offx += float32(lx)
	offy += float32(ly)
	offz += float32(lz)

	gl.Color3f(1, 1, 1)
	//    gl.Begin(gl.LINES)
	gl.LineWidth(4)

	if v.PxMesh == nil {
		v.PxMesh = GetCubeAsTriangles(v.BoxWidth, v.BoxHeight, v.BoxHeight)
	}
	v.PxMesh.Texture.Unbind()

	_, hy, hz := (v.UnitWidth * float32(v.Width) / 2), (v.UnitHeight * float32(v.Height) / 2), (v.UnitDepth * float32(v.Height) / 2)

	pal := v.Palette //v.Spec.GetPalette()

	glumby.MeshBuffer_Begin(gl.TRIANGLES)
	for _, cp := range v.cubepoints {

		ox, oy, oz := float32(cp.P.X)-offx, float32(uint8(v.Height)-cp.P.Y)-offy, -float32(cp.P.Z)-offz

		ox = ox*v.UnitWidth + v.BitmapPosX + float32(v.MPos.X)
		oy = oy*v.UnitHeight + hy + float32(v.MPos.Y)
		oz = oz*v.UnitDepth + hz + float32(v.MPos.Z)

		c := pal.Get(int(cp.C))

		r, g, b, a := c.ToFRGBA()

		v.PxMesh.SetColor(r, g, b, a)

		v.PxMesh.DrawWithMeshBuffer(ox, oy, oz)
	}
	glumby.MeshBuffer_End()

}

func calcNormalForTriangle(p1, p2, p3 mgl32.Vec3) mgl32.Vec3 {
	u := p2.Sub(p1)
	v := p3.Sub(p1)

	/*
		Nx = UyVz - UzVy
		Ny = UzVx - UxVz
		Nz = UxVy - UyVx
	*/
	n := mgl32.Vec3{
		u[1]*v[2] - u[2]*v[1],
		u[2]*v[0] - u[0]*v[2],
		u[0]*v[1] - u[1]*v[0],
	}
	return mgl32.Vec3{}.Sub(n.Normalize())
}

func calcNormalForTriangleInv(p1, p2, p3 mgl32.Vec3) mgl32.Vec3 {
	u := p2.Sub(p1)
	v := p3.Sub(p1)

	/*
		Nx = UyVz - UzVy
		Ny = UzVx - UxVz
		Nz = UxVy - UyVx
	*/
	n := mgl32.Vec3{
		u[1]*v[2] - u[2]*v[1],
		u[2]*v[0] - u[0]*v[2],
		u[0]*v[1] - u[1]*v[0],
	}
	return n.Normalize()
}

func (v *GraphicsLayer) DrawVectorData(data map[types.VectorType][]*types.Vector, isTurtle bool, turtleAttr *types.Vector) {

	lx, ly, lz := v.Spec.GetPos()
	var vx, vy, vz float32
	if isTurtle {
		vx, vy, vz = turtleAttr.X[0], turtleAttr.Y[0], turtleAttr.Z[0]
	}

	m := glumby.NewMesh(gl.LINES)
	m.Vertex3f(0, 0, 0)
	m.Color4f(1, 1, 1, 1)
	m.Vertex3f(0, 0, 0)
	m.Color4f(1, 0, 0, 1)

	tm := GetTriangle(v.BoxWidth, v.BoxHeight)
	qm := GetPlaneAsTriangles(v.BoxWidth, v.BoxHeight)

	if v.PxMesh == nil {
		v.PxMesh = GetCubeAsTriangles(20, 20, 20) //GetCubeAsTriangles(v.BoxWidth, v.BoxHeight, v.BoxDepth)
	}

	gl.Color3f(1, 1, 1)
	memindex := v.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
	gl.LineWidth(settings.GetLineWidth(memindex))

	//hx, hy := (v.UnitWidth * float32(v.Width) / 2), (v.UnitHeight * float32(v.Height) / 2)
	hx, hy := v.BitmapPosX, v.BitmapPosY

	pal := v.Palette // v.Spec.GetPalette()

	keys := make([]int, len(data))
	i := 0
	for t, _ := range data {
		keys[i] = int(t)
		i++
	}
	sort.Ints(keys)

	var vl []*types.Vector
	var t types.VectorType
	for _, tt := range keys {
		t = types.VectorType(tt)
		vl = data[t]
		//		//fmt.Printf("Vector type = %d\n", t)

		gl.BindTexture(gl.TEXTURE_2D, 0)

		//gl.Begin(gl.LINES)
		switch t {
		case types.VT_LINECUBE:
			//log.Printf("Drawing %d cubes\n", len(vl))
			//gl.PushAttrib(gl.ENABLE_BIT)
			//gl.Disable(gl.BLEND)
			//gl.Enable(gl.LINE_STIPPLE)
			//gl.LineStipple(1, 0x0003)
			glumby.MeshBuffer_Begin(gl.LINES)
			var mx *glumby.Mesh
			var v0, v1, v2, v3, v4, v5, v6, v7 mgl32.Vec3
			//var amt float32
			for _, vv := range vl {

				v0 = mgl32.Vec3{vv.X[0]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[0]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[0]*v.UnitDepth + float32(v.MPos.Z)}
				v1 = mgl32.Vec3{vv.X[1]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[1]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[1]*v.UnitDepth + float32(v.MPos.Z)}
				v2 = mgl32.Vec3{vv.X[2]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[2]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[2]*v.UnitDepth + float32(v.MPos.Z)}
				v3 = mgl32.Vec3{vv.X[3]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[3]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[3]*v.UnitDepth + float32(v.MPos.Z)}
				v4 = mgl32.Vec3{vv.X[4]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[4]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[4]*v.UnitDepth + float32(v.MPos.Z)}
				v5 = mgl32.Vec3{vv.X[5]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[5]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[5]*v.UnitDepth + float32(v.MPos.Z)}
				v6 = mgl32.Vec3{vv.X[6]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[6]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[6]*v.UnitDepth + float32(v.MPos.Z)}
				v7 = mgl32.Vec3{vv.X[7]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[7]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[7]*v.UnitDepth + float32(v.MPos.Z)}

				c := pal.Get(int(vv.RGBA))

				r, g, b, a := c.ToFRGBAI(vv.RGBA)

				mx = GetCubeAsLinesV(
					v0, v1, v2, v3, v4, v5, v6, v7,
				)

				mx.SetColor(r, g, b, a)

				mx.DrawWithMeshBuffer(float32(lx), float32(ly), float32(lz))
			}
			glumby.MeshBuffer_End()
			//gl.Disable(gl.LINE_STIPPLE)
			//gl.PopAttrib()
			//gl.Flush()

		case types.VT_PYRAMID:
			//log2.Printf("Drawing %d pyramids\n", len(vl))
			gl.Enable(gl.LIGHTING)
			glumby.MeshBuffer_Begin(gl.TRIANGLES)
			var mx *glumby.Mesh
			var ax, bx, cx, dx, ex mgl32.Vec3
			//var amt float32
			for _, vv := range vl {

				ax = mgl32.Vec3{vv.X[0]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[0]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[0]*v.UnitDepth + float32(v.MPos.Z)}
				bx = mgl32.Vec3{vv.X[1]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[1]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[1]*v.UnitDepth + float32(v.MPos.Z)}
				cx = mgl32.Vec3{vv.X[2]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[2]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[2]*v.UnitDepth + float32(v.MPos.Z)}
				dx = mgl32.Vec3{vv.X[3]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[3]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[3]*v.UnitDepth + float32(v.MPos.Z)}
				ex = mgl32.Vec3{vv.X[4]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[4]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[4]*v.UnitDepth + float32(v.MPos.Z)}

				c := pal.Get(int(vv.RGBA))

				r, g, b, a := c.ToFRGBAI(vv.RGBA)

				mx = GetPyramidAsTrianglesV(
					ax, bx, cx, dx, ex,
				)

				//	log2.Printf("Got mesh with %d vertices", mx.Size())

				mx.SetColor(r, g, b, a)

				mx.DrawWithMeshBuffer(float32(lx), float32(ly), float32(lz))
			}
			glumby.MeshBuffer_End()

		case types.VT_CUBE:
			//log.Printf("Drawing %d cubes\n", len(vl))
			glumby.MeshBuffer_Begin(gl.TRIANGLES)
			var mx *glumby.Mesh
			var v0, v1, v2, v3, v4, v5, v6, v7 mgl32.Vec3
			//var amt float32
			for _, vv := range vl {

				v0 = mgl32.Vec3{vv.X[0]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[0]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[0]*v.UnitDepth + float32(v.MPos.Z)}
				v1 = mgl32.Vec3{vv.X[1]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[1]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[1]*v.UnitDepth + float32(v.MPos.Z)}
				v2 = mgl32.Vec3{vv.X[2]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[2]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[2]*v.UnitDepth + float32(v.MPos.Z)}
				v3 = mgl32.Vec3{vv.X[3]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[3]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[3]*v.UnitDepth + float32(v.MPos.Z)}
				v4 = mgl32.Vec3{vv.X[4]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[4]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[4]*v.UnitDepth + float32(v.MPos.Z)}
				v5 = mgl32.Vec3{vv.X[5]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[5]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[5]*v.UnitDepth + float32(v.MPos.Z)}
				v6 = mgl32.Vec3{vv.X[6]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[6]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[6]*v.UnitDepth + float32(v.MPos.Z)}
				v7 = mgl32.Vec3{vv.X[7]*v.UnitWidth + hx + float32(v.MPos.X), vv.Y[7]*v.UnitHeight + hy + float32(v.MPos.Y), vv.Z[7]*v.UnitDepth + float32(v.MPos.Z)}

				// ox = ox*v.UnitWidth + hx + float32(v.MPos.X)  //- (amt*v.UnitWidth)/2
				// oy = oy*v.UnitHeight + hy + float32(v.MPos.Y) //- (amt*v.UnitHeight)/2
				// oz = oz*v.UnitDepth + float32(v.MPos.Z)       // + (amt*v.UnitDepth)/2

				c := pal.Get(int(vv.RGBA))

				r, g, b, a := c.ToFRGBAI(vv.RGBA)

				mx = GetCubeAsTrianglesV(
					v0, v1, v2, v3, v4, v5, v6, v7,
				)

				mx.SetColor(r, g, b, a)

				mx.DrawWithMeshBuffer(float32(lx), float32(ly), float32(lz))
			}
			glumby.MeshBuffer_End()
		case types.VT_LINE:
			gl.Disable(gl.LIGHTING)
			glumby.MeshBuffer_Begin(gl.LINES)
			for _, vv := range vl {
				ox, oy, oz, dx, dy, dz := vv.X[0], vv.Y[0], vv.Z[0], vv.X[1], vv.Y[1], vv.Z[1]

				if isTurtle {
					ox, oy, oz = ox+vx, oy+vy, oz+vz
					dx, dy, dz = dx+vx, dy+vy, dz+vz
				}

				ox = ox*v.UnitWidth + hx + float32(v.MPos.X)
				dx = dx*v.UnitWidth + hx + float32(v.MPos.X)
				oy = oy*v.UnitHeight + hy + float32(v.MPos.Y)
				dy = dy*v.UnitHeight + hy + float32(v.MPos.Y)
				oz = oz*v.UnitDepth + float32(v.MPos.Z)
				dz = dz*v.UnitDepth + float32(v.MPos.Z)

				//fmt.Println(v.UnitDepth)

				c := pal.Get(int(vv.RGBA))

				r, g, b, a := c.ToFRGBAI(vv.RGBA)

				//fmt.Printf("Vector: (%f, %f, %f) - (%f, %f, %f)\n", ox, oy, oz, dx, dy, dz)

				m.SetColor4f(0, r, g, b, a)
				m.SetColor4f(1, r, g, b, a)
				m.SetVertex3f(0, ox, oy, oz)
				m.SetVertex3f(1, dx, dy, dz)

				m.DrawWithMeshBuffer(float32(lx), float32(ly), float32(lz))
			}
			glumby.MeshBuffer_End()
			gl.Enable(gl.LIGHTING)

		case types.VT_TRIANGLE:

			//log2.Printf("Triangle type count = %d", len(vl))

			gl.Enable(gl.LIGHTING)
			glumby.MeshBuffer_Begin(gl.TRIANGLES)
			for _, vv := range vl {
				ax, ay, az, bx, by, bz, cx, cy, cz := vv.X[0], vv.Y[0], vv.Z[0], vv.X[1], vv.Y[1], vv.Z[1], vv.X[2], vv.Y[2], vv.Z[2]

				if isTurtle {
					ax, ay, az = ax+vx, ay+vy, az+vz
					bx, by, bz = bx+vx, by+vy, bz+vz
					cx, cy, cz = cx+vx, cy+vy, cz+vz
				}

				//fm.Printf("VecTriangle{ %d, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f} },\r\n", vv.RGBA, ax, ay, az, bx, by, bz, cx, cy, cz)

				ax = ax*v.UnitWidth + hx + float32(v.MPos.X)
				bx = bx*v.UnitWidth + hx + float32(v.MPos.X)
				cx = cx*v.UnitWidth + hx + float32(v.MPos.X)
				ay = ay*v.UnitHeight + hy + float32(v.MPos.Y)
				by = by*v.UnitHeight + hy + float32(v.MPos.Y)
				cy = cy*v.UnitHeight + hy + float32(v.MPos.Y)
				az = az*v.UnitDepth + float32(v.MPos.Z)
				bz = bz*v.UnitDepth + float32(v.MPos.Z)
				cz = cz*v.UnitDepth + float32(v.MPos.Z)

				c := pal.Get(int(vv.RGBA))

				r, g, b, a := c.ToFRGBAI(vv.RGBA)

				tm.SetColor4f(0, r, g, b, a)
				tm.SetColor4f(1, r, g, b, a)
				tm.SetColor4f(2, r, g, b, a)
				tm.SetVertex3f(0, ax, ay, az)
				tm.SetVertex3f(1, bx, by, bz)
				tm.SetVertex3f(2, cx, cy, cz)

				n := calcNormalForTriangle(mgl32.Vec3{ax, ay, az}, mgl32.Vec3{bx, by, bz}, mgl32.Vec3{cx, cy, cz})

				tm.SetNormal3f(0, n[0], n[1], n[2])
				tm.SetNormal3f(1, n[0], n[1], n[2])
				tm.SetNormal3f(2, n[0], n[1], n[2])

				tm.DrawWithMeshBuffer(float32(lx), float32(ly), float32(lz))
			}
			glumby.MeshBuffer_End()
			gl.Enable(gl.LIGHTING)

		case types.VT_TRIANGLE_LINE:

			//log2.Printf("Triangle type count = %d", len(vl))

			gl.Disable(gl.LIGHTING)
			glumby.MeshBuffer_Begin(gl.LINES)
			lm := glumby.NewMesh(gl.LINES)
			lm.Vertex3f(0, 0, 0)
			lm.Vertex3f(0, 0, 0)
			lm.Vertex3f(0, 0, 0)
			lm.Vertex3f(0, 0, 0)
			lm.Vertex3f(0, 0, 0)
			lm.Vertex3f(0, 0, 0)
			lm.Color4f(0, 0, 0, 0)
			lm.Color4f(0, 0, 0, 0)
			lm.Color4f(0, 0, 0, 0)
			lm.Color4f(0, 0, 0, 0)
			lm.Color4f(0, 0, 0, 0)
			lm.Color4f(0, 0, 0, 0)
			for _, vv := range vl {
				ax, ay, az, bx, by, bz, cx, cy, cz := vv.X[0], vv.Y[0], vv.Z[0], vv.X[1], vv.Y[1], vv.Z[1], vv.X[2], vv.Y[2], vv.Z[2]

				if isTurtle {
					ax, ay, az = ax+vx, ay+vy, az+vz
					bx, by, bz = bx+vx, by+vy, bz+vz
					cx, cy, cz = cx+vx, cy+vy, cz+vz
				}

				//fm.Printf("VecTriangle{ %d, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f} },\r\n", vv.RGBA, ax, ay, az, bx, by, bz, cx, cy, cz)

				ax = ax*v.UnitWidth + hx + float32(v.MPos.X)
				bx = bx*v.UnitWidth + hx + float32(v.MPos.X)
				cx = cx*v.UnitWidth + hx + float32(v.MPos.X)
				ay = ay*v.UnitHeight + hy + float32(v.MPos.Y)
				by = by*v.UnitHeight + hy + float32(v.MPos.Y)
				cy = cy*v.UnitHeight + hy + float32(v.MPos.Y)
				az = az*v.UnitDepth + float32(v.MPos.Z)
				bz = bz*v.UnitDepth + float32(v.MPos.Z)
				cz = cz*v.UnitDepth + float32(v.MPos.Z)

				c := pal.Get(int(vv.RGBA))

				r, g, b, a := c.ToFRGBAI(vv.RGBA)

				lm.SetColor4f(0, r, g, b, a)
				lm.SetColor4f(1, r, g, b, a)
				lm.SetColor4f(2, r, g, b, a)
				lm.SetColor4f(3, r, g, b, a)
				lm.SetColor4f(4, r, g, b, a)
				lm.SetColor4f(5, r, g, b, a)
				lm.SetVertex3f(0, ax, ay, az)
				lm.SetVertex3f(1, bx, by, bz)
				lm.SetVertex3f(2, bx, by, bz)
				lm.SetVertex3f(3, cx, cy, cz)
				lm.SetVertex3f(4, cx, cy, cz)
				lm.SetVertex3f(5, ax, ay, az)

				lm.DrawWithMeshBuffer(float32(lx), float32(ly), float32(lz))
			}
			glumby.MeshBuffer_End()
			gl.Enable(gl.LIGHTING)

		case types.VT_TURTLE:

			// TODO: graft in turtles here
			// if !isTurtle {
			// 	for _, vv := range vl {
			// 		v.DrawVectorData(turtleData, true, vv) // render of turtle data
			// 	}
			// }

		case types.VT_QUAD:
			gl.Enable(gl.LIGHTING)
			glumby.MeshBuffer_Begin(gl.TRIANGLES)
			for _, vv := range vl {
				ax, ay, az, bx, by, bz, cx, cy, cz, dx, dy, dz := vv.X[0], vv.Y[0], vv.Z[0], vv.X[1], vv.Y[1], vv.Z[1], vv.X[2], vv.Y[2], vv.Z[2], vv.X[3], vv.Y[3], vv.Z[3]

				if isTurtle {
					ax, ay, az = ax+vx, ay+vy, az+vz
					bx, by, bz = bx+vx, by+vy, bz+vz
					cx, cy, cz = cx+vx, cy+vy, cz+vz
					dx, dy, dz = dx+vx, dy+vy, dz+vz
				}

				//fm.Printf("VecTriangle{ %d, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f} },\r\n", vv.RGBA, ax, ay, az, bx, by, bz, cx, cy, cz)

				ax = ax*v.UnitWidth + hx + float32(v.MPos.X)
				bx = bx*v.UnitWidth + hx + float32(v.MPos.X)
				cx = cx*v.UnitWidth + hx + float32(v.MPos.X)
				dx = dx*v.UnitWidth + hx + float32(v.MPos.X)
				ay = ay*v.UnitHeight + hy + float32(v.MPos.Y)
				by = by*v.UnitHeight + hy + float32(v.MPos.Y)
				cy = cy*v.UnitHeight + hy + float32(v.MPos.Y)
				dy = dy*v.UnitHeight + hy + float32(v.MPos.Y)
				az = az*v.UnitDepth + float32(v.MPos.Z)
				bz = bz*v.UnitDepth + float32(v.MPos.Z)
				cz = cz*v.UnitDepth + float32(v.MPos.Z)
				dz = dz*v.UnitDepth + float32(v.MPos.Z)

				c := pal.Get(int(vv.RGBA))

				r, g, b, a := c.ToFRGBAI(vv.RGBA)

				qm.SetColor4f(0, r, g, b, a)
				qm.SetColor4f(1, r, g, b, a)
				qm.SetColor4f(2, r, g, b, a)
				qm.SetColor4f(3, r, g, b, a)
				qm.SetColor4f(4, r, g, b, a)
				qm.SetColor4f(5, r, g, b, a)
				qm.SetVertex3f(0, ax, ay, az)
				qm.SetVertex3f(1, bx, by, bz)
				qm.SetVertex3f(2, cx, cy, cz)
				qm.SetVertex3f(3, dx, dy, dz)
				qm.SetVertex3f(4, bx, by, bz)
				qm.SetVertex3f(5, cx, cy, cz)

				n := calcNormalForTriangle(mgl32.Vec3{ax, ay, az}, mgl32.Vec3{bx, by, bz}, mgl32.Vec3{cx, cy, cz})

				qm.SetNormal3f(0, n[0], n[1], n[2])
				qm.SetNormal3f(1, n[0], n[1], n[2])
				qm.SetNormal3f(2, n[0], n[1], n[2])
				qm.SetNormal3f(3, n[0], n[1], n[2])
				qm.SetNormal3f(4, n[0], n[1], n[2])
				qm.SetNormal3f(5, n[0], n[1], n[2])

				qm.DrawWithMeshBuffer(float32(lx), float32(ly), float32(lz))
			}
			glumby.MeshBuffer_End()
			gl.Enable(gl.LIGHTING)

		case types.VT_LINEQUAD:
			lqm := glumby.NewMesh(gl.LINES)
			for i := 0; i < 8; i++ {
				lqm.Vertex3f(0, 0, 0)
				lqm.Color4f(0, 0, 0, 0)
			}
			gl.Enable(gl.LIGHTING)
			glumby.MeshBuffer_Begin(gl.LINES)
			for _, vv := range vl {
				ax, ay, az, bx, by, bz, cx, cy, cz, dx, dy, dz := vv.X[0], vv.Y[0], vv.Z[0], vv.X[1], vv.Y[1], vv.Z[1], vv.X[2], vv.Y[2], vv.Z[2], vv.X[3], vv.Y[3], vv.Z[3]

				if isTurtle {
					ax, ay, az = ax+vx, ay+vy, az+vz
					bx, by, bz = bx+vx, by+vy, bz+vz
					cx, cy, cz = cx+vx, cy+vy, cz+vz
					dx, dy, dz = dx+vx, dy+vy, dz+vz
				}

				//fm.Printf("VecTriangle{ %d, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f} },\r\n", vv.RGBA, ax, ay, az, bx, by, bz, cx, cy, cz)

				ax = ax*v.UnitWidth + hx + float32(v.MPos.X)
				bx = bx*v.UnitWidth + hx + float32(v.MPos.X)
				cx = cx*v.UnitWidth + hx + float32(v.MPos.X)
				dx = dx*v.UnitWidth + hx + float32(v.MPos.X)
				ay = ay*v.UnitHeight + hy + float32(v.MPos.Y)
				by = by*v.UnitHeight + hy + float32(v.MPos.Y)
				cy = cy*v.UnitHeight + hy + float32(v.MPos.Y)
				dy = dy*v.UnitHeight + hy + float32(v.MPos.Y)
				az = az*v.UnitDepth + float32(v.MPos.Z)
				bz = bz*v.UnitDepth + float32(v.MPos.Z)
				cz = cz*v.UnitDepth + float32(v.MPos.Z)
				dz = dz*v.UnitDepth + float32(v.MPos.Z)

				c := pal.Get(int(vv.RGBA))

				r, g, b, a := c.ToFRGBAI(vv.RGBA)

				lqm.SetColor4f(0, r, g, b, a)
				lqm.SetColor4f(1, r, g, b, a)
				lqm.SetColor4f(2, r, g, b, a)
				lqm.SetColor4f(3, r, g, b, a)
				lqm.SetColor4f(4, r, g, b, a)
				lqm.SetColor4f(5, r, g, b, a)
				lqm.SetColor4f(6, r, g, b, a)
				lqm.SetColor4f(7, r, g, b, a)
				lqm.SetVertex3f(0, ax, ay, az)
				lqm.SetVertex3f(1, bx, by, bz)
				lqm.SetVertex3f(2, bx, by, bz)
				lqm.SetVertex3f(3, dx, dy, dz)
				lqm.SetVertex3f(4, dx, dy, dz)
				lqm.SetVertex3f(5, cx, cy, cz)
				lqm.SetVertex3f(6, cx, cy, cz)
				lqm.SetVertex3f(7, ax, ay, az)

				lqm.DrawWithMeshBuffer(float32(lx), float32(ly), float32(lz))
			}
			glumby.MeshBuffer_End()
			gl.Enable(gl.LIGHTING)

		case types.VT_SPHERE:

			gl.Enable(gl.LIGHTING)
			glumby.MeshBuffer_Begin(gl.TRIANGLES)
			for _, vv := range vl {
				ax, ay, az, rad := vv.X[0], vv.Y[0], vv.Z[0], vv.X[1]

				ax = ax*v.UnitWidth + hx + float32(v.MPos.X)
				ay = ay*v.UnitHeight + hy + float32(v.MPos.Y)
				az = az*v.UnitDepth + float32(v.MPos.Z)

				rad = rad * v.UnitWidth

				c := pal.Get(int(vv.RGBA))

				r, g, b, a := c.ToFRGBAI(vv.RGBA)

				mx := sphereBuilder.MeshTriangles(rad)

				mx.SetColor(r, g, b, a)

				mx.DrawWithMeshBuffer(ax+float32(lx), ay+float32(ly), az+float32(lz))

			}
			glumby.MeshBuffer_End()
			gl.Enable(gl.LIGHTING)

		case types.VT_LINECIRCLE, types.VT_CIRCLE:
			pt := gl.LINE_LOOP
			if t == types.VT_CIRCLE {
				pt = gl.TRIANGLE_FAN
			}
			gl.Enable(gl.LIGHTING)
			for _, vv := range vl {
				glumby.MeshBuffer_Begin(uint32(pt))
				ax, ay, az, vx, vy, vz, ux, uy, uz, rx, ry, rz, rad := vv.X[0], vv.Y[0], vv.Z[0], vv.X[1], vv.Y[1], vv.Z[1], vv.X[2], vv.Y[2], vv.Z[2], vv.X[3], vv.Y[3], vv.Z[3], vv.X[4]

				if isTurtle {
					ax, ay, az = ax+vx, ay+vy, az+vz
				}

				//log2.Printf("circle orientation (frontend): (%v, %v, %v), (%v, %v, %v), (%v, %v, %v)", vx, vy, vz, ux, uy, uz, rx, ry, rz)

				//fm.Printf("VecTriangle{ %d, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f} },\r\n", vv.RGBA, ax, ay, az, bx, by, bz, cx, cy, cz)

				ax = ax*v.UnitWidth + hx + float32(v.MPos.X)
				ay = ay*v.UnitHeight + hy + float32(v.MPos.Y)
				az = az*v.UnitDepth + float32(v.MPos.Z)

				c := pal.Get(int(vv.RGBA))

				r, g, b, a := c.ToFRGBAI(vv.RGBA)

				ccm := GetCircleAsLinesRel(
					uint32(pt),
					rad*v.UnitWidth,
					mgl32.Vec3{vx, vy, vz},
					mgl32.Vec3{ux, uy, uz},
					mgl32.Vec3{rx, ry, rz},
				)
				ccm.SetColor(r, g, b, a)

				ccm.DrawWithMeshBuffer(ax+float32(lx), ay+float32(ly), az+float32(lz))
				glumby.MeshBuffer_End()
			}
			gl.Enable(gl.LIGHTING)

		case types.VT_LINEARC, types.VT_ARC:
			pt := gl.LINE_STRIP
			if t == types.VT_ARC {
				pt = gl.TRIANGLE_FAN
			}
			gl.Enable(gl.LIGHTING)
			for _, vv := range vl {
				var sang float32
				glumby.MeshBuffer_Begin(uint32(pt))
				ax, ay, az, vx, vy, vz, ux, uy, uz, rx, ry, rz, rad, ang := vv.X[0], vv.Y[0], vv.Z[0], vv.X[1], vv.Y[1], vv.Z[1], vv.X[2], vv.Y[2], vv.Z[2], vv.X[3], vv.Y[3], vv.Z[3], vv.X[4], vv.Y[4]

				if isTurtle {
					ax, ay, az = ax+vx, ay+vy, az+vz
				}

				//log2.Printf("circle orientation (frontend): (%v, %v, %v), (%v, %v, %v), (%v, %v, %v)", vx, vy, vz, ux, uy, uz, rx, ry, rz)

				//fm.Printf("VecTriangle{ %d, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f} },\r\n", vv.RGBA, ax, ay, az, bx, by, bz, cx, cy, cz)

				ax = ax*v.UnitWidth + hx + float32(v.MPos.X)
				ay = ay*v.UnitHeight + hy + float32(v.MPos.Y)
				az = az*v.UnitDepth + float32(v.MPos.Z)

				c := pal.Get(int(vv.RGBA))

				r, g, b, a := c.ToFRGBAI(vv.RGBA)

				var ccm *glumby.Mesh

				if ang > 180 {
					ccm = GetArcAsLinesRel(
						uint32(pt),
						rad*v.UnitWidth,
						180,
						0,
						mgl32.Vec3{vx, vy, vz},
						mgl32.Vec3{ux, uy, uz},
						mgl32.Vec3{rx, ry, rz},
					)
					ccm.SetColor(r, g, b, a)

					ccm.DrawWithMeshBuffer(ax+float32(lx), ay+float32(ly), az+float32(lz))
					glumby.MeshBuffer_End()

					ang -= 180
					sang = 180

					glumby.MeshBuffer_Begin(uint32(pt))
				}

				ccm = GetArcAsLinesRel(
					uint32(pt),
					rad*v.UnitWidth,
					ang,
					sang,
					mgl32.Vec3{vx, vy, vz},
					mgl32.Vec3{ux, uy, uz},
					mgl32.Vec3{rx, ry, rz},
				)
				ccm.SetColor(r, g, b, a)

				ccm.DrawWithMeshBuffer(ax+float32(lx), ay+float32(ly), az+float32(lz))
				glumby.MeshBuffer_End()
			}
			gl.Enable(gl.LIGHTING)

		case types.VT_LINEPOLY, types.VT_POLY:
			pt := gl.LINE_LOOP
			if t == types.VT_POLY {
				pt = gl.TRIANGLE_FAN
			}
			gl.Enable(gl.LIGHTING)
			for _, vv := range vl {
				glumby.MeshBuffer_Begin(uint32(pt))
				ax, ay, az, vx, vy, vz, ux, uy, uz, rx, ry, rz, rad, sides := vv.X[0], vv.Y[0], vv.Z[0], vv.X[1], vv.Y[1], vv.Z[1], vv.X[2], vv.Y[2], vv.Z[2], vv.X[3], vv.Y[3], vv.Z[3], vv.X[4], vv.Y[4]

				if isTurtle {
					ax, ay, az = ax+vx, ay+vy, az+vz
				}

				//log2.Printf("circle orientation (frontend): (%v, %v, %v), (%v, %v, %v), (%v, %v, %v)", vx, vy, vz, ux, uy, uz, rx, ry, rz)

				//fm.Printf("VecTriangle{ %d, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f}, VecPoint{%f,%f,%f} },\r\n", vv.RGBA, ax, ay, az, bx, by, bz, cx, cy, cz)

				ax = ax*v.UnitWidth + hx + float32(v.MPos.X)
				ay = ay*v.UnitHeight + hy + float32(v.MPos.Y)
				az = az*v.UnitDepth + float32(v.MPos.Z)

				c := pal.Get(int(vv.RGBA))

				r, g, b, a := c.ToFRGBAI(vv.RGBA)

				ccm := GetPolyAsLinesRel(
					uint32(pt),
					rad*v.UnitWidth,
					sides,
					mgl32.Vec3{vx, vy, vz},
					mgl32.Vec3{ux, uy, uz},
					mgl32.Vec3{rx, ry, rz},
				)
				ccm.SetColor(r, g, b, a)

				ccm.DrawWithMeshBuffer(ax+float32(lx), ay+float32(ly), az+float32(lz))
				glumby.MeshBuffer_End()
			}
			gl.Enable(gl.LIGHTING)
		}

	}

}

func (v *GraphicsLayer) RenderVectors() {

	v.DrawVectorData(v.LastVectorMap, false, nil)
	//v.DrawVectorData(turtleData, true, nil)

}

func (d *GraphicsLayer) getMeshForWidth(c int, w int, z int) *glumby.Mesh {

	if d.Format == types.LF_SUPER_HIRES {
		return d.meshcache[w-1]
	}

	index := z*maxColors*maxStripeLength + c*maxStripeLength + w - 1

	m := d.meshcache[index]
	// if m == nil {
	// 	m = GetCubeAsTriangles(d.BoxWidth*float32(w), d.BoxHeight, d.BoxDepth)
	// 	d.meshcache[w] = m
	// }
	if m == nil {
		fmt.Printf("nil mesh for c=%d, w=%d\n", c, w)
	}
	return m

}

func (this *GraphicsLayer) Free() {
	// cleanup helpers
	if this.NumTextures > 0 {
		// Need to clean these up..
		var textures = make([]uint32, this.NumTextures)
		for i, tex := range this.ScreenTextures {
			textures[i] = tex.Handle()
		}
		gl.DeleteTextures(int32(this.NumTextures), &textures[0])
	}
	//this.MBO.Free()
	//this.MBO = nil
	this.Pixel = nil
}

func (d *GraphicsLayer) CheckVoxelWidths(pal *types.VideoPalette) {

	lsf := d.Spec.GetSubFormat()

	memindex := d.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
	zc := settings.ColorZone[memindex][int(d.Format)]

	var zones = 1
	if zc != nil {
		zones = zones + len(zc.Zones)
	}
	var depth uint8
	var z *settings.VideoPaletteZone
	var p *settings.VideoPalette

	for zone := 0; zone < zones; zone++ {

		for color := 0; color < pal.Size(); color++ {

			if zone == 0 {
				depth = pal.Get(color).Depth
			} else {
				//log2.Printf("zone-1=%d, len(zones)=%d", zone-1, len(zc.Zones))
				z = zc.Zones[zone-1]
				if z != nil {
					p = z.Palette
					depth = p.Get(color).Depth
				} else {
					depth = 20
				}
			}

			if d.lastVoxelDepths[zone*maxColors+color] != depth || d.meshcache[zone*maxColors*maxStripeLength+color*maxStripeLength] == nil {
				// refresh strip

				v := d.stdVoxelDepth
				if depth != 0 {
					v = depth
				}

				mult := float32(v) / 10

				d.lastVoxelDepthf[zone*maxColors+color] = 1 * d.BoxHeight * mult
				if lsf == types.LSF_VOXELS && d.Spec.GetFormat() != types.LF_LOWRES_WOZ {
					d.lastVoxelDepthf[zone*maxColors+color] = 2 * d.BoxHeight * mult
				}

				fmt.Printf("Zone index %d, Color index %d, depth=%f\n", zone, color, d.lastVoxelDepthf[color])

				for w := 1; w <= maxStripeLength; w++ {
					d.meshcache[zone*maxColors*maxStripeLength+color*maxStripeLength+w-1] = GetCubeAsTriangles(d.BoxWidth*float32(w), d.BoxHeight, d.lastVoxelDepthf[zone*maxColors+color])
				}

			}

			d.lastVoxelDepths[zone*maxColors+color] = depth

		}

	}

}

func (d *GraphicsLayer) RenderCubes2() {

	//fmt.Printf("usePaletteOffsets %v\n", d.UsePaletteOffsets)

	//var lc int = -1
	var idx int
	var rgba color.RGBA
	var color int
	var c, oc *types.VideoColor
	var x, y int
	var r, g, b, a float32
	var fx, fy, fz float32

	//fmt.Println("cubes")

	if d.PxMesh == nil {
		d.PxMesh = GetCubeAsTriangles(d.BoxWidth, d.BoxHeight, d.BoxDepth)
		fmt.Println(d.BoxWidth, d.BoxDepth)
	}

	// Start mesh processing
	gl.BindTexture(gl.TEXTURE_2D, 0)

	var oz float32

	pal := d.Spec.GetPalette()
	//bounds := d.Spec.GetBoundsRect()

	d.CheckVoxelWidths(&pal)

	//d.Changed = true

	gl.Enable(gl.CULL_FACE)

	GFXMBO.EnsureCapacity(d.Width, d.Height, 36)

	if GFXMBO.Count == 0 {
		return
	}

	for i := 0; i < GFXMBO.Count; i++ {
		mbo := GFXMBO.MBO(i)
		if mbo.GetFlushCount() > 1 {
			fmt.Printf(".")
			d.Changed = true
		}
		mbo.ResetCount()
	}
	//d.MBO.ResetCount()

	fmt.Println("Frame begin")
	memindex := d.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
	zc := settings.ColorZone[memindex][int(d.Format)]
	if d.DisableZones {
		zc = nil
	}

	var vc *settings.VideoColor

	if d.Changed || d.DepthChanged || d.Force || GFXMBOLastLayerID != d.Spec.GetID() {

		lx, ly, lz := d.Spec.GetPos()

		for i := 0; i < GFXMBO.Count; i++ {
			GFXMBO.MBO(i).Begin(gl.TRIANGLES)
		}

		bx := d.BitmapPosX - d.UnitWidth*float32(d.Width/2)

		var mboIdx int
		var colorwidth int
		var linesPerMBO = d.Height / GFXMBO.Count
		if linesPerMBO == 0 {
			linesPerMBO = d.Height
		}

		var lastDepth, depth uint8
		var dpz float32

		for y = 0; y < d.Height; y++ {

			mboIdx = y / linesPerMBO
			//fmt.Printf("y=%d, mboIdx=%d\n", y, mboIdx)

			colorwidth = 0
			//	lc = -1

			for x = 0; x < d.Width; x++ {
				idx = y*d.Width + x
				color = d.PixelData[idx]

				if d.Format == types.LF_SUPER_HIRES {
					rgba = RGB12ToRGBA(color)
					r = float32(rgba.R) / 255
					g = float32(rgba.G) / 255
					b = float32(rgba.B) / 255
					a = float32(rgba.A) / 255
					oz = 0
					//lc = 0
				} else {

					if zc != nil {
						vc = zc.GetColorAt(x, y, color)
					}

					oc = pal.Get(color % pal.Size())
					if vc == nil {
						//lc = color
						c = d.Palette.Get(color % d.Palette.Size())

						if d.UsePaletteOffsets {
							oz = (float32(oc.Offset) / 10) * d.BoxHeight
						}

						r = float32(c.Red) / 255
						g = float32(c.Green) / 255
						b = float32(c.Blue) / 255
						a = float32(c.Alpha) / 255

						depth = oc.Depth
					} else {
						c = &types.VideoColor{Red: vc.R, Green: vc.G, Blue: vc.B, Alpha: vc.A, Offset: vc.Offset, Depth: vc.Depth}

						//if d.UsePaletteOffsets {
						oz = (float32(vc.Offset) / 10) * d.BoxHeight
						//}

						r = float32(c.Red) / 255
						g = float32(c.Green) / 255
						b = float32(c.Blue) / 255
						a = float32(c.Alpha) / 255

						depth = vc.Depth
					}
				}

				if lastDepth != depth {
					mult := float32(depth) / 10
					dpz = 1 * d.BoxHeight * mult
					d.PxMesh = GetCubeAsTriangles(d.BoxWidth, d.BoxHeight, dpz)
				}

				colorwidth = 1

				fz = float32(d.MPos.Z) + d.Z + oz + 0.5*dpz
				fx = float32(d.MPos.X) + d.X + bx + d.UnitWidth*(float32(x)-0.5*float32(colorwidth)) //+ 0.5*d.UnitWidth
				fy = float32(d.MPos.Y) + d.Y + d.UnitHeight*float32(d.Height-y-1) + 0.5*d.UnitHeight
				d.PxMesh.SetColor(r, g, b, a)
				GFXMBO.MBO(mboIdx).Draw(fx+float32(lx), fy+float32(ly), fz+float32(lz), d.PxMesh)

				//lc = color
				lastDepth = depth

			}

		}

	}

	for i := 0; i < GFXMBO.Count; i++ {
		GFXMBO.MBO(i).Send(true)
	}
	GFXMBOLastLayerID = d.Spec.GetID()

	//fmt.Printf("Cubes done with %d flushes\n", d.MBO.GetFlushCount())

	gl.Disable(gl.CULL_FACE)

	d.DepthChanged = false

}

func (d *GraphicsLayer) RenderCubesNoCache() {

	//fmt.Printf("usePaletteOffsets %v\n", d.UsePaletteOffsets)

	//var lc int = -1
	var idx int
	var rgba color.RGBA
	var color int
	var c, oc *types.VideoColor
	var x, y int
	var r, g, b, a float32
	var fx, fy, fz float32

	//fmt.Println("cubes")

	if d.PxMesh == nil {
		d.PxMesh = GetCubeAsTriangles(d.BoxWidth, d.BoxHeight, d.BoxDepth)
		fmt.Println(d.BoxWidth, d.BoxDepth)
	}

	// Start mesh processing
	gl.BindTexture(gl.TEXTURE_2D, 0)

	var oz float32

	pal := d.Spec.GetPalette()
	//bounds := d.Spec.GetBoundsRect()

	d.CheckVoxelWidths(&pal)

	//d.Changed = true

	gl.Enable(gl.CULL_FACE)

	GFXMBO.EnsureCapacity(d.Width, d.Height, 36)

	if GFXMBO.Count == 0 {
		return
	}

	for i := 0; i < GFXMBO.Count; i++ {
		mbo := GFXMBO.MBO(i)
		if mbo.GetFlushCount() > 1 {
			fmt.Printf(".")
			d.Changed = true
		}
		mbo.ResetCount()
	}
	//d.MBO.ResetCount()

	fmt.Println("Frame begin")
	memindex := d.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
	zc := settings.ColorZone[memindex][int(d.Format)]
	if d.DisableZones {
		zc = nil
	}

	var vc *settings.VideoColor

	if d.Changed || d.DepthChanged || d.Force || GFXMBOLastLayerID != d.Spec.GetID() {

		lx, ly, lz := d.Spec.GetPos()

		for i := 0; i < GFXMBO.Count; i++ {
			GFXMBO.MBO(i).Begin(gl.TRIANGLES)
		}

		bx := d.BitmapPosX - d.UnitWidth*float32(d.Width/2)

		var mboIdx int
		var colorwidth int
		var linesPerMBO = d.Height / GFXMBO.Count
		if linesPerMBO == 0 {
			linesPerMBO = d.Height
		}

		var lastDepth, depth uint8
		var dpz float32

		for y = 0; y < d.Height; y++ {

			mboIdx = y / linesPerMBO
			//fmt.Printf("y=%d, mboIdx=%d\n", y, mboIdx)

			colorwidth = 0
			//	lc = -1

			for x = 0; x < d.Width; x++ {
				idx = y*d.Width + x
				color = d.PixelData[idx]

				if d.Format == types.LF_SUPER_HIRES {
					rgba = RGB12ToRGBA(color)
					r = float32(rgba.R) / 255
					g = float32(rgba.G) / 255
					b = float32(rgba.B) / 255
					a = float32(rgba.A) / 255
					oz = 0
					//lc = 0
				} else {

					if zc != nil {
						vc = zc.GetColorAt(x, y, color)
					}

					oc = pal.Get(color % pal.Size())
					if vc == nil {
						//lc = color
						c = d.Palette.Get(color % d.Palette.Size())

						if d.UsePaletteOffsets {
							oz = (float32(oc.Offset) / 10) * d.BoxHeight
						}

						r = float32(c.Red) / 255
						g = float32(c.Green) / 255
						b = float32(c.Blue) / 255
						a = float32(c.Alpha) / 255

						depth = oc.Depth
					} else {
						c = &types.VideoColor{Red: vc.R, Green: vc.G, Blue: vc.B, Alpha: vc.A, Offset: vc.Offset, Depth: vc.Depth}

						//if d.UsePaletteOffsets {
						oz = (float32(vc.Offset) / 10) * d.BoxHeight
						//}

						r = float32(c.Red) / 255
						g = float32(c.Green) / 255
						b = float32(c.Blue) / 255
						a = float32(c.Alpha) / 255

						depth = vc.Depth
					}
				}

				if lastDepth != depth {
					mult := float32(depth) / 10
					dpz = 1 * d.BoxHeight * mult
					d.PxMesh = GetCubeAsTriangles(d.BoxWidth, d.BoxHeight, dpz)
				}

				colorwidth = 1

				fz = float32(d.MPos.Z) + d.Z + oz + 0.5*dpz
				fx = float32(d.MPos.X) + d.X + bx + d.UnitWidth*(float32(x)-0.5*float32(colorwidth)) //+ 0.5*d.UnitWidth
				fy = float32(d.MPos.Y) + d.Y + d.UnitHeight*float32(d.Height-y-1) + 0.5*d.UnitHeight
				d.PxMesh.SetColor(r, g, b, a)
				GFXMBO.MBO(mboIdx).Draw(fx+float32(lx), fy+float32(ly), fz+float32(lz), d.PxMesh)

				//lc = color
				lastDepth = depth

			}

		}

	}

	for i := 0; i < GFXMBO.Count; i++ {
		GFXMBO.MBO(i).Send(true)
	}
	GFXMBOLastLayerID = d.Spec.GetID()

	//fmt.Printf("Cubes done with %d flushes\n", d.MBO.GetFlushCount())

	gl.Disable(gl.CULL_FACE)

	d.DepthChanged = false

}

func (d *GraphicsLayer) RenderCubes() {

	//fmt.Printf("usePaletteOffsets %v\n", d.UsePaletteOffsets)

	var lc int = -1
	var idx int
	var rgba color.RGBA
	var color int
	var c, oc *types.VideoColor
	var x, y int
	var r, g, b, a float32
	var fx, fy, fz float32

	fmt.Println("Frame begin")
	memindex := d.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
	zc := settings.ColorZone[memindex][int(d.Format)]
	if d.DisableZones {
		zc = nil
	}

	// if zc != nil {
	// 	d.RenderCubesNoCache()
	// 	return
	// }

	//fmt.Println("cubes")

	if d.PxMesh == nil {
		d.PxMesh = GetCubeAsTriangles(d.BoxWidth, d.BoxHeight, d.BoxDepth)
		fmt.Println(d.BoxWidth, d.BoxDepth)
	}

	// Start mesh processing
	gl.BindTexture(gl.TEXTURE_2D, 0)

	var oz float32

	pal := d.Spec.GetPalette()
	bounds := d.Spec.GetBoundsRect()

	bounds.X1 = uint16(d.Width - 1)

	d.CheckVoxelWidths(&pal)

	//d.Changed = true

	gl.Enable(gl.CULL_FACE)

	GFXMBO.EnsureCapacity(d.Width, d.Height, 36)

	if GFXMBO.Count == 0 {
		return
	}

	for i := 0; i < GFXMBO.Count; i++ {
		mbo := GFXMBO.MBO(i)
		if mbo.GetFlushCount() > 1 {
			fmt.Printf(".")
			d.Changed = true
		}
		mbo.ResetCount()
	}
	//d.MBO.ResetCount()

	var vc, lastvc *settings.VideoColor

	depthForColor := func(x, y int, lc int, zone int) float32 {
		if lc < 0 || lc >= maxColors {
			lc = 1
		}
		if zc == nil {
			return d.lastVoxelDepthf[zone*maxColors+lc]
		}
		var vc *settings.VideoColor
		vc = zc.GetColorAt(x, y, lc)
		if vc != nil {
			mult := float32(vc.Depth) / 10
			return 1 * d.BoxHeight * mult
		}
		return d.lastVoxelDepthf[zone*maxColors+lc]
	}

	if d.Changed || d.DepthChanged || d.Force || GFXMBOLastLayerID != d.Spec.GetID() {

		lx, ly, lz := d.Spec.GetPos()

		for i := 0; i < GFXMBO.Count; i++ {
			GFXMBO.MBO(i).Begin(gl.TRIANGLES)
		}

		bx := d.BitmapPosX - d.UnitWidth*float32(d.Width/2)

		var mboIdx int
		var colorwidth int
		var linesPerMBO int = d.Height / GFXMBO.Count
		if linesPerMBO == 0 {
			linesPerMBO = d.Height
		}

		var zoneIndex = 0

		for y = 0; y < d.Height; y++ {

			mboIdx = y / linesPerMBO
			//fmt.Printf("y=%d, mboIdx=%d\n", y, mboIdx)

			colorwidth = 0
			lc = -1

			for x = 0; x < d.Width; x++ {

				if zc == nil {
					zoneIndex = 0
				} else {
					zoneIndex = zc.GetZoneAt(x, y)
				}

				idx = y*d.Width + x
				color = d.PixelData[idx]

				if !bounds.Contains(uint16(x), uint16(y)) || d.IsTransparent(color, &pal) {
					if colorwidth > 0 {
						// Plot it

						d.PxMesh = d.getMeshForWidth(lc%pal.Size(), colorwidth, zoneIndex)
						//fz = float32(d.MPos.Z) + d.Z + oz + 0.5*d.UnitDepth
						fz = float32(d.MPos.Z) + d.Z + oz + 0.5*depthForColor(x, y, lc, zoneIndex)
						fx = float32(d.MPos.X) + d.X + bx + d.UnitWidth*(float32(x)-0.5*float32(colorwidth)) //+ 0.5*d.UnitWidth
						fy = float32(d.MPos.Y) + d.Y + d.UnitHeight*float32(d.Height-y-1) + 0.5*d.UnitHeight
						d.PxMesh.SetColor(r, g, b, a)
						GFXMBO.MBO(mboIdx).Draw(fx+float32(lx), fy+float32(ly), fz+float32(lz), d.PxMesh)
					}
					lc = color
					colorwidth = 0

					continue
				}

				lastvc = vc
				if zc != nil {
					vc = zc.GetColorAt(x, y, color)
				}

				if lc != color || colorwidth == maxStripeLength || lastvc != vc {

					if colorwidth > 0 {
						// Plot it
						d.PxMesh = d.getMeshForWidth(lc%pal.Size(), colorwidth, zoneIndex)
						fz = float32(d.MPos.Z) + d.Z + oz + 0.5*depthForColor(x, y, lc, zoneIndex)
						fx = float32(d.MPos.X) + d.X + bx + d.UnitWidth*(float32(x)-0.5*float32(colorwidth)) //+ 0.5*d.UnitWidth
						fy = float32(d.MPos.Y) + d.Y + d.UnitHeight*float32(d.Height-y-1) + 0.5*d.UnitHeight
						d.PxMesh.SetColor(r, g, b, a)
						GFXMBO.MBO(mboIdx).Draw(fx+float32(lx), fy+float32(ly), fz+float32(lz), d.PxMesh)
					}

					if d.Format == types.LF_SUPER_HIRES {
						rgba = RGB12ToRGBA(color)
						r = float32(rgba.R) / 255
						g = float32(rgba.G) / 255
						b = float32(rgba.B) / 255
						a = float32(rgba.A) / 255
						oz = 0
						lc = 0
					} else {

						oc = pal.Get(color % pal.Size())
						if vc == nil {
							lc = color
							c = d.Palette.Get(color % d.Palette.Size())

							if d.UsePaletteOffsets {
								oz = (float32(oc.Offset) / 10) * d.BoxHeight
							}

							r = float32(c.Red) / 255
							g = float32(c.Green) / 255
							b = float32(c.Blue) / 255
							a = float32(c.Alpha) / 255
						} else {
							c = &types.VideoColor{Red: vc.R, Green: vc.G, Blue: vc.B, Alpha: vc.A, Offset: vc.Offset, Depth: vc.Depth}

							//if d.UsePaletteOffsets {
							oz = (float32(vc.Offset) / 10) * d.BoxHeight
							//}

							r = float32(c.Red) / 255
							g = float32(c.Green) / 255
							b = float32(c.Blue) / 255
							a = float32(c.Alpha) / 255
						}
					}

					colorwidth = 0

				}

				colorwidth++

				lc = color

			}

			// and here
			if colorwidth > 0 {
				// Plot it
				d.PxMesh = d.getMeshForWidth(lc%pal.Size(), colorwidth, zoneIndex)
				fz = float32(d.MPos.Z) + d.Z + oz + 0.5*depthForColor(x, y, lc, zoneIndex)
				fx = float32(d.MPos.X) + d.X + bx + d.UnitWidth*(float32(x)-0.5*float32(colorwidth)) //+ 0.5*d.UnitWidth
				fy = float32(d.MPos.Y) + d.Y + d.UnitHeight*float32(d.Height-y-1) + 0.5*d.UnitHeight
				d.PxMesh.SetColor(r, g, b, a)
				GFXMBO.MBO(mboIdx).Draw(fx+float32(lx), fy+float32(ly), fz+float32(lz), d.PxMesh)
			}
		}

	}

	for i := 0; i < GFXMBO.Count; i++ {
		GFXMBO.MBO(i).Send(true)
	}
	GFXMBOLastLayerID = d.Spec.GetID()

	//fmt.Printf("Cubes done with %d flushes\n", d.MBO.GetFlushCount())

	gl.Disable(gl.CULL_FACE)

	d.DepthChanged = false

}

func (t *GraphicsLayer) RenderTextureLayers() {

	p := t.Spec.GetPalette()

	var fx, fy, fz float32

	lx, ly, lz := t.Spec.GetPos()

	gl.Enable(gl.ALPHA_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	GFXMBO.EnsureCapacity(1, 1, 12)

	for i := 0; i < t.NumTextures; i++ {

		if t.BitmapDirty[i] {
			t.ScreenTextures[i].SetSourceSame(t.BitmapLayers[i])
			t.BitmapDirty[i] = false
		}

		t.ScreenTextures[i].Bind()
		t.d[i].Mesh.SetColor(1, 1, 1, 1)

		oz := (float32(p.Get(i).Offset) / 10) * t.BoxHeight

		fz = float32(t.MPos.Z) + t.Z + oz + float32(lx)
		fx = float32(t.MPos.X) + t.X + t.BitmapPosX + float32(ly)
		fy = float32(t.MPos.Y) + t.Y + t.BitmapPosY + float32(lz)

		GFXMBO.MBO(0).Begin(gl.TRIANGLES)
		GFXMBO.MBO(0).Draw(fx, fy, fz, t.d[i].Mesh)
		GFXMBO.MBO(0).Send(true)
	}

	//fmt.Println("gl:rtl:end")

}

func (d *GraphicsLayer) Capture() {
	for i := 0; i < d.NumTextures; i++ {
		filename := fmt.Sprintf("%s-layer-%d.png", d.Spec.GetID(), i)
		f, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}

		if err := png.Encode(f, d.BitmapLayers[i]); err != nil {
			f.Close()
			log.Fatal(err)
		}

		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}
}

func (d *GraphicsLayer) RenderPoints(usePointSprites bool) {

	//fmt.Println("gl:rp:begin")

	var lc int = -1
	var rgba color.RGBA
	var color int
	var c, oc *types.VideoColor
	var x, y int
	var r, g, b, a float32
	var fx, fy, fz float32
	var idx int

	lx, ly, lz := d.Spec.GetPos()

	if d.PxMesh == nil {
		d.PxMesh = glumby.NewMesh(gl.POINTS)
		d.PxMesh.Vertex3f(0, 0, 0)
		d.PxMesh.Color4f(1, 1, 1, 1)
	}

	//gl.Disable(gl.TEXTURE_2D)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	gl.PointSize(HGRPixelSize)

	gl.Enable(gl.POINT_SMOOTH)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	//	var count int
	pal := d.Spec.GetPalette()
	bounds := d.Spec.GetBoundsRect()
	bounds.X1 = uint16(d.Width - 1)

	var useHalfPixelOffsets = (d.Format == types.LF_HGR_WOZ)
	var hpo float32

	GFXMBO.EnsureCapacity(1, 1, 1)

	memindex := d.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE

	zc := settings.ColorZone[memindex][d.Format]
	if d.DisableZones {
		zc = nil
	}

	if d.Changed || GFXMBOLastLayerID != d.Spec.GetID() {

		GFXMBO.MBO(0).Begin(gl.POINTS)

		var oz float32
		var xskip int = 1

		if !d.Spec.GetMono() && d.Format == types.LF_DHGR_WOZ {
			xskip = 2
		}

		//fmt.Printf("Screen width=%d (skip=%d)\n", d.Width, xskip)
		bx := d.BitmapPosX - d.UnitWidth*float32(d.Width/2)

		for y = d.Height - 1; y >= 0; y-- {
			for x = 0; x < d.Width; x += xskip {

				//fmt.Printf("%d,", x)

				idx = y*d.Width + x
				color = d.PixelData[idx]

				if !bounds.Contains(uint16(x), uint16(y)) || d.IsTransparent(color, &pal) {
					continue
				}

				if color == 0 {
					continue
				}

				if lc != color || zc != nil {
					lc = color

					if d.Format == types.LF_SUPER_HIRES {
						rgba = RGB12ToRGBA(color)
						r = float32(rgba.R) / 255
						g = float32(rgba.G) / 255
						b = float32(rgba.B) / 255
						a = float32(rgba.A) / 255
						oz = 0
					} else {

						var vc *settings.VideoColor
						if zc != nil {
							vc = zc.GetColorAt(x, y, color)
						}

						oc = pal.Get(color % pal.Size())
						if vc == nil {
							c = d.Palette.Get(color % d.Palette.Size())

							//if d.UsePaletteOffsets {
							oz = (float32(oc.Offset) / 10) * d.BoxHeight
							//}

							r = float32(c.Red) / 255
							g = float32(c.Green) / 255
							b = float32(c.Blue) / 255
							a = float32(c.Alpha) / 255
						} else {
							c = &types.VideoColor{Red: vc.R, Green: vc.G, Blue: vc.B, Alpha: vc.A, Offset: vc.Offset, Depth: vc.Depth}

							//if d.UsePaletteOffsets {
							oz = (float32(vc.Offset) / 10) * d.BoxHeight
							//}

							r = float32(c.Red) / 255
							g = float32(c.Green) / 255
							b = float32(c.Blue) / 255
							a = float32(c.Alpha) / 255
						}
					}

					d.PxMesh.SetColor(r, g, b, a)
				}

				// Plot it
				hpo = 0.5 * d.UnitWidth
				if color < 4 || !useHalfPixelOffsets {
					hpo = 0
				}

				fz = float32(d.MPos.Z) + d.Z + oz + float32(lx)
				fx = float32(d.MPos.X) + d.X + bx + d.UnitWidth*float32(x) + 0.5*d.UnitWidth + hpo + float32(ly)
				fy = float32(d.MPos.Y) + d.Y + d.UnitHeight*float32(d.Height-y-1) + 0.5*d.UnitHeight + float32(lz)

				GFXMBO.MBO(0).Draw(fx, fy, fz, d.PxMesh)
				//fmt.Printf("x,y=%d,%d\n", x, y)
			}

		}

	}

	GFXMBO.MBO(0).Send(true)
	GFXMBOLastLayerID = d.Spec.GetID()

	gl.Disable(gl.POINT_SMOOTH)
	gl.Disable(gl.BLEND)

	//	if usePointSprites {
	//		gl.Disable(gl.POINT_SPRITE_ARB)
	//	}

	//fmt.Println("gl:rp:end")

}

func (this *GraphicsLayer) Line(x0, y0, x1, y1, c int) {

	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	var sx, sy int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	for {
		this.Plot(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}

}

func (this *GraphicsLayer) MakeUpdatesLORES() {

	// iterate through data, update pixel mesh based on ram buffer
	var offsetFunc func(x, y int) int

	if this.Format == types.LF_LOWRES_LINEAR {
		offsetFunc = this.baseOffsetLinear
	} else if this.Format == types.LF_LOWRES_WOZ {
		if this.DoubleRes {
			offsetFunc = this.baseOffsetWozAlt
		} else {
			offsetFunc = this.baseOffsetWoz
		}
	} else {
		return
	}

	var forceAllCells bool

	this.Changed = false

	var ptr int
	var c0, c1 int
	for y := 0; y < this.Height; y++ {
		for x := 0; x < this.Width; x++ {
			if this.DoubleRes {
				ptr = offsetFunc(x, (y/2)*2)
			} else {
				ptr = offsetFunc(x*2, (y/2)*2)
			}
			ov := this.PreviousBuffer[ptr]
			v := this.framedata[ptr]
			if v != ov || forceAllCells || this.Force || this.sscanchanged[y] {
				this.Changed = true
				// get c value
				c0 = int(v & 0xf)
				c1 = int((v & 0xf0) >> 4)

				if this.DoubleRes && (x%2) == 0 {
					c0 = rol4bit(c0)
					c1 = rol4bit(c1)
				}

				switch y % 2 {
				case 0:
					this.PixelData[x+y*this.Width] = c0
				case 1:
					this.PixelData[x+y*this.Width] = c1
					// also copy to pb
					this.PreviousBuffer[ptr] = v
				}
			}
		}
		this.sscanchanged[y] = false
	}

	this.UpdateSpriteData()

}

func (this *GraphicsLayer) FetchUpdatesWozHGR() {

	// Snap whole of memory here
	this.framedata = this.Buffer.ReadSlice(0, this.Buffer.Size)
	for i, _ := range this.scanchanged {
		if this.scanchanged[i] {
			this.fscanchanged[i] = true
			this.scanchanged[i] = false
		}
	}

}

func (this *GraphicsLayer) UpdateSpriteData() {

	if !settings.SpritesEnabled[this.Buffer.GStart[0]/memory.OCTALYZER_INTERPRETER_SIZE] {
		return
	}

	var pixelWidth = 1
	var ys, ye = 0, 24
	var xs, xe = 0, 24
	var c, rc int
	var data [24][24]byte
	var yy, xx int
	var px, py, ax, ay int
	var sval int

	mode := this.Spec.GetID()
	var useSHR = (mode == "SHR1")
	var useSHR640 = useSHR && (this.Buffer.Read(0x7d00)&128) != 0
	if strings.HasPrefix(mode, "SHR") || strings.HasPrefix(mode, "DHR") {
		pixelWidth = 2
		if useSHR640 {
			pixelWidth = 1
		}
	}

	//this.sc.TestMove() // TODO: remove me
	activeSprites := this.sc.GetEnabledIndexes()
	//log2.Printf("active sprites = %+v", activeSprites)
	for _, id := range activeSprites {
		//log2.Printf("sprite %d is active", id)
		x, y, _, _, scl, bounds, ovcol := this.sc.GetSpriteAttr(id)
		xs = bounds.X
		ys = bounds.Y
		ye = ys + bounds.Size
		xe = xs + bounds.Size
		if xe > 24 {
			xe = 24
		}
		if ye > 24 {
			ye = 24
		}
		sval = int(scl) + 1
		data = this.sc.GetSpriteData(id)
		for yy = ys; yy < ye; yy++ {
			for xx = xs; xx < xe; xx++ {
				c = int(data[xx][yy])
				rc = c
				if ovcol != 0 && c != 0 {
					c = ovcol
					rc = c
				}
				if useSHR {
					if useSHR640 {
						rc = settings.DefaultSHRDitherPalette[c]
					} else {
						rc = settings.DefaultSHR320Palette[c]
					}
					//c = (c >> 8) | (c & 240) | ((c & 15) << 8)
				}
				if c != 0 && (px >= 0 && px < this.Width && py >= 0 && py < this.Height) {
					for py = 0; py < sval; py++ {
						for px = 0; px < sval; px++ {
							ax = x*pixelWidth + ((xx - xs) * sval * pixelWidth) + px*pixelWidth
							ay = y + ((yy - ys) * sval) + py

							if useSHR640 {
								switch ax % 4 {
								case 0:
									rc = settings.DefaultSHR320Palette[(c&3)+0x8]
								case 1:
									rc = settings.DefaultSHR320Palette[(c&3)+0xc]
								case 2:
									rc = settings.DefaultSHR320Palette[(c&3)+0x0]
								case 3:
									rc = settings.DefaultSHR320Palette[(c&3)+0x4]
								}
							}

							this.PlotPixel(ax, ay, rc)
							this.Plot(ax, ay, rc)
							if pixelWidth == 2 {
								if useSHR640 {
									switch (ax + 1) % 4 {
									case 0:
										rc = settings.DefaultSHR320Palette[(c&3)+0x8]
									case 1:
										rc = settings.DefaultSHR320Palette[(c&3)+0xc]
									case 2:
										rc = settings.DefaultSHR320Palette[(c&3)+0x0]
									case 3:
										rc = settings.DefaultSHR320Palette[(c&3)+0x4]
									}
								}
								this.PlotPixel(ax+1, ay, rc)
								this.Plot(ax+1, ay, rc)
							}
						}
						if ay < 200 {
							this.sscanchanged[ay] = true
						}
					}
				}
			}
		}
	}
}

func (this *GraphicsLayer) convertHGRToDHGRBits(data []uint64) []uint64 {
	var out_main, out_aux = make([]uint64, 40), make([]uint64, 40)
	var halfBitShift, bitset /*, lastbitset, lastHalfBitShift*/ bool
	var outbitptr = 0

	var outbit = func(index int) {
		if index >= 560 {
			return
		}
		bidx := index / 14
		bitidx := uint(6 - (index % 7))
		bitmask := uint64(1 << bitidx)
		if (index/7)%2 == 0 {
			out_main[bidx] |= bitmask
		} else {
			out_aux[bidx] |= bitmask
		}
	}

	// var clrbit = func(index int) {
	// 	if index >= 560 {
	// 		return
	// 	}
	// 	bidx := index / 7
	// 	// if bidx%2 == 1 {
	// 	// 	bidx--
	// 	// } else {
	// 	// 	bidx++
	// 	// }
	// 	bitidx := uint(6 - (index % 7))
	// 	bitmask := 0xff ^ uint64(1<<bitidx)
	// 	out[bidx] &= bitmask
	// }

	for i, v := range data {
		outbitptr = (14 * i) + 1
		halfBitShift = (v&128 != 0)
		if halfBitShift {
			outbitptr--
		}
		for bit := 0; bit < 7; bit++ {
			bitset = (v&64 != 0)
			if bitset {
				outbit(outbitptr)
				outbit(outbitptr + 1)
			}
			v <<= 1
			outbitptr += 2
		}
	}

	b := make([]uint64, 80) // we need to interleave the bytes
	for i, _ := range b {
		switch i % 2 {
		case 0:
			b[i] = out_aux[i/2]
		case 1:
			b[i] = out_main[i/2]
		}
	}

	return b
}

func (this *GraphicsLayer) MakeUpdatesWozHGRViaDHGR() {

	//fmt.Println("gl:muhgr:begin")
	this.Changed = false

	b := this.Spec.GetBoundsRect()
	b.X1 = uint16(this.Width - 1)
	boundsChanged := !b.Equals(this.lastBounds)
	this.lastBounds = b
	subformat := this.Spec.GetSubFormat()
	index := this.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE

	var redolines [192]bool

	var offset int
	for y := 0; y < this.Height; y++ {

		if !this.fscanchanged[y] && !boundsChanged && !this.Force && !this.sscanchanged[y] {
			continue
		}

		redolines[y] = true // mark line for 2nd pass
		if settings.UseVerticalBlend[index] {
			if y > 0 {
				redolines[y-1] = true
			}
			if y < this.Height-1 {
				redolines[y+1] = true
			}
		}

		this.fscanchanged[y] = false
		this.sscanchanged[y] = false

		offset = this.HControl.XYToOffset(7, y)

		rawdata := ConvertHGRToDHGRPattern(this.framedata[offset:offset+40], this.Spec.GetMono())

		inBounds := b.Contains(uint16(1), uint16(y))
		if inBounds {
			this.scandata[y] = ColorFlip(this.HControl.ColorsForScanLine(rawdata, this.Spec.GetMono()))
			for i, v := range this.scandata[y] {
				this.scandata[y][i] = 8 + v // bump up in the palette table
			}
		} else {
			this.scandata[y] = make([]int, this.Width)
		}

		for x := 0; x < this.Width; x++ {
			//this.PlotPixel(x, y, this.scandata[y][x])
			this.Plot(x, y, this.scandata[y][x])
		}

	}

	// 2nd pass
	if subformat == types.LSF_SINGLE_LAYER {
		for y := 0; y < this.Height; y++ {
			if !redolines[y] {
				continue
			}

			//this.PlotScanLine(y, 0) // clear it

			if settings.UseVerticalBlend[index] {
				var cc, ct, cb int
				for x := 0; x < this.Width; x++ {
					ct = 0
					cb = 0
					cc = this.scandata[y][x]
					if y > 0 {
						ct = this.scandata[y-1][x]
					}
					if y < this.Height-1 {
						cb = this.scandata[y+1][x]
					}

					this.PlotPixelBlended(x, y, cc, ct, cb, 8, 8, 23, 23)
				}
			} else {
				for x := 0; x < this.Width; x++ {
					this.PlotPixel(x, y, this.scandata[y][x])
				}
			}
		}
	}

	// update sprite data - we do this before, because we modify the
	this.UpdateSpriteData()

}

func (this *GraphicsLayer) MakeUpdatesWozHGR() {

	//fmt.Println("gl:muhgr:begin")
	this.Changed = false

	b := this.Spec.GetBoundsRect()
	boundsChanged := !b.Equals(this.lastBounds)
	this.lastBounds = b
	subformat := this.Spec.GetSubFormat()
	index := this.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE

	var redolines [192]bool

	for y := 0; y < this.Height; y++ {

		if !this.fscanchanged[y] && !boundsChanged && !this.Force && !this.sscanchanged[y] {
			continue
		}

		redolines[y] = true // mark line for 2nd pass
		if settings.UseVerticalBlend[index] {
			if y > 0 {
				redolines[y-1] = true
			}
			if y < this.Height-1 {
				redolines[y+1] = true
			}
		}

		this.fscanchanged[y] = false
		this.sscanchanged[y] = false

		var offset int

		inBounds := b.Contains(uint16(1), uint16(y))
		if inBounds {
			offset = this.HControl.XYToOffset(0, y)
			if len(this.framedata) > 0 {
				this.scandata[y] = this.HControl.ColorsForScanLine(this.framedata[offset:offset+40], this.Spec.GetMono())
			}
		} else {
			this.scandata[y] = make([]int, this.Width)
		}

		for x := 0; x < this.Width; x++ {
			//this.PlotPixel(x, y, this.scandata[y][x])
			this.Plot(x, y, this.scandata[y][x])
		}

	}

	// 2nd pass
	if subformat == types.LSF_SINGLE_LAYER {
		for y := 0; y < this.Height; y++ {
			if !redolines[y] {
				continue
			}

			//this.PlotScanLine(y, 0) // clear it

			if settings.UseVerticalBlend[index] {
				var cc, ct, cb int
				for x := 0; x < this.Width; x++ {
					ct = 0
					cb = 0
					cc = this.scandata[y][x]
					if y > 0 {
						ct = this.scandata[y-1][x]
					}
					if y < this.Height-1 {
						cb = this.scandata[y+1][x]
					}

					this.PlotPixelBlended(x, y, cc, ct, cb, 0, 4, 3, 7)
				}
			} else {
				for x := 0; x < this.Width; x++ {
					this.PlotPixel(x, y, this.scandata[y][x])
				}
			}
		}
	}

	// update sprite data - we do this before, because we modify the
	this.UpdateSpriteData()

}

func (this *GraphicsLayer) MakeUpdatesWozLORES() {

	// iterate through data, update pixel mesh based on ram buffer
	var offsetFunc func(x, y int) int

	if this.Format == types.LF_LOWRES_LINEAR {
		offsetFunc = this.baseOffsetLinear
	} else if this.Format == types.LF_LOWRES_WOZ {
		if this.DoubleRes {
			offsetFunc = this.baseOffsetWozAlt
		} else {
			offsetFunc = this.baseOffsetWoz
		}
	} else {
		return
	}

	var forceAllCells bool

	this.Changed = false

	b := this.Spec.GetBoundsRect()
	boundsChanged := !b.Equals(this.lastBounds)
	this.lastBounds = b

	var ptr int
	var c0, c1 int
	for y := 0; y < this.Height; y++ {
		for x := 0; x < this.Width; x++ {
			if this.DoubleRes {
				ptr = offsetFunc(x, (y/2)*2)
			} else {
				ptr = offsetFunc(x*2, (y/2)*2)
			}
			ov := this.PreviousBuffer[ptr]
			v := this.framedata[ptr]
			if v != ov || forceAllCells || boundsChanged || this.Force || this.sscanchanged[y] {
				this.Changed = true
				// get c value
				c0 = int(v & 0xf)
				c1 = int((v & 0xf0) >> 4)

				if this.DoubleRes && (x%2) == 0 {
					c0 = rol4bit(c0)
					c1 = rol4bit(c1)
				}

				if b.Contains(uint16(x), uint16(y)) {
					switch y % 2 {
					case 0:
						this.PlotPixel(x, y, c0)
						//this.PixelData[x+y*this.Width] = c0
						this.Plot(x, y, c0)
					case 1:
						this.PlotPixel(x, y, c1)
						//this.PixelData[x+y*this.Width] = c1
						this.Plot(x, y, c1)
						// also copy to pb
						this.PreviousBuffer[ptr] = v
					}
				} else {
					this.PlotPixel(x, y, 0)
					this.Plot(x, y, 0)
					this.PreviousBuffer[ptr] = v
				}
			}
		}
		this.sscanchanged[y] = false
	}

	this.UpdateSpriteData()

	this.Force = false

}

func (this *GraphicsLayer) FetchUpdatesWozDHGR() {

	// Snap whole of memory here
	this.framedata = this.Buffer.ReadSlice(0, this.Buffer.Size)
	for i, _ := range this.scanchanged {
		if this.scanchanged[i] {
			this.fscanchanged[i] = true
			this.scanchanged[i] = false
		}
	}

}

func (this *GraphicsLayer) MakeUpdatesWozDHGR() {

	this.Changed = false

	bounds := this.Spec.GetBoundsRect()
	boundsChanged := !bounds.Equals(this.lastBounds)
	this.lastBounds = bounds

	var offs_aux, offs_main int
	var aux_data, main_data []uint64

	dhr := this.HControl.(*hires.DHGRScreen)

	// DHGR High bit handling...
	if settings.DHGRHighBit[this.Spec.Index] != settings.LastDHGRHighBit[this.Spec.Index] {
		boundsChanged = true
		settings.LastDHGRHighBit[this.Spec.Index] = settings.DHGRHighBit[this.Spec.Index]
	}
	if settings.DHGRHighBit[this.Spec.Index] == settings.DHB_MIXED_AUTO {
		if settings.LastDHGRMode3Detected[this.Spec.Index] != settings.DHGRMode3Detected[this.Spec.Index] {
			settings.LastDHGRMode3Detected[this.Spec.Index] = settings.DHGRMode3Detected[this.Spec.Index]
			boundsChanged = true
		}
	}

	var isVoxels = this.Spec.GetSubFormat() == types.LSF_VOXELS

	for y := 0; y < this.Height; y++ {

		if !this.fscanchanged[y] && !boundsChanged && !this.sscanchanged[y] {
			continue
		}
		this.fscanchanged[y] = false
		this.sscanchanged[y] = false

		// check each scanline for updates
		offs_aux = this.HControl.XYToOffset(0, y)
		offs_main = this.HControl.XYToOffset(7, y)

		aux_data = this.framedata[offs_aux : offs_aux+40] // yay we have the scanline bytes
		main_data = this.framedata[offs_main : offs_main+40]

		b := make([]uint64, 80) // we need to interleave the bytes
		for i, _ := range b {
			switch i % 2 {
			case 0:
				b[i] = aux_data[i/2]
			case 1:
				b[i] = main_data[i/2]
			}
		}

		if isVoxels {
			this.scandata[y] = dhr.ColorsForScanLineOld(b, this.Spec.GetMono())
		} else {
			this.scandata[y] = dhr.ColorsForScanLine(b, this.Spec.GetMono())
		}

		for x := 0; x < this.Width; x += 1 {
			if !bounds.Contains(uint16(x), uint16(y)) {
				this.PlotPixel(x, y, 0)
				this.Plot(x, y, 0)
			} else {
				this.PlotPixel(x, y, this.scandata[y][x])
				this.Plot(x, y, this.scandata[y][x])
			}
		}
	}

	// update sprite data - we do this before, because we modify the
	this.UpdateSpriteData()

}

func (this *GraphicsLayer) MakeUpdatesXHGR() {

	this.Changed = false

	for y := 0; y < this.Height; y++ {

		for x := 0; x < this.Width; x++ {
			this.Plot(x, y, this.scandata[y][x])
			this.PlotPixel(x, y, this.scandata[y][x])
		}
	}

}

func (this *GraphicsLayer) MakeUpdatesSHR() {

	// var changeCount int

	this.Lock()
	defer this.Unlock()

	shr, ok := this.HControl.(*hires.SuperHiResBuffer)
	if ok {
		shr.LoadPaletteCache()
	}
	this.framedata = this.Buffer.ReadSliceCopy(0, this.Buffer.Size)

	//log2.Printf("Refresh bounds: %d, %d", this.MinRefresh, this.MaxRefresh)

	// defer func() {
	// 	this.MinRefresh = this.Height + 1
	// 	this.MaxRefresh = -1
	// }()

	memindex := this.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
	if settings.SHRFrameForce[memindex] {
		this.Force = true
		defer func() {
			settings.SHRFrameForce[memindex] = false
		}()
	}

	this.Changed = false

	//var singleLayer = (this.Spec.GetSubFormat() == types.LSF_SINGLE_LAYER)
	var ctrl []uint64

	for y := 0; y < this.Height; y++ {

		// scan not changed, and not forcing
		if !this.scanchanged[y] && !this.Force {
			continue
		}

		// changeCount++

		ctrl = []uint64{this.framedata[0x7d00+y]}
		ctrl = append(ctrl, this.framedata[y*160:y*160+160]...)
		this.scandata[y] = this.HControl.ColorsForScanLine(ctrl, false)

		for x, c := range this.scandata[y] {
			this.PlotPixel(x, y, c)
			this.Plot(x, y, c)
		}

		// mark scanline processed
		this.scanchanged[y] = false
	}

	// if changeCount > 0 {
	// 	log2.Printf("shr: %d changed/refreshed scanlines", changeCount)
	// }

	// update sprite data - we do this before, because we modify the
	this.UpdateSpriteData()

}

func (v *GraphicsLayer) MakeUpdatesVECTOR() {

	index := v.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
	settings.VBLock[index].Lock()
	defer settings.VBLock[index].Unlock()

	//	 this.VControl.Turtle
	vecmap := make(map[types.VectorType][]*types.Vector)

	vl := make(types.VectorList, 0)

	size := 14*int(v.Buffer.Read(0)) + 2
	fresh := (v.Buffer.Read(1) != 0)
	//state := v.Buffer.Read(0) & 0xff000000

	//log.Printf("Seeing vector buffer of size %d, fresh = %b\n", size, fresh)

	var nsize = size
	if size != v.LastVectorSize || fresh {

		if nsize != v.Buffer.Size {
			nsize = v.Buffer.Size
		}

		_ = vl.UnmarshalBinary(v.Buffer.ReadSlice(0, nsize))

		//	 //fmt.Printf("RenderVectors() called for %d vectors...\n", len(vl))

		for _, vv := range vl {
			list, ok := vecmap[vv.Type]
			if !ok {
				list = make([]*types.Vector, 0)
			}
			list = append(list, vv)
			vecmap[vv.Type] = list
		}

		v.Buffer.Write(1, 0)

	} else {
		// same vectors
		vecmap = v.LastVectorMap
	}

	v.LastVectorMap = vecmap
	v.LastVectorSize = size

}
