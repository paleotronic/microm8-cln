package ducktape

import (
	"bytes"
	"paleotronic.com/fmt"
	"testing"
)

func TestMessagePackUnpack(t *testing.T) {

	var kpei, kpeo Message

	kpei = Message{
		Target:  "Fred",
		Channel: "#display",
		Payload: []byte{7, 6, 5, 4, 3, 2, 1, 0},
	}

	b, err := kpei.MarshalBinary()
	if err != nil {
		t.Error(err.Error())
	}

	////fmt.Printf("%v\n", b)

	err = kpeo.UnmarshalBinary(b)
	if err != nil {
		t.Error(err.Error())
	}

	if kpei.Target != kpeo.Target {
		t.Error("Target mismatch after decode", kpei, "/", kpeo)
	}

	if kpei.Channel != kpeo.Channel {
		t.Error("Channel mismatch after decode", kpei, "/", kpeo)
	}

	if !bytes.Equal(kpei.Payload, kpeo.Payload) {
		t.Error("Payload mismatch after decode", kpei, "/", kpeo)
	}

}
