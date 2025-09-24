package video

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/nfnt/resize"
	"paleotronic.com/accelimage"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/gl"
	"paleotronic.com/glumby"
	"paleotronic.com/log"
	"paleotronic.com/octalyzer/video/font"
)

const defaultFont = "fonts/osdfont.yaml"

var (
	H_HALF   float32 = 1
	H_NORMAL float32 = 2
	H_DOUBLE float32 = 4
	H_QUAD   float32 = 8
	W_HALF   float32 = 1
	W_NORMAL float32 = 2
	W_DOUBLE float32 = 4
	W_QUAD   float32 = 8

	SW = 560 * 2
	SH = 384

	SelSX = 3068
	SelSY = 3069
	SelEX = 3070
	SelEY = 3071
)

var shadevals = [8]float32{1.000, 0.915, 0.825, 0.735, 0.645, 0.555, 0.465, 0.375}

var (
	fontOSD     *font.DecalFont
	fontsLoaded bool
	SnapShot    bool
	SnapCount   int
)

func InitFonts() {
	if fontsLoaded {
		return
	}
	font, err := font.LoadFromFile(defaultFont)
	if err != nil {
		panic(err)
	}
	fontOSD = font
	// for i, _ := range settings.DefaultFont {
	// 	//settings.DefaultFont[i] = font
	// }
	fontsLoaded = true
}

func ReloadFonts(name string) {
	fontpixel = name
	psTexture = nil
	fontsLoaded = true
}

type TextLayer struct {
	BaseLayer
	// ---
	Cache map[float32]map[float32]map[int]map[int]*glumby.Mesh
	//~ DataFG                 [][]*Decal
	//~ DataBG                 [][]*Decal
	PreviousBuffer         []uint64
	CX, LastCX             int
	CY, LastCY             int
	FlashOn                bool
	FlashInterval          int64
	CursorOn               bool
	CursorInterval         int64
	Changed, NeedRefresh   bool
	CursorVisible          bool
	CanCopy                bool
	SelActive              bool
	meshesBuilt            bool
	lastFullFrame          time.Time
	CursorDec, CursorDecBG *Decal
	UseAltSet              bool
	SwitchInterleave       bool
	Palette                types.VideoPalette
	lastChanginess         float32
	//MBO                                *glumby.MeshBufferObject
	ScreenTex                          *glumby.Texture
	ScreenTexW, ScreenTexH             int
	d                                  *Decal
	Bitmap                             *image.RGBA
	BitmapUpdate                       bool
	BitmapPosX, BitmapPosY, BitmapPosZ float32
	RedoArea                           image.Rectangle
	LastBounds                         types.LayerRect
	LastSelBounds                      types.LayerRect
	CellLastRefreshed                  [80 * 50]int64
	IM                                 *types.InlineImageManager
	CountInline                        int
	InlineMap                          map[rune]string
	InlineCache                        map[string]image.Image
	InlineBounds                       map[rune][]InlinePos
	InlinePalettize                    bool
	framedata                          []uint64
	scanchanged, fscanchanged          []bool
	lastSubFormat                      types.LayerSubFormat
	lineHeights                        [50]int
	ttp                                *TeleTextProcessor
	lastTTBuffer                       [ttxtRows * ttxtCols]TeleTextCell
}

func (t *TextLayer) Update() {
	if t.NeedRefresh || t.FlashChanged() || t.CursorChanged() || t.Changed || t.PosChanged {
		t.MakeUpdates()
	}
}

func (t *TextLayer) Render() {

	if sx, sy, ex, ey := t.Spec.GetBounds(); sx >= ex && sy >= ey {
		return
	}

	if t.lastAspect != t.Controller.GetAspect() {
		t.lastAspect = t.Controller.GetAspect()
		t.glWidth = t.glHeight * float32(t.lastAspect)
		t.CalcDims()
	}

	sx, sy, sz := t.Spec.GetPos()

	gl.Enable(gl.TEXTURE_2D)
	t.ScreenTex.Bind()
	// t.MBO.Begin(gl.TRIANGLES)
	glumby.MeshBuffer_Begin(gl.TRIANGLES)
	t.d.Mesh.DrawWithMeshBuffer(t.BitmapPosX+float32(sx), t.BitmapPosY+float32(sy), t.BitmapPosZ+float32(sz))
	glumby.MeshBuffer_End()
	// t.MBO.Draw(t.BitmapPosX, t.BitmapPosY, t.BitmapPosZ, t.d.Mesh)
	// t.MBO.Send(true)
	gl.Disable(gl.TEXTURE_2D)

	//log.Println("RENDERING #############################################################################################")

}

func (d *TextLayer) Fetch() {
	tmp := d.Buffer.ReadSlice(0, 4096)
	d.framedata = make([]uint64, len(tmp))
	for i, v := range tmp {
		d.framedata[i] = v
	}
	for i, _ := range d.scanchanged {
		if d.scanchanged[i] {
			d.fscanchanged[i] = true
			d.scanchanged[i] = false
		}
	}
}

