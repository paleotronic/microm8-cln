package cpu

type FEResponse int

const (
	FE_OK FEResponse = iota
	FE_HALTED
	FE_BREAKPOINT
	FE_BREAKPOINT_MEM
	FE_BREAKINTERRUPT
	FE_ILLEGALOPCODE
	FE_MEMORYPROTECT
	FE_CTRLBREAK
	FE_SLEEP
)

func (f FEResponse) String() string {
	switch f {
	case FE_HALTED:
		return "Halt"
	case FE_BREAKPOINT:
		return "Breakpoint"
	case FE_BREAKPOINT_MEM:
		return "Memory Breakpoint"
	case FE_BREAKINTERRUPT:
		return "Break Interrupt"
	case FE_ILLEGALOPCODE:
		return "Illegal Opcode"
	case FE_MEMORYPROTECT:
		return "Memory Violation"
	case FE_CTRLBREAK:
		return "User Interrupt"
	case FE_SLEEP:
		return "Sleep"
	}
	return "Unknown Event"
}

type CPU interface {
	FetchExecute() FEResponse
	ExecuteSliced() FEResponse
	Reset()
	ResetSliced()
	IsHalted() bool
	SetHalted(halted bool)
}
