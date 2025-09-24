package logo

import (
	"strings"

	"paleotronic.com/core/types"
	"paleotronic.com/log"
)

type LogoRunState struct {
	Stack                *LogoStack
	S                    *LogoScope
	ProcLine             int
	ProcStmt             int
	LastTargetStackLevel int
	LastReturn           *types.Token
}

func (d *LogoDriver) PauseExecution() {

	p := d.GetProcScope()
	if p == nil || strings.HasPrefix(p.ProcRef.Name, "_") {
		log.Printf("Not pausing from an immediate stack state.")
		return // can't pause in non-proc scope (eg. Immediate)
	}
	log.Printf("Pausing execution...")
	d.savedRunState = &LogoRunState{
		LastTargetStackLevel: 0,
		S:                    d.S,
		Stack:                d.Stack,
		LastReturn:           d.LastReturn,
	}
	d.Stack = NewLogoStack()
	d.S = nil

	d.ent.PutStr("PAUSED IN " + p.ProcRef.Name + "\r\n")
}

func (d *LogoDriver) ResumeExecution() {
	if d.savedRunState != nil {
		d.S = d.savedRunState.S
		d.Stack = d.savedRunState.Stack
		d.LastReturn = d.savedRunState.LastReturn
		d.savedRunState = nil
		d.hasResumed = true // flag so we know what is going on
		log.Printf("Resuming from pause state...")
	}
}
