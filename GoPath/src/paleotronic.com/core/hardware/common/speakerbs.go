package common

//import "paleotronic.com/fmt"

import (
	"sync"

	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
)

const RAWBUFFERS = 8

type SpeakerBitstream struct {
	samplebyte      int
	samplebit       int
	buffer          []uint64
	rawbuffer       [RAWBUFFERS][]float32
	rb              int
	buffersizebytes int
	lastbit         uint64
	lastamp         uint64
	bitcount        int
	onecount        int
	rawnzcount      int
	rawcount        int
	maxbits         int
	e               interfaces.Interpretable
	rate            int
	div             *int
	keepzeropackets bool
	sendraw         bool
	sendpacked      bool
	divfactor       float64
	flushMutex      sync.Mutex
	ramp            float32
	tps             int
}

func NewSpeakerBitstream(buffersize int, e interfaces.Interpretable, rate int, div *int) *SpeakerBitstream {
	this := &SpeakerBitstream{
		bitcount:        0,
		buffer:          make([]uint64, buffersize),
		buffersizebytes: buffersize,
		samplebit:       31,
		samplebyte:      0,
		maxbits:         32 * buffersize,
		e:               e,
		rate:            rate,
		div:             div,
		sendraw:         true,
		sendpacked:      true,
		keepzeropackets: false,
		divfactor:       1,
	}

	if this.sendraw {
		for i := 0; i < RAWBUFFERS; i++ {
			this.rawbuffer[i] = make([]float32, settings.RAWBufferSize)
		}
	}

	return this
}

func (sbs *SpeakerBitstream) Ramp() float32 {
	return sbs.ramp
}

func (sbs *SpeakerBitstream) SetRamp(v float32) {
	sbs.ramp = v
}

func (sbs *SpeakerBitstream) GetSendRaw() bool {
	return sbs.sendraw
}

func (sbs *SpeakerBitstream) SetDivFactor(f float64) {
	sbs.divfactor = f
}

func (sbs *SpeakerBitstream) WriteSample(amp int) {

	famp := float64(amp)

	famp = famp / (float64(*sbs.div) * sbs.divfactor)
	amp = int(famp * sbs.divfactor)

	if settings.MuteCPU {
		amp = 0
	}

	if sbs.sendraw {
		sbs.WriteSampleRaw(amp)
	}

	if sbs.sendpacked {
		sbs.WriteSamplePacked(amp)
	}

}

func (sbs *SpeakerBitstream) WriteSampleRaw(amp int) {

	if amp > 0 {
		sbs.rawnzcount++
	}

	sbs.rawbuffer[sbs.rb][sbs.rawcount] = float32(amp) * 0.25 * sbs.ramp
	sbs.rawcount++
	if sbs.rawcount >= settings.RAWBufferSize {
		sbs.FlushRaw()
	}

}

func (sbs *SpeakerBitstream) WriteSamplePacked(amp int) {

	var needed int
	if amp == 0 {
		needed = 1
	} else {
		needed = amp + 1 // number of 1s + a zero
		sbs.onecount++
	}

	bitsleft := sbs.maxbits - sbs.bitcount

	if bitsleft < needed {
		sbs.FlushPacked()
	}

	// generate bits
	for i := 0; i < needed; i++ {
		var bit uint64
		if amp > 0 && i < needed-1 {
			bit = 1
		}
		bitval := bit << uint64(sbs.samplebit)
		// At bit 31 we clear the byte
		if sbs.samplebit == 31 {
			sbs.buffer[sbs.samplebyte] = 0
		}
		// do bit
		if bit == 1 {
			sbs.buffer[sbs.samplebyte] |= bitval
		}
		sbs.samplebit--
		if sbs.samplebit < 0 {
			sbs.samplebyte++
			sbs.samplebit = 31
		}
		sbs.bitcount++ // keep track of how many bits this sample contains
	}

}

