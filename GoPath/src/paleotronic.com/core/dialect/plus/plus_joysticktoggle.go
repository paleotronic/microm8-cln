package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
    "errors"
)

type PlusJoyToggle struct {
	dialect.CoreFunction
}

func (this *PlusJoyToggle) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	at := this.ValueMap["a"]; a := at.AsInteger()
    bt := this.ValueMap["b"]; b := bt.AsInteger()
    
    if a < 0 || b < 0 || a > 3 || b > 3 {
       this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
       return errors.New("invalid parameters")
    }
    
    if a == b {
      this.Interpreter.GetMemoryMap().PaddleMap[this.Interpreter.GetMemIndex()][0] = 0
      this.Interpreter.GetMemoryMap().PaddleMap[this.Interpreter.GetMemIndex()][1] = 1
      this.Interpreter.GetMemoryMap().PaddleMap[this.Interpreter.GetMemIndex()][2] = 2
      this.Interpreter.GetMemoryMap().PaddleMap[this.Interpreter.GetMemIndex()][3] = 3
      return nil        
    }
    
    this.Interpreter.GetMemoryMap().PaddleMap[this.Interpreter.GetMemIndex()][a] = b
    this.Interpreter.GetMemoryMap().PaddleMap[this.Interpreter.GetMemIndex()][b] = a

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusJoyToggle) Syntax() string {

	/* vars */
	var result string

	result = "PDLSWITCH{}"

	/* enforce non void return */
	return result

}

func (this *PlusJoyToggle) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusJoyToggle(a int, b int, params types.TokenList) *PlusJoyToggle {
	this := &PlusJoyToggle{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "JOYTOGGLE"
    this.Raw = true
    
    this.NamedParams = []string{"a", "b"}
    this.NamedDefaults = []types.Token{
        *types.NewToken(types.NUMBER, "0"),
        *types.NewToken(types.NUMBER, "0"),
    }

	return this
}
