package restalgia

import (
	"math"
)

const (
	samplesForFilt = 10
)

type Filter interface {
	Filter(in float32) float32
}

type LowPassFilter struct {
	Memory     []float32
	CutOff     float32
	SampleRate float32
}

type HighPassFilter struct {
	LowPassFilter
}

// return the filtered value for in
func (this *LowPassFilter) Filter(in float32) float32 {

	this.Memory = append(this.Memory, in)
	if len(this.Memory) > samplesForFilt {
		o := len(this.Memory) - samplesForFilt
		this.Memory = this.Memory[o:]
	}

	if len(this.Memory) < samplesForFilt {
		return in
	}

	// can filter here
	RC := 1.0 / (this.CutOff * 2 * math.Pi)
	dt := 1.0 / this.SampleRate
	alpha := dt / (RC + dt)
	filteredArray := make([]float32, len(this.Memory))
	filteredArray[0] = this.Memory[0]
	for i := 1; i < len(this.Memory); i++ {
		filteredArray[i] = filteredArray[i-1] + (alpha * (this.Memory[i] - filteredArray[i-1]))
	}

	f := filteredArray[len(filteredArray)-1] // last value
	if f > 1 {
		f = 1
	} else if f < -1 {
		f = -1
	}

	return f

}

func NewLowPassFilter(sampleRate float32, cutoffFreq float32) *LowPassFilter {
	return &LowPassFilter{CutOff: cutoffFreq, SampleRate: sampleRate, Memory: make([]float32, 0)}
}

func NewHighPassFilter(sampleRate float32, cutoffFreq float32) *HighPassFilter {
	this := &HighPassFilter{}
	this.LowPassFilter = *NewLowPassFilter(sampleRate, cutoffFreq)
	return this
}

// return the filtered value for in
func (this *HighPassFilter) Filter(in float32) float32 {

	this.Memory = append(this.Memory, in)
	if len(this.Memory) > samplesForFilt {
		o := len(this.Memory) - samplesForFilt
		this.Memory = this.Memory[o:]
	}

	if len(this.Memory) < samplesForFilt {
		return in
	}

	// can filter here
	RC := 1.0 / (this.CutOff * 2 * math.Pi)
	dt := 1.0 / this.SampleRate
	alpha := RC / (RC + dt)
	filteredArray := make([]float32, len(this.Memory))
	filteredArray[0] = this.Memory[0]
	for i := 1; i < len(this.Memory); i++ {
		//filteredArray[i] = filteredArray[i-1] + (alpha * (this.Memory[i] - filteredArray[i-1]))

		// y[i] := α * y[i-1] + α * (x[i] - x[i-1])

		filteredArray[i] = alpha*filteredArray[i-1] + alpha*(this.Memory[i]-this.Memory[i-1])

		//filteredArray[i] = alpha * (filteredArray[i-1] + this.Memory[i] - this.Memory[i-1])

	}

	f := filteredArray[len(filteredArray)-1] // last value
	if f > 1 {
		f = 1
	} else if f < -1 {
		f = -1
	}

	return f

}
