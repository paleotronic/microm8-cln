package restalgia

import (
	"math"
)

const AYMAXDROP = 0.004
const AYENVCYCLES = 1020484

type EnvelopeGeneratorSimple struct {
	SampleRate         int
	TicksSinceTrigger  int
	TicksPerPhase      int
	TicksPerStep       int
	Shape              int
	Enabled            bool
	lastOut            float32
	frequency          float32
	vt                 [32]float32
	MaxRange           float32
	AllowSharpChange   bool
	RequestRecalculate bool
}

func NewEnvelopeGeneratorSimple(sampleRate int) *EnvelopeGeneratorSimple {
	return &EnvelopeGeneratorSimple{
		SampleRate:       sampleRate,
		TicksPerPhase:    1,
		AllowSharpChange: false,
		Shape:            0,
		vt:               buildAYVolTable(),
	}
}

func (egs *EnvelopeGeneratorSimple) Trigger() {
	egs.TicksSinceTrigger = 0
	//fmt.Printf("tpp calc: r: %d, f: %f\n", egs.SampleRate, egs.frequency)
	egs.Recalculate()
	egs.RequestRecalculate = false
}

func (egs *EnvelopeGeneratorSimple) Recalculate() {
	egs.TicksPerPhase = int(float64(egs.SampleRate) / float64(egs.frequency))
	if egs.TicksPerPhase == 0 {
		egs.TicksPerPhase = 1
	}
	egs.TicksPerStep = int(float64(egs.TicksPerPhase) / 32)
	if egs.TicksPerStep == 0 {
		egs.TicksPerStep = 1
	}

}

func (egs *EnvelopeGeneratorSimple) SetFrequency(f float32) {
	if f == egs.frequency {
		return
	}
	//fmt.Printf("setF = %f (was %f)\n", f, egs.frequency)
	if f <= 0 {
		return
	}
	egs.frequency = f
	egs.RequestRecalculate = true
	//egs.Recalculate()
}

func (egs *EnvelopeGeneratorSimple) SetEnabled(b bool) {
	egs.Enabled = b
	if egs.Enabled {
		egs.Trigger()
	}
}

func (egs *EnvelopeGeneratorSimple) SetShape(shape int) {
	egs.Shape = shape

	//if egs.Enabled {
	egs.Trigger()
	//}
}

func (this *EnvelopeGeneratorSimple) Clip(f float32) float32 {
	if this.AllowSharpChange {
		this.lastOut = f
		return f
	}

	if math.Abs(float64(this.lastOut)-float64(f)) > AYMAXDROP {
		if this.lastOut < f {
			this.lastOut += AYMAXDROP
			return this.lastOut
		}
		this.lastOut -= AYMAXDROP
		return this.lastOut
	}

	this.lastOut = f
	return f
}

func (egs *EnvelopeGeneratorSimple) GetAmplitude2f(l int, m float32) []float32 {
	s := make([]float32, l)
	for i, _ := range s {
		if i%2 == 0 {
			s[i] = egs.GetAmplitude(false) * m
		} else {
			s[i] = s[i-1]
		}
	}
	return s
}

func (egs *EnvelopeGeneratorSimple) GetAmplitude(peek bool) float32 {

	if !egs.Enabled {
		return egs.Clip(0)
	}

	if egs.TicksPerPhase < 1 {
		egs.TicksPerPhase = 1
	}

	var amp float32

	switch egs.Shape {
	case 0:
		amp = egs.highFallMin()
	case 1:
		amp = egs.highFallMin()
	case 2:
		amp = egs.highFallMin()
	case 3:
		amp = egs.highFallMin()
	case 4:
		amp = egs.lowRiseMin()
	case 5:
		amp = egs.lowRiseMin()
	case 6:
		amp = egs.lowRiseMin()
	case 7:
		amp = egs.lowRiseMin()
	case 8:
		amp = egs.highFallRepeat()
	case 9:
		amp = egs.highFallMin()
	case 10:
		amp = egs.fallRiseRepeat()
	case 11:
		amp = egs.highFallMax()
	case 12:
		amp = egs.lowRiseRepeat()
	case 13:
		amp = egs.lowRiseMax()
	case 14:
		amp = egs.riseFallRepeat()
	case 15:
		amp = egs.lowRiseMin()
	}

	idx := int(amp * 31)
	amp = egs.Clip(egs.vt[idx])

	if peek {
		return amp
	}

	egs.TicksSinceTrigger++
	//log.Printf("tst = %d, tpp = %d", egs.TicksSinceTrigger, egs.TicksPerPhase)
	if egs.TicksSinceTrigger%(2*egs.TicksPerPhase) == 0 && egs.RequestRecalculate {
		//fmt.Println("recalc")
		egs.Trigger()
	}

	return amp
}

