package apple2

import (
	"math"
	"strings"

	"paleotronic.com/core/hardware/common"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
)

//import "paleotronic.com/fmt"

const MAX_SPEAKER_BUFFERS = 2
const IDLE_STOP = 8192

type AppleSpeaker struct {
	SampleRate                           int
	CPUSpeed                             int
	LastCPUSpeed                         int
	TicksPerSample, AdjustTicksPerSample int
	TPSErrorMargin                       float64
	CurrentBuffer                        int
	Buffers                              [MAX_SPEAKER_BUFFERS][]uint64
	HQBuffers                            [MAX_SPEAKER_BUFFERS][]float32
	BufferSize                           int
	SpeakerState                         bool
	AccumulatedCycles                    int
	SampleCount                          int
	SampleNum                            uint64
	SampleCounter                        uint64
	LeapTicksEvery                       uint64
	LeapTickSign                         int
	e                                    interfaces.Interpretable
	sbs                                  *common.SpeakerBitstream
	HQ                                   bool
	HQLevelMax                           *int
	HQLevel                              int
	divfactor                            float64
	idleCycles                           int64
	idleSamples                          int
	nonIdleSamples                       int
	emptyCount                           int
	uid                                  string
	outputFuncNB, outputFunc             func(channel int, data []float32, loop bool, rate int)
	outputFuncCompressed                 func(channel int, data []uint64, loop bool, rate int, raw bool)
	//
	tps        int
	lastsecond int
	channel    int
}

func NewAppleSpeaker(e interfaces.Interpretable, uid string, channel int, sampleRate int, cpuClock int, bufferSize int) *AppleSpeaker {

	s := &AppleSpeaker{
		SampleRate:    sampleRate,
		CPUSpeed:      cpuClock,
		BufferSize:    bufferSize,
		CurrentBuffer: 0,
		e:             e,
		divfactor:     1,
		uid:           uid,
		channel:       channel,
	}

	for i := 0; i < MAX_SPEAKER_BUFFERS; i++ {
		s.Buffers[i] = make([]uint64, bufferSize)
		s.HQBuffers[i] = make([]float32, bufferSize)
	}

	s.SampleNum = 31
	s.SampleCount = 0

	s.CalcTicksPerSample(cpuClock)

	s.HQLevelMax = &settings.SpeakerBitstreamPsuedoLevels

	s.sbs = common.NewSpeakerBitstream(bufferSize, e, sampleRate, &settings.SpeakerBitstreamDiv)

	//fmt.Printf("SPEAKER: TPS = %d, Leapadjust %d every %d samples", s.TicksPerSample, s.LeapTickSign, s.LeapTicksEvery)

	return s
}

func NewAppleSpeakerHQ(e interfaces.Interpretable, uid string, channel int, sampleRate int, cpuClock int, bufferSize int, maxLevels int) *AppleSpeaker {

	s := NewAppleSpeaker(e, uid, channel, sampleRate, cpuClock, bufferSize*32)
	s.HQ = true
	s.HQLevelMax = &maxLevels

	return s
}

func (s *AppleSpeaker) Done() {
	s.Flush()
	s.e = nil
	s.sbs = nil
}

func (s *AppleSpeaker) Bind(of, ofnb func(int, []float32, bool, int), ofc func(int, []uint64, bool, int, bool)) {
	s.outputFunc = of
	s.outputFuncNB = ofnb
	s.outputFuncCompressed = ofc
}

func (s *AppleSpeaker) CalcTicksPerSample(cpuClock int) {

	if cpuClock == s.LastCPUSpeed {
		return
	}

	//fmt.Printf("CPU Speed has changed to %d...\n", cpuClock)

	// TPS is rounded...
	s.TicksPerSample = int(float64(cpuClock) / float64(s.SampleRate))
	//log2.Printf("setting ticks persample to %d", s.TicksPerSample)
	s.AdjustTicksPerSample = s.TicksPerSample
	s.TPSErrorMargin = (float64(cpuClock) / float64(s.SampleRate)) - math.Floor(float64(cpuClock)/float64(s.SampleRate)+0.5)

	s.LeapTicksEvery = uint64(math.Floor(1/math.Abs(s.TPSErrorMargin) + 0.5))
	if s.TPSErrorMargin < 0 {
		s.LeapTickSign = -1
	} else {
		s.LeapTickSign = 1
	}

	s.LastCPUSpeed = cpuClock

}

func (s *AppleSpeaker) AdjustClock(speed int) {
	//log2.Printf("Adjust clock to %d", speed)
	s.sbs.SetDivFactor(1020484 / float64(speed))
	s.divfactor = 1
	s.CPUSpeed = speed
	s.CalcTicksPerSample(s.CPUSpeed)
	s.HQLevel = 0
	if !s.HQ {
		s.sbs.Flush()
	} else {
		s.SampleCount = 0
	}
}

func (s *AppleSpeaker) ResetSpeaker() {
	s.idleCycles = 0
	s.AccumulatedCycles = 0
	s.HQLevel = 0
	s.SpeakerState = false
}

const RAMPUP = 2

