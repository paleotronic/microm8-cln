package plus

import (
	"paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"

	"errors"
)

type PlusAppDBQuery struct {
	dialect.CoreFunction
}

func (this *PlusAppDBQuery) FunctionExecute(params *types.TokenList) error {

	var stmtid int

	var e error

	if e = this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		// check dbname
		h := this.ValueMap["handle"]
		handle := h.AsInteger()

		// varname
		varname := this.ValueMap["varname"].Content

		if varname == "" {
			return errors.New("Needs varname")
		}

		// query
		query := this.ValueMap["query"].Content
		if query == "" {
			return errors.New("Needs query")
		}

		// try connect
		appsig := this.Interpreter.GetCode().Checksum()
		stmtid, e = s8webclient.CONN.AppDatabaseQuery("DQE", handle, appsig, query)

		// set varname to handle
		tl := this.Interpreter.GetDialect().Tokenize(runestring.Cast(this.ValueMap["varname"].Content))
		tl.Push(types.NewToken(types.ASSIGNMENT, "="))
		tl.Push(types.NewToken(types.NUMBER, utils.IntToStr(stmtid)))

		//this.Interpreter.PutStr(  this.Interpreter.TokenListAsString(*tl)+ "\r\n" )

		a := this.Interpreter.GetCode()

		this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, a, this.Interpreter.GetPC())
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(stmtid)))

	return nil
}

func (this *PlusAppDBQuery) Syntax() string {

	/* vars */
	var result string

	result = "appdb.connect{dbname}"

	/* enforce non void return */
	return result

}

func (this *PlusAppDBQuery) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusAppDBQuery(a int, b int, params types.TokenList) *PlusAppDBQuery {
	this := &PlusAppDBQuery{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "appdb.query{dbname}"

	this.NamedDefaults = []types.Token{*types.NewToken(types.NUMBER, "0"), *types.NewToken(types.STRING, ""), *types.NewToken(types.STRING, "")}
	this.NamedParams = []string{"handle", "query", "varname"}
	this.Raw = true
	this.MaxParams = 2
	this.MinParams = 2

	this.Hidden = true

	return this
}
