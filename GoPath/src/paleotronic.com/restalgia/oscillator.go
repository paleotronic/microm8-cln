package restalgia

import (
	"strings"
	//"paleotronic.com/fmt"
	"math"
	//	"os"
	//	"time"
)

type Oscillator struct {
	*RInfo
	Enabled             bool
	PulseWidth          float64
	dPulseWidth         float64
	wfCUSTOM            *WaveformCUSTOM
	wfBUZZER            *WaveformBUZZER
	envelope            *EnvelopeGenerator
	RealignFrequency    bool
	samplesPerCycle     float64
	wfPULSE             *WaveformPULSE
	wfSQUARE            *WaveformSQUARE
	wfNOISE             *WaveformNOISE
	wfADDITIVE          *WaveformAdditive
	wfFMSYNTH           *WaveformFM
	FrequencyMultiplier float64
	waveform            WAVEFORM
	phase               float64
	dPhase              float64
	wfSINE              *WaveformSINE
	wfTRIANGLE          *WaveformTRIANGLE
	wfTRISAW            *WaveformTRISAW
	wfSAWTOOTH          *WaveformSAWTOOTH
	Volume              float64
	dVolume             float64
	Frequency           float64
	dFrequency          float64
	sampleNum           float64
	SampleRate          int
	w                   Waveformer
	cc                  float64
	sc                  float64
}

func StringToWAVEFORM(s string) WAVEFORM {
	switch strings.ToLower(s) {
	case "sine":
		return SINE
	case "pulse":
		return PULSE
	case "noise":
		return NOISE
	case "sawtooth":
		return SAWTOOTH
	case "square":
		return SQUARE
	case "buzzer":
		return BUZZER
	case "custom":
		return CUSTOM
	case "wave":
		return CUSTOM
	case "additive":
		return ADDSYNTH
	case "fm":
		return FMSYNTH
	}
	return SINE
}

func (this *Oscillator) Toggle() {
	//wfCUSTOM.Stimulate();
}

func (this *Oscillator) SetFrequency(frequency float64) {
	this.Frequency = frequency
	this.RealignFrequency = true
	//this.RecalcFrequency()
	//this.SampleNum = 0;
}

func (this *Oscillator) SetFrequencySilent(frequency2 float64) {
	this.Frequency = frequency2
	this.RealignFrequency = true
	//this.RecalcFrequency()
	//this.SampleNum = 0;
}

func (this *Oscillator) SetSampleRate(sampleRate int) {
	this.SampleRate = sampleRate
	this.RealignFrequency = true
	//this.RecalcFrequency()
	// TODO: stuff
}

func (this *Oscillator) GetSampleNum() float64 {
	return this.sampleNum
}

func (this *Oscillator) GetWfSINE() *WaveformSINE {
	return this.wfSINE
}

func (this *Oscillator) GetwfADDITIVE() *WaveformAdditive {
	return this.wfADDITIVE
}

func (this *Oscillator) GetwfFMSYNTH() *WaveformFM {
	return this.wfFMSYNTH
}

func (this *Oscillator) GetWfSQUARE() *WaveformSQUARE {
	return this.wfSQUARE
}

func (this *Oscillator) SetWfSQUARE(wfSQUARE *WaveformSQUARE) {
	this.wfSQUARE = wfSQUARE
}

func (this *Oscillator) SetWfSAWTOOTH(wfSAWTOOTH *WaveformSAWTOOTH) {
	this.wfSAWTOOTH = wfSAWTOOTH
}

func (this *Oscillator) SetWfTRIANGLE(wfTRIANGLE *WaveformTRIANGLE) {
	this.wfTRIANGLE = wfTRIANGLE
}

func (this *Oscillator) SetVolume(volume float64) {
	this.Volume = volume
}

func (this *Oscillator) GetWaveform() WAVEFORM {
	return this.waveform
}

func (this *Oscillator) SetPulseWidthRadians(w float64) {
	this.PulseWidth = w
}

