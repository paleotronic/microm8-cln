package plus

import (
	"paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSetSKV struct {
	dialect.CoreFunction
}

func (this *PlusSetSKV) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	k := this.Stack.Shift().Content
	v := this.Stack.Shift().Content

	_ = s8webclient.CONN.SetKeyValue("SKS", k, v)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusSetSKV) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusSetSKV) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusSetSKV(a int, b int, params types.TokenList) *PlusSetSKV {
	this := &PlusSetSKV{}

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
