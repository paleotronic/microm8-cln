package apple2

import (
	"bytes"
	"paleotronic.com/log"
	// "io/ioutil"
	"time"

	"paleotronic.com/fmt"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/common"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"

	yaml "gopkg.in/yaml.v2"
)

const (
	SW1          = 0x01
	SW2_CTS      = 0x02 // Read = Jumper block SW2 and CTS
	ACIA_Data    = 0x08 // Read=Receive / Write=transmit
	ACIA_Status  = 0x09 // Read=Status / Write=Reset
	ACIA_Command = 0x0A
	ACIA_Control = 0x0B
)

type IOCardSSCState struct {
	lastInputByte     int
	SW1Setting        int
	SW2Setting        int
	SW2CTS            int
	RecvIRQEnabled    bool
	TransIRQEnabled   bool
	IRQTriggered      bool
	DTR               bool
	FullEcho          bool
	TransActive       bool
	DataBitPattern    int
	lastControl       int
	RecvIRQ           bool
	Cycles            int
	CyclesPerByte     int
	RecvSuspend       bool
	TxBufferFull      bool
	TxBufferByte      byte
	RxBufferFull      bool
	BaudRate          int
	StopBits          string
	Parity            string
	DataBits          int
	LineConfigChanged bool
}

type IOCardSSC struct {
	IOCard
	// data
	Device  common.SerialDevice
	running bool
	IOCardSSCState
	lastUseServer settings.SSCMode
	byteInterval  time.Duration
	lastRead      time.Time
	//
	commandChar        byte
	commandCharDisable bool
	nextIsCommand      bool

	store *bytes.Buffer

	Int  interfaces.Interpretable
	ACIA *common.MOS6551
}

func (d *IOCardSSC) GetYAML() []byte {
	b, _ := yaml.Marshal(d.IOCardSSCState)
	return b
}

func (d *IOCardSSC) SetYAML(b []byte) {
	_ = yaml.Unmarshal(b, &d.IOCardSSCState)
}

func (d *IOCardSSC) Init(slot int) {
	d.IOCard.Init(slot)
	d.Log("Initialising serial...")
	d.SW1Setting = settings.SSCDipSwitch1
	d.SW2Setting = settings.SSCDipSwitch2
	d.SW2CTS = 0x02
	d.DTR = true
	d.DataBitPattern = 0xff // bit mask
	d.running = true
	d.RecvIRQ = false
	d.TransActive = false
	d.TransIRQEnabled = false
	d.Parity = "N"
	d.TxBufferFull = false
	d.RxBufferFull = false
	d.commandChar = 0x09
	d.CyclesPerByte = 7653
	d.Int.SetCycleCounter(d)
	d.store = bytes.NewBuffer(nil)
	d.ACIA = common.NewMOS6551(common.NewMos6551ControlRegister())
	d.ACIA.Status.SetValue(0x00010000)
	d.configureSerialMode()
	//go d.poll()
}

func (d *IOCardSSC) GetStatusBits() (cts bool, dsr bool, dcd bool, ri bool) {
	if rd, isRaw := d.Device.(*common.SerialPortDevice); isRaw {
		return rd.GetStatusBits()
	}
	return false, d.Device != nil && d.Device.CanSend(), false, false
}

