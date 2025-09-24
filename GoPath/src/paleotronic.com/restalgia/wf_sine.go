package restalgia

//"paleotronic.com/fmt"

type WaveformSINE struct {
	Waveform
}

func (this *WaveformSINE) BitValue() WAVEFORM {
	// TODO Auto-generated method stub
	return SINE
}

func NewWaveformSINE(osc *Oscillator) *WaveformSINE {
	this := &WaveformSINE{}
	this.Waveform = NewWaveform(osc)
	// TODO Auto-generated constructor stub
	return this
}

func (this *WaveformSINE) ValueForInputSignal(z float64) float64 {
	// TODO Auto-generated method stub
	//////fmt.Println(z, math.Sin(z))
	return Sin(z)
}
