package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//"paleotronic.com/core/hardware/apple2helpers"
	//"paleotronic.com/utils"
	"strings"
)

type PlusTextWrap struct {
	dialect.CoreFunction
}

func wrap(text string, lineWidth int, start int) (wrapped string) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}
	wrapped = words[0]
	spaceLeft := lineWidth - len(wrapped) - start
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			wrapped += "\r\n" + word
			spaceLeft = lineWidth - len(word)
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + len(word)
		}
	}
	return
}

func (this *PlusTextWrap) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	text := params.Shift().Content
	width := params.Shift().AsInteger()
	start := 0
	if params.Size() > 0 {
		start = params.Shift().AsInteger()
	}

	ntext := wrap(text, width, start)

	this.Stack.Push(types.NewToken(types.STRING, ntext))

	return nil
}

func (this *PlusTextWrap) Syntax() string {

	/* vars */
	var result string

	result = "@text.wrap{text,col}"

	/* enforce non void return */
	return result

}

func (this *PlusTextWrap) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusTextWrap(a int, b int, params types.TokenList) *PlusTextWrap {
	this := &PlusTextWrap{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "@output.wrap"

	//this.NamedParams = []string{ "color" }
	//this.NamedDefaults = []types.Token{ *types.NewToken( types.INTEGER, "15" ) }
	//this.Raw = true
	this.MinParams = 2
	this.MaxParams = 3

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "text", Default: *types.NewToken(types.STRING, "")},
			dialect.FunctionParamDef{Name: "width", Default: *types.NewToken(types.NUMBER, "40")},
			dialect.FunctionParamDef{Name: "start", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}
