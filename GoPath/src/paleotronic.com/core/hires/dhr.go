package hires

/*

The memory mapping for the hi-res graphics display is analogous to the
technique used for the 80-column display.  The double hi-res display
interleaves bytes from the two different memory pages (auxiliary and
motherboard).  Seven bits from a byte in the auxiliary memory bank are
displayed first, followed by seven bits from the corresponding byte on the
motherboard.  The bits are shifted out the same way as in standard hi-res
(least-significant bit first).  In double hi-res, the most significant bit of
each byte is ignored; thus, no half-dot shift can occur.  (This feature is
important, as you will see when we examine double hi-res in color.)

		Page	AUX			MAIN
		1		$2000		$2000
		2		$4000		$4000

                        Horizontal Offset
       $00   $01   $02   $03         $24   $25   $26   $27
         M     M     M     M           M     M     M     M
      A  a  A  a  A  a  A  a        A  a  A  a  A  a  A  a
     |u  i |u  i |u  i |u  i        u  i |u  i |u  i |u  i |
Base |x |n |x |n |x |n |x |n        x |n |x |n |x |n |x |n |
_____|__|__|__|__|__|__|__|__       __|__|__|__|__|__|__|__|
$2000|     |     |     |     |     |     |     |     |     |
$2080|     |     |     |     |     |     |     |     |     |
$2100|     |     |     |     |     |     |     |     |     |
$2180|     |     |     |     |     |     |     |     |     |
$2200|     |     |     |     |     |     |     |     |     |
$2280|     |     |     |     |     |     |     |     |     |
$2300|     |     |     |     |     |     |     |     |     |
$2380|     |     |     |     |     |     |     |     |     |
$2028|     |     |     |      \     \    |     |     |     |
$20A8|     |     |     |       \     \   |     |     |     |
$2128|     |     |     |       /     /   |     |     |     |
$21A8|     |     |     |      /     /    |     |     |     |
$2228|     |     |     |     /     /     |     |     |     |
$22A8|     |     |     |    /     /      |     |     |     |
$2328|     |     |     |   /     /       |     |     |     |
$23A8|     |     |     |   \     \       |     |     |     |
$2050|     |     |     |    \     \      |     |     |     |
$20D0|     |     |     |     |     |     |     |     |     |
$2150|     |     |     |     |     |     |     |     |     |
$21D0|     |     |     |     |     |     |     |     |     |
$2250|     |     |     |     |     |     |     |     |     |
$22D0|     |     |     |     |     |     |     |     |     |
$2350|     |     |     |     |     |     |     |     |     |
$23D0|     |     |     |     |     |     |     |     |     |

            Figure 3 - Double Hi-Res Memory Map
*/

import (
	"math"

	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
)

var DHGRBytePatterns = [16][4]int{
	[4]int{0x00, 0x00, 0x00, 0x00}, // Black				lores(0)
	[4]int{0x08, 0x11, 0x22, 0x44}, // Magenta
	[4]int{0x44, 0x08, 0x11, 0x22}, // Brown
	[4]int{0x4c, 0x19, 0x33, 0x66}, // Orange
	[4]int{0x22, 0x44, 0x08, 0x11}, // Dark Green
	[4]int{0x2a, 0x55, 0x2a, 0x55}, // Grey 1
	[4]int{0x66, 0x4c, 0x19, 0x33}, // Green
	[4]int{0x6E, 0x5D, 0x3B, 0x77}, // Yellow
	[4]int{0x11, 0x22, 0x44, 0x08}, // Dark blue
	[4]int{0x19, 0x33, 0x66, 0x4C}, // Violet
	[4]int{0x55, 0x2A, 0x55, 0x2A}, // Grey 2
	[4]int{0x5D, 0x3B, 0x77, 0x6E}, // Pink
	[4]int{0x33, 0x66, 0x4C, 0x19}, // Medium Blue
	[4]int{0x3B, 0x77, 0x6E, 0x5D}, // Light Blue
	[4]int{0x77, 0x6E, 0x5D, 0x3B}, // Aqua
	[4]int{0x7F, 0x7F, 0x7F, 0x7F}, // White				lores(15)
}

var DHGRPaletteToLores = map[int]int{
	0x0: 0x0,
	0x1: 0x1,
	0x8: 0x2,
	0x9: 0x3,
	0x4: 0x4,
	0x5: 0x5,
	0xc: 0x6,
	0xd: 0x7,
	0x2: 0x8,
	0x3: 0x9,
	0xa: 0xa,
	0xb: 0xb,
	0x6: 0xc,
	0x7: 0xd,
	0xe: 0xe,
	0xf: 0xf,
}

