package types

import (
	"image"
	"math/rand"
	"strings"

	"paleotronic.com/core/memory"
)

const (
	SPRITE_ENABLED = 0
	SPRITE_INFO    = SPRITE_ENABLED + 2
	SPRITE_DATA    = SPRITE_INFO + 128
	SPRITE_BACK    = 130 + 48*128
)

type SpriteRotation int

const (
	SPRITE_ROTATE_0 SpriteRotation = iota
	SPRITE_ROTATE_90
	SPRITE_ROTATE_180
	SPRITE_ROTATE_270
)

type SpriteFlip int

const (
	SPRITE_FLIP_NONE SpriteFlip = iota
	SPRITE_FLIP_VERTICAL
	SPRITE_FLIP_HORIZONTAL
	SPRITE_FLIP_BOTH
)

type SpriteScale int

const (
	SPRITE_SCALE_1X SpriteScale = iota
	SPRITE_SCALE_2X
	SPRITE_SCALE_3X
)

type SpriteBounds struct {
	X, Y, Size int
}

type SpriteController struct {
	Index  int
	M      *memory.MemoryMap
	Base   int
	dx, dy int
}

func NewSpriteController(index int, m *memory.MemoryMap, baseaddr int) *SpriteController {
	return &SpriteController{
		Index: index,
		M:     m,
		Base:  baseaddr,
	}
}

func (s *SpriteController) Reset() {
	var empty [24][24]byte
	for i := 0; i < memory.MICROM8_MAX_SPRITES; i++ {
		s.SetEnabled(i, false)
		s.SetSpriteData(i, empty)
		s.SetSpriteAttr(i, 0, 0, 0, 0, 0, SpriteBounds{}, 0)
	}
}

func (s *SpriteController) GetEnabledStates() []bool {
	var b uint64
	var out = make([]bool, memory.MICROM8_MAX_SPRITES)
	for i := 0; i < memory.MICROM8_MAX_SPRITES; i++ {
		if i%64 == 0 {
			b = s.M.ReadInterpreterMemorySilent(s.Index, s.Base+SPRITE_ENABLED+(i/64))
		}
		out[i] = b&1 != 0
		b >>= 1
	}
	return out
}

func (s *SpriteController) GetEnabledIndexes() []int {
	var b uint64
	var out = make([]int, 0, memory.MICROM8_MAX_SPRITES)
	for i := 0; i < memory.MICROM8_MAX_SPRITES; i++ {
		if i%64 == 0 {
			b = s.M.ReadInterpreterMemorySilent(s.Index, s.Base+SPRITE_ENABLED+(i/64))
		}
		if b&1 != 0 {
			out = append(out, i)
		}
		b >>= 1
	}
	return out
}

func (s *SpriteController) GetEnabled(i int) bool {
	var b uint64
	b = s.M.ReadInterpreterMemorySilent(s.Index, s.Base+SPRITE_ENABLED+(i/64))
	b >>= (uint(i) % 64)
	return b&1 != 0
}

func (s *SpriteController) SetEnabled(i int, e bool) {
	var b uint64
	b = s.M.ReadInterpreterMemorySilent(s.Index, s.Base+SPRITE_ENABLED+(i/64))
	var bitmask uint64 = 1 << (uint(i) % 64)
	var clrmask uint64 = 0xFFFFFFFFFFFFFFFF ^ bitmask
	if e {
		b |= bitmask
	} else {
		b &= clrmask
	}
	s.M.WriteInterpreterMemory(s.Index, s.Base+SPRITE_ENABLED+(i/64), b)
}

