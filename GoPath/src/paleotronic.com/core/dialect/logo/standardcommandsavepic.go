package logo

import (
	"errors"
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/utils"
)

type StandardCommandSAVEPIC struct {
	dialect.Command
}

func (this *StandardCommandSAVEPIC) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	if tokens.Size() < 1 {
		return result, errors.New("I NEED A VALUE")
	}

	// get result
	tt, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	// save
	filename := tt.Content
	if files.GetExt(filename) == "" {
		filename += ".tsf"
	}
	if !strings.HasPrefix(filename, "/") {
		filename = "/" + strings.Trim(caller.GetWorkDir(), "/") + "/" + strings.Trim(filename, "/")
	}

	data, err := apple2helpers.VECTOR(caller).GetTurtle( this.Command.D.(*DialectLogo).Driver.GetTurtle() ).ToJSON()
	if err != nil {
		return 0, err
	}
	data = utils.XZBytes(data)

	err = files.WriteBytesViaProvider(files.GetPath(filename), files.GetFilename(filename), data)
	if err != nil {
		return 0, err
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandSAVEPIC) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
