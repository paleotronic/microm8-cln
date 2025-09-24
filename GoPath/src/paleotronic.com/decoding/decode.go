package decoding

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"time"

	"paleotronic.com/log"

	"paleotronic.com/decoding/ogg"
	"paleotronic.com/decoding/wav"
)

// AudioDecoder is a generalized audio decoder interface
type AudioDecoder interface {
	Samples() ([]float32, error)
	SamplesMono() ([]float32, error)
	Length() time.Duration
	SampleRate() int
	Channels() int
}

func NewAudio(r io.Reader) (AudioDecoder, error) {

	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(raw) < 4 {
		return nil, errors.New("File is corrupt")
	}

	if string(raw[0:4]) == "RIFF" {
		return wav.New(bytes.NewBuffer(raw))
	} else if string(raw[0:4]) == "OggS" {
		return ogg.New(bytes.NewBuffer(raw))
	}

	return nil, errors.New("Unsupported audio type: unrecognized magic " + string(raw[0:4]))

}

type AudioPlayer struct {
	buffer      []float32
	rate        int
	channels    int
	size        int
	leadin      int
	fadein      int
	running     bool
	paused      bool
	buffptr     int
	samplecount int
	p           Playable
	stime       time.Time
}

type Playable interface {
	PassMusicBufferNB(data []float32, loop bool, rate int, channels int)
	GetMusicVolume() (float32, bool)
	SetMusicVolume(v float32, locked bool)
}

func NewPlayer(output Playable, r io.Reader, leadinMS int, fadeinMS int) (*AudioPlayer, error) {
	ad, err := NewAudio(r)
	if err != nil {
		return nil, err
	}
	s, _ := ad.Samples()
	ap := &AudioPlayer{
		buffer:   s,
		rate:     ad.SampleRate(),
		channels: ad.Channels(),
		size:     len(s),
		p:        output,
		leadin:   leadinMS,
		fadein:   fadeinMS,
	}
	return ap, nil
}

func (p *AudioPlayer) Stop() {
	if p.running {
		p.running = false
		time.Sleep(50 * time.Millisecond)
	}
}

func (p *AudioPlayer) SetPaused(b bool) {
	if p.running {
		p.paused = b
		time.Sleep(50 * time.Millisecond)
	}
}

func (p *AudioPlayer) Start() {
	p.Stop()
	go p.play()
}

func (p *AudioPlayer) getVolume() float32 {

	volume, _ := p.p.GetMusicVolume()

	samples := p.fadein * p.rate / 1000

	if p.samplecount <= p.leadin {
		return 0
	}

	diff := p.samplecount - p.leadin
	if diff >= 0 && diff <= samples {
		m := float32(diff) / float32(samples)
		return volume * m
	}

	return volume
}

const maxSamples = 4096

func (p *AudioPlayer) play() {

	p.running = true
	p.samplecount = 0

	time.Sleep(time.Duration(p.leadin) * time.Millisecond)
	p.stime = time.Now()

	var chunk [maxSamples]float32

	for p.running {

		for i, _ := range chunk {

			if p.buffptr >= p.size {
				p.buffptr = 0
				p.fadein = 0
				p.samplecount = 0
			}
			chunk[i] = p.getVolume() * p.buffer[p.buffptr]
			p.buffptr++
			p.samplecount++

		}

		if p.p != nil {
			p.p.PassMusicBufferNB(chunk[:], false, p.rate, p.channels)
			log.Printf("Sent %d bytes of music...\n", len(chunk))
		}

		for p.paused {
			time.Sleep(50 * time.Millisecond)
		}
	}

}
