package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusBackdropCamTrack struct {
	dialect.CoreFunction
}

func (this *PlusBackdropCamTrack) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()

	var x types.Token
	if !this.Query {
		x = this.ValueMap["camtrack"]
		camtrack := x.AsExtended() != 0

		_, backdrop, camidx, opacity, zoom, zoomf, _ := this.Interpreter.GetMemoryMap().IntGetBackdrop(index)

		//fmt.Printf("backdrop=%s, opacity=%f, zoom=%f\n", backdrop, opacity, zoom)

		this.Interpreter.GetMemoryMap().IntSetBackdrop(index, backdrop, camidx, opacity, zoom, zoomf, camtrack)
		this.Stack.Push(types.NewToken(types.STRING, backdrop))
	} else {
		_, backdrop[index], _, _, _, _, _ = this.Interpreter.GetMemoryMap().IntGetBackdrop(index)
		this.Stack.Push(types.NewToken(types.STRING, backdrop[index]))
	}

	return nil
}

func (this *PlusBackdropCamTrack) Syntax() string {

	/* vars */
	var result string

	result = "BACKDROP{image}"

	/* enforce non void return */
	return result

}

func (this *PlusBackdropCamTrack) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusBackdropCamTrack(a int, b int, params types.TokenList) *PlusBackdropCamTrack {
	this := &PlusBackdropCamTrack{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BACKDROP"

	this.NamedParams = []string{"camtrack"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0.0"),
	}
	this.Raw = true
	this.MinParams = 1
	this.MaxParams = 1

	return this
}
