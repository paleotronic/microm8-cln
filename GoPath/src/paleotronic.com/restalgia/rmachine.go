package restalgia

import (
	"math"

	"paleotronic.com/fmt"

	"paleotronic.com/log"

	"paleotronic.com/core/settings"
)

// RMachineMaxVoices defines the maximum number of voices a single RMachine can control
const RMachineMaxVoices = 64

type RMachineController interface {
	opRead(port int) uint64
	opWrite(port int, opcode uint64)
	valueRead(port int) uint64
	valueWrite(port int, value uint64)
}

type rOpFunc func(rm *RMachine, v *Voice, value *uint64)

type RMachine struct {
	sampleRate  int
	VoiceSelect int
	BusValue    uint64
	Voices      [RMachineMaxVoices]*Voice
	Labels      [RMachineMaxVoices]string
	ops         [256]rOpFunc
	CountVoices int
	SlotId      int
	Names       map[string]int
	r           *RMachineRecorder
	p           *RMachinePlayer
}

// func (r *RMachine) AddVoice(name string, waveform WAVEFORM) {
// 	r.Voices[r.CountVoices] = NewVoice(label, r.sampleRate, waveform, 1.0)
// 	r.Labels[r.CountVoices] = label
// 	r.CountVoices++
// }

func NewRMachine(uid int) *RMachine {
	r := &RMachine{
		sampleRate: settings.SampleRate,
		Names:      make(map[string]int),
		SlotId:     uid,
	}
	r.Initialize()
	return r
}

func (r *RMachine) PushSamples(count int) {

}

func (r *RMachine) StartPlayback(filename string) error {
	if r.IsRecording() {
		r.r.Stop()
	}
	if r.IsPlaying() {
		r.p.Stop()
	}
	r.p = NewRMachinePlayer(r)
	return r.p.Start(filename)
}

func (r *RMachine) StartRecording(filename string) error {
	if r.r == nil {
		r.r = NewRMachineRecorder()
	}
	return r.r.Start(filename)
}

func (r *RMachine) StopPlaying() {
	if r.p != nil {
		r.p.Stop()
	}
}

func (r *RMachine) StopRecording() {
	if r.r != nil {
		r.r.Stop()
	}
}

func (r *RMachine) IsPlaying() bool {
	return r.p != nil && r.p.running
}

func (r *RMachine) IsRecording() bool {
	return r.r != nil && r.r.running
}

func (rm *RMachine) VoiceByName(name string) *Voice {
	idx := rm.indexOfName(name)
	if idx != -1 {
		return rm.Voices[idx]
	}
	return nil
}

func (rm *RMachine) indexOfName(name string) int {
	if idx, ok := rm.Names[name]; ok {
		return idx
	}
	return -1
}

func (rm *RMachine) indexOf(v *Voice) int {
	for i := 0; i < rm.CountVoices; i++ {
		if rm.Voices[i] == v {
			return i
		}
	}
	return -1
}

func (rm *RMachine) AddVoice(v *Voice) {
	i := rm.indexOf(v)
	if i == -1 {
		rm.Voices[rm.CountVoices] = v
		rm.Names[v.label] = rm.CountVoices
		rm.CountVoices++

		rm.DumpVoices()
	}
}

func (rm *RMachine) DumpVoices() {
	fmt.Println("Current voice table:-")
	for i := 0; i < rm.CountVoices; i++ {
		fmt.Printf("0x%.2x    %s\n", i, rm.Voices[i].label)
	}
	fmt.Println("")
}

func (rm *RMachine) RemoveVoice(v *Voice) {
	if rm.CountVoices == 0 {
		return
	}
	idx := rm.indexOf(v)
	for i := rm.CountVoices - 1; i > idx; i-- {
		rm.Voices[i-1] = rm.Voices[i]
	}
}

// ExecuteOpcode executes one opcode based on the current rmachine state
func (rm *RMachine) ExecuteOpcode(voice int, opcode int, value uint64) error {

	if voice > 64 {
		return nil
	}

	log.Printf("opcode %.2x for voice %x (%s)", opcode, voice, rm.Voices[voice].label)

	v := rm.Voices[voice]
	op := rm.ops[opcode&0xff]
	if v != nil && op != nil {
		op(rm, v, &value)
		if opcode >= 0x80 && rm.r != nil {
			rm.r.LogEvent(voice, opcode, value)
		}
	}

	return nil
}

