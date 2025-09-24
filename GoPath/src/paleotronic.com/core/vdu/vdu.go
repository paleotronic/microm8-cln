package vdu

import (
	"paleotronic.com/fmt"
	"math"
	"time"

	"paleotronic.com/log"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/runestring"
	"paleotronic.com/files"
	"paleotronic.com/core"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/hardware"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/utils"
)

const (
	PADDLE_PRESS_TIME int64 = 3000
	BUFFER_CHARS      int   = 16
	KeyBufferPurgeMS  int64 = 3000
)

type VDUCore struct {
	UseClassicHGR     bool
	SaveVideoMode     types.VideoMode
	SaveTextMemory    []uint
	SaveCursorX       int
	SaveCursorY       int
	SaveAttribute     types.VideoAttribute
	ColorFlip         bool
	LastKeyBufferRead int64
	BGColour          int
	CGColour          int
	ColourMem         []int
	TextMemory        *types.TXMemoryBuffer
	ShadowTextMemory  *types.TXMemoryBuffer
	VideoModes        types.VideoModeList
	PenDown           bool
	Echo              bool
	PaddleButtons     []bool
	CharMem           []rune
	CharacterCapture  string
	OutChannel        string
	PaddleVar         types.Variable
	TabWidth          int
	LastZ             int
	Scale             float64
	Prompt            string
	CursorVisible     bool
	Attribute         types.VideoAttribute
	LastGraphicsMode  *types.VideoMode
	Speed             int
	LastPaddleTime    []int64
	PaddleModifier    []float32
	CursorY           int
	LastX             int
	VideoMode         types.VideoMode
	DisplayPage       int
	LastC             int
	Memory            []int
	FGColour          int
	PaddleValues      []int
	Window            types.TextWindow
	LastButtonPress   []int64
	LastY             int
	AttrMem           []types.VideoAttribute
	CurrentPage       int
	CursorX           int
	Buffer            runestring.RuneString
	FeedBuffer        string
	OutputBuffer      string
	SuppressFormat    bool
	DosBuffer         string
	DosCommand        bool
	NextByteColor     bool
	MouseKeys         bool
	LastChar          rune
	BitmapMemory      []*hires.IndexedVideoBuffer
	WozBitmapMemory   []*hires.HGRScreen
	Mossy             *mos6502.Core6502
	NoBreak           bool
	TextMode          types.TextSize
	SaveTextMode      types.TextSize
	Specification     hardware.LayerBundle
}

type VDU struct {
	VDUCore
}

func (this *VDUCore) GetFGColour() int {
	return this.FGColour
}

func (this *VDUCore) SetBreak(v bool) {
	this.NoBreak = !v
}

func (this *VDUCore) Breakable() bool {
	return !this.NoBreak
}

func (this *VDUCore) StartSound() {
	// nothing
}

func (this *VDUCore) Beep() {
	// nothing
}

func (this *VDUCore) ClrHome() {

	/* vars */

	this.Clear()
	this.Home()
	//this.Reposition();

	//log.Panicln("Intended panic")

}

func (this *VDUCore) ClrHomeFull() {

	/* vars */

	this.ClearFull()
	this.Home()
	//this.Reposition();

	//log.Panicln("Intended panic")

}

func (this *VDUCore) Reposition() {
}

func (this *VDUCore) HandleEvent(e types.Event) {

	if e.Name == "VARCHANGE" {
		if e.Target == "SPEED" {
			this.Speed = e.IntParam
			// writeln("DEBUG: VDU Speed Change to ", e.IntParam);
		}
	}

}

func (this *VDUCore) PassWaveBuffer(data []float32) {

}

func (this *VDUCore) ExecNative(mem []int, a int, x int, y int, pc int, sr int, sp int, vdu interfaces.Display) {
	//
	if this.Mossy == nil {
		this.Mossy = mos6502.NewCore6502(mem, a, x, y, pc, sr, sp, vdu)
	}
	this.Mossy.PC = pc
	this.Mossy.P = sr
	this.Mossy.SP = sp
	this.Mossy.A = a
	this.Mossy.X = x
	this.Mossy.Y = y
	this.Mossy.ExecTillHalted()
}

// Shims here for cool things
func (this *VDUCore) SetMemoryValue(x, y int) bool {

	var z int

	result := true

	switch {
	case 8192 <= x && x < 16384:
		this.WozBitmapMemory[0].Poke(x-8192, byte(y&0xff), (this.VideoMode.ActualRows == this.VideoMode.Rows))
		if this.VideoMode.ActualRows == this.VideoMode.Rows {
			result = false
		}
	case 16384 <= x && x < 24576:
		this.WozBitmapMemory[1].Poke(x-16384, byte(y&0xff), (this.VideoMode.ActualRows == this.VideoMode.Rows))
		if this.VideoMode.ActualRows == this.VideoMode.Rows {
			result = false
		}
	case 1024 <= x && x < 3072:
		this.TextMemory.SetValue(x-1024, uint(y))
	default:
		switch x { /* FIXME - Switch statement needs cleanup */
		case 36:
			{
				this.SetCursorX(y % this.GetVideoMode().Columns)
				if this.GetCursorX() < this.GetWindow().Left {
					this.SetCursorX(this.GetWindow().Left)
				}
				if this.GetCursorX() > this.GetWindow().Right {
					this.SetCursorX(this.GetWindow().Right)
				}
				break
			}
		case 37:
			{
				this.SetCursorY(y % this.GetVideoMode().Rows)
				if this.GetCursorY() < this.GetWindow().Top {
					this.SetCursorY(this.GetWindow().Top)
				}
				if this.GetCursorY() > this.GetWindow().Bottom {
					this.SetCursorY(this.GetWindow().Bottom)
				}
				break
			}
		case 50:
			{
				switch y { /* FIXME - Switch statement needs cleanup */
				case 255:
					this.SetAttribute(types.VA_NORMAL)
					break
				case 63:
					this.SetAttribute(types.VA_INVERSE)
					break
				case 127:
					this.SetAttribute(types.VA_BLINK)
					break
				}
				break
			}
		case 51:
			this.SetPrompt(string(rune(y)))
			break
			//49168:  {
			//        this.Buffer = "";
			//}
		case 230:
			{
				switch y { /* FIXME - Switch statement needs cleanup */
				case 32:
					this.SetCurrentPage(0)
					break
				case 64:
					this.SetCurrentPage(1)
					break
				}
				break
			}
		case 234:
			{
				//AppleHiRES.CollisionCount = y
			}
		case 49236:
			{
				this.SetDisplayPage(0) // PAGE 1
				break
			}
		case 49237:
			{
				this.SetDisplayPage(1) // PAGE 2
				break
			}
		case 49323:
			{
				// GFX;
				if this.GetVideoMode().ActualRows == this.GetVideoMode().Rows {
					if this.GetLastGraphicsMode() == nil {
						this.SetVideoMode(this.GetVideoModes()[6])
					} else {
						this.SetVideoMode(*this.GetLastGraphicsMode())
					}
				}
				break
			}
		case 49232:
			{
				// GFX;
				if this.GetVideoMode().ActualRows == this.GetVideoMode().Rows {
					if this.GetLastGraphicsMode() == nil {
						//System.Err.Println("NULL");
						//fmt.Println("NIL")
						this.SetVideoMode(this.GetVideoModes()[6])
					} else {
						//System.Err.Println("MODE");
						//fmt.Println("MODE")
						this.SetVideoMode(*this.GetLastGraphicsMode())
					}
				}
				break
			}
		case 49233:
			{
				// TEXT;
				this.SetVideoMode(this.GetVideoModes()[5])
				break
			}
		case 49234:
			{
				// page full screen;
				z = this.CurrentMode()
				switch z { /* FIXME - Switch statement needs cleanup */
				case 6:
					this.SetVideoMode(this.GetVideoModes()[z+1])
					this.GfxClearSplit(0)
					break
				case 8:
					this.SetVideoMode(this.GetVideoModes()[z+1])
					break
				}
				break
			}
		case 49235:
			{
				// page full screen;
				z = this.CurrentMode()
				switch z { /* FIXME - Switch statement needs cleanup */
				case 7:
					this.SetVideoMode(this.GetVideoModes()[z-1])
					break
				case 9:
					this.SetVideoMode(this.GetVideoModes()[z-1])
					break
				}
				break
			}
		case 49238:
			{
				// Low res;
				z = this.CurrentMode()
				//System.Err.Println("Current mode is "+z);
				switch z { /* FIXME - Switch statement needs cleanup */
				case 8:
					this.SetVideoMode(this.GetVideoModes()[z-2])
					break
				case 9:
					this.SetVideoMode(this.GetVideoModes()[z-2])
					break
				}
				break
			}
		case 49239:
			{
				// Hi res;
				z = this.CurrentMode()
				switch z { /* FIXME - Switch statement needs cleanup */
				case 6:
					this.SetVideoMode(this.GetVideoModes()[z+2])
					break
				case 7:
					this.SetVideoMode(this.GetVideoModes()[z+2])
					break
				}
				break
			}
		default:
			{
				result = false
			}
		}
	}

	if result == true {
		//////fmt.Printf("Memory update mapped in VDU :)")
	}

	return result

}

