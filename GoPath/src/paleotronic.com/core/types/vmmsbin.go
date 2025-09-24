package types

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"sync"

	"paleotronic.com/fmt"

	"paleotronic.com/log"

	"paleotronic.com/core/memory"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type StrAllocMap map[int]StringPtr3b
type StrAllocPointers map[StringPtr3b][]int // addresses containing a copy of this pointer

// Response codes for handling undefined variable
type VarUndefinedResponse int

const (
	VUR_QUIET  VarUndefinedResponse = 0
	VUR_ERROR  VarUndefinedResponse = 1
	VUR_CREATE VarUndefinedResponse = 2
)

const (
	VENTRY_SIZE = 7
)

// VarManagerMSBIN stores microsoft basic style variable table
type VarManagerMSBIN struct {
	// Pointers to the interpreter reference
	mm    *memory.MemoryMap
	index int
	// Data pointers
	VARTAB int
	ARRTAB int
	FRETOP int
	STREND int
	MEMSIZ int
	// configs
	UndefResponse VarUndefinedResponse
	Cache         map[[2]byte]int
	CacheIndexed  map[[2]byte]int
	UseCache      bool
	m             sync.Mutex
}

// NewVarManagerMSBIN creates a new microsoft basic compatible var manager
func NewVarManagerMSBIN(mm *memory.MemoryMap, index int, vt, at, ft, se, ms int, uresp VarUndefinedResponse) *VarManagerMSBIN {

	this := &VarManagerMSBIN{
		mm:            mm,
		index:         index,
		VARTAB:        vt,
		ARRTAB:        at,
		FRETOP:        ft,
		STREND:        se,
		MEMSIZ:        ms,
		UndefResponse: uresp,
		Cache:         make(map[[2]byte]int),
		CacheIndexed:  make(map[[2]byte]int),
		UseCache:      true,
	}

	return this

}

func (vm *VarManagerMSBIN) GetIndex() int {
	return vm.index
}

func (vm *VarManagerMSBIN) GetMM() *memory.MemoryMap {
	return vm.mm
}

// Exists returns true if a given variable name exists
func (vm *VarManagerMSBIN) Exists(name string) bool {
	a, _, _ := vm.getAddressOfVar(name)
	return (a != -1)
}

func (vm *VarManagerMSBIN) Contains(name string) bool {
	a, _, _ := vm.getAddressOfVar(name)
	b, _, _ := vm.getAddressOfArray(name)
	return (a != -1) || (b != -1)
}

func (vm *VarManagerMSBIN) ContainsKey(name string) bool {
	return vm.Contains(name)
}

func (vm *VarManagerMSBIN) Dimensions(name string) []int {

	if vm.Exists(name) {
		return []int{1}
	}

	addr, _, _ := vm.getAddressOfArray(name)
	var msbin *MSBINArrayRecord = &MSBINArrayRecord{}

	msbin.ReadMemory(vm.mm, vm.index, addr)

	return msbin.DimData

}

func (vm *VarManagerMSBIN) Put(name string, v *Variable) {

	if v.IsArray() {
		// array
	} else {
		// scalar
	}

}

func (vm *VarManagerMSBIN) Get(name string) *Variable {

	if vm.Exists(name) {

		addr, _, _ := vm.getAddressOfVar(name)
		var msbin *MSBINRecord = &MSBINRecord{}
		msbin.ReadMemory(vm.mm, vm.index, addr)

		return &Variable{
			Name:           name,
			Driver:         VD_MICROSOFT,
			Context:        VDC_SCALAR,
			Map:            vm,
			Kind:           msbin.GetType(),
			AssumeLowIndex: true,
			Owner:          "",
		}

	} else if vm.ExistsIndexed(name) {

		addr, _, _ := vm.getAddressOfArray(name)
		var msbin *MSBINArrayRecord = &MSBINArrayRecord{}
		msbin.ReadMemory(vm.mm, vm.index, addr)

		return &Variable{
			Name:           name,
			Driver:         VD_MICROSOFT,
			Context:        VDC_ARRAY,
			Map:            vm,
			Kind:           msbin.GetType(),
			AssumeLowIndex: true,
			Owner:          "",
		}

	}

	return nil
}

func (vm *VarManagerMSBIN) GetValue(name string) (interface{}, error) {

	var msbin *MSBINRecord

	addr, vt, _ := vm.getAddressOfVar(name)
	if addr == -1 {
		// var exists
		switch vm.UndefResponse {
		case VUR_ERROR:
			return nil, errors.New("NO SUCH VARIABLE")
		case VUR_QUIET:
			switch vt {
			case VT_FLOAT:
				return NewFloat5b(0), nil
			case VT_INTEGER:
				return NewInteger2b(0), nil
			case VT_EXPRESSION:
				return NewFuncPtr5b(0, 0, 0), nil
			case VT_STRING:
				return NewStringPtr3b(0, 0), nil
			}
		case VUR_CREATE:
			switch vt {
			case VT_FLOAT:
				e := vm.Create(name, VT_FLOAT, NewFloat5b(0))
				if e != nil {
					return nil, e
				}
				addr, vt, _ = vm.getAddressOfVar(name)
			case VT_INTEGER:
				e := vm.Create(name, VT_FLOAT, NewInteger2b(0))
				if e != nil {
					return nil, e
				}
				addr, vt, _ = vm.getAddressOfVar(name)
			case VT_EXPRESSION:
				e := vm.Create(name, VT_FLOAT, NewFuncPtr5b(0, 0, 0))
				if e != nil {
					return nil, e
				}
				addr, vt, _ = vm.getAddressOfVar(name)
			case VT_STRING:
				ptr, e := vm.allocStringMemory("")
				if e != nil {
					return nil, e
				}
				e = vm.Create(name, VT_FLOAT, ptr)
				if e != nil {
					return nil, e
				}
				addr, vt, _ = vm.getAddressOfVar(name)
			}
		}
	}

	msbin = &MSBINRecord{}
	msbin.ReadMemory(vm.mm, vm.index, addr)

	switch vt {
	case VT_FLOAT:
		return msbin.GetFloatValue(), nil
	case VT_INTEGER:
		return msbin.GetIntValue(), nil
	case VT_STRING:
		return msbin.GetStringPointer(), nil
	case VT_EXPRESSION:
		return msbin.GetFuncPointer(), nil
	}

	return nil, errors.New("ERROR")

}

