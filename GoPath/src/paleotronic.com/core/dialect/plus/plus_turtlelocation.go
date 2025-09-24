package plus

import (
	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusTurtleLocation struct {
	dialect.CoreFunction
}

func (this *PlusTurtleLocation) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	xt := this.ValueMap["x"]
	x := xt.AsExtended()
	yt := this.ValueMap["y"]
	y := yt.AsExtended()
	zt := this.ValueMap["z"]
	z := zt.AsExtended()

	if this.Query {
		switch this.QueryVar {
		case "x":
			v := apple2helpers.VECTOR(this.Interpreter).Turtle().Position[0]
			this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))
		case "y":
			v := apple2helpers.VECTOR(this.Interpreter).Turtle().Position[1]
			this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))
		case "z":
			v := apple2helpers.VECTOR(this.Interpreter).Turtle().Position[2]
			this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))
		}
	} else {

		apple2helpers.VECTOR(this.Interpreter).Turtle().AddCommandV(types.TURTLE_SPOS, &mgl64.Vec3{x, y, z})
		apple2helpers.VECTOR(this.Interpreter).Render()

	}

	return nil
}

func (this *PlusTurtleLocation) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusTurtleLocation) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusTurtleLocation(a int, b int, params types.TokenList) *PlusTurtleLocation {
	this := &PlusTurtleLocation{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"x", "y", "z"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}
