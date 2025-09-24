package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
)

type PlusZoneRGBA struct {
	dialect.CoreFunction
	Fill bool
}

func (this *PlusZoneRGBA) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	tmp := this.ValueMap["r"]
	r := tmp.AsInteger()
	tmp = this.ValueMap["g"]
	g := tmp.AsInteger()
	tmp = this.ValueMap["b"]
	b := tmp.AsInteger()
	tmp = this.ValueMap["a"]
	a := tmp.AsInteger()
	tmp = this.ValueMap["c"]
	c := tmp.AsInteger()
	tmp = this.ValueMap["zone"]
	zone := tmp.AsInteger()

	apple2helpers.SetZonePaletteRGBA(
		this.Interpreter,
		zone,
		c,
		uint8(r),
		uint8(g),
		uint8(b),
		uint8(a),
	)

	return nil
}

func (this *PlusZoneRGBA) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusZoneRGBA) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusZoneRGBA(a int, b int, params types.TokenList) *PlusZoneRGBA {
	this := &PlusZoneRGBA{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"zone", "c", "r", "g", "b", "a"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "255"),
		*types.NewToken(types.NUMBER, "255"),
		*types.NewToken(types.NUMBER, "255"),
		*types.NewToken(types.NUMBER, "255"),
	}
	this.Raw = true

	return this
}
