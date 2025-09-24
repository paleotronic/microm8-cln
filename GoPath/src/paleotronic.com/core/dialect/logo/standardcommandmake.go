package logo

import (
	"fmt"
	"strings"

	//	"errors"
	//	"paleotronic.com/fmt"

	"paleotronic.com/log"

	"paleotronic.com/core/dialect"

	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandMAKE struct {
	dialect.Command
	Split bool
	Local bool
}

func (this *StandardCommandMAKE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	d := this.Command.D.(*DialectLogo)

	groups, err := d.Driver.ParseExprRLCollapse(tokens.Copy(), false)
	if err != nil {
		return result, err
	}
	if len(groups) > 2 {
		return result, fmt.Errorf("i don't know what to do with %s", groups[2].Content)
	}

	if len(groups) < 2 {
		return result, fmt.Errorf("not enough inputs to make")
	}

	name := strings.Trim(groups[0].Content, ":\"")

	log.Printf("*** {%s}: set var [%s] to [%s]", tlistStr("", tokens.Copy()), name, groups[1].Content)

	if v, s := d.Driver.GetVar(name); v != nil && s != nil {
		// set an existing var at the same scope we find it
		s.Vars.Set(name, groups[1])
	} else {
		// if not found we create a global var
		d.Driver.Globals.Set(name, groups[1])
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandMAKE) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
