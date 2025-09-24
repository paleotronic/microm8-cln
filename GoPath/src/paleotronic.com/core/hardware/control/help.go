package control

import (
	"strings"

	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type HelpItem struct {
	Label  string
	Line   int
	X0, X1 int
	Target string
}

type HelpPos struct {
	Voffset, Hoffset int
	Line, Column     int
	Filename         string
	Content          []runestring.RuneString
}

func (hi HelpItem) IsPresent(line, col int) bool {
	return line == hi.Line && col >= hi.X0 && col <= hi.X1
}

type HelpController struct {
	edit                *editor.CoreEdit
	title               string
	raw                 string
	helpbase            string
	defaultfile         string
	filename            string
	links               map[int][]*HelpItem
	mouseLine, mouseCol int
	stack               []*HelpPos
}

func NewHelpController(ent interfaces.Interpretable, title string, helpbase string, filename string) *HelpController {
	return &HelpController{
		title:       title,
		defaultfile: filename,
		filename:    filename,
		helpbase:    helpbase,
		links:       make(map[int][]*HelpItem),
	}
}

func NewHelpControllerString(ent interfaces.Interpretable, title string, str string) *HelpController {
	hc := &HelpController{
		title:       title,
		defaultfile: "",
		filename:    "",
		helpbase:    "",
		raw:         str,
		links:       make(map[int][]*HelpItem),
	}
	return hc
}

func HelpPresent(ent interfaces.Interpretable) {
	hc := NewHelpController(ent, "Help System", settings.HelpBase, "quickhelp")
	hc.Do(ent)
}

func (hc *HelpController) LoadText(filename string) (string, error) {

	if !strings.HasPrefix(filename, "/") {
		filename = hc.helpbase + "/" + filename
	}

	if files.GetExt(filename) == "" || files.GetExt(filename) == filename {
		filename += ".hlp"
	}

	fmt.Printf("Looking for file: %s\n", filename)

	var text string = "Oh, where oh where did my little help go!"
	fr, err := files.ReadBytesViaProvider(files.GetPath(filename), files.GetFilename(filename))
	if err == nil {
		text = utils.Unescape(string(fr.Content))
	}
	return text, err
}

func (hc *HelpController) Push() {
	hp := &HelpPos{}
	hp.Column = hc.edit.Column
	hp.Line = hc.edit.Line
	hp.Voffset = hc.edit.Voffset
	hp.Hoffset = hc.edit.Hoffset
	hp.Content = hc.edit.Content
	hp.Filename = hc.filename
	if hc.stack == nil {
		hc.stack = make([]*HelpPos, 0)
	}
	hc.stack = append(hc.stack, hp)
}

func (hc *HelpController) Pop() {
	if hc.stack == nil {
		hc.stack = make([]*HelpPos, 0)
	}
	if len(hc.stack) == 0 {
		return
	}
	hp := hc.stack[len(hc.stack)-1]
	hc.stack = hc.stack[0 : len(hc.stack)-1]
	hc.edit.Line = hp.Line
	hc.edit.Column = hp.Column
	hc.edit.Voffset = hp.Voffset
	hc.edit.Hoffset = hp.Hoffset
	hc.edit.Content = hp.Content
	hc.edit.SubTitle = files.GetFilename(hp.Filename)
	hc.filename = hp.Filename
}

func (hc *HelpController) Do(ent interfaces.Interpretable) {
	apple2helpers.MonitorPanel(ent, true)
	apple2helpers.TEXTMAX(ent)

	var text string
	if hc.raw != "" {
		text = hc.raw
	} else {
		text, _ = hc.LoadText(hc.defaultfile)
	}

	hc.edit = editor.NewCoreEdit(
		ent,
		hc.title,
		text,
		false,
		false,
	)

	hc.edit.SubTitle = files.GetFilename(hc.defaultfile)
	hc.filename = hc.defaultfile

	hc.edit.CursorScrollWindow = true
	hc.edit.CursorHidden = true
	hc.edit.UserProcessHighlight = hc.ProcessHighlight
	hc.edit.UserRecombine = hc.Recombine

	hc.edit.BarBG = 15
	hc.edit.BarFG = 2
	hc.edit.FGColor = 15
	hc.edit.BGColor = 2
	hc.edit.SelBG = 14
	hc.edit.SelFG = 2

	// make us the event handler
	hc.edit.SetEventHandler(hc)

	hc.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_Q,
		"Quit",
		true,
		hc.HelpQuitFunc,
	)
	hc.edit.Run()

	apple2helpers.MonitorPanel(ent, false)
}

func (hc *HelpController) HelpQuitFunc(edit *editor.CoreEdit) {
	edit.Running = false
	edit.Done()
}

