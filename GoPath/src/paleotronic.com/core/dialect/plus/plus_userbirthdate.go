package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
//	"paleotronic.com/runestring"
     "paleotronic.com/api"
	//"paleotronic.com/log"
)

type PlusUserDOB struct {
	dialect.CoreFunction
}

func (this *PlusUserDOB) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

    fn, _ := s8webclient.CONN.GetUserDOB()

	this.Stack.Push(types.NewToken(types.STRING, fn))

	return nil
}

func (this *PlusUserDOB) Syntax() string {

	/* vars */
	var result string

	result = "FIRSTNAME{}"

	/* enforce non void return */
	return result

}

func (this *PlusUserDOB) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusUserDOB(a int, b int, params types.TokenList) *PlusUserDOB {
	this := &PlusUserDOB{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "USER.FIRSTNAME"

	this.NamedParams = []string{}
	this.NamedDefaults = []types.Token{
	}
	this.Raw = true

	return this
}
