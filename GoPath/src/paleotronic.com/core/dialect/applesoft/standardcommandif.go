package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/exception"
)

type StandardCommandIF struct {
	dialect.Command
}

func (this *StandardCommandIF) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var then_pos int
	var else_pos int
	var to_index int
	var from_index int
	var nl int
	var kw_pos int
	var conditions types.TokenList
	var clause types.TokenList
	var res types.Token
	var inckw bool

	result = 0

	then_pos = tokens.IndexOf(types.KEYWORD, "THEN")
	else_pos = tokens.IndexOf(types.KEYWORD, "ELSE")
	kw_pos = tokens.IndexOf(types.KEYWORD, "GOTO")

	inckw = false
	if (then_pos == -1) && (kw_pos >= 0) {
		then_pos = kw_pos
		inckw = true
	}

	if then_pos == -1 {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	conditions = *tokens.SubList(0, then_pos)

	res = caller.ParseTokensForResult(conditions)

	// System.Out.Println( "DEBUG: IF RESULT ("+caller.TokenListAsString(conditions)+") returns "+res.AsInteger() );

	// removed free call here /* dont need this anymore */

	/* false && no else, just exit */
	if (res.AsInteger() == 0) && (else_pos == -1) {
		/* roll to next line */
		if caller.GetState() == types.RUNNING {
			b := caller.GetCode()
			nl = b.NextAfter(caller.GetPC().Line)
			caller.GetPC().Line = nl
			caller.GetPC().Statement = 0
			caller.GetPC().Token = 0
		} else {
			b := caller.GetDirectAlgorithm();
			nl = b.NextAfter(caller.GetLPC().Line)
			caller.GetLPC().Line = nl
			caller.GetLPC().Statement = 0
			caller.GetLPC().Token = 0
		}
		// removed free call here;
		return result, nil
	}

	to_index = tokens.Size()
	if inckw {
		from_index = then_pos
	} else {
		from_index = then_pos + 1
	}

	if (res.AsInteger() != 0) && (else_pos != -1) {
		to_index = else_pos
	} else if (res.AsInteger() == 0) && (else_pos != -1) {
		from_index = else_pos + 1
	}

	// removed free call here /* dont need res any more */

	clause = *tokens.SubList(from_index, to_index)

	if clause.Size() == 0 {
		clause.Push(types.NewToken(types.KEYWORD, "rem"))
	}

	if (clause.Size() == 1) && ((clause.Get(0).Type == types.NUMBER) || (clause.Get(0).Type == types.INTEGER)) {
		clause.UnShift(types.NewToken(types.KEYWORD, "goto"))
	}

	caller.GetDialect().ExecuteDirectCommand(clause, caller, Scope, &LPC)

	// removed free call here;

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandIF) Syntax() string {

	/* vars */
	var result string

	result = "IF"

	/* enforce non void return */
	return result

}
