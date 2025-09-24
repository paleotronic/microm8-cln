package woz

import (
	"errors"

	"fmt"
)

var nib53Values = []byte{
	0xab, 0xad, 0xae, 0xaf, 0xb5, 0xb6, 0xb7, 0xba,
	0xbb, 0xbd, 0xbe, 0xbf, 0xd6, 0xd7, 0xda, 0xdb,
	0xdd, 0xde, 0xdf, 0xea, 0xeb, 0xed, 0xee, 0xef,
	0xf5, 0xf6, 0xf7, 0xfa, 0xfb, 0xfd, 0xfe, 0xff,
}

var nib62Values = []byte{
	0x96, 0x97, 0x9a, 0x9b, 0x9d, 0x9e, 0x9f, 0xa6,
	0xa7, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xb2, 0xb3,
	0xb4, 0xb5, 0xb6, 0xb7, 0xb9, 0xba, 0xbb, 0xbc,
	0xbd, 0xbe, 0xbf, 0xcb, 0xcd, 0xce, 0xcf, 0xd3,
	0xd6, 0xd7, 0xd9, 0xda, 0xdb, 0xdc, 0xdd, 0xde,
	0xdf, 0xe5, 0xe6, 0xe7, 0xe9, 0xea, 0xeb, 0xec,
	0xed, 0xee, 0xef, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6,
	0xf7, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff,
}

var unNib62Values, unNib53Values []byte

func init() {
	unNib62Values = make([]byte, 256)
	for i, nib62 := range nib62Values {
		unNib62Values[int(nib62)&0x0ff] = 0x0ff & byte(i)
	}
	unNib53Values = make([]byte, 256)
	for i, _ := range unNib53Values {
		unNib53Values[i] = 0x20
	}
	for i, nib53 := range nib53Values {
		unNib53Values[int(nib53)&0x0ff] = 0x0ff & byte(i)
	}
}

func Decode44(in []byte) byte {
	if len(in) < 2 {
		return 0x00
	}
	b := (in[0] << 1) & 0xaa
	b |= in[1] & 0x55
	return b
}

func Encode44(i byte) []byte {
	a := make([]byte, 2)
	a[0] = (i >> 1) & 0x55
	a[0] |= 0xaa

	a[1] = i & 0x55
	a[1] |= 0xaa
	return a
}

// ExtractNextAddressAndDataNibbles pulls the next sector address and data
// from a nibble stream, assumes standard prologues and epilogues...
func ExtractNextAddressAndDataNibbles(src []byte, start int, prologue []byte) ([]byte, []byte, int) {

	// we assume we are either before or at the addresss prologue
	_, apPos := CollectUntilPattern(src, start, prologue, false)
	if apPos == -1 {
		return nil, nil, -1
	}

	// apPos represents the first byte of the prologue...
	apPos, apEnd := CollectUntilPattern(src, apPos, addrEpilogue, true)
	if apEnd == -1 {
		return nil, nil, -1
	}

	if apEnd-apPos > 32 {
		// could be a problem here, let's try a fallback
		apPosFallback, apEndFallback := CollectUntilPattern(src, apPos, addrEpilogueFallback, true)
		if apEndFallback-apPosFallback < 32 {
			apEnd = apEndFallback + 1
			apPos = apPosFallback
			//fmt.Println("Warning had to resort to fallback address epilogue")
		}
	}

	addressData := src[apPos:apEnd]

	_, dpPos := CollectUntilPattern(src, apEnd, dataPrologue, false)
	if dpPos == -1 {
		return nil, nil, -1
	}

	dpPos, dpEnd := CollectUntilPattern(src, dpPos, dataEpilogue, true)
	if dpEnd == -1 {
		return nil, nil, -1
	}

	sectorData := src[dpPos:dpEnd]

	return addressData, sectorData, dpEnd
}

// DeNibblizeAddress extacts volume, track, sector and checksum from address
// nibbles
func DeNibblizeAddress(address []byte) (volume, track, sector int, checksum byte) {
	// 3 + 2 + 2 + 2 + 2 + 3
	// start 	end		what
	// 0		3		prologue
	// 3 		5 		volume
	// 5		7		track
	// 7		9		sector
	// 9		11		checksum
	// 11		13		epilogue
	if len(address) != 14 {
		// for i, v := range address {
		// 	if i%16 == 0 {
		// 		fmt.Printf("\n%.4x:", i)
		// 	}
		// 	fmt.Printf(" %.2x", v)
		// }
		// fmt.Println()
		return -1, -1, -1, 0x00
	}
	volume = int(Decode44(address[3:5]))
	track = int(Decode44(address[5:7]))
	sector = int(Decode44(address[7:9]))
	checksum = Decode44(address[9:11])
	return volume, track, sector, checksum
}

