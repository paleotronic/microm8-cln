package applesoft

import (
	"paleotronic.com/core/dialect" //"paleotronic.com/core/hires"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandHGR4 struct {
	dialect.Command
}

func (this *StandardCommandHGR4) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	var result int = 0

	//caller.PutStr(string(rune(4)) + "PR#3\r\n")
	//log.Printf("HGR4 called")

	apple2helpers.SuperHiresDisable(caller, false)
	apple2helpers.CubeLayerEnable(caller, false, false)
	apple2helpers.VectorLayerEnable(caller, false, false)
	apple2helpers.TEXT80(caller)
	apple2helpers.SetRealCursorPos(caller, 0, 46)
	if mr, ok := caller.GetMemoryMap().InterpreterMappableAtAddress(caller.GetMemIndex(), 0xc000); ok {
		io := mr.(*apple2.Apple2IOChip)
		io.SetVidModeForce(
			apple2.VF_TEXT,
		)
		io.SetVidModeForce(
			apple2.VF_HIRES |
				apple2.VF_DHIRES |
				//apple2.VF_MIXED |
				apple2.VF_80COL,
		)
	}

	//caller.GetVDU().Clear;
	//caller.GetVDU().Home();
	caller.SetCurrentPage("DHR1")
	caller.SetDisplayPage("DHR1")
	//hires.GetAppleHiRES().HgrFill(caller.GetVDU().GetBitmapMemory()[caller.GetVDU().GetCurrentPage()%2], 0)
	apple2helpers.HGRClear(caller)
	apple2helpers.HTab(caller, 1)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandHGR4) Syntax() string {

	/* vars */
	var result string

	result = "HGR2"

	/* enforce non void return */
	return result

}
