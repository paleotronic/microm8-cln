package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionSCRN struct {
	dialect.CoreFunction
}

func NewStandardFunctionSCRN(a int, b int, params types.TokenList) *StandardFunctionSCRN {
	this := &StandardFunctionSCRN{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SCRN"

	this.MinParams = 2
	this.MaxParams = 3

	return this
}

func (this *StandardFunctionSCRN) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var x int
	var y int
	var i int
	//var r int
	//var s string
	//var ch rune

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	y = this.Stack.Pop().AsInteger()
	x = this.Stack.Pop().AsInteger()
	z := 0
	if this.Stack.Size() > 0 {
		z = this.Stack.Pop().AsInteger()
	}

	//if (this.Interpreter.VDU.VideoMode.ActualRows == this.Interpreter.VDU.VideoMode.Rows)
	//{
	//  this.Stack.Push( types.NewToken(types.NUMBER, utils.IntToStr( this.Interpreter.VDU[x,y / 2] )) );
	//  return;
	//}

	//i = this.GetEntity().GetVDU().ColorAt(x, y)

	modes := apple2helpers.GetActiveVideoModes(this.Interpreter)

	if len(modes) == 1 {
		switch modes[0] {
		case "LOGR":
			i = int(apple2helpers.LOGRGet40(this.Interpreter, x, y))
		case "DLGR":
			i = int(apple2helpers.LOGRGet80(this.Interpreter, x, y))
		case "CUBE":
			i = int(apple2helpers.CUBE(this.Interpreter).ColorAt(uint8(x), uint8(y), uint8(z)))
		}
	}

	// System.Out.Println("SCRN <================================================> Pixel at x = "+x+", y = "+y+" is "+i);

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(i)))

	return nil
}

func (this *StandardFunctionSCRN) Syntax() string {

	/* vars */
	var result string

	result = "SCRN(X, Y, Z)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionSCRN) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 2 );
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