func (sbs *SpeakerBitstream) Flush() {
	sbs.FlushRaw()
	sbs.FlushPacked()
}

func (sbs *SpeakerBitstream) Drain() {
	sbs.DrainRaw()
	sbs.DrainPacked()
}

func (sbs *SpeakerBitstream) DrainRaw() {

	if !sbs.sendraw {
		return
	}

	sbs.rawnzcount = 0
	sbs.rawcount = 0

	sbs.rb = (sbs.rb + 1) % RAWBUFFERS
	for i := 0; i < sbs.buffersizebytes; i++ {
		sbs.rawbuffer[sbs.rb][i] = 0
	}
}

func (sbs *SpeakerBitstream) DrainPacked() {

	if !sbs.sendpacked {
		return
	}

	// after sent
	sbs.bitcount = 0
	sbs.onecount = 0
	sbs.samplebit = 31
	sbs.samplebyte = 0

}

func (sbs *SpeakerBitstream) FlushRaw() {

	if !sbs.sendraw {
		return
	}

	sbs.flushMutex.Lock()
	defer sbs.flushMutex.Unlock()

	if sbs.rawnzcount > 0 {
		sbs.e.PassWaveBufferNB(0, sbs.rawbuffer[sbs.rb][:sbs.rawcount], false, sbs.rate)
		//fmt.Printf("BUFFER (%d toggles) %v\n", sbs.tps, sbs.rawbuffer[sbs.rb][:sbs.rawcount])
	}

	sbs.rawnzcount = 0
	sbs.rawcount = 0
	sbs.tps = 0

	sbs.rb = (sbs.rb + 1) % RAWBUFFERS
	sbs.rawbuffer[sbs.rb] = make([]float32, settings.RAWBufferSize)
}

func (sbs *SpeakerBitstream) FlushPacked() {

	if !sbs.sendpacked {
		return
	}

	sbs.flushMutex.Lock()
	defer sbs.flushMutex.Unlock()

	if sbs.onecount > 0 || sbs.keepzeropackets {

		bytes := (sbs.bitcount / 32) + 1
		if bytes > sbs.buffersizebytes {
			bytes = sbs.buffersizebytes
		}

		ndata := make([]uint64, 1+bytes)
		ndata[0] = uint64(sbs.bitcount)

		for i := 0; i < bytes; i++ {
			ndata[1+i] = sbs.buffer[i]
		}

		sbs.e.PassWaveBufferCompressed(0, ndata, false, sbs.rate, sbs.sendraw)

	}

	// after sent
	sbs.bitcount = 0
	sbs.onecount = 0
	sbs.samplebit = 31
	sbs.samplebyte = 0

}

func (sbs *SpeakerBitstream) DecodeBits(bitcount int, buffer []uint64, ampscale float32) []float32 {

	var fdata []float32

	fdata = make([]float32, 1)
	var bitnum int = 31
	var bindex int = 0
	var findex int = 0
	var bitsprocessed int = 0

	for bitsprocessed < bitcount {
		bitval := uint64(1 << uint64(bitnum))
		if buffer[bindex]&bitval != 0 {
			// 1: bump amp
			fdata[findex] += ampscale
		} else {
			// 0: move to next
			fdata = append(fdata, 0)
			findex++
		}
		bitsprocessed++
		bitnum--
		if bitnum < 0 {
			// next 'byte'
			bindex++
			bitnum = 31
		}
	}

	return fdata[0:findex]
}

func (sbs *SpeakerBitstream) Empty() bool {
	return sbs.bitcount == 0
}

//func init() {

//	indata := []int{0, 0, 0, 5, 6, 1, 0, 0, 1, 2, 3, 3, 1, 0, 0, 0, 7}
//	fmt.Println(indata)
//	sbs := NewSpeakerBitstream(2)
//	for _, f := range indata {
//		sbs.WriteSample(f)
//	}
//	if !sbs.Empty() {
//		sbs.Flush()
//	}
//	os.Exit(1)

//}
