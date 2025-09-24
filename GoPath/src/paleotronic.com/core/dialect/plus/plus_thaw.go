package plus

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
)

type PlusThaw struct {
	dialect.CoreFunction
}

func (this *PlusThaw) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slot := this.Interpreter.GetMemIndex()

	if !this.Query {
		t := this.ValueMap["path"]
		c := t.Content
		t = this.ValueMap["slot"]
		if t.AsInteger() != 0 {
			slot = t.AsInteger() - 1
		}

		// parts := strings.Split(c, "local/")
		// fullpath := files.GetUserDirectory(files.BASEDIR + "/" + parts[len(parts)-1])
		if !strings.HasPrefix(c, "/") {
			c = "/" + strings.Trim(this.Interpreter.GetWorkDir(), "/") + "/" + c
		}

		data, err := files.ReadBytesViaProvider(files.GetPath(c), files.GetFilename(c))

		if err == nil {
			settings.PureBootRestoreStateBin[slot] = data.Content
			this.Interpreter.GetMemoryMap().IntSetSlotRestart(slot, true)
			//log.Printf("Request restart in slot %d", slot)
		}

	} else {
		this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetWorkDir()))
	}

	return nil
}

func (this *PlusThaw) Syntax() string {

	/* vars */
	var result string

	result = "CD{path}"

	/* enforce non void return */
	return result

}

func (this *PlusThaw) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusThaw(a int, b int, params types.TokenList) *PlusThaw {
	this := &PlusThaw{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CD"

	this.NamedParams = []string{"path", "slot"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, "~"),
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}
