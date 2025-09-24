package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
)

type PlusZoneDepth struct {
	dialect.CoreFunction
	Fill bool
}

func (this *PlusZoneDepth) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	tmp := this.ValueMap["d"]
	d := tmp.AsInteger()
	tmp = this.ValueMap["c"]
	c := tmp.AsInteger()
	tmp = this.ValueMap["zone"]
	zone := tmp.AsInteger()

	apple2helpers.SetZonePaletteDepth(
		this.Interpreter,
		zone,
		c,
		uint8(d),
	)

	return nil
}

func (this *PlusZoneDepth) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusZoneDepth) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusZoneDepth(a int, b int, params types.TokenList) *PlusZoneDepth {
	this := &PlusZoneDepth{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"zone", "c", "d"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "20"),
	}
	this.Raw = true

	return this
}
