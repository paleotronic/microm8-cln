// +build !nox

package core

func (vm *VM) PreInitOverrides() {

	// log.Printf("memory init pattern")

	for i := 0; i < 65536; i++ {
		v := ((i / 2) % 2) * 0xff
		vm.RAM.WriteInterpreterMemorySilent(vm.Index, i, uint64(v))
		vm.RAM.WriteInterpreterMemorySilent(vm.Index, i+65536, uint64(v))
	}

}
