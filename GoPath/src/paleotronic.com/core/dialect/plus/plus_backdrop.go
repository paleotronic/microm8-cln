package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
)

var backdrop [memory.OCTALYZER_NUM_INTERPRETERS]string

type PlusBackdrop struct {
	dialect.CoreFunction
}

func (this *PlusBackdrop) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()

	var x types.Token
	if !this.Query {
		backdrop[index] = this.ValueMap["image"].Content
		x = this.ValueMap["opacity"]
		opacity := float32(x.AsExtended())
		x = this.ValueMap["zoom"]
		zoom := float32(x.AsExtended())
		x = this.ValueMap["zoomfactor"]
		zoomf := float32(x.AsExtended())
		x = this.ValueMap["camera"]
		camidx := x.AsInteger()
		x = this.ValueMap["camtrack"]
		camtrack := x.AsInteger() != 0
		this.Interpreter.GetMemoryMap().IntSetBackdrop(index, backdrop[index], camidx, opacity, zoom, zoomf, camtrack)
		this.Stack.Push(types.NewToken(types.STRING, backdrop[index]))
	} else {
		_, backdrop[index], _, _, _, _, _ = this.Interpreter.GetMemoryMap().IntGetBackdrop(index)
		this.Stack.Push(types.NewToken(types.STRING, backdrop[index]))
	}

	return nil
}

func (this *PlusBackdrop) Syntax() string {

	/* vars */
	var result string

	result = "BACKDROP{image}"

	/* enforce non void return */
	return result

}

func (this *PlusBackdrop) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusBackdrop(a int, b int, params types.TokenList) *PlusBackdrop {
	this := &PlusBackdrop{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BACKDROP"

	this.NamedParams = []string{"image", "opacity", "zoom", "zoomfactor", "camera", "camtrack"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, ""),
		*types.NewToken(types.NUMBER, "1.0"),
		*types.NewToken(types.NUMBER, "1.0"),
		*types.NewToken(types.NUMBER, "0.0"),
		*types.NewToken(types.NUMBER, "7"),
		*types.NewToken(types.NUMBER, "1"),
	}
	this.Raw = true
	this.MinParams = 1
	this.MaxParams = 5

	return this
}
