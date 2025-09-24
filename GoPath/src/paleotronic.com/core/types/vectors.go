package types

import (
	"errors"

	"github.com/go-gl/mathgl/mgl32"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/fmt"
)

type VectorType uint64

const (
	VT_TURTLE VectorType = iota
	VT_LINE
	VT_CURVE
	VT_POINT
	VT_CUBE
	VT_TRIANGLE_LINE
	VT_TRIANGLE
	VT_QUAD
	VT_LINECUBE
	VT_LINEQUAD
	VT_CIRCLE
	VT_LINECIRCLE
	VT_SPHERE
	VT_POLY
	VT_LINEPOLY
	VT_ARC
	VT_LINEARC
	VT_PYRAMID
)

func (v VectorType) String() string {
	switch v {
	case VT_TURTLE:
		return "TURTLE"
	case VT_LINE:
		return "LINE"
	case VT_CURVE:
		return "CURVE"
	case VT_POINT:
		return "POINT"
	case VT_CUBE:
		return "CUBE"
	case VT_TRIANGLE_LINE:
		return "TRIANGLE LINE"
	case VT_TRIANGLE:
		return "TRIANGLE"
	case VT_QUAD:
		return "QUAD"
	}
	return "?"
}

type Vector struct {
	Type        VectorType
	VectorCount int
	RGBA        uint64
	X, Y, Z     []float32
}

func NewVector(t VectorType, rgba uint64, coordinates ...float32) *Vector {
	if len(coordinates)%3 != 0 {
		return nil
	}
	v := &Vector{
		Type:        t,
		VectorCount: len(coordinates) / 3,
		RGBA:        rgba,
		X:           make([]float32, len(coordinates)/3),
		Y:           make([]float32, len(coordinates)/3),
		Z:           make([]float32, len(coordinates)/3),
	}
	for i, vv := range coordinates {
		switch i % 3 {
		case 0:
			v.X[i/3] = vv
		case 1:
			v.Y[i/3] = vv
		case 2:
			v.Z[i/3] = vv
		}
	}
	return v
}

func (v *Vector) Points() []mgl32.Vec3 {
	vl := make([]mgl32.Vec3, len(v.X))
	for i, _ := range v.X {
		vl[i] = mgl32.Vec3{v.X[i], v.Y[i], v.Z[i]}
	}
	return vl
}

func (v *Vector) String() string {
	return fmt.Sprintf("VECTOR TYPE %s: %+v", v.Type, *v)
}

func (v *Vector) MarshalBinary() ([]uint64, error) {

	//fmt.Println(v.String())
	needed := len(v.X)/2 + 1

	data := make([]uint64, 2+3*needed)

	data[0] = uint64(v.Type) | (uint64(v.VectorCount) << 32)
	data[1] = v.RGBA

	for i, vv := range v.X {
		idx := 2 + 0*needed + (i / 2)
		val := data[idx]
		switch i % 2 {
		case 0:
			val = Float2uint(vv)
		case 1:
			val |= (Float2uint(vv) << 32)
		}
		data[idx] = val
	}
	for i, vv := range v.Y {
		idx := 2 + 1*needed + (i / 2)
		val := data[idx]
		switch i % 2 {
		case 0:
			val = Float2uint(vv)
		case 1:
			val |= (Float2uint(vv) << 32)
		}
		data[idx] = val
	}
	for i, vv := range v.Z {
		idx := 2 + 2*needed + (i / 2)
		val := data[idx]
		switch i % 2 {
		case 0:
			val = Float2uint(vv)
		case 1:
			val |= (Float2uint(vv) << 32)
		}
		data[idx] = val
	}

	return data, nil
}

func (v *Vector) UnmarshalBinary(data []uint64) error {

	if len(data) < 2 {
		return errors.New("not enough data")
	}

	v.Type = VectorType(data[0]) & 0xffffffff
	v.VectorCount = int(data[0] >> 32)
	v.RGBA = data[1]

	needed := v.VectorCount/2 + 1
	neededSize := 2 + 3*needed
	if len(data) < neededSize {
		return errors.New("not enough data")
	}

	v.X = make([]float32, v.VectorCount)
	v.Y = make([]float32, v.VectorCount)
	v.Z = make([]float32, v.VectorCount)

	for i, _ := range v.X {
		idx := 2 + 0*needed + (i / 2)
		switch i % 2 {
		case 0:
			v.X[i] = Uint2Float(data[idx] & 0xffffffff)
		case 1:
			v.X[i] = Uint2Float(data[idx] >> 32)
		}
	}

	for i, _ := range v.Y {
		idx := 2 + 1*needed + (i / 2)
		switch i % 2 {
		case 0:
			v.Y[i] = Uint2Float(data[idx] & 0xffffffff)
		case 1:
			v.Y[i] = Uint2Float(data[idx] >> 32)
		}
	}

	for i, _ := range v.Z {
		idx := 2 + 2*needed + (i / 2)
		switch i % 2 {
		case 0:
			v.Z[i] = Uint2Float(data[idx] & 0xffffffff)
		case 1:
			v.Z[i] = Uint2Float(data[idx] >> 32)
		}
	}

	//fmt.Println("UNMARSHAL:", v.String())

	return nil
}

