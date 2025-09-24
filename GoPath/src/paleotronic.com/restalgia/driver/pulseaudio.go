// +build jlinux

package driver

import (
	"bytes"
	"encoding/binary"
	"time"

	"paleotronic.com/log"

	"github.com/mesilliac/pulse-simple"
	"paleotronic.com/restalgia"
)

type output struct {
	st      *pulse.Stream
	sf      func()
	mix     *restalgia.Mixer
	stopped bool
}

var SampleRate int

func (o *output) SetPullSource(mix *restalgia.Mixer) {
	o.mix = mix
}

func (o *output) SetStarvationFunc(sf func()) {
	o.sf = sf
}

func get(sampleRate, channels int) (Output, error) {
	o := new(output)
	var err error
	ss := pulse.SampleSpec{pulse.SAMPLE_FLOAT32LE, uint32(sampleRate), uint8(channels)}
	o.st, err = pulse.Playback("mog", "mog", &ss)
	if err != nil {
		return nil, err
	}
	SampleRate = sampleRate
	return o, nil
}

func (o *output) Push(samples []float32) {
	buf := new(bytes.Buffer)
	for _, s := range samples {
		_ = binary.Write(buf, binary.LittleEndian, s)
	}
	_, err := o.st.Write(buf.Bytes())
	if err != nil {
		log.Println(err)
	}
}

func (o *output) Start() {
	o.stopped = false
	go func() {
		for !o.stopped {
			if o.mix != nil {
				chunk := make([]float32, 1024)
				o.mix.FillStereo(chunk)
				o.Push(chunk)
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()
}

func (o *output) Stop() {
	o.stopped = true
}