const DHR_WHITE = 15
const DHR_BLACK = 0

// HGRScreen defines the screen memory
type DHGRScreen struct {
	Data           *memory.MemoryControlBlock
	SLUpdated      [192]int64
	Spritedata     map[int]int
	CollisionCount int
	LASTX, LASTY   int
	mixed          bool
}

func NewDHGRScreen(data *memory.MemoryControlBlock) *DHGRScreen {
	// if data.Size < 16384 {
	// 	panic(fmt.Sprintf("bad dhgr buffer - got %d bytes", data.Size))
	// }
	this := &DHGRScreen{}
	this.Data = data
	return this
}

func (hgr *DHGRScreen) MixedMono() bool {
	var b1, b2 []uint64
	var count, offset1, offset2 int
	var mixed bool
	for y := 0; y < 192; y++ {
		offset1 = hgr.XYToOffset(0, y)
		offset2 = hgr.XYToOffset(7, y)
		b1 = hgr.Data.ReadSlice(offset1, offset1+40)
		b2 = hgr.Data.ReadSlice(offset2, offset2+40)
		count = 0
		for x := 0; x < 40; x++ {
			if b1[x]&0x80 != 0 {
				count++
			}
			if b2[x]&0x80 != 0 {
				count++
			}
		}
		mixed = count > 0
		if mixed {
			//log.Printf("Scanline %d: count=%d", y, count)
			hgr.mixed = true
			return true
		}
	}
	hgr.mixed = false
	return false
}

// XYToOffset converts an x,y (560,192) to an offset
func (hgr *DHGRScreen) XYToOffset(x, y int) int {

	thirdOfScreen := y / 64         // 0,1,2
	textLine := y / 8               // 0 - 23
	textLineOfThird := textLine % 8 // 0 - 7
	scanLine := y % 8               // 0 - 7

	mempage := (1 + (x / 7)) % 2 // 0 = main, 1 = aux

	offset := (textLineOfThird * 128) + (40 * thirdOfScreen) + (1024 * scanLine) + (8192 * mempage)

	return offset
}

func (hgr *DHGRScreen) OffsetToScanline(offset int) int {
	thirdOfScreen := (offset % 128) / 40
	scanLine := offset / 1024
	textLineOfThird := (offset % 1024) / 128

	return 64*thirdOfScreen + 8*textLineOfThird + scanLine
}

func (hgr *DHGRScreen) PixelAt(x, y int) bool {

	offs := hgr.XYToOffset(x, y) + (x / 14)

	value := hgr.Data.Read(offs)
	bit := uint64(1) << uint64(x%7)

	return ((value & bit) == bit)
}

func (hgr *DHGRScreen) SetPixelAt(x, y int) {
	offs := hgr.XYToOffset(x, y) + (x / 14)

	value := hgr.Data.Read(offs)

	bit := uint64(1) << uint64(x%7)

	value = value | bit
	hgr.Data.Write(offs, value)
}

func (hgr *DHGRScreen) ClearPixelAt(x, y int) {
	offs := hgr.XYToOffset(x, y) + (x / 14)

	value := hgr.Data.Read(offs)

	bit := uint64(1) << uint64(x%7)

	value = value & (255 - bit)
	hgr.Data.Write(offs, value)
}

var dhgrBitPos = []uint64{6, 2, 5, 1, 4, 0, 3}

// Put a certain colour at a certain 140x192 position
func (hgr *DHGRScreen) Plot140(px, py int, c int) {

	ax := px * 4
	bx := ax + 3

	// ax is the first x bit at 560px, bx the last x bit at 560px that make up the 140px pixel
	for x := ax; x <= bx; x++ {

		//xind := (x % 28) / 7 // 0-3
		bitpattern := DHGRBytePatterns[c%16][0]

		// ?aaaabbb-?bccccdd-?ddeeeef-?fffgggg
		//    6   2     5   1     4   0     3
		//    0   4     8   12    16  20    24
		bit := uint64(6 - ((x - ax) % 7))
		bittest := 1 << bit

		if bitpattern&bittest != 0 {
			hgr.SetPixelAt(x, py)
		} else {
			hgr.ClearPixelAt(x, py)
		}

	}

}

