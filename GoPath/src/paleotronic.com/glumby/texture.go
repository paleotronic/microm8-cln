package glumby

import (
	//"errors"

	"image"
	"image/color"
	"image/draw"
	_ "image/gif" // used for image loading
	_ "image/jpeg"
	_ "image/png"
	"io" //	"paleotronic.com/log"
	"os"

	"paleotronic.com/fmt"
	"paleotronic.com/gl"
)

// Texture wrapper class
type Texture struct {
	source string
	handle uint32
	w, h   int
}

func GetMaxTextureSize() int32 {
	var v int32
	gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &v)
	return v
}

func NewSolidColorTexture(solid color.RGBA) *Texture {
	i := image.NewRGBA(image.Rect(0, 0, 1, 1))

	draw.Draw(i, image.Rect(0, 0, 1, 1), image.NewUniform(solid), image.ZP, draw.Src)

	return NewTextureFromRGBA(i)
}

// NewTextureFromRGBA creates a texture from a raw RGBA image
func NewTextureFromRGBA(rgba *image.RGBA) *Texture {
	var texture uint32
	gl.Enable(gl.TEXTURE_2D)
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.Disable(gl.TEXTURE_2D)

	return &Texture{handle: texture, w: rgba.Rect.Size().X, h: rgba.Rect.Size().Y}
}

// NewTexture creates a texture from a file (PNG, GIF, JPEG)
func NewTextureFromBytes(r io.Reader) (*Texture, error) {
	img, _, err := image.Decode(r)
	////fmt.Printntln(s)
	if err != nil {
		return nil, err
	}

	b := img.Bounds()
	m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m, m.Bounds(), img, b.Min, draw.Src)
	return NewTextureFromRGBA(m), nil
}

// NewTexture creates a texture from a file (PNG, GIF, JPEG)
func NewTexture(file string) (*Texture, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, err
	}

	b := img.Bounds()
	m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m, m.Bounds(), img, b.Min, draw.Src)

	return NewTextureFromRGBA(m), nil
}

// Texture handle for opengl
func (this *Texture) Handle() uint32 {
	return this.handle
}

// Texture Width
func (this *Texture) Width() int {
	return this.w
}

// Texture Height
func (this *Texture) Height() int {
	return this.h
}

// Update the image used for the texture, causing a rebind
func (this *Texture) SetSource(rgba *image.RGBA) {
	this.Bind()
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))
	this.w = rgba.Rect.Size().X
	this.h = rgba.Rect.Size().Y
	this.source = "image.RGBA"
	this.Unbind()

}

// Bind the texture to the GPU for processing
func (this *Texture) Bind() {
	if this == nil {
		//log.Fatalln("THIS is nil")
		return
	}

	gl.BindTexture(gl.TEXTURE_2D, this.handle)
}

// Unbind the texture from opengl
func (this *Texture) Unbind() {
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func NewTextureBlank(width, height int, color color.RGBA) *Texture {
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{color}, image.ZP, draw.Src)

	return NewTextureFromRGBA(rgba)
}

// Draw will draw part an RGBA image onto the texture at the given position
func (t *Texture) Draw(bind bool, x, y int, rgba *image.RGBA) {

	if bind {
		t.Bind()
	}

	gl.TexSubImage2D(
		gl.TEXTURE_2D,
		0,
		int32(x),
		int32(y),
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix),
	)

}

func (this *Texture) SetSourceSame(rgba *image.RGBA) {

	if rgba.Rect.Size().X != this.w {
		panic(fmt.Sprintf("not same width - expected %d, got %d", this.w, rgba.Rect.Size().X))
	}

	if rgba.Rect.Size().Y != this.h {
		panic(fmt.Sprintf("not same height - expected %d, got %d", this.h, rgba.Rect.Size().Y))
	}

	this.Bind()
	gl.TexSubImage2D(
		gl.TEXTURE_2D,
		0,
		0,
		0,
		int32(this.w),
		int32(this.h),
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix),
	)
	this.Unbind()

}

func (this *Texture) SetSourcePartial(x, y int32, rgba *image.RGBA) {

	//~ if rgba.Rect.Size().X != this.w {
	//~ panic(fmt.Sprintf("not same width - expected %d, got %d", this.w, rgba.Rect.Size().X))
	//~ }

	//~ if rgba.Rect.Size().Y != this.h {
	//~ panic(fmt.Sprintf("not same height - expected %d, got %d", this.h, rgba.Rect.Size().Y))
	//~ }

	this.Bind()
	gl.TexSubImage2D(
		gl.TEXTURE_2D,
		0,
		x,
		y,
		int32(rgba.Bounds().Max.X),
		int32(rgba.Bounds().Max.Y),
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix),
	)
	this.Unbind()

}

//~ func (t *Texture) Freshen() {
//~ rgba := t.Bitmap
//~ var texture uint32 = t.handle
//~ gl.Enable(gl.TEXTURE_2D)
//~ gl.BindTexture(gl.TEXTURE_2D, texture)

//~ gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(t.w), int32(t.h), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
//~ }

// Draw a glyph with up to 2 masks and colors
func (t *Texture) DrawGlyph(bind bool, x, y int, aglyph *image.RGBA, bglyph *image.RGBA, ocol, acol, bcol color.Color) {

	// assumptions aglyph and bglyph are same dimensions
	dst := image.NewRGBA(aglyph.Bounds())

	// original color...
	draw.Draw(dst, dst.Bounds(), &image.Uniform{ocol}, image.ZP, draw.Src)

	// 'a' glyph
	if aglyph != nil {
		src := &image.Uniform{acol}
		mask := aglyph
		mr := aglyph.Bounds()
		draw.DrawMask(dst, mr.Sub(mr.Min), src, image.ZP, mask, mr.Min, draw.Over)
	}

	// 'b' glyph
	if bglyph != nil {
		src := &image.Uniform{bcol}
		mask := bglyph
		mr := bglyph.Bounds()
		draw.DrawMask(dst, mr.Sub(mr.Min), src, image.ZP, mask, mr.Min, draw.Over)
	}

	// draw to bitmap

	//func Draw(dst Image, r image.Rectangle, src image.Image, sp image.Point, op Op)
	//draw.Draw( t.Bitmap, aglyph.Bounds().Add(image.Pt(x, y)), dst, image.ZP, draw.Over )

}