func (d *IOCardSSC) configureSerialMode() {
	log.Printf("SSC: reconfigure mode")
	if settings.SSCCardMode[d.Int.GetMemIndex()] == settings.SSCModeTelnetServer && !settings.IsSetBoolOverride(d.Int.GetMemIndex(), "ssc.disable.telnetserver") {
		d.Device = common.NewSerialTelnetServer(d.Int.GetMemIndex(), "localhost", fmt.Sprintf("%d", 1977+d.Int.GetMemIndex()))
	} else if settings.SSCCardMode[d.Int.GetMemIndex()] == settings.SSCModeVirtualModem {
		d.Device = common.NewSerialVirtualModem(settings.GetModemInitString(d.Int.GetMemIndex())) //&SerialDummyDevice{Data: []byte(" Hello world!")}
	} else if settings.SSCCardMode[d.Int.GetMemIndex()] == settings.SSCModeEmulatedESCP {
		d.Device = common.NewSerialPrinterEmu(
			common.NewESCPDevice(&common.PDFOutput{}, d.Int),
						      32000,
		)
	} else if settings.SSCCardMode[d.Int.GetMemIndex()] == settings.SSCModeEmulatedImageWriter {
		d.Device = common.NewSerialPrinterEmu(
			common.NewImageWriterIIDevice(&common.PDFOutput{}, d.Int),
						      2048,
		)
	} else if settings.SSCCardMode[d.Int.GetMemIndex()] == settings.SSCModeSerialRaw {
		port, err := common.NewSerialPortDevice(settings.SSCHardwarePort, 9600, "N", 8, "1", false)
		if err != nil {
			d.Device = common.NewSerialVirtualModem(settings.GetModemInitString(d.Int.GetMemIndex())) //&SerialDummyDevice{Data: []byte(" Hello world!")}
		} else {
			d.Device = port
		}
	}
	d.ACIA.Attach(d, d.TriggerIRQ)
}

func (d *IOCardSSC) HasInput() bool {
	return d.Device != nil && d.Device.IsConnected() && d.Device.InputAvailable()
}

func (d *IOCardSSC) CanSend() bool {
	return d.Device != nil && d.Device.IsConnected() && d.Device.CanSend()
}

func (d *IOCardSSC) Recv() int {
	if d.Device != nil && d.Device.IsConnected() && d.Device.InputAvailable() {
		return d.Device.GetInputByte()
	}
	return 0
}

func (d *IOCardSSC) Send(v int) {
	if d.Device != nil && d.Device.IsConnected() {
		d.Device.SendOutputByte(v)
	}
}

func (d *IOCardSSC) ChangeMode(baudRate int, dataBits int, parity string, stopBits string) {
	if rd, isRaw := d.Device.(*common.SerialPortDevice); isRaw {
		rd.ChangeMode(baudRate, dataBits, parity, stopBits)
	}
}

func (d *IOCardSSC) ImA() string {
	return "SSC"
}

func (d *IOCardSSC) Increment(n int) {
	d.Cycles += n
	d.poll()
	d.ACIA.Tick()
}

func (d *IOCardSSC) Decrement(n int) {

}

func (d *IOCardSSC) AdjustClock(n int) {

}

func (d *IOCardSSC) Done(slot int) {
	// dummy
	d.running = false
	if d.Device != nil {
		d.Device.Stop()
		d.Device = nil
	}
	// ts := time.Now().Unix()
	// fn := fmt.Sprintf("serialdump_%d.bin", ts)
	// ioutil.WriteFile(fn, d.store.Bytes(), 0755)
}

func (d *IOCardSSC) TriggerIRQ() {
	//fmt.Println("IRQ Triggered")
	cpu := apple2helpers.GetCPU(d.Int)
	cpu.PullIRQLine()
	d.IRQTriggered = true
}

func (d *IOCardSSC) Connected() bool {
	return d.Device != nil
}

func (d *IOCardSSC) inputAvail() bool {
	return !d.RecvSuspend && d.Device != nil && d.Device.InputAvailable() //&& d.Cycles >= d.CyclesPerByte
}

func (d *IOCardSSC) poll() {

	// for d.running {
	if d.inputAvail() {
		// we gots some inputs -- haha
		if d.RecvIRQEnabled {
			//d.RecvIRQ = true
			// Notify we have data
			d.TriggerIRQ()
		}
	}
	if d.lastUseServer != settings.SSCCardMode[d.Int.GetMemIndex()] {
		if d.Device != nil {
			d.Device.Stop()
		}
		d.configureSerialMode()
		d.lastUseServer = settings.SSCCardMode[d.Int.GetMemIndex()]
	}
	// }

}

func (d *IOCardSSC) controlString() string {
	return fmt.Sprintf("%d,%d%s%s", d.BaudRate, d.DataBits, d.Parity, d.StopBits)
}

