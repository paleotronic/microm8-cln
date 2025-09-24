package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandPLOT struct {
	dialect.Command
}

func (this *StandardCommandPLOT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var tla types.TokenListArray
	var tl types.TokenList
	var x int
	var y int

	result = 0

	tla = caller.SplitOnTokenWithBrackets(tokens, *types.NewToken(types.SEPARATOR, ","))
	tl = *types.NewTokenList()

	for _, i1 := range tla {
		vtok := caller.ParseTokensForResult(i1)
		tl.Push(&vtok)
	}

	if tl.Size() != 2 && tl.Size() != 3 {
		return result, exception.NewESyntaxError("PLOT expects 2 or 3 parameters")
	}

	//if ((tl.Get(0).Type != types.NUMBER) || (tl.Get(0).Type != types.NUMBER))
	//  return types.NewESyntaxError("PLOT expects numbers");

	hc := apple2helpers.GetCOLOR(caller)
	x = tl.Get(0).AsInteger()
	y = tl.Get(1).AsInteger()
	z := 0
	if tl.Size() >= 3 {
		z = tl.Get(2).AsInteger()
	}

	//caller.GetVDU().GrPlot(x, y, vtok.AsInteger())

	modes := apple2helpers.GetActiveVideoModes(caller)

	//fmt.Printf("modes = %v", modes)

	if len(modes) == 1 {
		switch modes[0] {
		case "LOGR":
			apple2helpers.LOGRPlot40(caller, uint64(x), uint64(y), uint64(hc))
		case "DLGR":
			apple2helpers.LOGRPlot80(caller, uint64(x), uint64(y), uint64(hc))
		case "CUBE":
			apple2helpers.CUBE(caller).Plot(uint8(x), uint8(y), uint8(z), uint8(hc))
		}
	}
	//System.Out.Println("PLOT <================================================> SetPixel at x = "+x+", y = "+y+" is "+vtok.AsInteger());

	if hc != 0 {
		this.Cost = (1000000000 / 800)
	} else {
		this.Cost = 0
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandPLOT) Syntax() string {

	/* vars */
	var result string

	result = "PLOT <line>"

	/* enforce non void return */
	return result

}
