package restalgia

import (
	"fmt"
	//"fmt"
	"regexp"
	"strings"

	//	"paleotronic.com/fmt"
	"math"

	"paleotronic.com/utils"

	//log2 "log"
)

type Voice struct {
	*RInfo
	Active       bool
	ticksPer32nd int
	TickCount    int
	NoteIndex    int
	NoteLength   int
	volume       float64
	sampleRate   int
	Pan, dPan    float64
	Notes        []string
	Tempo        int
	OSC          []*Oscillator
	ENV          *EnvelopeGeneratorSimple
	LFO          *LFO
	FILT         []Filter
	PrevOctave   int
	CustomOnly   bool
	Colour       *Voice
	ColourRatio  float32
	IsColour     bool
	parent       *Mixer
	context      int
}

func (this *Voice) AddFilter(f Filter) {
	this.FILT = append(this.FILT, f)
}

func (this *Voice) SetOSCEnabled(index int, on bool) {
	this.OSC[index].SetEnabled(on)
}

func (this *Voice) SetOSCPulseWidth(index int, pulseWidth float64) {
	this.OSC[index].SetPulseWidthRadians(pulseWidth)
}

func (this *Voice) SetEnvelope(attack int, decay int, sustain int, release int) {
	// this.ENV.SetEnvelope([]int{attack, decay, sustain, release})
	// this.ENV.Trigger()
}

func (this *Voice) Stop() {
	//this.ENV.Collapse()
}

func (this *Voice) Play(note string, octave int) {

	var f float64 = NT.Frequency(note, octave)

	if f <= 0 {
		return
	}

	this.SetFrequency(f)
}

func (this *Voice) GetPannedAmplitudes() []float32 {
	var vals []float32 = make([]float32, 2)

	// ensure its in bounds
	p := this.Pan + this.dPan
	if p < -1 {
		p = -1
	}
	if p > 1 {
		p = 1
	}

	p = (p + float64(1)) / float64(2)

	var amp float32 = this.GetAmplitude()

	var panleft float32 = 1 - float32(p)
	if panleft > 1 {
		panleft = 1
	}
	var panright float32 = 1 - panleft

	vals[0] = panleft * amp
	vals[1] = panright * amp

	return vals
}

func (this *Voice) SetOSCEnvelope(index int, attack int, decay int, sustain int, release int) {
	this.OSC[index].GetEnvelope().SetEnvelope([]int{attack, decay, sustain, release})
	this.OSC[index].Trigger()
}

func (this *Voice) GetAmplitude() float32 {
	if this.OSC[0].waveform == CUSTOM {
		return this.OSC[0].GetAmplitude() * float32(this.volume)
	}

	this.HandleNotes()

	// if this.label == "psgnoise0c0" || this.label == "psgtones0c0v0" {
	// 	this.DumpState()
	// }

	// now this gets fun ...
	var level float32 = 0
	var countEnabled int = 0

	var vLFO float64

	if this.LFO != nil {
		vLFO = this.LFO.GetVariance()
	}

	for x := 0; x < numOSC; x++ {

		if this.OSC[x] == nil {
			continue
		}

		if this.LFO != nil {
			//OldState = *this.OSC[x] // state state
			switch this.LFO.BoundControl {
			case LFO_PHASESHIFT:
				this.OSC[x].dPhase = 2 * math.Pi * vLFO
			case LFO_VOLUME:
				this.OSC[x].dVolume = this.OSC[x].Volume * vLFO
			case LFO_FREQUENCY:
				this.OSC[x].dFrequency = this.OSC[x].Frequency * vLFO
				this.OSC[x].RealignFrequency = true
			case LFO_PULSEWIDTH:
				this.OSC[x].dPulseWidth = this.OSC[x].PulseWidth * vLFO
			}
		}

		if this.OSC[x].IsEnabled() {
			level += this.OSC[x].GetAmplitude()
			countEnabled++
		}

		// if this.LFO != nil && this.OSC[x] != nil {
		// 	//Restore defaults
		// 	*this.OSC[x] = OldState
		// }
	}

	for _, f := range this.FILT {
		level = f.Filter(level)
	}

	//////fmt.Printf("ENV = %v\n", this.ENV.GetAmplitude())
	volume := float32(this.volume)
	if this.ENV.Enabled {
		//fmt.Printf("%s: Env on\n", this.label)
		volume = this.ENV.GetAmplitude(false)
	}

	me := level * 1.0 * volume

	if this.Colour != nil && this.ColourRatio > 0 {
		me = (1-this.ColourRatio)*me + (volume * this.ColourRatio * this.Colour.GetAmplitude())
	}

	// if this.ENV.Enabled {
	// 	me = me * this.ENV.GetAmplitude(false)
	// }

	return me

}

