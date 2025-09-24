package hires

import (
	"time"

	"paleotronic.com/fmt"
	//"paleotronic.com/fmt"
	"math"

	"paleotronic.com/core/memory"
)

const (
	stripeWidth = 128
	stripeCount = 192
	useMemColor = false
)

var DoubleWidth bool = false

/*
 * Even though we're dealing with Apple graphics, the screen map is still
 * heavily tied to the text screen layout.  Apple's HGR screen is divided
 * into 24 "blocks" of 8 lines per block (the main reason why when you do
 * a BLOAD to the HGR screen, you get the infamous "venetian blind" load.
 * Unfortunately, we end up stepping through memory (scanline-wise) as
 * follows:
 *
 *   0, 64, 128, 8, 72, 136, ..., 184, 1, 65, 129, ...
 *
 * (i.e. line 0 is the first line of the file, line 1 is the 25th line of
 * the file...wonderful, right?)
 *
 * If that isn't enough, every there are an additional 8 bytes of garbage
 * space every 3 scanlines (reminiscent of IBM's CGA architecture).
 *
 */

// HGRPixelMask contains the x % 7 bitmask
var HGRPixelMask [7]byte = [7]byte{1, 2, 4, 8, 16, 32, 64}

const (
	BLACK  = 0
	GREEN  = 1
	VIOLET = 2
	WHITE1 = 3
	BLACK1 = 4
	ORANGE = 5
	BLUE   = 6
	WHITE  = 7
)

var HGRHMask [7]byte = [7]byte{0x81, 0x82, 0x84, 0x88, 0x90, 0xA0, 0xC0}

type HGRControllable interface {
	ColorAt(x, y int) int
	ColorsForScanLine(b []uint64, mono bool) []int
	Fill(c int)
	Plot(x, y, c int)
	Shape(shape ShapeEntry, x int, y int, scl int, deg float64, c int, usecol bool)
	XYToOffset(x, y int) int
	GetCollisionCount() uint64
	SetCollisionCount(v uint64)
	Clear(fv uint64)
	HgrSpritePlot(x int, y int, col int)
	GetLastXY() (int, int)
	OffsetToScanline(offset int) int
	MixedMono() bool
}

func b(s string) byte {

	var b byte

	for _, ch := range s {
		b = b << 1
		if ch == '1' {
			b = b | 1
		}
	}

	return b
}

func init() {
	//	////fmt.Println(b("11111111"))
}

var HGRColorMasks [2][8]byte = [2][8]byte{
	[8]byte{
		b("00000000"),
		b("00101010"),
		b("01010101"),
		b("01111111"),
		b("10000000"),
		b("10101010"),
		b("11010101"),
		b("11111111"),
	},
	[8]byte{
		b("00000000"),
		b("01010101"),
		b("00101010"),
		b("01111111"),
		b("10000000"),
		b("11010101"),
		b("10101010"),
		b("11111111"),
	},
}

var HGRMASK int

var HGRColorHint [2][2][2]int = [2][2][2]int{
	[2][2]int{
		[2]int{VIOLET, GREEN},
		[2]int{BLUE, ORANGE},
	},
	[2][2]int{
		[2]int{GREEN, VIOLET},
		[2]int{ORANGE, BLUE},
	},
}

// HGRScreen defines the screen memory
type HGRScreen struct {
	Data           *memory.MemoryControlBlock
	SLUpdated      [192]int64
	Spritedata     map[int]int
	CollisionCount int
	LASTX, LASTY   int
}

func NewHGRScreen(data *memory.MemoryControlBlock) *HGRScreen {
	if data.Size < 8192 {
		panic(fmt.Sprintf("bad hgr buffer - got %d bytes", data.Size))
	}
	this := &HGRScreen{}
	this.Data = data
	return this
}

func (hgr *HGRScreen) MixedMono() bool {
	return false
}

func (hgr *HGRScreen) Clear(fv uint64) {
	for i := 0; i < hgr.Data.Size; i++ {
		hgr.Data.Write(i, fv)
	}
}