func (rm *RMachine) Initialize() {
	rm.ops[0x00] = initFunc
	rm.ops[0x01] = getVolume
	rm.ops[0x81] = setVolume
	rm.ops[0x02] = getFrequency
	rm.ops[0x82] = setFrequency
	rm.ops[0x03] = getEnvelopeAttack
	rm.ops[0x83] = setEnvelopeAttack
	rm.ops[0x04] = getEnvelopeDecay
	rm.ops[0x84] = setEnvelopeDecay
	rm.ops[0x05] = getEnvelopeSustain
	rm.ops[0x85] = setEnvelopeSustain
	rm.ops[0x06] = getEnvelopeRelease
	rm.ops[0x86] = setEnvelopeRelease
	rm.ops[0x07] = getWaveform
	rm.ops[0x87] = setWaveform
	rm.ops[0x08] = getLFOControl
	rm.ops[0x88] = setLFOControl
	rm.ops[0x09] = getLFOFreq
	rm.ops[0x89] = setLFOFreq
	rm.ops[0x0a] = getLFORatio
	rm.ops[0x8a] = setLFORatio
	rm.ops[0x0b] = getLFOWaveform
	rm.ops[0x8b] = setLFOWaveform
	rm.ops[0x0c] = getEnabled
	rm.ops[0x8c] = setEnabled
	rm.ops[0x0d] = getEnvShape
	rm.ops[0x8d] = setEnvShape
	rm.ops[0x0e] = getEnvShapeFreq
	rm.ops[0x8e] = setEnvShapeFreq
	rm.ops[0x0f] = getEnvShapeEnabled
	rm.ops[0x8f] = setEnvShapeEnabled
	rm.ops[0x10] = getColour
	rm.ops[0x90] = setColour
	rm.ops[0x11] = getColourRatio
	rm.ops[0x91] = setColourRatio
	rm.ops[0x12] = getIsColour
	rm.ops[0x92] = setIsColour
	rm.ops[0x13] = getPan
	rm.ops[0x93] = setPan
}

func (rm *RMachine) SelectVoice(v int) {
	if v >= RMachineMaxVoices || v < 0 || rm.Voices[v] == nil {
		return
	}
	rm.VoiceSelect = v
}

func initFunc(rm *RMachine, v *Voice, value *uint64) {
	v.SetVolume(0)
	v.SetFrequency(440)
	v.SetEnvelope(0, 0, 0, 0)
	*value = 1
}

func getEnvShape(rm *RMachine, v *Voice, value *uint64) {
	*value = uint64(v.ENV.Shape)
}

func setEnvShape(rm *RMachine, v *Voice, value *uint64) {
	v.ENV.SetShape(int(*value))
}

func getEnvShapeFreq(rm *RMachine, v *Voice, value *uint64) {
	*value = math.Float64bits(float64(v.ENV.frequency))
}

func setEnvShapeFreq(rm *RMachine, v *Voice, value *uint64) {
	v.ENV.SetFrequency(float32(math.Float64frombits(*value)))
}

func getVolume(rm *RMachine, v *Voice, value *uint64) {
	*value = math.Float64bits(v.volume)
}

func setVolume(rm *RMachine, v *Voice, value *uint64) {
	v.volume = math.Float64frombits(*value)
}

func getPan(rm *RMachine, v *Voice, value *uint64) {
	*value = math.Float64bits(v.Pan)
}

func setPan(rm *RMachine, v *Voice, value *uint64) {
	v.Pan = math.Float64frombits(*value)
}

func getEnabled(rm *RMachine, v *Voice, value *uint64) {
	if v.OSC[0].Enabled {
		*value = 1
	} else {
		*value = 0
	}
}

func setEnabled(rm *RMachine, v *Voice, value *uint64) {
	v.OSC[0].Enabled = (*value != 0)
}

func getFrequency(rm *RMachine, v *Voice, value *uint64) {
	*value = math.Float64bits(v.OSC[0].Frequency)
}

func setFrequency(rm *RMachine, v *Voice, value *uint64) {
	f := math.Float64frombits(*value)
	v.SetFrequency(f)
	log.Printf("*** set voice %s frequency to %f", v.label, f)
}

func getEnvelopeAttack(rm *RMachine, v *Voice, value *uint64) {
	if v.OSC[0].envelope != nil {
		*value = uint64(v.OSC[0].envelope.Attack)
	}
}

func setEnvelopeAttack(rm *RMachine, v *Voice, value *uint64) {
	if v.OSC[0].envelope != nil {
		v.OSC[0].envelope.Attack = int(*value)
		v.OSC[0].envelope.RecalculateEnvelope()
		v.OSC[0].envelope.Trigger()
	}
}

func getEnvelopeDecay(rm *RMachine, v *Voice, value *uint64) {
	if v.OSC[0].envelope != nil {
		*value = uint64(v.OSC[0].envelope.Decay)
	}
}

