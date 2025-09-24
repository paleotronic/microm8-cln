package spectrum

import "paleotronic.com/core/hardware/spectrum/snapshot"

// const (
// 	SCREEN_BASE_ADDR = 0x4000
// 	ATTR_BASE_ADDR   = 0x5800

// 	PIXELS_PER_TSTATE = 2 // The number of screen pixels painted per T-state

// 	FIRST_SCREEN_BYTE = 14336 // T-state when the first byte of the screen (16384) is displayed

// 	// Horizontal
// 	LINE_SCREEN       = ScreenWidth / PIXELS_PER_TSTATE // 128 T states of screen
// 	LINE_RIGHT_BORDER = 24                              // 24 T states of right border
// 	LINE_RETRACE      = 48                              // 48 T states of horizontal retrace
// 	LINE_LEFT_BORDER  = 24                              // 24 T states of left border

// 	TSTATES_PER_LINE = (LINE_RIGHT_BORDER + LINE_SCREEN + LINE_LEFT_BORDER + LINE_RETRACE) // 224 T states

// 	TStatesPerFrame = 69888
// 	InterruptLength = 32

// 	ScreenWidth  = 256
// 	ScreenHeight = 192
// )
var zzz snapshot.Z80

type SpectrumConfig struct {
	ScreenBaseAddr0 int // First Screen
	AttrBaseAddr0   int
	ScreenBaseAddr1 int // Second Screen (optional)
	AttrBaseAddr1   int
	PixelsPerTState int // # Pixels displayed by ULA per TState
	FirstScreenByte int // TState for first point of screen (0,0)
	TStatesPerLine  int // TStates per scanline
	TStatesPerFrame int // TStates per frame
	InterruptLength int // TStates per IRQ
	ScreenWidth     int
	ScreenHeight    int
	VSyncTiming     int
}

func NewSpectrum48KConfig() *SpectrumConfig {
	return &SpectrumConfig{
		ScreenBaseAddr0: 0x4000,
		AttrBaseAddr0:   0x5800,
		PixelsPerTState: 2,
		FirstScreenByte: 14336,
		TStatesPerLine:  224,
		TStatesPerFrame: 69888,
		InterruptLength: 32,
		ScreenWidth:     256,
		ScreenHeight:    192,
		VSyncTiming:     37264,
	}
}

func NewSpectrum128KConfig() *SpectrumConfig {
	return &SpectrumConfig{
		ScreenBaseAddr0: 0x4000,
		AttrBaseAddr0:   0x5800,
		PixelsPerTState: 2,
		FirstScreenByte: 14362,
		TStatesPerLine:  228,
		TStatesPerFrame: 70908,
		InterruptLength: 32,
		ScreenWidth:     256,
		ScreenHeight:    192,
		VSyncTiming:     37264,
	}
}
