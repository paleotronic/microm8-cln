package asm

import (
	"errors"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/utils"
)

var reImplied = regexp.MustCompile("^[ \t]*([A-Za-z]{3})$")
var reImmediate = regexp.MustCompile("^[ \t]*([A-Za-z]{3})[ \t]+#(.+)$")
var reAbsoluteX = regexp.MustCompile("^[ \t]*([A-Za-z]{3})[ \t]+(.+),[xX]$")
var reAbsoluteY = regexp.MustCompile("^[ \t]*([A-Za-z]{3})[ \t]+(.+),[yY]$")
var reIndirect = regexp.MustCompile("^[ \t]*([A-Za-z]{3})[ \t]+[(](.+)[)]$")
var reIndirectY = regexp.MustCompile("^[ \t]*([A-Za-z]{3})[ \t]+[(](.+)[)],[yY]$")
var reIndirectX = regexp.MustCompile("^[ \t]*([A-Za-z]{3})[ \t]+[(](.+),[xX][)]$")
var reAbsolute = regexp.MustCompile("^[ \t]*([A-Za-z]{3})[ \t]+(.+)$")
var reSymbol = regexp.MustCompile("^([_:]?[A-Za-z][A-Za-z0-9_]*):([ \t]*.*)$")
var reSymbolAlt = regexp.MustCompile("^([]:_]?[A-Za-z][A-Za-z0-9_]*)([ \t]*.*)$")
var reConst = regexp.MustCompile("^[ \t]*([A-Za-z_][A-Za-z0-9_]*)[ \t]*[=][ \t]*(.+)[ \t]*$")
var reConstEQU = regexp.MustCompile("^[ \t]*([A-Za-z][A-Za-z0-9_]*)[ \t]*[.]?(EQU|equ)[ \t]*(.+)[ \t]*$")
var reDotDirective = regexp.MustCompile("^[ \t]*[.]([A-Za-z]+)[ \t]+(.*)$")
var reStarDirective = regexp.MustCompile("^[ \t]*([*])?[ \t]*([=]|ORG|org)(.*)[ \t]*$")
var reBinDirective = regexp.MustCompile("(?i)^[ \t]*(BYT|DFB|HEX|WORD|STR|ASC|DB|DW|DA|DS)[ \t]+(.*)$")
var rePutDirective = regexp.MustCompile("(?i)^[ \t]*[.]?(put|include)[ \t]+(.*)$")
var reHexNumber = regexp.MustCompile("^(0x|[#]?[$])([0-9a-fA-F]+)$")
var reDecNumber = regexp.MustCompile("^[#]?([0-9]+)$")
var reBinaryNumber = regexp.MustCompile("^[#]?[%]([0-1]+)$")
var reString = regexp.MustCompile("^[\"](.*)[\"]$")
var reString2 = regexp.MustCompile("^['](.*)[']$")
var reChar = regexp.MustCompile("^['\"](.)['\"]$")
var reByteList = regexp.MustCompile("^(0x|[$])?([0-9a-fA-F]+)[,](.+)$")
var reExprList = regexp.MustCompile("^(.+)[,](.+)$")
var reMacroStart = regexp.MustCompile("(?i)^[ \t]*([A-Za-z][A-Za-z0-9_]*)?[ \t]*(MAC)[ \t]*$")
var reMacroEnd = regexp.MustCompile("(?i)^[ \t]*([<][<][<])[ \t]*$")
var reLinkDirective = regexp.MustCompile("(?i)^[ \t]*(LST|TYP|XC|DSK|DO |ELSE|FIN)[ \t]*(.*)?$")

const DEBUG = false

type AsmMacro struct {
	lines []string
}

func (m *AsmMacro) Generate(params []string) []string {

	out := make([]string, len(m.lines))
	for i, l := range m.lines {

		for j, par := range params {
			label := fmt.Sprintf("]%d", j+1)
			l = strings.Replace(l, label, strings.Trim(par, " \t"), -1)
		}

		out[i] = l

	}

	return out
}

type Asm6502 struct {
	c           *mos6502.Core6502
	Symbols     map[string]int
	Macros      map[string]*AsmMacro
	PassCount   int
	Unresolved  bool
	OptXC       bool
	OptLST      bool
	OptOUTPUT   string
	OptFileType int
	SearchPath  string
	reMacroCall *regexp.Regexp
}

