// main project main.go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type Field struct {
	Name     string
	Type     string
	JavaType string
	Scope    string
}

type Function struct {
	Name        string
	ReturnType  []string
	Params      []Field
	ClassName   string
	Content     []string
	Scope       string
	Constructor bool
}

type PatternReplacement struct {
	Pattern     string         // String version of the pattern
	regexp      *regexp.Regexp // Compiled version of the pattern
	Replacement string         // Replacement with $1..$n placeholders
	Needs       []string       // Any go libraries needed by the process
}

type ReplacementYAMLStruct struct {
	Replacements []PatternReplacement
}

var srcFile string
var packageName string
var inheritsFrom string
var patternFile string
var className string
var funcName string
var fields map[string]Field
var functions map[string]Function
var braceCount int
var vartypes map[string]string
var replaces map[*regexp.Regexp]*PatternReplacement
var currentFunc Function
var lastLine string
var constructorExtra []string
var importsExtra []string

func isIn(s string, list []string) bool {
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func splitLine(in string, splitdiscard string, splitkeep string) []string {
	chunk := ""
	var s []string
	var rs string
	var inQ bool = false
	var inQQ bool = false

	for _, r := range in {

		rs = string(r)

		if (strings.IndexAny(rs, splitdiscard) >= 0) && !(inQ || inQQ) {
			if chunk != "" {
				s = append(s, chunk)
				chunk = ""
			}
		} else if (strings.IndexAny(rs, splitkeep) >= 0) && !(inQ || inQQ) {
			if chunk != "" {
				s = append(s, chunk)
				chunk = ""
			}
			s = append(s, rs)
		} else {
			if (r == '\'') && (!inQQ) {
				inQ = !inQ
			} else if (r == '"') && (!inQ) {
				inQQ = !inQQ
			}
			chunk = chunk + rs
		}
	}

	if chunk != "" {
		s = append(s, chunk)
		chunk = ""
	}

	//////fmt.Printf("===> %q\n", s)

	return s
}

func (this Function) ParamSpec() string {
	var out []string

	out = []string(nil)

	for _, f := range this.Params {
		out = append(out, f.Name+" "+f.Type)
	}

	return strings.Join(out, ", ")
}

func DumpStruct() string {
	var out []string

	out = []string(nil)

	out = append(out, "type "+className+" struct {")
	if inheritsFrom != "" {
		out = append(out, "    "+inheritsFrom)
	}
	for _, v := range fields {
		out = append(out, "    "+v.Name+" "+v.Type)
	}
	out = append(out, "}")

	return strings.Join(out, "\n")
}

func (this Function) Dump() string {
	var out []string

	out = []string(nil)

	rts := ""
	if len(this.ReturnType) == 1 {
		rts = this.ReturnType[0]
	} else if len(this.ReturnType) > 1 {
		rts = "(" + strings.Join(this.ReturnType, ", ") + ")"
	}

	if !this.Constructor {
		out = append(out, "func (this *"+this.ClassName+") "+this.Name+"("+this.ParamSpec()+") "+rts+" {")
	} else {
		out = append(out, "func "+this.Name+"("+this.ParamSpec()+") "+rts+" {")
		out = append(out, "    this := "+this.ClassName+"{}")
		for _, extra := range constructorExtra {
			out = append(out, "    "+extra)
		}
	}
	out = append(out, this.Content...)
	if this.Constructor {
		out = append(out, "    return this")
	}
	out = append(out, "}")

	return strings.Join(out, "\n")
}

func getGoType(s string) string {

	r, ok := vartypes[s]
	if ok {
		return r
	}

	return s
}

func processDeclaration(line string, parts []string) {

	if strings.Contains(line, "= make([]") {
		ppp := strings.Split(line, "=")
		line = ppp[0]

		qqq := strings.Split(strings.Trim(ppp[0], " "), " ")

		constructorExtra = append(constructorExtra, "this."+qqq[len(qqq)-1]+" = "+ppp[1])
	}

	re_Proc := regexp.MustCompile("(public[ ]+|private[ ]+|protected[ ]+)?(static[ ]+)?([a-zA-Z0-9<>,]+)?([[])?([]])?[ ]+([a-zA-Z0-9_]+)[ ]*[(](.*)[)]")
	re_Var := regexp.MustCompile("(public[ ]+|private[ ]+|protected[ ]+)?(static[ ]+)?([a-zA-Z0-9<>,]+)([[])?([]])?[ ]+([a-zA-Z0-9_]+)[ ]*([=])?[ ]*(.+)?[;]?")

	procMatch := re_Proc.FindAllStringSubmatch(line, -1)
	varMatch := re_Var.FindAllStringSubmatch(line, -1)

	if len(procMatch) > 0 {
		//		log.Printf("Proc : %q\n", procMatch)

		rt_list := []string(nil)

		rt := procMatch[0][4] + procMatch[0][5] + procMatch[0][3]
		if rt == "void" {
			rt = ""
		}
		rt = getGoType(rt)

		if rt != "" {
			rt_list = append(rt_list, rt)
		}

		name := procMatch[0][6]

		isConst := false

		if (name == "") || (name == className) {
			name = "New" + className
			rt_list = []string{"*" + className}
			isConst = true
		}

		currentFunc = Function{Scope: procMatch[0][1], Constructor: isConst, Name: name, ReturnType: rt_list, ClassName: className, Content: []string(nil), Params: []Field(nil)}

		if currentFunc.Scope == "private" {
			currentFunc.Name = currentFunc.Name
		} else {
			currentFunc.Name = strings.Title(currentFunc.Name)
		}

		//process args
		argstr := strings.Trim(procMatch[0][7], " ")
		if argstr != "" {
			argparts := strings.Split(argstr, ",")
			for _, a := range argparts {
				a = strings.Trim(a, " ")
				a = strings.Replace(a, "  ", " ", -1)
				a = strings.Replace(a, "\t", " ", -1)
				fmt.Println(a)
				p := strings.Split(a, " ")
				ff := Field{Name: p[1], JavaType: p[0], Type: getGoType(p[0])}
				currentFunc.Params = append(currentFunc.Params, ff)
			}
		}

	} else if len(varMatch) > 0 {
		typ := varMatch[0][4] + varMatch[0][5] + varMatch[0][3]
		typ = getGoType(typ)
		name := varMatch[0][6]
		ff := Field{Name: name, Type: typ}
		fields[name] = ff
	}

}

func startsWith(s string, substr string) bool {
	if len(s) < len(substr) {
		return false
	}

	cmp := s[0:len(substr)]

	return (cmp == substr)
}

func endsWith(s string, substr string) bool {
	if len(s) < len(substr) {
		return false
	}

	cmp := s[len(s)-len(substr) : len(s)]

	return (cmp == substr)
}

func getWS(s string) string {

	var out string
	for _, r := range s {
		if r == ' ' {
			out = out + " "
		} else if r == '\t' {
			out = out + "\t"
		} else {
			return out
		}
	}
	return out
}

func parseLine(index int, line string) {

	fmt.Printf("%d) %s\n", index, line)

	// pre-replace
	//log.Printf("num replaces: %d", len(replaces))
	for reg, p := range replaces {
		if reg.MatchString(line) {
			line = reg.ReplaceAllString(line, p.Replacement)
			// If this replacement needs golang imports, add them
			for _, i := range p.Needs {
				if !isIn(i, importsExtra) {
					importsExtra = append(importsExtra, i)
				}
			}
		}
	}

	re := regexp.MustCompile("[.]([a-z])")

	line = re.ReplaceAllStringFunc(line, strings.ToUpper)

	line = strings.Replace(line, " new ", " New", -1)

	if strings.Contains(line, "throw ") {
		// subst throw to "return"
		// make sure func returns an error value
		contains := false
		for _, rt := range currentFunc.ReturnType {
			if rt == "error" {
				contains = true
			}
		}
		if !contains {
			currentFunc.ReturnType = append(currentFunc.ReturnType, "error")
		}
		line = strings.Replace(line, "throw ", "return ", -1)
	}

	if strings.Contains(line, "while ") {
		line = strings.Replace(line, "while ", "for ", -1)
	}

	oline := line
	line = strings.Trim(line, "\t ")
	if (len(line) > 2) && (line[0:2] == "//") {
		currentFunc.Content = append(currentFunc.Content, oline)
		return
	}

	if (line != "") && (line[len(line)-1] == ';') {
		line = line[0 : len(line)-1]
	}

	if (oline != "") && (oline[len(oline)-1] == ';') {
		oline = oline[0 : len(oline)-1]
	}

	parts := splitLine(line, " \t", "()+/-*[]{},=")

	if (len(parts) == 2) && (parts[0] != "try") && (parts[0] != "return") && (parts[0] != "/") {
		line = "var " + parts[1] + " " + getGoType(parts[0])
		oline = getWS(oline) + line
	}

	pullUp := false

	pullUp = (line == "{") || ((line == "else") && (endsWith(currentFunc.Content[len(currentFunc.Content)-1], "}"))) || ((len(currentFunc.Content) > 0) && (startsWith(line, "if ")) && (endsWith(currentFunc.Content[len(currentFunc.Content)-1], "else")))

	if (pullUp) && (len(currentFunc.Content) > 0) && (startsWith(strings.Trim(currentFunc.Content[len(currentFunc.Content)-1], " \t"), "//")) {
		pullUp = false
	}

	enclose := false

	if startsWith(lastLine, "if ") && !endsWith(lastLine, "{") && !endsWith(lastLine, "&&") && !endsWith(lastLine, "||") && (line != "{") {
		enclose = true
	}

	if startsWith(lastLine, "while ") && !endsWith(lastLine, "{") && !endsWith(lastLine, "&&") && !endsWith(lastLine, "||") && (line != "{") {
		enclose = true
	}

	if startsWith(lastLine, "else") && !endsWith(lastLine, "{") && (line != "{") && (!startsWith(line, "if ")) {
		enclose = true
	}

	// replace while with for

	if (len(parts) == 1) && (parts[0] == "}") && (braceCount == 2) {
		functions[currentFunc.Name] = currentFunc

	} else if braceCount >= 2 {

		if enclose {
			currentFunc.Content[len(currentFunc.Content)-1] += " {"
		}

		indent := ""
		if len(currentFunc.Content) > 0 {
			indent = getWS(currentFunc.Content[len(currentFunc.Content)-1])
		}

		if pullUp {
			currentFunc.Content[len(currentFunc.Content)-1] += " " + line
		} else {
			currentFunc.Content = append(currentFunc.Content, oline)
		}

		if enclose {
			currentFunc.Content = append(currentFunc.Content, indent+"}")
		}
	}

	// class decl
	if (len(parts) >= 3) && (parts[0] == "public") && (parts[1] == "class") {
		className = parts[2]
	}

	if (len(parts) > 1) && (braceCount == 1) {
		processDeclaration(line, parts)
	}

	// adjust brace count
	for _, r := range parts {
		if r == "{" {
			braceCount++
		}
		if r == "}" {
			braceCount--
		}
	}

	lastLine = line
}

func parseFile(data []string) []string {
	var out []string
	for i, line := range data {
		parseLine(i, line)

	}
	return out
}

func dumpFile(filename string) {
	fmt.Printf("package %s\n", packageName)
	fmt.Println()
	if len(importsExtra) > 0 {
		fmt.Println("import (")
		fmt.Println("\t\"" + strings.Join(importsExtra, "\"\n\t\"") + "\"")
		fmt.Println(")")
		fmt.Println()
	}
	fmt.Println(DumpStruct())
	fmt.Println()

	for _, ff := range functions {
		fmt.Println(ff.Dump())
		fmt.Println()
	}
}

func initStructures() {
	// init stuff
	fields = make(map[string]Field)
	functions = make(map[string]Function)
	vartypes = make(map[string]string)
	replaces = make(map[*regexp.Regexp]*PatternReplacement)
	constructorExtra = make([]string, 0)
}

func initTypeMappings() {
	vartypes["String"] = "string"
	vartypes["String[]"] = "[]string"
	vartypes["int"] = "int"
	vartypes["long"] = "int64"
	vartypes["int[]"] = "[]int"
	vartypes["char"] = "byte"
	vartypes["char[]"] = "[]byte"
	vartypes["float"] = "float32"
	vartypes["double"] = "float64"
	vartypes["float[]"] = "[]float32"
	vartypes["double[]"] = "[]float64"
	vartypes["[]double"] = "[]float64"
	vartypes["[]String"] = "[]string"
	vartypes["byte"] = "int8"
	vartypes["[]byte"] = "[]int8"
	vartypes["boolean"] = "bool"
}

func loadPatterns() {
	// A YAML file containing the following structure
	// Replacement:

	data, err := ioutil.ReadFile(patternFile)
	if err != nil {
		panic(err)
	}

	var rep ReplacementYAMLStruct

	err = yaml.Unmarshal(data, &rep)
	if err != nil {
		panic(err)
	}

	for _, p := range rep.Replacements {
		zz := p
		zz.regexp = regexp.MustCompile("(?i)" + zz.Pattern)
		replaces[zz.regexp] = &zz
	}

}

func main() {
	flag.StringVar(&srcFile, "src", "", "Specify the source file to process")
	flag.StringVar(&packageName, "pkg", "main", "Specify the target package name")
	flag.StringVar(&inheritsFrom, "base", "", "Specify struct basis")
	flag.StringVar(&patternFile, "patterns", "", "File containing regex patterns, and replacements")
	flag.Parse()

	if srcFile == "" {
		log.Fatalf("Need filename (got [%v])\n", srcFile)
	}

	data, err := readLines(srcFile)
	if err != nil {
		log.Fatalln(err.Error())
	}

	initStructures()
	initTypeMappings()
	if patternFile != "" {
		loadPatterns()
	}

	importsExtra = []string{}

	_ = parseFile(data)
	dumpFile("")

}