func setEnvelopeDecay(rm *RMachine, v *Voice, value *uint64) {
	if v.OSC[0].envelope != nil {
		v.OSC[0].envelope.Decay = int(*value)
		v.OSC[0].envelope.RecalculateEnvelope()
		v.OSC[0].envelope.Trigger()
	}
}

func getEnvelopeSustain(rm *RMachine, v *Voice, value *uint64) {
	if v.OSC[0].envelope != nil {
		*value = uint64(v.OSC[0].envelope.Sustain)
	}
}

func setEnvelopeSustain(rm *RMachine, v *Voice, value *uint64) {
	if v.OSC[0].envelope != nil {
		v.OSC[0].envelope.Sustain = int(*value)
		v.OSC[0].envelope.RecalculateEnvelope()
		v.OSC[0].envelope.Trigger()
	}
}

func getEnvelopeRelease(rm *RMachine, v *Voice, value *uint64) {
	if v.OSC[0].envelope != nil {
		*value = uint64(v.OSC[0].envelope.Release)
	}
}

func setEnvelopeRelease(rm *RMachine, v *Voice, value *uint64) {
	if v.OSC[0].envelope != nil {
		v.OSC[0].envelope.Release = int(*value)
		v.OSC[0].envelope.RecalculateEnvelope()
		v.OSC[0].envelope.Trigger()
	}
}

func getWaveform(rm *RMachine, v *Voice, value *uint64) {
	*value = uint64(v.OSC[0].waveform)
}

func setWaveform(rm *RMachine, v *Voice, value *uint64) {
	v.OSC[0].SetWaveform(WAVEFORM(*value))
}

func getLFOControl(rm *RMachine, v *Voice, value *uint64) {
	if v.LFO != nil {
		*value = uint64(v.LFO.BoundControl)
	}
}

func setLFOControl(rm *RMachine, v *Voice, value *uint64) {
	if v.LFO != nil {
		v.LFO.BoundControl = LFOControlType(*value)
	}
}

func getLFORatio(rm *RMachine, v *Voice, value *uint64) {
	if v.LFO != nil {
		*value = math.Float64bits(v.LFO.MaxDeviation)
	}
}

func setLFORatio(rm *RMachine, v *Voice, value *uint64) {
	if v.LFO != nil {
		v.LFO.MaxDeviation = math.Float64frombits(*value)
	}
}

func getLFOFreq(rm *RMachine, v *Voice, value *uint64) {
	if v.LFO == nil {
		*value = math.Float64bits(v.LFO.Frequency)
	}
}

func setLFOFreq(rm *RMachine, v *Voice, value *uint64) {
	if v.LFO != nil {
		v.LFO.SetFrequency(math.Float64frombits(*value))
	}
}

func getLFOWaveform(rm *RMachine, v *Voice, value *uint64) {
	if v.LFO != nil {
		*value = uint64(v.LFO.waveform)
	}
}

func setLFOWaveform(rm *RMachine, v *Voice, value *uint64) {
	if v.LFO != nil {
		v.LFO.SetWaveform(WAVEFORM(*value))
	}
}

func getVoiceIndex(r *RMachine, v *Voice) int {
	for i, vv := range r.Voices {
		if vv == v {
			return i
		}
	}
	return 999
}

func getColour(rm *RMachine, v *Voice, value *uint64) {
	if v.Colour != nil {
		*value = uint64(getVoiceIndex(rm, v.Colour))
	}
}

func setColour(rm *RMachine, v *Voice, value *uint64) {
	vi := int(*value)
	if vi >= 0 && vi < RMachineMaxVoices && rm.Voices[vi] != nil {
		v.Colour = rm.Voices[vi]
	} else {
		v.Colour = nil
	}
}

func getColourRatio(rm *RMachine, v *Voice, value *uint64) {
	*value = math.Float64bits(float64(v.ColourRatio))
}

func setColourRatio(rm *RMachine, v *Voice, value *uint64) {
	v.ColourRatio = float32(math.Float64frombits(*value))
}

func getIsColour(rm *RMachine, v *Voice, value *uint64) {
	if v.IsColour {
		*value = 1
	} else {
		*value = 0
	}
}

func setIsColour(rm *RMachine, v *Voice, value *uint64) {
	v.IsColour = (*value != 0)
}

func getEnvShapeEnabled(rm *RMachine, v *Voice, value *uint64) {
	if v.ENV != nil && v.ENV.Enabled {
		*value = 1
	} else {
		*value = 0
	}
}

func setEnvShapeEnabled(rm *RMachine, v *Voice, value *uint64) {
	if v.ENV != nil {
		v.ENV.Enabled = (*value != 0)
		// if v.ENV.Enabled {
		// 	v.ENV.Trigger()
		// } else {
		// 	v.ENV.Collapse()
		// }
	}
}