func (s *SpriteController) FindCollidingSprites(x, y int, bounds SpriteBounds, scl SpriteScale, ignoreId int) []int {
	var out = []int{}

	if bounds.Size == 0 {
		bounds.Size = 24
	}
	var a = image.Rect(
		x+bounds.X*int(scl+1),
		y+bounds.Y*int(scl+1),
		x+bounds.X*int(scl+1)+bounds.Size*int(scl+1),
		y+bounds.Y*int(scl+1)+bounds.Size*int(scl+1),
	)
	var b image.Rectangle

	enabledIds := s.GetEnabledIndexes()
	for _, id := range enabledIds {
		if id == ignoreId {
			continue
		}
		sx, sy, _, _, sscl, sbounds, _ := s.GetSpriteAttr(id)
		if sbounds.Size == 0 {
			sbounds.Size = 24
		}
		b = image.Rect(
			sx+sbounds.X*int(sscl+1),
			sy+sbounds.Y*int(sscl+1),
			sx+sbounds.X*int(sscl+1)+sbounds.Size*int(sscl+1),
			sy+sbounds.Y*int(sscl+1)+sbounds.Size*int(sscl+1),
		)

		if b.Intersect(a).Size().X > 0 {
			out = append(out, id)
		}
	}

	return out
}

func (s *SpriteController) GetSpriteAttr(i int) (x int, y int, rot SpriteRotation, flip SpriteFlip, scl SpriteScale, bounds SpriteBounds, color int) {
	if i < 0 || i >= memory.MICROM8_MAX_SPRITES {
		return
	}
	addr := s.Base + SPRITE_INFO + i
	memval := s.M.ReadInterpreterMemorySilent(s.Index, addr)
	// [x 12 bits::y 12 bits::rot 2::flip 2::scl 2::]
	x = int(memval) & 0x3ff
	y = int(memval>>10) & 0x3ff
	rot = SpriteRotation((memval >> 20) & 3)
	flip = SpriteFlip((memval >> 22) & 3)
	scl = SpriteScale((memval >> 24) & 3)
	// bounds bits x0: 26-30
	bounds.X = int(memval>>26) & 31
	bounds.Y = int(memval>>31) & 31
	bounds.Size = int(memval>>36) & 31
	if bounds.Size == 0 {
		bounds.Size = 24
	}
	color = int(memval>>41) & 15
	// next attr at 46
	return
}

func (s *SpriteController) SetSpriteAttr(i int, x, y int, rot SpriteRotation, flip SpriteFlip, scl SpriteScale, bounds SpriteBounds, color int) {
	var b = uint64(x) |
		(uint64(y) << 10) |
		(uint64(rot) << 20) |
		(uint64(flip) << 22) |
		(uint64(scl) << 24) |
		(uint64(bounds.X) << 26) |
		(uint64(bounds.Y) << 31) |
		(uint64(bounds.Size) << 36) |
		(uint64(color) << 41)

	addr := s.Base + SPRITE_INFO + i
	s.M.WriteInterpreterMemory(s.Index, addr, b)
}

func (s *SpriteController) GetSpriteData(i int) [24][24]byte {
	var addr = s.Base + SPRITE_DATA + i*48
	var data = s.M.BlockRead(s.Index, s.M.MEMBASE(s.Index)+addr, 48)
	return decodeSpriteData(data)
}

func (s *SpriteController) SetSpriteData(i int, data [24][24]byte) {
	var addr = s.Base + SPRITE_DATA + i*48
	memvals := encodeSpriteData(data)
	s.M.BlockWrite(s.Index, s.M.MEMBASE(s.Index)+addr, memvals)
	//
	s.SetSpriteAttr(i, 0, 0, 0, 0, 0, SpriteBounds{}, 0)
}

func (s *SpriteController) GetSpriteBacking(i int) [24][24]byte {
	var addr = s.Base + SPRITE_BACK + i*48
	var data = s.M.BlockRead(s.Index, s.M.MEMBASE(s.Index)+addr, 48)
	return decodeSpriteData(data)
}

func (s *SpriteController) SetSpriteBacking(i int, data [24][24]byte) {
	var addr = s.Base + SPRITE_BACK + i*48
	memvals := encodeSpriteData(data)
	s.M.BlockWrite(s.Index, addr, memvals)
}

func (s *SpriteController) GetScale(sno int) SpriteScale {
	_, _, _, _, scl, _, _ := s.GetSpriteAttr(sno)
	return scl
}