func (this *VDUCore) GetMemoryValue(addr int) (int, bool) {

	r := -1
	result := true

	switch {
	case 8192 <= addr && addr < 16384:
		r = int(this.WozBitmapMemory[0].Data[addr-8192])
	case 16384 <= addr && addr < 24576:
		r = int(this.WozBitmapMemory[1].Data[addr-16384])
	case 1024 <= addr && addr < 3072:
		r = int(this.TextMemory.GetValue(addr - 1024))
	default:
		result = false
	}

	return r, result
}

func (this *VDUCore) SetColorFlip(b bool) {
	this.ColorFlip = b
}

func (this *VDUCore) GetBGColour() int {
	return this.BGColour
}

func (this *VDUCore) RealPut(ch rune) {

	/* vars */
	var i int
	var ox int
	var oy int

	if runestring.Pos(rune(3), this.Buffer) > 0 && this.Breakable() {
		this.Speed = 255
	}

	ox = this.CursorX
	oy = this.CursorY

	//System.out.println("Screen memory = $"+Integer.toHexString(addr));

	//ch = (char) translateChar(ch);

	if ch == 13 {

		if (this.DosCommand) && (len(this.DosBuffer) > 0) {
			//doscommand = false;
			this.DosCommand = false
			interp := core.GetInstance()
			ent := interp.GetInterpreters()[interp.GetContext()]

			for ent.GetChild() != nil {
				ent = ent.GetChild()
			}

			//System.out.println("Received dos command: "+dosbuffer);
			//System.out.println("Current work directory: "+ent.WorkDir);
			s := this.DosBuffer
			this.DosBuffer = ""
			e := ent.GetDialect().ProcessDynamicCommand(ent, s)
			if e != nil {

				if !ent.HandleError() {
					ent.GetDialect().HandleException(ent, e)
				}

			}
			this.DosBuffer = ""
			return
		}
	}

	if this.NextByteColor {
		this.FGColour = (int(ch) & 0x0f)
		this.NextByteColor = false
		return
	}

	if this.DosCommand {
		if ch == 10 {
			this.DosCommand = false
			return
		}
		this.DosBuffer = this.DosBuffer + string(ch)
		return
	}

	// handle output
	if this.OutChannel != "" && (!this.DosCommand) {
		// send it out here..
		if ch == 4 {
			this.DosCommand = true
			this.DosBuffer = ""
			this.DosBuffer = ""
			return
		}

		this.CharacterCapture = this.CharacterCapture + string(ch)
		if ch == 10 {
			ss := []string{this.CharacterCapture}

			_ = utils.AppendTextFile(this.OutChannel, ss)
			this.CharacterCapture = ""
		}
		return
	}

	this.Reposition()
	switch { /* FIXME - Switch statement needs cleanup */
	case ch == 4:
		this.DosBuffer = ""
		this.DosCommand = true
		break
	case ch == 6:
		this.NextByteColor = true
		break
	case ch == 11:
		this.ClearToBottom()
		break
	case ch == 12:
		this.ClrHome()
		break
	case ch == 13:
		this.LF()
		this.CR()
		//System.out.print ( "\r" );
		break
	case ch == 10:
		if this.LastChar != '\r' {
			this.LF()
		}
		//System.out.print ( "\n" );
		break
	case ch == 14:
		this.Attribute = types.VA_NORMAL
		break
	case ch == 15:
		this.Attribute = types.VA_INVERSE
		break
	case ch == 17:
		this.SetVideoMode(this.VideoModes[5])
		break
	case ch == 18:
		this.SetVideoMode(this.VideoModes[0])
		break
	case ch == 25:
		this.ClrHome()
		break
	case ch == 26:
		this.ClearLine()
		break
	case ch == 28:
		this.CursorRight()
		break
	case ch == 29:
		this.ClearToEOL()
		break
	case ch == '\t':
		this.TAB()
		break
	case ch == 8:
		this.Backspace()
		break
	case ch == 136:
		this.Backspace()
		break
	case ch == 7:
		this.Beep()
		break
	case ch == 135:
		this.Beep()
		break
	case ch == 27:
		{
			if this.Attribute == types.VA_INVERSE {
				this.MouseKeys = true
				//writeln("MOUSE KEYS IS == ",this.MouseKeys);
			}
			break
		}
	case ch == 24:
		{
			if this.Attribute == types.VA_INVERSE {
				this.MouseKeys = false
			}
			//writeln("MOUSE KEYS IS == ",this.MouseKeys);
			break
		}
	case ch >= vduconst.COLOR0 && ch <= vduconst.COLOR15:
		{
			this.SetFGColour(int(ch - vduconst.COLOR0))
		}
	case ch >= vduconst.BGCOLOR0 && ch <= vduconst.BGCOLOR15:
		{
			this.SetBGColour(int(ch - vduconst.BGCOLOR0))
		}
	case ch == vduconst.INVERSE_ON:
		{
			this.ColorFlip = !this.ColorFlip
		}
	default:
		{
			os := this.XYToOffset(ox, oy)

			oa := this.Attribute
			if this.SuppressFormat {
				this.Attribute = types.VA_NORMAL
			}

			this.TextMemory.SetValue(os, uint(this.AsciiToPoke(int(ch))|((this.FGColour&255)<<16)))

			//laddr := ox + (oy * this.VideoMode.Columns)

			cx := (this.FGColour & 0xf) | ((this.BGColour & 0xf) << 4)
			if this.ColorFlip {
				cx = (this.BGColour & 0xf) | ((this.FGColour & 0xf) << 4)
			}

			//this.ShadowTextMemory.SetValue(laddr, this.AsciiToPoke(int(ch))|((cx&255)<<16))

			this.VideoMode.PutXYMemory(this.ShadowTextMemory, ox, oy, uint(this.AsciiToPoke(int(ch))), uint(cx), this.GetTextMode())

			if this.SuppressFormat {
				this.Attribute = oa
			}
			//System.out.println("Screen memory = $"+Integer.toHexString(1024+os)+", "+ch);

			//System.out.print(ch);
			//System.err.println("cx = "+CursorX+", cy = "+CursorY);
			//renderBufferUpdate( this.TextMemory, os, this.CursorX, this.CursorY );

			i = this.CursorX + (this.CursorY * this.VideoMode.Columns)
			if i < len(this.CharMem) {
				this.CharMem[i] = ch
				this.ColourMem[i] = this.FGColour + 16*this.BGColour
				if !this.MouseKeys {
					this.AttrMem[i] = this.Attribute
				} else {
					this.AttrMem[i] = types.VA_NORMAL
				}
				// move cursor right;
				this.CursorRight()
			}
			break
		}
	}

	this.LastChar = ch

}

func (this *VDUCore) Backspace() {

	/* vars */

	for this.HasOutput() {
		this.DoOutput()
	}

	//this.CursorLeft();
	//this.RealPut(' ');
	this.CursorLeft()
	//this.Reposition();

}