func (vm *VarManagerMSBIN) SetValue(name string, content interface{}) error {
	// first does it already exist
	var msbin *MSBINRecord = &MSBINRecord{}

	////fmt.Printf("In SetValue(%s, %v)\n", name, content)

	addr, _, _ := vm.getAddressOfVar(name)
	if addr == -1 {
		// var exists
		return errors.New("SYNTAX ERROR")
	}

	msbin.ReadMemory(vm.mm, vm.index, addr)

	//	//fmt.Printf("ReadMemory returns %s\n", msbin.String())

	vt := msbin.GetType()

	/* account for 2 byte name header */
	switch vt {
	case VT_FLOAT:
		f := content.(*Float5b)
		f.WriteMemory(vm.mm, vm.index, addr+2)
		return nil
	case VT_INTEGER:
		f := content.(*Integer2b)
		f.WriteMemory(vm.mm, vm.index, addr+2)
		return nil
	case VT_STRING:
		f := content.(*StringPtr3b)
		f.WriteMemory(vm.mm, vm.index, addr+2)
		return nil
	}

	return errors.New("SYNTAX ERROR")
}

func (vm *VarManagerMSBIN) ExistsIndexed(name string) bool {
	a, _, _ := vm.getAddressOfArray(name)
	return (a != -1)
}

func (vm *VarManagerMSBIN) GetValueIndexed(name string, index []int) (interface{}, error) {

	// first does it already exist
	var msbin *MSBINArrayRecord = &MSBINArrayRecord{}

	addr, vt, _ := vm.getAddressOfArray(name)
	if addr == -1 {
		// var exists
		switch vm.UndefResponse {
		case VUR_ERROR:
			return nil, errors.New("NO SUCH VARIABLE")
		case VUR_QUIET:
			switch vt {
			case VT_FLOAT:
				return NewFloat5b(0), nil
			case VT_INTEGER:
				return NewInteger2b(0), nil
			case VT_STRING:
				return NewStringPtr3b(0, 0), nil
			}
		case VUR_CREATE:
			switch vt {
			case VT_FLOAT:
				e := vm.CreateIndexed(name, VT_FLOAT, []int{10}, NewFloat5b(0))
				if e != nil {
					return nil, e
				}
				addr, vt, _ = vm.getAddressOfVar(name)
			case VT_INTEGER:
				e := vm.CreateIndexed(name, VT_FLOAT, []int{10}, NewInteger2b(0))
				if e != nil {
					return nil, e
				}
				addr, vt, _ = vm.getAddressOfVar(name)
			case VT_STRING:
				ptr, e := vm.allocStringMemory("")
				if e != nil {
					return nil, e
				}
				e = vm.CreateIndexed(name, VT_FLOAT, []int{10}, ptr)
				if e != nil {
					return nil, e
				}
				addr, vt, _ = vm.getAddressOfVar(name)
			}
		}
	}

	msbin.ReadMemory(vm.mm, vm.index, addr)

	return msbin.GetRecordAtIndex(vm.mm, vm.index, index)
}

// Set the value at at a given index
func (vm *VarManagerMSBIN) SetValueIndexed(name string, index []int, value interface{}) error {
	// first does it already exist
	var msbin *MSBINArrayRecord = &MSBINArrayRecord{}

	addr, _, _ := vm.getAddressOfArray(name)
	if addr == -1 {
		// var exists
		return errors.New("BAD SUBSCRIPT")
	}

	msbin.ReadMemory(vm.mm, vm.index, addr)

	return msbin.SetRecordAtIndex(vm.mm, vm.index, index, value)
}

// GetVarNames returns a list of variable names
func (vm *VarManagerMSBIN) GetVarNames() []string {
	varptr := vm.GetVector(vm.VARTAB) // start of VARTAB
	arrptr := vm.GetVector(vm.ARRTAB) // end of VARTAB

	found := false

	names := make([]string, 0)
	msbin := &MSBINRecord{}

	for varptr < arrptr && !found {

		msbin.ReadMemory(vm.mm, vm.index, varptr)
		names = append(names, msbin.GetName())
		varptr += 7

	}

	return names
}

// GetIndexedVarNames returns a list of variable names
func (vm *VarManagerMSBIN) GetVarNamesIndexed() []string {

	arrptr := vm.GetVector(vm.ARRTAB) // start of array
	strend := vm.GetVector(vm.STREND) // end of array

	found := false

	names := make([]string, 0)
	msbin := &MSBINArrayRecord{}

	for arrptr < strend-1 && !found {

		msbin.ReadMemory(vm.mm, vm.index, arrptr)
		names = append(names, msbin.GetName())

		offset := msbin.OffsetNext
		if offset < 5 {
			found = true
		}
		arrptr += int(offset)

	}

	return names
}

