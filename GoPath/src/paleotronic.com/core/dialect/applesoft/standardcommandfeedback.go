package applesoft

import (
	"strings"
	"time"

	s8webclient "paleotronic.com/api"
	"paleotronic.com/fmt"

	"github.com/atotto/clipboard"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"

	//	"paleotronic.com/files"
	//	"paleotronic.com/log"

	"paleotronic.com/filerecord"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type StandardCommandFEEDBACK struct {
	dialect.Command
	selected int
	save     bool
	textMode bool
	textFile string
	execute  int
	edit     *editor.CoreEdit
	Scope    *types.Algorithm
}

func NewStandardCommandFEEDBACK() *StandardCommandFEEDBACK {
	this := &StandardCommandFEEDBACK{}
	this.ImmediateMode = true
	return this
}

func (this *StandardCommandFEEDBACK) Syntax() string {
	// TODO Auto-generated method stub
	return "EDIT"
}

func (this *StandardCommandFEEDBACK) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	this.textMode = false

	// Purge keys
	caller.SetBuffer(runestring.NewRuneString())

	gotoLine := 0

	if tokens.Size() > 0 {
		s := ""
		for _, tt := range tokens.Content {
			s += tt.Content
		}
		tokens = *types.NewTokenList()

		if utils.StrToInt(s) != 0 {
			tokens.Push(types.NewToken(types.NUMBER, s))
		} else {
			tokens.Push(types.NewToken(types.STRING, s))
		}

		t, _ := caller.GetDialect().ParseTokensForResult(caller, tokens)

		if t.IsNumeric() {
			gotoLine = t.AsInteger()
		} else {
			this.textMode = true
			wd := strings.Trim(caller.GetWorkDir(), "/")
			this.textFile = t.Content
			if wd != "" && rune(this.textFile[0]) != '/' {
				this.textFile = wd + "/" + t.Content
			}
			this.textFile = strings.ToLower(this.textFile)
			fmt.Println("filename =", this.textFile)
		}
	}

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

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_H,
		"KeyHelp",
		true,
		ShowKeymap,
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

	return 0, nil
}

func (this *StandardCommandFEEDBACK) OnEditorExit(edit *editor.CoreEdit) {
	apple2helpers.Clearscreen(edit.Int)
}

func (this *StandardCommandFEEDBACK) OnEditorBegin(edit *editor.CoreEdit) {
	//edit.Int.PutStr(string(rune(7)))
	edit.Changed = false
}

func (this *StandardCommandFEEDBACK) OnEditorChange(edit *editor.CoreEdit) {

}

func (this *StandardCommandFEEDBACK) OnEditorMove(edit *editor.CoreEdit) {

}

func (this *StandardCommandFEEDBACK) OnMouseMove(edit *editor.CoreEdit, x, y int) {
}

func (this *StandardCommandFEEDBACK) OnMouseClick(edit *editor.CoreEdit, left, right bool) {
}

func (this *StandardCommandFEEDBACK) OnEditorKeypress(edit *editor.CoreEdit, ch rune) bool {
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

func FeedBackExit(this *editor.CoreEdit) {

	if this.Changed {
		this.Int.SetMemory(37, uint64(this.ReservedTop+this.Height))
		this.Int.SetMemory(36, 1)
		menu := YesNoCancelMenu()
		resp := this.Choice("Save and send your feedback?", menu)
		if resp.Key == 'Y' {
			e := this.Parent.(*StandardCommandFEEDBACK)
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

func FeedBackFG(this *editor.CoreEdit) {
	fg := GetColor("Foreground:", this)
	if fg != -1 {
		this.SetFGColor(rune(fg))
	}
}

func FeedBackBG(this *editor.CoreEdit) {
	fg := GetColor("Background:", this)
	if fg != -1 {
		this.SetBGColor(rune(fg))
	}
}

func FeedBackSave(this *editor.CoreEdit) {
	e := this.Parent.(*StandardCommandFEEDBACK)
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
