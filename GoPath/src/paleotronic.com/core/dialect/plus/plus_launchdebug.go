package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusLaunchDebug struct {
	dialect.CoreFunction
}

var StateLoadFunc func(fn string) error

func (this *PlusLaunchDebug) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	fn := this.ValueMap["file"].Content

	if StateLoadFunc != nil {
		StateLoadFunc(fn)
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusLaunchDebug) Syntax() string {

	/* vars */
	var result string

	result = "REBOOT{}"

	/* enforce non void return */
	return result

}

func (this *PlusLaunchDebug) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusLaunchDebug(a int, b int, params types.TokenList) *PlusLaunchDebug {
	this := &PlusLaunchDebug{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "REBOOT"
	this.Raw = true
	this.MinParams = 1
	this.MaxParams = 1
	this.NamedParams = []string{"file"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.STRING, "")}

	return this
}
