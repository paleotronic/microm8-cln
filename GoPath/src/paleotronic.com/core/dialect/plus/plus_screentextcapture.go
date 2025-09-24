package plus

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"

	"paleotronic.com/log"
)

type PlusScreenTextCapture struct {
	dialect.CoreFunction
}

func (this *PlusScreenTextCapture) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	x := this.ValueMap["line"]
	var line = x.AsInteger()
	s := this.ValueMap["start"]
	var start = s.AsInteger()
	e := this.ValueMap["end"]
	var end = e.AsInteger()

	if line == -1 {
		this.Stack.Push(types.NewToken(types.STRING, ""))
		return nil
	}

	lines := this.CopyText()

	for i, v := range lines {
		log.Printf("line %.2d: %s", i, v)
	}

	text := lines[line]
	outtext := ""
	for i, v := range text {
		if i >= start && i <= end {
			outtext += string(v)
		}
	}

	this.Stack.Push(types.NewToken(types.STRING, strings.Trim(outtext, " ")))

	return nil
}

func (this *PlusScreenTextCapture) CopyText() []string {

	//_, textlayers := apple2helpers.GetActiveLayers(this.Interpreter)
	lines := []string(nil)

	l := apple2helpers.GETHUD(this.Interpreter, "TEXT")
	lines = l.Control.GetStrings()

	tmp := make([]string, 0)
	for i, v := range lines {
		if i%2 == 0 {
			tmp = append(tmp, v)
		}
	}

	return tmp
}

func (this *PlusScreenTextCapture) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusScreenTextCapture) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusScreenTextCapture(a int, b int, params types.TokenList) *PlusScreenTextCapture {
	this := &PlusScreenTextCapture{}

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 2
	this.MaxParams = 2
	this.Raw = true

	this.NamedParams = []string{
		"line",
		"start",
		"end",
	}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "-1"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "79"),
	}

	return this
}
