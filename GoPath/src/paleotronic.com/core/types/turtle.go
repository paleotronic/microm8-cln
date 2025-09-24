package types

import (
	"encoding/json"
	"errors"
	"strings"

	"math"

	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/core/memory"
	"paleotronic.com/fmt" //	"math"

	"paleotronic.com/core/settings"

	"paleotronic.com/octalyzer/video/font"
)

// TurtleCommandType indicates the type of instruction
type TurtleCommandType int

const (
	TURTLE_CLEAR TurtleCommandType = iota
	TURTLE_HOME
	TURTLE_SETHOME
	TURTLE_PENUP
	TURTLE_PENDOWN
	TURTLE_FD
	TURTLE_BK
	TURTLE_UP
	TURTLE_DN
	TURTLE_LT
	TURTLE_RT
	TURTLE_RL
	TURTLE_RR
	TURTLE_PENCOL
	TURTLE_FILLCOL
	TURTLE_POSX
	TURTLE_POSY
	TURTLE_POSZ
	TURTLE_HEADING
	TURTLE_SHOW
	TURTLE_HIDE
	TURTLE_SPOS
	TURTLE_TRIANGLE
	TURTLE_SPHR
	TURTLE_PITCH
	TURTLE_ROLL
	TURTLE_CUBE
	TURTLE_QUAD
	TURTLE_LINECUBE
	TURTLE_LINEQUAD
	TURTLE_CIRCLE
	TURTLE_LINECIRCLE
	TURTLE_SPHERE
	TURTLE_RAISE
	TURTLE_LOWER
	TURTLE_POLY
	TURTLE_LINEPOLY
	TURTLE_LINEARC
	TURTLE_ARC
	TURTLE_LINETRIANGLE
	TURTLE_PYRAMID
	TURTLE_LINEPYRAMID
	TURTLE_PENOPACITY
	TURTLE_FILLOPACITY
	TURTLE_SHUFFLE_LEFT
	TURTLE_SHUFFLE_RIGHT
	TURTLE_GLYPH
	TURTLE_GLYPH_FONT
	TURTLE_GLYPH_DEPTH
	TURTLE_GLYPH_SIZE
	TURTLE_GLYPH_STRETCH
	TURTLE_GLYPH_COLOR
	TURTLE_GLYPH_FILLED
)

type TurtleBoundsMode int

const (
	WINDOW TurtleBoundsMode = iota
	FENCE
	WRAP
)

type TurtleCommand struct {
	Type   TurtleCommandType `json:"commandType,omitempty"`
	FValue float32           `json:"floatVal,omitempty"`
	IValue int32             `json:"intVal,omitempty"`
	VValue *mgl64.Vec3       `json:"vecVal,omitempty"`
	Tag    string            `json:"tag,omitempty"`
	Hidden bool
}

type Turtle struct {
	HomePosition mgl64.Vec3              `json:"homePosition,omitempty"`
	Position     mgl64.Vec3              `json:"position,omitempty"`
	ViewDir      mgl64.Vec3              `json:"viewDir,omitempty"`
	UpDir        mgl64.Vec3              `json:"upDir,omitempty"`
	RightDir     mgl64.Vec3              `json:"rightDir,omitempty"`
	Heading      float64                 `json:"heading,omitempty"`
	Pitch        float64                 `json:"pitch,omitempty"`
	Roll         float64                 `json:"roll,omitempty"`
	Track        []*TurtleCommand        `json:"track,omitempty"`
	PenUp        bool                    `json:"penUp,omitempty"`
	PenErase     bool                    `json:"penErase,omitempty"`
	Hide         bool                    `json:"hideTurtle,omitempty"`
	PenColor     uint32                  `json:"colorPen,omitempty"`
	FillColor    uint32                  `json:"colorFill,omitempty"`
	vb           *VectorBuffer           `json:"-"`
	BoundsMode   TurtleBoundsMode        `json:"boundsMode,omitempty"`
	bx1          float64                 `json:"boundsX1,omitempty"`
	by1          float64                 `json:"boundsY1,omitempty"`
	bx2          float64                 `json:"boundsX2,omitempty"`
	by2          float64                 `json:"boundsY2,omitempty"`
	FontSize     float64                 `json:"fontSize,omitempty"`
	FontDepth    float64                 `json:"fontDepth,omitempty"`
	FontStretch  float64                 `json:"fontStretch,omitempty"`
	FontFace     int                     `json:"fontFace,omitempty"`
	FontFilled   bool                    `json:"fontFill,omitempty"`
	FontData     map[int]*font.DecalFont `json:"-"`

	TurtleDefinition []*TurtleCommand

	tag string `json:"tag",omitempty`
}

func NewTurtle(vb *VectorBuffer) *Turtle {
	this := &Turtle{
		vb:               vb,
		Track:            make([]*TurtleCommand, 0),
		TurtleDefinition: make([]*TurtleCommand, 0),
	}
	this.Reset()
	_ = json.Unmarshal([]byte(turtleDef), &this.TurtleDefinition)
	return this
}

func (t *Turtle) ToJSON() ([]byte, error) {
	data, err := json.Marshal(t)
	return data, err
}

func (t *Turtle) FromJSON(data []byte) error {
	t.Reset()
	return json.Unmarshal(data, t)
}

// func (t *Turtle) SaveToFile(filename string) error {
// 	data, err := t.ToJSON()
// 	if err != nil {
// 		return err
// 	}
// 	return files.WriteBytesViaProvider(files.GetPath(filename), files.GetFilename(filename), data)
// }

// func (t *Turtle) LoadFromFile(filename string) error {
// 	fr, err := files.ReadBytesViaProvider(files.GetPath(filename), files.GetFilename(filename))
// 	if err != nil {
// 		return err
// 	}
// 	return t.FromJSON(fr.Content)
// }

func (t *Turtle) Reset() {
	t.Position = t.HomePosition
	t.ViewDir = mgl64.Vec3{0, 1, 0}
	t.UpDir = mgl64.Vec3{0, 0, -1}
	t.RightDir = mgl64.Vec3{1, 0, 0}
	t.PenUp = false
	t.Hide = false
	t.PenColor = 23
	t.FillColor = 23
	t.Heading = 0
	t.Pitch = 0
	t.Roll = 0
	t.FontFace = 0
	t.FontSize = 10
	t.FontData = make(map[int]*font.DecalFont)
	t.FontDepth = 1
	t.FontStretch = 1
	t.FontFilled = true
	//t.tag = ""
	t.SetBounds(-140, -96, 140, 96)
	// t.Track = make([]*TurtleCommand, 0)
	// t.vb.Vectors = make(VectorList, 0)
}

