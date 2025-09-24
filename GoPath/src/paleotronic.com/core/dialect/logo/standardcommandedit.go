package logo

import (
	"errors"
	"strings"
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/files" //"paleotronic.com/log"
	"paleotronic.com/fmt"   //"github.com/atotto/clipboard"
	"paleotronic.com/log"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type StandardCommandEDIT struct {
	dialect.Command
	selected int
	save     bool
	textMode bool
	richText bool
	textFile string
	execute  int
	edit     *editor.CoreEdit
	Scope    *types.Algorithm
}

func NewStandardCommandEDIT() *StandardCommandEDIT {
	this := &StandardCommandEDIT{}
	this.ImmediateMode = true
	this.NoTokens = true
	return this
}

func (this *StandardCommandEDIT) Syntax() string {
	// TODO Auto-generated method stub
	return "EDIT"
}

func Pad(s string, l int) string {

	for len(s) < l {
		s = " " + s
	}

	return s
}

func inList(s string, list []string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func (this *StandardCommandEDIT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	this.textMode = false
	settings.DisableMetaMode[caller.GetMemIndex()] = true

	d := caller.GetDialect().(*DialectLogo)
	d.Driver.KillAllCoroutines()
	time.Sleep(50 * time.Millisecond)

	// Purge keys
	caller.SetBuffer(runestring.NewRuneString())

	gotoLine := 0
	gotoError := ""

	this.richText = false

	fmt.Println("EDIT [", caller.TokenListAsString(tokens), "]")

	var logoProc string

	if tokens.Size() > 0 && caller.GetDialect().GetTitle() != "Logo" {
		s := ""
		for _, tt := range tokens.Content {
			s += tt.Content
		}
		tokens = *types.NewTokenList()

		fmt.Printf("Concat edit param [%s]\n", s)

		if utils.StrToInt(s) != 0 {
			tokens.Push(types.NewToken(types.NUMBER, s))
		} else if s != "" {
			tokens.Push(types.NewToken(types.STRING, s))
		}

		t, _ := caller.GetDialect().ParseTokensForResult(caller, tokens)

		if t.Type != types.INVALID {
			if t.IsNumeric() {
				gotoLine = t.AsInteger() - 1
			} else {
				this.textMode = true
				wd := strings.Trim(caller.GetWorkDir(), "/")
				this.textFile = t.Content
				if wd != "" && rune(this.textFile[0]) != '/' {
					this.textFile = wd + "/" + t.Content
				}
				this.textFile = strings.ToLower(this.textFile)
				ext := files.GetExt(this.textFile)

				fmt.Println("Ext =", ext)

				script, _, _ := files.IsLaunchable(ext)
				this.richText = !script

				if files.IsBinary(ext) {
					return 0, errors.New("CANNOT EDIT BINARY FILE")
				}

				fmt.Println("filename =", this.textFile)
			}
		}
	} else if tokens.Size() > 0 {
		// Logo
		logoProc = tokens.Shift().Content

	}

	fmt.Printf("Rich text file: %v\n", this.richText)

	tt := caller.GetCode()
	l := tt.GetLowIndex()
	//  h := tt.GetHighIndex()

	text := ""
	s := ""
	ft := ""

	var ln types.Line
	this.save = false
	var readonly bool

	if this.textMode {

		if files.IsEditPlain(files.GetExt(this.textFile)) {
			this.richText = false
		}

		fmt.Println("Text mode")

		if files.ExistsViaProvider(files.GetPath(this.textFile), files.GetFilename(this.textFile)) {

			b, err := files.ReadBytesViaProvider(files.GetPath(this.textFile), files.GetFilename(this.textFile))

			if err != nil {
				return 0, err
			}

			text = utils.Unescape(string(b.Content))

			if le := files.LockViaProvider(files.GetPath(this.textFile), files.GetFilename(this.textFile)); le == nil {
				readonly = false
			} else {
				readonly = true
			}

		} else {
			text = ""
		}

	} else {

		if caller.GetDialect().GetTitle() == "Logo" {

			var lines []string

			lines = caller.GetDialect().GetWorkspaceBody(false, logoProc)

			text = strings.Join(lines, "\r\n")

		} else {

			h := tt.GetHighIndex()
			nlen := len(utils.IntToStr(h))

			for l != -1 {
				/* display this line */
				//Str(l, s)
				//write(f, PadLeft(s, w)+' ');
				/* now formatted tokens */
				ln, _ = caller.GetCode().Get(l)
				s = Pad(utils.IntToStr(l), nlen) + "  "
				z := 0
				for _, stmt := range ln {

					stl := *stmt.SubList(0, stmt.Size())

					ft = caller.TokenListAsString(stl)

					//ft = strings.Replace( ft, string([]rune{34,4,34}), "CHR$(4)", -1 )

					if z > 0 {
						s = s + "\r\n"
						s = s + Pad(":", nlen) + "  "
					}

					s = s + ft

					z++
				}
				s = s + "\r\n\r\n"
				text = text + s

				/* next line */
				l = tt.NextAfter(l)
			}

		}

	}

rerunedit:

	//apple2helpers.TextSaveScreen(caller)
	apple2helpers.MonitorPanel(caller, true)
	apple2helpers.TEXTMAX(caller)
	//apple2helpers.Clearscreen(caller)

	caller.SetIgnoreSpecial(true)

	//	caller.GetVDU().SaveVDUState()
	//	caller.GetVDU().SetVideoMode(caller.GetVDU().GetVideoModes()[0])
	this.edit = editor.NewCoreEdit(caller, "EDIT (F2=Accept Changes, ESC=Cancel)", text, !readonly, (gotoError != ""))
	this.edit.ShowSwatch = this.textMode
	this.edit.AutoIndent = true
	UpdateSubTitle(this.edit)
	if gotoError != "" {
		this.edit.SubTitle = gotoError
		gotoError = ""
	}

	mode := "EDIT"
	if readonly {
		mode = "VIEW"
	}

	if this.textMode {
		if this.richText {
			this.edit.BarFG = 15
			this.edit.BarBG = 6
			this.edit.BGColor = 0
			this.edit.FGColor = 15
			this.edit.Title = mode + " (RICH TEXT) File: " + this.textFile
		} else {
			this.edit.BarFG = 15
			this.edit.BarBG = 6
			this.edit.BGColor = 0
			this.edit.FGColor = 15
			this.edit.Title = mode + " (PLAIN TEXT) File: " + this.textFile
		}
	} else {
		this.edit.BarFG = 15
		this.edit.BarBG = 3
		this.edit.BGColor = 0
		this.edit.FGColor = 15
		this.edit.SelBG = 5
		this.edit.Title = mode + " (BASIC)"
	}

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_Q,
		"Quit",
		true,
		EditorExit,
	)

	if !readonly {
		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_A,
			"AInd",
			true,
			EditorAutoIndent,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_K,
			"Kill",
			true,
			EditorCut,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_U,
			"Unkill",
			true,
			EditorUncut,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_O,
			"Save",
			true,
			EditorSave,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_Y,
			"Redo",
			true,
			EditorRedo,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_Z,
			"Undo",
			true,
			EditorUndo,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_X,
			"Cut",
			true,
			EditorCutSel,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_V,
			"Paste",
			true,
			EditorPaste,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_E,
			"Chars",
			true,
			EditorChar,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_H,
			"KeyHelp",
			true,
			ShowKeymap,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_I,
			"Ins",
			true,
			EditorToggleInsert,
		)

		if this.richText {

			this.edit.RegisterCommand(
				vduconst.SHIFT_CTRL_F,
				"FgCol",
				true,
				EditorFG,
			)

			this.edit.RegisterCommand(
				vduconst.SHIFT_CTRL_B,
				"BgCol",
				true,
				EditorBG,
			)

		}

	}

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_T,
		"40/80",
		true,
		EditorToggle4080,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_S,
		"Select",
		true,
		EditorSelect,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_C,
		"Copy",
		true,
		EditorCopy,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_W,
		"Where",
		true,
		EditorFind,
	)

	this.edit.Parent = this

	if this.textMode && this.richText && !readonly {
		this.edit.ShowColor = true
		this.edit.ShowOverwrite = true
	}
	this.edit.SetEventHandler(this)
	if gotoLine > 0 {
		this.edit.GotoLine(gotoLine)
	}

	this.edit.Run()
	//caller.GetVDU().RestoreVDUState()
	//apple2helpers.TextRestoreScreen(caller)
	defer apple2helpers.MonitorPanel(caller, false)
	caller.SetIgnoreSpecial(false)

	if this.save {
		if this.textMode && this.richText {
			sl := utils.Escape(this.edit.GetText())
			// save it
			errw := files.WriteBytesViaProvider(files.GetPath(this.textFile), files.GetFilename(this.textFile), []byte(sl))
			log.Printf("Errw = %v", errw)
			// if err == nil && !settings.PureBoot(caller.GetMemIndex()) {
			// 	apple2helpers.PutStr(caller, "Saved "+this.textFile)
			// }
		} else if this.textMode {
			tmp := this.edit.GetRawContent()
			sl := ""
			for _, rs := range tmp {
				if sl != "" {
					sl += "\r\n"
				}
				sl += string(rs.Runes)
			}

			//			fmt.Println(sl)
			// save it
			files.WriteBytesViaProvider(files.GetPath(this.textFile), files.GetFilename(this.textFile), []byte(sl))
			// if err == nil {
			// 	apple2helpers.PutStr(caller, "Saved "+this.textFile)
			// }
		} else if caller.GetDialect().GetTitle() == "Logo" {

			//pre-erase functions
			//caller.GetDialect().CleanDynaCommandsByName(logoProcErase)

			lines := strings.Split(this.edit.GetText(), "\r\n")

			// quick parse...
			for i, l := range lines {
				//log2.Printf("%d\t%s", i, l)
				err := caller.GetDialect().SyntaxValid(l)
				if err != nil {
					apple2helpers.BeepError(caller)
					gotoLine = i
					gotoError = err.Error()
					text = this.edit.GetText()
					this.save = false
					goto rerunedit
				}
			}

			caller.GetDialect().SaveState()
			settings.LogoSuppressDefines[caller.GetMemIndex()] = true
			for i, l := range lines {
				//log2.Printf("%d\t%s", i, l)
				err := caller.GetDialect().Parse(caller, l)
				if err != nil {
					apple2helpers.BeepError(caller)
					gotoLine = i
					gotoError = err.Error()
					text = this.edit.GetText()
					this.save = false
					goto rerunedit
				}
			}
			settings.LogoSuppressDefines[caller.GetMemIndex()] = false
			caller.GetDialect().RestoreState()
		} else {

			caller.SetCode(types.NewAlgorithm())
			tmp := this.edit.GetRawContent()
			s := time.Now()
			fmt.Printf("EDIT parsed by %s\n", caller.GetName())
			caller.GetDialect().SetSkipMemParse(true)
			for _, ss := range tmp {
				fmt.Println("*** ", string(ss.Runes), ss.Runes)
				if len(ss.Runes) > 0 {
					caller.Parse(string(ss.Runes))
				}
			}
			caller.GetDialect().SetSkipMemParse(false)
			fmt.Printf("Parse completed in %v\n", time.Since(s))

			if settings.AutosaveFilename[caller.GetMemIndex()] != "" {
				data := caller.GetDialect().GetWorkspace(caller)
				files.AutoSave(caller.GetMemIndex(), data)
			}
		}
	} else {
		if files.ExistsViaProvider(files.GetPath(this.textFile), files.GetFilename(this.textFile)) {
			// need to unlock it
			_ = files.UnlockViaProvider(files.GetPath(this.textFile), files.GetFilename(this.textFile))
		}
	}

	settings.DisableMetaMode[caller.GetMemIndex()] = false

	return 0, nil
}

