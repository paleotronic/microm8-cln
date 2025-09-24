package snapshot

import (
	"fmt"
	"log"
)

type Z80 struct {
	Data           []byte
	CPU            Z80State
	Border         int
	JoystickMode   byte
	Version        int
	Is128K         bool
	Port7ffd       byte
	Portfffd       byte
	Port1ffd       byte
	IF1Paged       bool
	ModifyHardware bool
	REmulation     bool
	AYEmulation    bool
	AYRegs         []byte
	MemPages       map[int][]byte
	CompressedData bool
}

func NewZ80FromData(data []byte) *Z80 {
	return &Z80{
		Data:     data,
		CPU:      Z80State{},
		AYRegs:   []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		MemPages: map[int][]byte{},
	}
}

func (z *Z80) Memory() []byte {
	var out = []byte{}
	pages := []int{8, 4, 5}
	if z.Is128K {
		pages = []int{3, 4, 5, 6, 7, 8, 9, 10}
	}
	for _, page := range pages {
		b, ok := z.MemPages[page]
		if !ok {
			b = make([]byte, 0x4000)
		}
		out = append(out, b...)
	}
	return out
}

func (z *Z80) Load() error {

	z80Header1, err := z.z80HeaderV1()
	if err != nil {
		return err
	}

	z80 := Z80State{}

	z80.RegA = z80Header1[0]
	z80.RegF = z80Header1[1]
	z80.RegC = z80Header1[2]
	z80.RegB = z80Header1[3]
	z80.RegL = z80Header1[4]
	z80.RegH = z80Header1[5]
	z80.RegPC = (uint16(z80Header1[6]) & 0xff) | (uint16(z80Header1[7]) << 8)
	z80.RegSP = (uint16(z80Header1[8]) & 0xff) | (uint16(z80Header1[9]) << 8)
	z80.RegI = z80Header1[10]

	RegR := z80Header1[11] & 0x7f
	if (z80Header1[12] & 0x01) != 0 {
		RegR |= 0x80
	}
	z80.RegR = RegR

	// TODO: border
	z.Border = (int(z80Header1[12]>>1) & 0x07)
	d12 := z80Header1[12]
	if d12 == 0xff {
		d12 = 0x01
	}
	z.CompressedData = d12&0x20 != 0

	z80.RegE = z80Header1[13]
	z80.RegD = z80Header1[14]
	z80.RegCx = z80Header1[15]
	z80.RegBx = z80Header1[16]
	z80.RegEx = z80Header1[17]
	z80.RegDx = z80Header1[18]
	z80.RegLx = z80Header1[19]
	z80.RegHx = z80Header1[20]
	z80.RegAx = z80Header1[21]
	z80.RegFx = z80Header1[22]
	z80.RegIY = (uint16(z80Header1[23]) & 0xff) | (uint16(z80Header1[24]) << 8)
	z80.RegIX = (uint16(z80Header1[25]) & 0xff) | (uint16(z80Header1[26]) << 8)

	if z80Header1[27] != 0 {
		z80.IFF1 = 1
	} else {
		z80.IFF1 = 0
	}

	if z80Header1[28] != 0 {
		z80.IFF2 = 1
	} else {
		z80.IFF2 = 0
	}

	switch z80Header1[29] & 0x03 {
	case 0:
		z80.ModeINT = 0
		break
	case 1:
		z80.ModeINT = 1
		break
	case 2:
		z80.ModeINT = 2
		break
	}

	if (z80Header1[29] & 0x04) != 0 {
		return fmt.Errorf("Issue 2 not supported")
	}

	z.JoystickMode = z80Header1[29] & 0xC0

	z.CPU = z80

	// if the PC is zero, it is a v2/v3 file
	databegin := 30
	if z.CPU.RegPC == 0 {
		hdrLen := (int(z.Data[30]) | (int(z.Data[31]) << 8)) & 0xffff
		switch hdrLen {
		case 23, 54, 55:
			log.Printf("secondary header is %d bytes", hdrLen)
			databegin += hdrLen + 2
			// * 32      2       Program counter
			z.CPU.RegPC = (uint16(z.Data[32]) | (uint16(z.Data[33]) << 8)) & 0xffff
			// * 34      1       Hardware mode (see below)
			z.ModifyHardware = (z.Data[37]&0x80 != 0)
			switch hdrLen {
			case 23:
				log.Printf("Machine id is: %s", v2ModelString(z.Data[34]))
				switch z.Data[34] {
				case 0, 1:
					z.Is128K = false
				case 3, 4:
					z.Is128K = true
				default:
					return fmt.Errorf("Unsupported hardware model (v2) of %s", v2ModelString(z.Data[34]))
				}
			case 54, 55:
				log.Printf("Machine id is: %s", v3ModelString(z.Data[34]))
				switch z.Data[34] {
				case 0, 1:
					z.Is128K = false
				case 4, 5:
					z.Is128K = true
				default:
					return fmt.Errorf("Unsupported hardware model (v3) of %s", v3ModelString(z.Data[34]))
				}
			}
			// * 35      1       If in SamRam mode, bitwise state of 74ls259.
			// 					For example, bit 6=1 after an OUT 31,13 (=2*6+1)
			// 					If in 128 mode, contains last OUT to 0x7ffd
			// 		            If in Timex mode, contains last OUT to 0xf4
			if z.Is128K {
				z.Port7ffd = z.Data[35]
			}
			// * 36      1       Contains 0xff if Interface I rom paged
			// 		             If in Timex mode, contains last OUT to 0xff
			z.IF1Paged = (z.Data[36] == 0xff)
			// * 37      1       Bit 0: 1 if R Register emulation on
			// 					Bit 1: 1 if LDIR emulation on
			// 		Bit 2: AY sound in use, even on 48K machines
			// 		Bit 6: (if bit 2 set) Fuller Audio Box emulation
			// 		Bit 7: Modify hardware (see below)
			z.AYEmulation = (z.Data[37]&0x04 != 0)
			z.REmulation = (z.Data[37]&0x01 != 0)

			// * 38      1       Last OUT to port 0xfffd (soundchip Register number)
			if z.AYEmulation {
				z.Portfffd = z.Data[38]
			}

			// * 39      16      Contents of the sound chip Registers
			if z.AYEmulation {
				z.AYRegs = z.Data[39:55]
			}

			// 	55      2       Low T state counter
			// 	57      1       Hi T state counter
			z.CPU.Tstates = 0 // jspeccy ignores this... (uint16(z.Data[55]) | (uint16(z.Data[56]) << 8)) | (uint16(z.Data[57]) << 16))

			if hdrLen > 23 {

				// VERSION 2 file
				tstate_low := uint(z.Data[55]) | (uint(z.Data[56]) << 8)
				tstate_hi := uint(z.Data[57] & 0x03)
				const T4 = 70908 / 4
				z.CPU.Tstates = int(((tstate_hi-3)%4)*T4 + (T4 - (tstate_low % T4) - 1))

				// 	58      1       Flag byte used by Spectator (QL spec. emulator)
				// 					Ignored by Z80 when loading, zero when saving
				// 	59      1       0xff if MGT Rom paged
				// 	60      1       0xff if Multiface Rom paged. Should always be 0.
				// 	61      1       0xff if 0-8191 is ROM, 0 if RAM
				// 	62      1       0xff if 8192-16383 is ROM, 0 if RAM
				// 	63      10      5 x keyboard mappings for user defined joystick
				// 	73      10      5 x ASCII word: keys corresponding to mappings above
				// 	83      1       MGT type: 0=Disciple+Epson,1=Disciple+HP,16=Plus D
				// 	84      1       Disciple inhibit button status: 0=out, 0ff=in
				// 	85      1       Disciple inhibit flag: 0=rom pageable, 0ff=not
				if hdrLen > 54 {
					// ** 86      1       Last OUT to port 0x1ffd
					z.Port1ffd = z.Data[86]
				}
			}

		default:
			return fmt.Errorf("Unrecognized Z80 header length: %d", hdrLen)
		}
	} else {

		if !z.CompressedData {
			log.Println("v1: not compressed")
			pages := []int{8, 4, 5}
			ptr := 30
			for _, page := range pages {
				if ptr+0x4000 <= len(z.Data) {
					log.Printf("found page %d of length %d at %d", page, 0x4000, ptr)
					z.MemPages[page] = z.Data[ptr : ptr+0x4000]
				}
				ptr += 0x4000
			}
		} else {
			log.Println("v1: compressed")
			d := z80_decompress(z.Data[30 : len(z.Data)-4])
			log.Printf("decompressed is %d bytes", len(d))
			if len(d) != 0xc000 && len(d) != 0x4000 {
				return fmt.Errorf("v1 load incorrect 16/48k compressed length: got %d", len(d))
			}
			z.MemPages[8] = d[0x0000:0x4000]
			if len(d) == 0xc000 {
				z.MemPages[4] = d[0x4000:0x8000]
				z.MemPages[5] = d[0x8000:0xc000]
			} else {
				z.MemPages[4] = make([]byte, 0x4000)
				z.MemPages[5] = make([]byte, 0x4000)
			}
		}

		return nil

	}

	// we got here, try to load data
	ptr := databegin
	log.Printf("starting memory page load from %d", ptr)
	for ptr < len(z.Data)-3 {
		blocklen := int(z.Data[ptr+0]) | (int(z.Data[ptr+1]) << 8)
		length := blocklen
		if blocklen == 0xffff {
			length = 0x4000
		}
		page := int(z.Data[ptr+2])
		log.Printf("found page %d of length %d at %d", page, length, ptr+3)
		ptr += 3
		var block []byte
		if ptr+length > len(z.Data) {
			return fmt.Errorf("file truncated during memory block read")
		}
		if length == 0x4000 {
			block = z.Data[ptr : ptr+length]
		} else {
			block = z80_decompress(z.Data[ptr : ptr+length])
		}
		if len(block) < 0x4000 {
			return fmt.Errorf("only got %d bytes for page %d", len(block), page)
		}
		z.MemPages[page] = block
		ptr += length
	}

	return nil

}

