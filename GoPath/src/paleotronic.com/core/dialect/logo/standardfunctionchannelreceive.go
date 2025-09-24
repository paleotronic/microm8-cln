package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type StandardFunctionCHANNELRECEIVE struct {
	dialect.CoreFunction
	D *DialectLogo
}

func NewStandardFunctionCHANNELRECEIVE(a int, b int, params types.TokenList, d *DialectLogo) *StandardFunctionCHANNELRECEIVE {
	this := &StandardFunctionCHANNELRECEIVE{D: d}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "RECEIVE"
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionCHANNELRECEIVE) FunctionExecute(params *types.TokenList) error {

	/* vars */

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	t := this.Stack.Shift()
	if t == nil || t.Type == types.LIST {
		return errors.New("FUNCTION EXPECTS 1 VALUE")
	}

	value, err := this.D.Driver.ChannelRecv(t.Content)
	if err != nil {
		return err
	}

	this.Stack.Push(value)

	return nil
}

func (this *StandardFunctionCHANNELRECEIVE) Syntax() string {

	/* vars */
	var result string

	result = "SIN(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionCHANNELRECEIVE) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
