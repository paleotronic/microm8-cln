package restalgia

type Waveformer interface {
	ValueForInputSignal(f float64) float64
}
