package hires

import (
	"paleotronic.com/fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"testing"
)

func TestKeypressPackUnpack(t *testing.T) {

	var hgr HGRScreen
	var offset int

	offset = hgr.XYToOffset(0, 0)

	if offset != 0 {
		t.Error("HGR memory offset (0,0) != 0")
	}

	offset = hgr.XYToOffset(0, 64)
	if offset != 40 {
		t.Error("HGR memory offset (0,64) != 40")
	}

	offset = hgr.XYToOffset(0, 128)
	if offset != 80 {
		t.Error("HGR memory offset (0,128) != 80")
	}

	offset = hgr.XYToOffset(0, 8)
	if offset != 128 {
		t.Error("HGR memory offset (0,8) != 128")
	}

	offset = hgr.XYToOffset(0, 72)
	if offset != 168 {
		t.Error("HGR memory offset (0,72) != 168")
	}

	offset = hgr.XYToOffset(0, 1)
	if offset != 1024 {
		t.Error("HGR memory offset (0,1) != 1024")
	}

}

func TestScreenPackUnpack(t *testing.T) {

	var hgr HGRScreen

	b, err := ioutil.ReadFile("files/SCREEN")
	if err != nil {
		t.Error("Failed to read file")
	}

	for i, v := range b {
		hgr.Data[i] = v
	}

	var colmap color.Palette
	colmap = append(colmap, color.RGBA{000, 000, 000, 255})
	colmap = append(colmap, color.RGBA{000, 255, 000, 255})
	colmap = append(colmap, color.RGBA{255, 000, 255, 255})
	colmap = append(colmap, color.RGBA{255, 255, 255, 255})
	colmap = append(colmap, color.RGBA{000, 000, 000, 255})
	colmap = append(colmap, color.RGBA{255, 128, 000, 255})
	colmap = append(colmap, color.RGBA{0, 128, 255, 255})
	colmap = append(colmap, color.RGBA{255, 255, 255, 255})

	img := image.NewRGBA(image.Rect(0, 0, 280, 192))

	for y := 0; y < 192; y++ {
		offset := hgr.XYToOffset(0, y)
		scanline := hgr.ColorsForScanLine(hgr.Data[offset : offset+40])
		for x := 0; x < 280; x++ {
			////fmt.Printf("%d, ", scanline[x])
			img.Set(x, y, colmap[scanline[x]])
		}
	}

	out, err := os.Create("files/output.png")
	if err != nil {
		t.Error("Failed to open output")
	}

	err = png.Encode(out, img)
	if err != nil {
		t.Error("Failed to write Png")
	}

}
