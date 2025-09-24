package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
)

type StandardCommandEXIT struct {
	dialect.Command
}

func NewStandardCommandEXIT() *StandardCommandEXIT {
	this := &StandardCommandEXIT{}
	this.ImmediateMode = true
	return this
}

func (this *StandardCommandEXIT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	//apple2helpers.PutStr(caller,"Trying to exit " + caller.GetName())

	if caller.GetParent() != nil {
		apple2helpers.TextRestoreScreen(caller.GetParent())
		caller.GetParent().GetDialect().InitVDU(caller.GetParent(), true)
		caller.GetParent().SetChild(nil)
		caller.SetParent(nil)
		caller.GetDialect().GetCommands()["text"].Execute(env, caller, tokens, Scope, LPC)
	} else {
		settings.VMLaunch[caller.GetMemIndex()] = nil
		settings.PureBootVolume[caller.GetMemIndex()] = ""
		settings.PureBootVolume2[caller.GetMemIndex()] = ""
		settings.PureBootSmartVolume[caller.GetMemIndex()] = ""
		settings.MicroPakPath = ""
		caller.GetMemoryMap().IntSetSlotRestart(caller.GetMemIndex(), true)
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandEXIT) Syntax() string {

	/* vars */
	var result string

	result = "END"

	/* enforce non void return */
	return result

}
