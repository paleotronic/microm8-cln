package common

type AYTimedGenerator struct {
	sampleRate      int
	clock           int64
	period          int
	counter         float64
	cyclesPerSample float64
	clocksPerPeriod int
}

func NewAYTimedGenerator(clock int64, sampleRate int) *AYTimedGenerator {
	tg := &AYTimedGenerator{}
	tg.SetPeriod(1)
	tg.SetRate(clock, sampleRate)
	tg.Reset()
	return tg
}

func (tg *AYTimedGenerator) StepsPerCycle() int {
	return 1
}

func (tg *AYTimedGenerator) SetRate(clock int64, sampleRate int) {
	if sampleRate == 0 {
		tg.sampleRate = 48000
	} else {
		tg.sampleRate = sampleRate
	}
	tg.clock = clock
	tg.cyclesPerSample = float64(clock) / float64(sampleRate)
}

func (tg *AYTimedGenerator) SetPeriod(period int) {
	tg.period = period
	if period == 0 {
		tg.period = 1
	}
	tg.clocksPerPeriod = (tg.period * tg.StepsPerCycle())
}

func (tg *AYTimedGenerator) UpdateCounter() int {
	tg.counter += tg.cyclesPerSample
	numStateChanges := 0
	for tg.counter >= float64(tg.clocksPerPeriod) {
		tg.counter -= float64(tg.clocksPerPeriod)
		numStateChanges++
	}
	return numStateChanges
}

func (tg *AYTimedGenerator) Reset() {
	tg.counter = 0
	tg.period = 1
}
