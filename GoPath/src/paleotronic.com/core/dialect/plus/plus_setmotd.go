package plus

import (
//	"paleotronic.com/log"
	//"errors"
//	"strings"

	//"paleotronic.com/files"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	"paleotronic.com/api"
	//"paleotronic.com/core/interfaces"
)

type PlusSetMOTD struct {
	dialect.CoreFunction
}

// params: 
// (1) hostname
// (2) name

func (this *PlusSetMOTD) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { 
		return e 
	}

	motdtext := params.Shift().Content
	
	_ = s8webclient.CONN.AddMOTD( motdtext )

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusSetMOTD) Syntax() string {

	/* vars */
	var result string

	result = "CONTROL{slot, target}"

	/* enforce non void return */
	return result

}

func (this *PlusSetMOTD) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusSetMOTD(a int, b int, params types.TokenList) *PlusSetMOTD {
	this := &PlusSetMOTD{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CONTROL"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}
