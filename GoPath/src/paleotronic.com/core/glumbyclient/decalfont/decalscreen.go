package decalfont

import (
	"paleotronic.com/log"
	//"os"
	//"image"
	"image/color"
	//"paleotronic.com/log"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"paleotronic.com/core/hires"
	"paleotronic.com/core/types"
	"paleotronic.com/glumby"
)

type DecalAttr rune

const (
	NORMAL    DecalAttr = 'N'
	FLASH     DecalAttr = 'F'
	INVERSE   DecalAttr = 'I'
	USEWOZHGR           = true
)

type DecalDuckScreen struct {
	Rows, Cols               int
	Width, Height            float32
	DW, DH                   float32
	Data                     []*Decal
	NormalFont               DecalDuckFont
	InvertedFont             DecalDuckFont
	CX, CY, LastCX, LastCY   float32
	CursorVisible            bool
	FlashInterval            int64
	BlinkInterval            int64
	BlinkOn                  bool
	FlashOn                  bool
	TextMemory               types.TXMemoryBuffer
	RedrawNextFrame          bool
	ForceFull                bool
	RedrawAddr               int
	RedrawData               []int
	RedrawCount              int
	BGColor                  color.RGBA
	lastFrame                []int
	Batch                    *DecalBatch
	Mesh, BrickMesh, DotMesh *glumby.Mesh
	Mode                     *types.VideoMode
	HGR                      [2]hires.IndexedVideoBuffer
	WozHGR                   [2]hires.HGRScreen
	ValidCharMin             int
	ValidCharMax             int
	DisplayPage, CurrentPage int
}

func (this *DecalDuckScreen) RedrawIfRequired() {

	// sync
	//for i, v := range this.RedrawData {
	//		this.TextMemory[this.RedrawAddr+i] = v
	//	}

	// get current exact state
	currentFrame := this.TextMemory.GetValues(0, this.TextMemory.Size())

	if this.ForceFull {
		currentFrame = make([]int, 2048)
		this.ForceFull = false
	}

	//r.Clear(image.Rect(0, 0, r.Bounds().Max.X, r.Bounds().Max.Y), this.BGColor)
	//r.ClearDepth(image.Rect(0, 0, 0, 0), 1.0)

	flashOn := (time.Now().UnixNano()%(this.FlashInterval*2) > this.FlashInterval)

	flashChange := (flashOn != this.FlashOn)

	batchChunk := make([]*Decal, 0)

	for y := 0; y < this.Rows; y++ {

		for x := 0; x < this.Cols; x++ {

			didx := (y * this.Cols) + x
			//idx := XYToOffset(this.Cols, x, y)
			idx := (y * this.Cols) + x

			if (didx < len(this.Data)) && (y >= this.ValidCharMin) && (y <= this.ValidCharMax) {
				batchChunk = append(batchChunk, this.Data[didx])
			}

			// do char if it has changed
			if (currentFrame[idx] != this.lastFrame[idx]) || (flashChange) {

				//rr := image.Rect(int(float32(x)*this.DW), int((float32(this.Rows)-float32(y)-1)*this.DH), 1+int((float32(x)+1)*this.DW-1), 1+int((float32(this.Rows)-float32(y))*this.DH-1))

				decal := this.Data[didx]
				memval := currentFrame[idx]
				ch := rune(PokeToAscii(memval))
				va := PokeToAttribute(memval)

				//log.Printf("Change @ x = %d, y = %d, ch = [%s]\n", x, y, string(ch))

				if (va == types.VA_INVERSE) || ((va == types.VA_BLINK) && flashOn) {
					if this.Cols == 80 {
						decal.Texture = this.InvertedFont.Chars80[ch]
					} else {
						decal.Texture = this.InvertedFont.Chars[ch]
					}
				} else {
					if this.Cols == 80 {
						decal.Texture = this.NormalFont.Chars80[ch]
					} else {
						decal.Texture = this.NormalFont.Chars[ch]
					}
				}

				//r.Clear(rr, this.BGColor)
				//r.Draw(image.Rect(0, 0, 0, 0), decal, camera)
				//r.Render()
			} else {
				//log.Printf("No Change @ x = %d, y = %d\n", x, y)
			}
		}

	}

	this.FlashOn = flashOn

	// render it
	this.Batch.Items = batchChunk
	this.Batch.Render()

	//this.Render(r, camera)

	//this.RenderCursor(r, camera)

	this.RedrawNextFrame = false
	this.lastFrame = currentFrame
}

