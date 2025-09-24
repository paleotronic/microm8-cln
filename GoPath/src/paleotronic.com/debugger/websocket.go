package debugger

import (
	"encoding/json"
	"errors"
	log2 "log"
	"strconv"
	"strings"
	"time"

	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/debugger/debugtypes"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
)

func parseNumber(str string) (int, bool) {
	if strings.HasPrefix(str, "$") {
		str = strings.Replace(str, "$", "0x", -1)
	} else if strings.HasPrefix(str, "%") {
		str = strings.Replace(str, "%", "0b", -1)
	}

	if strings.HasPrefix(str, "0b") {
		// parse binary string
		v, err := strconv.ParseInt(str[2:], 2, 64)
		if err != nil {
			return 0, false
		}
		return int(v), true
	}

	v, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		return 0, false
	}
	return int(v), true

}

func getMemoryBanks(addr int, action memory.MemoryAction) (int, int) {
	mm := DebuggerInstance.ent().GetMemoryMap()
	index := DebuggerInstance.ent().GetMemIndex()
	mmu := mm.BlockMapper[index]
	blocks := mmu.GetMappedBlocks(action)
	bank := addr / 256
	if bank == 0xc0 {
		return 0, 0
	}
	b := blocks[bank]

	if b == nil {
		return 0, 0
	}

	m := strings.HasPrefix(b.Label, "main.")
	a := strings.HasPrefix(b.Label, "aux.")

	switch {
	case !m && !a:
		return 0, 0
	case m && !a:
		return 1, 0
	case m && a:
		return 1, 1
	case !m && a:
		return 0, 1
	}

	return 0, 0
}

func getMemblock(addr int, forceAux bool, action memory.MemoryAction) *memory.MemoryBlock {
	mm := DebuggerInstance.ent().GetMemoryMap()
	index := DebuggerInstance.ent().GetMemIndex()
	mmu := mm.BlockMapper[index]
	blocks := mmu.GetMappedBlocks(action)
	bank := (addr / 256) % 256
	if bank == 0xc0 {
		return nil
	}
	b := blocks[bank]
	if b == nil {
		return nil
	}
	if strings.HasPrefix(b.Label, "main.") && forceAux {
		name := b.Label
		name = strings.Replace(name, "main.", "aux.", -1)
		b = mmu.Get(name)
	} else if strings.HasPrefix(b.Label, "aux.") && !forceAux {
		name := b.Label
		name = strings.Replace(name, "aux.", "main.", -1)
		b = mmu.Get(name)
	}
	return b
}

func searchMemory(address int, forceAux bool, values []int) *debugtypes.MemSearchResult {

	var pos = address + 1
	var matches = 0
	var matchStart = -1
	var v int
	for matches < len(values) && pos != address {
		b := getMemblock(pos, forceAux, memory.MA_READ)
		if b == nil || pos/256 == 0xc0 {
			v = -1
		} else {
			var temp uint64
			if b.Do(pos, memory.MA_READ, &temp) {
				v = int(temp & 0xff)
			}
		}
		if v == values[matches] || values[matches] == -1 {
			if matches == 0 {
				matchStart = pos
			}
			matches++
		} else {
			matches = 0
		}
		pos = (pos + 1) % 65536
	}

	if matches < len(values) {
		return nil
	}

	return &debugtypes.MemSearchResult{
		Search:    values,
		FoundAddr: matchStart,
		Aux:       forceAux,
	}

}

func liveRewindState(force bool) *debugtypes.LiveRewindState {
	s := &debugtypes.LiveRewindState{
		Enabled: DebuggerInstance.ent().IsRecordingVideo() || DebuggerInstance.ent().IsPlayingVideo() || force,
	}
	s.CanBack = s.Enabled
	s.CanForward = DebuggerInstance.ent().IsPlayingVideo()
	s.CanResume = DebuggerInstance.ent().IsPlayingVideo()
	if s.CanResume {
		p := DebuggerInstance.ent().GetPlayer()
		s.Backwards = p.IsBackwards()
		s.TimeFactor = p.GetTimeShift()
	}

	return s
}