func (s *SpriteController) GetFlip(sno int) SpriteFlip {
	_, _, _, flip, _, _, _ := s.GetSpriteAttr(sno)
	return flip
}

func (s *SpriteController) GetRotation(sno int) SpriteRotation {
	_, _, rot, _, _, _, _ := s.GetSpriteAttr(sno)
	return rot
}

func (s *SpriteController) SetX(sno int, v int) {
	_, y, rot, flip, scl, bounds, col := s.GetSpriteAttr(sno)
	s.SetSpriteAttr(sno, v, y, rot, flip, scl, bounds, col)
}

func (s *SpriteController) SetY(sno int, v int) {
	x, _, rot, flip, scl, bounds, col := s.GetSpriteAttr(sno)
	s.SetSpriteAttr(sno, x, v, rot, flip, scl, bounds, col)
}

func (s *SpriteController) SetBounds(sno int, v SpriteBounds) {
	x, y, rot, flip, scl, _, col := s.GetSpriteAttr(sno)
	s.SetSpriteAttr(sno, x, y, rot, flip, scl, v, col)
}

func (s *SpriteController) GetX(sno int) int {
	x, _, _, _, _, _, _ := s.GetSpriteAttr(sno)
	return x
}

func (s *SpriteController) GetY(sno int) int {
	_, y, _, _, _, _, _ := s.GetSpriteAttr(sno)
	return y
}

func (s *SpriteController) SetRotation(sno int, newrot SpriteRotation) {
	x, y, rot, flip, scl, bounds, col := s.GetSpriteAttr(sno)
	if rot == newrot {
		return
	}
	data := s.GetSpriteData(sno)

	switch rot {
	case SPRITE_ROTATE_0:
		switch newrot {
		case SPRITE_ROTATE_90:
			data = rotateRight(data, bounds)
		case SPRITE_ROTATE_180:
			data = rotateRight(data, bounds)
			data = rotateRight(data, bounds)
		case SPRITE_ROTATE_270:
			data = rotateLeft(data, bounds)
		}
	case SPRITE_ROTATE_90:
		switch newrot {
		case SPRITE_ROTATE_0:
			data = rotateLeft(data, bounds)
		case SPRITE_ROTATE_180:
			data = rotateRight(data, bounds)
		case SPRITE_ROTATE_270:
			data = rotateRight(data, bounds)
			data = rotateRight(data, bounds)
		}
	case SPRITE_ROTATE_180:
		switch newrot {
		case SPRITE_ROTATE_0:
			data = rotateLeft(data, bounds)
			data = rotateLeft(data, bounds)
		case SPRITE_ROTATE_90:
			data = rotateLeft(data, bounds)
		case SPRITE_ROTATE_270:
			data = rotateRight(data, bounds)
		}
	case SPRITE_ROTATE_270:
		switch newrot {
		case SPRITE_ROTATE_0:
			data = rotateRight(data, bounds)
		case SPRITE_ROTATE_90:
			data = rotateLeft(data, bounds)
			data = rotateLeft(data, bounds)
		case SPRITE_ROTATE_180:
			data = rotateLeft(data, bounds)
		}
	}

	s.SetSpriteData(sno, data)
	s.SetSpriteAttr(sno, x, y, newrot, flip, scl, bounds, col)
}

func (s *SpriteController) SetScale(sno int, newscl SpriteScale) {
	x, y, rot, flip, _, bounds, col := s.GetSpriteAttr(sno)
	s.SetSpriteAttr(sno, x, y, rot, flip, newscl, bounds, col)
}

func (s *SpriteController) SetColor(sno int, newcol int) {
	x, y, rot, flip, scl, bounds, _ := s.GetSpriteAttr(sno)
	s.SetSpriteAttr(sno, x, y, rot, flip, scl, bounds, newcol)
}

func (s *SpriteController) GetColor(sno int) int {
	_, _, _, _, _, _, col := s.GetSpriteAttr(sno)
	return col
}

