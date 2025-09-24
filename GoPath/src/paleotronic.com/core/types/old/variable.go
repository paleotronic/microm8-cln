// variable.go
package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"
	"paleotronic.com/fmt"

	"paleotronic.com/log"
	"paleotronic.com/utils"
    "paleotronic.com/core/memory"
)

type VariableType int

const (
	VT_STRING     VariableType = 1 + iota
	VT_BOOLEAN    VariableType = 1 + iota
	VT_INTEGER    VariableType = 1 + iota
	VT_FLOAT      VariableType = 1 + iota
	VT_EXPRESSION VariableType = 1 + iota
)

const (
	MAX_EXPRESSION_LENGTH = 256
	MAX_STRING_LENGTH     = 256
    MAX_COMPRESSED_STRING = MAX_STRING_LENGTH/4
	MAX_FLOAT_LENGTH      = 1
	MAX_INTEGER_LENGTH    = 1
	DEFAULT_NAME_ULEN      = 4
	DEFAULT_INDEX         = 1024
)

type Variable struct { 
	Content        []*VarValue
	ContentScalar  *VarValue
	Length         int 
	Dimensions     []int
	AssumeLowIndex bool
	Kind           VariableType
	Name           string
	Mutable        bool
	ZeroIsScalar   bool
	Owner          string
	Stacked        bool
	Shared         bool
	Map            *VarMap
	VarAddress     int
	VarSize        int
}

type VarMemoryManager interface {
	SetMemory(address int, value uint)
	GetMemory(address int) uint
    SetVM( vm VarManager )
    GetVM() VarManager
    GetMemoryMap() *memory.MemoryMap
    GetMemIndex() int
}

type StrAlloc struct {
	 Address int
     Length  int
}

type StringMemoryManager map[string]StrAlloc
type StringMemoryDealloc map[int]StrAlloc

type VarMap struct {
	Content   map[string]*Variable
	MaxLength int
	Preserve  []string
	Mgr       VarMemoryManager
	// Memory management
	VIndexBase int
	VIndexNext int
	VTableBase int // base address for vtable in memory
	VTableNext int // next slot address
	VStringBase int // base address for vtable in memory
	VStringNext int // next slot address
    
    // String allocation tracker
    StringMemory StringMemoryManager
    StringReuse  StringMemoryDealloc
    
}

func GetVarKey( name string, index int ) string {
	 return name + "." + utils.IntToStr(index)
}

func (this VariableType) String() string {
	switch this {
	case VT_BOOLEAN:
		return "BOOLEAN"
	case VT_EXPRESSION:
		return "EXPRESSION"
	case VT_FLOAT:
		return "FLOAT"
	case VT_INTEGER:
		return "INTEGER"
	case VT_STRING:
		return "STRING"
	}
	return "(other)"
}

func (this VariableType) Size() int {
	switch this {
	case VT_BOOLEAN:
		return MAX_INTEGER_LENGTH
	case VT_EXPRESSION:
		return MAX_EXPRESSION_LENGTH
	case VT_FLOAT:
		return MAX_FLOAT_LENGTH
	case VT_INTEGER:
		return MAX_INTEGER_LENGTH
	case VT_STRING:
		return MAX_COMPRESSED_STRING
	}
	return MAX_INTEGER_LENGTH
}

func (this VariableType) GetNeededAllocation(length int) int {
	return this.Size() * length
}

func NewVarMap(mvl int, mgr VarMemoryManager) *VarMap {
	this := &VarMap{Mgr: mgr, Content: make(map[string]*Variable), MaxLength: mvl, Preserve: []string{"color", "hcolor", "scale", "rot"}}

	this.VTableBase = 65536+3072+16384*2
	this.VTableNext = this.VTableBase+2
	this.VStringBase = 0x19400
	this.VStringNext = 0x19400
    this.StringMemory = make( StringMemoryManager )
    this.StringReuse  = make( StringMemoryDealloc )

	return this
}

func (this *VarMap) SetBase( address int, saddress int ) {

	this.VIndexBase = saddress
	this.VIndexNext = saddress
	this.VTableBase = address
	this.VTableNext = address
	//this.Clear()
//	//fmt.Printf( "*** Setting Variable storage base to $%04x\n", this.VIndexBase )
	//this.Mgr.SetMemory(this.VIndexBase, 0)
	this.VStringBase = saddress+DEFAULT_INDEX
	this.VStringNext = saddress+DEFAULT_INDEX
}

