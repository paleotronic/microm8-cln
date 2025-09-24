package logo

import (
	"errors"
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardCommandSHOW struct {
	dialect.Command
	PrintMode bool
}

func (this *StandardCommandSHOW) Syntax() string {

	/* vars */
	var result string

	result = "SHOW"

	/* enforce non void return */
	return result

}

func (this *StandardCommandSHOW) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	d := this.Command.D.(*DialectLogo)

	t, err := d.Driver.ParseExprRLCollapse(tokens.Copy(), false)
	if err != nil {
		return result, err
	}
	if t == nil {
		return result, errors.New("I NEED A VALUE")
	}

	for _, tt := range t {
		show(caller, tt, this.PrintMode)
		caller.PutStr(" ")
	}
	caller.PutStr("\r\n")

	/* enforce non void return */
	return result, nil

}

func show(caller interfaces.Interpretable, t *types.Token, printMode bool) {

	if t.Type == types.LIST || t.Type == types.TABLE {
		if !printMode {
			caller.PutStr("[")
		}
		for _, tt := range t.List.Content {
			// if i > 0 && !(tt.Type == types.WORD && len(tt.Content) > 1 && tt.Content[:1] == "\\") {
			// 	caller.PutStr(" ")
			// }
			show(caller, tt, printMode)
			// if i < t.List.Size()-1 {
			// 	caller.PutStr(" ")
			// }
		}
		if !printMode {
			caller.PutStr("]")
		}
	} else {

		//log.Printf("show for token [%s] (suffix: [%s])", t.Content, t.WSSuffix)

		if t.Type == types.WORD && !printMode {
			caller.PutStr("\"" + t.Content)
		} else if t.Type == types.WORD && printMode {
			if strings.HasPrefix(t.Content, "\\") {
				caller.PutStr(t.Content[1:])
			} else {
				caller.PutStr(t.Content)
			}
		} else if t.Type == types.NUMBER && printMode {
			caller.PutStr(
				utils.StrToFloatStrAppleLogo(t.Content),
			)
		} else {
			caller.PutStr(t.Content)
		}
		if t.WSSuffix != "" {
			caller.PutStr(t.WSSuffix)
		}
	}

}
