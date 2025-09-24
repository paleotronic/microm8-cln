package decalfont

import (
	"bufio"
	//	"paleotronic.com/fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/nfnt/resize"
	"paleotronic.com/glumby"
	"paleotronic.com/log"
)

type DecalDuckFont struct {
	GlyphWidth  int
	DotHeight   int
	col         color.Color
	ScaleX      float32
	SpacingH    int
	GlyphHeight int
	DotWidth    int
	ScaleY      float32
	Tex, Tex80  image.Image
	camera      *glumby.Camera
	Pix         image.Image
	TextHeight  int
	ScaleZ      float32
	cpl         int
	TextWidth   int
	SpacingV    int
	baseLine    int
	used        []*glumby.Mesh
	batchCount  int
	Chars       map[rune]*glumby.Texture
	Chars80     map[rune]*glumby.Texture
}

func LoadNormalFont() DecalDuckFont {
	tex40, e40 := loadPNG("images/pixel-glow-square.png")
	if e40 != nil {
		log.Fatal(e40)
	}
	tex80, e80 := loadPNG("images/pixel-brick.png")
	if e80 != nil {
		log.Fatal(e80)
	}
	this := NewDecalDuckFont("fonts/Pr21Normal_0.png", 32, 192, 7, 8, 1, 0, tex40, tex80)
	return this
}

func LoadInvertedFont() DecalDuckFont {
	tex40, e40 := loadPNG("images/pixel-glow-square.png")
	if e40 != nil {
		log.Fatal(e40)
	}
	tex80, e80 := loadPNG("images/pixel-brick.png")
	if e80 != nil {
		log.Fatal(e80)
	}
	this := NewDecalDuckFont("fonts/Pr21Inverted_0.png", 32, 192, 7, 8, 1, 0, tex40, tex80)
	return this
}

func loadPNG(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	r := bufio.NewReader(f)
	defer f.Close()

	return png.Decode(r)
}

func fillRGBA(img *image.RGBA, c color.RGBA) {
	//log.Printf("Fill image with bounds %d, %d with %v\n", img.Bounds().Max.X, img.Bounds().Max.Y, c)
	for y := 0; y <= img.Bounds().Max.Y; y++ {
		for x := 0; x <= img.Bounds().Max.X; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

func (this *DecalDuckFont) GetScaleZ() float32 {
	// TODO Auto-generated method stub
	return this.ScaleZ
}

func (this *DecalDuckFont) ReadBuffer(pix image.Image, ox, oy int, w, h int, t image.Image) *glumby.Texture {
	//System.Out.Println("Load char from ox = "+ox+", oy = "+oy);

	sb := image.NewRGBA(image.Rect(0, 0, this.GlyphWidth, this.GlyphHeight))
	fillRGBA(sb, color.RGBA{R: 0, G: 255, B: 0, A: 4})

	//solid := color.RGBA{255, 255, 255, 255}

	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			//Color cc = types.NewColor(p.GetPixel(ox+c, oy+r))
			cc := pix.At(ox+c, oy+(this.TextHeight-r-1))
			_, _, _, aa := cc.RGBA()
			//System.Out.Println("Pixel is r="+cc.R+", g="+cc.G+", b="+cc.B+", a="+cc.A);

			sr := t.Bounds()
			dp := image.Pt(c*this.DotWidth, r*this.DotHeight)
			rr := image.Rectangle{dp, dp.Add(sr.Size())}

			//draw.Draw(sb,
			//	image.Rect(0, 0, 10, 10),
			//	&image.Uniform{solid},
			//	image.ZP,
			//	draw.Over)

			if aa >= 32768 {
				//if (r+c)%2 == 1 {
				//sb.DrawPixmap(t, 0, 0, t.GetWidth(), t.GetHeight(), c * DotWidth, r * DotHeight, DotWidth, DotHeight)
				draw.Draw(sb,
					//image.Rect(c*this.DotWidth, r*this.DotHeight, (c+1)*this.DotWidth, (r+1)*this.DotHeight),
					rr,
					t, //&image.Uniform{solid}, //t,
					image.ZP,
					draw.Src)
				//////fmt.Print("#")
			} else {
				//////fmt.Print(" ")
			}
		}
		//System.Out.Println();
		//////fmt.Println()
	}

	tx := glumby.NewTextureFromRGBA(sb)

	return tx
}

