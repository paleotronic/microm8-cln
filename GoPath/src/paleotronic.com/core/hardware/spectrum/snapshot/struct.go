package snapshot

import (
	"log"
)

func uncompress_z80(data []byte) []byte {

	var buffer = make([]byte, 0, 0x4000)

	idx := 0
	for idx < len(data) && len(buffer) < 0x4000 {
		mem := data[idx] & 0xff
		idx++
		if mem != 0xED {
			buffer = append(buffer, mem)
		} else {
			mem2 := data[idx] & 0xff
			idx++
			if mem2 != 0xED {
				buffer = append(buffer, 0xED)
				buffer = append(buffer, mem2)
			} else {
				nreps := data[idx] & 0xff
				idx++
				value := data[idx] & 0xff
				idx++
				for nreps > 0 {
					nreps--
					buffer = append(buffer, value)
				}
			}
		}

	}

	if len(buffer) < 0x4000 {
		log.Printf("warning: block unpack got only %d bytes, not 16KB", len(buffer))
	}

	return buffer
}
