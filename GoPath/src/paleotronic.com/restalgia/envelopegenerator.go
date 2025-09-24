package restalgia

import (
	"math"
)

//"paleotronic.com/fmt"

type PHASE int

const (
	ATTACK  PHASE = 1 + iota
	DECAY   PHASE = 1 + iota
	SUSTAIN PHASE = 1 + iota
	RELEASE PHASE = 1 + iota
)

const (
	DECAYRATE = 3.0
	MAXDROP   = 0.01
)

type EnvelopeState int

const (
	esNONE EnvelopeState = iota
	esATTACK
	esDECAY
	esHIGH
	esLOW
	esLOOP0
	esLOOP1
	esLOOP2
	esLOOP3
)

func (e EnvelopeState) String() string {
	switch e {
	case esATTACK:
		return "A"
	case esDECAY:
		return "D"
	case esHIGH:
		return "h"
	case esLOW:
		return "l"
	}
	return ""
}

var EnvStateModes = [][]EnvelopeState{
	[]EnvelopeState{esDECAY, esLOW, esLOOP1}, // \_______
	[]EnvelopeState{esDECAY, esLOW, esLOOP1},
	[]EnvelopeState{esDECAY, esLOW, esLOOP1},
	[]EnvelopeState{esDECAY, esLOW, esLOOP1},
	[]EnvelopeState{esATTACK, esLOW, esLOOP1}, // /_______
	[]EnvelopeState{esATTACK, esLOW, esLOOP1},
	[]EnvelopeState{esATTACK, esLOW, esLOOP1},
	[]EnvelopeState{esATTACK, esLOW, esLOOP1},
	[]EnvelopeState{esDECAY, esLOOP0},           // \|\|\|\|
	[]EnvelopeState{esDECAY, esLOW, esLOOP1},    // \_______
	[]EnvelopeState{esDECAY, esATTACK, esLOOP0}, // \/\/\/\/
	[]EnvelopeState{esDECAY, esHIGH, esLOOP1},   // \|------
	[]EnvelopeState{esATTACK, esLOOP0},          // /|/|/|/|
	[]EnvelopeState{esATTACK, esHIGH, esLOOP1},  // \|\|\|\|
	[]EnvelopeState{esATTACK, esDECAY, esLOOP0}, // /\/\/\/\
	[]EnvelopeState{esATTACK, esLOW, esLOOP1},   // /_______
}

type EnvelopeGenerator struct {
	*RInfo
	Decay             int
	attackGradient    float64
	Sustain           int
	offsetMS          float64
	msPerSample       float64
	releaseGradient   float64
	sustainVolume     float64
	Attack            int
	decayGradient     float64
	SampleRate        int
	Release           int
	triggerTime       int64
	Enabled           bool
	AllowSharpChange  bool
	lastOut           float32
	State             int
	StateIndex        int
	StateFrequency    float64
	stateSampleCount  int
	stateSamplePeriod int
}

func NewEnvelopeGenerator(sampleRate int, attack int, decay int, sustain int, release int) *EnvelopeGenerator {
	this := &EnvelopeGenerator{}
	this.Attack = attack
	this.Decay = decay
	this.Sustain = sustain
	this.Release = release
	this.SampleRate = sampleRate
	this.AllowSharpChange = false
	this.State = -1
	this.initFields("envelopegenerator")

	this.RecalculateEnvelope()
	return this
}

