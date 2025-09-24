package apple2

import (
	//	"paleotronic.com/fmt"

	"math/rand"

	"paleotronic.com/log"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

//var mm2e map[*memory.MappedRegion]interfaces.Interpretable

type Apple2ZeroPage struct {
	*memory.MappedRegion
	e interfaces.Interpretable
}

func NewApple2ZeroPage(mm *memory.MemoryMap, globalbase int, base int, ent interfaces.Interpretable) *Apple2ZeroPage {

	this := &Apple2ZeroPage{}

	var rsh [256]memory.ReadSubscriptionHandler
	var esh [256]memory.ExecSubscriptionHandler
	var wsh [256]memory.WriteSubscriptionHandler

	this.MappedRegion = memory.NewMappedRegion(
		mm,
		ent.GetMemIndex(),
		globalbase,
		base,
		256,
		"Apple2IOZeroPage",
		rsh,
		esh,
		wsh,
	)

	this.e = ent

	// and return
	return this
}

func (this *Apple2ZeroPage) Done() {
	// cleanup
}

/*
	ReadCursorXHandler -- triggered when Cursor X address is peeked
*/
func (mr *Apple2ZeroPage) ReadCursorXHandler(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	x := apple2helpers.GetCursorRelativeX(e)
	return uint64(x / (80 / apple2helpers.GetFullColumns(e)))
}

/*
	ReadCursorYHandler -- triggered when Cursor Y address is peeked
*/
func (mr *Apple2ZeroPage) ReadCursorYHandler(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	y := apple2helpers.GetCursorY(e)
	return uint64(y / (48 / apple2helpers.GetFullRows(e)))
}

func (mr *Apple2ZeroPage) ReadWindowLeftMargin(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	return uint64(apple2helpers.GetTextLeftMargin(e))
}

func (mr *Apple2ZeroPage) ReadWindowTopMargin(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	return uint64(apple2helpers.GetTextTopMargin(e))
}

func (mr *Apple2ZeroPage) ReadWindowBottomMargin(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	return uint64(apple2helpers.GetTextBottomMargin(e))
}

func (mr *Apple2ZeroPage) ReadWindowWidth(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	return uint64(apple2helpers.GetTextWidth(e))
}

func (mr *Apple2ZeroPage) WriteCursorXHandler(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	y := apple2helpers.GetCursorY(e)
	x := apple2helpers.TEXT(e).FontW() * int(value)

	apple2helpers.Gotoxy(e, x, y)

	//x := apple2helpers.GetCursorRelativeX(e)
	//return uint64(x / (80 / apple2helpers.GetColumns(e)))
}

func (mr *Apple2ZeroPage) WriteCursorYHandler(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	x := apple2helpers.GetCursorX(e)
	y := apple2helpers.TEXT(e).FontH() * int(value)

	apple2helpers.Gotoxy(e, x, y)

	//x := apple2helpers.GetCursorRelativeX(e)
	//return uint64(x / (80 / apple2helpers.GetColumns(e)))
}

/*
Poke 32,X Set left margin of text window
Poke 33,X Set width of text window
Poke 34,X Set top margin of text window
Poke 35,X Set bottom margin of text window
*/

func (mr *Apple2ZeroPage) WriteWindowLeftMargin(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	apple2helpers.SetTextLeftMargin(e, int(value))
	//apple2helpers.ClipCursor(e)
}

func (mr *Apple2ZeroPage) WriteWindowTopMargin(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	apple2helpers.SetTextTopMargin(e, int(value))

	//apple2helpers.ClipCursor(e)
}

func (mr *Apple2ZeroPage) WriteWindowBottomMargin(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	apple2helpers.SetTextBottomMargin(e, int(value))

	//apple2helpers.ClipCursor(e)
}

func (mr *Apple2ZeroPage) WriteWindowWidth(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	apple2helpers.SetTextWidth(e, int(value))

	//apple2helpers.ClipCursor(e)
}

func (mr *Apple2ZeroPage) ReadPrompt(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}
	return uint64(e.GetPrompt()[0])
}

func (mr *Apple2ZeroPage) WritePrompt(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}
	e.SetPrompt(string(rune(value & 0xffff)))
}

func (mr *Apple2ZeroPage) ReadAttribute(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}
	va := apple2helpers.GetAttribute(e)
	switch va {
	case types.VA_NORMAL:
		return 255
	case types.VA_BLINK:
		return 127
	case types.VA_INVERSE:
		return 63
	}
	return 255
}

func (mr *Apple2ZeroPage) WriteAttribute(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	switch value { /* FIXME - Switch statement needs cleanup */
	case 255:
		apple2helpers.Attribute(e, types.VA_NORMAL)
		break
	case 63:
		apple2helpers.Attribute(e, types.VA_INVERSE)
		break
	case 127:
		apple2helpers.Attribute(e, types.VA_BLINK)
		break
	}
}

