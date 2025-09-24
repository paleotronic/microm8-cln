package common

// MOS6522 represents a 6522 Complex Interface Adapter chip
type MOS6522 struct {
	PA  byte
	PB  byte
	CA1 bool
	CA2 bool
	CB1 bool
	CB2 bool
	RES *bool
	IRQ bool

	ORB  byte
	IRB  byte
	ORA  byte
	IRA  byte
	DDRB byte
	DDRA byte
	T1C  uint16
	T1L  uint16
	T2LL byte
	T2C  uint16
	SR   byte
	ACR  byte
	PCR  byte
	IFR  byte
	IER  byte

	irqTarget   Interruptable
	customWrite func(address int, value byte)
	customRead  func(address int) byte
}

// NewMOS6522 creates an instance of the MOS6522
func NewMOS6522(irqTarget Interruptable, cw func(address int, value byte), cr func(address int) byte) *MOS6522 {
	m := &MOS6522{
		irqTarget:   irqTarget,
		customRead:  cr,
		customWrite: cw,
	}
	m.Reset()
	return m
}

// Reset resets the 6522 to power-on state
func (m *MOS6522) Reset() {
	m.PA = 0x00
	m.PB = 0x00
	m.CA1 = true
	m.CA2 = true
	m.CB1 = true
	m.CB2 = true
	m.RES = nil // Input
	m.IRQ = true

	m.ORA = 0xFF
	m.IRA = 0xFF
	m.ORB = 0xFF
	m.IRB = 0xFF
	m.DDRA = 0x00 // All inputs
	m.DDRB = 0x00 // All inputs
	m.T1C = 0xFFFF
	m.T1L = 0xFFFF
	m.T2LL = 0xFF
	m.T2C = 0xFFFF
	m.SR = 0x00
	m.ACR = 0x00
	m.PCR = 0x00
	m.IER = 0x80
	m.IFR = 0x00
}

// Decrement needed to satisfy cycle counter
func (m *MOS6522) Decrement(cycles int) {
	// not used
}

// AdjustClock needed to satisfy cycle counter
func (m *MOS6522) AdjustClock(speed int) {
	// not used
}

// Increment needed to provide timings
func (m *MOS6522) Increment(cycles int) {
	for i := 0; i < cycles; i++ {
		m.UpdateInterrupt()         // Update the IFR register
		if (m.PCR & 0x0C) == 0x08 { // CA2 is output and in pulse mode
			m.CA2 = true // No need to check current value of CA2, just set it to true
		}

		if (m.PCR & 0xE0) == 0xA0 { // CB2 is output and in pulse mode
			m.CB2 = true // No need to check current value of CB2, just set it to true
		}

		if m.T1C != 0 {
			m.T1C--
			if (m.T1C == 0) && (m.IER&0x40 != 0) { // Timer 1 time-out
				m.IFR |= 0x40 // Set 'time-out of timer 1' interrupt flag
			}
		} else {
			if m.ACR&0x40 != 0 { // Timer 1 is in free-run mode
				m.T1C = m.T1L // ...so reload counter
				if m.ACR&0x80 != 0 {
					m.PB ^= 0x80 // Toggle PB7 if T1 is in free-run mode
				}
			} else if m.ACR&0x80 != 0 {
				m.PB |= 0x80 // Set PB7 if T1 is in one-shot mode
			}
		}

		if (m.ACR & 0x20) != 0x20 { // Timer 2 is in timed interrupt mode
			if m.T2C != 0 {
				m.T2C--
				if (m.T2C == 0) && (m.IER&0x20 != 0) { // Timer 2 time-out
					m.IFR |= 0x20 // Set 'time-out of timer 2' interrupt flag
				}
			} else {
				m.T2C = uint16(m.T2LL) // Reload counter
			}
		}
	}
}

func (m *MOS6522) UpdateInterrupt() {
	oirq := m.IRQ
	if m.IFR&(m.IER&0x7F) != 0 {
		//log.Printf("Set bit 7 of IFR")
		m.IFR |= 0x80 // Update top bit of IFR
		m.IRQ = false // ... and the IRQ output
	} else {
		//log.Printf("Clear bit 7 of IFR")
		m.IFR &= 0x7F // Update top bit of IFR
		m.IRQ = true  // ... and the IRQ output
	}
	if oirq != m.IRQ && !m.IRQ {
		//log.Println("IRQ?")
		m.irqTarget.PullIRQLine()
	}
	//	log.Printf("IRQ=%v, IFR=%.8b", m.IRQ, m.IFR)
}

