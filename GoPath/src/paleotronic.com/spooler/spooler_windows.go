//go:build windows
// +build windows

// go:build windows

package spooler

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

type winSpooler struct {}

var psPrintCommand = `Start-Process -FilePath "[filename]" -Verb Print -PassThru | %{sleep 10;$_} | kill`

func (ws *winSpooler) SpoolPDF(filename string) error {
	log.Printf("PDFSpooler: attempting to spool output %s", filename)
	// tmp, err := ioutil.TempFile("/tmp", "powershell.*.ps1")
	// if err != nil {
	// 	return err
	// }
	// defer os.Remove(tmp.Name()) // clean-up when done
 //
 //
	// f, err := os.OpenFile(tmp.Name(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	// if err != nil {
	// 	return err
	// }
 //
	// _, err = f.WriteString(strings.ReplaceAll(psPrintCommand, "[filename]", filename))
	// if err != nil {
	// 	return err
	// }
 //
	// err = f.Close()
	// if err != nil {
	// log.Fatal(err)
	// }
 //
	// out, err := exec.Command("Powershell", "-file" , tmp.Name()).Output()
	// if err != nil {
	// 	return err
	// }

	cmd := exec.Command("powershell", "-nologo", "-noprofile")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		fmt.Fprintln(stdin, strings.ReplaceAll(psPrintCommand, "[filename]", filename))
	}()
	time.Sleep(50 * time.Millisecond)
	out, err := cmd.CombinedOutput()

	log.Printf("PDFSpooler: result = %s", string(out))
	return nil
}

func NewSpooler() PrintSpooler {
	return &winSpooler{}
}