type VectorList []*Vector

func (l *VectorList) MarshalBinary(max int) ([]uint64, error) {
	clip := len(*l)
	data := make([]uint64, 2)
	data[0] = uint64(clip)
	data[1] = 1
	zz := *l
	for z := 0; z < clip; z++ {
		v := *zz[z]
		d, _ := v.MarshalBinary()
		if len(data)+len(d) > max {
			data[0] = uint64(z)
			return data, errors.New("Out of vector memory")
		}
		data = append(data, d...)
	}

	return data, nil
}

func (l *VectorList) UnmarshalBinary(data []uint64) error {
	size := int(data[0])
	needed := 2 + size*8
	if len(data) < needed {
		return errors.New("not enough data")
	}
	offset := 2

	count := 0
	*l = make(VectorList, 0, size)
	for offset < len(data) && count < size {
		vectorcount := int(data[offset] >> 32)
		size := 2 + (vectorcount/2+1)*3
		if offset+size > len(data) {
			break
		}
		v := &Vector{}
		e := v.UnmarshalBinary(data[offset : offset+size])
		if e != nil {
			return e
		}
		*l = append(*l, v)
		count++
		offset += size // move to next vector
	}

	return nil
}

type VectorBuffer struct {
	//sync.Mutex
	Data        *memory.MemoryControlBlock
	Vectors     VectorList
	BaseAddress int
	Size        int
	turtle      map[int]*Turtle
	selected    int
	CubeMap     *CubeMap
}

func (vb *VectorBuffer) Lock() {
	if vb != nil {
		settings.VBLock[vb.BaseAddress/memory.OCTALYZER_INTERPRETER_SIZE].Lock()
	}
}

func (vb *VectorBuffer) Unlock() {
	if vb != nil {
		settings.VBLock[vb.BaseAddress/memory.OCTALYZER_INTERPRETER_SIZE].Unlock()
	}
}

func NewVectorBufferMapped(base int, bufferSize int, mr *memory.MappedRegion) *VectorBuffer {
	this := &VectorBuffer{
		BaseAddress: base,
		Size:        bufferSize,
		Data:        mr.Data,
		Vectors:     make(VectorList, 0),
	}
	this.CubeMap = NewCubeMap(80, 48, 48, this)
	this.turtle = map[int]*Turtle{
		1: NewTurtle(this),
	}
	this.selected = 1
	this.Turtle().SetPenColor(23)
	this.Render()
	return this
}

func (tb *VectorBuffer) GetTurtle(i int) *Turtle {
	t, ok := tb.turtle[i]
	if !ok {
		t = NewTurtle(tb)
		tb.turtle[i] = t
	}
	return t
}

func (tb *VectorBuffer) Turtle() *Turtle {

	_, ok := tb.turtle[tb.selected]
	if !ok {
		tb.turtle[tb.selected] = NewTurtle(tb)
	}

	return tb.turtle[tb.selected]
}

func (vb *VectorBuffer) ReadFromMemory() error {
	chunk := vb.Data.ReadSlice(0, vb.Data.Size)
	return vb.Vectors.UnmarshalBinary(chunk)
}

func (vb *VectorBuffer) WriteToMemory() error {
	chunk, e := vb.Vectors.MarshalBinary(vb.Size)
	if e != nil {
		return e
	}
	if len(chunk) > vb.Data.Size {
		return errors.New("Vector buffer overflow")
	}
	for i := len(chunk) - 1; i >= 0; i-- {
		vb.Data.Write(i, chunk[i])
	}

	return nil
}

func (vb *VectorBuffer) Render() error {
	vb.Lock()
	defer vb.Unlock()
	vb.Vectors = make(VectorList, 0)
	var err error
	for _, t := range vb.turtle {
		err = t.render()
		if err != nil {
			return err
		}
	}

	err = vb.WriteToMemory()
	//time.Sleep(20 * time.Millisecond)
	return err
}

func (vb *VectorBuffer) SelectTurtle(i int) {
	vb.selected = i
	vb.Render()
}

func (vb *VectorBuffer) DeleteTurtle(i int) {
	delete(vb.turtle, i)
	vb.Render()
}
