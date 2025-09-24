package applesoft

import (
	//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardCommandDRAW struct {
	dialect.Command
}

func (this *StandardCommandDRAW) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	//	var vtok types.Token
	var sn types.Token
	var xt types.Token
	var yt types.Token
	//	var cr types.CodeRef
	var tla types.TokenListArray
	var tlb types.TokenListArray
	//	var tl types.TokenList
	//	var i types.TokenList
	var x int
	var y int
	//	var z int
	var stAddr int
	var stNumShapes int
	var shIdx int
	var so int
	var shaddr int
	var hc int
	var rot float64
	var scl int
	var done bool
	var se hires.ShapeEntry

	result = 0

	stAddr = int(caller.GetMemory(0xe8) + (256 * caller.GetMemory(0xe9)))

	stNumShapes = int(caller.GetMemory(stAddr))

	tla = caller.GetDialect().SplitOnToken(tokens, *types.NewToken(types.KEYWORD, "at"))

	if tla.Size() != 2 {
		if tla.Size() == 0 {
			return result, exception.NewESyntaxError("SYNTAX ERROR")
		}
		// add last co-ords
		ttt := types.NewTokenList()
		ttt.Push(types.NewToken(types.NUMBER, utils.IntToStr(apple2helpers.LastX)))
		ttt.Push(types.NewToken(types.SEPARATOR, ","))
		ttt.Push(types.NewToken(types.NUMBER, utils.IntToStr(apple2helpers.LastY)))
		tla = tla.Add(*ttt)
	}

	tlb = caller.SplitOnTokenWithBrackets(tla[1], *types.NewToken(types.SEPARATOR, ","))
	if tlb.Size() != 2 {
		return 0, exception.NewESyntaxError("SYNTAX ERROR")
	}

	xt = caller.ParseTokensForResult(tlb[0])
	yt = caller.ParseTokensForResult(tlb[1])

	sn = caller.ParseTokensForResult(tla[0])

	shIdx = sn.AsInteger()
	x = xt.AsInteger()
	y = yt.AsInteger()

	apple2helpers.LastX = x
	apple2helpers.LastY = y

	if (shIdx < 1) || (shIdx > stNumShapes) {
		return 0, nil
	}

	so = stAddr + (shIdx * 2)

	shaddr = int(caller.GetMemory(so)+256*caller.GetMemory(so+1)) + stAddr

	for caller.GetMemory(shaddr) == 0 {
		shaddr++
	}

	done = false
	se = hires.NewShapeEntry()
	for !done {

		se = append(se, int(caller.GetMemory(shaddr)))

		if caller.GetMemory(shaddr) == 0 {
			done = true
		}

		if !done {
			shaddr++
		}
	}

	hc = apple2helpers.GetHCOLOR(caller)

	rot = float64(apple2helpers.GetROT(caller) % 64)
	scl = apple2helpers.GetSCALE(caller)

	apple2helpers.HGRShape(caller, se, x, y, scl, rot, hc, true)
	cc := apple2helpers.GetHGRCollisionCount(caller)
	caller.SetMemory(234, uint64(cc)&0xff)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandDRAW) Syntax() string {

	/* vars */
	var result string

	result = "DRAW <shape> AT <x>, <y>"

	/* enforce non void return */
	return result

}
