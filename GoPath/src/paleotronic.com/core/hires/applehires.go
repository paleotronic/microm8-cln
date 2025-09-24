package hires

import (
	"math"
)

const (
	ST_UP_WITHOUT_PLOT = 0
	ST_RT_WITHOUT_PLOT = 1
	ST_DN_WITHOUT_PLOT = 2
	ST_LT_WITHOUT_PLOT = 3
	ST_UP_WITH_PLOT    = 4
	ST_RT_WITH_PLOT    = 5
	ST_DN_WITH_PLOT    = 6
	ST_LT_WITH_PLOT    = 7
)

type AppleHiRES struct {
	PX1            int
	PY0            int
	LASTY          int
	IGNORELASTPLOT bool
	Spritedata     map[int]int
	CollisionCount int
	LASTX          int
	PX0            int
	PY1            int
}

type Plotter interface {
	HgrPlotHold(x, y int, c int)
}

var hires *AppleHiRES

func GetAppleHiRES() *AppleHiRES {
	if hires == nil {
		hires = &AppleHiRES{}
		hires.Spritedata = make(map[int]int)
	}
	return hires
}

func (this *AppleHiRES) GetCollisionCount() uint {
	 return uint( this.CollisionCount )
}

func (this *AppleHiRES) SetCollisionCount(v uint) {
	 this.CollisionCount = int(v)
}


func (this *AppleHiRES) HgrSpritePlot(x int, y int, col int) {
	idx := (y * 280) + x

	if (x < 0) || (x > 279) || (y < 0) || (y > 191) {
		return
	}

	if (idx < 280*192) && (idx >= 0) {
		this.Spritedata[idx] = col
	}
	this.LASTX = x
	this.LASTY = y
}

func (this *AppleHiRES) HgrPlot(bm *IndexedVideoBuffer, x int, y int, col int) {
	bm.Plot(x, y, col)

	this.LASTX = x
	this.LASTY = y
}

func (this *AppleHiRES) HgrScreen(bm *IndexedVideoBuffer, x int, y int) int {
	return bm.ColorAt(x, y)
}

func (this *AppleHiRES) HgrLine(bm *IndexedVideoBuffer, x0 int, y0 int, x1 int, y1 int, col int) {

	/* vars */
	var x int
	var delta_x int
	var step_x int
	var y int
	var delta_y int
	var step_y int
	var z int
	var t int
	var delta_z int
	var step_z int
	var swap_xy bool
	var swap_xz bool
	var drift_xy int
	var drift_xz int
	var cx int
	var cy int
	var cz int
	var c int

	var z0 = 0
	var z1 = 0

	//start && end points (change these values);
	//x0 = 0;     x1 = -2;
	//y0 = 0;     y1 = 5;
	//z0 = 0;     z1 = -10;

	//"steep" xy Line, make longest delta x plane  ;
	swap_xy = math.Abs(float64(y1-y0)) > math.Abs(float64(x1-x0))
	if swap_xy {
		//Swap(x0, y0);
		t = x0
		x0 = y0
		y0 = t
		//Swap(x1, y1);
		t = x1
		x1 = y1
		y1 = t
	}

	//do same for xz;
	swap_xz = math.Abs(float64(z1-z0)) > math.Abs(float64(x1-x0))
	if swap_xz {
		//Swap(x0, z0);
		t = x0
		x0 = z0
		z0 = t

		//Swap(x1, z1);
		t = x1
		x1 = z1
		z1 = t
	}

	//delta is Length in each plane;
	delta_x = int(math.Abs(float64(x1 - x0)))
	delta_y = int(math.Abs(float64(y1 - y0)))
	delta_z = int(math.Abs(float64(z1 - z0)))

	//drift controls when to step in "shallow" planes;
	//starting value keeps Line centred;
	drift_xy = (int)(delta_x / 2)
	drift_xz = (int)(delta_x / 2)

	//direction of line;
	step_x = 1
	if x0 > x1 {
		step_x = -1
	}
	step_y = 1
	if y0 > y1 {
		step_y = -1
	}
	step_z = 1
	if z0 > z1 {
		step_z = -1
	}

	//starting point;
	y = y0
	z = z0

	//step through longest delta (which we have swapped to x);
	x = x0
	for c = 0; c <= int(math.Abs(float64(x1-x0))); c++ {

		//copy position;
		cx = x
		cy = y
		cz = z

		//this.Plot3(x, y, z, col);
		//AppleHiRES.HgrPlot(bm, x, y, col);

		//unswap (in reverse);
		if swap_xz {
			//Swap(cx, cz);
			t = cx
			cx = cz
			cz = t
		}
		if swap_xy {
			//Swap(cx, cy);
			t = cx
			cx = cy
			cy = t
		}

		if col < 0 {
			col = 7 ^ this.HgrScreen(bm, cx, cy)
		}
		this.HgrPlot(bm, cx, cy, col)

		//update progress in other planes;
		drift_xy = drift_xy - delta_y
		drift_xz = drift_xz - delta_z

		//step in y plane;
		if drift_xy < 0 {
			y = y + step_y
			drift_xy = drift_xy + delta_x
		}

		//same in z;
		if drift_xz < 0 {
			z = z + step_z
			drift_xz = drift_xz + delta_x
		}

		x = x + step_x
	}

}

