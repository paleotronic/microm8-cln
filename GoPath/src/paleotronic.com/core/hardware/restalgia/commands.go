package restalgia

const (
	RF_initFunc           = 0x00
	RF_getVolume          = 0x01
	RF_setVolume          = 0x81
	RF_getFrequency       = 0x02
	RF_setFrequency       = 0x82
	RF_getEnvelopeAttack  = 0x03
	RF_setEnvelopeAttack  = 0x83
	RF_getEnvelopeDecay   = 0x04
	RF_setEnvelopeDecay   = 0x84
	RF_getEnvelopeSustain = 0x05
	RF_setEnvelopeSustain = 0x85
	RF_getEnvelopeRelease = 0x06
	RF_setEnvelopeRelease = 0x86
	RF_getWaveform        = 0x07
	RF_setWaveform        = 0x87
	RF_getLFOControl      = 0x08
	RF_setLFOControl      = 0x88
	RF_getLFOFreq         = 0x09
	RF_setLFOFreq         = 0x89
	RF_getLFORatio        = 0x0a
	RF_setLFORatio        = 0x8a
	RF_getLFOWaveform     = 0x0b
	RF_setLFOWaveform     = 0x8b
	RF_getEnabled         = 0x0c
	RF_setEnabled         = 0x8c
	RF_getEnvShape        = 0x0d
	RF_setEnvShape        = 0x8d
	RF_getEnvShapeFreq    = 0x0e
	RF_setEnvShapeFreq    = 0x8e
	RF_getEnvShapeEnabled = 0x0f
	RF_setEnvShapeEnabled = 0x8f
	RF_getColour          = 0x10
	RF_setColour          = 0x90
	RF_getColourRatio     = 0x11
	RF_setColourRatio     = 0x91
	RF_getIsColour        = 0x12
	RF_setIsColour        = 0x92
	RF_getPan             = 0x13
	RF_setPan             = 0x93
)
