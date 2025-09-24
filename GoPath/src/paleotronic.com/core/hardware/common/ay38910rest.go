package common

import (
	"paleotronic.com/log"

	"paleotronic.com/core/hardware/restalgia"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/fmt"
)

type AYRestalgiaEngine struct {
	index     int
	memindex  int
	Int       interfaces.Interpretable
	clock     int64
	buffer    [][]string
	toneport  [3]int
	noiseport int
	lastPan   float64
}

func NewAYRestalgiaEngine(ent interfaces.Interpretable, index int, clock int64) *AYRestalgiaEngine {
	r := &AYRestalgiaEngine{
		index:    index,
		memindex: ent.GetMemIndex(),
		Int:      ent,
		clock:    clock,
		buffer:   make([][]string, 0),
	}
	r.Build()
	return r
}

func (c *AYRestalgiaEngine) Build() {

	for t, _ := range c.toneport {
		vname := c.GetToneName(t)
		c.toneport[t] = c.Int.GetAudioPort(vname)
		fmt.Printf("*** Tone port %s = 0x%.2x\n", vname, c.toneport[t])
	}
	c.noiseport = c.Int.GetAudioPort(c.GetNoiseName())
	fmt.Printf("*** Noise port %s = 0x%.2x\n", c.GetNoiseName(), c.noiseport)

	restalgia.CommandI(c.Int, c.noiseport, restalgia.RF_setIsColour, 1)
	restalgia.CommandI(c.Int, c.noiseport, restalgia.RF_setEnvelopeAttack, 0)
	restalgia.CommandI(c.Int, c.noiseport, restalgia.RF_setEnvelopeDecay, 0)
	restalgia.CommandI(c.Int, c.noiseport, restalgia.RF_setEnvelopeSustain, 0)
	restalgia.CommandI(c.Int, c.noiseport, restalgia.RF_setEnvelopeRelease, 0)
	restalgia.CommandF(c.Int, c.noiseport, restalgia.RF_setFrequency, 1)
	restalgia.CommandF(c.Int, c.noiseport, restalgia.RF_setVolume, 1)

	for _, tp := range c.toneport {
		restalgia.CommandI(c.Int, tp, restalgia.RF_setColour, c.noiseport)
		restalgia.CommandF(c.Int, tp, restalgia.RF_setColourRatio, 0.0)
		restalgia.CommandI(c.Int, tp, restalgia.RF_setEnvelopeAttack, 0)
		restalgia.CommandI(c.Int, tp, restalgia.RF_setEnvelopeDecay, 0)
		restalgia.CommandI(c.Int, tp, restalgia.RF_setEnvelopeSustain, 0)
		restalgia.CommandI(c.Int, tp, restalgia.RF_setEnvelopeRelease, 0)
		restalgia.CommandF(c.Int, tp, restalgia.RF_setFrequency, 1)
		restalgia.CommandF(c.Int, tp, restalgia.RF_setVolume, 1)
		// if c.index == 0 {
		// 	restalgia.CommandF(c.Int, tp, restalgia.RF_setPan, -0.5)
		// } else {
		// 	restalgia.CommandF(c.Int, tp, restalgia.RF_setPan, 0.5)
		// }
	}

	c.CheckPan()

}

func (c *AYRestalgiaEngine) Done() {
	// TODO
}

func (c *AYRestalgiaEngine) GetToneName(index int) string {
	return fmt.Sprintf("psgtones%dc%dv%d", 0, c.index, index)
}

func (c *AYRestalgiaEngine) GetNoiseName() string {
	return fmt.Sprintf("psgnoise%dc%d", 0, c.index)
}

func (c *AYRestalgiaEngine) CheckPan() {
	var intended float64
	switch c.index {
	case 0:
		intended = settings.MockingBoardPSG0Bal
	case 1:
		intended = settings.MockingBoardPSG1Bal
	}

	if intended != c.lastPan {
		for _, tp := range c.toneport {
			restalgia.CommandF(c.Int, tp, restalgia.RF_setPan, intended)
		}
		c.lastPan = intended
	}
}

func (c *AYRestalgiaEngine) SetToneVolume(voice int, level int) {
	lv := AYVolTable[2*(level&0xf)]
	restalgia.CommandF(c.Int, c.toneport[voice], restalgia.RF_setVolume, float64(lv))
}

func (c *AYRestalgiaEngine) SetToneFreq(voice int, period int) {
	if period == 0 {
		period = 1
	}
	//fr := (float64(c.clock) / 16) / float64(period)
	fr := float64(c.clock) / (16 * float64(period))
	log.Printf("Period of %d -> frequency of %f", period, fr)
	restalgia.CommandF(c.Int, c.toneport[voice], restalgia.RF_setFrequency, fr)
}

