package plus

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusOverlayFilename struct {
	dialect.CoreFunction
}

func (this *PlusOverlayFilename) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()

	//var x types.Token
	if !this.Query {
		overlay := this.ValueMap["image"].Content

		if !strings.HasPrefix(overlay, "/") && this.Origin.GetWorkDir() != "" {
			overlay = "/" + strings.Trim(this.Origin.GetWorkDir(), "/") + "/" + overlay
		}

		this.Interpreter.GetMemoryMap().IntSetOverlay(index, overlay)
		this.Stack.Push(types.NewToken(types.STRING, overlay))
	} else {
		_, overlay := this.Interpreter.GetMemoryMap().IntGetOverlay(index)
		this.Stack.Push(types.NewToken(types.STRING, overlay))
	}

	return nil
}

func (this *PlusOverlayFilename) Syntax() string {

	/* vars */
	var result string

	result = "BACKDROP{image}"

	/* enforce non void return */
	return result

}

func (this *PlusOverlayFilename) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusOverlayFilename(a int, b int, params types.TokenList) *PlusOverlayFilename {
	this := &PlusOverlayFilename{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BACKDROP"

	this.NamedParams = []string{"image"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, ""),
	}
	this.Raw = true
	this.MinParams = 1
	this.MaxParams = 1

	return this
}
