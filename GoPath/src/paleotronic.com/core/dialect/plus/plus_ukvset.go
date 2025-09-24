package plus

import (
	"paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSetUKV struct {
	dialect.CoreFunction
}

func (this *PlusSetUKV) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	k := this.Stack.Shift().Content
	v := this.Stack.Shift().Content

	_ = s8webclient.CONN.SetKeyValue("SKU", k, v)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusSetUKV) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusSetUKV) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusSetUKV(a int, b int, params types.TokenList) *PlusSetUKV {
	this := &PlusSetUKV{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 2
	this.MaxParams = 2

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "key", Default: *types.NewToken(types.STRING, "")},
			dialect.FunctionParamDef{Name: "value", Default: *types.NewToken(types.STRING, "")},
		},
	)

	return this
}
