package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
)

type PlusZoneOffset struct {
	dialect.CoreFunction
	Fill bool
}

func (this *PlusZoneOffset) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	tmp := this.ValueMap["o"]
	o := tmp.AsInteger()
	tmp = this.ValueMap["c"]
	c := tmp.AsInteger()
	tmp = this.ValueMap["zone"]
	zone := tmp.AsInteger()

	apple2helpers.SetZonePaletteOffset(
		this.Interpreter,
		zone,
		c,
		int8(o),
	)

	return nil
}

func (this *PlusZoneOffset) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusZoneOffset) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusZoneOffset(a int, b int, params types.TokenList) *PlusZoneOffset {
	this := &PlusZoneOffset{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"zone", "c", "o"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}