func (mr *Apple2ZeroPage) ReadRandom(mm *memory.MappedRegion, address int) uint64 {
	return uint64(rand.Uint32() % 256)
}

func (mr *Apple2ZeroPage) ReadHCOLOR(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	i := *types.NewTokenList()
	i.Push(types.NewToken(types.VARIABLE, "HCOLOR"))
	vtok := e.ParseTokensForResult(i)
	hc := vtok.AsInteger()
	var r uint64
	// removed free call here;

	switch hc { /* FIXME - Switch statement needs cleanup */
	case 0:
		r = 0
		break
	case 1:
		r = 42
		break
	case 2:
		r = 85
		break
	case 3:
		r = 127
		break
	case 4:
		r = 128
		break
	case 5:
		r = 170
		break
	case 6:
		r = 213
		break
	case 7:
		r = 255
		break
	}

	return r
}

func (mr *Apple2ZeroPage) WriteHCOLOR(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	//	cc := 0
	//	switch value { /* FIXME - Switch statement needs cleanup */
	//	case 42:
	//		cc = 1
	//		break
	//	case 85:
	//		cc = 2
	//		break
	//	case 127:
	//		cc = 3
	//		break
	//	case 128:
	//		cc = 4
	//		break
	//	case 170:
	//		cc = 5
	//		break
	//	case 213:
	//		cc = 6
	//		break
	//	case 255:
	//		cc = 7
	//		break
	//	}

	//	ntl := *types.NewTokenList()
	//	ntl.Push(types.NewToken(types.VARIABLE, "hcolor"))
	//	ntl.Push(types.NewToken(types.ASSIGNMENT, "="))
	//	ntl.Push(types.NewToken(types.NUMBER, utils.IntToStr(cc)))
	//	a := e.GetCode()
	//	e.GetDialect().ExecuteDirectCommand(ntl, e, &a, e.GetLPC())

	hires.HGRMASK = int(value & 0xff)
	mm.Global.WriteInterpreterMemorySilent(e.GetMemIndex(), 228, value&0xff)
}

func (mr *Apple2ZeroPage) WriteROT(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	ntl := *types.NewTokenList()
	ntl.Push(types.NewToken(types.VARIABLE, "rot"))
	ntl.Push(types.NewToken(types.ASSIGNMENT, "="))
	ntl.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(value))))
	a := e.GetCode()
	e.GetDialect().ExecuteDirectCommand(ntl, e, a, e.GetLPC())
}

func (mr *Apple2ZeroPage) ReadROT(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	i := *types.NewTokenList()
	i.Push(types.NewToken(types.VARIABLE, "ROT"))
	vtok := e.ParseTokensForResult(i)
	return uint64(vtok.AsInteger())
}

func (mr *Apple2ZeroPage) WriteSPEED(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	ntl := *types.NewTokenList()
	ntl.Push(types.NewToken(types.VARIABLE, "speed"))
	ntl.Push(types.NewToken(types.ASSIGNMENT, "="))
	ntl.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(value))))
	a := e.GetCode()
	e.GetDialect().ExecuteDirectCommand(ntl, e, a, e.GetLPC())
}

func (mr *Apple2ZeroPage) ReadSPEED(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	i := *types.NewTokenList()
	i.Push(types.NewToken(types.VARIABLE, "SPEED"))
	vtok := e.ParseTokensForResult(i)
	return uint64(vtok.AsInteger())
}

func (mr *Apple2ZeroPage) WriteSCALE(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	ntl := *types.NewTokenList()
	ntl.Push(types.NewToken(types.VARIABLE, "scale"))
	ntl.Push(types.NewToken(types.ASSIGNMENT, "="))
	ntl.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(value))))
	a := e.GetCode()
	e.GetDialect().ExecuteDirectCommand(ntl, e, a, e.GetLPC())
}

func (mr *Apple2ZeroPage) ReadSCALE(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	i := *types.NewTokenList()
	i.Push(types.NewToken(types.VARIABLE, "SCALE"))
	vtok := e.ParseTokensForResult(i)
	return uint64(vtok.AsInteger())
}

func (mr *Apple2ZeroPage) Read222(mm *memory.MappedRegion, address int) uint64 {
	r := mm.Data.Read(222)
	mm.Data.Write(222, 0)
	return r
}

func (mr *Apple2ZeroPage) ReadLineLo(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	r := 1024 + apple2helpers.GetWozOffsetLine(e)

	return uint64(r % 256)
}

func (mr *Apple2ZeroPage) ReadLineHi(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	r := 1024 + apple2helpers.GetWozOffsetLine(e)

	return uint64(r / 256)
}