func (z *Z80) z80HeaderV1() ([]byte, error) {
	if len(z.Data) < 30 {
		return nil, fmt.Errorf("Not enough data: got only %d", len(z.Data))
	}
	return z.Data[:30], nil
}

func v2ModelString(b byte) string {
	switch b {
	case 0:
		return "48k"
	case 1:
		return "48k + If.1"
	case 2:
		return "SamRam"
	case 3:
		return "128k"
	case 4:
		return "128k + If.1"
	}
	return "unknown"
}

func v3ModelString(b byte) string {
	switch b {
	case 0:
		return "48k"
	case 1:
		return "48k + If.1"
	case 2:
		return "SamRam"
	case 3:
		return "48k + M.G.T"
	case 4:
		return "128k"
	case 5:
		return "128k + If.1"
	case 6:
		return "128k + M.G.T"
	}
	return "unknown"
}

func z80_decompress(in []byte) []byte {
	// The input is decompressed in 2 phases:
	//  1. Determine output size
	//  2. Decompress

	len_in := len(in)
	i := 0
	j := 0
	for i < len_in {
		if i+4 <= len_in {
			if (in[i+0] == 0xED) && (in[i+1] == 0xED) {
				count := in[i+2]
				j += int(count)
				i += 4
				continue
			}
		}

		i++
		j++
	}

	len_out := j
	out := make([]byte, len_out)

	i = 0
	j = 0
	for i < len_in {
		if i+4 <= len_in {
			if (in[i+0] == 0xED) && (in[i+1] == 0xED) {
				count := in[i+2]
				value := in[i+3]

				for jj := byte(0); jj < count; jj++ {
					out[j] = value
					j++
				}

				i += 4
				continue
			}
		}

		out[j] = in[i]
		i++
		j++
	}

	return out
}