func (this *AppleHiRES) HgrLineToSprite(bm *IndexedVideoBuffer, x0 int, y0 int, x1 int, y1 int, col int) {

	/* vars */
	var x int
	var delta_x int
	var step_x int
	var y int
	var delta_y int
	var step_y int
	var z int
	var t int
	var delta_z int
	var step_z int
	var swap_xy bool
	var swap_xz bool
	var drift_xy int
	var drift_xz int
	var cx int
	var cy int
	var cz int
	var c int

	var z0 = 0
	var z1 = 0

	//"steep" xy Line, make longest delta x plane  ;
	swap_xy = math.Abs(float64(y1-y0)) > math.Abs(float64(x1-x0))
	if swap_xy {
		//Swap(x0, y0);
		t = x0
		x0 = y0
		y0 = t
		//Swap(x1, y1);
		t = x1
		x1 = y1
		y1 = t
	}

	//do same for xz;
	swap_xz = math.Abs(float64(z1-z0)) > math.Abs(float64(x1-x0))
	if swap_xz {
		//Swap(x0, z0);
		t = x0
		x0 = z0
		z0 = t

		//Swap(x1, z1);
		t = x1
		x1 = z1
		z1 = t
	}

	//delta is Length in each plane;
	delta_x = int(math.Abs(float64(x1 - x0)))
	delta_y = int(math.Abs(float64(y1 - y0)))
	delta_z = int(math.Abs(float64(z1 - z0)))

	//drift controls when to step in "shallow" planes;
	//starting value keeps Line centred;
	drift_xy = int(delta_x / 2)
	drift_xz = int(delta_x / 2)

	//direction of line;
	step_x = 1
	if x0 > x1 {
		step_x = -1
	}
	step_y = 1
	if y0 > y1 {
		step_y = -1
	}
	step_z = 1
	if z0 > z1 {
		step_z = -1
	}

	//starting point;
	y = y0
	z = z0

	//step through longest delta (which we have swapped to x);
	x = x0
	for c = 0; c <= int(math.Abs(float64(x1-x0))); c++ {

		//copy position;
		cx = x
		cy = y
		cz = z

		//this.Plot3(x, y, z, col);
		//AppleHiRES.HgrPlot(bm, x, y, col);

		//unswap (in reverse);
		if swap_xz {
			//Swap(cx, cz);
			t = cx
			cx = cz
			cz = t
		}
		if swap_xy {
			//Swap(cx, cy);
			t = cx
			cx = cy
			cy = t
		}

		if col < 0 {
			col = 7 ^ this.HgrScreen(bm, cx, cy)
		}
		//AppleHiRES.HgrPlot(bm, cx, cy, col);
		this.HgrSpritePlot(cx, cy, col)

		//update progress in other planes;
		drift_xy = drift_xy - delta_y
		drift_xz = drift_xz - delta_z

		//step in y plane;
		if drift_xy < 0 {
			y = y + step_y
			drift_xy = drift_xy + delta_x
		}

		//same in z;
		if drift_xz < 0 {
			z = z + step_z
			drift_xz = drift_xz + delta_x
		}

		x = x + step_x
	}

}