func (this *Oscillator) SetFrequencyMultiplier(freqMult float64) {
	this.FrequencyMultiplier = freqMult
}

func (this *Oscillator) GetPhase() float64 {
	return this.phase
}

func (this *Oscillator) SetWfNOISE(wfNOISE *WaveformNOISE) {
	this.wfNOISE = wfNOISE
}

func (this *Oscillator) SetwfADDITIVE(wfADDITIVE *WaveformAdditive) {
	this.wfADDITIVE = wfADDITIVE
}

func (this *Oscillator) SetwfFMSYNTH(wfFMSYNTH *WaveformFM) {
	this.wfFMSYNTH = wfFMSYNTH
}

func (this *Oscillator) GetSamplesPerCycle() float64 {
	return this.samplesPerCycle
}

func (this *Oscillator) GetWfBUZZER() *WaveformBUZZER {
	return this.wfBUZZER
}

func (this *Oscillator) GetWfCUSTOM() *WaveformCUSTOM {
	return this.wfCUSTOM
}

func (this *Oscillator) GetWfPULSE() *WaveformPULSE {
	return this.wfPULSE
}

func (this *Oscillator) GetWfTRIANGLE() *WaveformTRIANGLE {
	return this.wfTRIANGLE
}

func (this *Oscillator) GetEnvelope() *EnvelopeGenerator {
	return this.envelope
}

func (this *Oscillator) Trigger() {
	this.envelope.Trigger()
}

func (this *Oscillator) GetAmplitude() float32 {

	eamp := this.envelope.GetAmplitude(false)

	if !this.Enabled || this.Volume+this.dVolume == 0 || eamp == 0 {
		return 0.0
	}

	// return the amplitude at a given moment -- get one "sample"
	this.sc++
	if int(this.sc) > this.SampleRate {
		this.sc = 0
		this.cc = 0
	}

	this.sampleNum++
	if float64(this.sampleNum) >= this.samplesPerCycle {
		this.cc++
		this.sampleNum -= this.samplesPerCycle
	}

	if this.RealignFrequency {
		this.RecalcFrequency()
		this.RealignFrequency = false
	}

	cyclePoint := ((this.sampleNum/this.samplesPerCycle)*2*math.Pi + (this.phase + this.dPhase))
	if cyclePoint > 2*math.Pi {
		cyclePoint -= 2 * math.Pi
	}

	if this.w == nil {
		panic("Error initializing restalgia")
	}

	vv := float32(this.w.ValueForInputSignal(cyclePoint))

	for vv == 999 {
		vv = float32(this.w.ValueForInputSignal(cyclePoint))
	}

	return vv * eamp * float32(this.Volume+this.dVolume)
}

func (this *Oscillator) SetEnabled(enabled bool) {
	this.Enabled = enabled
	this.Trigger()
}

func (this *Oscillator) GetPulseWidth() int64 {
	return int64(math.Floor((0.5 * (this.PulseWidth / (2 * math.Pi))) * 255))
}

func (this *Oscillator) GetWfSAWTOOTH() *WaveformSAWTOOTH {
	return this.wfSAWTOOTH
}

func NewOscillator(label string, wf WAVEFORM, freq int, sr int, on bool) *Oscillator {
	this := &Oscillator{}
	this.Frequency = float64(freq)
	this.FrequencyMultiplier = 1.0
	this.SampleRate = sr
	this.waveform = wf
	this.Enabled = on

	this.envelope = NewEnvelopeGenerator(sr, 0, 0, 0, 0)

	// Create our Waveforms
	this.wfNOISE = NewWaveformNOISE(this)
	this.wfSINE = NewWaveformSINE(this)
	this.wfSAWTOOTH = NewWaveformSAWTOOTH(this)
	this.wfTRIANGLE = NewWaveformTRIANGLE(this)
	this.wfPULSE = NewWaveformPULSE(this)
	this.wfCUSTOM = NewWaveformCUSTOM(this)
	this.wfBUZZER = NewWaveformBUZZER(this)
	this.wfTRISAW = NewWaveformTRISAW(this)
	this.wfADDITIVE = NewWaveformAdditive(this)
	this.wfFMSYNTH = NewWaveformFM(this)
	this.wfSQUARE = NewWaveformSQUARE(this)
	this.SetWaveform(wf)
	this.initFields(label)

	this.RecalcFrequency()
	return this
}