// ReadRegister reads a 6522 register
func (m *MOS6522) ReadRegister(register int) byte {
	var value byte

	// custom handler
	if m.customRead != nil {
		m.customRead(register)
	}

	switch register {
	case 0x00: // Input register B
		value = m.IRB
		m.IFR &= ^byte(0x10)
		if (m.PCR & 0xC0) == 0x80 { // CB2 is output and in handshake or pulse mode
			m.CB2 = false
		}
		break

	case 0x01: // Input register A
		value = m.IRA
		m.IFR &= ^byte(0x02)
		if (m.PCR & 0x0C) == 0x08 { // CA2 is output and in handshake or pulse mode
			m.CA2 = false
		}
		break

	case 0x02: // Data direction register B
		value = m.DDRB
		break

	case 0x03: // Data direction register A
		value = m.DDRA
		break

	case 0x04: // Timer 1 Low-Order Latches
		value = byte(m.T1L & 0x00FF)
		m.IFR &= ^byte(0x40) // Clear Timer 1 bit in IFR
		break

	case 0x05: // Timer 1 High-Order Counter
		value = byte((m.T1C & 0xFF00) >> 8)
		break

	case 0x06: // Timer 1 Low-Order Latches
		value = byte(m.T1L & 0x00FF)
		m.IFR &= ^byte(0x40) // Clear Timer 1 bit in IFR
		break

	case 0x07: // Timer 1 High-Order Latches
		value = byte((m.T1L & 0xFF00) >> 8)
		break

	case 0x08: // Timer 2 Low-Order Latches
		m.IFR &= ^byte(0x20) // Clear Timer 2 bit in IFR
		value = m.T2LL
		break

	case 0x09: // Timer 1 High-Order Counter
		value = byte((m.T1C & 0xFF00) >> 8)
		break

	case 0x0A: // Shift Register
		value = m.SR
		break

	case 0x0B: // Auxiliary Control Register
		value = m.ACR
		break

	case 0x0C: // Pehiperal Control Register
		value = m.PCR
		break

	case 0x0D: // Interrupt Flag Register
		value = m.IFR
		break

	case 0x0E: // Interrupt Enable Register
		value = m.IER | 0x80
		break

	case 0x0F: // Input Register A (No Handshake)
		value = m.IRA
		break
	}

	return value
}

// WriteRegister writes a value to a 6522 register
func (m *MOS6522) WriteRegister(register int, value byte) {
	switch register {
	case 0x00: // Output register B
		m.ORB = value
		m.PB &= ^m.DDRB             // Clear any output bits
		m.PB |= value               // And set them accordingly to the ORB register
		m.IFR &= ^byte(0x10)        // Clear the CB1 flag in the IFR register
		if (m.PCR & 0xC0) == 0x80 { // Is CB2 output and in handshake or pulse mode?
			m.CB2 = false // If so, then generate "Data Taken" on CB2
		}
		break

	case 0x01: // Output register A
		m.ORA = value
		m.PA &= ^m.DDRA             // Clear any output bits
		m.PA |= value               // And set them accordingly to the ORA register
		m.IFR &= ^byte(0x02)        // Clear the CA1 flag in the IFR register
		if (m.PCR & 0x0C) == 0x08 { // Is CA2 output and in handshake or pulse mode?
			m.CA2 = false // If so, then generate "Data Taken" on CA2
		}
		break

	case 0x02: // Data direction register B
		m.DDRB = value
		break

	case 0x03: // Data direction register A
		m.DDRA = value
		break

	case 0x04: // Timer 1 Low-Order Counter
		m.T1C &= 0xFF00
		m.T1C |= uint16(value)
		break

	case 0x05: // Timer 1 High-Order Counter
		m.T1C &= 0x00FF
		m.T1C |= uint16(value) << 8
		m.IFR &= ^byte(0x40) // Clear Timer 1 bit in IFR
		if m.ACR&0x80 != 0 { // If PB7 toggling enabled, then lower PB7 now
			m.PB &= 0x7F
		}
		break

	case 0x06: // Timer 1 Low-Order Latches
		m.T1L &= 0xFF00
		m.T1L |= uint16(value)
		break

	case 0x07: // Timer 1 High-Order Latches
		m.T1L &= 0x00FF
		m.T1L |= uint16(value) << 8
		m.IFR &= ^byte(0x40) // Clear Timer 1 bit in IFR
		break

	case 0x08: // Timer 2 Low-Order Counter
		m.T2C &= 0xFF00
		m.T2C |= uint16(value)
		break

	case 0x09: // Timer 2 High-Order Counter
		m.T2C &= 0x00FF
		m.T2C |= uint16(value) << 8
		m.IFR &= ^byte(0x20) // Clear Timer 2 bit in IFR
		break

	case 0x0A: // Shift Register
		m.SR = value
		break

	case 0x0B: // Auxiliary Control Register
		m.ACR = value
		break

	case 0x0C: // Pehiperal Control Register
		m.PCR = value
		if (value & 0x0E) == 0x0C {
			m.CA2 = false
		}
		if (value & 0x0E) == 0x0E {
			m.CA2 = true
		}
		if (value & 0xE0) == 0xC0 {
			m.CB2 = false
		}
		if (value & 0xE0) == 0xE0 {
			m.CB2 = true
		}
		break

	case 0x0D: // Interrupt Flag Register
		m.IFR = value
		break

	case 0x0E: // Interrupt Enable Register
		if value&0x80 != 0 {
			m.IER |= (value & 0x7F)
		} else {
			m.IER &= ^(value & 0x7F)
		}
		break

	case 0x0F: // Output Register A (No Handshake)
		m.ORA = value
		break
	}

	// custom handler
	if m.customWrite != nil {
		m.customWrite(register, value)
	}

}

