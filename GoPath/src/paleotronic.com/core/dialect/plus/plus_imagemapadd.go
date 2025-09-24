package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	"paleotronic.com/runestring"
)

type PlusImageMapAdd struct {
	dialect.CoreFunction
}

func (this *PlusImageMapAdd) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	rstr := this.Stack.Shift().Content
	fstr := this.Stack.Shift().Content
	
	mm := types.NewInlineImageManager( this.Interpreter.GetMemIndex(), this.Interpreter.GetMemoryMap() )
	
	rs := runestring.Cast(rstr)
	ch := rs.Runes[0]
	
	mm.Add( ch, fstr )

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusImageMapAdd) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusImageMapAdd) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusImageMapAdd(a int, b int, params types.TokenList) *PlusImageMapAdd {
	this := &PlusImageMapAdd{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}
