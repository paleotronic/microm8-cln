package restalgia

type FMOperator struct {
	A float64
	F float64
}

type WaveformFM struct {
	Waveform
	Data []FMOperator
}

func NewWaveformFM(osc *Oscillator) *WaveformFM {
	this := &WaveformFM{}
	this.Waveform = NewWaveform(osc)
	// TODO Auto-generated constructor stub
	this.Data = []FMOperator{}
	return this
}

func (this *WaveformFM) ValueForInputSignal(z float64) float64 {

	var m float64

	var maxscale float64 = 1

	for _, op := range this.Data {

		m = op.A * Sin(op.F*z+m)

		maxscale += op.A

	}

	return m / maxscale

}

func (this *WaveformFM) BitValue() WAVEFORM {
	// TODO Auto-generated method stub
	return FMSYNTH
}