// True if map contains variable with that name
func (this VarMap) Contains(n string) bool {
	_, ok := this.Content[this.CompliantName(n)]
	return ok
}

func (this VarMap) ContainsKey(n string) bool {
	_, ok := this.Content[this.CompliantName(n)]
	return ok
}

// Used internally to generate a compliant variable name
func (this *VarMap) CompliantName(n string) string {

	// if no restriction, return name
	if this.MaxLength == -1 {
		return n
	}

	// if name is marked for preservation, return name
	for _, pname := range this.Preserve {
		if pname == n {
			return n
		}
	}

	var suffix string
	if (n[len(n)-1] == '$') || (n[len(n)-1] == '%') {
		suffix = utils.Copy(n, len(n), 1)
		n = utils.Delete(n, len(n), 1)
	}

	if len(n) > this.MaxLength {
		n = utils.Copy(n, 1, this.MaxLength)
	}

	return strings.ToLower(n + suffix)
}

// Returns a Variable for a given name, or nil
func (this *VarMap) Get(n string) *Variable {
	cn := this.CompliantName(n)
	if this.Contains(cn) {
		return this.Content[cn]
	}
	return nil
}

func (this *VarMap) PutQuietly(n string, v *Variable) {
	cn := this.CompliantName(n)
	this.Content[cn] = v
	v.Map = this
}

// Maps a Variable to a particular key
func (this *VarMap) Put(n string, v *Variable) {

    if v == nil {
       panic("Variable pointer is nil: "+n)
    }

	cn := this.CompliantName(n)
	this.Content[cn] = v

	alloc := (v.Kind.Size() * v.Length)

	v.VarSize = alloc
	if v.Kind == VT_STRING {
		alloc = v.Length   // strings are just pointer to secondary memory
	}
	v.VarAddress = this.VTableNext
	this.VTableNext += alloc

//	//fmt.Printf( "*** Variable allocation %s at $%06x\n", cn, v.VarAddress )

	v.Map = this
	if v.Length == 1 {
		vvv, _ := v.GetContentScalar()
		v.FreshenScalar(vvv)
	} else {
		v.FreshenArray()
	}

	// Add to the index...
	namedata := PackName(cn, DEFAULT_NAME_ULEN*4)
	namedata = append(namedata, uint(v.Kind))
	namedata = append(namedata, uint(v.Length))
	namedata = append(namedata, uint(v.VarAddress))
	namedata = append(namedata, uint(v.VarSize))

	if v.Length > 1 {
		namedata = append(namedata, uint(len(v.Dimensions)))
		for _, n := range v.Dimensions {
			namedata = append(namedata, uint(n))
		}
	}

//	//fmt.Printf("*** Var %s metadata at $%04x\n", cn, this.VIndexNext)
	for i, vv := range namedata {
		this.Mgr.SetMemory( this.VIndexNext+i, vv )
	}
	this.VIndexNext += len(namedata)
	this.Mgr.SetMemory( this.VIndexNext, 0 )

}

func (this *VarMap) Defrost( zis bool ) {
	ptr := this.VIndexBase
	biggest := this.VTableBase
	sbiggest := this.VStringBase
	for this.Mgr.GetMemory(ptr) != 0 {

		namedata := make([]uint, DEFAULT_NAME_ULEN+4)

		for i := 0; i<len(namedata); i++ {
			namedata[i] = this.Mgr.GetMemory(ptr+i)
		}

		// decode
		vname := UnpackName(namedata[0:DEFAULT_NAME_ULEN])
		vkind  := VariableType(namedata[DEFAULT_NAME_ULEN+0])
		vlength  := int(namedata[DEFAULT_NAME_ULEN+1])
		vaddress  := int(namedata[DEFAULT_NAME_ULEN+2])
		vsize  := int(namedata[DEFAULT_NAME_ULEN+3])
		dims := make([]int, 1)
		dims[0] = 1

		if vlength > 1 {
			namedata = append(namedata, this.Mgr.GetMemory(ptr+DEFAULT_NAME_ULEN+4))
			ndims := int(namedata[DEFAULT_NAME_ULEN+4])
			if ndims > 10 {
				panic("dimensions seem wrong")
			}
			dims = make([]int, ndims)
			for i:=0; i<ndims; i++ {
				namedata = append(namedata, this.Mgr.GetMemory(ptr+DEFAULT_NAME_ULEN+5+i))
				dims[i] = int(namedata[DEFAULT_NAME_ULEN+5+i])
			}
		}

//		//fmt.Printf( "=== Defrost: %s(%d) %s (%d bytes at %d) %v\n", vname, vlength, vkind.String(), vsize, vaddress, dims  )

		// now let's do it
		if vlength == 1 {
			// Create a thing
			v := NewVariableP(vname, vkind, "", true)
			v.ZeroIsScalar = zis
			v.VarAddress = vaddress
			v.VarSize = vsize
			this.PutQuietly(vname, v)
			v.SyncScalar()
		} else {
			// Create array of thing
			v := NewVariablePA(vname, vkind, "", true, dims)
			v.ZeroIsScalar = zis
			v.VarAddress = vaddress
			v.VarSize = vsize
			this.PutQuietly(vname, v)
			v.SyncArray()
		}

		if vkind == VT_STRING {
			if vaddress + vsize > sbiggest {
				sbiggest = vaddress + vsize
			}
		} else {
			if vaddress + vsize > biggest {
				biggest = vaddress + vsize
			}
		}


		// move on
		ptr += len(namedata)
	}
	this.VIndexNext = ptr
	this.VTableNext  = biggest
	this.VStringNext = sbiggest
}