func (vm *VarManagerMSBIN) extendVarMemory(size int) error {

	arrtab := vm.GetVector(vm.ARRTAB)
	strend := vm.GetVector(vm.STREND)
	fretop := vm.GetVector(vm.FRETOP)

	if arrtab != strend {
		// must have some array data
		//for i := strend; i >= arrtab; i-- {
		//	vm.mm.WriteInterpreterMemory(vm.index, i+size, vm.mm.ReadInterpreterMemory(vm.index, i))
		//}

		vm.blockMoveUp(arrtab, arrtab+size, strend-arrtab)
	}

	strend += size
	arrtab += size

	if strend > fretop || arrtab > fretop {
		return errors.New("OUT OF MEMORY")
	}

	vm.SetVector(vm.ARRTAB, arrtab)
	vm.SetVector(vm.STREND, strend)

	//	//fmt.Printf("[msbin] Extend array start to %d, string end to %d\n", arrtab, strend)

	vm.CacheIndexed = make(map[[2]byte]int)
	vm.Cache = make(map[[2]byte]int)

	return nil

}

func (vm *VarManagerMSBIN) extendArrayMemory(size int) error {

	arrend := vm.GetVector(vm.STREND) - 1
	strend := vm.GetVector(vm.STREND)
	fretop := vm.GetVector(vm.FRETOP)

	strend += size
	arrend += size

	if arrend >= fretop {
		return errors.New("OUT OF MEMORY")
	}

	vm.SetVector(vm.STREND, strend)

	////fmt.Printf("[msbin] Extend array end to %d, string end to %d\n", arrend, strend)

	return nil

}

func (vm *VarManagerMSBIN) allocStringMemory(str string) (*StringPtr3b, error) {

	if utils.Len(str) > 255 {
		return &StringPtr3b{}, errors.New("STRING TOO LONG")
	}

	fretop := vm.GetVector(vm.FRETOP)
	strend := vm.GetVector(vm.STREND)

	size := utils.Len(str)

	allocAddr, e := vm.FindStringChunk(size)

	nfretop := fretop

	if e != nil {

		// realloc
		vm.CleanStrings()
		fretop = vm.GetVector(vm.FRETOP)
		nfretop = fretop - size

		if nfretop < strend {
			return &StringPtr3b{}, errors.New("OUT OF MEMORY")
		}
	}

	if allocAddr < nfretop {
		nfretop = allocAddr
	}

	rs := runestring.Cast(str)

	for i, ch := range rs.Runes {
		vm.mm.WriteInterpreterMemory(vm.index, i+allocAddr, uint64(ch))
	}

	vm.SetVector(vm.FRETOP, nfretop)

	//fmt.Printf("[msbin] Allocate string start %d, len %d\n", allocAddr, size)

	p := NewStringPtr3b(byte(size), allocAddr)

	return p, nil

}

func (vm *VarManagerMSBIN) CreateString(name, str string) error {

	ptr, e := vm.allocStringMemory(str)
	if e != nil {
		return e
	}

	return vm.Create(name, VT_STRING, ptr)

}

func (vm *VarManagerMSBIN) CreateStringIndexed(name string, capacity []int, str string) error {

	ptr, e := vm.allocStringMemory(str)
	if e != nil {
		return e
	}

	return vm.CreateIndexed(name, VT_STRING, capacity, ptr)

}

// Create a new var
func (vm *VarManagerMSBIN) Create(name string, kind VariableType, content interface{}) error {

	var msbin *MSBINRecord

	addr, _, nvarptr := vm.getAddressOfVar(name)
	if addr != -1 {
		// var exists
		return errors.New("VAR EXISTS")
	}

	content_type := reflect.TypeOf(content).String()

	msbin = &MSBINRecord{}

	msbin.SetName(name)
	msbin.SetType(kind)

	switch kind {
	case VT_FLOAT:
		if content_type != "*types.Float5b" {
			return errors.New("TYPE MISMATCH " + content_type)
		}
		msbin.SetFloatValue(content.(*Float5b))
	case VT_INTEGER:
		if content_type != "*types.Integer2b" {
			return errors.New("TYPE MISMATCH " + content_type)
		}
		msbin.SetIntValue(content.(*Integer2b))
	case VT_STRING:
		if content_type != "*types.StringPtr3b" {
			return errors.New("TYPE MISMATCH " + content_type)
		}
		msbin.SetStringPointer(content.(*StringPtr3b))
	case VT_EXPRESSION:
		if content_type != "*types.FuncPtr5b" {
			return errors.New("TYPE MISMATCH " + content_type)
		}
		msbin.SetFuncPointer(content.(*FuncPtr5b))
	default:
		return errors.New("UNKNOWN TYPE")
	}

	e := vm.extendVarMemory(VENTRY_SIZE)
	if e != nil {
		return e
	}

	// store the msbin at nvarptr
	msbin.WriteMemory(vm.mm, vm.index, nvarptr)
	////fmt.Println(msbin)

	return nil
}