func (s *AppleSpeaker) Tick() {
	s.idleCycles++

	s.AccumulatedCycles += 1
	n := s.AdjustTicksPerSample

	if s.SpeakerState {
		s.HQLevel++
	}

	if s.AccumulatedCycles >= n {
		//s.sbs.ramp = 1
		if s.idleCycles >= IDLE_STOP {
			s.SpeakerState = false // off
			s.idleSamples++
			switch s.idleSamples {
			case 0:
				s.sbs.SetRamp(0.98)
			case 1:
				s.sbs.SetRamp(0.96)
			case 2:
				s.sbs.SetRamp(0.92)
			case 3:
				s.sbs.SetRamp(0.84)
			case 4:
				s.sbs.SetRamp(0.68)
			case 5:
				s.sbs.SetRamp(0.36)
			case 6:
				s.sbs.SetRamp(0.20)
			case 7:
				s.sbs.SetRamp(0.12)
			case 8:
				s.sbs.SetRamp(0.08)
			case 9:
				s.sbs.SetRamp(0.06)
			default:
				s.sbs.SetRamp(0)
			}
		} else {
			s.sbs.SetRamp(1)
		}

		s.WriteSample()
		// Decrement
		s.AccumulatedCycles -= n
		// level
		if s.HQLevel > 0 {
			s.HQLevel /= 2
		}
	}
}

func (s *AppleSpeaker) Increment(n int) {
	for i := 0; i < n; i++ {
		s.Tick()
	}
}

func (s *AppleSpeaker) AddHQLevel(n int) {
	s.HQLevel += n
}

func (s *AppleSpeaker) SubHQLevel(n int) {
	s.HQLevel -= n
	if s.HQLevel < 0 {
		s.HQLevel = 0
	}
}

func (s *AppleSpeaker) ResetHQLevel() {
	s.HQLevel = 0
}

func (s *AppleSpeaker) Decrement(n int) {
	s.AccumulatedCycles -= n
}

func (s *AppleSpeaker) ToggleSpeaker(isWrite bool) {

	//fmt2.Printf(".")

	if isWrite {
		s.HQLevel += 2
		s.SpeakerState = !s.SpeakerState
		//fmt.Print("w")
	} else {
		s.SpeakerState = !s.SpeakerState
		//fmt.Print("r")
	}
	s.idleCycles = 0
	s.idleSamples = 0
}

func (s *AppleSpeaker) WriteSample() {

	//	s.CalcTicksPerSample(s.CPUSpeed)

	if s.HQ {
		s.WriteSampleHQ()
	} else {
		s.WriteSamplePacked()
	}
}

// Use new packing method - kudos to Melody!
func (s *AppleSpeaker) WriteSamplePacked() {
	s.SampleCounter++

	if settings.AudioUsesLeapTicks {
		if s.SampleCounter%s.LeapTicksEvery == 0 {
			s.AdjustTicksPerSample = s.TicksPerSample + s.LeapTickSign
		} else {
			s.AdjustTicksPerSample = s.TicksPerSample
		}
	}

	s.sbs.WriteSample(s.HQLevel)
	s.HQLevel = 0
}

func (s *AppleSpeaker) ImA() string {
	return "SPEAKER-" + strings.ToUpper(s.uid)
}

func (s *AppleSpeaker) Flush() {
	if s.HQ {
		if s.SampleCount != 0 || (s.channel == 1 && settings.RecordC020[s.e.GetMemIndex()]) {
			s.outputFunc(s.channel, s.HQBuffers[s.CurrentBuffer][:s.SampleCount], false, s.SampleRate)
			s.SampleCount = 0
			s.CurrentBuffer = (s.CurrentBuffer + 1) % MAX_SPEAKER_BUFFERS
		}
	} else {
		s.sbs.Flush()
	}
}

// WriteSampleCompressed outputs one sample to the current WBuffer
func (s *AppleSpeaker) WriteSampleCompressed() {

	if s.Buffers[s.CurrentBuffer] == nil || len(s.Buffers[s.CurrentBuffer]) == 0 {
		return
	}

	s.SampleCounter++ // keep track of samples for leap samples etc

	if settings.AudioUsesLeapTicks {
		if s.SampleCounter%s.LeapTicksEvery == 0 {
			s.AdjustTicksPerSample = s.TicksPerSample + s.LeapTickSign
		} else {
			s.AdjustTicksPerSample = s.TicksPerSample
		}
	}

	if s.SampleNum == 31 {
		s.Buffers[s.CurrentBuffer][s.SampleCount] = 0 // clean sample
	}

	if s.SpeakerState {
		// if we are flipping the speaker then we need to mark a 1 in the correct bit
		s.Buffers[s.CurrentBuffer][s.SampleCount] = s.Buffers[s.CurrentBuffer][s.SampleCount] | (1 << s.SampleNum)
	}

	if s.SampleNum > 0 {
		s.SampleNum--
	} else {
		s.SampleNum = 31 // new byte
		s.SampleCount++
		if s.SampleCount >= s.BufferSize {
			s.outputFuncCompressed(s.channel, s.Buffers[s.CurrentBuffer], false, s.SampleRate, s.sbs.GetSendRaw())
			s.SampleCount = 0
			// Rotate buffer
			s.CurrentBuffer = (s.CurrentBuffer + 1) % MAX_SPEAKER_BUFFERS
		}
	}

}

func (s *AppleSpeaker) ProcessEvent(name string, addr int, value *uint64, action memory.MemoryAction) (bool, bool) {
	// This will only ever be used for intercepting the speaker 0xC03x range so we don't care
	s.ToggleSpeaker((action == memory.MA_WRITE))

	return false, false
}
