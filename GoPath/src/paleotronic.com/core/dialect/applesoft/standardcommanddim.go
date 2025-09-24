package applesoft

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardCommandDIM struct {
	dialect.Command
}

func (this *StandardCommandDIM) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

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
	//var s string

	result = 0

	if (tokens.Size() < 1) || (tokens.LPeek().Type != types.VARIABLE) {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	tla = caller.SplitOnTokenWithBrackets(tokens, *types.NewToken(types.SEPARATOR, ","))

	dld = make([]int, 1)
	dld[0] = 10

	for _, tl := range tla {

		v = *tl.Shift()
		if v.Type != types.VARIABLE {
			return result, exception.NewESyntaxError("DIM expects variable")
		}

		if tl.Size() == 0 {
			dln = make([]int, len(dld))
			dln = dld
		} else {
			/* get dimensions */
			//caller.PutStr(caller.TokenListAsString(tl)+"

			dln, _ = caller.IndicesFromTokens(tl, "(", ")")
			if len(dln) == 0 {
				return result, exception.NewESyntaxError("Invalid indices for array")
			}
		}

		/* offset by 1 */
//		for i = 0; i <= len(dln)-1; i++ {
//			dln[i] = dln[i] + 1
//		}


		//s = "" /* default value */

		/* now check if (it exists */
		if vv = caller.GetLocal().Get(v.Content); vv != nil {

			if vv.IsArray() {
				eq := utils.IntSliceEq(dln, vv.Dimensions())
				if !eq {
					return result, exception.NewESyntaxError("redeclared array")
				}
				return 0, nil
			}
			/* not an array - let's redim it */
			//s, _ = vv.GetContentScalar()
			// removed free call here;
		}

		/* okay - lets create it */
		vt := types.VT_FLOAT

		ch = rune(v.Content[len(v.Content)-1])

		switch ch { /* FIXME - Switch statement needs cleanup */
		case '%':
			vt = types.VT_INTEGER
			break
		case '#':
			vt = types.VT_FLOAT
			break
		case '!':
			vt = types.VT_FLOAT
			break
		case '$':
			vt = types.VT_STRING
			break
		case '@':
			vt = types.VT_BOOLEAN
			break
		}

		ns := ""
		if vt == types.VT_STRING {
			ns = ""
		} else {
			ns = "0"
		}

		vv, e := types.NewVariablePA(caller.GetLocal(), strings.ToLower(v.Content), vt, ns, true, dln)
		if e != nil {
           return 0, e
        }
    	//vv.SetContentScalar(s)
		vv.Owner = caller.GetName()
		//caller.CreateVarLower(strings.ToLower(v.Content), vv)

	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandDIM) Syntax() string {

	/* vars */
	var result string

	result = "DIM"

	/* enforce non void return */
	return result

}
