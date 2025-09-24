package plus

import (
	"paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/filerecord"
	"paleotronic.com/utils"
	//"paleotronic.com/fmt"
	"time"
)

type PlusBugClose struct {
	dialect.CoreFunction
}

func (this *PlusBugClose) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {

		ct := this.ValueMap["id"]
		id := int64(ct.AsInteger())
		comment := this.ValueMap["comment"].Content

		if id < 1 {
			this.Interpreter.PutStr("Please specify an id.")
			this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
			return nil
		}

		bug, _ := s8webclient.CONN.GetBugByID(id)

		if bug.DefectID != id {
			this.Interpreter.PutStr("Could not find bug with that id.")
			this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
			return nil
		}

		if comment == "" {
			this.Interpreter.PutStr("You can't close a bug with a comment.")
			this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
			return nil
		}

		bug.Comments = append(
			bug.Comments,
			filerecord.BugComment{
				User:    s8webclient.CONN.Username,
				Created: time.Now(),
				Content: comment,
			},
		)

		bug.State = filerecord.BS_CLOSED

		_ = s8webclient.CONN.CreateUpdateBug(*bug)

	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusBugClose) Syntax() string {

	/* vars */
	var result string

	result = "EXIT{}"

	/* enforce non void return */
	return result

}

func (this *PlusBugClose) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusBugClose(a int, b int, params types.TokenList) *PlusBugClose {
	this := &PlusBugClose{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "bug.close"
	this.Raw = true
	this.NamedParams = []string{"id", "comment"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.STRING, ""),
	}

	return this
}
