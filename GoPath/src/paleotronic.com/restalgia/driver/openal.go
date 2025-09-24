// +build shit

package driver

import (
	"time"

	"github.com/vova616/go-openal/openal"
	//"paleotronic.com/fmt"
	//"math"
	"paleotronic.com/restalgia"
	"paleotronic.com/core/settings"
)

var (
	//SampleRate int
	startTime  time.Time
)

const (
	num_buffers = 32
)

type output struct {
	ch chan []int16
	// Openal state here
	device  *openal.Device
	context *openal.Context
	source  openal.Source
	buffers [num_buffers]openal.Buffer
	nextbuf int
	stopped bool
	started time.Time
	sf      func()
	mix     *restalgia.Mixer
}

func get(sampleRate, channels int) (Output, error) {

	o := output{
		device: openal.OpenDevice(""),
		ch:     make(chan []int16, 2),
	}

	o.context = o.device.CreateContext()
	o.context.Activate()
	o.source = openal.NewSource()
	o.source.SetPitch(1)
	o.source.SetGain(1)
	o.source.SetPosition(0, 0, 0)
	o.source.SetVelocity(0, 0, 0)
	o.source.SetLooping(false)

	for i := 0; i < num_buffers; i++ {
		o.buffers[i] = openal.NewBuffer()
	}

	settings.SampleRate = sampleRate

	go o.start()
	return &o, nil
}

func (o *output) Push(samples []float32) {
	cdata := make([]int16, len(samples))
	for i, v := range samples {
		cdata[i] = int16(32767 * v)
	}
	o.ch <- cdata
}

func UpSample(in []float32, current, target int, channels int) []float32 {
	if current == target {
		return in
	}
	ratio := (float32(channels) * float32(target)) / float32(current)
	// how much we need in the outbuffer
	insize := len(in)
	outsize := int(ratio * float32(insize))
	//	//fmt.Println(outsize)
	out := make([]float32, outsize)

	for i, _ := range out {
		out[i] = in[int(float32(i)/ratio)]
	}

	return out
}

func (o *output) PullFloats() []int16 {

	if o.mix == nil {
		return []int16(nil)
	}

	var buffer []float32 = make([]float32, 576)

	o.mix.FillStereo(buffer)

	is := make([]int16, len(buffer))
	for i, v := range buffer {

		//~ if v < 0 {
		//~ is[i] = int16( 0x8000*v )
		//~ } else {
		//~ is[i] = int16( 0x7fff*v )
		//~ }

		is[i] = int16(v * 32767)

	}

	return is
}

func (o *output) start() {

	o.started = time.Now()

	var b openal.Buffer

	for !o.stopped {

		n := o.source.BuffersProcessed()
		q := o.source.BuffersQueued()
		if n > 0 {
			// free buffer to fill
			//log.Println("Recycled buffer")

			//~ if len(o.ch) == 0 && o.mix != nil {
			//~ //fmt.Println("fill")
			//~ buff := make([]float32, 384)
			//~ o.mix.FillStereo(buff)
			//~ o.Push(buff)
			//~ }

			//~ if len(o.ch) == 0 && o.sf != nil && time.Since(o.started) > 5*time.Second {
			//~ o.sf()
			//~ o.started = time.Now()
			//~ }

			//chunk := <-o.ch
			chunk := o.PullFloats()

			if len(chunk) != 0 {
				b = o.source.UnqueueBuffer()
				b.SetDataInt(openal.FormatStereo16, chunk, int32(SampleRate))
				o.source.QueueBuffer(b)
				//log.Printf( "Queued %d bytes of real data\n", b.GetSize() )

				if o.source.State() != openal.Playing {
					o.source.Play()
				}
			}

		} else if q < num_buffers {
			// any buffer

			//log.Println("Using buffer")

			//~ if len(o.ch) == 0 && o.mix != nil {
			//~ //fmt.Println("fill")
			//~ buff := make([]float32, 384)
			//~ o.mix.FillStereo(buff)
			//~ o.Push(buff)
			//~ }

			//~ if len(o.ch) == 0 && o.sf != nil && time.Since(o.started) > 5*time.Second {
			//~ o.sf()
			//~ o.started = time.Now()
			//~ }

			//chunk := <-o.ch
			chunk := o.PullFloats()

			if len(chunk) != 0 {
				b = o.buffers[o.nextbuf]
				o.nextbuf = (o.nextbuf + 1) % num_buffers
				b.SetDataInt(openal.FormatStereo16, chunk, int32(SampleRate))
				o.source.QueueBuffer(b)
				//log.Printf( "Queued %d bytes of real data\n", b.GetSize() )

				if o.source.State() != openal.Playing {
					o.source.Play()
				}
			}

		} else {
			time.Sleep(1 * time.Millisecond)
		}

	}

}

func (o *output) SetPullSource(mix *restalgia.Mixer) {
	o.mix = mix
}

func (o *output) SetStarvationFunc(sf func()) {
	o.sf = sf
}

func (o *output) Stop() {
	o.source.Stop()
	o.stopped = true
}

func (o *output) Start() {
	o.source.Play()
	o.stopped = false
	//	go o.start()
}

func (o *output) play() {
	o.source.Play()
}
