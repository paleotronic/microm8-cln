package hires

//"paleotronic.com/core/dialect"
//"paleotronic.com/core/interfaces"
//"paleotronic.com/core/types"
import "paleotronic.com/core/memory"
import "math"

const (
	BitsPerPixel int = 8
	ByteSize     int = 32
)

var (
	OffsetMask = [4]uint64{
		0xff000000,
		0x00ff0000,
		0x0000ff00,
		0x000000ff,
	}
	OffsetMaskClear = [4]uint64{
		0x00ffffff,
		0xff00ffff,
		0xffff00ff,
		0xffffff00,
	}
	OffsetRotate = [4]uint64{
		24,
		16,
		8,
		0,
	}
)

type IndexedVideoBuffer struct {
	duplicate      map[int]int
	data           *memory.MemoryControlBlock
	width          int
	height         int
	txns           map[int]uint64
	Spritedata     map[int]int
	CollisionCount int
	LASTX, LASTY   int
}

func (hgr *IndexedVideoBuffer) GetLastXY() (int, int) {
	return hgr.LASTX, hgr.LASTY
}

func (hgr *IndexedVideoBuffer) OffsetToScanline(chunkaddr int) int {
	return (chunkaddr * (ByteSize / BitsPerPixel)) / hgr.width
}

func (this *IndexedVideoBuffer) Plot(x int, y int, c int) {
	if (x < 0) || (x >= this.width) || (y < 0) || (y >= this.height) {
		return
	}

	c = c & 0xff
	chunkaddr := ((y * this.width) + x) / (ByteSize / BitsPerPixel)
	chunk := this.data.Read(chunkaddr)
	byteoffset := ((y * this.width) + x) % (ByteSize / BitsPerPixel)

	chunk = (chunk & OffsetMaskClear[byteoffset]) | (uint64(c) << OffsetRotate[byteoffset])
	this.data.Write(chunkaddr, chunk)

	this.LogPixel(chunkaddr, chunk, x, y, c)

	//this.LASTX, this.LASTY = x, y
}

func (this *IndexedVideoBuffer) GetCollisionCount() uint64 {
	return uint64(this.CollisionCount)
}

func (this *IndexedVideoBuffer) SetCollisionCount(v uint64) {
	this.CollisionCount = int(v)
}

func (this *IndexedVideoBuffer) GetHeight() int {
	return this.height
}

func (this *IndexedVideoBuffer) UnpackPixels(chunk uint64) []uint64 {
	r := make([]uint64, 4)
	for i := 0; i < 4; i++ {
		r[i] = (chunk & OffsetMask[i]) >> OffsetRotate[i]
	}
	return r
}

func (this *IndexedVideoBuffer) Repaint() {
	//synchronized(this.Txns) {
	this.txns = make(map[int]uint64)
	this.txns[-2] = 0 // repaint signal
	//}
}

func (this *IndexedVideoBuffer) MixedMono() bool {
	return false
}

func (this *IndexedVideoBuffer) GetData() []uint64 {
	return this.data.ReadSlice(0, this.data.Size)
}

func NewIndexedVideoBuffer(width int, height int, data *memory.MemoryControlBlock) *IndexedVideoBuffer {
	this := &IndexedVideoBuffer{}
	this.width = width
	this.height = height
	this.data = data
	this.txns = make(map[int]uint64)
	return this
}

func (this *IndexedVideoBuffer) ClearTransactions() {
	//synchronized(this.Txns) {
	this.txns = make(map[int]uint64)
	//}
}

func (this *IndexedVideoBuffer) GetWidth() int {
	return this.width
}

func (this *IndexedVideoBuffer) CheckDuplicate(x int, y int, c int) {
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

func (this *IndexedVideoBuffer) Clear(fv uint64) {
	for i := 0; i < this.data.Size; i++ {
		this.data.Write(i, fv)
	}
}

func (this *IndexedVideoBuffer) Fill(c int) {

	c = c & 0xff

	chunk := (c << 24) | (c << 16) | (c << 8) | c
	for i := 0; i < this.data.Size; i++ {
		this.data.Write(i, uint64(chunk))
	}

	this.ClearTransactions()
	this.LogPixel(-1, uint64(chunk), -1, -1, c)
}

func (this *IndexedVideoBuffer) HasTransactions() bool {
	return (len(this.txns) != 0)
}

func (this *IndexedVideoBuffer) GetTransactions() map[int]uint64 {
	//HashMap<Integer,Integer> old
	old := make(map[int]uint64)

	//synchronized (txns) {
	old = this.txns
	this.txns = make(map[int]uint64)
	//}

	return old
}

func (this *IndexedVideoBuffer) ColorAt(x int, y int) int {
	if (x < 0) || (x >= this.width) || (y < 0) || (y >= this.height) {
		return 0
	}

	chunkaddr := ((y * this.width) + x) / (ByteSize / BitsPerPixel)
	chunk := this.data.Read(chunkaddr)
	byteoffset := ((y * this.width) + x) % (ByteSize / BitsPerPixel)

	chunk = (chunk & OffsetMask[byteoffset]) >> OffsetRotate[byteoffset]
	return int(chunk)
}

func (this *IndexedVideoBuffer) LogPixel(chunkaddr int, chunk uint64, x int, y int, c int) {

	// pause rendering if needed for dup pixels
	//	if x != -1 {
	//		this.CheckDuplicate(x, y, c)
	//	}
	//	//synchronized(this.Txns) {
	//	this.txns[chunkaddr] = chunk
	//}
}

func (this *IndexedVideoBuffer) HgrSpritePlot(x int, y int, col int) {
	idx := (y * 280) + x

	if (x < 0) || (x > 279) || (y < 0) || (y > 191) {
		return
	}

	if (idx < 280*192) && (idx >= 0) {
		this.Spritedata[idx] = col
	}

}

func (this *IndexedVideoBuffer) Shape(shape ShapeEntry, x int, y int, scl int, deg float64, c int, usecol bool) {

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

	/*
		switch scl {
		case 1:
			deg = (deg / 45) * 45                      // 8
			break
		case 2:
			deg = int((float64(deg) / 22.5) * 22.5)    // 12
			break
		case 3:
			deg = int((float64(deg) / 22.5) * 22.5)    // 28
			break
		case 4:
			deg = int((float64(deg) / 11.25) * 11.25)  // 48
			break
		case 5:
			deg = int((float64(deg) / 11.25) * 11.25)  //
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
		}
	*/
	//	var lastCode = -1

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

func (hgr *IndexedVideoBuffer) ColorsForScanLine(b []uint64, mono bool) []int {

	repline := make([]int, 280)

	for i, _ := range repline {
		repline[i] = int((b[i/4] & OffsetMask[i%4]) >> OffsetRotate[i%4])
	}

	return repline

}

// YToRowIndex converts a y index to memory offset
func (hgr *IndexedVideoBuffer) XYToOffset(x, y int) int {

	return 0
}
