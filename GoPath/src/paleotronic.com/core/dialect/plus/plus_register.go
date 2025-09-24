package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	"paleotronic.com/files"
	"paleotronic.com/core/exception"
	//"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/api"
	"strings"
	"paleotronic.com/fmt"
)

type PlusAuthRegister struct {
	dialect.CoreFunction
}

func (this *PlusAuthRegister) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if params.Size() < 9 {
		return exception.NewESyntaxError("REGISTER expects USER, PASSWORD, NAME, GENDER, AGE, CELL, LOCATION, ERROR TRAP, RESPONSE")
	}

	if !params.Get(7).IsNumeric() {
		return exception.NewESyntaxError("ERROR TRAP must BE a number")
	}

	u := strings.ToLower(params.Shift().Content)
	p := strings.ToLower(params.Shift().Content)
	f := strings.ToLower(params.Shift().Content)
    g := strings.ToLower(params.Shift().Content)
    b := strings.ToLower(params.Shift().Content)
    c := strings.ToLower(params.Shift().Content)
    location := strings.ToLower(params.Shift().Content)
    
	l := params.Shift().AsInteger()
	sv := strings.ToLower(params.Shift().Content)

	var err error
	s := s8webclient.CONN
	err = s.Register(u, p, f, g, b, c, location)
	
	fmt.Println("Register error result:", err)

	// Setup sesstion var
	if err != nil {
		fmt.Println("Jumping to line", l)
		list := *types.NewTokenList()
		list.Add(types.NewToken(types.VARIABLE, sv))
		list.Add(types.NewToken(types.ASSIGNMENT, "="))
		list.Add(types.NewToken(types.STRING, err.Error()))

		LPC := *this.Interpreter.GetPC()
		Scope := this.Interpreter.GetCode()

		this.Interpreter.GetDialect().ExecuteDirectCommand(list, this.Interpreter, Scope, &LPC)

		list = *types.NewTokenList()
		list.Add(types.NewToken(types.NUMBER, utils.IntToStr(l)))
		this.Interpreter.GetDialect().GetCommands()["goto"].Execute(nil, this.Interpreter, list, Scope, LPC)

		files.System = true

		if err != nil {
			return err
		}
	} else {
		
		fmt.Println("Continuing...")
		
		// Setup sesstion var
		list := *types.NewTokenList()
		list.Add(types.NewToken(types.VARIABLE, sv))
		list.Add(types.NewToken(types.ASSIGNMENT, "="))
		list.Add(types.NewToken(types.STRING, s.Session))

		files.System = false

		LPC := *this.Interpreter.GetPC()
		Scope := this.Interpreter.GetCode()

		this.Interpreter.GetDialect().ExecuteDirectCommand(list, this.Interpreter, Scope, &LPC)
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusAuthRegister) Syntax() string {

	/* vars */
	var result string

	result = "Register{name,pass,onerror,var}"

	/* enforce non void return */
	return result

}

func (this *PlusAuthRegister) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.STRING)
	result = append(result, types.STRING)
	result = append(result, types.STRING)
	result = append(result, types.STRING)
	result = append(result, types.STRING)
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusAuthRegister(a int, b int, params types.TokenList) *PlusAuthRegister {
	this := &PlusAuthRegister{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Register"
	this.MinParams = 9
	this.MaxParams = 9
	this.Hidden = true

	return this
}
