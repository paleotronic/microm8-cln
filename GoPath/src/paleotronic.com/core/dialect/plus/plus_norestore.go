package plus

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusNoRestore struct {
	dialect.CoreFunction
}

func (this *PlusNoRestore) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	param := strings.ToUpper(params.Shift().Content)

	switch {
	case param == "1":
		this.Interpreter.SetSaveAndRestoreText(false)
	case param == "0":
		this.Interpreter.SetSaveAndRestoreText(true)
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusNoRestore) Syntax() string {

	/* vars */
	var result string

	result = "NORESTORE{TRUE|FALSE}"

	/* enforce non void return */
	return result

}

func (this *PlusNoRestore) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusNoRestore(a int, b int, params types.TokenList) *PlusNoRestore {
	this := &PlusNoRestore{}

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
