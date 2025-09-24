package common

import (
	"paleotronic.com/fmt"
	"paleotronic.com/log"
	"strings"
)

/*
	MOS6551 ACIA Chip implementation
	================================
	Based on spec at: http://archive.6502.org/datasheets/mos_6551_acia.pdf
*/

/*
	CONTROL REGISTER
*/

type mos6551BaudRate int

var baudrates = [16]int{
	115200,
	50,
	75,
	110,
	135,
	150,
	300,
	600,
	1200,
	1800,
	2400,
	3600,
	4800,
	7200,
	9600,
	19200,
}

const (
	br16XExternalClock mos6551BaudRate = iota
	br50
	br75
	br110
	br135
	br150
	br300
	br600
	br1200
	br1800
	br2400
	br3600
	br4800
	br7200
	br9600
	br19200
)

func (br mos6551BaudRate) String() string {
	switch br {
	case br16XExternalClock:
		return "16xExternal"
	case br50:
		return "50 baud"
	case br75:
		return "75 baud"
	case br110:
		return "110 baud"
	case br135:
		return "135 baud"
	case br150:
		return "150 baud"
	case br300:
		return "300 baud"
	case br600:
		return "600 baud"
	case br1200:
		return "1200 baud"
	case br1800:
		return "1800 baud"
	case br2400:
		return "2400 baud"
	case br3600:
		return "3600 baud"
	case br4800:
		return "4800 baud"
	case br7200:
		return "7200 baud"
	case br9600:
		return "9600 baud"
	case br19200:
		return "19200 baud"
	}
	return "Unknown"
}

type mos6551ReceiverClockSource int

const (
	rcsExternalReceiverClock mos6551ReceiverClockSource = iota
	rcsBaudRateGenerator
)

func (rcs mos6551ReceiverClockSource) String() string {
	if rcs != 0 {
		return "Baud rate generator"
	}
	return "External receiver clock"
}

type mos6551DataWordLength int

const (
	dwl8Bit mos6551DataWordLength = iota
	dwl7Bit
	dwl6bit
	dwl5bit
)

func (dwl mos6551DataWordLength) String() string {
	switch dwl {
	case dwl8Bit:
		return "8 bit word"
	case dwl7Bit:
		return "7 bit word"
	case dwl6bit:
		return "6 bit word"
	case dwl5bit:
		return "5 bit word"
	}
	return "Unknown word length"
}

func (dwl mos6551DataWordLength) Mask() int {
	switch dwl {
	case dwl8Bit:
		return 0xff
	case dwl7Bit:
		return 0x7f
	case dwl6bit:
		return 0x3f
	case dwl5bit:
		return 0x1f
	}
	return 0xff
}

type mos6551StopBits int

const (
	sb1StopBit mos6551StopBits = iota
	sb2StopBit
)

func (sb mos6551StopBits) String() string {
	if sb == sb2StopBit {
		return "2 stop bits"
	}
	return "1 stop bit"
}

type mos6551ControlRegister struct {
	stopBits            mos6551StopBits
	dataWordLength      mos6551DataWordLength
	receiverClockSource mos6551ReceiverClockSource
	baudRate            mos6551BaudRate
}

func NewMos6551ControlRegister() *mos6551ControlRegister {
	return &mos6551ControlRegister{
		stopBits:            sb1StopBit,
		dataWordLength:      dwl8Bit,
		receiverClockSource: rcsBaudRateGenerator,
		baudRate:            br9600,
	}
}

func (cr *mos6551ControlRegister) SetValue(v int) {
	cr.stopBits = mos6551StopBits((v >> 7) & 1)
	cr.dataWordLength = mos6551DataWordLength((v >> 5) & 3)
	cr.receiverClockSource = mos6551ReceiverClockSource((v >> 4) & 1)
	cr.baudRate = mos6551BaudRate(v & 15)
}

func (cr mos6551ControlRegister) GetValue() int {
	return int(cr.stopBits)<<7 | int(cr.dataWordLength)<<5 | int(cr.receiverClockSource)<<4 | int(cr.baudRate)
}

func (r *mos6551ControlRegister) Reset() {
	r.SetValue(0x00)
}

func (cr mos6551ControlRegister) String() string {

	return fmt.Sprintf(
		"%s, %s, %s, %s",
		cr.baudRate,
		cr.dataWordLength,
		cr.stopBits,
		cr.receiverClockSource,
	)

}

