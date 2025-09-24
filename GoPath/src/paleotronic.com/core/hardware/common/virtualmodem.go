package common

import (
	"bytes"
	"encoding/csv"
	"regexp"
	"strings"
	"sync"
	"time"

	"paleotronic.com/files"

	"paleotronic.com/fmt"
)

var execBlacklist = []string{
	"rm",
	"cp",
	"parted",
	"del",
	"deltree",
	"mv",
	"move",
	"format",
	"fdisk",
	"diskutil",
	"reboot",
	"shutdown",
	"chmod",
	"chown",
	"sudo",
	"su",
}

type SerialVirtualModem struct {
	Input         []byte
	OutputChunk   chan byte
	Output        chan byte
	lastChars     string
	commandBuffer string
	Device        SerialDevice
	m             sync.Mutex
	r             bool
}

func (d *SerialVirtualModem) IsConnected() bool {

	if d.Device != nil && !d.Device.IsConnected() {
		d.Device = nil
		d.doOutput([]byte("LOST CARRIER"))
	}

	return true
}

func (d *SerialVirtualModem) CanSend() bool {
	return d.IsConnected()
}

func (d *SerialVirtualModem) InputAvailable() bool {
	if len(d.Output) > 0 {
		return true
	}

	//return d.Device != nil && d.Device.InputAvailable()
	return false
}


func (d *SerialVirtualModem) getOutput() byte {
	v := <-d.Output
	return v
}

func (d *SerialVirtualModem) GetInputByte() int {

	if !d.InputAvailable() {
		return 0
	}

	if len(d.Output) > 0 {
		return int(d.getOutput())
	}

	// v := d.Device.GetInputByte()

	// return int(v)
	return 0
}

func (d *SerialVirtualModem) SendOutputByte(value int) {

	if d.Device == nil || !d.Device.IsConnected() {
		// not connected to something
		d.doOutput([]byte{byte(value)})
		if value == 13 {
			if len(d.Input) > 0 {
				d.handleCommand(string(d.Input))
				d.Input = make([]byte, 0)
			}
		} else {
			d.Input = append(d.Input, byte(value))
		}
	} else {
		d.lastChars += string(rune(value))
		if len(d.lastChars) > 10 {
			d.lastChars = d.lastChars[1:]
		}
		if strings.HasSuffix(d.lastChars, "+++") {
			d.Device.Stop()
			d.Device = nil
			d.Drain()
			d.doOutput([]byte(string(rune(27)) + "c" + "\r\nNO CARRIER\r\n"))
			d.lastChars = ""
			time.Sleep(5 * time.Millisecond)
			d.Input = []byte(nil)
			d.SendOutputByte(13)
			return
		}
		d.Device.SendOutputByte(value)
	}

}

func (d *SerialVirtualModem) doOutput(data []byte) {
	for _, v := range data {
		//fmt.Print(string(rune(v)))
		d.OutputChunk <- v
	}
	//d.OutputChunk <- data
}

func (d *SerialVirtualModem) Stop() {
	if d.Device != nil {
		d.Device.Stop()
		d.Device = nil
		d.doOutput([]byte("\r\nSTOPPED\r\n"))
	}
	d.r = false
}

