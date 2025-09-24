package editor

import (
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/fmt"
)

type Viewer struct {
	edit *CoreEdit
	Int  interfaces.Interpretable
}

func NewViewer(ent interfaces.Interpretable, title string, text string) *Viewer {

	viewer := &Viewer{
		edit: NewCoreEdit(ent, title, text, false, false),
		Int:  ent,
	}
	viewer.edit.SetEventHandler(viewer)

	return viewer

}

func (this *Viewer) Do() {

	gfx, hud := apple2helpers.GetActiveLayers(this.Int)

	defer func() {
		apple2helpers.SetActiveLayers(this.Int, gfx, hud)
	}()

	this.edit.BarBG = 15
	this.edit.BarFG = 0
	this.edit.BGColor = 0
	this.edit.FGColor = 15

	this.edit.RegisterCommand(vduconst.SHIFT_CTRL_Q, "Quit", true, this.Quit)

	this.edit.Run()

}

func (this *Viewer) OnEditorMove(edit *CoreEdit) {

}

func (this *Viewer) OnEditorExit(edit *CoreEdit) {
	apple2helpers.Clearscreen(edit.Int)
}

func (this *Viewer) OnEditorBegin(edit *CoreEdit) {
	//edit.Int.PutStr(string(rune(7)))
}

func (this *Viewer) OnEditorChange(edit *CoreEdit) {

}

func (this *Viewer) OnMouseMove(edit *CoreEdit, x, y int) {
	fmt.Printf("Mouse at (%d, %d)\n", x, y)

	if y > 0 && y < 46 {
		tl := edit.Voffset + int(y-1)
		if tl < len(edit.Content) {
			edit.GotoLine(tl)
			edit.Display()
		}
	}

	edit.MouseMoved = true
}

func (this *Viewer) OnMouseClick(edit *CoreEdit, left, right bool) {

	if left && edit.MY > 0 && edit.MY < 46 {

		tl := edit.Voffset + int(edit.MY-1)
		if tl < len(edit.Content) {

			edit.GotoLine(tl)
			edit.Int.GetMemoryMap().KeyBufferAdd(edit.Int.GetMemIndex(), 13)

		}

	}

}

func (this *Viewer) OnEditorKeypress(edit *CoreEdit, ch rune) bool {
	return false
}

func (this *Viewer) Quit(edit *CoreEdit) {
	edit.Running = false
	edit.Changed = false
}
