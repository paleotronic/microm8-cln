package logo

import (
	"strings"
	"sync"

	"paleotronic.com/core/types"
)

// LogoVarTable is a name based map of Variables
type LogoVarTable struct {
	s      sync.RWMutex
	m      map[string]*types.Token
	buried map[string]bool
}

func NewLogoVarTable() *LogoVarTable {
	return &LogoVarTable{
		m: map[string]*types.Token{},
	}
}

func (vt *LogoVarTable) Get(name string) *types.Token {
	vt.s.RLock()
	defer vt.s.RUnlock()
	name = strings.ToLower(strings.TrimLeft(strings.TrimLeft(name, ":"), "\""))
	return vt.m[name]
}

func (vt *LogoVarTable) Set(name string, value *types.Token) {
	vt.s.Lock()
	defer vt.s.Unlock()
	name = strings.ToLower(strings.TrimLeft(strings.TrimLeft(name, ":"), "\""))
	vt.m[name] = value.Copy()
	//log2.Printf("set var %s -> %s", name, tokenStr("", value))
}

func (vt *LogoVarTable) Exists(name string) bool {
	vt.s.RLock()
	defer vt.s.RUnlock()
	name = strings.ToLower(strings.TrimLeft(strings.TrimLeft(name, ":"), "\""))
	_, ok := vt.m[name]
	return ok
}

func (vt *LogoVarTable) Bury(name string) {
	vt.s.Lock()
	defer vt.s.Unlock()
	name = strings.ToLower(strings.TrimLeft(strings.TrimLeft(name, ":"), "\""))
	vt.buried[name] = true
}

func (vt *LogoVarTable) Unbury(name string) {
	vt.s.Lock()
	defer vt.s.Unlock()
	name = strings.ToLower(strings.TrimLeft(strings.TrimLeft(name, ":"), "\""))
	vt.buried[name] = false
}

func (vt *LogoVarTable) Erase(name string) {
	vt.s.Lock()
	defer vt.s.Unlock()
	name = strings.ToLower(strings.TrimLeft(strings.TrimLeft(name, ":"), "\""))
	delete(vt.buried, name)
	delete(vt.m, name)
}
