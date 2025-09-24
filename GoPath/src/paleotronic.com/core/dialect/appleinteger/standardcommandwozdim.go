package appleinteger

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"strings"
)

type StandardCommandWozDIM struct {
	dialect.Command
}

func (this *StandardCommandWozDIM) Syntax() string {

	/* vars */
	var result string

	result = "DIM"

	/* enforce non void return */
	return result

}

func (this *StandardCommandWozDIM) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var v types.Token
	//TokenList tl;
	var tla types.TokenListArray
	var dld []int
	var dln []int
	var vv *types.Variable
	var ch rune
	//	var i int
	var s string

	result = 0

	if (tokens.Size() < 1) || ((tokens.LPeek().Type != types.VARIABLE) && (tokens.LPeek().Type != types.KEYWORD)) {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	tla = caller.SplitOnTokenWithBrackets(tokens, *types.NewToken(types.SEPARATOR, ","))

	dld = make([]int, 1)
	dld[0] = 10

	for _, tl := range tla {

		v = *tl.Shift()
		if (v.Type != types.VARIABLE) && (v.Type != types.KEYWORD) {
			return result, exception.NewESyntaxError("DIM expects variable")
		}

		if tl.Size() == 0 {
			//SetLength(dln, dld.Size());
			dln = make([]int, len(dld))
			dln = dld
		} else {
			/* get dimensions */
			//caller.GetVDU().PutStr(caller.TokenListAsString(tl)+PasUtil.CRLF);
			dln, _ = caller.IndicesFromTokens(tl, "(", ")")
			if len(dln) == 0 {
				return result, exception.NewESyntaxError("Invalid indices for array")
			}
		}

		/* offset by 1 */
//		for i := 0; i <= len(dln)-1; i++ {
//			dln[i] = dln[i] + 1
//		}


		s = "" /* default value */

		/* now check if (it exists */
		if caller.GetLocal().Exists(strings.ToLower(v.Content)) {

			vv = caller.GetLocal().Get(strings.ToLower(v.Content))

			if vv.IsArray() {
				return result, exception.NewESyntaxError("DIM ERR")
			}

			/* not an array - let's redim it */
			s, _ = vv.GetContentScalar()
			// removed free call here;
		}

		/* okay - lets create it */
		vt := types.VT_INTEGER

		ch = rune(v.Content[len(v.Content)-1])

		switch ch { /* FIXME - Switch statement needs cleanup */
		case '%':
			vt = types.VT_INTEGER
			break
		case '$':
			vt = types.VT_STRING
			break
		case '@':
			vt = types.VT_BOOLEAN
			break
		}

		if vt != types.VT_STRING {
			vv, _ = types.NewVariablePA(caller.GetLocal(), strings.ToLower(v.Content), vt, s, true, dln)
		} else {
			//vv, _ = types.NewVariableP(caller.GetLocal(), strings.ToLower(v.Content), vt, s, true)
            //dln[0] = dln[0] - 1
            ee := caller.GetLocal().CreateIndexed(
                strings.ToLower(v.Content), 
                types.VT_STRING, 
                dln, 
                s)
            if ee != nil {
               return 0, ee
            }
		}

		//vv.Owner = caller.GetName()
		//caller.CreateVarLower(strings.ToLower(v.Content), *vv)

	}

	/* enforce non void return */
	return result, nil

}
