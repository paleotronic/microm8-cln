package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
//	"paleotronic.com/runestring"
     "paleotronic.com/api"
	//"paleotronic.com/log"
	"strings"
)

type PlusUserFirstname struct {
	dialect.CoreFunction
}

func (this *PlusUserFirstname) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

    fn, _ := s8webclient.CONN.GetUserFirstName()
    
    parts := strings.Split(fn, " ")

	this.Stack.Push(types.NewToken(types.STRING, parts[0]))

	return nil
}

func (this *PlusUserFirstname) Syntax() string {

	/* vars */
	var result string

	result = "FIRSTNAME{}"

	/* enforce non void return */
	return result

}

func (this *PlusUserFirstname) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusUserFirstname(a int, b int, params types.TokenList) *PlusUserFirstname {
	this := &PlusUserFirstname{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "USER.FIRSTNAME"

	this.NamedParams = []string{}
	this.NamedDefaults = []types.Token{
	}
	this.Raw = true

	return this
}
