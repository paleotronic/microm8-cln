package plus

import (
//	"strings"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
//	"paleotronic.com/files"
)

type PlusDOSCHAIN struct {
	dialect.CoreFunction
}

func (this *PlusDOSCHAIN) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {
		filename := this.ValueMap["filename"].Content

        if filename == "" {
           return nil
        }

        if rune(filename[0]) != '/' {
           filename = this.Interpreter.GetWorkDir() + filename
        }

        ent := this.Interpreter

		tl := types.NewTokenList()
        tl.Push( types.NewToken(types.STRING, filename) )
        a := ent.GetCode()
		_, e := ent.GetDialect().GetCommands()["load"].Execute(nil, ent, *tl, a, *ent.GetLPC())
    	if e != nil {
           return e
        }
        a = ent.GetCode()
        pc := ent.GetPC()
        pc.Line = a.GetLowIndex()
        pc.Statement = 0
        pc.Token = 0
        ent.SetPC( pc )
        ent.SetState( types.RUNNING )

	} else {
		this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetWorkDir()))
	}

	return nil
}

func (this *PlusDOSCHAIN) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusDOSCHAIN) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusDOSCHAIN(a int, b int, params types.TokenList) *PlusDOSCHAIN {
	this := &PlusDOSCHAIN{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{ "filename" }
	this.NamedDefaults = []types.Token{ *types.NewToken( types.STRING, "" ) }
	this.Raw = true

	return this
}
