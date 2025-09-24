package plus

import (
	"paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	//"errors"
)

type PlusAppDBResultCount struct {
	dialect.CoreFunction
}

func (this *PlusAppDBResultCount) FunctionExecute(params *types.TokenList) error {

	var stmtid int

	var e error

	if e = this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	stmtid = params.Shift().AsInteger()
	count, e := s8webclient.CONN.DBResultCount(stmtid)

	if e != nil {
		return e
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(count)))

	return nil
}

func (this *PlusAppDBResultCount) Syntax() string {

	/* vars */
	var result string

	result = "appdb.resultcount{stmtid}"

	/* enforce non void return */
	return result

}

func (this *PlusAppDBResultCount) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusAppDBResultCount(a int, b int, params types.TokenList) *PlusAppDBResultCount {
	this := &PlusAppDBResultCount{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "appdb.resultcount{stmtid}"

	this.MaxParams = 1
	this.MinParams = 1

	this.Hidden = true

	return this
}
