package disk

import (
	"bytes"
	"io"
)

var NIBBLE_62 = []byte{
	0x96, 0x97, 0x9a, 0x9b, 0x9d, 0x9e, 0x9f, 0xa6,
	0xa7, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xb2, 0xb3,
	0xb4, 0xb5, 0xb6, 0xb7, 0xb9, 0xba, 0xbb, 0xbc,
	0xbd, 0xbe, 0xbf, 0xcb, 0xcd, 0xce, 0xcf, 0xd3,
	0xd6, 0xd7, 0xd9, 0xda, 0xdb, 0xdc, 0xdd, 0xde,
	0xdf, 0xe5, 0xe6, 0xe7, 0xe9, 0xea, 0xeb, 0xec,
	0xed, 0xee, 0xef, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6,
	0xf7, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff}

var prologue62 = []byte{0xd5, 0xaa, 0x96}
var prologue53 = []byte{0xd5, 0xaa, 0xb5}

func (d *DSKWrapper) NibblizeTrack(track int) ([]byte, []byte) {
	var output = bytes.NewBuffer([]byte(nil))
	var s = bytes.NewBuffer([]byte(nil))

	const gap1 = 128
	const gap2 = 7
	const gap3 = 16

	d.writeJunkBytes(output, s, gap1)
	for sector := 0; sector < STD_SECTORS_PER_TRACK; sector++ {
		// Address block - 14 bytes
		volid := d.DOSVolumeID
		if volid == 0 {
			volid = 254
		}
		d.writeAddressBlock(output, s, track, sector, volid, prologue62)
		// gap2
		d.writeJunkBytes(output, s, gap2)
		// Data block
		d.nibblizeBlock62(output, s, track, d.CurrentSectorOrder[sector], d.Data)
		// gap3
		d.writeJunkBytes(output, s, gap3) // 14 + 7 + 342 + 16
	}
	// 0-14 junk bytes
	return output.Bytes(), s.Bytes()
}

func (d *DSKWrapper) Nibblize() []byte {
	if len(d.Data) == STD_DISK_BYTES {
		return d.Nibblize62()
	} else if len(d.Data) == STD_DISK_BYTES_OLD {
		return d.Nibblize53()
	}
	return make([]byte, 232960)
}

func (d *DSKWrapper) Nibblize62() []byte {

	if len(d.Data) != STD_DISK_BYTES {
		return make([]byte, 232960)
	}

	data := d.Data

	output := bytes.NewBuffer([]byte(nil))
	s := bytes.NewBuffer([]byte(nil))

	for track := 0; track < STD_TRACKS_PER_DISK; track++ {
		//d.writeJunkBytes(output, 48);
		for sector := 0; sector < STD_SECTORS_PER_TRACK; sector++ {
			//gap2 := int((rand.Float32() * 5.0) + 4)
			gap2 := 6
			// 15 junk bytes
			d.writeJunkBytes(output, s, 15)
			// Address block - 14 bytes
			volid := d.DOSVolumeID
			if volid == 0 {
				volid = 254
			}
			d.writeAddressBlock(output, s, track, sector, volid, prologue62)
			// 6 junk bytes
			d.writeJunkBytes(output, s, gap2)
			// Data block
			d.nibblizeBlock62(output, s, track, d.CurrentSectorOrder[sector], data)
			// 32 junk bytes
			d.writeJunkBytes(output, s, 38-gap2)
		}
	}

	return output.Bytes()

}

