package restalgia

import (
	"math"
	"math/rand"

	"paleotronic.com/utils"
)

type WaveformNOISE struct {
	Waveform
	lastValue float64
	lastZ     float64
	scale     float64
	altMethod bool
}

func NewWaveformNOISE(osc *Oscillator) *WaveformNOISE {
	this := &WaveformNOISE{altMethod: true}
	this.Waveform = NewWaveform(osc)
	return this
}

func (this *WaveformNOISE) ValueForInputSignal(z float64) float64 {

	if !this.altMethod {

		var chance float64 = this.GetOscillator().GetFrequency() / 2000

		var r float64 = rand.Float64()

		var amp float64 = this.lastValue

		if r < chance {

			amp = rand.Float64()*2 - 1
			this.lastValue = amp

		}

		return amp
	} else {
		if this.lastZ > z {
			this.lastValue = 2*utils.Random() - 1
		} else if this.lastZ < math.Pi && z >= math.Pi {
			this.lastValue = 2*utils.Random() - 1
		}

		this.lastZ = z

		return this.lastValue
	}
}

func (this *WaveformNOISE) BitValue() WAVEFORM {
	// TODO Auto-generated method stub
	return NOISE
}
