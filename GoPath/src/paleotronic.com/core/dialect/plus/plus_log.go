package plus

import (
	"log"
	"paleotronic.com/core/dialect" //	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"   //	"paleotronic.com/utils"
)

type PlusLog struct {
	dialect.CoreFunction
}

func (this *PlusLog) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	s := ""
	for _, p := range this.Stack.Content {
		if s != "" {
			s += " "
		}
		s += p.Content
	}
	log.Printf("vm#%d: %s", this.Interpreter.GetMemIndex()+1, s)

	return nil

}

func (this *PlusLog) Syntax() string {

	/* vars */
	var result string

	result = "ALTCASE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusLog) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusLog(a int, b int, params types.TokenList) *PlusLog {
	this := &PlusLog{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"

	// this.NamedParams = []string{"mode"}
	// this.NamedDefaults = []types.Token{*types.NewToken(types.INTEGER, "0")}
	// this.Raw = true
	this.MinParams = 1
	this.MaxParams = 20

	return this
}