func (d *DSKWrapper) Nibblize53() []byte {

	if len(d.Data) != STD_DISK_BYTES_OLD {
		return make([]byte, 232960)
	}

	data := d.Data

	output := bytes.NewBuffer([]byte(nil))
	s := bytes.NewBuffer([]byte(nil))

	for track := 0; track < STD_TRACKS_PER_DISK; track++ {
		d.writeJunkBytes(output, s, 48)
		for sector := 0; sector < STD_SECTORS_PER_TRACK_OLD; sector++ {
			//gap2 := int((rand.Float32() * 5.0) + 4)
			gap2 := 6
			// 15 junk bytes
			d.writeJunkBytes(output, s, 0)
			// Address block - 14 bytes
			volid := d.DOSVolumeID
			if volid == 0 {
				volid = 254
			}
			d.writeAddressBlock(output, s, track, sector, volid, prologue53)
			// 6 junk bytes
			d.writeJunkBytes(output, s, gap2)
			// Data block
			d.nibblizeBlock53(output, s, track /*d.CurrentSectorOrder[sector]*/, sector, data)
			// 32 junk bytes
			d.writeJunkBytes(output, s, 27)
		}
		d.writeJunkBytes(output, s, 576)
	}

	// f, e := os.Create("out-53-conv.nib")
	// if e == nil {
	// 	f.Write(output.Bytes())
	// 	f.Close()
	// }

	return output.Bytes()

}

func (d *DSKWrapper) NibbleOffsetToTS(offset int) (int, int) {
	offset = offset - (offset % 256)
	c := offset / 256
	sector := c % SECTOR_COUNT
	track := (c - sector) / SECTOR_COUNT
	return track, sector
}

const (
	DataSize53  = 411 // 410 + checksum
	ChunkSize53 = 51
	ThreeSize   = (ChunkSize53 * 3) + 1
)

func (d *DSKWrapper) nibblizeBlock53(output io.Writer, s io.Writer, track, sector int, nibbles []byte) {
	offset := ((track * STD_SECTORS_PER_TRACK_OLD) + sector) * 256

	//	fmt2.Printf("offset to track %d, sector %d is %d", track, sector, offset)

	var top [ChunkSize53*5 + 1]byte    // (255 / 0xff) +1
	var threes [ChunkSize53*3 + 1]byte // (153 / 0x99) +1
	var i, chunk int

	sctBuf := d.Data[offset : offset+256]
	sctPtr := 0

	/*
	 * Split the bytes into sections.
	 */
	chunk = ChunkSize53 - 1
	for i = 0; i < len(top)-1; i += 5 {
		var three1, three2, three3, three4, three5 byte

		three1 = sctBuf[sctPtr+0]
		three2 = sctBuf[sctPtr+1]
		three3 = sctBuf[sctPtr+2]
		three4 = sctBuf[sctPtr+3]
		three5 = sctBuf[sctPtr+4]

		sctPtr += 5

		top[chunk] = three1 >> 3
		top[chunk+ChunkSize53*1] = three2 >> 3
		top[chunk+ChunkSize53*2] = three3 >> 3
		top[chunk+ChunkSize53*3] = three4 >> 3
		top[chunk+ChunkSize53*4] = three5 >> 3

		threes[chunk] = (three1&0x07)<<2 | (three4&0x04)>>1 | (three5&0x04)>>2
		threes[chunk+ChunkSize53*1] = (three2&0x07)<<2 | (three4 & 0x02) | (three5&0x02)>>1
		threes[chunk+ChunkSize53*2] = (three3&0x07)<<2 | (three4&0x01)<<1 | (three5 & 0x01)

		chunk--
	}

	/*
	 * Handle the last byte.
	 */
	var val byte
	val = sctBuf[sctPtr]
	sctPtr++
	top[255] = val >> 3
	threes[ThreeSize-1] = val & 0x07

	output.Write([]byte{0x0d5, 0x0aa, 0x0ad})
	s.Write([]byte{0, 0, 0})
	/*
	 * Write the bytes.
	 */
	var chksum byte
	for i := len(threes) - 1; i >= 0; i-- {
		//assert(threes[i] < sizeof(NIBBLE_53));
		output.Write([]byte{NIBBLE_53[threes[i]^chksum]})
		s.Write([]byte{0})
		chksum = threes[i]
	}

	for i := 0; i < 256; i++ {
		output.Write([]byte{NIBBLE_53[top[i]^chksum]})
		s.Write([]byte{0})
		chksum = top[i]
	}

	//printf("Enc checksum value is 0x%02x\n", chksum);
	output.Write([]byte{NIBBLE_53[chksum]})
	s.Write([]byte{0})

	output.Write([]byte{0x0de, 0x0aa, 0x0eb})
	s.Write([]byte{0, 0, 0})

}

