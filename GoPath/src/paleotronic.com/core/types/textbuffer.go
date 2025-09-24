package types

import (
	"errors"
	"math"
	"strings"
	"unicode/utf8"

	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/fmt"
)

/*
TextBuffer provides a memory store (backed onto TXMemoryBuffer) for manipulating
screen style data
*/
const (
	hBump           = 1024
	vBump           = 2048
	bufferWidth     = 80
	bufferHeight    = 48
	bufferSize      = 4096
	xPos            = 4094
	yPos            = 4095
	sX              = 4093
	sY              = 4092
	eX              = 4091
	eY              = 4090
	cVis            = 4089
	cSize           = 4088
	noNLOutOfBounds = true
	CanCopy         = 3067
	SelSX           = 3068
	SelSY           = 3069
	SelEX           = 3070
	SelEY           = 3071
)

type MemoryPublisher interface {
	SetMemory(addr int, value uint64)
	GetMemory(addr int) uint64
}

type TextBufferWindow struct {
	SX, SY, EX, EY int
}

type TextCursorPos struct {
	CX, CY int
}

type TextBuffer struct {
	SwitchInterleave bool
	BaseAddress      int
	Size             int
	Data             *memory.MemoryControlBlock
	Font             TextSize
	Attribute        VideoAttribute
	Linear           bool
	FGColor, BGColor uint64
	Shade            uint64
	CX, CY           int
	SX, SY, EX, EY   int // Drawable rectangle
	Windows          map[string]TextBufferWindow
	CStack           []TextCursorPos
	States           [][]uint64
	UseAlt           bool
	Int              MemoryPublisher
	// Event handlers
	OnTBChange func(tb *TextBuffer)
	OnTBCheck  func(tb *TextBuffer)
}

var cmFilled uint64 = 0xffffffff
var cmSpace uint64 = 0

//func NewTextBuffer(linear bool, base int, font TextSize) *TextBuffer {
//	this := &TextBuffer{
//		Linear:      linear,
//		BaseAddress: base,
//		Size:        bufferSize,
//		Font:        font,
//		Attribute:   VA_NORMAL,
//		Data:        make([]uint64, bufferSize),
//		EX:          bufferWidth - 1,
//		EY:          bufferHeight - 1,
//		FGColor:     15,
//		BGColor:     0,
//	}
//	this.ClearScreen()
//	return this
//}

func NewTextBufferMapped(linear bool, base int, font TextSize, mr *memory.MappedRegion, mp MemoryPublisher) *TextBuffer {
	this := &TextBuffer{
		Linear:           linear,
		BaseAddress:      base,
		Size:             bufferSize,
		Font:             font,
		Attribute:        VA_NORMAL,
		Data:             mr.Data,
		EX:               bufferWidth - 1,
		EY:               bufferHeight - 1,
		FGColor:          15,
		SwitchInterleave: false,
		BGColor:          0,
		Shade:            0,
		Windows:          make(map[string]TextBufferWindow),
		CStack:           make([]TextCursorPos, 0),
		States:           make([][]uint64, 0),
		Int:              mp,
	}
	this.Data.UseMM = false
	this.RegisterWriteMirrors(mr.Data)
	if memory.WarmStart {
		this.Sync()
	} else {
		this.ClearScreen()
		this.Windows["DEFAULT"] = TextBufferWindow{0, 0, 79, 47}
	}
	this.CopyOn()
	this.FullRefresh()
	return this
}

func hAdv(ts TextSize) int {
	w := 1
	ww := int(ts) / 4
	switch ww {
	case 0:
		w = 1
	case 1:
		w = 2
	case 2:
		w = 4
	case 3:
		w = 8
	}
	return w
}

func vAdv(ts TextSize) int {
	w := 1
	ww := int(ts) % 4
	switch ww {
	case 0:
		w = 1
	case 1:
		w = 2
	case 2:
		w = 4
	case 3:
		w = 8
	}
	return w
}

func wozVInterlace(y int) int {
	return y % 2
}

func wozHInterlace(x int) int {
	return x % 2
}

func wozHInterlaceAlt(x int) int {
	return (x + 1) % 2
}

func EncodePosInfo(sx, sy int) uint64 {
	return uint64(sx)<<40 | uint64(sy)<<48
}

func DecodePosInfo(v uint64) (int, int) {
	return int(v >> 40 & 0xff), int(v >> 48 & 0xff)
}

// baseOffsetModXY returns the given memory offset address,
func baseOffsetWozModXY(x, y int) (int, int, int) {
	return wozVInterlace(y)*vBump + wozHInterlace(x)*hBump, x / 2, y / 2
}

func baseOffsetWozModXYAlt(x, y int) (int, int, int) {
	return wozVInterlace(y)*vBump + wozHInterlaceAlt(x)*hBump, x / 2, y / 2
}

// Returns entire memory offset
func baseOffsetWoz(x, y int) int {
	base, mx, my := baseOffsetWozModXY(x, y)
	// At this point base refers to the interlaced Quadrant base
	// address in memory, mx and my are divd by 2 to yield co-ords
	// ammmenable to the standard memory map Woz calcs
	jump := ((my % 8) * 128) + ((my / 8) * 40) + mx
	return base + jump
}

