package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/types"
	//"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/api"
	"paleotronic.com/filerecord"
	"paleotronic.com/utils"
	//"paleotronic.com/fmt"
	//"time"
	// "strings"
	//"github.com/atotto/clipboard"
)

type PlusBugLoad struct {
	dialect.CoreFunction
	bug  *filerecord.BugReport
	edit *editor.CoreEdit
}

func (this *PlusBugLoad) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {

		ct := this.ValueMap["id"]
		id := int64(ct.AsInteger())

		if id < 1 {
			this.Interpreter.PutStr("Please specify an id.")
			this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
			return nil
		}

		this.bug, _ = s8webclient.CONN.GetBugByID(id)

		if this.bug.DefectID != id {
			this.Interpreter.PutStr("Could not find bug with that id.")
			this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
			return nil
		}

		listing := -1
		for i, att := range this.bug.Attachments {
			if att.Name == "LISTING" {
				listing = i
			}
		}

		if listing == -1 {
			this.Interpreter.PutStr("No listing attached.")
			this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
			return nil
		}

		this.Interpreter.SetFeedBuffer(string(this.bug.Attachments[listing].Content))

		tl := types.NewTokenList()
		tl.Push(types.NewToken(types.STRING, "!"))

		a := this.Interpreter.GetDirectAlgorithm()

		this.Interpreter.GetDialect().GetCommands()["load"].Execute(
			nil,
			this.Interpreter,
			*tl,
			a,
			*this.Interpreter.GetLPC(),
		)

		this.Interpreter.PutStr("Loaded listing.\r\n")

	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusBugLoad) Syntax() string {

	/* vars */
	var result string

	result = "EXIT{}"

	/* enforce non void return */
	return result

}

func (this *PlusBugLoad) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusBugLoad(a int, b int, params types.TokenList) *PlusBugLoad {
	this := &PlusBugLoad{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "bug.show"
	this.Raw = true
	this.NamedParams = []string{"id"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
	}

	return this
}
