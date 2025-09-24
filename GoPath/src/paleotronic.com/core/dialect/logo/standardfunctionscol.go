package logo

import (
	//	"math"

	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionSCOL struct {
	dialect.CoreFunction
}

func (this *StandardFunctionSCOL) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.TABLE)
	result = append(result, types.LIST)

	/* enforce non void return */
	return result

}

func NewStandardFunctionSCOL(a int, b int, params types.TokenList) *StandardFunctionSCOL {
	this := &StandardFunctionSCOL{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ITEM"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionSCOL) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	thing := this.Stack.Shift()
	args := this.Stack.Shift()

	if thing.Type != types.TABLE {
		return errors.New("first param should be table")
	}

	if args.Type != types.LIST || args.List.Size() < 2 {
		return errors.New("second param should be list")
	}

	row := args.List.Get(0).AsInteger()
	value := args.List.Get(1)
	variance := types.NewToken(types.NUMBER, "0")
	if args.List.Size() > 2 {
		variance = args.List.Get(2)
	}

	list, err := TableSearchColumn(thing, row, value, variance)
	if err != nil {
		return err
	}

	r := types.NewToken(types.LIST, "")
	r.List = list

	this.Stack.Push(r)

	return nil
}

func (this *StandardFunctionSCOL) Syntax() string {

	/* vars */
	var result string

	result = "ITEM index word|list"

	/* enforce non void return */
	return result

}