func (pc *Turtle) RotateAxis(a float64, axis mgl64.Vec3) {
	r := mgl64.DegToRad(a)

	m := mgl64.HomogRotate3D(r, axis)
	//pc.Position = mgl64.TransformCoordinate(pc.Position, m)
	pc.UpDir = mgl64.TransformNormal(pc.UpDir, m).Normalize()
	pc.RightDir = mgl64.TransformNormal(pc.RightDir, m).Normalize()
	pc.ViewDir = mgl64.TransformNormal(pc.ViewDir, m).Normalize()

	//	//fmt.Printf("After rotate of %f, viewdir is %f, %f, %f (total %f)\n", a, pc.ViewDir[0], pc.ViewDir[1], pc.ViewDir[2], pc.ViewDir[0]+pc.ViewDir[1]+pc.ViewDir[2])
}

func (t *Turtle) LoadModelFromTag(tag string) {
	t.tag = ""
	newc := make([]*TurtleCommand, 0)
	for _, c := range t.Track {
		if strings.ToLower(c.Tag) == strings.ToLower(tag) {
			newc = append(newc, &TurtleCommand{
				Type:   c.Type,
				FValue: c.FValue,
				IValue: c.IValue,
				VValue: c.VValue,
			})
		}
	}
	t.TurtleDefinition = newc
	//log2.Printf("defined turtle to contain %d steps", len(t.TurtleDefinition))
}

func (t *Turtle) SetTag(tag string) {
	t.tag = tag
}

func (t *Turtle) DeleteTag(tag string) {
	t.tag = ""
	newc := make([]*TurtleCommand, 0, len(t.Track))
	for _, c := range t.Track {
		if c.Tag != tag {
			newc = append(newc, c)
		}
	}
	t.Track = newc
}

func (t *Turtle) SetHideTag(tag string, hidden bool) {
	for _, c := range t.Track {
		if c.Tag != tag {
			continue
		}
		c.Hidden = hidden
	}
}

func (t *Turtle) SetBounds(x1, y1, x2, y2 float64) {
	t.bx1, t.by1 = x1, y1
	t.bx2, t.by2 = x2, y2
}

func (t *Turtle) Clamp() {
	for t.Heading >= 360 {
		t.Heading -= 360
	}
	for t.Heading < 0 {
		t.Heading += 360
	}
	for t.Pitch >= 360 {
		t.Pitch -= 360
	}
	for t.Pitch < 0 {
		t.Pitch += 360
	}
	for t.Roll >= 360 {
		t.Roll -= 360
	}
	for t.Roll < 0 {
		t.Roll += 360
	}
}

// Turn turtle left or right by an angle
func (t *Turtle) RotateLR(angle float64) {

	t.RotateAxis(angle, t.UpDir)
	t.Heading += angle
	t.Clamp()

}

func (t *Turtle) RotateRoll(angle float64) {

	t.RotateAxis(angle, t.ViewDir)
	t.Roll += angle
	t.Clamp()

}

// Pitch turtle up or down by an angle (deg)
func (t *Turtle) RotateUD(angle float64) {

	t.RotateAxis(angle, t.RightDir)
	t.Pitch += angle
	t.Clamp()

}

func (t *Turtle) Undo() {
	t.RemoveCommands(1)
}

func (t *Turtle) DoHome() {
	t.AddCommand(TURTLE_HOME, 0, 0)
}

func (t *Turtle) PosX(x float32) {
	t.AddCommand(TURTLE_POSX, x, 0)
}

func (t *Turtle) SetH(x float32) {
	t.AddCommand(TURTLE_HEADING, x, 0)
}

func (t *Turtle) PosY(y float32) {
	t.AddCommand(TURTLE_POSY, y, 0)
}

func (t *Turtle) PosZ(z float32) {
	t.AddCommand(TURTLE_POSZ, z, 0)
}

func (t *Turtle) Glyph(ch rune) {
	t.AddCommand(TURTLE_GLYPH, 0, int32(ch))
}

func (t *Turtle) GlyphSize(v float32) {
	t.AddCommand(TURTLE_GLYPH_SIZE, v, 0)
}

func (t *Turtle) GlyphStretch(v float32) {
	t.AddCommand(TURTLE_GLYPH_STRETCH, v, 0)
}

func (t *Turtle) GlyphDepth(v float32) {
	t.AddCommand(TURTLE_GLYPH_DEPTH, v, 0)
}

func (t *Turtle) GlyphFace(v int) {
	t.AddCommand(TURTLE_GLYPH_FONT, 0, int32(v))
}

func (t *Turtle) GlyphFilled(v int) {
	t.AddCommand(TURTLE_GLYPH_FILLED, 0, int32(v))
}

func (t *Turtle) GlyphColor(v int) {
	t.AddCommand(TURTLE_GLYPH_COLOR, 0, int32(v))
}

func (t *Turtle) Arc(r float32, angle float32, solid bool) {
	if solid {
		t.AddCommandV(TURTLE_ARC, &mgl64.Vec3{float64(r), float64(angle), 0})
	} else {
		t.AddCommandV(TURTLE_LINEARC, &mgl64.Vec3{float64(r), float64(angle), 0})
	}
}

func (t *Turtle) Circle(r float32, solid bool) {
	if solid {
		t.AddCommand(TURTLE_CIRCLE, r, 0)
	} else {
		t.AddCommand(TURTLE_LINECIRCLE, r, 0)
	}
}

func (t *Turtle) Sphere(r float32, solid bool) {

	t.AddCommand(TURTLE_SPHERE, r, 0)

}

func (t *Turtle) GetPyramidPoints(size mgl64.Vec3) []mgl64.Vec3 {
	base, height := size[0], size[1]
	ctr := t.Position

	a := t.GetRelativePointCustom(ctr, mgl64.Vec3{-base / 2, 0, -base / 2})
	b := t.GetRelativePointCustom(ctr, mgl64.Vec3{-base / 2, 0, base / 2})
	c := t.GetRelativePointCustom(ctr, mgl64.Vec3{base / 2, 0, base / 2})
	d := t.GetRelativePointCustom(ctr, mgl64.Vec3{base / 2, 0, -base / 2})
	e := t.GetRelativePointCustom(ctr, mgl64.Vec3{0, height, 0})

	return []mgl64.Vec3{
		a, b, c, d, e,
	}
}

