package common

const rngSeed = 0x003333
const BIT17 = 0x010000

type AYNoiseGenerator struct {
	*AYTimedGenerator
	rng   int
	state bool
	r     *AYRestalgiaEngine
}

func NewAYNoiseGenerator(clock int64, sampleRate int, r *AYRestalgiaEngine) *AYNoiseGenerator {
	return &AYNoiseGenerator{
		rng:              rngSeed,
		AYTimedGenerator: NewAYTimedGenerator(clock, sampleRate),
		r:                r,
	}
}

func (n *AYNoiseGenerator) SetPeriod(p int) {
	n.r.SetNoiseFreq(p)
}

func (n *AYNoiseGenerator) StepsPerCycle() int {
	return 4
}

func (n *AYNoiseGenerator) Step() {
	stateChanges := n.UpdateCounter()
	for i := 0; i < stateChanges; i++ {
		n.UpdateRng()
	}
}

func (n *AYNoiseGenerator) UpdateRng() {
	if (n.rng & 1) != 0 {
		n.rng = (n.rng ^ 0x24000) >> 1
	} else {
		n.rng = n.rng >> 1
	}
	if (n.rng & 1) == 1 {
		n.state = !n.state
	}
}

func (n *AYNoiseGenerator) IsOn() bool {
	return n.state
}
