package common

import "math"

type AYEnvelopeGenerator struct {
	*AYTimedGenerator
	shape              int
	hold               bool
	attk               bool
	alt                bool
	cont               bool
	direction          int
	amplitude          int
	start1high         bool
	start2high         bool
	oneShot            bool
	oddEven            bool
	effectiveAmplitude int
}

func NewAYEnvelopeGenerator(clock int64, sampleRate int) *AYEnvelopeGenerator {
	return &AYEnvelopeGenerator{
		AYTimedGenerator: NewAYTimedGenerator(clock, sampleRate),
	}
}

func (e *AYEnvelopeGenerator) StepsPerCycle() int {
	return 8
}

func (e *AYEnvelopeGenerator) SetPeriod(p int) {
	if p > 0 {
		e.AYTimedGenerator.SetPeriod(p)
	} else {
		e.clocksPerPeriod = e.StepsPerCycle() / 2
	}
}

func (e *AYEnvelopeGenerator) Step() {
	stateChanges := e.UpdateCounter()
	total := 0
	for i := 0; i < stateChanges; i++ {
		e.amplitude += e.direction
		if e.amplitude > 15 || e.amplitude < 0 {
			if e.oddEven {
				e.SetPhase(e.start1high)
			} else {
				e.SetPhase(e.start2high)
			}
			e.oddEven = !e.oddEven
			if e.hold {
				e.direction = 0
			}
		}
		total += e.amplitude
	}
	if stateChanges == 0 {
		e.effectiveAmplitude = e.amplitude
	} else {
		e.effectiveAmplitude = int(math.Min(15, float64(total)/float64(stateChanges)))
	}
}

func xor(a, b bool) bool {
	return a != b
}

func (e *AYEnvelopeGenerator) SetShape(shape int) {
	e.oddEven = false
	e.counter = 0
	e.cont = (shape & 8) != 0
	e.attk = (shape & 4) != 0
	e.alt = (shape & 2) != 0
	e.hold = ((shape ^ 8) & 9) != 0

	e.start1high = !e.attk
	e.start2high = e.cont && !(xor(xor(e.attk, e.alt), e.hold))

	e.SetPhase(e.start1high)
}

func (e *AYEnvelopeGenerator) SetPhase(isHigh bool) {
	if isHigh {
		e.amplitude = 15
		e.direction = -1
	} else {
		e.amplitude = 0
		e.direction = 1
	}
}

func (e *AYEnvelopeGenerator) GetEffectiveAmplitude() int {
	return e.effectiveAmplitude
}

func (e *AYEnvelopeGenerator) GetAmplitude() int {
	return e.amplitude
}

func (e *AYEnvelopeGenerator) Reset() {
	e.AYTimedGenerator.Reset()
	e.SetShape(0)
}
