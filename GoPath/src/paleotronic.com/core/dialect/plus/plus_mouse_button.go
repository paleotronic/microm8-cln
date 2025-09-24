package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types" //	"paleotronic.com/core/memory"
	"paleotronic.com/utils"      //	"time"
	//	"paleotronic.com/core/settings"
)

type PlusMouseButton struct {
	dialect.CoreFunction
}

func (this *PlusMouseButton) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()

	if this.Query {

		var l, _ = mm.IntGetMouseButtons(index)
		var v float64

		if l {
			v = 1
		}

		this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))

	}

	return nil
}

func (this *PlusMouseButton) Syntax() string {

	/* vars */
	var result string

	result = "ANG{ax,ay,az}"

	/* enforce non void return */
	return result

}

func (this *PlusMouseButton) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusMouseButton(a int, b int, params types.TokenList) *PlusMouseButton {
	this := &PlusMouseButton{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMANG"
	this.MinParams = 1
	this.MaxParams = 3
	this.Raw = true
	this.NamedParams = []string{"x", "y"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.INTEGER, "0"),
		*types.NewToken(types.INTEGER, "0"),
	}
	this.EvalSingleParam = true

	return this
}
