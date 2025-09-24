package logo

import (
	"sort"
	"strings"

	"paleotronic.com/core/dialect/parcel"
	"paleotronic.com/core/types"
)

// GetWorkspaceBody returns the textual representation of the
// workspace
func (d *LogoDriver) GetWorkspaceBody(includeProcs bool, includeVars bool) []string {

	var out = []string{}
	var addblock = func(b []string) {
		if len(out) > 0 {
			out = append(out, "")
		}
		out = append(out, b...)
	}

	// output proc list (non-buried)
	var keys = []string{}
	if includeProcs {
		for procname := range d.Procs {
			keys = append(keys, procname)
		}
		sort.Strings(keys)
		for _, key := range keys {
			p, ok := d.GetProc(key)
			if !ok || p.Buried {
				continue
			}
			lines := p.GetCode()
			addblock(lines)
		}
	}

	// output var list (non-buried)
	if includeVars {
		keys = []string{}
		for varname := range d.Globals.m {
			keys = append(keys, varname)
		}
		sort.Strings(keys)
		varlines := []string{}
		for _, key := range keys {
			if d.Globals.buried[key] {
				continue
			}
			v := d.Globals.Get(key)
			if v == nil {
				continue
			}
			if v.IsPropList {
				txt := d.DumpObjectStruct(v, false, "PPROP \""+key+" ")
				varlines = append(varlines, txt)
			} else {
				txt := d.DumpObjectStruct(v, false, "MAKE \""+key+" ")
				varlines = append(varlines, txt)
			}
		}
		addblock(varlines)
	}

	return out
}

func (d *LogoDriver) DumpObjectStruct(t *types.Token, printMode bool, buffer string) string {
	if t.Type == types.COMMANDLIST || t.Type == types.EXPRESSIONLIST {
		if !printMode {
			buffer += "("
		}
		for i, tt := range t.List.Content {
			if i > 0 {
				buffer += " "
			}
			buffer = d.DumpObjectStruct(tt, printMode, buffer)
		}
		if !printMode {
			buffer += ")"
		}
	} else if t.Type == types.LIST {
		if !printMode {
			buffer += "["
		}
		for i, tt := range t.List.Content {
			if i > 0 {
				buffer += " "
			}
			buffer = d.DumpObjectStruct(tt, printMode, buffer)
		}
		if !printMode {
			buffer += "]"
		}
	} else {
		if t.Type == types.WORD && !printMode {
			buffer += "\"" + t.Content
		} else {
			buffer += t.Content
		}
	}
	return buffer
}

func (d *LogoDriver) CountUnresolvedInObject(t *types.Token, count int) int {
	if t.Type == types.COMMANDLIST || t.Type == types.EXPRESSIONLIST {
		for _, tt := range t.List.Content {
			count = d.CountUnresolvedInObject(tt, count)
		}
	} else if t.Type == types.LIST {
		for _, tt := range t.List.Content {
			count = d.CountUnresolvedInObject(tt, count)
		}
	} else {
		if t.Type == types.IDENTIFIER {
			count++
		}
	}
	return count
}

func (d *LogoDriver) ReresolveSymbols(l *parcel.Lexer) {
	names := d.FindProcsWithUnresolvedSymbols()
	if len(names) > 0 {
		d.Printf("Reresolving symbols for %+v", names)
		d.SuppressConsole = true
		_ = d.ReparseProcs(l, names)
		d.SuppressConsole = false
	}
}

func (d *LogoDriver) SetBuryProc(name string, buried bool) {
	p, ok := d.GetProc(name)
	if !ok {
		return
	}
	p.Buried = buried
}

func (d *LogoDriver) SetBuryVar(name string, buried bool) {
	if buried {
		d.Globals.Bury(name)
	} else {
		d.Globals.Unbury(name)
	}
}

func (d *LogoDriver) EraseGlobal(name string) {
	d.Globals.Erase(name)
}

func (d *LogoDriver) EraseProc(name string) {
	if _, ok := d.GetProc(name); ok {
		delete(d.Procs, strings.ToLower(name))
	}
}

func (d *LogoDriver) HasCBreak() bool {
	return d.ent != nil && d.ent.GetMemoryMap().KeyBufferHasBreak(d.ent.GetMemIndex())
}

func (d *LogoDriver) HasReboot() bool {
	if d.ent != nil && d.ent.VM().IsDying() {
		return true
	}
	d.ent.WaitForWorld()
	return d.ent != nil && d.ent.GetMemoryMap().IntGetSlotRestart(d.ent.GetMemIndex())
}

func (d *LogoDriver) SetBuryAllProcs(bury bool) {
	for _, p := range d.Procs {
		p.Buried = bury
	}
}

func (d *LogoDriver) SetBuryAllVars(bury bool) {
	for name, _ := range d.Globals.m {
		if bury {
			d.Globals.Bury(name)
		} else {
			d.Globals.Unbury(name)
		}
	}
}

func (d *LogoDriver) EraseAllProcs() {
	d.Procs = map[string]*LogoProc{}
}

func (d *LogoDriver) GetProcList() []string {
	var out = []string{}
	for k, _ := range d.Procs {
		out = append(out, k)
	}
	return out
}
