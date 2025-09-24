package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusBackdropOpacity struct {
	dialect.CoreFunction
}

func (this *PlusBackdropOpacity) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()

	var x types.Token
	if !this.Query {
		x = this.ValueMap["opacity"]
		opacity := float32(x.AsExtended())

		_, backdrop, camidx, _, zoom, zoomf, camtrack := this.Interpreter.GetMemoryMap().IntGetBackdrop(index)

		//fmt.Printf("backdrop=%s, opacity=%f, zoom=%f\n", backdrop, opacity, zoom)

		this.Interpreter.GetMemoryMap().IntSetBackdrop(index, backdrop, camidx, opacity, zoom, zoomf, camtrack)
		this.Stack.Push(types.NewToken(types.STRING, backdrop))
	} else {
		_, backdrop[index], _, _, _, _, _ = this.Interpreter.GetMemoryMap().IntGetBackdrop(index)
		this.Stack.Push(types.NewToken(types.STRING, backdrop[index]))
	}

	return nil
}

func (this *PlusBackdropOpacity) Syntax() string {

	/* vars */
	var result string

	result = "BACKDROP{image}"

	/* enforce non void return */
	return result

}

func (this *PlusBackdropOpacity) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusBackdropOpacity(a int, b int, params types.TokenList) *PlusBackdropOpacity {
	this := &PlusBackdropOpacity{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BACKDROP"

	this.NamedParams = []string{"opacity"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "1.0"),
	}
	this.Raw = true
	this.MinParams = 1
	this.MaxParams = 1

	return this
}
