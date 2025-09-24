package common

import (
	"paleotronic.com/fmt"
)

type DipSwitchBlock struct {
	name  string
	value byte
}

func NewDipSwitchBlock(name string, value byte) *DipSwitchBlock {
	return &DipSwitchBlock{value: value, name: name}
}

func (ds *DipSwitchBlock) SetOff(n int) {
	bit := 8 - n
	ds.value |= (1 << bit)
}

func (ds *DipSwitchBlock) IsOn(n int) bool {
	bit := 8 - n
	return ds.value&(1<<bit) == 0
}

func (ds *DipSwitchBlock) IsOff(n int) bool {
	bit := 8 - n
	return ds.value&(1<<bit) != 0
}

func (ds *DipSwitchBlock) SetOn(n int) {
	bit := 8 - n
	ds.value &= (0xff ^ (1 << bit))
}

func (ds *DipSwitchBlock) Byte() byte {
	return ds.value
}

func (ds *DipSwitchBlock) String() string {
	str := ds.name + ":\n"
	for i := 1; i <= 8; i++ {
		if ds.IsOn(i) {
			str += fmt.Sprintf(" %s-%d: OFF |--O| ON\n", ds.name, i)
		} else {
			str += fmt.Sprintf(" %s-%d: OFF |O--| ON\n", ds.name, i)
		}
	}
	return str
}

func init() {
	sw1 := NewDipSwitchBlock("SW1", 0xff)
	sw1.SetOn(4)
	sw1.SetOn(6)
	sw1.SetOn(7)
	fmt.Println(sw1.String())
	fmt.Printf("0x%.2x\n", sw1.Byte())
	sw2 := NewDipSwitchBlock("SW2", 0x04)
	fmt.Println(sw2.String())
	//os.Exit(0)
}
