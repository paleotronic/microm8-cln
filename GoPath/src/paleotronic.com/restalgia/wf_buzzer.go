package restalgia

import ()

type WaveformBUZZER struct {
	Waveform
	buffer    []float64
	buffptr   int
	lastValue float64
}

func (this *WaveformBUZZER) ValueForInputSignal(z float64) float64 {

	return this.lastValue
}

func (this *WaveformBUZZER) BitValue() WAVEFORM {
	return BUZZER
}

func (this *WaveformBUZZER) Stimulate() {
	switch {
	case this.lastValue > 0:
		this.lastValue = -1.0
	default:
		this.lastValue = 1.0
	}
}

func NewWaveformBUZZER(osc *Oscillator) *WaveformBUZZER {
	this := &WaveformBUZZER{}
	this.buffer = make([]float64, 0)
	this.Waveform = NewWaveform(osc)
	return this
}
