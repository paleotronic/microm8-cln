package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"strings"
)

type PlusMkDir struct {
	dialect.CoreFunction
}

func (this *PlusMkDir) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {
		t := this.ValueMap["path"]
		c := t.Content

		wd := this.Interpreter.GetWorkDir()
		wd = strings.Trim(wd, "/")

		fullpath := wd + "/" + c

		if c == "~" {
			fullpath = ""
		}

		if fullpath != "" {

			e := files.MkdirViaProvider(fullpath)

			if e == nil {
				this.Interpreter.PutStr("Ok\r\n")
			} else {
				this.Interpreter.PutStr(e.Error()+"\r\n")
			}

		}
	}
	this.Stack.Push(types.NewToken(types.NUMBER, "0"))

	return nil
}

func (this *PlusMkDir) Syntax() string {

	/* vars */
	var result string

	result = "MKDIR{path}"

	/* enforce non void return */
	return result

}

func (this *PlusMkDir) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusMkDir(a int, b int, params types.TokenList) *PlusMkDir {
	this := &PlusMkDir{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "MKDIR"

	this.NamedParams = []string{ "path" }
	this.NamedDefaults = []types.Token{ *types.NewToken( types.STRING, "" ) }
	this.Raw = true

	return this
}