func (t *Turtle) GetCubePoints(size mgl64.Vec3) []mgl64.Vec3 {
	w, h, d := size[0], size[1], size[2]
	basefl := t.Position
	baserl := t.GetRelativePointCustom(basefl, mgl64.Vec3{0, 0, d})
	baserr := t.GetRelativePointCustom(baserl, mgl64.Vec3{w, 0, 0})
	basefr := t.GetRelativePointCustom(basefl, mgl64.Vec3{w, 0, 0})
	topfl := t.GetRelativePointCustom(basefl, mgl64.Vec3{0, h, 0})
	toprl := t.GetRelativePointCustom(baserl, mgl64.Vec3{0, h, 0})
	topfr := t.GetRelativePointCustom(basefr, mgl64.Vec3{0, h, 0})
	toprr := t.GetRelativePointCustom(baserr, mgl64.Vec3{0, h, 0})
	return []mgl64.Vec3{
		basefl, baserl, baserr, basefr,
		topfl, toprl, toprr, topfr,
	}
}

func (t *Turtle) Cube(x, y, z float32, solid bool) {
	if solid {
		t.AddCommandV(TURTLE_CUBE, &mgl64.Vec3{float64(x), float64(y), float64(z)})
	} else {
		t.AddCommandV(TURTLE_LINECUBE, &mgl64.Vec3{float64(x), float64(y), float64(z)})
	}
}

func (t *Turtle) Pyramid(x, y float32, solid bool) {
	if solid {
		t.AddCommandV(TURTLE_PYRAMID, &mgl64.Vec3{float64(x), float64(y), 0})
	} else {
		t.AddCommandV(TURTLE_LINEPYRAMID, &mgl64.Vec3{float64(x), float64(y), 0})
	}
}

func (t *Turtle) Quad(x, y float32, solid bool) {
	if solid {
		t.AddCommandV(TURTLE_QUAD, &mgl64.Vec3{float64(x), float64(y), 0})
	} else {
		t.AddCommandV(TURTLE_LINEQUAD, &mgl64.Vec3{float64(x), float64(y), 0})
	}
}

func (t *Turtle) Home() {
	x1, y1, z1 := t.Position[0], t.Position[1], t.Position[2]
	x2, y2, z2 := t.HomePosition[0], t.HomePosition[1], t.HomePosition[2]

	v := NewVector(
		VT_LINE,
		uint64(t.PenColor),
		float32(x1), float32(y1), float32(z1),
		float32(x2), float32(y2), float32(z2),
	)

	if !t.PenUp {
		t.vb.Vectors = append(t.vb.Vectors, v)
	}

	t.Reset()
}

func (t *Turtle) SetHome(x, y, z float64) {
	t.AddCommandV(TURTLE_SETHOME, &mgl64.Vec3{x, y, z})
}

func (t *Turtle) SetPos(x, y, z float64) {

	// set vectors
	//t.PristineRotation(t.Heading, t.Pitch, t.Roll)

	if !t.PenUp {
		x1, y1, z1 := t.Position[0], t.Position[1], t.Position[2]
		x2, y2, z2 := x, y, z

		v := NewVector(
			VT_LINE,
			uint64(t.PenColor),
			float32(x1), float32(y1), float32(z1),
			float32(x2), float32(y2), float32(z2),
		)
		t.vb.Vectors = append(t.vb.Vectors, v)
	}

	t.Position[0], t.Position[1], t.Position[2] = x, y, z

	////fmt.Println(t.String())
}

func (t *Turtle) DrawTriangle(amount1 float64, amount2 float64, vt VectorType, c uint32) error {

	max := (0x10000-2)/8 - 10
	if len(t.vb.Vectors) >= max {
		return errors.New("Too many vectors")
	}

	x1, y1, z1 := t.Position[0], t.Position[1], t.Position[2] // current postion
	p2 := t.GetRelativePoint(mgl64.Vec3{0, amount2, 0})       // up distance
	p3 := t.GetRelativePoint(mgl64.Vec3{amount1, 0, 0})       // right distance

	v := NewVector(
		vt,
		uint64(c),
		float32(x1), float32(y1), float32(z1),
		float32(p2[0]), float32(p2[1]), float32(p2[2]),
		float32(p3[0]), float32(p3[1]), float32(p3[2]),
	)

	//if !t.PenUp {
	t.vb.Vectors = append(t.vb.Vectors, v)
	//}

	return nil
}

func (t *Turtle) DrawQuad(x, y float64, k VectorType, c uint32) error {

	max := (0x10000-2)/8 - 10
	if len(t.vb.Vectors) >= max {
		return errors.New("Too many vectors")
	}

	x1, y1, z1 := t.Position[0], t.Position[1], t.Position[2] // current postion
	p2 := t.GetRelativePoint(mgl64.Vec3{0, y, 0})             // up distance
	p3 := t.GetRelativePoint(mgl64.Vec3{x, 0, 0})             // right distance
	p4 := t.GetRelativePoint(mgl64.Vec3{x, y, 0})

	v := NewVector(
		k,
		uint64(c),
		float32(x1), float32(y1), float32(z1),
		float32(p2[0]), float32(p2[1]), float32(p2[2]),
		float32(p3[0]), float32(p3[1]), float32(p3[2]),
		float32(p4[0]), float32(p4[1]), float32(p4[2]),
	)

	//if !t.PenUp {
	t.vb.Vectors = append(t.vb.Vectors, v)
	//}

	return nil
}

func (t *Turtle) DrawSphere(r float64, k VectorType, c uint32) error {

	max := (0x10000-2)/8 - 10
	if len(t.vb.Vectors) >= max {
		return errors.New("Too many vectors")
	}

	x1, y1, z1 := t.Position[0], t.Position[1], t.Position[2] // current postion

	v := NewVector(
		k,
		uint64(c),
		float32(x1), float32(y1), float32(z1),
		float32(r), float32(r), float32(r),
	)

	t.vb.Vectors = append(t.vb.Vectors, v)

	return nil
}

func (t *Turtle) DrawPoly(size float64, sides float64, k VectorType, c uint32) error {

	max := (0x10000-2)/8 - 10
	if len(t.vb.Vectors) >= max {
		return errors.New("Too many vectors")
	}

	x1, y1, z1 := t.Position[0], t.Position[1], t.Position[2] // current postion

	v := NewVector(
		k,
		uint64(c),
		float32(x1), float32(y1), float32(z1),
		float32(t.ViewDir[0]), float32(t.ViewDir[1]), float32(t.ViewDir[2]),
		float32(t.UpDir[0]), float32(t.UpDir[1]), float32(t.UpDir[2]),
		float32(t.RightDir[0]), float32(t.RightDir[1]), float32(t.RightDir[2]),
		float32(size), float32(sides), float32(0),
	)

	t.vb.Vectors = append(t.vb.Vectors, v)

	return nil
}

