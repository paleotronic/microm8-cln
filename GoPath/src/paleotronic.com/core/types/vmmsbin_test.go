package types
/*
import (
	   "paleotronic.com/fmt"
	   "testing"
       "paleotronic.com/core/memory"
)

func getMemorySetup() *VarManagerMSBIN {

	 // Setup memory
	 RAM := memory.NewMemoryMap()

     // Setup Applesoft style memory map
     vm := NewVarManagerMSBIN(
     	RAM,
        0,
        105,
        107,
        111,
        109,
        115,
        VUR_QUIET,
     )

     varmem := 2304
     fretop := 38400

     vm.SetVector( vm.VARTAB, varmem )
     vm.SetVector( vm.ARRTAB, varmem )
     vm.SetVector( vm.STREND, varmem+1 )
     vm.SetVector( vm.FRETOP, fretop )
     vm.SetVector( vm.MEMSIZ, fretop )

     return vm
}

func TestSimpleVarExists(t *testing.T) {

	 vm := getMemorySetup()

     // Check if a variable exists
     if vm.Exists("AA") {
     	t.Error("AA should not exist yet")
     }

     // Define the variable
     var e error
     e = vm.Create("A", VT_FLOAT, NewFloat5b(3.142))
     if e != nil {
     	t.Error(e)
     }
     e = vm.Create("BB", VT_FLOAT, NewFloat5b(3.142))
     if e != nil {
     	t.Error(e)
     }
     e = vm.Create("B%", VT_INTEGER, NewInteger2b(3))
     if e != nil {
     	t.Error(e)
     }
     e = vm.CreateString( "a$", "sdgfdgdfgdfgfd" )
     if e != nil {
     	t.Error(e)
     }

     data := make([]uint, 7)
     for i, _ := range data {
     	 data[i] = vm.mm.ReadInterpreterMemory( 0, 2318+i )
     }
     //fmt.Println(data)

     // Check if a variable exists
     if !vm.Exists("B%") {
     	t.Error("B should exist now after vm.Create")
     }

     v, e := vm.GetValue("a$")
     if e != nil {
     	t.Error(e)
     }

     //fmt.Println( v.(*StringPtr3b).FetchString(vm.mm, vm.index) )

     e = vm.CreateIndexed( "XY%", VT_INTEGER, []int{5,10,5}, NewInteger2b(27) )
     if e != nil {
     	t.Error(e)
     }

     e = vm.CreateStringIndexed( "ZZ$", []int{10}, "cat" )
     if e != nil {
     	t.Error(e)
     }

     if !vm.ExistsIndexed("ZZ$") {
     	t.Error("ZZ$ should exist now after vm.CreateIndexed")
     }

     //fmt.Println("Found ZZ$")

     vv, _ := vm.allocStringMemory("cabbage")
     //fmt.Println(vv.GetPointer(), vv.GetLength())

     e = vm.SetValueIndexed( "ZZ$", []int{5}, vv )
     if e != nil {
     	t.Error(e)
     }

     ss, e := vm.GetValueIndexed( "ZZ$", []int{5} )

     vv = ss.(*StringPtr3b)

     if e != nil {
     	t.Error(e)
     }

     //fmt.Println(vv.GetPointer(), vv.GetLength())
     //fmt.Println( "!!!", vv.FetchString(vm.mm, vm.index) )

     //fmt.Println( vm.GetVarNames() )
     //fmt.Println( vm.GetVarNamesIndexed() )

     // force
     //t.Error("end")

     vvv := vm.Get("A")
     if vvv == nil {
	     t.Error("A should exist")
     }

     if vvv.Kind != VT_FLOAT {
	     t.Error("A should be a VT_FLOAT, but is not")
     }

     //fmt.Println(NewFloat5b(0).String())


     t.Error("force")


}

func TestFree(t *testing.T) {

	 vm := getMemorySetup()

	 e := vm.CreateStringIndexed("A$", []int{10}, "cat")
	 if e != nil {
		 t.Error(e)
	 }

	 fbefore := vm.CleanStrings()

	 fafter := vm.CleanStrings()

	 if fafter != fbefore {
		 t.Error( fmt.Sprintf("before (%d) and after (%d) should be the same\n", fbefore, fafter) )
	 }

}
*/
