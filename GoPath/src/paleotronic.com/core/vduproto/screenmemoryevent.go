package vduproto

import (
	"errors"
	"paleotronic.com/core/types"
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

type ScreenMemoryEvent struct {
	LayerID byte
	Offset  int
	Content []uint
	X       int
	Y       int
}

func (this ScreenMemoryEvent) MarshalBinary() ([]byte, error) {

	data := []byte{byte(types.MtScreenMemoryEvent)}

	data = append(data, this.LayerID)
	data = append(data, byte(this.Offset%256), byte(this.Offset/256))
	data = append(data, byte(this.X), byte(this.Y))

	tmp := make([]byte, 0)

	for _, v := range this.Content {
		// pack rune lo, rune hi, col byte
		r := rune(v & 0xffff)
		c := (v & 0xff0000) >> 16
		b := (v & 0xf000000) >> 24
		tmp = append(tmp, byte(r%256), byte(r/256), byte(c), byte(b))
	}

	var b bytes.Buffer
	w := gzip.NewWriter( &b )
	w.Write(tmp)
	w.Close()

	data = append(data, b.Bytes()...)

	return data, nil

}

func (this *ScreenMemoryEvent) UnmarshalBinary(data []byte) error {

	if data[0] != byte(types.MtScreenMemoryEvent) {
		return errors.New("wrong type")
	}

	if len(data) < 6 {
		return errors.New("not enough data")
	}

	this.LayerID = data[1]
	this.Offset = int(data[2]) + 256*int(data[3])
	this.X = int(data[4])
	this.Y = int(data[5])

	r, ee := gzip.NewReader( bytes.NewBuffer(data[6:]) )
	if ee != nil {
		return ee
	}
	defer r.Close()
	rem, ee := ioutil.ReadAll(r)
	if ee != nil {
		return ee
	}

	if len(rem)%4 != 0 {
		return errors.New("incorrect length")
	}

	this.Content = make([]uint, 0)

	p := 0
	for p < len(rem) {
		v := uint(rem[p]) + 256*uint(rem[p+1]) + 65536*uint(rem[p+2]) + (uint(rem[p+3]) << 24)
		this.Content = append(this.Content, v)
		// move
		p += 4
	}

	return nil

}
