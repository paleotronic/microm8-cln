package applesoft

import (
	"math"
	//	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardCommandASNEXT struct {
	dialect.Command
}

func (this *StandardCommandASNEXT) Syntax() string {

	/* vars */
	var result string

	result = "NEXT"

	/* enforce non void return */
	return result

}

func (this *StandardCommandASNEXT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var nxt types.CodeRef
	var tla types.TokenListArray
	var forvarname string
	var inLoop bool
	var t types.Token
	var ls types.LoopState
	var clause types.TokenList
	var tl types.TokenList
	var match bool
	var varindex int

	result = 0
	this.Cost = 1000000000 / 1100

	if (caller.GetLoopStack().Size() == 0) || (caller.GetLoopVariable() == "") {
		return result, exception.NewESyntaxError("next without for")
	}

	if tokens.Size() == 0 {
		if caller.GetLoopStack().Size() > 0 {
			forvarname = caller.GetLoopStack().Get(caller.GetLoopStack().Size() - 1).VarName
		} else {
			forvarname = caller.GetLoopVariable()
		}
		tokens.Push(types.NewToken(types.VARIABLE, forvarname))
	}

	tla = caller.SplitOnToken(tokens, *types.NewToken(types.SEPARATOR, ","))

	match = false

	varindex = 0
	for (varindex < tla.Size()) && (!match) {
		tl = *tla.Get(varindex)

		forvarname = tl.Get(0).Content

		z := caller.GetLoopStack().Size() - 1
		for !match && z >= 0 {
			match = caller.GetLoopStack().Get(z).VarName == forvarname
			if !match {
				z--
			}
		}

		if !match {
			varindex++
			continue
		}

		ls = *caller.GetLoopStack().Get(z)

		inLoop = true

		if ls.Step > 0 || ls.Start < ls.Finish {
			/* increment */
			clause = *types.NewTokenList()
			clause.Push(types.NewToken(types.VARIABLE, forvarname))
			clause.Push(types.NewToken(types.ASSIGNMENT, "="))
			clause.Push(types.NewToken(types.VARIABLE, forvarname))
			clause.Push(types.NewToken(types.OPERATOR, "+"))
			clause.Push(types.NewToken(types.NUMBER, utils.FloatToStr(math.Abs(ls.Step))))
			caller.GetDialect().ExecuteDirectCommand(clause, caller, Scope, &LPC)
			// removed free call here /*free tokens*/

			clause = *types.NewTokenList()
			clause.Push(types.NewToken(types.VARIABLE, forvarname))
			clause.Push(types.NewToken(types.ASSIGNMENT, "<="))
			clause.Push(types.NewToken(types.NUMBER, utils.FloatToStr(math.Abs(ls.Finish))))

			t = caller.ParseTokensForResult(clause)

			inLoop = (t.AsInteger() != 0)
		} else if ls.Step < 0 || ls.Start > ls.Finish {
			/* decrement */
			clause = *types.NewTokenList()
			clause.Push(types.NewToken(types.VARIABLE, forvarname))
			clause.Push(types.NewToken(types.ASSIGNMENT, "="))
			clause.Push(types.NewToken(types.VARIABLE, forvarname))
			clause.Push(types.NewToken(types.OPERATOR, "-"))
			clause.Push(types.NewToken(types.NUMBER, utils.FloatToStr(math.Abs(ls.Step))))
			caller.GetDialect().ExecuteDirectCommand(clause, caller, Scope, &LPC)

			clause = *types.NewTokenList()
			clause.Push(types.NewToken(types.VARIABLE, forvarname))
			clause.Push(types.NewToken(types.ASSIGNMENT, ">="))
			clause.Push(types.NewToken(types.NUMBER, utils.FloatToStr(ls.Finish)))

			t = caller.ParseTokensForResult(clause)

			inLoop = (t.AsInteger() != 0)
		} else {
			inLoop = ((ls.Step != 0) || (ls.Start != ls.Finish))
		}

		//writeln( "WATCH: inLoop == ",inLoop);

		if inLoop {

			if caller.GetState() == types.DIRECTRUNNING {
				caller.GetLPC().Line = ls.Entry.Line
				caller.GetLPC().Statement = ls.Entry.Statement
				caller.GetLPC().Token = ls.Entry.Token
				caller.GetLPC().SubIndex = -1
			} else {
				caller.GetPC().Line = ls.Entry.Line
				caller.GetPC().Statement = ls.Entry.Statement
				caller.GetPC().Token = ls.Entry.Token
				caller.GetPC().SubIndex = -1
			}

			result = 2
			return result, nil
		}

		// collapse the stack one level;
		if varindex == tla.Size()-1 {
			nxt = caller.GetNextStatement(LPC)
		} else {
			nxt = *types.NewCodeRefCopy(LPC)
		}

		// remove top entry
		if z > caller.GetLoopBase() {
			for caller.GetLoopStack().Size() > z {
				caller.GetLoopStack().Remove(caller.GetLoopStack().Size() - 1)
			}
		}

		if caller.GetState() == types.DIRECTRUNNING {
			caller.GetLPC().Line = nxt.Line
			caller.GetLPC().Statement = nxt.Statement
			caller.GetLPC().Token = nxt.Token
			caller.GetLPC().SubIndex = (varindex + 1) * -1
		} else {
			caller.GetPC().Line = nxt.Line
			caller.GetPC().Statement = nxt.Statement
			caller.GetPC().Token = nxt.Token
			caller.GetPC().SubIndex = (varindex + 1) * -1
		}

		// inc && fallback;
		varindex++
	}

	if !match {
		return result, exception.NewESyntaxError("next without for")
	}

	data, _ := caller.GetLoopStack().MarshalBinary()
	for i, v := range data {
		caller.SetMemory(LOOPSTACK_ADDRESS+i, v)
	}

	/* enforce non void return */
	return result, nil

}
