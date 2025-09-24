package applesoft

import (
	//"paleotronic.com/fmt"
	//"paleotronic.com/log"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardCommandPOKE struct {
	dialect.Command
}

func (this *StandardCommandPOKE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	//var vtok types.Token
	//var cr types.CodeRef
	var tla types.TokenListArray
	var tl types.TokenList
	//TokenList  i;
	var x int
	var y int
	//	var z int

	result = 0

	tla = caller.SplitOnTokenWithBrackets(tokens, *types.NewToken(types.SEPARATOR, ","))
	tl = *types.NewTokenList()

	for _, i := range tla {
		//writeln("----------------==================> ", caller.TokenListAsString(i));
		vtok := caller.ParseTokensForResult(i)
		//writeln("result: ", vtok.Type, '/', vtok.Content);
		// removed free call here /*free tokens*/
		tl.Push(&vtok)
	}

	if tl.Size() != 2 {
		return result, exception.NewESyntaxError("POKE expects 2 parameters")
	}

	list := []types.TokenType{types.INTEGER, types.NUMBER}

	if !(tl.Get(0).IsType(list)) && (tl.Get(0).IsType(list)) {
		return result, exception.NewESyntaxError("POKE expects numbers")
	}

	x = tl.Get(0).AsInteger() % (65536 * 2)

	if x < 0 {
		x = 65536 + x
	}

	y = tl.Get(1).AsInteger()

	/* free tokens */
	// removed free call here;

	//writeln("POKE ",x,", ", y);
	if (x == 2053) && (utils.Pos("INTEGER", caller.GetDialect().GetTitle()) > 0) {
		//System.Err.Println("Is integer");
		if caller.GetFirstString() != "" {
			//System.Err.Println("first string set");
			v := caller.GetVar(caller.GetFirstString())
			ss, _ := v.GetContentScalar()
			ss = string(rune(y))
			v.SetContentScalar(ss)
		} else {
			caller.SetMemory(2053, uint64(y))
		}
	} else if x >= 1024 && x <= 2047 {
		n := int(caller.GetMemory(x)&0xffff0000) | (y & 65535)
		caller.SetMemory(x, uint64(n))
	} else {
		////fmt.Printf("Set memory %d to %d\n", x, y&255)
		caller.SetMemory(x, uint64(y)&0xff)
	}

	if x >= 32 && x <= 37 {
		txt := apple2helpers.GETHUD(caller, "TEXT")
		if txt != nil {
			caller.TBCheck(txt.Control)
		}
	}

	if x == 230 {
		if y == 32 {
			caller.SetCurrentPage("HGR1")
		} else if y == 64 {
			caller.SetCurrentPage("HGR2")
		}
	}

	if x == 51 {
		caller.SetPrompt(string(rune(y)))
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandPOKE) Syntax() string {

	/* vars */
	var result string

	result = "POKE <addr>, <value>"

	/* enforce non void return */
	return result

}
