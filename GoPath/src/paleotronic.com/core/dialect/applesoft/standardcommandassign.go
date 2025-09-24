package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/runestring"
)

type StandardCommandASSIGN struct {
	dialect.Command
}

func (this *StandardCommandASSIGN) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	//Token t;
	var v types.Token
	//var d types.Token
	//var ref types.CodeRef
	var clause types.TokenList
	var tl types.TokenList
	var pl types.TokenList
	var tla types.TokenListArray

	result = 0

	if (tokens.Size() < 1) || (tokens.LPeek().Type != types.VARIABLE) {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	tla = caller.SplitOnTokenWithBrackets(tokens, *types.NewToken(types.SEPARATOR, ","))

	pl = *types.NewTokenList()
	for _, tl1 := range tla {
		v = caller.ParseTokensForResult(tl1)
		// removed free call here /*free tokens*/
		pl.Push(&v)
	}

	if pl.Size() != 2 {
		return result, exception.NewESyntaxError("two parameters required")
	}

	// frist param could be a parsable;
	tl = *caller.GetDialect().Tokenize(runestring.Cast(pl.Get(0).Content))

	clause = *types.NewTokenList()
	//clause.Push( types.NewToken(types.VARIABLE, pl.Get(0).Content) );
	for _, t := range tl.Content {
		clause.Push(t)
	}
	clause.Push(types.NewToken(types.ASSIGNMENT, "="))
	clause.Push(pl.Get(1))

	caller.SetOuterVars(true)
	caller.GetDialect().ExecuteDirectCommand(clause, caller, Scope, &LPC)
	caller.SetOuterVars(false)

	// removed free call here /*tokens*/
	// removed free call here /*tokens*/

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandASSIGN) Syntax() string {

	/* vars */
	var result string

	result = "ASSIGN <name>, <value>"

	/* enforce non void return */
	return result

}
