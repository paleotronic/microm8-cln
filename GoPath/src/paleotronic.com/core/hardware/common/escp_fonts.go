package common

import (
	"bytes"
	"io"
	"strings"

	"paleotronic.com/octalyzer/assets"
)

var fxFontList = []string{
	"FXMatrix105MonoComprExpItalic",
	"FXMatrix105MonoComprExpRegular",
	"FXMatrix105MonoComprItalic",
	"FXMatrix105MonoComprRegular",
	"FXMatrix105MonoEliteExpItalic",
	"FXMatrix105MonoEliteExpRegular",
	"FXMatrix105MonoEliteItalic",
	"FXMatrix105MonoEliteRegular",
	"FXMatrix105MonoPicaExpItalic",
	"FXMatrix105MonoPicaExpRegular",
	"FXMatrix105MonoPicaItalic",
	"FXMatrix105MonoPicaRegular",
}

func getFXFontData(name string) io.Reader {
	return bytes.NewBuffer(assets.MustAsset("fonts/" + name + ".ttf"))
}

type fxProportionalMode bool

func (m fxProportionalMode) String() string {
	if m {
		return "Prop"
	}
	return "Mono"
}

type fxPitch int

const (
	fxPitchPica      fxPitch = 0
	fxPitchElite     fxPitch = 1
	fxPitchCondensed fxPitch = 2
)

func (m fxPitch) String() string {
	switch m {
	case fxPitchPica:
		return "Pica"
	case fxPitchElite:
		return "Elite"
	case fxPitchCondensed:
		return "Compr"
	}
	return "Pica"
}

type fxExpandedMode bool

func (m fxExpandedMode) String() string {
	if m {
		return "Exp"
	}
	return ""
}

type fxDoubleStrikeMode bool

func (m fxDoubleStrikeMode) String() string {
	if m {
		return "Dbl"
	}
	return ""
}

type fxUnderlineMode bool

func (m fxUnderlineMode) String() string {
	if m {
		return "UL"
	}
	return ""
}

type fxSubMode int

const (
	fxSubModeNone        fxSubMode = 0
	fxSubModeSubscript   fxSubMode = 1
	fxSubModeSuperscript fxSubMode = 2
)

func (m fxSubMode) String() string {
	switch m {
	case fxSubModeNone:
		return ""
	case fxSubModeSubscript:
		return "Sub"
	case fxSubModeSuperscript:
		return "Sup"
	}
	return ""
}

type fxStyle int

const (
	fxStyleRegular    fxStyle = 0
	fxStyleBold       fxStyle = 1
	fxStyleItalic     fxStyle = 2
	fxStyleBoldItalic fxStyle = 3
)

func (m fxStyle) String() string {
	switch m {
	case fxStyleRegular:
		return "Regular"
	case fxStyleBold:
		return ""
	case fxStyleItalic:
		return "Italic"
	case fxStyleBoldItalic:
		return "Italic"
	}
	return "Regular"
}

type fxFontStyle struct {
	imgWriter    bool
	proportional fxProportionalMode
	pitch        fxPitch
	expanded     fxExpandedMode
	doubleStrike fxDoubleStrikeMode
	underline    fxUnderlineMode
	submode      fxSubMode
	style        fxStyle
}

func (m fxFontStyle) String() string {
	if m.imgWriter {
		if m.proportional {
			return "imgwriter-variable"
		} else {
			return "imgwriter-draft"
		}
	}
	return strings.Join(
		[]string{
			"FXMatrix",
			"105",
			m.proportional.String(),
			m.pitch.String(),
			m.expanded.String(),
			//m.doubleStrike.String(),
			//m.underline.String(),
			//m.submode.String(),
			m.style.String(),
		},
		"",
	)
}

func (m fxFontStyle) Cpi() float64 {
	switch m.pitch.String() + m.expanded.String() {
	case "Pica":
		return 10.0
	case "Elite":
		return 12.0
	case "Compr":
		return 17.16
	case "PicaExp":
		return 5
	case "EliteExp":
		return 6
	case "ComprExp":
		return 8.58
	}
	return 10
}
