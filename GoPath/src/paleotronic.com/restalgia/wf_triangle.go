package restalgia

type WaveformTRIANGLE struct {
	Waveform
}

func NewWaveformTRIANGLE(osc *Oscillator) *WaveformTRIANGLE {
	this := &WaveformTRIANGLE{}
	this.Waveform = NewWaveform(osc)
	// TODO Auto-generated constructor stub
	return this
}

func (this *WaveformTRIANGLE) ValueForInputSignal(z float64) float64 {
	f := (this.Oscillator.Frequency + this.Oscillator.dFrequency) * this.Oscillator.FrequencyMultiplier

	numHarmonics := this.Oscillator.SampleRate / (4 * int(f))
	if numHarmonics > sqMaxH {
		numHarmonics = sqMaxH
	}

	var amp float64
	var fac float64
	var val float64
	for n := 1; n <= numHarmonics; n++ {
		fac = 2*float64(n) - 1
		amp = float64(1) / float64(fac*fac)
		switch n % 2 {
		case 0:
			val += amp * Sin(fac*z)
		case 1:
			val -= amp * Sin(fac*z)
		}

	}

	return val
}

func (this *WaveformTRIANGLE) BitValue() WAVEFORM {
	// TODO Auto-generated method stub
	return TRIANGLE
}