// Returns a list of Variable names (keys)
func (this *VarMap) Keys() []string {
	v := make([]string, len(this.Content))
	i := 0
	for k := range this.Content {
		v[i] = k
		i++
	}
	return v
}

// Returns a list of Variables (values)
func (this *VarMap) Values() []*Variable {
	v := make([]*Variable, len(this.Content))
	i := 0
	for _, c := range this.Content {
		v[i] = c
		i++
	}
	return v
}

// Empty the Variable Map
func (this *VarMap) Clear() {
	this.Content = make(map[string]*Variable)
	this.SetBase(this.VIndexBase, 0x19000)
}

/* Variable funcs */

// create variable
func NewVariable() *Variable {
	return &Variable{Name: "", Kind: VT_STRING, AssumeLowIndex: true, Length: 1, Mutable: true}
}

func NewVariableP(n string, t VariableType, v string, m bool) *Variable {
	return &Variable{Name: n, Kind: t, AssumeLowIndex: true, Length: 1, Mutable: m, ContentScalar: NewVarValue(v), Dimensions: []int{1}, Content: []*VarValue{NewVarValue("")}}
}

func NewVariablePZ(n string, t VariableType, v string, m bool) *Variable {
	this := &Variable{Name: n, Kind: t, AssumeLowIndex: true, Length: 1, Mutable: m, ZeroIsScalar: true, ContentScalar: NewVarValue(v), Dimensions: []int{1}, Content: []*VarValue{NewVarValue(v)}}
	return this
}

func NewVariablePA(n string, t VariableType, v string, m bool, dl []int) *Variable {

	this, _ := NewVariableArray(n, t, v, m, dl)

	return this
}

func (v Variable) IsScalar() bool {
	return (v.Length == 1)
}

func (v Variable) IsArray() bool {
	return !v.IsScalar()
}

func (v Variable) IsMutable() bool {
	return v.Mutable
}

func (v Variable) GetKind() VariableType {
	return v.Kind
}

func (vv Variable) FlattenIndex(index []int) (int, error) {
	if len(vv.Dimensions) != len(index) {
		//panic("Invalid dimensions used to address variable")
		return -1, errors.New("Invalid dimensions used to address variable")
	}

	for c, value := range index {
		if (value >= vv.Dimensions[c]) || (value < 0) {
			//panic("Invalid dimensions used to address ARRAY variable")
			return -1, errors.New("Invalid indices used to address array variable")
		}
	}

	v := index[len(vv.Dimensions)-1]
	m := 1

	for c := len(vv.Dimensions) - 2; c >= 0; c-- {
		v = v + vv.Dimensions[c+1]*index[c]*m
		m = m * vv.Dimensions[c+1]
	}

	return v, nil
}

func (vv Variable) GetContentScalar() (string, error) {
	if (vv.Length != 1) && (!vv.AssumeLowIndex) {
		return "", errors.New("Attempt to access array in a scalar context")
	}

	if vv.ZeroIsScalar {
		return vv.Content[0].String(), nil
	}

	return vv.ContentScalar.String(), nil
}