// Create a new indexed var
func (vm *VarManagerMSBIN) CreateIndexed(name string, kind VariableType, capacity []int, content interface{}) error {

	// first does it already exist
	var msbin *MSBINArrayRecord

	addr, _, nvarptr := vm.getAddressOfArray(name)
	if addr != -1 {
		// var exists
		return errors.New("REDIM'D ARRAY")
	}

	msbin = &MSBINArrayRecord{}
	msbin.SetName(name)
	msbin.SetType(kind)
	msbin.DimCount = len(capacity)
	msbin.DimData = capacity

	bytesneeded := msbin.Size() // size including header and data allocation
	////fmt.Printf("[msbin] Array needs %d bytes of storage at %d\n", bytesneeded, nvarptr)
	////fmt.Printf("[msbin] Current header is %v\n", msbin)

	msbin.OffsetNext = uint16(bytesneeded) // Set pointer to next array record

	// try allocate the memory
	e := vm.extendArrayMemory(bytesneeded)
	if e != nil {
		return e
	}

	// Got memory // write header and read it back
	msbin.WriteMemory(vm.mm, vm.index, nvarptr)
	msbin.ReadMemory(vm.mm, vm.index, nvarptr)

	// Now we need to populate the array structure
	for i := 0; i < msbin.DataCount(); i++ {
		ptr := msbin.DataStart + i*msbin.ItemSize()
		switch kind {
		case VT_FLOAT:
			content.(*Float5b).WriteMemory(vm.mm, vm.index, ptr)
		case VT_INTEGER:
			content.(*Integer2b).WriteMemory(vm.mm, vm.index, ptr)
		case VT_EXPRESSION:
			content.(*FuncPtr5b).WriteMemory(vm.mm, vm.index, ptr)
		case VT_STRING:
			content.(*StringPtr3b).WriteMemory(vm.mm, vm.index, ptr)
		}
	}

	return nil

}

func (vm *VarManagerMSBIN) GetVector(base int) int {
	addr := vm.mm.ReadInterpreterMemory(vm.index, base) + 256*vm.mm.ReadInterpreterMemory(vm.index, base+1)
	return int(addr)
}

func (vm *VarManagerMSBIN) SetVector(base int, value int) {

	log.Printf("======================================> Setting vector %d to %d\n", base, value)

	if value < 0 {
		panic("bad msbin vector")
	}

	vm.mm.WriteInterpreterMemory(vm.index, base, uint64(value)%256)
	vm.mm.WriteInterpreterMemory(vm.index, base+1, uint64(value)/256)
}

// Returns Applesoft/MS-BASIC 2 char varname
func getShortVarName(name string) ([2]uint64, VariableType) {

	var vt VariableType = VT_FLOAT

	name = strings.ToUpper(name)
	if strings.HasSuffix(name, "$") {
		name = name[0 : len(name)-1]
		vt = VT_STRING
	}
	if strings.HasSuffix(name, "%") {
		name = name[0 : len(name)-1]
		vt = VT_INTEGER
	}
	if len(name) > 2 {
		name = name[0:2]
	}

	var result [2]uint64
	result[0] = uint64(name[0])
	if len(name) > 1 {
		result[1] = uint64(name[1])
	}

	// Type type bits
	if vt == VT_INTEGER || vt == VT_STRING {
		result[1] = result[1] | 128
	}
	if vt == VT_INTEGER || vt == VT_EXPRESSION {
		result[0] = result[0] | 128
	}

	return result, vt
}

func (vm *VarManagerMSBIN) getAddressOfVar(strname string) (int, VariableType, int) {

	varptr := vm.GetVector(vm.VARTAB) // start of VARTAB
	arrptr := vm.GetVector(vm.ARRTAB) // end of VARTAB

	found := false

	name, kind := getShortVarName(strname) // 2 byte name

	var key [2]byte = [2]byte{byte(name[0]), byte(name[1])}
	addr, ex := vm.Cache[key]
	if ex {
		////fmt.Printf("[msbin] Found var [%s] in cache at address %d\n", strname, addr)
		return addr, kind, arrptr
	}

	for varptr < arrptr && !found {

		a := vm.mm.ReadInterpreterMemory(vm.index, varptr)
		b := vm.mm.ReadInterpreterMemory(vm.index, varptr+1)

		////fmt.Printf("[msbin] Examine var header [%d, %d] vs target [%d, %d]\n", a, b, name[0], name[1])

		found := (name[0] == a && name[1] == b)

		if found {
			////fmt.Printf("[msbin] Found at %d\n", varptr)
			if vm.UseCache {
				vm.m.Lock()
				vm.Cache[key] = varptr
				vm.m.Unlock()
			}
			return varptr, kind, arrptr
		}

		varptr += 7

	}

	return -1, kind, varptr
}

func (vm *VarManagerMSBIN) getAddressOfArray(strname string) (int, VariableType, int) {

	arrptr := vm.GetVector(vm.ARRTAB) // start of array
	strend := vm.GetVector(vm.STREND) // end of array

	found := false

	name, kind := getShortVarName(strname) // 2 byte name

	var key [2]byte = [2]byte{byte(name[0]), byte(name[1])}
	addr, ex := vm.CacheIndexed[key]
	if ex {
		////fmt.Println("[msbin] Found var in cache")
		return addr, kind, strend - 1
	}

	for arrptr < strend-1 && !found {

		////fmt.Printf("[msbin] searching for array at %d\n", arrptr)

		a := vm.mm.ReadInterpreterMemory(vm.index, arrptr)
		b := vm.mm.ReadInterpreterMemory(vm.index, arrptr+1)

		found := (name[0] == a && name[1] == b)

		if found {
			if vm.UseCache {
				vm.m.Lock()
				vm.CacheIndexed[key] = arrptr
				vm.m.Unlock()
			}
			return arrptr, kind, strend - 1
		}

		// follow the pointer chain
		offset := vm.GetVector(arrptr + 2)
		////fmt.Printf("[msbin] Offset to next array is %d\n", offset)
		if offset < 5 {
			return -1, kind, strend - 1
		}
		arrptr += offset

	}

	////fmt.Println("[msbin-array] array not found...")

	return -1, kind, arrptr

}

