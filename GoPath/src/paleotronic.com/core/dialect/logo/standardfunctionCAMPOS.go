package logo

import (
	//	"strings"
	//	"time"

	"paleotronic.com/core/dialect" //	"paleotronic.com/runestring"
	"paleotronic.com/core/types"   //	"paleotronic.com/core/interfaces"
	"paleotronic.com/utils"
)

type StandardFunctionCAMPOS struct {
	dialect.CoreFunction
}

func (this *StandardFunctionCAMPOS) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionCAMPOS(a int, b int, params types.TokenList) *StandardFunctionCAMPOS {
	this := &StandardFunctionCAMPOS{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "POS3"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionCAMPOS) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()
	cindex := 1
	control := types.NewOrbitController(mm, index, cindex)
	a := control.GetPosition()

	l := types.NewToken(types.LIST, "")
	l.List = types.NewTokenList()
	l.List.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a[0])))
	l.List.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a[1])))
	l.List.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a[2])))

	this.Stack.Push(l)

	return nil
}

func (this *StandardFunctionCAMPOS) Syntax() string {

	/* vars */
	var result string

	result = "POS3 word list"

	/* enforce non void return */
	return result

}
