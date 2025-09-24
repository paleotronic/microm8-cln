package logo

import "time"

type LogoDriverCommand string

const (
	CommandSuspend    LogoDriverCommand = "suspend"
	CommandResume     LogoDriverCommand = "resume"
	CommandPause      LogoDriverCommand = "pause"
	CommandCont       LogoDriverCommand = "cont"
	CommandStop       LogoDriverCommand = "stop"
	CommandDefinesOff LogoDriverCommand = "defines-off"
	CommandDefinesOn  LogoDriverCommand = "defines-on"
)

func (d *LogoDriver) SendCommand(command LogoDriverCommand) {
	d.Commands <- command
}

func (d *LogoDriver) ClearCommands() {
	if d.Commands != nil {
		close(d.Commands)
	}
	d.Commands = make(chan LogoDriverCommand, 3)
}

func (d *LogoDriver) CheckCommands() {
	if len(d.Commands) > 0 {
		cmd := <-d.Commands
		switch cmd {
		case CommandPause:
			d.PauseExecution()
		case CommandCont:
			d.ResumeExecution()
		case CommandSuspend:
			d.Paused = true
		case CommandResume:
			d.Paused = false
		case CommandStop:
			d.ThrowTopLevel()
		case CommandDefinesOff:
			d.DisableDefineMsgs = true
		case CommandDefinesOn:
			d.DisableDefineMsgs = false
		}
	}
}

func (d *LogoDriver) CheckPaused() {
	for d.Paused && !d.HasReboot() && !d.HasCBreak() {
		time.Sleep(5 * time.Millisecond)
		d.CheckCommands()
	}
}
