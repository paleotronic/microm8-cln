package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types" //	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandCAMBEHIND struct {
	dialect.Command
}

func (this *StandardCommandCAMBEHIND) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	// v, err := this.Command.D.ParseTokensForResult(caller, tokens)
	// if err != nil {
	// 	return result, err
	// }
	// if v == nil || v.Type == types.LIST {
	// 	return result, errors.New("I NEED A VALUE")
	// }

	d := this.Command.D.(*DialectLogo)
	// of, ob := d.Driver.Tracking.FollowPosition, d.Driver.Tracking.FollowBehind
	// d.Driver.Tracking.FollowPosition = true
	// d.Driver.Tracking.FollowBehind = true
	d.Driver.TrackBehind()
	// d.Driver.Tracking.FollowPosition = of
	// d.Driver.Tracking.FollowBehind = ob

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandCAMBEHIND) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
