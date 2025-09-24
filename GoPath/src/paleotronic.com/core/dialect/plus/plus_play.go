package plus

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/utils"
)

type PlusPlay struct {
	dialect.CoreFunction
}

func (this *PlusPlay) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	filename := this.Stack.Shift().Content

	p := files.GetPath(filename)
	f := files.GetFilename(filename)

	wait := false
	if this.Stack.Size() > 0 {
		wait = this.Stack.Shift().AsInteger() != 0
	}

	if strings.HasSuffix(filename, ".rst") {
		this.Interpreter.GetMemoryMap().IntSetRestalgiaPath(this.Interpreter.GetMemIndex(), filename, wait)
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
		return nil
	}

	_, err := apple2helpers.PlayAudio(this.Interpreter, p, f, wait)
	if err != nil {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
	}

	return nil
}

func (this *PlusPlay) Syntax() string {

	/* vars */
	var result string

	result = "PLAY{soundfile}"

	/* enforce non void return */
	return result

}

func (this *PlusPlay) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusPlay(a int, b int, params types.TokenList) *PlusPlay {
	this := &PlusPlay{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PLAY"
	this.MinParams = 1
	this.MaxParams = 2

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "file", Default: *types.NewToken(types.STRING, "")},
			dialect.FunctionParamDef{Name: "block", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}
