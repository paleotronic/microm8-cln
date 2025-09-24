package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
    "paleotronic.com/api"
    "paleotronic.com/fmt"
)

type PlusGetRemotes struct {
	dialect.CoreFunction
}

func (this *PlusGetRemotes) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {

		filter := this.ValueMap["filter"].Content
		listvar := this.ValueMap["listvar"].Content
		countvar := this.ValueMap["countvar"].Content
		max := utils.StrToInt(this.ValueMap["max"].Content)

		//log.Printf("Parameters: %s, %s, %d\n", namevar, msgvar, max)

		remotes, err := s8webclient.CONN.GetRemoteInstances(filter)
        if len(remotes) > max {
		   remotes = remotes[0:max]
        }

		if err == nil {

			for i := 0; i<len(remotes); i++ {
				//log.Printf("*** <%s> %s\n", n[i], m[i])
				rr := remotes[i]
                info := fmt.Sprintf("%s:%d:%d", rr.Host, rr.Port, rr.PID )

				tl := types.NewTokenList()
				tl.Push( types.NewToken(types.VARIABLE, listvar) )
				tl.Push( types.NewToken(types.OBRACKET, "(") )
				tl.Push( types.NewToken(types.NUMBER, utils.IntToStr(i+1)))
				tl.Push( types.NewToken(types.CBRACKET, ")") )
				tl.Push( types.NewToken(types.ASSIGNMENT, "=") )
				tl.Push( types.NewToken(types.STRING, info) )
				a := this.Interpreter.GetCode()
				this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, a, this.Interpreter.GetPC())
				//log.Println(tl.AsString())

				tl = types.NewTokenList()
				tl.Push( types.NewToken(types.VARIABLE, countvar) )
				tl.Push( types.NewToken(types.ASSIGNMENT, "=") )
				tl.Push( types.NewToken(types.NUMBER, utils.IntToStr(i+1)))
				this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, a, this.Interpreter.GetPC())

				//log.Println(tl.AsString())
			}

			//cvar.SetContentScalar( utils.IntToStr(len(n)) )
		} else {
			a := this.Interpreter.GetCode()
			tl := types.NewTokenList()
			tl.Push( types.NewToken(types.VARIABLE, countvar) )
			tl.Push( types.NewToken(types.ASSIGNMENT, "=") )
			tl.Push( types.NewToken(types.NUMBER, utils.IntToStr(-1)))
			this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, a, this.Interpreter.GetPC())
			//log.Println(tl.AsString())
		}

	}

	return nil
}

func (this *PlusGetRemotes) Syntax() string {

	/* vars */
	var result string

	result = "GetMessages{namevar,msgvar,max}"

	/* enforce non void return */
	return result

}

func (this *PlusGetRemotes) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusGetRemotes(a int, b int, params types.TokenList) *PlusGetRemotes {
	this := &PlusGetRemotes{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "GETREMOTES"

	this.NamedParams = []string{ "filter", "listvar", "countvar", "max" }
	this.NamedDefaults = []types.Token{
		*types.NewToken( types.STRING, "all" ),
		*types.NewToken( types.STRING, "RM$" ),
		*types.NewToken( types.STRING, "COUNT" ),
		*types.NewToken( types.INTEGER, "20" ),
	}
	this.Raw = true

	return this
}
