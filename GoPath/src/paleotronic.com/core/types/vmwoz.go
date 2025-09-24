package types

import (
	"bytes"
	"errors"
	"strings"

	//    "paleotronic.com/fmt"

	"paleotronic.com/core/memory"
)

const (
	INTEGER_LOMEM = 2048
	INTEGER_HIMEM = 38400
)

/*
  Example:
  		   A=-1
           bytes:
           C1 00 06 08 FF FF

           NM DS NADDR VALUE
           -- -- ----- -----
           A  0  2054  -1
*/

type VarManagerWOZ struct {
	// Pointers to the interpreter reference
	mm    *memory.MemoryMap
	index int
	// Data pointers
	VARBOT int // first address in variable table (INTEGER_LOMEM) 74,75
	VARTOP int // end of variable memory 204,205
	BASBOT int // bottom-most address used by basic code 202,203
	BASTOP int // top of program memory (INTEGER_HIMEM) 76, 77
	// configs
	UndefResponse VarUndefinedResponse
	Cache         map[[100]byte]int
}

// NewVarManagerWOZ creates a new microsoft basic compatible var manager
func NewVarManagerWOZ(mm *memory.MemoryMap, index int, vb, vt, bb, bt int, uresp VarUndefinedResponse) *VarManagerWOZ {

	this := &VarManagerWOZ{
		mm:            mm,
		index:         index,
		VARBOT:        vb,
		VARTOP:        vt,
		BASBOT:        bb,
		BASTOP:        bt,
		UndefResponse: uresp,
		Cache:         make(map[[100]byte]int),
	}

	return this

}

const (
	WOZ_VARNAME_MAX = 100
)

type WOZVarRecord struct {
	Name   [100]byte // up to 100 bytes, bit 128 set
	DSP    byte      // 01 = display changes during execution, 00 = not
	NextLo byte      // Next variable address
	NextHi byte
	VData  []byte // Integer 2 bytes lo, hi
	// String 0-255 bytes (high) + ST (<128)
}

func (w *WOZVarRecord) GetSize() int {
	namelen := len(w.GetName())
	return namelen + 3 + len(w.VData)
}

func (w *WOZVarRecord) GetNextAddress() int {
	return int(w.NextLo) + 256*int(w.NextHi)
}

// ReadMemory extracts a variables representation from memory
func (w *WOZVarRecord) ReadMemory(mm *memory.MemoryMap, index int, address int) {

	for i, _ := range w.Name {
		w.Name[i] = 0
	}

	i := 0
	v := mm.ReadInterpreterMemory(index, address+i)
	for v > 1 && i < 100 {
		w.Name[i] = byte(v)
		i++
		v = mm.ReadInterpreterMemory(index, address+i)
	}
	w.DSP = byte(mm.ReadInterpreterMemory(index, address+i))
	i++
	w.NextLo = byte(mm.ReadInterpreterMemory(index, address+i))
	i++
	w.NextHi = byte(mm.ReadInterpreterMemory(index, address+i))
	i++
	nextaddr := w.GetNextAddress()
	count := nextaddr - (address + i)
	if count > 49152 {
		count = 1
	}
	w.VData = make([]byte, count)

	for z, _ := range w.VData {
		w.VData[z] = byte(mm.ReadInterpreterMemory(index, address+i+z))
	}

}

// WriteMemory encodes a variables representation to memory
func (w *WOZVarRecord) WriteMemory(mm *memory.MemoryMap, index int, address int) {

	// Encode name data
	count := 0
	for i, v := range w.Name {
		if v > 1 {
			mm.WriteInterpreterMemory(index, address+count, uint64(w.Name[i]))
			count++
		}
	}

	// DSP
	mm.WriteInterpreterMemory(index, address+count, uint64(w.DSP&1))
	count++

	// Next Var address (lo, hi)
	mm.WriteInterpreterMemory(index, address+count, uint64(w.NextLo))
	count++
	mm.WriteInterpreterMemory(index, address+count, uint64(w.NextHi))
	count++

	// datasize
	for _, v := range w.VData {
		mm.WriteInterpreterMemory(index, address+count, uint64(v))
		count++
	}

}