func (vv Variable) ParseContent(v string) (string, error) {

	if (vv.Length != 1) && (!vv.AssumeLowIndex) {
		return "", errors.New("Attempt to access array in scalar context")
	}

	//this.SetContent(-1, v);
	if !vv.Mutable {
		return "", errors.New("Attempt to redeclare constant")
	}

	if vv.Kind == VT_STRING {
        if len(v) > MAX_STRING_LENGTH {
           return v, errors.New("STRING TOO LONG")
        }
		return v, nil
	}

	if vv.Kind == VT_INTEGER {
		if v == "" {
			v = "0"
		}
		v = utils.StrToIntStr(v)
		return v, nil
	}

	if vv.Kind == VT_FLOAT {
		if v == "" {
			v = "0"
		}

		d, _ := strconv.ParseFloat(utils.NumberPart(utils.FlattenASCII(v)), 32)
		if d == math.Floor(d) {
			v = utils.StrToIntStr(v)
		} else {
			v = utils.StrToFloatStr(v)
		}
		return v, nil
	}

	if vv.Kind == VT_BOOLEAN {
		if v == "" {
			v = "false"
		}

		b := (strings.ToLower(v) == "true") || (strings.ToLower(v) == "yes")
		v = "false"
		if b {
			v = "true"
		}
		return v, nil
	}

	if vv.Kind == VT_EXPRESSION {
		if v == "" {
			v = "0"
		}
		return v, nil
	}

	return v, nil
}

func (vv *Variable) SyncScalar() {
	// into memory
	if vv.VarAddress != 0 {

		offset := vv.VarAddress
		v      := ""
		switch vv.Kind {
		case VT_INTEGER:
			v = utils.IntToStr( int(vv.Map.Mgr.GetMemory(offset)) )
		case VT_FLOAT:
			v = utils.FloatToStr( float64( Uint2Float(vv.Map.Mgr.GetMemory(offset))) )
		case VT_STRING:
        
            realaddr := int( vv.Map.Mgr.GetMemory(offset) )
        
        	if realaddr == 0 {
               v = ""
            } else {
               count := int(vv.Map.Mgr.GetMemory(realaddr))
               v = ""
               for i:=0; i<count; i++ {
               	   v += string( rune( vv.Map.Mgr.GetMemory(realaddr+1+i) ) )
               }
            }
            
		case VT_EXPRESSION:
            realaddr := int( vv.Map.Mgr.GetMemory(offset) )
        
        	if realaddr == 0 {
               v = ""
            } else {
               count := int(vv.Map.Mgr.GetMemory(realaddr))
               v = ""
               for i:=0; i<count; i++ {
               	   v += string( rune( vv.Map.Mgr.GetMemory(realaddr+1+i) ) )
               }
            }
		}

		vv.SetContentScalar(v)

//		ss, _ := vv.GetContentScalar()

//		//fmt.Printf( "--> Set value to [%s]\n", ss)

	}
}

func (vv *Variable) SyncArray() {
	for i:=0; i<vv.Length; i++ {
		vv.SyncIndex(i)
	}
}


func (vv *Variable) SyncIndex(rindex int) {
	// into memory
	if vv.VarAddress != 0 {

       size := vv.Kind.Size()
       if vv.Kind == VT_STRING {
       	  size = 1
       }
    
		offset := vv.VarAddress + size*rindex
		v      := ""
		switch vv.Kind {
		case VT_INTEGER:
			v = utils.IntToStr( int(vv.Map.Mgr.GetMemory(offset)) )
		case VT_FLOAT:
			v = utils.FloatToStr( float64( Uint2Float(vv.Map.Mgr.GetMemory(offset))) )
		case VT_STRING:
            realaddr := int( vv.Map.Mgr.GetMemory(offset) )
        
        	if realaddr == 0 {
               v = ""
            } else {
               count := int(vv.Map.Mgr.GetMemory(realaddr))
               v = ""
               for i:=0; i<count; i++ {
               	   v += string( rune( vv.Map.Mgr.GetMemory(realaddr+1+i) ) )
               }
            }
		case VT_EXPRESSION:
            realaddr := int( vv.Map.Mgr.GetMemory(offset) )
        
        	if realaddr == 0 {
               v = ""
            } else {
               count := int(vv.Map.Mgr.GetMemory(realaddr))
               v = ""
               for i:=0; i<count; i++ {
               	   v += string( rune( vv.Map.Mgr.GetMemory(realaddr+1+i) ) )
               }
            }
		}

		vv.SetContent(rindex, v)

//		//fmt.Printf( "--> Set value %d to [%s]\n", rindex, v )

	}
}