func (t *TextLayer) Dimension(data *memory.MemoryControlBlock) {

	t.Buffer = data
	t.PreviousBuffer = make([]uint64, data.Size)

}

func (t *TextLayer) DecodePackedCharLegacy(memval uint64) (rune, types.VideoAttribute, uint64, uint64, uint64, float32, float32, bool) {

	f := t.PokeToAsciiApple
	if t.Format == types.LF_BBC_TELE_TEXT {
		f = t.PokeToAsciiBBC
	}

	ch := rune(f(memval, t.UseAltSet))
	va := t.PokeToAttribute(memval, t.UseAltSet)
	shade := (memval >> 28) & 0x07
	colidx := (memval >> 16) & 0x0f
	bcolidx := (memval >> 20) & 0x0f
	tm := types.TextSize((memval >> 24) & 15)
	solid := false

	if settings.HighContrastUI {
		if colidx > 0 {
			colidx = 15
		}
		if bcolidx > 0 {
			bcolidx = 15
		}
	}

	var w float32 = W_NORMAL
	var h float32 = H_NORMAL

	ww := int(tm) / 4
	hh := int(tm) % 4

	switch ww {
	case 0:
		w = W_HALF
	case 1:
		w = W_NORMAL
	case 2:
		w = W_DOUBLE
	case 3:
		w = W_QUAD
	}

	switch hh {
	case 0:
		h = H_HALF
	case 1:
		h = H_NORMAL
	case 2:
		h = H_DOUBLE
	case 3:
		h = H_QUAD
	}

	return ch, va, colidx, bcolidx, shade, w, h, solid
}

// PlotBM updates the bitmap version of the screen
func (this *TextLayer) PlotBlank(wv, hv float32, solid bool, x, y int) {

	xp := (this.ScreenTexW / this.Width) * x
	yp := (this.ScreenTexH / this.Height) * y

	w := (this.ScreenTexW / this.Width) * int(wv)
	h := (this.ScreenTexH / this.Height) * int(hv)

	target := image.Rect(xp, yp, xp+w, yp+h) // Char pos

	//draw.Draw(this.Bitmap, target, image.Transparent, image.ZP, draw.Src)
	accelimage.FillRGBA(this.Bitmap, target, color.RGBA{0, 0, 0, 0})

}

func (this *TextLayer) PlotPixelFont(ch rune, colidx, bcolidx, shade int, va types.VideoAttribute, wv, hv float32, solid bool, x, y int) {

	if ch > 255 && ch < 512 {
		ch -= 96
	}

	pw := this.Bitmap.Bounds().Dx() / this.Width
	ph := this.Bitmap.Bounds().Dy() / this.Height

	//log.Printf("pw, ph, width, height = %d, %d", pw, ph, this.Width, this.Height)

	xp := pw * x
	yp := ph * y

	w := pw * int(wv)
	h := ph * int(hv)

	target := image.Rect(xp, yp, xp+w, yp+h) // Char pos

	if (va == types.VA_BLINK && this.FlashOn) || va == types.VA_INVERSE {
		c := colidx
		colidx = bcolidx
		bcolidx = c
	}

	fontNormal := settings.DefaultFont[this.Spec.Index]
	if this.Spec.GetID() == "OOSD" || this.Spec.GetID() == "MONI" {
		fontNormal = fontOSD
	}

	if this.Spec.GetFormat() == types.LF_BBC_TELE_TEXT {
		fontNormal = teleTextFont
	}

	glyph_n, glyphExists := fontNormal.GlyphsN[ch]
	glyph_i, _ := fontNormal.GlyphsI[ch]

	// do BG
	av := color.RGBA{0, 255, 0, 255}

	accelimage.FillRGBA(this.Bitmap, target, color.RGBA{0, 0, 0, 0})

	list, ok := this.InlineBounds[ch]

	if ok {

		p := InlinePos{-1, -1, -1, -1}
		for _, pp := range list {
			if x >= pp.SX && x <= pp.EX && y >= pp.SY && y < pp.EY {
				p = pp
				break
			}
		}

		if p.SX == -1 {
			return
		}

		imageW, imageH := p.Dimensions((this.ScreenTexW / this.Width), (this.ScreenTexH / this.Height))

		mappedimage := this.GetInlineImageByRune(ch, imageW, imageH)

		if mappedimage == nil {
			draw.Draw(this.Bitmap, target, &image.Uniform{av}, image.ZP, draw.Src)
			return
		}

		// mappedimage is defined
		op := image.Point{(this.ScreenTexW / this.Width) * (x - p.SX), (this.ScreenTexH / this.Height) * (y - p.SY)}
		//draw.Draw(this.Bitmap, target, mappedimage, op, draw.Src)
		accelimage.DrawImageRGBAOffset(
			this.Bitmap,
			target,
			mappedimage.(*image.RGBA),
			op,
		)
		return

	}

	if !glyphExists {
		if ch != 0 && ch != 32 {
			log.Printf("Glyph for %s does not exist\n", string(ch))
		}
		return
	}
	//fmt.RPrintf("Glyph for code %d exists!\n", ch)

	if ch == 0 {
		return
	}

	fg := this.Palette.Get(colidx).ToColorNRGBA(uint8(255 - (32 * shade)))
	bg := this.Palette.Get(bcolidx).ToColorNRGBA(uint8(255 - (32 * shade)))

	solid = settings.HighContrastUI && (this.Spec.GetID() == "OOSD" || this.Spec.GetID() == "MONI") && ch != 32

	if solid {
		fg.A = 255
		bg.A = 255
	}

	if !this.IsTransparent(colidx, &this.Palette) || solid {
		this.DrawGlyph(this.Bitmap, glyph_n, fontNormal.TextWidth, fontNormal.TextHeight, target, fg)
	}
	if !this.IsTransparent(bcolidx, &this.Palette) || solid {
		this.DrawGlyph(this.Bitmap, glyph_i, fontNormal.TextWidth, fontNormal.TextHeight, target, bg)
	}
	this.BitmapUpdate = true
}

