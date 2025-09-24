package logo

import (
	//	"strings"
//	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
//	"paleotronic.com/core/interfaces"
//	"paleotronic.com/core/hardware/apple2helpers"
//	"paleotronic.com/runestring"
	"paleotronic.com/utils"
	"paleotronic.com/files"
)

type StandardFunctionREADPOS struct {
	dialect.CoreFunction
}

func (this *StandardFunctionREADPOS) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionREADPOS(a int, b int, params types.TokenList) *StandardFunctionREADPOS {
	this := &StandardFunctionREADPOS{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "READPOS"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionREADPOS) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	p, e := files.DOSREADPOS( files.GetPath(files.Reader), files.GetFilename(files.Reader) )

	this.Stack.Push( types.NewToken( types.NUMBER, utils.FormatFloat("", float64(p)) ) )

	return e
}

func (this *StandardFunctionREADPOS) Syntax() string {

	/* vars */
	var result string

	result = "READPOS word list"

	/* enforce non void return */
	return result

}
