package restalgia

//import	"paleotronic.com/fmt"
//	"time"
//	"math"
//import "log"

type SampleBuffer struct {
	Data []float32
	Rate int
}

type WaveformCUSTOM struct {
	Waveform
	buffer      []float32
	bufferrate  int
	buffptr     int
	LateSamples int
	loop        bool
	Playing     bool
	SamplesLeft int
	lastval     float64
	bufferqueue chan SampleBuffer
	Blocking    bool
}

func (this *WaveformCUSTOM) Drop() {
	for len(this.bufferqueue) > 0 {
		_ = <-this.bufferqueue
	}
}

func (this *WaveformCUSTOM) ValueForInputSignal(z float64) float64 {

	var v float64 = this.lastval

	// handle new queued buffer thing
	if this.buffer == nil || this.buffptr >= len(this.buffer) {
		this.buffer = nil
		if len(this.bufferqueue) > 0 {
			bb := <-this.bufferqueue
			this.buffer = resampleIfNeeded(bb.Data, bb.Rate, this.Waveform.GetOscillator().GetSampleRate())
			//this.buffer = resampleInterpolateIfNeeded( bb.Data, bb.Rate, this.Waveform.GetOscillator().GetSampleRate() )
			this.bufferrate = bb.Rate
			this.buffptr = 0
		}
	}

	if this.buffer != nil {

		this.LateSamples = 0

		if this.buffptr >= len(this.buffer) && this.loop {
			this.buffptr = 0 // for looping sound
			this.Playing = true
			this.SamplesLeft = len(this.buffer) - this.buffptr
		}

		if this.buffptr < len(this.buffer) {
			v = float64(this.buffer[this.buffptr])
			this.buffptr++
			this.Playing = true
			this.SamplesLeft = len(this.buffer) - this.buffptr
		} else {
			this.Playing = false
			this.SamplesLeft = 0
		}
	} else {
		this.Playing = false
		this.SamplesLeft = 0
		this.LateSamples++
		//log2.Printf("late samples %d", this.LateSamples)
	}

	this.lastval = v

	return v
}

func (this *WaveformCUSTOM) BitValue() WAVEFORM {
	return CUSTOM
}

func (this *WaveformCUSTOM) Waiting() int {
	return len(this.bufferqueue)
}

func (this *WaveformCUSTOM) Stimulate(b []float32, loop bool, block bool, rate int) {
	// do something here causes wave to fli

	this.bufferqueue <- SampleBuffer{Data: b, Rate: rate}

	this.loop = false
	this.Playing = true
}

func NewWaveformCUSTOM(osc *Oscillator) *WaveformCUSTOM {
	this := &WaveformCUSTOM{}
	this.buffer = make([]float32, 0)
	this.bufferqueue = make(chan SampleBuffer, 4)
	this.Waveform = NewWaveform(osc)

	return this
}

func resampleIfNeeded(in []float32, current, target int) []float32 {

	if current == target {
		return in
	}

	if current == 0 {
		return in
	}

	ratio := float32(target) / float32(current)
	// how much we need in the outbuffer
	insize := len(in)
	outsize := int(ratio * float32(insize))
	//	//fmt.Println(outsize)

	//fmt.Printf("I'm creating a samplebuffer of size %d (original size was %d, ratio is %f)\n", outsize, insize, ratio)

	out := make([]float32, outsize)

	for i, _ := range out {
		out[i] = in[int(float32(i)/ratio)]
	}

	return out
}

func resampleIfNeededA(in []float32, current, target int) []float32 {

	if current == target {
		return in
	}

	ratio := float32(target) / float32(current)
	// how much we need in the outbuffer
	insize := len(in)
	outsize := int(ratio * float32(insize))
	//	//fmt.Println(outsize)
	out := make([]float32, outsize)

	var avg float32 = 0

	for i, _ := range out {
		v := in[int(float32(i)/ratio)]
		avg = (avg + v) / 2
		out[i] = v
	}

	return out
}

func resampleInterpolateIfNeeded(in []float32, current, target int) []float32 {

	if current == target {
		return in
	}
	ratio := float32(target) / float32(current)
	// how much we need in the outbuffer
	insize := len(in)
	outsize := int(ratio * float32(insize))
	//	//fmt.Println(outsize)
	out := make([]float32, outsize)

	var inptr int
	var infrc float32
	for i, _ := range out {
		inptr = int(float32(i) / ratio)
		infrc = (float32(i) / ratio) - float32(inptr)
		if infrc < 0.3 {
			out[i] = in[inptr]
		} else if infrc > 0.7 && inptr < len(in)-1 {
			out[i] = in[inptr]*0.3 + in[inptr+1]*0.7
		} else if inptr < len(in)-1 {
			out[i] = in[inptr]*0.5 + in[inptr+1]*0.5
		} else {
			out[i] = in[inptr]
		}
	}

	return out
}