func baseOffsetWozAlt(x, y int) int {
	base, mx, my := baseOffsetWozModXYAlt(x, y)
	// At this point base refers to the interlaced Quadrant base
	// address in memory, mx and my are divd by 2 to yield co-ords
	// ammmenable to the standard memory map Woz calcs
	jump := ((my % 8) * 128) + ((my / 8) * 40) + mx
	return base + jump
}

func baseOffsetLinear(x, y int) int {
	return (x % bufferWidth) + (y%bufferHeight)*bufferWidth
}

func offsetToXYLinear(offset int) (int, int) {
	y := offset / bufferWidth
	x := offset % bufferWidth
	return x, y
}

func offsetToXYWoz(offset int) (int, int) {
	ymod2 := offset / vBump           // 0 or 1
	xmod2 := (offset % vBump) / hBump // 0 or 1
	jump := offset % hBump
	mx := (jump % 128) % 40
	my := ((jump % 128) / 40 * 8) + ((jump / 128) % 8)
	return mx*2 + xmod2, my*2 + ymod2
}

func offsetToXYWozAlt(offset int) (int, int) {
	ymod2 := offset / vBump           // 0 or 1
	xmod2 := (offset % vBump) / hBump // 0 or 1
	jump := offset % hBump
	mx := (jump % 128) % 40
	my := ((jump % 128) / 40 * 8) + ((jump / 128) % 8)
	return mx*2 + (xmod2+1)%2, my*2 + ymod2
}

func decodePackedCharLegacy(memval uint64) (rune, VideoAttribute, uint64, uint64, uint64, int, int) {
	ch := rune(pokeToAscii(memval))
	va := pokeToAttribute(memval)
	shade := (memval >> 28) & 0x7
	colidx := (memval >> 16) & 0x0f
	bcolidx := (memval >> 20) & 0x0f
	tm := TextSize((memval >> 24) & 15)
	ww := int(tm) / 4
	hh := int(tm) % 4
	w := 1
	h := 1
	switch ww {
	case 0:
		w = 1
	case 1:
		w = 2
	case 2:
		w = 4
	case 3:
		w = 8
	}
	switch hh {
	case 0:
		h = 1
	case 1:
		h = 2
	case 2:
		h = 4
	case 3:
		h = 8
	}
	return ch, va, colidx, bcolidx, shade, w, h
}

func encodePackedCharLegacy(ch rune, va VideoAttribute, colidx uint64, bcolidx uint64, shade uint64, ww int, hh int, usealt bool) uint64 {
	w := 0

	switch ww {
	case 1:
		w = 0
	case 2:
		w = 1
	case 4:
		w = 2
	case 8:
		w = 3
	}
	h := 0
	switch hh {
	case 1:
		h = 0
	case 2:
		h = 1
	case 4:
		h = 2
	case 8:
		h = 3
	}
	return uint64(asciiToPoke(int(ch), va, usealt)) | uint64((colidx&0xf)<<16) | uint64((bcolidx&0xf)<<20) | uint64(w*4+h)<<24 | ((shade & 0x7) << 28)
}

func EncodePackedCharLegacy(ch rune, va VideoAttribute, colidx uint64, bcolidx uint64, shade uint64, ww int, hh int, usealt bool) uint64 {
	w := 0
	switch ww {
	case 1:
		w = 0
	case 2:
		w = 1
	case 4:
		w = 2
	case 8:
		w = 3
	}
	h := 0
	switch hh {
	case 1:
		h = 0
	case 2:
		h = 1
	case 4:
		h = 2
	case 8:
		h = 3
	}
	return uint64(asciiToPoke(int(ch), va, usealt)) | uint64((colidx&0xf)<<16) | uint64((bcolidx&0xf)<<20) | uint64(w*4+h)<<24 | ((shade & 0x7) << 28)
}

func Between(v, lo, hi uint64) bool {
	return ((v >= lo) && (v <= hi))
}

func pokeToAscii(v uint64) int {
	highbit := v & 1024
	v = v & 1023
	if Between(v, 0, 31) {
		return int((64 + (v % 32)) | highbit)
	}
	if Between(v, 32, 63) {
		return int((32 + (v % 32)) | highbit)
	}
	if Between(v, 64, 95) {
		return int((64 + (v % 32)) | highbit)
	}
	if Between(v, 96, 127) {
		return int((32 + (v % 32)) | highbit)
	}
	if Between(v, 128, 159) {
		return int((64 + (v % 32)) | highbit)
	}
	if Between(v, 160, 191) {
		return int((32 + (v % 32)) | highbit)
	}
	if Between(v, 192, 223) {
		return int((64 + (v % 32)) | highbit)
	}
	if Between(v, 224, 255) {
		return int((96 + (v % 32)) | highbit)
	}
	if Between(v, 256, 287) {
		return int(v | highbit)
	}
	return int(v | highbit)
}

func pokeToAttribute(v uint64) VideoAttribute {
	v = v & 1023
	va := VA_INVERSE
	if (v & 64) > 0 {
		va = VA_BLINK
	}
	if (v & 128) > 0 {
		va = VA_NORMAL
	}
	if (v & 256) > 0 {
		va = VA_NORMAL
	}
	return va
}

