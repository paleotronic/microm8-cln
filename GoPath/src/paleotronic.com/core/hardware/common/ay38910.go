package common

import (
	"paleotronic.com/core/interfaces"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
)

const AYClockSpeed = 1020484
const AYSamplesBuffer = 4096

var AYVolTable [32]float32 = buildAYVolTable()

type AYRegister int

const (
	AFine AYRegister = iota
	ACoarse
	BFine
	BCoarse
	CFine
	CCoarse
	NoisePeriod
	Enable
	AVol
	BVol
	CVol
	EnvFine
	EnvCoarse
	EnvShape
	PortA
	PortB
	AYMaxReg
)

func (r AYRegister) valid() bool {
	return r >= 0 && r < AYMaxReg
}

var AYRegMax = [AYMaxReg]int{
	255,
	15,
	255,
	15,
	255,
	15,
	31,
	255,
	31,
	31,
	31,
	255,
	255,
	15,
	255,
	255,
}

var AYPrefOrder = []AYRegister{
	Enable, EnvShape, EnvCoarse, EnvFine, NoisePeriod, AVol, BVol, CVol,
	AFine, ACoarse, BFine, BCoarse, CFine, CCoarse,
}

type AYBusControl int

const (
	aybcReset    AYBusControl = 0
	aybcInactive AYBusControl = 4
	aybcRead     AYBusControl = 5
	aybcWrite    AYBusControl = 6
	aybcLatch    AYBusControl = 7
)

func (a AYBusControl) valid() bool {
	return a >= aybcInactive && a <= aybcLatch
}

type AY38910State struct {
	basereg     int
	bus         int
	control     int
	clock       int64
	rate        int
	mask        int
	index       int // index for restalgia
	selectedReg AYRegister
	regValues   [AYMaxReg]int
	cycles      int
}

type AY38910 struct {
	AY38910State
	label             string
	Int               interfaces.Interpretable
	r                 *AYRestalgiaEngine
	channels          [3]*AYChannel
	noiseGenerator    *AYNoiseGenerator
	envelopeGenerator *AYEnvelopeGenerator
}

func NewAY38910(label string, base int, clockspeed int64, sampleRate int, ddrMask int, index int, ent interfaces.Interpretable) *AY38910 {
	a := &AY38910{
		AY38910State: AY38910State{
			basereg: base,
			clock:   clockspeed,
			rate:    sampleRate,
			mask:    ddrMask,
			index:   index,
		},
		Int: ent,
		r:   NewAYRestalgiaEngine(ent, index, clockspeed),
	}
	for i := 0; i < len(a.channels); i++ {
		a.channels[i] = NewAYChannel(clockspeed, sampleRate, i, a.r)
	}
	a.noiseGenerator = NewAYNoiseGenerator(clockspeed, sampleRate, a.r)
	a.envelopeGenerator = NewAYEnvelopeGenerator(clockspeed, sampleRate)
	a.Reset()
	return a
}

func (c *AY38910) AddCycles(cycles int) {
	c.cycles += cycles
}

func (c *AY38910) SetControl(val int) {
	cmd := AYBusControl(val)

	//fmt.Printf("Bus control recvd: 0x%.2x -> chip #%d\n", cmd, c.index)

	if !cmd.valid() && cmd != aybcReset {
		return
	}
	switch cmd {
	case aybcReset:
		c.Reset()
	case aybcInactive:
		break
	case aybcLatch:
		c.selectedReg = AYRegister(c.bus & 0x0f)
		break
	case aybcRead:
		c.bus = c.getReg(c.selectedReg)
		break
	case aybcWrite:
		c.setReg(c.selectedReg, c.bus)
		break
	}
}

func (c *AY38910) ReadRegister(reg AYRegister) int {
	return c.getReg(reg)
}

func (c *AY38910) State() []int {
	var out = make([]int, 16)
	for i, _ := range out {
		out[i] = c.getReg(AYRegister(i))
	}
	return out
}

func (c *AY38910) getReg(reg AYRegister) int {
	if !reg.valid() {
		return -1
	}
	val := c.regValues[int(reg)] & 0xff
	//log.Printf("getReg called with register %d, value %d", reg, val)
	return val
}