func NewAsm6502Custom(cpu *mos6502.Core6502) *Asm6502 {
	this := &Asm6502{c: cpu, Symbols: make(map[string]int), Macros: make(map[string]*AsmMacro)}
	return this
}

func NewAsm6502() *Asm6502 {
	this := &Asm6502{c: mos6502.NewCore6502(nil, 0, 0, 0, 0, 0, 0, nil), Symbols: make(map[string]int), Macros: make(map[string]*AsmMacro)}
	return this
}

func NewAsm65C02() *Asm6502 {
	this := &Asm6502{c: mos6502.NewCore65C02(nil, 0, 0, 0, 0, 0, 0, nil), Symbols: make(map[string]int), Macros: make(map[string]*AsmMacro)}
	return this
}

func (a *Asm6502) DecideValueIndex(hpop int, count int, values []string) int {

	// low end
	if hpop == 0 {
		return 0
	}

	// high end
	if hpop+count-1 >= len(values) {
		return len(values) - count
	}

	return hpop
}

func (a *Asm6502) Numeric(in string) (int, error) {

	if strings.HasPrefix(in, "#") {
		in = in[1:]
	}

	if reChar.MatchString(in) {
		m := reChar.FindAllStringSubmatch(in, -1)
		return int(m[0][1][0]), nil
	}

	if v, ok := a.Symbols[strings.ToLower(in)]; ok {
		return v, nil
	}

	if reHexNumber.MatchString(in) {
		m := reHexNumber.FindAllStringSubmatch(in, -1)
		return utils.StrToInt("0x" + m[0][2]), nil
	}

	if reDecNumber.MatchString(in) {
		m := reDecNumber.FindAllStringSubmatch(in, -1)
		return utils.StrToInt(m[0][1]), nil
	}

	if reBinaryNumber.MatchString(in) {
		m := reBinaryNumber.FindAllStringSubmatch(in, -1)
		i, err := strconv.ParseInt(m[0][1], 2, 64)
		return int(i), err
	}

	if a.PassCount == 1 {
		if DEBUG {
			fmt.Printf("... Forward reference %s will be resolved  on pass %d (Sticking a cow 0xBEEF in it for now)\n", in, a.PassCount+1)
		}
		a.Unresolved = true
		return 0xbeef, nil // return dummy value as we may need a forward lookup
	}

	return -1, errors.New("Unresolved Symbol (2nd pass): " + in)
}

