package logo

import (
	"errors"
	"fmt"
	"strings" //	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	//"errors"
)

type StandardCommandLOAD struct {
	dialect.Command
}

func (this *StandardCommandLOAD) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	tokens, e := this.Expect(caller, tokens, 1)
	if e != nil {
		return result, e
	}

	if tokens.Size() < 1 {
		return result, errors.New("Need filename")
	}

	// Got a filename
	filename := tokens.Get(0).Content

	if !strings.HasPrefix(filename, "/") && caller.GetWorkDir() != "" {
		filename = "/" + strings.Trim(caller.GetWorkDir(), "/") + "/" + filename
	}

	if !strings.HasSuffix(filename, ".lgo") {
		filename += ".lgo"
	}

	data, e := files.ReadBytesViaProvider(files.GetPath(strings.ToLower(filename)), files.GetFilename(strings.ToLower(filename)))
	if e != nil {
		return result, e
	}

	lines := strings.Split(string(data.Content), "\r\n")
	// for _, l := range lines {
	// 	if l != "" {
	// 		tl := caller.GetDialect().Tokenize(runestring.Cast(l))
	// 		scope := caller.GetDirectAlgorithm()
	// 		caller.GetDialect().SetSilentDefines(true)
	// 		caller.GetDialect().ExecuteDirectCommand(*tl, caller, scope, caller.GetLPC())
	// 		caller.GetDialect().SetSilentDefines(false)
	// 		caller.SaveCPOS()
	// 		caller.SetNeedsPrompt(true)
	// 	}
	// }
	d := this.Command.D.(*DialectLogo)
	d.Driver.NoResolveProcEntry = true
	for i, l := range lines {
		if l != "" {
			d.SuppressError = true
			err := d.Parse(caller, l)
			d.SuppressError = false
			if err != nil {
				d.Driver.NoResolveProcEntry = false
				err = fmt.Errorf(
					"%s\r\n%v, line %d",
					strings.Trim(l, "\r\n"),
					err.Error(),
					i,
				)
				return result, err
			}
		}
	}
	d.Driver.NoResolveProcEntry = false
	d.Driver.ReresolveSymbols(d.Lexer)

	caller.SaveCPOS()
	caller.SetNeedsPrompt(true)

	/* enforce non void return */
	return result, e

}

func (this *StandardCommandLOAD) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
