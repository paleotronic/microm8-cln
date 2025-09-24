// +build !remint

package main

import (
	"bytes"
	"os"
	"unsafe"

	"image"
	"image/jpeg"
	"image/png"

	//"github.com/disintegration/imaging"
	"paleotronic.com/gl"
)

func ScreenShot(x, y, width, height int) *image.NRGBA {

	if width == 0 || height == 0 {
		return nil
	}

	i := image.NewRGBA(image.Rect(0, 0, width, height))
	data := unsafe.Pointer(&i.Pix[0])

	gl.ReadPixels(int32(x), int32(y), int32(width), int32(height), gl.RGBA, gl.UNSIGNED_BYTE, data)

	// because GL, this image is upsidedown...
	ni := image.NewNRGBA(i.Bounds())

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			ni.Set(x, height-y-1, i.At(x, y))
		}
	}

	return ni

}

func ScreenShotPNG(x, y, width, height int, file string) error {

	raw := ScreenShot(x, y, width, height)

	f, e := os.Create(file)
	if e != nil {
		return e
	}

	png.Encode(f, raw)

	return f.Close()
}

func ScreenShotJPG(x, y, width, height int) []byte {

	raw := ScreenShot(x, y, width, height)

	if raw == nil {
		return []byte(nil)
	}

	b := bytes.NewBuffer([]byte(nil))

	jpeg.Encode(b, raw, nil)

	return b.Bytes()
}

var enc = png.Encoder{CompressionLevel: -2}

func ScreenShotPNGBytes(x, y, width, height int) []byte {

	raw := ScreenShot(x, y, width, height)
	//scaled := resize.Resize(342, 192, raw, resize.Bilinear)

	if raw == nil {
		return []byte(nil)
	}

	b := bytes.NewBuffer([]byte(nil))

	enc.Encode(b, raw)

	return b.Bytes()
}
