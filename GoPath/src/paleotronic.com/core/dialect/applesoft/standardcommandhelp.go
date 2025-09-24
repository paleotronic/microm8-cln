package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/control"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
)

type StandardCommandHELP struct {
	dialect.Command
	edit *editor.CoreEdit
}

func HelpExit(this *editor.CoreEdit) {

	this.Running = false
}

func (this *StandardCommandHELP) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0
	base := "dialect"
	dialect := caller.GetDialect().GetShortName()
	command := ""
	helpfile := base + "/" + dialect

	if tokens.Size() > 0 {
		for tokens.Size() > 0 {
			tmp := tokens.Shift().Content
			if command != "" {
				command += " "
			}
			command += tmp
		}
		helpfile += "/" + command
	}

	h := control.NewHelpController(caller, caller.GetDialect().GetLongName()+" Help", settings.HelpBase, helpfile)
	h.Do(caller)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandHELP) Syntax() string {

	/* vars */
	var result string

	result = "HELP"

	/* enforce non void return */
	return result

}
