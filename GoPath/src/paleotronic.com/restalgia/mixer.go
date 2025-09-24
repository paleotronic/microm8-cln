package restalgia

import (
	"bufio"
	"math"
	"os" //"time"

	"paleotronic.com/core/settings"
	"paleotronic.com/fmt"
)

/*
	Mixing functions
*/

const (
	attenuationPerStream = 0.2 // Per stream attenuation
)

type Mixer struct {
	*RInfo
	SlotSelect int
	Slots      [settings.NUMSLOTS]*RMachine
	Volume     float32
	output     *os.File
	b          *bufio.Writer
	isMuted    bool
}

func NewMixer() *Mixer {

	m := &Mixer{
		Volume:     0.5,
		SlotSelect: 0,
	}

	for i := 0; i < settings.NUMSLOTS; i++ {
		m.Slots[i] = NewRMachine(i)
	}

	m.initFields("mixer")

	return m

}

func (this *Mixer) IsRecording() bool {
	return this.output != nil && this.b != nil
}

func (this *Mixer) StopRecording() {
	if this.output != nil {
		this.b.Flush()
		this.output.Close()
		this.output = nil
		this.b = nil
	}
}

func (this *Mixer) StartRecording(filename string) {
	this.StopRecording()
	var e error
	this.output, e = os.Create(filename)
	if e != nil {
		panic(e)
	}
	this.b = bufio.NewWriter(this.output)
}

