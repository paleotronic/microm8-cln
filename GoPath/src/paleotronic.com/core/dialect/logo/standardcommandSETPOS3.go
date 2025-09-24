package logo

import (
	"errors"

	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandSETPOS3 struct {
	dialect.Command
}

func (this *StandardCommandSETPOS3) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	vv, e := this.Command.D.ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	if vv == nil || vv.Type != types.LIST || vv.List.Size() != 3 {
		return result, errors.New("I NEED 3 VALUES")
	}

	list := vv.List

	x := list.Get(0).AsExtended()
	y := list.Get(1).AsExtended()
	z := list.Get(2).AsExtended()

	/* enforce non void return */
	apple2helpers.VECTOR(caller).GetTurtle(this.Command.D.(*DialectLogo).Driver.GetTurtle()).AddCommandV(types.TURTLE_SPOS, &mgl64.Vec3{x, y, z})
	apple2helpers.VECTOR(caller).Render()

	return result, nil

}

func (this *StandardCommandSETPOS3) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
