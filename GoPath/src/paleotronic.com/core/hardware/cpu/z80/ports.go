package z80

import (
	"log"
)

func (z *CoreZ80) ReadPort(address uint16) byte {
	log.Printf("Z80 Port read 0x%.4x", address)
	return 0x00
}

func (z *CoreZ80) WritePort(address uint16, b byte) {
	log.Printf("Z80 Port write 0x%.2x -> 0x%.4x", b, address)
}

func (z *CoreZ80) ReadPortInternal(address uint16, contend bool) byte {
	log.Printf("Z80 Port read internal 0x%.4x", address)
	return 0x00
}

func (z *CoreZ80) WritePortInternal(address uint16, b byte, contend bool) {
	log.Printf("Z80 Port write internal 0x%.2x -> 0x%.4x", b, address)
}

func (z *CoreZ80) ContendPortPreio(address uint16) {

}

func (z *CoreZ80) ContendPortPostio(address uint16) {

}