// GetType returns the type of variable (VT_INTEGER, VT_STRING)
func (w *WOZVarRecord) GetType() VariableType {

	n := w.GetName()
	if strings.HasSuffix(n, "$") {
		return VT_STRING
	}
	return VT_INTEGER

}

// GetName returns the name of variable (converted to ASCII)
func (w *WOZVarRecord) GetName() string {
	out := ""
	for _, ch := range w.Name {
		if ch >= 128 {
			out += string(rune(ch) - 128)
		} else if ch == 64 {
			out += "$"
		}
	}
	return out
}

// DataCount returns the number of elements
func (w *WOZVarRecord) DataCount() int {

	if w.GetType() == VT_STRING {
		return 1
	} else {
		return len(w.VData) / w.ItemSize()
	}

}

// ItemSize returns the element size
func (w *WOZVarRecord) ItemSize() int {

	if w.GetType() == VT_STRING {
		return len(w.VData) - 1 // account for ST byte
	} else {
		return 2
	}

}

func (w *WOZVarRecord) GetIntValue() (*Integer2b, error) {
	if w.GetType() != VT_INTEGER {
		return nil, errors.New("SYNTAX ERR")
	}
	v := &Integer2b{
		hi: w.VData[1],
		lo: w.VData[0],
	}
	return v, nil
}

func (w *WOZVarRecord) SetIntValue(v *Integer2b) error {
	if w.GetType() != VT_INTEGER {
		return errors.New("SYNTAX ERR")
	}
	w.VData[0] = v.lo
	w.VData[1] = v.hi
	return nil
}

func (w *WOZVarRecord) GetIntValueIndexed(i int) (*Integer2b, error) {
	if w.GetType() != VT_INTEGER {
		return nil, errors.New("SYNTAX ERR")
	}
	if i >= 0 && i < w.DataCount() {
		v := &Integer2b{
			hi: w.VData[1+i*2],
			lo: w.VData[0+i*2],
		}
		return v, nil
	}
	return nil, errors.New("RANGE ERR")
}

func (w *WOZVarRecord) SetIntValueIndexed(i int, v *Integer2b) error {
	if w.GetType() != VT_INTEGER {
		return errors.New("SYNTAX ERR")
	}
	if i >= 0 && i < w.DataCount() {
		w.VData[0+i*2] = v.lo
		w.VData[1+i*2] = v.hi
		return nil
	}
	return errors.New("RANGE ERR")
}

func (w *WOZVarRecord) GetStringValue() (string, error) {
	if w.GetType() != VT_STRING {
		return "", errors.New("SYNTAX ERR")
	}
	out := ""
	end := false
	for i, v := range w.VData {
		if v == 0x1e {
			end = true
		}
		if i < len(w.VData)-1 && !end {
			out += string(rune(v & 127))
		}
	}
	return out, nil
}

func (w *WOZVarRecord) SetStringValue(s string) error {
	if w.GetType() != VT_STRING {
		return errors.New("SYNTAX ERR")
	}
	//    fmt.Println(len(w.VData))
	if len(s) > len(w.VData)-1 {
		return errors.New("STR OVFL ERR")
	}
	for i, v := range s {
		w.VData[i] = byte(v | 128)
	}
	w.VData[len(s)] = 0x1e
	return nil
}

/*
* Functions for the Var Manager
 */

// Exists() returns true if variable exists
func (w *VarManagerWOZ) Exists(name string) bool {
	addr, _, _ := w.GetVariableAddress(name)
	return (addr != -1)
}

// GetValue() returns value of specified variable
func (w *VarManagerWOZ) GetValue(name string) (interface{}, error) {

	addr, vt, _ := w.GetVariableAddress(name)

	if addr == -1 {
		// handle var not existing
		switch w.UndefResponse {
		// quiet implies we just return an "empty" value
		case VUR_QUIET:
			switch vt {
			case VT_STRING:
				return "", nil
			case VT_INTEGER:
				return &Integer2b{}, nil
			}
		case VUR_ERROR:
			return nil, errors.New("SYNTAX ERR")
		case VUR_CREATE:
			return nil, errors.New("SYNTAX ERR")
		}
	}

	var woz *WOZVarRecord = &WOZVarRecord{}
	woz.ReadMemory(w.mm, w.index, addr)

	switch woz.GetType() {
	case VT_STRING:
		return woz.GetStringValue()
	case VT_INTEGER:
		return woz.GetIntValue()
	}

	return nil, errors.New("SYNTAX ERR")
}