func (mr *Apple2ZeroPage) WriteCurrentPage(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	switch value {
	case 32:
		e.SetCurrentPage("HGR1")
	case 64:
		e.SetCurrentPage("HGR2")
	}
}

func (mr *Apple2ZeroPage) WriteCollisionReg(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	apple2helpers.SetHGRCollisionCount(e, value)
}

func (mr *Apple2ZeroPage) ReadCurrentPage(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	switch e.GetCurrentPage() {
	case "HGR1":
		return 32
	case "HGR2":
		return 64
	}

	return 32
}

func (mr *Apple2ZeroPage) ReadCollisionReg(mm *memory.MappedRegion, address int) uint64 {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	log.Println("In ReadCollisionReg() ZP handler")

	return apple2helpers.GetHGRCollisionCount(e)
}

// 1 = Mask ZP handler during ML, 0 = Don't mask ZP handler
func (mr *Apple2ZeroPage) SetZPMaskState(mm *memory.MappedRegion, address int, value uint64) {
	// remove mutex
	// remove mutex
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}
	mp, ex := e.GetMemoryMap().InterpreterMappableByLabel(e.GetMemIndex(), "Apple2IOZeroPage")
	if ex {
		//fmt.Printf("Set ZP Mask state = %v\n", (value != 0))
		mp.SetMaskEnabled(value != 0)
	}
}

func (mr *Apple2ZeroPage) RelativeWrite(offset int, value uint64) {

	if offset >= mr.Size {
		return // ignore write outside our bounds
	}

	if !settings.SlotZPEmu[mr.e.GetMemIndex()] {
		mr.Data.Data[0][offset] = value
		return
	}

	//fmt.Printf("Zero page hit for write @ %d\n", offset)

	// switch here
	switch offset {
	case 36:
		mr.WriteCursorXHandler(mr.MappedRegion, offset, value)
	case 37:
		mr.WriteCursorYHandler(mr.MappedRegion, offset, value)
	case 32:
		mr.WriteWindowLeftMargin(mr.MappedRegion, offset, value)
	case 33:
		mr.WriteWindowWidth(mr.MappedRegion, offset, value)
	case 34:
		mr.WriteWindowTopMargin(mr.MappedRegion, offset, value)
	case 35:
		mr.WriteWindowBottomMargin(mr.MappedRegion, offset, value)
	case 51:
		//debug.PrintStack()
		mr.WritePrompt(mr.MappedRegion, offset, value)
	case 50:
		mr.WriteAttribute(mr.MappedRegion, offset, value)
	case 230:
		mr.WriteCurrentPage(mr.MappedRegion, offset, value)
	case 234:
		mr.WriteCollisionReg(mr.MappedRegion, offset, value)
	default:
		mr.Data.Data[0][offset] = value
	}
}

/* RelativeRead handles a read within this regions address space */
func (mr *Apple2ZeroPage) RelativeRead(offset int) uint64 {
	if offset >= mr.Size {
		return 0 // ignore read outside our bounds
	}

	if !settings.SlotZPEmu[mr.e.GetMemIndex()] {
		return mr.Data.Data[0][offset]
	}

	// switch here
	switch offset {
	case 36:
		return mr.ReadCursorXHandler(mr.MappedRegion, offset)
	case 37:
		return mr.ReadCursorYHandler(mr.MappedRegion, offset)
	case 32:
		return mr.ReadWindowLeftMargin(mr.MappedRegion, offset)
	case 33:
		return mr.ReadWindowWidth(mr.MappedRegion, offset)
	case 34:
		return mr.ReadWindowTopMargin(mr.MappedRegion, offset)
	case 35:
		return mr.ReadWindowBottomMargin(mr.MappedRegion, offset)
	case 51:
		return mr.ReadPrompt(mr.MappedRegion, offset)
	case 50:
		return mr.ReadAttribute(mr.MappedRegion, offset)
	case 78:
		return mr.ReadRandom(mr.MappedRegion, offset)
	case 79:
		return mr.ReadRandom(mr.MappedRegion, offset)
	case 222:
		return mr.Read222(mr.MappedRegion, offset)
	case 41:
		return mr.ReadLineHi(mr.MappedRegion, offset)
	case 40:
		return mr.ReadLineLo(mr.MappedRegion, offset)
	case 230:
		return mr.ReadCurrentPage(mr.MappedRegion, offset)
	case 234:
		return mr.ReadCollisionReg(mr.MappedRegion, offset)
	}

	return mr.Data.Data[0][offset]
}

func (mr *Apple2ZeroPage) Read(address int) uint64 {
	return mr.RelativeRead(address - mr.Base)
}

func (mr *Apple2ZeroPage) Exec(address int) {
	mr.RelativeExec(address - mr.Base)
}

func (mr *Apple2ZeroPage) Write(address int, value uint64) {
	mr.RelativeWrite(address-mr.Base, value)
}
