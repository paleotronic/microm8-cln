package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"strings"
)

type StandardCommandDECLARE struct {
	dialect.Command
}

func (this *StandardCommandDECLARE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var typetok types.Token
	var nametok types.Token
	//var tok types.Token
	var vtok types.Token
	//var front types.Token
	//var back types.Token
	//var dtok types.Token
	var tla types.TokenListArray
	var tl types.TokenList
	var dl []int

	result = 0

	/* handle var declaration */
	if tokens.Size() < 2 {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	typetok = *tokens.Shift()
	nametok = *tokens.Shift()

	if typetok.Type != types.TYPE {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	if nametok.Type != types.VARIABLE {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	vtok = *types.NewToken(types.STRING, "")

	dl = make([]int, 0)

	if tokens.Size() > 0 {

		// split on '=' to see if (we have a value;
		tla = caller.SplitOnToken(tokens, *types.NewToken(types.ASSIGNMENT, ""))

		if tla.Size() > 1 {
			vt := caller.VariableTypeFromString(typetok.Content)
			//writeln("                                                ", vt);
			if vt != types.VT_EXPRESSION {
				vtok = caller.ParseTokensForResult(*tla.Get(tla.Size() - 1))
			} else {
				vtok.Content = caller.TokenListAsString(*tla.Get(tla.Size() - 1))
				//env.VDU.PutStr("[\"+vtok.Content+\"]");
			}
		}

		tl = *tla.Get(0)

		if tl.Size() > 0 {
			// might be dimensions;
			var err error
			dl, err = caller.IndicesFromTokens(tl, "(", ")")
			if err != nil {
				return result, err
			}
		}

	}

	/* if (we get here try create var */
	/* does it exist? */
	if caller.ExistsVar(strings.ToLower(nametok.Content)) {
		return result, exception.NewESyntaxError("REDECLARED VARIABLE " + nametok.Content)
	}

	if len(dl) == 0 {
		/* create scalar */
		vt := caller.VariableTypeFromString(typetok.Content)
		//env.Global.Get(strings.ToLower(nametok.Content)) = types.NewVariable(strings.ToLower(nametok.Content), vt, vtok.Content, true);
		//env.Global.Get(strings.ToLower(nametok.Content)).Owner = caller.Name;
		v, _ := types.NewVariableP(caller.GetLocal(), strings.ToLower(nametok.Content), vt, vtok.Content, true)
		v.Owner = caller.GetName()
	} else {
		/* create array */
		vt := caller.VariableTypeFromString(typetok.Content)
		//env.Global.Get(strings.ToLower(nametok.Content)) = types.NewVariable()Array(strings.ToLower(nametok.Content), vt, vtok.Content, true, dl);
		//env.Global.Get(strings.ToLower(nametok.Content)).Owner = caller.Name;
		v, _ := types.NewVariablePA(caller.GetLocal(), strings.ToLower(nametok.Content), vt, vtok.Content, true, dl)

		//caller.GetLocal().CreateIndexed(nametok.Content, vt, dl, )

		v.Owner = caller.GetName()
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandDECLARE) Syntax() string {

	/* vars */
	var result string

	result = "DECLARE <type> <name>[(size)]"

	/* enforce non void return */
	return result

}
