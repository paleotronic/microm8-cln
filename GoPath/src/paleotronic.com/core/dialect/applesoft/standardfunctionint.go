package applesoft

import (
	//"paleotronic.com/log"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	"math"
    //"paleotronic.com/fmt"
)

type StandardFunctionINT struct {
	dialect.CoreFunction
}

func (this *StandardFunctionINT) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionINT(a int, b int, params types.TokenList) *StandardFunctionINT {
	this := &StandardFunctionINT{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "INT"

	return this
}

func (this *StandardFunctionINT) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value float64

	//log.Println("int() called with", params.Size(), "params...")

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	value = this.Stack.Pop().AsExtended()

    fvalue := math.Floor(value)
    if fvalue > value {
       fvalue -= 1
    }

	if ((value < 0) && (fvalue != value)) {
	  this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(fvalue-1))))
	} else {
	  this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(fvalue))))
	}

	return nil
}

func (this *StandardFunctionINT) Syntax() string {

	/* vars */
	var result string

	result = "INT(<number>)"

	/* enforce non void return */
	return result

}
