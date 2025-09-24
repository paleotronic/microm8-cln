package plus

import (
	"errors"
	"io/ioutil"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/files"

	//	"paleotronic.com/core/memory"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusVMRestore struct {
	dialect.CoreFunction
}

func (this *PlusVMRestore) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	st := params.Shift()
	slotid := st.AsInteger() - 1
	freeze := params.Shift().Content

	if slotid < 0 || slotid >= memory.OCTALYZER_NUM_INTERPRETERS {
		return errors.New("invalid slot")
	}

	// restore freeze
	data, err := ioutil.ReadFile(freeze)
	if err != nil {
		fp, err := files.ReadBytesViaProvider(files.GetPath(freeze), files.GetFilename(freeze))
		if err != nil {
			return err
		}
		data = fp.Content
	}
	settings.PureBootRestoreStateBin[slotid] = data
	this.Interpreter.GetMemoryMap().IntSetSlotRestart(slotid, true)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusVMRestore) Syntax() string {

	/* vars */
	var result string

	result = "SETCLASSICCPU{}"

	/* enforce non void return */
	return result

}

func (this *PlusVMRestore) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusVMRestore(a int, b int, params types.TokenList) *PlusVMRestore {
	this := &PlusVMRestore{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SETCLASSICCPU"

	this.MaxParams = 2
	this.MinParams = 2

	return this
}
