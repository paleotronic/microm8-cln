package plus

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type PlusTriggerDef struct {
	dialect.CoreFunction
}

func (this *PlusTriggerDef) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slotid := this.Interpreter.GetMemIndex()
	condition := this.Stack.Shift().Content
	lineno := this.Stack.Shift().AsInteger()

	if this.Stack.Size() > 0 {
		slotid = this.Stack.Shift().AsInteger()
	}

	condition = strings.Replace(condition, "\"", "'", -1)
	tl := this.Interpreter.GetDialect().Tokenize(runestring.Cast(condition))

	this.Interpreter.AddTrigger(slotid, tl, lineno)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusTriggerDef) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusTriggerDef) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusTriggerDef(a int, b int, params types.TokenList) *PlusTriggerDef {
	this := &PlusTriggerDef{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 2
	this.MaxParams = 3

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "condition", Default: *types.NewToken(types.STRING, "")},
			dialect.FunctionParamDef{Name: "lineno", Default: *types.NewToken(types.NUMBER, "1")},
			dialect.FunctionParamDef{Name: "vm", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}