func (hgr *DHGRScreen) Plot560(x, y int, c int) {
	switch c % 2 {
	case 0:
		hgr.ClearPixelAt(x, y)
	case 1:
		hgr.SetPixelAt(x, y)
	}
}

func (hgr *DHGRScreen) Plot(x, y, c int) {

	if x < 0 || x > 139 || y < 0 || y > 191 {
		return
	}

	hgr.Plot140(x, y, c)

}

// Woz you bastard ;)
func (hgr *DHGRScreen) ColorsForScanLineOld(b []uint64, mono bool) []int {

	index := hgr.Data.GStart[0] / memory.OCTALYZER_INTERPRETER_SIZE

	line := make([]int, 560)

	rbval := 0
	//lastrbval := rbval

	var monobits bool

	bitcols := []int{DHR_BLACK, DHR_WHITE}

	var useColor = make([]bool, len(b))
	var mixed = (settings.DHGRHighBit[index] == settings.DHB_MIXED_AUTO && settings.DHGRMode3Detected[index]) || (settings.DHGRHighBit[index] == settings.DHB_MIXED_ON)

	for i, v := range b {
		useColor[i] = (v & 0x80) != 0
	}

	for j, abyte := range b {

		monobits = mono || (mixed && !useColor[j])

		for k := 0; k < 7; k++ {
			x := j*7 + k
			abit := int(abyte & 0x01)
			rbval = ((rbval << 1) & 0xf) | abit
			abyte = abyte >> 1
			if !mono && !monobits {
				if x%4 == 3 {
					line[x-0] = DHGRPaletteToLores[rbval]
					line[x-1] = DHGRPaletteToLores[rbval]
					line[x-2] = DHGRPaletteToLores[rbval]
					line[x-3] = DHGRPaletteToLores[rbval]
					if x+1 < 560 {
						line[x+1] = DHGRPaletteToLores[rbval]
					}
					if x+2 < 560 {
						line[x+2] = DHGRPaletteToLores[rbval]
					}
					//lastrbval = rbval
				}
				//line[x] = DHGRPaletteToLores[rbval]
				//lastrbval = rbval
			} else {
				line[x] = bitcols[abit]
				//lastrbval = rbval
			}
		}
	}

	return line

}

func (hgr *DHGRScreen) ColorAt(x, y int) int {

	x *= 4

	offs_aux := hgr.XYToOffset(0, y)
	offs_main := hgr.XYToOffset(7, y)

	aux_data := hgr.Data.ReadSlice(offs_aux, offs_aux+40) // yay we have the scanline bytes
	main_data := hgr.Data.ReadSlice(offs_main, offs_main+40)

	b := make([]uint64, 80) // we need to interleave the bytes
	for i, _ := range b {
		switch i % 2 {
		case 0:
			b[i] = aux_data[i/2]
		case 1:
			b[i] = main_data[i/2]
		}
	}

	sline := hgr.ColorsForScanLineOld(b, false)

	if x < 0 || x >= len(sline) {
		return 0
	}

	return sline[x]
}

func (hgr *DHGRScreen) Fill(c int) {
	for y := 0; y < 192; y++ {
		for x := 0; x < 140; x++ {
			hgr.Plot(x, y, c)
		}
	}
}

func (hgr *DHGRScreen) Clear(fv uint64) {
	for i := 0; i < hgr.Data.Size; i++ {
		hgr.Data.Write(i, fv)
	}
}

func (hgr *DHGRScreen) GetLastXY() (int, int) {
	return hgr.LASTX, hgr.LASTY
}

func (this *DHGRScreen) HgrSpritePlot(x int, y int, col int) {
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

func (this *DHGRScreen) GetCollisionCount() uint64 {
	return uint64(this.CollisionCount)
}

func (this *DHGRScreen) SetCollisionCount(v uint64) {
	this.CollisionCount = int(v)
}

func (this *DHGRScreen) Shape(shape ShapeEntry, x int, y int, scl int, deg float64, c int, usecol bool) {

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
				xx := k % 560
				yy := k / 560

				//				oc := pp.ColorAt(xx, yy)
				if usecol {
					c = this.Spritedata[k]
					pp.Plot(xx, yy, c)
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
				//bm[k] = c;
				//bm.Plot(xx, yy, c)
				//pp.Plot(xx, yy, c)

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
		//this.LASTX = int(ox)
		//this.LASTY = int(oy)
	}

	// draw sprite buffer
}