func (hc *HelpController) OnMouseClick(edit *editor.CoreEdit, left, right bool) {

	fmt.Printf("Mouse Buttons at (%v, %v)\n", left, right)

	if left && edit.MY > 0 && edit.MY < 46 {

		tl := edit.Voffset + int(edit.MY-1)
		tc := edit.Hoffset + int(edit.MX-1)
		if tl < len(edit.Content) {

			hc.mouseCol = tc
			hc.mouseLine = tl

			himatch, _, _ := hc.LinkAtPos(tl, tc)
			if himatch != nil {
				fmt.Printf("Enter on %s -> %s\n", himatch.Label, himatch.Target)
				text, err := hc.LoadText(himatch.Target)
				if err == nil {
					hc.NewStack(text, himatch.Target)
				} else {
					fmt.Println("Failed to load document")
				}
			}
		}

	}

}

func (hc *HelpController) OnMouseMove(edit *editor.CoreEdit, x, y int) {

	hc.mouseLine = edit.Voffset + int(y-1)
	hc.mouseCol = edit.Hoffset + int(x-1)
	edit.Display()

	edit.MouseMoved = true
}

func (this *HelpController) OnEditorExit(edit *editor.CoreEdit) {
	apple2helpers.Clearscreen(edit.Int)
}

func (this *HelpController) OnEditorBegin(edit *editor.CoreEdit) {
	//edit.Int.PutStr(string(rune(7)))
}

func (this *HelpController) OnEditorChange(edit *editor.CoreEdit) {

}

func (this *HelpController) OnEditorMove(edit *editor.CoreEdit) {

}

func (this *HelpController) NewStack(text string, filename string) {
	this.edit.SubTitle = files.GetFilename(filename)
	this.Push()
	this.edit.SetText(text)
	this.edit.Line = 0
	this.edit.Column = 0
	this.edit.Hoffset = 0
	this.edit.Voffset = 0
	this.edit.Display()
	this.mouseLine = 0
	this.mouseCol = 0
}

func (this *HelpController) OnEditorKeypress(edit *editor.CoreEdit, ch rune) bool {

	if ch == 13 || ch == 32 {

		tl := this.mouseLine
		tc := this.mouseCol

		himatch, _, _ := this.LinkAtPos(tl, tc)
		if himatch != nil {
			fmt.Printf("Enter on %s -> %s\n", himatch.Label, himatch.Target)
			text, err := this.LoadText(himatch.Target)
			if err == nil {
				this.NewStack(text, himatch.Target)
			} else {
				fmt.Println("Failed to load document")
			}
		}

		return true

	}

	if ch == vduconst.CSR_DOWN {
		fmt.Println("tab")
		ni := this.NextLink(this.mouseLine, this.mouseCol, -1, -1)
		if ni != nil {
			this.mouseLine, this.mouseCol = ni.Line, ni.X0
			this.edit.Display()
		} else {
			this.edit.CursorPageDown()
			ni = this.NextLink(this.mouseLine, this.mouseCol, -1, -1)
			if ni != nil {
				this.mouseLine, this.mouseCol = ni.Line, ni.X0
				this.edit.Display()
			}
		}
		return true
	}

	if ch == vduconst.CSR_UP {
		fmt.Println("tab")
		ni := this.PrevLink(this.mouseLine, this.mouseCol, -1, -1)
		if ni != nil {
			this.mouseLine, this.mouseCol = ni.Line, ni.X0
			this.edit.Display()
		} else {
			this.edit.CursorPageUp()
			ni = this.PrevLink(this.mouseLine, this.mouseCol, -1, -1)
			if ni != nil {
				this.mouseLine, this.mouseCol = ni.Line, ni.X0
				this.edit.Display()
			}
		}
		return true
	}

	if ch == 27 {
		// go back
		if this.stack == nil || len(this.stack) == 0 {
			this.HelpQuitFunc(this.edit)
			return true
		}
		this.Pop()
		return true
	}

	return false
}

// NextLink looks for the next available hyperlink between lines sl / el, nil
func (this *HelpController) NextLink(l, c int, sl, el int) *HelpItem {

	if sl == -1 {
		sl = this.edit.Voffset
	}
	if el == -1 {
		el = this.edit.Voffset + this.edit.Height - 1
	}

	// set current position
	tc, tl := c, l
	if tl < sl {
		tl = sl
	}

	// something at current pos?
	hi, _, _ := this.LinkAtPos(tl, tc)
	if hi != nil {
		tc = hi.X1 + 1 // end of current link before searching
		sl = hi.Line
	}

	// nothing at current pos or on existing line...
	for line := sl; line <= el; line++ {

		if line < len(this.edit.Content) {

			hr := this.ProcessHighlight(line, this.edit.Content[line])

			for tc < len(hr.Data.Runes) {
				hi, _, _ := this.LinkAtPos(line, tc)
				if hi != nil {
					return hi
				}
				tc++
			}

		}

		tc = 0

	}

	return nil
}

