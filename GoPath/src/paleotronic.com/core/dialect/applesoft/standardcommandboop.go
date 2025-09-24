package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandBOOP struct {
	dialect.Command
}

func NewStandardCommandBOOP() *StandardCommandBOOP {
	this := &StandardCommandBOOP{}
	this.UseStates = true
	return this
}

func (this *StandardCommandBOOP) StateInit(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	caller.PutStr("init\r\n")
	caller.SetSubState( types.ESS_EXEC )
	
	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandBOOP) StateExec(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	caller.PutStr("exec\r\n")
	caller.SetSubState( types.ESS_DONE )
	
	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandBOOP) StateDone(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	caller.PutStr("done\r\n")
	panic("testing exceptions")
	
	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandBOOP) Syntax() string {

	/* vars */
	var result string

	result = "BOOP"

	/* enforce non void return */
	return result

}
