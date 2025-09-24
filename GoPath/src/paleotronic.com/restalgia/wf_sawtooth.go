package restalgia

type WaveformSAWTOOTH struct {
	Waveform
}

func NewWaveformSAWTOOTH(osc *Oscillator) *WaveformSAWTOOTH {
	this := &WaveformSAWTOOTH{}
	this.Waveform = NewWaveform(osc)
	// TODO Auto-generated constructor stub
	return this
}

func (this *WaveformSAWTOOTH) ValueForInputSignal(z float64) float64 {

	f := (this.Oscillator.Frequency + this.Oscillator.dFrequency) * this.Oscillator.FrequencyMultiplier

	numHarmonics := this.Oscillator.SampleRate / (2 * int(f))
	if numHarmonics > sqMaxH {
		numHarmonics = sqMaxH
	}

	var amp float64
	var fac float64
	var val float64
	for n := 1; n <= numHarmonics; n++ {
		amp = float64(1) / float64(n)
		fac = float64(n)
		val += amp * Sin(fac*z)
	}

	return val

}

func (this *WaveformSAWTOOTH) BitValue() WAVEFORM {
	// TODO Auto-generated method stub
	return SAWTOOTH
}
