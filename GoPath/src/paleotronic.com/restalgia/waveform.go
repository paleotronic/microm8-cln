package restalgia

type WAVEFORM int

const (
	SINE     WAVEFORM = 1 + iota
	SAWTOOTH WAVEFORM = 1 + iota
	SQUARE   WAVEFORM = 1 + iota
	PULSE    WAVEFORM = 1 + iota
	NOISE    WAVEFORM = 1 + iota
	TRIANGLE WAVEFORM = 1 + iota
	CUSTOM   WAVEFORM = 1 + iota
	BUZZER   WAVEFORM = 1 + iota
	TRISAW   WAVEFORM = 1 + iota
	FMSYNTH  WAVEFORM = 1 + iota
	ADDSYNTH WAVEFORM = 1 + iota
	BITPOP   WAVEFORM = 1 + iota
)

func (w WAVEFORM) String() string {
	switch w {
	case SINE:
		return "sine"
	case SAWTOOTH:
		return "sawtooth"
	case SQUARE:
		return "square"
	case PULSE:
		return "pulse"
	case NOISE:
		return "noise"
	case TRIANGLE:
		return "triangle"
	case CUSTOM:
		return "pcm"
	case BUZZER:
		return "buzzer"
	case TRISAW:
		return "trisaw"
	case FMSYNTH:
		return "fm"
	case ADDSYNTH:
		return "additive"
	}
	return "unknown"
}

const (
	bitmapSize = 65536
)

type Waveform struct {
	Oscillator *Oscillator
}

func (this *Waveform) ValueForInputSignal(f float64) float64 {
	return f
}

func (this *Waveform) Value(z float64) float64 {
	//for (z > 2*Math.PI) {
	//	z -= 2*Math.PI;
	//}
	//int index = (int)Math.Round((z / (2*Math.PI)) * bitmapSize);
	//return bitmap[ index % bitmapSize ];
	return this.ValueForInputSignal(z)
}

func (this *Waveform) ValuesForWaveforms(waves []*Waveform, z float64) []float64 {

	// create an array the size we need
	var values []float64 = make([]float64, len(waves)+1)

	idx := 0
	for _, wf := range waves {
		values[idx] = wf.ValueForInputSignal(z)
		idx++
	}
	values[idx] = this.ValueForInputSignal(z)

	return values

}

func (this *Waveform) Average(values []float64) float64 {
	var sum float64 = 0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

//~ func (this *Waveform) InitializeBitmap() {
//~ // calculates one cycle for a waveform
//~ var res float64 = (2 * math.Pi) / float64(bitmapSize)

//~ for x := 0; x < bitmapSize; x++ {
//~ this.bitmap[x] = this.ValueForInputSignal(res * float64(x))
//~ }
//~ }

func NewWaveform(osc *Oscillator) Waveform {
	this := Waveform{}
	//this.bitmap = make([]float64, bitmapSize)
	this.Oscillator = osc
	//this.InitializeBitmap()
	return this
}

func (this *Waveform) GetOscillator() *Oscillator {
	return this.Oscillator
}

func (this *Waveform) SetOscillator(oscillator *Oscillator) {
	this.Oscillator = oscillator
}

func (this *Waveform) Stimulate(buffer []float32) {
	// do nothing here
}

func (this *Waveform) Combinate(waves []*Waveform, z float64) float64 {
	// now we average all the rest
	var values []float64 = this.ValuesForWaveforms(waves, z)

	return this.Average(values)
}
