package logo

import (
	"log"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandOUTPUT struct {
	dialect.Command
}

func (this *StandardCommandOUTPUT) Syntax() string {

	/* vars */
	var result string

	result = "OUTPUT"

	/* enforce non void return */
	return result

}

func (this *StandardCommandOUTPUT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	d := this.Command.D.(*DialectLogo)

	rtok, e := d.Driver.ParseExprRLCollapse(tokens.Copy(), false)
	if len(rtok) == 0 {
		log.Printf("output: rtok is empty!")
	}
	if e != nil {
		return result, e
	}

	d.Driver.ReturnFromProc(rtok[0])

	/* enforce non void return */
	return result, e

}