func (c *AYRestalgiaEngine) SetToneEnvEnabled(voice int, b bool) {
	v := 0
	if b {
		v = 1
	}
	restalgia.CommandI(c.Int, c.toneport[voice], restalgia.RF_setEnvShapeEnabled, v)
}

func (c *AYRestalgiaEngine) SetToneEnvState(voice int, p int) {
	restalgia.CommandI(c.Int, c.toneport[voice], restalgia.RF_setEnvShape, p)
}

func (c *AYRestalgiaEngine) SetToneEnvFreq(voice int, period int) {
	if period == 0 {
		period = 1
	}
	//vn := c.GetToneName(voice)
	plo := float64(period & 0xff)
	phi := float64((period & 0xff00) >> 8)
	fr := float64(c.clock) / (256*plo + 65536*phi)

	//log.Printf("Period %d yields frequency %f", period, fr)

	restalgia.CommandF(c.Int, c.toneport[voice], restalgia.RF_setEnvShapeFreq, fr)
}

func (c *AYRestalgiaEngine) SetToneColour(voice int, colour float32) {
	restalgia.CommandF(c.Int, c.toneport[voice], restalgia.RF_setColourRatio, float64(colour))
}

func (c *AYRestalgiaEngine) SetToneEnabled(voice int, b bool) {
	v := 0
	if b {
		v = 1
	}
	restalgia.CommandI(c.Int, c.toneport[voice], restalgia.RF_setEnabled, v)
}

func (c *AYRestalgiaEngine) SetNoiseFreq(period int) {
	period = period + 2
	fr := float64(c.clock) / (16 * float64(period))
	restalgia.CommandF(c.Int, c.noiseport, restalgia.RF_setVolume, 1)
	restalgia.CommandF(c.Int, c.noiseport, restalgia.RF_setFrequency, fr)
}

// func (c *AYRestalgiaEngine) InitVoice(voice int) {
// 	vn := c.index*3 + voice
// 	c.SendRestalgia(
// 		[]string{
// 			fmt.Sprintf("use mixer.voices.trk%d", vn),
// 			fmt.Sprintf("set instrument \"WAVE=TRIANGLE:ADSR=0,0,1000,1:VOLUME=1.0\""),
// 			fmt.Sprintf("set volume %f", 0),
// 		},
// 	)
// }

// func (c *AYRestalgiaEngine) EnableOSC(voice int, osc int, on bool) {
// 	vn := c.index*3 + voice
// 	c.SendRestalgia(
// 		[]string{
// 			fmt.Sprintf("use mixer.voices.trk%d.oscillators.osc%d", vn, osc),
// 			fmt.Sprintf("set enabled %v", on),
// 		},
// 	)
// }

// func (c *AYRestalgiaEngine) SetOSCVolume(voice int, osc int, level float32) {
// 	vn := c.index*3 + voice
// 	c.SendRestalgia(
// 		[]string{
// 			fmt.Sprintf("use mixer.voices.trk%d.oscillators.osc%d", vn, osc),
// 			fmt.Sprintf("set volume %f", level),
// 		},
// 	)
// }

// func (c *AYRestalgiaEngine) SetOSCPeriod(voice int, osc int, period int) {
// 	if period == 0 {
// 		period = 1
// 	}
// 	vn := c.index*3 + voice
// 	fr := (float64(c.clock) / 16) / float64(period)
// 	c.SendRestalgia(
// 		[]string{
// 			fmt.Sprintf("use mixer.voices.trk%d.oscillators.osc%d", vn, osc),
// 			fmt.Sprintf("set frequency %f", fr),
// 		},
// 	)
// }

// func (c *AYRestalgiaEngine) InitVoiceNoise(voice int) {
// 	vn := c.index*3 + voice
// 	c.SendRestalgia(
// 		[]string{
// 			fmt.Sprintf("use mixer.voices.trk%d", vn),
// 			fmt.Sprintf("set instrument \"WAVE=SQUARE:ADSR=0,0,1000,1:VOLUME=1.0;WAVE=NOISE:ADSR=0,0,1000,1:VOLUME=0\""),
// 			fmt.Sprintf("set volume %f", 0),
// 		},
// 	)
// }

// func (c *AYRestalgiaEngine) ChangeInstrument(voice int, inst string) {
// 	vn := c.index*3 + voice
// 	c.SendRestalgia(
// 		[]string{
// 			fmt.Sprintf("use mixer.voices.trk%d", vn),
// 			fmt.Sprintf("set instrument \"%s\"", inst),
// 		},
// 	)
// }

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
