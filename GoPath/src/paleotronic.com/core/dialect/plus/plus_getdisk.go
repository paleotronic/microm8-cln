package plus

import (
	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type PlusGetDisk struct {
	dialect.CoreFunction
}

func (this *PlusGetDisk) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	fmt.Printf("valuemap: %+v\n", this.ValueMap)

	drivet := this.ValueMap["drive"]
	drive := drivet.AsInteger()
	fmt.Printf("Drive #%d\n", drive)
	switch drive {
	case 0:
		this.Stack.Push(types.NewToken(types.STRING, settings.PureBootVolume[this.Interpreter.GetMemIndex()]))
	case 1:
		this.Stack.Push(types.NewToken(types.STRING, settings.PureBootVolume2[this.Interpreter.GetMemIndex()]))
	case 2:
		this.Stack.Push(types.NewToken(types.STRING, settings.PureBootSmartVolume[this.Interpreter.GetMemIndex()]))
	}

	return nil

}

func (this *PlusGetDisk) Syntax() string {

	/* vars */
	var result string

	result = "ALTCASE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusGetDisk) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusGetDisk(a int, b int, params types.TokenList) *PlusGetDisk {
	this := &PlusGetDisk{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"
	this.MinParams = 1
	this.MaxParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "drive", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}
