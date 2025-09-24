package plus

import (
	"paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"

	"errors"
)

type PlusAppDBConnect struct {
	dialect.CoreFunction
}

func (this *PlusAppDBConnect) FunctionExecute(params *types.TokenList) error {

	var id int = -1

	var e error

	if e = this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		// check dbname
		dbname := this.ValueMap["dbname"].Content
		if dbname == "" {
			return errors.New("No appname")
		}
		// varname
		varname := this.ValueMap["varname"].Content

		if varname == "" {
			return errors.New("Needs varname")
		}

		// try connect
		appsig := this.Interpreter.GetCode().Checksum()
		id, e = s8webclient.CONN.AppDatabaseConnect("DCA", dbname, appsig)

		// set varname to handle
		tl := types.NewTokenList()
		vn := this.ValueMap["varname"]
		tl.Push(types.NewToken(types.VARIABLE, vn.Content))
		tl.Push(types.NewToken(types.ASSIGNMENT, "="))
		tl.Push(types.NewToken(types.NUMBER, utils.IntToStr(id)))

		//this.Interpreter.PutStr(  this.Interpreter.TokenListAsString(*tl)+ "\r\n" )

		a := this.Interpreter.GetCode()

		this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, a, this.Interpreter.GetPC())
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(id)))

	return nil
}

func (this *PlusAppDBConnect) Syntax() string {

	/* vars */
	var result string

	result = "appdb.connect{dbname}"

	/* enforce non void return */
	return result

}

func (this *PlusAppDBConnect) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusAppDBConnect(a int, b int, params types.TokenList) *PlusAppDBConnect {
	this := &PlusAppDBConnect{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "appdb.connect{dbname}"

	this.NamedDefaults = []types.Token{*types.NewToken(types.STRING, ""), *types.NewToken(types.VARIABLE, "")}
	this.NamedParams = []string{"dbname", "varname"}
	this.Raw = true
	this.MaxParams = 2
	this.MinParams = 2

	this.Hidden = true

	return this
}
