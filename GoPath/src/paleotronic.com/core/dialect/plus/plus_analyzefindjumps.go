package plus

import (
	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/ducktape"
)

type AJumpCounter struct {
	TargetCount int
	Jumps       map[int]int
}

func NewAJumpCounter(target int) *AJumpCounter {
	return &AJumpCounter{TargetCount: target, Jumps: make(map[int]int)}
}

func (ajc *AJumpCounter) Process(msg *ducktape.DuckTapeBundle) bool {

	if msg.ID != "JMP" {
		return true
	}

	fmt.Print(msg.ID + ".")

	//isJump := msg.Payload[8] != 0
	//from := int(msg.Payload[0])<<24 | int(msg.Payload[1]<<16) | int(msg.Payload[2]<<8) | int(msg.Payload[3])
	to := int(msg.Payload[4])<<24 | int(msg.Payload[5])<<16 | int(msg.Payload[6])<<8 | int(msg.Payload[7])

	ajc.Jumps[to] = ajc.Jumps[to] + 1

	return true

}

func (ajc *AJumpCounter) GetAddressesByJumpCounts(c int) []int {
	out := make([]int, 0)
	for addr, num := range ajc.Jumps {
		if num == c {
			out = append(out, addr)
		}
	}
	return out
}

type PlusAnalyzeFindJump struct {
	dialect.CoreFunction
}

func (this *PlusAnalyzeFindJump) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	filename := params.Shift().Content
	target := params.Shift().AsInteger()

	af := NewAJumpCounter(target)
	m := map[string]interfaces.AnalyzerFunc{
		"JMP": af.Process,
	}
	this.Interpreter.AnalyzeRecording(filename, m)

	for _, addr := range af.GetAddressesByJumpCounts(target) {
		s := fmt.Sprintf("$%.4x  %d times\r\n", addr, target)
		this.Interpreter.PutStr(s)
	}

	return nil

}

func (this *PlusAnalyzeFindJump) Syntax() string {

	/* vars */
	var result string

	result = "ALTCASE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusAnalyzeFindJump) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusAnalyzeFindJump(a int, b int, params types.TokenList) *PlusAnalyzeFindJump {
	this := &PlusAnalyzeFindJump{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"

	this.MinParams = 2
	this.MaxParams = 2

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "recording", Default: *types.NewToken(types.STRING, "")},
			dialect.FunctionParamDef{Name: "count", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}