// NextLink looks for the next available hyperlink between lines sl / el, nil
func (this *HelpController) PrevLink(l, c int, sl, el int) *HelpItem {

	if sl == -1 {
		sl = this.edit.Voffset
	}
	if el == -1 {
		el = this.edit.Voffset + this.edit.Height - 1
	}

	// set current position
	tc, tl := c, l
	if tl > el {
		tl = el
	}

	// something at current pos?
	hi, _, _ := this.LinkAtPos(tl, tc)
	if hi != nil {
		tc = hi.X0 - 1 // end of current link before searching
		el = hi.Line
	}

	// nothing at current pos or on existing line...
	for line := el; line >= sl; line-- {

		if line < len(this.edit.Content) {

			hr := this.ProcessHighlight(line, this.edit.Content[line])

			if tc == -1 {
				tc = len(hr.Data.Runes) - 1
			}

			for tc >= 0 {
				hi, _, _ := this.LinkAtPos(line, tc)
				if hi != nil {
					return hi
				}
				tc--
			}

		}

		tc = -1

	}

	return nil
}

func (this *HelpController) LinkAtPos(tl, tc int) (*HelpItem, int, int) {
	items, ok := this.links[tl]
	if ok {
		for i, hi := range items {
			if hi.IsPresent(tl, tc) {
				return hi, i, len(items)
			}
		}
	}
	return nil, 0, 0
}

func (this *HelpController) ProcessHighlight(lineno int, s runestring.RuneString) editor.HRec {
	hr := editor.HRec{}

	// Link format:  [[ name ]] - internal wiki
	// Link format: [[ url | name ]] external wiki

	// color defaults
	fgcol := this.edit.FGColor
	hfgcol := 13
	bgcol := this.edit.BGColor
	linkcol := this.edit.SelBG

	items := make([]*HelpItem, 0)
	collect := runestring.Cast("")
	inlink := false

	lecount := 0
	var lastCh rune

	for _, ch := range s.Runes {
		switch {
		case ch >= vduconst.COLOR0 && ch <= vduconst.COLOR15:
			//
			fmt.Printf("color code %d\n", ch-vduconst.COLOR0)
			fgcol = int(ch - vduconst.COLOR0)
		case ch >= vduconst.BGCOLOR0 && ch <= vduconst.BGCOLOR15:
			//
			fmt.Printf("bgcolor code %d\n", ch-vduconst.BGCOLOR0)
			bgcol = int(ch - vduconst.BGCOLOR0)
		case ch == '*' && len(hr.Data.Runes) == 0:
			hr.Data.Runes = append(hr.Data.Runes, 1117)
			hr.Colour.Runes = append(hr.Colour.Runes, rune(hfgcol|(bgcol<<4)))
		case ch == '=' && len(hr.Data.Runes) == 0:
			lecount++
		case ch == '[' && lastCh == '[' && !inlink:
			collect = runestring.Cast("")
			inlink = true
			hr.Data.Runes = hr.Data.Runes[:len(hr.Data.Runes)-1]
			hr.Colour.Runes = hr.Colour.Runes[:len(hr.Colour.Runes)-1]
		case ch == ']' && lastCh == ']' && inlink:
			// trim off last char
			if len(collect.Runes) > 0 {
				collect.Runes = collect.Runes[:len(collect.Runes)-1]
			}

			inlink = false
			// TODO process collect here
			hi := &HelpItem{}
			hi.Line = lineno
			hi.X0 = len(hr.Data.Runes)
			parts := strings.Split(string(collect.Runes), "|")
			if len(parts) > 1 {
				hi.Label = strings.Trim(parts[1], " \t\r\n")
				hi.Target = strings.Trim(parts[0], " \t\r\n")
			} else {
				hi.Target = strings.Trim(parts[0], " \t\r\n")
				hi.Label = hi.Target
			}
			hi.X1 = hi.X0 + len(hi.Label) - 1
			for _, lch := range hi.Label {
				hr.Data.Runes = append(hr.Data.Runes, lch)

				if hi.Line == this.edit.Line && this.edit.Column >= hi.X0 && this.edit.Column <= hi.X1 {
					// link hover
					hr.Colour.Runes = append(hr.Colour.Runes, rune(bgcol|(linkcol<<4)))
				} else if hi.Line == this.mouseLine && this.mouseCol >= hi.X0 && this.mouseCol <= hi.X1 {
					// link hover
					hr.Colour.Runes = append(hr.Colour.Runes, rune(bgcol|(linkcol<<4)))
				} else {
					hr.Colour.Runes = append(hr.Colour.Runes, rune(linkcol|(bgcol<<4)))
				}
			}
			items = append(items, hi)
			//}
		default:
			if inlink {
				collect.Runes = append(collect.Runes, ch)
			} else {
				hrcount := lecount - 1
				hr.Data.Runes = append(hr.Data.Runes, ch)

				if hrcount < 1 {
					hr.Colour.Runes = append(hr.Colour.Runes, rune(fgcol|(bgcol<<4)))
				} else {
					hr.Colour.Runes = append(hr.Colour.Runes, rune(hfgcol|(bgcol<<4)))
				}
			}
		}
		lastCh = ch
	}

	// update links
	this.links[lineno] = items

	return hr
}

func (this *HelpController) Recombine(lineno int, hr editor.HRec) runestring.RuneString {

	return hr.Data

}
