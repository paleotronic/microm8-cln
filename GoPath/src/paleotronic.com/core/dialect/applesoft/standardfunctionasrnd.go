package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionASRND struct {
	dialect.CoreFunction
	//vars instance
	lastval float64
}

func (this *StandardFunctionASRND) Lastval() float64 {

	/* vars */
	var result float64
	result = this.lastval

	/* enforce non void return */
	return result

}

func (this *StandardFunctionASRND) Setlastval(v float64) {

	/* vars */

	this.lastval = v

}

func NewStandardFunctionASRND(a int, b int, params types.TokenList) *StandardFunctionASRND {
	this := &StandardFunctionASRND{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "RND"
	this.lastval = 0

	return this
}

func (this *StandardFunctionASRND) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	value = this.Stack.Pop().AsExtended()

	if value > 0 {
		this.lastval = utils.Random() //math.Random()
		this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(this.lastval)))
	} else if value == 0 {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(this.lastval)))
	} else {
		utils.PSeed( int64(value) )
		this.lastval = utils.Random()
		this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(this.lastval)))
	}

	return nil
}

func (this *StandardFunctionASRND) Syntax() string {

	/* vars */
	var result string

	result = "RND(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionASRND) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