/*
	COMMAND REGISTER
*/

type mos6551EchoMode int

const (
	emOff mos6551EchoMode = iota
	emOn
)

func (m mos6551EchoMode) String() string {
	if m == emOn {
		return "Echo=On"
	}
	return "Echo=Off"
}

type mos6551DataTerminalReady int

const (
	dtrOff mos6551DataTerminalReady = iota
	dtrOn
)

func (m mos6551DataTerminalReady) String() string {
	if m == dtrOn {
		return "DTR=On"
	}
	return "DTR=Off"
}

type mos6551ReceiverInterruptEnable int

const (
	rieOn mos6551ReceiverInterruptEnable = iota
	rieOff
)

func (m mos6551ReceiverInterruptEnable) String() string {
	if m == rieOn {
		return "rxIRQ=On"
	}
	return "rxIRQ=Off"
}

type mos6551ParityCheckControl int

const (
	pccNoParity    mos6551ParityCheckControl = 0
	pccOddParity   mos6551ParityCheckControl = 1
	pccEvenParity  mos6551ParityCheckControl = 3
	pccMarkParity  mos6551ParityCheckControl = 5
	pccSpaceParity mos6551ParityCheckControl = 7
)

func (p mos6551ParityCheckControl) String() string {
	switch p {
	case pccNoParity:
		return "No parity"
	case pccOddParity:
		return "Odd parity"
	case pccEvenParity:
		return "Even parity"
	case pccMarkParity:
		return "Mark parity"
	case pccSpaceParity:
		return "Space Parity"
	}
	return "Unknown"
}

type mos6551TransmitterControl int

const (
	tcTxDisabledRTSHigh mos6551TransmitterControl = iota
	tcTxEnabledRTSLow
	tcTxDisabledRTSLow
	tcTxDisabledBRK
)

func (t mos6551TransmitterControl) String() string {
	switch t {
	case tcTxDisabledRTSHigh:
		return "TxIRQ=no,RTS=high"
	case tcTxEnabledRTSLow:
		return "TxIRQ=yes,RTS=low"
	case tcTxDisabledRTSLow:
		return "TxIRQ=no,RTS=low"
	case tcTxDisabledBRK:
		return "TxIRQ=no,RTS=low,BRK"
	}
	return "Unknown"
}

type mos6551CommandRegister struct {
	parityCheck      mos6551ParityCheckControl
	echoMode         mos6551EchoMode
	txIRQControl     mos6551TransmitterControl
	rxIRQControl     mos6551ReceiverInterruptEnable
	dataTerminaReady mos6551DataTerminalReady
}

func (r *mos6551CommandRegister) SetValue(v int) {
	r.parityCheck = mos6551ParityCheckControl((v >> 5) & 7)
	r.echoMode = mos6551EchoMode((v >> 4) & 1)
	r.txIRQControl = mos6551TransmitterControl((v >> 2) & 3)
	r.rxIRQControl = mos6551ReceiverInterruptEnable((v >> 1) & 1)
	r.dataTerminaReady = mos6551DataTerminalReady(v & 1)
}

func (r mos6551CommandRegister) GetValue() int {
	return int(r.parityCheck<<5) |
		int(r.echoMode<<4) |
		int(r.txIRQControl<<2) |
		int(r.rxIRQControl<<1) |
		int(r.dataTerminaReady)
}

func (r *mos6551CommandRegister) Reset() {
	r.SetValue(0x02)
}

func (r mos6551CommandRegister) String() string {

	return fmt.Sprintf(
		"%s, %s, %s, %s, %s",
		r.parityCheck,
		r.echoMode,
		r.txIRQControl,
		r.rxIRQControl,
		r.dataTerminaReady,
	)

}

/*
	STATUS REGISTER
*/

type mos6551StatusRegisterFlag int

const (
	// bit 7
	srIRQ mos6551StatusRegisterFlag = 0x80 // IRQ has occurred
	srDSR mos6551StatusRegisterFlag = 0x40 // Data Set Ready
	srDCD mos6551StatusRegisterFlag = 0x20 // Data Carrier Detect
	srTXO mos6551StatusRegisterFlag = 0x10 // Tx OK (buffer empty)
	srRXO mos6551StatusRegisterFlag = 0x08 // Rx OK (buffer full)
	// Error states
	srOVR mos6551StatusRegisterFlag = 0x04 // Overrun Error
	srFRM mos6551StatusRegisterFlag = 0x02 // Framing Error
	srPAR mos6551StatusRegisterFlag = 0x01 // Parity Error
)

