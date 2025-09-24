// variable.go
package types

import (
	"errors"
	"reflect"

	"paleotronic.com/utils"
	//	"paleotronic.com/fmt"
)

type VariableType int

const (
	VT_STRING     VariableType = 1 + iota
	VT_BOOLEAN    VariableType = 1 + iota
	VT_INTEGER    VariableType = 1 + iota
	VT_FLOAT      VariableType = 1 + iota
	VT_EXPRESSION VariableType = 1 + iota
)

type VarDriver int

const (
	// Applesoft id's
	VD_MICROSOFT VarDriver = iota
	VD_WOZ
)

type VarDriverContext int

const (
	VDC_SCALAR VarDriverContext = iota
	VDC_ARRAY
)

const (
	MAX_EXPRESSION_LENGTH = 256
	MAX_STRING_LENGTH     = 256
	MAX_COMPRESSED_STRING = MAX_STRING_LENGTH / 4
	MAX_FLOAT_LENGTH      = 1
	MAX_INTEGER_LENGTH    = 1
	DEFAULT_NAME_ULEN     = 4
	DEFAULT_INDEX         = 1024
)

type Variable struct {
	Name           string
	Driver         VarDriver
	Context        VarDriverContext
	Map            VarManager
	Kind           VariableType
	AssumeLowIndex bool
	Owner          string
}

func (v *Variable) IsArray() bool {
	return (v.Context == VDC_ARRAY)
}

func (v *Variable) Dimensions() []int {
	return v.Map.Dimensions(v.Name)
}

func (v *Variable) DeflectValue(content string) (interface{}, error) {

	switch v.Driver {
	case VD_MICROSOFT:
		// stuff
		switch v.Kind {
		case VT_INTEGER:
			return NewInteger2b(int16(utils.StrToInt(content))), nil
		case VT_FLOAT:
			return NewFloat5b(utils.StrToFloat64(content)), nil
		case VT_STRING:
			msbin := v.Map.(*VarManagerMSBIN)
			// if val, err := msbin.GetValue(v.Name); err == nil {
			// 	var ptr = val.(*StringPtr3b)
			// 	if int(ptr.GetLength()) >= len(content) {
			// 		//log2.Printf("using existing pointer for string... %+v", ptr.GetPointer())
			// 		return ptr, nil
			// 	}
			// }
			//msbin.CleanStrings()
			i, err := msbin.allocStringMemory(content)
			msbin.DumpStrings()
			return i, err
		default:
			return nil, errors.New("TYPE ERR")
		}
	case VD_WOZ:
		// handle  mapping of integer basic
		switch v.Kind {
		case VT_INTEGER:
			return NewInteger2b(int16(utils.StrToInt(content))), nil
		case VT_STRING:
			return content, nil
		default:
			return nil, errors.New("TYPE ERR")
		}
	}

	return nil, errors.New("TYPE ERR")
}

func (v *Variable) ReflectValue(content interface{}) (string, error) {

	// Use type reflection here
	contentType := reflect.TypeOf(content).String()

	switch contentType {
	case "string":
		return content.(string), nil
	case "*types.Float5b":
		return content.(*Float5b).String(), nil
	case "*types.Integer2b":
		return content.(*Integer2b).String(), nil
	case "*types.StringPtr3b":
		return content.(*StringPtr3b).FetchString(v.Map.GetMM(), v.Map.GetIndex()), nil
	}

	return "", errors.New("Unknown type: " + contentType)
}

func (v *Variable) SetContentScalar(value string) error {

	i, e := v.DeflectValue(value)
	//	//fmt.Printf("%s.SetContentScalar(%s)\n", v.Name, value)
	if e != nil {
		return e
	}
	//	//fmt.Printf("About to call v.Map.SetValue(%s, %v)n", v.Name, i)
	e = v.Map.SetValue(v.Name, i)
	if e != nil {
		// doesnt exist,,,
		e2 := v.Map.Create(v.Name, v.Kind, i)
		if e2 != nil {
			return e
		}
	}
	//_, _ := v.Map.GetValue(v.Name)
	//	//fmt.Printf("v.Map.GetValue(%s) gives %vn", v.Name, z)
	return nil

}

func (v *Variable) SetContentByIndex(a, b int, index []int, value string) error {

	i, e := v.DeflectValue(value)
	if e != nil {
		return e
	}
	e = v.Map.SetValueIndexed(v.Name, index, i)
	if e != nil {
		switch v.Kind {
		case VT_FLOAT:
			v.Map.CreateIndexed(v.Name, v.Kind, []int{10}, NewFloat5b(0))
		case VT_INTEGER:
			v.Map.CreateIndexed(v.Name, v.Kind, []int{10}, NewInteger2b(0))
		case VT_STRING:
			v.Map.CreateStringIndexed(v.Name, []int{10}, "")
		}
		e = v.Map.SetValueIndexed(v.Name, index, i)
	}
	return nil

}

func (v *Variable) GetContentScalar() (string, error) {
	content, e := v.Map.GetValue(v.Name)
	if e != nil {
		return "", e
	}

	ss, ee := v.ReflectValue(content)
	//	//fmt.Printf("%s.GetContentScalar() returns %s\n", v.Name, ss)

	return ss, ee
}

func (v *Variable) GetContentByIndex(a int, b int, index []int) (string, error) {
	content, e := v.Map.GetValueIndexed(v.Name, index)
	if e != nil {
		return "", e
	}

	return v.ReflectValue(content)
}

// ---------------------------------------------------------------------------------

// NewVariablePA creates a new array type variable, returns nil on err
func NewVariablePA(vm VarManager, name string, vt VariableType, content string, mutable bool, dims []int) (*Variable, error) {

	v := &Variable{
		Map:            vm,
		AssumeLowIndex: true,
		Context:        VDC_ARRAY,
		Driver:         vm.GetDriver(),
		Name:           name,
		Kind:           vt,
	}

	i, e := v.DeflectValue(content)
	if e != nil {
		return nil, e
	}

	e = vm.CreateIndexed(name, vt, dims, i)
	//fmt.Println(e)

	return v, e
}

// NewVariablePA creates a new array type variable, returns nil on err
func NewVariableP(vm VarManager, name string, vt VariableType, content string, mutable bool) (*Variable, error) {

	v := &Variable{
		Map:            vm,
		AssumeLowIndex: true,
		Context:        VDC_SCALAR,
		Driver:         vm.GetDriver(),
		Name:           name,
		Kind:           vt,
	}

	i, e := v.DeflectValue(content)
	if e != nil {
		return nil, e
	}

	e = vm.Create(name, vt, i)

	return v, e
}

// NewVariablePZ is needed for backware compatibility
func NewVariablePZ(vm VarManager, name string, vt VariableType, content string, mutable bool) (*Variable, error) {
	return NewVariableP(vm, name, vt, content, mutable)
}
