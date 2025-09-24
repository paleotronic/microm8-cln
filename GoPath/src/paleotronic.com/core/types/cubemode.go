package types

import (
	"paleotronic.com/log"

	"paleotronic.com/core/memory"
)

const maxPoints = 65534 * 2

// CubePoint defines the position of a cube.  Used as a key to
// accessing the CubeScreen
type CubePoint struct {
	X, Y, Z uint8
}

type CubePointColor struct {
	P CubePoint
	C uint8
}

// CubeScreen defines a fixed size map of cubes
type CubeScreen struct {
	baseAddr   int
	bufferSize int
	//cubes         map[CubePoint]uint8 // position ->> color
	nullcolor     uint8
	buffer        *memory.MemoryControlBlock
	DelayedRender bool
	cubeAddress   map[CubePoint]int
	usedAddress   [maxPoints]bool
	earliestFree  int
	maxUsed       int
}

// NewCubeScreen initialises an instance of CubeScreen with the given
// size.
func NewCubeScreen(base int, bufferSize int, mr *memory.MemoryControlBlock) *CubeScreen {

	this := &CubeScreen{
		nullcolor: 0,
		//cubes:        make(map[CubePoint]uint8),
		buffer:       mr,
		baseAddr:     base,
		bufferSize:   bufferSize,
		cubeAddress:  make(map[CubePoint]int),
		earliestFree: 0,
		maxUsed:      0,
	}
	this.Clear()

	return this

}

// Return index or -1 if point does not exist
func (c *CubeScreen) indexOfPoint(x uint8, y uint8, z uint8) int {
	idx, ok := c.cubeAddress[CubePoint{x, y, z}]
	if ok {
		return idx
	}
	return -1
}

// ColorAt returns col of cube at location, or the nullcolor if there is no
// cube.  Also returns nullcolor if the x, y, z is outside of the
// bounds of the 3 cube space.
func (c *CubeScreen) ColorAt(x, y, z uint8) uint8 {
	_, col := c.ColorAtIndex(c.indexOfPoint(x, y, z))
	return col
}

func (c *CubeScreen) ColorAtIndex(idx int) (CubePoint, uint8) {
	if idx == -1 {
		return CubePoint{}, 0
	}
	offset := 2 + idx/2
	u := c.buffer.Read(offset)
	if idx%2 == 0 {
		u >>= 32
	}
	return CubePoint{
		uint8(u >> 24),
		uint8(u >> 16),
		uint8(u >> 8),
	}, uint8(u & 0xff)
}

func (c *CubeScreen) SetColorAtIndex(idx int, p CubePoint, col uint8) {
	if idx == -1 {
		return
	}
	offset := 2 + idx/2
	u := c.buffer.Read(offset)
	if idx%2 == 0 {
		u = (u & 0x00000000ffffffff) | (uint64(col) << 32) | (uint64(p.Z) << 40) | (uint64(p.Y) << 48) | (uint64(p.X) << 56)
	} else {
		u = (u & 0xffffffff00000000) | uint64(col) | (uint64(p.Z) << 8) | (uint64(p.Y) << 16) | (uint64(p.X) << 24)
	}
	c.buffer.Write(offset, u)
}

// Plot puts a cube of specified color at the given co-ordinates
func (c *CubeScreen) Plot(x, y, z, col uint8) {
	//log.Printf("CubeScreen.Plot(%d,%d,%d,%d)\n", x, y, z, col)

	if c.maxUsed >= maxPoints && col != 0 {
		return
	}

	idx := c.indexOfPoint(x, y, z)
	log.Printf("Index of point (%d, %d, %d) -> %d", x, y, z, idx)

	key := CubePoint{x, y, z}

	// if idx != -1 && col == c.nullcolor {
	// 	delete(c.cubeAddress, key)
	// 	c.markUsed(idx, false)
	// } else if idx != -1 {
	// 	c.markUsed(idx, true)
	// 	c.SetColorAtIndex(idx, key, col)
	// 	c.cubeAddress[key] = idx
	// } else {
	// 	idx = c.nextUnused()
	// 	c.markUsed(idx, true)
	// 	c.SetColorAtIndex(idx, key, col)
	// 	c.cubeAddress[key] = idx
	// }

	if idx != -1 {
		if col == c.nullcolor {
			delete(c.cubeAddress, key)
			c.SetColorAtIndex(idx, key, col)
			c.markUsed(idx, false)
		} else {
			c.markUsed(idx, true)
			c.SetColorAtIndex(idx, key, col)
			c.cubeAddress[key] = idx
		}
	} else if col != c.nullcolor {
		idx = c.nextUnused()
		c.markUsed(idx, true)
		c.SetColorAtIndex(idx, key, col)
		c.cubeAddress[key] = idx
	}

	c.Render()
	c.dbg()
}

func (c *CubeScreen) nextUnused() int {
	if c.earliestFree >= len(c.usedAddress) {
		return -1
	}
	for c.usedAddress[c.earliestFree] && c.earliestFree < maxPoints {
		c.earliestFree++
	}
	if c.earliestFree >= maxPoints {
		return -1
	}
	if c.earliestFree >= c.maxUsed {
		c.maxUsed = c.earliestFree
	}
	//c.earliestFree++
	c.dbg()
	return c.earliestFree
}

func (c *CubeScreen) markUsed(idx int, used bool) {
	c.usedAddress[idx] = used
	if !used && idx < c.earliestFree {
		c.earliestFree = idx
	} else if !used && idx == c.maxUsed && c.maxUsed > 0 {
		c.maxUsed--
	}
	c.dbg()
}