const (
	DataSize53      = 411 // 410 + checksum
	ChunkSize53     = 51
	ThreeSize       = (ChunkSize53 * 3) + 1
	InvInvalidValue = 0x20
)

// DeNibblizeData converts 411 5+3 encoded bytes (with p/e) into 256 bytes
// of data (sector) and 1 byte (checksum)
func DeNibblizeData13(data []byte) (sector []byte, checksum byte) {

	sector = make([]byte, 256)
	buffer := data[3 : len(data)-4] // remove prologue and epilogues
	checksum = data[len(data)-4]
	var base [256]byte
	var threes [ThreeSize]uint8
	var chksum byte
	var decodedVal byte
	var i int
	var idx int

	/*
	 * Pull the 410 bytes out, convert them from disk bytes to 5-bit
	 * values, and arrange them into a DOS-like pair of buffers.
	 */
	idx = 0
	for i = ThreeSize - 1; i >= 0; i-- {
		decodedVal = unNib53Values[buffer[idx]]
		//fmt.Printf("Decode nib (%.2x) -> %.2x", buffer[idx], decodedVal)
		idx++
		if decodedVal == InvInvalidValue {
			//fmt.Println("got invalid value")
			return
		}
		chksum ^= decodedVal
		threes[i] = chksum
	}

	for i = 0; i < 256; i++ {
		decodedVal = unNib53Values[buffer[idx]]
		//fmt.Printf("Decode nib (%.2x) -> %.2x\n", buffer[idx], decodedVal)
		idx++
		if decodedVal == InvInvalidValue {
			//fmt.Println("got invalid value 2nd loop")
			return
		}
		chksum ^= decodedVal
		base[i] = (chksum << 3)
	}

	/*
	 * Convert this pile of stuff into 256 data bytes.
	 */
	var bufPtr int

	bufPtr = 0
	for i = ChunkSize53 - 1; i >= 0; i-- {
		var three1, three2, three3, three4, three5 byte

		three1 = threes[i]
		three2 = threes[ChunkSize53+i]
		three3 = threes[ChunkSize53*2+i]
		three4 = (three1&0x02)<<1 | (three2 & 0x02) | (three3&0x02)>>1
		three5 = (three1&0x01)<<2 | (three2&0x01)<<1 | (three3 & 0x01)

		sector[bufPtr+0] = base[i] | ((three1 >> 2) & 0x07)
		sector[bufPtr+1] = base[ChunkSize53+i] | ((three2 >> 2) & 0x07)
		sector[bufPtr+2] = base[ChunkSize53*2+i] | ((three3 >> 2) & 0x07)
		sector[bufPtr+3] = base[ChunkSize53*3+i] | (three4 & 0x07)
		sector[bufPtr+4] = base[ChunkSize53*4+i] | (three5 & 0x07)

		bufPtr += 5
	}

	/*
	 * Convert the very last byte, which is handled specially.
	 */
	sector[bufPtr] = base[255] | (threes[ThreeSize-1] & 0x07)

	return sector, checksum

}

// DeNibblizeData converts 343 6+2 encoded bytes (with p/e) into 256 bytes
// of data (sector) and 1 byte (checksum)
func DeNibblizeData(data []byte) (sector []byte, checksum byte) {

	sector = make([]byte, 256)
	source := data[3 : len(data)-4] // remove prologue and epilogues
	checksum = data[len(data)-4]
	temp := make([]byte, 342)
	var last byte

	var current int
	for i := len(temp) - 1; i > 255; i-- {
		t := unNib62Values[0xff&source[current]]
		current++
		temp[i] = t ^ last
		last ^= t
	}
	for i := 0; i < 256; i++ {
		t := unNib62Values[0x0ff&source[current]]
		current++
		temp[i] = t ^ last
		last ^= t
	}
	p := len(temp) - 1
	for i := 0; i < 256; i++ {
		a := (temp[i] << 2)
		a = a + ((temp[p] & 1) << 1) + ((temp[p] & 2) >> 1)
		sector[i] = a
		temp[p] = temp[p] >> 2
		p--
		if p < 256 {
			p = len(temp) - 1
		}
	}

	return sector, checksum
}

