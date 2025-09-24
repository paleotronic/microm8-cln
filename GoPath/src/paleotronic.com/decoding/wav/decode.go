package wav

import (
	"io"
	"time"

	"github.com/mjibson/go-dsp/wav"
)

type WaveBlob struct {
	w *wav.Wav
}

func New(r io.Reader) (*WaveBlob, error) {

	w, err := wav.New(r)
	if err != nil {
		return nil, err
	}

	wb := &WaveBlob{
		w: w,
	}

	return wb, nil

}

func (wb *WaveBlob) Channels() int {
	return int(wb.w.NumChannels)
}

func (wb *WaveBlob) SampleRate() int {
	return int(wb.w.SampleRate)
}

func (wb *WaveBlob) Length() time.Duration {
	return wb.w.Duration
}

func (wb *WaveBlob) Samples() ([]float32, error) {
	return wb.w.ReadFloats(wb.w.Samples)
}

func (wb *WaveBlob) SamplesMono() ([]float32, error) {
	return wb.Samples()
}
