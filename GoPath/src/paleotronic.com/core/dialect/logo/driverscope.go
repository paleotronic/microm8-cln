package logo

import (
	"errors"
	"fmt"

	"paleotronic.com/core/dialect/parcel"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type LogoState int

const (
	lsStopped LogoState = iota
	lsContinue
	lsHalt
	lsCall
	lsReturn
)

type LogoScope struct {
	ProcRef                     *LogoProc // actual proc reference
	ProcLine                    int       // line of code
	ProcStmt                    int       // stmt in line
	StmtIterCount, StmtIterMax  int       // current stmt iteration count
	StmtIterExpr                *types.TokenList
	StmtIterTarget, StmtIterDir *float64
	StmtIterVar                 *string
	ReturnValue                 *types.Token  // set if output has been called
	Vars                        *LogoVarTable // scoped local vars including vars to proc call
	Virtual                     bool
	StmtIterStep                float64
}

func NewLogoScope(proc *LogoProc, itercount int) *LogoScope {
	if itercount == 0 {
		itercount = 1
	}
	return &LogoScope{
		StmtIterCount: itercount,
		StmtIterMax:   itercount,
		StmtIterStep:  1,
		ProcRef:       proc,
		ProcLine:      0,
		ProcStmt:      0,
		ReturnValue:   nil,
		Vars:          NewLogoVarTable(),
	}
}

func (d *LogoScope) IterCount() int {
	return d.StmtIterMax - d.StmtIterCount
}

func (d *LogoDriver) GetVar(name string) (*types.Token, *LogoScope) {
	for i := d.Stack.Size() - 1; i >= 0; i-- {
		if d.Stack.Get(i).Virtual {
			continue // don't search virtual scopes (repeats, etc)
		}
		v := d.Stack.Get(i).Vars.Get(name)
		if v != nil {
			return v, d.Stack.Get(i)
		}
	}
	if d.Globals.Exists(name) {
		return d.Globals.Get(name), nil
	}
	return SharedVars.Get(name), nil
}

func (d *LogoDriver) SetVarShared(name string, value *types.Token) {
	SharedVars.Set(name, value)
}

func (d *LogoDriver) SetVar(name string, value *types.Token, global bool) {

	if global {
		d.Globals.Set(name, value)
	} else {
		_, scope := d.GetVar(name)
		if scope == nil {
			scope = d.S
		}
		scope.Vars.Set(name, value)
	}

}

func (d *LogoDriver) SetVarLocal(name string, value *types.Token) {
	s := d.GetProcScope()
	if s != nil {
		s.Vars.Set(name, value)
	} else {
		d.Globals.Set(name, value)
	}
}

func (d *LogoDriver) SetScopeContext(proc *LogoProc, line, stmt int) {
	d.S.ProcRef = proc
	d.S.ProcLine = line
	d.S.ProcStmt = stmt
}

func (d *LogoDriver) GetScopeContext() (*LogoProc, int, int) {
	return d.S.ProcRef, d.S.ProcLine, d.S.ProcStmt
}

// CreateProcScope creates a new stack frame with a proc and args - args are seeded to
// local context vars
func (d *LogoDriver) CreateProcScope(proc *LogoProc, args *types.TokenList) error {

	currentProcScope := d.GetProcScope()
	var replaceScope bool
	if proc == currentProcScope.ProcRef && !proc.ContainsOutput { // APRIL -- only clean recursion if not a function...
		// we are recursing here... so let's artificially collapse any virtual scopes down to proc level
		d.Printf("Looks like we are recursing :) down the rabbit hole we go...")
		for d.S != currentProcScope {
			d.DestroyScope()
		}
		replaceScope = true
	}

	c := NewLogoScope(proc, 1)
	if len(proc.Arguments) > 0 {
		if args.Size() < len(proc.Arguments) {
			return fmt.Errorf("not enough inputs to %s, expect %d", proc.Name, len(proc.Arguments))
		}
		if args.Size() > len(proc.Arguments) {
			return fmt.Errorf("i don't know what to do with %s", d.DumpObjectStruct(args.Get(len(proc.Arguments)), false, ""))
		}
		for i, argname := range proc.Arguments {
			c.Vars.Set(argname, args.Get(i).Copy())
		}
	}

	if replaceScope {
		d.Stack.Pop() // pop off the old scope for the same proc
		if e := d.Stack.Push(c); e != nil {
			return e
		}
		d.Printf("looping scope %s", c.ProcRef.Name)
	} else {
		d.Printf("entering scope %s", c.ProcRef.Name)
		if e := d.Stack.Push(c); e != nil {
			return e
		}
	}
	d.S = c
	d.State = lsCall
	return nil
}

func (d *LogoDriver) CreateCommandScope(l *parcel.Lexer, lines []string) error {
	p, err := d.CreateProc(l, "__immediate."+utils.IntToStr(d.Stack.Size()), []string{}, lines)
	if err != nil {
		return err
	}
	d.Printf("entering scope %s", p.Name)
	s := NewLogoScope(p, 1)
	s.Virtual = true
	if d.S != nil {
		s.StmtIterStep = d.S.StmtIterStep
	}
	if e := d.Stack.Push(s); e != nil {
		return e
	}
	d.S = s
	d.State = lsCall
	return nil
}

func (d *LogoDriver) CreateWhileBlockScope(expr *types.TokenList, list *types.TokenList) error {
	p, err := d.CreateBlock(fmt.Sprintf("__while.%d", d.Stack.Size()), list)
	if err != nil {
		return err
	}
	d.Printf("entering scope %s", p.Name)
	s := NewLogoScope(p, 1)
	s.StmtIterExpr = expr
	s.Virtual = true
	if d.S != nil {
		s.StmtIterStep = d.S.StmtIterStep
	}
	if e := d.Stack.Push(s); e != nil {
		return e
	}
	d.S = s
	d.State = lsCall
	return nil
}

func (d *LogoDriver) CreateRepeatBlockScope(count int, list *types.TokenList) error {
	p, err := d.CreateBlock(fmt.Sprintf("__repeat.%d", d.Stack.Size()), list)
	if err != nil {
		return err
	}
	d.Printf("entering scope %s", p.Name)
	s := NewLogoScope(p, count)
	s.Virtual = true
	if d.S != nil {
		s.StmtIterStep = d.S.StmtIterStep
	}
	if e := d.Stack.Push(s); e != nil {
		return e
	}
	d.S = s
	d.State = lsCall
	return nil
}

func (d *LogoDriver) CreateForBlockScope(varname string, target float64, dir float64, list *types.TokenList) error {
	p, err := d.CreateBlock(fmt.Sprintf("__for.%d", d.Stack.Size()), list)
	if err != nil {
		return err
	}
	d.Printf("entering scope %s", p.Name)
	s := NewLogoScope(p, 1)
	s.StmtIterTarget = &target
	s.StmtIterDir = &dir
	if d.S != nil {
		s.StmtIterStep = d.S.StmtIterStep
	}
	s.StmtIterVar = &varname
	s.Virtual = true
	if e := d.Stack.Push(s); e != nil {
		return e
	}
	d.S = s
	d.State = lsCall
	return nil
}

func (d *LogoDriver) CreateCoroutineScope(list *types.TokenList, id string) error {
	p, err := d.CreateBlock(fmt.Sprintf("__coroutine.%s.%d", id, d.Stack.Size()), list)
	if err != nil {
		return err
	}
	d.Printf("entering coroutine scope %s", p.Name)
	s := NewLogoScope(p, 1)
	if d.S != nil {
		s.StmtIterStep = d.S.StmtIterStep
	}
	s.Virtual = true
	if e := d.Stack.Push(s); e != nil {
		return e
	}
	d.S = s
	d.State = lsCall
	return nil
}

func (d *LogoDriver) DestroyScope() (*LogoScope, error) {
	s := d.Stack.Pop()
	d.S = d.Stack.Top()
	if s == nil {
		return nil, errors.New("EOF")
	}
	d.Printf("exiting scope %s", s.ProcRef.Name)
	return s, nil
}

func (d *LogoScope) GetNext() (int, int) {

	if d == nil {
		return -1, -1
	}

	line := d.ProcLine
	stmt := d.ProcStmt

	if line == -1 {
		return -1, -1
	}

	if line >= len(d.ProcRef.Lines) {
		stmt = 0
		// out of lines in current scope
		d.StmtIterCount--
		if d.StmtIterCount > 0 {
			stmt = 0
			line = 0
		} else {
			stmt = -1
			line = -1
		}
	} else {
		stmt++
		if stmt >= len(d.ProcRef.Lines[line].Statements) {
			stmt = 0
			line++
			if line >= len(d.ProcRef.Lines) {
				// out of lines in current scope
				d.StmtIterCount--
				if d.StmtIterCount > 0 {
					stmt = 0
					line = 0
				} else {
					stmt = -1
					line = -1
				}
			}
		}
	}
	return line, stmt
}

func (d *LogoDriver) GetProcScope() *LogoScope {
	for i := d.Stack.Size() - 1; i > 0; i-- {
		if !d.Stack.Get(i).Virtual {
			return d.Stack.Get(i)
		}
	}
	return d.Stack.Get(0)
}

func (d *LogoDriver) ThrowTopLevel() {
	for d.Stack.Size() > 0 {
		d.DestroyScope()
	}
}

func (d *LogoDriver) CurrentProc() (string, int) {
	if d.S != nil && d.S.ProcRef != nil {
		return d.S.ProcRef.Name, d.S.ProcLine
	}
	return "", 0
}