func scanFuncNIL(ax, ay int, c color.RGBA) color.RGBA {
	return c
}

func scanFunc(ax, ay int, c color.RGBA) color.RGBA {
	d := c
	if ay%2 == 1 && !settings.DisableScanlines {
		d.A = uint8(settings.ScanLineIntensity * float32(d.A))
	}
	return d
}

func (this *TextLayer) DrawGlyph(bitmap *image.RGBA, gdata []font.DecalPoint, fw, fh int, target image.Rectangle, c color.RGBA) {

	pxw, pxh := target.Dx()/fw, target.Dy()/fh // size of a "pixel"

	f := scanFunc
	id := this.Spec.GetID()
	if id == "OOSD" || id == "MONI" || (id != "TEXT" && id != "TXT2" && settings.HighContrastUI) {
		f = scanFuncNIL
	}

	for _, p := range gdata {
		xps, yps := p.X*pxw, p.Y*pxh
		xpe, ype := xps+pxw, yps+pxh

		point := image.Rect(target.Min.X+xps, target.Min.Y+yps, target.Min.X+xpe, target.Min.Y+ype)

		accelimage.FillRGBAWithFilter(bitmap, point, c, f)
	}

}

func NewTextLayer(width, height int, glWidth float32, glHeight float32, data *memory.MemoryControlBlock, spec *types.LayerSpecMapped, useTex *glumby.Texture, useBitmap *image.RGBA) *TextLayer {

	//log2.Printf("NewTextLayer(%s)", spec.GetID())

	InitFonts()

	this := &TextLayer{}

	this.BaseLayer.Spec = spec

	this.InlineBounds = make(map[rune][]InlinePos)

	this.UseAltSet = false
	this.Width = width
	this.Height = height
	this.glWidth = glWidth
	this.glHeight = glHeight
	this.BitmapPosX = this.glWidth / 2
	this.BitmapPosY = this.glHeight / 2
	this.d = NewDecal(float32(this.ScreenTexW), float32(this.ScreenTexH))
	index := data.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
	if spec.GetID() == "OOSD" {
		this.Controller = types.NewOrbitControllerOSD(data.GetMM(), index)
	} else {
		this.Controller = types.NewOrbitController(data.GetMM(), index, -1)
	}
	this.CalcDims()
	this.InlineMap = make(map[rune]string)
	this.InlineCache = make(map[string]image.Image)
	this.FlashInterval = 1000
	this.CursorInterval = 500
	this.InlinePalettize = true
	this.CursorVisible = true
	//this.MBO = &glumby.MeshBufferObject{}
	//~ this.MBOC = &glumby.MeshBufferObject{}
	this.ScreenTexH = SH
	this.ScreenTexW = SW
	this.Format = this.Spec.GetFormat()
	if this.Format == types.LF_BBC_TELE_TEXT {
		this.ScreenTexH = 500
		this.ScreenTexW = 800
	}
	if useTex == nil {
		useTex = glumby.NewTextureBlank(this.ScreenTexW, this.ScreenTexH, color.RGBA{0, 0, 0, 0})
	}
	this.ScreenTex = useTex //
	this.IM = types.NewInlineImageManager(spec.Index, spec.Mm)

	this.d.Texture = this.ScreenTex
	this.d.Mesh = GetPlaneAsTrianglesInv(glWidth, glHeight)

	this.Bitmap = useBitmap
	if this.Bitmap == nil {
		this.Bitmap = image.NewRGBA(image.Rect(0, 0, this.ScreenTexW, this.ScreenTexH))
	}
	draw.Draw(this.Bitmap, this.Bitmap.Bounds(), image.Transparent, image.ZP, draw.Src)

	this.d.Texture.SetSourceSame(this.Bitmap) // force update

	this.Palette = spec.GetPalette()

	this.scanchanged = make([]bool, height)
	this.fscanchanged = make([]bool, height)

	this.Dimension(data)

	if this.Format == types.LF_BBC_TELE_TEXT {
		this.ttp = &TeleTextProcessor{}
	}

	return this
}

