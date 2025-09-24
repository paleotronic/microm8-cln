package plus

import (
	"paleotronic.com/fmt"

	"paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/filerecord"
	"paleotronic.com/utils"
	//    "strings"
	"time"
)

type PlusBugList struct {
	dialect.CoreFunction
	Type filerecord.BugType
	edit *editor.CoreEdit
	list []filerecord.BugReport
	bug  *filerecord.BugReport
	show *PlusBugShow
}

func BuglistQuit(this *editor.CoreEdit) {

	this.Done()

}

func BuglistClose(this *editor.CoreEdit) {

	bl := this.Parent.(*PlusBugList)

	bl.Close("Bug closed by " + s8webclient.CONN.Username)
	bl.Refresh()

}

func BuglistFix(this *editor.CoreEdit) {

	bl := this.Parent.(*PlusBugList)

	bl.Fix()
	bl.Refresh()

}

func BuglistAssign(this *editor.CoreEdit) {

	bl := this.Parent.(*PlusBugList)

	person := this.GetCRTLine("Reassign to: ")

	bl.ReAssign(person)
	bl.Refresh()

}

func (this *PlusBugList) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if this.show == nil {
		this.show = NewPlusBugShow(0, 0, *types.NewTokenList())
		this.show.Interpreter = this.Interpreter
	}

	apple2helpers.MonitorPanel(this.Interpreter, true)
	apple2helpers.TEXTMAX(this.Interpreter)

	this.edit = editor.NewCoreEdit(this.Interpreter, this.Type.String()+" list", "", false, true)
	this.edit.BarBG = 1
	this.edit.BarFG = 15
	this.edit.BGColor = 8
	this.edit.FGColor = 15
	this.Refresh()
	this.edit.SetEventHandler(this)
	this.edit.Title = this.Type.String() + " list "

	this.edit.Parent = this

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_Q,
		"Quit",
		true,
		BuglistQuit,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_F,
		"Fixed",
		true,
		BuglistFix,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_C,
		"Close",
		true,
		BuglistClose,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_A,
		"Assign",
		true,
		BuglistAssign,
	)

	this.edit.Run()

	apple2helpers.MonitorPanel(this.Interpreter, false)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusBugList) Refresh() {
	this.list, _ = s8webclient.CONN.GetBugList(this.Type)

	var out string

	if len(this.list) == 0 {
		out += "No bugs either created by or assigned to you.\r\n"
	} else {

		out += fmt.Sprintf("%-5s %-10s %-10s %-3s %-3s %-40s\r\n", "ID", "State", "Assigned", "Com", "Att", "Summary")

		for _, bug := range this.list {

			out += fmt.Sprintf("%-5d %-10s %-10s %-3d %-3d %-40s\r\n", bug.DefectID, bug.State.String(), bug.Assigned, len(bug.Comments), len(bug.Attachments), bug.Summary)

		}

	}

	this.edit.SetText(out)

	// do thing
	this.OnEditorMove(this.edit)
}

func (this *PlusBugList) Syntax() string {

	/* vars */
	var result string

	result = "EXIT{}"

	/* enforce non void return */
	return result

}

