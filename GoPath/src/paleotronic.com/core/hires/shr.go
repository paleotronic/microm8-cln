package hires

//"paleotronic.com/core/dialect"
//"paleotronic.com/core/interfaces"
//"paleotronic.com/core/types"
import (
	"log"
	"math"
	"strconv"

	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
)

const (
	shrWidth  = 640
	shrHeight = 200
)

type SHRColor int

type SHRPalette [16]SHRColor

var (
	palette640 = SHRPalette{
		0x000,
		0x00f,
		0xff0,
		0xfff,
		0x000,
		0xd00,
		0x0e0,
		0xfff,
		0x000,
		0x00f,
		0xff0,
		0xfff,
		0x000,
		0xd00,
		0x0e0,
		0xfff,
	}
	ditherTable = [16][4]int{
		{1, 1, 1, 1},
		{2, 1, 2, 1},
		{3, 1, 3, 1},
		{4, 1, 4, 1},
		{1, 2, 1, 2},
		{2, 2, 2, 2},
		{3, 2, 3, 2},
		{4, 2, 4, 2},
		{1, 3, 1, 3},
		{2, 3, 2, 3},
		{3, 3, 3, 3},
		{4, 3, 4, 3},
		{1, 4, 1, 4},
		{2, 4, 2, 4},
		{3, 4, 3, 4},
		{4, 4, 4, 4},
	}
	palette640dithered = SHRPalette{
		0x000, // black
		0x00b, // blue
		0xbb0, // yellow
		0xddd, // grey
		0xb00, // red
		0xb0b, // purple
		0xb80, // orange
		0xf88, // ltred
		0x0b0, // green
		0x0aa, // aqua
		0x8a0, // lime
		0x8f8, // ltgrn
		0xddd, // grey2
		0x88f, // ltblue
		0xff8, // ltyellow
		0xfff, // white
	}
)

type SuperHiResBuffer struct {
	duplicate         map[int]int
	data              *memory.MemoryControlBlock
	width             int
	height            int
	txns              map[int]uint64
	Spritedata        map[int]int
	CollisionCount    int
	LASTX, LASTY      int
	InterleavedMemory bool
	palettes          [16]*SHRPalette
	r, g, b           float32
	tint              bool
}

// Calculate memory offset into 32KB bankm based on interleaving mode
func (hgr *SuperHiResBuffer) offset(o int) int {
	if hgr.InterleavedMemory {
		return ((o+1)%2)*16384 + o/2
	}
	return o
}

func (hgr *SuperHiResBuffer) Init640() {
	for i := 0; i < 16; i++ {
		hgr.setVideoPalette(i, palette640)
	}
	for y := 0; y < 200; y++ {
		hgr.data.Write(0x7d00+y, 128)
		for a := 0; a < 160; a++ {
			hgr.data.Write(160*y+a, 0)
		}
	}
}

func (hgr *SuperHiResBuffer) Init320() {
	for i := 0; i < 16; i++ {
		var p SHRPalette
		for i, v := range settings.DefaultSHR320Palette {
			p[i] = SHRColor(v)
		}
		hgr.setVideoPalette(i, p)
	}
	for y := 0; y < 200; y++ {
		hgr.data.Write(0x7d00+y, 0)
		for a := 0; a < 160; a++ {
			hgr.data.Write(160*y+a, 0)
		}
	}
}

func (hgr *SuperHiResBuffer) GetPalette() SHRPalette {
	if hgr.GetWidth() == 640 {
		return palette640dithered
	}
	return *hgr.getVideoPalette(0)
}

func (hgr *SuperHiResBuffer) SetTint(r, g, b float32, tint bool) {
	//log.Printf("SetTint(%f, %f, %f, %v)", r, g, b, tint)
	hgr.tint = tint
	hgr.r = r
	hgr.g = g
	hgr.b = b
	hgr.LoadPaletteCache()
}

func (hgr *SuperHiResBuffer) getScanlineOffset(y int) int {
	if y < 0 || y > 199 {
		return 0
	}
	return 160 * y
}

func (hgr *SuperHiResBuffer) grabChunk(begin, end int) []byte {
	size := end - begin
	out := make([]byte, size)
	for i, _ := range out {
		out[i] = byte(hgr.data.Read(hgr.offset(i)) & 0xff)
	}
	return out
}

func (hgr *SuperHiResBuffer) getOffsetForScanlineConfig(y int) int {
	return 0x7d00 + y
}

func (hgr *SuperHiResBuffer) ClearPaletteCache(p int) {
	log.Printf("Clearing cache for palette %d", p)
	hgr.palettes[p%16] = nil
}

