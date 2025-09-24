package beeper

import (
	"github.com/gordonklaus/portaudio"
	timer "time"
	"paleotronic.com/restalgia/driver"
	"paleotronic.com/core/memory"
	"time"
)

type Beeper struct {
	Stream *portaudio.Stream
	Output driver.Output
	SampleRate float64
	Frequency float64  // Hz
	NewFrequency float64 // 0 if no change
	SamplesPerCycle int
	SampleCount int
	Smooth bool
	Signal float32
	mm *memory.MemoryMap
	lastMonitor uint
	wait time.Duration
	bias time.Duration
	Calibrate bool
}

func NewBeeper(f float64, m *memory.MemoryMap) *Beeper {

	this := &Beeper{mm: m}

	var err error

	h, err := portaudio.DefaultHostApi()
	this.Stream, err = portaudio.OpenStream(portaudio.HighLatencyParameters(nil, h.DefaultOutputDevice), this.FeedStream)

	if err != nil {
		panic(err)
	}

	this.SampleRate = this.Stream.Info().SampleRate // actual audio rate

//	this.Output, err = driver.Get(44100, 2, nil)
//	if err != nil {
//		panic(err)
//	}

//	this.SampleRate = driver.SampleRate
	this.Frequency  = f
	this.NewFrequency = -1
	this.Smooth = true
	this.Signal = -1
	this.wait = time.Duration(1000000000 / this.SampleRate)
	this.bias = 0
	//this.wait = 1000

	//this.Output.Start()

	//go this.DoSound()

	this.Stream.Start()

	return this

}

func (this *Beeper) SetFrequency(f float64) {

	if this.Smooth {
		this.NewFrequency = (this.Frequency + f) / 2
	} else {
		this.NewFrequency = f
	}

}

//func (this *Beeper) Toggle( out []float32 ) {

//	for i, _ := range out {
//		out[i] = this.Signal
//	}

//}

func (this *Beeper) FeedStream( out [][]float32 ) {

	var now = timer.Now()
	var start = now

	for i, _ := range out[0] {
		now = timer.Now()
		out[0][i] = this.GetSample()
		out[1][i] = out[0][i]

		for timer.Since(now) < this.wait-this.bias {
			//
		}
	}

	since := timer.Since(start)
	if since > time.Duration(len(out[0]))*this.wait {
		this.bias = this.bias + 600
	}

}

func (this *Beeper) Toggle() {
	if this.Signal == 0 {
		this.Signal = 1
	} else {
		this.Signal = 0
	}
//	this.Output.Push([]float32{
//		1,1,0,0,1,1,0,0,
//		1,1,0,0,1,1,0,0,
//	})
}

func (this *Beeper) GetSample() float32 {

//	n := this.mm.Data[memory.OCTALYZER_SPEAKER_TOGGLE]//ReadGlobal(memory.OCTALYZER_SPEAKER_TOGGLE)
//	if n != this.lastMonitor {
//		this.Signal = - this.Signal
//	}

//	this.lastMonitor = n

	return this.Signal

}

func (this *Beeper) DoSound() {

	var v float32
	//var now = timer.Now()

	for {
		//now = timer.Now()
		v = this.GetSample()
		this.Output.Push( []float32{ v, v } )
//		for timer.Since(now) < this.wait/2 {
//			//
//		}
	}

}

//func (this *Beeper) GetSample() float32 {

//	if  this.SampleCount == 0 {

//		if this.NewFrequency != -1 {
//			this.Frequency = this.NewFrequency
//			this.NewFrequency = -1
//			if this.Frequency < 1 {
//				this.Frequency = 1
//			}
//		}

//		this.SamplesPerCycle = int(this.SampleRate / this.Frequency)

//	}

//	var v float32

//	if this.SampleCount < (this.SamplesPerCycle/2) {
//		v = -1
//	} else {
//		v = 1
//	}

//	this.SampleCount = (this.SampleCount + 1) % this.SamplesPerCycle

//	if this.Frequency < 20 {
//		return 0
//	}

//	return v

//}
