package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/api"
	)

type PlusRevoke struct {
	dialect.CoreFunction
}

func (this *PlusRevoke) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {

		user := this.ValueMap["username"].Content
		mode := this.ValueMap["action"].Content
		group := this.ValueMap["group"].Content

		if mode == "read" {
			e := s8webclient.CONN.RevokeReadGroup( user, group )
			if e != nil {
				this.Interpreter.PutStr( e.Error() )
			} else {
				this.Interpreter.PutStr( "Ok" )
			}
		} else if mode == "write" {
			e := s8webclient.CONN.RevokeWriteGroup( user, group )
			if e != nil {
				this.Interpreter.PutStr( e.Error() )
			} else {
				this.Interpreter.PutStr( "Ok" )
			}
		}

	} else {
		this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetWorkDir()))
	}

	return nil
}

func (this *PlusRevoke) Syntax() string {

	/* vars */
	var result string

	result = "REVOKE{username,mode,group}"

	/* enforce non void return */
	return result

}

func (this *PlusRevoke) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusRevoke(a int, b int, params types.TokenList) *PlusRevoke {
	this := &PlusRevoke{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "REVOKE"

	this.NamedParams = []string{"username", "action", "group"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.STRING, ""), *types.NewToken(types.STRING, "read"), *types.NewToken(types.STRING, "")}
	this.Raw = true

	return this
}
