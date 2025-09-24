package plus

import (
	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	"paleotronic.com/ducktape"
	"paleotronic.com/encoding/mempak"
)

type APatternCollector struct {
	MemorySequence []uint64
	AddrSequence   map[int]int
	TotalSequences map[int]int
}

func NewAPatternCollector(seq []uint64) *APatternCollector {
	return &APatternCollector{MemorySequence: seq, AddrSequence: make(map[int]int), TotalSequences: make(map[int]int)}
}

func (ajc *APatternCollector) Process(msg *ducktape.DuckTapeBundle) bool {

	if msg.ID != "BMU" {
		return true
	}

	data := msg.Payload
	fullcount := int(data[0])<<16 | int(data[1])<<8 | int(data[2])
	count := fullcount / 2
	idx := 3

	ss := 0
	ee := count - 1

	for i := 0; i < fullcount; i++ {
		end := idx + 13
		if end >= len(data) {
			end = len(data)
		}
		chunk := data[idx:end]

		_, addr, value, read, size, e := mempak.Decode(chunk)
		if e != nil {
			break
		}

		if i < ss || i > ee {
			idx += size
			continue
		}

		if !read {

			a := addr % memory.OCTALYZER_INTERPRETER_SIZE
			ajc.CheckPattern(a, value)

		}

		idx += size

	}

	return true

}

func (apc *APatternCollector) CheckPattern(addr int, value uint64) {

	c := apc.AddrSequence[addr]
	if c >= len(apc.MemorySequence) {
		return
	}

	if apc.MemorySequence[c] == value {
		c++
		apc.AddrSequence[addr] = c
		if c == len(apc.MemorySequence) {
			fmt.Printf("Sequence found for 0x%.4x\n", addr)
			apc.TotalSequences[addr] = apc.TotalSequences[addr] + 1
			delete(apc.AddrSequence, addr)
		}
	} else if c > 0 {
		delete(apc.AddrSequence, addr)
	}

}

type PlusAnalyzeFindSeq struct {
	dialect.CoreFunction
}

func (this *PlusAnalyzeFindSeq) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	filename := params.Shift().Content
	seq := make([]uint64, len(params.Content))
	for i, v := range params.Content {
		seq[i] = uint64(v.AsInteger())
	}

	af := NewAPatternCollector(seq)
	m := map[string]interfaces.AnalyzerFunc{
		"BMU": af.Process,
	}
	this.Interpreter.AnalyzeRecording(filename, m)

	for addr, times := range af.TotalSequences {
		s := fmt.Sprintf("$%.4x  %d times\r\n", addr, times)
		this.Interpreter.PutStr(s)
	}

	return nil

}

func (this *PlusAnalyzeFindSeq) Syntax() string {

	/* vars */
	var result string

	result = "ALTCASE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusAnalyzeFindSeq) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusAnalyzeFindSeq(a int, b int, params types.TokenList) *PlusAnalyzeFindSeq {
	this := &PlusAnalyzeFindSeq{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"

	this.MinParams = 3
	this.MaxParams = 20

	return this
}
