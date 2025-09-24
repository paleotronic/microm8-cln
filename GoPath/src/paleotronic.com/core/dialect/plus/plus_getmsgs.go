package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
//	"paleotronic.com/utils"
	//"paleotronic.com/log"
)

type PlusGetMsgs struct {
	dialect.CoreFunction
}

func (this *PlusGetMsgs) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	//~ if !this.Query {

		//~ namevar := this.ValueMap["namevar"].Content
		//~ msgvar := this.ValueMap["msgvar"].Content
		//~ countvar := this.ValueMap["countvar"].Content
		//~ max := utils.StrToInt(this.ValueMap["max"].Content)

		//~ //log.Printf("Parameters: %s, %s, %d\n", namevar, msgvar, max)

		//~ n, m, err := this.Interpreter.GetChatMessages(max)

		//~ //log.Printf("n = %v\n", n)
		//~ //log.Printf("m = %v\n", m)
		//~ //log.Printf("err = %v\n", err)

		//~ if err == nil {

			//~ for i := 0; i<len(n); i++ {
				//~ //log.Printf("*** <%s> %s\n", n[i], m[i])
				//~ tl := types.NewTokenList()
				//~ tl.Push( types.NewToken(types.VARIABLE, namevar) )
				//~ tl.Push( types.NewToken(types.OBRACKET, "(") )
				//~ tl.Push( types.NewToken(types.NUMBER, utils.IntToStr(i+1)))
				//~ tl.Push( types.NewToken(types.CBRACKET, ")") )
				//~ tl.Push( types.NewToken(types.ASSIGNMENT, "=") )
				//~ tl.Push( types.NewToken(types.STRING, n[i]) )
				//~ a := this.Interpreter.GetCode()

				//~ //log.Println(tl.AsString())

				//~ this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, &a, this.Interpreter.GetPC())

				//~ tl = types.NewTokenList()
				//~ tl.Push( types.NewToken(types.VARIABLE, msgvar) )
				//~ tl.Push( types.NewToken(types.OBRACKET, "(") )
				//~ tl.Push( types.NewToken(types.NUMBER, utils.IntToStr(i+1)))
				//~ tl.Push( types.NewToken(types.CBRACKET, ")") )
				//~ tl.Push( types.NewToken(types.ASSIGNMENT, "=") )
				//~ tl.Push( types.NewToken(types.STRING, m[i]) )
				//~ this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, &a, this.Interpreter.GetPC())

				//~ //log.Println(tl.AsString())

				//~ tl = types.NewTokenList()
				//~ tl.Push( types.NewToken(types.VARIABLE, countvar) )
				//~ tl.Push( types.NewToken(types.ASSIGNMENT, "=") )
				//~ tl.Push( types.NewToken(types.NUMBER, utils.IntToStr(i+1)))
				//~ this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, &a, this.Interpreter.GetPC())

				//~ //log.Println(tl.AsString())
			//~ }

			//~ //cvar.SetContentScalar( utils.IntToStr(len(n)) )
		//~ } else {
			//~ a := this.Interpreter.GetCode()
			//~ tl := types.NewTokenList()
			//~ tl.Push( types.NewToken(types.VARIABLE, countvar) )
			//~ tl.Push( types.NewToken(types.ASSIGNMENT, "=") )
			//~ tl.Push( types.NewToken(types.NUMBER, utils.IntToStr(-1)))
			//~ this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, &a, this.Interpreter.GetPC())
			//~ //log.Println(tl.AsString())
		//~ }

	//~ }

	return nil
}

func (this *PlusGetMsgs) Syntax() string {

	/* vars */
	var result string

	result = "GetMessages{namevar,msgvar,max}"

	/* enforce non void return */
	return result

}

func (this *PlusGetMsgs) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusGetMsgs(a int, b int, params types.TokenList) *PlusGetMsgs {
	this := &PlusGetMsgs{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "GETMESSAGES"

	this.NamedParams = []string{ "namevar", "msgvar", "countvar", "max" }
	this.NamedDefaults = []types.Token{
		*types.NewToken( types.STRING, "NA$" ),
		*types.NewToken( types.STRING, "MG$" ),
		*types.NewToken( types.STRING, "MC" ),
		*types.NewToken( types.INTEGER, "1" ),
	}
	this.Raw = true

	return this
}