func (this *Oscillator) initFields(label string) {
	this.RInfo = NewRInfo(label, "Oscillator")
	this.RInfo.AddAttribute(
		"frequency",
		"float64",
		false,
		func(index int) interface{} {
			return this.Frequency
		},
		func(index int, v interface{}) {
			this.Frequency = v.(float64)
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
}

func (this *Oscillator) GetSamplesForDuration(seconds float64) int64 {
	return int64(math.Floor(float64(this.SampleRate) * seconds))
}

func (this *Oscillator) GetFrequency() float64 {
	return this.Frequency
}

func (this *Oscillator) RecalcFrequency() {
	nspc := float64(this.SampleRate) / ((this.Frequency + this.dFrequency) * this.FrequencyMultiplier)
	this.samplesPerCycle = nspc
}

func (this *Oscillator) GetSampleRate() int {
	return this.SampleRate
}

func (this *Oscillator) SetSampleNum(sampleNum float64) {
	this.sampleNum = sampleNum
}

func (this *Oscillator) SetWfSINE(wfSINE *WaveformSINE) {
	this.wfSINE = wfSINE
}

func (this *Oscillator) GetVolume() float64 {
	return this.Volume
}

func (this *Oscillator) SetWaveform(waveform WAVEFORM) {
	this.waveform = waveform

	switch this.waveform {
	case TRIANGLE:
		this.w = this.wfTRIANGLE
		break
	case CUSTOM:
		this.w = this.wfCUSTOM
		break
	case FMSYNTH:
		this.w = this.wfFMSYNTH
		break
	case ADDSYNTH:
		this.w = this.wfADDITIVE
		break
	case NOISE:
		this.w = this.wfNOISE
		break
	case SINE:
		this.w = this.wfSINE
		break
	case TRISAW:
		this.w = this.wfTRISAW
		break
	case SQUARE:
		this.w = this.wfSQUARE
		break
	case PULSE:
		this.w = this.wfPULSE
		break
	case SAWTOOTH:
		this.w = this.wfSAWTOOTH
		break
	case BUZZER:
		this.w = this.wfBUZZER
		break
	}
}

func (this *Oscillator) GetPulseWidthRadians() float64 {
	return this.PulseWidth + this.dPulseWidth
}

func (this *Oscillator) SetPhaseShift(shift float64) {
	this.phase = shift
}

func (this *Oscillator) SetPhase(phase float64) {
	this.phase = phase
}

func (this *Oscillator) GetWfNOISE() *WaveformNOISE {
	return this.wfNOISE
}

func (this *Oscillator) GetFrequencyMultiplier() float64 {
	return this.FrequencyMultiplier
}

func (this *Oscillator) SetSamplesPerCycle(samplesPerCycle float64) {
	this.samplesPerCycle = samplesPerCycle
}

func (this *Oscillator) SetWfCUSTOM(wfCUSTOM *WaveformCUSTOM) {
	this.wfCUSTOM = wfCUSTOM
}

func (this *Oscillator) SetWfPULSE(wfPULSE *WaveformPULSE) {
	this.wfPULSE = wfPULSE
}

func (this *Oscillator) SetEnvelope(envelope *EnvelopeGenerator) {
	this.envelope = envelope
}

func (this *Oscillator) IsEnabled() bool {
	return this.Enabled
}

func (this *Oscillator) SetPulseWidth(pulseWidth float64) {
	this.PulseWidth = pulseWidth
}
