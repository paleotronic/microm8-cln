package plus

import (
	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types" //	"paleotronic.com/core/memory"
	"paleotronic.com/utils"      //	"time"
	//	"paleotronic.com/core/settings"
)

type PlusTurtleCamera struct {
	dialect.CoreFunction
}

func (this *PlusTurtleCamera) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()
	cindex := mm.GetCameraConfigure(index)
	control := types.NewOrbitController(mm, index, cindex)

	if this.Query {

		mpos := mgl64.Vec3{0, 0, 0}
		vect := apple2helpers.VECTOR(this.Interpreter)
		vl := apple2helpers.GETGFX(this.Interpreter, "VCTR")
		t := vect.Turtle()
		tpos := mgl64.Vec3{
			t.Position[0],
			t.Position[1],
			t.Position[2],
		}

		glh := float64(types.CHEIGHT)
		glw := float64(types.CHEIGHT) * control.GetAspect()

		width := float64(vl.GetWidth())
		height := float64(vl.GetHeight())

		apos := utils.TurtleCoordinatesToGL(glw, glh, width, height, mpos, tpos)

		var v float64

		switch this.QueryVar {
		case "x":
			v = apos.X()
		case "y":
			v = apos.Y()
		case "z":
			v = apos.Z()
		}

		this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))

	}

	return nil
}

func (this *PlusTurtleCamera) Syntax() string {

	/* vars */
	var result string

	result = "ANG{ax,ay,az}"

	/* enforce non void return */
	return result

}

func (this *PlusTurtleCamera) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusTurtleCamera(a int, b int, params types.TokenList) *PlusTurtleCamera {
	this := &PlusTurtleCamera{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMANG"
	this.MinParams = 1
	this.MaxParams = 3
	this.Raw = true
	this.NamedParams = []string{"x", "y", "z"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.INTEGER, "0"),
		*types.NewToken(types.INTEGER, "0"),
		*types.NewToken(types.INTEGER, "0"),
	}
	this.EvalSingleParam = true

	return this
}
