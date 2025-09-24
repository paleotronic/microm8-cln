package plus

import (
	"paleotronic.com/log"
	"strings"

	"paleotronic.com/files"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusDelete struct {
	dialect.CoreFunction
}

func (this *PlusDelete) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {
		t := this.ValueMap["path"]
		c := t.Content

		wd := this.Interpreter.GetWorkDir()
		wd = strings.Trim(wd, "/")

		fullpath := c

		if wd != "" {
			fullpath = wd + "/" + c
		}

		if c == "~" {
			fullpath = ""
		}

		log.Printf("fullpath = %s\n", fullpath)

		p := files.GetPath(fullpath)
		f := files.GetFilename(fullpath)

		// test if it exists
		if files.ExistsViaProvider(p, f) {

			e := files.DeleteViaProvider(fullpath)

			if e == nil {
				this.Interpreter.PutStr("Ok")
			} else {
				this.Interpreter.PutStr("Delete failed")
			}

		} else {
			this.Interpreter.PutStr("Path does not exist")
		}

	} else {
		this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetWorkDir()))
	}

	return nil
}

func (this *PlusDelete) Syntax() string {

	/* vars */
	var result string

	result = "DEL{path}"

	/* enforce non void return */
	return result

}

func (this *PlusDelete) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusDelete(a int, b int, params types.TokenList) *PlusDelete {
	this := &PlusDelete{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "DEL"

	this.NamedParams = []string{"path"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.STRING, "~")}
	this.Raw = true

	return this
}