func (vv *Variable) FreshenScalar(v string) {
			// into memory
			if vv.VarAddress != 0 {

				offset := vv.VarAddress
				switch vv.Kind {
				case VT_INTEGER:
					vv.Map.Mgr.SetMemory(offset, uint(utils.StrToInt(v)))
				case VT_FLOAT:
					vv.Map.Mgr.SetMemory(offset, uint(Float2uint(utils.StrToFloat(v))))
				case VT_STRING:
					if len(v) > 255 {
						v = v[0:255]
					}
                    
                    // Get allocation, changed or not
                    ptr := vv.Map.RequestStringAllocation(vv.Name, len(v)+1, 0)
                    vv.Map.Mgr.SetMemory(offset, uint(ptr.Address))
                    
                    // do this properly..
                    vv.Map.Mgr.SetMemory(ptr.Address, uint(len(v)))
                    for i, ch := range v {
						vv.Map.Mgr.SetMemory(ptr.Address+1+i, uint(ch))
					}
                    
				case VT_EXPRESSION:
					if len(v) > 255 {
						v = v[0:255]
					}
                    
                    // Get allocation, changed or not
                    ptr := vv.Map.RequestStringAllocation(vv.Name, len(v)+1, 0)
                    vv.Map.Mgr.SetMemory(offset, uint(ptr.Address))
                    
                    // do this properly..
                    vv.Map.Mgr.SetMemory(ptr.Address, uint(len(v)))
                    for i, ch := range v {
						vv.Map.Mgr.SetMemory(ptr.Address+1+i, uint(ch))
					}
				}

			}
}

func (vv *Variable) SetContentScalar(v string) error {
	s, err := vv.ParseContent(v)
	if err == nil {
		if !vv.Shared {
			vv.ContentScalar.Assign(s, vv.Stacked)
			if vv.ZeroIsScalar {
				vv.Content[0].Assign(s, vv.Stacked)
			}
			// into memory
			vv.FreshenScalar(v)
		} else {
			// Share handling

		}
		return nil
	}
	return err
}

func (vv *Variable) FreshenArray() {
	for i:=0; i<vv.Length; i++ {
		vv.FreshenIndex(i, vv.Content[i].String())
	}
}

func (vv *Variable) FreshenIndex( rindex int, v string ) {
			if vv.VarAddress != 0 {

                size := vv.Kind.Size()
                if vv.Kind == VT_STRING {
                   size = 1
                }
            
				offset := vv.VarAddress + size*rindex
				switch vv.Kind {
				case VT_INTEGER:
					vv.Map.Mgr.SetMemory(offset, uint(utils.StrToInt(v)))
				case VT_FLOAT:
					vv.Map.Mgr.SetMemory(offset, uint(Float2uint(utils.StrToFloat(v))))
				case VT_STRING:
					if len(v) > 255 {
						v = v[0:255]
					}
                    
                    // Get allocation, changed or not
                    ptr := vv.Map.RequestStringAllocation(vv.Name, len(v)+1, 0)
                    vv.Map.Mgr.SetMemory(offset, uint(ptr.Address))
                    
                    // do this properly..
                    vv.Map.Mgr.SetMemory(ptr.Address, uint(len(v)))
                    for i, ch := range v {
						vv.Map.Mgr.SetMemory(ptr.Address+1+i, uint(ch))
					}
				case VT_EXPRESSION:
					if len(v) > 255 {
						v = v[0:255]
					}
                    
                    // Get allocation, changed or not
                    ptr := vv.Map.RequestStringAllocation(vv.Name, len(v)+1, 0)
                    vv.Map.Mgr.SetMemory(offset, uint(ptr.Address))
                    
                    // do this properly..
                    vv.Map.Mgr.SetMemory(ptr.Address, uint(len(v)))
                    for i, ch := range v {
						vv.Map.Mgr.SetMemory(ptr.Address+1+i, uint(ch))
					}
				}

			}
}

func (vv *Variable) SetContent(rindex int, v string) error {
	s, err := vv.ParseContent(v)
	if err == nil {
		if !vv.Shared {
			if vv.Content[rindex] == nil {
				vv.Content[rindex] = NewVarValue(v)
			}
			vv.Content[rindex].Assign(s, vv.Stacked)

			// Update memory
			vv.FreshenIndex(rindex, v)

			log.Printf("Set %d to %s\n", rindex, s)
		} else {
			// Share handling
		}
		return nil
	}
	return err
}

