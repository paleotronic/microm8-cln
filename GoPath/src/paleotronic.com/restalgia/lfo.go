package restalgia

import (
	"strings"
)

type LFOControlType int

const (
	LFO_NONE       LFOControlType = iota
	LFO_VOLUME     LFOControlType = iota
	LFO_FREQUENCY  LFOControlType = iota
	LFO_PHASESHIFT LFOControlType = iota
	LFO_FREQMULT   LFOControlType = iota
	LFO_PULSEWIDTH LFOControlType = iota
	LFO_HIPASS     LFOControlType = iota
	LFO_LOPASS     LFOControlType = iota
	LFO_PAN        LFOControlType = iota
)

func (t LFOControlType) String() string {
	switch t {
	case LFO_NONE:
		return "NONE"
	case LFO_VOLUME:
		return "VOLUME"
	case LFO_FREQUENCY:
		return "FREQUENCY"
	case LFO_PHASESHIFT:
		return "PHASESHIFT"
	case LFO_FREQMULT:
		return "FREQMULT"
	case LFO_PULSEWIDTH:
		return "PULSEWIDTH"
	case LFO_HIPASS:
		return "HIPASS"
	case LFO_LOPASS:
		return "LOPASS"
	case LFO_PAN:
		return "PAN"
	}
	return "NONE"
}

func getLFOType(s string) LFOControlType {
	switch strings.ToUpper(s) {
	case "NONE":
		return LFO_NONE
	case "VOLUME":
		return LFO_VOLUME
	case "FREQUENCY":
		return LFO_FREQUENCY
	case "FREQMULT":
		return LFO_FREQMULT
	case "PHASESHIFT":
		return LFO_PHASESHIFT
	case "PULSEWIDTH":
		return LFO_PULSEWIDTH
	case "HIPASS":
		return LFO_HIPASS
	case "LOPASS":
		return LFO_LOPASS
	case "PAN":
		return LFO_PAN
	}
	return LFO_NONE
}

type LFO struct {
	*Oscillator         // Oscillator providing the frequency
	BoundVoice   *Voice // Voice this LFO belongs to
	BoundControl LFOControlType
	MaxDeviation float64
}

// NewLFO creates a new LFO instance...
func NewLFO(wf WAVEFORM, frequency int, voice *Voice, control LFOControlType, dev float64) *LFO {

	this := &LFO{
		Oscillator:   NewOscillator("lfo", wf, frequency, voice.sampleRate, true),
		BoundVoice:   voice,
		BoundControl: control,
		MaxDeviation: dev,
	}

	this.Oscillator.Volume = 1

	this.initFields("lfo")

	return this
}

func (this *LFO) GetVariance() float64 {
	return this.MaxDeviation * float64(this.GetAmplitude())
}

func (this *LFO) initFields(label string) {
	this.RInfo = NewRInfo(label, "LFO")
	this.RInfo.AddAttribute(
		"frequency",
		"float64",
		false,
		func(index int) interface{} {
			return this.Frequency
		},
		func(index int, v interface{}) {
			this.SetFrequency(v.(float64))
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"freqmult",
		"float64",
		false,
		func(index int) interface{} {
			return this.FrequencyMultiplier
		},
		func(index int, v interface{}) {
			this.FrequencyMultiplier = v.(float64)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"waveform",
		"WAVEFORM",
		false,
		func(index int) interface{} {
			return this.waveform
		},
		func(index int, v interface{}) {
			this.SetWaveform(v.(WAVEFORM))
			//this.Trigger()
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"enabled",
		"bool",
		false,
		func(index int) interface{} {
			return this.Enabled
		},
		func(index int, v interface{}) {
			this.Enabled = v.(bool)
			if this.Enabled {
				this.SetVolume(1)
				this.Trigger()
			} else {
				this.SetVolume(0)
			}
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"pulsewidth",
		"float64",
		false,
		func(index int) interface{} {
			return this.PulseWidth
		},
		func(index int, v interface{}) {
			this.PulseWidth = v.(float64)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"phase",
		"float64",
		false,
		func(index int) interface{} {
			return this.phase
		},
		func(index int, v interface{}) {
			this.phase = v.(float64)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"binding",
		"string",
		false,
		func(index int) interface{} {
			return this.BoundControl.String()
		},
		func(index int, v interface{}) {
			this.SetBoundControl(v.(string))
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"ratio",
		"float64",
		false,
		func(index int) interface{} {
			return this.MaxDeviation
		},
		func(index int, v interface{}) {
			this.MaxDeviation = v.(float64)
		},
		func() int {
			return 1
		},
	)
}

func (this *LFO) SetBoundControl(s string) {
	c := getLFOType(s)
	this.BoundControl = c
}
