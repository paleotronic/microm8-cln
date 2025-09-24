package plus

import (
	//	"strings"
	"paleotronic.com/files"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusDOSOPEN struct {
	dialect.CoreFunction
}

func (this *PlusDOSOPEN) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		filename := this.ValueMap["filename"].Content

		if filename == "" {
			return nil
		}

		if rune(filename[0]) != '/' {
			filename = this.Interpreter.GetWorkDir() + filename
		}

		p := files.GetPath(filename)
		f := files.GetFilename(filename)

		e := files.DOSOPEN(p, f, 0)
		if e != nil {
			this.Interpreter.PutStr(e.Error() + "\r\n")
		}

	} else {
		this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetWorkDir()))
	}

	return nil
}

func (this *PlusDOSOPEN) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusDOSOPEN) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusDOSOPEN(a int, b int, params types.TokenList) *PlusDOSOPEN {
	this := &PlusDOSOPEN{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"filename"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.STRING, "")}
	this.Raw = true

	return this
}