func (this *DecalDuckFont) LoadCharacters(start int, end int) {

	for idx := start; idx <= end; idx++ {
		v := idx - start
		ox := (v % this.cpl) * (this.TextWidth + this.SpacingH)
		oy := (v / this.cpl) * (this.TextHeight + this.SpacingV)

		//System.Out.Println("Char is "+idx);
		sb := this.ReadBuffer(this.Pix, ox, oy, this.TextWidth, this.TextHeight, this.Tex)
		this.Chars[rune(idx)] = sb
		sb80 := this.ReadBuffer(this.Pix, ox, oy, this.TextWidth, this.TextHeight, this.Tex80)
		this.Chars80[rune(idx)] = sb80
	}

	//os.Exit(0)

}

func (this *DecalDuckFont) GetScaleX() float32 {
	return this.ScaleX
}

func (this *DecalDuckFont) SetScaleX(scaleX float32) {
	this.ScaleX = scaleX
}

func (this *DecalDuckFont) SetScale(x float32, y float32) {
	this.SetScaleX(x)
	this.SetScaleY(y)
}

func (this *DecalDuckFont) Draw(text string, attr DecalAttr, c int, r int, w int, h int, screen []*Decal) {

	if (c < 0) || (c >= w) || (r < 0) || (r >= h) {
		return
	}

	//System.Out.Println("DecalFont Draw called ["+text+"]");

	for _, ch := range text {
		sb, exists := this.Chars[ch]
		if !exists {
			continue // cant plot what we dont have
		}

		if w == 80 {
			sb, exists = this.Chars80[ch]
		}

		//System.Out.Println("Got DECAL char "+ch);

		index := c + (r * w)
		if (index < 0) || (index >= len(screen)) {
			continue
		}

		decal := screen[index]
		//if (w != 80)
		decal.Texture = sb
		decal.Name = string(ch) + string(attr)
		//decal.SetWidth( this.TextWidth * this.ScaleX );
		//decal.SetHeight( this.TextHeight * this.ScaleY );

		// advance
		c++
	}

}

func NewDecalDuckFont(file string, start int, end int, textWidth int, textHeight int, hSpace int, vSpace int, pixel image.Image, pixel80 image.Image) DecalDuckFont {
	this := DecalDuckFont{}

	pix, err := loadPNG(file)
	if err != nil {
		log.Fatalf("Unable to load image buffer from %s: %s\n", file, err.Error())
	}

	//if file == "fonts/Pr21Inverted_0.png" {
	//	log.Fatalf("%s(0,0) [%v]: %d, %d, %d, %d\n", file, cc, r, g, b, a)
	//}
	this.Pix = pix

	width := pix.Bounds().Max.X
	//height := pix.Bounds().Max.Y

	this.GlyphHeight = 56
	this.GlyphWidth = 56

	this.TextWidth = textWidth
	this.TextHeight = textHeight
	this.SpacingH = hSpace
	this.SpacingV = vSpace
	//this.Tex = pixel
	//this.Tex80 = pixel80
	this.DotWidth = this.GlyphWidth / this.TextWidth
	this.DotHeight = this.GlyphHeight / this.TextHeight

	this.cpl = width / (this.TextWidth + this.SpacingH)

	this.Chars = make(map[rune]*glumby.Texture)
	this.Chars80 = make(map[rune]*glumby.Texture)

	this.Tex = resize.Resize(uint(this.DotWidth), uint(this.DotHeight), pixel, resize.Lanczos3)
	this.Tex80 = resize.Resize(uint(this.DotWidth), uint(this.DotHeight), pixel80, resize.Lanczos3)

	this.LoadCharacters(start, end)

	return this
}

func (this *DecalDuckFont) GetColor() color.Color {
	return this.col
}

func (this *DecalDuckFont) SetColor(col color.Color) {
	this.col = col
}

func (this *DecalDuckFont) GetScaleY() float32 {
	return this.ScaleY
}

func (this *DecalDuckFont) SetScaleY(scaleY float32) {
	this.ScaleY = scaleY
}
