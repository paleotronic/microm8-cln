// +build

package driver

import (
	"paleotronic.com/fmt"

	"github.com/gordonklaus/portaudio"
	"paleotronic.com/core/settings"
	"paleotronic.com/octalyzer/errorreport"
	"paleotronic.com/restalgia"
)

type port struct {
	st  *portaudio.Stream
	mix *restalgia.Mixer
	sf  func()
}

func init() {
	err := portaudio.Initialize()

	if err != nil {

		errorreport.GracefulErrorReport(
			"The PortAudio subsystem has failed to initialize properly.",
			err,
		)

	}
}

func get(sampleRate, channels int) (Output, error) {
	o := &port{}
	var err error
	h, e := portaudio.DefaultHostApi()
	if e != nil {
		errorreport.GracefulErrorReport("PortAudio failed to obtain the default host API.", e)
		return nil, err
	}
	//o.st, err = portaudio.OpenStream(portaudio.HighLatencyParameters(nil, h.DefaultOutputDevice), o.Fetch)

	o.st, err = portaudio.OpenStream(portaudio.LowLatencyParameters(nil, h.DefaultOutputDevice), o.Fetch)
	if err != nil {
		errorreport.GracefulErrorReport("PortAudio failed to open the default audio device ("+h.DefaultOutputDevice.Name+").\nIf there is another device, please set it as your OS default and try again.", e)
		return nil, err
	}

	// o.st, err = portaudio.OpenDefaultStream(0, 2, 44100, portaudio.FramesPerBufferUnspecified)
	// if err != nil {
	// 	errorreport.GracefulErrorReport("PortAudio failed to open the default stream.", err)
	// 	return nil, err
	// }

	settings.SampleRate = int(o.st.Info().SampleRate)
	fmt.Printf("Portaudio is reporting a samplerate of %d.\n", settings.SampleRate)

	return o, nil
}

func (p *port) Push(samples []float32) {
	// No-op: This implementation uses mixer-based pull instead of push
}

// Fetch fills the output buffer with audio samples from the mixer.
func (p *port) Fetch(out []float32) {
	if p.mix != nil {
		p.mix.FillStereo(out)
	}
}

func (o *port) SetStarvationFunc(sf func()) {
	o.sf = sf
}

func (o *port) SetPullSource(mix *restalgia.Mixer) {
	o.mix = mix
}

func (p *port) Stop() {
	p.st.Stop()
}

func (p *port) Start() {
	p.st.Start()
}