func (hgr *SuperHiResBuffer) LoadPaletteCache() {
	for i := 0; i < 16; i++ {
		hgr.palettes[i] = nil
		_ = hgr.getVideoPalette(i)
	}
}

func (hgr *SuperHiResBuffer) setVideoPalette(pidx int, p SHRPalette) {
	pidx = pidx % 16
	pbegin := 0x7e00 + pidx*32

	for i, c := range p {
		hgr.data.Write(pbegin+i*2+1, uint64(c)>>8)
		hgr.data.Write(pbegin+i*2+0, uint64(c&0xff))
	}

	hgr.palettes[pidx] = &p
}

func (hgr *SuperHiResBuffer) getVideoPalette(pidx int) *SHRPalette {
	pidx = pidx % 16
	if hgr.palettes[pidx] != nil {
		return hgr.palettes[pidx]
	}
	pbegin := 0x7e00 + pidx*32
	pdata := hgr.data.ReadSlice(pbegin, pbegin+32)
	pal := SHRPalette{}
	for i := 0; i < 16; i++ {
		rr := pdata[i*2+1] & 0xf
		gg := (pdata[i*2+0] & 0xf0) >> 4
		bb := (pdata[i*2+0] & 0x0f)

		if hgr.tint {
			aa := (float32(rr) + float32(gg) + float32(bb)) / 3

			rr = uint64(aa * hgr.r)
			gg = uint64(aa * hgr.g)
			bb = uint64(aa * hgr.b)
		}

		pal[i] = SHRColor((rr << 8) | (gg << 4) | bb)
	}
	hgr.palettes[pidx] = &pal
	return &pal
}

func (hgr *SuperHiResBuffer) GetLastXY() (int, int) {
	return hgr.LASTX, hgr.LASTY
}

func (hgr *SuperHiResBuffer) OffsetToScanline(chunkaddr int) int {
	if chunkaddr < 32000 {
		return hgr.offset(chunkaddr) / 160
	}
	return 0
}

func bin(s string) int64 {
	b, _ := strconv.ParseInt(s, 2, 64)
	return b
}

var (
	mask640p0 = uint64(bin("00111111"))
	mask640p1 = uint64(bin("11001111"))
	mask640p2 = uint64(bin("11110011"))
	mask640p3 = uint64(bin("11111100"))
)

func (this *SuperHiResBuffer) Plot(x int, y int, c int) {

	c = c % 16

	if (x < 0) || (x >= this.GetWidth()) || (y < 0) || (y >= this.height) {
		return
	}

	offset := 160 * y

	switch this.GetWidth() {
	case 640:
		// convert to dither
		c = ditherTable[c][x%4] - 1
		byt := this.data.Read(offset + x/4)
		switch x % 4 {
		case 0:
			byt = (byt & mask640p0) | (uint64((c & 3)) << 6)
		case 1:
			byt = (byt & mask640p1) | (uint64((c & 3)) << 4)
		case 2:
			byt = (byt & mask640p2) | (uint64((c & 3)) << 2)
		case 3:
			byt = (byt & mask640p3) | uint64((c & 3))
		}
		this.data.Write(offset+x/4, byt)
	case 320:
		byt := this.data.Read(offset + x/2)
		if x%2 == 0 {
			byt = (byt & 0x0f) | (uint64(c&0xf) << 4)
		} else {
			byt = (byt & 0xf0) | uint64(c&0xf)
		}
		this.data.Write(offset+x/2, byt)
	}
	//this.LASTX, this.LASTY = x, y
}

func (this *SuperHiResBuffer) GetCollisionCount() uint64 {
	return uint64(this.CollisionCount)
}

func (this *SuperHiResBuffer) SetCollisionCount(v uint64) {
	this.CollisionCount = int(v)
}

func (this *SuperHiResBuffer) GetHeight() int {
	return this.height
}

func (this *SuperHiResBuffer) UnpackPixels(chunk uint64) []uint64 {
	r := make([]uint64, 4)
	for i := 0; i < 4; i++ {
		r[i] = (chunk & OffsetMask[i]) >> OffsetRotate[i]
	}
	return r
}

func (this *SuperHiResBuffer) Repaint() {
	//synchronized(this.Txns) {
	this.txns = make(map[int]uint64)
	this.txns[-2] = 0 // repaint signal
	//}
}

func (this *SuperHiResBuffer) MixedMono() bool {
	return false
}

func (this *SuperHiResBuffer) GetScanData(y int) []uint64 {

	if y == 0 {
		this.LoadPaletteCache()
	}

	var out = make([]uint64, 161)
	out[0] = this.data.Read(0x7d00 + y)
	d := this.data.ReadSlice(y*160, y*160+160)
	for i, v := range d {
		out[i+1] = v
	}
	return out
}

