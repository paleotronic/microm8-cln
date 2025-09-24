package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	"paleotronic.com/fmt"
)

type PlusRecordPlay struct {
	dialect.CoreFunction
}

func (this *PlusRecordPlay) FunctionExecute(params *types.TokenList) error {

	fmt.Println("PLAY")

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	fmt.Println("Params OK")

//	backend.REBOOT_NEEDED = true
	fn := this.ValueMap["file"].Content
	fmt.Println("fn =",fn)
	this.Interpreter.PlayRecording( fn )

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusRecordPlay) Syntax() string {

	/* vars */
	var result string

	result = "REBOOT{}"

	/* enforce non void return */
	return result

}

func (this *PlusRecordPlay) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusRecordPlay(a int, b int, params types.TokenList) *PlusRecordPlay {
	this := &PlusRecordPlay{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "REBOOT"
	this.Raw = true
	this.MinParams = 1
	this.MaxParams = 1
	this.NamedParams = []string{"file"}
	this.NamedDefaults = []types.Token{ *types.NewToken(types.STRING, "") }

	return this
}