// Resolve attempts to resolve a given input
// simple expression parser
func (asm *Asm6502) Resolve(input string) (int, error) {
	validops := "*/+-<>" // this is used as order of precedence
	validsep := " \t"
	values := make([]string, 0)
	ops := make([]string, 0)

	var chunk string
	var inqq bool
	for _, ch := range input {

		if len(values) == 0 && len(ops) == 0 && ch == '*' {
			log.Printf("Getting CPU %d\n", asm.Symbols["*"])
			values = append(values, fmt.Sprintf("$%x", asm.Symbols["*"]))
		} else if rune(ch) == '"' {
			inqq = !inqq
			chunk += string(rune(ch))
		} else if strings.Index(validops, string(rune(ch))) > -1 && !inqq {
			// got an op
			if chunk != "" {
				values = append(values, chunk)
				chunk = ""
			}
			ops = append(ops, string(rune(ch)))
		} else if strings.Index(validsep, string(rune(ch))) > -1 && !inqq {
			// got a sep
			if chunk != "" {
				values = append(values, chunk)
				chunk = ""
			}
		} else {
			chunk += string(rune(ch))
		}
	}

	if chunk != "" {
		values = append(values, chunk)
	}

	// log.Printf("Values: %v\n", values)
	// log.Printf("Ops: %v\n", ops)

	// now parse it
	for len(ops) > 0 {

		hidx := -1
		hival := -1
		for opidx, op := range ops {
			i := strings.Index(validops, op)
			if i > hidx {
				hidx = i
				hival = opidx
			}
		}

		op := ops[hival]

		ops = append(ops[0:hival], ops[hival+1:]...)

		// process highest ops
		needed := 2
		if op == ">" || op == "<" {
			needed = 1
		}

		if len(values) < needed {
			return -1, errors.New("Error parsing expression (too few values for " + op + "): " + strings.Join(values, ", "))
		}

		validx := asm.DecideValueIndex(hival, needed, values)

		var a string
		var b string
		var aval int
		var bval int
		var e error

		if needed == 2 {
			a = values[validx]
			b = values[validx+1]
			values = append(values[0:validx], values[validx+2:]...)
			aval, e = asm.Numeric(a)
			if e != nil {
				return -1, e
			}
			bval, e = asm.Numeric(b)
			if e != nil {
				return -1, e
			}

			fmt.Printf("%d %s %d = ", aval, op, bval)

		} else {
			a = values[validx]
			values = append(values[0:validx], values[validx+1:]...)
			aval, e = asm.Numeric(a)
			if e != nil {
				return -1, e
			}
			fmt.Printf("%s %d = ", op, aval)
		}

		// a and b are values, apply the op
		r := "0"
		switch op {
		case "*":
			r = utils.IntToStr(aval * bval)
		case "/":
			r = utils.IntToStr(aval / bval)
		case "+":
			r = utils.IntToStr(aval + bval)
		case "-":
			r = utils.IntToStr(aval - bval)
		case "<":
			r = utils.IntToStr(aval % 256)
		case ">":
			r = utils.IntToStr(aval / 256)
		}

		fmt.Printf("%s\n", r)

		av := values[0:validx]
		bv := values[validx:]

		fmt.Printf("VALUES B4: %v %v\n", av, bv)

		values = make([]string, 0)

		for _, v := range av {
			values = append(values, v)
		}
		values = append(values, r)
		for _, v := range bv {
			values = append(values, v)
		}

		fmt.Printf("VALUES: %v\n", values)
	}

	if len(values) > 1 {
		return -1, errors.New("Invalid expression: " + input)
	}
	if len(values) == 0 {
		return -1, errors.New("Invalid expression: " + input)
	}

	return asm.Numeric(values[0])
}

func (a *Asm6502) GetAddressMode(in string) (string, string, []mos6502.MODEENUM) {

	if reImplied.MatchString(in) {
		m := reImplied.FindAllStringSubmatch(in, -1)
		return m[0][1], "", []mos6502.MODEENUM{mos6502.MODE_IMPLIED}
	}

	if reImmediate.MatchString(in) {
		m := reImmediate.FindAllStringSubmatch(in, -1)
		return m[0][1], m[0][2], []mos6502.MODEENUM{mos6502.MODE_IMMEDIATE}
	}

	if reIndirectX.MatchString(in) {
		m := reIndirectX.FindAllStringSubmatch(in, -1)
		return m[0][1], m[0][2], []mos6502.MODEENUM{mos6502.MODE_INDIRECT_ZP_X, mos6502.MODE_INDIRECT_X}
	}

	if reIndirectY.MatchString(in) {
		m := reIndirectY.FindAllStringSubmatch(in, -1)
		return m[0][1], m[0][2], []mos6502.MODEENUM{mos6502.MODE_INDIRECT_ZP_Y}
	}

	if reIndirect.MatchString(in) {
		m := reIndirect.FindAllStringSubmatch(in, -1)
		return m[0][1], m[0][2], []mos6502.MODEENUM{mos6502.MODE_INDIRECT_ZP, mos6502.MODE_INDIRECT, mos6502.MODE_INDIRECT_NMOS}
	}

	if reAbsoluteX.MatchString(in) {
		m := reAbsoluteX.FindAllStringSubmatch(in, -1)
		return m[0][1], m[0][2], []mos6502.MODEENUM{mos6502.MODE_ZEROPAGE_X, mos6502.MODE_ABSOLUTE_X}
	}

	if reAbsoluteY.MatchString(in) {
		m := reAbsoluteY.FindAllStringSubmatch(in, -1)
		return m[0][1], m[0][2], []mos6502.MODEENUM{mos6502.MODE_ABSOLUTE_Y}
	}

	if reAbsolute.MatchString(in) {
		m := reAbsolute.FindAllStringSubmatch(in, -1)
		return m[0][1], m[0][2], []mos6502.MODEENUM{mos6502.MODE_ABSOLUTE, mos6502.MODE_RELATIVE, mos6502.MODE_ZEROPAGE}
	}

	return "", "", []mos6502.MODEENUM(nil)

}

