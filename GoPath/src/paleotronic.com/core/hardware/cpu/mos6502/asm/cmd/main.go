package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"paleotronic.com/core/hardware/cpu/mos6502/asm"
	"paleotronic.com/utils"
)

var target = flag.String("target", "6502", "Target CPU to assemble for")
var input = flag.String("i", "", "Input file")
var output = flag.String("o", "out", "Output file basename (.o added)")
var symbols = flag.Bool("s", true, "Print out symbols at end.")

func fatalf(s string, i ...interface{}) {
	fmt.Printf(s, i...)
	fmt.Println()
	os.Exit(1)
}

func fatal(s string) {
	fmt.Println(s)
	os.Exit(1)
}

func main() {

	var a *asm.Asm6502

	flag.Parse()

	switch *target {
	case "6502":
		a = asm.NewAsm6502()
	case "65C02":
		a = asm.NewAsm65C02()
	default:
		fatalf("Unknown cpu type: '%s'\n", *target)
	}

	if *input == "" {
		fatal("I need a filename as input (-i filename)")
	}

	searchPath := filepath.Dir(*input)
	if searchPath != *input {
		a.SearchPath = searchPath
	} else {
		a.SearchPath = "."
	}

	lines, e := utils.ReadTextFile(*input)
	if e != nil {
		fatalf("Error reading input file: %s - %s\n", *input, e.Error())
	}

	if *output == "out" {
		*output = strings.Replace(*input, ".asm", "", -1)
	}

	codeblocks, lno, line, e := a.AssembleMultipass(
		lines,
		0x4000,
	)

	if e != nil {
		fmt.Printf("Assembly failed at line %d:\n", lno)
		fmt.Println("  " + line)
		fmt.Println(e.Error())
	} else {
		a.DumpFiles(*output, codeblocks)
	}

	if *symbols {
		a.DumpSymsStderr()
	}

}
