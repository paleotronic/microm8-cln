// entity_stackentry.go
package interfaces

import "paleotronic.com/core/types"
import "bytes"
import "encoding/binary"
import "strings"
import "errors"

type StackEntry struct {
	PC             *types.CodeRef // Yes (2 uint64)
	Code           *types.Algorithm
	State          types.EntityState
	Locals         types.VarManager
	IsolateVars    bool
	VarPrefix      string
	TokenStack     *types.TokenList
	ErrorTrap      *types.CodeRef
	CurrentDialect Dialecter
	DataMap        *types.TokenMap
	CreatedTokens  *types.TokenList
	LoopVariable   string  // Yes (4 uint64)
	LoopStep       float32 // Yes (1 uint64)
	LoopStackSize  int     // Yes (1 uint64)
	Registers      types.BRegisters
	CurrentSub     string
}

func NewStackEntry() *StackEntry {
	this := &StackEntry{}

	return this
}

func (this *StackEntry) MarshalBinary() ([]uint64, error) {
	data := packName(this.LoopVariable, 16)
	data = append(data, float2uint(this.LoopStep))
	data = append(data, uint64(this.LoopStackSize))
	data = append(data, uint64(this.PC.Line))
	data = append(data, uint64(this.PC.Statement))

	return data, nil
}

func (this *StackEntry) UnmarshalBinary(data []uint64) error {

	if len(data) < 8 {
		return errors.New("not enough data")
	}

	this.PC = types.NewCodeRef()
	this.TokenStack = types.NewTokenList()
	this.CreatedTokens = types.NewTokenList()
	this.DataMap = types.NewTokenMap()

	this.LoopVariable = unpackName(data[0:4])
	this.LoopStep = uint2Float(data[4])
	this.LoopStackSize = int(data[5])
	this.PC.Line = int(data[6])
	this.PC.Statement = int(data[7])

	return nil
}

// Helper for converting floats to []byte
func float2uint(f float32) uint64 {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, f)
	if err != nil {
		return 0
	}
	b := buf.Bytes()
	return uint64(b[0])<<24 | uint64(b[1])<<16 | uint64(b[2])<<8 | uint64(b[3])
}

func bytes2Uint(b []byte) uint64 {
	return uint64(b[0])<<24 | uint64(b[1])<<16 | uint64(b[2])<<8 | uint64(b[3])
}

func uint2Bytes(u uint64) []byte {
	data := make([]byte, 4)

	data[0] = byte((u & 0xff000000) >> 24)
	data[1] = byte((u & 0x00ff0000) >> 16)
	data[2] = byte((u & 0x0000ff00) >> 8)
	data[3] = byte(u & 0x000000ff)

	return data
}

func uint2Float(u uint64) float32 {
	data := make([]byte, 4)

	data[0] = byte((u & 0xff000000) >> 24)
	data[1] = byte((u & 0x00ff0000) >> 16)
	data[2] = byte((u & 0x0000ff00) >> 8)
	data[3] = byte(u & 0x000000ff)

	var f float32
	b := bytes.NewBuffer(data)
	_ = binary.Read(b, binary.LittleEndian, &f)
	return f
}

func packName(name string, l int) []uint64 {
	if len(name) > l {
		name = name[0:l]
	}
	for len(name) < l {
		name += " "
	}
	b := []byte(name)

	data := make([]uint64, 0)

	for len(b) > 0 {
		if len(b) >= 4 {
			chunk := b[0:4]
			b = b[4:]
			data = append(data, bytes2Uint(chunk))
		}
	}

	return data
}

func unpackName(data []uint64) string {

	out := make([]byte, 0)

	for _, u := range data {
		b := uint2Bytes(u)
		out = append(out, b...)
	}

	s := string(out)

	return strings.Trim(s, " ")
}
