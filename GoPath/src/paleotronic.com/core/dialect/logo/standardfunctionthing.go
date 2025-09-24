package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"

	"log"
)

type StandardFunctionTHING struct {
	dialect.CoreFunction
	D interfaces.Dialecter
}

func NewStandardFunctionTHING(a int, b int, params types.TokenList, d interfaces.Dialecter) *StandardFunctionTHING {
	this := &StandardFunctionTHING{
		D: d,
	}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SIN"

	return this
}

func (this *StandardFunctionTHING) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value string

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	value = this.Stack.Pop().Content
	log.Printf("thing: value is %s", value)

	d := this.D.(*DialectLogo)
	v, _ := d.Driver.GetVar(value)
	if v == nil {
		v = types.NewToken(types.STRING, "")
	}
	log.Printf("value of thing %s is: %s", value, d.Driver.DumpObjectStruct(v, false, ""))

	this.Stack.Push(v.Copy())

	return nil
}

func (this *StandardFunctionTHING) Syntax() string {

	/* vars */
	var result string

	result = "SIN(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionTHING) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}
