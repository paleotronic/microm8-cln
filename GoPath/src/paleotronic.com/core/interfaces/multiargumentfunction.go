package interfaces

import (
	//	"paleotronic.com/fmt"
	"strings"

	"paleotronic.com/core/exception"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type MultiArgumentFunction struct {
	Arguments  types.TokenList
	Expression types.TokenList
}

func NewMultiArgumentFunction(arglist types.TokenList, expr types.TokenList) *MultiArgumentFunction {
	this := &MultiArgumentFunction{}
	this.Arguments = arglist
	this.Expression = expr
	return this
}

func (this *MultiArgumentFunction) Evaluate(ent Interpretable, args types.TokenList) (types.Token, error) {

	/* vars */
	var result types.Token
	var cr, opc types.CodeRef
	var scope *types.Algorithm
	var i int
	var vtok types.Token
	var ntok types.Token
	var rtok types.Token
	var ecopy types.TokenList

	result = *types.NewToken(types.INVALID, "")

	////fmt.Printf( "MAF called with args (%d): %v\n", args.Size(), args )

	if args.Size() != this.Arguments.Size() {
		return result, exception.NewESyntaxError("Function expected " + utils.IntToStr(this.Arguments.Size()) + " but got " + utils.IntToStr(args.Size()))
	}

	/* we cheat && create a Newstack frame from the calculation which allows us to create
	 * local vars with the specified names / values */

	if ent.GetState() == types.RUNNING {
		cr = *types.NewCodeRefCopy(*ent.GetPC())
		opc = *types.NewCodeRefCopy(*ent.GetPC())
		scope = ent.GetCode()
	} else {
		cr = *types.NewCodeRefCopy(*ent.GetLPC())
		opc = *types.NewCodeRefCopy(*ent.GetLPC())
		scope = ent.GetDirectAlgorithm()
	}

	/* Newstack frame*/
	//      procedure Call( target: CodeRef; c: algorithm; s: EntityState; iso: Boolean; prefix: String; stackstate: TokenList; dia: Dialect );
	ent.Call(cr, scope, ent.GetState(), false, "", *ent.GetTokenStack(), ent.GetDialect())

	ecopy = *this.Expression.SubList(0, this.Expression.Size())

	/* setup parameters */
	for i = 0; i <= args.Size()-1; i++ {
		vtok = *args.Get(i)
		ntok = *this.Arguments.Get(i)

		// replace instances of var ntok with value
		for i, t := range ecopy.Content {
			if t.Type == types.VARIABLE && strings.ToLower(t.Content) == strings.ToLower(ntok.Content) {
				s := vtok.Content
				if vtok.Type == types.VARIABLE {
					vv := ent.GetLocal().Get(vtok.Content)
					if vv == nil {
						s = "0"
					} else {
						s, _ = vv.GetContentScalar()
					}
				}
				ecopy.Content[i] = types.NewToken(types.NUMBER, s)
			}
		}

		/*

						vt := types.VT_STRING
						switch vtok.Type {
						case types.NUMBER:
							vt = types.VT_FLOAT
							break
						case types.INTEGER:
							vt = types.VT_FLOAT
							break
						case types.BOOLEAN:
							vt = types.VT_FLOAT
							break
						}

			            // Create arg if does not exist
			            var e error
			            argvar := ent.GetLocal().Get( ntok.Content )
			            if argvar == nil {
			               argvar, e = types.NewVariableP(ent.GetLocal(), strings.ToLower(ntok.Content), vt, "0", true)
			            }

						//System.Out.Println( "--- Set: " + ntok.Content + " to [" + vtok.Content + "]");

			            //fmt.Printf("*** FN Set: %s to %s\n", ntok.Content, vtok.Content )

						if vtok.Type == types.VARIABLE {

			                var tmp *types.Variable
			                var strval string = "0"
			                //var e error

							if ent.GetLocal().Exists(vtok.Content) {
							   tmp = ent.GetLocal().Get(strings.ToLower(vtok.Content))
			                   strval, e = tmp.GetContentScalar()
			                   if e != nil {
			                   	  return result, e
			                   }
			                }

							e = argvar.SetContentScalar( strval )
			                if e != nil {
			                   return result, e
			                }

						} else {
							e = argvar.SetContentScalar( vtok.Content )
			                if e != nil {
			                   return result, e
			                }
						} */

	}

	//fmt.Printf("FN: %s\n", ent.TokenListAsString(ecopy) )

	/* ready to evaluate */

	rtok = ent.ParseTokensForResult(ecopy)

	/* copy to result token */
	result.Type = rtok.Type
	result.Content = rtok.Content

	/* return to old frame */
	ent.Return(true)
	// reset program counter
	if ent.GetState() == types.RUNNING {
		ent.SetPC(&opc)
	} else {
		ent.SetLPC(&opc)
	}

	/* enforce non void return */
	return result, nil

}