func (this *VDUCore) CursorUp() {

	/* vars */

	if this.CursorY > this.Window.Top {
		//this.CursorY = this.CursorY - this.VideoMode.VAdvance(this.TextMode)
		this.CursorY -= 1
	}
	this.Reposition()

}

func (this *VDUCore) ClearToEOL() {

	/* vars */
	var i int
	var r int
	var c int

	r = this.CursorY
	c = this.CursorX

	for c = this.CursorX; c <= this.Window.Right; c++ {
		i = c + (this.VideoMode.Columns * r)

		if i < len(this.CharMem) {
			this.CharMem[i] = ' '
			this.ColourMem[i] = this.FGColour + 16*this.BGColour
			this.AttrMem[i] = this.Attribute
		}

		to := this.XYToOffset(c, r)
		this.TextMemory.SetValue(to, uint(160|(this.FGColour<<16)|(this.BGColour<<20)))

		//laddr := c + (r * this.VideoMode.Columns)
		//this.ShadowTextMemory.SetValue(laddr, 160|(this.FGColour<<16)|(this.BGColour<<20))
		cx := (this.BGColour << 4) | this.FGColour

		this.VideoMode.PutXYMemory(this.ShadowTextMemory, c, r, 160, uint(cx), this.GetTextMode())

	}

}

func (this *VDUCore) PutStr(s string) error {
	if s == "" {
		return nil
	}

	if s == "null" {
		return exception.NewESyntaxError("NULL TO OUTPUT ERROR")
	}
	/* vars */
	//char ch;

	//this.Reposition();
	for _, ch := range s {
		this.Put(ch)
	}

	//System.Err.Println("PUTSTR ["+s+"]");

	//this.DoOutput();
	return nil
}

func (this *VDUCore) ScrollGFX() {

	/* vars */
	var r int
	var c int

	for r = 1; r < this.VideoMode.Rows; r++ {
		for c = 0; c < this.VideoMode.Columns; c++ {
			//this.CharMem[((r-1)*this.VideoMode.Columns)+c] = this.CharMem[((r)*this.VideoMode.Columns)+c]
			//this.ColourMem[((r-1)*this.VideoMode.Columns)+c] = this.ColourMem[((r)*this.VideoMode.Columns)+c]
			//this.AttrMem[((r-1)*this.VideoMode.Columns)+c] = this.AttrMem[((r)*this.VideoMode.Columns)+c]
			to := this.XYToOffset(c, r-1)
			from := this.XYToOffset(c, r)
			this.TextMemory.SetValue(to, this.TextMemory.GetValue(from))

			laddr := this.VideoMode.LinearOffsetXY(c, r-1)
			laddrf := this.VideoMode.LinearOffsetXY(c, r)
			this.ShadowTextMemory.SetValue(laddr, this.ShadowTextMemory.GetValue(laddrf))

		}
	}

	fillChar := 0

	//r = this.Window.Bottom - (this.VideoMode.VAdvance(this.TextMode) - 1)
	r = this.VideoMode.Rows - 1
	//for r = this.Window.Bottom - (this.VideoMode.VAdvance(this.TextMode) - 1); r <= this.Window.Bottom; r++ {

	for c = 0; c < this.VideoMode.Columns; c++ {
		//this.CharMem[((r)*this.VideoMode.Columns)+c] = ' '
		//this.ColourMem[((r)*this.VideoMode.Columns)+c] = this.FGColour + 16*this.BGColour
		//this.AttrMem[((r)*this.VideoMode.Columns)+c] = this.Attribute

		to := this.XYToOffset(c, r)

		cx := (this.BGColour << 4) | this.FGColour

		this.VideoMode.PutXYMemory(this.ShadowTextMemory, c, r, uint(fillChar), uint(cx), this.GetTextMode())

		if to >= 0 {
			this.TextMemory.SetValue(to, uint(fillChar|(this.FGColour<<16)|(this.BGColour<<20)))
			this.VideoMode.PutXYMemory(this.ShadowTextMemory, c, r, uint(fillChar), uint(cx), this.GetTextMode())
		}
	}

	//}

}

func (this *VDUCore) GetStrings() {

}

func (this *VDUCore) SetPaddleButton(index int, value bool) {
	this.LastButtonPress[index%4] = (time.Now().UnixNano() / 1000000)
	this.PaddleButtons[index%4] = value
}

func (this *VDUCore) CursorLeft() {

	/* vars */

	//this.CursorX = this.CursorX - this.VideoMode.HAdvance(this.TextMode)
	this.CursorX -= 1
	if this.CursorX < this.Window.Left {
		this.CursorX = this.Window.Right
		this.CursorUp()
	}
	this.Reposition()

}

func (this *VDUCore) HasOutput() bool {

	/* vars */
	var result bool

	if this.OutputBuffer == "" {
		this.OutputBuffer = ""
	}

	result = (len(this.OutputBuffer) > 0)

	/* enforce non void return */
	return result

}

func (this *VDUCore) GetKey() rune {

	/* vars */
	var result rune

	result = 0
	if len(this.Buffer.Runes) > 0 {
		result = this.Buffer.Runes[0]
		this.Buffer = runestring.Delete(this.Buffer, 1, 1)
	}

	/* enforce non void return */
	return result

}

func (this *VDUCore) SetPaddleModifier(index int, value float32) {
	this.LastPaddleTime[index%4] = time.Now().UnixNano() / 1000000
	this.PaddleModifier[index%4] = value
}

func (this *VDUCore) CurrentMode() int {

	/* vars */
	var result int
	//VideoMode vm;
	var i int

	result = -1
	i = 0
	for _, vm := range this.VideoModes {
		if vm.Equals(&this.VideoMode) {
			result = i
			return result
		}
		i++
	}

	/* enforce non void return */
	return result

}

func (this *VDUCore) LF() {

	/* vars */
	//this.Flush();

	this.CursorDown()
	this.Reposition()

}

func (this *VDUCore) PollPaddleButtons() {
	for i := 0; i < len(this.PaddleButtons); i++ {
		if ((time.Now().UnixNano()/1000000)-this.LastButtonPress[i] > PADDLE_PRESS_TIME) && (this.PaddleButtons[i]) {
			this.PaddleButtons[i] = false
		}
	}
}

func (this *VDUCore) Normal() {

	/* vars */

	//TextColor(lightgray);

}

func (this *VDUCore) SetTextMode(tm types.TextSize) {
	this.TextMode = tm
}

func (this *VDUCore) GetTextMode() types.TextSize {
	return this.TextMode
}

func (this *VDUCore) SetVideoMode(vm types.VideoMode) {

	/* vars */
	var size int

	this.ColorFlip = false

	this.VideoMode = vm

	this.CurrentPage = 0
	this.DisplayPage = 0

	this.Window.Top = vm.DefaultWindow.Top
	this.Window.Bottom = vm.DefaultWindow.Bottom
	this.Window.Left = vm.DefaultWindow.Left
	this.Window.Right = vm.DefaultWindow.Right

	this.TextMode = vm.GetDefaultTW() // Establish text mode bits per vm

	size = vm.Rows * vm.Columns // always rows * columns no matter what

	rcol := this.ColourMem
	rchar := this.CharMem
	rattr := this.AttrMem

	this.ColourMem = make([]int, size)
	this.CharMem = make([]rune, size)
	this.AttrMem = make([]types.VideoAttribute, size)

	if rcol != nil {
		for ii := 0; ii < int(math.Min(float64(len(rcol)), float64(size))); ii++ {
			this.ColourMem[ii] = rcol[ii]
			this.CharMem[ii] = rchar[ii]
			this.AttrMem[ii] = rattr[ii]
		}
	}

	// reset colour maybe;
	this.FGColour = 15
	this.BGColour = 0

	this.LastX = vm.Width / 2
	this.LastY = vm.Height / 2
	this.LastZ = 0

	this.CursorX = this.Window.Left

	if this.CursorY < this.Window.Top {
		this.CursorY = this.Window.Top
	}

	// save mode for reference later on
	if vm.ActualRows != vm.Rows {
		this.LastGraphicsMode = &vm
		if vm.Width > 50 {
			this.Window.Top = 0
			this.Window.Bottom = vm.Rows - 1
			this.Window.Left = 0
			this.Window.Right = vm.Columns - 1
		}
	}

}

