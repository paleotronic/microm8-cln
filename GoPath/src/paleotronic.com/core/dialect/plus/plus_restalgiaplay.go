package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
)

var audiofile [memory.OCTALYZER_NUM_INTERPRETERS]string

type PlusRestalgiaPlay struct {
	dialect.CoreFunction
}

func (this *PlusRestalgiaPlay) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()

	var x types.Token
	if !this.Query {
		audiofile[index] = this.ValueMap["file"].Content
		loop := x.AsInteger() != 0
		this.Interpreter.GetMemoryMap().IntSetRestalgiaPath(index, audiofile[index], loop)
		this.Stack.Push(types.NewToken(types.STRING, audiofile[index]))
	} else {
		_, audiofile[index], _ = this.Interpreter.GetMemoryMap().IntGetRestalgiaPath(index)
		this.Stack.Push(types.NewToken(types.STRING, audiofile[index]))
	}

	return nil
}

func (this *PlusRestalgiaPlay) Syntax() string {

	/* vars */
	var result string

	result = "BACKDROP{image}"

	/* enforce non void return */
	return result

}

func (this *PlusRestalgiaPlay) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusRestalgiaPlay(a int, b int, params types.TokenList) *PlusRestalgiaPlay {
	this := &PlusRestalgiaPlay{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BACKDROP"

	this.NamedParams = []string{"file", "loop"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, ""),
		*types.NewToken(types.NUMBER, "1"),
	}
	this.Raw = true
	this.MinParams = 1
	this.MaxParams = 5

	return this
}
