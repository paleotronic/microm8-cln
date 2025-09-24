package plus

import (
	"strings"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
)

type PlusMount struct {
	dialect.CoreFunction
}

func (this *PlusMount) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {
		t := this.ValueMap["path"]
		c := t.Content
		dt := this.ValueMap["drive"]
		drive := dt.AsInteger()

		wd := this.Interpreter.GetWorkDir()
		wd = strings.Trim(wd, "/")

		fullpath := wd + "/" + c

		p := files.GetPath(fullpath)
		f := strings.ToLower(files.GetFilename(fullpath))

		// test if it exists
		if files.ExistsViaProvider( p, f ) {
			if files.GetExt(f) == "dsk" {
				mount, e := files.MountDSKImage( p, f, drive )
				if e == nil {
					this.Interpreter.SetWorkDir(mount)
					this.Interpreter.PutStr("Switched to directory "+mount+"\r\n")
				}
			}
		}

	} 

	return nil
}

func (this *PlusMount) Syntax() string {

	/* vars */
	var result string

	result = "CD{path}"

	/* enforce non void return */
	return result

}

func (this *PlusMount) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusMount(a int, b int, params types.TokenList) *PlusMount {
	this := &PlusMount{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CD"

	this.NamedParams = []string{ "path", "drive" }
	this.NamedDefaults = []types.Token{ *types.NewToken( types.STRING, "" ), *types.NewToken( types.NUMBER, "0" ) }
	this.Raw = true

	return this
}
