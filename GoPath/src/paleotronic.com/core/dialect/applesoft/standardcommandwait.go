package applesoft

import (
//	"paleotronic.com/fmt"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandWAIT struct {
	dialect.Command
}

func (this *StandardCommandWAIT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	//var vtok types.Token
	//var cr types.CodeRef
	var tla types.TokenListArray
	var tl types.TokenList
	//TokenList  i;
	//var ntl types.TokenList
	//var x int
	var y int
	//var z int
	var cc int

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

	res := 0

	if tl.Size() == 3 {
		res = tl.Pop().AsInteger()
	}

	if tl.Size() != 2 {
		return result, exception.NewESyntaxError("WAIT expects 2 parameters")
	}

	list := []types.TokenType{types.NUMBER, types.INTEGER}

	if !(tl.Get(0).IsType(list)) && (tl.Get(0).IsType(list)) {
		return result, exception.NewESyntaxError("WAIT expects numbers")
	}

	addr := tl.Get(0).AsInteger()

	if addr < 0 {
		addr = 65536 + addr
	}

	y = (tl.Get(1).AsInteger() & 255)

	//z = caller.GetMemory()[x] & y;

	r := 0

	r = int(caller.GetMemory(addr&65535))

	////fmt.Printf("Wait() r == %d\n", r)

	//z = (r & y) & (~res & 0xff);   // (160 & 128) ^ (127)

	bb := 128
	exit := false
	for xx := 0; xx < 8; xx++ {

		if (y & bb) == bb {
			// look at this bit...
			//System.Err.Println("*** Check bit "+bb);
			cc = res & bb
			if cc != 0 {
				//System.Err.Println("Exit if 0");
				exit = ((r & bb) == 0)
			} else {
				//System.Err.Println("Exit if 1");
				exit = ((r & bb) != 0)
			}
			//System.Err.Println("exit = "+exit);
		}

		if exit {
			break
		}

		bb = bb >> 1

	}

	//System.Err.Println( "("+r+" AND "+y+") AND "+(255-res)+" == "+z );

	if !exit {
		////fmt.Println("I should not exit")
		if caller.IsRunningDirect() {
			caller.GetLPC().SubIndex = caller.GetLPC().SubIndex + 1
		} else {
			caller.GetPC().SubIndex = caller.GetPC().SubIndex + 1
		}
	}
	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandWAIT) Syntax() string {

	/* vars */
	var result string

	result = "WAIT <addr>,<bits>"

	/* enforce non void return */
	return result

}
