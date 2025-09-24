package vdu

import (
	"paleotronic.com/core/types"
)

const (
	TB_FIXED_WIDTH  = 80
	TB_FIXED_HEIGHT = 488
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

type TextBuffer struct {
	Data             *types.TXMemoryBuffer
	CursorX, CursorY int
	TextMode         TextSize
	FGColor, BGColor int
}

func NewTextBuffer() *TextBuffer {
	this := &TextBuffer{}
	this.Data = types.NewTXMemoryBuffer(TB_FIXED_WIDTH * TB_FIXED_HEIGHT)
	this.CursorX = 0
	this.CursorY = 0
	this.TextMode = W_NORMAL_H_NORMAL
	return this
}

func (this TextBuffer) HAdvance() int {
	tw := this.TextMode
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

func (this TextBuffer) VAdvance() int {
	tw := this.TextMode
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

func (this *TextBuffer) PutXY(x, y int, ch rune) {

	cx := (uint(this.FGColor&0xf) << 4) | uint(this.BGColor&0xf)

	os := x + (y * TB_FIXED_WIDTH)
	v := uint(uint(ch) | ((cx & 255) << 16) | (uint(this.TextMode) << 24))
	this.Data.SetValue(os%this.Data.Size(), v)
}
