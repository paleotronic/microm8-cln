package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusSendMsg struct {
	dialect.CoreFunction
}

func (this *PlusSendMsg) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	//~ if !this.Query {
		//~ msg := this.ValueMap["message"].Content

		//~ if msg != "" {
			//~ if rune(msg[0]) == '/' {
				//~ this.Interpreter.SendRemIntMessage("SCM", []byte(msg[1:]), true)
			//~ } else if rune(msg[0]) == '@' {
				//~ this.Interpreter.SendRemIntMessage("CAM", []byte(msg[1:]), true)
			//~ } else {
				//~ this.Interpreter.SendChatMessage(msg)
			//~ }
		//~ }

	//~ }

	//~ this.Stack.Push(types.NewToken(types.STRING, backdrop))

	return nil
}

func (this *PlusSendMsg) Syntax() string {

	/* vars */
	var result string

	result = "SendMessage{msg}"

	/* enforce non void return */
	return result

}

func (this *PlusSendMsg) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusSendMsg(a int, b int, params types.TokenList) *PlusSendMsg {
	this := &PlusSendMsg{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SENDMESSAGE"

	this.NamedParams = []string{ "message" }
	this.NamedDefaults = []types.Token{ *types.NewToken( types.STRING, "" ) }
	this.Raw = true

	return this
}
