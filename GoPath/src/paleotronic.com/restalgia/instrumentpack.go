package restalgia

import (
//	"paleotronic.com/fmt"
)

type InstrumentPack struct {
	Frequency  []float64
	Volume     []float64
	Panning    []float64
	Count      int
	Instrument []*Instrument
	Voice      []*Voice
}

func NewInstrumentPack() *InstrumentPack {
	this := &InstrumentPack{}
	this.Instrument = make([]*Instrument, numVOICES)
	this.Voice = make([]*Voice, numVOICES)
	this.Frequency = make([]float64, numVOICES)
	this.Volume = make([]float64, numVOICES)
	this.Panning = make([]float64, numVOICES)
	for x := 0; x < numVOICES; x++ {
		this.Instrument[x] = nil
		this.Frequency[x] = -1.0
		this.Voice[x] = nil
		this.Volume[x] = -1
		this.Panning[x] = -2
	}
	return this
}

func (this *InstrumentPack) Add(v *Voice, i *Instrument, f float64, vol float64, panning float64) {
	this.Instrument[this.Count] = i
	this.Voice[this.Count] = v
	this.Frequency[this.Count] = f
	this.Volume[this.Count] = vol
	this.Panning[this.Count] = panning
	this.Count++
}

func (this *InstrumentPack) Apply() {
	for x := 0; x < numVOICES; x++ {
		if this.Voice[x] == nil {
			return
		}

		if this.Instrument[x] != nil {
			////fmt.Printntf("Apply instrument to voice %d, %v, %v, %v, %v\n", x, this.Instrument[x].params, this.Frequency[x], this.Volume[x], this.Panning[x])
			this.Instrument[x].Apply(this.Voice[x])
		}
		if this.Frequency[x] > 0 {
			this.Voice[x].SetFrequency(this.Frequency[x])
		}
		if this.Volume[x] >= 0 {
			this.Voice[x].SetVolume(this.Volume[x])
		}
		if this.Panning[x] >= -1 {
			this.Voice[x].Pan = this.Panning[x]
		}
	}
}
