package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
)

type PlusBlockPC struct {
	dialect.CoreFunction
}

func (this *PlusBlockPC) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	i := params.Shift().AsInteger()
	settings.AllowPerspectiveChanges = (i == 0)

	return nil
}

func (this *PlusBlockPC) Syntax() string {

	/* vars */
	var result string

	result = "CGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusBlockPC) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusBlockPC(a int, b int, params types.TokenList) *PlusBlockPC {
	this := &PlusBlockPC{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CGCOLOR"

	//this.NamedParams = []string{ "color" }
	//this.NamedDefaults = []types.Token{ *types.NewToken( types.INTEGER, "15" ) }
	//this.Raw = true

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "enabled", Default: *types.NewToken(types.NUMBER, "0")},
		},
	)

	return this
}
