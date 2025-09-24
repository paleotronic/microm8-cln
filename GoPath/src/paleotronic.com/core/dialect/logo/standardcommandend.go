package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
)

type StandardCommandEND struct {
	dialect.Command
}

func (this *StandardCommandEND) Syntax() string {

	/* vars */
	var result string

	result = "RETURN"

	/* enforce non void return */
	return result

}

func (this *StandardCommandEND) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	d := this.Command.D.(*DialectLogo)

	if d.Driver.PendingProcName != "" {
		_, err := d.Driver.StoreProc(
			d.Lexer,
			d.Driver.PendingProcName,
			d.Driver.PendingProcArgs,
			d.Driver.PendingProcStatements,
		)
		if err == nil && !settings.LogoSuppressDefines[caller.GetMemIndex()] {
			caller.PutStr("DEFINED " + d.Driver.PendingProcName + "\r\n")
		}
		d.Driver.PendingProcName = ""
		caller.SetPrompt(d.OldPrompt)
		if !d.Driver.NoResolveProcEntry {
			d.Driver.ReresolveSymbols(d.Lexer)
		}
		return result, err
	}

	if d.OldPrompt != "" {
		caller.SetPrompt(d.OldPrompt)
	}

	/* enforce non void return */
	return result, nil

}
