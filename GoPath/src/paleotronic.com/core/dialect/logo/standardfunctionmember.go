package logo

import (
	"strings"
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"

	"fmt"
)

type StandardFunctionMEMBER struct {
	dialect.CoreFunction
}

func (this *StandardFunctionMEMBER) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionMEMBER(a int, b int, params types.TokenList) *StandardFunctionMEMBER {
	this := &StandardFunctionMEMBER{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "MEMBER"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionMEMBER) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	subthing := this.Stack.Shift()
	thing := this.Stack.Shift()

	if thing == nil || subthing == nil {
		this.Stack.Push(types.NewToken(types.STRING, ""))
		return nil
	}

	if thing.Type == types.LIST {
		index := thing.List.IndexOfToken(subthing)
		fmt.Printf("member.index=%d\n", index)
		if index < 0 {
			this.Stack.Push(types.NewToken(types.STRING, ""))
			return nil
		}
		this.Stack.Push(thing.SubListAsToken(index, thing.List.Size()))
	} else {
		//		n := len(thing.Content)
		index := strings.Index(thing.Content, subthing.Content)
		if index < 0 {
			this.Stack.Push(types.NewToken(types.STRING, ""))
			return nil
		}
		this.Stack.Push(types.NewToken(types.STRING, thing.Content[index:]))
	}

	//this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a+b)))

	return nil
}

func (this *StandardFunctionMEMBER) Syntax() string {

	/* vars */
	var result string

	result = "MEMBER index word|list"

	/* enforce non void return */
	return result

}
