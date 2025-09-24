package plus

import (
	"bytes"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
)

var PNGColorSpace byte

type PlusPNG2HGR struct {
	dialect.CoreFunction
}

func (this *PlusPNG2HGR) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	/*
	 * 0 = SuperFrog
	 * 1 = Atkinson
	 * 2 = Burkes
	 * 3 = Stucki
	 * 4 = FloydSteinberg
	 * 5 = Sierra3Row
	 * 6 = JaJuNi
	 * 7 = Bayer4x4
	 * -1 = no dither
	 */
	index := this.Interpreter.GetMemIndex()
	rect := settings.ImageDrawRect[index]

	if !this.Query {
		filename := this.ValueMap["image"].Content
		g := this.ValueMap["brightness"]
		brightness := float32(g.AsExtended())
		pc := this.ValueMap["perceptual"]
		md := this.ValueMap["method"]
		perceptual := (pc.AsInteger() != 0)
		method := md.AsInteger() % 8

		p := files.GetPath(filename)
		f := files.GetFilename(filename)

		if files.ExistsViaProvider(p, f) {
			b, e := files.ReadBytesViaProvider(p, f)

			if e != nil {
				this.Interpreter.PutStr("I/O Error")
				return e
			}

			bb := bytes.NewBuffer(b.Content)

			switch method {
			case -1:
				apple2helpers.HGRDither(this.Interpreter, bb, []int{}, brightness, apple2helpers.None, perceptual, rect)
			case 0:
				apple2helpers.HGRDither(this.Interpreter, bb, []int{}, brightness, apple2helpers.SuperFrog, perceptual, rect)
			case 1:
				apple2helpers.HGRDither(this.Interpreter, bb, []int{}, brightness, apple2helpers.Atkinson, perceptual, rect)
			case 2:
				apple2helpers.HGRDither(this.Interpreter, bb, []int{}, brightness, apple2helpers.Burkes, perceptual, rect)
			case 3:
				apple2helpers.HGRDither(this.Interpreter, bb, []int{}, brightness, apple2helpers.Stucki, perceptual, rect)
			case 4:
				apple2helpers.HGRDither(this.Interpreter, bb, []int{}, brightness, apple2helpers.FloydSteinberg, perceptual, rect)
			case 5:
				apple2helpers.HGRDither(this.Interpreter, bb, []int{}, brightness, apple2helpers.Sierra3Row, perceptual, rect)
			case 6:
				apple2helpers.HGRDither(this.Interpreter, bb, []int{}, brightness, apple2helpers.JaJuNi, perceptual, rect)
			case 7:
				apple2helpers.HGRDitherBayer4x4(this.Interpreter, bb, []int{}, brightness, perceptual)
			}

		}

	}

	this.Stack.Push(types.NewToken(types.STRING, ""))

	return nil
}

func (this *PlusPNG2HGR) Syntax() string {

	/* vars */
	var result string

	result = "BACKDROP{image}"

	/* enforce non void return */
	return result

}

func (this *PlusPNG2HGR) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusPNG2HGR(a int, b int, params types.TokenList) *PlusPNG2HGR {
	this := &PlusPNG2HGR{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PNG2HGR"

	this.NamedParams = []string{"image", "brightness", "method", "perceptual"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, ""),
		*types.NewToken(types.NUMBER, "0.997"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}

func NewPlusPNG2HGRH(a int, b int, params types.TokenList) *PlusPNG2HGR {
	this := &PlusPNG2HGR{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PNG2HGR"
	this.Hidden = true

	this.NamedParams = []string{"image", "brightness", "method", "perceptual"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, ""),
		*types.NewToken(types.NUMBER, "0.997"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}
