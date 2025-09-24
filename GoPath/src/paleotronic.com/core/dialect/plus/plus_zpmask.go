package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
//	"paleotronic.com/core/memory"
	"paleotronic.com/utils"
//	"time"
)

type PlusZPMask struct {
	dialect.CoreFunction
}

func (this *PlusZPMask) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

//	on := (this.Stack.Shift().AsInteger() != 0)

	//s :=  this.Stack.Shift().AsInteger()
	//e :=  this.Stack.Shift().AsInteger()

	//this.Interpreter.GetVDU().CamMove(float32(x),float32(y),float32(z))

//	mp, ex := this.Interpreter.GetMemoryMap().InterpreterMappableByLabel(this.Interpreter.GetMemIndex(), "Apple2IOZeroPage")

	//if !ex {
	//	return nil
	//}

	//if on {
	//	mp.SetMaskBit(s, e)
	//} else {
	//	mp.ClearMaskBit(s, e)
	//}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusZPMask) Syntax() string {

	/* vars */
	var result string

	result = "ZPMASK{state,start,end}"

	/* enforce non void return */
	return result

}

func (this *PlusZPMask) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusZPMask(a int, b int, params types.TokenList) *PlusZPMask {
	this := &PlusZPMask{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ZPMASK"

	return this
}
