package video

import "time"

const (
	ZXVideoBytesPerLine = 32
)

func (this *GraphicsLayer) FetchUpdatesZXVideo() {
	//fmt.Printf("scooping data for render %d bytes @ $%.4x\n", this.Buffer.Size, this.Buffer.GStart[0])
	// Capture framedata and palette at Fetch cycle
	this.framedata = this.Buffer.ReadSlice(0, this.Buffer.Size)
	//fmt.Printf("%+v\n", this.Buffer.Read(0))
	this.fscanchanged = this.scanchanged
	for i, _ := range this.scanchanged {
		this.scanchanged[i] = false
	}
}

func (this *GraphicsLayer) ZXGetScanByteAtXY(buffer []uint64, x int, y int) uint64 {
	base := 0x0000
	ybits := ((y & 7) << 8) | ((y & 192) << 5) | ((y & 56) << 2)
	offs := base + (ybits | (x & 31))
	return buffer[offs]
}

func (this *GraphicsLayer) ZXGetAttrAtXY(buffer []uint64, x int, y int) (fg int, bg int, flash bool) {
	base := 0x1800
	offs := base + ((y / 8) * ZXVideoBytesPerLine) + x
	b := buffer[offs]
	fg = (int(b) & 7)
	bg = ((int(b) >> 3) & 7)
	if b&64 != 0 {
		fg |= 8
		bg |= 8
	}
	flash = (b & 128) != 0
	return
}

func (this *GraphicsLayer) ZXColorsForScanline(buffer []uint64, y int) []int {
	var out = make([]int, this.Width)
	if y < 0 || y > this.Height {
		return out
	}
	var b uint64
	var fg, bg, tc int
	var flash bool
	var flashOn = time.Now().UnixNano()%1000000000 >= 500000000
	for x := 0; x < ZXVideoBytesPerLine; x++ {
		b = this.ZXGetScanByteAtXY(buffer, x, y)
		fg, bg, flash = this.ZXGetAttrAtXY(buffer, x, y)
		if flash && flashOn {
			tc = fg
			fg = bg
			bg = tc
		}
		for i := 0; i < 8; i++ {
			if b&128 != 0 {
				out[x*8+i] = fg
			} else {
				out[x*8+i] = bg
			}
			b <<= 1
		}
	}
	return out
}

func (this *GraphicsLayer) MakeUpdatesZXVideo() {

	this.Changed = false

	var out []int

	for y := 0; y < this.Height; y++ {
		out = this.ZXColorsForScanline(this.framedata, y)
		for x, c := range out {
			this.PlotPixel(x, y, int(c))
			this.Plot(x, y, int(c))
		}
	}

}
