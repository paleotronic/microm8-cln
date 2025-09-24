// +build darwin freebsd netbsd openbsd linux

package common

import (
	"os/exec"
	"strings"
	"syscall"

	"paleotronic.com/fmt"
)

func (e *Executor) Stop() {
	if e.r {
		e.r = false
		fmt.Println("killing process")
		syscall.Kill(-e.cmd.Process.Pid, syscall.SIGKILL)
	}
}

func NewExecutor(commandString string) *Executor {

	e := &Executor{
		cmd:         exec.Command("sh", "-c", commandString),
		toCommand:   make(chan byte, 1024),
		fromCommand: make(chan byte, 24576),
		echo:        !strings.Contains(commandString, "telnet"),
		crToLF:      !strings.Contains(commandString, "telnet"),
		addCRtoLF:   true,
	}

	e.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return e

}
