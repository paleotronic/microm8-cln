package plus

import (
	"paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusGetSKV struct {
	dialect.CoreFunction
}

func (this *PlusGetSKV) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	k := this.Stack.Shift().Content

	v, _ := s8webclient.CONN.GetKeyValue("GKS", k)

	this.Stack.Push(types.NewToken(types.STRING, v))

	return nil
}

func (this *PlusGetSKV) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusGetSKV) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusGetSKV(a int, b int, params types.TokenList) *PlusGetSKV {
	this := &PlusGetSKV{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 1
	this.MaxParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "key", Default: *types.NewToken(types.STRING, "")},
		},
	)

	return this
}
