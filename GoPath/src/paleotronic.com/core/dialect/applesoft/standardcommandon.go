package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"strings"
)

type StandardCommandON struct {
	dialect.Command
}

func (this *StandardCommandON) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var action types.Token
	var varname types.Token
	//var tok types.Token
	var tla types.TokenListArray
	var tl types.TokenList
	//TokenList  tt;
	var ve types.TokenList
	var cl types.TokenList
	var gotopos int
	var gosubpos int
	var p int
	var x int
	//var y int
	//var cr types.CodeRef

	result = 0

	gotopos = tokens.IndexOfN(0, types.KEYWORD, "goto")
	gosubpos = tokens.IndexOfN(0, types.KEYWORD, "gosub")

	p = -1

	if gotopos >= 0 {
		p = gotopos
	}

	if gosubpos >= 0 {
		p = gosubpos
	}

	if p == -1 {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	action = *tokens.Get(p)

	ve = *tokens.SubList(0, p)
	cl = *tokens.SubList(p+1, tokens.Size())

	//apple2helpers.PutStr(caller, "DEBUG: "+caller.Dialect.TokenListAsString(ve)+PasUtil.CRLF);

	varname = caller.ParseTokensForResult(ve)

	//writeln( varname.Type, '/', varname.Content );

	if !((varname.Type == types.NUMBER) || (varname.Type == types.BOOLEAN) || (varname.Type == types.INTEGER)) {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	x = varname.AsInteger()

	if (action.Type != types.KEYWORD) || ((strings.ToLower(action.Content) != "goto") && (strings.ToLower(action.Content) != "gosub")) {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	/* keyword is okay */
	tla = caller.SplitOnTokenWithBrackets(cl, *types.NewToken(types.SEPARATOR, ","))

	tl = *types.NewTokenList()
	for _, tt := range tla {
		tok := caller.ParseTokensForResult(tt)

		//caller.PutStr(tok.Content + " ")

		if tok.Type != types.NUMBER {
			return result, exception.NewESyntaxError("SYNTAX ERROR")
		}

		tl.Push(&tok)
	}

	/* valid index */
	if x < 1 {
		return result, nil /* fall through */
	}

	if x > tl.Size() {
		return result, nil
	}

	cl.Clear()
	cl.Push(&action)
	cl.Push(tl.Get(x - 1))

	caller.GetDialect().ExecuteDirectCommand(cl, caller, Scope, &LPC)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandON) Syntax() string {

	/* vars */
	var result string

	result = "ON"

	/* enforce non void return */
	return result

}
