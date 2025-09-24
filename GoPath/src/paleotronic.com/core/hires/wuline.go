package hires

import (
	"math"
)

// drop in for line drawing algorithm

type Plotable interface {
	Plot(x, y int, c int)
	ColorAt(x, y int) int
}

type Plotable3 interface {
	Plot(x, y, z int, c int)
	ColorAt(x, y, z int) int
}

func ipart(x float64) int {
	return int(x)
}

func round(x float64) int {
	return ipart(x + 0.5)
}

func fpart(x float64) float64 {
	if x < 0 {
		return 1 - (x - math.Floor(x))
	}
	return x - math.Floor(x)
}

func rfpart(x float64) float64 {
	return 1 - fpart(x)
}

func swap(a, b *float64) {
	t := *a
	*a = *b
	*b = t
}

func WuLine(x0, y0, x1, y1 float64, c int, p HGRControllable) {
	steep := math.Abs(y1-y0) > math.Abs(x1-x0)

	if steep {
		swap(&x0, &y0)
		swap(&x1, &y1)
	}

	if x0 > x1 {
		swap(&x0, &x1)
		swap(&y0, &y1)
	}

	dx := x1 - x0
	dy := y1 - y0
	gradient := dy / dx

	// handle first }point
	xend := round(x0)
	yend := y0 + gradient*(float64(xend)-x0)
	//xgap := rfpart(x0 + 0.5)
	xpxl1 := xend // this will be used in the main loop
	ypxl1 := ipart(yend)
	if steep {
		p.Plot(ypxl1, xpxl1, c)
		p.Plot(ypxl1+1, xpxl1, c)
	} else {
		p.Plot(xpxl1, ypxl1, c)
		p.Plot(xpxl1, ypxl1+1, c)
	}
	intery := yend + gradient // first y-intersection for the main loop

	// handle second }point
	xend = round(x1)
	yend = y1 + gradient*(float64(xend)-x1)
	//	xgap = fpart(x1 + 0.5)
	xpxl2 := xend //this will be used in the main loop
	ypxl2 := ipart(yend)
	if steep {
		p.Plot(ypxl2, xpxl2, c)
		p.Plot(ypxl2+1, xpxl2, c)
	} else {
		p.Plot(xpxl2, ypxl2, c)
		p.Plot(xpxl2, ypxl2+1, c)
	}

	// main loop
	for x := xpxl1 + 1; x < xpxl2-1; x++ {
		if steep {
			p.Plot(ipart(intery), x, c)
			p.Plot(ipart(intery)+1, x, c)
		} else {
			p.Plot(x, ipart(intery), c)
			p.Plot(x, ipart(intery)+1, c)
		}
		intery = intery + gradient
	}
}


func BrenshamLine(x0, y0, x1, y1 int, p int, pp HGRControllable) {
    // implemented straight from WP pseudocode
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
        pp.Plot(x0, y0, p)
        //if p & 3 == 3 {
           pp.Plot(x0+1, y0, p)
       // }
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

func BrenshamLineSprite(x0, y0, x1, y1 int, p int, pp HGRControllable) {
    // implemented straight from WP pseudocode
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
        pp.HgrSpritePlot(x0, y0, p)
        //if p & 3 == 3 {
        //   pp.HgrSpritePlot(x0+1, y0, p)
       // }
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

func Swap( a, b *int ) {
	c := *a
	*a = *b
	*b = c
}

func Abs(a int) int {
	if a < 0 {
		a = -a
	}
	return a
}

func BrenshamLine3D( p Plotable3, x0, y0, z0, x1, y1, z1 int, c int ) {
    var x, delta_x, step_x int
    var y, delta_y, step_y int
    var z, delta_z, step_z int
    var swap_xy, swap_xz bool
    var drift_xy, drift_xz int
    var cx, cy, cz int 	

   	swap_xy = (Abs(y1 - y0) > Abs(x1 - x0))
    if swap_xy {
        Swap(&x0, &y0)
        Swap(&x1, &y1)
	}

    swap_xz = (Abs(z1 - z0) > Abs(x1 - x0))
    if swap_xz {
        Swap(&x0, &z0)
        Swap(&x1, &z1)
	}

    delta_x = Abs(x1 - x0)
    delta_y = Abs(y1 - y0)
    delta_z = Abs(z1 - z0)

    drift_xy  = (delta_x / 2)
    drift_xz  = (delta_x / 2)

	step_x = 1
	if (x0 > x1) { step_x = -1 }
    step_y = 1
	if (y0 > y1) { step_y = -1 }
    step_z = 1  
    if (z0 > z1) { step_z = -1 }

	y = y0
	z = z0  

    //step through longest delta (which we have swapped to x)
    for x = x0; x <= x1; x += step_x {
        
        //copy position
        cx = x
        cy = y    
        cz = z

        //unswap (in reverse)
        if swap_xz { 
	        Swap(&cx, &cz)
        }
        if swap_xy {
	        Swap(&cx, &cy)
        }

        // plot
        p.Plot(cx, cy, cz, c)
        
        //update progress in other planes
        drift_xy = drift_xy - delta_y
        drift_xz = drift_xz - delta_z

        //step in y plane
        if drift_xy < 0 {
            y = y + step_y
            drift_xy = drift_xy + delta_x
    	}
        
        //same in z
        if drift_xz < 0 {
            z = z + step_z
            drift_xz = drift_xz + delta_x
    	}
    	
	}
	  	
}