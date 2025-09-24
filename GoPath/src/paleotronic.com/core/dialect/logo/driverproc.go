package logo

import (
	"strings"

	"paleotronic.com/core/dialect/parcel"
	"paleotronic.com/core/types"

	"paleotronic.com/log"
)

type LogoLine struct {
	Statements []*types.TokenList
}

type LogoProc struct {
	l              *parcel.Lexer
	Name           string
	Arguments      []string
	RawDef         []string
	Lines          []*LogoLine
	ContainsOutput bool
	Buried         bool // is the proc buried?
}

func (d *LogoDriver) CreateProc(l *parcel.Lexer, name string, args []string, code []string) (*LogoProc, error) {
	var proc = &LogoProc{
		Name:      name,
		Arguments: args,
		RawDef:    code,
		Lines:     []*LogoLine{},
	}

	for _, line := range code {
		d.Printf("parsing line: %s", line)
		stmts, err := d.ParseMulti(l, line)
		d.Printf("parse result: %v, %v", stmts, err)
		if err != nil {
			log.Println("[create-proc] parsing failed")
			return nil, err
		}
		for _, s := range stmts {
			if s.Get(0).Type == types.KEYWORD && (strings.ToLower(s.Get(0).Content) == "output" || strings.ToLower(s.Get(0).Content) == "op") {
				proc.ContainsOutput = true
			}
		}
		proc.Lines = append(proc.Lines,
			&LogoLine{
				Statements: stmts,
			},
		)
	}

	log.Printf("[create-proc] created proc: %s [%+v] %+v", name, args, code)

	return proc, nil
}

func (d *LogoDriver) StoreProc(l *parcel.Lexer, name string, args []string, code []string) (*LogoProc, error) {
	proc, err := d.CreateProc(l, name, args, code)
	if err != nil {
		return nil, err
	}

	d.Procs[strings.ToLower(name)] = proc

	return proc, nil
}

func (d *LogoDriver) CreateBlock(name string, code *types.TokenList) (*LogoProc, error) {
	var proc = &LogoProc{
		Name:      name,
		Arguments: []string{},
		RawDef:    []string{},
		Lines:     []*LogoLine{},
	}
	stmts := d.SplitMulti(code)
	for _, s := range stmts {
		if s.Get(0).Type == types.KEYWORD && (strings.ToLower(s.Get(0).Content) == "output" || strings.ToLower(s.Get(0).Content) == "op") {
			proc.ContainsOutput = true
		}
	}
	proc.Lines = append(proc.Lines,
		&LogoLine{
			Statements: stmts,
		},
	)
	return proc, nil
}

func (d *LogoDriver) HasProc(name string) bool {
	_, ok := d.GetProc(name)
	return ok
}

func (d *LogoDriver) HasFunc(name string) bool {
	p, ok := d.GetProc(name)
	return ok && p.ContainsOutput
}

func (d *LogoDriver) GetProc(name string) (*LogoProc, bool) {
	p, ok := d.Procs[strings.ToLower(name)]
	return p, ok
}

func (d *LogoProc) GetCode() []string {
	head := "TO " + d.Name
	if len(d.Arguments) > 0 {
		head += " " + strings.Join(d.Arguments, " ")
	}
	var body = []string{head}
	body = append(body, d.RawDef...)
	body = append(body, "END")
	return body
}

func (d *LogoDriver) ReparseProc(l *parcel.Lexer, name string) error {
	pdata, ok := d.GetProc(name)
	if !ok {
		return nil
	}
	_, err := d.StoreProc(l, name, pdata.Arguments, pdata.RawDef)
	return err
}

func (d *LogoDriver) ReparseProcs(l *parcel.Lexer, names []string) error {
	for _, name := range names {
		err := d.ReparseProc(l, name)
		if err != nil {
			return err
		}
	}
	return nil
}

// FindProcsWithUnresolvedSymbols locates procs that don't have fully resolved
// symbols..
func (d *LogoDriver) FindProcsWithUnresolvedSymbols() []string {
	var out = []string{}
	for name, p := range d.Procs {
		var count = 0
		for _, l := range p.Lines {
			for _, s := range l.Statements {
				for _, t := range s.Content {
					count = d.CountUnresolvedInObject(t, count)
					if count > 0 {
						break
					}
				}
				if count > 0 {
					break
				}
			}
			if count > 0 {
				break
			}
		}
		if count > 0 {
			out = append(out, name)
		}
	}
	return out
}
