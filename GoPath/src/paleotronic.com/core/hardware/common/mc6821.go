package common

// MC6821 provides an implementation of the motorola 6821 PIA
type MC6821 struct {
	pa   byte
	pb   byte
	ca1  bool
	ca2  bool
	cb1  bool
	cb2  bool
	irqa bool
	irqb bool

	dataa byte
	datab byte
	ddra  byte
	ddrb  byte
	ctrla byte
	ctrlb byte
}

// NewMC6821 creates an instance of the MC6821
func NewMC6821() *MC6821 {
	m := &MC6821{}
	m.Reset()
	return m
}

// Reset resets the MC6821 to power up state
func (m *MC6821) Reset() {
	m.pa = 0x00
	m.pb = 0x00
	m.ca1 = false // CA1 is input only
	m.ca2 = false // CA2 is input or output
	m.cb1 = false // CB1 is input only
	m.cb2 = false // CB2 is input or output
	m.irqa = true // IRQ lines are active low
	m.irqb = true
	m.dataa = 0
	m.datab = 0
	m.ddra = 0
	m.ddrb = 0
	m.ctrla = 0
	m.ctrlb = 0
}

// ReadRegister reads a register in the 6821
func (m *MC6821) ReadRegister(register int) byte {
	var value byte

	switch register {
	case 0x00: // PRA or DDRA
		if m.ctrla&0x04 != 0 {
			value = m.dataa
			m.ctrla &= 0x3F // Clear irqa1 and irqa2 bits in the control register
		} else {
			value = m.ddra
		}
		break

	case 0x01: // CRA
		value = m.ctrla
		break

	case 0x02: // PRB or DDRB
		if m.ctrlb&0x04 != 0 {
			value = m.datab
			m.ctrlb &= ^byte(0x3F) // Clear irqb1 and irqb2 bits in the control register
		} else {
			value = m.ddrb
		}
		break

	case 0x03: // CRB
		value = m.ctrlb
		break
	}
	m.updateIRQ()

	return value
}

// WriteRegister writes a value to a register in the 6821
func (m *MC6821) WriteRegister(register int, value byte) {

	switch register {
	case 0x00: // PRA or DDRA
		if m.ctrla&0x04 != 0 {
			m.dataa &= ^m.ddra          // Mask out input bits
			m.dataa |= (value & m.ddra) // Only set output bits
			m.pa = m.dataa              // Set PA port according to the DATAA register
		} else {
			m.ddra = value
		}
		break

	case 0x01: // CRA
		m.ctrla = (value & 0x3F)
		if (value & 0x30) == 0x30 {
			if value&0x08 != 0 {
				m.ca2 = true
			} else {
				m.ca2 = false
			}
		}
		break

	case 0x02: // PRB or DDRB
		if m.ctrlb&0x04 != 0 {
			m.datab &= ^m.ddrb          // Mask out input bits
			m.datab |= (value & m.ddrb) // Only set output bits
			m.pa = m.dataa              // Set PB port according to the DATAB register
		} else {
			m.ddrb = (value & 0x3F)
		}
		break

	case 0x03: // CRB
		m.ctrlb = value
		if (value & 0x30) == 0x30 {
			if value&0x08 != 0 {
				m.cb2 = true
			} else {
				m.cb2 = false
			}
		}
		break
	}
	m.updateIRQ()
}

// updateIRQ updates the IRQ line state
func (m *MC6821) updateIRQ() {
	if (m.ctrla&0x80 != 0) || (m.ctrlb&0x80 != 0) {
		m.irqa = true
	} else {
		m.irqa = false
	}

	if (m.ctrla&0x40 != 0) || (m.ctrlb&0x40 != 0) {
		m.irqb = true
	} else {
		m.irqb = false
	}
}

// SetPA sets PA internal state
func (m *MC6821) SetPA(value byte) {
	m.dataa &= m.ddra
	m.dataa |= (value & ^m.ddra)
}

// SetPB sets PB internal state (dummied)
func (m *MC6821) SetPB(value byte) {
	// do nothing?
}

// SetCA1 sets CA1 internal state
func (m *MC6821) SetCA1(value bool) {
	if ((value != m.ca1) && (m.ctrla&0x01 != 0) && (m.ctrla&0x02 != 0) && (value == true)) ||
		((value != m.ca1) && (m.ctrla&0x01 != 0) && (m.ctrla&0x02 == 0) && (value == false)) {
		m.ctrla |= 0x80
		if (m.ctrla & 0x28) == 0x20 {
			m.ca2 = true // Read strobe with CA1 restore
		}
	}
	m.ca1 = value
}

// SetCA2 sets CA2 internal state
func (m *MC6821) SetCA2(value bool) {
	if (m.ctrla & 0x20) != 0x20 { // We can only set CA2 if it's selected as an input
		if ((value != m.ca2) && (m.ctrla&0x08 != 0) && (m.ctrla&0x10 != 0) && (value == true)) ||
			((value != m.ca2) && (m.ctrla&0x08 != 0) && (m.ctrla&0x10 == 0) && (value == false)) {
			m.ctrla |= 0x40
		}
		m.ca2 = value
	}
}

// SetCB1 sets CB1 internal state
func (m *MC6821) SetCB1(value bool) {
	if ((value != m.cb1) && (m.ctrlb&0x01 != 0) && (m.ctrlb&0x02 != 0) && (value == true)) ||
		((value != m.cb1) && (m.ctrlb&0x01 != 0) && (m.ctrlb&0x02 == 0) && (value == false)) {
		m.ctrlb |= 0x80
		if (m.ctrlb & 0x28) == 0x20 {
			m.cb2 = true // Read strobe with CB1 restore
		}
	}
	m.cb1 = value
}

// SetCB2 sets CB2 internal state
func (m *MC6821) SetCB2(value bool) {
	if (m.ctrla & 0x20) != 0x20 { // We can only set CA2 if it's selected as an input
		if ((value != m.cb2) && (m.ctrlb&0x08 != 0) && (m.ctrlb&0x10 != 0) && (value == true)) ||
			((value != m.cb2) && (m.ctrlb&0x08 != 0) && (m.ctrlb&0x10 == 0) && (value == false)) {
			m.ctrlb |= 0x40
		}
		m.cb2 = value
	}
}