func (this *Voice) DumpState() string {
	return fmt.Sprintf(
		"Voice: %s, volume: %f, freq: %f, env enabled: %v, env shape: %d, env volume: %f, color ratio: %f\n",
		this.label,
		this.volume,
		this.OSC[0].Frequency,
		this.ENV.Enabled,
		this.ENV.Shape,
		this.ENV.GetAmplitude(true),
		this.ColourRatio,
	)
}

func (this *Voice) SetColour(v *Voice) {
	this.Colour = v
}

func (this *Voice) SetColourRatio(f float32) {
	if f < 0 {
		f = 0
	} else if f > 1 {
		f = 1
	}
	this.ColourRatio = f
}

// GetSamplesf returns 'count' samples from the device
func (this *Voice) GetSamplesf(count int) []float32 {
	out := make([]float32, count)
	for i, _ := range out {
		out[i] = this.GetAmplitude()
	}
	return out
}

// GetSamplesf returns 'count' samples from the device
func (this *Voice) GetSamples2f(count int) []float32 {
	out := make([]float32, count)
	var tmp = []float32{0, 0}
	for i, _ := range out {
		if this.Pan == 0 {
			switch i % 2 {
			case 0:
				out[i] = this.GetAmplitude()
			case 1:
				out[i] = out[i-1]
			}
		} else {
			switch i % 2 {
			case 0:
				tmp = this.GetPannedAmplitudes()
				out[i] = tmp[0]
			case 1:
				out[i] = tmp[1]
			}
		}
	}
	return out
}

func (v *Voice) IsAudible() bool {
	osc := v.OSC[0]
	if v.IsColour {
		return false
	} else if !osc.Enabled {
		return false
	} else if osc.GetWaveform() == CUSTOM && !osc.GetWfCUSTOM().Playing {
		return false
	} else if v.volume == 0 {
		if !v.ENV.Enabled {
			return false // nothing to hear...
		}
		// so env is enabled...
	} else if len(v.Notes) == 0 && osc.GetEnvelope().GetAmplitude(true) == 0 {
		return false
	}
	return true
}

func (this *Voice) GetVolume() float64 {
	return this.volume
}

func (this *Voice) SetVolume(volume float64) {
	this.volume = volume
}

func (this *Voice) SetOSCFrequencyMultiplier(index int, frequency float64) {
	this.OSC[index].SetFrequencyMultiplier(frequency)
}

func (this *Voice) SetOSCPhaseShift(index int, shift float64) {
	this.OSC[index].SetPhaseShift(shift)
}

func (this *Voice) SetFrequency(frequency float64) {
	for x := 0; x < numOSC; x++ {
		this.OSC[x].SetFrequency(frequency)
		this.OSC[x].Trigger()
	}
	//	this.ENV.Trigger()
}

func (this *Voice) SetFrequencySlide(frequency float64) {
	for x := 0; x < numOSC; x++ {
		this.OSC[x].SetFrequencySilent(frequency)
		//this.OSC[x].Trigger()
	}
	//this.ENV.Trigger()
}

