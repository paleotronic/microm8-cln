package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/utils"
	"strings"
)

type PlusTurtleLoadState struct {
	dialect.CoreFunction
}

func (this *PlusTurtleLoadState) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	filename := this.ValueMap["file"].Content

	if files.GetExt(filename) == "" {
		filename += ".tsf"
	}
	if !strings.HasPrefix(filename, "/") {
		filename = "/" + strings.Trim(this.Interpreter.GetWorkDir(), "/") + "/" + strings.Trim(filename, "/")
	}

	if filename != "" {
		if this.Interpreter.GetDialect().GetShortName() == "logo" {

			fr, err := files.ReadBytesViaProvider(files.GetPath(filename), files.GetFilename(filename))
			if err != nil {
				return err
			}
			fr.Content = utils.UnXZBytes(fr.Content)

			err = apple2helpers.VECTOR(this.Interpreter).Turtle().FromJSON(fr.Content)
			apple2helpers.VECTOR(this.Interpreter).Render()
			return err

		}
	}

	return nil
}

func (this *PlusTurtleLoadState) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusTurtleLoadState) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusTurtleLoadState(a int, b int, params types.TokenList) *PlusTurtleLoadState {
	this := &PlusTurtleLoadState{}

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
