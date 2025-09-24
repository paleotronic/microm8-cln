package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	"paleotronic.com/files"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/api"
	"strings"
)

type PlusAuthLogin struct {
	dialect.CoreFunction
}

func (this *PlusAuthLogin) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if params.Size() < 4 {
		return exception.NewESyntaxError("LOGIN expects USER, PASSWORD, ERROR TRAP, SESSION VAR")
	}

	if !params.Get(2).IsNumeric() {
		return exception.NewESyntaxError("ERROR TRAP must BE a number")
	}

	u := strings.ToLower(params.Shift().Content)
	p := strings.ToLower(params.Shift().Content)
	l := params.Shift().AsInteger()
	sv := strings.ToLower(params.Shift().Content)

	//
	code, err := s8webclient.CONN.Login(u, p)

	if err != nil || code >= 300 {
		// failed
		list := *types.NewTokenList()
		list.Add(types.NewToken(types.NUMBER, utils.IntToStr(l)))
		Scope := this.Interpreter.GetCode()
		LPC := *this.Interpreter.GetPC()
		this.Interpreter.GetDialect().GetCommands()["goto"].Execute(nil, this.Interpreter, list, Scope, LPC)

		files.System = true

		if err != nil {
			return err
		}
	} else {
		apple2helpers.PutStr(this.Interpreter, "Ok\r\n")

		// Setup sesstion var
		list := *types.NewTokenList()
		list.Add(types.NewToken(types.VARIABLE, sv))
		list.Add(types.NewToken(types.ASSIGNMENT, "="))
		list.Add(types.NewToken(types.STRING, s8webclient.CONN.Session))

		Scope := this.Interpreter.GetCode()
		LPC := this.Interpreter.GetPC()
		this.Interpreter.GetDialect().ExecuteDirectCommand(list, this.Interpreter, Scope, LPC)

		files.System = false
        
        // Hook here to get dir of projects folder
        _, _, _ = files.ReadDirViaProvider("/projects", "*.*")
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusAuthLogin) Syntax() string {

	/* vars */
	var result string

	result = "Login{name,pass,onerror,var}"

	/* enforce non void return */
	return result

}

func (this *PlusAuthLogin) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusAuthLogin(a int, b int, params types.TokenList) *PlusAuthLogin {
	this := &PlusAuthLogin{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Login"
	this.MinParams = 4
	this.MaxParams = 4
	this.Hidden = true

	return this
}