func (egs *EnvelopeGeneratorSimple) highFallMin() float32 {

	// [\_____]

	if egs.TicksSinceTrigger < egs.TicksPerPhase {
		psc := float32(1) / float32(egs.TicksPerPhase)
		return float32(1) - psc*float32(egs.TicksSinceTrigger)
	}

	return 0
}

func (egs *EnvelopeGeneratorSimple) lowRiseMin() float32 {

	// [/_____]

	if egs.TicksSinceTrigger < egs.TicksPerPhase {
		psc := float32(1) / float32(egs.TicksPerPhase)
		return psc * float32(egs.TicksSinceTrigger)
	}

	return 0
}

func (egs *EnvelopeGeneratorSimple) lowRiseMax() float32 {

	// [/_____]

	if egs.TicksSinceTrigger < egs.TicksPerPhase {
		psc := float32(1) / float32(egs.TicksPerPhase)
		return psc * float32(egs.TicksSinceTrigger)
	}

	return 1
}

func (egs *EnvelopeGeneratorSimple) highFallMax() float32 {

	// [\-----]

	if egs.TicksSinceTrigger < egs.TicksPerPhase {
		psc := float32(1) / float32(egs.TicksPerPhase)
		return 1 - psc*float32(egs.TicksSinceTrigger)
	}

	return 1
}

func (egs *EnvelopeGeneratorSimple) highFallRepeat() float32 {

	// [\\\\\\]
	psc := float32(1) / float32(egs.TicksPerPhase)
	return float32(1) - psc*float32(egs.TicksSinceTrigger%egs.TicksPerPhase)
}

func (egs *EnvelopeGeneratorSimple) lowRiseRepeat() float32 {

	// [//////]
	psc := float32(1) / float32(egs.TicksPerPhase)
	return psc * float32(egs.TicksSinceTrigger%egs.TicksPerPhase)
}

func (egs *EnvelopeGeneratorSimple) fallRiseRepeat() float32 {

	// [\/\/\/]
	ticks := egs.TicksSinceTrigger % (2 * egs.TicksPerPhase)
	psc := float32(1) / float32(egs.TicksPerPhase)

	if ticks < egs.TicksPerPhase {
		return 1 - psc*float32(ticks%egs.TicksPerPhase)
	}

	return psc * float32(ticks%egs.TicksPerPhase)

}

func (egs *EnvelopeGeneratorSimple) riseFallRepeat() float32 {
	// [\/\/\/]
	//fmt.Printf("tst=%d, tps=%d, 2*tps=%d\n", egs.TicksSinceTrigger, egs.TicksPerPhase, 2*egs.TicksPerPhase)
	ticks := egs.TicksSinceTrigger % (2 * egs.TicksPerPhase)
	psc := float32(1) / float32(egs.TicksPerPhase)

	if ticks < egs.TicksPerPhase {
		return psc * float32(ticks%egs.TicksPerPhase)
	}

	return 1 - psc*float32(ticks%egs.TicksPerPhase)

}

func buildAYVolTable() [32]float32 {
	var t [32]float32
	out := float32(1)
	for i := 31; i > 0; i-- {
		t[i] = out
		out /= 1.188502227
	}
	t[0] = 0
	return t
}