func (this *Voice) GetSamplesForDuration(seconds float64) int64 {
	return this.OSC[0].GetSamplesForDuration(seconds)
}

func (this *Voice) SetOSCWaveform(i int, wf WAVEFORM) {
	this.OSC[i].SetWaveform(wf)
}

func (this *Voice) TicksForNoteLength(noteLength int) int {
	switch noteLength % 10 {
	case 0:
		return this.ticksPer32nd // 1/32
	case 1:
		return this.ticksPer32nd * 2 // 1/16
	case 2:
		return this.ticksPer32nd * 3 // 3/32
	case 3:
		return this.ticksPer32nd * 4 // 1/8
	case 4:
		return this.ticksPer32nd * 6 // 3/16
	case 5:
		return this.ticksPer32nd * 8 // 1/4
	case 6:
		return this.ticksPer32nd * 12 // 3/8
	case 7:
		return this.ticksPer32nd * 16 // 1/2
	case 8:
		return this.ticksPer32nd * 24 // 3/4
	case 9:
		return this.ticksPer32nd * 32 // 1/1
	}
	return this.ticksPer32nd
}

func NewVoice(label string, sampleRate int, osc WAVEFORM, volume float64) *Voice {
	this := &Voice{}
	this.OSC = make([]*Oscillator, numOSC)

	this.PrevOctave = 3

	this.volume = volume
	this.sampleRate = sampleRate
	this.ticksPer32nd = sampleRate / 8
	this.Tempo = 120
	this.Pan = 0
	this.NoteLength = 7 // half note
	this.initFields(label)

	// create oscillators
	for x := 0; x < numOSC; x++ {
		this.OSC[x] = NewOscillator(fmt.Sprintf("osc%d", x), osc, 440, sampleRate, true)
	}

	this.ENV = NewEnvelopeGeneratorSimple(sampleRate)

	this.FILT = make([]Filter, 0)

	// if label == "speaker" {
	// 	log2.Printf("adding lopass to voice %s", label)
	// 	this.FILT = []Filter{
	// 		NewHighPassFilter(float32(sampleRate), 1000),
	// 	}
	// }

	// this.LFO = NewLFO(SINE, 80, this, LFO_FREQUENCY, 0.30)
	// this.LFO.Trigger()

	return this
}

