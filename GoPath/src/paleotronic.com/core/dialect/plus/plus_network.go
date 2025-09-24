package plus

import (
	"paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusNetworkOK struct {
	dialect.CoreFunction
}

func (this *PlusNetworkOK) FunctionExecute(params *types.TokenList) error {

	//if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	c := 0
	if s8webclient.CONN.IsConnected() {
		c = 1
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(c)))

	return nil
}

func (this *PlusNetworkOK) Syntax() string {

	/* vars */
	var result string

	result = "BOOTTIME{v}"

	/* enforce non void return */
	return result

}

func (this *PlusNetworkOK) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	//	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusNetworkOK(a int, b int, params types.TokenList) *PlusNetworkOK {
	this := &PlusNetworkOK{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "UPTIME"

	return this
}
