package vdu

import (
	"paleotronic.com/core/types"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/runestring"
	"paleotronic.com/files"
	"paleotronic.com/core/hires"

)

/* Apple2Display satisfies the Display interface */
type Apple2Display struct {
}


func (this *Apple2Display) SetClassicHGR(v bool)  {

	return
}


func (this *Apple2Display) HandleEvent(e types.Event)  {

	return
}


func (this *Apple2Display) RealPut(ch rune)  {

	return
}


func (this *Apple2Display) PutStr(s string) error {
	var e error

	return e
}


func (this *Apple2Display) SetPaddleButton(index int, value bool)  {

	return
}


func (this *Apple2Display) SetPaddleModifier(index int, value float32)  {

	return
}


func (this *Apple2Display) SetVideoMode(vm types.VideoMode)  {

	return
}


func (this *Apple2Display) InitDisplay(vm types.VideoMode)  {

	return
}


func (this *Apple2Display) GfxClear(idx int)  {

	return
}


func (this *Apple2Display) ColorAt(x, y int) int {
	var i int

	return i
}


func (this *Apple2Display) Swap(a *int, b *int)  {

	return
}


func (this *Apple2Display) SetDisplayPage(i int)  {

	return
}


func (this *Apple2Display) CharAt(x int, y int) int {
	var i int

	return i
}


func (this *Apple2Display) SetPage(i int)  {

	return
}


func (this *Apple2Display) SetCursorVisible(v bool)  {

	return
}


func (this *Apple2Display) PaddleButton(i int)  {

	return
}


func (this *Apple2Display) InsertCharToBuffer(ch rune)  {

	return
}


func (this *Apple2Display) ProcessKeyBuffer(ent interfaces.Interpretable)  {

	return
}


func (this *Apple2Display) Put(ch rune)  {

	return
}


func (this *Apple2Display) RollPaddle(i int, amount int)  {

	return
}


func (this *Apple2Display) SetCursorX(x int)  {

	return
}


func (this *Apple2Display) SetCursorY(y int)  {

	return
}


func (this *Apple2Display) SetSpeed(s int)  {

	return
}


func (this *Apple2Display) SetPaddleValues(z int, v int)  {

	return
}


func (this *Apple2Display) GetPaddleValues(z int) int {
	var i int

	return i
}


func (this *Apple2Display) SetPaddleButtons(z int, v bool)  {

	return
}


func (this *Apple2Display) SetBuffer(s runestring.RuneString)  {

	return
}


func (this *Apple2Display) SetMemory(v []int)  {

	return
}


func (this *Apple2Display) SetFeedBuffer(s string)  {

	return
}


func (this *Apple2Display) SetOutChannel(s string)  {

	return
}


func (this *Apple2Display) SetSuppressFormat(v bool)  {

	return
}


func (this *Apple2Display) SetAttribute(v types.VideoAttribute)  {

	return
}


func (this *Apple2Display) GfxClearSplit(v int)  {

	return
}


func (this *Apple2Display) RegenerateWindow(v []int)  {

	return
}


func (this *Apple2Display) RegenerateMemory(ent interfaces.Interpretable)  {

	return
}


func (this *Apple2Display) SetPrompt(s string)  {

	return
}


func (this *Apple2Display) SetCurrentPage(i int)  {

	return
}


func (this *Apple2Display) GetPaddleButtons(i int) bool {
	var b bool

	return b
}


func (this *Apple2Display) SetTabWidth(i int)  {

	return
}


func (this *Apple2Display) SetTextMemory(m *types.TXMemoryBuffer)  {

	return
}


func (this *Apple2Display) SetShadowTextMemory(m *types.TXMemoryBuffer)  {

	return
}


func (this *Apple2Display) GrPlot(x, y, c int)  {

	return
}


func (this *Apple2Display) SetLastX(x int)  {

	return
}


func (this *Apple2Display) SetLastY(y int)  {

	return
}


func (this *Apple2Display) XYToOffset(x, y int) int {
	var i int

	return i
}


func (this *Apple2Display) GrVertLine(x, y0, y1, c int)  {

	return
}


func (this *Apple2Display) GrHorizLine(x0, x1, y, c int)  {

	return
}


func (this *Apple2Display) HgrPlot(x2, y2 int, hc int)  {

	return
}


func (this *Apple2Display) HgrLine(x1, y1, x2, y2 int, hc int)  {

	return
}


func (this *Apple2Display) HgrFill(hc int)  {

	return
}


func (this *Apple2Display) HgrShape(shape hires.ShapeEntry, x int, y int, scl int, deg int, c int, usecol bool)  {

	return
}


func (this *Apple2Display) SetFGColour(c int)  {

	return
}


func (this *Apple2Display) SetBGColour(c int)  {

	return
}


func (this *Apple2Display) SetBGColourTriple(r, g, b int)  {

	return
}


func (this *Apple2Display) SetMemoryValue(addr, value int) bool {
	var b bool

	return b
}


func (this *Apple2Display) GetMemoryValue(addr int) (int, bool) {
	var i int
	var b bool

	return i, b
}


func (this *Apple2Display) ExecNative(mem []int, a int, x int, y int, pc int, sr int, sp int, vdu interfaces.Display)  {

	return
}


func (this *Apple2Display) PassWaveBuffer(data []float32)  {

	return
}


func (this *Apple2Display) AssetCheck(p, f string) (*files.FilePack, error) {
	var fp *files.FilePack
	var e error

	return fp, e
}


func (this *Apple2Display) PlayWave(p, f string) (bool, error) {
	var b bool
	var e error

	return b, e
}


func (this *Apple2Display) PNGSplash(p, f string) (bool, error) {
	var b bool
	var e error

	return b, e
}


func (this *Apple2Display) PNGBackdrop(p, f string) (bool, error) {
	var b bool
	var e error

	return b, e
}


func (this *Apple2Display) LoadEightTrack(p, f string) (bool, error) {
	var b bool
	var e error

	return b, e
}


func (this *Apple2Display) SetBreak(v bool)  {

	return
}


func (this *Apple2Display) SetColorFlip(b bool)  {

	return
}


func (this *Apple2Display) SetCGColour(c int)  {

	return
}


func (this *Apple2Display) SendRestalgiaEvent(kind byte, content string)  {

	return
}


func (this *Apple2Display) CamDolly(r float32)  {

	return
}


func (this *Apple2Display) CamZoom(r float32)  {

	return
}


func (this *Apple2Display) CamPos(x, y, z float32)  {

	return
}


func (this *Apple2Display) CamPivPnt(x, y, z float32)  {

	return
}


func (this *Apple2Display) CamMove(x, y, z float32)  {

	return
}


func (this *Apple2Display) CamRot(x, y, z float32)  {

	return
}


func (this *Apple2Display) CamOrbit(x, y float32)  {

	return
}


func (this *Apple2Display) SetTextMode(tm types.TextSize)  {

	return
}


func (this *Apple2Display) Reconnect(ip string)  {

	return
}


func (this *Apple2Display) HColorAt(x, y int) int {
	var i int

	return i
}

