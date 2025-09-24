package appleinteger

import (
	"paleotronic.com/log"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
//    "paleotronic.com/utils"
	"strings"
	"errors"
)

type StandardCommandWozFOR struct {
	dialect.Command
}

func (this *StandardCommandWozFOR) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var eq_idx int
	var to_idx int
	var step_idx int
	var fromtok types.TokenList
	var steptok types.TokenList
	var totok types.TokenList
	var clause types.TokenList
	var step float64
	var ls types.LoopState

	result = 0

	if tokens.Size() < 5 {
		return result, exception.NewESyntaxError("Syntax Error")
	}

	step = 1.0

	eq_idx = tokens.IndexOf(types.ASSIGNMENT, "=")
	to_idx = tokens.IndexOf(types.KEYWORD, "TO")
	
	if eq_idx < 0 || to_idx < 0 {
		return result, errors.New("SYNTAX ERROR")
	}

	step_idx = tokens.IndexOf(types.KEYWORD, "STEP")

	tvar := tokens.Get(0)

	fromtok = *tokens.SubList(eq_idx+1, to_idx)
	tlo := caller.ParseTokensForResult(fromtok)

	if step_idx != -1 {
		steptok = *tokens.SubList(step_idx+1, tokens.Size())
		stepv := caller.ParseTokensForResult(steptok)
		step = stepv.AsExtended()
	} else {
		step_idx = tokens.Size()
	}

	totok = *tokens.SubList(to_idx+1, step_idx)
	thi := caller.ParseTokensForResult(totok)

	forvarname := strings.ToLower(tvar.Content)

	z := caller.IndexOfLoop(forvarname)
	if z > -1 {
		caller.GetLoopStack().Remove(z)
	}

	log.Println("forvarname", forvarname)

	if !caller.GetLocal().Exists(strings.ToLower(tvar.Content)) {
		v, _ := types.NewVariableP(caller.GetLocal(), strings.ToLower(tvar.Content), types.VT_INTEGER, "0", true)
		v.Owner = caller.GetName()
	} else {
		if caller.GetLocal().Get(strings.ToLower(tvar.Content)).Kind != types.VT_INTEGER {
			return result, exception.NewESyntaxError("TYPE ERR")
		}
	}

	clause = *types.NewTokenList()
	clause.Push(tvar)
	clause.Push(types.NewToken(types.ASSIGNMENT, "="))
	clause.Push(&tlo)
	caller.GetDialect().ExecuteDirectCommand(clause, caller, Scope, &LPC)
	// removed free call here /*free tokens*/

	/* create loop state */
	ls = types.LoopState{}
	ls.VarName = strings.ToLower(tvar.Content)
	ls.Start = tlo.AsExtended()
	ls.Finish = thi.AsExtended()
	ls.Step = step
	/* handle scope */
	ls.Code = *Scope

	ls.Entry = caller.GetNextStatement(LPC)

	caller.GetLoopStack().Add(ls)

	// Now create a stack point;
	caller.SetLoopVariable(ls.VarName)
	caller.SetLoopStep(ls.Step)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandWozFOR) Syntax() string {

	/* vars */
	var result string

	result = "FOR VN == N1 TO N2 [STEP N3]"

	/* enforce non void return */
	return result

}
