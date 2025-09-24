package plus

import (
	"time"

	"github.com/atotto/clipboard"
	s8webclient "paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/filerecord"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type PlusFeedback struct {
	dialect.CoreFunction
	textMode bool
	textFile string
	save     bool
	edit     *editor.CoreEdit
}

func (this *PlusFeedback) FunctionExecute(params *types.TokenList) error {

	settings.DisableMetaMode[this.Interpreter.GetMemIndex()] = true
	defer func() {
		settings.DisableMetaMode[this.Interpreter.GetMemIndex()] = false
	}()

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	caller := this.Interpreter

	this.textMode = false

	// Purge keys
	caller.SetBuffer(runestring.NewRuneString())

	gotoLine := 0

	//  h := tt.GetHighIndex()

	text := ""

	this.save = false

	this.textMode = true
	text = ""

	apple2helpers.MonitorPanel(caller, true)
	apple2helpers.TEXTMAX(caller)
	apple2helpers.Clearscreen(caller)

	caller.SetIgnoreSpecial(true)

	//	caller.GetVDU().SaveVDUState()
	//	caller.GetVDU().SetVideoMode(caller.GetVDU().GetVideoModes()[0])
	this.edit = editor.NewCoreEdit(caller, "EDIT (F2=Accept Changes, ESC=Cancel)", text, true, !this.textMode)
	this.edit.ShowSwatch = this.textMode

	if this.textMode {
		this.edit.BarFG = 15
		this.edit.BarBG = 6
		this.edit.BGColor = 2
		this.edit.FGColor = 15
		this.edit.SelBG = 2
		this.edit.SelFG = 12
		this.edit.Title = "Please enter your feedback... "
	}

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_Q,
		"Quit",
		true,
		FeedBackExit,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_K,
		"Kill",
		true,
		FeedBackCut,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_U,
		"Unkill",
		true,
		FeedBackUncut,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_O,
		"Save",
		true,
		FeedBackSave,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_Y,
		"Redo",
		true,
		FeedBackRedo,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_Z,
		"Undo",
		true,
		FeedBackUndo,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_S,
		"Select",
		true,
		FeedBackSelect,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_C,
		"Copy",
		true,
		FeedBackCopy,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_X,
		"Cut",
		true,
		FeedBackCutSel,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_V,
		"Paste",
		true,
		FeedBackPaste,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_W,
		"Where",
		true,
		FeedBackFind,
	)

	//~ this.edit.RegisterCommand(
	//~ vduconst.SHIFT_CTRL_F,
	//~ "FgCol",
	//~ true,
	//~ FeedBackFG,
	//~ )

	//~ this.edit.RegisterCommand(
	//~ vduconst.SHIFT_CTRL_B,
	//~ "BgCol",
	//~ true,
	//~ FeedBackBG,
	//~ )

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_E,
		"Chars",
		true,
		FeedBackChar,
	)

	this.edit.Parent = this

	if this.textMode {
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
	apple2helpers.MonitorPanel(caller, false)
	caller.SetIgnoreSpecial(false)

	if this.save {
		CreateFeedback(caller, "", this.edit.GetText())
		caller.PutStr("Your feedback has been sent...\r\n")
	}

	return nil
}

func (this *PlusFeedback) OnEditorExit(edit *editor.CoreEdit) {
	apple2helpers.Clearscreen(edit.Int)
}

func (this *PlusFeedback) OnEditorBegin(edit *editor.CoreEdit) {
	//edit.Int.PutStr(string(rune(7)))
	edit.Changed = false
}

func (this *PlusFeedback) OnEditorChange(edit *editor.CoreEdit) {

}

func (this *PlusFeedback) OnEditorMove(edit *editor.CoreEdit) {

}

func (this *PlusFeedback) OnMouseMove(edit *editor.CoreEdit, x, y int) {
}

func (this *PlusFeedback) OnMouseClick(edit *editor.CoreEdit, left, right bool) {
}

func (this *PlusFeedback) OnEditorKeypress(edit *editor.CoreEdit, ch rune) bool {
	//System.Err.Println("Got FeedBack keypress: "+ch);

	if ch == vduconst.DELETE {
		this.edit.DeleteCurrentLine()
	}

	if ch == vduconst.INSERT {
		this.edit.Overwrite = !this.edit.Overwrite
	}

	if ch == vduconst.PASTE {
		// paste
		text, err := clipboard.ReadAll()
		if err == nil {
			rs := runestring.NewRuneString()
			rs.Append(text)
			edit.Int.SetBuffer(rs)
		}
		return true
	}

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
		}

	}

	return false
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

