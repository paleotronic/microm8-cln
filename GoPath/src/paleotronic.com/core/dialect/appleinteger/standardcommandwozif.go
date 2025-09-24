package appleinteger

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandWozIF struct {
	dialect.Command
}

func (this *StandardCommandWozIF) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var then_pos int
	var else_pos int
	var to_index int
	var from_index int
	//	var nl int
	var conditions types.TokenList
	var clause types.TokenList
	var res types.Token
	var cr types.CodeRef

	result = 0

	then_pos = tokens.IndexOf(types.KEYWORD, "THEN")
	else_pos = tokens.IndexOf(types.KEYWORD, "ELSE")

	if then_pos == -1 {
		return result, exception.NewESyntaxError("IF without THEN")
	}

	conditions = *tokens.SubList(0, then_pos)

	//System.Out.Println(caller.TokenListAsString(conditions));

	res = caller.ParseTokensForResult(conditions)

	// removed free call here /* dont need this anymore */

	/* false && no else, just exit */
	if (res.AsInteger() == 0) && (else_pos == -1) {
		/* roll to next line */
		if caller.GetState() == types.RUNNING {
			cr = caller.GetNextStatement(*caller.GetPC())
			caller.GetPC().Line = cr.Line
			caller.GetPC().Statement = cr.Statement
			caller.GetPC().Token = 0
		} else {
			cr = caller.GetNextStatement(*caller.GetLPC())
			caller.GetLPC().Line = cr.Line
			caller.GetLPC().Statement = cr.Statement
			caller.GetLPC().Token = 0
		}
		// removed free call here;
		return result, nil
	}

	to_index = tokens.Size()
	from_index = then_pos + 1

	if (res.AsInteger() != 0) && (else_pos != -1) {
		to_index = else_pos
	} else if (res.AsInteger() == 0) && (else_pos != -1) {
		from_index = else_pos + 1
	}

	// removed free call here /* dont need res any more */

	clause = *tokens.SubList(from_index, to_index)

	if clause.Size() == 0 {
		return result, exception.NewESyntaxError("Badly formed IF statement")
	}

	if (clause.Size() == 1) && ((clause.Get(0).Type == types.NUMBER) || (clause.Get(0).Type == types.INTEGER)) {
		clause.UnShift(types.NewToken(types.KEYWORD, "goto"))
	}

	caller.GetDialect().ExecuteDirectCommand(clause, caller, Scope, &LPC)

	// removed free call here;

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandWozIF) Syntax() string {

	/* vars */
	var result string

	result = "IF"

	/* enforce non void return */
	return result

}