func (this *SuperHiResBuffer) GetData() []uint64 {
	return this.data.ReadSlice(0, this.data.Size)
}

func NewSuperHiResBuffer(data *memory.MemoryControlBlock) *SuperHiResBuffer {
	data.UseMM = false
	this := &SuperHiResBuffer{}
	this.width = 640
	this.height = 200
	this.data = data
	this.txns = make(map[int]uint64)
	this.InterleavedMemory = false
	this.tint = false
	this.r = 1
	this.g = 0.5
	this.b = 0.7
	return this
}

func (this *SuperHiResBuffer) ClearTransactions() {
	//synchronized(this.Txns) {
	this.txns = make(map[int]uint64)
	//}
}

func (this *SuperHiResBuffer) GetWidth() int {
	if this.data.Read(0x7d00)&128 != 0 {
		return 640
	}
	return 320
}

func (this *SuperHiResBuffer) CheckDuplicate(x int, y int, c int) {
	if true {
		return
	}

	index := (y * this.width) + x

	//if (c != 0) {
	//	return;
	//}

	//Integer ref = this.Duplicate.Get(index)

	_, exists := this.duplicate[index]

	for exists {
		//synchronized (this.Data) {
		if len(this.txns) == 0 {
			this.duplicate = make(map[int]int)
		}
		//}
		// perform check again
		_, exists = this.duplicate[index]
	}

	// now its okay... log the pixel and return
	this.duplicate[index] = 1
}

func (this *SuperHiResBuffer) Clear(fv uint64) {
	switch this.GetWidth() {
	case 640:
		this.Init640()
	case 320:
		this.Init320()
	}
}

func (this *SuperHiResBuffer) Fill(c int) {

	for y := 0; y < this.height; y++ {
		for x := 0; x < this.GetWidth(); x++ {
			this.Plot(x, y, c)
		}
	}

}

func (this *SuperHiResBuffer) HasTransactions() bool {
	return (len(this.txns) != 0)
}

func (this *SuperHiResBuffer) GetTransactions() map[int]uint64 {
	//HashMap<Integer,Integer> old
	old := make(map[int]uint64)

	//synchronized (txns) {
	old = this.txns
	this.txns = make(map[int]uint64)
	//}

	return old
}

func (this *SuperHiResBuffer) ColorAt(x int, y int) int {
	if (x < 0) || (x >= this.GetWidth()) || (y < 0) || (y >= this.height) {
		return 0
	}

	b := this.GetScanData(y)
	var colorFill = b[0]&32 != 0

	var c int
	var byt int

	switch this.GetWidth() {
	case 640:
		byt = int(b[1+x/4])
		switch x % 4 {
		case 0:
			c = ((byt >> 6) & 3) + 0x8
		case 1:
			c = ((byt >> 4) & 3) + 0xc
		case 2:
			c = ((byt >> 2) & 3) + 0x0
		case 3:
			c = (byt & 3) + 0x4
		}
		return c
	case 320:
		byt = int(b[1+x/2])
		if x%2 == 0 {
			byt = byt >> 4
		}
		byt &= 0xf
		if byt != 0 || !colorFill {
			c = byt
		}
		return c
	}

	return 0
}

func (this *SuperHiResBuffer) LogPixel(chunkaddr int, chunk uint64, x int, y int, c int) {

	// pause rendering if needed for dup pixels
	//	if x != -1 {
	//		this.CheckDuplicate(x, y, c)
	//	}
	//	//synchronized(this.Txns) {
	//	this.txns[chunkaddr] = chunk
	//}
}

func (this *SuperHiResBuffer) HgrSpritePlot(x int, y int, col int) {
	idx := (y * 280) + x

	if (x < 0) || (x > 279) || (y < 0) || (y > 191) {
		return
	}

	if (idx < 280*192) && (idx >= 0) {
		this.Spritedata[idx] = col
	}

}

