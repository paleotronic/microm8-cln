package types

import (
	"paleotronic.com/log"
)

// CubePos defines the position of a cube.  Used as a key to
// accessing the CubeMap
type CubePos [3]int

// CubeMap defines a fixed size map of cubes
type CubeMap struct {
	cubes                         map[CubePos]int // position ->> color
	maxwidth, maxheight, maxdepth int
	nullcolor                     int
	vb                            *VectorBuffer
}

// NewCubeMap initialises an instance of CubeMap with the given
// size.
func NewCubeMap(w, h, d int, vbuff *VectorBuffer) *CubeMap {

	this := &CubeMap{
		maxwidth:  w,
		maxheight: h,
		maxdepth:  d,
		cubes:     make(map[CubePos]int),
		vb:        vbuff,
	}

	return this

}

// Within returns true if the given x, y, z is within the
// CubeMap's viable co-ordinates, false otherwise
func (c *CubeMap) Within(x, y, z int) bool {
	if x < 0 || y < 0 || z < 0 {
		return false
	}
	if x >= c.maxwidth || y >= c.maxheight || z >= c.maxdepth {
		return false
	}
	return true
}

// ColorAt returns col of cube at location, or the nullcolor if there is no
// cube.  Also returns nullcolor if the x, y, z is outside of the
// bounds of the 3 cube space.
func (c *CubeMap) ColorAt(x, y, z int) int {

	if !c.Within(x, y, z) {
		return c.nullcolor
	}

	key := CubePos{x, y, z}

	col, ok := c.cubes[key]

	if !ok {
		return c.nullcolor
	}

	return col
}

// Plot puts a cube of specified color at the given co-ordinates
func (c *CubeMap) Plot(x, y, z, col int) {
	log.Printf("CubeMap.Plot(%d,%d,%d,%d)\n", x, y, z, col)

	if !c.Within(x, y, z) {
		return
	}

	key := CubePos{x, y, z}

	if col == c.nullcolor {
		// erase
		delete(c.cubes, key)
	} else {
		c.cubes[key] = col
	}

	c.Render()
}

// Clear removes all plotted cubes.
func (c *CubeMap) Clear() {
	c.cubes = make(map[CubePos]int)
	c.Render()
}

// Render populates memory with an encoded representation of the CubeMap
func (c *CubeMap) Render() {
	c.vb.Vectors = make(VectorList, 0) // empty the vector list
	for key, col := range c.cubes {
		// we need to adjust the x, y coordinates to be centered
		ay := c.maxheight - key[1] - 1
		x1, y1, z1 := float32(key[0]-c.maxwidth/2), float32(ay-c.maxheight/2), float32(key[2])
		v := NewVector(
			VT_CUBE,
			uint64(col),
			x1, y1, z1,
			1, 1, 1,
		)
		c.vb.Vectors = append(c.vb.Vectors, v)
		log.Println(key, v)
	}

	// Publish
	c.vb.WriteToMemory()
}

// Swap 2 ints
func Swap(a, b *int) {
	c := *a
	*a = *b
	*b = c
}

// Abs returns the absolute value of the specified int
func Abs(a int) int {
	if a < 0 {
		a = -a
	}
	return a
}

// Line plots a line between 2 points
func (cm *CubeMap) Line(x0, y0, z0, x1, y1, z1 int, c int) {
	var x, delta_x, step_x int
	var y, delta_y, step_y int
	var z, delta_z, step_z int
	var swap_xy, swap_xz bool
	var drift_xy, drift_xz int
	var cx, cy, cz int

	swap_xy = (Abs(y1-y0) > Abs(x1-x0))
	if swap_xy {
		Swap(&x0, &y0)
		Swap(&x1, &y1)
	}

	swap_xz = (Abs(z1-z0) > Abs(x1-x0))
	if swap_xz {
		Swap(&x0, &z0)
		Swap(&x1, &z1)
	}

	delta_x = Abs(x1 - x0)
	delta_y = Abs(y1 - y0)
	delta_z = Abs(z1 - z0)

	drift_xy = (delta_x / 2)
	drift_xz = (delta_x / 2)

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
		cm.Plot(cx, cy, cz, c)

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