func (c *CubeScreen) dbg() {
	log.Printf("earliestFree = %d, maxUsed = %d, count points = %d", c.earliestFree, c.maxUsed, len(c.cubeAddress))
}

// Clear removes all plotted cubes.
func (c *CubeScreen) Clear() {
	//c.cubes = make(map[CubePoint]uint8)
	c.cubeAddress = make(map[CubePoint]int)
	for i, _ := range c.usedAddress {
		c.usedAddress[i] = false
	}
	c.earliestFree = 0
	c.maxUsed = 0
	c.SetColorAtIndex(0, CubePoint{0, 0, 0}, 0)
	c.dbg()
	c.Render()
}

// Render populates memory with an encoded representation of the CubeScreen
func (c *CubeScreen) Render() {
	c.buffer.Write(0, uint64(c.maxUsed))
	c.buffer.Write(1, 1)
}

func (c *CubeScreen) Size() int {
	return int(c.buffer.Read(0))
}

func (c *CubeScreen) Changed() bool {
	return c.buffer.Read(1) != 0
}

func (c *CubeScreen) GetMap() []CubePointColor {
	l := make([]CubePointColor, 0, len(c.cubeAddress))
	c.maxUsed = int(c.buffer.Read(0))
	for idx := 0; idx <= c.maxUsed; idx++ {
		p, col := c.ColorAtIndex(idx)
		//log.Printf("Point %+v -> %d", p, col)
		if col != 0 {
			l = append(l, CubePointColor{P: p, C: col})
		}
	}
	//log.Printf("GetMap() returns %d points.", len(l))
	c.buffer.Write(1, 0)
	return l
}

func (c *CubeScreen) Line3d(gx0, gy0, gz0, gx1, gy1, gz1 float64, col uint8) {
	c.Bresenham3D(
		int(gx0),
		int(gy0),
		int(gz0),
		int(gx1),
		int(gy1),
		int(gz1),
		col,
	)
	c.Render()
}

func abs(x int) int {
	if x < 0 {
		return 0 - x
	}
	return x
}

func iif(cond bool, a, b int) int {
	if cond {
		return a
	}
	return b
}

func (c *CubeScreen) Bresenham3D(x1, y1, z1, x2, y2, z2 int, col uint8) {

	var i, dx, dy, dz, l, m, n, x_inc, y_inc, z_inc, err_1, err_2, dx2, dy2, dz2 int
	var point [3]int

	point[0] = x1
	point[1] = y1
	point[2] = z1
	dx = x2 - x1
	dy = y2 - y1
	dz = z2 - z1
	x_inc = iif(dx < 0, -1, 1)
	l = abs(dx)
	y_inc = iif(dy < 0, -1, 1)
	m = abs(dy)
	z_inc = iif(dz < 0, -1, 1)
	n = abs(dz)
	dx2 = l << 1
	dy2 = m << 1
	dz2 = n << 1

	if (l >= m) && (l >= n) {
		err_1 = dy2 - l
		err_2 = dz2 - l
		for i = 0; i < l; i++ {
			if point[0] >= 0 && point[0] <= 255 && point[1] >= 0 && point[1] <= 255 && point[2] >= 0 && point[2] <= 255 {
				c.Plot(
					uint8(point[0]),
					uint8(point[1]),
					uint8(point[2]),
					col,
				)
			}
			if err_1 > 0 {
				point[1] += y_inc
				err_1 -= dx2
			}
			if err_2 > 0 {
				point[2] += z_inc
				err_2 -= dx2
			}
			err_1 += dy2
			err_2 += dz2
			point[0] += x_inc
		}
	} else if (m >= l) && (m >= n) {
		err_1 = dx2 - m
		err_2 = dz2 - m
		for i = 0; i < m; i++ {
			if point[0] >= 0 && point[0] <= 255 && point[1] >= 0 && point[1] <= 255 && point[2] >= 0 && point[2] <= 255 {
				c.Plot(
					uint8(point[0]),
					uint8(point[1]),
					uint8(point[2]),
					col,
				)
			}
			if err_1 > 0 {
				point[0] += x_inc
				err_1 -= dy2
			}
			if err_2 > 0 {
				point[2] += z_inc
				err_2 -= dy2
			}
			err_1 += dx2
			err_2 += dz2
			point[1] += y_inc
		}
	} else {
		err_1 = dy2 - n
		err_2 = dx2 - n
		for i = 0; i < n; i++ {
			if point[0] >= 0 && point[0] <= 255 && point[1] >= 0 && point[1] <= 255 && point[2] >= 0 && point[2] <= 255 {
				c.Plot(
					uint8(point[0]),
					uint8(point[1]),
					uint8(point[2]),
					col,
				)
			}
			if err_1 > 0 {
				point[1] += y_inc
				err_1 -= dz2
			}
			if err_2 > 0 {
				point[0] += x_inc
				err_2 -= dz2
			}
			err_1 += dy2
			err_2 += dx2
			point[2] += z_inc
		}
	}
	if point[0] >= 0 && point[0] <= 255 && point[1] >= 0 && point[1] <= 255 && point[2] >= 0 && point[2] <= 255 {
		c.Plot(
			uint8(point[0]),
			uint8(point[1]),
			uint8(point[2]),
			col,
		)
	}
	defer c.Render()
}