// -- MSBINStruct
type MSBINRecord struct {
	Name [2]byte
	Data [5]byte
}

func (msbin *MSBINRecord) ReadMemory(mm *memory.MemoryMap, index int, address int) {
	msbin.Name[0] = byte(mm.ReadInterpreterMemory(index, address+0))
	msbin.Name[1] = byte(mm.ReadInterpreterMemory(index, address+1))
	msbin.Data[0] = byte(mm.ReadInterpreterMemory(index, address+2))
	msbin.Data[1] = byte(mm.ReadInterpreterMemory(index, address+3))
	msbin.Data[2] = byte(mm.ReadInterpreterMemory(index, address+4))
	msbin.Data[3] = byte(mm.ReadInterpreterMemory(index, address+5))
	msbin.Data[4] = byte(mm.ReadInterpreterMemory(index, address+6))
}

func (msbin *MSBINRecord) WriteMemory(mm *memory.MemoryMap, index int, address int) {
	mm.WriteInterpreterMemory(index, address+0, uint64(msbin.Name[0]))
	mm.WriteInterpreterMemory(index, address+1, uint64(msbin.Name[1]))
	mm.WriteInterpreterMemory(index, address+2, uint64(msbin.Data[0]))
	mm.WriteInterpreterMemory(index, address+3, uint64(msbin.Data[1]))
	mm.WriteInterpreterMemory(index, address+4, uint64(msbin.Data[2]))
	mm.WriteInterpreterMemory(index, address+5, uint64(msbin.Data[3]))
	mm.WriteInterpreterMemory(index, address+6, uint64(msbin.Data[4]))
}

func (msbin *MSBINRecord) GetType() VariableType {
	h1 := (msbin.Name[0]&128 == 128)
	h2 := (msbin.Name[1]&128 == 128)

	switch {
	case h1 && h2:
		return VT_INTEGER
	case h1 && !h2:
		return VT_EXPRESSION
	case !h1 && h2:
		return VT_STRING
	case !h1 && !h2:
		return VT_FLOAT
	}

	return VT_FLOAT
}

func (msbin *MSBINRecord) SetType(v VariableType) {

	h1 := msbin.Name[0] & 127
	h2 := msbin.Name[1] & 127

	switch v {
	case VT_STRING:
		h2 = h2 | 128
	case VT_EXPRESSION:
		h1 = h1 | 128
	case VT_INTEGER:
		h1 = h1 | 128
		h2 = h2 | 128
	}

	msbin.Name[0] = h1
	msbin.Name[1] = h2

}

func (msbin *MSBINRecord) GetName() string {

	vt := msbin.GetType()

	// extract name
	name := string(rune(msbin.Name[0]&127)) + string(rune(msbin.Name[1]&127))
	if name[1] == 0 {
		name = name[0:1]
	}

	if vt == VT_STRING {
		name += "$"
	} else if vt == VT_INTEGER {
		name += "%"
	}

	return name

}

func (msbin *MSBINRecord) SetName(v string) {

	name, nvt := getShortVarName(v)

	msbin.Name[0] = byte(name[0])
	msbin.Name[1] = byte(name[1])

	msbin.SetType(nvt)

}

func (msbin *MSBINRecord) SetIntValue(v *Integer2b) {
	msbin.Data[0] = v.hi
	msbin.Data[1] = v.lo
}

func (msbin *MSBINRecord) GetIntValue() *Integer2b {
	v := &Integer2b{
		hi: msbin.Data[0],
		lo: msbin.Data[1],
	}
	return v
}

func (msbin *MSBINRecord) SetFloatValue(v *Float5b) {
	msbin.Data[0] = v.exp
	msbin.Data[1] = v.m4
	msbin.Data[2] = v.m3
	msbin.Data[3] = v.m2
	msbin.Data[4] = v.m1
}

func (msbin *MSBINRecord) GetFloatValue() *Float5b {
	v := &Float5b{
		exp: msbin.Data[0],
		m4:  msbin.Data[1],
		m3:  msbin.Data[2],
		m2:  msbin.Data[3],
		m1:  msbin.Data[4],
	}
	return v
}

func (msbin *MSBINRecord) SetStringPointer(v *StringPtr3b) {
	msbin.Data[0] = v.length
	msbin.Data[1] = v.hi
	msbin.Data[2] = v.lo
}

func (msbin *MSBINRecord) GetStringPointer() *StringPtr3b {
	v := &StringPtr3b{
		length: msbin.Data[0],
		hi:     msbin.Data[1],
		lo:     msbin.Data[2],
	}
	return v
}

func (msbin *MSBINRecord) SetFuncPointer(v *FuncPtr5b) {
	msbin.Data[0] = v.hi
	msbin.Data[1] = v.lo
	msbin.Data[2] = v.vhi
	msbin.Data[3] = v.vlo
	msbin.Data[4] = v.fb
}

func (msbin *MSBINRecord) GetFuncPointer() *FuncPtr5b {
	v := &FuncPtr5b{
		hi:  msbin.Data[0],
		lo:  msbin.Data[1],
		vhi: msbin.Data[2],
		vlo: msbin.Data[3],
		fb:  msbin.Data[4],
	}
	return v
}