func (this *StandardCommandEDIT) OnEditorExit(edit *editor.CoreEdit) {
	apple2helpers.Clearscreen(edit.Int)
}

func (this *StandardCommandEDIT) OnEditorBegin(edit *editor.CoreEdit) {
	//edit.Int.PutStr(string(rune(7)))
	edit.Changed = false
}

func (this *StandardCommandEDIT) OnEditorChange(edit *editor.CoreEdit) {

}

func (this *StandardCommandEDIT) OnEditorMove(edit *editor.CoreEdit) {

}

func (this *StandardCommandEDIT) OnMouseMove(edit *editor.CoreEdit, x, y int) {

	if edit.MBL {

		if !edit.SelMode {
			// start selmode here
			edit.SelectMode(true)
		}

		// move cursor
		tl := edit.Voffset + int(y-1)
		tc := edit.Hoffset + int(x)
		if tl < len(edit.Content) {

			edit.GotoLine(tl)
			edit.GotoColumn(tc)
			edit.Display()

		}

	}

}

func (this *StandardCommandEDIT) OnMouseClick(edit *editor.CoreEdit, left, right bool) {

	fmt.Printf("Mouse Buttons at (%v, %v)\n", left, right)

	if left && edit.MY == 0 {

		edit.Int.GetMemoryMap().KeyBufferAdd(edit.Int.GetMemIndex(), vduconst.PAGE_UP)

	} else if left && edit.MY > 0 && edit.MY < 46 {

		tl := edit.Voffset + int(edit.MY-1)
		tc := edit.Hoffset + int(edit.MX)
		if tl < len(edit.Content) {

			edit.GotoLine(tl)
			edit.GotoColumn(tc)
			edit.Display()

		}

	} else if left && edit.MY >= 46 {

		edit.Int.GetMemoryMap().KeyBufferAdd(edit.Int.GetMemIndex(), vduconst.PAGE_DOWN)

	}

}

