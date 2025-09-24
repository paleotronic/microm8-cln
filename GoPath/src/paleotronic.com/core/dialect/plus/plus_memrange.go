package plus

import (
	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusMemRange struct {
	dialect.CoreFunction
}

func (this *PlusMemRange) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	base := this.Stack.Shift().AsInteger()

	c := this.Stack.Shift()

	count := c.AsInteger()

	fmt.Printf("memrange count = %s\n", c.Content)

	value := uint64(this.Stack.Shift().AsInteger())

	for i := base; i < base+count; i++ {
		v := this.Interpreter.GetMemory(i)
		if v == value {
			this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
			return nil
		}
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusMemRange) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusMemRange) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusMemRange(a int, b int, params types.TokenList) *PlusMemRange {
	this := &PlusMemRange{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 3
	this.MaxParams = 3

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "base", Default: *types.NewToken(types.NUMBER, "4096")},
			dialect.FunctionParamDef{Name: "count", Default: *types.NewToken(types.NUMBER, "1024")},
		},
	)

	return this
}