func asciiToPoke(v int, va VideoAttribute, usealt bool) int {
	highbit := v & 1024
	v = v & 1023
	if v > 255 {
		return v | highbit
	}
	v = (v & 127)
	if va == VA_NORMAL {
		return (v + 128) | highbit
	}
	if va == VA_INVERSE && !usealt {
		if (v >= ' ') && (v < '@') {
			return v | highbit
		} else if (v >= 96) && (v <= 127) {
			return (v - 64) | highbit
		} else {
			return (v % 32) | highbit
		}
	}
	if va == VA_INVERSE && usealt {
		if (v >= ' ') && (v < '@') {
			return v | highbit
		} else if (v >= 96) && (v <= 127) {
			return v | highbit
		} else {
			return (v % 32) | highbit
		}
	}
	// flash
	if (v >= 64) && (v < 96) {
		return (64 + (v % 32)) | highbit
	}
	if (v >= 32) && (v < 64) {
		return (96 + (v % 32)) | highbit
	}
	return v | highbit
}

func (tb *TextBuffer) OffsetToY(offset int) int {
	_, y := offsetToXYWozAlt(offset)
	return y / 2
}

func (tb *TextBuffer) RegisterWriteMirrors(mcb *memory.MemoryControlBlock) {
	mm := mcb.GetMM()
	index := mcb.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE
	mm.WriteMirrorsClear(index)
	mm.WriteMirrorRegister(index, memory.NewWriteMirror(
		0x0400,
		0x0400,
		[]int{0x30000},
		func(mm *memory.MemoryMap, index int, address int, value uint64) {
			//fmt.Printf("Mirror write to %d of %d\n", address, value)
			mm.WriteInterpreterMemorySilent(index, address, cmFilled)
		},
	),
	)
	mm.WriteMirrorRegister(index, memory.NewWriteMirror(
		0x10400,
		0x0400,
		[]int{0x30800},
		func(mm *memory.MemoryMap, index int, address int, value uint64) {
			//fmt.Printf("Mirror write to %d of %d\n", address, value)
			mm.WriteInterpreterMemorySilent(index, address, cmFilled)
		},
	),
	)

}

func (tb *TextBuffer) CheckState() {
	if tb.OnTBCheck != nil {
		tb.OnTBCheck(tb)
	}
}

func (tb *TextBuffer) PutState() {
	if tb.OnTBChange != nil {
		tb.OnTBChange(tb)
	}
}

func (tb *TextBuffer) SaveState() {

	// save current buffer
	tmp := make([]uint64, tb.Data.Size)
	for i := range tmp {
		tmp[i] = tb.Data.Read(i)
	}
	tb.States = append(tb.States, tmp)
}

func (tb *TextBuffer) RestoreState() {

	if len(tb.States) == 0 {
		return
	}
	tmp := tb.States[len(tb.States)-1]
	tb.States = tb.States[0 : len(tb.States)-1]

	for i, v := range tmp {
		tb.Data.Write(i, v)
	}

	tb.Sync()
	tb.FullRefresh()
}

func (tb *TextBuffer) HideCursor() {

	v := tb.Data.Read(cVis)

	bitmask := uint64(127 + 256)

	nv := (v & bitmask) | 128

	if v == nv {
		return
	}

	tb.Data.Write(cVis, nv)
}

func (tb *TextBuffer) ShowCursor() {

	v := tb.Data.Read(cVis)

	bitmask := uint64(127 + 256)

	nv := (v & bitmask)
	if v == nv {
		return
	}

	tb.Data.Write(cVis, tb.Data.Read(cVis)&bitmask)
}

func (tb *TextBuffer) FullRefresh() {
	tb.Data.Write(cVis, (tb.Data.Read(cVis)&255)|256)
}

func (tb *TextBuffer) NormalChars() {

	bitmask := uint64(191 + 256)

	tb.Data.Write(cVis, tb.Data.Read(cVis)&bitmask)
	tb.UseAlt = false
}

func (tb *TextBuffer) AltChars() {
	bitmask := uint64(191 + 256)
	tb.Data.Write(cVis, (tb.Data.Read(cVis)&bitmask)|64)
	tb.UseAlt = true
}

func (tb *TextBuffer) CopyOn() {

	bitmask := uint64(239 + 256)

	tb.Data.Write(cVis, (tb.Data.Read(cVis)&bitmask)|16)
	//tb.SwitchInterleave = true
}

func (tb *TextBuffer) CopyOff() {
	bitmask := uint64(239 + 256)

	tb.Data.Write(cVis, tb.Data.Read(cVis)&bitmask)
	//tb.SwitchInterleave = false
}

func (tb *TextBuffer) SwitchedInterleave() {

	bitmask := uint64(223 + 256)

	tb.Data.Write(cVis, (tb.Data.Read(cVis)&bitmask)|32)
	tb.SwitchInterleave = true
}

func (tb *TextBuffer) NormalInterleave() {
	bitmask := uint64(223 + 256)

	tb.Data.Write(cVis, tb.Data.Read(cVis)&bitmask)
	tb.SwitchInterleave = false
}

func (tb *TextBuffer) Repos() {
	tb.Data.Write(xPos, uint64(tb.CX))
	tb.Data.Write(yPos, uint64(tb.CY))
	tb.Data.Write(sX, uint64(tb.SX))
	tb.Data.Write(sY, uint64(tb.SY))
	tb.Data.Write(eX, uint64(tb.EX))
	tb.Data.Write(eY, uint64(tb.EY))
	tb.Data.Write(cSize, uint64(tb.Font)|(tb.FGColor<<8)|(tb.BGColor<<16))
	tb.PutState()
}

