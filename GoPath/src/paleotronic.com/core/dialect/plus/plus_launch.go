package plus

import (
	"strings"

	"paleotronic.com/files"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	//	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
	"fmt"
)

type PlusLaunch struct {
	dialect.CoreFunction
}

func (this *PlusLaunch) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		file := this.ValueMap["file"].Content
		fmt.Println(file)
		index := this.Interpreter.GetMemIndex()
		fr, err := files.ReadBytesViaProvider(files.GetPath(file), files.GetFilename(file))
		if err != nil {
			fmt.Println(err)
			return err
		}
		if strings.HasSuffix(file, ".frz") {
			settings.PureBootRestoreStateBin[index] = fr.Content
			this.Interpreter.GetMemoryMap().IntSetSlotRestart(index, true)
		}
	}

	return nil

}

func (this *PlusLaunch) Syntax() string {

	/* vars */
	var result string

	result = "ALTCASE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusLaunch) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusLaunch(a int, b int, params types.TokenList) *PlusLaunch {
	this := &PlusLaunch{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"

	this.NamedParams = []string{"file"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.STRING, "")}
	this.Raw = true

	return this
}
