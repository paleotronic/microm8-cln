package appleinteger

import (
	"math"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	"strings"
)

type StandardCommandWozNEXT struct {
	dialect.Command
}

func (this *StandardCommandWozNEXT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

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

	if (caller.GetLoopStack().Size() == 0) || (caller.GetLoopVariable() == "") {
		return result, exception.NewESyntaxError("next without for")
	}

	if tokens.Size() == 0 {

		forvarname = caller.GetLoopVariable()
		tokens.Push(types.NewToken(types.VARIABLE, forvarname))

	}

	tla = caller.SplitOnToken(tokens, *types.NewToken(types.SEPARATOR, ","))

	match = false

	varindex = 0
	for (varindex < tla.Size()) && (!match) {
		tl = *tla.Get(varindex)

		forvarname = strings.ToLower(tl.Get(0).Content)

		//match = caller.LoopStates.ContainsKey(forvarname);

		z := 0
		for !match && z < caller.GetLoopStack().Size() {
			match = caller.GetLoopStack().Get(z).VarName == forvarname
			if !match {
				z++
			}
		}

		if !match {
			varindex++
			continue
		}

		ls = *caller.GetLoopStack().Get(z)

		inLoop = true

		//if ((strings.ToLower(forvarname) == 'x'));
		//writeln( "Next related to for loop at line ", ls.Entry.Line, " step == ", ls.Step);

		if ls.Step > 0 {
			/* increment */
			clause = *types.NewTokenList()
			clause.Push(types.NewToken(types.VARIABLE, forvarname))
			clause.Push(types.NewToken(types.ASSIGNMENT, "="))
			clause.Push(types.NewToken(types.VARIABLE, forvarname))
			clause.Push(types.NewToken(types.OPERATOR, "+"))
			clause.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", math.Abs(ls.Step))))
			caller.GetDialect().ExecuteDirectCommand(clause, caller, Scope, &LPC)
			// removed free call here /*free tokens*/

			clause = *types.NewTokenList()
			clause.Push(types.NewToken(types.VARIABLE, forvarname))
			clause.Push(types.NewToken(types.ASSIGNMENT, "<="))
			clause.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", math.Abs(ls.Finish))))

			t = caller.ParseTokensForResult(clause)

			// removed free call here /*free tokens*/

			//apple2helpers.PutStr(caller,"? "+caller.TokenListAsString(clause) + " == "+t.Content + utils.CRLF);

			inLoop = (t.AsInteger() != 0)
		} else if ls.Step < 0 {
			/* decrement */
			clause = *types.NewTokenList()
			clause.Push(types.NewToken(types.VARIABLE, forvarname))
			clause.Push(types.NewToken(types.ASSIGNMENT, "="))
			clause.Push(types.NewToken(types.VARIABLE, forvarname))
			clause.Push(types.NewToken(types.OPERATOR, "-"))
			clause.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", math.Abs(ls.Step))))
			caller.GetDialect().ExecuteDirectCommand(clause, caller, Scope, &LPC)
			// removed free call here /*free tokens*/

			clause = *types.NewTokenList()
			clause.Push(types.NewToken(types.VARIABLE, forvarname))
			clause.Push(types.NewToken(types.ASSIGNMENT, ">="))
			clause.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", math.Abs(ls.Finish))))

			t = caller.ParseTokensForResult(clause)

			//apple2helpers.PutStr(caller,"? "+caller.TokenListAsString(clause) + " == "+t.Content + utils.CRLF);

			// removed free call here /*free tokens*/

			inLoop = (t.AsInteger() != 0)
		}

		//writeln( "WATCH: inLoop == ",inLoop);

		if inLoop {
			// loop back;
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

			//writeln( "Return to Line == \", ls.Entry.Line, \", statement == ", ls.Entry.Statement);

			//apple2helpers.PutStr(caller,"LOOP Reset to line.Equals("+utils.IntToStr(LPC.Line)+"), statement == "+utils.IntToStr(LPC.Statement)+utils.CRLF);
			result = 2
			return result, nil
		}

		// collapse the stack one level;
		//apple2helpers.PutStr(caller,"Fall out of loop"+utils.CRLF);
		if varindex == tla.Size()-1 {
			nxt = caller.GetNextStatement(LPC)
		} else {
			nxt = *types.NewCodeRefCopy(LPC)
		}

		//for (!strings.ToLower(caller.LoopVariable).Equals(strings.ToLower(forvarname))) {
		//	  System.Err.Println("Collapse stack state as not var not "+forvarname+" ( was "+caller.LoopVariable+")");
		//	  caller.Return();
		//}
		//System.Err.Println("Collapse state for loop - "+caller.LoopVariable);
		//caller.Return();

		// remove top entry
		for caller.GetLoopStack().Size() > z {
			caller.GetLoopStack().Remove(caller.GetLoopStack().Size() - 1)
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

		//System.Err.Println("Exit for - next loop "+forvarname+" to "+utils.IntToStr(nxt.Line)+", "+utils.IntToStr(nxt.Statement));

		// inc && fallback;
		varindex++
	}

	if !match {
		return result, exception.NewESyntaxError("next with for")
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandWozNEXT) Syntax() string {

	/* vars */
	var result string

	result = "NEXT"

	/* enforce non void return */
	return result

}