func (this *StandardCommandEDIT) OnEditorKeypress(edit *editor.CoreEdit, ch rune) bool {
	//System.Err.Println("Got editor keypress: "+ch);

	if ch == 27 && edit.SelMode {
		edit.SelectMode(false)
		return true
	}

	if ch == vduconst.DELETE {
		this.edit.DeleteCurrentLine()
	}

	if ch == vduconst.INSERT {
		this.edit.Overwrite = !this.edit.Overwrite
	}

	//~ if ch == vduconst.PASTE || ch == vduconst.SHIFT_CTRL_V {
	//~ // paste
	//~ text, err := clipboard.ReadAll()
	//~ if err == nil {
	//~ rs := runestring.NewRuneString()
	//~ rs.Append(text)
	//~ edit.Int.SetBuffer(rs)
	//~ }
	//~ return true
	//~ }

	if this.textMode {

		// Double toggle control key method
		if ch >= vduconst.COLOR0 && ch <= vduconst.COLOR15 {
			cval := ch - vduconst.COLOR0
			edit.SetFGColor(cval)

			return true
		} else if ch >= vduconst.BGCOLOR0 && ch <= vduconst.BGCOLOR15 {
			cval := ch - vduconst.BGCOLOR0

			edit.SetBGColor(cval)
			return true
		} else if ch >= vduconst.SHADE0 && ch <= vduconst.SHADE7 {
			//fmt.Println("ALT + number")
			cval := (ch - vduconst.SHADE0)
			edit.SetShade(cval)
			return true
		} else if ch == vduconst.CTRL_SPACE {
			fg := edit.FGColor
			bg := edit.BGColor
			edit.FGColor = bg
			edit.BGColor = fg
			return true
		}

	}

	return false
}

