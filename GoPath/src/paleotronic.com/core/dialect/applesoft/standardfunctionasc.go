package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
    "unicode/utf8"
)

type StandardFunctionASC struct {
	dialect.CoreFunction
}

func (this *StandardFunctionASC) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value string

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	value = this.Stack.Pop().Content

	//if ((value.Size() > 0) && (this.Interpreter.Dialect.Title == "INTEGER"));
	//   this.Stack.Push( types.NewToken( types.NUMBER, IntToStr((Ord(value.Get(1)) && 127)+128) ) );
	//else
	if len(value) > 0 {
        r, _ := utf8.DecodeRuneInString(value)

		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(r))))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))
	}

	return nil
}

func (this *StandardFunctionASC) Syntax() string {

	/* vars */
	var result string

	result = "ASC(<string>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionASC) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionASC(a int, b int, params types.TokenList) *StandardFunctionASC {
	this := &StandardFunctionASC{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ASC"

	return this
}