func (d *SerialVirtualModem) handleCommand(c string) {

	out := ""
	for _, ch := range c {
		if ch >= 32 && ch <= 127 {
			out += string(ch)
		}
	}
	c = out

	fmt.Println("got command string:", c)
	c = strings.ToLower(strings.Trim(c, " "))

	if strings.HasPrefix(c, "at") {
		if d.Device != nil {
			d.Device.Stop()
			d.Device = nil
		}

		// drain buffer
		d.Drain()

		if strings.HasPrefix(c, "atc") {
			// arbitrary telnet code
		}

		if strings.HasPrefix(c, "atw") {
			host := "telnet.wmflabs.org"
			port := "23"
			d.Device = NewSerialTelnetDevice(host, port)
			if d.Device.IsConnected() {
				d.doOutput([]byte("\r\nCONNECT " + c + "\r\n"))
				return
			} else {
				d.doOutput([]byte("\r\nCONNECT FAILED\r\n"))
				return
			}
		}

		if strings.HasPrefix(c, "ati") {
			out := []string{
				"Welcome to the microM8 virtual modem!",
				"Commands:",
				"ATL[A-Z]     Available BBSes starting with letter(s).",
				"ATDT[0-999]  Connect to a BBS by number.",
				"+++          Disconnect from BBS.",
				"ATI          Show this information",
			}
			d.doOutput([]byte("\r\n" + strings.Join(out, "\r\n") + "\r\n"))
			return
		}

		var reg = regexp.MustCompile("(?i)^ATD[A-Z]+[ ]*([0-9]+)$")
		if reg.MatchString(c) {
			m := reg.FindAllStringSubmatch(c, -1)
			number := m[0][1]
			b, err := files.ReadBytesViaProvider("/boot/defaults", "bbslist.csv")
			if err != nil {
				d.doOutput([]byte("\r\nNO LIST\r\n"))
				return
			}
			buff := bytes.NewBuffer(b.Content)
			r := csv.NewReader(buff)
			records, err := r.ReadAll()
			if err != nil {
				d.doOutput([]byte("\r\nBAD LIST\r\n"))
				return
			}
			for _, record := range records {
				if len(record) < 3 {
					continue
				}
				if record[0] == number {
					// got it
					address := record[2]
					parts := strings.SplitN(address, ":", 2)
					host := parts[0]
					port := parts[1]
					d.Device = NewSerialTelnetDevice(host, port)
					if d.Device.IsConnected() {
						d.doOutput([]byte("\r\nCONNECT " + c + "\r\n"))
						return
					} else {
						d.doOutput([]byte("\r\nCONNECT FAILED\r\n"))
						return
					}
				}
			}
			d.doOutput([]byte("\r\nNOT FOUND\r\n"))
			return
		}

		if strings.HasPrefix(c, "atl") {
			prefix := strings.ToLower(strings.Trim(c[3:], " "))
			if prefix != "" {
				b, err := files.ReadBytesViaProvider("/boot/defaults", "bbslist.csv")
				if err != nil {
					d.doOutput([]byte("\r\nNO LIST\r\n"))
					return
				}
				buff := bytes.NewBuffer(b.Content)
				r := csv.NewReader(buff)
				records, err := r.ReadAll()
				if err != nil {
					d.doOutput([]byte("\r\nBAD LIST\r\n"))
					return
				}
				for _, record := range records {
					if len(record) < 3 {
						continue
					}
					if strings.HasPrefix(strings.ToLower(record[1]), prefix) {
						s := record[1]
						d.doOutput([]byte("ATDT" + record[0] + " => " + s + "\r\n"))
					}
				}
				return
			}
		}

		if strings.HasPrefix(c, "atx") {
			c := strings.ToLower(strings.Trim(c[3:], " "))
			parts := strings.Split(c, " ")
			for _, word := range execBlacklist {
				if strings.ToLower(parts[0]) == word {
					d.doOutput([]byte("\r\nNOPE\r\n"))
					return
				}
			}

			if d.Device != nil {
				d.Device.Stop()
				d.Device = nil
				// d.doOutput([]byte("\r\nOK\r\n"))
			}

			if strings.ToLower(parts[0]) == "telnet" {
				host := ""
				port := ""
				if len(parts) > 1 {
					host = parts[1]
				}
				if len(parts) > 2 {
					port = parts[2]
				}
				d.Device = NewSerialTelnetDevice(host, port)
				if d.Device.IsConnected() {
					// d.doOutput([]byte("\r\nCONNECT " + c + "\r\n"))
				}
			} else {
				d.Device = NewSerialCommandDevice(c, []string(nil))
				// d.doOutput([]byte("\r\nCONNECT " + c + "\r\n"))
			}
		}

		d.doOutput([]byte("\r\nOK\r\n"))
		return
	}

}

func (d *SerialVirtualModem) Drain() {
	for len(d.OutputChunk) > 0 {
		<-d.OutputChunk
	}
	for len(d.Output) > 0 {
		<-d.Output
	}
}

func NewSerialVirtualModem(str string) *SerialVirtualModem {
	m := &SerialVirtualModem{
		commandBuffer: "",
		Device:        nil,
		Input:         make([]byte, 0),
		Output:        make(chan byte, 1),
		OutputChunk:   make(chan byte, 24576),
		lastChars:     "",
		r:             true,
	}
	if str != "-" {
		m.handleCommand(str)
	}
	go func() {
		for m.r {
			if len(m.Output) == 0 {
				if len(m.OutputChunk) > 0 {
					m.Output <- <-m.OutputChunk
				} else if m.Device != nil && m.Device.InputAvailable() {
					m.OutputChunk <- byte(m.Device.GetInputByte())
				}
				time.Sleep(2 * time.Millisecond)
			} else {
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	return m
}