type mos6551StatusRegister struct {
	state    mos6551StatusRegisterFlag
	callback func()
}

func (r *mos6551StatusRegister) SetValue(v int) {
	r.state = mos6551StatusRegisterFlag(v & 0xff)
}

func (r mos6551StatusRegister) GetValue() int {
	if r.callback != nil {
		r.callback() // allows dynamic updating of status
	}
	return int(r.state)
}

func (r mos6551StatusRegister) Reset() {
	r.state = 0x10
}

func (r mos6551StatusRegister) ProgramReset() {
	r.state = r.state & 0xfd
}

func (r mos6551StatusRegister) String() string {
	labels := []string{
		"IRQ",
		"DSR",
		"DCD",
		"TX",
		"RX",
		"eOV",
		"eFR",
		"ePA",
	}

	out := []string(nil)
	for i, v := range labels {
		bit := uint(7 - i)
		mask := 1 << bit
		if int(r.state)&mask != 0 {
			out = append(out, v+"=true")
		} else {
			out = append(out, v+"=false")
		}
	}
	return strings.Join(out, ", ")
}

func (r *mos6551StatusRegister) IsSet(flag mos6551StatusRegisterFlag) bool {
	return r.state&flag == flag
}

func (r *mos6551StatusRegister) Set(flag mos6551StatusRegisterFlag) {
	r.state |= flag
}

func (r *mos6551StatusRegister) Clear(flag mos6551StatusRegisterFlag) {
	r.state &= (0xff ^ flag)
}

/*
	MOS6551
*/

type MOS6551TxRx interface {
	Recv() int
	Send(v int)
	CanSend() bool
	HasInput() bool
	ChangeMode(baudRate int, dataBits int, parity string, stopBits string)
	GetStatusBits() (cts bool, dsr bool, dcd bool, ri bool)
}

type MOS6551 struct {
	cycles             int64
	cyclesPerBaudClock int64

	Control mos6551ControlRegister
	Status  mos6551StatusRegister
	Command mos6551CommandRegister
	// device
	device MOS6551TxRx `yaml:"-"`
	// lastValues
	rxByte int
	txByte int
	// pins
	RTS bool // request to send
	//
	IRQ func() `yaml:"-"`
}

func (m *MOS6551) UpdatePort() {
	baudRate := baudrates[int(m.Control.baudRate)]
	dataBits := 8
	switch m.Control.dataWordLength {
	case dwl7Bit:
		dataBits = 7
	case dwl6bit:
		dataBits = 6
	case dwl5bit:
		dataBits = 5
	}
	parity := "N"
	switch m.Command.parityCheck {
	case pccNoParity:
		parity = "N"
	case pccOddParity:
		parity = "O"
	case pccEvenParity:
		parity = "E"
	case pccMarkParity:
		parity = "M"
	case pccSpaceParity:
		parity = "S"
	}
	stopBits := "1"
	switch m.Control.stopBits {
	case sb1StopBit:
		stopBits = "1"
	case sb2StopBit:
		stopBits = "2"
	}
	log.Printf("MOS6551: reconfigure %d,%d,%s,%s", baudRate, dataBits, parity, stopBits)
	m.device.ChangeMode(baudRate, dataBits, parity, stopBits)
}

func (m *MOS6551) Tick() {
	dataBits := 8
	switch m.Control.dataWordLength {
	case dwl7Bit:
		dataBits = 7
	case dwl6bit:
		dataBits = 6
	case dwl5bit:
		dataBits = 5
	}
	stopBits := 1
	switch m.Control.stopBits {
	case sb1StopBit:
		stopBits = 1
	case sb2StopBit:
		stopBits = 2
	}
	bits := 1 + dataBits + stopBits
	m.cyclesPerBaudClock = int64(bits * 1020484 / baudrates[int(m.Control.baudRate)])
	m.cycles++
	if m.cycles > m.cyclesPerBaudClock {
		m.cycles -= m.cyclesPerBaudClock
		if m.device != nil {
			m.update()
		}
	} else if m.cycles % 100 == 0 {
		m.UpdateStatus()
	}
}

func (m *MOS6551) Reset() {
	m.Command.Reset()
	m.Control.Reset()
	m.Status.Reset()
}