func (hgr *HGRScreen) Poke(addr int, value uint64, text bool) {

	scanLine := addr / 1024
	rem := addr % 1024

	//offset := (textLineOfThird * 128) + (40 * thirdOfScreen)
	textLineOfThird := rem / 128
	rem = rem % 128

	thirdOfScreen := rem / 40

	if thirdOfScreen > 2 {
		thirdOfScreen = 2
	}

	//xbyte := addr - ((textLineOfThird * 128) + (40 * thirdOfScreen) + (1024 * scanLine))

	y := (thirdOfScreen * 64) + (8 * textLineOfThird) + scanLine

	if !text {
		hgr.SLUpdated[y] = time.Now().UnixNano()
	}
	hgr.Data.Write(addr, value)

	//////fmt.Printf("Marking update for scanline %d\n", y)

}

func (hgr *HGRScreen) GetScanLineChange() ([]uint64, int) {

	now := time.Now().UnixNano()

	for i, b := range hgr.SLUpdated {
		// look for scanline > 1ms old (more likely completed)
		if (b > 0) && (now-b > 5000000) {
			offset := hgr.XYToOffset(0, i)
			hgr.SLUpdated[i] = 0
			return hgr.Data.ReadSlice(offset, offset+40), i
		}
	}

	return []uint64(nil), -1
}

func (hgr *HGRScreen) Plot(x, y, c int) {
	hgr.plot(x, y, c)
	if DoubleWidth && x < 279 {
		hgr.plot(x+1, y, c)
	}
}

func (hgr *HGRScreen) plot(x, y, c int) {

	if x < 0 || x > 279 || y < 0 || y > 191 {
		return
	}

	offs := hgr.XYToOffset(x, y) + (x / 7)

	if offs < 0 || offs >= hgr.Data.Size {
		return
	}

	b := hgr.Data.Read(offs)

	hndx := x / 7                                // 1
	hmask := uint64(HGRHMask[x%7])               // 0x81
	hcolor := uint64(HGRColorMasks[hndx%2][c%8]) // 0xff

	if useMemColor {
		hcolor = uint64(HGRMASK) // assume this is kept updated from memory

		if hndx%2 == 1 {
			if (hcolor&127 != 0) && (hcolor&127 != 127) {
				hcolor = hcolor ^ 127
			}
		}
	}

	a := hcolor
	a = a ^ b     // a = 11111111
	a = a & hmask // a = 10000001

	a = a ^ b // a = 10000001

	hgr.Data.Write(offs, a)

}

func (hgr *HGRScreen) GetLastXY() (int, int) {
	return hgr.LASTX, hgr.LASTY
}

func (hgr *HGRScreen) Fill(c int) {
	for y := 0; y < 192; y++ {
		for x := 0; x < 280; x++ {
			hgr.Plot(x, y, c)
		}
	}
}

