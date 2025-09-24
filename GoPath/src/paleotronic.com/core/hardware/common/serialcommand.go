package common

import (
	"io"
	"os/exec"
	"time"

	"paleotronic.com/fmt"
)

type SerialCommandDevice struct {
	Command string
	Args    []string
	child   *Executor
}

func (d *SerialCommandDevice) Stop() {
	if d.child != nil {
		d.child.Stop()
		d.child = nil
	}
}

func (d *SerialCommandDevice) CanSend() bool {
	return d.IsConnected()
}

func (d *SerialCommandDevice) IsConnected() bool {
	return d.child != nil && d.child.r
}

func (d *SerialCommandDevice) InputAvailable() bool {
	return len(d.child.fromCommand) > 0
}

func (d *SerialCommandDevice) GetInputByte() int {

	if !d.InputAvailable() {
		return 0
	}

	v := <-d.child.fromCommand

	return int(v)
}

func (d *SerialCommandDevice) SendOutputByte(value int) {
	d.child.toCommand <- byte(value)
}

func NewSerialCommandDevice(command string, args []string) *SerialCommandDevice {

	s := &SerialCommandDevice{
		Command: command,
		Args:    args,
		child:   NewExecutor(command),
	}

	s.child.Start()

	fmt.Println("Running command", command)

	return s

}

type Executor struct {
	fromCommand chan byte
	toCommand   chan byte
	cmd         *exec.Cmd
	r           bool
	echo        bool
	crToLF      bool
	addCRtoLF   bool
}

func (e *Executor) Start() {
	if e.r {
		return
	}
	e.r = true
	stdout, _ := e.cmd.StdoutPipe()
	stdin, _ := e.cmd.StdinPipe()
	fmt.Print(e.cmd.Start())

	go func(out io.ReadCloser) {

		fmt.Println("starting reader for stdout")

		buff := make([]byte, 8)

		for e.r {

			fmt.Println("b4 read")
			n, err := out.Read(buff)
			fmt.Printf("read %d bytes, %v\n", n, err)
			if err != nil {
				e.r = false
				e.fromCommand <- byte('E')
				e.fromCommand <- byte('O')
				e.fromCommand <- byte('F')
				e.fromCommand <- byte('\r')
				e.fromCommand <- byte('\n')
				return
			}
			fmt.Println("after read")

			if n > 0 {
				for _, v := range buff[0:n] {
					if e.addCRtoLF && v == 10 {
						e.fromCommand <- 13
					}
					fmt.Printf("recv: %.2x\n", v)
					e.fromCommand <- v
					time.Sleep(time.Millisecond * 3)
				}
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}

	}(stdout)

	go func(in io.WriteCloser) {

		for e.r {
			select {
			case b := <-e.toCommand:

				if e.echo {
					e.fromCommand <- b
				}

				if e.crToLF && b == 13 {
					b = 10
				}

				fmt.Printf("send: %.2x\n", b)
				_, err := in.Write([]byte{b})
				fmt.Printf("Write %v\n", err)
				if err != nil {
					e.r = false
					e.fromCommand <- byte('E')
					e.fromCommand <- byte('O')
					e.fromCommand <- byte('F')
					e.fromCommand <- byte('\r')
					e.fromCommand <- byte('\n')
					return
				}
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}

	}(stdin)
}
