package restalgia

import (
	"math"

	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
)

var createVoiceCallback func(index int, port int, label string, inst string)
var removeVoiceCallback func(index int, port int, label string)
var directCallback func(index int, voice int, opcode int, value uint64) uint64
var audioPortCallback func(name string) int

func SetDirectVoiceCallback(f func(index int, voice int, opcode int, value uint64) uint64) {
	directCallback = f
}

func SetCreateVoiceCallback(cvc func(index int, port int, label string, inst string)) {
	createVoiceCallback = cvc
}

func SetRemoveVoiceCallback(cvc func(index int, port int, label string)) {
	removeVoiceCallback = cvc
}
func SetAudioPortCallback(apcb func(name string) int) {
	audioPortCallback = apcb
}

func CreateVoice(ent interfaces.Interpretable, port int, label string, inst string) {
	if createVoiceCallback != nil {
		createVoiceCallback(ent.GetMemIndex(), port, label, inst)
	}
}

func RemoveVoice(ent interfaces.Interpretable, port int, label string) {
	if removeVoiceCallback != nil {
		removeVoiceCallback(ent.GetMemIndex(), port, label)
	}
}

func CommandF(ent interfaces.Interpretable, voice int, opcode int, value float64) float64 {
	if voice >= 64 {
		return -1
	}

	//	fmt.Printf("DEBUG: port 0x%.2x: %s %f\n", voice, RestCommandString(opcode), value)

	if directCallback == nil {
		mm := ent.GetMemoryMap()
		base := mm.MEMBASE(ent.GetMemIndex())
		address := memory.MICROM8_VOICE_PORT_BASE + 2*voice
		mm.WriteGlobal(ent.GetMemIndex(), base+address+1, math.Float64bits(value))
		mm.WriteGlobal(ent.GetMemIndex(), base+address+0, uint64(opcode))
		return math.Float64frombits(mm.ReadGlobal(ent.GetMemIndex(), base+address+1))
	}

	return math.Float64frombits(directCallback(ent.GetMemIndex(), voice, opcode, math.Float64bits(value)))
}

func CommandFD(voice int, opcode int, value float64) float64 {
	if voice >= 64 {
		return -1
	}

	//	fmt.Printf("DEBUG: port 0x%.2x: %s %f\n", voice, RestCommandString(opcode), value)

	if directCallback == nil {
		return 0
	}

	return math.Float64frombits(directCallback(0, voice, opcode, math.Float64bits(value)))
}

func CommandI(ent interfaces.Interpretable, voice int, opcode int, value int) int {
	if voice >= 64 {
		return -1
	}

	//	fmt.Printf("DEBUG: port 0x%.2x: %s %d\n", voice, RestCommandString(opcode), int(value))

	if directCallback == nil {
		mm := ent.GetMemoryMap()
		base := mm.MEMBASE(ent.GetMemIndex())
		address := memory.MICROM8_VOICE_PORT_BASE + 2*voice
		mm.WriteGlobal(ent.GetMemIndex(), base+address+1, uint64(value))
		mm.WriteGlobal(ent.GetMemIndex(), base+address+0, uint64(opcode))
		return int(mm.ReadGlobal(ent.GetMemIndex(), base+address+1))
	}
	return int(directCallback(ent.GetMemIndex(), voice, opcode, uint64(value)))
}

func CommandID(voice int, opcode int, value int) int {
	if voice >= 64 {
		return -1
	}

	//	fmt.Printf("DEBUG: port 0x%.2x: %s %d\n", voice, RestCommandString(opcode), int(value))

	if directCallback == nil {
		return 0
	}
	return int(directCallback(0, voice, opcode, uint64(value)))
}

func GetAudioPort(name string) int {
	if audioPortCallback == nil {
		return 0
	}
	return audioPortCallback(name)
}
