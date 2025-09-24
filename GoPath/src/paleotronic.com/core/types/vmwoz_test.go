package types

/*

import (
	"paleotronic.com/fmt"
	"testing"

	"paleotronic.com/core/memory"
)

func getMemorySetupW() *VarManagerWOZ {

	// Setup memory
	RAM := memory.NewMemoryMap()

	// Setup Applesoft style memory map
	vm := NewVarManagerWOZ(
		RAM,
		0,
		74,
		204,
		202,
		76,
		VUR_QUIET,
	)

	varmem := 2048
	fretop := 38400

	vm.SetVector(vm.VARBOT, varmem)
	vm.SetVector(vm.VARTOP, varmem)
	vm.SetVector(vm.BASTOP, fretop)
	vm.SetVector(vm.BASBOT, fretop)

	return vm
}

func TestWOZVars(t *testing.T) {

	vm := getMemorySetupW()

	e := vm.Create("A", VT_INTEGER, NewInteger2b(9000))
	if e != nil {
		t.Error(e)
	}
    
	e = vm.CreateIndexed("A$", VT_STRING, []int{3}, "CAT")
	if e != nil {
		t.Error(e)
	}

	e = vm.CreateIndexed("B", VT_INTEGER, []int{5}, NewInteger2b(257))
	if e != nil {
		t.Error(e)
	}

	vm.SetValueIndexed("B", []int{3}, NewInteger2b(9000))

	vv, e := vm.GetValueIndexed("B", []int{3})
	if e != nil {
		t.Error(e)
	}

	n := vv.(*Integer2b).GetValue()
	if n != 9000 {
		t.Error("Expected B(3) to contain 9000 but got", n)
	}

	s := make([]uint, 32)
	for i, _ := range s {
		s[i] = vm.mm.ReadInterpreterMemory(0, 2048+i)
	}
	//fmt.Println("DUMP:", s)

	//fmt.Println(vm.GetVarNamesIndexed())

	//t.Error("done")

}

*/
