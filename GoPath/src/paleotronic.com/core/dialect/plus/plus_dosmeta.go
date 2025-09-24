package plus

import (
	"paleotronic.com/log"

	"paleotronic.com/files"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/api"
)

type PlusMetaMod struct {
	dialect.CoreFunction
}

func (this *PlusMetaMod) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {

		filename := this.ValueMap["filename"].Content
		if filename == "" {
			return nil
		}

		delete(this.ValueMap, "filename")

		p, f := files.GetPath(filename), files.GetFilename(filename)

		// does file exist?
		fr, e := files.ReadBytesViaProvider(p, f)
		if e != nil {
			return e
		}

		log.Printf("Current MetaData: %v\n", fr.MetaData)

		// file exists... build new meta data
		for n, v := range this.ValueMap {
			fr.AddMeta(n, v.Content)
		}

		log.Printf("New MetaData: %v\n", fr.MetaData)

		files.MetaUpdateViaProvider(p, f, fr.MetaData)

	} else {
		this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetWorkDir()))
	}

	return nil
}

func (this *PlusMetaMod) Syntax() string {

	/* vars */
	var result string

	result = "METAMOD{filename, name=value, name=value}"

	/* enforce non void return */
	return result

}

func (this *PlusMetaMod) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusMetaMod(a int, b int, params types.TokenList) *PlusMetaMod {
	this := &PlusMetaMod{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "METAMOD"

	this.NamedParams = []string{"filename"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.STRING, "")}
	this.Raw = true

	return this
}
