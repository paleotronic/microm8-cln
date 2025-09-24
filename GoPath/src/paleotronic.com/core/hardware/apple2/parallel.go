package apple2

import (
	"bytes"
	"log"
	// "io/ioutil"
	"time"

	"paleotronic.com/core/hardware/common"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/fmt"
)

type IOCardParallel struct {
	IOCard
	// data
	//
	Int       interfaces.Interpretable
	lp        common.Parallel
	driver    common.OutputDriver
	lastWrite time.Time
	// in        chan byte
	// terminate chan bool
	store *bytes.Buffer
}

func (d *IOCardParallel) Log(format string, items ...interface{}) {
	log.Printf(d.Name+": "+format, items...)
}

func (d *IOCardParallel) Init(slot int) {
	d.IOCard.Init(slot)
	d.Log("Initialising parallel...")
	// d.in = make(chan byte, 24576) // print buffer
	// d.terminate = make(chan bool)
	if settings.ParallelPassThrough {
		var err error
		d.lp, err = common.NewPassThroughParallel(settings.ParallelLinePrinter)
		if err != nil {
			d.Log("Failed to open %s: %s. Falling back to ESCP emulation.", settings.ParallelLinePrinter, err.Error())
			d.lp = common.NewESCPDevice(&common.PDFOutput{}, d.Int)
		}
	} else {
		d.lp = common.NewESCPDevice(&common.PDFOutput{}, d.Int)
		//d.lp = common.NewImageWriterIIDevice(&common.PDFOutput{}, d.Int)
	}
	d.store = bytes.NewBuffer(nil)
}

func (d *IOCardParallel) Done(slot int) {
	//d.terminate <- true
	if d.lp != nil {
		d.lp.Close()
	}
	// ts := time.Now().Unix()
	// fn := fmt.Sprintf("paralleldump_%d.bin", ts)
	// ioutil.WriteFile(fn, d.store.Bytes(), 0755)
}

func (d *IOCardParallel) HandleIO(register int, value *uint64, eventType IOType) {

	//fmt.RPrintf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))

	switch eventType {
	case IOT_READ:
		switch register {
		case 0x04:
			*value = 0xff
		default:
			fmt.RPrintf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))
		}
	case IOT_WRITE:
		switch register {
		case 0x00:
			//fmt.RPrintf("0x%.2x/", *value)
			d.lp.Write([]byte{byte(*value)})
			d.store.Write([]byte{byte(*value)})
		case 0x02:
			//fmt.RPrintf("0x%.2x/", *value)

		default:
			fmt.RPrintf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))
		}
	}

}

func NewIOCardParallel(mm *memory.MemoryMap, index int, ent interfaces.Interpretable) *IOCardParallel {
	this := &IOCardParallel{}
	this.SetMemory(mm, index)
	this.Int = ent
	//this.SetROM([]uint64{0x24, 0xEA, 0x4C})
	this.Name = "IOCardParallel"

	return this
}
