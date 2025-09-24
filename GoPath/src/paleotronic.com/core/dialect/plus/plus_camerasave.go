package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	//	"paleotronic.com/utils"
	//	"time"
	"strings"
	//	"paleotronic.com/core/settings"
	"paleotronic.com/fmt"

	"github.com/atotto/clipboard"
)

type PlusCameraSave struct {
	dialect.CoreFunction
}

func (this *PlusCameraSave) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if params.Size() == 0 {

		index := this.Interpreter.GetMemIndex()
		mm := this.Interpreter.GetMemoryMap()
		cindex := mm.GetCameraConfigure(index)
		control := types.NewOrbitController(mm, index, cindex)

		s := control.String()

		s = strings.Replace(s, "\"", "'", -1)

		this.Stack.Push(types.NewToken(types.STRING, s[1:len(s)-1]))

		// put in clipboard
		clipboard.WriteAll(s[1 : len(s)-1])

	} else {

		s := "{" + params.Shift().Content + "}"

		s = strings.Replace(s, "'", "\"", -1)

		index := this.Interpreter.GetMemIndex()
		mm := this.Interpreter.GetMemoryMap()
		cindex := mm.GetCameraConfigure(index)
		control := types.NewOrbitController(mm, index, cindex)

		fmt.Println("JSON:", s)

		control.FromString(s)

		this.Stack.Push(types.NewToken(types.NUMBER, "0"))

	}

	return nil
}

func (this *PlusCameraSave) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusCameraSave) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusCameraSave(a int, b int, params types.TokenList) *PlusCameraSave {
	this := &PlusCameraSave{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 0
	this.MaxParams = 1

	return this
}