func (this *TextLayer) CalcDims() {
	this.UnitWidth = this.glWidth / float32(this.Width)
	this.UnitHeight = this.glHeight / float32(this.Height)
	this.UnitWidth = this.glWidth / float32(this.Width)
	this.UnitHeight = this.glHeight / float32(this.Height)
	this.BoxWidth = this.UnitWidth
	this.BoxHeight = this.UnitHeight
	this.d.Mesh = GetPlaneAsTrianglesInv(this.glWidth, this.glHeight)
	// this.Controller.SetPos(float64(this.glWidth/2), float64(this.glHeight/2), types.CDIST)
	// this.Controller.SetLookAt(mgl64.Vec3{float64(this.glWidth / 2), float64(this.glHeight / 2), 0})
}

const cmFilled = 0xffffffff
const cmSpace = 0x00000000

func (this *TextLayer) ProcessEvent(label string, addr int, value *uint64, action memory.MemoryAction) (bool, bool) {

	if this.Format == types.LF_TEXT_WOZ {
		var offset int

		switch string(label[len(label)-1]) {
		case "0":
			offset = addr - (this.Buffer.GStart[0] % memory.OCTALYZER_INTERPRETER_SIZE)
		case "1":
			offset = 1024 + addr - (this.Buffer.GStart[1] % memory.OCTALYZER_INTERPRETER_SIZE)
		case "2":
			offset = 2048 + addr - (this.Buffer.GStart[2] % memory.OCTALYZER_INTERPRETER_SIZE)
		case "3":
			offset = 3072 + addr - (this.Buffer.GStart[3] % memory.OCTALYZER_INTERPRETER_SIZE)
		}

		if offset%128 >= 120 {
			return true, true
		}
		y := this.OffsetToY(offset)
		this.scanchanged[y] = true
	}

	return true, true
}

func (this *TextLayer) GetText(r types.LayerRect) string {
	var idx int
	var ch rune
	var nv uint64

	data := this.Buffer.ReadSlice(0, 4096)

	lines := []string(nil)

	for y := 0; y < this.Height; y++ {
		s := ""
		for x := 0; x < this.Width; x++ {
			if this.InSelection(r, uint16(y), uint16(x)) {
				if this.Format == 0 {
					// linear
					idx = this.baseOffsetLinear(x, y)
				} else {
					if this.SwitchInterleave || this.Spec.GetSubFormat() == types.LSF_FIXED_80_24 {
						idx = this.baseOffsetWozAlt(x, y)
					} else {
						idx = this.baseOffsetWoz(x, y)
					}
				}

				nv = data[idx]
				isFill := (nv & (1 << 32)) != 0

				if nv != cmFilled && nv != cmSpace && !isFill {

					ch, _, _, _, _, _, _, _ = this.DecodePackedCharLegacy(nv)
					if ch >= 32 && ch <= 255 {
						s += string(ch)
					} else if ch > 255 && ch < 2048 {
						nch, ok := font.HighChar2Unicode[ch]
						if ok {
							ch = nch
						} else {
							ch = font.HighChar2Unicode[0]
						}
						s += string(ch)
					}

				}

			}
		}
		if s != "" {
			lines = append(lines, strings.TrimRight(s, " "))
		}
	}

	s := strings.Join(lines, "\r\n")

	return s
}

func (this *TextLayer) CopyText(r types.LayerRect) {

	// this.CanCopy = true

	// var idx int
	// var ch rune
	// var nv uint64

	// data := this.Buffer.ReadSlice(0, 4096)

	// lines := []string(nil)

	// for y := 0; y < this.Height; y++ {
	// 	s := ""
	// 	for x := 0; x < this.Width; x++ {
	// 		if this.InSelection(r, uint16(y), uint16(x)) {
	// 			if this.Format == 0 {
	// 				// linear
	// 				idx = this.baseOffsetLinear(x, y)
	// 			} else {
	// 				if this.SwitchInterleave || this.Spec.GetSubFormat() == types.LSF_FIXED_80_24 {
	// 					idx = this.baseOffsetWozAlt(x, y)
	// 				} else {
	// 					idx = this.baseOffsetWoz(x, y)
	// 				}
	// 			}

	// 			nv = data[idx]
	// 			isFill := (nv & (1 << 32)) != 0

	// 			if nv != cmFilled && nv != cmSpace && !isFill {

	// 				ch, _, _, _, _, _, _, _ = this.DecodePackedCharLegacy(nv)
	// 				if ch >= 32 && ch <= 255 {
	// 					s += string(ch)
	// 				} else if ch > 255 && ch < 2048 {
	// 					nch, ok := font.HighChar2Unicode[ch]
	// 					if ok {
	// 						ch = nch
	// 					} else {
	// 						ch = font.HighChar2Unicode[0]
	// 					}
	// 					s += string(ch)
	// 				}

	// 			}

	// 		}
	// 	}
	// 	if s != "" {
	// 		lines = append(lines, strings.TrimRight(s, " "))
	// 	}
	// }

	// s := strings.Join(lines, "\r\n")
	s := this.GetText(r)

	fmt.RPrintf("Clipboard: [%s]\n", s)

	if s != "" {
		clipboard.WriteAll(s)
	}
}