// MatchOpcode will find the Opcodes matching the result of GetAddressMode()
func (a *Asm6502) MatchOpcode(m string, p string, modes []mos6502.MODEENUM) ([]*mos6502.Op6502, string) {
	r := make([]*mos6502.Op6502, 0)

	// Fix to support 2 byte BRK
	if strings.ToLower(m) == "brk" && p != "" {
		r = append(r,
			&mos6502.Op6502{
				Cycles:         7,
				AddressingMode: "IMMEDIATE",
				Bytes:          2,
				Opcode:         0x00,
				Fetch:          mos6502.IMMEDIATE,
				FetchMode:      mos6502.MODE_IMMEDIATE,
			},
		)
		return r, p
	}

	for _, opcode := range a.c.Opref {
		if opcode == nil {
			continue
		}
		if strings.ToLower(m) == strings.ToLower(opcode.Description[0:3]) {
			for _, mode := range modes {
				if mode == opcode.FetchMode {
					r = append(r, opcode)
					//fmt.Println(opcode)
				}
			}
		}
	}
	return r, p
}

func (asm *Asm6502) PreProcess(lines []string) ([]string, error) {

	exts := []string{"s", "asm", "a65"}

	out := make([]string, 0)
	currentMacro := ""

	for _, l := range lines {

		if reConstEQU.MatchString(l) {

			if strings.Contains(l, ";") {
				parts := strings.Split(l, ";")
				l = parts[0]
			}

			fmt.Printf("line: %s\n", l)

			m := reConstEQU.FindAllStringSubmatch(l, -1)
			symbol := m[0][1]
			value := m[0][3]

			v, e := asm.Resolve(value)

			asm.Symbols[strings.ToLower(symbol)] = v
			fmt.Printf("*** Define symbol %s -> %v\n", strings.ToUpper(symbol), v)

			if e != nil {
				panic(e)
			}

			continue
		}

		if reLinkDirective.MatchString(l) {

			if strings.Contains(l, ";") {
				parts := strings.Split(l, ";")
				l = parts[0]
			}

			m := reLinkDirective.FindAllStringSubmatch(l, -1)
			command := strings.ToUpper(m[0][1])
			param := strings.ToUpper(m[0][2])
			switch command {
			case "LST":
				asm.OptLST = (param == "ON")
			case "XC":
				asm.OptXC = (param == "ON")
			case "DSK":
				asm.OptOUTPUT = param
			case "TYP":
				asm.OptFileType, _ = asm.Resolve(param)
			case "DO":
				eval, _ := asm.Resolve(param)
				fmt.Printf("Conditional: DO %s = %d\n", param, eval)
			}
			continue
		}

		if asm.reMacroCall != nil && asm.reMacroCall.MatchString(l) {

			m := asm.reMacroCall.FindAllStringSubmatch(l, -1)
			log.Printf("Macromatch: %+v", m)
			leader := m[0][1]
			macroName := m[0][2]
			params := m[0][3]

			if macro, ok := asm.Macros[macroName]; ok {

				parts := strings.Split(params, ";")
				newlines := macro.Generate(parts)

				log.Printf("MACRO EXPAND: %s: %s", leader, macroName)
				for i, v := range newlines {
					log.Printf("MAC#%d: %s", i, v)
				}

				if leader != "" {
					out = append(out, leader)
				}

				out = append(out, newlines...)

				continue

			}

		}

		if strings.Trim(l, " \t\r\n") == "<<<" {

			keys := make([]string, 0)
			for k, _ := range asm.Macros {
				keys = append(keys, k)
			}

			head := "(?i)^([:_A-Za-z0-9]+)?[ \t]*("
			tail := ")[ \t]*(.+)$"
			p := head + strings.Join(keys, "|") + tail
			log.Printf("macro regexp = /%s/", p)
			asm.reMacroCall = regexp.MustCompile(p)

			log.Printf("*** Defined macro: %s", currentMacro)
			currentMacro = ""
			continue

		} else if currentMacro != "" {

			asm.Macros[currentMacro].lines = append(asm.Macros[currentMacro].lines, l)
			fmt.Printf("%s: %s\n", currentMacro, l)
			continue

		} else if reMacroStart.MatchString(l) {

			m := reMacroStart.FindAllStringSubmatch(l, -1)
			symbol := m[0][1]
			currentMacro = symbol
			asm.Macros[currentMacro] = &AsmMacro{
				lines: make([]string, 0),
			}
			log.Printf("*** Start macro: %s", symbol)
			continue

		} else if rePutDirective.MatchString(l) {
			m := rePutDirective.FindAllStringSubmatch(l, -1)
			name := m[0][2]
			found := false

			if strings.HasSuffix(name, ".s") {

				crud, err := utils.ReadTextFile(asm.SearchPath + "/" + name)
				if err == nil {
					fmt.Printf("Loaded %s\n", name)
					out = append(out, crud...)
					found = true
					continue
				}

			} else {

				for _, ext := range exts {
					crud, err := utils.ReadTextFile(asm.SearchPath + "/" + name + "." + ext)
					if err == nil {
						fmt.Printf("Loaded %s\n", name+"."+ext)
						out = append(out, crud...)
						found = true
						continue
					}
				}
			}
			if !found {
				return lines, errors.New("Unable to load file " + name)
			}
		} else {
			out = append(out, l)
		}

	}

	return out, nil

}

