package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//"paleotronic.com/utils"

	"time"
)

type StandardFunctionREADCHAR struct {
	dialect.CoreFunction
}

func (this *StandardFunctionREADCHAR) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionREADCHAR(a int, b int, params types.TokenList) *StandardFunctionREADCHAR {
	this := &StandardFunctionREADCHAR{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "READCHAR"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionREADCHAR) FunctionExecute(params *types.TokenList) error {

	/* vars */
	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	for this.Interpreter.GetMemory( 49152 ) < 128 {
		time.Sleep(5 * time.Millisecond)
	}

	// now got char
	

	this.Stack.Push(types.NewToken(types.WORD, string(rune(this.Interpreter.GetMemory( 49152 ) & 0xff7f))))

	this.Interpreter.SetMemory(49168,0)

	return nil
}

func (this *StandardFunctionREADCHAR) Syntax() string {

	/* vars */
	var result string

	result = "READCHAR a b"

	/* enforce non void return */
	return result

}
