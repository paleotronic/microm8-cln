package types

import (
	"math/rand"
	"testing"

	"paleotronic.com/core/memory"

	"fmt"
)

func TestEncodeDecodeSprite(t *testing.T) {

	var sprite [24][24]byte
	for x, _ := range sprite {
		for y, _ := range sprite[x] {
			sprite[x][y] = byte(rand.Intn(16))
		}
	}
	t.Logf("sprite data: %+v", sprite)

	data := encodeSpriteData(sprite)

	sprite2 := decodeSpriteData(data)

	for x, _ := range sprite {
		for y, _ := range sprite[x] {
			if sprite[x][y] != sprite2[x][y] {
				t.Fatalf("data at %d,%d is different after decode: %d vs %d", x, y, sprite[x][y], sprite2[x][y])
			}
		}
	}

}

func TestSpriteEnableDisableState(t *testing.T) {

	m := memory.NewMemoryMap()
	sc := NewSpriteController(0, m, memory.MICROM8_SPRITE_CONTROL_BASE)

	spritesToEnable := []int{1, 25, 77, 99, 127}

	for _, sno := range spritesToEnable {
		sc.SetEnabled(sno, true)
	}

	list := sc.GetEnabledIndexes()

	for i, v := range list {
		if spritesToEnable[i] != v {
			t.Fatal("didn't get the same list back")
		}
	}

}

func TestSetGetInfo(t *testing.T) {

	m := memory.NewMemoryMap()
	sc := NewSpriteController(0, m, memory.MICROM8_SPRITE_CONTROL_BASE)

	x, y, rot, flip, scl, b, col := sc.GetSpriteAttr(10)
	x = 50
	y = 20
	rot = 1
	flip = 2
	scl = 2
	sc.SetSpriteAttr(10, x, y, rot, flip, scl, b, col)

	x2, y2, rot2, flip2, scl2, _, col2 := sc.GetSpriteAttr(10)

	if x2 != x {
		t.Fatal("x does not match")
	}

	if y2 != y {
		t.Fatal("y does not match")
	}

	if rot2 != rot {
		t.Fatal("rot does not match")
	}

	if flip2 != flip {
		t.Fatal("flip does not match")
	}

	if scl2 != scl {
		t.Fatal("scl does not match")
	}

	if col2 != col {
		t.Fatal("col does not match")
	}

}

func dump(data [24][24]byte, bounds SpriteBounds) {
	fmt.Println("dump data")
	for y := bounds.Y; y < bounds.Y+bounds.Size; y++ {
		for x := bounds.X; x < bounds.X+bounds.Size; x++ {
			if data[x][y] != 0 {
				fmt.Printf("%.2x ", data[x][y])
			} else {
				fmt.Printf("   ")
			}
		}
		fmt.Println()
	}
}

func TestRotateSprite(t *testing.T) {

	var sprite [24][24]byte
	for x, _ := range sprite {
		for y, _ := range sprite[x] {
			sprite[x][y] = byte(x + y)
		}
	}

	bounds := SpriteBounds{0, 0, 8}
	dump(sprite, bounds)
	sprite = rotateLeft(sprite, bounds)
	dump(sprite, bounds)
	sprite = rotateRight(sprite, bounds)
	dump(sprite, bounds)
	sprite = flipVertical(sprite, bounds)
	dump(sprite, bounds)
	sprite = flipVertical(sprite, bounds)
	sprite = flipHorizontal(sprite, bounds)
	dump(sprite, bounds)

}

func TestEncodeDecodeRLE(t *testing.T) {
	var sprite [24][24]byte
	for x, _ := range sprite {
		for y, _ := range sprite[x] {
			if rand.Intn(100) > 75 {
				sprite[x][y] = byte(rand.Intn(16))
			}
		}
	}
	t.Logf("sprite data: %+v", sprite)

	rledata := encodeRLE(sprite)
	t.Logf("rlestrings = %+v", rledata)
	t.Logf("%d bytes", len(rledata))

	sprite2 := decodeRLE(rledata)

	for x, _ := range sprite {
		for y, _ := range sprite[x] {
			if sprite[x][y] != sprite2[x][y] {
				t.Fatalf("data at %d,%d is different after decode: %d vs %d", x, y, sprite[x][y], sprite2[x][y])
			}
		}
	}

	//	t.Fail()
}

func TestDecodePredefinedRLE(t *testing.T) {
	var testSprite = "0U1F1QF1Q0Z0Q1FP1FP1Q0Z0Q1F1F1F1Q0Z1Z0Z0QB3B3BQ3B0Z0PBW30Z0PBW30Z0QB9BS30Z0SBS0Z0S121S210Z0P1P21S21P0Z01P21S21P0Z01P21S21P0Z01P21S21P0Z01P21S21P0Z0BP2Q02QBP0Z0PB2Q02QB0Z0R2Q02Q0Z0S2Q02Q0Z0S2Q02Q0Z0S2Q02Q0Z0S2Q02Q0Z0Q1S01T0Z1S01T0U"
	sprite := decodeRLE(testSprite)
	dump(sprite, SpriteBounds{0, 0, 24})
}
