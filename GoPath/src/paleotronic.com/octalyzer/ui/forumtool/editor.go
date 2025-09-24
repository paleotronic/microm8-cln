package forumtool

import (
	"errors"

	"github.com/atotto/clipboard"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/runestring"
)

func (this *ForumApp) OnEditorChange(edit *editor.CoreEdit) {

}

func (this *ForumApp) OnEditorMove(edit *editor.CoreEdit) {

}

func (this *ForumApp) OnMouseMove(edit *editor.CoreEdit, x, y int) {
}

func (this *ForumApp) OnMouseClick(edit *editor.CoreEdit, left, right bool) {
}

func (this *ForumApp) OnEditorKeypress(edit *editor.CoreEdit, ch rune) bool {
	//System.Err.Println("Got Editor keypress: "+ch);

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

	//if this.textMode {

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

	//}

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

func EditorExit(this *editor.CoreEdit) {

	if this.Changed {
		this.Int.SetMemory(37, uint64(this.ReservedTop+this.Height))
		this.Int.SetMemory(36, 1)
		menu := YesNoCancelMenu()
		resp := this.Choice("Save and send your feedback?", menu)
		if resp.Key == 'Y' {
			e := this.Parent.(*ForumApp)
			e.save = true
		}
		if resp.Key == 'C' {
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

func EditorSave(this *editor.CoreEdit) {
	e := this.Parent.(*ForumApp)
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

func (this *ForumApp) OnEditorExit(edit *editor.CoreEdit) {
	apple2helpers.Clearscreen(edit.Int)
}

func (this *ForumApp) OnEditorBegin(edit *editor.CoreEdit) {
	//edit.Int.PutStr(string(rune(7)))
	edit.Changed = false
}

func (this *ForumApp) editMessage(forumId int32, parentId int32, text string) (string, error) {

	settings.DisableMetaMode[this.Int.GetMemIndex()] = true
	defer func() {
		settings.DisableMetaMode[this.Int.GetMemIndex()] = false
	}()

	caller := this.Int

	// Purge keys
	caller.SetBuffer(runestring.NewRuneString())

	gotoLine := 0

	//  h := tt.GetHighIndex()

	apple2helpers.MonitorPanel(caller, true)
	apple2helpers.TEXTMAX(caller)
	apple2helpers.Clearscreen(caller)

	caller.SetIgnoreSpecial(true)

	//	caller.GetVDU().SaveVDUState()
	//	caller.GetVDU().SetVideoMode(caller.GetVDU().GetVideoModes()[0])
	this.edit = editor.NewCoreEdit(caller, "EDIT (F2=Accept Changes, ESC=Cancel)", text, true, false)
	this.edit.ShowSwatch = true

	this.edit.BarFG = 15
	this.edit.BarBG = 6
	this.edit.BGColor = 0
	this.edit.FGColor = 15
	this.edit.SelBG = 2
	this.edit.SelFG = 12
	this.edit.Title = "Please enter your message... "

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_Q,
		"Quit",
		true,
		EditorExit,
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
		vduconst.SHIFT_CTRL_W,
		"Where",
		true,
		EditorFind,
	)

	// ~ this.edit.RegisterCommand(
	// ~ vduconst.SHIFT_CTRL_F,
	// ~ "FgCol",
	// ~ true,
	// ~ EditorFG,
	// ~ )

	// ~ this.edit.RegisterCommand(
	// ~ vduconst.SHIFT_CTRL_B,
	// ~ "BgCol",
	// ~ true,
	// ~ EditorBG,
	// ~ )

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_E,
		"Chars",
		true,
		EditorChar,
	)

	this.edit.Parent = this
	this.edit.ShowColor = true
	this.edit.ShowOverwrite = true

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
		// publish
		return this.edit.GetText(), nil
	}

	return "", errors.New("Edit aborted")

}
