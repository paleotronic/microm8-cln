package plus

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/utils"
)

type PlusBGMusic struct {
	dialect.CoreFunction
}

func (this *PlusBGMusic) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	filename := this.ValueMap["file"].Content
	a := this.ValueMap["leadin"]
	leadin := a.AsInteger()
	b := this.ValueMap["fadein"]
	fadein := b.AsInteger()

	if filename == "" {
		this.Interpreter.StopMusic()
		return nil
	}

	if !strings.HasPrefix(filename, "/") && this.Interpreter.GetWorkDir() != "" {
		filename = strings.Trim(this.Interpreter.GetWorkDir(), "/") + "/" + filename
	}

	p := files.GetPath(filename)
	f := files.GetFilename(filename)

	err := this.Interpreter.PlayMusic(p, f, leadin, fadein)
	if err != nil {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
	}

	return nil
}

func (this *PlusBGMusic) Syntax() string {

	/* vars */
	var result string

	result = "PLAY{soundfile}"

	/* enforce non void return */
	return result

}

func (this *PlusBGMusic) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusBGMusic(a int, b int, params types.TokenList) *PlusBGMusic {
	this := &PlusBGMusic{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PLAY"
	this.MinParams = 1
	this.MaxParams = 3
	this.NamedParams = []string{"file", "leadin", "fadein"}
	this.Raw = true
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, ""),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
	}

	return this
}
