package common

import (
	"bytes"
	"io"
	"math"

	"paleotronic.com/log"

	"github.com/mjibson/go-dsp/wav"
)

type Cassette struct {
	Samples            []float32
	lastSample         float32
	StartPoints        []int
	Position           int
	HighThreshold      float32
	LowThreshold       float32
	ZeroThreshold      float32
	ZeroThresholdCount int
	SampleRate         int
	ClockRate          int64
	CyclesPerSample    float64
	CycleAccumulator   float64
	TapeRunning        bool
	Bit                bool
	LastBitRead        int64
	Globalcycles       int64
}

func NewCassette() *Cassette {
	return &Cassette{
		Samples:            []float32{},
		Position:           0,
		HighThreshold:      -1,
		LowThreshold:       0,
		ZeroThreshold:      0.1,
		ZeroThresholdCount: 4096,
		ClockRate:          1020484,
		SampleRate:         44100,
	}
}

func NewCassetteFromData(data []byte, clocks int64) (*Cassette, error) {
	c := NewCassette()
	b := bytes.NewBuffer(data)
	err := c.LoadWAVE(b, clocks)
	return c, err
}

func cleanupWAVE(in []float32, normalize bool, scale float32, smooth bool, smoothpc float32, sampleRate int) []float32 {
	out := make([]float32, len(in))
	var min, max, ctr float32
	for i, v := range in {
		if min > v {
			min = v
		}
		if max < v {
			max = v
		}
		out[i] = v
	}

	// center wave if needed
	ctr = (min + max) / 2
	if ctr != 0 {
		for i, v := range out {
			out[i] = v - ctr
		}
	}

	// normallize
	if normalize {
		var max float32
		for _, v := range out {
			if math.Abs(float64(v)) > float64(max) {
				max = float32(math.Abs(float64(v)))
			}
		}
		for i, v := range out {
			out[i] = (v / max) * scale
		}
	}

	// scale
	for i, v := range out {
		out[i] = v * scale
	}

	// smooth
	if smooth {
		var minSmooth = int(float32(sampleRate) * smoothpc)
		var sameCount int
		var lastV float32 = -99999476 // junk value
		for i, v := range out {
			if v == lastV {
				sameCount++
			} else {
				if sameCount >= minSmooth {
					end := i - 1 // last value
					start := end - sameCount + 1
					for j := start; j <= end; j++ {
						out[j] = 0 // smooth value
					}
				}
				sameCount = 0
			}
			lastV = v
		}
		if sameCount >= minSmooth {
			end := len(out) - 1 // last value
			start := end - sameCount + 1
			for j := start; j <= end; j++ {
				out[j] = 0 // smooth value
			}
		}
	}

	return out
}

func (c *Cassette) LoadWAVE(r io.Reader, clocks int64) error {

	log.Printf("[tape] data-clk = %d", clocks)

	wr, err := wav.New(r)
	if err != nil {
		return err
	}

	var raw []float32

	raw, _ = wr.ReadFloats(wr.Samples)

	if len(raw)%int(wr.NumChannels) != 0 {
		panic("on the dance floor")
	}

	log.Printf("[tape::loadWAVE] read %d samples", len(raw))
	for len(raw) > 0 && wr.NumChannels > 1 {
		var fl = make([]float32, len(raw)/int(wr.NumChannels))
		for i := 0; i < len(raw); i++ {
			if i%int(wr.NumChannels) == 0 {
				fl[i/int(wr.NumChannels)] = raw[i]
			}
		}

		raw = fl
	}

	raw = cleanupWAVE(raw, true, 1, true, 0.05, int(wr.SampleRate))

	// var min, max float32
	// for _, v := range raw {
	// 	if min > v {
	// 		min = v
	// 	}
	// 	if max < v {
	// 		max = v
	// 	}
	// }
	// zval := (max - min) / 2
	// for i, v := range raw {
	// 	raw[i] = v - zval // center on zero
	// }

	c.Samples = raw
	c.Position = 0
	c.StartPoints = c.FindSilenceIndexes()
	//c.SampleRate = int(wr.SampleRate)
	c.SetRate(clocks, int(wr.SampleRate))
	log.Printf("[tape] loaded wave data comprising %d samples, with %d silences", len(c.Samples), len(c.StartPoints))
	log.Printf("[tape] samplerate = %d, cycles per sample = %f", c.SampleRate, c.CyclesPerSample)
	return nil

}

func (c *Cassette) Begin() {
	c.Position = 0
}

func (c *Cassette) End() {
	c.Position = len(c.Samples)
}