func EditorExit(this *editor.CoreEdit) {

	if this.Changed {
		// this.Int.SetMemory(37, uint64(this.ReservedTop+this.Height))
		// this.Int.SetMemory(36, 1)

		apple2helpers.Gotoxy(
			this.Int,
			1*this.HSkip,
			(this.ReservedTop+this.Height)*this.VSkip,
		)

		menu := YesNoCancelMenu()
		resp := this.Choice("Save changes?", menu)
		if resp.Key == 'Y' || resp.Key == 'y' {
			e := this.Parent.(*StandardCommandEDIT)
			e.save = true
		}
		if resp.Key == 'C' || resp.Key == 'c' {
			return
		}
	}

	this.Running = false
}

func EditorUncut(this *editor.CoreEdit) {
	this.UncutLines()
}

func EditorCut(this *editor.CoreEdit) {
	this.CutLine()
}

func EditorAutoIndent(this *editor.CoreEdit) {
	this.AutoIndent = !this.AutoIndent
	UpdateSubTitle(this)
}

func UpdateSubTitle(this *editor.CoreEdit) {
	s := []string(nil)
	if this.Overwrite {
		s = append(s, "OVR")
	} else {
		s = append(s, "INS")
	}
	if this.AutoIndent {
		s = append(s, "IND")
	}
	this.SubTitle = strings.Join(s, ",")
}

