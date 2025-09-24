package video

import (
	"paleotronic.com/octalyzer/video/font"
)

var fontpixel string = "images/pixel-glow-crop-sl2.png"
var fontpixelhalf string = "images/pixel-glow-crop-half.png"
var fontpixel80 string = "images/pixel-brick.png"

var READYFONTS bool

func LoadNormalFont() *font.DecalFont {
	// tex40, e40 := font.LoadPNG(fontpixel)
	// if e40 != nil {
	// 	log.Fatal(e40)
	// }
	// tex80, e80 := font.LoadPNG(fontpixel80)
	// if e80 != nil {
	// 	log.Fatal(e80)
	// }
	this := font.NewDecalFont(
		&font.Font{
			Name:    "Apple ][",
			Width:   7,
			Height:  8,
			Inverse: true,
			Sources: []font.FontRange{
				font.FontRange{Source: "fonts/Pr21Normal_0.png", Start: 32, End: 288, XPad: 1},
				font.FontRange{Source: "fonts/Pr21Alt_0.png", Start: 1024 + 32, End: 1024 + 512, XPad: 1},
			},
		},
	)

	READYFONTS = true

	return this
}
