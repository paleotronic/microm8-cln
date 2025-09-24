package main

import (
	//"paleotronic.com/fmt"
	"paleotronic.com/restalgia"
	"paleotronic.com/restalgia/driver"
)

func main() {
	output, err := driver.Get(44100, 1, nil)
	if err != nil {
		panic(err)
	}
	////fmt.Printntln("Got audio stream to use")
	inst := restalgia.NewInstrument("WAVE=BUZZER:VOLUME=0.01:ADSR=50,0,4000,200")
	voice := restalgia.NewVoice(44100, restalgia.SINE, 1.0)
	inst.Apply(voice)
	voice.SetVolume(1.0)
	voice.SetFrequency(3000)
	output.Start()
	n := voice.GetSamplesForDuration(5.0)
	buffer := make([]float32, n)
	for x := 0; x < len(buffer); x++ {
		buffer[x] = voice.GetAmplitude()
		if x%8 == 0 {
			voice.OSC[0].GetWfBUZZER().Stimulate()
		}
		//////fmt.Print(buffer[x])
	}
	output.Push(buffer)

	output.Stop()
}
