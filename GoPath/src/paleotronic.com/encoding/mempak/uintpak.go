package mempak

import "paleotronic.com/fmt"
import "errors"
import "paleotronic.com/utils"

type StreamPack struct {
	Data  []byte
	Count byte
}

func (s *StreamPack) Add(v byte) {
	if len(s.Data) == 0 {
		s.Data = append(s.Data, v)
	} else {
		if s.Data[len(s.Data)-1] == v {
			if s.Count == 0 {
				s.Data = append(s.Data, v)
			}
			s.Count++
			if s.Count == 255 {
				s.Data = append(s.Data, s.Count)
				s.Count = 0
			}
		} else {
			if s.Count > 0 {
				s.Data = append(s.Data, s.Count)
				s.Count = 0
			}
			s.Data = append(s.Data, v)
		}
	}
}

// Add a slice of bytes to the packed structure
func (s *StreamPack) AddSlice(v []byte) {
	for _, bb := range v {
		s.Add(bb)
	}
	if s.Count > 0 {
		s.Data = append(s.Data, s.Count)
		s.Count = 0
	}
}

// Unwind the original stream from the packed one.
func (s *StreamPack) Unwind() ([]byte, error) {

	out := make([]byte, 0)

	i := 0
	var lastIsRepeat bool = false
	for i < len(s.Data) {
		v := s.Data[i]
		if len(out) == 0 {
			out = append(out, v)
			lastIsRepeat = false
		} else if (s.Data[i-1] == v) && (!lastIsRepeat) {
			i++

			r := s.Data[i]
			for c := 0; c < int(r); c++ {
				out = append(out, v)
			}
			lastIsRepeat = true

		} else {
			out = append(out, v)
			lastIsRepeat = false
		}

		i++
	}

	return out, nil
}

// PackSliceUints takes a uint64 slice and converts it to a stream of runlength encoded bytes
func PackSliceUints(data []uint64) []byte {
	//var encoded StreamPack

	count := len(data)

	var in []byte = make([]byte, count*8)
	for i, v := range data {
		in[i+count*0] = byte(v & 0xff)
		in[i+count*1] = byte((v >> 8) & 0xff)
		in[i+count*2] = byte((v >> 16) & 0xff)
		in[i+count*3] = byte((v >> 24) & 0xff)
		in[i+count*4] = byte((v >> 32) & 0xff)
		in[i+count*5] = byte((v >> 40) & 0xff)
		in[i+count*6] = byte((v >> 48) & 0xff)
		in[i+count*7] = byte((v >> 56) & 0xff)
	}

	// pre-load lengths of each stream at start
	// 2 bytes per length = 4 x 3 = 12 bytes
	//encoded.AddSlice(in)

	encoded := utils.GZIPBytes(in)

	final := make([]byte, 0)
	ll := len(data)
	final = append(final, byte(ll&0xff))
	final = append(final, byte((ll>>8)&0xff))

	final = append(final, encoded...)

	//	fmt.Printf("-> Encoded %d uints to %d bytes\n", len(data), len(final) )

	return final
}

func UnpackSliceUints(data []byte) ([]uint64, error) {

	var out []uint64
	//var s StreamPack

	if len(data) < 2 {
		return out, errors.New("Not enough data")
	}

	count := int(data[0]) | (int(data[1]) << 8)

	encoded := data[2:]

	out = make([]uint64, count)
	u := utils.UnGZIPBytes(encoded)

	if len(u) != count*8 {
		return out, errors.New(fmt.Sprintf("Expected byte count does not match - expected %d, got %d", count*4, len(u)))
	}

	for i, v := range u {
		z := i % count
		switch i / count {
		case 0:
			out[z] = out[z] | uint64(v)
		case 1:
			out[z] = out[z] | (uint64(v) << 8)
		case 2:
			out[z] = out[z] | (uint64(v) << 16)
		case 3:
			out[z] = out[z] | (uint64(v) << 24)
		case 4:
			out[z] = out[z] | (uint64(v) << 32)
		case 5:
			out[z] = out[z] | (uint64(v) << 40)
		case 6:
			out[z] = out[z] | (uint64(v) << 48)
		case 7:
			out[z] = out[z] | (uint64(v) << 56)
		}
	}

	return out, nil
}
