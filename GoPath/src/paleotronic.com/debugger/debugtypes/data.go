package debugtypes

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

/*
	Implements a nicer interface to debugging the system.
*/

type MemSearchResult struct {
	Search    []int
	FoundAddr int
	Aux       bool
}

type CPUInstructionDecode struct {
	Address     int
	Bytes       []int
	Instruction string
	Cycles      int
	Historic    bool
}

type CPUState struct {
	PC      int   // Program counter
	A, X, Y int   // Registers
	P       int   // Processor status
	SP      int   // Stack Pointer
	CC      int64 // Cumulative Cycles
	Speed   float64
	//Ahead   []CPUInstructionDecode
	IsRecording bool
	ForceUpdate bool
}

type CPUStack struct {
	Values []int
}

type CPUMemoryRead struct {
	PC      int // Program counter of read instruction
	Address int // Address being read
	Value   int // Value read
}

type CPUMemoryWrite struct {
	PC      int // Program counter of write instruction
	Address int // Address being written
	Value   int // Value being written
}

type CPUMode struct {
	State string
}

type CPUInstructions struct {
	Instructions []CPUInstructionDecode
}

type CPUMemory struct {
	Address int
	Memory  []int
}

type BreakpointActionType int

const (
	BABreak      BreakpointActionType = 0
	BAText       BreakpointActionType = 1
	BAChime      BreakpointActionType = 2
	BATraceOn    BreakpointActionType = 3
	BATraceOff   BreakpointActionType = 4
	BAJump       BreakpointActionType = 5
	BACount      BreakpointActionType = 6
	BASpeed      BreakpointActionType = 7
	BALogToTrace BreakpointActionType = 8
	BARecordOn   BreakpointActionType = 9
	BARecordOff  BreakpointActionType = 10
)

type BreakpointAction struct {
	Type BreakpointActionType `json:"Type,omitempty"`
	Arg0 interface{}          `json:"Arg0,omitempty"`
	Arg1 interface{}          `json:"Arg1,omitempty"`
}

type CPUBreakpoint struct {
	Disabled, Ephemeral bool
	ValueA              *int
	ValueX              *int
	ValueY              *int
	ValuePC             *int
	ValueSP             *int
	ValueP              *int
	WriteAddress        *int
	WriteValue          *int
	ReadAddress         *int
	ReadValue           *int
	Main                *int
	Aux                 *int
	Counter             int
	MaxCount            *int
	Action              *BreakpointAction
}

func (bp *CPUBreakpoint) ShouldBreak(PC, A, X, Y, SP, P, WA, WV, RA, RV, MAIN, AUX int) bool {
	if bp.Disabled {
		return false
	}

	if bp.ValueA != nil && A != *bp.ValueA {
		return false
	}
	if bp.ValueX != nil && X != *bp.ValueX {
		return false
	}
	if bp.ValueY != nil && Y != *bp.ValueY {
		return false
	}
	if bp.ValuePC != nil && PC != *bp.ValuePC {
		return false
	}
	if bp.ValueSP != nil && SP != *bp.ValueSP {
		return false
	}
	if bp.ValueP != nil && P&*bp.ValueP != *bp.ValueP {
		return false
	}
	if bp.WriteAddress != nil && WA != *bp.WriteAddress {
		return false
	}
	if bp.WriteValue != nil && WV != *bp.WriteValue {
		return false
	}
	if bp.ReadAddress != nil && RA != *bp.ReadAddress {
		return false
	}
	if bp.ReadValue != nil && RV != *bp.ReadValue {
		return false
	}
	if bp.Main != nil && MAIN != *bp.Main {
		return false
	}
	if bp.Aux != nil && AUX != *bp.Aux {
		return false
	}
	do := bp.ValueA != nil ||
		bp.ValueX != nil ||
		bp.ValueY != nil ||
		bp.ValuePC != nil ||
		bp.ValueSP != nil ||
		bp.ValueP != nil ||
		bp.WriteAddress != nil ||
		bp.WriteValue != nil ||
		bp.ReadAddress != nil ||
		bp.ReadValue != nil ||
		bp.Main != nil ||
		bp.Aux != nil

	if !do {
		return false
	}

	if bp.MaxCount != nil && *bp.MaxCount > 0 {
		//log.Printf("* bp.MaxCount = %d, incrementing", *bp.MaxCount)
		bp.Counter++
		//log.Printf("* counter is now %d", bp.Counter)
		do = bp.Counter == *bp.MaxCount
		//log.Printf("* condition met = %v", do)
	}

	return do
}

var reBPArg = regexp.MustCompile("(?i)^(OFF|PC|A|X|Y|SP|P|WA|WV|RA|RV|MAIN|AUX|ACTID|ACTTXT|ACTJMP|ACTSPD|EPH|MAXCNT)[=](.+)$")