// SetValue() sets the value of an existing variable
func (w *VarManagerWOZ) SetValue(name string, value interface{}) error {

	addr, vt, _ := w.GetVariableAddress(name)

	if addr == -1 {
		if vt == VT_INTEGER {
			return w.Create(name, vt, value)
		} else {
			return errors.New("STR OVFL ERR")
		}
	}

	// exists
	var woz *WOZVarRecord = &WOZVarRecord{}
	woz.ReadMemory(w.mm, w.index, addr)

	switch woz.GetType() {
	case VT_STRING:
		e := woz.SetStringValue(value.(string))
		if e != nil {
			return e
		}
	case VT_INTEGER:
		e := woz.SetIntValue(value.(*Integer2b))
		if e != nil {
			return e
		}
	}

	woz.WriteMemory(w.mm, w.index, addr)
	return nil

}

func (w *VarManagerWOZ) ExistsIndexed(name string) bool {

	addr, _, _ := w.GetVariableAddress(name)

	if addr == -1 {
		return false
	}

	var woz *WOZVarRecord = &WOZVarRecord{}
	woz.ReadMemory(w.mm, w.index, addr)

	return (woz.DataCount() > 1)

}

func (w *VarManagerWOZ) GetValueIndexed(name string, index []int) (interface{}, error) {

	if len(index) != 1 {
		return nil, errors.New("SYNTAX ERR")
	}

	addr, vt, _ := w.GetVariableAddress(name)

	if addr == -1 {
		// handle var not existing
		switch w.UndefResponse {
		// quiet implies we just return an "empty" value
		case VUR_QUIET:
			switch vt {
			case VT_STRING:
				return "", nil
			case VT_INTEGER:
				return &Integer2b{}, nil
			}
		case VUR_ERROR:
			return nil, errors.New("SYNTAX ERR")
		case VUR_CREATE:
			return nil, errors.New("SYNTAX ERR")
		}
	}

	var woz *WOZVarRecord = &WOZVarRecord{}
	woz.ReadMemory(w.mm, w.index, addr)

	switch woz.GetType() {
	case VT_STRING:
		return nil, errors.New("SYNTAX ERR")
	case VT_INTEGER:
		return woz.GetIntValueIndexed(index[0])
	}

	return nil, errors.New("SYNTAX ERR")

}

func (w *VarManagerWOZ) SetValueIndexed(name string, index []int, value interface{}) error {

	if len(index) != 1 {
		return errors.New("SYNTAX ERR")
	}

	addr, vt, _ := w.GetVariableAddress(name)

	if addr == -1 {
		if vt == VT_INTEGER {
			return errors.New("RANGE ERR")
		} else {
			return errors.New("SYNTAX ERR")
		}
	}

	// exists
	var woz *WOZVarRecord = &WOZVarRecord{}
	woz.ReadMemory(w.mm, w.index, addr)

	switch woz.GetType() {
	case VT_STRING:
		return errors.New("SYNTAX ERR")
	case VT_INTEGER:
		e := woz.SetIntValueIndexed(index[0], value.(*Integer2b))
		if e != nil {
			return e
		}
	}

	woz.WriteMemory(w.mm, w.index, addr)
	return nil

}

func (w *VarManagerWOZ) GetVarNames() []string {

	out := make([]string, 0)

	varptr := w.GetVector(w.VARBOT)
	varend := w.GetVector(w.VARTOP)

	var woz *WOZVarRecord = &WOZVarRecord{}

	for varptr < varend {

		// read var definition
		woz.ReadMemory(w.mm, w.index, varptr)

		out = append(out, woz.GetName())

		ovptr := varptr
		varptr = woz.GetNextAddress()

		if ovptr > varptr {
			return out
		}
	}

	// didn't find it
	return out

}

func (w *VarManagerWOZ) GetVarNamesIndexed() []string {
	out := make([]string, 0)

	varptr := w.GetVector(w.VARBOT)
	varend := w.GetVector(w.VARTOP)

	var woz *WOZVarRecord = &WOZVarRecord{}

	for varptr < varend {

		// read var definition
		woz.ReadMemory(w.mm, w.index, varptr)

		if woz.DataCount() > 1 {
			out = append(out, woz.GetName())
		}

		ovptr := varptr
		varptr = woz.GetNextAddress()

		if ovptr > varptr {
			return out
		}

	}

	// didn't find it
	return out
}