func (this *VDUCore) InitDisplay(vm types.VideoMode) {
	this.Scale = 1.0
	this.PenDown = true
	this.Echo = false
	this.Window = *types.NewTextWindow()
	this.SetVideoMode(vm)
	this.FGColour = 15
	this.BGColour = 0
	//this.Clear();
	this.ClrHome()
	this.Reposition()
	this.LastC = 7
	this.Buffer = runestring.NewRuneString()
	this.Prompt = "]"
	this.Attribute = types.VA_NORMAL
	this.Speed = 255
	this.DisplayPage = 0
	this.CurrentPage = 0
	this.PaddleValues = make([]int, 4)
	this.PaddleValues[0] = 128
	this.PaddleValues[1] = 128
	this.PaddleValues[2] = 128
	this.PaddleValues[3] = 128
	this.PaddleModifier = make([]float32, 4)
	this.PaddleModifier[0] = 0
	this.PaddleModifier[1] = 0
	this.PaddleModifier[2] = 0
	this.PaddleModifier[3] = 0
	this.PaddleButtons = make([]bool, 4)
	this.PaddleButtons[0] = false
	this.PaddleButtons[1] = false
	this.PaddleButtons[2] = false
	this.PaddleButtons[3] = false
	this.LastPaddleTime = make([]int64, 4)
	this.OutputBuffer = ""
	this.Buffer = runestring.NewRuneString()
}

func (this *VDUCore) HColorAt(x, y int) int {
	return this.WozBitmapMemory[this.CurrentPage%2].ColorAt(x, y)
}

func (this *VDUCore) ColorAt(x, y int) int {
	cy := y / 2
	cx := x
	addr := this.XYToOffset(cx, cy)

	//System.out.println("Checking memory at "+(1024+addr));

	if (addr < 0) || (addr >= this.TextMemory.Size()) {
		addr = 0
	}

	v := this.TextMemory.GetValue(addr)
	if y%2 == 0 {
		// low nibble
		return int(v & 15)
	} else {
		// hi nibble
		return int(v & 240) / 16
	}
}

func (this *VDUCore) GfxClear(idx int) {

	if this.VideoMode.Width == 40 {
		for r := 0; r < (this.VideoMode.Height / 2); r++ {
			for c := 0; c < this.VideoMode.Width; c++ {
				addr := this.XYToOffset(c, r)
				this.TextMemory.SetValue(addr, uint(0|(this.FGColour<<16)))

				//laddr := c + (r * this.VideoMode.Columns)
				//this.ShadowTextMemory.SetValue(laddr, 0|(this.FGColour<<16))
				cx := (this.BGColour << 4) | this.FGColour

				this.VideoMode.PutXYMemory(this.ShadowTextMemory, c, r, 0, uint(cx), this.GetTextMode())
			}
		}
	}

}

func (this *VDUCore) CR() {

	/* vars */
	//this.Flush();

	this.CursorX = this.Window.Left
	this.Reposition()

}

func (this *VDUCore) Swap(a *int, b *int) {
	var c *int
	c = a
	a = b
	b = c
}

func (this *VDUCore) Bold() {

	/* vars */

	//TextColor(lightblue);

}

func (this *VDUCore) Render() {

	/* vars */

	/* This is a stub method for the graphical class */

}

func (this *VDUCore) CursorRight() {

	/* vars */

	//this.CursorX = this.CursorX + this.VideoMode.HAdvance(this.TextMode)
	this.CursorX += 1
	if this.CursorX > this.Window.Right {
		this.CursorX = this.Window.Left
		this.CursorDown()
	}
	this.Reposition()

}

func (this *VDUCore) Home() {

	/* vars */
	this.Flush()

	this.CursorX = this.Window.Left
	this.CursorY = this.Window.Top
	//this.Reposition();

}

func (this *VDUCore) GetString() runestring.RuneString {

	/* vars */
	var result runestring.RuneString

	this.Reposition()
	//this.PutStr(prompt);
	//this.Reposition();
	result = runestring.NewRuneString()

	if runestring.Pos('\r', this.Buffer) > 0 {

		result = runestring.Copy(this.Buffer, 1, runestring.Pos('\r', this.Buffer)-1)
		this.Buffer = runestring.Delete(this.Buffer, 1, len(result.Runes)+1)
		this.PutStr("\r\n")

	}

	/* enforce non void return */
	return result

}

func (this *VDUCore) TAB() {

	/* vars */
	// this.Flush();

	//this.CursorRight();
	this.RealPut(' ')
	for ((this.CursorX - this.Window.Left) % this.TabWidth) != 0 {
		this.RealPut(' ')
	}
	//this.CursorRight();

}

func (this *VDUCore) DoPrompt() {

	/* vars */

	//this.Flush();
	this.PutStr(this.Prompt)

}

func (this *VDUCore) SetDisplayPage(i int) {

	/* vars */
	this.DisplayPage = (i % this.VideoMode.MaxPages)

}

func Between(v, lo, hi int) bool {
	return ((v >= lo) && (v <= hi))
}

func (this *VDUCore) PokeToAttribute(v int) types.VideoAttribute {

	v = v & 1023

	va := types.VA_INVERSE
	if (v & 64) > 0 {
		va = types.VA_BLINK
	}
	if (v & 128) > 0 {
		va = types.VA_NORMAL
	}
	if (v & 256) > 0 {
		va = types.VA_NORMAL
	}
	return va
}

