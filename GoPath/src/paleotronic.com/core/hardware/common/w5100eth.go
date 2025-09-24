package common

// W5100RamSize defines std ram size
const W5100RamSize = 32768

// W5100RegID represents Register Id Lookups
const (
	W5100SlotRegIDMode     = 0x04
	W5100SlotRegIDAddrHi   = 0x05
	W5100SlotRegIDAddrLo   = 0x06
	W5100SlotRegIDDataPort = 0x07
)

/*
REG						OFFS    SIZE
Mode Register (MR) 		$0 		1
Gateway Address 		$1 		4
Subnet Mask 			$5 		4
MAC Address 			$9 		6
Source IP Address 		$F 		4
Interrupt (IR) 			$15 	1
Interrupt Mask (IMR) 	$16 	1
Retry Count (RCR) 		$17 	2
RX Memory Size (RMSR) 	$1A 	1
TX Memory Size (TMSR) 	$1B 	1
*/

// W5100Mode represents type mapping for w5100 mode register
type W5100Mode byte

const (
	W5100ModeIndirectBusMode W5100Mode = 1 << iota
	W5100ModeAddressAutoIncrement
	W5100ModeUnused
	W5100ModePPPoEMode
	W5100ModePingBlockMode
	W5100ModeResBit5
	W5100ModeResBit6
	W5100ModeSoftwareReset

	W5100ClearResetMask W5100Mode = W5100ModeIndirectBusMode | W5100ModeAddressAutoIncrement | W5100ModeUnused | W5100ModePPPoEMode | W5100ModePingBlockMode | W5100ModeResBit5 | W5100ModeResBit6
	W5100ModeReset      W5100Mode = W5100ModeIndirectBusMode | W5100ModeAddressAutoIncrement
)

// Register Addresses and sizes
const wModeReg = 0x00
const wGatewayReg = 0x01
const wGatewayRegSize = 0x04
const wSubnetMask = 0x05
const wSubnetMaskSize = 0x04
const wMACAddr = 0x09
const wMACAddrSize = 0x06
const wSourceIP = 0x0f
const wSourceIPSize = 0x04
const wIRQ = 0x15
const wIRQSize = 0x01
const wIRQMask = 0x16
const wIRQMaskSize = 0x01
const wRetryCount = 0x17
const wRetryCountSize = 0x02
const wRXMemSize = 0x1a
const wRXMemSizeSize = 0x01
const wTXMemSize = 0x1b
const wTXMemSizeSize = 0x01

// Base Addresses and sizes
const wTXBase = 0x4000
const wTXSize = 0x2000
const wRXBase = 0x6000
const wRXSize = 0x2000

// W5100EthernetController type is the actual controller
type W5100EthernetController struct {
	RAM     [W5100RamSize]byte
	Address int
}

func NewW5100EthernetController() *W5100EthernetController {
	w := &W5100EthernetController{}
	w.SetMode(W5100ModeReset)
	return w
}

func (w *W5100EthernetController) Stop() {
	// TODO: cleanup code here
}

func (w *W5100EthernetController) GetMode() W5100Mode {
	return W5100Mode(w.RAM[0])
}

func (w *W5100EthernetController) SetMode(m W5100Mode) {
	w.RAM[0] = byte(m)
	if m&W5100ModeSoftwareReset != 0 {
		// reset the device
		w.Reset()
	}
}

func (w *W5100EthernetController) GetAddressHigh() byte {
	return byte((w.Address >> 8) & 0xff)
}

func (w *W5100EthernetController) SetAddressHigh(b byte) {
	w.Address = (w.Address & 0xff00) | int(b)
}

func (w *W5100EthernetController) GetAddressLow() byte {
	return byte(w.Address & 0xff)
}

func (w *W5100EthernetController) SetAddressLow(b byte) {
	w.Address = w.Address | (int(b) << 8)
}

func (w *W5100EthernetController) GetDataPort() byte {
	v := w.RAM[w.Address&0x7fff]
	w.Address++
	return v
}

func (w *W5100EthernetController) SetDataPort(b byte) {
	w.RAM[w.Address&0x7fff] = b
	w.Address++
}

func (w *W5100EthernetController) Reset() {
	// TODO: Reset handler here.
	// we don't reset the address pointers here.

	// clear buffers
	for i := 0x00; i < wTXSize; i++ {
		w.RAM[wTXBase+i] = 0x00
	}
	var rxTest = []byte("this is a test message")
	for i := 0x00; i < wRXSize; i++ {
		if i < len(rxTest) {
			w.RAM[wRXBase+i] = rxTest[i]
		} else {
			w.RAM[wRXBase+i] = 0x00
		}
	}
	w.Address = wRXBase
	// Done: clear reset flag
	w.RAM[0] = w.RAM[0] & byte(W5100ClearResetMask) // clear bit 7
}