func EditorToggleInsert(this *editor.CoreEdit) {
	this.Overwrite = !this.Overwrite
	UpdateSubTitle(this)
}

func EditorChar(this *editor.CoreEdit) {
	fg := GetCharacter("Character:", this)
	if fg != -1 {
		if this.Overwrite {
			this.CursorOverwriteChar(rune(fg))
		} else {
			this.CursorInsertChar(rune(fg))
		}
	}
}

func EditorFG(this *editor.CoreEdit) {
	fg := GetColor("Foreground:", this)
	if fg != -1 {
		this.SetFGColor(rune(fg))
	}
}

func EditorBG(this *editor.CoreEdit) {
	fg := GetColor("Background:", this)
	if fg != -1 {
		this.SetBGColor(rune(fg))
	}
}

func EditorSave(this *editor.CoreEdit) {
	e := this.Parent.(*StandardCommandEDIT)
	e.save = true
	this.Running = false
}

func EditorUndo(this *editor.CoreEdit) {
	this.Undo()
}

func EditorCopy(this *editor.CoreEdit) {
	this.CopySelection()
	this.SelectMode(false)
}

func EditorCutSel(this *editor.CoreEdit) {
	this.CopySelection()
	this.DelSelection()
	this.SelectMode(false)
}

func EditorPaste(this *editor.CoreEdit) {
	fmt.Println("EditorPaste")
	this.DelSelection()
	this.SelectMode(false)
	this.Paste()
}

func EditorRedo(this *editor.CoreEdit) {
	this.Redo()
}

func EditorSelect(this *editor.CoreEdit) {
	this.SelectMode(!this.SelMode)
}

func EditorFind(this *editor.CoreEdit) {
	this.Int.SetMemory(37, uint64(this.ReservedTop+this.Height))
	this.Int.SetMemory(36, 0)
	apple2helpers.ClearToBottom(this.Int)
	term := this.GetCRTLine("Find (" + this.SearchTerm + "):")
	if term != "" {
		line, col := this.SearchForward(term)
		if line != -1 {
			this.GotoLine(line)
			this.GotoColumn(col)
		} else {
			r, g, b, a := this.Int.GetMemoryMap().GetBGColor(this.Int.GetMemIndex())
			this.Int.GetMemoryMap().SetBGColor(this.Int.GetMemIndex(), 20, 20, 20, 255)
			apple2helpers.BeepError(this.Int)
			this.Int.GetMemoryMap().SetBGColor(this.Int.GetMemIndex(), r, g, b, a)
		}
	} else if this.SearchTerm != "" {
		line, col := this.SearchForward(this.SearchTerm)
		if line != -1 {
			this.GotoLine(line)
			this.GotoColumn(col)
		} else {
			r, g, b, a := this.Int.GetMemoryMap().GetBGColor(this.Int.GetMemIndex())
			this.Int.GetMemoryMap().SetBGColor(this.Int.GetMemIndex(), 20, 20, 20, 255)
			apple2helpers.BeepError(this.Int)
			this.Int.GetMemoryMap().SetBGColor(this.Int.GetMemIndex(), r, g, b, a)
		}
	}
	this.Int.SetMemory(37, uint64(this.ReservedTop+this.Height))
	this.Int.SetMemory(36, 0)
	apple2helpers.ClearToBottom(this.Int)
}

func EditorToggle4080(edit *editor.CoreEdit) {
	if apple2helpers.GetColumns(edit.Int) == 80 {
		//edit.Int.SetVideoMode(edit.Int.GetVideoModes()[5])
		apple2helpers.MODE40(edit.Int)
		edit.ReservedBot = 4
		edit.VSkip = 2
		edit.HSkip = 2
		edit.ResetDimensions()
		edit.Display()

	} else {
		//edit.Int.SetVideoMode(edit.Int.GetVideoModes()[0])
		apple2helpers.TEXTMAX(edit.Int)
		edit.ReservedBot = 2
		edit.VSkip = 1
		edit.HSkip = 1
		edit.ResetDimensions()
		edit.Display()
	}
}