func (c *AY38910) setReg(reg AYRegister, val int) {
	log.Printf("setReg called with register %d, value %d", reg, val)
	if reg.valid() {
		//val &= AYRegMax[reg]
		c.regValues[int(reg)] = val
		c.WriteReg(reg, val&0xff)
	}
}

func (c *AY38910) SetBus(val int) {
	c.bus = val
}

func (c *AY38910) GetBaseReg() int {
	return c.basereg
}

func (c *AY38910) GetBus() int {
	return c.bus
}

func (c *AY38910) Reset() {
	for _, r := range AYPrefOrder {
		if r != Enable {
			c.setReg(r, 0)
		} else {
			c.setReg(r, 255)
		}
	}
	c.envelopeGenerator.Reset()
	c.noiseGenerator.Reset()
	for _, ch := range c.channels {
		ch.Reset()
	}
}

func (c *AY38910) WriteReg(r AYRegister, value int) {

	var force bool

	value = value & 0x0ff
	switch r {
	case ACoarse:
		c.channels[0].SetPeriod(c.getReg(AFine) + (c.getReg(ACoarse) << 8))
		//fmt2.Printf("PeriodA = %d\n", c.getReg(AFine)+(c.getReg(ACoarse)<<8))
		//c.SetVoiceFreq(0, c.getReg(AFine)+(c.getReg(ACoarse)<<8))
	case AFine:
		//fmt2.Printf("PeriodA = %d\n", c.getReg(AFine)+(c.getReg(ACoarse)<<8))
		c.channels[0].SetPeriod(c.getReg(AFine) + (c.getReg(ACoarse) << 8))
		//c.SetVoiceFreq(0, c.getReg(AFine)+(c.getReg(ACoarse)<<8))
		break
	case BCoarse:
		c.channels[1].SetPeriod(c.getReg(BFine) + (c.getReg(BCoarse) << 8))
		//c.SetVoiceFreq(1, c.getReg(BFine)+(c.getReg(BCoarse)<<8))
	case BFine:
		//fmt2.Printf("PeriodB = %d\n", c.getReg(BFine)+(c.getReg(BCoarse)<<8))
		c.channels[1].SetPeriod(c.getReg(BFine) + (c.getReg(BCoarse) << 8))
		//c.SetVoiceFreq(1, c.getReg(BFine)+(c.getReg(BCoarse)<<8))
		break
	case CCoarse:
		c.channels[2].SetPeriod(c.getReg(CFine) + (c.getReg(CCoarse) << 8))
		//c.SetVoiceFreq(2, c.getReg(CFine)+(c.getReg(CCoarse)<<8))
	case CFine:
		//fmt2.Printf("PeriodC = %d\n", c.getReg(CFine)+(c.getReg(CCoarse)<<8))
		c.channels[2].SetPeriod(c.getReg(CFine) + (c.getReg(CCoarse) << 8))
		//c.SetVoiceFreq(2, c.getReg(CFine)+(c.getReg(CCoarse)<<8))
		break
	case NoisePeriod:
		// if value == 0 {
		// 	value = 32
		// }
		//fmt.Printf("PeriodN = %d\n", value+16)
		c.noiseGenerator.SetPeriod(value)
		//c.channels[0].SetNoisePeriod(value + 16)
		//c.channels[1].SetNoisePeriod(value + 16)
		//c.channels[2].SetNoisePeriod(value + 16)
		c.noiseGenerator.counter = 0
		break
	case Enable:
		// fmt.Printf(
		// 	"Enable = At=%v, An=%v, Bt=%v, Bn=%v, Ct=%v, Cn=%v\n",
		// 	(value&1) == 0,
		// 	(value&8) == 0,
		// 	(value&2) == 0,
		// 	(value&16) == 0,
		// 	(value&4) == 0,
		// 	(value&32) == 0,
		// )
		c.channels[0].SetActive((value & 1) == 0)
		c.channels[0].SetNoiseActive((value & 8) == 0)
		c.channels[1].SetActive((value & 2) == 0)
		c.channels[1].SetNoiseActive((value & 16) == 0)
		c.channels[2].SetActive((value & 4) == 0)
		c.channels[2].SetNoiseActive((value & 32) == 0)
		force = true
		break
	case AVol:
		//fmt.Printf("AmplitudeA = %d\n", value)
		c.channels[0].SetAmplitude(value)
		//c.SetVoiceVolume(0, value)
		break
	case BVol:
		//fmt.Printf("AmplitudeB = %d\n", value)
		c.channels[1].SetAmplitude(value)
		//c.SetVoiceVolume(1, value)
		break
	case CVol:
		//	fmt.Printf("AmplitudeC = %d\n", value)
		c.channels[2].SetAmplitude(value)
		//c.SetVoiceVolume(2, value)
		break
	case EnvFine:
		c.envelopeGenerator.SetPeriod(c.getReg(EnvFine) + 256*c.getReg(EnvCoarse))
		c.channels[0].SetEnvPeriod(c.getReg(EnvFine) + 256*c.getReg(EnvCoarse))
		c.channels[1].SetEnvPeriod(c.getReg(EnvFine) + 256*c.getReg(EnvCoarse))
		c.channels[2].SetEnvPeriod(c.getReg(EnvFine) + 256*c.getReg(EnvCoarse))
	case EnvCoarse:
		//fmt.Printf("PeriodEnv = %d\n", c.getReg(EnvFine)+(c.getReg(EnvCoarse)<<8))
		c.envelopeGenerator.SetPeriod(c.getReg(EnvFine) + 256*c.getReg(EnvCoarse))
		c.channels[0].SetEnvPeriod(c.getReg(EnvFine) + 256*c.getReg(EnvCoarse))
		c.channels[1].SetEnvPeriod(c.getReg(EnvFine) + 256*c.getReg(EnvCoarse))
		c.channels[2].SetEnvPeriod(c.getReg(EnvFine) + 256*c.getReg(EnvCoarse))
		break
	case EnvShape:
		//fmt.Printf("Shape of Env = %d\n", value)
		c.envelopeGenerator.SetShape(value)
		c.channels[0].SetEnvShape(value)
		c.channels[1].SetEnvShape(value)
		c.channels[2].SetEnvShape(value)
		break
	case PortA:
	case PortB:
		break
	}

	c.DumpChip()

	if c.cycles > 300 || force {
		//c.r.FlushRestalgia()
		c.cycles = 0
	}
}