func (this *TextLayer) CheckInlines() {
	c := this.IM.Count()
	if c != this.CountInline {
		// we need to process updates
		this.InlineMap = this.IM.GetMap()
		fmt.Printf("INLINEMAP: {%v}\n", this.InlineMap)
		this.InlineBounds = make(map[rune][]InlinePos)
	}
	this.CountInline = c
}

func bytesToPng(d []byte) (image.Image, error) {

	b := bytes.NewBuffer(d)

	i, e := png.Decode(b)

	return i, e
}

func (this *TextLayer) GetInlineImageByRune(ch rune, w, h int) image.Image {

	filename, ok := this.InlineMap[ch]
	if !ok {
		return nil
	}

	key := fmt.Sprintf("%s:%d:%d", filename, w, h)

	// now check cache
	i, iok := this.InlineCache[key]
	if !iok {
		fp, e := files.ReadBytesViaProvider(files.GetPath(filename), files.GetFilename(filename))
		if e != nil {
			this.InlineCache[key] = &image.Uniform{color.RGBA{255, 255, 0, 255}}
			fmt.Println(e)
			return this.InlineCache[key]
		}
		i, e = bytesToPng(fp.Content)
		if e != nil {
			this.InlineCache[key] = &image.Uniform{color.RGBA{255, 255, 0, 255}}
			fmt.Println(e)
			return this.InlineCache[key]
		}

		// now create a resized version
		ri := resize.Resize(uint(w), uint(h), i, resize.NearestNeighbor)

		if this.InlinePalettize {

			// we need to limit the image colors
			//DitherImage(palette *types.VideoPalette, pimg image.Image, gamma float32, matrix *DiffusionMatrix, perceptual bool) image.Image
			ri = apple2helpers.DitherImage(&this.Palette, ri, 1.00, apple2helpers.FloydSteinberg, false)
		}

		this.InlineCache[key] = ri

		return ri
	}

	return i
}

func (this *TextLayer) InSelection(r types.LayerRect, l, c uint16) bool {
	if !this.CanCopy {
		return false
	}

	el := r.Y1
	sl := r.Y0
	ec := r.X1
	sc := r.X0

	if sl == el && ec < sc {
		el = r.Y1
		sl = r.Y0
		ec = r.X0
		sc = r.X1
	} else if el < sl {
		el = r.Y0
		sl = r.Y1
		ec = r.X0
		sc = r.X1
	}

	// line based
	if l > sl && l < el {
		return true
	} else if l == sl && l == el && c >= sc && c <= ec {
		return true
	} else if l == sl && sl != el && c >= sc {
		return true
	} else if l == el && sl != el && c <= ec {
		return true
	}

	return false

}

func (t *TextLayer) StartSelect(x, y int) {
	t.SelActive = true
	t.Buffer.Write(SelSX, uint64(x))
	t.Buffer.Write(SelSY, uint64(y))
	t.Buffer.Write(SelEX, uint64(x))
	t.Buffer.Write(SelEY, uint64(y))
}

func (t *TextLayer) DragSelect(x, y int) {

	if !t.SelActive {
		return
	}

	t.CanCopy = ((t.Buffer.Read(4089) & 16) == 16)

	if !t.CanCopy {
		t.Buffer.Write(SelSX, 0)
		t.Buffer.Write(SelSY, 0)
		t.Buffer.Write(SelEX, 0)
		t.Buffer.Write(SelEY, 0)
		return
	}

	empty := types.LayerRect{}
	selbounds := types.LayerRect{}
	selbounds.X0 = uint16(t.Buffer.Read(SelSX))
	selbounds.Y0 = uint16(t.Buffer.Read(SelSY))
	selbounds.X1 = uint16(t.Buffer.Read(SelEX))
	selbounds.Y1 = uint16(t.Buffer.Read(SelEY))

	oy := y
	ox := x

	if y > 0 {
		lh := t.lineHeights[y-1]
		if lh > 1 {
			y = y + lh - 2
			x = 79
		}
	}

	if y >= t.Height {
		y = t.Height - 1
	}

	lh := t.lineHeights[y]
	if lh > 1 {
		y = y + lh - 1
		x = 79
	}

	if selbounds == empty {
		// start
		t.Buffer.Write(SelSX, uint64(ox))
		t.Buffer.Write(SelSY, uint64(oy))
		t.Buffer.Write(SelEX, uint64(x))
		t.Buffer.Write(SelEY, uint64(y))
	} else {
		t.Buffer.Write(SelEX, uint64(x))
		t.Buffer.Write(SelEY, uint64(y))
	}

}

func (t *TextLayer) DoneSelect() {

	if !t.SelActive {
		return
	}

	t.SelActive = false

	t.CanCopy = ((t.Buffer.Read(4089) & 16) == 16)

	selbounds := types.LayerRect{}
	selbounds.X0 = uint16(t.Buffer.Read(SelSX))
	selbounds.Y0 = uint16(t.Buffer.Read(SelSY))
	selbounds.X1 = uint16(t.Buffer.Read(SelEX))
	selbounds.Y1 = uint16(t.Buffer.Read(SelEY))

	t.CopyText(selbounds)

	t.Buffer.Write(SelSX, 0)
	t.Buffer.Write(SelSY, 0)
	t.Buffer.Write(SelEX, 0)
	t.Buffer.Write(SelEY, 0)

	t.Buffer.Write(4089, t.Buffer.Read(4089)|256)

}