func (w *VarManagerWOZ) Create(name string, kind VariableType, content interface{}) error {

	if w.Exists(name) {
		return w.SetValue(name, content)
	}

	// not exist
	bname, _, _ := w.GetWOZVariableName(name)
	varend := w.GetVector(w.VARTOP)

	var woz *WOZVarRecord = &WOZVarRecord{}
	// create object
	woz.Name = bname
	woz.DSP = 0
	switch kind {
	case VT_INTEGER:
		woz.VData = make([]byte, 2)
		i := content.(*Integer2b)
		e := woz.SetIntValue(i)
		if e != nil {
			return e
		}
	case VT_STRING:
		str := content.(string)
		woz.VData = make([]byte, len(str)+1)
		e := woz.SetStringValue(str)
		if e != nil {
			return e
		}
	default:
		return errors.New("SYNTAX ERR")
	}

	// Got here and set variable
	bytesneeded := woz.GetSize()
	nextaddr := varend + bytesneeded
	woz.NextLo = byte(nextaddr % 256)
	woz.NextHi = byte(nextaddr / 256)

	// write it out
	woz.WriteMemory(w.mm, w.index, varend)
	w.SetVector(w.VARTOP, nextaddr)

	return nil
}

func (w *VarManagerWOZ) CreateIndexed(name string, kind VariableType, capacity []int, content interface{}) error {

	if w.Exists(name) {
		return errors.New("DIM ERR")
	}

	if len(capacity) != 1 {
		return errors.New("SYNTAX ERR")
	}

	// not exist
	bname, _, _ := w.GetWOZVariableName(name)
	varend := w.GetVector(w.VARTOP)

	var woz *WOZVarRecord = &WOZVarRecord{}
	// create object
	woz.Name = bname
	woz.DSP = 0
	switch kind {
	case VT_INTEGER:
		woz.VData = make([]byte, 2*(capacity[0]+1))
		i := content.(*Integer2b)
		for z := 0; z < (capacity[0] + 1); z++ {
			e := woz.SetIntValueIndexed(z, i)
			if e != nil {
				return e
			}
		}
	case VT_STRING:
		str := content.(string)
		if capacity[0] > 255 {
			return errors.New("STR OVFL ERR")
		}
		woz.VData = make([]byte, capacity[0]+1)
		e := woz.SetStringValue(str)
		if e != nil {
			return e
		}
	default:
		return errors.New("SYNTAX ERR")
	}

	// Got here and set variable
	bytesneeded := woz.GetSize()
	nextaddr := varend + bytesneeded
	woz.NextLo = byte(nextaddr % 256)
	woz.NextHi = byte(nextaddr / 256)

	// write it out
	woz.WriteMemory(w.mm, w.index, varend)
	w.SetVector(w.VARTOP, nextaddr)

	return nil

}

func (w *VarManagerWOZ) CreateString(name, str string) error {

	if w.Exists(name) {
		return w.SetValue(name, str)
	}

	// not exists
	return w.Create(name, VT_STRING, str)

}

func (w *VarManagerWOZ) CreateStringIndexed(name string, capacity []int, str string) error {
	return errors.New("SYNTAX ERR")
}

// private service codes

func (vm *VarManagerWOZ) GetVector(base int) int {
	addr := vm.mm.ReadInterpreterMemory(vm.index, base) + 256*vm.mm.ReadInterpreterMemory(vm.index, base+1)
	return int(addr)
}

func (vm *VarManagerWOZ) SetVector(base int, value int) {
	vm.mm.WriteInterpreterMemory(vm.index, base, uint64(value)%256)
	vm.mm.WriteInterpreterMemory(vm.index, base+1, uint64(value)/256)
}

