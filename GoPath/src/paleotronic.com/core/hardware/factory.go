package hardware

import (
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/spectrum"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
)

// FactoryProduce yields an instance of a particular memory mapped region controller
func FactoryProduce(mm *memory.MemoryMap, globalbase int, base int, component string, ent interfaces.Interpretable, misc map[string]map[interface{}]interface{}, options string, ms MachineSpec) memory.Mappable {

	switch component {
	case "apple2iochip":
		return apple2.NewApple2IOChip(mm, globalbase, base, ent, misc, options, int64(ms.CPU.Clocks), int64(ms.CPU.VerticalRetraceCycles), int64(ms.CPU.VBlankCycles), int64(ms.CPU.ScanCycles), int64(ms.CPU.FPS))
	case "apple2zeropage":
		return apple2.NewApple2ZeroPage(mm, globalbase, base, ent)
	case "zxspectrum":
		return spectrum.NewZXSpectrum(mm, globalbase, base, ent, misc)
		// case "shiela":
		// 	return bbc.NewSHIELA(mm, globalbase, base, ent, misc)
	}

	return nil

}
