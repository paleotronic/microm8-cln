package applesoft

import (
//	"paleotronic.com/fmt"
	"math"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardCommandRENUMBER struct {
	dialect.Command
}

func NewStandardCommandRENUMBER() *StandardCommandRENUMBER {
	this := &StandardCommandRENUMBER{}
	this.ImmediateMode = true
	return this
}

func (this *StandardCommandRENUMBER) PadLeft(str string, width int) string {
	for len(str) < width {
		str = " " + str
	}

	return str
}

func (this *StandardCommandRENUMBER) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var l int
	var h int
	var w int
	var z int
	var s string
	//	var ft string
	//	var ln types.Line
	//var stmt types.Statement
	//var t types.Token
	//	var lt types.Token
	var ntl types.TokenList
	//var tl types.TokenList
	var tla types.TokenListArray

	result = 0

	ntl = *types.NewTokenList()

	//fmt.Println("S =", tokens.Size())

	if tokens.Size() == 2 {
		l = int(math.Abs(float64(tokens.LPeek().AsInteger())))
		h = int(math.Abs(float64(tokens.RPeek().AsInteger())))
		tokens = *types.NewTokenList()
		tokens.Push(types.NewToken(types.NUMBER, utils.IntToStr(1000)))
		tokens.Push(types.NewToken(types.SEPARATOR, ","))
		tokens.Push(types.NewToken(types.NUMBER, utils.IntToStr(10)))
	}

	//fmt.Println("S =", tokens.Size())

	tla = caller.SplitOnTokenWithBrackets(tokens, *types.NewToken(types.SEPARATOR, ","))

	for _, tl1 := range tla {
		t := caller.ParseTokensForResult(tl1)
		ntl.Push(&t)
	}

	tokens = *ntl.SubList(0, ntl.Size())

	a := caller.GetCode()
	//	b := &a

	l = 1000
	h = 10

	s = utils.IntToStr(h)
	w = len(s) + 1
	if w < 4 {
		w = 4
	}

	if l < 0 {
		return result, nil
	}

	/* now take extra params */
	if (tokens.Size() > 0) && ((tokens.Get(0).Type == types.NUMBER) || (tokens.Get(0).Type == types.INTEGER)) {
		z = tokens.Get(0).AsInteger()
		//env.VDU.PutStr(IntToStr(z)+PasUtil.CRLF);
		l = z
	}

	if (tokens.Size() > 1) && ((tokens.Get(1).Type == types.NUMBER) || (tokens.Get(1).Type == types.INTEGER)) {
		z = tokens.Get(1).AsInteger()
		//writeln( "h will be set to ", z );
		//nv.VDU.PutStr(IntToStr(z)+PasUtil.CRLF);
		h = z
	}

	//writeln( "l is set to ", l );
	//fmt.Println("L =", l)
	//fmt.Println("H =", h)

	newcode := caller.GetDialect().Renumber(*a, l, h)

	caller.SetCode(&newcode)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandRENUMBER) Syntax() string {

	/* vars */
	var result string

	result = "RENUMBER"

	/* enforce non void return */
	return result

}
