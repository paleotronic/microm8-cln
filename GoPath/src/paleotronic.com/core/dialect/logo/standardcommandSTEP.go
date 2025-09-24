package logo

import (
	"errors"
	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types" //	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandSTEP struct {
	dialect.Command
}

const StepVarName = "__for_step__"

func (this *StandardCommandSTEP) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	v, err := this.Command.D.ParseTokensForResult(caller, tokens)
	if err != nil {
		return result, err
	}
	if v == nil || v.Type == types.LIST {
		return result, errors.New("I NEED A VALUE")
	}

	step := math.Abs(v.AsExtended())

	// Step sets stepping in current scope..
	d := this.Command.D.(*DialectLogo)

	// if strings.HasPrefix(d.Driver.S.ProcRef.Name, "__imm") {
	// 	d.Driver.Globals.Set(StepVarName, types.NewToken(types.NUMBER, utils.FloatToStr(step)))
	// } else {
	d.Driver.S.StmtIterStep = step
	// }

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandSTEP) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