func YesNoMenu() map[rune]*editor.CoreEditHook {
	m := make(map[rune]*editor.CoreEditHook)

	m['Y'] = &editor.CoreEditHook{
		Key:         'Y',
		Description: "Yes",
		Hook:        nil,
		Visible:     true,
	}
	m['N'] = &editor.CoreEditHook{
		Key:         'N',
		Description: "No",
		Hook:        nil,
		Visible:     true,
	}

	return m
}

func YesNoCancelMenu() map[rune]*editor.CoreEditHook {
	m := make(map[rune]*editor.CoreEditHook)

	m['Y'] = &editor.CoreEditHook{
		Key:         'Y',
		Description: "Yes",
		Hook:        nil,
		Visible:     true,
	}
	m['N'] = &editor.CoreEditHook{
		Key:         'N',
		Description: "No",
		Hook:        nil,
		Visible:     true,
	}
	m['C'] = &editor.CoreEditHook{
		Key:         'C',
		Description: "Cancel",
		Hook:        nil,
		Visible:     true,
	}

	return m
}

func GetColor(prompt string, this *editor.CoreEdit) int {

	items := []editor.GridItem{
		editor.GridItem{FG: 15, BG: 0, Label: "  ", ID: 0},
		editor.GridItem{FG: 15, BG: 1, Label: "  ", ID: 1},
		editor.GridItem{FG: 15, BG: 2, Label: "  ", ID: 2},
		editor.GridItem{FG: 15, BG: 3, Label: "  ", ID: 3},
		editor.GridItem{FG: 15, BG: 4, Label: "  ", ID: 4},
		editor.GridItem{FG: 15, BG: 5, Label: "  ", ID: 5},
		editor.GridItem{FG: 15, BG: 6, Label: "  ", ID: 6},
		editor.GridItem{FG: 15, BG: 7, Label: "  ", ID: 7},
		editor.GridItem{FG: 15, BG: 8, Label: "  ", ID: 8},
		editor.GridItem{FG: 15, BG: 9, Label: "  ", ID: 9},
		editor.GridItem{FG: 15, BG: 10, Label: "  ", ID: 10},
		editor.GridItem{FG: 15, BG: 11, Label: "  ", ID: 11},
		editor.GridItem{FG: 15, BG: 12, Label: "  ", ID: 12},
		editor.GridItem{FG: 15, BG: 13, Label: "  ", ID: 13},
		editor.GridItem{FG: 15, BG: 14, Label: "  ", ID: 14},
		editor.GridItem{FG: 15, BG: 15, Label: "  ", ID: 15},
	}

	gi := this.GridChooser(
		prompt,
		items,
		22,
		9,
		4,
		2,
	)

	return gi

}

func GetCharacter(prompt string, this *editor.CoreEdit) int {

	items := make([]editor.GridItem, 0)

	fg := 15
	bg := 2

	if settings.HighContrastUI {
		bg = 0
	}

	for i := 1024; i < 1024+128; i++ {
		items = append(items,
			editor.GridItem{
				ID:    i,
				FG:    fg,
				BG:    bg,
				Label: string(rune(i)),
			},
		)
	}

	gi := this.GridChooser(
		prompt,
		items,
		22,
		4,
		3,
		2,
	)

	return gi

}

// ----------------------------------------------------------------------------
// RenderTokens will take a raw line (hr.Data), and return a matching color set
// for it.
// ----------------------------------------------------------------------------
//func RenderTokens(rawline runestring.RuneString, dialect *interfaces.Dialecter) runestring.RuneString {

//	chunk := ""

//	for i, ch := range rawline {
//		switch ch {
//		case " ":
//			if chunk != "" {
//				tl := dialect.Tokenize(chunk)
//			}
//		}
//	}

//}

func ShowKeymap(this *editor.CoreEdit) {

	data, e := files.ReadBytesViaProvider("/system/help", "keymap.t")
	if e != nil {
		return
	}

	ue := utils.Unescape(string(data.Content))
	lines := strings.Split(ue, "\r\n")

	vpos := (apple2helpers.GetRows(this.Int) - len(lines)) / 2
	this.Int.SetMemory(37, uint64(vpos))
	this.Int.SetMemory(36, 0)
	this.Int.PutStr(ue)

	for this.Int.GetMemoryMap().KeyBufferGetLatest(this.Int.GetMemIndex()) != 27 {
		time.Sleep(50 * time.Millisecond)
	}
}