// GetVariableAddress() returns the address of the specified variable
func (w *VarManagerWOZ) GetVariableAddress(name string) (int, VariableType, int) {

	//fmt.Printf("Looking for [%s]\n", name)

	bname, vt, namelen := w.GetWOZVariableName(name)

	varptr := w.GetVector(w.VARBOT)
	varend := w.GetVector(w.VARTOP)

	if addr, ex := w.Cache[bname]; ex {
		return addr, vt, varend
	}

	found := false
	var woz *WOZVarRecord = &WOZVarRecord{}

	for varptr < varend && !found {

		// read var definition
		woz.ReadMemory(w.mm, w.index, varptr)
		if len(woz.GetName()) != namelen {
			varptr = woz.GetNextAddress()
			continue
		}

		//        fmt.Printf("Var search are %v and %v equal?\n", bname[0:namelen], woz.Name[0:namelen] )

		if bytes.Compare(bname[0:namelen], woz.Name[0:namelen]) == 0 {
			//            fmt.Println("MATCHED")
			// matching name
			w.Cache[bname] = varptr
			return varptr, vt, varend
		}

		varptr = woz.GetNextAddress()

	}

	// didn't find it
	return -1, vt, varend

}

// GetWOZVariableName() returns a 100 byte potential formatted name
func (w *VarManagerWOZ) GetWOZVariableName(name string) ([100]byte, VariableType, int) {
	name = strings.ToUpper(name)

	var outname [100]byte
	var vt VariableType = VT_INTEGER
	var l int = len(name)
	if l > 100 {
		l = 100
	}

	for i, v := range name {
		if i >= WOZ_VARNAME_MAX {
			continue
		}
		if rune(v) == '$' {
			outname[i] = 0x40
			vt = VT_STRING
		} else {
			outname[i] = byte(v | 128)
		}
	}

	return outname, vt, l
}

func (vm *VarManagerWOZ) GetDriver() VarDriver {
	return VD_WOZ
}

func (vm *VarManagerWOZ) Clear() {
	vm.SetVector(vm.VARTOP, vm.GetVector(vm.VARBOT))
}

func (vm *VarManagerWOZ) CleanStrings() int {
	// do nothing here
	return vm.GetVector(vm.BASBOT) - vm.GetVector(vm.VARTOP)
}

func (vm *VarManagerWOZ) Contains(name string) bool {
	a, _, _ := vm.GetVariableAddress(name)
	return (a != -1)
}

func (vm *VarManagerWOZ) ContainsKey(name string) bool {
	return vm.Contains(name)
}

func (vm *VarManagerWOZ) Dimensions(name string) []int {

	addr, _, _ := vm.GetVariableAddress(name)
	var msbin *WOZVarRecord = &WOZVarRecord{}

	msbin.ReadMemory(vm.mm, vm.index, addr)

	return []int{msbin.DataCount()}

}

func (vm *VarManagerWOZ) Put(name string, v *Variable) {

	if v.IsArray() {
		// array
	} else {
		// scalar
	}

}

func (vm *VarManagerWOZ) Get(name string) *Variable {

	if vm.Exists(name) {

		addr, _, _ := vm.GetVariableAddress(name)
		var msbin *WOZVarRecord = &WOZVarRecord{}
		msbin.ReadMemory(vm.mm, vm.index, addr)

		if msbin.DataCount() > 1 {

			return &Variable{
				Name:           name,
				Driver:         VD_WOZ,
				Context:        VDC_ARRAY,
				Map:            vm,
				Kind:           msbin.GetType(),
				AssumeLowIndex: true,
				Owner:          "",
			}

		} else {

			return &Variable{
				Name:           name,
				Driver:         VD_WOZ,
				Context:        VDC_SCALAR,
				Map:            vm,
				Kind:           msbin.GetType(),
				AssumeLowIndex: true,
				Owner:          "",
			}

		}

	}

	return nil
}

func (vm *VarManagerWOZ) GetIndex() int {
	return vm.index
}

func (vm *VarManagerWOZ) GetMM() *memory.MemoryMap {
	return vm.mm
}

func (vm *VarManagerWOZ) SetHiBound(fretop int) {
	vm.SetVector(vm.BASBOT, fretop)
}

func (vm *VarManagerWOZ) SetLoBound(vartab int) {
	vm.SetVector(vm.VARBOT, vartab)
	vm.SetVector(vm.VARTOP, vartab)
}

func (vm *VarManagerWOZ) GetFree() int {
	strend := vm.GetVector(vm.VARTOP)
	fretop := vm.GetVector(vm.BASBOT)

	return fretop - strend
}
