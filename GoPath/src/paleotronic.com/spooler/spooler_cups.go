// +build !windows
// go:build !windows

package spooler

import (
	"log"
	"os/exec"
)

type cupsSpooler struct {}

func (cs *cupsSpooler) SpoolPDF(filename string) error {
	log.Printf("PDFSpooler: attempting to spool output %s", filename)
	cmd := exec.Command("lp", filename)
	return cmd.Run()
}

func NewSpooler() PrintSpooler {
	return &cupsSpooler{}
}
