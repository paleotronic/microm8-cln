package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	"strconv"
)

type StandardFunctionVAL struct {
	dialect.CoreFunction
}

func (this *StandardFunctionVAL) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value string
	var e float64
	//var code int

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	value = this.Stack.Pop().Content

	// remove spaces

	//try {
	//	e = types.Double.ParseDouble(utils.NumberPart(value))
	//}
	//catch (Exception ex) {
	//	e = 0
	//}

    value = utils.NumberPart(value)

	e, err := strconv.ParseFloat(value, 32)
	if err != nil {
		e = 0
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", e)))

	return nil
}

func (this *StandardFunctionVAL) Syntax() string {

	/* vars */
	var result string

	result = "VAL(<string>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionVAL) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionVAL(a int, b int, params types.TokenList) *StandardFunctionVAL {
	this := &StandardFunctionVAL{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VAL"

	return this
}