func FeedBackExit(this *editor.CoreEdit) {

	if this.Changed {
		this.Int.SetMemory(37, uint64(this.ReservedTop+this.Height))
		this.Int.SetMemory(36, 1)
		menu := YesNoCancelMenu()
		resp := this.Choice("Save and send your feedback?", menu)
		if resp.Key == 'Y' {
			e := this.Parent.(*PlusFeedback)
			e.save = true
		}
		if resp.Key == 'C' {
			return
		}
	}

	this.Running = false
}

func FeedBackUncut(this *editor.CoreEdit) {
	this.UncutLines()
}

func FeedBackCut(this *editor.CoreEdit) {
	this.CutLine()
}

func GetCharacter(prompt string, this *editor.CoreEdit) int {

	items := make([]editor.GridItem, 0)

	for i := 1024; i < 1024+128; i++ {
		items = append(items,
			editor.GridItem{
				ID:    i,
				FG:    15,
				BG:    2,
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

func FeedBackChar(this *editor.CoreEdit) {
	fg := GetCharacter("Character:", this)
	if fg != -1 {
		if this.Overwrite {
			this.CursorOverwriteChar(rune(fg))
		} else {
			this.CursorInsertChar(rune(fg))
		}
	}
}

func FeedBackSave(this *editor.CoreEdit) {
	e := this.Parent.(*PlusFeedback)
	e.save = true
	this.Running = false
}

func FeedBackUndo(this *editor.CoreEdit) {
	this.Undo()
}

func FeedBackCopy(this *editor.CoreEdit) {
	this.CopySelection()
	this.SelectMode(false)
}

func FeedBackCutSel(this *editor.CoreEdit) {
	this.CopySelection()
	this.DelSelection()
	this.SelectMode(false)
}

func FeedBackPaste(this *editor.CoreEdit) {
	this.DelSelection()
	this.SelectMode(false)
	this.Paste()
}

func FeedBackRedo(this *editor.CoreEdit) {
	this.Redo()
}

func FeedBackSelect(this *editor.CoreEdit) {
	this.SelectMode(!this.SelMode)
}

func FeedBackFind(this *editor.CoreEdit) {
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

func FeedBackToggle4080(edit *editor.CoreEdit) {
	if apple2helpers.GetColumns(edit.Int) == 80 {
		//edit.Int.SetVideoMode(edit.Int.GetVideoModes()[5])
		apple2helpers.TEXT40(edit.Int)
		edit.Width = 40
		edit.Display()
	} else {
		//edit.Int.SetVideoMode(edit.Int.GetVideoModes()[0])
		apple2helpers.TEXT80(edit.Int)
		edit.Width = 80
		edit.Display()
	}
}

func CreateFeedback(this interfaces.Interpretable, summary string, body string) {

	if this.GetDiskImage() != "" {
		body = "Disk image: " + this.GetDiskImage() + "\r\n" + body
	}

	if summary == "" {
		summary = "Feedback from " + s8webclient.CONN.Username
	}
	capture := false

	bug := filerecord.BugReport{
		Summary: summary,
		Body:    body,
		Created: time.Now(),
		Creator: s8webclient.CONN.Username,
		Type:    filerecord.BT_BUG,
	}
	bug.Filename = this.GetFileRecord().FileName
	bug.Filepath = this.GetFileRecord().FilePath

	if capture {
		tmp, _ := this.FreezeBytes()
		att := filerecord.BugAttachment{
			Created: time.Now(),
			Content: utils.GZIPBytes(tmp),
			Name:    "Compressed run state",
		}
		attlist := filerecord.BugAttachment{
			Created: time.Now(),
			Content: utils.GZIPBytes([]byte(this.GetCode().String())),
			Name:    "LISTING",
		}
		bug.Attachments = []filerecord.BugAttachment{att, attlist}
	}

	_ = s8webclient.CONN.CreateUpdateBug(bug)

}

func (this *PlusFeedback) Syntax() string {

	/* vars */
	var result string

	result = "ALTCASE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusFeedback) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusFeedback(a int, b int, params types.TokenList) *PlusFeedback {
	this := &PlusFeedback{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"

	this.MaxParams = 1
	this.MinParams = 0

	return this
}
