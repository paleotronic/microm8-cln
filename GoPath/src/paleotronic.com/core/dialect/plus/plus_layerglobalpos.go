package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/fmt"
)

type PlusLayerGlobalPos struct {
	dialect.CoreFunction
}

func (this *PlusLayerGlobalPos) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		xt := this.ValueMap["x"]
		x := xt.AsExtended()
		yt := this.ValueMap["y"]
		y := yt.AsExtended()
		//zt := this.ValueMap["z"]; z := zt.AsInteger()
		st := this.ValueMap["slotid"]
		sid := st.AsInteger() - 1

		// assume percents
		//x = int(float32(x)/100 * 1706 - 213)
		//y = int(float32(y)/100 * 960 - 8)
		//z = 0

		xpc := float64(x-12.5) / 100
		ypc := float64(y) / 100

		ent := this.Interpreter

		if sid != -1 {
			//fmt.Println("=========================================================================")
			//fmt.Println("slotid =", sid)
			if sid >= 0 && sid < this.Interpreter.GetProducer().GetNumInterpreters() {
				ent = this.Interpreter.GetProducer().GetInterpreter(sid)
				//fmt.Println("Got interpreter remote")
			}
		}

		ent.GetProducer().SetMasterLayerPos(ent.GetMemIndex(), xpc, ypc)

		//fmt.Println(ent.GetMemIndex())

		//if layer == "*" {
		//	ent.PositionAllLayers(int16(x), int16(y), int16(z))
		//} else {
		//	if !ent.HUDLayerSetPos(layer, int16(x), int16(y), int16(z)) {
		//		ent.GFXLayerSetPos(layer, int16(x), int16(y), int16(z))
		//	}
		//}
	}

	this.Stack.Push(types.NewToken(types.STRING, "1"))

	return nil
}

func (this *PlusLayerGlobalPos) Syntax() string {

	/* vars */
	var result string

	result = "LAYER.MASTERPOS{id,x,y,z}"

	/* enforce non void return */
	return result

}

func (this *PlusLayerGlobalPos) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusLayerGlobalPos(a int, b int, params types.TokenList) *PlusLayerGlobalPos {
	this := &PlusLayerGlobalPos{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "LAYER.MASTERPOS"

	this.NamedParams = []string{"slotid", "x", "y", "z"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.INTEGER, "-1"),
		*types.NewToken(types.STRING, "13"),
		*types.NewToken(types.STRING, "0"),
		*types.NewToken(types.STRING, "0"),
	}
	this.Raw = true

	return this
}
