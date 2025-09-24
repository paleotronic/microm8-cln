package plus

import (
	"paleotronic.com/fmt"
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
)

type PlusCopy struct {
	dialect.CoreFunction
}

func (this *PlusCopy) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		src := this.ValueMap["src"].Content
		dest := this.ValueMap["dest"].Content

		if src == "" {
			return nil
		}

		if dest == "" {
			dest = "."
		}

		wd := this.Interpreter.GetWorkDir()
		wd = strings.Trim(wd, "/")

		var srcpath, destpath string

		if rune(src[0]) != '/' {
			srcpath = wd + "/" + src
		} else {
			srcpath = src
		}

		if rune(dest[0]) != '/' {
			destpath = wd + "/" + dest
		} else {
			destpath = dest
		}

		if dest == "~" {
			destpath = "/" + files.GetFilename(srcpath)
		}

		if dest == "." {
			destpath = wd + "/" + files.GetFilename(srcpath)
		}

		if files.GetFilename(destpath) == "" {
			destpath = destpath + files.GetFilename(srcpath)
		}

		srcpath = strings.Replace(srcpath, "//", "/", -1)
		destpath = strings.Replace(destpath, "//", "/", -1)

		fmt.Printf("srcpath=%s, destpath=%s\n", srcpath, destpath)

		p := files.GetPath(srcpath)
		f := files.GetFilename(srcpath)

		// this.Interpreter.PutStr(p+":"+f+":"+destpath)

		// test if it exists
		if files.ExistsViaProvider(p, f) {

			e := files.CopyFileViaProviders(srcpath, destpath)
			if e != nil {
				this.Interpreter.PutStr(e.Error())
			} else {
				this.Interpreter.PutStr("Ok")
			}

		} else {
			this.Interpreter.PutStr("File Not Found")
		}

	} else {
		this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetWorkDir()))
	}

	return nil
}

func (this *PlusCopy) Syntax() string {

	/* vars */
	var result string

	result = "COPY{src,dest}"

	/* enforce non void return */
	return result

}

func (this *PlusCopy) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusCopy(a int, b int, params types.TokenList) *PlusCopy {
	this := &PlusCopy{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "COPY"

	this.NamedParams = []string{"src", "dest"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.STRING, ""), *types.NewToken(types.STRING, ".")}
	this.Raw = true
	this.EvalSingleParam = false

	return this
}
