package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
)

type PlusZoneInit struct {
	dialect.CoreFunction
	Fill bool
}

func (this *PlusZoneInit) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	apple2helpers.InitZones(
		this.Interpreter,
	)

	return nil
}

func (this *PlusZoneInit) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusZoneInit) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusZoneInit(a int, b int, params types.TokenList) *PlusZoneInit {
	this := &PlusZoneInit{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{}
	this.NamedDefaults = []types.Token{}
	this.Raw = true

	return this
}