func (msbin *MSBINRecord) String() string {
	name := msbin.GetName()
	kind := msbin.GetType()
	switch kind {
	case VT_STRING:
		v := msbin.GetStringPointer()
		return fmt.Sprintf("[msbin] %s is STRING, length %d at %d\n", name, v.GetLength(), v.GetPointer())
	case VT_EXPRESSION:
		v := msbin.GetFuncPointer()
		return fmt.Sprintf("[msbin] %s is DEF FN, function @%d, var @%d\n", name, v.GetPointer(), v.GetArgPointer())
	case VT_FLOAT:
		v := msbin.GetFloatValue()
		return fmt.Sprintf("[msbin] %s is REAL, value %s\n", name, v.String())
	case VT_INTEGER:
		v := msbin.GetIntValue()
		return fmt.Sprintf("[msbin] %s is INTEGER, value %s\n", name, v.String())
	}
	return "[msbin] Unknown (corrupt memory?)"
}

// MSBINArrayRecord inherits the basic structure from MSBINRecord
type MSBINArrayRecord struct {
	MSBINRecord        // So we get the naming stuff for free
	OffsetNext  uint16 // offset to next array
	DimCount    int    // num dimensions
	DimData     []int
	DataStart   int // memory address for data
}

func (msbin *MSBINArrayRecord) ReadMemory(mm *memory.MemoryMap, index int, address int) {
	// Name
	msbin.Name[0] = byte(mm.ReadInterpreterMemory(index, address+0))
	msbin.Name[1] = byte(mm.ReadInterpreterMemory(index, address+1))
	// OffsetNext
	msbin.OffsetNext = uint16(mm.ReadInterpreterMemory(index, address+2)) + 256*uint16(mm.ReadInterpreterMemory(index, address+3))
	// DimCount
	msbin.DimCount = int(mm.ReadInterpreterMemory(index, address+4))
	// DimData
	msbin.DimData = make([]int, msbin.DimCount)
	for i, _ := range msbin.DimData {
		msbin.DimData[i] = 256*int(mm.ReadInterpreterMemory(index, address+5+(i*2))) + int(mm.ReadInterpreterMemory(index, address+6+(i*2)))
	}
	msbin.DataStart = address + 5 + msbin.DimCount*2
}

func (msbin *MSBINArrayRecord) WriteMemory(mm *memory.MemoryMap, index int, address int) {
	// Name
	mm.WriteInterpreterMemory(index, address+0, uint64(msbin.Name[0]))
	mm.WriteInterpreterMemory(index, address+1, uint64(msbin.Name[1]))
	// OffsetNext
	mm.WriteInterpreterMemory(index, address+2, uint64(msbin.OffsetNext%256))
	mm.WriteInterpreterMemory(index, address+3, uint64(msbin.OffsetNext/256))
	// DimCount
	mm.WriteInterpreterMemory(index, address+4, uint64(msbin.DimCount%256))
	// DimData
	for i, v := range msbin.DimData {
		mm.WriteInterpreterMemory(index, address+5+i*2, uint64(v/256))
		mm.WriteInterpreterMemory(index, address+6+i*2, uint64(v%256))
	}
}

func (msbin *MSBINArrayRecord) getFlatIndex(index []int) (int, error) {

	if len(index) != msbin.DimCount {
		return 0, errors.New("BAD SUBSCRIPT")
	}

	for i, d := range msbin.DimData {
		if index[i] > d || index[i] < 0 {
			return 0, errors.New("BAD SUBSCRIPT")
		}
	}

	v := index[len(msbin.DimData)-1]
	m := 1

	// NOTE: We add one to the DimData value because a dim of 4 really means 0-4 (5)
	for c := len(msbin.DimData) - 2; c >= 0; c-- {
		v = v + (msbin.DimData[c+1]+1)*index[c]*m
		m = m * (msbin.DimData[c+1] + 1)
	}

	return v, nil
}

func (msbin *MSBINArrayRecord) ItemSize() int {
	datasize := 5
	vt := msbin.GetType()
	switch vt {
	case VT_FLOAT:
		datasize = 5
	case VT_STRING:
		datasize = 3
	case VT_INTEGER:
		datasize = 2
	case VT_EXPRESSION:
		datasize = 5
	}
	return datasize
}

// Return array size in bytes
func (msbin *MSBINArrayRecord) Size() int {

	size := 5 + (msbin.DimCount * 2) + (msbin.DataCount() * msbin.ItemSize())

	return size

}

func (msbin *MSBINArrayRecord) DataCount() int {

	size := 1

	for _, v := range msbin.DimData {
		size = size * (v + 1)
	}

	return size
}

// Retrieve record from memory
func (msbin *MSBINArrayRecord) GetRecordAtIndex(mm *memory.MemoryMap, index int, aindex []int) (interface{}, error) {

	findex, e := msbin.getFlatIndex(aindex)
	if e != nil {
		return nil, e
	}

	// index is valid at this point
	dataaddr := msbin.DataStart + findex*msbin.ItemSize()

	//	//fmt.Printf("[msbin-array] Data for index %v located at %d\n", aindex, dataaddr)

	vt := msbin.GetType()

	switch vt {
	case VT_FLOAT:
		f := NewFloat5b(0)
		f.ReadMemory(mm, index, dataaddr)
		return f, nil
	case VT_INTEGER:
		f := NewInteger2b(0)
		f.ReadMemory(mm, index, dataaddr)
		return f, nil
	case VT_STRING:
		f := NewStringPtr3b(0, 0)
		f.ReadMemory(mm, index, dataaddr)
		return f, nil
	}

	return nil, nil

}

