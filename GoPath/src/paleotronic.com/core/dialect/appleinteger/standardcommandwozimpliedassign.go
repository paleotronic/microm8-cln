package appleinteger

import (
	"paleotronic.com/fmt"
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardCommandWozIMPLIEDASSIGN struct {
	dialect.Command
}

func (this *StandardCommandWozIMPLIEDASSIGN) FindAssignmentSymbol(tokens types.TokenList) int {
	var eqidx int
	eqidx = -1
	idx := 0
	bc := 0
	for (idx < tokens.Size()) && (eqidx == -1) {
		tt := tokens.Get(idx)
		if tt.Type == types.CBRACKET {
			bc--
		} else if tt.Type == types.OBRACKET || tt.Type == types.FUNCTION || tt.Type == types.PLUSFUNCTION {
			bc++
		} else if (tt.Type == types.ASSIGNMENT) && (bc == 0) {
			eqidx = idx
		}
		idx++
	}

	return eqidx
}

func (this *StandardCommandWozIMPLIEDASSIGN) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	tla := types.NewTokenListArray()
	var tl types.TokenList
	var expr types.TokenList
	var dl []int
	var adl []int
	var vv *types.Variable
	var ch rune
	var i int
	var eqidx int
	//	var z int
	//	var e types.Event
	var cval string
	var ival string

	result = 0
	this.Cost = 1

	/* handle var declaration: <varname> == XXX */
	if tokens.Size() < 3 {
		return result, exception.NewESyntaxError("Syntax Error")
	}

	nametok := tokens.Shift()

	if nametok.Type != types.VARIABLE {
		return result, exception.NewESyntaxError("Syntax Error")
	}

	vtok := *types.NewToken(types.STRING, "")

	dl = make([]int, 0)
	//dl[0] = 0

	if tokens.Size() > 0 {

		eqidx = this.FindAssignmentSymbol(tokens)

		if eqidx == -1 {
			return result, exception.NewESyntaxError("Syntax Error")
		}

		tl = *tokens.SubList(0, eqidx)
		expr = *tokens.SubList(eqidx+1, tokens.Size())

		//vtok = caller.ParseTokensForResult( tla.Get(tla.Size()-1) );
		vtok = caller.ParseTokensForResult(expr)

		if tl.Size() > 0 {
			// might be dimensions;
			dl, _ = caller.IndicesFromTokens(tl, "(", ")")
		}

	}

	/* if (we get here try create var */
	/* does it exist? */
	if caller.GetLocal().Exists(strings.ToLower(nametok.Content)) {

		//caller.GetVDU().PutStr(nametok.Content+" exists");

		vv = caller.GetLocal().Get(strings.ToLower(nametok.Content))

		if vv.Kind == types.VT_EXPRESSION {
			vtok.Content = caller.TokenListAsString(tla[tla.Size()-1])
		}

		//v.GetContentByIndex(this.Dialect.ArrayDimDefault, this.Dialect.ArrayDimMax, dl);
		if (vv.Kind == types.VT_INTEGER) || (vv.Kind == types.VT_FLOAT) {
			vtok.Content = utils.Flatten7Bit(vtok.Content)
		}

		if vv.Kind != types.VT_STRING {

			if len(dl) > 0 {
				ov, _ := vv.GetContentByIndex(caller.GetDialect().GetArrayDimDefault(), caller.GetDialect().GetArrayDimMax(), dl)
				ee := vv.SetContentByIndex(caller.GetDialect().GetArrayDimDefault(), caller.GetDialect().GetArrayDimMax(), dl, vtok.Content)
				if ee != nil {
					return 0, ee
				}
				if caller.IsDebug() {
					msg := fmt.Sprintf("Existing var %s(%v): <- %s (was %s)", nametok.Content, dl, vtok.Content, ov)
					caller.Log("VAR", msg)
				}
			} else if vv.IsArray() {
				dl = make([]int, len(vv.Dimensions()))
				for i = 0; i <= len(dl)-1; i++ {
					dl[i] = 0
				}
				ov, _ := vv.GetContentByIndex(caller.GetDialect().GetArrayDimDefault(), caller.GetDialect().GetArrayDimMax(), dl)
				ee := vv.SetContentByIndex(caller.GetDialect().GetArrayDimDefault(), caller.GetDialect().GetArrayDimMax(), dl, vtok.Content)
				if ee != nil {
					return 0, ee
				}
				if caller.IsDebug() {
					msg := fmt.Sprintf("Existing var %s(%v): <- %s (was %s)", nametok.Content, dl, vtok.Content, ov)
					caller.Log("VAR", msg)
				}
			} else {
				ov, _ := vv.GetContentScalar()
				ee := vv.SetContentScalar(vtok.Content)
				if ee != nil {
					return 0, ee
				}
				if caller.IsDebug() {
					msg := fmt.Sprintf("Existing var %s: <- %s (was %s)", nametok.Content, vtok.Content, ov)
					caller.Log("VAR", msg)
				}
			}

			if (caller.IsRunning()) && (caller.GetDialect().GetWatchVars().ContainsKey(strings.ToLower(nametok.Content))) {
				apple2helpers.PutStr(caller, "#"+utils.IntToStr(LPC.Line)+":"+strings.ToUpper(nametok.Content)+" ")
			}

		} else {
			/* handle crap here for indexing */
			if (caller.IsRunning()) && (caller.GetDialect().GetWatchVars().ContainsKey(strings.ToLower(nametok.Content))) {
				apple2helpers.PutStr(caller, "#"+utils.IntToStr(LPC.Line)+":"+strings.ToUpper(nametok.Content)+" ")
			}

			if len(dl) == 0 {
				ov, _ := vv.GetContentScalar()
				ee := vv.SetContentScalar(vtok.Content)
				//fmt.Println(vv.Name, vtok.Content, ee)
				if ee != nil {
					return 0, ee
				}
				if caller.IsDebug() {
					msg := fmt.Sprintf("Existing var %s: <- %s (was %s)", nametok.Content, vtok.Content, ov)
					caller.Log("VAR", msg)
				}
			} else if len(dl) == 1 {
				/* put string in at index */
				i = dl[0] - 1
				cval, _ = vv.GetContentScalar()
				ov := cval
				ival = vtok.Content

				for len(cval) < i {
					cval = cval + " "
				}

				if len(cval)-1 > i {
					cval = cval[0:i]
				}

				cval = cval + ival

				vtok.Content = cval
				if (vv.Kind == types.VT_INTEGER) || (vv.Kind == types.VT_FLOAT) {
					vtok.Content = utils.Flatten7Bit(vtok.Content)
				}

				ee := vv.SetContentScalar(vtok.Content)
				if ee != nil {
					return 0, ee
				}

				if caller.IsDebug() {
					msg := fmt.Sprintf("Existing var %s: <- %s (was %s)", nametok.Content, vtok.Content, ov)
					caller.Log("VAR", msg)
				}

			} else {
				return result, exception.NewESyntaxError("SYNTAX ERROR")
			}
		}

	} else {
		/* create it need to imply type */

		//caller.GetVDU().PutStr(nametok.Content+" does not exist");

		vt := types.VT_INTEGER

		ch = rune(nametok.Content[len(nametok.Content)-1])

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
		case '&':
			{
				vt = types.VT_EXPRESSION
				vtok.Content = caller.TokenListAsString(*tla.Get(tla.Size() - 1))
			}
		}

		if (vt == types.VT_INTEGER) || (vt == types.VT_FLOAT) {
			vtok.Content = utils.Flatten7Bit(vtok.Content)
		}

		if len(dl) > 0 {
			//return types.NewESyntaxError("Must declare an array using dim");
			//SetLength(adl, 1);
			adl = make([]int, 1)
			adl[0] = 10
			vv, _ = types.NewVariablePA(caller.GetLocal(), strings.ToLower(nametok.Content), vt, vtok.Content, true, adl)
			//vv.ZeroIsScalar = true
			caller.CreateVarLower(strings.ToLower(nametok.Content), *vv)
			//apple2helpers.PutStr(caller,"after create");
			caller.GetVar(strings.ToLower(nametok.Content)).Owner = caller.GetName()
		} else {
			//apple2helpers.PutStr(caller,"before create");
			//adl = make([]int, 1)
			//adl[0] = 1
			var e error
			vv, e = types.NewVariablePZ(caller.GetLocal(), strings.ToLower(nametok.Content), vt, vtok.Content, true)
			if e != nil {
				return 0, e
			}
			//vv.ZeroIsScalar = true
			vv.SetContentScalar(vtok.Content)
			caller.CreateVarLower(strings.ToLower(nametok.Content), *vv)
			//apple2helpers.PutStr(caller,"after create");
			caller.GetVar(strings.ToLower(nametok.Content)).Owner = caller.GetName()
			//apple2helpers.PutStr(caller,"after set owner");

			if (!caller.IsRunning()) && (!caller.IsRunningDirect()) {
				apple2helpers.PutStr(caller, "Ok: "+strings.ToLower(nametok.Content)+" == "+vtok.AsString()+"\r\n")
			}
		}

	}

	//	if strings.ToUpper(nametok.Content) == "SPEED" {
	//		e = types.Event{}
	//		e.Name = "VARCHANGE"
	//		e.Target = strings.ToUpper(nametok.Content)
	//		e.IntParam = vtok.AsInteger()
	//		caller.HandleEvent(e)
	//	}

	//	if strings.ToUpper(nametok.Content) == "JOYPAD" {
	//		e = types.Event{}
	//		e.Name = "VARCHANGE"
	//		e.Target = strings.ToUpper(nametok.Content)
	//		e.IntParam = vtok.AsInteger()
	//		caller.HandleEvent(e)
	//		Paddle.SetPaddleValuesFromVar(caller.GetVar("joypad"))
	//	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandWozIMPLIEDASSIGN) Syntax() string {

	/* vars */
	var result string

	result = "IMPLIEDASSIGN <type> <name>[(size)]"

	/* enforce non void return */
	return result

}