func (asm *Asm6502) AssembleMultipass(lines []string, pc int) (map[int][]byte, int, string, error) {

	var err error
	lines, err = asm.PreProcess(lines)
	if err != nil {
		log.Println(err.Error())
		return nil, 0, "", err
	}
	lines, err = asm.PreProcess(lines)
	if err != nil {
		log.Println(err.Error())
		return nil, 0, "", err
	}

	asm.PassCount = 0

	// Pass #1
	codeblocks, lno, line, e := asm.Assemble(
		lines,
		pc,
	)

	if e == nil && asm.Unresolved && asm.PassCount == 1 {
		return asm.Assemble(
			lines,
			pc,
		)
	}

	return codeblocks, lno, line, e

}

func (asm *Asm6502) Assemble(lines []string, pc int) (map[int][]byte, int, string, error) {

	//asm := NewAsm6502()
	asm.PassCount++        // increment pass counter
	asm.Unresolved = false // reset unresolved symbol check

	fmt.Printf("... pass #%d\n", asm.PassCount)

	codeblocks := make(map[int][]byte)

	base := pc

	for lineno, line := range lines {

		asm.Symbols["*"] = pc

		fmt.Printf("CPU.PC = 0x%.4x    %s\n", pc, line)

		line = strings.TrimRight(line, " \t\r\n")

		n := strings.Index(line, ";")
		if n > -1 {
			line = line[0:n]
		}

		if strings.HasPrefix(line, "*") {
			line = ""
		}

		if strings.Trim(line, " \t\r\n") == "" {
			continue
		}

		if reConstEQU.MatchString(line) {
			//fmt.Println("const")
			m := reConstEQU.FindAllStringSubmatch(line, -1)
			symname := m[0][1]
			remainder := strings.Trim(m[0][3], "\t \r\n")

			ival, e := asm.Resolve(remainder)
			if e != nil {
				return codeblocks, lineno, line, e
			}
			asm.Symbols[strings.ToLower(symname)] = ival
			fmt.Printf("--> Defined symbol %s at %.4x\n", symname, pc)
			continue
		} else if reConst.MatchString(line) {
			//fmt.Println("const")
			m := reConst.FindAllStringSubmatch(line, -1)
			symname := m[0][1]
			remainder := strings.Trim(m[0][2], "\t \r\n")

			ival, e := asm.Resolve(remainder)
			if e != nil {
				return codeblocks, lineno, line, e
			}
			asm.Symbols[strings.ToLower(symname)] = ival
			//fmt.Printf("--> Defined symbol %s at %.4x\n", symname, pc)
			continue
		} else if reSymbol.MatchString(line) {
			m := reSymbol.FindAllStringSubmatch(line, -1)
			symname := m[0][1]
			remainder := m[0][2]
			asm.Symbols[strings.ToLower(symname)] = pc
			fmt.Printf("--> Defined symbol %s at %.4x\n", symname, pc)
			line = remainder
		} else if reSymbolAlt.MatchString(line) {
			m := reSymbolAlt.FindAllStringSubmatch(line, -1)
			symname := m[0][1]
			remainder := m[0][2]
			asm.Symbols[strings.ToLower(symname)] = pc
			fmt.Printf("--> Defined symbol %s at %.4x\n", symname, pc)
			line = remainder
		} else if reStarDirective.MatchString(line) {
			m := reStarDirective.FindAllStringSubmatch(line, -1)
			dirname := m[0][1]
			preamble := m[0][3]
			log.Printf("* DIR = %+v", m)

			id := strings.ToUpper(dirname)
			if id == "" && m[0][2] != "=" {
				id = strings.ToUpper(m[0][2])
			}

			fmt.Printf("Dir: %s, Value: %s\n", id, preamble)

			switch {
			case id == "ORG":
				ival, e := asm.Resolve(preamble)
				if e != nil {
					return codeblocks, lineno, line, e
				}
				pc = ival
				base = pc
			case id == "*":
				ival, e := asm.Resolve(preamble)
				if e != nil {
					return codeblocks, lineno, line, e
				}
				pc = ival
				base = pc
			default:
				return codeblocks, lineno, line, errors.New("Invalid *directive: ." + dirname)
			}

			continue

		}

		if line == "" {
			continue
		}

		if reBinDirective.MatchString(line) {
			m := reBinDirective.FindAllStringSubmatch(line, -1)
			dirname := m[0][1]
			preamble := strings.Trim(m[0][2], " ")
			id := strings.ToUpper(dirname)
			fmt.Printf("Dir: %s, Value: %s\n", dirname, preamble)

			switch {
			case id == "BYT" || id == "BYTE" || id == "DB" || id == "DFB":
				preamble = strings.Trim(preamble, " \t\r\n")

				if strings.HasPrefix(preamble, "#") {
					preamble = preamble[1:]
				}

				if strings.HasSuffix(preamble, ",") {
					preamble = preamble[0 : len(preamble)-1]
				}

				var b []byte
				if reByteList.MatchString(preamble) || reExprList.MatchString(preamble) {
					parts := strings.Split(preamble, ",")
					for _, pv := range parts {
						ival, e := asm.Resolve(strings.Trim(pv, " \t\r\n"))
						if e != nil {
							return codeblocks, lineno, line, e
						}
						b = append(b, byte(ival))
					}
				} else if reDecNumber.MatchString(preamble) || reHexNumber.MatchString(preamble) {
					ival, e := asm.Resolve(preamble)
					if e != nil {
						return codeblocks, lineno, line, e
					}
					b = []byte{byte(ival % 256)}
				} else if reString.MatchString(preamble) {
					s := preamble[1 : len(preamble)-1]
					b = []byte(s)
				} else if ival, err := asm.Resolve(preamble); err == nil {
					b = []byte{byte(ival)}
				} else {
					return codeblocks, lineno, line, errors.New("Not a .byte value: " + preamble)
				}

				codeblocks[base] = append(codeblocks[base], b...)
				pc += len(b)
			case id == "HEX":
				var b []byte
				parts := strings.Split(preamble, ",")
				for _, pv := range parts {

					if !strings.HasPrefix(pv, "$") && !strings.HasPrefix(pv, "0x") {
						pv = "$" + pv
					}

					ival, e := asm.Resolve(strings.Trim(pv, " \t\r\n"))
					if e != nil {
						return codeblocks, lineno, line, e
					}
					b = append(b, byte(ival))
				}
				codeblocks[base] = append(codeblocks[base], b...)
				pc += len(b)
			case id == "DS":
				if preamble == "\\" {
					preamble = "1"
				}
				ival, _ := asm.Resolve(preamble)
				b := make([]byte, ival)
				codeblocks[base] = append(codeblocks[base], b...)
				pc += len(b)
			case id == "WORD" || id == "DW" || id == "DA" || id == "DDB":

				parts := strings.Split(preamble, ",")

				b := make([]byte, 0)
				for _, p := range parts {
					ival, e := asm.Resolve(p)
					if e != nil {
						return codeblocks, lineno, line, e
					}
					if id == "DDB" {
						b = append(b, byte(ival/256), byte(ival%256))
					} else {
						b = append(b, byte(ival%256), byte(ival/256))
					}
				}
				codeblocks[base] = append(codeblocks[base], b...)
				pc += len(b)
			case id == "STR" || id == "ASC":
				//fmt.Println("doing string", preamble)
				if !reString.MatchString(preamble) && !reString2.MatchString(preamble) {
					return codeblocks, lineno, line, errors.New(fmt.Sprintf("Not a string: [%s]", preamble))
				}
				s := preamble[1 : len(preamble)-1]
				fmt.Println(s)
				b := []byte(s)
				codeblocks[base] = append(codeblocks[base], b...)
				pc += len(b)
			default:
				return codeblocks, lineno, line, errors.New("Invalid directive: " + dirname)
			}

			continue
		}

		if reDotDirective.MatchString(line) {
			m := reDotDirective.FindAllStringSubmatch(line, -1)
			dirname := m[0][1]
			preamble := m[0][2]
			//fmt.Printf("--> Got directive .%s %s\n", dirname, preamble)

			id := strings.ToUpper(dirname)

			switch {
			case id == "ORG":
				ival, e := asm.Resolve(preamble)
				if e != nil {
					return codeblocks, lineno, line, e
				}
				pc = ival
				base = pc
			case id == "BYT" || id == "BYTE" || id == "DB":
				preamble = strings.Trim(preamble, " \t\r\n")
				var b []byte
				if reDecNumber.MatchString(preamble) || reHexNumber.MatchString(preamble) {
					ival, e := asm.Resolve(preamble)
					if e != nil {
						return codeblocks, lineno, line, e
					}
					b = []byte{byte(ival % 256)}
				} else if reByteList.MatchString(preamble) {
					parts := strings.Split(preamble, ",")
					for _, pv := range parts {
						ival, e := asm.Resolve(strings.Trim(pv, " \t\r\n"))
						if e != nil {
							return codeblocks, lineno, line, e
						}
						b = append(b, byte(ival))
					}
				} else if reString.MatchString(preamble) {
					s := preamble[1 : len(preamble)-1]
					b = []byte(s)
				} else {
					return codeblocks, lineno, line, errors.New("Not a .byte value: " + preamble)
				}

				codeblocks[base] = append(codeblocks[base], b...)
				pc += len(b)
			case id == "PUT":

			case id == "EXPORT":
				//
			case id == "WORD":
				ival, e := asm.Resolve(preamble)
				if e != nil {
					return codeblocks, lineno, line, e
				}
				b := []byte{byte(ival % 256), byte(ival / 256)}
				codeblocks[base] = append(codeblocks[base], b...)
				pc += len(b)
			case id == "STR":
				//fmt.Println("doing string", preamble)
				if !reString.MatchString(preamble) {
					return codeblocks, lineno, line, errors.New("Not a string: " + preamble)
				}
				s := preamble[1 : len(preamble)-1]
				fmt.Println(s)
				b := []byte(s)
				codeblocks[base] = append(codeblocks[base], b...)
				pc += len(b)
			default:
				return codeblocks, lineno, line, errors.New("Invalid directive: ." + dirname)
			}

			continue
		}

		line = strings.TrimLeft(line, " \t\r\n")

		if line == "" {
			continue
		}

		//log.Printf("Checking for valid syntax: {%s}\n", strings.Trim(line, " \t\r\n"))

		r, operand := asm.MatchOpcode(asm.GetAddressMode(strings.Trim(line, " \t\r\n")))

		if len(r) == 0 {
			return codeblocks, lineno, line, errors.New("Parse error: [" + line + "]")
		}

		for _, v := range r {
			pick := false
			if len(r) == 1 {
				pick = true
			}
			ival := 0
			if v.GetBytes() > 1 {
				// has operand
				var e error
				ival, e = asm.Resolve(operand)
				if e != nil {
					return codeblocks, lineno, line, e
				}

				if ival > 256 && v.GetBytes() > 2 {
					pick = true
				} else if ival < 256 && v.GetBytes() == 2 {
					pick = true
				}

			}
			if pick {

				if base == pc {
					codeblocks[base] = make([]byte, 0)
				}

				bytes := make([]byte, v.GetBytes())
				bytes[0] = byte(v.GetOpCode())

				hexstr := fmt.Sprintf("%.2X", bytes[0])

				if v.GetBytes() == 2 {
					if v.FetchMode == mos6502.MODE_RELATIVE {
						ival = ival - pc - v.GetBytes()
					}
					bytes[1] = byte(int8(ival))
					hexstr = fmt.Sprintf("%.2X %.2X", bytes[0], bytes[1])
				} else if v.GetBytes() == 3 {
					bytes[1] = byte(ival % 256)
					bytes[2] = byte(ival / 256)
					hexstr = fmt.Sprintf("%.2X %.2X %.2X", bytes[0], bytes[1], bytes[2])
				}

				if DEBUG {
					fmt.Printf("%.4X: %s\n", pc, hexstr)
				}
				pc += v.GetBytes()

				codeblocks[base] = append(codeblocks[base], bytes...)

				break
			}
		}

	}

	return codeblocks, 0, "", nil

}

