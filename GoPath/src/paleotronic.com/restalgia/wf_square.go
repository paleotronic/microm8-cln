package restalgia

import "math"

type WaveformSQUARE struct {
	Waveform
	factor float64
	cheap  bool
}

func NewWaveformSQUARE(osc *Oscillator) *WaveformSQUARE {
	this := &WaveformSQUARE{}
	this.Waveform = NewWaveform(osc)
	this.factor = 4
	this.cheap = true
	// TODO Auto-generated constructor stub
	return this
}

func (this *WaveformSQUARE) clip(v float64) float64 {
	if v < -1 {
		return -1
	}
	if v > 1 {
		return 1
	}
	return v
}

const sqMaxH = 64

func (this *WaveformSQUARE) ValueForInputSignal(z float64) float64 {

	if this.cheap {
		if z < math.Pi {
			return 1
		}
		return -1
	}

	f := (this.Oscillator.Frequency + this.Oscillator.dFrequency) * this.Oscillator.FrequencyMultiplier

	if f == 0 {
		f = 1
	}

	numHarmonics := this.Oscillator.SampleRate / (2 * int(f))
	if numHarmonics > sqMaxH {
		numHarmonics = sqMaxH
	}

	var amp float64
	var fac float64
	var val float64
	for n := 1; n <= numHarmonics; n++ {
		fac = 2*float64(n) - 1
		amp = float64(1) / float64(fac)
		val += amp * Sin(fac*z)
	}

	return val
}

func (this *WaveformSQUARE) GetPulseWidth() float64 {
	return this.GetOscillator().GetPulseWidthRadians()
}

func (this *WaveformSQUARE) BitValue() WAVEFORM {
	// TODO Auto-generated method stub
	return SQUARE
}
