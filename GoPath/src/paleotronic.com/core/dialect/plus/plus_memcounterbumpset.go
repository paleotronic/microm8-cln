package plus

import (
	"strings"

	"paleotronic.com/fmt"

	"paleotronic.com/core/hardware/apple2helpers"

	"paleotronic.com/core/hardware/cpu/mos6502"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusMemCounterBumpSet struct {
	dialect.CoreFunction
}

func (this *PlusMemCounterBumpSet) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var mode string
	var addr int
	var counter int

	mode = strings.ToLower(params.Shift().Content)
	addr = params.Shift().AsInteger()
	counter = params.Shift().AsInteger()

	var f mos6502.AddressFlags = 0
	switch mode {
	case "exec":
		f = mos6502.AF_EXEC_BUMP
	case "write":
		f = mos6502.AF_WRITE_BUMP
	case "read":
		f = mos6502.AF_READ_BUMP
	}

	fmt.Print(mode, "bump counter ")

	cpu := apple2helpers.GetCPU(this.Interpreter)
	v := cpu.SpecialFlag[addr]
	if counter >= 0 && counter < 255 {
		// set
		v |= (f | (mos6502.AddressFlags(counter) << 32))
		fmt.Println("set", counter)
	} else {
		// clear
		if v != 0 {
			v &= (0xffffffff ^ f)
		}
		fmt.Println("clear")
	}
	cpu.SpecialFlag[addr] = v
	cpu.HasSpecialFlags = true

	this.Interpreter.GetMemoryMap().IntBumpCounter(this.Interpreter.GetMemIndex(), counter)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(counter))))

	return nil
}

func (this *PlusMemCounterBumpSet) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusMemCounterBumpSet) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusMemCounterBumpSet(a int, b int, params types.TokenList) *PlusMemCounterBumpSet {
	this := &PlusMemCounterBumpSet{}

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 3
	this.MaxParams = 3

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "mode", Default: *types.NewToken(types.STRING, "")},
			dialect.FunctionParamDef{Name: "address", Default: *types.NewToken(types.NUMBER, "1")},
			dialect.FunctionParamDef{Name: "counter", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}
