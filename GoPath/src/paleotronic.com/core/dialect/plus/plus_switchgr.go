package plus

import (
	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	//	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSwitchHGR struct {
	dialect.CoreFunction
}

func (this *PlusSwitchHGR) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	//	if !this.Query {
	//		q := this.ValueMap["enabled"]
	//		this.Interpreter.GetMemoryMap().WriteGlobal(this.Interpreter.GetMemoryMap().MEMBASE(this.Interpreter.GetMemIndex())+memory.OCTALYZER_INTERPRETER_PROFILE, uint64(q.AsInteger()))
	//	}
	fmt.Println("Deprecated function")

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusSwitchHGR) Syntax() string {

	/* vars */
	var result string

	result = "SETCLASSICHGR{}"

	/* enforce non void return */
	return result

}

func (this *PlusSwitchHGR) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusSwitchHGR(a int, b int, params types.TokenList) *PlusSwitchHGR {
	this := &PlusSwitchHGR{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SETCLASSICHGR"

	this.NamedDefaults = []types.Token{*types.NewToken(types.NUMBER, "0")}
	this.NamedParams = []string{"enabled"}
	this.Raw = true

	return this
}
