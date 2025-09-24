package applesoft

import (
	"time"

	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionPEEK struct {
	dialect.CoreFunction
	a int
	//functionExecute void
	c types.CodeRef
}

func NewStandardFunctionPEEK() *StandardFunctionPEEK {

	this := &StandardFunctionPEEK{}

	/* vars */
	this.CoreFunction = *dialect.NewCoreFunction(0, 0, *types.NewTokenList())
	this.Name = "PEEK"

	return this
}

func (this *StandardFunctionPEEK) FunctionExecute(params *types.TokenList) error {
	/* vars */
	var addr int
	var r int
	//var s string

	for _, vtok1 := range params.Content {
		if vtok1.Type == types.INTEGER {
			vtok1.Type = types.NUMBER
		}
	}

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	addr = this.Stack.Pop().AsInteger()

	if addr < 0 {
		addr = 65536 + addr
	}

	if addr >= 32 && addr <= 37 {
		txt := apple2helpers.GETHUD(this.Interpreter, "TEXT")
		if txt != nil {
			this.Interpreter.TBChange(txt.Control)
		}
	}

	r = 0
	if (addr == 2053) && (utils.Pos("INTEGER", this.Interpreter.GetDialect().GetTitle()) > 0) {
		//if (this.Interpreter.Dialect.Title.Contains("INTEGER")) {
		//System.Err.Println("Is integer");
		if this.Interpreter.GetFirstString() != "" {
			//System.Err.Println("first string set");
			v := this.Interpreter.GetVar(this.Interpreter.GetFirstString())
			ss, _ := v.GetContentScalar()
			if len(ss) > 0 {
				//System.Err.Println("Var is not zero length");
				r = int(ss[0])
			} else {
				r = int(this.Interpreter.GetMemory(2053))
			}
		} else {
			r = int(this.Interpreter.GetMemory(2053))
		}
		//} else {
		//	r = this.Interpreter.Memory[2053];
		//}
		//}
	} else if addr == 1403 {
		r = apple2helpers.GetCursorX(this.Interpreter)
	} else if (addr >= 1024) && (addr < 1024+4096) {
		r = int(this.Interpreter.GetMemory(addr) & 0xffff) // strip color
	} else {
		r = int(this.Interpreter.GetMemory(addr % (65536 * 2)))
		fmt.Println(r)
		r = r & 0xff
	}

	this.Stack.Clear()
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(r)))

	//System.Err.Println("PEEK("+addr+") called RETURNS "+r);

	time.Sleep(time.Duration(1000000000 / 800))

	return nil
}

func (this *StandardFunctionPEEK) Syntax() string {

	/* vars */
	var result string

	result = "PEEK(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionPEEK) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
