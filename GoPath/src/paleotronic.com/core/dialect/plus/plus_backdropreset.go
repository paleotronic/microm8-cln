package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusBackdropReset struct {
	dialect.CoreFunction
}

func (this *PlusBackdropReset) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()

	//var x types.Token
	this.Interpreter.GetMemoryMap().IntSetBackdrop(index, "", 7, 1, 16, 0, false)
	this.Interpreter.GetMemoryMap().IntSetBackdropPos(index, types.CWIDTH/2, types.CHEIGHT/2, -types.CWIDTH/2)

	return nil
}

func (this *PlusBackdropReset) Syntax() string {

	/* vars */
	var result string

	result = "BACKDROP{image}"

	/* enforce non void return */
	return result

}

func (this *PlusBackdropReset) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusBackdropReset(a int, b int, params types.TokenList) *PlusBackdropReset {
	this := &PlusBackdropReset{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BACKDROP"
	this.MinParams = 0
	this.MaxParams = 1

	return this
}
