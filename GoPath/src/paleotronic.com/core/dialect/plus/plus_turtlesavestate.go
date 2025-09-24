package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/utils"
	"strings"
)

type PlusTurtleSaveState struct {
	dialect.CoreFunction
}

func (this *PlusTurtleSaveState) FunctionExecute(params *types.TokenList) error {

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

			data, err := apple2helpers.VECTOR(this.Interpreter).Turtle().ToJSON()
			if err != nil {
				return err
			}
			data = utils.XZBytes(data)

			err = files.WriteBytesViaProvider(files.GetPath(filename), files.GetFilename(filename), data)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func (this *PlusTurtleSaveState) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusTurtleSaveState) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusTurtleSaveState(a int, b int, params types.TokenList) *PlusTurtleSaveState {
	this := &PlusTurtleSaveState{}

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
