package restalgia

import (
	"math"
)

type WaveformTRISAW struct {
	Waveform
}

func NewWaveformTRISAW(osc *Oscillator) *WaveformTRISAW {
	this := &WaveformTRISAW{}
	this.Waveform = NewWaveform(osc)
	// TODO Auto-generated constructor stub
	return this
}

func (this *WaveformTRISAW) ValueForInputSignal(z float64) float64 {

	a := float64(1)
	p := float64(math.Pi * 2)

	y1 := ((-2 * a) / math.Pi) * Atan(1/Tan((z*math.Pi)/p))
	y2 := ((2 * a) / math.Pi) * Asin(Sin(((2*math.Pi)/(p))*z))

	return (y1 + y2) / 2

}

func (this *WaveformTRISAW) BitValue() WAVEFORM {
	// TODO Auto-generated method stub
	return TRISAW
}
