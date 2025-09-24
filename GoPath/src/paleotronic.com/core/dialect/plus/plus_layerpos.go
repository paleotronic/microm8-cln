package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusLayerPos struct {
	dialect.CoreFunction
}

func (this *PlusLayerPos) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		layer := this.ValueMap["id"].Content
		xt := this.ValueMap["x"]
		x := xt.AsExtended()
		yt := this.ValueMap["y"]
		y := yt.AsExtended()
		zt := this.ValueMap["z"]
		z := zt.AsExtended()
		st := this.ValueMap["slotid"]
		sid := st.AsInteger() - 1

		ent := this.Interpreter

		if sid != -1 {
			if sid >= 0 && sid < this.Interpreter.GetProducer().GetNumInterpreters() {
				ent = this.Interpreter.GetProducer().GetInterpreter(sid)
			}
		}

		if layer == "*" {
			ent.MoveAllLayers(x, y, z)
		} else {
			if !ent.HUDLayerMovePos(layer, x, y, z) {
				ent.GFXLayerMovePos(layer, x, y, z)
			}
		}
	}

	this.Stack.Push(types.NewToken(types.STRING, "1"))

	return nil
}

func (this *PlusLayerPos) Syntax() string {

	/* vars */
	var result string

	result = "LAYER.POS{id,x,y,z}"

	/* enforce non void return */
	return result

}

func (this *PlusLayerPos) FunctionParams() []types.TokenType {

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

func NewPlusLayerPos(a int, b int, params types.TokenList) *PlusLayerPos {
	this := &PlusLayerPos{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "LAYER.POS"

	this.NamedParams = []string{"slotid", "id", "x", "y", "z"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.INTEGER, "-1"),
		*types.NewToken(types.STRING, "*"),
		*types.NewToken(types.STRING, "0"),
		*types.NewToken(types.STRING, "0"),
		*types.NewToken(types.STRING, "0"),
	}
	this.Raw = true

	return this
}
