package applesoft

import (
	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandVLIN struct {
	dialect.Command
}

func (this *StandardCommandVLIN) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	//var vtok types.Token
	//var atok types.Token
	//	var cr types.CodeRef
	var tla types.TokenListArray
	var tls types.TokenListArray
	var tl types.TokenList
	//	var i types.TokenList
	var y1 int
	var y2 int
	var x int
	var v int

	result = 0

	tls = caller.GetDialect().SplitOnToken(tokens, *types.NewToken(types.KEYWORD, "at"))

	//caller.PutStr(IntToStr(tls.Size())+PasUtil.CRLF);

	if tls.Size() != 2 {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	tla = caller.SplitOnTokenWithBrackets(*tls.Get(0), *types.NewToken(types.SEPARATOR, ","))
	tl = *types.NewTokenList()

	for _, i1 := range tla {
		vtok := caller.ParseTokensForResult(i1)
		tl.Push(&vtok)
	}

	if tl.Size() != 2 {
		return result, exception.NewESyntaxError("VLIN expects 2 parameters")
	}

	if (tl.Get(0).Type != types.NUMBER) && (tl.Get(0).Type != types.INTEGER) {
		return result, exception.NewESyntaxError("VLIN expects numbers")
	}

	hc := apple2helpers.GetCOLOR(caller)
	atok := caller.ParseTokensForResult(*tls.Get(1))
	y1 = tl.Get(0).AsInteger()
	y2 = tl.Get(1).AsInteger()
	x = atok.AsInteger()
	//caller.GetVDU().Line( x, y1, x, y2, vtok.AsInteger() );
	if y1 > y2 {
		v = y1
		y1 = y2
		y2 = v
	}

	//writeln("-----------------------------------------> y1 == \", y1,\", y2 == ", y2);

	//	for v = y1; v <= y2; v++ {
	//		caller.GetVDU().GrPlot(x, v, vtok.AsInteger())
	//	}
	// if apple2helpers.GetColumns(caller) == 80 {
	// 	apple2helpers.LOGRVLine80(caller, uint64(y1), uint64(y2), uint64(x), uint64(hc))
	// } else {
	// 	apple2helpers.LOGRVLine40(caller, uint64(y1), uint64(y2), uint64(x), uint64(hc))
	// }

	modes := apple2helpers.GetActiveVideoModes(caller)

	if len(modes) == 1 {
		switch modes[0] {
		case "LOGR", "LGR2":
			apple2helpers.LOGRVLine40(caller, uint64(y1), uint64(y2), uint64(x), uint64(hc))
		case "DLGR":
			apple2helpers.LOGRVLine80(caller, uint64(y1), uint64(y2), uint64(x), uint64(hc))
		case "CUBE":
			apple2helpers.CUBE(caller).Bresenham3D(
				int(x), int(y1), 0,
				int(x), int(y2), 0,
				uint8(hc),
			)
		}
	} else {
		apple2helpers.LOGRVLine40(caller, uint64(y1), uint64(y2), uint64(x), uint64(hc))
	}

	if hc != 0 {
		this.Cost = int64(math.Abs(float64(y1-y2))) * (1000000000 / 800) / 2
	} else {
		this.Cost = 0
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandVLIN) Syntax() string {

	/* vars */
	var result string

	result = "VLIN y1,y2 AT x"

	/* enforce non void return */
	return result

}