func (asm *Asm6502) DumpOutput(codeblocks map[int][]byte) {

	perline := 16

	for base, bytes := range codeblocks {
		for i, v := range bytes {
			if i%perline == 0 {
				fmt.Println()
				fmt.Printf("%.4X:", base+i)
			}
			fmt.Printf(" %.2X", v)
		}
		fmt.Println()
	}

}

func (asm *Asm6502) DumpSyms(out interfaces.Interpretable) {
	for k, v := range asm.Symbols {
		out.PutStr(fmt.Sprintf("%s=$%X, ", strings.ToUpper(k), v))
	}
	out.PutStr("\r\n")
}

func (asm *Asm6502) DumpSymsStderr() {

	maxwidth := 80
	width := 0
	keys := make([]string, 0)
	for k, v := range asm.Symbols {
		keys = append(keys, k)
		l := len(fmt.Sprintf("%s = $%X    ", strings.ToUpper(k), v))
		if l > width {
			width = l
		}
	}

	perline := (maxwidth / width) + 1
	sort.Strings(keys)

	for i, k := range keys {
		v := asm.Symbols[k]
		s := fmt.Sprintf("%s = $%X    ", strings.ToUpper(k), v)
		for len(s) < width {
			s += " "
		}
		os.Stderr.WriteString(s)
		if i%perline == perline-1 {
			os.Stderr.WriteString("\n")
		}
	}
	os.Stderr.WriteString("\n")
}