func (tb *TextBuffer) Sync() {
	tb.CX = int(tb.Data.Read(xPos))
	tb.CY = int(tb.Data.Read(yPos))
	tb.SX = int(tb.Data.Read(sX))
	tb.SY = int(tb.Data.Read(sY))
	tb.EX = int(tb.Data.Read(eX))
	tb.EY = int(tb.Data.Read(eY))
	tb.Font = TextSize(tb.Data.Read(cSize) & 0xf)
	tb.FGColor = tb.Data.Read(cSize) >> 8
	tb.BGColor = tb.Data.Read(cSize) >> 16
	v := tb.Data.Read(cVis)
	tb.SwitchInterleave = (v&32 != 0)
	tb.UseAlt = (v&64 != 0)
}

// Core functions
// PutValueXY plots a data value into the buffer, based on an 80x48 co-ordinate
func (tb *TextBuffer) PutValueXY(x, y int, value uint64) {
	var offset int
	if tb.Linear {
		offset = baseOffsetLinear(x, y)
	} else {
		if tb.SwitchInterleave {
			offset = baseOffsetWozAlt(x, y)
		} else {
			offset = baseOffsetWoz(x, y)
		}
	}

	tb.Data.Write(offset, value)
}

// GetValueXY gets a data value from the buffer, based on an 80x48 co-ordinate
func (tb *TextBuffer) GetValueXY(x, y int) uint64 {
	var offset int
	if tb.Linear {
		offset = baseOffsetLinear(x, y)
	} else {
		if tb.SwitchInterleave {
			offset = baseOffsetWozAlt(x, y)
		} else {
			offset = baseOffsetWoz(x, y)
		}
	}
	return tb.Data.Read(offset)
}

func (tb *TextBuffer) SetFGColorXY(x, y int, c uint64) {
	memval := tb.GetValueXY(x, y)
	ch, va, fg, bg, shade, w, h := decodePackedCharLegacy(memval)
	fg = c
	memval = encodePackedCharLegacy(ch, va, fg, bg, shade, w, h, false)
	tb.PutValueXY(x, y, memval)
}

func (tb *TextBuffer) SetBGColorXY(x, y int, c uint64) {
	memval := tb.GetValueXY(x, y)
	ch, va, fg, bg, shade, w, h := decodePackedCharLegacy(memval)
	bg = uint64(c)
	memval = encodePackedCharLegacy(ch, va, fg, bg, shade, w, h, false)
	tb.PutValueXY(x, y, memval)
}

// Empty fills the buffer with zeros, which is a kind of null-reference (useful for getting a known state)
func (tb *TextBuffer) Empty() {
	for i := 0; i < tb.Data.Size; i++ {
		tb.Data.Write(i, 0)
	}
}

func (tb *TextBuffer) CursorRight() {
	xskip := hAdv(tb.Font)
	if !tb.MoveCursorX(xskip) {
		// cannot go right..
		tb.CX = tb.SX
		tb.CursorDown()
	}
}

func (tb *TextBuffer) CursorLeft() {
	xskip := hAdv(tb.Font)
	yskip := vAdv(tb.Font)
	if !tb.MoveCursorX(-xskip) {
		// hit left edge
		if tb.MoveCursorY(-yskip) {
			// must be at top-left region, whatevs
			// now go left from right edge
			tb.CX = tb.EX + 1
			tb.MoveCursorX(-xskip)
		}
	}
}

//
func (tb *TextBuffer) ThisLineHeight() int {
	var v int = 1
	var xskip int = 1
	var yskip int = 1
	for x := tb.SX; x <= tb.EX; x += xskip {
		memval := tb.GetValueXY(x, tb.CY)
		_, _, _, _, _, xskip, yskip = decodePackedCharLegacy(memval)
		if yskip > v {
			v = yskip
		}
	}
	return v
}