func handleCommand(msg debugtypes.WebSocketCommand) debugtypes.WebSocketMessage {

	//log.Printf("Received command: %s -> %v", msg.Type, msg.Args)

	switch msg.Type {
	case "live-rewind":
		//rlog.Printf("slotid = %d", DebuggerInstance.slotid)
		//rlog.Printf("IsPlaying = %v", DebuggerInstance.ent().IsPlayingVideo())
		//rlog.Printf("IsRecording = %v", DebuggerInstance.ent().IsRecordingVideo())
		if len(msg.Args) > 0 {
			verb := msg.Args[0]
			switch verb {
			case "jump":
				if len(msg.Args) > 1 {
					state := msg.Args[1]
					count, ok := parseNumber(state)
					if ok && DebuggerInstance.ent().IsPlayingVideo() {
						p := DebuggerInstance.ent().GetPlayer()
						p.Jump(count)
						// servicebus.InjectServiceBusMessage(
						// 	DebuggerInstance.ent().GetMemIndex(),
						// 	servicebus.PlayerJump,
						// 	count,
						// )
					}
				}
			case "set":
				if len(msg.Args) > 1 {
					state := msg.Args[1]
					switch state {
					case "true":
						if DebuggerInstance.ent().IsPlayingVideo() {
							DebuggerInstance.ent().BreakIntoVideo()
						}
						DebuggerInstance.ent().StopRecording()
						DebuggerInstance.ent().RecordToggle(DebuggerInstance.Config.FullCPURecord)
					case "false":
						if DebuggerInstance.ent().IsPlayingVideo() {
							p := DebuggerInstance.ent().GetPlayer()
							p.SetNoResume(true)
							DebuggerInstance.ent().BreakIntoVideo()
						}
						DebuggerInstance.ent().StopRecording()
						return debugtypes.WebSocketMessage{
							Ok: true,
							Payload: &debugtypes.LiveRewindState{
								CanBack:    false,
								CanForward: false,
								CanResume:  false,
								Enabled:    false,
							},
							Type: "live-rewind-response",
						}
					}
				}
			case "pause":
				if !DebuggerInstance.ent().IsPlayingVideo() {
					DebuggerInstance.ent().BackVideo()
					for !DebuggerInstance.ent().IsPlayingVideo() {
						time.Sleep(1 * time.Millisecond)
					}
				}
				//for DebuggerInstance.ent().GetPlayer().GetTimeShift() > 0 {
				servicebus.InjectServiceBusMessage(
					DebuggerInstance.slotid,
					servicebus.PlayerPause,
					"",
				)
				//}
			case "resume-cpu-paused":
				DebuggerInstance.PauseCPU()
				if DebuggerInstance.ent().IsPlayingVideo() {
					servicebus.SendServiceBusMessage(
						DebuggerInstance.slotid,
						servicebus.PlayerResume,
						"",
					)
				}
				cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
				for cpu.RunState != mos6502.CrsPaused {
					time.Sleep(time.Millisecond)
				}
				return debugtypes.WebSocketMessage{
					Ok: true,
					Payload: &debugtypes.CPUState{
						PC:    cpu.PC,
						A:     cpu.A,
						X:     cpu.X,
						Y:     cpu.Y,
						SP:    cpu.SP,
						P:     cpu.P,
						CC:    cpu.GlobalCycles,
						Speed: cpu.UserWarp,
					},
					Type: "pause-response",
				}
			case "resume":
				if DebuggerInstance.ent().IsPlayingVideo() {
					//DebuggerInstance.ent().BreakIntoVideo()
					// DebuggerInstance.ent().StopRecording()
					// DebuggerInstance.ent().RecordToggle()
					servicebus.InjectServiceBusMessage(
						DebuggerInstance.slotid,
						servicebus.PlayerResume,
						"",
					)
				}
			case "reset":
				if DebuggerInstance.ent().ReplayVideo() {
					for !DebuggerInstance.ent().IsPlayingVideo() {
						time.Sleep(1 * time.Millisecond)
					}
					return debugtypes.WebSocketMessage{
						Ok:      true,
						Payload: liveRewindState(true),
						Type:    "live-rewind-response",
					}
				}
			case "back":
				if DebuggerInstance.ent().BackVideo() {
					for !DebuggerInstance.ent().IsPlayingVideo() {
						time.Sleep(1 * time.Millisecond)
					}
					return debugtypes.WebSocketMessage{
						Ok:      true,
						Payload: liveRewindState(true),
						Type:    "live-rewind-response",
					}
				}
			case "forwards":
				if DebuggerInstance.ent().ForwardVideo() {
					for !DebuggerInstance.ent().IsPlayingVideo() {
						time.Sleep(1 * time.Millisecond)
					}
					return debugtypes.WebSocketMessage{
						Ok:      true,
						Payload: liveRewindState(true),
						Type:    "live-rewind-response",
					}
				}
			case "forwards1x":
				if DebuggerInstance.ent().ForwardVideo1x() {
					for !DebuggerInstance.ent().IsPlayingVideo() {
						time.Sleep(1 * time.Millisecond)
					}
					DebuggerInstance.ent().GetPlayer().SetBackwards(false)
					DebuggerInstance.ent().GetPlayer().SetTimeShift(1)
					return debugtypes.WebSocketMessage{
						Ok:      true,
						Payload: liveRewindState(true),
						Type:    "live-rewind-response",
					}
				}
			}
			return debugtypes.WebSocketMessage{
				Ok:      true,
				Payload: liveRewindState(false),
				Type:    "live-rewind-response",
			}
		}
	case "set-heatmap":
		mmu := DebuggerInstance.ent().GetMemoryMap().BlockMapper[DebuggerInstance.ent().GetMemIndex()]
		if len(msg.Args) > 0 {
			// Args: 0 = mode, 1 =  bank
			mode, ok := parseNumber(msg.Args[0])
			if !ok || mode < 0 || mode > int(settings.HMExecCombined) {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Mode invalid",
					Type:    "enable-heatmap-response",
				}
			}
			bank := 0
			if len(msg.Args) > 1 {
				bank, ok = parseNumber(msg.Args[1])
				if !ok || bank < 0 || bank > 255 {
					return debugtypes.WebSocketMessage{
						Ok:      false,
						Payload: "Bank invalid",
						Type:    "enable-heatmap-response",
					}
				}
			}
			log.Printf("setting heatmap mode = %d, bank (%s) = %d", mode, msg.Args[1], bank)
			mmu.SetHeatMapMode(DebuggerInstance.ent().GetMemIndex(), settings.HeatMapMode(mode), bank)
			hm := mmu.HeatMap
			mmu.CoolHeatmap()
			return debugtypes.WebSocketMessage{
				Ok:      true,
				Payload: hm,
				Type:    "heatmap-response",
			}
		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "Args required for enable",
				Type:    "enable-heatmap-response",
			}
		}
	case "get-heatmap":
		// REMOVEME: turn on bank 0 main heatmap
		mmu := DebuggerInstance.ent().GetMemoryMap().BlockMapper[DebuggerInstance.ent().GetMemIndex()]
		hm := mmu.HeatMap
		mmu.CoolHeatmap()
		log2.Printf("Sending heatmap data... %+v", hm)
		return debugtypes.WebSocketMessage{
			Ok:      true,
			Payload: hm,
			Type:    "heatmap-response",
		}
	case "apple-reset-6502":
		cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
		mr, ok := DebuggerInstance.ent().GetMemoryMap().InterpreterMappableAtAddress(DebuggerInstance.ent().GetMemIndex(), 0xc000)
		if ok {
			mr.(*apple2.Apple2IOChip).Reset()
			mr.(*apple2.Apple2IOChip).ResetMemory(false)
			for i := 1024; i < 2048; i++ {
				DebuggerInstance.ent().GetMemoryMap().WriteInterpreterMemorySilent(DebuggerInstance.ent().GetMemIndex(), i, 0xa0)
			}
		}
		cpu.RequestReset(64166)
		DebuggerInstance.ResetCounters()
		return debugtypes.WebSocketMessage{
			Ok: true,
			Payload: &debugtypes.CPUState{
				PC:    cpu.PC,
				A:     cpu.A,
				X:     cpu.X,
				Y:     cpu.Y,
				SP:    cpu.SP,
				P:     cpu.P,
				CC:    cpu.GlobalCycles,
				Speed: cpu.UserWarp,
			},
			Type: "state-response",
		}
	case "reset-6502":
		cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
		cpu.RequestReset(-1)
		return debugtypes.WebSocketMessage{
			Ok: true,
			Payload: &debugtypes.CPUState{
				PC:    cpu.PC,
				A:     cpu.A,
				X:     cpu.X,
				Y:     cpu.Y,
				SP:    cpu.SP,
				P:     cpu.P,
				CC:    cpu.GlobalCycles,
				Speed: cpu.UserWarp,
			},
			Type: "state-response",
		}
	case "get-settings":
		return debugtypes.WebSocketMessage{
			Ok:      true,
			Payload: DebuggerInstance.Config,
			Type:    "get-settings-response",
		}
	case "set-val":
		if len(msg.Args) > 0 {
			var name = msg.Args[0]
			if len(msg.Args) > 1 {
				value, ok := parseNumber(msg.Args[1])
				if !ok {
					return debugtypes.WebSocketMessage{
						Ok:      false,
						Payload: "Setting value invalid",
						Type:    "set-val-response",
					}
				}
				DebuggerInstance.SetVal(name, value)
				if strings.HasPrefix(name, "6502.") {
					cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
					return debugtypes.WebSocketMessage{
						Ok: true,
						Payload: &debugtypes.CPUState{
							PC:    cpu.PC,
							A:     cpu.A,
							X:     cpu.X,
							Y:     cpu.Y,
							SP:    cpu.SP,
							P:     cpu.P,
							CC:    cpu.GlobalCycles,
							Speed: cpu.UserWarp,
						},
						Type: "state-response",
					}
				}
				return debugtypes.WebSocketMessage{
					Ok:      true,
					Payload: DebuggerInstance.Config,
					Type:    "get-settings-response",
				}
			} else {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Setting value required",
					Type:    "set-val-response",
				}
			}
		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "Setting name required",
				Type:    "set-val-response",
			}
		}
	case "trace":
		if len(msg.Args) > 0 {
			verb := msg.Args[0]
			msg := DebuggerInstance.Trace(verb)
			return debugtypes.WebSocketMessage{
				Ok:      true,
				Payload: msg,
				Type:    "trace-response",
			}
		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "mode name required",
				Type:    "trace-response",
			}
		}
	case "toggle-flag":
		if len(msg.Args) > 0 {
			name := msg.Args[0]
			DebuggerInstance.ToggleCPUFlag(name)
			return debugtypes.WebSocketMessage{
				Ok:      true,
				Payload: "Flag " + name + " toggle requested",
				Type:    "toggle-flag-response",
			}
		}
		return debugtypes.WebSocketMessage{
			Ok:      false,
			Payload: "Flag name required",
			Type:    "toggle-flag-response",
		}

	case "toggle-switch":
		if len(msg.Args) > 0 {
			name := msg.Args[0]
			DebuggerInstance.ToggleSoftSwitch(name)
			return debugtypes.WebSocketMessage{
				Ok:      true,
				Payload: "Switch " + name + " toggle requested",
				Type:    "toggle-switch-response",
			}
		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "Switch name required",
				Type:    "toggle-switch-response",
			}
		}
	case "pause":
		DebuggerInstance.PauseCPU()
		cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
		cpu.PauseNextRTS = false
		for cpu.RunState != mos6502.CrsPaused {
			time.Sleep(time.Millisecond)
		}
		return debugtypes.WebSocketMessage{
			Ok: true,
			Payload: &debugtypes.CPUState{
				PC:    cpu.PC,
				A:     cpu.A,
				X:     cpu.X,
				Y:     cpu.Y,
				SP:    cpu.SP,
				P:     cpu.P,
				CC:    cpu.GlobalCycles,
				Speed: cpu.UserWarp,
			},
			Type: "pause-response",
		}
	case "step-out":
		DebuggerInstance.StepCPUOut()
		cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
		// for cpu.RunState == mos6502.CrsStepOut {
		// 	time.Sleep(time.Millisecond)
		// }
		return debugtypes.WebSocketMessage{
			Ok: true,
			Payload: &debugtypes.CPUState{
				PC:    cpu.PC,
				A:     cpu.A,
				X:     cpu.X,
				Y:     cpu.Y,
				SP:    cpu.SP,
				P:     cpu.P,
				CC:    cpu.GlobalCycles,
				Speed: cpu.UserWarp,
			},
			Type: "step-response",
		}
	case "state-save":
		err := DebuggerInstance.SaveState()
		if err != nil {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "Save failed: " + err.Error(),
				Type:    "state-save-response",
			}
		}
		return debugtypes.WebSocketMessage{
			Ok:      true,
			Payload: "Save ok",
			Type:    "state-save-response",
		}
	case "step-over":
		DebuggerInstance.StepCPUOver()
		cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
		for cpu.RunState == mos6502.CrsStepOver {
			time.Sleep(time.Millisecond)
		}
		return debugtypes.WebSocketMessage{
			Ok: true,
			Payload: &debugtypes.CPUState{
				PC:    cpu.PC,
				A:     cpu.A,
				X:     cpu.X,
				Y:     cpu.Y,
				SP:    cpu.SP,
				P:     cpu.P,
				CC:    cpu.GlobalCycles,
				Speed: cpu.UserWarp,
			},
			Type: "step-response",
		}
	case "step":
		DebuggerInstance.StepCPU()
		cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
		for cpu.RunState == mos6502.CrsSingleStep {
			time.Sleep(time.Millisecond)
		}
		return debugtypes.WebSocketMessage{
			Ok: true,
			Payload: &debugtypes.CPUState{
				PC:    cpu.PC,
				A:     cpu.A,
				X:     cpu.X,
				Y:     cpu.Y,
				SP:    cpu.SP,
				P:     cpu.P,
				CC:    cpu.GlobalCycles,
				Speed: cpu.UserWarp,
			},
			Type: "step-response",
		}
	case "continue-out":
		DebuggerInstance.ContinueCPUOut()
		return debugtypes.WebSocketMessage{
			Ok:      true,
			Payload: "continue",
			Type:    "continue-response",
		}
	case "continue":
		DebuggerInstance.ContinueCPU()
		return debugtypes.WebSocketMessage{
			Ok:      true,
			Payload: "continue",
			Type:    "continue-response",
		}
	case "decode-dasm":
		if len(msg.Args) > 0 {
			var address int
			var ok bool
			address, ok = parseNumber(msg.Args[0])
			if !ok {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Address param invalid",
					Type:    "decode-dasm-response",
				}
			}
			var count = DebuggerInstance.Config.CPUInstructionBacklog + DebuggerInstance.Config.CPUInstructionLookahead
			if len(msg.Args) > 1 {
				count, ok = parseNumber(msg.Args[1])
				if !ok {
					return debugtypes.WebSocketMessage{
						Ok:      false,
						Payload: "Count param invalid",
						Type:    "decode-dasm-response",
					}
				}
			}
			cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
			instr := make([]debugtypes.CPUInstructionDecode, int(count))
			for i, _ := range instr {
				code, desc, cycles := cpu.DecodeInstruction(int(address))
				instr[i].Address = int(address)
				instr[i].Bytes = code
				instr[i].Instruction = desc
				instr[i].Cycles = cycles
				address += len(code) % 65536
			}
			return debugtypes.WebSocketMessage{
				Ok: true,
				Payload: &debugtypes.CPUInstructions{
					Instructions: instr,
				},
				Type: "decode-dasm-response",
			}
		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "No params",
				Type:    "decode-dasm-response",
			}
		}
	case "decode":
		if len(msg.Args) > 0 {
			var address int
			var ok bool
			address, ok = parseNumber(msg.Args[0])
			if !ok {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Address param invalid",
					Type:    "decode-response",
				}
			}
			var count = DebuggerInstance.Config.CPUInstructionBacklog + DebuggerInstance.Config.CPUInstructionLookahead
			if len(msg.Args) > 1 {
				count, ok = parseNumber(msg.Args[1])
				if !ok {
					return debugtypes.WebSocketMessage{
						Ok:      false,
						Payload: "Count param invalid",
						Type:    "decode-response",
					}
				}
			}
			cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
			instr := make([]debugtypes.CPUInstructionDecode, int(count))
			for i, _ := range instr {
				if i < cpuHistory {
					code, desc, cycles := cpu.DecodeInstruction(int(DebuggerInstance.LastPC[i]))
					instr[i].Address = int(DebuggerInstance.LastPC[i])
					instr[i].Bytes = code
					instr[i].Instruction = desc
					instr[i].Cycles = cycles
					instr[i].Historic = true
				} else {
					code, desc, cycles := cpu.DecodeInstruction(int(address))
					instr[i].Address = int(address)
					instr[i].Bytes = code
					instr[i].Instruction = desc
					instr[i].Cycles = cycles
					address += len(code) % 65536
				}
			}
			return debugtypes.WebSocketMessage{
				Ok: true,
				Payload: &debugtypes.CPUInstructions{
					Instructions: instr,
				},
				Type: "decode-response",
			}
		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "No params",
				Type:    "decode-response",
			}
		}
	case "sendkey":
		if len(msg.Args) > 0 {
			keycode, ok := parseNumber(msg.Args[0])
			if !ok {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "KeyCode param not valid",
					Type:    "sendkey-response",
				}
			}
			DebuggerInstance.SendKey(keycode)
			return debugtypes.WebSocketMessage{
				Ok:      true,
				Payload: "KeyCode param not valid",
				Type:    "sendkey-response",
			}
		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "KeyCode argument missing",
				Type:    "sendkey-response",
			}
		}
	case "setwarp":
		if len(msg.Args) > 0 {
			value, err := strconv.ParseFloat(msg.Args[0], 64)
			if err != nil {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Warp param invalid",
					Type:    "set-warp-response",
				}
			}
			cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
			cpu.SetWarpUser(value)
			cpu.CalcTiming()
			log.Printf("Setting CPU warp to %f", value)
			return debugtypes.WebSocketMessage{
				Ok:      true,
				Payload: "Warp set",
				Type:    "set-warp-response",
			}
		}
		return debugtypes.WebSocketMessage{
			Ok:      false,
			Payload: "Warp param invalid",
			Type:    "set-warp-response",
		}
	case "getstack":
		if len(msg.Args) > 0 {
			max, ok := parseNumber(msg.Args[0])
			if !ok {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "max param invalid",
					Type:    "getstack-response",
				}
			}
			cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
			return debugtypes.WebSocketMessage{
				Ok: true,
				Payload: &debugtypes.CPUStack{
					Values: cpu.GetStack(max),
				},
				Type: "getstack-response",
			}
		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "max param invalid",
				Type:    "getstack-response",
			}
		}
	case "setpc":
		if len(msg.Args) > 0 {
			address, ok := parseNumber(msg.Args[0])
			if !ok {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Address param invalid",
					Type:    "setpc-response",
				}
			}
			cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
			cpu.PC = address
			return debugtypes.WebSocketMessage{
				Ok:      true,
				Payload: "PC set",
				Type:    "setpc-response",
			}
		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "Address param invalid",
				Type:    "setpc-response",
			}
		}
	case "memsearch":
		// args: startaddress, auxbank, sequence
		if len(msg.Args) < 3 {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "memsearch needs 3 args minimum",
				Type:    "memsearch-response",
			}
		}
		// address
		address, ok := parseNumber(msg.Args[0])
		if !ok {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "Address param invalid",
				Type:    "memsearch-response",
			}
		}
		// auxmain
		aux, ok := parseNumber(msg.Args[1])
		if !ok {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "Aux param invalid",
				Type:    "memsearch-response",
			}
		}
		useaux := aux != 0
		// sequence...
		values := make([]int, 0)
		for _, s := range msg.Args[2:] {
			if strings.HasPrefix(s, "?") {
				values = append(values, -1)
				continue
			}
			v, ok := parseNumber(s)
			if !ok {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Invalid value " + s,
					Type:    "memsearch-response",
				}
			}
			values = append(values, int(v))
		}
		r := searchMemory(address, useaux, values)

		if r != nil {
			// TODO: Replace with real search...
			return debugtypes.WebSocketMessage{
				Ok:      true,
				Payload: r,
				Type:    "memsearch-response",
			}
		}
	case "get-memlock":
		if len(msg.Args) > 0 {
			address, ok := parseNumber(msg.Args[0])
			if !ok {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Address param invalid",
					Type:    "memset-response",
				}
			}
			cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
			_, ok = cpu.LockValue[address]
			return debugtypes.WebSocketMessage{
				Ok:      true,
				Payload: ok,
				Type:    "get-memlock-response",
			}
		}
	case "memset":
		if len(msg.Args) > 0 {
			address, ok := parseNumber(msg.Args[0])
			if !ok {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Address param invalid",
					Type:    "memset-response",
				}
			}
			var value = 0
			if len(msg.Args) > 1 {
				value, ok = parseNumber(msg.Args[1])
				if !ok || value < 0 || value > 0xff {
					return debugtypes.WebSocketMessage{
						Ok:      false,
						Payload: "Value param invalid",
						Type:    "memset-response",
					}
				}
			}
			var count = 1
			var forceAux = 0
			var lockMem = 0
			if len(msg.Args) > 2 {
				forceAux, ok = parseNumber(msg.Args[2])
				if !ok {
					return debugtypes.WebSocketMessage{
						Ok:      false,
						Payload: "ForceAux param invalid",
						Type:    "memory-response",
					}
				}
			}
			if len(msg.Args) > 3 {
				lockMem, ok = parseNumber(msg.Args[3])
				if !ok {
					return debugtypes.WebSocketMessage{
						Ok:      false,
						Payload: "LockMem param invalid",
						Type:    "memory-response",
					}
				}
				// Apply or remove lock
				cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
				if lockMem != 0 {
					cpu.LockValue[address] = uint64(value & 0xff)
				} else {
					delete(cpu.LockValue, address)
				}
			}
			for i := 0; i < count; i++ {
				b := getMemblock(address+i%65536, forceAux != 0, memory.MA_WRITE)
				if b != nil {
					var v uint64
					b.Do(address+i%65536, memory.MA_READ, &v)
					v = (v & 0xffffffffffffff00) | uint64(value&0xff)
					b.Do(address+i%65536, memory.MA_WRITE, &v)
				}
			}
			return debugtypes.WebSocketMessage{
				Ok:      true,
				Payload: fmt.Sprintf("Wrote $%.2x %d times from $%.4x", value, count, address),
				Type:    "memset-response",
			}
		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "No params",
				Type:    "decode-response",
			}
		}
	case "memory":
		if len(msg.Args) > 0 {
			address, ok := parseNumber(msg.Args[0])
			if !ok {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Address param invalid",
					Type:    "memory-response",
				}
			}
			var count = 20
			if len(msg.Args) > 1 {
				count, ok = parseNumber(msg.Args[1])
				if !ok {
					return debugtypes.WebSocketMessage{
						Ok:      false,
						Payload: "Count param invalid",
						Type:    "memory-response",
					}
				}
			}
			var forceAux = 0
			if len(msg.Args) > 2 {
				forceAux, ok = parseNumber(msg.Args[2])
				if !ok {
					return debugtypes.WebSocketMessage{
						Ok:      false,
						Payload: "ForceAux param invalid",
						Type:    "memory-response",
					}
				}
			}
			m := make([]int, int(count))
			for i, _ := range m {
				b := getMemblock(address+i%65536, forceAux != 0, memory.MA_READ)
				if b == nil {
					m[i] = 0x00
				} else {
					var v uint64
					if b.Do(address+i%65536, memory.MA_READ, &v) {
						m[i] = int(v) & 0xff
					}
				}
			}
			return debugtypes.WebSocketMessage{
				Ok: true,
				Payload: &debugtypes.CPUMemory{
					Memory:  m,
					Address: int(address),
				},
				Type: "memory-response",
			}
		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "No params",
				Type:    "decode-response",
			}
		}
	case "detach":
		if DebuggerInstance.ent().IsRecordingDiscVideo() {
			DebuggerInstance.ent().StopRecording()
		} else if DebuggerInstance.ent().IsPlayingVideo() {
			p := DebuggerInstance.ent().GetPlayer()
			p.SetNoResume(true)
			DebuggerInstance.ent().BreakIntoVideo()
		}
		DebuggerInstance.ContinueCPU()
		DebuggerInstance.Detach()
		cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
		cpu.SetWarpUser(1.0)
	case "attach":
		if len(msg.Args) > 0 {
			slotid, ok := parseNumber(msg.Args[0])
			if ok && slotid >= 0 && slotid <= 7 {
				DebuggerInstance.AttachSlot(int(slotid))
				return debugtypes.WebSocketMessage{
					Ok:      true,
					Payload: slotid,
					Type:    "attach-response",
				}
			} else {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Failed to attach",
					Type:    "attach-response",
				}
			}
		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "Failed to attach",
				Type:    "attach-response",
			}
		}
	case "listbp":
		return debugtypes.WebSocketMessage{
			Ok: true,
			Payload: &debugtypes.CPUBreakpointList{
				Breakpoints: DebuggerInstance.GetBreakpoints(),
			},
			Type: "listbp-response",
		}
	case "enbp":
		if len(msg.Args) > 0 {
			bpid, ok := parseNumber(msg.Args[0])
			bpid -= 1
			if !ok && msg.Args[0] == "*" {
				DebuggerInstance.EnableAllBreakpoints()
				return debugtypes.WebSocketMessage{
					Ok: true,
					Payload: &debugtypes.CPUBreakpointList{
						Breakpoints: DebuggerInstance.GetBreakpoints(),
					},
					Type: "listbp-response",
				}
			} else if ok && bpid >= 0 && int(bpid) < len(DebuggerInstance.Breakpoints) {
				DebuggerInstance.EnableBreakpoint(int(bpid))
				return debugtypes.WebSocketMessage{
					Ok: true,
					Payload: &debugtypes.CPUBreakpointList{
						Breakpoints: DebuggerInstance.GetBreakpoints(),
					},
					Type: "listbp-response",
				}
			} else {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Failed to disable breakpoint",
					Type:    "clrbp-response",
				}
			}
		}
	case "disbp":
		if len(msg.Args) > 0 {
			bpid, ok := parseNumber(msg.Args[0])
			bpid -= 1
			if !ok && msg.Args[0] == "*" {
				DebuggerInstance.DisableAllBreakpoints()
				return debugtypes.WebSocketMessage{
					Ok: true,
					Payload: &debugtypes.CPUBreakpointList{
						Breakpoints: DebuggerInstance.GetBreakpoints(),
					},
					Type: "listbp-response",
				}
			} else if ok && bpid >= 0 && int(bpid) < len(DebuggerInstance.Breakpoints) {
				DebuggerInstance.DisableBreakpoint(int(bpid))
				return debugtypes.WebSocketMessage{
					Ok: true,
					Payload: &debugtypes.CPUBreakpointList{
						Breakpoints: DebuggerInstance.GetBreakpoints(),
					},
					Type: "listbp-response",
				}
			} else {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Failed to disable breakpoint",
					Type:    "clrbp-response",
				}
			}
		}
	case "clrbp":
		if len(msg.Args) > 0 {
			bpid, ok := parseNumber(msg.Args[0])
			bpid -= 1
			if !ok && msg.Args[0] == "*" {
				DebuggerInstance.ClearAllBreakpoints()
				return debugtypes.WebSocketMessage{
					Ok: true,
					Payload: &debugtypes.CPUBreakpointList{
						Breakpoints: DebuggerInstance.GetBreakpoints(),
					},
					Type: "listbp-response",
				}
			} else if ok && bpid >= 0 && int(bpid) < len(DebuggerInstance.Breakpoints) {
				DebuggerInstance.RemoveBreakpoint(int(bpid))
				return debugtypes.WebSocketMessage{
					Ok: true,
					Payload: &debugtypes.CPUBreakpointList{
						Breakpoints: DebuggerInstance.GetBreakpoints(),
					},
					Type: "listbp-response",
				}
			} else {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Failed to remove breakpoint",
					Type:    "clrbp-response",
				}
			}
		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "Failed to attach",
				Type:    "clrbp-response",
			}
		}
	case "setbp":
		if len(msg.Args) > 0 {
			log.Printf("updbp args=[%v]", msg.Args)
			var bp = &debugtypes.CPUBreakpoint{}

			for _, arg := range msg.Args {
				if !bp.ParseArg(arg) {
					return debugtypes.WebSocketMessage{
						Ok:      false,
						Payload: "Failed to set breakpoint: bad arg: " + arg,
						Type:    "setbp-response",
					}
				}
			}

			DebuggerInstance.AddBreakpoint(bp)
			return debugtypes.WebSocketMessage{
				Ok: true,
				Payload: &debugtypes.CPUBreakpointList{
					Breakpoints: DebuggerInstance.GetBreakpoints(),
				},
				Type: "listbp-response",
			}

		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "Failed to set breakpoint",
				Type:    "setbp-response",
			}
		}
	case "setbpcounter":
		if len(msg.Args) > 0 {
			idx, ok := parseNumber(msg.Args[0])
			if !ok {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Failed to update breakpoint counter: bad index: " + msg.Args[0],
					Type:    "setbp-response",
				}
			}
			if idx < 0 || idx >= len(DebuggerInstance.Breakpoints) {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Failed to update breakpoint counter: out of bounds index: " + msg.Args[0],
					Type:    "setbp-response",
				}
			}
			value := 0
			if len(msg.Args) > 1 {
				value, ok = parseNumber(msg.Args[1])
				if !ok {
					return debugtypes.WebSocketMessage{
						Ok:      false,
						Payload: "Failed to update breakpoint counter: bad value: " + msg.Args[0],
						Type:    "setbp-response",
					}
				}
				DebuggerInstance.SetBreakpointCounter(idx, value)
				return debugtypes.WebSocketMessage{
					Ok: true,
					Payload: &debugtypes.CPUBreakpointList{
						Breakpoints: DebuggerInstance.GetBreakpoints(),
					},
					Type: "listbp-response",
				}
			}
		}
	case "get-lastbpmsg":
		if DebuggerInstance.LastBPMessage != "" {
			msg := DebuggerInstance.LastBPMessage
			DebuggerInstance.LastBPMessage = ""
			return debugtypes.WebSocketMessage{
				Ok:      true,
				Payload: msg,
				Type:    "debug-message",
			}
		}
	case "updbp":
		if len(msg.Args) > 0 {
			log.Printf("updbp args=[%v]", msg.Args)
			idx, ok := parseNumber(msg.Args[0])
			if !ok {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Failed to update breakpoint: bad index: " + msg.Args[0],
					Type:    "setbp-response",
				}
			}
			if idx < 0 || idx >= len(DebuggerInstance.Breakpoints) {
				return debugtypes.WebSocketMessage{
					Ok:      false,
					Payload: "Failed to update breakpoint: out of bounds index: " + msg.Args[0],
					Type:    "setbp-response",
				}
			}

			var bp = &debugtypes.CPUBreakpoint{}

			for _, arg := range msg.Args[1:] {
				if !bp.ParseArg(arg) {
					return debugtypes.WebSocketMessage{
						Ok:      false,
						Payload: "Failed to update breakpoint: bad arg: " + arg,
						Type:    "setbp-response",
					}
				}
			}

			DebuggerInstance.UpdateBreakpoint(idx, bp)
			return debugtypes.WebSocketMessage{
				Ok: true,
				Payload: &debugtypes.CPUBreakpointList{
					Breakpoints: DebuggerInstance.GetBreakpoints(),
				},
				Type: "listbp-response",
			}

		} else {
			return debugtypes.WebSocketMessage{
				Ok:      false,
				Payload: "Failed to set breakpoint",
				Type:    "setbp-response",
			}
		}
	}
	return debugtypes.WebSocketMessage{
		Ok:      false,
		Payload: "Unrecognised command",
		Type:    "error",
	}
}

func (d *Debugger) sendMessage(msg *debugtypes.WebSocketMessage) error {
	if d.socket == nil {
		return errors.New("Not connected")
	}
	j, _ := json.Marshal(msg)
	return d.socket.WriteMessage(1, j)
}

func (d *Debugger) SendMessage(kind string, payload interface{}, ok bool) error {
	msg := &debugtypes.WebSocketMessage{
		Type:    kind,
		Payload: payload,
		Ok:      ok,
	}

	defer func() {
		r := recover()
		if r != nil {
			log.Printf("Not sent")
		}
	}()

	d.sendQueue <- msg
	return nil
}
