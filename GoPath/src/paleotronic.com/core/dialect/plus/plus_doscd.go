package plus

import (
	"paleotronic.com/fmt"
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
)

type PlusCd struct {
	dialect.CoreFunction
}

func (this *PlusCd) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		t := this.ValueMap["path"]
		c := t.Content

		wd := this.Interpreter.GetWorkDir()
		wd = strings.Trim(wd, "/")

		fullpath := wd + "/" + c

		if c == "~" {
			fullpath = ""
		}

		if c == ".." {
			parts := strings.Split(wd, "/")
			parts = parts[0 : len(parts)-1]
			if len(parts) == 0 {
				fullpath = ""
			} else {
				fullpath = strings.Join(parts, "/")
			}
		} else if rune(c[0]) == '/' {
			fullpath = c
		}

		p := files.GetPath(fullpath)
		f := files.GetFilename(fullpath)

		fmt.Printf("fullpath = %s\n", fullpath)

		// test if it exists
		if fullpath == "" || files.ExistsViaProvider(p, f) {
			this.Interpreter.SetWorkDir(fullpath)
			if fullpath == "" {
				fullpath = "/"
			}
			this.Interpreter.PutStr("Working dir is now: " + fullpath)
		} else {
			this.Interpreter.PutStr("Directory does not exist")
		}

	} else {
		this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetWorkDir()))
	}

	return nil
}

func (this *PlusCd) Syntax() string {

	/* vars */
	var result string

	result = "CD{path}"

	/* enforce non void return */
	return result

}

func (this *PlusCd) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusCd(a int, b int, params types.TokenList) *PlusCd {
	this := &PlusCd{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CD"

	this.NamedParams = []string{"path"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.STRING, "~")}
	this.Raw = true

	return this
}
