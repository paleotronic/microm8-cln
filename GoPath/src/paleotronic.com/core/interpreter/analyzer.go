package interpreter

import (
	"paleotronic.com/fmt"
	"time"

	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/ducktape"
	"paleotronic.com/encoding/mempak"
)

func (p *Player) AnalyzeEvent(msg *ducktape.DuckTapeBundle, m map[string]interfaces.AnalyzerFunc) bool {
	// call stubs from here
	// must be setup before main analyze method is called.

	for msgType, handler := range m {
		if msgType == msg.ID {
			return handler(msg)
		}
	}

	return true
}

func (p *Player) Analyze(m map[string]interfaces.AnalyzerFunc) bool {

	var e error
	var delta int
	var msg *ducktape.DuckTapeBundle

	p.start = time.Now()
	vdelta := 0

	for e == nil {

		delta, msg, e = p.Next()
		if e != nil {
			break
		}

		vdelta += delta

		if e == nil {
			if !p.AnalyzeEvent(msg, m) {
				return false
			}
		}
	}

	if e != nil {
		fmt.Println(e)
	}

	return false

}

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

	//isJump := msg.Payload[8] != 0
	//from := int(msg.Payload[0])<<24 | int(msg.Payload[1]<<16) | int(msg.Payload[2]<<8) | int(msg.Payload[3])
	to := int(msg.Payload[4])<<24 | int(msg.Payload[5]<<16) | int(msg.Payload[6]<<8) | int(msg.Payload[7])

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

type APatternCollector struct {
	MemorySequence []uint64
	AddrSequence   map[int]int
}

func NewAPatternCollector(seq []uint64) *APatternCollector {
	return &APatternCollector{MemorySequence: seq}
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
		fmt.Printf("Sequence found for 0x%.4x", addr)
	} else if c > 0 {
		delete(apc.AddrSequence, addr)
	}

}
