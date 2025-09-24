package plus

import (
	"paleotronic.com/fmt"
	"strings"

	"paleotronic.com/core/hardware/apple2helpers"

	"paleotronic.com/utils"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
)

type PlusDir struct {
	dialect.CoreFunction
}

func (this *PlusDir) FunctionExecute(params *types.TokenList) error {

	fmt.Printf("function invoked with %v\n", params.AsString())

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		c := this.ValueMap["path"].Content
		//c := t.Content
		s := this.ValueMap["spec"].Content

		fmt.Printf("path=%s, spec=%s\n", c, s)

		wd := this.Interpreter.GetWorkDir()
		wd = strings.Trim(wd, "/")

		fullpath := c

		if wd != "" && !strings.HasPrefix(c, "/") {
			fullpath = "/" + wd + "/" + c
		}

		if c == "~" {
			fullpath = ""
		}

		if !strings.HasSuffix(fullpath, "/") {
			fullpath = fullpath + "/"
		}

		fmt.Printf("ls debug: fullpath=%s, pattern=%s\n", fullpath, s)

		p := files.GetPath(fullpath)
		f := files.GetFilename(fullpath)

		// test if it exists
		if fullpath == "" || files.ExistsViaProvider(p, f) {

			// Dir files
			d, f, e := files.ReadDirViaProvider(fullpath, s)

			if e != nil {
				this.Interpreter.PutStr("I/O Error: " + e.Error() + "\r\n")
				return nil
			}
			if fullpath == "" {
				fullpath = "/"
			}
			this.Interpreter.PutStr("Directory of " + fullpath + "\r\n")
			for _, de := range d {
				this.Interpreter.PutStr(fmt.Sprintf(
					"%-20s %-3s %-8d\r\n",
					utils.Overflow(de.Name, "...", 20),
					"dir",
					de.Size,
				))
			}
			for _, fe := range f {
				name := fe.Name
				if fe.Extension != "" {
					name = name + "." + fe.Extension
				}
				extra := ""
				if apple2helpers.GetColumns(this.Interpreter) == 80 {
					extra = fe.Description
				}
				this.Interpreter.PutStr(fmt.Sprintf(
					"%-20s %-3s %-8d %s\r\n",
					utils.Overflow(fe.Name, "...", 20),
					fe.Extension,
					fe.Size,
					utils.Overflow(extra, "...", 38),
				))
			}

		} else {
			this.Interpreter.PutStr("Directory does not exist")
		}

	} else {
		this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetWorkDir()))
	}

	return nil
}

func (this *PlusDir) Syntax() string {

	/* vars */
	var result string

	result = "DIR{path,spec}"

	/* enforce non void return */
	return result

}

func (this *PlusDir) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusDir(a int, b int, params types.TokenList) *PlusDir {
	this := &PlusDir{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "DIR"

	this.NamedParams = []string{"path", "spec"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.STRING, ""), *types.NewToken(types.STRING, "*.*")}
	this.Raw = true

	return this
}