func (vv *Variable) SetContentByIndex(def int, max int, index []int, v string) error {
	if (len(vv.Dimensions) == 1) && (vv.Dimensions[0] == 1) && (vv.AssumeLowIndex) && (index[0] > 0) {
		/* auto redim array */
		vv.Dimensions = make([]int, len(index))
		for i := 0; i <= len(vv.Dimensions)-1; i++ {
			vv.Dimensions[i] = def + 1
		}
		/* Redim data space */
		vv.Length = 1
		for c := 0; c <= len(vv.Dimensions)-1; c++ {
			vv.Length = vv.Length * vv.Dimensions[c]
		}
		//System.out.println("Array length.equals(",Length);
		vv.Content = make([]*VarValue, vv.Length)
		//this.Content = v; /* forces evaluation */
		vvv := ""
		if vv.Kind != VT_STRING {
			vvv = "0"
		}
		for c := 1; c <= vv.Length-1; c++ {
			e := vv.SetContent(c, vvv)
            if e != nil {
               return e
            }
		}
	} else if (len(vv.Dimensions) == 1) && (vv.Dimensions[0] <= index[0]) && (vv.AssumeLowIndex) {
		newsize := index[0] + def
		stuff := make([]*VarValue, newsize)
		for z := 0; z < len(stuff); z++ {
			if z < len(vv.Content) {
				stuff[z] = vv.Content[z]
			} else {
				stuff[z] = NewVarValue("")
			}
		}
		vv.Dimensions[0] = newsize
		vv.Content = stuff
		vv.Length = newsize
	}

	fidx, err := vv.FlattenIndex(index)

	log.Println("Flattened index is", fidx)

	if err != nil {
		return err
	}

	e := vv.SetContent(fidx, v)

	return e
}

func (vv *Variable) GetContentByIndex(def int, max int, index []int) (string, error) {
	if (len(vv.Dimensions) == 1) && (vv.Dimensions[0] == 1) && (vv.AssumeLowIndex) && (index[0] > 0) && (!vv.ZeroIsScalar) {
		/* auto redim array */
		vv.Dimensions = make([]int, len(index))
		for i := 0; i <= len(vv.Dimensions)-1; i++ {
			vv.Dimensions[i] = def + 1
		}
		/* Redim data space */
		vv.Length = 1
		for c := 0; c <= len(vv.Dimensions)-1; c++ {
			vv.Length = vv.Length * vv.Dimensions[c]
		}
		//System.out.println("Array length.equals(",Length);
		vv.Content = make([]*VarValue, vv.Length)
		//this.Content = v; /* forces evaluation */
		vvv := ""
		if vv.Kind != VT_STRING {
			vvv = "0"
		}
		for c := 1; c <= vv.Length-1; c++ {
			_ = vv.SetContent(c, vvv)
		}
	}

	fidx, err := vv.FlattenIndex(index)

	if err != nil {
		return "", err
	}

    var result string
    if vv.Content[fidx] != nil {
	  result = vv.Content[fidx].String()
    }
	if result == "" {
		if vv.Kind != VT_STRING {
			result = "0"
		}
	}

	return result, nil
}

func NewVariableArray(n string, t VariableType, v string, mutable bool, dim []int) (*Variable, error) {

	this := NewVariable()

	this.Dimensions = dim
	this.Length = 1
	for c := 0; c <= len(this.Dimensions)-1; c++ {
		this.Length = this.Length * this.Dimensions[c]
	}
	//System.out.println("Array length.equals(",Length);
	log.Printf("Creating %s = %d\r\n", n, this.Length)
	this.Content = make([]*VarValue, this.Length)
	for x := 0; x < this.Length; x++ {
		this.Content[x] = NewVarValue(v)
	}
	this.ContentScalar = NewVarValue(v)
	this.Name = n
	this.Kind = t
	this.Mutable = true
	//this.Content = v; /* forces evaluation */
	for c := 0; c <= this.Length-1; c++ {
		err := this.SetContent(c, v)
		if err != nil {
			return nil, err
		}
	}
	err := this.SetContentScalar(v)
	this.Mutable = mutable

	if err != nil {
		return nil, err
	}

	return this, nil
}

func NewVariableScalar(n string, t VariableType, v string, mutable bool) (*Variable, error) {
	this := NewVariable()

	this.Length = 1
	this.Dimensions = make([]int, 1)
	this.Dimensions[0] = 1
	this.Content = make([]*VarValue, 1)
	this.Content[0] = NewVarValue(v)
	this.Name = n
	this.Kind = t
	this.Mutable = mutable
	this.ContentScalar = NewVarValue(v)
	err := this.SetContentScalar(v) /* forces evaluation */
	if err != nil {
		return nil, err
	}
	return this, nil
}

