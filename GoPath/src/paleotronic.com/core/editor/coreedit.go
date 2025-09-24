package editor

import (
	//"paleotronic.com/fmt"
	//	"paleotronic.com/fmt"

	"fmt"
	"regexp" //  "paleotronic.com/core/dialect"

	//"paleotronic.com/fmt"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/log"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

const (
	MAX_LINE_LENGTH = 1024
)

type CoreEditListener interface {
	OnEditorKeypress(this *CoreEdit, ch rune) bool
	OnEditorBegin(this *CoreEdit)
	OnEditorChange(this *CoreEdit)
	OnEditorExit(this *CoreEdit)
	OnEditorMove(this *CoreEdit)
	OnMouseMove(this *CoreEdit, x, y int)
	OnMouseClick(this *CoreEdit, left, right bool)
}

type HRec struct {
	Data, Colour runestring.RuneString
}

type CoreEditFunc func(this *CoreEdit)

type CoreEditHook struct {
	Key         rune
	Description string // Description of command
	Hook        CoreEditFunc
	Visible     bool
}

var menuFunc func(e interfaces.Interpretable)

func SetMenuHook(f func(e interfaces.Interpretable)) {
	menuFunc = f
}

func TestMenu(e interfaces.Interpretable) {
	if menuFunc != nil {
		menuFunc(e)
	}
}

func (this *CoreEditHook) GetShortcutInfo() (string, string) {
	keyname := string(this.Key)
	if this.Key >= vduconst.CTRL_A && this.Key <= vduconst.CTRL_Z {
		keyname = "^" + string(this.Key-vduconst.CTRL_A+65)
	}
	if this.Key >= vduconst.SHIFT_CTRL_A && this.Key <= vduconst.SHIFT_CTRL_Z {
		keyname = string(rune(1054)) + "^" + string(this.Key-vduconst.SHIFT_CTRL_A+65)
	}

	if this.Key == vduconst.SHIFT_CTRL_ENTER {
		keyname = string(rune(1054)) + "^" + string(rune(1120))
	}

	if this.Key >= vduconst.F1 && this.Key <= vduconst.F9 {
		keyname = "F" + string(this.Key-vduconst.F1+49)
	}
	if this.Key == vduconst.DELETE {
		keyname = "DEL"
	}
	if this.Key == vduconst.INSERT {
		keyname = "INS"
	}
	return keyname, this.Description
}

type EditorLineChange struct {
	Lines            []runestring.RuneString
	Line, Column     int
	Hoffset, Voffset int
}

func NewEditorCopy(current *CoreEdit) *EditorLineChange {
	this := &EditorLineChange{}
	this.Lines = make([]runestring.RuneString, len(current.Content))
	for i, _ := range current.Content {
		this.Lines[i] = current.Content[i].SubString(0, len(current.Content[i].Runes))
	}
	this.Line = current.Line
	this.Column = current.Column
	this.Voffset = current.Voffset
	this.Hoffset = current.Hoffset
	return this
}

type CoreEditColors struct {
	BarBG   int
	BarFG   int
	SelFG   int
	SelBG   int
	FGColor int
	BGColor int
}

type CoreEdit struct {
	CoreEditColors
	ForceExit                 bool
	AutoIndent                bool
	Mutable                   bool
	Line                      int
	Column                    int
	Width                     int
	Content                   []runestring.RuneString
	ReservedBot               int
	Hoffset                   int
	EventHandler              CoreEditListener
	Highlight                 bool
	Title                     string
	SubTitle                  string
	ReservedRight             int
	Int                       interfaces.Interpretable
	ReservedTop               int
	ReservedLeft              int
	MouseHidden               bool
	Height                    int
	Voffset                   int
	Running                   bool
	Tabstop                   int
	CursorScrollWindow        bool
	Shade                     int
	InsFGColor, InsBGColor    int
	ShowColor                 bool
	ShowOverwrite             bool
	Overwrite                 bool
	Inverse                   int
	SearchTerm                string // current search term
	Changed                   bool
	Hooks                     map[rune]*CoreEditHook
	LastCutLine               int
	CutBuffer                 []runestring.RuneString
	Parent                    interface{}
	SelStartCol, SelStartLine int
	SelEndCol, SelEndLine     int
	SelMode                   bool
	SelRect                   bool
	UndoBuffer                []*EditorLineChange
	UndoPointer               int
	UndoIgnore                bool
	ShowSwatch                bool
	MX, MY                    uint64
	MBL, MBR                  bool
	MouseMoved                bool
	UserProcessHighlight      func(lineno int, s runestring.RuneString) HRec
	UserRecombine             func(lineno int, hr HRec) runestring.RuneString
	CursorHidden              bool
	VSkip, HSkip              int
	lastHCMode                bool
	RealColors                CoreEditColors // saved color state
}

func (this *CoreEdit) InitMouse() bool {

	r := this.Int.GetMemoryMap()
	idx := this.Int.GetMemIndex()

	this.MX, this.MY = r.IntGetMousePos(idx)
	this.MBL, this.MBR = r.IntGetMouseButtons(idx)

	result := false

	return result
}

func (this *CoreEdit) ReadMouse() bool {

	r := this.Int.GetMemoryMap()
	idx := this.Int.GetMemIndex()

	nx, ny := r.IntGetMousePos(idx)
	lb, rb := r.IntGetMouseButtons(idx)

	result := false

	if lb != this.MBL || rb != this.MBR {
		if this.EventHandler != nil {
			this.EventHandler.OnMouseClick(this, lb, rb)
		}
		this.MBL, this.MBR = lb, rb
		result = true
	}

	if nx != this.MX || ny != this.MY {
		if this.EventHandler != nil {
			this.EventHandler.OnMouseMove(this, int(nx), int(ny))
		}
		this.MX, this.MY = nx, ny
		result = true
	}

	return result
}

func (this *CoreEdit) CutLine() {
	// If we changed line, then we should clear the cut buffer
	if this.LastCutLine != this.Line {
		this.CutBuffer = make([]runestring.RuneString, 0)
	}
	// Current line into the cut buffer
	this.CutBuffer = append(this.CutBuffer, runestring.Copy(this.Content[this.Line], 1, len(this.Content[this.Line].Runes)))
	this.DeleteCurrentLine()
	this.LastCutLine = this.Line
}

func (this *CoreEdit) UncutLines() {

	this.Push()

	this.InsertLine(len(this.CutBuffer))
	log.Printf("Inserted %d lines\n", len(this.CutBuffer))
	for i, l := range this.CutBuffer {
		this.Content[this.Line+i] = l
	}

	this.LastCutLine = -1
}

func (this *CoreEdit) SelectMode(b bool) {
	this.SelMode = b
	if b {
		this.SelStartCol = this.Column
		this.SelStartLine = this.Line
		this.SelEndCol = this.Column
		this.SelEndLine = this.Line
		this.SelRect = this.Overwrite
	}
}

func (this *CoreEdit) WordRight() bool {
	hr := this.ProcessHighlight(this.Line, this.Content[this.Line])
	c := this.Column
	if len(hr.Data.Runes) == 0 || c >= len(hr.Data.Runes) {
		return false
	}
	if hr.Data.Runes[c] == 32 {
		for c < len(hr.Data.Runes) && hr.Data.Runes[c] == 32 {
			c++
		}
		//c--
	} //else {
	for c < len(hr.Data.Runes) && hr.Data.Runes[c] != 32 {
		c++
	}
	//}
	// if c < len(hr.Data.Runes) {
	// 	c++
	// }
	for this.Column < c {
		this.CursorRight()
	}
	return true
}

func (this *CoreEdit) WordLeft() bool {
	hr := this.ProcessHighlight(this.Line, this.Content[this.Line])
	c := this.Column - 1
	if len(hr.Data.Runes) == 0 || c >= len(hr.Data.Runes) {
		for this.Column > 0 {
			this.CursorLeft()
		}
		return false
	}
	for c > 0 && hr.Data.Runes[c] == 32 {
		c--
	}
	for c > 0 && hr.Data.Runes[c] != 32 {
		c--
	}
	if c == 0 {
		c--
	}
	c++
	for this.Column > c {
		this.CursorLeft()
	}
	return true
}

func (this *CoreEdit) InsertLine(n int) {
	// make new thing
	this.Push()
	nc := make([]runestring.RuneString, len(this.Content)+n)
	for i := 0; i < len(nc); i++ {
		if i < this.Line {
			nc[i] = this.Content[i]
		} else if i < this.Line+n {
			nc[i] = runestring.Cast("")
		} else {
			nc[i] = this.Content[i-n]
		}
	}
	this.Content = nc
}

func (this *CoreEdit) Done() {
	this.Running = false
}

func (this *CoreEdit) GetLine() int {
	return this.Line
}

func (this *CoreEdit) GetColumn() int {
	return this.Column
}

func (this *CoreEdit) DeleteCurrentLine() {
	this.Push()
	if len(this.Content) == 1 {
		this.Content[0].Assign("")
	} else {
		a := this.Content[0:this.Line]
		b := this.Content[this.Line+1:]
		this.Content = append(a, b...)
	}
}

func lorune(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
}

func (this *CoreEdit) FindNextStart(index int, r rune) {
	i := this.Line + 1
	found := false

	for (!found) && (i < len(this.Content)) {
		hr := this.ProcessHighlight(i, this.Content[i])
		if len(hr.Data.Runes) < index+1 {
			i++
			continue
		}
		found = (lorune(hr.Data.Runes[index]) == lorune(r))
		if !found {
			i++
		}
	}

	if found {
		for this.Line < i {
			this.CursorDown()
		}
	}
}

func (this *CoreEdit) FindNextPrefix(prefix runestring.RuneString) bool {
	i := this.Line

	if prefix.Length() == 1 {
		i = this.Line + 1
	}

	found := false

	for (!found) && (i < len(this.Content)) {
		hr := this.ProcessHighlight(i, this.Content[i])
		if hr.Data.Length() < prefix.Length() {
			i++
			continue
		}
		found = hr.Data.HasPrefix(prefix)
		if !found {
			i++
		}
	}

	if found {
		for this.Line < i {
			this.CursorDown()
		}
	}

	return found
}

func rc(cond bool, c1, c2 rune) rune {
	if cond {
		return c1
	}
	return c2
}

func (this *CoreEdit) ProcessHighlight(lineno int, s runestring.RuneString) HRec {

	if this.UserProcessHighlight != nil {
		return this.UserProcessHighlight(lineno, s)
	}

	r := HRec{}
	r.Data = runestring.NewRuneString()
	r.Colour = runestring.NewRuneString()
	col := rune(this.FGColor | (this.BGColor << 4))
	bcol := rune(this.BGColor)
	shade := rune(0)
	inv := rune(0)
	nextCC := false
	for _, ch := range s.Runes {
		if nextCC {
			col = ch
			nextCC = false
			continue
		}
		if ch == 6 {
			nextCC = true
			continue
		}
		if ch >= vduconst.COLOR0 && ch <= vduconst.COLOR15 {
			col = rc(settings.HighContrastUI, 0, rune(ch-vduconst.COLOR0))
			continue
		}
		if ch >= vduconst.BGCOLOR0 && ch <= vduconst.BGCOLOR15 {
			bcol = rc(settings.HighContrastUI, 15, rune(ch-vduconst.BGCOLOR0))
			continue
		}
		if ch >= vduconst.SHADE0 && ch <= vduconst.SHADE7 {
			shade = rc(settings.HighContrastUI, 0, rune(ch-vduconst.SHADE0))
			continue
		}
		if ch == vduconst.INVERSE_ON {
			if inv == 0 {
				inv = 256
			} else {
				inv = 0
			}
			continue
		}
		r.Data.Append(string(ch))
		r.Colour.Append(string(col | (bcol << 4) | inv | (shade << 16)))
	}
	return r
}

func (this *CoreEdit) DisplayContent() {
	// run through content -- voffset -> voffset + this.Height-1
	//                     -- this.Hoffset -> this.Hoffset + this.Width-1

	for i := 0; i < this.Height; i++ {

		apple2helpers.SetBGColor(this.Int, uint64(this.BGColor))
		apple2helpers.SetFGColor(this.Int, uint64(this.FGColor))

		apple2helpers.Gotoxy(this.Int, this.ReservedLeft*this.HSkip, (this.ReservedTop+i)*this.VSkip)

		realLine := this.Voffset + i

		apple2helpers.Attribute(this.Int, types.VA_NORMAL)

		if (realLine >= 0) && (realLine < len(this.Content)) {
			s := this.Content[realLine]

			r := this.ProcessHighlight(realLine, s)

			if (realLine == this.Line) && (this.Highlight) {
				dd := ""
				rr := rune((this.SelBG << 4) | this.SelFG)
				for zz := 0; zz < len(r.Colour.Runes); zz++ {
					dd = dd + string(rr)
				}
				r.Colour.Assign(dd)
				for len(r.Colour.Runes) < this.Width {
					r.Colour.Runes = append(r.Colour.Runes, rr)
					r.Data.Runes = append(r.Data.Runes, ' ')
				}
			}

			if this.Hoffset > 0 {
				if len(r.Data.Runes) > this.Hoffset {
					r.Data = runestring.Delete(r.Data, 1, this.Hoffset)
					r.Colour = runestring.Delete(r.Colour, 1, this.Hoffset)
				} else {
					r.Data.Assign(" ")
					r.Colour.Assign(string(rune(this.FGColor | (this.BGColor << 4))))
				}
			}
			r.Data = runestring.Copy(r.Data, 1, this.Width)
			r.Colour = runestring.Copy(r.Colour, 1, this.Width)

			// now display it
			if len(r.Data.Runes) > 0 {
				for zz := 0; zz < len(r.Data.Runes); zz++ {

					realCol := this.Hoffset + zz

					if r.Colour.Runes[zz]&256 != 0 {
						apple2helpers.ColorFlip = true
					} else {
						apple2helpers.ColorFlip = false
					}

					fgcol := r.Colour.Runes[zz] & 15
					bgcol := (r.Colour.Runes[zz] >> 4) & 15

					if this.InSelection(realLine, realCol) {
						fgcol = rune(this.BarFG)
						bgcol = rune(this.BarBG)
					}

					apple2helpers.SetFGColor(this.Int, uint64(fgcol))
					apple2helpers.SetBGColor(this.Int, uint64(bgcol))
					apple2helpers.SetShade(this.Int, uint64((r.Colour.Runes[zz]>>16)&7))
					if r.Data.Runes[zz] != 7 {
						this.Int.PutStr(string(r.Data.Runes[zz]))
					}
				}

			}

			for zz := len(r.Data.Runes); zz < this.Width; zz++ {
				realCol := this.Hoffset + zz

				fgcol := this.FGColor
				bgcol := this.BGColor

				if this.InSelection(realLine, realCol) {
					fgcol = this.BarFG
					bgcol = this.BarBG
				}

				apple2helpers.SetFGColor(this.Int, uint64(fgcol))
				apple2helpers.SetBGColor(this.Int, uint64(bgcol))
				apple2helpers.SetShade(this.Int, 0)
				this.Int.PutStr(" ")
			}

			apple2helpers.ColorFlip = false
			apple2helpers.SetFGColor(this.Int, uint64(this.FGColor))
			apple2helpers.SetBGColor(this.Int, uint64(this.BGColor))
			apple2helpers.SetShade(this.Int, 0)

			for apple2helpers.GetCursorX(this.Int) != 0 {
				this.Int.PutStr(" ")
			}
		} else {
			s := " "
			this.Int.PutStr(s)
			for apple2helpers.GetCursorX(this.Int) != 0 {
				this.Int.PutStr(" ")
			}
		}

	}

	apple2helpers.SetFGColor(this.Int, uint64(this.FGColor))
	apple2helpers.SetBGColor(this.Int, uint64(this.BGColor))
	apple2helpers.Attribute(this.Int, types.VA_NORMAL)
	apple2helpers.SetShade(this.Int, 0)
}

func (this *CoreEdit) GetContent() []runestring.RuneString {

	return this.Content
}

func (this *CoreEdit) GetRawContent() []runestring.RuneString {

	c := make([]runestring.RuneString, 0)

	for i, r := range this.Content {
		hr := this.ProcessHighlight(i, r)
		c = append(c, hr.Data)
	}

	return c
}

func (this *CoreEdit) ToggleInverse() {
	f := this.FGColor
	b := this.BGColor
	this.BGColor = f
	this.FGColor = b
}

func (this *CoreEdit) CursorRight() {
	if this.CursorScrollWindow {
		if this.Hoffset < MAX_LINE_LENGTH-this.Width {
			this.Hoffset++
		}
		this.Column = this.Hoffset
		return
	}

	// normal scrolling
	if this.Column+1 < MAX_LINE_LENGTH {
		this.Column++
		if this.Column >= this.Hoffset+this.Width {
			this.Hoffset++
		}
	}
	this.CheckSel()
}

func (this *CoreEdit) CountLinesToTop(regex string) int {
	count := 0
	rc := regexp.MustCompile(regex)
	for i := this.Line - 1; i >= 0; i-- {
		s := this.Content[i]
		r := this.ProcessHighlight(i, s)
		//System.Out.Println(r.Data);

		if rc.MatchString(string(r.Data.Runes)) {
			count++
		}
	}
	return count
}

func (this *CoreEdit) GotoLine(l int) {
	if l >= len(this.Content) {
		l = len(this.Content) - 1
	}
	//~ this.Line = 0
	//~ this.Voffset = 0
	for this.Line < l {
		this.CursorDown()
	}
	for this.Line > l {
		this.CursorUp()
	}
}

func (this *CoreEdit) GotoColumn(l int) {
	hr := this.ProcessHighlight(this.Line, this.Content[this.Line])
	if l >= len(hr.Data.Runes) {
		l = len(hr.Data.Runes)
	}
	//~ this.Column = 0
	//~ this.Hoffset = 0
	for this.Column < l {
		this.CursorRight()
	}
	for this.Column > l {
		this.CursorLeft()
	}
}

func (this *CoreEdit) PromptString(label string, def string) string {
	//~ this.Int.SetMemory(36, 0)
	//~ this.Int.SetMemory(37, uint64(apple2helpers.GetRows(this.Int)-1))

	apple2helpers.Gotoxy(this.Int, 0, (apple2helpers.GetRows(this.Int)-1)*this.VSkip)

	result := this.GetCRTLine(label)
	if result == "" {
		result = def
	}

	return result
}

func (this *CoreEdit) GetCRTLine(promptString string) string {

	command := ""
	collect := true
	display := this.Int

	cb := this.Int.GetProducer().GetMemoryCallback(this.Int.GetMemIndex())

	display.PutStr(promptString)

	if cb != nil {
		cb(this.Int.GetMemIndex())
	}

	for collect {
		if cb != nil {
			cb(this.Int.GetMemIndex())
		}
		apple2helpers.TextShowCursor(this.Int)
		for this.Int.GetMemory(49152) < 128 {
			time.Sleep(5 * time.Millisecond)
			if cb != nil {
				cb(this.Int.GetMemIndex())
			}
			if this.Int.VM().IsDying() {
				return ""
			}
		}
		apple2helpers.TextHideCursor(this.Int)
		ch := rune(this.Int.GetMemory(49152) & 0xff7f)
		this.Int.SetMemory(49168, 0)
		switch ch {
		case 10:
			{
				display.SetSuppressFormat(true)
				display.PutStr("\r\n")
				display.SetSuppressFormat(false)
				return command
			}
		case 13:
			{
				display.SetSuppressFormat(true)
				display.PutStr("\r\n")
				display.SetSuppressFormat(false)
				return command
			}
		case 127:
			{
				if len(command) > 0 {
					command = utils.Copy(command, 1, len(command)-1)
					display.Backspace()
					display.SetSuppressFormat(true)
					display.PutStr(" ")
					display.SetSuppressFormat(false)
					display.Backspace()
					if cb != nil {
						cb(this.Int.GetMemIndex())
					}
				}
				break
			}
		default:
			{

				display.SetSuppressFormat(true)
				display.RealPut(rune(ch))
				display.SetSuppressFormat(false)

				if cb != nil {
					cb(this.Int.GetMemIndex())
				}

				command = command + string(ch)
				break
			}
		}
	}

	if cb != nil {
		cb(this.Int.GetMemIndex())
	}

	return command

}

func (this *CoreEdit) CheckPalette() {
	if this.lastHCMode != settings.HighContrastUI {
		this.lastHCMode = settings.HighContrastUI
		this.CheckContrast()
		apple2helpers.SetBGColor(this.Int, uint64(this.BGColor))
		apple2helpers.SetFGColor(this.Int, uint64(this.FGColor))
		apple2helpers.Clearscreen(this.Int)
		this.Display()
		apple2helpers.TEXT(this.Int).FullRefresh()
	}
}

func (this *CoreEdit) CheckContrast() {
	if settings.HighContrastUI {
		this.BarBG = 0
		this.BarFG = 15
		this.SelFG = 15
		this.SelBG = 0
		this.FGColor = 0
		this.BGColor = 15
	} else {
		this.SelFG = this.RealColors.SelFG
		this.SelBG = this.RealColors.SelBG
		this.BGColor = this.RealColors.BGColor
		this.FGColor = this.RealColors.FGColor
		this.BarBG = this.RealColors.BarBG
		this.BarFG = this.RealColors.BarFG
	}
}

func (this *CoreEdit) DisplayHeader() {

	this.CheckContrast()

	if this.ReservedTop == 0 {
		return
	}
	//~ this.Int.SetMemory(36, 0)
	//~ this.Int.SetMemory(37, 0)

	apple2helpers.Gotoxy(this.Int, 0, 0)

	if settings.HighContrastUI {
		apple2helpers.SetBGColor(this.Int, uint64(this.BarFG))
		apple2helpers.SetFGColor(this.Int, uint64(this.BarBG))
	} else {
		apple2helpers.SetBGColor(this.Int, uint64(this.BarBG))
		apple2helpers.SetFGColor(this.Int, uint64(this.BarFG))
	}

	pad := func() string {
		if !settings.HighContrastUI {
			return " "
		}
		switch apple2helpers.GetCursorX(this.Int) % 2 {
		case 0:
			return string(rune(1057))
		case 1:
			return string(rune(1058))
		}
		return " "
	}

	//apple2helpers.Attribute(this.Int, types.VA_INVERSE)
	this.Int.PutStr(pad() + " " + this.Title + " ")
	if this.SubTitle != "" {
		this.Int.PutStr("(" + this.SubTitle + ") ")
	}
	for apple2helpers.GetCursorX(this.Int) != 0 {
		this.Int.PutStr(pad())
	}

	if this.ShowSwatch {
		apple2helpers.SetCursorX(this.Int, 79)
		apple2helpers.SetCursorY(this.Int, 0)
		apple2helpers.SetFGColor(this.Int, uint64(this.FGColor))
		apple2helpers.SetBGColor(this.Int, uint64(this.BGColor))
		this.Int.PutStr(string(rune(1129)))
	}

	//apple2helpers.Attribute(this.Int, types.VA_NORMAL)
	apple2helpers.SetBGColor(this.Int, uint64(this.BGColor))
	apple2helpers.SetFGColor(this.Int, uint64(this.FGColor))
}

func (this *CoreEdit) Run() int {

	// save real colors
	this.RealColors = CoreEditColors{
		BGColor: this.BGColor,
		FGColor: this.FGColor,
		BarBG:   this.BarBG,
		BarFG:   this.BarFG,
		SelBG:   this.SelBG,
		SelFG:   this.SelFG,
	}

	this.CheckContrast()

	this.Int.GetMemoryMap().KeyBufferEmpty(this.Int.GetMemIndex())

	os := this.Int.GetGFXLayerState()
	ns := make([]bool, len(os))
	this.Int.SetGFXLayerState(ns)

	if this.EventHandler != nil {
		this.EventHandler.OnEditorBegin(this)
	}

	shortcutinfo := this.GetShortcutInfo(nil)

	if len(shortcutinfo) > 0 {
		rowsneeded := len(shortcutinfo)/8 + 1

		if this.ReservedBot < rowsneeded {
			this.ReservedBot = rowsneeded
		}
	}

	needRefresh := true
	m := this.Int.GetMemoryMap()

	this.InsBGColor = this.BGColor
	this.InsFGColor = this.FGColor

	apple2helpers.SetClientCopy(this.Int, false)

	apple2helpers.SetBGColor(this.Int, uint64(this.BGColor))
	apple2helpers.SetFGColor(this.Int, uint64(this.FGColor))
	txt := apple2helpers.TEXT(this.Int)
	txt.SetWindow(0, 0, 79, 47)
	apple2helpers.Clearscreen(this.Int)

	this.InitMouse()
	this.MouseMoved = false

	for this.Running {

		if needRefresh {
			this.Display()
			needRefresh = false
		}

		ol := this.Line
		oc := this.Column

		if !this.Int.GetProducer().AmIActive(this.Int) {
			fmt.Printf("I (%d) am dead\n", this.Int.GetUUID())
			return -1
		}

		for m.KeyBufferSize(this.Int.GetMemIndex()) == 0 {
			if this.Int != nil && this.Int.VM() != nil && this.Int.VM().IsDying() {
				return -1
			}
			if !this.Int.GetProducer().AmIActive(this.Int) {
				fmt.Println("Quitting as not active!!!")
				return -1
			}
			time.Sleep(5 * time.Millisecond)
			if this.Int.GetMemoryMap().IntGetSlotMenu(this.Int.GetMemIndex()) {
				if menuFunc != nil {
					menuFunc(this.Int)
				}
				this.Int.GetMemoryMap().IntSetSlotMenu(this.Int.GetMemIndex(), false)
			}
			if this.ReadMouse() {
				needRefresh = true
			} else {
				this.CheckScroll()
			}
			this.CheckPalette()
		}
		ch := rune(m.KeyBufferGet(this.Int.GetMemIndex()))
		//		fmt.Println("rune", int(ch))

		//this.Int.SetBuffer(runestring.Delete(this.Int.GetBuffer(), 1, 1))
		if this.EventHandler != nil {
			if !this.EventHandler.OnEditorKeypress(this, ch) {
				this.OnEditorKeypress(ch)
			}
		} else {
			this.OnEditorKeypress(ch)
		}
		needRefresh = true

		if (oc != this.Column || ol != this.Line) && this.EventHandler != nil {
			this.EventHandler.OnEditorMove(this)
		}
		//} else {
		//
		//	time.Sleep(time.Millisecond * 80)
		//}
	}

	apple2helpers.SetClientCopy(this.Int, true)

	if this.EventHandler != nil {
		this.EventHandler.OnEditorExit(this)
	}

	//this.Int.RestoreVDUState();

	this.Int.SetGFXLayerState(os)

	return 0
}

func (this *CoreEdit) CheckSel() {
	if !this.SelMode {
		return
	}
	this.SelEndCol = this.Column
	this.SelEndLine = this.Line
}

func (this *CoreEdit) InSelection(l, c int) bool {
	if !this.SelMode {
		return false
	}

	el := this.SelEndLine
	sl := this.SelStartLine
	ec := this.SelEndCol
	sc := this.SelStartCol

	if sl == el && ec < sc {
		sl = this.SelEndLine
		el = this.SelStartLine
		sc = this.SelEndCol
		ec = this.SelStartCol
	} else if el < sl {
		sl = this.SelEndLine
		el = this.SelStartLine
		sc = this.SelEndCol
		ec = this.SelStartCol
	}

	if this.SelRect {
		// rect based
		return l >= sl && l <= el && c >= sc && c <= ec
	} else {
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
	}

	return false

}

func (this *CoreEdit) CursorHome() {
	for this.Column > 0 {
		this.CursorLeft()
	}
	this.CheckSel()
}

func (this *CoreEdit) CursorEnd() {

	hr := this.ProcessHighlight(this.Line, this.Content[this.Line])

	for this.Column < len(hr.Data.Runes) {
		this.CursorRight()
	}
	this.CheckSel()
}

func (this *CoreEdit) CursorLeft() {

	if this.CursorScrollWindow {
		if this.Hoffset > 0 {
			this.Hoffset--
		}
		this.Column = this.Hoffset
		return
	}

	if this.Column-1 >= 0 {
		this.Column--
		if this.Column < this.Hoffset {
			this.Hoffset--
		}
	}
	this.CheckSel()
}

func (this *CoreEdit) GetText() string {
	out := ""
	for _, l := range this.Content {
		if out != "" {
			out = out + "\r\n"
		}
		out = out + string(l.Runes)
	}

	return out
}

func (this *CoreEdit) OnEditorKeypress(ch rune) {

	cec, ok := this.Hooks[ch]

	if ok {
		// exec command
		//		fmt.Printf("Got hook for 0x%.4x\n", ch)
		cec.Hook(this)
	} else if ch >= vduconst.BGCOLOR0 && ch <= vduconst.BGCOLOR15 {
		this.InsBGColor = int(ch - vduconst.BGCOLOR0)
	} else if ch >= vduconst.COLOR0 && ch <= vduconst.COLOR15 {
		this.InsFGColor = int(ch - vduconst.COLOR0)
	} else if ((ch >= 32) && (ch < 127)) || ((ch >= 1024) && (ch <= 1024+127)) {
		if this.Overwrite {
			this.CursorOverwriteChar(ch)
		} else {
			this.CursorInsertChar(ch)
		}
	} else {
		switch {
		case ch == vduconst.INVERSE_ON:
			this.ToggleInverse()
		case ch == vduconst.HOME:
			{
				this.CursorHome()
				break
			}
		//~ case vduconst.PASTE:
		//~ {
		//~ if err == nil {
		//~ text = strings.Replace(text, "\r\n", "\r", -1)
		//~ this.PasteBuffer = runestring.NewRuneString()
		//~ this.PasteBuffer.Append(text)
		//~ }
		//~ }
		case ch == vduconst.END:
			{
				this.CursorEnd()
				break
			}
		case ch == vduconst.PAGE_UP || ch == vduconst.SHIFT_CSR_UP:
			{
				this.CursorPageUp()
				break
			}
		case ch == vduconst.PAGE_DOWN || ch == vduconst.SHIFT_CSR_DOWN:
			{
				this.CursorPageDown()
				break
			}
		case ch == vduconst.CSR_UP:
			{
				this.CursorUp()
				break
			}
		case ch == vduconst.CSR_DOWN:
			{
				this.CursorDown()
				break
			}
		case ch == vduconst.CSR_LEFT:
			{
				this.CursorLeft()
				break
			}
		case ch == vduconst.CSR_RIGHT:
			{
				this.CursorRight()
				break
			}
		case ch == vduconst.SHIFT_CSR_LEFT:
			this.WordLeft()
		case ch == vduconst.SHIFT_CSR_RIGHT:
			this.WordRight()
		case ch == 13:
			{
				this.CursorCarriageReturn()
				break
			}
		case ch == 127:
			{
				this.CursorBackspace()
				break
			}
		case ch == 9:
			{
				this.CursorTab()
				break
			}
		}
	}

}

func (this *CoreEdit) CursorPageUp() {
	for i := 0; i < this.Height/2; i++ {
		this.CursorUp()
	}
	this.CheckSel()
}

func (this *CoreEdit) CursorDown() {

	if this.CursorScrollWindow {
		if this.Voffset < len(this.Content)-this.Height {
			this.Voffset++
		}
		this.Line = this.Voffset
		return
	}

	if this.Line+1 < len(this.Content) {
		this.Line++
		if this.Line >= this.Voffset+this.Height {
			this.Voffset++
		}
	}

	this.CheckSel()
}

func (this *CoreEdit) CursorCarriageReturn() {
	if !this.Mutable {
		return
	}

	hr := this.ProcessHighlight(this.Line, this.Content[this.Line])

	currLineLength := len(hr.Data.Runes)
	nextLine := this.Line + 1
	// insert types.Newline at position
	var indent int
	var v rune
	for indent, v = range hr.Data.Runes {
		if v != 32 {
			break
		}
	}

	this.Content = append(this.Content[:nextLine], append([]runestring.RuneString{runestring.Cast("")}, this.Content[nextLine:]...)...)

	//  if this.Column > currLineLength {
	//      this.Column = currLineLength
	//  }

	if this.Column < currLineLength {
		// need to split the line

		this.Push()

		first := runestring.Cast("")
		first.AppendSlice(hr.Data.Runes[0:this.Column])
		second := runestring.Delete(hr.Data, 1, len(first.Runes))

		firstc := runestring.Cast("")
		firstc.AppendSlice(hr.Colour.Runes[0:this.Column])
		secondc := runestring.Delete(hr.Colour, 1, len(firstc.Runes))

		if this.AutoIndent {
			for i := 0; i < indent; i++ {
				second.Runes = append([]rune{32}, second.Runes...)
				secondc.Runes = append([]rune{rune(this.FGColor) + 16*rune(this.BGColor)}, secondc.Runes...)
			}
		}

		ff := HRec{Data: first, Colour: firstc}
		ss := HRec{Data: second, Colour: secondc}

		this.Content[this.Line] = this.Recombine(this.Line, ff)
		this.Content[nextLine] = this.Recombine(nextLine, ss)

		if len(first.Runes) == 0 {
			apple2helpers.Clearscreen(this.Int)
		}
	}
	this.CursorDown()
	this.CursorHome()

	if this.AutoIndent {
		for i := 0; i < indent; i++ {
			this.CursorRight()
		}
	}

	this.Changed = true
}

func (this *CoreEdit) GetLineLength(i int) int {
	s := this.Content[i]
	r := this.ProcessHighlight(i, s)
	return len(r.Data.Runes)
}

func (this *CoreEdit) ResetDimensions() {
	this.Height = apple2helpers.GetRows(this.Int) - this.ReservedTop - this.ReservedBot
	this.Width = apple2helpers.GetColumns(this.Int) - this.ReservedLeft - this.ReservedRight
}

func NewCoreEdit(v interfaces.Interpretable, title string, content string, mutable bool, highlight bool) *CoreEdit {
	this := &CoreEdit{}
	this.Int = v
	this.Mutable = mutable
	this.Highlight = highlight
	this.Title = title

	this.Running = true
	this.ReservedBot = 2
	this.ReservedTop = 1
	this.Width = apple2helpers.GetColumns(this.Int)
	this.Height = apple2helpers.GetRows(this.Int)
	this.HSkip = 80 / this.Width
	this.VSkip = 48 / this.Height

	this.FGColor = 15

	this.MY = 99

	this.BarFG = 0
	this.BarBG = 15

	this.SelFG = 0
	this.SelBG = 15

	this.SetText(content)

	this.Height = apple2helpers.GetRows(this.Int) - this.ReservedTop - this.ReservedBot
	this.Width = apple2helpers.GetColumns(this.Int) - this.ReservedLeft - this.ReservedRight
	//this.Int.SaveVDUState();
	apple2helpers.Clearscreen(this.Int)
	//this.Display()
	if len(this.Content) == 0 {
		this.Content = []runestring.RuneString{runestring.Cast("")}
	}

	// init hook handler
	this.Hooks = make(map[rune]*CoreEditHook)
	this.CutBuffer = make([]runestring.RuneString, 0)
	this.LastCutLine = -1

	this.UndoBuffer = make([]*EditorLineChange, 0)

	return this
}

func (this *CoreEdit) RegisterCommand(key rune, desc string, v bool, f func(this *CoreEdit)) {
	this.Hooks[key] = &CoreEditHook{
		Description: desc,
		Hook:        f,
		Visible:     v,
		Key:         key,
	}
}

func (this *CoreEdit) Undo() {

	if this.UndoPointer == 0 {
		return // no undos
	}

	if len(this.UndoBuffer) == 0 {
		return
	}

	this.UndoPointer--
	state := this.UndoBuffer[this.UndoPointer]

	// apply it
	this.Content = state.Lines
	this.Line = state.Line
	this.Column = state.Column
	this.Voffset = state.Voffset
	this.Hoffset = state.Hoffset

}

func (this *CoreEdit) Redo() {

	if this.UndoPointer+1 >= len(this.UndoBuffer) {
		return // no undos
	}

	if len(this.UndoBuffer) == 0 {
		return
	}

	this.UndoPointer++
	state := this.UndoBuffer[this.UndoPointer]

	// apply it
	this.Content = state.Lines
	this.Line = state.Line
	this.Column = state.Column
	this.Voffset = state.Voffset
	this.Hoffset = state.Hoffset

}

func (this *CoreEdit) Push() {

	if this.UndoIgnore {
		return
	}

	// once you try to do something new, future redos are lost
	if this.UndoPointer < len(this.UndoBuffer) {
		this.UndoBuffer = this.UndoBuffer[0:this.UndoPointer]
	}

	state := NewEditorCopy(this)
	this.UndoBuffer = append(this.UndoBuffer, state)

	this.UndoPointer = len(this.UndoBuffer)
}

func (this *CoreEdit) CheckScroll() bool {
	y := this.MY
	if !this.MouseMoved {
		return false
	}
	if y == 0 {
		if this.Line > 0 {
			this.Int.GetMemoryMap().KeyBufferAdd(this.Int.GetMemIndex(), vduconst.CSR_UP)
			return true
		}
	} else if y == 47 {
		if this.Line < len(this.Content)-1 {
			this.Int.GetMemoryMap().KeyBufferAdd(this.Int.GetMemIndex(), vduconst.CSR_DOWN)
			return true
		}
	}
	return false
}

func (this *CoreEdit) Pop() {
	// revert if avail
	if len(this.UndoBuffer) == 0 {
		return
	}

	state := this.UndoBuffer[len(this.UndoBuffer)-1]
	this.UndoBuffer = this.UndoBuffer[0 : len(this.UndoBuffer)-1]
	// apply it
	this.Content = state.Lines
	this.Line = state.Line
	this.Column = state.Column
	this.Voffset = state.Voffset
	this.Hoffset = state.Hoffset

	this.UndoPointer = len(this.UndoBuffer)
}

func (this *CoreEdit) SetText(content string) {

	this.Content = make([]runestring.RuneString, 0)
	chunk := ""
	lastChar := ' '
	if content != "" {

		for _, ch := range content {

			switch ch {
			case 13:
				if lastChar != 6 {
					this.Content = append(this.Content, runestring.Cast(chunk))
					chunk = ""
				} else {
					chunk = chunk + string(ch)
				}
			case 10:
				if lastChar != 13 && lastChar != 6 {
					this.Content = append(this.Content, runestring.Cast(chunk))
					chunk = ""
				} else if lastChar != 13 {
					chunk = chunk + string(ch)
				}
			default:
				chunk = chunk + string(ch)
			}

			lastChar = ch
		}

		if chunk != "" {
			this.Content = append(this.Content, runestring.Cast(chunk))
		}
	}
}

func (this *CoreEdit) DisplayFooter() {
	if this.ReservedBot == 0 {
		return
	}
	//~ this.Int.SetMemory(37, uint64(this.ReservedTop+this.Height))
	//~ this.Int.SetMemory(36, 0)

	apple2helpers.Gotoxy(this.Int, 0, (this.ReservedTop+this.Height)*this.VSkip)

	sc := this.GetShortcutInfo(nil)

	if len(sc) == 0 {
		//apple2helpers.Attribute(this.Int, types.VA_INVERSE)
		apple2helpers.SetBGColor(this.Int, uint64(this.BarBG))
		apple2helpers.SetFGColor(this.Int, uint64(this.BarFG))
		this.Int.PutStr(" L" + utils.IntToStr(this.Line) + " C" + utils.IntToStr(this.Column))
		if this.ShowColor {
			oc := apple2helpers.GetFGColor(this.Int)
			ob := apple2helpers.GetBGColor(this.Int)
			this.Int.PutStr(" COL:")
			apple2helpers.SetFGColor(this.Int, uint64(this.BGColor))
			apple2helpers.SetBGColor(this.Int, uint64(this.FGColor))
			this.Int.PutStr(string(rune(1058)))
			apple2helpers.SetFGColor(this.Int, oc)
			apple2helpers.SetBGColor(this.Int, ob)
			if this.Inverse != 0 {
				this.Int.PutStr(" INV")
			}
		}
		if this.ShowOverwrite {
			if this.Overwrite {
				this.Int.PutStr(" OVR")
			} else {
				this.Int.PutStr(" INS")
			}
		}
		for apple2helpers.GetCursorX(this.Int) != 0 {
			this.Int.PutStr(" ")
		}
	} else {

		perline := 9
		spacing := 9
		if apple2helpers.GetColumns(this.Int) == 40 {
			perline = 4
		}

		for i, v := range sc {
			//~ this.Int.SetMemory(37, uint64(this.ReservedTop+this.Height+(i/perline)))
			//~ this.Int.SetMemory(36, uint64(i%perline)*10)

			x := (i % perline) * spacing
			y := this.ReservedTop + this.Height + (i / perline)

			apple2helpers.Gotoxy(this.Int, x*this.HSkip, y*this.VSkip)

			apple2helpers.SetBGColor(this.Int, uint64(this.BarBG))
			apple2helpers.SetFGColor(this.Int, uint64(this.BarFG))
			apple2helpers.PutStr(this.Int, v[0])
			//~ this.Int.SetMemory(36, uint64(i%perline)*10+3)

			apple2helpers.Gotoxy(this.Int, ((i%perline)*9+3)*this.HSkip, y*this.VSkip)

			apple2helpers.SetFGColor(this.Int, uint64(this.FGColor))
			apple2helpers.SetBGColor(this.Int, uint64(this.BGColor))
			apple2helpers.PutStr(this.Int, v[1])
		}

	}

	apple2helpers.SetFGColor(this.Int, uint64(this.FGColor))
	apple2helpers.SetBGColor(this.Int, uint64(this.BGColor))
	apple2helpers.Attribute(this.Int, types.VA_NORMAL)
}

func (this *CoreEdit) PositionCursor() {
	// y: reservedTop + (line-voffset);
	// x: reservedLeft + (this.Column-this.Hoffset);
	apple2helpers.Gotoxy(this.Int, (this.ReservedLeft+(this.Column-this.Hoffset))*this.HSkip, (this.ReservedTop+(this.Line-this.Voffset))*this.VSkip)

	//~ this.Int.SetMemory(36, uint64(this.ReservedLeft+(this.Column-this.Hoffset)))
	//~ this.Int.SetMemory(37, uint64(this.ReservedTop+(this.Line-this.Voffset)))
}

func (this *CoreEdit) CursorPageDown() {
	for i := 0; i < this.Height/2; i++ {
		this.CursorDown()
	}
	this.CheckSel()
}

func (this *CoreEdit) CursorUp() {

	if this.CursorScrollWindow {
		if this.Voffset > 0 {
			this.Voffset--
		}
		this.Line = this.Voffset
		return
	}

	if this.Line-1 >= 0 {
		this.Line--
		if this.Line < this.Voffset {
			this.Voffset--
		}
	}
	this.CheckSel()
}

func (this *CoreEdit) SetEventHandler(ev CoreEditListener) {
	this.EventHandler = ev
}

func (this *CoreEdit) Display() {

	// hide cursor
	//	this.Int.SetCursorVisible(false)

	apple2helpers.TextHideCursor(this.Int)

	this.DisplayHeader()
	this.DisplayContent()
	this.DisplayFooter()

	// rehome cursor before allowing it to be shown
	this.PositionCursor()

	cb := this.Int.GetProducer().GetMemoryCallback(this.Int.GetMemIndex())

	if cb != nil {
		cb(this.Int.GetMemIndex())
	}

	if !this.CursorHidden {
		apple2helpers.TextShowCursor(this.Int)
	}
	// show cursor
	//	this.Int.SetCursorVisible(true)

}

func (this *CoreEdit) Empty() {
	this.Push()
	this.Content = []runestring.RuneString{runestring.NewRuneString()}
}

func (this *CoreEdit) CursorOverwriteChar(ch rune) {
	if !this.Mutable {
		return
	}

	hr := this.ProcessHighlight(this.Line, this.Content[this.Line])

	thisLine := hr.Data
	thisColr := hr.Colour
	cs := rune(this.InsFGColor | (this.InsBGColor << 4) | this.Inverse | (this.Shade << 16))
	this.Changed = true
	fillcs := rune(15 | this.Inverse | (this.Shade << 16))

	this.Push()

	if this.Column >= len(thisLine.Runes) {
		for len(thisLine.Runes) < this.Column {
			thisLine.Append(" ")
			thisColr.Append(string(fillcs))
		}
		thisLine.Append(string(ch))
		thisColr.Append(string(cs))

		hr.Colour = thisColr
		hr.Data = thisLine
		this.Content[this.Line] = this.Recombine(this.Line, hr)
		this.CursorRight()
		return
	} else {
		// chars
		thisLine.Runes[this.Column] = ch
		// color
		thisColr.Runes[this.Column] = cs
		// merge
		hr.Colour = thisColr
		hr.Data = thisLine
		this.Content[this.Line] = this.Recombine(this.Line, hr)
		this.CursorRight()
	}
}

func (this *CoreEdit) CursorInsertChar(ch rune) {
	if !this.Mutable {
		return
	}

	if this.SelMode && !this.SelRect {
		this.DelSelection()
		this.SelMode = false
	}

	hr := this.ProcessHighlight(this.Line, this.Content[this.Line])

	thisLine := hr.Data
	thisColr := hr.Colour
	cs := rune(this.InsFGColor | (this.InsBGColor << 4) | this.Inverse | (this.Shade << 16))
	fillcs := rune(15 | this.Inverse | (this.Shade << 16))

	this.Changed = true

	this.Push()

	if this.Column > len(thisLine.Runes) {
		for len(thisLine.Runes) < this.Column {
			thisLine.Append(" ")
			thisColr.AppendSlice([]rune{fillcs})
		}
		thisLine.Append(string(ch))
		thisColr.AppendSlice([]rune{cs})
		hr.Colour = thisColr
		hr.Data = thisLine
		this.Content[this.Line] = this.Recombine(this.Line, hr)
		this.CursorRight()
		return
	} else {
		// chars
		first := thisLine.Runes[0:this.Column]
		last := thisLine.Runes[this.Column:]
		thisLine.Assign("")
		thisLine.AppendSlice(first)
		thisLine.Append(string(ch))
		thisLine.AppendSlice(last)
		// color
		first = thisColr.Runes[0:this.Column]
		last = thisColr.Runes[this.Column:]
		thisColr.Assign("")
		thisColr.AppendSlice(first)
		thisColr.AppendSlice([]rune{cs})
		thisColr.AppendSlice(last)
		// merge
		hr.Colour = thisColr
		hr.Data = thisLine
		this.Content[this.Line] = this.Recombine(this.Line, hr)
		this.CursorRight()
	}

	////fmt.Printf("Insert char [%v] with attr [%d]\n", ch, cs)
}

func (this *CoreEdit) CursorTab() {
	// tbc
	if !this.Mutable {
		return
	}
}

func (this *CoreEdit) CursorBackspace() {

	if !this.Mutable {
		return
	}

	if this.Column == 0 {
		if this.Line == 0 {
			return
		}

		this.Push()

		// join this line to previous line
		oldline := this.Line

		hr := this.ProcessHighlight(this.Line, this.Content[this.Line])
		pr := this.ProcessHighlight(this.Line-1, this.Content[this.Line-1])

		thisLine := hr.Data
		thisColr := hr.Colour
		prevLine := pr.Data
		prevColr := pr.Colour

		// get old length here
		oldlen := len(prevLine.Runes)

		prevLine.AppendRunes(thisLine)
		prevColr.AppendRunes(thisColr)

		hr.Data = prevLine
		hr.Colour = prevColr

		this.Content[this.Line-1] = this.Recombine(this.Line-1, hr)

		this.CursorUp()
		for this.Column < oldlen {
			this.CursorRight()
		}
		this.Content = append(this.Content[0:oldline], this.Content[oldline+1:]...)
		apple2helpers.Clearscreen(this.Int)
	} else {
		// remove character at x
		hr := this.ProcessHighlight(this.Line, this.Content[this.Line])

		thisLine := hr.Data
		thisColr := hr.Colour
		if this.Column > len(thisLine.Runes) {
			this.CursorLeft()
		} else {

			this.Push()

			thisLine = runestring.Delete(thisLine, this.Column, 1)
			thisColr = runestring.Delete(thisColr, this.Column, 1)
			hr.Data = thisLine
			hr.Colour = thisColr

			this.Content[this.Line] = this.Recombine(this.Line, hr)
			this.CursorLeft()
		}
	}

	this.Changed = true
}

func (this *CoreEdit) Recombine(lineno int, hr HRec) runestring.RuneString {

	if this.UserRecombine != nil {
		return this.UserRecombine(lineno, hr)
	}

	rs := runestring.NewRuneString()
	cc := 15
	bc := 0
	shade := 0
	inv := 0

	//////fmt.Printf("HR.Colour = %d, HR.Data = %d\n", len(hr.Colour.Runes), len(hr.Data.Runes))

	for i := 0; i < len(hr.Data.Runes); i++ {
		if cc&15 != int(hr.Colour.Runes[i]&15) {
			rs.AppendSlice([]rune{vduconst.COLOR0 + hr.Colour.Runes[i]&15})
			cc = int(hr.Colour.Runes[i] & 15)
		}
		if bc&15 != int((hr.Colour.Runes[i]>>4)&15) {
			rs.AppendSlice([]rune{vduconst.BGCOLOR0 + (hr.Colour.Runes[i]>>4)&15})
			bc = int((hr.Colour.Runes[i] >> 4) & 15)
		}
		if shade&15 != int((hr.Colour.Runes[i]>>16)&7) {
			rs.AppendSlice([]rune{vduconst.SHADE0 + (hr.Colour.Runes[i]>>16)&7})
			shade = int((hr.Colour.Runes[i] >> 16) & 7)
		}
		if inv != int(hr.Colour.Runes[i]&256) {
			rs.AppendSlice([]rune{vduconst.INVERSE_ON})
			inv = int(hr.Colour.Runes[i] & 256)
			////fmt.Printf("Inverse toggled at %d\n", i)
		}
		rs.AppendSlice([]rune{hr.Data.Runes[i]})
	}

	if inv != 0 {
		rs.AppendSlice([]rune{vduconst.INVERSE_ON})
	}

	if cc != 15 {
		rs.AppendSlice([]rune{vduconst.COLOR15})
	}

	if bc != 0 {
		rs.AppendSlice([]rune{vduconst.BGCOLOR0})
	}

	if shade != 0 {
		rs.AppendSlice([]rune{vduconst.SHADE0})
	}

	return rs
}

func (this *CoreEdit) SetFGColor(c rune) {
	this.InsFGColor = int(c)
}

func (this *CoreEdit) SetBGColor(c rune) {
	this.InsBGColor = int(c)
}

func (this *CoreEdit) GetFGColor() rune {
	return rune(this.InsFGColor)
}

func (this *CoreEdit) GetBGColor() rune {
	return rune(this.InsBGColor)
}

func (this *CoreEdit) SetShade(c rune) {
	this.Shade = int(c) & 7
}

func (this *CoreEdit) GetShade() rune {
	return rune(this.Shade)
}

// SearchForward will look for the next occurrence of the string
// in the current edit buffer, returning the line and column
// or -1, -1 if not found.
func (this *CoreEdit) SearchForward(st string) (int, int) {

	if st != "" {
		this.SearchTerm = st
	}

	if this.SearchTerm == "" {
		return -1, -1
	}

	// search through stripped lines
	match := false
	matchline, matchcolumn := -1, -1

	l := this.Line
	c := this.Column + 1

	for !match && l < len(this.Content) {

		hr := this.ProcessHighlight(l, this.Content[l])

		s := string(hr.Data.Runes)
		if c > 0 {
			if c < len(s) {
				s = s[c:]
			} else {
				l++
				if l >= len(this.Content) {
					return -1, -1
				}
				hr = this.ProcessHighlight(l, this.Content[l])
				s = string(hr.Data.Runes)
				c = 0
			}
		}

		matchcolumn = strings.Index(strings.ToLower(s), strings.ToLower(this.SearchTerm))
		if matchcolumn > -1 {
			// got match
			matchcolumn = matchcolumn + c
			matchline = l
			return matchline, matchcolumn
		}

		l++
		c = 0
	}

	return matchline, matchcolumn

}

// ReplaceAll will replace all instances of oldstr with
// newstr
func (this *CoreEdit) ReplaceAll(oldstr, newstr string) {

	this.Push()

	l := 0
	c := 0

	this.Column = 0
	this.Line = 0

	l, c = this.SearchForward(oldstr)
	for c != -1 {
		// GetLine in question
		hr := this.ProcessHighlight(l, this.Content[l])

		bData := hr.Data.Runes[0:c]
		bCols := hr.Colour.Runes[0:c]

		aData := hr.Data.Runes[c+len(oldstr):]
		aCols := hr.Colour.Runes[c+len(oldstr):]

		rs := runestring.NewRuneString()
		rs.Append(newstr)

		var cr rune
		if len(bCols) > 0 {
			cr = bCols[len(bCols)-1]
		} else if len(aCols) > 0 {
			cr = aCols[len(aCols)-1]
		} else {
			cr = rune(this.FGColor)
		}

		nd := HRec{}
		nd.Data.Runes = append(nd.Data.Runes, bData...)
		nd.Colour.Runes = append(nd.Colour.Runes, bCols...)
		// now new chars
		nd.Data.Runes = append(nd.Data.Runes, rs.Runes...)
		for len(nd.Colour.Runes) < len(nd.Data.Runes) {
			nd.Colour.Runes = append(nd.Colour.Runes, cr)
		}
		// now end
		nd.Data.Runes = append(nd.Data.Runes, aData...)
		nd.Colour.Runes = append(nd.Colour.Runes, aCols...)

		// process
		this.Content[l] = this.Recombine(l, nd)

		// again
		l, c = this.SearchForward(oldstr)
	}

	return
}

func (this *CoreEdit) GetShortcutInfo(hooks map[rune]*CoreEditHook) [][2]string {
	r := make([][2]string, 0)

	if hooks == nil {
		hooks = this.Hooks
	}

	keylist := make([]int, 0)
	for k, v := range hooks {
		if v.Visible {
			keylist = append(keylist, int(k))
		}
	}

	sort.Ints(keylist)

	for _, k := range keylist {
		kk, dd := hooks[rune(k)].GetShortcutInfo()
		r = append(r, [2]string{kk, dd})
	}

	return r
}

func (this *CoreEdit) Choice(prompt string, hooks map[rune]*CoreEditHook) *CoreEditHook {

	sc := this.GetShortcutInfo(hooks)

	option := rune(0)

	//~ this.Int.SetMemory(37, uint64(this.ReservedTop+this.Height-1))
	//~ this.Int.SetMemory(36, 0)

	apple2helpers.Gotoxy(this.Int, 0, (this.ReservedTop+this.Height-1)*this.VSkip)

	apple2helpers.ClearToBottom(this.Int)
	this.Int.PutStr(prompt)

	for option == 0 {

		perline := 9
		spacing := 9
		if apple2helpers.GetColumns(this.Int) == 40 {
			perline = 4
		}

		for i, v := range sc {
			//~ this.Int.SetMemory(37, uint64(this.ReservedTop+this.Height+(i/perline)))
			//~ this.Int.SetMemory(36, uint64(i%perline)*10)

			x := (i % perline) * spacing
			y := this.ReservedTop + this.Height + (i / perline)

			apple2helpers.Gotoxy(this.Int, x*this.HSkip, y*this.VSkip)

			apple2helpers.SetBGColor(this.Int, uint64(this.BarBG))
			apple2helpers.SetFGColor(this.Int, uint64(this.BarFG))
			apple2helpers.PutStr(this.Int, v[0])
			//~ this.Int.SetMemory(36, uint64(i%perline)*10+3)

			apple2helpers.Gotoxy(this.Int, ((i%perline)*9+3)*this.HSkip, y*this.VSkip)

			apple2helpers.SetFGColor(this.Int, uint64(this.FGColor))
			apple2helpers.SetBGColor(this.Int, uint64(this.BGColor))
			apple2helpers.PutStr(this.Int, v[1])
		}

		option = this.GetCRTKey()
		if option >= 'a' && option <= 'z' {
			option -= 32
		}

		for k, _ := range hooks {
			//fmt.Printf("Compare %v to %v\n", option, k)
			if k == option {
				return hooks[option]
			}
		}
		option = 0

	}

	return nil
}

func (this *CoreEdit) GetCRTKey() rune {

	command := rune(0)

	cb := this.Int.GetProducer().GetMemoryCallback(this.Int.GetMemIndex())

	if cb != nil {
		cb(this.Int.GetMemIndex())
	}

	if cb != nil {
		cb(this.Int.GetMemIndex())
	}
	apple2helpers.TextShowCursor(this.Int)
	for this.Int.GetMemory(49152) < 128 {
		time.Sleep(5 * time.Millisecond)
		if cb != nil {
			cb(this.Int.GetMemIndex())
		}
		if this.Int.VM().IsDying() {
			return command
		}
	}
	apple2helpers.TextHideCursor(this.Int)
	command = rune(this.Int.GetMemory(49152) & 0xff7f)
	this.Int.SetMemory(49168, 0)

	if cb != nil {
		cb(this.Int.GetMemIndex())
	}

	return command

}

func (this *CoreEdit) CopySelection() {
	if !this.SelMode {
		return
	}

	this.CutBuffer = make([]runestring.RuneString, 0)

	el := this.SelEndLine
	sl := this.SelStartLine
	ec := this.SelEndCol
	sc := this.SelStartCol

	if sl == el && ec < sc {
		sl = this.SelEndLine
		el = this.SelStartLine
		sc = this.SelEndCol
		ec = this.SelStartCol
	} else if el < sl {
		sl = this.SelEndLine
		el = this.SelStartLine
		sc = this.SelEndCol
		ec = this.SelStartCol
	}

	textual := ""

	for l := sl; l <= el; l++ {
		line := this.Content[l]
		hr := this.ProcessHighlight(l, line)

		s := len(hr.Data.Runes)
		e := 0

		for c := 0; c < len(hr.Data.Runes); c++ {
			if this.InSelection(l, c) {
				if c < s {
					s = c
				}
				if c > e {
					e = c
				}
			}
		}

		nhr := HRec{}
		nhr.Data = hr.Data.SubString(s, e+1)
		nhr.Colour = hr.Colour.SubString(s, e+1)

		this.CutBuffer = append(this.CutBuffer, this.Recombine(l, nhr))

		if textual != "" {
			textual += "\r\n"
		}
		textual += string(this.Recombine(l, nhr).Runes)

	}

	if textual != "" {
		clipboard.WriteAll(textual)
	}
}

func (this *CoreEdit) DelSelection() {

	if !this.SelMode {
		return
	}

	this.Push() // save current state
	this.UndoIgnore = true

	el := this.SelEndLine
	sl := this.SelStartLine
	ec := this.SelEndCol
	sc := this.SelStartCol

	if sl == el && ec < sc {
		sl = this.SelEndLine
		el = this.SelStartLine
		sc = this.SelEndCol
		ec = this.SelStartCol
	} else if el < sl {
		sl = this.SelEndLine
		el = this.SelStartLine
		sc = this.SelEndCol
		ec = this.SelStartCol
	}

	if this.Overwrite {

		// handle overwrite delete...
		for l := sl; l <= el; l++ {
			for c := sc; c <= ec; c++ {
				this.Line = l
				this.Column = c
				this.CursorOverwriteChar(' ')
			}
		}

	} else {

		for l := sl; l <= el; l++ {
			line := this.Content[l]
			hr := this.ProcessHighlight(l, line)

			nhr := HRec{}

			for c := 0; c < len(hr.Data.Runes); c++ {
				if !this.InSelection(l, c) {
					nhr.Data.Runes = append(nhr.Data.Runes, hr.Data.Runes[c])
					nhr.Colour.Runes = append(nhr.Colour.Runes, hr.Colour.Runes[c])
				}
			}

			r := this.Recombine(l, nhr)
			this.Content[l] = r

		}

		// delete lines...
		for l := el - 1; l > sl; l-- {
			this.Line = l
			this.DeleteCurrentLine()
		}

		if sl != el {
			this.Line = sl + 1
			this.Column = 0
			this.CursorBackspace()
		}

	}

	this.Line = sl
	this.Column = sc

	this.UndoIgnore = false

}

func (this *CoreEdit) Paste() {

	this.Push()
	this.UndoIgnore = true

	col := this.Column
	line := this.Line

	ofg := this.InsFGColor
	obg := this.InsBGColor
	osh := this.Shade

	//	fmt.Println("PASTING")

	text, err := clipboard.ReadAll()
	if err == nil {
		text = strings.Replace(text, "\r\n", "\r", -1)
		text = strings.Replace(text, "\n", "\r", -1)

		if len(text) != 0 {
			this.CutBuffer = make([]runestring.RuneString, 0)
			parts := strings.Split(text, "\r")
			for _, p := range parts {
				r := runestring.Cast(p)
				this.CutBuffer = append(this.CutBuffer, r)
			}
		}

	}

	if len(this.CutBuffer) > 0 {

		for i, l := range this.CutBuffer {

			if this.Overwrite {
				this.Column = col
			}

			hr := this.ProcessHighlight(i, l)

			for i, ch := range hr.Data.Runes {

				cr := hr.Colour.Runes[i]

				fgcol := int(cr & 15)
				bgcol := int((cr >> 4) & 15)
				shade := int((cr >> 16) & 7)

				this.FGColor = fgcol
				this.BGColor = bgcol
				this.Shade = shade

				if this.Overwrite {
					this.CursorOverwriteChar(ch)
				} else {
					this.CursorInsertChar(ch)
				}

			}

			if this.Overwrite {
				for line+i >= len(this.Content) {
					this.Content = append(this.Content, runestring.Cast(""))
				}
				if i < len(this.CutBuffer)-1 {
					this.CursorDown()
				}
			} else {
				if i < len(this.CutBuffer)-1 {
					this.CursorCarriageReturn()
				}
			}
		}

	}

	this.InsFGColor = ofg
	this.InsBGColor = obg
	this.Shade = osh

	this.UndoIgnore = false
}

/*
	GridChooser provides a way to choose an item from a grid of items, eg a colour palette, a char
	picker etc.
*/

type GridItem struct {
	FG    int
	BG    int
	ID    int
	Label string
}

/*        __
@@ @@|@@|
     --
@@ @@ @@
*/

func (this *CoreEdit) GridChooser(prompt string, items []GridItem, hmargin, vmargin int, itemwidth, itemheight int) int {

	width := apple2helpers.GetTextWidth(this.Int) - 2*hmargin
	height := apple2helpers.GetRows(this.Int) - 2*vmargin

	if settings.HighContrastUI {
		apple2helpers.SetBGColor(this.Int, 0)
		apple2helpers.SetFGColor(this.Int, 15)
	} else {
		apple2helpers.SetBGColor(this.Int, 2)
		apple2helpers.SetFGColor(this.Int, 15)
	}

	apple2helpers.TextDrawBox(
		this.Int,
		hmargin,
		vmargin,
		width,
		height,
		prompt,
		true,
		false,
	)

	selected := 0 // index of selected item
	itemsperh := (width / itemwidth) - 1
	itemsperv := (height / itemheight) - 1

	soffset := 0
	maxitems := itemsperh * itemsperv

	var done bool

	fg := apple2helpers.GetFGColor(this.Int)
	bg := apple2helpers.GetBGColor(this.Int)

	for !done {
		// Draw grid
		for i := soffset; i < len(items) && i < soffset+maxitems; i++ {
			item := items[i]
			x := hmargin + (i%itemsperh)*itemwidth + 1
			y := vmargin + ((i-soffset)/itemsperh)*itemheight + 1
			//~ this.Int.SetMemory(36, uint64(x))
			//~ this.Int.SetMemory(37, uint64(y))

			apple2helpers.Gotoxy(this.Int, x, y)

			apple2helpers.SetBGColor(this.Int, bg)
			apple2helpers.SetFGColor(this.Int, fg)

			if i == selected {
				this.Int.PutStr("[")
			} else {
				this.Int.PutStr(" ")
			}
			apple2helpers.SetBGColor(this.Int, uint64(item.BG))
			apple2helpers.SetFGColor(this.Int, uint64(item.FG))
			this.Int.PutStr(item.Label)
			apple2helpers.SetBGColor(this.Int, bg)
			apple2helpers.SetFGColor(this.Int, fg)
			if i == selected {
				this.Int.PutStr("]")
			} else {
				this.Int.PutStr(" ")
			}
		}

		// read a key
		for this.Int.GetMemoryMap().KeyBufferSize(this.Int.GetMemIndex()) == 0 {
			time.Sleep(50 * time.Millisecond)
		}

		ch := this.Int.GetMemoryMap().KeyBufferGetLatest(this.Int.GetMemIndex())

		switch ch {
		case vduconst.CSR_LEFT:
			if selected > 0 {
				selected--
			}
		case vduconst.CSR_RIGHT:
			if selected < len(items)-1 {
				selected++
			}
		case vduconst.CSR_DOWN:
			if selected < len(items)-itemsperh {
				selected += itemsperh
			}
		case vduconst.CSR_UP:
			if selected >= itemsperh {
				selected -= itemsperh
			}
		case 13:
			done = true
			break
		case 27:
			return -1
		}

		// if we are here, correct view
		for soffset > selected {
			soffset -= itemsperh
		}
		for selected >= soffset+maxitems {
			soffset += itemsperh
		}

	}

	return items[selected].ID
}

func (this *CoreEdit) ShowKeymap(prompt string, hmargin, vmargin int) int {

	//	inlines := []string{
	//		"123456789-=",
	//		"QWERTYUIOP[]",
	//		"ASDFGHJKL;'",
	//		"ZXCVBNM,./",
	//	}

	items := make([]string, 0)
	//	for _, l := range inlines {

	//	}

	width := apple2helpers.GetTextWidth(this.Int) - 2*hmargin
	height := apple2helpers.GetRows(this.Int) - 2*vmargin

	if settings.HighContrastUI {
		apple2helpers.SetBGColor(this.Int, 0)
		apple2helpers.SetFGColor(this.Int, 15)
	} else {
		apple2helpers.SetBGColor(this.Int, 2)
		apple2helpers.SetFGColor(this.Int, 15)
	}

	apple2helpers.TextDrawBox(
		this.Int,
		hmargin,
		vmargin,
		width,
		height,
		prompt,
		true,
		false,
	)

	var done bool

	//	fg := apple2helpers.GetFGColor(this.Int)
	//	bg := apple2helpers.GetBGColor(this.Int)

	for !done {
		// Draw lines
		for i, v := range items {
			x := 1
			y := vmargin + 1 + i
			//~ this.Int.SetMemory(36, uint64(x))
			//~ this.Int.SetMemory(37, uint64(y))

			apple2helpers.Gotoxy(this.Int, x, y)

			this.Int.PutStr(v)
		}

		// read a key
		for this.Int.GetMemoryMap().KeyBufferSize(this.Int.GetMemIndex()) == 0 {
			time.Sleep(50 * time.Millisecond)
		}

		ch := this.Int.GetMemoryMap().KeyBufferGetLatest(this.Int.GetMemIndex())

		switch ch {
		case 13:
			done = true
			break
		case 27:
			return -1
		}

	}

	return 0
}
