package types

import (
	"errors"
	//"paleotronic.com/fmt"
)

type TextSize uint32

const (
	W_HALF_H_HALF     TextSize = 0
	W_HALF_H_NORMAL   TextSize = 1
	W_HALF_H_DOUBLE   TextSize = 2
	W_HALF_H_QUAD     TextSize = 3
	W_NORMAL_H_HALF   TextSize = 4
	W_NORMAL_H_NORMAL TextSize = 5
	W_NORMAL_H_DOUBLE TextSize = 6
	W_NORMAL_H_QUAD   TextSize = 7
	W_DOUBLE_H_HALF   TextSize = 8
	W_DOUBLE_H_NORMAL TextSize = 9
	W_DOUBLE_H_DOUBLE TextSize = 10
	W_DOUBLE_H_QUAD   TextSize = 11
	W_QUAD_H_HALF     TextSize = 12
	W_QUAD_H_NORMAL   TextSize = 13
	W_QUAD_H_DOUBLE   TextSize = 14
	W_QUAD_H_QUAD     TextSize = 15
)

type TextWindow struct {
	Top, Left, Bottom, Right int
}

func (this *TextWindow) MarshalBinary() ([]byte, error) {

	data := make([]byte, 4)

	data[0] = byte(this.Left)
	data[1] = byte(this.Right)
	data[2] = byte(this.Bottom)
	data[3] = byte(this.Top)

	return data, nil
}

func (this *TextWindow) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return errors.New("Not enough data")
	}

	this.Left = int(data[0])
	this.Right = int(data[1])
	this.Bottom = int(data[2])
	this.Top = int(data[3])

	return nil
}

type VideoMode struct {
	Palette       VideoPalette
	DefaultWindow TextWindow
	Height        int
	Rows          int
	Columns       int
	ActualRows    int
	Anchor        int
	Width         int
	MaxPages      int
}

func (vm VideoMode) GetDefaultTW() TextSize {
	if vm.Rows == 48 {
		return W_HALF_H_HALF
	} else {
		if vm.Columns == 80 {
			return W_HALF_H_NORMAL
		} else if vm.Columns == 40 {
			return W_NORMAL_H_NORMAL
		}
	}
	return W_NORMAL_H_NORMAL
}

func (vm VideoMode) HSkip() int {
	return 80 / vm.Columns
}

func (vm VideoMode) VSkip() int {
	return 48 / vm.Rows
}

func (vm VideoMode) HAdvance(tw TextSize) int {
	switch {
	case tw == W_QUAD_H_HALF || tw == W_QUAD_H_NORMAL || tw == W_QUAD_H_DOUBLE || tw == W_QUAD_H_QUAD:
		return 8
	case tw == W_DOUBLE_H_HALF || tw == W_DOUBLE_H_NORMAL || tw == W_DOUBLE_H_DOUBLE || tw == W_DOUBLE_H_QUAD:
		return 4
	case tw == W_NORMAL_H_HALF || tw == W_NORMAL_H_NORMAL || tw == W_NORMAL_H_DOUBLE || tw == W_NORMAL_H_QUAD:
		return 2
	case tw == W_HALF_H_HALF || tw == W_HALF_H_NORMAL || tw == W_HALF_H_DOUBLE || tw == W_HALF_H_QUAD:
		return 1
	}

	return 1
}

func (vm VideoMode) VAdvance(tw TextSize) int {

	switch {
	case tw == W_QUAD_H_QUAD || tw == W_DOUBLE_H_QUAD || tw == W_NORMAL_H_QUAD || tw == W_HALF_H_QUAD:
		return 8
	case tw == W_QUAD_H_DOUBLE || tw == W_DOUBLE_H_DOUBLE || tw == W_NORMAL_H_DOUBLE || tw == W_HALF_H_DOUBLE:
		return 4
	case tw == W_QUAD_H_NORMAL || tw == W_DOUBLE_H_NORMAL || tw == W_NORMAL_H_NORMAL || tw == W_HALF_H_NORMAL:
		return 2
	case tw == W_QUAD_H_HALF || tw == W_DOUBLE_H_HALF || tw == W_NORMAL_H_HALF || tw == W_HALF_H_HALF:
		return 1
	}
	return 1
}

