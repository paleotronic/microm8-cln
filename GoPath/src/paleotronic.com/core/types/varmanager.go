package types

import (
	"paleotronic.com/core/memory"
)

// VarManager is an abstracted variable manager for a given system
// that implements storage appropriately.
type VarManager interface {
	Exists(name string) bool
	GetValue(name string) (interface{}, error)
	SetValue(name string, value interface{}) error
	ExistsIndexed(name string) bool
	GetValueIndexed(name string, index []int) (interface{}, error)
	SetValueIndexed(name string, index []int, value interface{}) error
	GetVarNames() []string
	GetVarNamesIndexed() []string
	Create(name string, kind VariableType, content interface{}) error
	CreateIndexed(name string, kind VariableType, capacity []int, content interface{}) error
	CreateString(name, str string) error
	CreateStringIndexed(name string, capacity []int, str string) error
	ContainsKey(name string) bool
	Contains(name string) bool
	Put(name string, v *Variable)
	Get(name string) *Variable
	GetMM() *memory.MemoryMap
	GetIndex() int
	Dimensions(name string) []int
	GetDriver() VarDriver
	Clear()
	CleanStrings() int
	SetLoBound( address int )
	SetHiBound( address int )
    GetFree() int
}
