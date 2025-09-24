package restalgia

import (
	"paleotronic.com/fmt"
	"testing"
)

func TestKeypressPackUnpack(t *testing.T) {

	song := NewSong("test.song")

	////fmt.Printntln(song)

	s := 0
	for song.Playing {
		if song.PullSampleMono() != 0 {
			//////fmt.Println(v)
			s++
		}
	}

	////fmt.Printntf("Samples non zero = %d\n", s)

}
