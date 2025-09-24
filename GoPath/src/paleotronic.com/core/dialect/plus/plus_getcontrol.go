package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//"paleotronic.com/utils"
//	"paleotronic.com/log"
)

type PlusGetControls struct {
	dialect.CoreFunction
}

func (this *PlusGetControls) FunctionExecute(params *types.TokenList) error {

	//~ if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	//~ if !this.Query {

		//~ control := this.ValueMap["control"].Content
		//~ varname := this.ValueMap["varname"].Content

		//~ log.Printf("Got params: %s, %s\n", control, varname)

		//~ s, e := this.Interpreter.GetControlState( control )
		//~ if e != nil {
			//~ a := this.Interpreter.GetCode()
			//~ tl := types.NewTokenList()
			//~ tl.Push( types.NewToken( types.VARIABLE, varname ) )
			//~ tl.Push( types.NewToken( types.ASSIGNMENT, "=" ) )
			//~ tl.Push( types.NewToken( types.STRING, s ) )
			//~ this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, &a, this.Interpreter.GetPC())
			//~ log.Println(tl.AsString())
		//~ }

	//~ }

	return nil
}

func (this *PlusGetControls) Syntax() string {

	/* vars */
	var result string

	result = "GetControls{type,varname}"

	/* enforce non void return */
	return result

}

func (this *PlusGetControls) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusGetControls(a int, b int, params types.TokenList) *PlusGetControls {
	this := &PlusGetControls{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "GETCONTROLS"

	this.NamedParams = []string{ "control", "varname" }
	this.NamedDefaults = []types.Token{
		*types.NewToken( types.STRING, "keys" ),
		*types.NewToken( types.VARIABLE, "CO$" ),
	}
	this.Raw = true

	return this
}
