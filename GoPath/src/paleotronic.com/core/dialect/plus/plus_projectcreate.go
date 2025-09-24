package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/api"
//	"paleotronic.com/files"
)

//var backdrop string

type PlusProjectCreate struct {
	dialect.CoreFunction
}

func (this *PlusProjectCreate) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {
		name := this.ValueMap["name"].Content

		if name != "" {
			e := s8webclient.CONN.CreateProjectDir( name )
			if e != nil {
				this.Stack.Push(types.NewToken(types.STRING, e.Error()))
			} else {
				this.Stack.Push(types.NewToken(types.STRING, "Ok"))
			}
		}

	} else {

		this.Stack.Push(types.NewToken(types.STRING, ""))

	}
	return nil
}

func (this *PlusProjectCreate) Syntax() string {

	/* vars */
	var result string

	result = "PROJCREATE{name}"

	/* enforce non void return */
	return result

}

func (this *PlusProjectCreate) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusProjectCreate(a int, b int, params types.TokenList) *PlusProjectCreate {
	this := &PlusProjectCreate{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PROJCREATE"

	this.NamedParams = []string{ "name" }
	this.NamedDefaults = []types.Token{ *types.NewToken( types.STRING, "" ) }
	this.Raw = true

	return this
}