func (t *Turtle) DrawCircle(r float64, k VectorType, c uint32) error {

	max := (0x10000-2)/8 - 10
	if len(t.vb.Vectors) >= max {
		return errors.New("Too many vectors")
	}

	x1, y1, z1 := t.Position[0], t.Position[1], t.Position[2] // current postion

	v := NewVector(
		k,
		uint64(c),
		float32(x1), float32(y1), float32(z1),
		float32(t.ViewDir[0]), float32(t.ViewDir[1]), float32(t.ViewDir[2]),
		float32(t.UpDir[0]), float32(t.UpDir[1]), float32(t.UpDir[2]),
		float32(t.RightDir[0]), float32(t.RightDir[1]), float32(t.RightDir[2]),
		float32(r), float32(r), float32(r),
	)

	t.vb.Vectors = append(t.vb.Vectors, v)

	return nil
}

func (t *Turtle) DrawArc(r, a float64, k VectorType, c uint32) error {

	max := (0x10000-2)/8 - 10
	if len(t.vb.Vectors) >= max {
		return errors.New("Too many vectors")
	}

	x1, y1, z1 := t.Position[0], t.Position[1], t.Position[2] // current postion

	v := NewVector(
		k,
		uint64(c),
		float32(x1), float32(y1), float32(z1),
		float32(t.ViewDir[0]), float32(t.ViewDir[1]), float32(t.ViewDir[2]),
		float32(t.UpDir[0]), float32(t.UpDir[1]), float32(t.UpDir[2]),
		float32(t.RightDir[0]), float32(t.RightDir[1]), float32(t.RightDir[2]),
		float32(r), float32(a), float32(0),
	)

	t.vb.Vectors = append(t.vb.Vectors, v)

	return nil
}

func (t *Turtle) GetQuat() (float32, float32, float32, float32) {

	// {cos a/2, (sin a/2) n_x, (sin a/2) n_y, (sin a/2) n_z}

	a := mgl64.DegToRad(t.Roll)
	w := math.Cos(a / 2)
	x := math.Sin(a/2) * t.ViewDir[0]
	y := math.Sin(a/2) * t.ViewDir[1]
	z := math.Sin(a/2) * t.ViewDir[2]

	return float32(x), float32(y), float32(z), float32(w)
}

func (t *Turtle) DrawPyramid(base, height float64, k VectorType, c uint32) error {

	max := (0x10000-2)/8 - 10
	if len(t.vb.Vectors) >= max {
		return errors.New("Too many vectors")
	}

	vts := t.GetPyramidPoints(mgl64.Vec3{base, height, 0})
	out := make([]float32, len(vts)*3)
	for i, v := range vts {
		out[i*3+0] = float32(v[0])
		out[i*3+1] = float32(v[1])
		out[i*3+2] = float32(v[2])
	}

	v := NewVector(
		k,
		uint64(c),
		out...,
	)

	//if !t.PenUp {
	t.vb.Vectors = append(t.vb.Vectors, v)
	//}

	return nil
}

func (t *Turtle) GetFont() (*font.DecalFont, error) {
	fontid := t.FontFace

	if f, ok := t.FontData[fontid]; ok {
		return f, nil
	}

	index := t.vb.BaseAddress / memory.OCTALYZER_INTERPRETER_SIZE
	if fontid >= 0 && fontid < len(settings.AuxFonts[index]) {
		fontName := settings.AuxFonts[index][fontid]
		f, err := font.LoadFromFile(fontName)
		if err != nil {
			return nil, err
		}
		t.FontData[fontid] = f
		return f, nil
	}

	return nil, errors.New("NO SUCH FONT")
}

func (t *Turtle) DrawGlyph(fontsize float64, ch rune, c uint32) error {
	max := (0x10000-2)/8 - 10
	if len(t.vb.Vectors) >= max {
		return errors.New("Too many vectors")
	}

	f, err := t.GetFont()
	if err != nil {
		return err
	}

	gd, ok := f.GlyphsN[ch]
	if !ok {
		return nil
	}

	vt := VT_LINECUBE
	if t.FontFilled {
		vt = VT_CUBE
	}

	sizeh := fontsize / float64(f.TextHeight)
	sizew := sizeh * t.FontStretch
	depth := t.FontDepth

	var p = t.Position
	var offset mgl64.Vec3
	var e error
	for _, pt := range gd {
		// each is a point
		offset = mgl64.Vec3{float64(pt.X) * sizew, float64(f.TextHeight-pt.Y-1) * sizeh, 0}
		t.Position = t.GetRelativePointCustom(p, offset)
		e = t.DrawCube(sizew, sizeh, depth, vt, c)
		if e != nil {
			t.Position = p
			return e
		}
	}
	t.Position = p
	ops := t.PenUp
	t.PenUp = true
	t.Move(sizew*float64(f.TextWidth), 0)
	t.PenUp = ops

	return nil
}

func (t *Turtle) DrawCube(x, y, z float64, k VectorType, c uint32) error {

	max := (0x10000-2)/8 - 10
	if len(t.vb.Vectors) >= max {
		return errors.New("Too many vectors")
	}

	vts := t.GetCubePoints(mgl64.Vec3{x, y, z})
	out := make([]float32, len(vts)*3)
	for i, v := range vts {
		out[i*3+0] = float32(v[0])
		out[i*3+1] = float32(v[1])
		out[i*3+2] = float32(v[2])
	}

	v := NewVector(
		k,
		uint64(c),
		out...,
	)

	//if !t.PenUp {
	t.vb.Vectors = append(t.vb.Vectors, v)
	//}

	return nil
}

func (t *Turtle) Move(amount float64, axis int) error {

	max := (0x10000-2)/8 - 6
	if len(t.vb.Vectors) >= max {
		return errors.New("Too many vectors")
	}

	// set vectors
	//t.PristineRotation(t.Heading, t.Pitch, t.Roll)
	mv := mgl64.Vec3{0, amount, 0}
	if axis == 2 {
		mv = mgl64.Vec3{0, 0, amount}
	} else if axis == 0 {
		mv = mgl64.Vec3{amount, 0, 0}
	}

	if t.BoundsMode == WRAP {

		vlist := t.GetRelativePointWithWrap(mv)
		l := vlist[len(vlist)-1].Points()
		t.Position = mgl64.Vec3{
			float64(l[len(l)-1][0]),
			float64(l[len(l)-1][1]),
			float64(l[len(l)-1][2]),
		}

		if !t.PenUp {
			t.vb.Vectors = append(t.vb.Vectors, vlist...)
		}

	} else {

		x1, y1, z1 := t.Position[0], t.Position[1], t.Position[2]

		//t.Position[0], t.Position[1], t.Position[2] = t.GetMoveTarget(amount)
		t.Position = t.GetRelativePoint(mv)

		x2, y2, z2 := t.Position[0], t.Position[1], t.Position[2]

		v := NewVector(
			VT_LINE,
			uint64(t.PenColor),
			float32(x1), float32(y1), float32(z1),
			float32(x2), float32(y2), float32(z2),
		)

		if t.BoundsMode == FENCE {
			if x2 < t.bx1 || x2 > t.bx2 || y2 < t.by1 || y2 > t.by2 {
				return errors.New("TURTLE OUT OF BOUNDS")
			}
		}

		if !t.PenUp {
			t.vb.Vectors = append(t.vb.Vectors, v)
		}

	}

	////fmt.Println(t.String())
	return nil
}