func (this *EnvelopeGenerator) initFields(label string) {
	this.RInfo = NewRInfo(label, "EnvelopeGenerator")
	this.RInfo.AddAttribute(
		"enabled",
		"bool",
		false,
		func(index int) interface{} {
			return this.Enabled
		},
		func(index int, v interface{}) {
			//fmt.Printf("env.enabled=%v\n", v.(bool))
			this.Enabled = v.(bool)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"attack",
		"int",
		false,
		func(index int) interface{} {
			return this.Attack
		},
		func(index int, v interface{}) {
			this.Attack = v.(int)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"decay",
		"int",
		false,
		func(index int) interface{} {
			return this.Decay
		},
		func(index int, v interface{}) {
			this.Decay = v.(int)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"sustain",
		"int",
		false,
		func(index int) interface{} {
			return this.Sustain
		},
		func(index int, v interface{}) {
			this.Sustain = v.(int)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"release",
		"int",
		false,
		func(index int) interface{} {
			return this.Release
		},
		func(index int, v interface{}) {
			this.Release = v.(int)
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"state",
		"int",
		false,
		func(index int) interface{} {
			return this.State
		},
		func(index int, v interface{}) {
			//fmt.Printf("env.state=%v\n", v.(int))
			this.SetState(v.(int))
		},
		func() int {
			return 1
		},
	)
	this.RInfo.AddAttribute(
		"statefrequency",
		"float64",
		false,
		func(index int) interface{} {
			return this.StateFrequency
		},
		func(index int, v interface{}) {
			//fmt.Printf("env.statefrequency=%v\n", v.(float64))
			this.SetStateFrequency(v.(float64))
		},
		func() int {
			return 1
		},
	)
	// this.RInfo.AddAttribute(
	// 	"adsr",
	// 	"[]int",
	// 	true,
	// 	func(index int) interface{} {
	// 		switch index {
	// 		case 0:
	// 			return this.Attack
	// 		case 1:
	// 			return this.Decay
	// 		case 2:
	// 			return this.Sustain
	// 		case 3:
	// 			return this.Release
	// 		}
	// 		return this.Attack
	// 	},
	// 	func(index int, v interface{}) {
	// 		i := v.(int)
	// 		switch index {
	// 		case 0:
	// 			this.Attack = i
	// 		case 1:
	// 			this.Decay = i
	// 		case 2:
	// 			this.Sustain = i
	// 		case 3:
	// 			this.Release = i
	// 		}
	// 	},
	// 	func() int {
	// 		return 4
	// 	},
	// )

}

func (this *EnvelopeGenerator) ResetState() {
	this.stateSamplePeriod = int(float64(this.SampleRate) / (2 * this.StateFrequency))
	this.stateSampleCount = 0
	this.StateIndex = 0
}

func (this *EnvelopeGenerator) SetState(state int) {
	this.State = state
	//fmt.Printf("state=%d\n", this.State)
	if this.State != -1 {
		this.ResetState()
	}
}

func (this *EnvelopeGenerator) SetStateFrequency(f float64) {
	this.StateFrequency = f
	this.ResetState()
}

func (this *EnvelopeGenerator) SetEnvelope(values []int) {
	this.Attack = values[0]
	this.Decay = values[1]
	this.Sustain = values[2]
	this.Release = values[3]

	this.RecalculateEnvelope()
}

func (this *EnvelopeGenerator) RecalculateEnvelope() {

	// based on the sample rate, work out ms per sample
	this.msPerSample = 1000.0 / float64(this.SampleRate)

	if this.Attack > 0 {
		this.attackGradient = 1.0 / float64(this.Attack) // how many ms to peak volume?
	} else {
		this.attackGradient = 0 // instant peak volume
	}

	if this.Decay > 0 {
		this.decayGradient = 1.0 / (float64(this.Decay) * DECAYRATE)
		this.sustainVolume = 1.0 - (float64(this.Decay) * this.decayGradient)
	} else {
		this.decayGradient = 0
		this.sustainVolume = 1.0
	}

	if this.Release > 0 {
		this.releaseGradient = this.sustainVolume / float64(this.Release)
	} else {
		this.releaseGradient = 0
	}

	//System.Out.Println( "ATTACK Gradient = "+attackGradient+" per ms" );
	//System.Out.Println( "DECAY Gradient = "+decayGradient+" per ms" );
	//System.Out.Println( "RELEASE Gradient = "+releaseGradient+" per ms" );
	//	////fmt.Printf("AG = %v, DG = %v, RG = %v\n", this.attackGradient, this.decayGradient, this.releaseGradient)
}

func (this *EnvelopeGenerator) Trigger() {
	// start envelope now
	this.Enabled = true
	this.RecalculateEnvelope()
	this.offsetMS = 0
	this.ResetState()
}

func (this *EnvelopeGenerator) Clip(f float32) float32 {
	if this.AllowSharpChange {
		this.lastOut = f
		return f
	}

	if math.Abs(float64(this.lastOut)-float64(f)) > MAXDROP {
		if this.lastOut < f {
			this.lastOut += MAXDROP
			return this.lastOut
		}
		this.lastOut -= MAXDROP
		return this.lastOut
	}

	this.lastOut = f
	return f
}

func (this *EnvelopeGenerator) GetStateAmplitude(peek bool) float32 {

	perSample := float32(1) / float32(this.stateSamplePeriod)

	phase := EnvStateModes[this.State][this.StateIndex]
	//fmt.Printf("%s", phase.String())

	var amp float32

	switch phase {
	case esATTACK:
		amp = perSample * (float32(this.stateSampleCount) / float32(this.stateSamplePeriod))
	case esDECAY:
		amp = 1 - (perSample * (float32(this.stateSampleCount) / float32(this.stateSamplePeriod)))
	case esHIGH:
		amp = 1
	case esLOW:
		amp = 0
	}

	if !peek {
		// increment
		this.stateSampleCount++
		if this.stateSampleCount >= this.stateSamplePeriod {
			this.StateIndex++
			if this.StateIndex >= len(EnvStateModes[this.State]) {
				this.StateIndex = 0
			}
			// check for loops
			newphase := EnvStateModes[this.State][this.StateIndex]
			if newphase >= esLOOP0 && newphase <= esLOOP3 {
				this.StateIndex = int(newphase - esLOOP0)
			}
		}
	}

	//fmt.Printf("(%f), ", amp)

	return amp

}

func (this *EnvelopeGenerator) GetAmplitude(peek bool) float32 {

	if !this.Enabled {
		return 0.0
	}

	if this.State >= 0 {
		return this.GetStateAmplitude(peek)
	}

	var tst float64 = this.offsetMS
	var offset float64 = tst

	var current PHASE = ATTACK

	if tst >= float64(this.Attack) {
		current = DECAY
		offset = tst - float64(this.Attack)
	}
	if tst >= float64(this.Attack+this.Decay) {
		current = SUSTAIN
		offset = tst - float64(this.Attack) - float64(this.Decay)
	}
	if tst >= float64(this.Attack+this.Decay+this.Sustain) {
		current = RELEASE
		offset = tst - float64(this.Attack) - float64(this.Decay) - float64(this.Sustain)
	}

	var amp float64 = 1.0
	switch current {
	case ATTACK:
		amp = offset * this.attackGradient
		break
	case DECAY:
		amp = 1.0 - (this.decayGradient * offset)
		break
	case SUSTAIN:
		amp = this.sustainVolume
		break
	case RELEASE:
		amp = this.sustainVolume - (this.releaseGradient * offset)
		break
	}

	//if (current == PHASE.ATTACK)
	//System.Out.Println("Volume is "+(100*amp)+" at offset "+Math.Round(offsetMS)+"|| "+current);

	if amp < 0 {
		amp = 0
	}
	if amp > 1 {
		amp = 1
	}

	if !peek {
		this.offsetMS += this.msPerSample
	}

	return this.Clip(float32(amp))
}

func (this *EnvelopeGenerator) Collapse() {
	// TODO Auto-generated method stub
	this.Enabled = false
	this.lastOut = 0
}
