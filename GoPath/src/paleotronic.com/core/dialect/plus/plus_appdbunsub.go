package plus

import (
	//"paleotronic.com/runestring"
	"paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	//"errors"
)

type PlusAppDBUnsub struct {
	dialect.CoreFunction
}

func (this *PlusAppDBUnsub) FunctionExecute(params *types.TokenList) error {

	//	var stmtid int

	var e error

	if e = this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	s8webclient.CONN.DBUnsubscribeALL()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusAppDBUnsub) Syntax() string {

	/* vars */
	var result string

	result = "appdb.connect{dbname}"

	/* enforce non void return */
	return result

}

func (this *PlusAppDBUnsub) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusAppDBUnsub(a int, b int, params types.TokenList) *PlusAppDBUnsub {
	this := &PlusAppDBUnsub{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "appdb.query{dbname}"
	this.MinParams = 0
	this.MaxParams = 0

	this.Hidden = true

	return this
}