func (bp *CPUBreakpoint) ParseArg(str string) bool {

	if !reBPArg.MatchString(str) {
		return false
	}
	m := reBPArg.FindAllStringSubmatch(str, -1)
	fieldname := strings.ToUpper(m[0][1])
	value := strings.ToUpper(m[0][2])
	if strings.HasPrefix(value, "$") {
		value = strings.Replace(value, "$", "0x", -1)
	} else if strings.HasPrefix(value, "%") {
		value = strings.Replace(value, "%", "0b", -1)
	}
	v, err := strconv.ParseInt(value, 0, 64)
	if err != nil && fieldname != "ACTTXT" && fieldname != "ACTSPD" {
		return false
	}
	vv := int(v)
	switch fieldname {
	case "MAXCNT":
		bp.MaxCount = &vv
	case "EPH":
		bp.Ephemeral = (vv != 0)
	case "ACTID":
		bp.Action = &BreakpointAction{Type: BreakpointActionType(vv)}
	case "ACTTXT":
		bp.Action.Arg0 = strings.Replace(strings.ToUpper(m[0][2]), "_", " ", -1)
	case "ACTJMP":
		bp.Action.Arg0 = vv
	case "ACTSPD":
		i, _ := strconv.ParseFloat(value, 64)
		if i < 0.25 {
			i = 0.25
		}
		if i > 4 {
			i = 4
		}
		bp.Action.Arg0 = i
	case "OFF":
		bp.Disabled = vv != 0
	case "PC":
		bp.ValuePC = &vv
	case "A":
		bp.ValueA = &vv
	case "X":
		bp.ValueX = &vv
	case "Y":
		bp.ValueY = &vv
	case "SP":
		bp.ValueSP = &vv
	case "P":
		bp.ValueP = &vv
	case "WA":
		bp.WriteAddress = &vv
	case "WV":
		bp.WriteValue = &vv
	case "RA":
		bp.ReadAddress = &vv
	case "RV":
		bp.ReadValue = &vv
	case "MAIN":
		bp.Main = &vv
	case "AUX":
		bp.Aux = &vv
	}

	return true
}

func (bp *CPUBreakpoint) String() string {
	parts := []string(nil)
	if bp.Ephemeral {
		parts = append(parts, "EPH=1")
	}
	if bp.Disabled {
		parts = append(parts, "OFF=1")
	}
	if bp.ValuePC != nil {
		parts = append(parts, fmt.Sprintf("PC=$%.4x", *bp.ValuePC))
	}
	if bp.ValueA != nil {
		parts = append(parts, fmt.Sprintf("A=$%.2x", *bp.ValueA))
	}
	if bp.ValueX != nil {
		parts = append(parts, fmt.Sprintf("X=$%.2x", *bp.ValueX))
	}
	if bp.ValueY != nil {
		parts = append(parts, fmt.Sprintf("Y=$%.2x", *bp.ValueY))
	}
	if bp.ValueSP != nil {
		parts = append(parts, fmt.Sprintf("SP=$%.4x", *bp.ValueSP))
	}
	if bp.ValueP != nil {
		parts = append(parts, fmt.Sprintf("P=$%.2x", *bp.ValueP))
	}
	if bp.WriteAddress != nil {
		parts = append(parts, fmt.Sprintf("WA=$%.4x", *bp.WriteAddress))
	}
	if bp.WriteValue != nil {
		parts = append(parts, fmt.Sprintf("WV=$%.2x", *bp.WriteValue))
	}
	if bp.ReadAddress != nil {
		parts = append(parts, fmt.Sprintf("RA=$%.4x", *bp.ReadAddress))
	}
	if bp.ReadValue != nil {
		parts = append(parts, fmt.Sprintf("RV=$%.2x", *bp.ReadValue))
	}
	if bp.Main != nil {
		parts = append(parts, fmt.Sprintf("MAIN=%d", *bp.Main))
	}
	if bp.Aux != nil {
		parts = append(parts, fmt.Sprintf("AUX=%d", *bp.Aux))
	}
	if bp.MaxCount != nil {
		parts = append(parts, fmt.Sprintf("MAXCNT=%d", *bp.MaxCount))
	}
	return strings.Join(parts, " ")
}

type CPUBreakpointList struct {
	Breakpoints []*CPUBreakpoint
}

type VideoSoftSwitches struct {
	Switches []SoftSwitchInfo
}

type MemorySoftSwitches struct {
	Switches []SoftSwitchInfo
}

type SoftSwitchInfo struct {
	Name           string
	Enabled        bool
	EnabledText    string
	EnableAddress  int
	DisabledText   string
	DisableAddress int
	StatusAddress  int
}

type CPUActionType int

const (
	CatResume CPUActionType = iota
	CatPause
	CatStep
	CatStepOver
	CatStepN
)

type CPUAction struct {
	Type CPUActionType
}

type WebSocketCommand struct {
	Type string
	Args []string
}

type WebSocketMessage struct {
	Type    string
	Payload interface{}
	Ok      bool
}

type MemoryWatchRegion struct {
	Base int
	Size int
}

type MemoryWatchRegionList struct {
	Regions []MemoryWatchRegion
}

type LiveRewindState struct {
	Enabled bool
	// Feature Related
	CanResume  bool
	CanBack    bool
	CanForward bool
	// Playback Related
	TimeFactor  float64
	Backwards   bool
	Block       int
	TotalBlocks int
}
