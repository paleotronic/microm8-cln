package common

//var VolTable [16]int = buildVolTable()

type AYChannel struct {
	*AYTimedGenerator
	noiseActive                 bool
	inverted                    bool
	amplitude                   int
	useEnvGen                   bool
	active                      bool
	r                           *AYRestalgiaEngine
	index                       int
	nperiod                     int
	nmult                       float32
	lastInst                    string
	lastActive, lastNoiseActive bool
}

func NewAYChannel(clock int64, sampleRate int, index int, r *AYRestalgiaEngine) *AYChannel {
	c := &AYChannel{
		AYTimedGenerator: NewAYTimedGenerator(clock, sampleRate),
		index:            index,
		r:                r,
		nmult:            1,
	}
	//	c.r.InitVoiceNoise(c.index)
	c.SetPeriod(1)
	c.checkInst()
	return c
}

func (c *AYChannel) StepsPerCycle() int {
	return 8
}

func (c *AYChannel) SetAmplitude(amp int) {
	c.useEnvGen = (amp & 0x010) != 0
	c.r.SetToneEnvEnabled(c.index, c.useEnvGen)
	if !c.useEnvGen {
		c.amplitude = (amp & 0x0F)
		c.r.SetToneVolume(c.index, amp&0xf)
	}
}

func (c *AYChannel) SetEnvPeriod(p int) {
	c.r.SetToneEnvFreq(c.index, p)
}

func (c *AYChannel) SetEnvShape(p int) {
	c.r.SetToneEnvState(c.index, p)
}

func (c *AYChannel) SetActive(active bool) {
	c.active = active
	c.checkInst()
}

func (c *AYChannel) SetNoiseActive(active bool) {
	c.noiseActive = active
	c.checkInst()
}

func (c *AYChannel) SetNoisePeriod(period int) {
	c.nperiod = period
	c.checkInst()
}

func (c *AYChannel) SetPeriod(period int) {
	c.period = period
	c.AYTimedGenerator.SetPeriod(period)
	c.r.SetToneFreq(c.index, period)
}

func (c *AYChannel) Reset() {
	c.AYTimedGenerator.Reset()
	c.amplitude = 0
	c.useEnvGen = false
	c.active = false
	c.noiseActive = false
	c.inverted = false
	c.checkInst()
	c.r.SetToneEnvEnabled(c.index, c.useEnvGen)
	c.r.SetToneVolume(c.index, c.amplitude)
	c.r.SetToneColour(c.index, 0.0)
}

func (c *AYChannel) checkInst() {

	c.r.CheckPan()

	if c.active == c.lastActive && c.noiseActive == c.lastNoiseActive {
		return
	}

	//	fmt.Printf("v%d: n=%v t=%v\n", c.index, c.noiseActive, c.active)

	switch {
	case c.active && !c.noiseActive:
		c.r.SetToneEnabled(c.index, true)
		c.r.SetToneColour(c.index, 0.0)
		c.r.SetToneFreq(c.index, c.period)
		//c.r.SetToneVolume(c.index, c.amplitude)
	case c.active && c.noiseActive:
		c.r.SetToneEnabled(c.index, true)
		c.r.SetToneColour(c.index, 0.5)
		c.r.SetToneFreq(c.index, c.period)
		//c.r.SetToneVolume(c.index, c.amplitude)
	case !c.active && c.noiseActive:
		c.r.SetToneEnabled(c.index, true)
		c.r.SetToneColour(c.index, 1.0)
		//c.r.SetToneVolume(c.index, c.amplitude)
	case !c.active && !c.noiseActive:
		c.r.SetToneEnabled(c.index, false)
		c.r.SetToneColour(c.index, 0.0)
		//c.r.SetToneVolume(c.index, 0) // mute channel
	}

	c.lastActive, c.lastNoiseActive = c.active, c.noiseActive

}