// SetPA sets internal state
func (m *MOS6522) SetPA(value byte) {

	m.PA &= m.DDRA            // Preserve the output bits
	m.PA |= (value & ^m.DDRA) // ...and set the input bits

	if (m.ACR & 0x01) != 0x01 { // Is input latching disabled?
		m.IRA = m.PA // If so, also set IRA
	}
}

// SetPB sets internal state
func (m *MOS6522) SetPB(value byte) {

	m.PB &= m.DDRB            // Preserve the output bits
	m.PB |= (value & ^m.DDRB) // ...and set the input bits

	if (m.ACR & 0x02) != 0x02 { // Is input latching disabled?
		m.IRB = m.PB // If so, also set IRB
	}
}

// SetCA1 sets internal state
func (m *MOS6522) SetCA1(value bool) {
	if ((value != m.CA1) && (m.PCR&0x01 != 0) && (value == true)) ||
		((value != m.CA1) && ((m.PCR & 0x01) == 0) && (value == false)) { // ((CA1 positive active edge) || (CA1 negative active edge))
		m.IRA = m.PA  // Latch PA into IRA
		m.IFR |= 0x02 // ...and set flag in the IFR
	}
	m.CA1 = value
}

// SetCA2 sets internal state
func (m *MOS6522) SetCA2(value bool) {
	if (m.PCR & 0x08) == 0x00 { // CA2 must be in input mode to set its value
		if ((value != m.CA2) && ((m.PCR & 0x0C) == 0x04) && (value == true)) ||
			((value != m.CA2) && ((m.PCR & 0x0C) == 0x00) && (value == false)) { // ((CA2 positive active edge) || (CA2 negative active edge))
			m.IFR |= 0x01
		}
		m.CA2 = value
	}
}

// SetCB1 sets internal state
func (m *MOS6522) SetCB1(value bool) {
	if ((value != m.CB1) && (m.PCR&0x10 != 0) && (value == true)) ||
		((value != m.CB1) && (m.PCR&0x01 == 0) && (value == false)) { // ((CB1 positive active edge) || (CB1 negative active edge))
		m.IRB = m.PB  // Latch PB into IRB
		m.IFR |= 0x10 // ...and set flag in the IFR
	}
	m.CB1 = value
}

// SetCB2 sets internal state
func (m *MOS6522) SetCB2(value bool) {
	if (m.PCR & 0x80) == 0x00 { // CB2 must be in input mode to set its value
		if ((value != m.CA1) && ((m.PCR & 0x0C) == 0x04) && (value == true)) ||
			((value != m.CA1) && ((m.PCR & 0x0C) == 0x00) && (value == false)) { // ((CB2 positive active edge) || (CB2 negative active edge))
			m.IFR |= 0x08
		}
		m.CB2 = value
	}
}
