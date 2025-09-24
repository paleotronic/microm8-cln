package beeper

import (
	"runtime"
	"time"
	"sync"
	"math"
	"paleotronic.com/log"
	"paleotronic.com/restalgia/driver"
//	"github.com/go-gl/glfw/v3.1/glfw"
	//"os"
	clock "time"
)

type Speaker struct {
	Stream  driver.Output
	counter float64
	level float32
	idleCycles int
	BUFFER_SIZE int
	MIN_PLAYBACK_BUFFER int
	TICKS_PER_SAMPLE float64
	TICKS_PER_SAMPLE_FLOOR float64
	volume float32
	SampleRate float64
	speakerBit bool
	primaryBuffer []float32
	secondaryBuffer []float32
	bufferPos int
	bufferMutex sync.Mutex
	playbackTimer *time.Ticker
	running bool
}

func NewSpeaker() *Speaker {
	this := &Speaker{}

	var err error
	this.Stream, err = driver.Get(44100, 2, nil)

	if err != nil {
		panic(err)
	}

	this.SampleRate = driver.SampleRate // actual audio rate
	this.BUFFER_SIZE = int( 0.4 * this.SampleRate )
	this.MIN_PLAYBACK_BUFFER = 64

	this.TICKS_PER_SAMPLE = 1000000 / this.SampleRate
	this.TICKS_PER_SAMPLE_FLOOR = math.Floor(this.TICKS_PER_SAMPLE)

	this.Reconfigure()

	log.Printf( "*** Created speaker (rate=%f, buffer=%d, ticks=%f)\n", this.SampleRate, this.BUFFER_SIZE, this.TICKS_PER_SAMPLE )

	//os.Exit(0)

	this.volume = 1

	return this
}

func (this *Speaker) Suspend() {
	if !this.running {
		return
	}
	this.playbackTimer.Stop()
	this.running = false
	this.Stream.Stop()
}

func (this *Speaker) Resume() {
	if this.running {
		return
	}
	this.running = true
	this.playbackTimer = time.NewTicker( 208*time.Millisecond )


	go func() {
		for this.running {
			select {
				case <- this.playbackTimer.C:
					this.PlayCurrentBuffer()
			}
//			if this.bufferPos > 4096 {
//				this.PlayCurrentBuffer()
//			}
		}
	}()

	go func() {
		runtime.LockOSThread()
		clock := clock.NewTimer(18)
		//var i int
		//var maxticks = int(this.TICKS_PER_SAMPLE_FLOOR)
		for this.running {
			this.counter = this.TICKS_PER_SAMPLE
			this.Tick()
			_ = clock.LoopInterval()
		}
	}()

	this.Stream.Start()
}

func (this *Speaker) PlayCurrentBuffer() {
	var buffer []float32
	var length int

	this.bufferMutex.Lock()

	length = this.bufferPos
	buffer = this.primaryBuffer
	this.primaryBuffer = this.secondaryBuffer
	this.bufferPos = 0

	this.bufferMutex.Unlock()

	this.secondaryBuffer = buffer
	this.Stream.Push( buffer[0:length] )

	log.Printf("---> Wrote %d samples\n", length)
}

func (this *Speaker) Tick() {

	if this.speakerBit {
		this.level++
	}

	this.counter += 1
	this.idleCycles++

	if this.counter > this.TICKS_PER_SAMPLE {

		var sample float32

		if this.speakerBit {
			sample = 1
		} else {
			sample = -1
		}

		for this.bufferPos >= len(this.primaryBuffer) {
			time.Sleep(1*time.Microsecond)
		}

		this.bufferMutex.Lock()

			index := this.bufferPos
			this.primaryBuffer[index] = sample
			this.primaryBuffer[index+1] = sample

			this.bufferPos += 2

		this.bufferMutex.Unlock()


		this.level = 0
		this.counter -= this.TICKS_PER_SAMPLE_FLOOR
	}

}

func (this *Speaker) ToggleSpeaker(write bool) {

	if write {
		this.level += 0.2
	} else {
		this.speakerBit = !this.speakerBit
	}
	this.idleCycles = 0
}

func (this *Speaker) Reconfigure() {
	if this.primaryBuffer != nil && this.secondaryBuffer != nil {
		return
	}
	this.BUFFER_SIZE = 20000
	this.primaryBuffer = make([]float32, this.BUFFER_SIZE)
	this.secondaryBuffer = make([]float32, this.BUFFER_SIZE)
}