func (asm *Asm6502) DumpFiles(basename string, codeblocks map[int][]byte) error {

	for base, bytes := range codeblocks {
		filename := fmt.Sprintf("%s-0x%.4x.o", basename, base)
		fmt.Printf("Writing %s...", filename)
		e := utils.WriteBinaryFile(filename, bytes)
		if e != nil {
			fmt.Println("FAIL")
			return e
		}
		fmt.Println("OK!")
	}

	return nil

}

func (asm *Asm6502) DumpFilesNFS(basename string, codeblocks map[int][]byte, out interfaces.Interpretable) error {

	for base, bytes := range codeblocks {
		filename := fmt.Sprintf("%s%.4xh.s", basename, base)
		out.PutStr(fmt.Sprintf("Writing %s...\r\n", filename))
		e := files.WriteBytesViaProvider(files.GetPath(filename), files.GetFilename(filename), bytes)
		if e != nil {
			return e
		}
	}

	return nil

}

func (asm *Asm6502) DumpFilesRAM(codeblocks map[int][]byte, out interfaces.Interpretable) error {

	for base, bytes := range codeblocks {
		filename := fmt.Sprintf("$%.4x", base)
		out.PutStr(fmt.Sprintf("Writing block to RAM at %s...\r\n", filename))
		for i, v := range bytes {
			out.SetMemory(base+i, uint64(v))
		}
	}

	return nil

}
