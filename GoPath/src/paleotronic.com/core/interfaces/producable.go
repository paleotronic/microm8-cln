package interfaces

import (
	"time"

	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
)

type Presentable interface {
	Apply(context string, ent Interpretable) error
}

type Producable interface {
	BootStrap(path string) error
	PauseMicroControls()
	ResumeMicroControls()
	SetPState(index int, p Presentable, source string)
	SetPresentationSource(index int, source string)
	CallEndRemotes()
	//HasUnblockedInterpreters() bool
	Stop()
	StopAudio()
	StopMicroControls()
	AmIActive(ent Interpretable) bool
	HasRunningInterpreters() bool
	//Execute() (int, time.Duration)
	SetContext(v int) error
	Parse(s string)
	Broadcast(msg types.InfernalMessage)
	//CreateInterpreter(slot int, name string, dia Dialecter, spec string, uuid uint64) (Interpretable, error)
	//Reboot()
	Run()
	GetMinWait() (int, time.Duration)
	SetNeedsPrompt(v bool)
	GetContext() int
	GetInterpreterList() [memory.OCTALYZER_NUM_INTERPRETERS]Interpretable
	Post(index int)
	Activate(slotid int)
	GetNumInterpreters() int
	GetMemoryCallback(index int) func(index int)
	SetMasterLayerPos(index int, x, y float64)
	GetMasterLayerPos(index int) (float64, float64)
	DropInterpreter(slotid int)
	RestartInterpreter(slotid int)
	//GetNextSlot() int
	Select(slot int)
	SetInputContext(slot int) error
	GetInterpreter(slot int) Interpretable
	DropVM(slot int)
}
