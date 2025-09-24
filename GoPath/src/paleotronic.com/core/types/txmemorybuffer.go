package types

import (
	//"paleotronic.com/log"
	"sync"
)

type TXMemoryBufferChange struct {
	Start int
	End   int
}

type TXMemoryBuffer struct {
	Data   []uint
	mutex  sync.Mutex
	log    []TXMemoryBufferChange
	silent bool
}

// size of buffer
func (this *TXMemoryBuffer) Size() int {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return len(this.Data)
}

// Create an instance of the TX memory buffer
func NewTXMemoryBuffer(size int) *TXMemoryBuffer {
	this := &TXMemoryBuffer{Data: make([]uint, size)}
	return this
}

func (this *TXMemoryBuffer) Silent(s bool) {
	this.mutex.Lock()

	defer this.mutex.Unlock()

	this.silent = s
}

// Set a range of values in the buffer
func (this *TXMemoryBuffer) SetValues(offset int, data []uint) {
	this.mutex.Lock()

	defer this.mutex.Unlock()
	var mx int

	for i := 0; i < len(data); i++ {
		if offset+i >= len(this.Data) {
			break
		}
		this.Data[offset+i] = data[i]
		mx = offset + i
	}

	if !this.silent {
		this.log = append(this.log, TXMemoryBufferChange{Start: offset, End: mx})
	}
}

// Set a single value in the buffer
func (this *TXMemoryBuffer) SetValue(offset int, data uint) {

	this.mutex.Lock()

	defer this.mutex.Unlock()

	//logChange := false

	if offset < len(this.Data) {

		//logChange = (this.Data[offset] != data)

		this.Data[offset] = data
		//log.Printf("Updated %d into addr %d\n", data, offset)
	}

	if !this.silent {
		this.log = append(this.log, TXMemoryBufferChange{Start: offset, End: offset})
	}

}

// Get a range of values from the buffer
func (this *TXMemoryBuffer) GetValues(offset, count int) []uint {
	this.mutex.Lock()

	defer this.mutex.Unlock()

	var data []uint

	if offset >= len(this.Data) {
		return make([]uint, 0)
	}

	if count+offset >= len(this.Data) {
		count = len(this.Data) - offset - 1
	}

	data = this.Data[offset : offset+count+1]
	// this could change on the fly which is bad so we copy the actual values
	newslice := make([]uint, len(data))
	for i, v := range data {
		newslice[i] = v
	}

	return newslice
}

// Return a single value
func (this *TXMemoryBuffer) GetValue(offset int) uint {
	this.mutex.Lock()

	defer this.mutex.Unlock()

	if (offset < 0) || (offset >= len(this.Data)) {
		return 0
	}

	return this.Data[offset]
}

// Fill buffer with a value
func (this *TXMemoryBuffer) Fill(value uint) {
	this.mutex.Lock()

	defer this.mutex.Unlock()

	for i := 0; i < len(this.Data); i++ {
		this.Data[i] = value
	}

	this.log = make([]TXMemoryBufferChange, 0)

	if !this.silent {
		this.log = append(this.log, TXMemoryBufferChange{Start: 0, End: len(this.Data) - 1})
	}

}

func (this *TXMemoryBuffer) GetChangeLog() []TXMemoryBufferChange {
	this.mutex.Lock()

	defer this.mutex.Unlock()
	d := this.log
	this.log = make([]TXMemoryBufferChange, 0)

	return d
}

func (this *TXMemoryBuffer) PurgeLog() {
	this.mutex.Lock()

	defer this.mutex.Unlock()

	this.log = make([]TXMemoryBufferChange, 0)
}

func (this *TXMemoryBuffer) GetChangedData() (int, []uint) {
	this.mutex.Lock()

	defer this.mutex.Unlock()

	var offset int = -1
	var data = make([]uint, 0)

	if len(this.log) == 0 {
		return offset, data
	}

	// we have changes to process
	lo := len(this.Data)
	hi := 0

	l := this.log
	this.log = make([]TXMemoryBufferChange, 0)

	for _, cr := range l {
		if cr.Start < lo {
			lo = cr.Start
		}
		if cr.End > hi {
			hi = cr.End
		}
	}

	data = this.Data[lo : hi+1]

	// this could change on the fly which is bad so we copy the actual values
	newslice := make([]uint, len(data))
	for i, v := range data {
		newslice[i] = v
	}

	return lo, newslice
}
