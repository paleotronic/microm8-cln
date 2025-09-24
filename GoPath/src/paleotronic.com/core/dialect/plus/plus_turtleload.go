package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/runestring"
	"strings"
)

type PlusTurtleLoad struct {
	dialect.CoreFunction
}

func (this *PlusTurtleLoad) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	filename := this.ValueMap["file"].Content
	if filename != "" {
		if this.Interpreter.GetDialect().GetShortName() == "logo" {

			data, e := files.ReadBytesViaProvider(files.GetPath(strings.ToLower(filename)), files.GetFilename(strings.ToLower(filename)))
			if e != nil {
				return e
			}

			lines := strings.Split(string(data.Content), "\r\n")
			for _, l := range lines {
				if l != "" {
					tl := this.Interpreter.GetDialect().Tokenize(runestring.Cast(l))
					scope := this.Interpreter.GetDirectAlgorithm()
					this.Interpreter.GetDialect().SetSilentDefines(true)
					this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, scope, this.Interpreter.GetLPC())
					this.Interpreter.GetDialect().SetSilentDefines(false)
					this.Interpreter.SaveCPOS()
					this.Interpreter.SetNeedsPrompt(true)
				}
			}
			for _, l := range lines {
				if l != "" {
					tl := this.Interpreter.GetDialect().Tokenize(runestring.Cast(l))
					scope := this.Interpreter.GetDirectAlgorithm()
					this.Interpreter.GetDialect().SetSilentDefines(true)
					this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, scope, this.Interpreter.GetLPC())
					this.Interpreter.GetDialect().SetSilentDefines(false)
					this.Interpreter.SaveCPOS()
					this.Interpreter.SetNeedsPrompt(true)
				}
			}

		}
	}

	return nil
}

func (this *PlusTurtleLoad) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusTurtleLoad) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusTurtleLoad(a int, b int, params types.TokenList) *PlusTurtleLoad {
	this := &PlusTurtleLoad{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"file"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, ""),
	}
	this.Raw = true

	return this
}
