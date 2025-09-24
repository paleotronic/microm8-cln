package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types" //	"paleotronic.com/core/memory"
	"paleotronic.com/utils"      //	"time"
	//	"paleotronic.com/core/settings"
)

type PlusMousePos struct {
	dialect.CoreFunction
}

func (this *PlusMousePos) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()

	if this.Query {

		var x, y = mm.IntGetMousePos(index)
		var v float64

		switch this.QueryVar {
		case "x":
			v = float64(x)
		case "y":
			v = float64(y)
		}

		this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))

	}

	return nil
}

func (this *PlusMousePos) Syntax() string {

	/* vars */
	var result string

	result = "ANG{ax,ay,az}"

	/* enforce non void return */
	return result

}

func (this *PlusMousePos) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusMousePos(a int, b int, params types.TokenList) *PlusMousePos {
	this := &PlusMousePos{}

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
