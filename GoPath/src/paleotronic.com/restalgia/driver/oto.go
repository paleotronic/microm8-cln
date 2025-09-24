// +build darwin linux windows

package driver

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/ebitengine/oto/v3"
	"paleotronic.com/core/settings"
	"paleotronic.com/fmt"
	"paleotronic.com/octalyzer/errorreport"
	"paleotronic.com/restalgia"
)

type otoDriver struct {
	ctx        *oto.Context
	player     *oto.Player
	mix        *restalgia.Mixer
	sf         func()
	reader     *otoReader
	sampleRate int
}

// otoReader implements io.Reader to provide audio data to Oto
type otoReader struct {
	driver     *otoDriver
	tempBuffer []float32
}

func (r *otoReader) Read(p []byte) (n int, err error) {
	// Calculate how many float32 samples we need
	samplesNeeded := len(p) / 4 // 4 bytes per float32

	// Reuse buffer if possible
	if len(r.tempBuffer) < samplesNeeded {
		r.tempBuffer = make([]float32, samplesNeeded)
	}
	samples := r.tempBuffer[:samplesNeeded]

	// Get float32 samples from mixer
	if r.driver.mix != nil {
		r.driver.mix.FillStereo(samples)
	} else {
		// Fill with silence if no mixer
		for i := range samples {
			samples[i] = 0
		}
	}

	// Convert float32 samples directly to bytes without intermediate buffer
	for i, sample := range samples {
		// Clamp the sample to [-1, 1] range
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}

		// Convert to uint32 bits and write directly
		bits := math.Float32bits(sample)
		binary.LittleEndian.PutUint32(p[i*4:], bits)
	}

	return len(p), nil
}

func init() {
	// Oto initialization is done per-context, not globally
}

func get(sampleRate, channels int) (Output, error) {
	o := &otoDriver{
		sampleRate: sampleRate,
	}

	// Create reader for this driver
	o.reader = &otoReader{driver: o}

	// Create Oto context options with buffer size
	// Use a buffer of ~50ms for balance between latency and stability
	bufferSize := time.Duration(17) * time.Millisecond

	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: channels,
		Format:       oto.FormatFloat32LE,
		BufferSize:   bufferSize,
	}

	// Create context
	ctx, readyChan, err := oto.NewContext(op)
	if err != nil {
		errorreport.GracefulErrorReport(
			"The Oto audio subsystem has failed to initialize properly.",
			err,
		)
		return nil, err
	}

	// Wait for context to be ready
	<-readyChan
	o.ctx = ctx

	// Create player
	o.player = ctx.NewPlayer(o.reader)

	// Set a larger player buffer size for smoother playback
	// Note: *oto.Player implements BufferSizeSetter directly
	// Set to ~100ms worth of audio data
	playerBufferSize := 4096  //sampleRate * channels * 4 * 100 / 1000 // 4 bytes per float32
	o.player.SetBufferSize(playerBufferSize)

	// Update global sample rate
	settings.SampleRate = sampleRate
	fmt.Printf("Oto is using a samplerate of %d with buffer size %v.\n", settings.SampleRate, bufferSize)

	return o, nil
}

func (o *otoDriver) Push(samples []float32) {
	// No-op: This implementation uses mixer-based pull instead of push
}

func (o *otoDriver) SetStarvationFunc(sf func()) {
	o.sf = sf
}

func (o *otoDriver) SetPullSource(mix *restalgia.Mixer) {
	o.mix = mix
}

func (o *otoDriver) Stop() {
	if o.player != nil {
		o.player.Pause()
	}
}

func (o *otoDriver) Start() {
	if o.player != nil {
		o.player.Play()
	}
}