type InlinePos struct {
	SX, SY int
	EX, EY int
}

func (ip InlinePos) Dimensions(cw, ch int) (int, int) {
	return (ip.EX - ip.SX) * cw, (ip.EY - ip.SY) * ch
}

func (t *TextLayer) LoadInlines(data []uint64) {
	// process inline changes if any
	t.CheckInlines()

	var idx int
	var w, h float32
	var ch rune

	t.InlineBounds = make(map[rune][]InlinePos)

	for y := t.Height - 1; y >= 0; y -= 1 {
		for x := t.Width - 1; x >= 0; x -= 1 {
			if t.Format == 0 || t.Format == 13 {
				// linear
				idx = t.baseOffsetLinear(x, y)
			} else {
				if t.SwitchInterleave {
					idx = t.baseOffsetWozAlt(x, y)
				} else {
					idx = t.baseOffsetWoz(x, y)
				}
			}

			nv := data[idx]

			ch, _, _, _, _, w, h, _ = t.DecodePackedCharLegacy(nv)

			if len(t.InlineMap) > 0 {
				_, ok := t.InlineMap[ch]
				if ok {
					// mapped...
					list, ok := t.InlineBounds[ch]
					if !ok {
						list = make([]InlinePos, 0)
					}
					if len(list) == 0 {
						// add it
						list = append(list, InlinePos{SX: x, SY: y, EX: x + int(w), EY: y + int(h)})
					} else {
						lidx := len(list) - 1
						// maybe it could adjoin an existing one?
						dx := list[lidx].SX - x
						dy := list[lidx].SY - y

						if dx <= int(w) || dy <= int(h) {
							// continuous
							if x < list[lidx].SX {
								list[lidx].SX = x
							}
							if y < list[lidx].SY {
								list[lidx].SY = y
							}
						} else if x >= list[lidx].SX && x <= list[lidx].EX && y >= list[lidx].SY && y <= list[lidx].EY {
							// same deal... but no need to expand
						} else {
							list = append(list, InlinePos{SX: x, SY: y, EX: x + int(w), EY: y + int(h)})
						}
					}

					t.InlineBounds[ch] = list // update list
				}
			}

		}
	}

}

func (t *TextLayer) PointIsInline(x, y int) bool {
	for _, list := range t.InlineBounds {

		for _, b := range list {

			if x >= b.SX && x <= b.EX && y >= b.SY && y <= b.EY {
				return true
			}

		}
	}
	return false
}

type TextCell struct {
	x, y         int
	w, h         int
	ch           rune
	fgcol, bgcol int
	refreshed    bool
}

func (tc *TextCell) IsEmpty() bool {
	return (tc.ch == ' ') || (tc.bgcol == 0 && tc.fgcol == 0)
}

type TextCellManager struct {
	data [80][50]*TextCell
}

func (tcm *TextCellManager) Get(x, y int) *TextCell {

	return tcm.data[x][y]

}

func (tcm *TextCellManager) Add(tc *TextCell) (bool, *TextCell) {

	c := tcm.data[tc.x][tc.y]
	if c != nil {
		// there is a cell already here... is it something we cannot draw over?
		if !c.IsEmpty() {
			return false, c
		}
	}

	for y := tc.y; y < tc.y+tc.h; y++ {
		for x := tc.x; x < tc.x+tc.w; x++ {
			if x < 80 && y < 50 {
				tcm.data[x][y] = tc
			}
		}
	}

	return true, tc
}

func (t *TextLayer) MakeUpdatesTT() {

	var forceAll bool

	sf := t.BaseLayer.Spec.GetSubFormat()

	if sf != t.lastSubFormat {
		forceAll = true
	}

	if t.TintChanged && t.Spec.GetID() != "OOSD" {
		t.Palette = t.Spec.GetPalette()
		if t.Tint != nil {
			pa := t.Palette.Desaturate().Tint(t.Tint.Red, t.Tint.Green, t.Tint.Blue)
			t.Palette = *pa
		}
		t.TintChanged = false
		forceAll = true
	}

	bounds := t.Spec.GetBoundsRect()
	//boundsChanged := !bounds.Equals(t.LastBounds)

	ttbuffer := t.ttp.ProcessBuffer(t.framedata)

	var cell TeleTextCell
	for y := 0; y < ttxtRows; y++ {
		for x := 0; x < ttxtCols; x++ {
			cell = ttbuffer[y*ttxtCols+x]
			//log.Printf("x=%d, y=%d, celldata=%+v", x, y, cell)
			if cell.displayMode == ttxtCharConceal || (cell.displayMode == ttxtCharFlash && !t.FlashOn) {
				cell.ch = ' '
			}
			va := types.VA_NORMAL
			if cell.displayMode == ttxtCharFlash {
				va = types.VA_BLINK
			}
			if cell != t.lastTTBuffer[y*ttxtCols+x] || forceAll || t.FlashChanged() {
				// cell changed so render
				t.PlotPixelFont(
					cell.ch,
					cell.colidx,
					cell.bcolidx,
					0,
					va,
					1,
					1,
					true,
					x,
					y,
				)
			}
		}
	}

	// end sync up
	t.lastTTBuffer = ttbuffer

	t.Changed = false
	t.NeedRefresh = false
	t.LastCX = t.CX
	t.LastCY = t.CY
	t.PosChanged = false
	t.LastBounds = bounds
	//t.LastSelBounds = selbounds

	if t.BitmapUpdate {
		t.ScreenTex.SetSourceSame(t.Bitmap)
		t.BitmapUpdate = false
	}

	//	t.lastSubFormat = sf
}

