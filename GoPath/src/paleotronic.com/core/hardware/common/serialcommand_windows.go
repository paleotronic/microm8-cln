// +build windows

package common

import (
	"os/exec"
	"strings"

	"paleotronic.com/fmt"
)

func (e *Executor) Stop() {
	if e.r {
		e.r = false
		fmt.Println("killing process")
		e.cmd.Process.Kill()
	}
}

func NewExecutor(commandString string) *Executor {

	e := &Executor{
		cmd:         exec.Command("powershell", "/c", commandString),
		toCommand:   make(chan byte, 1024),
		fromCommand: make(chan byte, 24576),
		echo:        !strings.Contains(commandString, "telnet"),
		crToLF:      !strings.Contains(commandString, "telnet"),
		addCRtoLF:   true,
	}

	return e

}
