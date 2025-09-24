package plus

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusNoBreak struct {
	dialect.CoreFunction
}

func (this *PlusNoBreak) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	param := strings.ToUpper(params.Shift().Content)

	switch {
	case param == "1":
		this.Interpreter.SetBreakable(false)
	case param == "0":
		this.Interpreter.SetBreakable(true)
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusNoBreak) Syntax() string {

	/* vars */
	var result string

	result = "NOBREAK{TRUE|FALSE}"

	/* enforce non void return */
	return result

}

func (this *PlusNoBreak) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusNoBreak(a int, b int, params types.TokenList) *PlusNoBreak {
	this := &PlusNoBreak{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "NOBREAK"

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "enabled", Default: *types.NewToken(types.NUMBER, "0")},
		},
	)

	return this
}
