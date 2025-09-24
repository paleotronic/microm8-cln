package plus

import (
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type PlusKeyType struct {
	dialect.CoreFunction
}

func (this *PlusKeyType) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	str := this.ValueMap["keys"].Content
	x := this.ValueMap["cps"]
	settings.PasteCPS = x.AsInteger()
	block := this.ValueMap["block"].Content != "0"

	// for _, ch := range str {
	// 	// for this.Interpreter.GetMemoryMap().KeyBufferSize(this.Interpreter.GetMemIndex()) > 0 {
	// 	// 	time.Sleep(1 * time.Millisecond)
	// 	// }
	// 	this.Interpreter.GetMemoryMap().KeyBufferAdd(this.Interpreter.GetMemIndex(), uint64(ch))
	// 	time.Sleep(delayMS * time.Millisecond)
	// }
	this.Interpreter.SetPasteBuffer(runestring.Cast(str))
	if block {
		for this.Interpreter.GetPasteBuffer().Length() > 0 {
			time.Sleep(25 * time.Millisecond)
		}
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusKeyType) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusKeyType) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusKeyType(a int, b int, params types.TokenList) *PlusKeyType {
	this := &PlusKeyType{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "keys", Default: *types.NewToken(types.STRING, "")},
			dialect.FunctionParamDef{Name: "cps", Default: *types.NewToken(types.NUMBER, "5")},
			dialect.FunctionParamDef{Name: "block", Default: *types.NewToken(types.NUMBER, "0")},
		},
	)

	return this
}
