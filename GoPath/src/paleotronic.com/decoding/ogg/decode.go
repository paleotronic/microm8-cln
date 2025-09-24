package ogg

import (
	"io"
	"time"

	"paleotronic.com/fmt"

	"bytes"
	"io/ioutil"

	"paleotronic.com/decoding/ogg/internal/vorbis"
)

type VorbisBlob struct {
	v *vorbis.Vorbis
	d time.Duration
}

func New(r io.Reader) (*VorbisBlob, error) {

	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	v, err := vorbis.New(bytes.NewBuffer(raw))
	if err != nil {
		return nil, err
	}

	d, err := vorbis.Length(raw)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Length is %v\n", d)

	vb := &VorbisBlob{
		v: v,
		d: d,
	}

	return vb, nil

}

func (vb *VorbisBlob) Channels() int {
	return vb.v.Channels
}

func (vb *VorbisBlob) SampleRate() int {
	return vb.v.SampleRate
}

func (vb *VorbisBlob) Length() time.Duration {
	return vb.d
}

func (vb *VorbisBlob) Samples() ([]float32, error) {

	var r []float32 = make([]float32, 0, 16384)

	var err error
	var chunk []float32

	for err == nil {
		chunk, err = vb.v.Decode()
		if err == nil {
			r = append(r, chunk...)
		}
	}

	if err.Error() == "unexpected EOF" {
		err = nil
	}

	return r, err
}

func (vb *VorbisBlob) SamplesMono() ([]float32, error) {

	var r []float32 = make([]float32, 0, 16384)

	var err error
	var chunk []float32

	for err == nil {
		chunk, err = vb.v.Decode()
		if err == nil {
			for i, v := range chunk {
				if i%vb.v.Channels == 0 {
					r = append(r, v)
				}
			}
		}
	}

	if err.Error() == "unexpected EOF" {
		err = nil
	}

	return r, err
}