func (s *SpriteController) SetFlip(sno int, newflip SpriteFlip) {
	// get current state
	x, y, rot, flip, scl, bounds, col := s.GetSpriteAttr(sno)
	if flip == newflip {
		return
	}
	data := s.GetSpriteData(sno)
	switch flip {
	case SPRITE_FLIP_NONE:
		switch newflip {
		case SPRITE_FLIP_HORIZONTAL:
			data = flipHorizontal(data, bounds)
		case SPRITE_FLIP_VERTICAL:
			data = flipVertical(data, bounds)
		case SPRITE_FLIP_BOTH:
			data = flipHorizontal(data, bounds)
			data = flipVertical(data, bounds)
		}
	case SPRITE_FLIP_HORIZONTAL:
		switch newflip {
		case SPRITE_FLIP_NONE:
			data = flipHorizontal(data, bounds)
		case SPRITE_FLIP_BOTH:
			data = flipVertical(data, bounds)
		case SPRITE_FLIP_VERTICAL:
			data = flipHorizontal(data, bounds)
			data = flipVertical(data, bounds)
		}
	case SPRITE_FLIP_VERTICAL:
		switch newflip {
		case SPRITE_FLIP_NONE:
			data = flipVertical(data, bounds)
		case SPRITE_FLIP_BOTH:
			data = flipHorizontal(data, bounds)
		case SPRITE_FLIP_HORIZONTAL:
			data = flipVertical(data, bounds)
			data = flipHorizontal(data, bounds)
		}
	case SPRITE_FLIP_BOTH:
		switch newflip {
		case SPRITE_FLIP_NONE:
			data = flipVertical(data, bounds)
			data = flipHorizontal(data, bounds)
		case SPRITE_FLIP_VERTICAL:
			data = flipHorizontal(data, bounds)
		case SPRITE_FLIP_HORIZONTAL:
			data = flipVertical(data, bounds)
		}
	}
	s.SetSpriteData(sno, data)
	s.SetSpriteAttr(sno, x, y, rot, newflip, scl, bounds, col)
}

const maxreps = 11

// Get definition returns the RLE version of the sprite data
func (s *SpriteController) GetDefinition(sno int) string {
	data := s.GetSpriteData(sno)
	return encodeRLE(data)
}

func (s *SpriteController) SetDefinition(sno int, rledata string) {
	data := decodeRLE(rledata)
	s.SetSpriteData(sno, data)
}

// Some play test stuff

func (s *SpriteController) TestMode() {
	var sprite [24][24]byte
	for x, _ := range sprite {
		for y, _ := range sprite[x] {
			sprite[x][y] = byte(rand.Intn(16))
		}
	}
	sno := 127
	s.SetEnabled(sno, true)
	s.SetSpriteData(sno, sprite)
	x, y, rot, flip, scl, bounds, col := s.GetSpriteAttr(sno)
	scl = SPRITE_SCALE_1X
	x = 50
	y = 30
	bounds.Size = 24
	s.SetSpriteAttr(sno, x, y, rot, flip, scl, bounds, col)
	s.dx = 1
	s.dy = -1
}

func (s *SpriteController) TestMove() {
	sno := 127
	x, y, rot, flip, scl, bounds, col := s.GetSpriteAttr(sno)
	x += s.dx
	y += s.dy
	var r bool
	if x < 0 || x > 279 {
		s.dx = -s.dx
		x += s.dx
		r = true
	}
	if y < 0 || y > 159 {
		s.dy = -s.dy
		y += s.dy
		r = true
	}
	s.SetSpriteAttr(sno, x, y, rot, flip, scl, bounds, col)
	if r {
		data := s.GetSpriteData(sno)
		data = rotateLeft(data, bounds)
		s.SetSpriteData(sno, data)
	}
}

// ------------------------------------ internal methods