func (m *MOS6551) Attach(d MOS6551TxRx, irqf func()) {
	m.device = d
	m.IRQ = irqf
	m.Reset()
}

func (m *MOS6551) Detach() {
	m.device = nil
	m.Reset()
}

func (m *MOS6551) UpdateStatus() {
	var s mos6551StatusRegisterFlag = m.Status.state
	_, dsr, _, _ := m.device.GetStatusBits()
	if dsr {
		s &= (0xff ^ srDSR)
	} else {
		s |= srDSR
	}
	changed := (m.Status.state & srDSR) != (s & srDSR)
	if changed {
		log.Printf("MOS6551: DSR state change %v -> %v", m.Status.state&srDSR == 0, s&srDSR == 0)
	}
	m.Status.SetValue(int(s))
}

func (m *MOS6551) update() {

	// GetStatusBits() (cts bool, dsr bool, dcd bool, ri bool)

	var s mos6551StatusRegisterFlag = m.Status.state

	if m.device != nil {
		dsr := !m.Status.IsSet(srDSR)
		if dsr {
			s &= (0xff ^ srDSR)
		} else {
			s |= srDSR
		}

		//log.Printf("MOS6551: CTS = %v, DSR = %v, DCD = %v, RI = %v", cts, dsr, dcd, ri)

		// If recv empty, but has data, populate data and set srRXO
		if !m.Status.IsSet(srRXO) && m.device.HasInput() && !m.RTS {
			m.rxByte = m.device.Recv()
			s |= srRXO // Receive buffer full
			fmt.Printf("RECV: %.2x : '%s'\n", m.rxByte, string(rune(m.rxByte)))

			if m.Command.rxIRQControl == rieOn && m.IRQ != nil {
				m.IRQ()
			}
		}
		// if tx is full, and can send, send data and set srTXO, RTS must be high not low!!
		// if !m.Status.IsSet(srTXO) && m.RTS {
		// 	if m.device.CanSend() {
		// 		if dsr {
		// 			m.device.Send(m.txByte)
  //
		// 			s |= srTXO // Transmit buffer empty
		// 			m.RTS = false
  //
		// 			if m.Command.txIRQControl == tcTxEnabledRTSLow && m.IRQ != nil {
		// 				m.IRQ()
		// 			}
		// 		}
		// 	}
		// }


		if !m.Status.IsSet(srTXO) && m.device.CanSend() && m.RTS {

			//fmt.Printf("SEND: 0x%.2x\n", m.txByte)

			if dsr {
				m.device.Send(m.txByte)

				s |= srTXO // Transmit buffer empty
				m.RTS = false

				if m.Command.txIRQControl == tcTxEnabledRTSLow && m.IRQ != nil {
					m.IRQ()
				}
			}

		} else {
			s |= srTXO
		}
	} else {
		// clear bits if disconnected
		s |= srDCD
		s |= srDSR
	}

	changed := (m.Status.state & srDSR) != (s & srDSR)
	if changed {
		log.Printf("MOS6551: DSR state change %v -> %v", m.Status.state&srDSR == 0, s&srDSR == 0)
	}

	m.Status.SetValue(int(s))
}

func (m *MOS6551) Send(v int) error {
	//if m.Command.dataTerminaReady != dtrOn {
	//	return nil
	//}
	if m.Status.IsSet(srDSR) {
		log.Printf("trying to send byte while DSR high :(")
		return nil
	}
	if !m.Status.IsSet(srTXO) {
		return nil
	}
	m.txByte = v
	m.Status.Clear(srTXO)
	m.RTS = true
	return nil
}

func (m *MOS6551) Recv() (int, error) {
	//if m.Command.dataTerminaReady != dtrOn {
	//	return m.rxByte, nil
	//}
	v := m.rxByte
	m.Status.Clear(srRXO)
	return v, nil
}

func (m *MOS6551) String() string {

	return fmt.Sprintf(
		`
MOS6551 State:
  Status  : %s
  Control : %s
  Command : %s
  RxBuffer: 0x%.2x
  TxBuffer: 0x%.2x
`,
		m.Status,
		m.Control,
		m.Command,
		m.rxByte,
		m.txByte,
	)

}

func NewMOS6551(control *mos6551ControlRegister) *MOS6551 {
	m := &MOS6551{}
	m.Reset()
	if control != nil {
		m.Control = *control
	}
	m.update()
	fmt.Println(m.String())

	return m
}
