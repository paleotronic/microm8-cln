package plus

import (
//	"strings"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
)

type PlusDOSCLOSE struct {
	dialect.CoreFunction
}

func (this *PlusDOSCLOSE) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {
		filename := this.ValueMap["filename"].Content

        if filename == "" {
           _ = files.DOSCLOSEALL()
           this.Interpreter.SetOutChannel("")
           this.Interpreter.SetInChannel("")
           return nil
        }

        if rune(filename[0]) != '/' {
           filename = this.Interpreter.GetWorkDir() + filename
        }

        p := files.GetPath(filename)
        f := files.GetFilename(filename)

        e := files.DOSCLOSE(p, f)
        if e != nil {
           this.Interpreter.PutStr( e.Error()+"\r\n" )
           this.Interpreter.SetOutChannel("")
           this.Interpreter.SetInChannel("")
        }

	} else {
		this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetWorkDir()))
	}

	return nil
}

func (this *PlusDOSCLOSE) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusDOSCLOSE) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusDOSCLOSE(a int, b int, params types.TokenList) *PlusDOSCLOSE {
	this := &PlusDOSCLOSE{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{ "filename" }
	this.NamedDefaults = []types.Token{ *types.NewToken( types.STRING, "" ) }
	this.Raw = true

	return this
}
