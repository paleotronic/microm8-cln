package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/octalyzer/ui/chat"
)

type PlusBuiltinChat struct {
	dialect.CoreFunction
}

func (this *PlusBuiltinChat) FunctionExecute(params *types.TokenList) error {

	//if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	//	backend.REBOOT_NEEDED = true
	startChan := ""
	if params.Size() > 0 {
		startChan = params.Get(0).Content
	}
	app := chat.NewChatClient(this.Interpreter, startChan)
	app.Run()

	return nil
}

func (this *PlusBuiltinChat) Syntax() string {

	/* vars */
	var result string

	result = "REBOOT{}"

	/* enforce non void return */
	return result

}

func (this *PlusBuiltinChat) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusBuiltinChat(a int, b int, params types.TokenList) *PlusBuiltinChat {
	this := &PlusBuiltinChat{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "REBOOT"
	this.MinParams = 0
	this.MaxParams = 1

	return this
}