func (t *Turtle) RemoveCommands(n int) {
	for len(t.Track) > 0 && n > 0 {
		t.Track = t.Track[0 : len(t.Track)-1]
		n--
	}
}

// Add command adds a turtle command
func (t *Turtle) AddCommand(ct TurtleCommandType, f float32, i int32) {
	tc := &TurtleCommand{Type: ct, FValue: f, IValue: i, Tag: t.tag}
	//log.Printf("add command: +%v", *tc)
	t.Track = append(t.Track, tc)
}

func (t *Turtle) AddCommandV(ct TurtleCommandType, v *mgl64.Vec3) {
	tc := &TurtleCommand{Type: ct, VValue: v, Tag: t.tag}
	//log.Printf("add command: +%v", *tc)
	t.Track = append(t.Track, tc)
}

func (t *Turtle) Clean() {
	t.Reset()
	t.Track = make([]*TurtleCommand, 0)
}

func (t *Turtle) ClearScreen() {
	t.Reset()
	t.Track = make([]*TurtleCommand, 0)
}

func (t *Turtle) Poly(size float32, sides float32, solid bool) {
	if solid {
		t.AddCommandV(TURTLE_POLY, &mgl64.Vec3{float64(size), float64(sides), 0})
	} else {
		t.AddCommandV(TURTLE_LINEPOLY, &mgl64.Vec3{float64(size), float64(sides), 0})
	}
}

func (t *Turtle) Raise(amount float32) {
	t.AddCommand(TURTLE_RAISE, amount, 0)
}

func (t *Turtle) Lower(amount float32) {
	t.AddCommand(TURTLE_LOWER, amount, 0)
}

func (t *Turtle) ShuffleLeft(amount float32) {
	t.AddCommand(TURTLE_SHUFFLE_LEFT, amount, 0)
}

func (t *Turtle) ShuffleRight(amount float32) {
	t.AddCommand(TURTLE_SHUFFLE_RIGHT, amount, 0)
}

func (t *Turtle) Forward(amount float32) {
	t.AddCommand(TURTLE_FD, amount, 0)
}

func (t *Turtle) Triangle(amount1, amount2 float32, solid bool) {
	if solid {
		t.AddCommandV(TURTLE_TRIANGLE, &mgl64.Vec3{float64(amount1), float64(amount2), 0})
	} else {
		t.AddCommandV(TURTLE_LINETRIANGLE, &mgl64.Vec3{float64(amount1), float64(amount2), 0})
	}
}

func (t *Turtle) Backward(amount float32) {
	t.AddCommand(TURTLE_BK, amount, 0)
}

func (t *Turtle) Left(amount float32) {
	t.AddCommand(TURTLE_LT, amount, 0)
}

func (t *Turtle) Right(amount float32) {
	t.AddCommand(TURTLE_RT, amount, 0)
}

func (t *Turtle) RollLeft(amount float32) {
	t.AddCommand(TURTLE_RL, amount, 0)
}

func (t *Turtle) RollRight(amount float32) {
	t.AddCommand(TURTLE_RR, amount, 0)
}

func (t *Turtle) Up(amount float32) {
	t.AddCommand(TURTLE_UP, amount, 0)
}

func (t *Turtle) Down(amount float32) {
	t.AddCommand(TURTLE_DN, amount, 0)
}

func (t *Turtle) SetPenColor(cval int32) {
	t.AddCommand(TURTLE_PENCOL, 0, cval)
}

func (t *Turtle) SetPenOpacity(cval int32) {
	t.AddCommand(TURTLE_PENOPACITY, 0, cval)
}

func (t *Turtle) SetFillColor(cval int32) {
	t.AddCommand(TURTLE_FILLCOL, 0, cval)
}

func (t *Turtle) SetFillOpacity(cval int32) {
	t.AddCommand(TURTLE_FILLOPACITY, 0, cval)
}

func (t *Turtle) SetShow() {
	t.AddCommand(TURTLE_SHOW, 0, 0)
}

func (t *Turtle) SetHide() {
	t.AddCommand(TURTLE_HIDE, 0, 0)
}

func (t *Turtle) SetPenDown() {
	t.AddCommand(TURTLE_PENDOWN, 0, 0)
}

func (t *Turtle) SetPenUp() {
	t.AddCommand(TURTLE_PENUP, 0, 0)
}

