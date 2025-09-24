package applesoft

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type StandardCommandXLIST struct {
	dialect.Command
	openedit bool
	execute  int
	Scope    *types.Algorithm
	edit     *editor.CoreEdit
	selected int
}

func NewStandardCommandXLIST() *StandardCommandXLIST {
	this := &StandardCommandXLIST{}
	this.ImmediateMode = true
	return this
}

func (this *StandardCommandXLIST) OnEditorKeypress(edit *editor.CoreEdit, ch rune) bool {
	//System.Err.Println("Got editor keypress: "+ch);

	if ch == 27 {
		edit.Done()
		return true
	}

	if ch == vduconst.F2 {
		// get lines
		this.openedit = true
		edit.Done()
		return true
	}

	return false
}

func (this *StandardCommandXLIST) TokenListAsString(tokens types.TokenList) string {

	/* vars */
	var result string
	//Token tok;

	result = ""
	for _, tok := range tokens.Content {
		if result != "" {
			result = result + " "
		}

		switch tok.Type {
		case types.KEYWORD:
			{
				if strings.ToLower(tok.Content) == "rem" {
					result = result + col(13) + strings.ToUpper(tok.AsString())
				} else {
					result = result + col(15) + strings.ToUpper(tok.AsString()) + col(5)
				}
				break
			}
		case types.FUNCTION:
			{
				result = result + col(11) + strings.ToUpper(tok.AsString()) + col(5)
				break
			}
		case types.DYNAMICKEYWORD:
			{
				result = result + col(15) + strings.ToUpper(tok.AsString()) + col(5)
				break
			}
		case types.ASSIGNMENT:
			{
				result = result + col(15) + strings.ToUpper(tok.AsString()) + col(5)
				break
			}
		case types.OPERATOR:
			{
				result = result + col(15) + strings.ToUpper(tok.AsString()) + col(5)
				break
			}
		case types.COMPARITOR:
			{
				result = result + col(15) + strings.ToUpper(tok.AsString()) + col(5)
				break
			}
		case types.LOGIC:
			{
				result = result + col(15) + strings.ToUpper(tok.AsString()) + col(5)
				break
			}
		case types.STRING:
			{
				result = result + col(14) + tok.AsString() + col(5)
				break
			}
		case types.VARIABLE:
			{
				result = result + col(12) + strings.ToUpper(tok.AsString()) + col(5)
				break
			}
		case types.NUMBER:
			{
				result = result + col(6) + strings.ToUpper(tok.AsString()) + col(5)
				break
			}
		case types.INTEGER:
			{
				result = result + col(6) + strings.ToUpper(tok.AsString()) + col(5)
				break
			}
		default:
			{
				result = result + tok.AsString()
				break
			}
		}

	}

	/* enforce non void return */
	return result

}

func col(n int) string {
	if settings.HighContrastUI {
		return string(rune(6)) + string(rune(0))
	}
	return string(rune(6)) + string(rune(n))
}

func (this *StandardCommandXLIST) Syntax() string {
	// TODO Auto-generated method stub
	return "XLIST"
}

func (this *StandardCommandXLIST) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	settings.DisableMetaMode[caller.GetMemIndex()] = true

	tt := caller.GetCode()

	l := tt.GetLowIndex()
	h := tt.GetHighIndex()

	indent := len(utils.IntToStr(h))

	// Purge keys
	caller.SetBuffer(runestring.NewRuneString())

	text := ""
	s := ""
	ft := ""

	var ln types.Line

	for l != -1 {
		/* display this line */
		//Str(l, s)
		//write(f, PadLeft(s, w)+' ');
		/* now formatted tokens */
		//int indent = Integer.ToString(l).Length();

		ln, _ = tt.Get(l)
		s = ""
		z := 0
		for _, stmt := range ln {
			label := utils.IntToStr(l)
			if z > 0 {
				label = ":"
				s = s + "\r\n"
			}
			for len(label) < indent {
				label = " " + label
			}

			s = s + col(5) + label + col(5) + "  "
			stl := *stmt.SubList(0, stmt.Size())
			ft = this.TokenListAsString(stl)
			//ft = strings.Replace( ft, string([]rune{34,4,34}), "CHR$(4)", -1 )
			s = s + ft

			z++
		}
		s = s + "\r\n \r\n"
		text = text + s

		/* next line */
		l = tt.NextAfter(l)
	}

	//	caller.GetVDU().SaveVDUState()
	//	caller.GetVDU().SetVideoMode(caller.GetVDU().GetVideoModes()[0])

	//apple2helpers.TextSaveScreen(caller)
	apple2helpers.MonitorPanel(caller, true)
	apple2helpers.TEXTMAX(caller)
	caller.SetIgnoreSpecial(true)

	this.edit = editor.NewCoreEdit(caller, "VIEW", text, false, false)

	this.edit.BarFG = 15
	this.edit.BarBG = 6
	this.edit.FGColor = 15
	this.edit.BGColor = 2
	this.edit.Title = "XLIST"

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_Q,
		"Quit",
		true,
		EditorExit,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_E,
		"Edit",
		true,
		EditorEdit,
	)

	this.edit.CursorScrollWindow = true
	this.edit.SetEventHandler(this)
	this.openedit = false
	this.edit.Run()
	//apple2helpers.TextRestoreScreen(caller)
	apple2helpers.MonitorPanel(caller, false)
	caller.SetIgnoreSpecial(false)

	//	caller.GetVDU().RestoreVDUState()

	if this.openedit {
		ll := this.edit.CountLinesToTop("^[ ]*[0-9]+[ ]+.+$")
		b := runestring.NewRuneString()
		b.Append("edit " + utils.IntToStr(ll) + "\r")
		caller.SetBuffer(b)
	}

	settings.DisableMetaMode[caller.GetMemIndex()] = false

	return 0, nil
}

func EditorEdit(this *editor.CoreEdit) {

	l := this.Line + 1

	this.Running = false

	tl := types.NewTokenList()
	tl.Push(types.NewToken(types.KEYWORD, "edit"))
	tl.Push(types.NewToken(types.NUMBER, utils.IntToStr(l)))

	code := this.Int.GetCode()

	this.Int.GetDialect().ExecuteDirectCommand(*tl, this.Int, code, this.Int.GetLPC())

}

func (this *StandardCommandXLIST) OnEditorExit(edit *editor.CoreEdit) {
	///this.edit.VDU.ClrHome()
	apple2helpers.Clearscreen(this.edit.Int)
}

func (this *StandardCommandXLIST) OnEditorBegin(edit *editor.CoreEdit) {
	//this.edit.Int.PutStr(string(rune(7)))
}

func (this *StandardCommandXLIST) OnEditorChange(edit *editor.CoreEdit) {
}

func (this *StandardCommandXLIST) OnEditorMove(edit *editor.CoreEdit) {

}

func (this *StandardCommandXLIST) OnMouseMove(edit *editor.CoreEdit, x, y int) {
}

func (this *StandardCommandXLIST) OnMouseClick(edit *editor.CoreEdit, left, right bool) {
}
