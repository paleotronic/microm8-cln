package plus

import (
	//"paleotronic.com/files"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusMetaData struct {
	dialect.CoreFunction
}

func (this *PlusMetaData) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	name := this.Stack.Shift().Content
	fr := this.Interpreter.GetFileRecord()
	p := &fr
	value := p.GetMeta(name, "")

	this.Stack.Push(types.NewToken(types.STRING, value))

	return nil
}

func (this *PlusMetaData) Syntax() string {

	/* vars */
	var result string

	result = "METADATA{NAME,VALUE}"

	/* enforce non void return */
	return result

}

func (this *PlusMetaData) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusMetaData(a int, b int, params types.TokenList) *PlusMetaData {
	this := &PlusMetaData{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "METADATA"

	//	this.NamedParams = []string{"name", "value"}
	//	this.NamedDefaults = []types.Token{
	//		*types.NewToken(types.STRING, ""),
	//		*types.NewToken(types.STRING, ""),
	//	}
	//	this.Raw = true

	return this
}
