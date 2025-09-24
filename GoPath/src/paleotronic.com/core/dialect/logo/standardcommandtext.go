package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandTEXT struct {
	dialect.Command
}

func (this *StandardCommandTEXT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	//cx := caller.GetVDU().GetCursorX()
	cy := apple2helpers.GetCursorY(caller)

	if apple2helpers.GetRows(caller) == 0 {
		cy = 23
	}

	//caller.GetVDU().Clear;
	//caller.GetVDU().CursorY = caller.GetVDU().Window.Bottom;
	caller.SetMemory(65536-16303, 0)

	//	w := apple2helpers.GetColumns(caller)

	//	caller.SetMemory(32, 0)
	//	caller.SetMemory(33, uint(w))
	//	caller.SetMemory(34, 0)
	//	caller.SetMemory(35, 24)

	apple2helpers.SetRealWindow(caller, 0, 0, 79, 47)

	apple2helpers.PutStr(caller, "\r") // horiz home cursor
	cx := apple2helpers.GetCursorX(caller)
	apple2helpers.Gotoxy(caller, cx, cy)
	apple2helpers.Attribute(caller, types.VA_NORMAL)
	apple2helpers.CubeLayerEnable(caller, false, false)
	apple2helpers.VectorLayerEnable(caller, false, false)
	//caller.GetVDU().LastGraphicsMode = null;

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandTEXT) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