func (this *Variable) MarshalBinary() ([]uint, error) {
	//
	data := make([]uint, 0)
	data = append(data, uint(this.Kind))
	d := []byte(this.Name)
	data = append(data, uint(len(d)))
	for _, v := range d {
		data = append(data, uint(v))
	}
	data = append(data, uint(this.Length))

	if this.Length > 1 {
		data = append(data, uint(len(this.Dimensions)))
		for _, v := range this.Dimensions {
			data = append(data, uint(v))
		}

		// Encode datas
		for _, vv := range this.Content {
			s := vv.GetValue() // s is a string
			// length
			d := []byte(s)
			data = append(data, uint(len(d)))
			for _, v := range d {
				data = append(data, uint(v))
			}
		}
	} else {
		s := this.ContentScalar.GetValue() // s is a string
		// length
		d := []byte(s)
		data = append(data, uint(len(d)))
		for _, v := range d {
			data = append(data, uint(v))
		}
	}

	return data, nil
}

func (this *Variable) UnmarshalBinary(data []uint) (int, error) {

	this.Kind = VariableType(data[0])

	// Decode the name
	namelen := int(data[1])
	namedata := data[2 : 2+namelen]
	name := ""
	for _, v := range namedata {
		name += string(rune(v))
	}

	idx := namelen + 2

	// get length
	length := int(data[idx])

	idx++
	dims := make([]int, 0)
	if length > 1 {
		numdims := int(data[idx])
		for i := 0; i < numdims; i++ {
			idx++
			dims = append(dims, int(data[idx]))
		}
	} else {
		dims = append(dims, 1)
	}

	this.Dimensions = dims
	this.Length = length
	this.Name = name

	if length > 1 {
		this.Content = make([]*VarValue, length)
		for i := 0; i < length; i++ {
			l := int(data[idx])
			idx++
			valuedata := data[idx : idx+l]
			s := ""
			for _, v := range valuedata {
				s += string(rune(v))
			}
			this.Content[i] = NewVarValue(s)
			idx += l
		}
	} else {
		l := int(data[idx])
		idx++
		valuedata := data[idx : idx+l]
		s := ""
		for _, v := range valuedata {
			s += string(rune(v))
		}
		this.ContentScalar = NewVarValue(s)
		idx += l
	}

	return idx, nil

}

func (this *VarMap) MarshalBinary() ([]uint, error) {

	data := make([]uint, 0)

	keys := make([]string, 0)
	for k, _ := range this.Content {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := this.Content[k]
		chunk, e := v.MarshalBinary()
		if e != nil {
			return data, e
		}
		data = append(data, chunk...)
	}

	return data, nil
}

// Helper for converting floats to []byte
func Float2uint(f float32) uint {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, f)
	if err != nil {
		return 0
	}
	b := buf.Bytes()
	return uint(b[0])<<24 | uint(b[1])<<16 | uint(b[2])<<8 | uint(b[3])
}

func bytes2Uint(b []byte) uint {
	return uint(b[0])<<24 | uint(b[1])<<16 | uint(b[2])<<8 | uint(b[3])
}

func uint2Bytes(u uint) []byte {
	data := make([]byte, 4)

	data[0] = byte((u & 0xff000000) >> 24)
	data[1] = byte((u & 0x00ff0000) >> 16)
	data[2] = byte((u & 0x0000ff00) >> 8)
	data[3] = byte(u & 0x000000ff)

	return data
}

func Uint2Float(u uint) float32 {
	data := make([]byte, 4)

	data[0] = byte((u & 0xff000000) >> 24)
	data[1] = byte((u & 0x00ff0000) >> 16)
	data[2] = byte((u & 0x0000ff00) >> 8)
	data[3] = byte(u & 0x000000ff)

	var f float32
	b := bytes.NewBuffer(data)
	_ = binary.Read(b, binary.LittleEndian, &f)
	return f
}

func PackName( name string, l int) []uint {
	if len(name) > l {
		name = name[0:l]
	}
	for len(name) < l {
		name += string(rune(0))
	}
	b := []byte(name)

	data := make([]uint, 0)

	for len(b) > 0 {
		if len(b) >= 4 {
			chunk := b[0:4]
			b = b[4:]
			data = append(data, bytes2Uint(chunk))
		}
	}

	return data
}

func UnpackName( data []uint ) string {

	out := make([]byte, 0)

	for _, u := range data {
		b := uint2Bytes(u)
        
        for _, bb := range b {
        	if bb != 0 {
        	   out = append(out, bb)
            }
        }
	}

	s := string(out)

	return strings.Trim(s, " ")
}

