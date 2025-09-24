package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types" //	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/fmt"
)

type StandardCommandCOLORS struct {
	dialect.Command
}

func (this *StandardCommandCOLORS) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	txt := apple2helpers.TEXT(caller)
	ofg, obg := txt.FGColor, txt.BGColor
	txt.PutStr("Cols: [")
	for i := 0; i < 16; i++ {
		label := fmt.Sprintf("%-2d", i)
		txt.BGColor = uint64(i)
		txt.FGColor = 15
		if i >= 10 {
			txt.FGColor = 0
		}
		txt.PutStr(label)
	}
	txt.FGColor = ofg
	txt.BGColor = obg
	txt.PutStr("]\r\n")

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandCOLORS) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
