package logo

import (
	"strings"
	//	"math"
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionGPROP struct {
	dialect.CoreFunction
}

func (this *StandardFunctionGPROP) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionGPROP(a int, b int, params types.TokenList) *StandardFunctionGPROP {
	this := &StandardFunctionGPROP{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "GPROP"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionGPROP) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	// var name
	if params.Get(0).Type == types.LIST {
		params = params.Get(0).List.SubList(0, params.Get(0).List.Size())
	}

	o := params.Shift()
	if o == nil {
		return errors.New("I NEED 2 OBJECTS")
	}
	name := strings.ToLower(o.Content)
	if !strings.HasPrefix(name, ":") {
		name = ":" + name
	}

	//
	f := params.Shift()
	if f == nil {
		return errors.New("I NEED 2 OBJECTS")
	}
	field := strings.ToLower(f.Content)

	lobj := this.Interpreter.GetData(name)

	if lobj == nil {
		// empty list
		el := types.NewToken(types.LIST, "")
		el.List = types.NewTokenList()
		this.Stack.Push(el)
	} else {
		// find property
		foundIdx := -1
		for i := 0; i < lobj.List.Size(); i += 2 {
			t := lobj.List.Get(i)
			if strings.ToLower(t.Content) == strings.ToLower(field) {
				foundIdx = i
				break
			}
		}
		if foundIdx == -1 {
			// empty list
			el := types.NewToken(types.LIST, "")
			el.List = types.NewTokenList()
			this.Stack.Push(el)
		} else {
			this.Stack.Push(lobj.List.Content[foundIdx+1])
		}
	}

	return nil
}

func (this *StandardFunctionGPROP) Syntax() string {

	/* vars */
	var result string

	result = "GPROP object"

	/* enforce non void return */
	return result

}
