package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/log"
	"paleotronic.com/utils"
)

type StandardCommandHPLOT struct {
	dialect.Command
}

func (this *StandardCommandHPLOT) Syntax() string {

	/* vars */
	var result string

	result = "HPLOT x, y"

	/* enforce non void return */
	return result

}

func (this *StandardCommandHPLOT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	//var vtok *types.Token
	//	var cr types.CodeRef
	var tlt types.TokenListArray
	var tla types.TokenListArray
	var tl types.TokenList
	//TokenList  i;
	//	var first types.TokenList
	//	var second types.TokenList
	//TokenList  tokchunk;
	var x1 int
	var y1 int
	var x2 int
	var y2 int
	//	var to_pos int
	var hc int
	var ii int

	/*$I config.Inc*/

	result = 0

	//	i = *types.NewTokenList()
	//	i.Push(types.NewToken(types.VARIABLE, "HCOLOR"))
	//	vtok := caller.ParseTokensForResult(i)
	//	hc = vtok.AsInteger()

	hc = apple2helpers.GetHCOLOR(caller)

	tlt = caller.GetDialect().SplitOnToken(tokens, *types.NewToken(types.KEYWORD, "to"))

	x1 = apple2helpers.LastX
	y1 = apple2helpers.LastY

	ii = 0
	for _, tokchunk := range tlt {
		if (tokchunk.Size() == 0) && (ii == 0) {
			tokchunk.Push(types.NewToken(types.NUMBER, utils.IntToStr(apple2helpers.LastX)))
			tokchunk.Push(types.NewToken(types.SEPARATOR, ","))
			tokchunk.Push(types.NewToken(types.NUMBER, utils.IntToStr(apple2helpers.LastY)))
		}

		tla = caller.SplitOnTokenWithBrackets(tokchunk, *types.NewToken(types.SEPARATOR, ","))
		tl = *types.NewTokenList()

		for _, i1 := range tla {
			vtok := caller.ParseTokensForResult(i1)
			tl.Push(&vtok)
		}

		if tl.Size() != 2 {
			return 0, exception.NewESyntaxError("HPLOT expects 2 parameters")
		}

		if (tl.Get(0).Type != types.NUMBER) && (tl.Get(0).Type != types.INTEGER) {
			return 0, exception.NewESyntaxError("HPLOT expects numbers")
		}

		x2 = tl.Get(0).AsInteger()
		y2 = tl.Get(1).AsInteger()

		tl.Clear()

		if ii == 0 {
			if hc == 0 {
				log.Printf("HCOLOR(0).Plot(%d, %d) at Line %d\n", x2, y2, LPC.Line)
			}
			apple2helpers.HGRPlot(caller, uint64(x2), uint64(y2), uint64(hc))
			apple2helpers.LastX = x2
			apple2helpers.LastY = y2
		} else {
			if hc == 0 {
				log.Printf("HCOLOR(0).Line(%d, %d) - (%d, %d) at Line %d\n", x1, y2, x2, y2, LPC.Line)
			}
			apple2helpers.HGRLine(caller, uint64(x1), uint64(y1), uint64(x2), uint64(y2), uint64(hc))
			apple2helpers.LastX = x2
			apple2helpers.LastY = y2
		}

		x1 = x2
		y1 = y2

		/* increment */
		ii++

		tokchunk.Clear()
	}

	/* update memory locations */
	caller.SetMemory(224, uint64(apple2helpers.LastX%256))
	caller.SetMemory(225, uint64(apple2helpers.LastX/256))
	caller.SetMemory(226, uint64(apple2helpers.LastY%256))
	/* enforce non void return */
	return result, nil

}