// Retrieve record from memory
func (msbin *MSBINArrayRecord) SetRecordAtIndex(mm *memory.MemoryMap, index int, aindex []int, content interface{}) error {

	findex, e := msbin.getFlatIndex(aindex)
	if e != nil {
		return e
	}

	// index is valid at this point
	dataaddr := msbin.DataStart + findex*msbin.ItemSize()

	//	//fmt.Printf("[msbin-array] Data for index %v located at %d\n", aindex, dataaddr)

	vt := msbin.GetType()

	switch vt {
	case VT_FLOAT:
		f := content.(*Float5b)
		f.WriteMemory(mm, index, dataaddr)
		return nil
	case VT_INTEGER:
		f := content.(*Integer2b)
		f.WriteMemory(mm, index, dataaddr)
		return nil
	case VT_STRING:
		f := content.(*StringPtr3b)
		f.WriteMemory(mm, index, dataaddr)
		return nil
	}

	return nil

}

func (vm *VarManagerMSBIN) GetDriver() VarDriver {
	return VD_MICROSOFT
}

func (vm *VarManagerMSBIN) Clear() {
	vm.SetVector(vm.ARRTAB, vm.GetVector(vm.VARTAB))
	vm.SetVector(vm.STREND, vm.GetVector(vm.VARTAB)+1)
	vm.SetVector(vm.FRETOP, vm.GetVector(vm.MEMSIZ))
}

func (vm *VarManagerMSBIN) blockMoveUp(oAddr, nAddr, size int) {
	log.Printf("Block memory move old = %d, new = %d, size = %d\n", oAddr, nAddr, size)
	for i := size - 1; i >= 0; i-- {
		vm.mm.WriteInterpreterMemory(vm.index, nAddr+i, vm.mm.ReadInterpreterMemory(vm.index, oAddr+i))
	}
}

// CleanStrings() defragments the string memory
func (vm *VarManagerMSBIN) CleanStrings() int {

	var usedBlocks = make(StrAllocMap)
	var blockUsedBy = make(StrAllocPointers)

	// Scan variables
	varptr := vm.GetVector(vm.VARTAB) // start of VARTAB
	arrptr := vm.GetVector(vm.ARRTAB) // end of VARTAB
	strend := vm.GetVector(vm.STREND) // end of ARRTAB
	memsiz := vm.GetVector(vm.MEMSIZ) // end of ARRTAB

	msbin := &MSBINRecord{}

	for varptr < arrptr {

		msbin.ReadMemory(vm.mm, vm.index, varptr)

		if msbin.GetType() == VT_STRING {
			// get its allocatiion record
			ptr := msbin.GetStringPointer()
			usedBlocks[ptr.GetPointer()] = *ptr

			slice, ok := blockUsedBy[*ptr]
			if !ok {
				slice = make([]int, 0)
			}
			slice = append(slice, varptr+2)
			blockUsedBy[*ptr] = slice
		}

		varptr += 7

	}

	amsbin := &MSBINArrayRecord{}

	for arrptr < strend-1 {

		amsbin.ReadMemory(vm.mm, vm.index, arrptr)

		if amsbin.GetType() == VT_STRING {

			// Need to scan the memory space
			count := amsbin.DataCount()
			for i := 0; i < count; i++ {
				dataaddr := amsbin.DataStart + i*3
				sptr := &StringPtr3b{}
				sptr.ReadMemory(vm.mm, vm.index, dataaddr)
				usedBlocks[sptr.GetPointer()] = *sptr
				slice, ok := blockUsedBy[*sptr]
				if !ok {
					slice = make([]int, 0)
				}
				slice = append(slice, dataaddr)
				blockUsedBy[*sptr] = slice
			}

		}

		// move to next
		offset := amsbin.OffsetNext
		if offset < 5 {
			break
		}
		arrptr += int(offset)

	}

	// At this stage usedBlocks contains a full map of all the blocks
	// Currently active in the variable space.
	// We can now crawl this in an ordered way to reclaim some free
	// space for stuff.

	keys := make([]int, 0)
	for i, _ := range usedBlocks {
		keys = append(keys, i)
	}
	sort.Ints(keys)

	high := memsiz

	// high represents the top of the compressed space
	moved := false
	for i := len(keys) - 1; i >= 0; i-- {
		curraddr := keys[i]
		currblk := usedBlocks[curraddr]

		// is there scope to move this block up?
		if currblk.Top() < high {
			moved = true
			diff := high - currblk.Top()

			newaddr := currblk.GetPointer() + diff // now location

			vm.blockMoveUp(curraddr, newaddr, int(currblk.GetLength()))

			//fmt.Printf(">> Relocating string memory at %d to %d (up %d bytes)\n", curraddr, newaddr, diff)

			// we need to redo all the pointers for currblk
			newblk := NewStringPtr3b(currblk.GetLength(), newaddr)
			refs := blockUsedBy[currblk]
			for _, ptraddr := range refs {
				// write newblk to ptraddr
				//	//fmt.Printf("** Updating string pointer stored at %d\n", ptraddr)
				newblk.WriteMemory(vm.mm, vm.index, ptraddr)
			}

			high = newaddr // high becomes byte before new string alloc
		} else {
			high = currblk.GetPointer()
		}
	}

	// update fretop to below the last string

	if moved {
		vm.SetVector(vm.FRETOP, high)
	}

	fretop := vm.GetVector(vm.FRETOP)

	return fretop - strend
}

