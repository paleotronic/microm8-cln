package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type StandardFunctionLOGOID struct {
	dialect.CoreFunction
	D *DialectLogo
}

func NewStandardFunctionLOGOID(a int, b int, params types.TokenList, d *DialectLogo) *StandardFunctionLOGOID {
	this := &StandardFunctionLOGOID{D: d}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ROUTINE"
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionLOGOID) FunctionExecute(params *types.TokenList) error {

	/* vars */

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	t := this.Stack.Shift()
	if t == nil || t.Type != types.LIST {
		return errors.New("FUNCTION EXPECTS A LIST")
	}

	c, err := this.D.Driver.SpawnCoroutine(t.List.Copy())
	if err != nil {
		return err
	}

	this.Stack.Push(types.NewToken(types.NUMBER, c))

	return nil
}

func (this *StandardFunctionLOGOID) Syntax() string {

	/* vars */
	var result string

	result = "SIN(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionLOGOID) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
