package hardware

type FreezeFile struct {

}

type BlockID byte

const (
	FB_REL BlockID = 1 + iota
	FB_ABS BlockID = 1 + iota
	FB_SPC BlockID = 1 + iota
	FB_HUD BlockID = 1 + iota
	FB_GFX BlockID = 1 + iota
)

// Single memory block
type FreezeBlock struct {
	Class  [3]byte // REL, ABS, SPC, HUD, GFX
	Origin byte
	Start  uint
	Size   uint
	UData  []uint
	BData  []byte
}
