package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionREPCOUNT struct {
	dialect.CoreFunction
	D *DialectLogo
}

func (this *StandardFunctionREPCOUNT) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionREPCOUNT(a int, b int, params types.TokenList, d *DialectLogo) *StandardFunctionREPCOUNT {
	this := &StandardFunctionREPCOUNT{D: d}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "COS"
	this.MaxParams = 0
	this.MinParams = 0

	return this
}

func (this *StandardFunctionREPCOUNT) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	//log.Printf("Stack size is: %d", this.D.Driver.Stack.Size())

	c := this.D.Driver.Stack.Top().IterCount()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(c+1)))

	return nil
}

func (this *StandardFunctionREPCOUNT) Syntax() string {

	/* vars */
	var result string

	result = "COS(<number>)"

	/* enforce non void return */
	return result

}
