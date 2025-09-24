package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
)

type PlusWrapper struct {
	dialect.CoreFunction
	wf interfaces.Functioner
}

func (this *PlusWrapper) FunctionExecute(params *types.TokenList) error {

	slotid := params.Shift().AsInteger() - 1
	this.wf.SetEntity(this.Interpreter.GetProducer().GetInterpreter(slotid % memory.OCTALYZER_NUM_INTERPRETERS))

	return this.wf.FunctionExecute(params)

}

func (this *PlusWrapper) Syntax() string {

	return this.wf.Syntax()

}

func (this *PlusWrapper) FunctionParams() []types.TokenType {

	/* vars */

	result := this.wf.FunctionParams()

	result = append([]types.TokenType{types.INTEGER}, result...)

	/* enforce non void return */
	return result

}

func NewPlusWrapper(a int, b int, params types.TokenList, wrapped interfaces.Functioner) *PlusWrapper {
	this := &PlusWrapper{}

	/* vars */
	this.wf = wrapped
	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = wrapped.GetName()
	this.NamedDefaults = wrapped.GetNamedDefaults()
	this.NamedParams = wrapped.GetNamedParams()
	this.Raw = wrapped.GetRaw()
	this.MinParams = wrapped.GetMinParams() + 1
	this.MaxParams = wrapped.GetMaxParams() + 1

	return this
}
