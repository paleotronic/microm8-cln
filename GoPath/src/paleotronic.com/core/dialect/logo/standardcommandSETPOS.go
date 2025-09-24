package logo

import (
	"errors"

	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
)

type StandardCommandSETPOS struct {
	dialect.Command
}

func (this *StandardCommandSETPOS) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	t, _ := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)

	if t == nil {
		return result, errors.New("I NEED A VALUE")
	}

	list := t

	fmt.Printf("%s token.list = %+v\n", list.Type, list.List)

	if list.Type != types.LIST || list.List.Size() < 2 {
		return result, errors.New("I NEED A LIST OF 2")
	}

	x := list.List.Get(0).AsExtended()
	y := list.List.Get(1).AsExtended()
	z := apple2helpers.VECTOR(caller).GetTurtle( this.Command.D.(*DialectLogo).Driver.GetTurtle() ).Position[2]
	if list.List.Size() > 2 {
		z = list.List.Get(2).AsExtended()
	}

	/* enforce non void return */
	apple2helpers.VECTOR(caller).GetTurtle( this.Command.D.(*DialectLogo).Driver.GetTurtle() ).AddCommandV(types.TURTLE_SPOS, &mgl64.Vec3{x, y, z})
	apple2helpers.VECTOR(caller).Render()

	return result, nil

}

func (this *StandardCommandSETPOS) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
