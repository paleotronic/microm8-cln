package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/utils"
)

type PlusRotPal struct {
	dialect.CoreFunction
}

func (this *PlusRotPal) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }
	
	low := params.Shift().AsInteger()
	high := params.Shift().AsInteger()
	change := params.Shift().AsInteger()

	apple2helpers.RotatePalette(this.Interpreter, low, high, change)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))


	return nil
}

func (this *PlusRotPal) Syntax() string {

	/* vars */
	var result string

	result = "CGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusRotPal) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusRotPal(a int, b int, params types.TokenList) *PlusRotPal {
	this := &PlusRotPal{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CGCOLOR"
	this.MinParams = 3
	this.MaxParams = 3

	//this.NamedParams = []string{ "color" }
	//this.NamedDefaults = []types.Token{ *types.NewToken( types.INTEGER, "15" ) }
	//this.Raw = true

	return this
}
