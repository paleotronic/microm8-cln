package video

import (
	"paleotronic.com/octalyzer/video/font"
)

const ttxtRows = 25
const ttxtCols = 40

const (
	// text
	ttxtColorRed     = 0x81
	ttxtColorGreen   = 0x82
	ttxtColorYellow  = 0x83
	ttxtColorBlue    = 0x84
	ttxtColorMagenta = 0x85
	ttxtColorCyan    = 0x86
	ttxtColorWhite   = 0x87
	ttxtFlash        = 0x88
	ttxtSteady       = 0x89
	ttxtNormalHeight = 0x8c
	ttxtDoubleHeight = 0x8d
	// graphics
	ttxtGraphicsRed        = 0x91
	ttxtGraphicsGreen      = 0x92
	ttxtGraphicsYellow     = 0x93
	ttxtGraphicsBlue       = 0x94
	ttxtGraphicsMagenta    = 0x95
	ttxtGraphicsCyan       = 0x96
	ttxtGraphicsWhite      = 0x97
	ttxtConcealDisplay     = 0x98
	ttxtContiguousGraphics = 0x99
	ttxtSeparatedGraphics  = 0x9a
	ttxtBlackBackground    = 0x9c
	ttxtNewBackground      = 0x9d
	ttxtHoldGraphics       = 0x9e
	ttxtReleaseGraphics    = 0x9f
)

var teleTextFont *font.DecalFont

func init() {
	var err error
	teleTextFont, err = font.LoadFromFile("fonts/teletext.yaml")
	if err != nil {
		panic(err)
	}
}

// Codes assumed at new line start
var lineStartCodes = []byte{
	ttxtColorWhite,
	ttxtSteady,
	ttxtNormalHeight,
	ttxtContiguousGraphics,
	ttxtBlackBackground,
	ttxtReleaseGraphics,
}

type TeleTextDisplayMode int

const (
	ttxtCharSteady TeleTextDisplayMode = iota
	ttxtCharFlash
	ttxtCharConceal
)

type TeleTextHeight int

const (
	ttxtHeightNormal TeleTextHeight = iota
	ttxtHeightDouble
)

type TeleTextGraphicsMode int

const (
	ttxtGraphicsContiguous TeleTextGraphicsMode = iota
	ttxtGraphicsSeparated
)

type TeleTextCell struct {
	ch              rune
	colidx, bcolidx int
	displayMode     TeleTextDisplayMode
	height          TeleTextHeight
}

type TeleTextProcessor struct {
	bgColor      int
	fgColor      int
	heldGraphic  byte
	displayMode  TeleTextDisplayMode
	height       TeleTextHeight
	graphicsMode TeleTextGraphicsMode
	holdGraphics bool
	graphics     bool
	buffer       [ttxtRows * ttxtCols]TeleTextCell
	x, y         int
}

func (ttp *TeleTextProcessor) Process(code byte, hold bool) {

	visible := false

	switch code {
	// fg colors
	case ttxtColorRed:
		ttp.fgColor = 1
		ttp.graphics = false
	case ttxtColorGreen:
		ttp.fgColor = 2
		ttp.graphics = false
	case ttxtColorYellow:
		ttp.fgColor = 3
		ttp.graphics = false
	case ttxtColorBlue:
		ttp.fgColor = 4
		ttp.graphics = false
	case ttxtColorMagenta:
		ttp.fgColor = 5
		ttp.graphics = false
	case ttxtColorCyan:
		ttp.fgColor = 6
		ttp.graphics = false
	case ttxtColorWhite:
		ttp.fgColor = 7
		ttp.graphics = false
	// bg
	case ttxtBlackBackground:
		ttp.bgColor = 0
	case ttxtNewBackground:
		ttp.fgColor = ttp.fgColor
	// flash
	case ttxtFlash:
		ttp.displayMode = ttxtCharFlash
	case ttxtSteady:
		ttp.displayMode = ttxtCharSteady
	case ttxtConcealDisplay:
		ttp.displayMode = ttxtCharConceal
	// graphics
	case ttxtHoldGraphics:
		ttp.holdGraphics = true
	case ttxtReleaseGraphics:
		ttp.holdGraphics = false
	case ttxtSeparatedGraphics:
		ttp.graphicsMode = ttxtGraphicsSeparated
	case ttxtContiguousGraphics:
		ttp.graphicsMode = ttxtGraphicsContiguous
	case ttxtGraphicsRed:
		ttp.fgColor = 1
		ttp.graphics = true
	case ttxtGraphicsGreen:
		ttp.fgColor = 2
		ttp.graphics = true
	case ttxtGraphicsYellow:
		ttp.fgColor = 3
		ttp.graphics = true
	case ttxtGraphicsBlue:
		ttp.fgColor = 4
		ttp.graphics = true
	case ttxtGraphicsMagenta:
		ttp.fgColor = 5
		ttp.graphics = true
	case ttxtGraphicsCyan:
		ttp.fgColor = 6
		ttp.graphics = true
	case ttxtGraphicsWhite:
		ttp.fgColor = 7
		ttp.graphics = true
	default:
		visible = true
		// hold last code with bit 5 set
		if code&32 != 0 {
			ttp.heldGraphic = code
		}
	}

	ch := rune(code)
	if ttp.graphics {
		ch += 512 // remember to encode glyphs here
		if ttp.graphicsMode == ttxtGraphicsSeparated {
			ch += 256
		}
	}

	if !visible {
		ch = 32 // blank space
	}

	ttp.buffer[ttp.x+ttp.y*ttxtCols] = TeleTextCell{
		ch:          ch,
		bcolidx:     ttp.bgColor,
		colidx:      ttp.fgColor,
		displayMode: ttp.displayMode,
		height:      ttp.height,
	}

	if !hold {
		ttp.x += 1
		if ttp.x >= ttxtCols {
			ttp.x -= ttxtCols
		}
	}
}

func (ttp *TeleTextProcessor) NewLine() {
	ttp.y += 1
	if ttp.y >= ttxtRows {
		ttp.y = ttp.y - ttxtRows
	}
	ttp.x = 0
	for _, code := range lineStartCodes {
		ttp.Process(code, true)
	}
}

func (ttp *TeleTextProcessor) ProcessBuffer(in []uint64) [ttxtRows * ttxtCols]TeleTextCell {
	if len(in) < 1000 {
		return ttp.buffer
	}
	ttp.y = -1
	var code byte
	var idx int
	for y := 0; y < ttxtRows; y++ {
		ttp.NewLine()
		for x := 0; x < ttxtCols; x++ {
			idx = y*ttxtCols + x
			//fmt.Printf("x=%d,y=%d,idx=%d\n", x, y, idx)
			code = byte(in[idx])
			ttp.Process(code, false)
		}
	}
	return ttp.buffer
}