func (c *Cassette) Start() {
	if c.TapeRunning {
		log.Printf("[tape] Tape already running...")
		return
	}
	log.Printf("[tape] starting at position %d", c.Position)
	c.TapeRunning = true
}

func (c *Cassette) Stop() {
	if !c.TapeRunning {
		log.Printf("[tape] tape already stopped")
		return
	}
	log.Printf("[tape] tape stopped at position %d", c.Position)
	c.TapeRunning = false
}

func (c *Cassette) PrevSilence() {
	p := c.Position
	defer func() {
		log.Printf("[tape] tape wound back to position %d", c.Position)
	}()
	for i := len(c.StartPoints) - 1; i >= 0; i-- {
		v := c.StartPoints[i]
		if v < p {
			c.Position = v
			return
		}
	}
	// start of tape
	c.Position = 0
}

func (c *Cassette) NextSilence() {
	p := c.Position
	defer func() {
		log.Printf("[tape] tape wound forward to position %d", c.Position)
	}()
	for i := 0; i < len(c.StartPoints); i++ {
		v := c.StartPoints[i]
		if v > p {
			c.Position = v
			return
		}
	}
	// end of tape
	c.Position = len(c.Samples)
	c.TapeRunning = false
}

func (c *Cassette) SetRate(clock int64, sampleRate int) {
	if sampleRate == 0 {
		c.SampleRate = 48000
	} else {
		c.SampleRate = sampleRate
	}
	log.Printf("[tape] Set CPU clock to %d, sample rate to %d", clock, sampleRate)
	c.ClockRate = clock
	c.CyclesPerSample = float64(clock) / float64(sampleRate)
}

func (c *Cassette) FindSilenceIndexes() []int {
	var silences = []int{}
	var zeroCount int
	var lastZeroStart int
	var inZeroes, isZero bool
	var lastV float32
	var countT = 0
	for i, v := range c.Samples {
		if math.Signbit(float64(v)) != math.Signbit(float64(lastV)) {
			countT++
		}
		isZero = math.Abs(float64(v)) < float64(c.ZeroThreshold)
		if !isZero {
			if inZeroes && zeroCount > c.ZeroThresholdCount {
				silences = append(silences, lastZeroStart)
			}
			inZeroes = false
			lastZeroStart = -1
		} else {
			if !inZeroes {
				lastZeroStart = i // possible zero start here
				inZeroes = true
				zeroCount = 1
			} else {
				zeroCount++ // increment
			}
		}
		lastV = v
	}
	log.Printf("[tape] cassette image contains %d discrete blocks of silence with %d transitions", len(silences), countT)
	// at end we can ignore any silence (end of tape)
	return silences
}

// Sample feed subroutine
func (c *Cassette) PullSample() bool {
	if c.Position < len(c.Samples) {
		c.TapeRunning = true
	}
	//log.Printf("[tape] reading bit state of %v", c.Bit)
	c.LastBitRead = c.Globalcycles
	return c.Bit
}

// Increment winds the clock forward n cycles
func (c *Cassette) Increment(n int) {
	c.Globalcycles += int64(n)
	if c.TapeRunning && c.Globalcycles-c.LastBitRead > 100000 {
		log.Printf("[tape] Stopping the tape")
		c.TapeRunning = false
	}

	if !c.TapeRunning {
		return
	}

	c.CycleAccumulator += float64(n)

	//log.Printf("sample time")
	for c.CycleAccumulator > c.CyclesPerSample {
		c.CycleAccumulator -= c.CyclesPerSample
		c.Position++ // increment a sample
		if c.Position >= len(c.Samples) {
			c.Bit = false
			c.TapeRunning = false
			log.Printf("[tape] tape stopped")
		} else {
			// check sample
			if c.Position%c.SampleRate == 0 {
				log.Printf("[tape] at position %d secs", c.Position/c.SampleRate)
			}
			s := c.Samples[c.Position]
			//
			// if math.Signbit(float64(s)) != math.Signbit(float64(c.lastSample)) {
			// 	c.Bit = !c.Bit
			// }
			c.Bit = math.Signbit(float64(s))
			//if s != 0 {
			c.lastSample = s
			//}
		}
	}
}

// Decrement winds the clock back n cycles
func (c *Cassette) Decrement(n int) {
	// not required in this implementation
}

// AdjustClock is called if we need to change clock spee
func (c *Cassette) AdjustClock(n int) {
	c.SetRate(int64(n), c.SampleRate)
}

// IsA returns the type of cycle counted device
func (c *Cassette) ImA() string {
	return "Cassette Tape"
}