func (t *TextLayer) MakeUpdates() {
	if t.Format == types.LF_BBC_TELE_TEXT {
		t.MakeUpdatesTT()
	} else {
		t.MakeUpdatesApple()
	}
}

func (t *TextLayer) MakeUpdatesApple() {

	index := t.Buffer.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
	t.UseAltSet = t.Buffer.GetMM().IntGetAltChars(index)

	t.RedoArea = image.Rect(0, 0, t.ScreenTexW, 0)

	//data := t.Buffer.ReadSlice(0, 4096)

	t.LoadInlines(t.framedata)

	t.CX = int(t.Buffer.Read(4094))
	t.CY = int(t.Buffer.Read(4095))

	forceAll := ((t.Buffer.Read(4089) & 256) != 0)
	//forceAll := true

	t.Buffer.Write(4089, t.Buffer.Read(4089)&255)

	t.CursorVisible = ((t.Buffer.Read(4089) & 128) != 128)

	//t.UseAltSet = ((t.Buffer.Read(4089) & 64) == 64)

	t.CanCopy = ((t.Buffer.Read(4089) & 16) == 16)

	tmpswi := ((t.Buffer.Read(4089) & 32) == 32)
	if tmpswi != t.SwitchInterleave {
		forceAll = true // important
	}
	t.SwitchInterleave = tmpswi

	if t.TintChanged && t.Spec.GetID() != "OOSD" {
		t.Palette = t.Spec.GetPalette()
		if t.Tint != nil {
			pa := t.Palette.Desaturate().Tint(t.Tint.Red, t.Tint.Green, t.Tint.Blue)
			t.Palette = *pa
		}
		t.TintChanged = false
		forceAll = true
	}

	//	forceAll = true

	cx := t.CX
	cy := t.CY

	var idx int
	var ch rune
	var colidx uint64
	var bcolidx uint64
	var shade uint64
	var nv, ov uint64
	var va types.VideoAttribute
	var w float32
	var h float32
	var solid bool
	var tcm TextCellManager

	if t.Palette.Size() == 0 {
		t.Palette = t.Spec.GetPalette()
	}

	bounds := t.Spec.GetBoundsRect()
	boundsChanged := !bounds.Equals(t.LastBounds)

	selbounds := types.LayerRect{}
	if t.CanCopy && t.SelActive {
		selbounds.X0 = uint16(t.Buffer.Read(SelSX))
		selbounds.Y0 = uint16(t.Buffer.Read(SelSY))
		selbounds.X1 = uint16(t.Buffer.Read(SelEX))
		selbounds.Y1 = uint16(t.Buffer.Read(SelEY))
	}

	var isCursor bool
	var isLastCursor bool

	var now int64 = time.Now().UnixNano()

	xskip := 1
	yskip := 1
	forceSize := false

	var indexFunc func(x, y int) int
	indexFunc = t.baseOffsetLinear
	if t.Format != 0 {
		if t.SwitchInterleave {
			indexFunc = t.baseOffsetWozAlt
		} else {
			indexFunc = t.baseOffsetWoz
		}
	}

	if !forceAll {
		forceAll = t.NeedRefresh
	}

	sf := t.BaseLayer.Spec.GetSubFormat()

	if sf != t.lastSubFormat {
		forceAll = true
		//fmt.RPrintln("LAYER STATE CHANGED: " + t.Spec.String())
	}

	var useZero = sf != types.LSF_FREEFORM

	if sf != types.LSF_FREEFORM {
		t.CursorVisible = false // disable softcursor
		forceSize = true
		switch sf {
		case types.LSF_FIXED_40_24:
			indexFunc = t.baseOffsetWoz
			xskip = 2
			yskip = 2
		case types.LSF_FIXED_80_24:
			indexFunc = t.baseOffsetWozAlt
			xskip = 1
			yskip = 2
		case types.LSF_FIXED_40_48:
			indexFunc = t.baseOffsetWoz
			xskip = 2
			yskip = 1
		case types.LSF_FIXED_80_48:
			indexFunc = t.baseOffsetWozAlt
			xskip = 1
			yskip = 1
		}
	}

	if t.Format == types.LF_BBC_TELE_TEXT {
		log.Println("Format is teletext")
		indexFunc = t.baseOffsetLinear
		xskip = 1
		yskip = 1
		forceSize = true
	}

	for y := 0; y < t.Height; y += yskip {

		for x := 0; x < t.Width; x += xskip {

			idx = indexFunc(x, y)

			nv = t.framedata[idx]
			ov = t.PreviousBuffer[idx]
			t.PreviousBuffer[idx] = nv

			ch, va, colidx, bcolidx, shade, w, h, solid = t.DecodePackedCharLegacy(nv)

			// if x == 0 && y == 1 {
			// 	if ov != nv {
			// 		fmt.RPrintf("ch=%v, colidx=%v bcolidx=%v, w=%v, h=%v\n", ch, colidx, bcolidx, w, h)
			// 	}
			// }

			if forceSize {
				w = float32(xskip)
				h = float32(yskip)
				colidx = 15
				bcolidx = 0
				shade = 0
				if t.Format == types.LF_BBC_TELE_TEXT {
					w = 2
					h = 2
				}
			}

			if !bounds.Contains(uint16(x), uint16(y)) {
				if boundsChanged {
					t.PlotBlank(w, h, solid, x, y)
					tcm.Add(&TextCell{
						bgcol: 0,
						fgcol: 15,
						ch:    ' ',
						x:     x,
						y:     y,
						w:     int(w),
						h:     int(h),
					})
				}
				continue
			}

			isCursor = (y == cy && x == cx)
			isLastCursor = (y == t.LastCY && x == t.LastCX)

			boundsSelect := t.SelActive && t.InSelection(selbounds, uint16(y), uint16(x))

			if boundsChanged == false {
				boundsChanged = t.SelActive && (selbounds != t.LastSelBounds)
			}

			isInline := t.PointIsInline(x, y)

			c := tcm.Get(x, y)
			tcmOverlayRedraw := false
			if c != nil {
				tcmOverlayRedraw = c.refreshed && !boundsSelect
			}

			if tcmOverlayRedraw || isInline || boundsSelect || forceAll || boundsChanged || va == types.VA_BLINK || isCursor || isLastCursor || (nv != ov) {

				//If we see a space or a fill
				if nv == cmSpace && !useZero {
					continue
					// ch = ' '
					// w = 1
					// h = 1
					// bcolidx = 2
				}

				if nv == cmFilled {
					if ov != cmFilled && ov != cmSpace && c == nil {
						t.PlotBlank(1, 1, true, x, y)
					}
					continue
				}

				if isCursor && t.CursorOn && t.CursorVisible {
					if w == 1 {
						va = types.VA_INVERSE
						ch = ' '
					} else {
						ch = 127
					}
				}

				// if a selection we invert it...
				if boundsSelect && nv != cmSpace {
					a := colidx
					colidx = bcolidx
					bcolidx = a
				}

				// the text cell manager decides if we can draw this char...
				if ok, cell := tcm.Add(&TextCell{
					bgcol: int(bcolidx),
					fgcol: int(colidx),
					ch:    ch,
					x:     x,
					y:     y,
					w:     int(w),
					h:     int(h),
				}); ok {
					t.PlotPixelFont(ch, int(colidx), int(bcolidx), int(shade), va, w, h, solid, x, y)
					t.CellLastRefreshed[y*80+x] = now // mark cell as redrawn
					cell.refreshed = true
					// if x == 0 && y == 1 {
					// 	fmt.RPrintln("plot ok")
					// }
				}
			}
		}
	}

	//~ // update last cursor pos and set changed flag false
	t.Changed = false
	t.NeedRefresh = false
	t.LastCX = t.CX
	t.LastCY = t.CY
	t.PosChanged = false
	t.LastBounds = bounds
	t.LastSelBounds = selbounds

	if t.BitmapUpdate {
		t.ScreenTex.SetSourceSame(t.Bitmap)
		//t.ScreenTex.SetSourcePartial(int32(t.RedoArea.Min.X), int32(t.RedoArea.Min.Y), t.Bitmap.SubImage(t.RedoArea).(*image.RGBA))
		t.BitmapUpdate = false
		//fmt.Printf("Bitmap changed in area %d, %d - %d, %d\n", t.RedoArea.Min.X, t.RedoArea.Min.Y, t.RedoArea.Max.X, t.RedoArea.Max.Y )
	}

	t.lastSubFormat = sf

}

func (t *TextLayer) UpdateCursor(x, y int) {
	t.CX = x
	t.CY = y
	t.Changed = true
}

func (t *TextLayer) CursorChanged() bool {

	result := false

	con := ((time.Now().UnixNano()/1000000)%t.CursorInterval > t.CursorInterval/2)

	if con != t.CursorOn {
		result = true
	}

	t.CursorOn = con
	return result
}

func (t *TextLayer) FlashChanged() bool {

	result := false

	fon := ((time.Now().UnixNano()/1000000)%t.FlashInterval > t.FlashInterval/2)

	if fon != t.FlashOn {
		result = true
	}

	t.FlashOn = fon
	return result
}