const (
	stdTracks             = 35
	stdSectorsPerTrack    = 16
	stdSectorsPerTrack32  = 13
	stdBytesPerSector     = 256
	stdDSKSize            = stdTracks * stdSectorsPerTrack * stdBytesPerSector
	stdDSKSize32          = stdTracks * stdSectorsPerTrack32 * stdBytesPerSector
	std525NibblesPerTrack = 6656
	std525Nibbles         = std525NibblesPerTrack * stdTracks
)

var dskSectorOrder = [stdSectorsPerTrack]int{
	0x00, 0x07, 0x0E, 0x06, 0x0D, 0x05, 0x0C, 0x04,
	0x0B, 0x03, 0x0A, 0x02, 0x09, 0x01, 0x08, 0x0F,
}

var dskSectorOrder32 = [stdSectorsPerTrack32]int{
	0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc,
}

func DeNibblizeImage(data []byte, sectorOrdering []int) (dsk []byte, err error) {

	if sectorOrdering == nil {
		sectorOrdering = dskSectorOrder[:]
	}

	dsk = make([]byte, stdDSKSize)

	if len(data) != std525Nibbles {
		return nil, errors.New("Invalid nibbles length")
	}

	for track := 0; track < stdTracks; track++ {
		//fmt.Printf("Extracting track %d\n", track)
		trackdata := data[track*std525NibblesPerTrack : (track+1)*std525NibblesPerTrack]
		ptr := 0
		sectorsFound := 0
		for ptr != -1 && ptr < std525NibblesPerTrack {
			address, data, nextPos := ExtractNextAddressAndDataNibbles(trackdata, ptr, addrPrologue)
			if address == nil {
				break
			}
			// address decode
			//fmt.Printf("Address block is: %+v\n", address)
			volume, atrack, asector, _ := DeNibblizeAddress(address)
			if volume == -1 {
				return nil, fmt.Errorf("Failed to decode address field on track %d, pos %d", track, ptr)
			}
			if atrack != track {
				return nil, fmt.Errorf("Got track %d, expected %d", atrack, track)
			}
			if asector >= stdSectorsPerTrack {
				return nil, fmt.Errorf("Sector id %d is greater than maximum %d", asector, stdSectorsPerTrack-1)
			}
			//fmt.Printf("--> Track %d, Sector %d, Volume %d\n", atrack, asector, volume)
			// sector data decode
			sectorBytes, _ := DeNibblizeData(data)
			offset := track*stdBytesPerSector*stdSectorsPerTrack + sectorOrdering[asector]*stdBytesPerSector
			for i, b := range sectorBytes {
				dsk[offset+i] = b
			}
			// count
			sectorsFound++
			// move to next
			ptr = nextPos
		}
	}

	return dsk, err
}

func DeNibblizeImage13(data []byte, sectorOrdering []int) (dsk []byte, err error) {

	//if sectorOrdering == nil {
	sectorOrdering = dskSectorOrder32[:]
	//}

	dsk = make([]byte, stdDSKSize32)

	if len(data) != std525Nibbles {
		return nil, errors.New("Invalid nibbles length")
	}

	for track := 0; track < stdTracks; track++ {
		//fmt.Printf("Extracting track %d\n", track)
		trackdata := data[track*std525NibblesPerTrack : (track+1)*std525NibblesPerTrack]
		ptr := 0
		sectorsFound := 0
		for ptr != -1 && ptr < std525NibblesPerTrack {
			address, data, nextPos := ExtractNextAddressAndDataNibbles(trackdata, ptr, addrPrologue13)
			if address == nil {
				break
			}
			// address decode
			//fmt.Printf("Address block is: %+v\n", address)
			volume, atrack, asector, _ := DeNibblizeAddress(address)
			if volume == -1 {
				return nil, fmt.Errorf("Failed to decode address field on track %d, pos %d", track, ptr)
			}
			if atrack != track {
				return nil, fmt.Errorf("Got track %d, expected %d", atrack, track)
			}
			if asector >= stdSectorsPerTrack {
				return nil, fmt.Errorf("Sector id %d is greater than maximum %d", asector, stdSectorsPerTrack-1)
			}
			//fmt.Printf("--> Track %d, Sector %d, Volume %d\n", atrack, asector, volume)
			// sector data decode
			sectorBytes, _ := DeNibblizeData13(data)
			offset := track*stdBytesPerSector*stdSectorsPerTrack32 + sectorOrdering[asector]*stdBytesPerSector
			//fmt.Printf("--> writing block to offset %d\n", offset)
			for i, b := range sectorBytes {
				dsk[offset+i] = b
			}
			// count
			sectorsFound++
			// move to next
			ptr = nextPos
		}
	}

	return dsk, err
}