func (t *Turtle) RenderCommands(commands []*TurtleCommand) error {

	for i, tc := range commands {
		if tc.Hidden {
			continue
		}
		switch tc.Type {
		case TURTLE_SETHOME:
			t.HomePosition[0] = tc.VValue[0]
			t.HomePosition[1] = tc.VValue[1]
			t.HomePosition[2] = tc.VValue[2]
		case TURTLE_ROLL:
			t.SetPitchHeadingRoll(t.Pitch, t.Heading, float64(tc.FValue))
		case TURTLE_PITCH:
			t.SetPitchHeadingRoll(float64(tc.FValue), t.Heading, t.Roll)
		case TURTLE_SPHR:
			t.SetPitchHeadingRoll(tc.VValue[0], tc.VValue[1], tc.VValue[2])
		case TURTLE_HEADING:
			t.SetPitchHeadingRoll(t.Pitch, float64(tc.FValue), t.Roll)
		case TURTLE_CLEAR:
			t.vb.Vectors = make(VectorList, 0)
			t.Track = make([]*TurtleCommand, 0)
		case TURTLE_SPOS:
			t.SetPos(tc.VValue[0], tc.VValue[1], tc.VValue[2])
		case TURTLE_POSX:
			t.SetPos(float64(tc.FValue), t.Position[1], t.Position[2])
		case TURTLE_POSY:
			t.SetPos(t.Position[0], float64(tc.FValue), t.Position[2])
		case TURTLE_POSZ:
			t.SetPos(t.Position[0], t.Position[1], float64(tc.FValue))
		case TURTLE_HOME:
			t.Home()
		case TURTLE_FILLCOL:
			t.FillColor = (t.FillColor & 0xff000000) | uint32(tc.IValue&15)
		case TURTLE_FILLOPACITY:
			t.FillColor = (t.FillColor & 15) | (uint32(255-tc.IValue) << 24)
		case TURTLE_PENCOL:
			t.PenColor = (t.PenColor & 0xff000000) | uint32(tc.IValue&15)
		case TURTLE_PENOPACITY:
			t.PenColor = (t.PenColor & 15) | (uint32(255-tc.IValue) << 24)
		case TURTLE_PENUP:
			t.PenUp = true
		case TURTLE_PENDOWN:
			t.PenUp = false
		case TURTLE_FD:
			e := t.Move(float64(tc.FValue), 1)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
			t.UpdateTracking()
		case TURTLE_BK:
			e := t.Move(float64(-tc.FValue), 1)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
			t.UpdateTracking()
		case TURTLE_SHUFFLE_LEFT:
			e := t.Move(float64(-tc.FValue), 0)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
			t.UpdateTracking()
		case TURTLE_SHUFFLE_RIGHT:
			e := t.Move(float64(tc.FValue), 0)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
			t.UpdateTracking()
		case TURTLE_RAISE:
			e := t.Move(float64(-tc.FValue), 2)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
			t.UpdateTracking()
		case TURTLE_LOWER:
			e := t.Move(float64(tc.FValue), 2)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
			t.UpdateTracking()
		case TURTLE_LT:
			t.RotateLR(float64(-tc.FValue))
			t.UpdateTracking()
		case TURTLE_RT:
			t.RotateLR(float64(tc.FValue))
			t.UpdateTracking()
		case TURTLE_RL:
			t.RotateRoll(float64(tc.FValue))
			t.UpdateTracking()
		case TURTLE_RR:
			t.RotateRoll(float64(-tc.FValue))
			t.UpdateTracking()
		case TURTLE_UP:
			t.RotateUD(float64(tc.FValue))
			t.UpdateTracking()
		case TURTLE_DN:
			t.RotateUD(float64(-tc.FValue))
			t.UpdateTracking()
		case TURTLE_SHOW:
			t.Hide = false
		case TURTLE_HIDE:
			t.Hide = true
		case TURTLE_QUAD:
			e := t.DrawQuad(float64(tc.VValue[0]), float64(tc.VValue[1]), VT_QUAD, t.FillColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_LINEQUAD:
			e := t.DrawQuad(float64(tc.VValue[0]), float64(tc.VValue[1]), VT_LINEQUAD, t.PenColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_CUBE:
			e := t.DrawCube(float64(tc.VValue[0]), float64(tc.VValue[1]), float64(tc.VValue[2]), VT_CUBE, t.FillColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_LINECUBE:
			e := t.DrawCube(float64(tc.VValue[0]), float64(tc.VValue[1]), float64(tc.VValue[2]), VT_LINECUBE, t.PenColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_LINETRIANGLE:
			e := t.DrawTriangle(tc.VValue[0], tc.VValue[1], VT_TRIANGLE_LINE, t.PenColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_TRIANGLE:
			var e error
			if tc.VValue != nil {
				e = t.DrawTriangle(tc.VValue[0], tc.VValue[1], VT_TRIANGLE, t.FillColor)
			} else {
				e = t.DrawTriangle(float64(tc.FValue), float64(tc.FValue), VT_TRIANGLE, t.FillColor)
			}
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_POLY:
			e := t.DrawPoly(float64(tc.VValue[0]), float64(tc.VValue[1]), VT_POLY, t.FillColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_LINEPOLY:
			e := t.DrawPoly(float64(tc.VValue[0]), float64(tc.VValue[1]), VT_LINEPOLY, t.PenColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_PYRAMID:
			e := t.DrawPyramid(float64(tc.VValue[0]), float64(tc.VValue[1]), VT_PYRAMID, t.FillColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_CIRCLE:
			e := t.DrawCircle(float64(tc.FValue), VT_CIRCLE, t.FillColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_LINECIRCLE:
			e := t.DrawCircle(float64(tc.FValue), VT_LINECIRCLE, t.PenColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_ARC:
			e := t.DrawArc(tc.VValue[0], tc.VValue[1], VT_ARC, t.FillColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_LINEARC:
			e := t.DrawArc(tc.VValue[0], tc.VValue[1], VT_LINEARC, t.PenColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_SPHERE:
			e := t.DrawSphere(float64(tc.FValue), VT_SPHERE, t.FillColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_GLYPH:
			e := t.DrawGlyph(t.FontSize, rune(tc.IValue), t.FillColor)
			if e != nil {
				fmt.Println(e)
				t.Track = t.Track[0:i]
				t.DrawTurtleVectors(7)
				return e
			}
		case TURTLE_GLYPH_DEPTH:
			t.FontDepth = float64(tc.FValue)
		case TURTLE_GLYPH_SIZE:
			t.FontSize = float64(tc.FValue)
		case TURTLE_GLYPH_STRETCH:
			t.FontStretch = float64(tc.FValue)
		case TURTLE_GLYPH_FONT:
			t.FontFace = int(tc.IValue)
		case TURTLE_GLYPH_FILLED:
			t.FontFilled = int(tc.IValue) != 0
		}
	}

	return nil
}

func (t *Turtle) render() error {

	// INIT POSITION
	t.HomePosition = mgl64.Vec3{0, 0, 0}
	t.Reset()
	// Render out to the vector buffer attached

	//t.DumpJSON()
	var err error
	err = t.RenderCommands(t.Track)
	if err != nil {
		return err
	}
	if !t.Hide {
		//log2.Printf("drawing %d steps", len(t.TurtleDefinition))
		t.FillColor &= 15
		t.PenColor &= 15
		err = t.RenderCommands(t.TurtleDefinition)
		if err != nil {
			return err
		}
	}

	// Finally, do the turtle
	//t.DrawTurtleVectors(7)

	// Update memory image with data
	//t.vb.WriteToMemory()

	return nil
}

func (t *Turtle) DumpJSON() {
	//	j, _ := json.Marshal(t.Track)
	//fmt2.Println(string(j))
}

func (t *Turtle) DrawTurtleVectors(size float64) {

	if t.Hide {
		return
	}

	p1 := t.GetRelativePoint(mgl64.Vec3{-size / 2, 0, 0}) // left bottom
	p2 := t.GetRelativePoint(mgl64.Vec3{size / 2, 0, 0})  // right bottom
	p3 := t.GetRelativePoint(mgl64.Vec3{0, size, 0})      // top center

	p4 := t.GetRelativePoint(mgl64.Vec3{-size / 4, 0, 0}) // left bottom
	p5 := t.GetRelativePoint(mgl64.Vec3{size / 4, 0, 0})  // right bottom
	p6 := t.GetRelativePoint(mgl64.Vec3{0, size / 2, 0})  // top center

	v1 := NewVector(
		VT_LINE,
		15,
		float32(p1[0]), float32(p1[1]), float32(p1[2]),
		float32(p2[0]), float32(p2[1]), float32(p2[2]),
	)

	v2 := NewVector(
		VT_LINE,
		15,
		float32(p2[0]), float32(p2[1]), float32(p2[2]),
		float32(p3[0]), float32(p3[1]), float32(p3[2]),
	)

	v3 := NewVector(
		VT_LINE,
		15,
		float32(p3[0]), float32(p3[1]), float32(p3[2]),
		float32(p1[0]), float32(p1[1]), float32(p1[2]),
	)

	v4 := NewVector(
		VT_LINE,
		14,
		float32(p4[0]), float32(p4[1]), float32(p4[2]),
		float32(p5[0]), float32(p5[1]), float32(p5[2]),
	)

	v5 := NewVector(
		VT_LINE,
		14,
		float32(p5[0]), float32(p5[1]), float32(p5[2]),
		float32(p6[0]), float32(p6[1]), float32(p6[2]),
	)

	v6 := NewVector(
		VT_LINE,
		14,
		float32(p6[0]), float32(p6[1]), float32(p6[2]),
		float32(p4[0]), float32(p4[1]), float32(p4[2]),
	)

	t.vb.Vectors = append(t.vb.Vectors, v1, v2, v3, v4, v5, v6)

	// pack orientation of turtle
	// t.vb.Vectors = append(t.vb.Vectors,
	// 	NewVector(
	// 		VT_TURTLE,
	// 		15,
	// 		float32(t.Position[0]),
	// 		float32(t.Position[1]),
	// 		float32(t.Position[2]),
	// 		float32(t.Heading),
	// 		float32(t.Pitch),
	// 		float32(t.Roll),
	// 	),
	// )

}

func (t *Turtle) String() string {
	return fmt.Sprintf(
		"Pos: %.1f,%.1f,%.1f, facing %.1f,%.1f,%.1f, right is %.1f,%.1f,%.1f, up is %.1f,%.1f,%.1f\n",
		t.Position[0], t.Position[1], t.Position[2],
		t.ViewDir[0], t.ViewDir[1], t.ViewDir[2],
		t.RightDir[0], t.RightDir[1], t.RightDir[2],
		t.UpDir[0], t.UpDir[1], t.UpDir[2],
	)
}

/*
   Notes:
   (1)   Origin point is current turtle position in X,Y,Z space.
   (2)   Pitch plane
*/

func (t *Turtle) GetRelativePointCustom(origin, relative mgl64.Vec3) mgl64.Vec3 {
	viewdir, updir, rightdir := t.ViewDir, t.UpDir, t.RightDir

	result := origin

	// X
	result[0] = result[0] + relative[0]*rightdir[0]
	result[1] = result[1] + relative[0]*rightdir[1]
	result[2] = result[2] + relative[0]*rightdir[2]

	// Y
	result[0] = result[0] + relative[1]*viewdir[0]
	result[1] = result[1] + relative[1]*viewdir[1]
	result[2] = result[2] + relative[1]*viewdir[2]

	// Z
	result[0] = result[0] + relative[2]*updir[0]
	result[1] = result[1] + relative[2]*updir[1]
	result[2] = result[2] + relative[2]*updir[2]

	return result

}

func (t *Turtle) GetRelativePoint(relative mgl64.Vec3) mgl64.Vec3 {
	return t.GetRelativePointCustom(t.Position, relative)
}

func (t *Turtle) GetRelativePointWithWrap(r mgl64.Vec3) []*Vector {
	viewdir, updir, rightdir := t.ViewDir, t.UpDir, t.RightDir

	var result mgl64.Vec3

	out := make([]*Vector, 0)
	remaining := r.Len()
	relative := r.Normalize()

	origin := t.Position

	v := NewVector(
		VT_LINE,
		uint64(t.PenColor),
		float32(origin[0]), float32(origin[1]), float32(origin[2]),
		float32(origin[0]), float32(origin[1]), float32(origin[2]),
	)

	for remaining > 0 {

		result = origin

		// X
		result[0] = result[0] + relative[0]*rightdir[0]
		result[1] = result[1] + relative[0]*rightdir[1]
		result[2] = result[2] + relative[0]*rightdir[2]

		// Y
		result[0] = result[0] + relative[1]*viewdir[0]
		result[1] = result[1] + relative[1]*viewdir[1]
		result[2] = result[2] + relative[1]*viewdir[2]

		// Z
		result[0] = result[0] + relative[2]*updir[0]
		result[1] = result[1] + relative[2]*updir[1]
		result[2] = result[2] + relative[2]*updir[2]

		var oob bool

		if result[0] > t.bx2 {
			result[0] = t.bx1
			oob = true
		}
		if result[0] < t.bx1 {
			result[0] = t.bx2
			oob = true
		}
		if result[1] > t.by2 {
			result[1] = t.by1
			oob = true
		}
		if result[1] < t.by1 {
			result[1] = t.by2
			oob = true
		}

		if oob {
			// out of bounds...
			v.X[1] = float32(origin[0])
			v.Y[1] = float32(origin[1])
			v.Z[1] = float32(origin[2])
			out = append(out, v)
			// reset start point to corrected point
			v = NewVector(
				VT_LINE,
				uint64(t.PenColor),
				float32(result[0]), float32(result[1]), float32(result[2]),
				float32(result[0]), float32(result[1]), float32(result[2]),
			)
		}

		remaining--
		origin = result // start from here next time...
	}

	// final vector
	v.X[1] = float32(result[0])
	v.Y[1] = float32(result[1])
	v.Z[1] = float32(result[2])
	out = append(out, v)

	return out

}

func (t *Turtle) SetHeading(a float32) {
	//t.AddCommand(TURTLE_PENCOL, 0, cval)
	t.ViewDir = mgl64.Vec3{0, 1, 0}
	t.UpDir = mgl64.Vec3{0, 0, -1}
	t.RightDir = mgl64.Vec3{1, 0, 0}
	t.Heading = 0
	t.Pitch = 0
	t.Roll = 0
	t.RotateLR(float64(a))
}

func (t *Turtle) SetPitchHeadingRoll(p, h, r float64) {
	t.ViewDir = mgl64.Vec3{0, 1, 0}
	t.UpDir = mgl64.Vec3{0, 0, -1}
	t.RightDir = mgl64.Vec3{1, 0, 0}
	t.Heading = 0
	t.Pitch = 0
	t.Roll = 0
	t.RotateUD(float64(p))
	t.RotateLR(float64(h))
	t.RotateRoll(float64(r))
}

func (t *Turtle) SetRoll(a float32) {
	t.AddCommand(TURTLE_ROLL, a, 0)
}

func (t *Turtle) SetPitch(a float32) {
	t.AddCommand(TURTLE_PITCH, a, 0)
}

func (t *Turtle) SetBoundsMode(bm TurtleBoundsMode) {
	t.BoundsMode = bm
	t.vb.Render()
}

func (t *Turtle) UpdateTracking() {

	// update relative camera position...

}

// defs here
var turtleDef = `[{"commandType":20},{"commandType":13,"intVal":15},{"commandType":12,"floatVal":180},{"commandType":8,"floatVal":90},{"commandType":3},{"commandType":6,"floatVal":1.25},{"commandType":4},{"commandType":7,"floatVal":90},{"commandType":9,"floatVal":90},{"commandType":8,"floatVal":20},{"commandType":5,"floatVal":3.75},{"commandType":6,"floatVal":3.75},{"commandType":7,"floatVal":20},{"commandType":10,"floatVal":90},{"commandType":8,"floatVal":10.05},{"commandType":5,"floatVal":7.3},{"commandType":3},{"commandType":6,"floatVal":7.3},{"commandType":4},{"commandType":7,"floatVal":10.05},{"commandType":9,"floatVal":90},{"commandType":7,"floatVal":20},{"commandType":6,"floatVal":3.73},{"commandType":8,"floatVal":20},{"commandType":10,"floatVal":64},{"commandType":5,"floatVal":8},{"commandType":9,"floatVal":128},{"commandType":5,"floatVal":8},{"commandType":3},{"commandType":6,"floatVal":8},{"commandType":10,"floatVal":128},{"commandType":6,"floatVal":8},{"commandType":9,"floatVal":64},{"commandType":7,"floatVal":20},{"commandType":5,"floatVal":3.73},{"commandType":8,"floatVal":20},{"commandType":10,"floatVal":90},{"commandType":3},{"commandType":8,"floatVal":90},{"commandType":5,"floatVal":2.55},{"commandType":7,"floatVal":90},{"commandType":12,"floatVal":180},{"commandType":4},{"commandType":13,"intVal":15},{"commandType":9,"floatVal":90},{"commandType":8,"floatVal":20},{"commandType":5,"floatVal":3.75},{"commandType":3},{"commandType":6,"floatVal":3.75},{"commandType":4},{"commandType":7,"floatVal":20},{"commandType":10,"floatVal":90},{"commandType":8,"floatVal":10.05},{"commandType":5,"floatVal":7.3},{"commandType":3},{"commandType":6,"floatVal":7.3},{"commandType":4},{"commandType":7,"floatVal":10.05},{"commandType":9,"floatVal":90},{"commandType":7,"floatVal":20},{"commandType":6,"floatVal":3.73},{"commandType":3},{"commandType":19},{"commandType":3},{"commandType":8,"floatVal":20},{"commandType":5,"floatVal":3.5},{"commandType":10,"floatVal":90},{"commandType":8,"floatVal":90},{"commandType":5,"floatVal":1.25},{"commandType":7,"floatVal":90},{"commandType":13,"intVal":3},{"commandType":12,"floatVal":180},{"commandType":3},{"commandType":8,"floatVal":90},{"commandType":5,"floatVal":1.275},{"commandType":8,"floatVal":90},{"commandType":5,"floatVal":0.12},{"commandType":7,"floatVal":90},{"commandType":10,"floatVal":90},{"commandType":12,"floatVal":5},{"commandType":7,"floatVal":5},{"commandType":14,"intVal":2},{"commandType":22,"floatVal":1.28},{"commandType":11,"floatVal":5},{"commandType":8,"floatVal":5},{"commandType":12,"floatVal":180},{"commandType":11,"floatVal":5},{"commandType":8,"floatVal":5},{"commandType":14,"intVal":7},{"commandType":22,"floatVal":1.28},{"commandType":12,"floatVal":5},{"commandType":7,"floatVal":5},{"commandType":12,"floatVal":180},{"commandType":10,"floatVal":180},{"commandType":12,"floatVal":5},{"commandType":7,"floatVal":5},{"commandType":14,"intVal":6},{"commandType":22,"floatVal":1.28},{"commandType":11,"floatVal":5},{"commandType":8,"floatVal":5},{"commandType":12,"floatVal":180},{"commandType":11,"floatVal":5},{"commandType":8,"floatVal":5},{"commandType":14,"intVal":10},{"commandType":22,"floatVal":1.28},{"commandType":12,"floatVal":5},{"commandType":7,"floatVal":5},{"commandType":12,"floatVal":180},{"commandType":10,"floatVal":90},{"commandType":7,"floatVal":90},{"commandType":5,"floatVal":0.12},{"commandType":10,"floatVal":180},{"commandType":7,"floatVal":180},{"commandType":8,"floatVal":0.2},{"commandType":5,"floatVal":3.5},{"commandType":4},{"commandType":5,"floatVal":3.67},{"commandType":3},{"commandType":6,"floatVal":7.17},{"commandType":4},{"commandType":8,"floatVal":90},{"commandType":20},{"commandType":5,"floatVal":0.125},{"commandType":7,"floatVal":90},{"commandType":12,"floatVal":2},{"commandType":14,"intVal":1},{"commandType":7,"floatVal":2},{"commandType":22,"floatVal":3.5},{"commandType":8,"floatVal":2},{"commandType":12,"floatVal":180},{"commandType":11,"floatVal":4},{"commandType":8,"floatVal":2},{"commandType":14,"intVal":9},{"commandType":22,"floatVal":3.5},{"commandType":7,"floatVal":2},{"commandType":12,"floatVal":2},{"commandType":8,"floatVal":90},{"commandType":5,"floatVal":0.25},{"commandType":7,"floatVal":90},{"commandType":12,"floatVal":2},{"commandType":14,"intVal":4},{"commandType":7,"floatVal":2},{"commandType":22,"floatVal":3.5},{"commandType":8,"floatVal":2},{"commandType":12,"floatVal":180},{"commandType":11,"floatVal":4},{"commandType":14,"intVal":14},{"commandType":8,"floatVal":2},{"commandType":22,"floatVal":3.5},{"commandType":7,"floatVal":2},{"commandType":12,"floatVal":2},{"commandType":8,"floatVal":90},{"commandType":5,"floatVal":0.125},{"commandType":6,"floatVal":1.25},{"commandType":5,"floatVal":2.5},{"commandType":3},{"commandType":6,"floatVal":1.25},{"commandType":7,"floatVal":90},{"commandType":12,"floatVal":180},{"commandType":4}]`

func angleBetweenVectors(a, b mgl64.Vec3) float64 {
	ct := a.Dot(b) / a.Len() / b.Len()
	return mgl64.RadToDeg(math.Acos(ct))
}