func (this *VDUCore) AsciiToPoke(v int) int {

	highbit := v & 1024
	v = v & 1023

	if v > 255 {
		return v | highbit
	}

	v = (v & 127)

	if this.Attribute == types.VA_NORMAL {
		return (v + 128) | highbit
	}

	if this.Attribute == types.VA_INVERSE {
		if (v >= ' ') && (v < '@') {
			return v | highbit
		} else if (v >= 96) && (v <= 127) {
			return (v - 64) | highbit
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

func (this *VDUCore) PokeToAscii(v int) int {

	highbit := v & 1024

	v = v & 1023

	if Between(v, 0, 31) {
		return (64 + (v % 32)) | highbit
	}

	if Between(v, 32, 63) {
		return (32 + (v % 32)) | highbit
	}

	if Between(v, 64, 95) {
		return (64 + (v % 32)) | highbit
	}

	if Between(v, 96, 127) {
		return (32 + (v % 32)) | highbit
	}

	if Between(v, 128, 159) {
		return (64 + (v % 32)) | highbit
	}

	if Between(v, 160, 191) {
		return (32 + (v % 32)) | highbit
	}

	if Between(v, 192, 223) {
		return (64 + (v % 32)) | highbit
	}

	if Between(v, 224, 255) {
		return (96 + (v % 32)) | highbit
	}

	if Between(v, 256, 287) {
		return v | highbit
	}

	return v | highbit
}

func (this *VDUCore) ClearLine() {

	/* vars */
	var i int
	var r int
	var c int

	r = this.CursorY
	c = this.CursorX

	for c = this.Window.Left; c <= this.Window.Right; c++ {
		i = c + (this.VideoMode.Columns * r)
		this.CharMem[i] = ' '
		this.ColourMem[i] = this.FGColour + 16*this.BGColour
		this.AttrMem[i] = this.Attribute

		to := this.XYToOffset(c, r)
		this.TextMemory.SetValue(to, uint(this.AsciiToPoke(' ')|(this.FGColour<<16)|(this.BGColour<<20)))


		cx := (this.BGColour << 4) | this.FGColour

		this.VideoMode.PutXYMemory(this.ShadowTextMemory, c, r, 160, uint(cx), this.GetTextMode())
	}

}

func (this *VDUCore) DoOutput() {

	/* vars */
	var mx int
	var s string
	//char ch;

	/* output an amount of the buffer */
	mx = len(this.OutputBuffer)
	if (this.Speed + 1) < mx {
		mx = this.Speed + 1
	}

	s = utils.Copy(this.OutputBuffer, 1, mx)
	this.OutputBuffer = utils.Delete(this.OutputBuffer, 1, mx)

	for _, ch := range s {
		this.RealPut(ch)
	}

	//writeln("Flushed ",mx," characters");

}

func (this *VDUCore) CursorDown() {

	/* vars */

	//this.CursorY = this.CursorY + this.VideoMode.VAdvance(this.TextMode)
	this.CursorY += 1
	if this.CursorY > this.Window.Bottom {
		/* scroll display up one line */
		this.Scroll()
		this.CursorY = this.Window.Bottom
	}
	this.Reposition()

}

func (this *VDUCore) CharAt(x int, y int) int {

	/* vars */
	var result int
	var ch rune

	result = 0
	if x < 0 {
		return result
	}
	if x >= this.VideoMode.Columns {
		return result
	}
	if y < 0 {
		return result
	}
	if y >= this.VideoMode.Rows {
		return result
	}

	ch = this.CharMem[x+(this.VideoMode.Columns*y)]
	if ch != ' ' {
		result = 15
	}

	/* enforce non void return */
	return result

}

func (this *VDUCore) ProcessAudio() {
	// stubbed
}

func (this *VDUCore) Scroll() {

	/* vars */
	var r int
	var c int

	for r = this.Window.Top + 1; r <= this.Window.Bottom; r++ {
		for c = this.Window.Left; c <= this.Window.Right; c++ {
			//this.CharMem[((r-1)*this.VideoMode.Columns)+c] = this.CharMem[((r)*this.VideoMode.Columns)+c]
			//this.ColourMem[((r-1)*this.VideoMode.Columns)+c] = this.ColourMem[((r)*this.VideoMode.Columns)+c]
			//this.AttrMem[((r-1)*this.VideoMode.Columns)+c] = this.AttrMem[((r)*this.VideoMode.Columns)+c]
			to := this.XYToOffset(c, r-1)
			from := this.XYToOffset(c, r)
			this.TextMemory.SetValue(to, this.TextMemory.GetValue(from))

			laddr := this.VideoMode.LinearOffsetXY(c, r-1)
			laddrf := this.VideoMode.LinearOffsetXY(c, r)
			this.ShadowTextMemory.SetValue(laddr, this.ShadowTextMemory.GetValue(laddrf))

		}
	}

	fillChar := 160

	//r = this.Window.Bottom - (this.VideoMode.VAdvance(this.TextMode) - 1)

	for r = this.Window.Bottom; r <= this.Window.Bottom; r++ {

		for c = this.Window.Left; c <= this.Window.Right; c++ {
			//this.CharMem[((r)*this.VideoMode.Columns)+c] = ' '
			//this.ColourMem[((r)*this.VideoMode.Columns)+c] = this.FGColour + 16*this.BGColour
			//this.AttrMem[((r)*this.VideoMode.Columns)+c] = this.Attribute

			to := this.XYToOffset(c, r)

			cx := (this.BGColour << 4) | this.FGColour

			this.VideoMode.PutXYMemory(this.ShadowTextMemory, c, r, 160, uint(cx), this.GetTextMode())

			if to >= 0 {
				this.TextMemory.SetValue(to, uint(fillChar|(this.FGColour<<16)|(this.BGColour<<20)))
				this.VideoMode.PutXYMemory(this.ShadowTextMemory, c, r, uint(fillChar), uint(cx), this.GetTextMode())
			}
		}

	}

}

func (this *VDUCore) WindowOn() {

	/* vars */

	//window(1,1,this.VideoMode.Columns,this.VideoMode.ActualRows);

}

func (this *VDUCore) SetPage(i int) {

	/* vars */

	this.CurrentPage = (i % this.VideoMode.MaxPages)

}

func (this *VDUCore) UpdatePaddleValues() {
	for i := 0; i < 4; i++ {

		//this.PaddleValues[i] = this.PaddleValues[i] + Paddle.ValueForFloat(this.PaddleModifier[i])
		if this.PaddleValues[i] > 255 {
			this.PaddleValues[i] = 255
		}
		if this.PaddleValues[i] < 0 {
			this.PaddleValues[i] = 0
		}
	}
}

func (this *VDUCore) SetCursorVisible(v bool) {
	this.CursorVisible = v
}

func (this *VDUCore) CtrlBreak() bool {

	/* vars */
	var result bool

	result = false

	/* enforce non void return */
	return result

}

func (this *VDUCore) PaddleButton(i int) {
	idx := (i % 4)

	this.PaddleButtons[idx] = true
}

func (this *VDUCore) IsCursorVisible() bool {
	return this.CursorVisible
}

func NewVDU(vm *types.VideoMode) *VDU {

	/* vars */
	this := &VDU{}

	avp := *types.NewVideoPalette()
	avp.Add(types.NewVideoColor(0, 0, 0, 255))       // blk
	avp.Add(types.NewVideoColor(144, 23, 64, 255))   // dk red
	avp.Add(types.NewVideoColor(64, 44, 165, 255))   // dk blue
	avp.Add(types.NewVideoColor(208, 67, 229, 255))  // magenta
	avp.Add(types.NewVideoColor(0, 105, 64, 255))    // dk green blue
	avp.Add(types.NewVideoColor(128, 128, 128, 255)) // mid gray
	avp.Add(types.NewVideoColor(47, 149, 229, 255))  // cyan
	avp.Add(types.NewVideoColor(191, 171, 255, 255)) //lt blue
	avp.Add(types.NewVideoColor(64, 36, 0, 255))     // dk brown
	avp.Add(types.NewVideoColor(208, 106, 26, 255))  // orange
	avp.Add(types.NewVideoColor(128, 128, 128, 255)) // mid gray
	avp.Add(types.NewVideoColor(255, 150, 191, 255)) // lt red
	avp.Add(types.NewVideoColor(47, 188, 26, 255))   // green
	avp.Add(types.NewVideoColor(191, 211, 90, 255))  // green yellow
	avp.Add(types.NewVideoColor(111, 232, 191, 255)) // light cyan
	avp.Add(types.NewVideoColor(255, 255, 255, 255))

	havp := *types.NewVideoPalette()

	havp.Add(types.NewVideoColor(0, 0, 0, 255))       // blk
	havp.Add(types.NewVideoColor(144, 23, 64, 255))   // dk red
	havp.Add(types.NewVideoColor(64, 44, 165, 255))   // dk blue
	havp.Add(types.NewVideoColor(208, 67, 229, 255))  // magenta
	havp.Add(types.NewVideoColor(0, 105, 64, 255))    // dk green blue
	havp.Add(types.NewVideoColor(128, 128, 128, 255)) // mid gray
	havp.Add(types.NewVideoColor(47, 149, 229, 255))  // cyan
	havp.Add(types.NewVideoColor(191, 171, 255, 255)) //lt blue
	havp.Add(types.NewVideoColor(64, 36, 0, 255))     // dk brown
	havp.Add(types.NewVideoColor(208, 106, 26, 255))  // orange
	havp.Add(types.NewVideoColor(128, 128, 128, 255)) // mid gray
	havp.Add(types.NewVideoColor(255, 150, 191, 255)) // lt red
	havp.Add(types.NewVideoColor(47, 188, 26, 255))   // green
	havp.Add(types.NewVideoColor(191, 211, 90, 255))  // green yellow
	havp.Add(types.NewVideoColor(111, 232, 191, 255)) // light cyan
	havp.Add(types.NewVideoColor(255, 255, 255, 255))

	havp.Add(types.NewVideoColor(0, 0, 0, 255))       // BLACK
	havp.Add(types.NewVideoColor(53, 167, 51, 255))   // GREEN
	havp.Add(types.NewVideoColor(174, 76, 204, 255))  // VIOLET
	havp.Add(types.NewVideoColor(209, 217, 223, 255)) // WHITE
	havp.Add(types.NewVideoColor(18, 26, 32, 255))    // BLACK
	havp.Add(types.NewVideoColor(174, 105, 51, 255))  // ORANGE
	havp.Add(types.NewVideoColor(53, 138, 204, 255))  // BLUE
	havp.Add(types.NewVideoColor(224, 224, 224, 255)) // WHITE

	this.VideoModes = make(types.VideoModeList, 0)
	this.VideoModes.Add(*types.NewVideoMode(280, 192, 24, 80, 24, 1, avp, 1, 1))
	this.VideoModes.Add(*types.NewVideoMode(40, 40, 24, 80, 4, 1, avp, 1, 0.50)) // GR 40 x 40 pixels
	this.VideoModes.Add(*types.NewVideoMode(40, 48, 24, 80, 0, 1, avp, 1, 0.50)) // GR 40 x 48 pixels (fs)
	this.VideoModes.Add(*types.NewVideoMode(280, 160, 24, 80, 4, 1, havp, 2, 1)) // GR 40 x 40 pixels
	this.VideoModes.Add(*types.NewVideoMode(280, 192, 24, 80, 0, 1, havp, 2, 1)) // GR 40 x 48 pixels (fs)

	this.VideoModes.Add(*types.NewVideoMode(280, 192, 24, 40, 24, 1, avp, 1, 1))
	this.VideoModes.Add(*types.NewVideoMode(40, 40, 24, 40, 4, 1, avp, 1, 0.50)) // GR 40 x 40 pixels
	this.VideoModes.Add(*types.NewVideoMode(40, 48, 24, 40, 0, 1, avp, 1, 0.50)) // GR 40 x 48 pixels (fs)
	this.VideoModes.Add(*types.NewVideoMode(280, 160, 24, 40, 4, 1, havp, 2, 1)) // HGR 40 x 40 pixels
	this.VideoModes.Add(*types.NewVideoMode(280, 192, 24, 40, 0, 1, havp, 2, 1)) // HGR 40 x 48 pixels (fs)

	// Super modes
	this.VideoModes.Add(*types.NewVideoMode(280, 192, 48, 80, 48, 1, avp, 1, 1))
	this.VideoModes.Add(*types.NewVideoMode(40, 40, 48, 80, 8, 1, avp, 1, 0.50)) // GR 40 x 40 pixels
	this.VideoModes.Add(*types.NewVideoMode(40, 48, 48, 80, 0, 1, avp, 1, 0.50)) // GR 40 x 48 pixels (fs)
	this.VideoModes.Add(*types.NewVideoMode(280, 160, 48, 80, 8, 1, havp, 2, 1)) // GR 40 x 40 pixels
	this.VideoModes.Add(*types.NewVideoMode(280, 192, 48, 80, 0, 1, havp, 2, 1)) // GR 40 x 48 pixels (fs)

	this.VideoModes.Add(*types.NewVideoMode(280, 192, 48, 40, 48, 1, avp, 1, 1))
	this.VideoModes.Add(*types.NewVideoMode(40, 40, 48, 40, 8, 1, avp, 1, 0.50)) // GR 40 x 40 pixels
	this.VideoModes.Add(*types.NewVideoMode(40, 48, 48, 40, 0, 1, avp, 1, 0.50)) // GR 40 x 48 pixels (fs)
	this.VideoModes.Add(*types.NewVideoMode(280, 160, 48, 40, 8, 1, havp, 2, 1)) // GR 40 x 40 pixels
	this.VideoModes.Add(*types.NewVideoMode(280, 192, 48, 40, 0, 1, havp, 2, 1)) // GR 40 x 48 pixels (fs)

	log.Println("Number of video modes listed = ", len(this.VideoModes))

	this.TextMemory = types.NewTXMemoryBuffer(4096)
	this.TextMemory.Silent(true)
	this.ShadowTextMemory = types.NewTXMemoryBuffer(this.TextMemory.Size())

	this.InitDisplay(this.VideoModes[0])

	for i := 0; i < 4; i++ {
		this.LastPaddleTime[i] = time.Now().UnixNano() / 1000000
		this.PaddleModifier[i] = 1
	}

	// Video Buffers
	this.BitmapMemory = make([]*hires.IndexedVideoBuffer, 2)
	this.BitmapMemory[0] = hires.NewIndexedVideoBuffer(280, 192)
	this.BitmapMemory[1] = hires.NewIndexedVideoBuffer(280, 192)

	// Woz Compatible Buffers
	this.WozBitmapMemory = make([]*hires.HGRScreen, 2)
	this.WozBitmapMemory[0] = &hires.HGRScreen{}
	this.WozBitmapMemory[1] = &hires.HGRScreen{}

	return this

}

func (this *VDUCore) GetCurrentPage() int {
	return this.CurrentPage
}

func (this *VDUCore) SetLastX(x int) {
	this.LastX = x
}

func (this *VDUCore) SetLastY(y int) {
	this.LastY = y
}

func (this *VDUCore) GetBitmapMemory() []*hires.IndexedVideoBuffer {
	return this.BitmapMemory
}

func (this *VDUCore) Click() {
}

func (this *VDUCore) HgrShape(shape hires.ShapeEntry, x int, y int, scl int, deg int, c int, usecol bool) {
	hires.GetAppleHiRES().HgrShape(this, this.GetBitmapMemory()[this.GetCurrentPage()%2], shape, x, y, scl, deg, c, usecol)
}

func (this *VDUCore) HgrFill(hc int) {
	hires.GetAppleHiRES().HgrFill(this.GetBitmapMemory()[this.GetCurrentPage()%2], hc)
}

func (this *VDUCore) HgrPlotHold(x2, y2 int, hc int) {
	hires.GetAppleHiRES().HgrPlot(this.GetBitmapMemory()[this.GetCurrentPage()%2], x2, y2, hc)
}

func (this *VDUCore) HgrPlot(x2, y2 int, hc int) {
	hires.GetAppleHiRES().HgrPlot(this.GetBitmapMemory()[this.GetCurrentPage()%2], x2, y2, hc)
}

func (this *VDUCore) HgrLine(x1, y1, x2, y2 int, hc int) {
	hires.GetAppleHiRES().HgrLine(this.GetBitmapMemory()[this.GetCurrentPage()%2], x1, y1, x2, y2, hc)
}

func (this *VDUCore) GetAttribute() types.VideoAttribute {
	return this.Attribute
}

func (this *VDUCore) GetLastGraphicsMode() *types.VideoMode {
	return this.LastGraphicsMode
}

func (this *VDUCore) Flush() {
	for this.HasOutput() {
		this.DoOutput()
	}
}

func (this *VDUCore) Dump() {

}

func (this *VDUCore) InsertCharToBuffer(ch rune) {

	/* vars */

	if ch == 0 {
		return
	}

	/* fix for android devices */
	//if (ch == '\n') && (Gdx.App.GetType() == ApplicationType.Android) {
	//  ch = '\r'
	//}

	this.Buffer.Append(string(ch))
	//Buffer = "" + ch;

	if len(this.Buffer.Runes) > BUFFER_CHARS {
		this.Buffer = runestring.Delete(this.Buffer, 1, len(this.Buffer.Runes)-BUFFER_CHARS)
	}

	log.Println("Adding char to VDU Buffer - ", ch)

}

func (this *VDUCore) SetSuppressFormat(v bool) {
	this.SuppressFormat = v
}

func (this *VDUCore) SetPaddleValues(z int, v int) {
	this.PaddleValues[z] = v
}

func (this *VDUCore) SetPaddleButtons(z int, v bool) {
	this.PaddleButtons[z] = v
}

func (this *VDUCore) GetMemory() []int {
	return this.Memory
}

func (this *VDUCore) SetMemory(v []int) {
	this.Memory = v
}

func (this *VDUCore) GetOutChannel() string {
	return this.OutChannel
}

func (this *VDUCore) SetOutChannel(v string) {
	this.OutChannel = v
}

func (this *VDUCore) GetFeedBuffer() string {
	return this.FeedBuffer
}

func (this *VDUCore) SetFGColour(v int) {
	this.FGColour = v
}

func (this *VDUCore) SetBGColour(v int) {
	this.BGColour = v
}

func (this *VDUCore) SetFeedBuffer(v string) {
	this.FeedBuffer = v
}

func (this *VDUCore) SetBuffer(s runestring.RuneString) {
	this.Buffer = s
}

func (this *VDUCore) SetCursorX(v int) {
	this.CursorX = v
}

func (this *VDUCore) SetCursorY(v int) {
	this.CursorY = v
}

func (this *VDUCore) GetBuffer() runestring.RuneString {
	return this.Buffer
}

func (this *VDUCore) ProcessKeyBuffer(ent interfaces.Interpretable) {
	if len(this.FeedBuffer) > 0 {

		i := int(this.FeedBuffer[0])

		//if (i == VDU.CSR_UP) i = 11;
		//if (i == VDU.CSR_DOWN) i = 10;
		//if (i == VDU.CSR_LEFT) i = 8;
		//if (i == VDU.CSR_RIGHT) i = 21;

		this.FeedBuffer = utils.Delete(this.FeedBuffer, 1, 1)
		//this.Buffer = "";

		if (i >= 'A') && (i <= 'Z') {
			i = i + 32
		} else if (i >= 'a') && (i <= 'z') {
			i = i - 32
		}

		i = (128 | i)
		ent.SetMemory(49168, 0)
	} else if len(this.Buffer.Runes) > 0 {

		i := int(this.Buffer.Runes[0])

		if i == vduconst.CSR_UP {
			i = 11
		}
		if i == vduconst.CSR_DOWN {
			i = 10
		}
		if i == vduconst.CSR_LEFT {
			i = 8
		}
		if i == vduconst.CSR_RIGHT {
			i = 21
		}

		this.Buffer = runestring.Delete(this.Buffer, 1, 1)
		//this.Buffer = "";

		if (i >= 'A') && (i <= 'Z') {
			i = i + 32
		} else if (i >= 'a') && (i <= 'z') {
			i = i - 32
		}

		i = (128 | i)
		ent.SetMemory(49168, 0)
	}

}

func (this *VDUCore) PurgeKeyBuffer() {
	if KeyBufferPurgeMS < 0 {
		return
	}

	diff := (time.Duration(time.Now().Nanosecond()) / time.Millisecond) - time.Duration(this.LastKeyBufferRead)
	if diff > time.Duration(KeyBufferPurgeMS) {
		//System.err.println("Purging unread keystrokes...");
		this.Buffer = runestring.NewRuneString()
		this.LastKeyBufferRead = int64(time.Duration(time.Now().Nanosecond()) / time.Millisecond)
	}
}

func (this *VDUCore) RegenerateMemory(ent interfaces.Interpretable) {
	ent.SetMemorySilent(32, uint(this.Window.Left))
	ent.SetMemorySilent(33, uint(this.Window.Right+1-this.Window.Left))
	ent.SetMemorySilent(34, uint(this.Window.Top))
	ent.SetMemorySilent(35, uint(this.Window.Bottom+1))
}

func (this *VDUCore) RegenerateWindow(memory []int) {
	left := memory[32]   // 0-79
	width := memory[33]  // 1-80
	top := memory[34]    // 0-23
	bottom := memory[35] // 1-24
	this.Window.Left = left
	this.Window.Right = left + width - 1
	this.Window.Top = top
	this.Window.Bottom = bottom - 1

	if (this.CursorX < this.Window.Left) && (this.CursorX == 0) {
		this.CursorX = this.Window.Left
	}

	if this.CursorX > this.Window.Right {
		this.CursorX = this.Window.Left
	}
}

func (this *VDUCore) SetAttribute(a types.VideoAttribute) {
	this.Attribute = a
}

func (this *VDUCore) SetCurrentPage(p int) {
	this.CurrentPage = p
}

func (this *VDUCore) SetPrompt(s string) {
	this.Prompt = s
}

func (this *VDUCore) SetTabWidth(i int) {
	this.TabWidth = i
}

func (this *VDUCore) GetCursorX() int {
	return this.CursorX
}

func (this *VDUCore) GetCursorY() int {
	return this.CursorY
}

func (this *VDUCore) GetSpeed() int {
	return this.Speed
}

func (this *VDUCore) SetSpeed(v int) {
	this.Speed = v
}

func (this *VDUCore) SetSupressFormat(v bool) {
	this.SuppressFormat = v
}

func (this *VDUCore) GetSuppressFormat() bool {
	return this.SuppressFormat
}

func (this *VDUCore) GetVideoMode() types.VideoMode {
	return this.VideoMode
}

func (this *VDUCore) GetVideoModes() types.VideoModeList {
	return this.VideoModes
}

func (this *VDUCore) GetLastX() int {
	return this.LastX
}

func (this *VDUCore) GetLastY() int {
	return this.LastY
}

func (this *VDUCore) GetLastZ() int {
	return this.LastZ
}

func (this *VDUCore) GetTabWidth() int {
	return this.TabWidth
}

func (this *VDUCore) GetTextMemory() *types.TXMemoryBuffer {
	return this.TextMemory
}

func (this *VDUCore) GetShadowTextMemory() *types.TXMemoryBuffer {
	return this.ShadowTextMemory
}

func (this *VDUCore) GetWindow() *types.TextWindow {
	return &this.Window
}

func (this *VDUCore) GfxClearSplit(i int) {
	if this.VideoMode.Width == 40 {
		for r := 20; r < (this.VideoMode.Height / 2); r++ {
			for c := 0; c < this.VideoMode.Width; c++ {
				addr := this.XYToOffset(c, r)
				//laddr := c + (r * this.VideoMode.Width)
				this.TextMemory.SetValue(addr, uint(0|(this.FGColour<<16)))
				//this.ShadowTextMemory.SetValue(laddr, 0|(this.FGColour<<16))
				cx := (this.BGColour << 4) | this.FGColour
				this.VideoMode.PutXYMemory(this.ShadowTextMemory, c, r, 0, uint(cx), this.GetTextMode())

			}
		}
	}
}

func (this *VDUCore) SetTextMemory(m *types.TXMemoryBuffer) {
	this.TextMemory = m
}

func (this *VDUCore) SetShadowTextMemory(m *types.TXMemoryBuffer) {
	this.ShadowTextMemory = m
}

func (this *VDUCore) GrVertLine(x, y0, y1, c int) {
	for y := y0; y <= y1; y++ {
		this.GrPlot(x, y, c)
	}
}

func (this *VDUCore) GrHorizLine(x0, x1, y, c int) {
	for x := x0; x <= x1; x++ {
		this.GrPlot(x, y, c)
	}
}

func (this *VDUCore) GrPlot(x, y, c int) {

	//log.Printf("GRPLOT called with (x=%d, y=%d, c=%d)\n", x, y, c)

	c = c & 15
	//this.Cubes.plot ( x, this.VideoMode.Height-y-1, 0, c );
	cy := y / 2
	cx := x
	addr := this.XYToOffset(cx, cy)
	laddr := this.VideoMode.LinearOffsetXY(cx, cy)
	if (addr < 0) || (addr >= this.TextMemory.Size()) {
		addr = 0
		laddr = 0
	}
	v := this.TextMemory.GetValue(addr)

	hi := (v / 16) & 15
	lo := v & 15

	if (y % 2) == 0 {
		// low nibble
		lo = uint(c)
	} else {
		// hi nibble
		hi = uint(c)
	}

	this.TextMemory.SetValue(addr, ((hi*16)+lo)|uint(this.FGColour<<16))
	this.ShadowTextMemory.SetValue(laddr, ((hi*16)+lo)|uint(this.FGColour<<16)|uint(this.TextMode<<24))
}

func (this *VDUCore) GetPaddleButtons(z int) bool {
	return this.PaddleButtons[z%len(this.PaddleButtons)]
}

func (this *VDUCore) GetPaddleValues(z int) int {
	return this.PaddleValues[z%len(this.PaddleValues)]
}

func (this *VDUCore) GetPrompt() string {
	return this.Prompt
}

func (this *VDUCore) Put(ch rune) {

	/* vars */

	//this.OutputBuffer = this.OutputBuffer + ch;
	this.RealPut(ch)

}

func (this *VDUCore) OffsetToX(address int) int {
	if this.VideoMode.Columns == 80 {
		return this.OffsetToX80(address)
	} else {
		return this.OffsetToX40(address)
	}
}

func (this *VDUCore) OffsetToX40(address int) int {
	return int((address % 128) % 40)
}

func (this *VDUCore) OffsetToX80(address int) int {
	var off int = 1
	if address >= 1024 {
		off = 0
	}

	return (2 * ((address % 128) % 40)) + off
}

func (this *VDUCore) OffsetToY(address int) int {
	return ((address % 128) / 40 * 8) + ((address / 128) % 8)
}

func (this *VDUCore) XYToOffset80(x, y int) int {
	var offset int = 0
	if (x % 2) == 0 {
		offset = 1024
	}

	return ((y % 8) * 128) + ((y / 8) * 40) + (x / 2) + offset
}

func (this *VDUCore) XYToOffset40(x, y int) int {
	return ((y % 8) * 128) + ((y / 8) * 40) + x
}

func (this *VDUCore) XYToOffset(x, y int) int {
	if this.VideoMode.Columns == 80 {
		return this.XYToOffset80(x, y)
	} else {
		return this.XYToOffset40(x, y)
	}
}

func (this *VDUCore) ClearFull() {

	/* vars */
	// var i int
	var r int
	var c int

	this.Flush()

	/* clearing clears the current "window" */

	for r = 0; r <= this.VideoMode.Rows-1; r++ {
		for c = 0; c <= this.VideoMode.Columns-1; c++ {
			//i = (r * this.VideoMode.Columns) + c
			//            this.CharMem[i] = ' '
			//            this.ColourMem[i] = this.FGColour + 16*this.BGColour
			//            this.AttrMem[i] = types.VA_NORMAL

			to := this.XYToOffset(c, r)
			this.TextMemory.SetValue(to, uint(this.AsciiToPoke(' ')|(this.FGColour<<16)|(this.BGColour<<20)))

			//laddr := c + (r * this.VideoMode.Columns)
			cx := (this.BGColour << 4) | this.FGColour
			this.VideoMode.PutXYMemory(this.ShadowTextMemory, c, r, 160, uint(cx), this.GetTextMode())

			//this.ShadowTextMemory.SetValue(laddr, this.AsciiToPoke(' ')|(this.FGColour<<16)|(this.BGColour<<20))
		}
	}

}

func (this *VDUCore) Clear() {

	/* vars */
	var i int
	var r int
	var c int

	this.Flush()

	/* clearing clears the current "window" */

	for r = this.Window.Top; r <= this.Window.Bottom; r++ {
		for c = this.Window.Left; c <= this.Window.Right; c++ {
			i = (r * this.VideoMode.Columns) + c
			this.CharMem[i] = ' '
			this.ColourMem[i] = this.FGColour + 16*this.BGColour
			this.AttrMem[i] = types.VA_NORMAL

			to := this.XYToOffset(c, r)
			this.TextMemory.SetValue(to, uint(this.AsciiToPoke(' ')|(this.FGColour<<16)|(this.BGColour<<20)))

			//laddr := c + (r * this.VideoMode.Columns)
			cx := (this.BGColour << 4) | this.FGColour
			this.VideoMode.PutXYMemory(this.ShadowTextMemory, c, r, 160, uint(cx), this.GetTextMode())

			//this.ShadowTextMemory.SetValue(laddr, this.AsciiToPoke(' ')|(this.FGColour<<16)|(this.BGColour<<20))
		}
	}

}

func (this *VDUCore) RollPaddle(i int, amount int) {
	idx := (i % 4)

	ms_since := time.Now().Unix() - this.LastPaddleTime[idx]

	if ms_since < 150 {
		this.PaddleModifier[idx]++
	} else {
		this.PaddleModifier[idx] = 1
	}

	this.PaddleValues[idx] = (this.PaddleValues[idx] + int(float32(amount)*this.PaddleModifier[idx]))
	if this.PaddleValues[idx] > 255 {
		this.PaddleValues[idx] = 255
	}
	if this.PaddleValues[idx] < 0 {
		this.PaddleValues[idx] = 0
	}

	this.LastPaddleTime[idx] = time.Now().UnixNano() / 1000000
}

func (this *VDUCore) ClearToBottom() {

	/* vars */
	var r int
	var c int

	r = this.CursorY
	c = this.CursorX
	for r <= this.Window.Bottom {
		//i = c + (this.VideoMode.Columns * r)
		//        this.CharMem[i] = ' '
		//        this.ColourMem[i] = this.FGColour + 16*this.BGColour
		//        this.AttrMem[i] = this.Attribute

		to := this.XYToOffset(c, r)
		this.TextMemory.SetValue(to, uint(160|(this.FGColour<<16)|(this.BGColour<<20)))

		// laddr := c + (r * this.VideoMode.Columns)
		//this.ShadowTextMemory.SetValue(laddr, 160|(this.FGColour<<16)|(this.BGColour<<20))
		cx := (this.BGColour << 4) | this.FGColour
		this.VideoMode.PutXYMemory(this.ShadowTextMemory, c, r, 160, uint(cx), this.GetTextMode())

		c++
		if c > this.Window.Right {
			c = this.Window.Left
			r++
		}
	}

}

func (this *VDUCore) SaveVDUState() {
	this.SaveVideoMode = this.VideoMode
	this.SaveTextMemory = this.ShadowTextMemory.GetValues(0, this.TextMemory.Size())
	this.SaveCursorX = this.CursorX
	this.SaveCursorY = this.CursorY
	this.SaveAttribute = this.Attribute
	this.SaveTextMode  = this.TextMode
}

func (this *VDUCore) RestoreVDUState() {
	this.SetVideoMode(this.SaveVideoMode)
	this.ShadowTextMemory.SetValues(0, this.SaveTextMemory)
	this.CursorX = this.SaveCursorX
	this.CursorY = this.SaveCursorY
	this.Attribute = this.SaveAttribute
	this.TextMode  = this.SaveTextMode
}


func (this *VDUCore) AssetCheck(p, f string) (*files.FilePack, error) {
	return nil, nil
}

func (this *VDUCore) PlayWave(p, f string) (bool, error) {
	return false, nil
}

func (this *VDUCore) PNGSplash(p, f string) (bool, error) {
	return false, nil
}

func (this *VDUCore) PNGBackdrop(p, f string) (bool, error) {
	return false, nil
}

func (this *VDUCore) LoadEightTrack(p, f string) (bool, error) {
	return false, nil
}

func (this *VDUCore) CamResetAll() {

}

func (this *VDUCore) CamResetLoc() {

}

func (this *VDUCore) CamResetAngle() {

}

func (this *VDUCore) CamLock() {

}

func (this *VDUCore) CamUnlock() {

}

func (this *VDUCore) CamDolly(r float32) {

}

func (this *VDUCore) CamZoom(r float32) {

}

func (this *VDUCore) CamPos(x, y, z float32) {

}

func (this *VDUCore) CamPivPnt(x, y, z float32) {

}

func (this *VDUCore) CamMove(x, y, z float32) {

}

func (this *VDUCore) CamRot(x, y, z float32) {

}

func (this *VDUCore) CamOrbit(x, y float32) {

}

func (this *VDUCore) GetCGColour() int {
	return 0
}

func (this *VDUCore) IsClassicHGR() bool {
	return true
}

func (this *VDUCore) Reconnect(s string) {

}

func (this *VDUCore) SendRestalgiaEvent(b byte, s string) {

}

func (this *VDUCore) SetBGColourTriple(r,g,b int) {

}
