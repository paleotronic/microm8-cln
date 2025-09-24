package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"

	"paleotronic.com/microtracker/mock"
	"paleotronic.com/microtracker/tracker"
)

var filename = flag.String("i", "", "input sng file to assemble")
var outfile = flag.String("o", "song.asm", "output asm file")

func main() {

	flag.Parse()

	if *filename == "" {
		log.Fatalf("You need a filename")
	}

	e := &MockInt{}
	song := tracker.NewSong(120, mock.New(e, 0xc400))

	data, err := ioutil.ReadFile(*filename)
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	err = song.LoadBuffer(bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Error loading song: %v", err)
	}

	encoder := tracker.NewASMEncoder()
	encoder.Encode(song, *outfile)
}
