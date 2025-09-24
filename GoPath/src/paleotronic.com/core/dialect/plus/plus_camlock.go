package plus

import (
	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/memory"
	"strings"

	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusCamLock struct {
	dialect.CoreFunction
}

func (this *PlusCamLock) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	param := strings.ToUpper(params.Shift().Content)

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()
	cindex := mm.GetCameraConfigure(index)
	control := types.NewOrbitController(mm, index, cindex)

	switch {
	case param == "TRUE":
		control.SetLockView(true)
	case param == "FALSE":
		control.SetLockView(false)
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusCamLock) Syntax() string {

	/* vars */
	var result string

	result = "CAMLOCK{}"

	/* enforce non void return */
	return result

}

func (this *PlusCamLock) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusCamLock(a int, b int, params types.TokenList) *PlusCamLock {
	this := &PlusCamLock{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMLOCK"

	return this
}