func (this *SuperHiResBuffer) Shape(shape ShapeEntry, x int, y int, scl int, deg float64, c int, usecol bool) {

	/* vars */
	var i int
	var s int
	var nx int
	var ny int
	var px int
	var py int
	var ox float64
	var oy float64
	var b int
	var v int
	var draw bool
	var r float64

	pp := this

	this.Spritedata = make(map[int]int)

	i = 0

	ox = float64(x)
	oy = float64(y)

	px = 0
	py = 0

	/// zero collisions
	this.CollisionCount = 0

	r = math.Pi * deg / 180

	//writeln("DRAWSHAPE:");

	for i < len(shape) {
		b = shape[i]
		if b == 0 {
			for k, _ := range this.Spritedata {
				xx := k % 280
				yy := k / 280

				oc := pp.ColorAt(xx, yy)
				if usecol {
					c = this.Spritedata[k]
					if (oc != 0) && (c != 0) {
						this.CollisionCount++
					}
				} else {
					c = 7 - oc

					if oc > 3 {
						c = 7 - (oc & 3)
					} else {
						c = 3 - (oc & 3)
					}

					if oc == 0 || oc == 4 {
						this.CollisionCount++
					}
				}
				//bm[k] = c;
				//bm.Plot(xx, yy, c)
				pp.Plot(xx, yy, c)

				// update area
			}
			this.CollisionCount = this.CollisionCount % 256
			//System.Out.Println("!!!!!!!!!!!!!!!!!!!!!! Drawing sprite with "+spritedata.Size()+" points  YIELDS COLLISION COUNT "+collisionCount);
			//System.Err.Println("====== Shape end");
			return
		}

		// display byte
		//System.Err.Println("NEW Shape byte value = "+b);

		s = 0
		for (s < 3) && (b > 0) {
			v = (b & 7)

			if (s == 2) && (v == 0) {
				//System.Err.Println("!!!!! Skip s3/v0 -- NO MOVE NO PLOT");
			} else {

				nx = px
				ny = py

				draw = false

				var angle float64 = 0

				//System.Err.Println( "-----> Shape instruction code = "+ST_DESC[v]+"("+v+")" );

				switch v {
				case ST_UP_WITHOUT_PLOT:
					angle = deg
					break
				case ST_RT_WITHOUT_PLOT:
					angle = deg + 90
					break
				case ST_DN_WITHOUT_PLOT:
					angle = deg + 180
					break
				case ST_LT_WITHOUT_PLOT:
					angle = deg + 270
					break
				case ST_UP_WITH_PLOT:
					{
						angle = deg
						draw = true
						break
					}
				case ST_RT_WITH_PLOT:
					{
						angle = deg + 90
						draw = true
						break
					}
				case ST_DN_WITH_PLOT:
					{
						angle = deg + 180
						draw = true
						break
					}
				case ST_LT_WITH_PLOT:
					{
						angle = deg + 270
						draw = true
						break
					}
				}

				for angle >= 360 {
					angle -= 360
				}

				r = math.Pi * angle / 180

				/// ox, oy last point so starting point

				//boolean ignore = (v == 0) && (lastCode == 0);

				//if (!ignore) {
				for ss := 0; ss < scl; ss++ {
					if draw {
						//coll test
						if usecol {
							this.HgrSpritePlot(int(Round(ox)), int(Round(oy)), c)
						} else {
							this.HgrSpritePlot(int(Round(ox)), int(Round(oy)), 0-c)
						}
					}

					this.LASTX = int(Round(ox))
					this.LASTY = int(Round(oy))

					ox = ox + math.Sin(r)
					oy = oy - math.Cos(r)
				}

				px = nx
				py = ny

				//lastCode = v

			}

			b = (b >> 3)

			/* inc s */
			s++
		}

		i++
	}

	// draw sprite buffer

}

func (hgr *SuperHiResBuffer) ColorsForScanLine(b []uint64, mono bool) []int {

	repline := make([]int, 640)

	// raw data
	// 0     scanline control
	// 1-160 raw data
	if len(b) != 161 {
		return repline
	}
	var use640 = b[0]&128 != 0
	var colorFill = b[0]&32 != 0
	var usePalette = int(b[0] & 0xf)

	//log2.Printf("Scanline: 640? %v, useFill: %v, palette: %x", use640, colorFill, usePalette)

	var c int
	var byt int

	var p = hgr.getVideoPalette(usePalette)

	if use640 {
		c = 0
		for i := 0; i < 640; i++ {
			byt = int(b[1+i/4])
			switch i % 4 {
			case 0:
				c = ((byt >> 6) & 3) + 0x8
			case 1:
				c = ((byt >> 4) & 3) + 0xc
			case 2:
				c = ((byt >> 2) & 3) + 0x0
			case 3:
				c = (byt & 3) + 0x4
			}
			repline[i] = int(p[c])
		}
	} else {
		c = 0
		//log2.Printf("BYTE: %+v", b[1:17])
		for i := 0; i < 320; i++ {
			byt = int(b[1+i/2])
			if i%2 == 0 {
				byt = byt >> 4
			}
			byt &= 0xf
			if byt != 0 || !colorFill {
				c = byt
			}
			repline[i*2+0] = int(p[c])
			repline[i*2+1] = int(p[c])
		}
	}

	//log2.Printf("CFSL: %+v", repline[0:16])

	return repline

}

// YToRowIndex converts a y index to memory offset
func (hgr *SuperHiResBuffer) XYToOffset(x, y int) int {

	return 0
}
