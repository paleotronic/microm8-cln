package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
)

type PlusZoneReset struct {
	dialect.CoreFunction
	Fill bool
}

func (this *PlusZoneReset) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	tmp := this.ValueMap["zone"]
	zone := tmp.AsInteger()

	apple2helpers.ResetZonePalette(
		this.Interpreter,
		zone,
	)

	return nil
}

func (this *PlusZoneReset) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusZoneReset) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusZoneReset(a int, b int, params types.TokenList) *PlusZoneReset {
	this := &PlusZoneReset{}

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
