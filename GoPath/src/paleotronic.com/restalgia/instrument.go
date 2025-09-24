package restalgia

import (
	"math"
	"strconv"
	"strings"

	"paleotronic.com/fmt"
)

type StringParams map[string]string

type Instrument struct {
	oNum   int
	params [numOSC]StringParams
}

func (this *Instrument) Apply(v *Voice) {
	// apply settings to a voice

	// strip filters
	v.FILT = make([]Filter, 0)

	if v.LFO != nil {
		v.LFO.BoundVoice = nil
		v.LFO = nil
	}

	for x := 0; x < numOSC; x++ {

		osc := v.GetOSC(x)

		osc.dFrequency = 0
		osc.dPhase = 0
		osc.dPulseWidth = 0
		osc.dVolume = 0

		// has it got params?
		if len(this.params[x]) == 0 {
			// turn it off
			osc.SetEnabled(false)
		} else {
			// turn it on and set defaults
			osc.SetEnabled(true)
			osc.GetEnvelope().SetEnvelope([]int{0, 0, 0, 0})
			osc.SetVolume(1.0)
			osc.SetWaveform(SINE)
			osc.SetFrequencyMultiplier(1.0)
			osc.SetPhase(0)
			osc.SetPulseWidth(math.Pi)

			// now process the params
			for key, value := range this.params[x] {
				if key == "WAVE" {
					osc.SetWaveform(this.AsWAVEFORM(value))
				} else if key == "AOPS" {
					z := this.AsFloatArray(value)
					n := len(z) / 2
					ops := make([]AdditiveOperator, n)
					for i := 0; i < n; i++ {
						ops[i].Amplitude = z[0]
						ops[i].Harmonic = int(z[1])
						z = z[2:]
					}
					osc.GetwfADDITIVE().Harmonics = ops
				} else if key == "FMOPS" {
					z := this.AsFloatArray(value)
					n := len(z) / 2
					ops := make([]FMOperator, n)
					for i := 0; i < n; i++ {
						ops[i].A = z[0]
						ops[i].F = z[1]
						z = z[2:]
					}
					osc.GetwfFMSYNTH().Data = ops
				} else if key == "ADSR" {
					osc.GetEnvelope().SetEnvelope(this.AsIntArray(value))
				} else if key == "VOLUME" {
					osc.SetVolume(this.AsDouble(value))
				} else if key == "FREQMULT" {
					osc.SetFrequencyMultiplier(this.AsDouble(value))
				} else if key == "PHASESHIFT" {
					osc.SetPhaseShift(this.AsDouble(value))
				} else if key == "PULSEWIDTH" {
					osc.SetPulseWidth(this.AsDouble(value))
				} else if key == "HIPASS" {
					v.AddFilter(NewHighPassFilter(float32(v.sampleRate), float32(this.AsDouble(value))))
				} else if key == "LOPASS" {
					v.AddFilter(NewLowPassFilter(float32(v.sampleRate), float32(this.AsDouble(value))))
				} else if key == "LFO" {
					// LFO=SINE,80,0.30,PULSEWIDTH
					parts := strings.Split(value, ",")
					if len(parts) == 5 {
						wf := strings.ToUpper(parts[0])
						hz, err := strconv.ParseFloat(parts[1], 64)
						if err != nil {
							hz = 80
						}
						fr, err := strconv.ParseFloat(parts[2], 64)
						if err != nil || fr < 0 {
							fr = 0.30
						}
						lfoBind := LFO_NONE
						switch strings.ToUpper(parts[3]) {
						case "PULSEWIDTH":
							lfoBind = LFO_PULSEWIDTH
						case "PHASESHIFT":
							lfoBind = LFO_PHASESHIFT
						case "VOLUME":
							lfoBind = LFO_VOLUME
						case "FREQUENCY":
							lfoBind = LFO_FREQUENCY
						}
						pw, err := strconv.ParseFloat(parts[4], 64)
						if err != nil {
							pw = math.Pi
						}

						// Now what to do
						if lfoBind == LFO_NONE {
							if v.LFO != nil {
								v.LFO.BoundVoice = nil
								v.LFO = nil
							}
						} else {
							if v.LFO != nil {
								v.LFO.SetWaveform(StringToWAVEFORM(wf))
								v.LFO.BoundControl = lfoBind
								v.LFO.BoundVoice = v
								v.LFO.MaxDeviation = fr
								v.LFO.SetFrequency(hz)
								v.LFO.PulseWidth = pw
								v.LFO.Trigger()
							} else {
								v.LFO = NewLFO(
									StringToWAVEFORM(wf),
									int(hz),
									v,
									lfoBind,
									fr,
								)
								v.LFO.PulseWidth = pw
								v.LFO.Trigger()
							}
						}
					}
				}
			}
		}
	}

}

func NewInstrument(params string) *Instrument {
	this := &Instrument{}
	for x := 0; x < numOSC; x++ {
		this.params[x] = make(StringParams)
	}
	this.Process(params)
	return this
}

func (this *Instrument) Process(params string) {
	// process string params

	fmt.Printf("Inst params: %s\n", params)

	oscparam := strings.Split(params, ";")

	for _, osc := range oscparam {
		osc = strings.Trim(osc, " ")

		parts := strings.Split(osc, ":")

		for _, item := range parts {
			item = strings.Trim(item, " ")
			//System.Out.Println( "LINE = "+item );
			nv := strings.Split(item, "=")
			this.params[this.oNum][strings.Trim(nv[0], " ")] = strings.Trim(nv[1], " ")
			//System.Out.Println( "Instrument param (OSC#"+oNum+"): "+nv[0]+" = "+nv[1] );
		}

		this.oNum++

	}
}

func (this *Instrument) AsWAVEFORM(name string) WAVEFORM {
	var wf WAVEFORM = SINE

	s := strings.ToUpper(name)

	if s == "BUZZER" {
		wf = BUZZER
		return wf
	}

	if s == "SAWTOOTH" {
		wf = SAWTOOTH
		return wf
	}

	if s == "SQUARE" {
		wf = SQUARE
		return wf
	}

	if s == "TRISAW" {
		wf = TRISAW
		return wf
	}

	if s == "TRIANGLE" {
		wf = TRIANGLE
		return wf
	}

	if s == "PULSE" {
		wf = PULSE
		return wf
	}

	if s == "NOISE" {
		wf = NOISE
		return wf
	}

	if s == "CUSTOM" {
		wf = CUSTOM
		return wf
	}

	if s == "ADD" {
		wf = ADDSYNTH
		return wf
	}

	if s == "FM" {
		wf = FMSYNTH
		return wf
	}

	return wf
}

func (this *Instrument) AsIntArray(p string) []int {
	v := strings.Split(p, ",")
	res := make([]int, len(v))
	idx := 0
	for _, is := range v {
		is = strings.Trim(is, " ")
		res[idx] = 0
		tmp, _ := strconv.ParseInt(is, 10, 32)
		res[idx] = int(tmp)
		idx++
	}
	return res
}

func (this *Instrument) AsFloatArray(p string) []float64 {
	v := strings.Split(p, ",")
	res := make([]float64, len(v))
	idx := 0
	for _, is := range v {
		is = strings.Trim(is, " ")
		res[idx] = 0
		tmp, _ := strconv.ParseFloat(is, 64)
		res[idx] = tmp
		idx++
	}
	return res
}

func (this *Instrument) AsDouble(p string) float64 {
	var v float64
	v, _ = strconv.ParseFloat(p, 64)
	return v
}
