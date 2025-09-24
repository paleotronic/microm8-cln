package video

const (
	hBump        = 1024
	vBump        = 2048
	bufferWidth  = 80
	bufferHeight = 48
	bufferSize   = 4096
	xPos         = 4094
	yPos         = 4095
)

func (t *BaseLayer) wozVInterlace(y int) int {
	return y % 2
}

func (t *BaseLayer) wozHInterlace(x int) int {
	return x % 2
}

func (t *BaseLayer) wozHInterlaceAlt(x int) int {
	return (x + 1) % 2
}

// baseOffsetModXY returns the given memory offset address,
func (t *BaseLayer) baseOffsetWozModXY(x, y int) (int, int, int) {
	return t.wozVInterlace(y)*vBump + t.wozHInterlace(x)*hBump, x / 2, y / 2
}

func (t *BaseLayer) baseOffsetWozModXYAlt(x, y int) (int, int, int) {
	return t.wozVInterlace(y)*vBump + t.wozHInterlaceAlt(x)*hBump, x / 2, y / 2
}

// Returns entire memory offset
func (t *BaseLayer) baseOffsetWoz(x, y int) int {
	base, mx, my := t.baseOffsetWozModXY(x, y)

	// At this point base refers to the interlaced Quadrant base
	// address in memory, mx and my are divd by 2 to yield co-ords
	// ammmenable to the standard memory map Woz calcs

	jump := ((my % 8) * 128) + ((my / 8) * 40) + mx

	return base + jump
}

func (t *BaseLayer) baseOffsetWozAlt(x, y int) int {
	base, mx, my := t.baseOffsetWozModXYAlt(x, y)

	// At this point base refers to the interlaced Quadrant base
	// address in memory, mx and my are divd by 2 to yield co-ords
	// ammmenable to the standard memory map Woz calcs

	jump := ((my % 8) * 128) + ((my / 8) * 40) + mx

	return base + jump
}
func (t *BaseLayer) baseOffsetLinear(x, y int) int {
	return (x % t.Width) + (y%t.Height)*t.Width
}
