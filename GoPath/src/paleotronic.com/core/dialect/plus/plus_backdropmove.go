package plus

import (
	"paleotronic.com/utils"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusBackdropMove struct {
	dialect.CoreFunction
}

func (this *PlusBackdropMove) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()

	var tx, ty, tz types.Token
	if !this.Query {

		ox, oy, oz := this.Interpreter.GetMemoryMap().IntGetBackdropPos(index)

		tx = this.ValueMap["x"]
		ty = this.ValueMap["y"]
		tz = this.ValueMap["z"]
		x, y, z := tx.AsExtended(), ty.AsExtended(), tz.AsExtended()

		this.Interpreter.GetMemoryMap().IntSetBackdropPos(index, ox+x, oy+y, oz+z)
		this.Stack.Push(types.NewToken(types.NUMBER, "0"))
	} else {
		x, y, z := this.Interpreter.GetMemoryMap().IntGetBackdropPos(index)
		var v float64

		switch this.QueryVar {
		case "x":
			v = x
		case "y":
			v = y
		case "z":
			v = z
		}

		this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))
	}

	return nil
}

func (this *PlusBackdropMove) Syntax() string {

	/* vars */
	var result string

	result = "BACKDROP{image}"

	/* enforce non void return */
	return result

}

func (this *PlusBackdropMove) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusBackdropMove(a int, b int, params types.TokenList) *PlusBackdropMove {
	this := &PlusBackdropMove{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BACKDROP"

	this.NamedParams = []string{"x", "y", "z"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, utils.FloatToStr(types.CWIDTH/2)),
		*types.NewToken(types.NUMBER, utils.FloatToStr(types.CHEIGHT/2)),
		*types.NewToken(types.NUMBER, utils.FloatToStr(-types.CWIDTH/2)),
	}
	this.Raw = true
	this.MinParams = 1
	this.MaxParams = 3

	return this
}