func decodeRLE(str string) [24][24]byte {
	str = strings.ToLower(str)
	var out [24][24]byte
	var x, y int
	var c byte
	var repcount int
	var adv = func() bool {
		x++
		if x >= 24 {
			x = 0
			y++
			if y >= 24 {
				return false
			}
		}
		return true
	}
	for _, ch := range str {
		switch ch {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			c = byte(ch - '0')
			out[x][y] = c
			repcount = 0
			if !adv() {
				return out
			}
		case 'a', 'b', 'c', 'd', 'e', 'f':
			c = 10 + byte(ch-'a')
			out[x][y] = c
			repcount = 0
			if !adv() {
				return out
			}
		case 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
			repcount = int(ch-'p') + 1
			for repcount > 0 {
				out[x][y] = c
				repcount--
				if !adv() {
					return out
				}
			}
		}
	}
	return out
}

func encodeRLE(data [24][24]byte) string {
	var counts = []string{"", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	var hex = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
	var out = ""
	var repcount = 0
	var col = byte(0xff)
	var newc byte
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			newc = data[x][y]
			if newc != col {
				if repcount > 0 {
					out += counts[repcount]
					repcount = 0
				}
				out += hex[newc]
			} else if repcount < 11 {
				repcount++
			} else {
				if repcount > 0 {
					out += counts[repcount]
					repcount = 0
				}
				out += hex[newc]
			}
			col = newc
		}
	}
	if repcount > 0 {
		out += counts[repcount]
	}
	return strings.ToUpper(out)
}

func decodeSpriteData(data []uint64) [24][24]byte {
	var out [24][24]byte
	var bitmask uint64
	var sb uint
	var b uint64
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			b = data[y*2+(x/16)]
			bitmask = 0xf000000000000000 >> ((uint(x) % 16) * 4)
			sb = (15 - (uint(x) % 16)) * 4
			out[x][y] = byte((b & bitmask) >> sb)
		}
	}
	return out
}

func encodeSpriteData(data [24][24]byte) []uint64 {
	var out = make([]uint64, 48)
	var sb uint
	var b uint64
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			b = out[y*2+(x/16)]
			sb = (15 - (uint(x) % 16)) * 4
			out[y*2+(x/16)] = b | uint64(data[x][y])<<sb
		}
	}
	return out
}

// transformSprite performs an aribitrary transform...
func transformSprite(data [24][24]byte, bounds SpriteBounds, f func(bounds SpriteBounds, x, y int) (nx, ny int)) [24][24]byte {
	if f == nil {
		return data
	}
	var out = data
	var c byte
	var nx, ny int
	for y := 0; y < bounds.Size; y++ {
		for x := 0; x < bounds.Size; x++ {
			c = data[bounds.X+x][bounds.Y+y]
			nx, ny = f(bounds, x, y)
			if nx >= 0 && nx <= 23 && ny >= 0 && ny <= 23 {
				out[nx][ny] = c
			}
		}
	}
	return out
}

func trnsfFlipV(bounds SpriteBounds, x, y int) (nx, ny int) {
	nx = bounds.X + x
	ny = bounds.Y + (bounds.Size - y - 1)
	return
}

func trnsfFlipH(bounds SpriteBounds, x, y int) (nx, ny int) {
	nx = bounds.X + (bounds.Size - x - 1)
	ny = bounds.Y + y
	return
}

func trnsfRotL(bounds SpriteBounds, x, y int) (nx, ny int) {
	ny = bounds.Y + (bounds.Size - x - 1)
	nx = bounds.X + y
	return
}

func trnsfRotR(bounds SpriteBounds, x, y int) (nx, ny int) {
	nx = bounds.X + (bounds.Size - y - 1)
	ny = bounds.Y + x
	return
}

func flipVertical(data [24][24]byte, bounds SpriteBounds) [24][24]byte {
	return transformSprite(data, bounds, trnsfFlipV)
}

func flipHorizontal(data [24][24]byte, bounds SpriteBounds) [24][24]byte {
	return transformSprite(data, bounds, trnsfFlipH)
}

func rotateLeft(data [24][24]byte, bounds SpriteBounds) [24][24]byte {
	return transformSprite(data, bounds, trnsfRotL)
}

func rotateRight(data [24][24]byte, bounds SpriteBounds) [24][24]byte {
	return transformSprite(data, bounds, trnsfRotR)
}
