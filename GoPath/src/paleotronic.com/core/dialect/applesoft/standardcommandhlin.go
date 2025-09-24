package applesoft

import (
	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandHLIN struct {
	dialect.Command
}

func (this *StandardCommandHLIN) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	//var vtok types.Token
	//var atok types.Token
	//	var cr types.CodeRef
	var tla types.TokenListArray
	var tls types.TokenListArray
	//	var i, tl types.TokenList
	//TokenList  i;
	var x1 int
	var x2 int
	var y int
	//	var atpos int
	var v int

	result = 0

	tls = caller.GetDialect().SplitOnToken(tokens, *types.NewToken(types.KEYWORD, "at"))

	//caller.PutStr(IntToStr(tls.Size())+PasUtil.CRLF);

	if tls.Size() != 2 {
		return result, exception.NewESyntaxError("SYNTAX ERROR - tls Size()")
	}

	tla = caller.SplitOnTokenWithBrackets(*tls.Get(0), *types.NewToken(types.SEPARATOR, ","))
	tl := *types.NewTokenList()

	for _, i1 := range tla {
		vtok := caller.ParseTokensForResult(i1)
		tl.Push(&vtok)
	}

	if tl.Size() != 2 {
		return result, exception.NewESyntaxError("HLIN expects 2 parameters")
	}

	if (tl.Get(0).Type != types.NUMBER) && (tl.Get(0).Type != types.INTEGER) {
		return result, exception.NewESyntaxError("HLIN expects numbers")
	}

	hc := apple2helpers.GetCOLOR(caller)
	atok := caller.ParseTokensForResult(*tls.Get(1))
	x1 = tl.Get(0).AsInteger()
	x2 = tl.Get(1).AsInteger()
	y = atok.AsInteger()
	//caller.GetVDU().Line( x1, y, x2, y, vtok.AsInteger() );
	if x1 > x2 {
		v = x1
		x1 = x2
		x2 = v
	}
	//	for v = x1; v <= x2; v++ {
	//		caller.GetVDU().GrPlot(v, y, vtok.AsInteger())
	//	}
	//caller.GetVDU().GrHorizLine(x1, x2, y, vtok.AsInteger())
	// if apple2helpers.GetColumns(caller) == 80 {
	// 	apple2helpers.LOGRHLine80(caller, uint64(x1), uint64(x2), uint64(y), uint64(hc))
	// } else {
	// 	apple2helpers.LOGRHLine40(caller, uint64(x1), uint64(x2), uint64(y), uint64(hc))
	// }

	modes := apple2helpers.GetActiveVideoModes(caller)

	//fmt.Printf("modes = %v", modes)

	if len(modes) == 1 {
		switch modes[0] {
		case "LOGR", "LGR2":
			apple2helpers.LOGRHLine40(caller, uint64(x1), uint64(x2), uint64(y), uint64(hc))
		case "DLGR":
			apple2helpers.LOGRHLine80(caller, uint64(x1), uint64(x2), uint64(y), uint64(hc))
		case "CUBE":
			apple2helpers.CUBE(caller).Bresenham3D(
				int(x1), int(y), 0,
				int(x2), int(y), 0,
				uint8(hc),
			)
		}
	} else {
		apple2helpers.LOGRHLine40(caller, uint64(x1), uint64(x2), uint64(y), uint64(hc))
	}

	if hc != 0 {
		this.Cost = int64(math.Abs(float64(x1-x2))*(1000000000/800)) / 2
	} else {
		this.Cost = 0
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandHLIN) Syntax() string {

	/* vars */
	var result string

	result = "HLIN x1,x2 AT y"

	/* enforce non void return */
	return result

}