func (d *IOCardSSC) HandleIO(register int, value *uint64, eventType IOType) {

	switch eventType {
	case IOT_READ:

		switch register {
		case 0x00:

		case SW1:
			d.SW1Read(value)

		case SW2_CTS:
			d.SW2CTSRead(value)

		case 0x03:

		case 0x04:

		case 0x05:

		case 0x06:

		case 0x07:

		case ACIA_Data:
			v, _ := d.ACIA.Recv()
			*value = uint64(v)
			//d.DataRead(value)
			//log.Printf("SSC read: $%2x", *value&0xff)
			//d.RecvIRQ = false

		case ACIA_Status:

			d.ACIA.UpdateStatus()
			*value = uint64(d.ACIA.Status.GetValue())
			//if !d.ACIA.RTS {
			//	*value = 0x10
			//} else {
			//	*value = 0x00
			//}
			//if d.ACIA.Status.IsSet(0x08) {
			//	*value |= 0x08
			//}
			//log.Printf("SSC: read status of 0x%.8b", *value)
			//d.StatusRead(value)

		case ACIA_Command:
			*value = uint64(d.ACIA.Command.GetValue())
			//d.CommandRead(value)

		case ACIA_Control:
			*value = uint64(d.ACIA.Control.GetValue())
			//d.ControlRead(value)

		case 0x0C:

		case 0x0D:

		case 0x0E:

		case 0x0F:

		default:
			//fmt.RPrintf("SSC tried to read register 0x%.2x\n", register)

		}

		//fmt.Printf("Will return %d\n", *value)

		break

	case IOT_WRITE:
		switch register {
		case 0x00:

		case 0x01:

		case 0x02:

		case 0x03:

		case 0x04:

		case 0x05:

		case 0x06:

		case 0x07:

		case ACIA_Data:
			//log.Printf("SSC: write of byte 0x%.2X", byte(*value))
			//d.DataWrite(value)
			if d.nextIsCommand {
				switch byte(*value) {
				case 'Z':
					d.commandCharDisable = true
				}
				d.nextIsCommand = false
			} else {
				if !d.commandCharDisable && byte(*value) == d.commandChar {
					d.nextIsCommand = true
				} else {
					d.ACIA.Send(int(*value))
				}
			}

		case ACIA_Status:
			//d.StatusRead(value)
			d.ACIA.Status.ProgramReset()

		case ACIA_Command:

			//d.CommandWrite(value)
			d.ACIA.Command.SetValue(int(*value))
			d.ACIA.UpdatePort()
			//d.ACIA.Reset()
			log.Printf("SSC: write of COMMAND byte 0x%.8b", byte(*value))
		case ACIA_Control:
			d.ACIA.Control.SetValue(int(*value))
			d.ACIA.UpdatePort()
			//d.ControlWrite(value)
			log.Printf("SSC: write of CONTROL byte 0x%.8b", byte(*value))
		case 0x0C:

		case 0x0D:

		case 0x0E:

		case 0x0F:

		default:
			log.Printf("SSC: tried to write register 0x%.2d -> 0x%.2x\n", *value, register)
		}
		break

	}

	//log.Printf("IOCardSSC: %s of register %.2x (value == 0x%.8bb / %.2x)\n", eventType.String(), register, *value, *value)

}

func (d *IOCardSSC) SW1Read(value *uint64) {
	if settings.SSCDipSwitch1 != d.SW1Setting {
		d.SW1Setting = settings.SSCDipSwitch1
	}
	*value = uint64(d.SW1Setting)
}

func (d *IOCardSSC) SW2CTSRead(value *uint64) {
	if settings.SSCDipSwitch2 != d.SW2Setting {
		d.SW2Setting = settings.SSCDipSwitch2
	}

	*value = uint64(d.SW2Setting) & 0x0FE
	// if port is connected and ready to send another byte, set CTS bit on
	if d.Connected() && d.inputAvail() {
		*value |= 0x00
	} else {
		*value |= 0x01
	}
}

func NewIOCardSSC(mm *memory.MemoryMap, index int, ent interfaces.Interpretable) *IOCardSSC {
	this := &IOCardSSC{}
	this.SetMemory(mm, index)
	this.Int = ent
	this.Name = "IOCardSSC"

	return this
}