// LinearOffsetXY returns the actual memory reference for a given co-ordinate
func (vm VideoMode) LinearOffsetXY(x, y int) int {
	hs := vm.HSkip()   		// 2
	vs := vm.VSkip()   		// 2
	xx := (hs * x) % 80
	yy := (vs * y) % 48
	////fmt.Printf( "LinearOffsetXY(%d, %d) = %d, %d\n", x, y, xx, yy )
	return xx + (80 * yy)
}

func (vm *VideoMode) PutXYMemory(txmb *TXMemoryBuffer, x, y int, ch uint, cx uint, tm TextSize) {
	os := vm.LinearOffsetXY(x, y)
	v := uint(ch | ((cx & 255) << 16) | (uint(tm) << 24))
	txmb.SetValue(os%txmb.Size(), v)
}

func (vm *VideoMode) GetXYMemory(txmb *TXMemoryBuffer, x, y int) uint {
	os := vm.LinearOffsetXY(x, y)
	return txmb.GetValue(os % txmb.Size())
}

func (this VideoMode) MarshalBinary() ([]byte, error) {
	data := make([]byte, 9)
	data[0] = MtVideoMode // placeholder id
	data[1] = byte(this.Height % 256)
	data[2] = byte(this.Height / 256)
	data[3] = byte(this.Width % 256)
	data[4] = byte(this.Width / 256)
	data[5] = byte(this.Columns)
	data[6] = byte(this.Rows)
	data[7] = byte(this.ActualRows)
	data[8] = byte(this.MaxPages)

	b, _ := this.DefaultWindow.MarshalBinary()

	data = append(data, b...)

	c, _ := this.Palette.MarshalBinary()

	data = append(data, c...)

	return data, nil
}

func (this *VideoMode) UnmarshalBinary(data []byte) error {

	if len(data) < 13 {
		return errors.New("Not enough data")
	}

	if data[0] != MtVideoMode {
		return errors.New("wrong type")
	}

	this.Height = int(data[1]) + 256*int(data[2])
	this.Width = int(data[3]) + 256*int(data[4])
	this.Columns = int(data[5])
	this.Rows = int(data[6])
	this.ActualRows = int(data[7])
	this.MaxPages = int(data[8])

	d := &TextWindow{}

	//////fmt.Println(data[9:13])

	_ = d.UnmarshalBinary(data[9:13])

	this.DefaultWindow = *d

	if len(data) > 13 {
		_ = this.Palette.UnmarshalBinary(data[13:])
	}

	return nil
}

type VideoModeList []VideoMode

func (this *VideoModeList) Add(vm VideoMode) {
	a := *this
	a = append(a, vm)
	*this = a
}

func (this *VideoMode) Equals(vm *VideoMode) bool {
	return (this.Rows == vm.Rows) && (this.Columns == vm.Columns) && (this.ActualRows == vm.ActualRows) && (this.Width == vm.Width) && (this.Height == vm.Height) && (len(this.Palette.Items) == len(vm.Palette.Items))
}

func NewTextWindow() *TextWindow {
	return &TextWindow{}
}

func NewVideoMode(w int, h int, r int, c int, ar int, anc int, cp VideoPalette, mp int, pixelaspect float64) *VideoMode {
	this := &VideoMode{}

	this.Width = w
	this.Height = h
	this.Rows = r
	this.Columns = c
	this.ActualRows = ar
	this.Anchor = anc
	this.Palette = cp
	this.MaxPages = mp

	// now stuff;
	this.DefaultWindow = *NewTextWindow()
	this.DefaultWindow.Left = 0
	this.DefaultWindow.Right = c - 1
	this.DefaultWindow.Bottom = r - 1
	if w < 50 {
		this.DefaultWindow.Top = r - ar
	} else {
		this.DefaultWindow.Top = 0
	}

	return this
}


