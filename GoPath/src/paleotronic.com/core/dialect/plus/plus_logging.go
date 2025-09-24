package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusLogging struct {
	dialect.CoreFunction
}

func (this *PlusLogging) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		q := this.ValueMap["enabled"]
		m := q.AsInteger()
		switch m {
		case 0:
			this.Interpreter.StopRecording()
			this.Interpreter.GetMemoryMap().IntSetLogMode(0, 0)
		case 1:
			this.Interpreter.StopRecording()
			this.Interpreter.GetMemoryMap().IntSetLogMode(0, 1)
		case 2:
			this.Interpreter.StopRecording()
			this.Interpreter.GetMemoryMap().IntSetLogMode(0, 2)
			this.Interpreter.RecordToggle(false)
		}
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusLogging) Syntax() string {

	/* vars */
	var result string

	result = "DEBUG{}"

	/* enforce non void return */
	return result

}

func (this *PlusLogging) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusLogging(a int, b int, params types.TokenList) *PlusLogging {
	this := &PlusLogging{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "LOGGING"

	this.NamedDefaults = []types.Token{*types.NewToken(types.NUMBER, "0")}
	this.NamedParams = []string{"enabled"}
	this.Raw = true

	return this
}
