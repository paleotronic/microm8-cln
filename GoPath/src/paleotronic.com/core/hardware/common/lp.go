package common

import (
	"os"
	"time"
)

type Parallel interface {
	Close() error
	Write(b []byte) (int, error)
	Stop()
	BufferCount() int
}

type PassThroughParallel struct {
	in        chan byte
	running   bool
	terminate chan bool
	device    string
	file      *os.File
}

func NewPassThroughParallel(device string) (*PassThroughParallel, error) {
	ptp := &PassThroughParallel{
		in:        make(chan byte, 24576),
		running:   false,
		device:    device,
		terminate: make(chan bool),
	}
	f, err := os.OpenFile(device, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0770)
	if err != nil {
		return nil, err
	}
	ptp.file = f
	go ptp.Process()
	return ptp, nil
}

func (d *PassThroughParallel) BufferCount() int {
	return len(d.in)
}

func (d *PassThroughParallel) Write(b []byte) (int, error) {
	for _, bb := range b {
		d.in <- bb
	}
	return len(b), nil
}

func (d *PassThroughParallel) Close() error {
	d.running = false
	if d.file != nil {
		d.file.Close()
	}
	return nil
}

func (d *PassThroughParallel) Stop() {
	d.Close()
}

func (d *PassThroughParallel) flush() {
	for len(d.in) > 0 {
		var ch = <-d.in
		d.file.Write([]byte{ch})
	}
	close(d.in)
}

func (d *PassThroughParallel) Process() {
	d.running = true
	for d.running {
		select {

		case _ = <-d.terminate:
			d.running = false
			return

		case ch := <-d.in:
			d.file.Write([]byte{ch})

		default:
			time.Sleep(1 * time.Second)
		}
	}
	d.flush()
}