func (c *AY38910) SetRate(clock int64, sampleRate int) {
	for _, ch := range c.channels {
		ch.SetRate(clock, sampleRate)
	}
	c.noiseGenerator.SetRate(clock, sampleRate)
	c.envelopeGenerator.SetRate(clock, sampleRate)
	c.Reset()
}

func (c *AY38910) DumpChip() {
	return
	fmt.Printf("+-[ Chip #%d ]----------------------------------------------------+\n", c.index)
	fmt.Printf(
		"%5s %5s %5s %5s %5s %5s\n",
		"Atone", "Anois",
		"Btone", "Bnois",
		"Ctone", "Cnois",
	)
	value := c.getReg(Enable)
	fmt.Printf(
		"%5v %5v %5v %5v %5v %5v\n\n",
		(value&1) == 0,
		(value&8) == 0,
		(value&2) == 0,
		(value&16) == 0,
		(value&4) == 0,
		(value&32) == 0,
	)
	fmt.Printf(
		"%5s %5s %5s\n",
		"AVol", "BVol", "CVol",
	)
	fmt.Printf(
		"%5d %5d %5d\n\n",
		c.getReg(AVol),
		c.getReg(BVol),
		c.getReg(CVol),
	)
	fmt.Printf(
		"%5s %5s %5s\n",
		"AEnv", "BEnv", "CEnv",
	)
	fmt.Printf(
		"%5v %5v %5v\n\n",
		c.getReg(AVol)&16 != 0,
		c.getReg(BVol)&16 != 0,
		c.getReg(CVol)&16 != 0,
	)
	fmt.Printf(
		"%5s %5s %5s %5s\n",
		"APer", "BPer", "CPer", "NPer",
	)
	fmt.Printf(
		"%5d %5d %5d %5d\n",
		c.getReg(ACoarse)*256+c.getReg(AFine),
		c.getReg(BCoarse)*256+c.getReg(BFine),
		c.getReg(CCoarse)*256+c.getReg(CFine),
		c.getReg(NoisePeriod),
	)
	fmt.Printf("+----------------------------------------------------------------+\n\n")
}