func (this *PlusBugList) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusBugList(a int, b int, params types.TokenList) *PlusBugList {
	this := &PlusBugList{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "EXIT"
	this.Type = filerecord.BugType(a)

	return this
}

func (this *PlusBugList) OnEditorMove(edit *editor.CoreEdit) {
	if edit.Line > 0 {
		l := edit.Line - 1
		if l < len(this.list) {
			this.bug = &this.list[l]
			this.edit.Title = fmt.Sprintf(this.Type.String()+" list (%d selected) - Enter: SHOW", this.bug.DefectID)
		} else {
			this.bug = nil
			this.edit.Title = fmt.Sprintf(this.Type.String() + " list - None selected")
		}
	} else {
		this.bug = nil
		this.edit.Title = fmt.Sprintf(this.Type.String() + " list - None selected")
	}
}

func (this *PlusBugList) OnEditorBegin(edit *editor.CoreEdit) {
}

func (this *PlusBugList) OnEditorChange(edit *editor.CoreEdit) {
}

func (this *PlusBugList) OnEditorExit(edit *editor.CoreEdit) {
}

func (this *PlusBugList) OnMouseMove(edit *editor.CoreEdit, x, y int) {
	fmt.Printf("Mouse at (%d, %d)\n", x, y)
}

func (this *PlusBugList) OnMouseClick(edit *editor.CoreEdit, left, right bool) {

	fmt.Printf("Mouse Buttons at (%v, %v)\n", left, right)

	if left && edit.MY == 0 {

		edit.Int.GetMemoryMap().KeyBufferAdd(edit.Int.GetMemIndex(), vduconst.PAGE_UP)

	} else if left && edit.MY > 0 && edit.MY < 46 {

		tl := edit.Voffset + int(edit.MY-1)
		if tl < len(edit.Content) {

			edit.GotoLine(tl)

			edit.Int.GetMemoryMap().KeyBufferAdd(edit.Int.GetMemIndex(), 13)

			fmt.Println("Clicky")

		}

	} else if left && edit.MY >= 46 {

		edit.Int.GetMemoryMap().KeyBufferAdd(edit.Int.GetMemIndex(), vduconst.PAGE_DOWN)

	}

}
func (this *PlusBugList) OnEditorKeypress(edit *editor.CoreEdit, ch rune) bool {

	// lowercase the rune
	if ch >= 'A' && ch <= 'Z' {
		ch += 32
	}

	if ch == 13 && this.bug != nil {
		tl := types.NewTokenList()
		tl.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(this.bug.DefectID))))
		this.show.FunctionExecute(tl)
		apple2helpers.MonitorPanel(this.Interpreter, true)
		this.Refresh()
	} else if ch == 'c' && this.bug != nil {
		this.Close("Bug closed by " + s8webclient.CONN.Username)
		this.Refresh()
	} else if ch == 't' && this.bug != nil {
		this.Test()
		this.Refresh()
	} else if ch == 'f' && this.bug != nil {
		this.Working()
		this.Refresh()
	} else if ch == 'o' && this.bug != nil {
		this.Open()
		this.Refresh()
	}

	return false
}

func (this *PlusBugList) ReAssign(target string) {
	this.bug.Assigned = target
	_ = s8webclient.CONN.CreateUpdateBug(*this.bug)
	this.Refresh()
}

func (this *PlusBugList) Close(comment string) {
	this.bug.State = filerecord.BS_CLOSED
	this.bug.Comments = append(
		this.bug.Comments,
		filerecord.BugComment{
			Created: time.Now(),
			User:    s8webclient.CONN.Username,
			Content: comment,
		},
	)
	_ = s8webclient.CONN.CreateUpdateBug(*this.bug)
	this.Refresh()
}

func (this *PlusBugList) Comment(comment string) {
	this.bug.Comments = append(
		this.bug.Comments,
		filerecord.BugComment{
			Created: time.Now(),
			User:    s8webclient.CONN.Username,
			Content: comment,
		},
	)
	_ = s8webclient.CONN.CreateUpdateBug(*this.bug)
	this.Refresh()
}

func (this *PlusBugList) Working() {
	this.bug.State = filerecord.BS_INPROGRESS
	this.bug.Assigned = s8webclient.CONN.Username
	this.bug.Comments = append(
		this.bug.Comments,
		filerecord.BugComment{
			Created: time.Now(),
			User:    s8webclient.CONN.Username,
			Content: "State changed to In-Progress",
		},
	)
	_ = s8webclient.CONN.CreateUpdateBug(*this.bug)
	this.Refresh()
}

func (this *PlusBugList) Test() {
	this.bug.State = filerecord.BS_RETEST
	this.bug.Assigned = s8webclient.CONN.Username
	this.bug.Comments = append(
		this.bug.Comments,
		filerecord.BugComment{
			Created: time.Now(),
			User:    s8webclient.CONN.Username,
			Content: "State changed to Retest",
		},
	)
	_ = s8webclient.CONN.CreateUpdateBug(*this.bug)
	this.Refresh()
}

func (this *PlusBugList) Fix() {
	this.bug.State = filerecord.BS_FIXED
	this.bug.Assigned = s8webclient.CONN.Username
	this.bug.Comments = append(
		this.bug.Comments,
		filerecord.BugComment{
			Created: time.Now(),
			User:    s8webclient.CONN.Username,
			Content: "State changed to Fixed",
		},
	)
	_ = s8webclient.CONN.CreateUpdateBug(*this.bug)
	this.Refresh()
}

func (this *PlusBugList) Open() {
	this.bug.State = filerecord.BS_OPEN
	this.bug.Assigned = s8webclient.CONN.Username
	this.bug.Comments = append(
		this.bug.Comments,
		filerecord.BugComment{
			Created: time.Now(),
			User:    s8webclient.CONN.Username,
			Content: "State changed to Open",
		},
	)
	_ = s8webclient.CONN.CreateUpdateBug(*this.bug)
	this.Refresh()
}
