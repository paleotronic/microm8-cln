package restalgia

type AdditiveOperator struct {
	Harmonic  int
	Amplitude float64
}

type WaveformAdditive struct {
	Waveform
	Harmonics    []AdditiveOperator
	MaxHarmonics int
}

func NewWaveformAdditive(osc *Oscillator) *WaveformAdditive {
	this := &WaveformAdditive{}
	this.Waveform = NewWaveform(osc)

	this.Harmonics = []AdditiveOperator{}

	return this
}

func (this *WaveformAdditive) ValueForInputSignal(z float64) float64 {

	m := Sin(z) // fundamental frequency

	var maxscale float64 = 1

	if this.MaxHarmonics == 0 {
		for _, op := range this.Harmonics {
			m = m + (op.Amplitude * Sin(z*float64(op.Harmonic)))
			maxscale += op.Amplitude
		}
	} else {
		mh := this.MaxHarmonics
		if mh > len(this.Harmonics) {
			mh = len(this.Harmonics)
		}
		for _, op := range this.Harmonics[:mh] {
			m = m + (op.Amplitude * Sin(z*float64(op.Harmonic)))
			maxscale += op.Amplitude
		}

	}

	return m / maxscale

}

func (this *WaveformAdditive) BitValue() WAVEFORM {
	// TODO Auto-generated method stub
	return ADDSYNTH
}