func (this *DecalDuckScreen) SetCell(x, y int, ch rune, attr DecalAttr) {

	if attr == NORMAL {
		this.NormalFont.Draw(string(ch), attr, x, y, int(this.Width), int(this.Height), this.Data)
	} else {
		this.InvertedFont.Draw(string(ch), attr, x, y, int(this.Width), int(this.Height), this.Data)
	}

}

func (this *DecalDuckScreen) SetCursorVisible(cursorOn bool) {
	this.CursorVisible = cursorOn
}

func (this *DecalDuckScreen) RenderCursor() {

	// do we need to change???

	cursorOn := (time.Now().UnixNano()%(this.BlinkInterval*2) > this.BlinkInterval)

	if this.CY < float32(this.ValidCharMin) {
		cursorOn = false
	}

	if !this.CursorVisible {
		cursorOn = false
	}

	if cursorOn {

		fx := this.DW * float32(this.CX)
		fy := float32(this.Rows-int(this.CY)-1)*this.DH + 0
		rr := glumby.Vector3{X: float32(fx + this.DW/2), Y: float32(fy + this.DH/2), Z: 1}

		var decal *Decal
		if this.Cols == 80 {
			decal = &Decal{Texture: this.InvertedFont.Chars80[32], Mesh: this.GetDecalMesh(this.DW/2, this.DH/2, 0.0), Name: "cursor", Position: rr}
		} else {
			decal = &Decal{Texture: this.NormalFont.Chars[127], Mesh: this.GetDecalMesh(this.DW/2, this.DH/2, 0.0), Name: "cursor", Position: rr}
		}

		this.Batch.Items = []*Decal{decal}
		this.Batch.Render()

	}

	this.BlinkOn = cursorOn

	this.LastCX = this.CX
	this.LastCY = this.CY

}

func (this *DecalDuckScreen) Render() {

	//	log.Println(r.Bounds())
	//r.Clear(image.Rect(0, 0, 1100, 748), gfx.Color{0, 0, 1, 0.5})

	//	log.Println(camera.Bounds())

	if (this.Mode != nil) && (this.Mode.ActualRows != this.Mode.Rows) {
		if this.Mode.Width > 50 {
			this.RenderAppleIIHGR(0)
		} else {
			this.RenderAppleIIGR()
		}
	}

	this.RedrawIfRequired()
	this.RenderCursor()

}

func (this *DecalDuckScreen) GetDotMesh() *glumby.Mesh {

	if this.DotMesh != nil {
		return this.DotMesh
	} else {
		m := glumby.NewMesh(gl.POINTS)

		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(0, 0, 0)

		this.DotMesh = m

		return this.DotMesh
	}

}

func (this *DecalDuckScreen) GetDecalMesh(w, h, d float32) *glumby.Mesh {

	if this.Mesh != nil {
		return this.Mesh
	} else {
		m := glumby.NewMesh(gl.QUADS)

		m.Normal3f(0, 0, 1)
		m.TexCoord2f(0, 0)
		m.Vertex3f(-1*w, -1*h, 0)

		m.Normal3f(0, 0, 1)
		m.TexCoord2f(1, 0)
		m.Vertex3f(1*w, -1*h, 0)

		m.Normal3f(0, 0, 1)
		m.TexCoord2f(1, 1)
		m.Vertex3f(1*w, 1*h, 0)

		m.Normal3f(0, 0, 1)
		m.TexCoord2f(0, 1)
		m.Vertex3f(-1*w, 1*h, 0)

		this.Mesh = m
	}

	return this.Mesh
}

