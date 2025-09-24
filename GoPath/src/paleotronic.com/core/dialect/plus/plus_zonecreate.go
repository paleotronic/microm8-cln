package plus

import (
	"log"
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
)

type PlusZoneCreate struct {
	dialect.CoreFunction
	Fill bool
}

func (this *PlusZoneCreate) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		log.Printf("err: %v", e)
		return e
	}

	tmp := this.ValueMap["x0"]
	x0 := tmp.AsInteger()
	tmp = this.ValueMap["y0"]
	y0 := tmp.AsInteger()
	tmp = this.ValueMap["x1"]
	x1 := tmp.AsInteger()
	tmp = this.ValueMap["y1"]
	y1 := tmp.AsInteger()
	tmp = this.ValueMap["zone"]
	zone := tmp.AsInteger()

	zc := apple2helpers.GetZoneConfig(this.Interpreter)
	if zc == nil {
		apple2helpers.CreateZoneConfig(this.Interpreter)
	}
	apple2helpers.SetZone(
		this.Interpreter,
		zone,
		x0, y0, x1, y1,
	)

	return nil
}

func (this *PlusZoneCreate) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusZoneCreate) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusZoneCreate(a int, b int, params types.TokenList) *PlusZoneCreate {
	this := &PlusZoneCreate{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"zone", "x0", "y0", "x1", "y1"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}