func (tb *TextBuffer) CursorUp() {
	// Move up (or dont)
	yskip := vAdv(tb.Font)
	tb.MoveCursorY(-yskip)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (tb *TextBuffer) CursorDown() {
	yskip := vAdv(tb.Font)
	//hh := vAdv(tb.Font)
	if !tb.MoveCursorY(yskip) {
		tb.ScrollByWindow(yskip)
		tb.CY = tb.EY - yskip + 1
	}
}

func (tb *TextBuffer) ScrollBy(lines int) {
	var m uint64
	for y := lines; y < bufferHeight; y++ {
		for x := 0; x < bufferWidth; x++ {
			m = tb.GetValueXY(x, y)
			tb.PutValueXY(x, y-lines, m)
		}
	}
	tb.Fill(true, 0, bufferHeight-lines, bufferWidth-1, bufferHeight-1, ' ', tb.Attribute, tb.FGColor, tb.BGColor, tb.Shade)
	//tb.FullRefresh()
}

func (tb *TextBuffer) ScrollByWindow(lines int) {
	var m uint64
	//fmt.Printf("Scroll Window: SX=%d, SY=%d, EX=%d, EY=%d, lines=%d\n", tb.SX, tb.SY+lines, tb.EX, tb.EY, lines)
	for y := tb.SY + lines; y <= tb.EY; y++ {
		for x := tb.SX; x <= tb.EX; x++ {
			m = tb.GetValueXY(x, y)
			tb.PutValueXY(x, y-lines, m)
		}
	}

	//time.Sleep(3 * time.Second)

	tb.Fill(true, tb.SX, tb.EY+1-lines, tb.EX, tb.EY, ' ', tb.Attribute, tb.FGColor, tb.BGColor, tb.Shade)

	tb.FullRefresh()
}

func (tb *TextBuffer) FillAttr(empty bool, sx, sy, ex, ey int, ch rune, va VideoAttribute, colidx, bcolidx, shade uint64) {
	if empty {
		for y := sy; y <= ey; y += 1 {
			for x := sx; x <= ex; x += 1 {
				tb.PutValueXY(x, y, cmSpace) // prezero this data
			}
		}
	}
	yskip := vAdv(tb.Font)
	xskip := hAdv(tb.Font)
	memval := encodePackedCharLegacy(ch, va, colidx, bcolidx, shade, xskip, yskip, tb.UseAlt)
	for y := sy; y <= ey; y += yskip {
		for x := sx; x <= ex; x += xskip {
			oval := tb.GetValueXY(x, y)
			tb.PutValueXY(x, y, (memval&0xffff0000)|(oval&0xffff))
		}
	}
	//tb.FullRefresh()
}

func (tb *TextBuffer) FillSequence(empty bool, sx, sy, ex, ey int, ch []rune, va VideoAttribute, colidx, bcolidx, shade uint64) {
	if empty {
		for y := sy; y <= ey; y += 1 {
			for x := sx; x <= ex; x += 1 {
				tb.PutValueXY(x, y, cmSpace) // prezero this data
			}
		}
	}
	yskip := vAdv(tb.Font)
	xskip := hAdv(tb.Font)

	for y := sy; y <= ey; y += yskip {
		for x := sx; x <= ex; x += xskip {
			memval := encodePackedCharLegacy(ch[x%len(ch)], va, colidx, bcolidx, shade, xskip, yskip, tb.UseAlt)
			tb.PutValueXY(x, y, memval)
		}
	}
	//tb.FullRefresh()
}

func (tb *TextBuffer) Fill(empty bool, sx, sy, ex, ey int, ch rune, va VideoAttribute, colidx, bcolidx, shade uint64) {
	if empty {
		for y := sy; y <= ey; y += 1 {
			for x := sx; x <= ex; x += 1 {
				tb.PutValueXY(x, y, cmSpace) // prezero this data
			}
		}
	}
	yskip := vAdv(tb.Font)
	xskip := hAdv(tb.Font)
	memval := encodePackedCharLegacy(ch, va, colidx, bcolidx, shade, xskip, yskip, tb.UseAlt)
	for y := sy; y <= ey; y += yskip {
		for x := sx; x <= ex; x += xskip {
			tb.PutValueXY(x, y, memval)
		}
	}
	//tb.FullRefresh()
}

func (tb *TextBuffer) ClearScreen() {
	tb.Fill(true, 0, 0, bufferWidth-1, bufferHeight-1, ' ', tb.Attribute, tb.FGColor, tb.BGColor, tb.Shade)
	tb.HomeCursor()
	tb.FullRefresh()
}

func (tb *TextBuffer) ClearScreenAttr() {
	tb.FillAttr(true, 0, 0, bufferWidth-1, bufferHeight-1, ' ', tb.Attribute, tb.FGColor, tb.BGColor, tb.Shade)
	tb.HomeCursor()
	tb.FullRefresh()
}

func (tb *TextBuffer) ClearScreenWindow() {
	tb.Fill(true, tb.SX, tb.SY, tb.EX, tb.EY, ' ', VA_NORMAL, tb.FGColor, tb.BGColor, tb.Shade)
	tb.HomeCursorWindow()
	tb.FullRefresh()
}

func (tb *TextBuffer) Put(ch rune) {

	// zero old memory
	xskip := hAdv(tb.Font)
	yskip := vAdv(tb.Font)
	/*
		tval := encodePackedCharLegacy(' ', tb.Attribute, tb.FGColor, tb.BGColor, tb.Shade, xskip, yskip, tb.UseAlt)

		omemval := tb.GetValueXY(tb.CX, tb.CY)
		_, _, _, _, _, ww, hh := decodePackedCharLegacy(omemval)
		if omemval != cmFilled && omemval != cmSpace && (ww != xskip || hh != yskip) {
			// get data
			fmt.Printf("Clearing old char (%d,%d) to (%d,%d)\n", tb.CX, tb.CY, tb.CX+ww-1, tb.CY+hh-1)
			for y := tb.CY; y < tb.CY+hh; y++ {
				for x := tb.CX; x < tb.CX+ww; x++ {
					if (x%xskip == 0) && (y%yskip == 0) {
						tb.PutValueXY(x, y, tval)
					} else {
						tb.PutValueXY(x, y, cmFilled)
					}
				}
			}
		} else if omemval == cmFilled {
			fmt.Printf("Skip put at (%d, %d) due to FILLED attribute\n", tb.CX, tb.CY)
			tb.CursorRight()
			//tb.PutState()
			return
		}*/

	// Write chat / move cursor...
	//    log.Printf("Putting char [%c] at (%d,%d)\n", ch, tb.CX, tb.CY)
	//r := uint64(utils.Random()*256) << 32
	memval := encodePackedCharLegacy(ch, tb.Attribute, tb.FGColor, tb.BGColor, tb.Shade, xskip, yskip, tb.UseAlt) // | r
	for y := tb.CY; y < tb.CY+yskip; y++ {
		for x := tb.CX; x < tb.CX+xskip; x++ {
			if x == tb.CX && y == tb.CY {
				tb.PutValueXY(x, y, memval)
				if ch > 32 && ch < 127 && tb.FGColor != 0 {
					settings.BlueScreen = false
				}
			} else if x < 80 && y < 48 {
				tb.PutValueXY(x, y, cmFilled)
			}
		}
	}
	tb.CursorRight()

	//tb.PutState()
}

func (tb *TextBuffer) CR() {
	tb.CX = tb.SX
	tb.Repos()
}

func (tb *TextBuffer) LF() {
	if noNLOutOfBounds && tb.CY > tb.EY {
		return
	}

	tb.CursorDown()
	tb.Repos()
}

func (tb *TextBuffer) GetStrings() []string {
	out := make([]string, 0)
	for y := 0; y < bufferHeight; y++ {
		s := ""
		for x := 0; x < bufferWidth; x++ {
			memval := tb.GetValueXY(x, y)
			if memval != 0 && memval != cmFilled && memval != cmSpace {
				ch, _, _, _, _, _, _ := decodePackedCharLegacy(memval)
				s = s + string(ch)
			}
		}
		out = append(out, s+"\n")
	}
	return out
}

func (tb *TextBuffer) RuneAt(x, y int) (rune, error) {
	memval := tb.GetValueXY(x, y)
	if memval == 0 {
		return rune(0), errors.New("No character")
	}
	ch, _, _, _, _, _, _ := decodePackedCharLegacy(memval)
	return ch, nil
}

func (tb *TextBuffer) ClearToEOL() {
	// clear from current CX all the way to bufferWidth, in steps based on stuff
	sx, sy, ex, ey := tb.CX, tb.CY, bufferWidth-1, tb.CY+vAdv(tb.Font)-1
	tb.Fill(false, sx, sy, ex, ey, ' ', tb.Attribute, tb.FGColor, tb.BGColor, tb.Shade)
}

func (tb *TextBuffer) ClearToEOLWindow() {
	// clear from current CX all the way to bufferWidth, in steps based on stuff
	sx, sy, ex, ey := tb.CX, tb.CY, tb.EX, tb.CY+vAdv(tb.Font)-1
	tb.Fill(false, sx, sy, ex, ey, ' ', VA_NORMAL, tb.FGColor, tb.BGColor, tb.Shade)
}

func (tb *TextBuffer) ClearToBottom() {
	// clear from current CX all the way to bufferWidth, in steps based on stuff
	sx, sy, ex, ey := tb.CX, tb.CY+vAdv(tb.Font), bufferWidth-1, bufferHeight-1
	tb.Fill(false, sx, sy, ex, ey, ' ', tb.Attribute, tb.FGColor, tb.BGColor, tb.Shade)
	tb.ClearToEOL() // implied
}

func (tb *TextBuffer) ClearToBottomWindow() {
	// clear from current CX all the way to bufferWidth, in steps based on stuff
	sx, sy, ex, ey := tb.SX, tb.CY+vAdv(tb.Font), tb.EX, tb.EY
	tb.Fill(false, sx, sy, ex, ey, ' ', VA_NORMAL, tb.FGColor, tb.BGColor, tb.Shade)
	tb.ClearToEOLWindow() // implied
}

// Move cursor to literal X, Y even if no char there
func (tb *TextBuffer) GotoXY(x, y int) {

	if x < tb.SX {
		x = tb.SX
	}
	if x > tb.EX {
		x = tb.EX
	}
	if y < tb.SY {
		y = tb.SY
	}
	if y > tb.EY {
		y = tb.EY
	}

	//tb.CX, tb.CY = x*tb.FontW(), y*tb.FontH()
	tb.CX, tb.CY = x, y
	tb.Repos()
}

func (tb *TextBuffer) GotoXYWindow(x, y int) {
	tb.GotoXY(tb.SX+x, tb.SY+y)
}

func (tb *TextBuffer) PopCursor() {
	if len(tb.CStack) > 0 {
		i := len(tb.CStack) - 1
		tb.CX, tb.CY = tb.CStack[i].CX, tb.CStack[i].CY
		tb.CStack = tb.CStack[0:i]
		tb.Repos()
	}
}

func (tb *TextBuffer) PushCursor() {
	tb.CStack = append(tb.CStack, TextCursorPos{tb.CX, tb.CY})
}

// Move cursor to valid char closest to X, Y
func (tb *TextBuffer) GotoXYSmart(x, y int) {
	memval := tb.GetValueXY(x, y)
	for (memval == cmSpace || memval == cmFilled) && y < tb.EY {
		x += 1
		if x > tb.EX {
			x = tb.SX
			y += 1
			if y > tb.EY {
				return
			}
		}
		memval = tb.GetValueXY(x, y)
	}
	// If we are here we found the next char
	tb.CX, tb.CY = x, y
	tb.Repos()
}

func (tb *TextBuffer) HomeCursor() {
	tb.CX, tb.CY = 0, 0
	tb.Repos()
}

func (tb *TextBuffer) HomeCursorWindow() {
	tb.CX, tb.CY = tb.SX, tb.SY
	tb.Repos()
}

func (tb *TextBuffer) HorizontalTab(count int) {
	// this is postion '1' so we advance count-1 times
	ncx := tb.SX + (count-1)*hAdv(tb.Font)
	if ncx > tb.EX {
		l := ncx / 80
		ncx = ncx % 80
		for i := 0; i < l; i++ {
			tb.CursorDown()
		}
	}
	tb.CX = ncx
	tb.Repos()
}

func (tb *TextBuffer) VerticalTab(count int) {
	// this is postion '1' so we advance count-1 times
	ncy := 0 + (count-1)*vAdv(tb.Font)
	//    if ncy > tb.EY {
	//	   ncy = tb.EY
	//    }

	tb.CY = ncy
	tb.Repos()
}

func (tb *TextBuffer) zHorizontalTab(count int) {
	tb.CX = tb.SX
	// this is postion '1' so we advance count-1 times
	for x := 1; x < count; x++ {
		tb.CursorRight()
	}
}

func (tb *TextBuffer) zVerticalTab(count int) {
	tb.CY = tb.SY
	// this is postion '1' so we advance count-1 times
	for y := 1; y < count; y++ {
		tb.CursorDown()
	}
}

func (tb *TextBuffer) SetWindow(sx, sy, ex, ey int) {
	tb.SX, tb.SY, tb.EX, tb.EY = sx, sy, ex, ey
	//tb.PutState()
}

func (tb *TextBuffer) AddNamedWindow(name string, sx, sy, ex, ey int) {
	tb.Windows[name] = TextBufferWindow{sx, sy, ex, ey}
}

func (tb *TextBuffer) SetNamedWindow(name string) {
	w, ok := tb.Windows[name]
	if ok {
		tb.SetWindow(w.SX, w.SY, w.EX, w.EY)
	}
}

func (tb *TextBuffer) GetXYWindow() (int, int) {
	return tb.CX - tb.SX, tb.CY - tb.SY
}

func (tb *TextBuffer) Backspace() {
	ox, oy := tb.CX, tb.CY
	tb.CursorLeft()
	if tb.CX != ox || tb.CY != oy {
		// success
		yskip := vAdv(tb.Font)
		xskip := hAdv(tb.Font)
		memval := encodePackedCharLegacy(' ', tb.Attribute, tb.FGColor, tb.BGColor, tb.Shade, xskip, yskip, tb.UseAlt)
		tb.PutValueXY(tb.CX, tb.CY, memval)
	}
}

// MoveCursorX tries to move the cursor by dx which can be positive or negative.
// units are 80x48 screen units.
// It will take into account nearest valid char.
// Result: true if cursor was able to move horizontally, false if not
func (tb *TextBuffer) MoveCursorX(dx int) bool {
	//log.Printf( "Move X(%d) requested from (%d,%d)\n", dx, tb.CX, tb.CY )
	amount := int(math.Abs(float64(dx))) / dx
	//found := false
	px := tb.CX + amount*int(math.Abs(float64(dx)))

	if px < tb.SX {
		return false // at left of visible region
	}
	if px > tb.EX {
		return false // at right of visible region
	}

	// Found a spot
	tb.CX = px
	tb.Repos()
	//log.Printf( "After Move X(%d) requested to (%d,%d)\n", dx, tb.CX, tb.CY )

	return true
}

// MoveCursorY tries to move the cursor by dy which can be positive or negative.
// units are 80x48 screen units.
// It will take into account nearest valid char.
// Result: true if cursor was able to move vertcally, false if not
func (tb *TextBuffer) MoveCursorY(dy int) bool {
	amount := int(math.Abs(float64(dy))) / dy
	//	found := false
	py := tb.CY + amount*int(math.Abs(float64(dy)))
	if py > tb.EY {
		return false // at right of visible region
	}

	// Found a spot
	tb.CY = py
	tb.Repos()
	return true
}

func (tb *TextBuffer) NLIN() {
	if tb.CX > tb.SX {
		tb.CR()
		tb.LF()
	}
}

func (tb *TextBuffer) TAB(tn int) {
	//	 log.Println("Tab called", tn)
	tbm := hAdv(tb.Font) * tn
	if (tb.CX % tbm) == 0 {
		tb.CursorRight()
	}
	for (tb.CX % tbm) != 0 {
		tb.CursorRight()
	}
}

func (tb *TextBuffer) SetSizeAndClear(size TextSize) {
	tb.Font = size
	tb.SX = 0
	tb.EX = 79
	tb.SY = 0
	tb.EY = 47
	tb.ClearScreen()
	tb.HomeCursor()
	tb.FullRefresh()
}

func (tb *TextBuffer) SetSizeAndClearAttr(size TextSize) {
	tb.Font = size
	tb.SX = 0
	tb.EX = 79
	tb.SY = 0
	tb.EY = 47
	tb.ClearScreenAttr()
	tb.HomeCursor()
	tb.FullRefresh()
}

func (tb *TextBuffer) SetSizeAndClearWindow(size TextSize) {
	tb.Font = size
	tb.ClearScreenWindow()
	tb.HomeCursorWindow()
	tb.FullRefresh()
}

func (tb *TextBuffer) GetOffsetXY(x, y int) int {
	return baseOffsetWoz(x, y)
}

func (tb *TextBuffer) FontW() int {
	return hAdv(tb.Font)
}

func (tb *TextBuffer) FontH() int {
	return vAdv(tb.Font)
}

func (tb *TextBuffer) DrawTextBoxAuto(content []string, shadow, window bool) {

	var maxw int
	for _, l := range content {
		if utf8.RuneCountInString(l)+2 > maxw {
			maxw = utf8.RuneCountInString(l) + 2
		}
	}
	if maxw > 78 {
		maxw = 78
	}

	h := len(content) + 2
	w := maxw + 2
	x := (80 - w) / 2
	y := (48 - h) / 2

	tb.DrawTextBox(x, y, w, h, strings.Join(content, "\r\n"), shadow, window)

}

// DrawBox will draw a box using the current font, fg, bg etc.
func (tb *TextBuffer) DrawTextBox(x, y, w, h int, content string, shadow bool, window bool) {
	sx, sy := tb.FontW()*x, tb.FontH()*y
	ex, ey := tb.FontW()*w+sx-1, tb.FontH()*h+sy-1

	// validate bounds
	if sx < 0 || sx >= 80 || sy < 0 || sy >= 48 || ex < 0 || ex >= 80 || ey < 0 || ey >= 48 {
		return
	}

	// text bounds - depends on shadow
	tsx, tsy, tex, tey := sx, sy, ex, ey

	if shadow {
		tex, tey = ex-tb.FontW(), ey-tb.FontH()
	}

	// draw the main box...
	// empty bool, sx, sy, ex, ey int, ch rune, va VideoAttribute, colidx, bcolidx uint64

	if shadow {
		// right side
		tb.FillSequence(true, tsx+tb.FontW(), tsy+tb.FontH(), ex-tb.FontW()+1, tey+tb.FontH(), []rune{1057, 1058}, tb.Attribute, tb.FGColor, tb.BGColor, tb.Shade)
		tb.FillSequence(true, tsx+tb.FontW(), ey-tb.FontH()+1, tex+tb.FontW(), ey-tb.FontH()+1, []rune{1057, 1058}, tb.Attribute, tb.FGColor, tb.BGColor, tb.Shade)
	}

	tb.Fill(true, tsx, tsy, tex, tey, ' ', tb.Attribute, tb.FGColor, tb.BGColor, tb.Shade)

	// handle content
	osx, osy, oex, oey, ocx, ocy := tb.SX, tb.SY, tb.EX, tb.EY, tb.CX, tb.CY
	if content != "" {
		// apply window
		tb.SX, tb.SY, tb.EX, tb.EY = tsx, tsy, tex, tey
		tb.CX, tb.CY = tsx, tsy
		for _, ch := range content {
			switch rune(ch) {
			case '\r':
				tb.CR()
			case '\n':
				tb.LF()
			default:
				tb.Put(rune(ch))
			}
		}
		tb.SX, tb.SY, tb.EX, tb.EY, tb.CX, tb.CY = osx, osy, oex, oey, ocx, ocy
	}

	if window {
		tb.SX, tb.SY, tb.EX, tb.EY, tb.CX, tb.CY = tsx, tsy, tex, tey, tsx, tsy
	}
}

func (tb *TextBuffer) PutStr(s string) {

	//tb.CheckState()

	for _, ch := range s {

		if ch >= vduconst.BGCOLOR0 && ch <= vduconst.BGCOLOR15 {
			tb.BGColor = uint64(ch - vduconst.BGCOLOR0)
			continue
		}

		if ch >= vduconst.COLOR0 && ch <= vduconst.COLOR15 {
			tb.FGColor = uint64(ch - vduconst.COLOR0)
			continue
		}

		switch ch {
		case 13:
			tb.CR()
		case 10:
			tb.LF()
		case 8:
			tb.Backspace()
		case 9:
			tb.TAB(8)
		default:
			tb.Put(ch)
		}
	}

	//tb.PutState()

}

func (tb *TextBuffer) Printf(format string, v ...interface{}) {

	s := fmt.Sprintf(format, v...)

	tb.PutStr(s)

}

func (tb *TextBuffer) RotatePalette(low int, high int, change int) {

	r := high - low + 1

	var xskip int = 1
	var memval uint64

	for y := 0; y < 48; y += 1 {

		for x := 0; x < 80; x += xskip {

			memval = tb.GetValueXY(x, y)

			if memval == cmFilled {
				continue
			}

			ch, va, c, b, shade, w, h := decodePackedCharLegacy(memval)

			colidx := int(c)
			bcolidx := int(b)

			xskip = w
			if colidx >= low && colidx <= high {
				colidx = ((colidx - low) + change)
				if colidx < 0 {
					colidx += r
				}
				colidx = (colidx % r) + low
			}
			if bcolidx >= low && bcolidx <= high {
				bcolidx = ((bcolidx - low) + change)
				if bcolidx < 0 {
					bcolidx += r
				}
				bcolidx = (bcolidx % r) + low
			}

			tb.PutValueXY(x, y, encodePackedCharLegacy(ch, va, uint64(colidx), uint64(bcolidx), shade, w, h, tb.UseAlt))

		}

	}

}
