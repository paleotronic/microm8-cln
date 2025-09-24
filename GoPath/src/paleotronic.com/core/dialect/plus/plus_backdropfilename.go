package plus

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"

	"paleotronic.com/fmt"
)

type PlusBackdropFilename struct {
	dialect.CoreFunction
}

func (this *PlusBackdropFilename) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()

	//var x types.Token
	if !this.Query {
		backdrop[index] = this.ValueMap["image"].Content

		fmt.Printf("backdrop=%s\n", backdrop[index])

		if !strings.HasPrefix(backdrop[index], "/") && this.Origin.GetWorkDir() != "" {
			backdrop[index] = "/" + strings.Trim(this.Origin.GetWorkDir(), "/") + "/" + backdrop[index]
			fmt.Printf("backdrop.ammended=%s\n", backdrop[index])
		}

		_, _, camidx, opacity, zoom, zoomf, camtrack := this.Interpreter.GetMemoryMap().IntGetBackdrop(index)

		this.Interpreter.GetMemoryMap().IntSetBackdrop(index, backdrop[index], camidx, opacity, zoom, zoomf, camtrack)
		this.Stack.Push(types.NewToken(types.STRING, backdrop[index]))
	} else {
		_, backdrop[index], _, _, _, _, _ = this.Interpreter.GetMemoryMap().IntGetBackdrop(index)
		this.Stack.Push(types.NewToken(types.STRING, backdrop[index]))
	}

	return nil
}

func (this *PlusBackdropFilename) Syntax() string {

	/* vars */
	var result string

	result = "BACKDROP{image}"

	/* enforce non void return */
	return result

}

func (this *PlusBackdropFilename) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusBackdropFilename(a int, b int, params types.TokenList) *PlusBackdropFilename {
	this := &PlusBackdropFilename{}

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
