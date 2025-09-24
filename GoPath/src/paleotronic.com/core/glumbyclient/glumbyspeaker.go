package main

/*
Simple PC Speaker style implementation
*/

import (
	"paleotronic.com/fmt"
	"time"

	"paleotronic.com/core/vduproto"
	"paleotronic.com/restalgia"
	"paleotronic.com/restalgia/driver"
)

const (
	AUDIORATE = 44100
	CHANNELS  = 1
	BUFFSIZE  = 128
)

var Audio driver.Output
var AudioError error
var SPEAKER *restalgia.Voice
var CHUNKCHAN chan []float32
var CLICKNANO int64
var count int
var value float32 = -1

func InitAudio() {
	Audio, AudioError = driver.Get(AUDIORATE, CHANNELS)
	if AudioError != nil {
		panic(AudioError)
	}
	// Got here things are okay to init Restalgia
	inst := restalgia.NewInstrument("WAVE=BUZZER:VOLUME=1.0:ADSR=0,0,4000,0")
	SPEAKER = restalgia.NewVoice(AUDIORATE, restalgia.BUZZER, 1.0)
	inst.Apply(SPEAKER)
	SPEAKER.SetVolume(1.0)
	SPEAKER.SetFrequency(3000)
	CHUNKCHAN = make(chan []float32)

	// commence streaming
	go AudioLoop()
	//go AudioGenerator()
	go ClackyTheChicken()

	CLICKNANO = time.Now().UnixNano()
}

func AudioGenerator() {

	buffer := make([]float32, BUFFSIZE)
	count := 0

	for {

		buffer[count] = SPEAKER.GetAmplitude()

		count = count + 1
		if count >= len(buffer) {
			CHUNKCHAN <- buffer
			count = 0
			buffer = make([]float32, BUFFSIZE)
		}
	}

}

func AudioLoop() {
	Audio.Start()
	for {

		select {
		case chunk := <-CHUNKCHAN:
			Audio.Push(chunk)
		}

	}
	Audio.Stop()
}

func ClackyTheChicken() {
	CLICKNANO = -1
	for {
		select {
		case sout := <-spkChan:
			z := sout.Data.(vduproto.ClickEvent)
			ClickSpeaker(z.Data)
		}
	}
}

// Click speaker *at least* <nanos> ns after last click
func ClickSpeaker(data []byte) {

	count++
	if count%1000 == 0 {
		////fmt.Println(count)
	}

	//SPEAKER.GetOSC(0).GetWfBUZZER().Stimulate()
	buffer := make([]float32, len(data))

	for i := 0; i < len(data); i++ {
		if data[i] != 0 {
			value = -value
		}
		buffer[i] = value
	}

	// CHUNKCHAN <- buffer
	Audio.Push(buffer)

}
