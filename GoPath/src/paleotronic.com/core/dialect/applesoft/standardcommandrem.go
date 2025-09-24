package applesoft

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
)

const REM_PRAGMA = "#"
const REM_LABEL = "!"

type StandardCommandREM struct {
	dialect.Command
}

func NewStandardCommandREM() *StandardCommandREM {
	this := &StandardCommandREM{}
	this.NoTokens = true
	this.Cost = 1
	return this
}

func (this *StandardCommandREM) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandREM) Syntax() string {

	/* vars */
	var result string

	result = "REM comments"

	/* enforce non void return */
	return result

}

// Process any special fields
func (this *StandardCommandREM) BeforeRun(caller interfaces.Interpretable) {
	fmt.Println("REM.BeforeRun() being called")

	code := caller.GetCode()

	caller.ClearLabels()

	for lno, ln := range code.C {
		for _, st := range ln {
			t := st.LPeek()
			if t.Type == types.KEYWORD && strings.ToLower(t.Content) == "rem" {
				// Got a rem statement...
				if st.Size() < 2 {
					continue
				}

				s := st.Get(1).Content // get rem text
				if strings.HasPrefix(s, REM_LABEL) {
					// handle label
					parts := strings.Split(s[1:], " ")
					label := strings.ToLower(parts[0])
					caller.SetLabel(label, lno)

					fmt.Printf("Setup label [%s] at line %d...\n", label, lno)

				} else if strings.HasPrefix(s, REM_PRAGMA) {
					// handle pragma
					parts := strings.Split(s[1:], " ")
					pragma := strings.ToLower(parts[0])
					caller.SetPragma(pragma)
				}
			}
		}
	}

}