func (this *Mixer) initFields(label string) {
	this.RInfo = NewRInfo(label, "Mixer")
	this.RInfo.AddAttribute(
		"volume",
		"float32",
		false,
		func(index int) interface{} {
			return this.Volume
		},
		func(index int, v interface{}) {
			this.Volume = v.(float32)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"deleted",
		"string",
		false,
		func(index int) interface{} {
			return ""
		},
		func(index int, v interface{}) {
			s := v.(string)
			this.DeleteVoice(this.SlotSelect, s)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"created",
		"string",
		false,
		func(index int) interface{} {
			return ""
		},
		func(index int, v interface{}) {
			s := v.(string)
			this.CreateVoice(this.SlotSelect, s)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddObject(
		"voices",
		"Voice",
		true,
		func(index int) RQueryable {
			return this.Slots[this.SlotSelect].Voices[index]
		},
		func(index int, v RQueryable) {
			this.Slots[this.SlotSelect].Voices[index] = v.(*Voice)
		},
		func() int {
			return this.Slots[this.SlotSelect].CountVoices
		},
	)
}

func (this *Mixer) GetClass() string {
	return "mixer"
}

func (this *Mixer) SetMute(b bool) {
	this.isMuted = b
}

func (this *Mixer) IsMuted() bool {
	return this.isMuted
}

func (this *Mixer) VolumeUp() {
	this.Volume += 0.1
	if this.Volume > 1 {
		this.Volume = 1
	}
}

func (this *Mixer) VolumeDown() {
	this.Volume -= 0.1
	if this.Volume < 0 {
		this.Volume = 0
	}
}

func (this *Mixer) GetVolume() float32 {
	return this.Volume
}

func (this *Mixer) RemoveVoice(slot int, rv *Voice) {
	this.Slots[slot].RemoveVoice(rv)
}

func (this *Mixer) AddVoice(slot int, av *Voice) {
	this.Slots[slot].AddVoice(av)
}

// func (this *Mixer) Fill(b []float32) {

// 	count := 0 //len(this.Voices)

// 	chunk := make([][]float32, len(this.Voices))

// 	for _, v := range this.Voices {
// 		if !v.Active {
// 			continue
// 		}
// 		/*if v.OSC[0].GetWaveform() == CUSTOM && !v.OSC[0].GetWfCUSTOM().Playing {
// 			continue
// 		} else if v.OSC[0].GetWaveform() != CUSTOM && v.GetVolume() == 0 {
// 			continue
// 		}*/
// 		chunk[count] = v.GetSamplesf(len(b))
// 		count++
// 	}

// 	mvol := float32(1)
// 	if this.isMuted {
// 		mvol = 0
// 	}

// 	for idx, _ := range b {

// 		buff := make([]float32, count)
// 		for i := 0; i < count; i++ {
// 			buff[i] = chunk[i][idx]
// 		}

// 		// mix to single value
// 		b[idx] = mixChannelsMonoSimple(buff) * this.Volume * mvol

// 	}

// 	if this.output != nil {
// 		out := make([]byte, len(b))
// 		for i, bb := range b {
// 			out[i] = byte(127 * bb)
// 			this.output.Write(out)
// 		}
// 	}

// }

func (this *Mixer) Close() {
	this.StopRecording()
}

func (this *Mixer) FindVoice(slot int, name string) *Voice {

	//fmt.Printf("Mixer has %d slots (ss = %d)\n", len(this.Slots), this.SlotSelect)
	if this.Slots[slot] == nil {
		return nil
	}

	for _, v := range this.Slots[slot].Voices {
		if v == nil {
			continue
		}
		if v.label == name {
			return v
		}
	}
	return nil
}

func (this *Mixer) FindVoicePort(slot int, name string) int {

	//fmt.Printf("Mixer has %d slots (ss = %d)\n", len(this.Slots), this.SlotSelect)
	if this.Slots[slot] == nil {
		return -1
	}

	for idx, v := range this.Slots[slot].Voices {
		if v == nil {
			continue
		}
		if v.label == name {
			return idx
		}
	}
	return -1
}

func (this *Mixer) SetupVoice(slot int, port int, name string, instr string) *Voice {

	fmt.Printf("Create voice %s with rate %d\n", name, settings.SampleRate)

	v := NewVoice(name, settings.SampleRate, SQUARE, 0.0)
	v.parent = this
	v.context = slot
	i := NewInstrument(instr)
	i.Apply(v)
	v.SetVolume(0)
	this.Slots[slot].Voices[port] = v
	if port > this.Slots[slot].CountVoices-1 {
		this.Slots[slot].CountVoices = port + 1
	}
	this.Slots[slot].DumpVoices()
	return v
}

func (this *Mixer) DestroyVoice(slot int, port int, name string) {
	//
}

func (this *Mixer) CreateVoice(slot int, name string) *Voice {
	v := this.FindVoice(slot, name)
	if v == nil {
		v = NewVoice(name, settings.SampleRate, SQUARE, 1.0)
		v.parent = this
		v.context = slot
		this.AddVoice(slot, v)
	}
	return v
}

func (this *Mixer) DeleteVoice(slot int, name string) *Voice {
	v := this.FindVoice(slot, name)
	if v != nil {
		this.RemoveVoice(slot, v)
	}
	return v
}

func (this *Mixer) DumpState() {
	if this.Slots[this.SlotSelect].CountVoices == 0 {
		//fmt.Printf("slot %d has zero voices!\n", this.SlotSelect)
		return
	}

	// canhear := make([]string, 0, this.Slots[this.SlotSelect].CountVoices)

	for i := 0; i < this.Slots[this.SlotSelect].CountVoices; i++ {

		v := this.Slots[this.SlotSelect].Voices[i]

		if v.IsAudible() {
			fmt.Printf("ON: %s\n", v.label, v.DumpState())
		}
	}

}

func (this *Mixer) VoiceSetVolume(label string, level float64) {
	v := this.FindVoice(this.SlotSelect, label)
	if v != nil {
		v.SetVolume(level)
	}
}

func (this *Mixer) VoiceGetVolume(label string) float64 {
	v := this.FindVoice(this.SlotSelect, label)
	if v != nil {
		return v.GetVolume()
	}
	return 0
}

func (this *Mixer) FillStereo(b []float32) {

	//fmt.Printf("%d ", this.Slots[this.SlotSelect].CountVoices)

	this.Volume = float32(settings.MixerVolume)

	if this.Slots[this.SlotSelect].CountVoices == 0 || settings.BlueScreen {
		//fmt.Printf("slot %d has zero voices!\n", this.SlotSelect)
		for i, _ := range b {
			b[i] = 0
		}
		return
	}

	count := 0 //len(this.Voices)

	chunk := make([][]float32, this.Slots[this.SlotSelect].CountVoices)

	// canhear := make([]string, 0, this.Slots[this.SlotSelect].CountVoices)
	var canhear bool

	for i := 0; i < this.Slots[this.SlotSelect].CountVoices; i++ {

		v := this.Slots[this.SlotSelect].Voices[i]

		if settings.TemporaryMute && v.label != "beep" {
			continue
		}

		canhear = v.IsAudible()

		if !canhear && v.ENV == nil {
			continue
		}

		if !canhear && v.ENV != nil && v.ENV.Enabled {
			chunk[count] = v.ENV.GetAmplitude2f(len(b), 1.5)
			count++
			continue
		}

		if !canhear {
			continue
		}

		// canhear = append(canhear, v.DumpState())
		chunk[count] = v.GetSamples2f(len(b))
		count++
	}

	mvol := float32(1)
	if this.isMuted {
		mvol = 0
	}

	for idx := range b {

		buff := make([]float32, count)
		for i := 0; i < count; i++ {
			buff[i] = chunk[i][idx]
		}

		// mix to single value
		b[idx] = mixChannelsMonoSimple(buff) * this.Volume * mvol

	}

	if this.b != nil {
		out := make([]byte, len(b)*2)
		var val int16
		for i, bb := range b {
			val = int16(32767 * bb)
			out[i*2+1] = byte(val >> 8)
			out[i*2+0] = byte(val & 0xff)
		}
		_, err := this.b.Write(out)
		if err != nil {
			panic(err)
		}
	}

	// t := time.Now()
	// if t.Unix()%5 == 0 && len(canhear) > 0 {
	// 	fmt.Printf("Audible voices: %v\n", canhear)
	// }

}

// Convert logarithmic decibels to linear
func decibel2linear(decibels float32) float32 {
	return float32(math.Pow(10.0, float64(decibels)/20.0))
}

// func mixChannelsMono(voices []float32) float32 {
// 	return mixChannelsMonoSimple(voices)
// }

// mix any number of mono streams together
func mixChannelsMono(voices []float32) float32 {

	var result float32 = 0.0

	var channels int = len(voices)

	var attenuationFactor float32 = decibel2linear(-1.0 * attenuationPerStream * float32(channels))

	for x := 0; x < channels; x++ {
		result += attenuationFactor * voices[x]
	}

	if result > 1 {
		result = 1
	} else if result < -1 {
		result = -1
	}

	return result
}

// mix any number of mono streams together
func mixChannelsMonoSimpleOld(voices []float32) float32 {

	//fmt.Printf("Mix(%d)\n", len(voices))

	var result float32 = 0.0

	if len(voices) == 0 {
		return result
	}

	var channels int = len(voices)

	for x := 0; x < channels; x++ {
		result += voices[x]
	}

	return result / float32(channels)
}

func mixLogarithmicRangeCompression(i float32) float32 {
	if i < -1 {
		return float32(-math.Log(-float64(i)-0.85)/14 - 0.75)
	} else if i > 1 {
		return float32(math.Log(float64(i)-0.85)/14 + 0.75)
	} else {
		return (i / 1.61803398875)
	}
}

func mixChannelsMonoSimple(voices []float32) float32 {

	var result float32 = 0.0

	if len(voices) == 0 {
		return result
	}

	var channels int = len(voices)

	for x := 0; x < channels; x++ {
		result += voices[x] * 0.75
	}

	return result
}

func mixChannelsMonoSimpleAB(voices []float32) float32 {
	var result float32 = 0.0

	if len(voices) == 0 {
		return result
	}

	if len(voices) == 1 {
		return voices[0]
	}

	// result = s1 + s2 + sN
	var factor float32 = float32(len(voices)) - 1
	for x := 0; x < len(voices); x++ {
		result += voices[x]
		factor = factor * voices[x]
	}

	return result - factor
}

func mixClip(v, limit float32) float32 {
	if v < -limit {
		v = -limit
	} else if v > limit {
		v = limit
	}
	return v
}

// mix any number of stereo streams together
func mixChannelsStereo(left []float32, right []float32) []float32 {

	var result []float32 = make([]float32, 2)
	result[0] = 0.0
	result[1] = 0.0

	var channels int = len(left)

	var attenuationFactor float32 = decibel2linear(-1.0 * attenuationPerStream * float32(channels))

	for x := 0; x < channels; x++ {
		result[0] += attenuationFactor * left[x]
		result[1] += attenuationFactor * right[x]
	}

	return result
}
