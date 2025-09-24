package restalgia

import (
	"math"
)

type WaveformPULSE struct {
	Waveform
	pulseWidth float64
}

func NewWaveformPULSE(osc *Oscillator) *WaveformPULSE {
	this := &WaveformPULSE{}
	this.Waveform = NewWaveform(osc)
	this.pulseWidth = math.Pi / 2
	// TODO Auto-generated constructor stub
	return this
}

func (this *WaveformPULSE) ValueForInputSignal(z float64) float64 {

	hcA := this.GetPulseWidth()
	hcB := 2*math.Pi - hcA
	//zpc := z / (2 * math.Pi)

	var zmod float64
	var zpc float64
	if z < hcA {
		zpc = z / (2 * hcA)
		zmod = zpc * 2 * math.Pi
	} else {
		zpc = (z-hcA)/(2*hcB) + 0.5
		zmod = zpc * 2 * math.Pi
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
		val += amp * Sin(fac*zmod)
	}

	return val

}

func (this *WaveformPULSE) GetPulseWidth() float64 {
	return this.GetOscillator().GetPulseWidthRadians()
}

func (this *WaveformPULSE) BitValue() WAVEFORM {
	// TODO Auto-generated method stub
	return PULSE
}

/*
Notes:
	|-------------2pi---------------|

			hca = pw		 hcb=2pi - pw

	+--------------------+
	|                    |
	|                    |
	+					 +          +
						 |          |
						 |          |
						 +----------+

	|-------2/3----------|----1/3---|

	a cycle = 2 * a half cycle
	b cycle = 2 * b half cycle



*/