func (this *AppleHiRES) HgrShape(pp Plotter, bm *IndexedVideoBuffer, shape ShapeEntry, x int, y int, scl int, deg int, c int, usecol bool) {

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
	//	var x1 int
	//	var y1 int
	//	var x2 int
	//	var y2 int
	//	var ex int
	//	var ey int
	var b int
	var v int
	var draw bool
	//	var lastdraw bool
	var r float64

	//	var LastX, LastY int

	//	lastdraw = false

	this.Spritedata = make(map[int]int)

	i = 0
	//	LastX = x
	//	LastY = y

	ox = float64(x)
	oy = float64(y)

	px = 0
	py = 0

	/// zero collisions
	this.CollisionCount = 0

	//System.Err.Println("======= Shape Start");

	//
	switch scl {
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
	}

	//	var lastCode = -1

	r = math.Pi * (float64(deg) / float64(180))

	//writeln("DRAWSHAPE:");

	for i < len(shape) {
		b = shape[i]
		if b == 0 {
			for k, _ := range this.Spritedata {
				xx := k % bm.GetWidth()
				yy := k / bm.GetWidth()

				oc := bm.ColorAt(xx, yy)
				if usecol {
					c = this.Spritedata[k]
					if (oc != 0) && (c != 0) {
						this.CollisionCount++
					}
				} else {
					c = 7 - bm.ColorAt(xx, yy)
					if c == 7 {
						this.CollisionCount++
					}
				}
				//bm[k] = c;
				//bm.Plot(xx, yy, c)
				pp.HgrPlotHold(xx, yy, c)

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

				var angle = 0

				//System.Err.Println( "-----> Shape instruction code = "+ST_DESC[v]+"("+v+")" );

				switch v {
				case ST_UP_WITHOUT_PLOT:
					angle = (deg % 360)
					break
				case ST_RT_WITHOUT_PLOT:
					angle = ((deg + 90) % 360)
					break
				case ST_DN_WITHOUT_PLOT:
					angle = ((deg + 180) % 360)
					break
				case ST_LT_WITHOUT_PLOT:
					angle = ((deg + 270) % 360)
					break
				case ST_UP_WITH_PLOT:
					{
						angle = (deg % 360)
						draw = true
						break
					}
				case ST_RT_WITH_PLOT:
					{
						angle = ((deg + 90) % 360)
						draw = true
						break
					}
				case ST_DN_WITH_PLOT:
					{
						angle = ((deg + 180) % 360)
						draw = true
						break
					}
				case ST_LT_WITH_PLOT:
					{
						angle = ((deg + 270) % 360)
						draw = true
						break
					}
				}

				r = math.Pi * (float64(angle) / float64(180))

				/// ox, oy last point so starting point

				//boolean ignore = (v == 0) && (lastCode == 0);

				//if (!ignore) {
				for ss := 0; ss < scl; ss++ {
					if draw {
						// coll test
						if usecol {

							this.HgrSpritePlot(int(math.Ceil(ox)), int(math.Ceil(oy)), c)
						} else {
							this.HgrSpritePlot(int(math.Ceil(ox)), int(math.Ceil(oy)), 0-c)
						}
					}

					this.LASTX = int(ox)
					this.LASTY = int(oy)

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

func (this *AppleHiRES) HgrFill(bm *IndexedVideoBuffer, col int) {
	bm.Fill(col)
}
