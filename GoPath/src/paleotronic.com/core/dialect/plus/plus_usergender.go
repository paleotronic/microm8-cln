package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
//	"paleotronic.com/runestring"
     "paleotronic.com/api"
	//"paleotronic.com/log"
)

type PlusUserGender struct {
	dialect.CoreFunction
}

func (this *PlusUserGender) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

    fn, _ := s8webclient.CONN.GetUserGender()

	this.Stack.Push(types.NewToken(types.STRING, fn))

	return nil
}

func (this *PlusUserGender) Syntax() string {

	/* vars */
	var result string

	result = "FIRSTNAME{}"

	/* enforce non void return */
	return result

}

func (this *PlusUserGender) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusUserGender(a int, b int, params types.TokenList) *PlusUserGender {
	this := &PlusUserGender{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "USER.FIRSTNAME"

	this.NamedParams = []string{}
	this.NamedDefaults = []types.Token{
	}
	this.Raw = true

	return this
}