func (this *DecalDuckScreen) GetBrickMesh(w, h, d float32) *glumby.Mesh {

	if this.BrickMesh != nil {
		return this.BrickMesh
	} else {
		m := glumby.NewMesh(gl.QUADS)

		m.Normal3f(0, 0, 1)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(-1*w, -1*h, 1*d)

		m.Normal3f(0, 0, 1)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(1*w, -1*h, 1*d)

		m.Normal3f(0, 0, 1)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(1*w, 1*h, 1*d)

		m.Normal3f(0, 0, 1)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(-1*w, 1*h, 1*d)

		m.Normal3f(0, 0, -1)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(-1*w, -1*h, -1*d)

		m.Normal3f(0, 0, -1)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(-1*w, 1*h, -1*d)

		m.Normal3f(0, 0, -1)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(1*w, 1*h, -1*d)

		m.Normal3f(0, 0, -1)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(1*w, -1*h, -1*d)

		m.Normal3f(0, 1, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(-1*w, 1*h, -1*d)

		m.Normal3f(0, 1, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(-1*w, 1*h, 1*d)

		m.Normal3f(0, 1, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(1*w, 1*h, 1*d)

		m.Normal3f(0, 1, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(1*w, 1*h, -1*d)

		m.Normal3f(0, -1, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(-1*w, -1*h, -1*d)

		m.Normal3f(0, -1, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(1*w, -1*h, -1*d)

		m.Normal3f(0, -1, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(1*w, -1*h, 1*d)

		m.Normal3f(0, -1, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(-1*w, -1*h, 1*d)

		m.Normal3f(1, 0, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(1*w, -1*h, -1*d)

		m.Normal3f(1, 0, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(1*w, 1*h, -1*d)

		m.Normal3f(1, 0, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(1*w, 1*h, 1*d)

		m.Normal3f(1, 0, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(1*w, -1*h, 1*d)

		m.Normal3f(-1, 0, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(-1*w, -1*h, -1*d)

		m.Normal3f(-1, 0, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(-1*w, -1*h, 1*d)

		m.Normal3f(-1, 0, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(-1*w, 1*h, 1*d)

		m.Normal3f(-1, 0, 0)
		m.Color4f(0, 0, 0, 0)
		m.Vertex3f(-1*w, 1*h, -1*d)

		this.BrickMesh = m
	}

	return this.BrickMesh
}

func (this *DecalDuckScreen) ApplyMode(vm *types.VideoMode) {

	this.Mode = vm

	// Save old data states?
	lf := this.lastFrame
	tm := this.TextMemory

	this.Configure(this.Width, this.Height, vm.Columns, vm.Rows)

	this.lastFrame = lf
	this.TextMemory = tm

	// Actual rows support
	this.ValidCharMin = 0
	this.ValidCharMax = vm.Rows - 1

	if vm.ActualRows < vm.Rows {
		this.ValidCharMin = vm.Rows - vm.ActualRows
	}

}

func (this *DecalDuckScreen) Configure(w, h float32, c, r int) {
	this.Mesh = nil // force mesh reconfiguration
	this.BrickMesh = nil

	this.Rows = r
	this.Cols = c
	this.Width = w
	this.Height = h

	this.DW = this.Width / float32(this.Cols)
	this.DH = this.Height / float32(this.Rows)

	log.Printf("=========================> Decal Size is %f x %f\n", this.DW, this.DH)

	//os.Exit(0)

	//	log.Printf("Char cell dimensions are %v x %v\n", this.DW, this.DH)

	this.Data = make([]*Decal, r*c)

	this.FlashInterval = 250000000
	this.BlinkInterval = 900000000
	this.CursorVisible = true
	this.CX = 0
	this.CY = 0
	this.TextMemory = *types.NewTXMemoryBuffer(2048)
	this.TextMemory.Silent(true) // turn off change logging
	this.lastFrame = make([]int, 2048)
	this.BGColor = color.RGBA{0, 0, 0, 255}

	cardMesh := this.GetDecalMesh(this.DW/2, this.DH/2, 0.0)
	this.BrickMesh = this.GetBrickMesh(this.Width/80, this.DH/4, 0.5)
	this.DotMesh = this.GetDotMesh()

	for yy := 0; yy < this.Rows; yy++ {
		for xx := 0; xx < this.Cols; xx++ {
			fx := this.DW * float32(xx)
			fy := float32(this.Rows-yy-1)*this.DH + 0
			idx := (yy * this.Cols) + xx

			ru := rune(32)

			this.Data[idx] = &Decal{}
			this.Data[idx].Mesh = cardMesh
			this.Data[idx].Name = string(ru) + string(NORMAL)
			this.Data[idx].Texture = this.NormalFont.Chars[ru]
			//this.Data[idx].SetScale(lmath.Vec3{float64(this.DW), float64(this.DH)}, 1)
			this.Data[idx].Position = glumby.Vector3{X: float32(fx + this.DW/2), Y: float32(fy + this.DH/2), Z: 1}

			// tint stuff?

		}
	}
}

func (this *DecalDuckScreen) RenderAppleIIHGRWoz(idx int) {

	gl.BindTexture(gl.TEXTURE_2D, 0)

	if this.Mode == nil {
		//log.Println("NULL MODE")
		return
	}

	if this.Mode.Rows == this.Mode.ActualRows {
		//log.Println("Same rows")
		return
	}

	videoRows := this.Mode.Rows - this.Mode.ActualRows
	maxy := videoRows * 8
	maxx := 280

	if videoRows == 0 {
		//log.Println("No rows")
		return
	}

	dotMesh := this.GetDotMesh()

	gl.PointSize(this.DW / 3.5)
	gl.Enable(gl.POINT_SMOOTH)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	glumby.MeshBuffer_Begin(gl.POINTS)
	// Assume we are good -
	for y := 0; y < maxy; y++ {
		offs := this.WozHGR[idx].XYToOffset(0, y)
		scanline := this.WozHGR[idx].ColorsForScanLine(this.WozHGR[idx].Data[offs : offs+40])
		for x := 0; x < maxx; x++ {
			v := scanline[x]
			if v > 0 {
				fx := (this.DW / 7) * float32(x)
				fy := float32(maxy-y-1)*(this.DH/8) + (this.DH * float32(this.Mode.ActualRows))
				fz := float32(0) // use 1.1 to prevent z-fighting
				c := this.Mode.Palette.Get(v)

				//log.Printf("COL = %v\n", c)

				if c == nil {
					return
				}

				r := float32(c.Red) / float32(255)
				g := float32(c.Green) / float32(255)
				b := float32(c.Blue) / float32(255)
				a := float32(c.Alpha) / float32(255)

				dotMesh.SetColor(r, g, b, a*0.9)
				dotMesh.DrawWithMeshBuffer(fx, fy, fz)
			}
		}
	}
	glumby.MeshBuffer_End()

}

func (this *DecalDuckScreen) RenderAppleIIHGR(idx int) {

	gl.BindTexture(gl.TEXTURE_2D, 0)

	if this.Mode == nil {
		//log.Println("NULL MODE")
		return
	}

	if this.Mode.Rows == this.Mode.ActualRows {
		//log.Println("Same rows")
		return
	}

	videoRows := this.Mode.Rows - this.Mode.ActualRows
	maxy := videoRows * 8
	maxx := 280

	if videoRows == 0 {
		//log.Println("No rows")
		return
	}

	dotMesh := this.GetDotMesh()

	gl.PointSize(this.DW / 3.5)
	gl.Enable(gl.POINT_SMOOTH)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	glumby.MeshBuffer_Begin(gl.POINTS)
	// Assume we are good -
	for y := 0; y < maxy; y++ {
		for x := 0; x < maxx; x++ {
			v := this.HGR[idx].ColorAt(x, y)
			if v > 0 {
				fx := (this.DW / 7) * float32(x)
				fy := float32(maxy-y-1)*(this.DH/8) + (this.DH * float32(this.Mode.ActualRows))
				fz := float32(0) // use 1.1 to prevent z-fighting
				c := this.Mode.Palette.Get(v)

				//log.Printf("COL = %v\n", c)

				if c == nil {
					return
				}

				r := float32(c.Red) / float32(255)
				g := float32(c.Green) / float32(255)
				b := float32(c.Blue) / float32(255)
				a := float32(c.Alpha) / float32(255)

				dotMesh.SetColor(r, g, b, a*0.9)
				dotMesh.DrawWithMeshBuffer(fx, fy, fz)
			}
		}
	}
	glumby.MeshBuffer_End()

}

// Renders the GR mode screen - assume camera is set already
func (this *DecalDuckScreen) RenderAppleIIGR() {

	gl.BindTexture(gl.TEXTURE_2D, 0)

	if this.Mode == nil {
		//log.Println("NULL MODE")
		return
	}

	if this.Mode.Rows == this.Mode.ActualRows {
		//log.Println("Same rows")
		return
	}

	videoRows := this.Mode.Rows - this.Mode.ActualRows
	maxy := videoRows * 2
	maxx := 40

	if videoRows == 0 {
		//log.Println("No rows")
		return
	}

	brickMesh := this.GetBrickMesh(this.Width/40, this.DH/4, this.DH/4)

	// get current exact state
	currentFrame := this.TextMemory.GetValues(0, this.TextMemory.Size())

	glumby.MeshBuffer_Begin(gl.QUADS)

	//log.Printf("%d x %d\n", maxx, maxy)

	for y := 0; y < maxy; y++ {

		for x := 0; x < maxx; x++ {
			// offset to videomemory
			//idx := XYToOffset(this.Cols, x, y/2)
			idx := ((y / 2) * this.Cols) + x

			d := currentFrame[idx]
			v := d & 0xf
			if (y % 2) == 1 {
				v = (d >> 4) & 0xf
			}

			if v != 0 {
				//log.Printf("Brick at %d, %d\n", x, y)

				// Do cuboid
				fx := this.DW*float32(x) + this.DH/2
				fy := float32(maxy-y-1)*(this.DH/2) + (this.DH * float32(this.Mode.ActualRows)) + this.DH/2
				fz := float32(0) // use 1.1 to prevent z-fighting

				//log.Printf("P size = %d\n", this.Mode.Palette.Size())

				c := this.Mode.Palette.Get(v)

				//log.Printf("COL = %v\n", c)

				if c == nil {
					return
				}

				r := float32(c.Red) / float32(255)
				g := float32(c.Green) / float32(255)
				b := float32(c.Blue) / float32(255)
				a := float32(c.Alpha) / float32(255)

				brickMesh.SetColor(r, g, b, a*0.9)
				brickMesh.DrawWithMeshBuffer(fx, fy, fz)

				//log.Printf("DrawMesh called %f,%f,%f,%f\n", r, g, b, a)

			}

		}

	}

	glumby.MeshBuffer_End()

}

func NewDecalDuckScreen(w, h float32, c, r int, fontNormal, fontInverted DecalDuckFont) *DecalDuckScreen {
	this := &DecalDuckScreen{}

	this.ForceFull = true

	this.NormalFont = fontNormal
	this.InvertedFont = fontInverted

	this.Configure(w, h, c, r)

	this.HGR[0] = *hires.NewIndexedVideoBuffer(280, 192)
	this.HGR[1] = *hires.NewIndexedVideoBuffer(280, 192)

	this.HGR[0].Fill(0)
	this.HGR[1].Fill(0)

	this.WozHGR[0] = hires.HGRScreen{}
	this.WozHGR[1] = hires.HGRScreen{}

	this.Batch = NewDecalBatch()

	log.Println("Created screen")

	return this
}

func (this *DecalDuckScreen) HFill2D(c int) {
	this.HGR[this.CurrentPage].Fill(c)
	this.WozHGR[this.CurrentPage].Fill(c)
}

func (this *DecalDuckScreen) HPlot2D(x, y, c int) {

	this.HGR[this.CurrentPage].Plot(x, y, c)
	this.WozHGR[this.CurrentPage].Plot(x, y, c)
	//	log.Printf("SetTextMemory(%d,%d)\n", idx, v)

}

func (this *DecalDuckScreen) Plot2D(x, y, c int) {

	idx := x + (y/2)*this.Mode.Width

	v := this.TextMemory.GetValue(idx)

	if (y % 2) == 0 {
		v = (v & 0xf0) | (c & 0xf)
	} else {
		v = (v & 0xf) | ((c & 0xf) << 4)
	}

	this.TextMemory.SetValue(idx, v)

	//	log.Printf("SetTextMemory(%d,%d)\n", idx, v)

}

func (this *DecalDuckScreen) HLine2D(x0, y0, x1, y1, c int) {

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
		this.HPlot2D(x0, y0, c)
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

func (this *DecalDuckScreen) Line2D(x0, y0, x1, y1, c int) {

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
		this.Plot2D(x0, y0, c)
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
