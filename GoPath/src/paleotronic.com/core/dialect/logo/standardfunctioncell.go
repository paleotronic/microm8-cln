package logo

import (
	//	"math"

	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionCELL struct {
	dialect.CoreFunction
}

func (this *StandardFunctionCELL) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.TABLE)
	result = append(result, types.LIST)

	/* enforce non void return */
	return result

}

func NewStandardFunctionCELL(a int, b int, params types.TokenList) *StandardFunctionCELL {
	this := &StandardFunctionCELL{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ITEM"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionCELL) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	thing := this.Stack.Shift()
	index := this.Stack.Shift()

	if thing.Type != types.TABLE {
		return errors.New("first param should be table")
	}

	if index.Type != types.LIST || index.List.Size() != 2 {
		return errors.New("second param should be list length two")
	}

	r, c := index.List.Get(0).AsInteger(), index.List.Get(1).AsInteger()

	value, err := TableGetCell(thing, r, c)
	if err != nil {
		return err
	}

	this.Stack.Push(value.Copy())

	//this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a+b)))

	return nil
}

func (this *StandardFunctionCELL) Syntax() string {

	/* vars */
	var result string

	result = "ITEM index word|list"

	/* enforce non void return */
	return result

}
