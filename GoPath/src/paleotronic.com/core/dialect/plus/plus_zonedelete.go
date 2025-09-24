package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
)

type PlusZoneDelete struct {
	dialect.CoreFunction
	Fill bool
}

func (this *PlusZoneDelete) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	tmp := this.ValueMap["zone"]
	zone := tmp.AsInteger()

	apple2helpers.DeleteZoneFromConfig(
		this.Interpreter,
		zone,
	)

	return nil
}

func (this *PlusZoneDelete) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusZoneDelete) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusZoneDelete(a int, b int, params types.TokenList) *PlusZoneDelete {
	this := &PlusZoneDelete{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"zone"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}
