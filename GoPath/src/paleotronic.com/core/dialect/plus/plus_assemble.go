package plus

import (
	"errors"
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/cpu/mos6502/asm"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/utils"
)

type PlusAssemble struct {
	dialect.CoreFunction
}

func (this *PlusAssemble) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		fmt.Println(e)
		return e
	}

	if !this.Query {

		input := this.ValueMap["input"].Content
		output := this.ValueMap["output"].Content
		cpu := this.ValueMap["cpu"].Content

		if input == "" {
			this.Interpreter.PutStr("input file needed")
			return errors.New("input file needed")
		}

		var a *asm.Asm6502

		switch strings.ToUpper(cpu) {
		case "6502":
			a = asm.NewAsm6502()
		case "65C02":
			a = asm.NewAsm65C02()
		default:
			this.Interpreter.PutStr("Unknown CPU: " + cpu + "\r\n")
			return errors.New("Unknown CPU: " + cpu)
		}

		if output == "out" {
			output = strings.Replace(input, ".asm", "", -1)
			output = strings.Replace(output, ".ASM", "", -1)
		}

		raw, e := files.ReadBytesViaProvider(files.GetPath(input), files.GetFilename(input))
		if e != nil {
			return e
		}

		lines := strings.Split(string(raw.Content), "\r\n")

		codeblocks, lno, line, e := a.AssembleMultipass(
			lines,
			0x4000,
		)

		if e != nil {
			this.Interpreter.PutStr(fmt.Sprintf("Assembly failed at line %d:\r\n", lno))
			this.Interpreter.PutStr("  " + line + "\r\n")
			return e
		} else {
			if output == "" {
				a.DumpFilesRAM(codeblocks, this.Interpreter)
			} else {
				a.DumpFilesNFS(output, codeblocks, this.Interpreter)
			}
			a.DumpSyms(this.Interpreter)
		}

	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusAssemble) Syntax() string {

	/* vars */
	var result string

	result = "SETCLASSICHGR{}"

	/* enforce non void return */
	return result

}

func (this *PlusAssemble) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusAssemble(a int, b int, params types.TokenList) *PlusAssemble {
	this := &PlusAssemble{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ASM"

	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, ""),
		*types.NewToken(types.STRING, ""),
		*types.NewToken(types.STRING, "6502"),
	}
	this.NamedParams = []string{"input", "output", "cpu"}
	this.Raw = true

	this.MinParams = 1
	this.MaxParams = 10

	return this
}
