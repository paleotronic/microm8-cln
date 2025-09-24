package plus

import (
	"paleotronic.com/fmt"
	"strings"

	"paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	//"errors"
)

type PlusAppDBFetch struct {
	dialect.CoreFunction
}

func (this *PlusAppDBFetch) FunctionExecute(params *types.TokenList) error {

	var stmtid int

	var e error

	if e = this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	stmtid = params.Shift().AsInteger()
	template := params.Shift().Content
	varname := params.Shift().Content

	rec, e := s8webclient.CONN.DBResultFetch(stmtid)

	if e != nil {
		return e
	}

	// got a record
	wanted := strings.Split(template, ",")

	for i, field := range wanted {
		targetidx := fmt.Sprintf("%d", i+1)
		value, exists := rec[field]
		if !exists {
			value = ""
		}
		tl := types.NewTokenList()
		tl.Push(types.NewToken(types.VARIABLE, varname))
		tl.Push(types.NewToken(types.OBRACKET, "("))
		tl.Push(types.NewToken(types.NUMBER, targetidx))
		tl.Push(types.NewToken(types.CBRACKET, ")"))
		tl.Push(types.NewToken(types.ASSIGNMENT, "="))
		tl.Push(types.NewToken(types.STRING, fmt.Sprintf("%v", value)))

		a := this.Interpreter.GetCode()

		this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, a, this.Interpreter.GetPC())
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(len(wanted))))

	return nil
}

func (this *PlusAppDBFetch) Syntax() string {

	/* vars */
	var result string

	result = "appdb.resultcount{stmtid}"

	/* enforce non void return */
	return result

}

func (this *PlusAppDBFetch) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.STRING)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusAppDBFetch(a int, b int, params types.TokenList) *PlusAppDBFetch {
	this := &PlusAppDBFetch{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "appdb.resultcount{stmtid,template,target}"

	this.MaxParams = 3
	this.MinParams = 3

	this.Hidden = true

	return this
}
