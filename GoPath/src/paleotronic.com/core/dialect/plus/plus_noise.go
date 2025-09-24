package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
	"paleotronic.com/utils"

	//	"paleotronic.com/fmt"
	"time"
)

type PlusNoise struct {
	dialect.CoreFunction
}

func (this *PlusNoise) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		//fmt.Println("e = ",e)
		return e
	}

	//fmt.Printf("@music.tone called with params list %v\n", params.AsString())

	pitch := 0
	duration := 0

	pitch = this.Stack.Shift().AsInteger()
	duration = this.Stack.Shift().AsInteger()

	//fmt.Printf( "TONE{%d, %d}\n", pitch, duration )

	//base := this.Interpreter.GetMemoryMap().MEMBASE(this.Interpreter.GetMemIndex())

	//this.Interpreter.GetMemoryMap().WriteGlobal(base+memory.OCTALYZER_SPEAKER_MS, uint64(duration))
	//this.Interpreter.GetMemoryMap().WriteGlobal(base+memory.OCTALYZER_SPEAKER_FREQ, uint64(pitch))

	cmd := fmt.Sprintf(`
use mixer.voices.boom
set instrument "WAVE=NOISE:VOLUME=1.0:ADSR=0,0,%d,1"
set frequency %d
set volume 1.0
	`, duration, pitch)
	this.Interpreter.PassRestBufferNB(cmd)

	time.Sleep(time.Millisecond * time.Duration(duration))
	this.Interpreter.PassRestBufferNB("set volume 0.0")

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusNoise) Syntax() string {

	/* vars */
	var result string

	result = "SOUND{f,delay}"

	/* enforce non void return */
	return result

}

func (this *PlusNoise) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusNoise(a int, b int, params types.TokenList) *PlusNoise {
	this := &PlusNoise{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SOUND"
	this.MinParams = 2
	this.MaxParams = 2

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "frequency", Default: *types.NewToken(types.NUMBER, "440")},
			dialect.FunctionParamDef{Name: "duration", Default: *types.NewToken(types.NUMBER, "500")},
		},
	)

	return this
}
