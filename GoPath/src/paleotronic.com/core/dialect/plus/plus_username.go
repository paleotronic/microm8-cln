package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/runestring"
	"paleotronic.com/api"
	//"paleotronic.com/log"
)

type PlusUserName struct {
	dialect.CoreFunction
}

func (this *PlusUserName) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	fn := s8webclient.CONN.Username

	this.Stack.Push(types.NewToken(types.STRING, fn))

	return nil
}

func (this *PlusUserName) Syntax() string {

	/* vars */
	var result string

	result = "FIRSTNAME{}"

	/* enforce non void return */
	return result

}

func (this *PlusUserName) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusUserName(a int, b int, params types.TokenList) *PlusUserName {
	this := &PlusUserName{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "USER.FIRSTNAME"

	this.MinParams = 0
	this.MaxParams = 0

	return this
}
