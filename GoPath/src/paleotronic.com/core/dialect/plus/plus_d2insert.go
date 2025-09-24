package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusDiskIIInsert struct {
	dialect.CoreFunction
}

func (this *PlusDiskIIInsert) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slotid := this.Stack.Shift().AsInteger()
	volume := this.Stack.Shift().Content

	servicebus.SendServiceBusMessage(
		this.Interpreter.GetMemIndex(),
		servicebus.DiskIIInsertFilename,
		servicebus.DiskTargetString{
			Drive:    slotid,
			Filename: volume,
		},
	)

	//if hardware.DiskInsert(this.Interpreter, slotid, volume, false) {
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
	//} else {
	//	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))
	//}

	return nil
}

func (this *PlusDiskIIInsert) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusDiskIIInsert) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusDiskIIInsert(a int, b int, params types.TokenList) *PlusDiskIIInsert {
	this := &PlusDiskIIInsert{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 2
	this.MaxParams = 2

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "drive", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "volume", Default: *types.NewToken(types.STRING, "")},
		},
	)

	return this
}
