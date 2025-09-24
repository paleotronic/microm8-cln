package applesoft

import (
	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardFunctionTAB struct {
	dialect.CoreFunction
}

func (this *StandardFunctionTAB) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value int

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	value = this.Stack.Pop().AsInteger()

	// here we need to account for buffered output first
	//cx := this.Interpreter.GetCursorX()
	//cy := this.Interpreter.GetCursorY()

	if ((this.Interpreter.GetCursorX() + 1) < value) && (value <= this.Interpreter.GetColumns()) {

		for this.Interpreter.GetCursorX() < value-1 {
			this.Interpreter.PutStr(" ")
		}
	}

	this.Stack.Push(types.NewToken(types.STRING, ""))

	return nil
}

func (this *StandardFunctionTAB) Syntax() string {

	/* vars */
	var result string

	result = "TAB(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionTAB) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionTAB(a int, b int, params types.TokenList) *StandardFunctionTAB {
	this := &StandardFunctionTAB{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "TAB"

	return this
}