func (this *Voice) initFields(label string) {
	this.RInfo = NewRInfo(label, "Voice")
	this.RInfo.AddAttribute(
		"enabled",
		"bool",
		false,
		func(index int) interface{} {
			return this.OSC[0].Enabled
		},
		func(index int, v interface{}) {
			this.OSC[0].SetEnabled(v.(bool))
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"notes",
		"string",
		false,
		func(index int) interface{} {
			return ""
		},
		func(index int, v interface{}) {
			this.AddNoteStream(v.(string))
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"iscolour",
		"bool",
		false,
		func(index int) interface{} {
			return this.IsColour
		},
		func(index int, v interface{}) {
			this.IsColour = v.(bool)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"instrument",
		"string",
		false,
		func(index int) interface{} {
			return ""
		},
		func(index int, v interface{}) {
			inst := NewInstrument(v.(string))
			inst.Apply(this)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"frequency",
		"float64",
		false,
		func(index int) interface{} {
			return this.OSC[0].Frequency
		},
		func(index int, v interface{}) {
			this.SetFrequency(v.(float64))
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"frequencyslide",
		"float64",
		false,
		func(index int) interface{} {
			return this.OSC[0].Frequency
		},
		func(index int, v interface{}) {
			this.SetFrequencySlide(v.(float64))
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"volume",
		"float64",
		false,
		func(index int) interface{} {
			return this.volume
		},
		func(index int, v interface{}) {
			this.volume = v.(float64)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"pan",
		"float64",
		false,
		func(index int) interface{} {
			return this.Pan
		},
		func(index int, v interface{}) {
			this.Pan = v.(float64)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"colourratio",
		"float32",
		false,
		func(index int) interface{} {
			return this.ColourRatio
		},
		func(index int, v interface{}) {
			this.ColourRatio = v.(float32)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"colour",
		"string",
		false,
		func(index int) interface{} {
			if this.Colour != nil {
				return this.Colour.label
			}
			return ""
		},
		func(index int, v interface{}) {
			s := v.(string)
			this.Colour = this.parent.FindVoice(this.context, s)
		},
		func() int {
			return 1
		},
	)
	// this.RInfo.AddObject(
	// 	"envelope",
	// 	"EnvelopeGenerator",
	// 	false,
	// 	func(index int) RQueryable {
	// 		return this.ENV
	// 	},
	// 	func(index int, v RQueryable) {
	// 		this.ENV = v.(*EnvelopeGenerator)
	// 	},
	// 	func() int {
	// 		return 1
	// 	},
	// )
	this.RInfo.AddObject(
		"lfo",
		"LFO",
		false,
		func(index int) RQueryable {
			if this.LFO == nil {
				this.LFO = NewLFO(SINE, 5, this, LFO_NONE, 0.03)
			}
			return this.LFO
		},
		func(index int, v RQueryable) {
			if this.LFO == nil {
				this.LFO = NewLFO(SINE, 5, this, LFO_NONE, 0.03)
			}
			this.LFO = v.(*LFO)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddObject(
		"oscillators",
		"Oscillator",
		true,
		func(index int) RQueryable {
			return this.OSC[index]
		},
		func(index int, v RQueryable) {
			this.OSC[index] = v.(*Oscillator)
		},
		func() int {
			return len(this.OSC)
		},
	)
}

func (this *Voice) HandleNotes() {
	/* play back music if it is available */
	if this.Notes == nil {
		return
	}

	if this.NoteIndex >= len(this.Notes) {
		this.NoteIndex = 0
		this.Notes = make([]string, 0)
		return
	}

	if this.TickCount != 0 {
		this.TickCount = (this.TickCount + 1) % this.TicksForNoteLength(this.NoteLength)
		return
	}

	noteRegex := regexp.MustCompile("^[ABCDEFG][#b]?[0-8]?$")
	tempoRegex := regexp.MustCompile("^T[0-9]+$")
	lengthRegex := regexp.MustCompile("^L[0-9]$")
	octaveRegex := regexp.MustCompile("^O[0-9]$")
	volumeRegex := regexp.MustCompile("^V[0-9]+$")
	splitRegex := regexp.MustCompile("[0-9]+")
	instRegex := regexp.MustCompile("^I.+$")

	if this.TickCount == 0 {
		// start note

		note := this.Notes[this.NoteIndex]
		this.NoteIndex++

		////fmt.Printf("[restalgia::music] ================== New instruction %s\n", note)

		if instRegex.MatchString(note) {
			////fmt.Printntf("[restalgia::music] Next inst [%s] is INSTRUMENT", note)
			i := NewInstrument(utils.Delete(note, 1, 1))
			i.Apply(this)
			return
		} else if note == "R" {
			////fmt.Printntf("[restalgia::music] Next inst [%s] is REST", note)
			this.TickCount = (this.TickCount + 1) % this.TicksForNoteLength(this.NoteLength)
		} else if noteRegex.MatchString(note) {
			////fmt.Printntf("[restalgia::music] Next inst [%s] is NOTE\n", note)
			parts := splitRegex.Split(note, 2)
			n := parts[0]
			o := strings.Replace(note, n, "", -1)
			if o == "" {
				o = utils.IntToStr(this.PrevOctave)
			}
			octave := utils.StrToInt(o)
			this.PrevOctave = octave
			this.Play(n, octave)

			////fmt.Printntf("[restalgia::music] Play note %s, octave %d\n", n, octave)
			this.TickCount = (this.TickCount + 1) % this.TicksForNoteLength(this.NoteLength)
		} else if tempoRegex.MatchString(note) {
			// tempo change
			////fmt.Printntf("[restalgia::music] Next inst [%s] is TEMPO\n", note)
			v := strings.Replace(note, "T", "", -1)
			this.Tempo = utils.StrToInt(v)

			samplesPerMinute := 60 * this.sampleRate
			wholeNotesPerMinute := this.Tempo / 4
			samplesPerWholeNote := samplesPerMinute / wholeNotesPerMinute

			this.ticksPer32nd = samplesPerWholeNote / 32
			this.TickCount = 0
			////fmt.Printntf("[restalgia::music] Tempo change to %d BPM\n", this.Tempo)
			return
		} else if lengthRegex.MatchString(note) {
			// tempo change
			////fmt.Printntf("[restalgia::music] Next inst [%s] is LENGTH\n", note)
			v := strings.Replace(note, "L", "", -1)
			this.NoteLength = utils.StrToInt(v)
			this.TickCount = 0
			////fmt.Printntf("[restalgia::music] Note length change %d\n", this.NoteLength)
			return
		} else if octaveRegex.MatchString(note) {
			////fmt.Printntf("[restalgia::music] Next inst [%s] is OCTAVE\n", note)
			v := strings.Replace(note, "O", "", -1)
			this.PrevOctave = utils.StrToInt(v)
			return
		} else if volumeRegex.MatchString(note) {
			////fmt.Printntf("[restalgia::music] Next inst [%s] is VOLUME\n", note)
			v := strings.Replace(note, "V", "", -1)
			vol := utils.StrToInt(v)
			if vol > 100 {
				vol = 100
			}
			this.SetVolume(float64(vol) / 100)
			return
		}
	}

	//////fmt.Printf("[restalgia::music] Tickcount is %d\n", this.TickCount)

}

func (this *Voice) GetOSC(index int) *Oscillator {
	return this.OSC[index]
}

func (this *Voice) SetOSC(index int, o *Oscillator) {
	this.OSC[index] = o
}

func (this *Voice) GetOSCVolume(index int) float64 {
	return this.OSC[index].GetVolume()
}

func (this *Voice) SetOSCVolume(index int, bias float64) {
	this.OSC[index].SetVolume(bias)
}

func (this *Voice) SetOSCFrequency(x int, frequency int) {
	this.OSC[x].SetFrequency(float64(frequency))
	this.OSC[x].Trigger()
}

func (this *Voice) AddInstrumentChange(content string) {

	if this.Notes == nil || len(this.Notes) == 0 {
		i := NewInstrument(content)
		i.Apply(this)
		return
	}

	n := this.Notes
	n = append(n, "I"+content)
	this.Notes = n
}

func (this *Voice) AddNoteStream(content string) {

	// SOUND PLAY 0, "A#5"
	parts := make([]string, 0)
	ap := false

	chunk := ""
	for _, ch := range content {
		if (ch >= 'A' && ch <= 'G') || ch == 'L' || ch == 'T' || ch == 'O' || ch == 'V' || ch == 'R' {
			if len(chunk) > 0 {
				parts = append(parts, chunk)
			}
			chunk = string(ch)
		} else if ch != ' ' && ch != ';' && ch != '+' {
			chunk = chunk + string(ch)
		} else if ch == '+' {
			ap = true
		}
	}
	if len(chunk) > 0 {
		parts = append(parts, chunk)
	}

	if !ap {
		this.Notes = make([]string, 0)
	}

	n := this.Notes

	for _, s := range parts {
		n = append(n, s)
		////fmt.Printntf("[restalgia::music] cache note %s\n", s)
	}

	this.Notes = n

	this.TickCount = 0
	this.NoteIndex = 0

	// tempo is 1/4 notes per minute
	// so...
	samplesPerMinute := 60 * this.sampleRate
	wholeNotesPerMinute := this.Tempo / 4
	samplesPerWholeNote := samplesPerMinute / wholeNotesPerMinute

	this.ticksPer32nd = samplesPerWholeNote / 32

}