func (hgr *HGRScreen) HLine(x0, y0, x1, y1, c int) {

	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	var sx, sy int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	for {
		hgr.Plot(x0, y0, c)
		//if c & 3 == 3 {
		hgr.Plot(x0+1, y0, c)
		//}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}

}

// YToRowIndex converts a y index to memory offset
func (hgr *HGRScreen) XYToOffset(x, y int) int {

	thirdOfScreen := y / 64         // 0,1,2
	textLine := y / 8               // 0 - 23
	textLineOfThird := textLine % 8 // 0 - 7
	scanLine := y % 8               // 0 - 7

	offset := (textLineOfThird * 128) + (40 * thirdOfScreen) + (1024 * scanLine)

	return offset
}

func (hgr *HGRScreen) OffsetToScanline(offset int) int {

	thirdOfScreen := (offset % 128) / 40
	scanLine := offset / 1024
	textLineOfThird := (offset % 1024) / 128

	return 64*thirdOfScreen + 8*textLineOfThird + scanLine
}

// ColorAt determines what PIXEL color would be shown based on the bitplanes
func (hgr *HGRScreen) ColorAt(x, y int) int {
	offs := hgr.XYToOffset(0, y)

	b := hgr.Data.ReadSlice(offs, offs+40) // yay we have the scanline bytes

	sline := hgr.ColorsForScanLine(b, false)

	return sline[x]
}

// ColorAt determines what PIXEL color would be shown based on the bitplanes
func (hgr *HGRScreen) PixelAt(x, y int) bool {
	offs := hgr.XYToOffset(x, y) + (x / 7)

	value := hgr.Data.Read(offs)
	bit := uint64(1) << uint64(x%7)

	return ((value & bit) == bit)
}

func (hgr *HGRScreen) SetPixelAt(x, y int) {
	offs := hgr.XYToOffset(x, y) + (x / 7)

	value := hgr.Data.Read(offs)

	bit := uint64(1) << uint64(x%7)

	value = value | bit
	hgr.Data.Write(offs, value)
}

func (hgr *HGRScreen) ClearPixelAt(x, y int) {
	offs := hgr.XYToOffset(x, y) + (x / 7)

	value := hgr.Data.Read(offs)

	bit := uint64(1) << uint64(x%7)

	value = value & (255 - bit)
	hgr.Data.Write(offs, value)
}

// Woz you bastard ;)
func (hgr *HGRScreen) ColorsForScanLine(b []uint64, isMono bool) []int {

	line := make([]int, 280)
	mono := make([]int, 280)
	repline := make([]int, 280)

	// each byte of line
	var x int
	//var lastBit uint64
	for j, abyte := range b {

		which := int((abyte & 0x80) >> 7)

		for k := 0; k < 7; k++ {
			x = j*7 + k
			abit := abyte & 0x01

			line[x] = HGRColorHint[j%2][which][k%2] * int(abit)

			if x%14 == 6 {
				nextBit := b[j+1] & 1
				if line[x] == VIOLET && int((b[j+1]&0x80)>>7) != 0 && nextBit == 0 {
					line[x] = BLUE
				} // else if line[x] == BLUE && int((b[j+1]&0x80)>>7) == 0 && nextBit == 0 {
				// 	line[x] = VIOLET
				// }
			}

			abyte = abyte >> 1
			if abit != 0 {
				mono[j*7+k] = 15
			}

			//lastBit = abit
		}

	}

	//index := hgr.Data.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE

	if isMono {
		return mono
	}

	// // now reprocess
	for k := 0; k < 280; k++ {
		if k == 0 {
			if mono[k] == 15 && mono[k+1] == 15 {
				repline[k] = WHITE
			} else {
				repline[k] = line[k]
			}
		} else if k == 279 {
			if mono[k] == 15 && mono[k-1] == 15 {
				repline[k] = WHITE
			} else {
				repline[k] = line[k]
			}
		} else {
			//log2.Printf("mono[%d] = %d", k, mono[k])
			if mono[k] == 15 && (mono[k-1] == 15 || mono[k+1] == 15) {
				repline[k] = WHITE
			} else if mono[k] == 0 && mono[k-1] == mono[k+1] {
				repline[k] = line[k-1]
			} else {
				repline[k] = line[k]
			}
		}
	}

	// return repline
	return repline

}

// Woz you bastard ;)
func (hgr *HGRScreen) ColorsForScanLineNew(b []uint64, isMono bool) []int {

	repline := make([]int, 280)
	line := make([]int, 280)
	mono := make([]int, 280)

	var pal = [2][2][4]int{
		[2][4]int{
			[4]int{BLACK, VIOLET, GREEN, WHITE},
			[4]int{BLACK, BLUE, ORANGE, WHITE},
		},
		[2][4]int{
			[4]int{BLACK, GREEN, VIOLET, WHITE},
			[4]int{BLACK, ORANGE, BLUE, WHITE},
		},
	}

	var chunk int
	var x, c int // pixel (zero to 140)
	var mask int
	var palette int
	var hds1, hds2 bool
	for i := 0; i < 20; i++ {
		// 14 bit chunk of pixels
		chunk = (int(b[i*2+1]&0x7f) << 7) | int(b[i*2]&0x7f)
		// hds states
		hds1 = (b[i*2]&128 != 0)
		hds2 = (b[i*2+1]&128 != 0)
		// produce the 7 pixels
		for p := 0; p < 14; p++ {
			mask = ((mask << 1) | (chunk & 1)) & 3 // 2 pixels
			// handle mono output
			if isMono {
				if mask&1 != 0 {
					mono[x] = 15
				}
			} else {
				// handle color output
				if p == 0 {
					palette = 0
					if hds1 {
						palette = 1
					}
				} else if p == 7 {
					palette = 0
					if hds2 {
						palette = 1
					}
				}
				// got palette and mask
				c = pal[x%2][palette][mask]
				line[x] = c
			}
			// move over 1 pixel
			chunk >>= 1
			x++
		}
	}

	if isMono {
		return mono
	}

	for k := 0; k < 280; k++ {
		if k == 0 {
			if mono[k] == 15 && mono[k+1] == 15 {
				repline[k] = WHITE
			} else {
				repline[k] = line[k]
			}
		} else if k == 279 {
			if mono[k] == 15 && mono[k-1] == 15 {
				repline[k] = WHITE
			} else {
				repline[k] = line[k]
			}
		} else {
			//log2.Printf("mono[%d] = %d", k, mono[k])
			if mono[k] == 15 && (mono[k-1] == 15 || mono[k+1] == 15) {
				repline[k] = WHITE
			} else if mono[k] == 0 && mono[k-1] == mono[k+1] {
				repline[k] = line[k-1]
			} else {
				repline[k] = line[k]
			}
		}
	}

	// return repline
	return repline

}

func (this *HGRScreen) HgrSpritePlot(x int, y int, col int) {
	idx := (y * 280) + x

	if (x < 0) || (x > 279) || (y < 0) || (y > 191) {
		return
	}

	if (idx < 280*192) && (idx >= 0) {
		this.Spritedata[idx] = col
	}
	//this.LASTX = x
	//	this.LASTY = y
}

func (this *HGRScreen) GetCollisionCount() uint64 {
	return uint64(this.CollisionCount)
}

func (this *HGRScreen) SetCollisionCount(v uint64) {
	this.CollisionCount = int(v)
}

func Round(v float64) float64 {
	return math.Floor(v + 0.5)
}

func (this *HGRScreen) Shape(shape ShapeEntry, x int, y int, scl int, deg float64, c int, usecol bool) {

	/* vars */
	//	var z int
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
	//	var lastdraw bool
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

	/*switch scl {
	case 1:
		deg = (deg / 45) * 45
		break
	case 2:
		deg = int((float64(deg) / 22.5) * 22.5)
		break
	case 3:
		deg = int((float64(deg) / 22.5) * 22.5)
		break
	case 4:
		deg = int((float64(deg) / 11.25) * 11.25)
		break
	case 5:
		deg = int((float64(deg) / 11.25) * 11.25)
		break
	case 6:
		deg = int((float64(deg) / 9) * 9)
		break
	case 7:
		deg = int((float64(deg) / 7.5) * 7.5)
		break
	case 8:
		deg = int((float64(deg) / 6.42857) * 6.42857)
		break
	default:
		deg = int((float64(deg) / 5.625) * 5.625)
		break
	}*/

	r = math.Pi * deg / 180

	for i < len(shape) {
		b = shape[i]
		if b == 0 {
			for k, _ := range this.Spritedata {
				xx := k % 280
				yy := k / 280

				//				oc := pp.ColorAt(xx, yy)
				if usecol {
					c = this.Spritedata[k]
					before := pp.PixelAt(xx, yy)
					pp.Plot(xx, yy, c)
					after := pp.PixelAt(xx, yy)
					if before == after {
						this.CollisionCount++
					}
				} else {
					oc := pp.ColorAt(xx, yy)
					c = 7 - oc
					if !pp.PixelAt(xx, yy) {
						this.CollisionCount++
						this.SetPixelAt(xx, yy)
					} else {
						this.ClearPixelAt(xx, yy)
					}

				}
			}
			this.CollisionCount = this.CollisionCount % 256
			return
		}

		// display byte
		//fmt.Println("NEW Shape byte value = ",b,"at",i);

		s = 0
		for (s < 3) && (b > 0) {
			v = (b & 7)

			if (s == 2) && (v == 0) {
				//fmt.Println("!!!!! Skip s3/v0 -- NO MOVE NO PLOT");
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

				//fmt.Printf("scale = %d\n", scl)

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
		//this.LASTX = int(ox)
		//this.LASTY = int(oy)
	}

	// draw sprite buffer
}
