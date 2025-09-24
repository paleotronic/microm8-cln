package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/api"
	"paleotronic.com/files"
)

//var backdrop string

type PlusProjectUse struct {
	dialect.CoreFunction
}

func (this *PlusProjectUse) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {
		name := this.ValueMap["name"].Content

		if name != "" {

			//files.Project = true

			ex, _, _, e := s8webclient.CONN.ProjectStatus( name )

			if e == nil {

				if ex  {
					files.SetProject(name)
				} else {
					this.Interpreter.PutStr("Project does not exist: "+name+"\r\n")
				}

			}

		}

	} else {

		this.Stack.Push(types.NewToken(types.STRING, ""))

	}
	return nil
}

func (this *PlusProjectUse) Syntax() string {

	/* vars */
	var result string

	result = "PROJECT.USE{name}"

	/* enforce non void return */
	return result

}

func (this *PlusProjectUse) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusProjectUse(a int, b int, params types.TokenList) *PlusProjectUse {
	this := &PlusProjectUse{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PROJECT.USE"

	this.NamedParams = []string{ "name" }
	this.NamedDefaults = []types.Token{ *types.NewToken( types.STRING, "" ) }
	this.Raw = true

	return this
}
