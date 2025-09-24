package plus

import (
	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusTurtleHeading struct {
	dialect.CoreFunction
}

func (this *PlusTurtleHeading) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	xt := this.ValueMap["pitch"]
	x := xt.AsExtended()
	yt := this.ValueMap["heading"]
	y := yt.AsExtended()
	zt := this.ValueMap["roll"]
	z := zt.AsExtended()

	if this.Query {
		switch this.QueryVar {
		case "pitch":
			v := apple2helpers.VECTOR(this.Interpreter).Turtle().Pitch
			this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))
		case "heading":
			v := apple2helpers.VECTOR(this.Interpreter).Turtle().Heading
			this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))
		case "roll":
			v := apple2helpers.VECTOR(this.Interpreter).Turtle().Roll
			this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))
		}
	} else {

		apple2helpers.VECTOR(this.Interpreter).Turtle().AddCommandV(types.TURTLE_SPHR, &mgl64.Vec3{x, y, z})
		apple2helpers.VECTOR(this.Interpreter).Render()

	}

	return nil
}

func (this *PlusTurtleHeading) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusTurtleHeading) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusTurtleHeading(a int, b int, params types.TokenList) *PlusTurtleHeading {
	this := &PlusTurtleHeading{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"pitch", "heading", "roll"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}
