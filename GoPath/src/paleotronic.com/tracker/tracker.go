package tracker

import (
	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/ducktape/client"
	"paleotronic.com/filerecord"
)

type Tracker struct {
	Connection *client.DuckTapeClient
	Items      []*filerecord.BugReport
	edit       *editor.CoreEdit
}

func NewTracker(c *client.DuckTapeClient) *Tracker {
	return &Tracker{Connection: c}
}

func (t *Tracker) Run() {

	t.Items, _ = t.GetAllBugs()

}

func (this *Tracker) OnEditorKeypress(edit *editor.CoreEdit, ch rune) bool {
	return false
}

func (this *Tracker) OnEditorExit(edit *editor.CoreEdit) {
	///this.edit.VDU.ClrHome()
	apple2helpers.Clearscreen(this.edit.Int)
}

func (this *Tracker) OnEditorBegin(edit *editor.CoreEdit) {
	//this.edit.Int.PutStr(string(rune(7)))
}

func (this *Tracker) OnEditorChange(edit *editor.CoreEdit) {
}

func (this *Tracker) GetAllBugs() ([]*filerecord.BugReport, error) {

	data, err := []*filerecord.BugReport{
		&filerecord.BugReport{DefectID: 22, Summary: "Foo crashes", Body: "Stuff is broken", Filename: "foo.a", Assigned: "frog21", Creator: "april"},
	}, error(nil)

	return data, err

}

func (this *Tracker) InputText(title string, current string) string {

}