// Scan through blocks allocated and group contiguous free blocks into a single block
func (this *VarMap) GroupStringFreeblocks() {

	 newmap := make(map[int]StrAlloc)
     
     sorted := make([]int, 0)
     for i, _ := range this.StringReuse {
     	 sorted = append( sorted, i )
     }
     sort.Ints(sorted)
     
     for i := len(sorted)-1; i>=0; i-- {
     
     	 address := sorted[i]        
         oldblock := this.StringReuse[address]    
         
         nextblockstart := oldblock.Address + oldblock.Length
         
         nextblock, ex := newmap[nextblockstart]
         
         if ex {
         	nextblock.Length += oldblock.Length
            nextblock.Address = oldblock.Address
            
            delete( newmap, nextblockstart )
            newmap[ nextblock.Address ] = nextblock
            
            //fmt.Printf("[STRING] Merged 2 blocks to yield %d bytes at 0x%x\n", nextblock.Length, nextblock.Address )
         } else {
           // add block
           newmap[ address ] = oldblock
         }
     
     }
     
     this.StringReuse = newmap

}

func (this *VarMap) RequestStringAllocation( name string, length int, index int ) StrAlloc {

	 keyname := GetVarKey(name, index)
     
     if len(this.StringReuse) > 2 {
     	this.GroupStringFreeblocks()
     }
     
     sa, ok := this.StringMemory[keyname]
     
     if !ok {
     	// try memory reuse
        for i, asa := range this.StringReuse {
        	if asa.Length >= length {
            
               nsa := asa
               nsa.Length = length // only take needed amount
               
               // adjust asa for free amount
               asa.Address += length
               asa.Length  -= length
            
               this.StringMemory[keyname] = nsa
               delete( this.StringReuse, i )
 
         	   //fmt.Printf("[STRING] Using freed allocation for %s: Address: 0x%x, Length: %d\n", keyname, nsa.Address, nsa.Length )
               
               if asa.Length > 0 {
               	  this.StringReuse[ asa.Address ] = asa
                  //fmt.Printf("[STRING] Heaping free partial block from %s: Address: 0x%x, Length: %d\n", keyname, asa.Address, asa.Length )
               }
               
               return nsa
            }
        }
        
        // get fresh one
        varAddress := this.VStringNext
		this.VStringNext += length
        sa.Address = varAddress
        sa.Length  = length
        this.StringMemory[keyname] = sa
        
        //fmt.Printf("[STRING] Created allocation for %s: Address: 0x%x, Length: %d\n", keyname, sa.Address, sa.Length )
        
        return sa
     } else {
        // existing allocation
        if sa.Length >= length {
          //fmt.Printf("[STRING] Keeping allocation for %s: Address: 0x%x, Length: %d\n", keyname, sa.Address, sa.Length )
          
          diff := sa.Length - length
          if diff > 2 {
          	 // trim and free
             nb := StrAlloc{ Address: sa.Address+length, Length: diff }
             this.StringReuse[ nb.Address ] = nb
             
             sa.Length = length
             this.StringMemory[keyname] = sa // trim block      
             
             //fmt.Printf("[STRING] Heaping free partial block from %s: Address: 0x%x, Length: %d\n", keyname, nb.Address, nb.Length )
          }
          
          return sa
        }
        
        // discard current block
        delete( this.StringMemory, keyname )
         
        // search discarded blocks for one of sufficient size
        this.StringReuse[ sa.Address ] = sa
        for i, asa := range this.StringReuse {
        	if asa.Length >= length {
            
               nsa := asa
               nsa.Length = length // only take needed amount
               
               // adjust asa for free amount
               asa.Address += length
               asa.Length  -= length
            
               this.StringMemory[keyname] = nsa
               delete( this.StringReuse, i )
 
         	   //fmt.Printf("[STRING] Using freed allocation for %s: Address: 0x%x, Length: %d\n", keyname, nsa.Address, nsa.Length )
               
               if asa.Length > 0 {
               	  this.StringReuse[ asa.Address ] = asa
                  //fmt.Printf("[STRING] Heaping free partial block from %s: Address: 0x%x, Length: %d\n", keyname, asa.Address, asa.Length )
               }
               
               return nsa
            }
        }
        
        // if not found then allocate new block
        varAddress := this.VStringNext
		this.VStringNext += length
        sa.Address = varAddress
        sa.Length  = length
        this.StringMemory[keyname] = sa
     	//fmt.Printf("[STRING] Created allocation for %s: Address: 0x%x, Length: %d\n", keyname, sa.Address, sa.Length )
        return sa
     }

}