func (vm *VarManagerMSBIN) SetHiBound(fretop int) {
	vm.SetVector(vm.MEMSIZ, fretop)
	vm.SetVector(vm.FRETOP, fretop)
}

func (vm *VarManagerMSBIN) SetLoBound(vartab int) {
	vm.SetVector(vm.VARTAB, vartab)
	vm.SetVector(vm.ARRTAB, vartab)
	vm.SetVector(vm.STREND, vartab+1)
}

func (vm *VarManagerMSBIN) GetFree() int {
	strend := vm.GetVector(vm.STREND)
	fretop := vm.GetVector(vm.FRETOP)
	vm.DumpStrings()
	return fretop - strend
}

func (vm *VarManagerMSBIN) DumpStrings() {

	var usedBlocks = make(StrAllocMap)

	// Scan variables
	varptr := vm.GetVector(vm.VARTAB) // start of VARTAB
	arrptr := vm.GetVector(vm.ARRTAB) // end of VARTAB
	strend := vm.GetVector(vm.STREND) // end of ARRTAB
	//memsiz := vm.GetVector(vm.MEMSIZ) // end of ARRTAB

	msbin := &MSBINRecord{}

	for varptr < arrptr {

		msbin.ReadMemory(vm.mm, vm.index, varptr)

		if msbin.GetType() == VT_STRING {
			// get its allocatiion record
			ptr := msbin.GetStringPointer()
			usedBlocks[ptr.GetPointer()] = *ptr

			//fmt2.Printf("%d: %s = [%s]\n", ptr.GetPointer(), msbin.GetName(), ptr.FetchString(vm.mm, vm.index))

		}

		varptr += 7

	}

	amsbin := &MSBINArrayRecord{}

	for arrptr < strend-1 {

		amsbin.ReadMemory(vm.mm, vm.index, arrptr)

		if amsbin.GetType() == VT_STRING {

			// Need to scan the memory space
			count := amsbin.DataCount()
			for i := 0; i < count; i++ {
				dataaddr := amsbin.DataStart + i*amsbin.ItemSize()
				sptr := &StringPtr3b{}
				sptr.ReadMemory(vm.mm, vm.index, dataaddr)
				usedBlocks[sptr.GetPointer()] = *sptr

				//fmt.Printf("%d: %s(%d) = [%s]\n", sptr.GetPointer(), amsbin.GetName(), i, sptr.FetchString(vm.mm, vm.index))
			}

		}

		// move to next
		offset := amsbin.OffsetNext
		if offset < 5 {
			break
		}
		arrptr += int(offset)

	}

}

func (vm *VarManagerMSBIN) FindStringChunk(length int) (int, error) {

	if length == 0 {
		length = 1
	}

	var usedBlocks = make(StrAllocMap)

	// Scan variables
	varptr := vm.GetVector(vm.VARTAB) // start of VARTAB
	arrptr := vm.GetVector(vm.ARRTAB) // end of VARTAB
	strend := vm.GetVector(vm.STREND) // end of ARRTAB
	memsiz := vm.GetVector(vm.MEMSIZ) // end of ARRTAB

	msbin := &MSBINRecord{}

	for varptr < arrptr {

		msbin.ReadMemory(vm.mm, vm.index, varptr)

		if msbin.GetType() == VT_STRING {
			// get its allocatiion record
			ptr := msbin.GetStringPointer()
			if ptr.GetLength() > 0 {
				usedBlocks[ptr.GetPointer()] = *ptr
			}
		}

		varptr += 7

	}

	amsbin := &MSBINArrayRecord{}

	for arrptr < strend-1 {

		amsbin.ReadMemory(vm.mm, vm.index, arrptr)

		if amsbin.GetType() == VT_STRING {

			// Need to scan the memory space
			count := amsbin.DataCount()
			for i := 0; i < count; i++ {
				dataaddr := amsbin.DataStart + i*3
				sptr := &StringPtr3b{}
				sptr.ReadMemory(vm.mm, vm.index, dataaddr)
				if sptr.GetLength() > 0 {
					usedBlocks[sptr.GetPointer()] = *sptr
				}
			}

		}

		// move to next
		offset := amsbin.OffsetNext
		if offset < 5 {
			break
		}
		arrptr += int(offset)

	}

	// At this stage usedBlocks contains a full map of all the blocks
	// Currently active in the variable space.
	// We can now crawl this in an ordered way to reclaim some free
	// space for stuff.

	keys := make([]int, 0)
	for i, _ := range usedBlocks {
		keys = append(keys, i)
	}
	sort.Ints(keys)

	high := memsiz

	// high represents the top of the compressed space
	//	moved := false
	for i := len(keys) - 1; i >= 0; i-- {
		curraddr := keys[i]
		currblk := usedBlocks[curraddr]

		if currblk.Top() < high {
			diff := high - currblk.Top()
			if diff >= length {
				newaddr := high - length
				return newaddr, nil
			}
			high = currblk.GetPointer()
		} else {
			high = currblk.GetPointer()
		}
	}

	fretop := vm.GetVector(vm.FRETOP)
	if high-length >= fretop {
		return high - length, nil
	}

	nfretop := fretop - length

	if nfretop < strend {
		return -1, errors.New("OUT OF MEMORY")
	}

	return nfretop, nil
}
