package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	"paleotronic.com/fmt"
)

type PlusRecordSlice struct {
	dialect.CoreFunction
}

func (this *PlusRecordSlice) FunctionExecute(params *types.TokenList) error {

	fmt.Println("PLAY")

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	fmt.Println("Params OK")

//	backend.REBOOT_NEEDED = true
	fn := this.ValueMap["file"].Content
	newfn := this.ValueMap["outfile"].Content
	sms := this.ValueMap["start"]; startms := sms.AsInteger()
	ems := this.ValueMap["end"]; endms := ems.AsInteger()
	e := this.Interpreter.SliceRecording( fn, newfn, startms, endms )

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return e
}

func (this *PlusRecordSlice) Syntax() string {

	/* vars */
	var result string

	result = "REBOOT{}"

	/* enforce non void return */
	return result

}

func (this *PlusRecordSlice) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusRecordSlice(a int, b int, params types.TokenList) *PlusRecordSlice {
	this := &PlusRecordSlice{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "REBOOT"
	this.Raw = true
	this.MinParams = 1
	this.MaxParams = 4
	this.NamedParams = []string{"file", "outfile", "start", "end"}
	this.NamedDefaults = []types.Token{ 
		*types.NewToken(types.STRING, ""),
		*types.NewToken(types.STRING, ""),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "86400"),
	}

	return this
}
