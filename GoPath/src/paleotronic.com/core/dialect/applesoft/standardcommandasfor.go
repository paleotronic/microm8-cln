package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"errors"
)

type StandardCommandASFOR struct {
	dialect.Command
}

func (this *StandardCommandASFOR) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var eq_idx int
	var to_idx int
	var step_idx int
	var tvar types.Token
	var tlo types.Token
	var thi types.Token
	var stepv types.Token
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
	step_idx = tokens.IndexOf(types.KEYWORD, "STEP")
	
	if eq_idx < 0 || to_idx < 0 {
		return result, errors.New("SYNTAX ERROR")
	}

	tvar = *tokens.Get(0)

	fromtok = *tokens.SubList(eq_idx+1, to_idx)
	tlo = caller.ParseTokensForResult(fromtok)

	if step_idx != -1 {
		//writeln("steptok");
		steptok = *tokens.SubList(step_idx+1, tokens.Size())
		stepv = caller.ParseTokensForResult(steptok)
		step = stepv.AsExtended()
		// removed free call here /*free tokens*/
	} else {
		step_idx = tokens.Size()
	}

	totok = *tokens.SubList(to_idx+1, step_idx)
	thi = caller.ParseTokensForResult(totok)
	forvarname := tvar.Content

	z := caller.IndexOfLoopFromBase(forvarname)
	if z > -1 {
		caller.GetLoopStack().Remove(z)
	}

	if !caller.GetLocal().Exists(forvarname) {
		caller.GetLocal().Create(forvarname, types.VT_FLOAT, types.NewFloat5b(tlo.AsExtended()))
	} else {
		if caller.GetLocal().Get(forvarname).Kind != types.VT_FLOAT {
			return result, exception.NewESyntaxError("FOR variable "+forvarname+" must be a NUMBER")
		}
	}


	/* init loop var */

	clause = *types.NewTokenList()
	clause.Push(&tvar)
	clause.Push(types.NewToken(types.ASSIGNMENT, "="))
	clause.Push(&tlo)
	caller.GetDialect().ExecuteDirectCommand(clause, caller, Scope, &LPC)

	/* create loop state */
	ls = types.LoopState{}
	ls.VarName = tvar.Content
	ls.Start = tlo.AsExtended()
	ls.Finish = thi.AsExtended()
	ls.Step = step

	/* handle scope */
	ls.Code = *Scope
	ls.Entry = caller.GetNextStatement(LPC)

	caller.GetLoopStack().Add(ls)

	/* if (we got here without error make sure loop is on the stack */

	caller.SetLoopVariable(ls.VarName)
	caller.SetLoopStep(ls.Step)

	data, _ := caller.GetLoopStack().MarshalBinary()
	for i, v := range data {
		caller.SetMemory(LOOPSTACK_ADDRESS+i, v)
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandASFOR) Syntax() string {

	/* vars */
	var result string

	result = "FOR VN == N1 TO N2 [STEP N3]"

	/* enforce non void return */
	return result

}