func (d *DSKWrapper) nibblizeBlock62(output io.Writer, s io.Writer, track, sector int, nibbles []byte) {

	//log.Printf("NibblizeBlock(%d, %d)", track, sector)

	offset := ((track * SECTOR_COUNT) + sector) * 256
	temp := make([]int, 342)
	for i := 0; i < 256; i++ {
		temp[i] = int((nibbles[offset+i] & 0x0ff) >> 2)
	}
	hi := 0x001
	med := 0x0AB
	low := 0x055

	for i := 0; i < 0x56; i++ {
		value := ((nibbles[offset+hi] & 1) << 5) |
			((nibbles[offset+hi] & 2) << 3) |
			((nibbles[offset+med] & 1) << 3) |
			((nibbles[offset+med] & 2) << 1) |
			((nibbles[offset+low] & 1) << 1) |
			((nibbles[offset+low] & 2) >> 1)
		temp[i+256] = int(value)
		hi = (hi - 1) & 0x0ff
		med = (med - 1) & 0x0ff
		low = (low - 1) & 0x0ff
	}
	output.Write([]byte{0x0d5, 0x0aa, 0x0ad})
	s.Write([]byte{0, 0, 0})

	last := 0
	for i := len(temp) - 1; i > 255; i-- {
		value := temp[i] ^ last
		output.Write([]byte{NIBBLE_62[value]})
		s.Write([]byte{0})
		last = temp[i]
	}
	for i := 0; i < 256; i++ {
		value := temp[i] ^ last
		output.Write([]byte{NIBBLE_62[value]})
		s.Write([]byte{0})
		last = temp[i]
	}
	// Last data byte used as checksum
	output.Write([]byte{NIBBLE_62[last]})
	s.Write([]byte{0})
	output.Write([]byte{0x0de, 0x0aa, 0x0eb})
	s.Write([]byte{0, 0, 0})
}

func (d *DSKWrapper) writeJunkBytes(output io.Writer, s io.Writer, i int) {
	for c := 0; c < i; c++ {
		output.Write([]byte{0xff})
		s.Write([]byte{0xff})
	}
}

func (d *DSKWrapper) writeAddressBlock(output io.Writer, s io.Writer, track, sector int, volumeNumber int, prologue []byte) {
	output.Write(prologue)
	s.Write([]byte{0, 0, 0})

	var checksum int = 0x00
	// volume
	checksum ^= volumeNumber
	output.Write(d.Encode44(byte(volumeNumber)))
	s.Write([]byte{0, 0})
	// track
	checksum ^= track
	output.Write(d.Encode44(byte(track)))
	s.Write([]byte{0, 0})
	// sector
	checksum ^= sector
	output.Write(d.Encode44(byte(sector)))
	s.Write([]byte{0, 0})
	// checksum
	output.Write(d.Encode44(byte(checksum & 0x0ff)))
	s.Write([]byte{0, 0})

	output.Write([]byte{0xde, 0xaa, 0xeb})
	s.Write([]byte{0, 0, 0})
}

func (d *DSKWrapper) Decode44(in []byte) byte {
	if len(in) < 2 {
		return 0x00
	}
	b := (in[0] << 1) & 0xaa
	b |= in[1] & 0x55
	return b
}

func (d *DSKWrapper) Encode44(i byte) []byte {
	a := make([]byte, 2)
	a[0] = (i >> 1) & 0x55
	a[0] |= 0xaa

	a[1] = i & 0x55
	a[1] |= 0xaa
	return a
}
