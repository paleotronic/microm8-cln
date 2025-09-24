package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
    "paleotronic.com/filerecord"
	"paleotronic.com/api"
    //"paleotronic.com/fmt"
    "time"
)

type PlusBugCreate struct {
	dialect.CoreFunction
    Type filerecord.BugType
}

func (this *PlusBugCreate) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {

       summary := this.ValueMap["summary"].Content
       body    := this.ValueMap["body"].Content
       ct      := this.ValueMap["capture"]
       capture := (ct.AsInteger() != 0)

       if summary == "" {
       	  this.Interpreter.PutStr("Please include a summary at least")
          this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
          return nil
       }

       bug := filerecord.BugReport{
       	   Summary: summary,
           Body: body,
           Created: time.Now(),
           Creator: s8webclient.CONN.Username,
           Type: this.Type,
       }
       bug.Filename = this.Interpreter.GetFileRecord().FileName
       bug.Filepath = this.Interpreter.GetFileRecord().FilePath

       if capture && this.Type == filerecord.BT_BUG {
       	  tmp,_ := this.Interpreter.FreezeBytes()
       	  att := filerecord.BugAttachment{
          	  Created: time.Now(),
              Content: utils.GZIPBytes(tmp),
              Name: "Compressed run state",
          }
          attlist := filerecord.BugAttachment{
          	  Created: time.Now(),
              Content: utils.GZIPBytes( []byte(this.Interpreter.GetCode().String()) ),
              Name: "LISTING",
          }
          bug.Attachments = []filerecord.BugAttachment{ att, attlist }
       }

       _ = s8webclient.CONN.CreateUpdateBug( bug )

    }

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusBugCreate) Syntax() string {

	/* vars */
	var result string

	result = "EXIT{}"

	/* enforce non void return */
	return result

}

func (this *PlusBugCreate) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusBugCreate(a int, b int, params types.TokenList) *PlusBugCreate {
	this := &PlusBugCreate{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "bug.create"
    this.Raw  = true
    this.NamedParams = []string{ "summary", "body", "capture" }
    this.NamedDefaults = []types.Token{
         *types.NewToken( types.STRING, "" ),
         *types.NewToken( types.STRING, "" ),
         *types.NewToken( types.NUMBER, "1" ),
    }
    this.Type = filerecord.BugType(a)

	return this
}
