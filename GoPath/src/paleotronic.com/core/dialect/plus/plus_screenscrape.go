package plus

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"

	"paleotronic.com/log"
)

type PlusScreenScrape struct {
	dialect.CoreFunction
}

func (this *PlusScreenScrape) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var str = strings.ToLower(this.ValueMap["string"].Content)
	x := this.ValueMap["line"]
	var line = x.AsInteger()

	lines := this.CopyText()

	for i, v := range lines {
		log.Printf("line %.2d: %s", i, v)
	}

	s := 0
	e := 23
	if line != -1 {
		s = line
		e = line
	}

	for l := s; l <= e; l++ {
		ss := strings.ToLower(lines[l])
		if strings.Contains(ss, str) {
			this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
			return nil
		}
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusScreenScrape) CopyText() []string {

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

func (this *PlusScreenScrape) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusScreenScrape) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusScreenScrape(a int, b int, params types.TokenList) *PlusScreenScrape {
	this := &PlusScreenScrape{}

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 2
	this.MaxParams = 2
	this.Raw = true

	this.NamedParams = []string{
		"string",
		"line",
	}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, ""),
		*types.NewToken(types.NUMBER, "-1"),
	}

	return this
}
