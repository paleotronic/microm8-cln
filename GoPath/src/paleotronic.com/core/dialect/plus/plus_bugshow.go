package plus

import (
	"paleotronic.com/fmt"
	"strings"
	"time"

	"paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/filerecord"
	"paleotronic.com/utils"
)

type PlusBugShow struct {
	dialect.CoreFunction
	bug  *filerecord.BugReport
	edit *editor.CoreEdit
}

func (this *PlusBugShow) FunctionExecute(params *types.TokenList) error {

	this.CoreFunction.FunctionExecute(params)

	if !this.Query {

		ct := this.ValueMap["id"]
		id := int64(ct.AsInteger())

		if id < 1 {
			this.Interpreter.PutStr("Please specify an id.")
			this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
			return nil
		}

		this.bug, _ = s8webclient.CONN.GetBugByID(id)

		if this.bug.DefectID != id {
			this.Interpreter.PutStr("Could not find bug with that id.")
			this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
			return nil
		}

		apple2helpers.MonitorPanel(this.Interpreter, true)

		apple2helpers.TEXTMAX(this.Interpreter)

		this.edit = editor.NewCoreEdit(this.Interpreter, "View Bug #"+utils.IntToStr(int(this.bug.DefectID)), "", false, true)
		this.Refresh()
		this.edit.BarBG = 1
		this.edit.BarFG = 15
		this.edit.BGColor = 8
		this.edit.FGColor = 15
		this.edit.SetEventHandler(this)
		this.edit.Title = "View Bug #" + utils.IntToStr(int(this.bug.DefectID)) + " - Esc + E(x)it, (c)omment, c(l)ose, (r)eassign"

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_Q,
			"Quit",
			true,
			BuglistQuit,
		)

		this.edit.Run()

		//apple2helpers.TextRestoreScreen(this.Interpreter)
		apple2helpers.MonitorPanel(this.Interpreter, false)

	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusBugShow) OnMouseMove(edit *editor.CoreEdit, x, y int) {
	fmt.Printf("Mouse at (%d, %d)\n", x, y)

	if y > 0 && y < 46 {
		tl := edit.Voffset + int(y-1)
		if tl < len(edit.Content) {
			edit.GotoLine(tl)
			edit.Display()
		}
	}
}

func (this *PlusBugShow) OnMouseClick(edit *editor.CoreEdit, left, right bool) {
}

func (this *PlusBugShow) Refresh() {
	out := ""

	out += "ID        : " + utils.IntToStr(int(this.bug.DefectID)) + "\r\n"
	out += "Summary   : " + this.bug.Summary + "\r\n"
	out += "State     : " + this.bug.State.String() + "\r\n"
	out += "Assigned  : " + this.bug.Assigned + "\r\n"
	out += "Created by: " + this.bug.Creator + "\r\n"
	out += "Program   : " + this.bug.Filename + "\r\n"
	out += "Path      : " + this.bug.Filepath + "\r\n"
	out += "Date      : " + this.bug.Created.String() + "\r\n"
	out += "Body      : " + strings.Replace(this.bug.Body, "\n", "\r\n", -1) + "\r\n"

	if this.bug.Body == "" {
		out += "(none)\r\n"
	}

	for i, c := range this.bug.Comments {
		out += "\r\n"
		out += "Comment #" + utils.IntToStr(i+1) + " by " + c.User + " at " + c.Created.String() + "\r\n"
		out += "   " + c.Content
		out += "\r\n"
	}

	this.edit.SetText(out)
}

func (this *PlusBugShow) Syntax() string {

	/* vars */
	var result string

	result = "EXIT{}"

	/* enforce non void return */
	return result

}

func (this *PlusBugShow) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusBugShow(a int, b int, params types.TokenList) *PlusBugShow {
	this := &PlusBugShow{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "bug.show"
	this.Raw = true
	this.NamedParams = []string{"id"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
	}

	return this
}

func (this *PlusBugShow) OnEditorBegin(edit *editor.CoreEdit) {
}

func (this *PlusBugShow) OnEditorChange(edit *editor.CoreEdit) {
}

func (this *PlusBugShow) OnEditorExit(edit *editor.CoreEdit) {
}

func (this *PlusBugShow) OnEditorMove(edit *editor.CoreEdit) {

}

func (this *PlusBugShow) OnEditorKeypress(edit *editor.CoreEdit, ch rune) bool {

	return false
}

func (this *PlusBugShow) ReAssign(target string) {
	this.bug.Assigned = target
	_ = s8webclient.CONN.CreateUpdateBug(*this.bug)
	this.Refresh()
}

func (this *PlusBugShow) Close(comment string) {
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

func (this *PlusBugShow) Comment(comment string) {
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

func (this *PlusBugShow) Working() {
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

func (this *PlusBugShow) Test() {
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

func (this *PlusBugShow) Fix() {
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

func (this *PlusBugShow) Open() {
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
