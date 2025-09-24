package applesoft

import (
	"errors"
	"strings"

	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/log"
	"paleotronic.com/utils"
)

type StandardCommandIMPLIEDASSIGN struct {
	dialect.Command
}

func ValidateType(vt types.VariableType, t *types.Token) error {

	e := errors.New("TYPE MISMATCH ERROR")

	switch vt {
	case types.VT_STRING:
		if t.Type != types.STRING {
			return e
		}
	case types.VT_FLOAT:
		if t.Type != types.NUMBER && t.Type != types.INTEGER {
			return e
		}
	case types.VT_INTEGER:
		if t.Type != types.NUMBER && t.Type != types.INTEGER {
			return e
		}
		// fix the value
		f := utils.StrToFloat(utils.NumberPart(t.Content))

		if f < 0 && float32(int(f)) != f {
			f -= 1
		}

		t.Content = utils.IntToStr(int(f))
	}

	return nil
}

func (this *StandardCommandIMPLIEDASSIGN) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var nametok types.Token
	var vtok types.Token
	tla := types.NewTokenListArray()
	var tl types.TokenList
	var expr types.TokenList
	var dl []int
	var adl []int
	//var vv types.Variable
	var ch rune
	var i int
	var eqidx int
	var e types.Event

	result = 0
	this.Cost = 1

	////fmt.Println("IMPASS:", caller.TokenListAsString(tokens))

	//writeln( "IMP ASS: "+caller.TokenListAsString(tokens) );

	/* handle var declaration: <varname> == types.XXX */
	if tokens.Size() < 3 {
		return result, exception.NewESyntaxError("Syntax Error")
	}

	nametok = *tokens.Shift()

	if nametok.Type != types.VARIABLE {
		return result, exception.NewESyntaxError("Syntax Error")
	}

	vtok = *types.NewToken(types.STRING, "")

	dl = make([]int, 0)

	if tokens.Size() > 0 {

		eqidx = this.FindAssignmentSymbol(tokens)

		if eqidx == -1 {
			return result, exception.NewESyntaxError("Syntax Error")
		}

		tl = *tokens.SubList(0, eqidx)
		expr = *tokens.SubList(eqidx+1, tokens.Size())

		tmptok, serr := caller.GetDialect().ParseTokensForResult(caller, expr)
		if serr != nil {
			return 0, serr
		}
		vtok = *tmptok

		if tl.Size() > 0 {
			dl, _ = caller.IndicesFromTokens(tl, "(", ")")
		}

	}

	/* if (we get here try create var */
	/* does it exist? */
	if caller.GetLocal().Exists(strings.ToLower(nametok.Content)) || caller.GetLocal().ExistsIndexed(strings.ToLower(nametok.Content)) {

		vvv := caller.GetLocal().Get(nametok.Content)

		////fmt.Println("kind:", vvv.Kind)

		if vvv.Kind == types.VT_EXPRESSION {
			vtok.Content = caller.TokenListAsString(*tla.Get(tla.Size() - 1))
		}

		eee := ValidateType(vvv.Kind, &vtok)
		if eee != nil {
			return 0, errors.New(eee.Error() + "(" + vtok.Content + ")")
		}

		if len(dl) > 0 {
			log.Println("Setting existing dimed var")
			log.Println(dl)
			log.Println(vtok.Content)
			ov, _ := vvv.GetContentByIndex(caller.GetDialect().GetArrayDimDefault(), caller.GetDialect().GetArrayDimMax(), dl)
			ee := vvv.SetContentByIndex(caller.GetDialect().GetArrayDimDefault(), caller.GetDialect().GetArrayDimMax(), dl, vtok.Content)

			if ee != nil {
				return 0, errors.New(ee.Error() + "(" + vtok.Content + ")")
			}

			if caller.IsDebug() {
				msg := fmt.Sprintf("Existing var %s(%v): <- %s (was %s)", nametok.Content, dl, vtok.Content, ov)
				caller.Log("VAR", msg)
			}

		} else {

			ov, _ := vvv.GetContentScalar()

			ee := vvv.SetContentScalar(vtok.Content)
			if ee != nil {
				return 0, errors.New(ee.Error() + "(" + vtok.Content + ")")
			}

			if caller.IsDebug() {
				msg := fmt.Sprintf("Existing var %s: <- %s (was %s)", nametok.Content, vtok.Content, ov)
				caller.Log("VAR", msg)
			}
		}

		if (caller.IsRunning()) && (caller.GetDialect().GetWatchVars().ContainsKey(strings.ToLower(nametok.Content))) {
			//apple2helpers.PutStr(caller,"#\"+IntToStr(LPC.Line)+\":" + strings.ToUpper(nametok.Content) + ' ')
		}

	} else {
		/* create it need to imply type */

		//caller.PutStr(nametok.Content+" does not exist");

		vt := types.VT_FLOAT

		ch = rune(nametok.Content[len(nametok.Content)-1])

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
		case '&':
			{
				vt = types.VT_EXPRESSION
				vtok.Content = caller.TokenListAsString(*tla.Get(tla.Size() - 1))
				break
			}
		}

		eee := ValidateType(vt, &vtok)
		if eee != nil {
			return 0, errors.New(eee.Error() + "(" + vtok.Content + ")")
		}

		if len(dl) > 0 {
			//return types.NewESyntaxError("Must declare an array using dim");
			adl = make([]int, len(dl))
			for i = 0; i <= len(adl)-1; i++ {
				adl[i] = 11
			}
			//adl.Get(0) = 10;
			vvv := ""
			if vt != types.VT_STRING {
				vvv = "0"
			}
			v, e := types.NewVariablePA(caller.GetLocal(), strings.ToLower(nametok.Content), vt, vvv, true, adl)
			if e != nil {
				return 0, e
			}
			//apple2helpers.PutStr(caller,"after create");
			v.Owner = caller.GetName()
			// now set indexed value
			v.SetContentByIndex(caller.GetDialect().GetArrayDimDefault(), caller.GetDialect().GetArrayDimMax(), dl, vtok.Content)

			if (caller.IsRunning()) && (caller.GetDialect().GetWatchVars().ContainsKey(strings.ToLower(nametok.Content))) {
				//apple2helpers.PutStr(caller,"#" + utils.IntToStr(LPC.Line) + ":" + strings.ToUpper(nametok.Content) + ' ')
			}
		} else {
			//apple2helpers.PutStr(caller,"before create");

			if len(vtok.Content) > 255 {
				return 0, errors.New("STRING TOO LONG")
			}

			v, e := types.NewVariableP(caller.GetLocal(), strings.ToLower(nametok.Content), vt, vtok.Content, true)
			if e != nil {
				return 0, e
			}
			//apple2helpers.PutStr(caller,"after create");
			v.Owner = caller.GetName()
			//apple2helpers.PutStr(caller,"after set owner");
		}

	}

	if strings.ToUpper(nametok.Content) == "SPEED" {
		e.Name = "VARCHANGE"
		e.Target = strings.ToUpper(nametok.Content)
		e.IntParam = vtok.AsInteger()
		caller.HandleEvent(e)
	}

	if strings.ToUpper(nametok.Content) == "HCOLOR" {
		e.Name = "VARCHANGE"
		e.Target = strings.ToUpper(nametok.Content)
		e.IntParam = vtok.AsInteger()
		caller.HandleEvent(e)
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandIMPLIEDASSIGN) FindAssignmentSymbol(tokens types.TokenList) int {
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

func (this *StandardCommandIMPLIEDASSIGN) Syntax() string {

	/* vars */
	var result string

	result = "IMPLIEDASSIGN <type> <name>[(size)]"

	/* enforce non void return */
	return result

}
