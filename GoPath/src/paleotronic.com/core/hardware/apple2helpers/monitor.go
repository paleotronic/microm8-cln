package apple2helpers

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"log"

	"paleotronic.com/debugger/debugtypes"

	"github.com/atotto/clipboard"
	"paleotronic.com/core/hardware/cpu"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/hardware/cpu/mos6502/asm"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/vduconst" //	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/octalyzer/bus" //"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/runestring"    //	"paleotronic.com/core/memory"
	"paleotronic.com/utils"

	debugclient "paleotronic.com/debugger/client"
)

type ROMHandler interface {
	DoCall(addr int, caller interfaces.Interpretable, passtocpu bool) bool
}

var (
	reBreak = regexp.MustCompile("^([A-Fa-f0-9]+)?I$")

	reBreakR = regexp.MustCompile("^([A-Fa-f0-9]+)?R$")
	reBreakW = regexp.MustCompile("^([A-Fa-f0-9]+)?W$")
	reBreakZ = regexp.MustCompile("^([A-Fa-f0-9]+)?Z$")

	reTrace    = regexp.MustCompile("^([A-Fa-f0-9]+)?T$")
	reStep     = regexp.MustCompile("^([A-Fa-f0-9]+)?S$")
	reExec     = regexp.MustCompile("^([A-Fa-f0-9]+)?G$")
	reDisass   = regexp.MustCompile("^([A-Fa-f0-9]+)?L$")
	reDisplay  = regexp.MustCompile("^([A-Fa-f0-9]+)([.]([A-Fa-f0-9]+))?$")
	reEntry    = regexp.MustCompile("^([A-Fa-f0-9]+)?:(([ ]*[A-Fa-f0-9]{2}){1,16})(.+)?$")
	reEntryASM = regexp.MustCompile("^([A-Fa-f0-9]+)?:(([ ]*[A-Za-z]{3})([ \t]*.*))$")
	reMove     = regexp.MustCompile("^([A-Fa-f0-9]+)[<]([A-Fa-f0-9]+)[.]([A-Fa-f0-9]+)M$")
	reVerify   = regexp.MustCompile("^([A-Fa-f0-9]+)[<]([A-Fa-f0-9]+)[.]([A-Fa-f0-9]+)V$")
)

type Monitor struct {
	Int     interfaces.Interpretable
	CPU     *mos6502.Core6502
	Buffer  string
	Asm     *asm.Asm6502
	command string
	c       *debugclient.DebugClient
	paused  bool
	lines   chan string
}

func NewMonitor(i interfaces.Interpretable) *Monitor {
	this := &Monitor{
		Int:   i,
		lines: make(chan string, 4096),
	}

	this.CPU = GetCPU(i)
	this.CPU.ROM = DoCall

	this.c = debugclient.NewDebugClient(i.GetMemIndex(), "localhost", "9502", "/api/websocket/debug")

	return this
}

func NewMonitorInvoke(i interfaces.Interpretable, pc int, a, x, y int, sp int) *Monitor {
	this := NewMonitor(i)

	this.CPU = GetCPU(i)
	this.CPU.ROM = DoCall
	this.CPU.PC = pc
	this.CPU.A = a
	this.CPU.X = x
	this.CPU.Y = y
	this.CPU.Set_nz(this.CPU.A)
	this.CPU.SP = sp

	return this
}

func (this *Monitor) Msg(s string) {
	this.lines <- s
}

func (this *Monitor) GetCRTLine(promptString string) string {

	command := ""
	collect := true
	display := this.Int

	this.Int.GetMemoryMap().KeyBufferEmpty(this.Int.GetMemIndex())

	cb := this.Int.GetProducer().GetMemoryCallback(this.Int.GetMemIndex())

	display.PutStr(promptString)

	if cb != nil {
		cb(this.Int.GetMemIndex())
	}

	for collect {
		if cb != nil {
			cb(this.Int.GetMemIndex())
		}

		TextShowCursor(this.Int)
		for this.Int.GetMemoryMap().KeyBufferPeek(this.Int.GetMemIndex()) == 0 && this.Buffer == "" {

			if len(this.Int.GetPasteBuffer().Runes) > 0 {
				this.Buffer = string(this.Int.GetPasteBuffer().Runes)
				this.Int.SetPasteBuffer(runestring.Cast(""))
				fmt.Printf("Putting into buffer [%s]\n", this.Buffer)
			}

			if this.Int.GetMemoryMap().IntGetSlotInterrupt(this.Int.GetMemIndex()) {
				TextSaveScreen(this.Int)
				this.Int.DoCatalog()
				MonitorPanel(this.Int, true)
				//settings.SlotInterrupt[ent.GetMemIndex()] = false
				this.Int.GetMemoryMap().IntSetSlotInterrupt(this.Int.GetMemIndex(), false)
				TextRestoreScreen(this.Int)
			}

			if this.Int.GetMemoryMap().IntGetSlotRestart(this.Int.GetMemIndex()) {
				log.Printf("Exiting due to restart")
				return ""
			}

			if this.Int.GetMemoryMap().IntGetSlotMenu(this.Int.GetMemIndex()) {
				return ""
			}

			//	for this.Int.GetMemory(49152) < 128 && this.Buffer == "" {
			time.Sleep(1 * time.Millisecond)

			if this.Int.VM().IsDying() {
				return ""
			}
		}
		TextHideCursor(this.Int)

		var ch rune

		if this.Buffer != "" {
			ch = rune(this.Buffer[0])
			this.Buffer = this.Buffer[1:]
		} else {
			ch = rune(this.Int.GetMemoryMap().KeyBufferGet(this.Int.GetMemIndex()))
		}
		this.Int.SetMemory(49168, 0)
		switch ch {
		case vduconst.PASTE:
			// paste code
			this.Buffer, _ = clipboard.ReadAll()
		case vduconst.SHIFT_CTRL_V:
			// paste code
			this.Buffer, _ = clipboard.ReadAll()
		case 3:
			return "q"
		case 10:
			{
				if command != "" {
					display.SetSuppressFormat(true)
					display.PutStr("\r\n")
					display.SetSuppressFormat(false)
					return command
				}
			}
		case 13:
			{
				if command != "" {
					display.SetSuppressFormat(true)
					display.PutStr("\r\n")
					display.SetSuppressFormat(false)
					return command
				}
			}
		case 127:
			{
				if len(command) > 0 {
					command = utils.Copy(command, 1, len(command)-1)
					display.Backspace()
					display.SetSuppressFormat(true)
					display.PutStr(" ")
					display.SetSuppressFormat(false)
					display.Backspace()
					if cb != nil {
						cb(this.Int.GetMemIndex())
					}
				}
				break
			}
		default:
			{

				display.SetSuppressFormat(true)
				display.RealPut(rune(ch))
				display.SetSuppressFormat(false)

				if cb != nil {
					cb(this.Int.GetMemIndex())
				}

				command = command + string(ch)
				break
			}
		}
	}

	if cb != nil {
		cb(this.Int.GetMemIndex())
	}

	return command

}

func (this *Monitor) ScreenOn(c int) {
	MonitorPanel(this.Int, true)

	//~ this.Int.SetMemory(65536-16304, 0) // GFX
	//~ this.Int.SetMemory(65536-16297, 0) // Hires
	//~ this.Int.SetMemory(65536-16299, 0) // Page 2
	//~ this.Int.SetMemory(65536-16302, 0) // Fullscreen
	//~ this.Int.SetMemory(65536-16303, 0)

	SetFGColor(this.Int, 15)
	SetBGColor(this.Int, 2)

	if c > 0 {
		TEXT80(this.Int)
	}

}

func (this *Monitor) ScreenOff() {
	MonitorPanel(this.Int, false)
	if !settings.PureBoot(this.Int.GetMemIndex()) {
		this.Int.SetMemory(0xc051, 0) // TEXT
		this.Int.SetMemory(0xc056, 0) // LORES
		this.Int.SetMemory(0xc054, 0) // Page 1
		this.Int.SetMemory(0xc053, 0) // Fullscreen
		this.PutStr("Test message\r\n")
	}
}

func (this *Monitor) Invoke(r cpu.FEResponse) bool {
	this.ScreenOn(1)
	this.PutStr(fmt.Sprintf("%s triggered at $%.4X.\r\n", r.String(), this.CPU.PC))
	this.connect()
	this.run()
	MonitorPanel(this.Int, false)
	if !settings.PureBoot(this.Int.GetMemIndex()) {
		TEXT40(this.Int)
	}
	return !this.CPU.Halted
}

func (this *Monitor) Break() {
	this.ScreenOn(1)
	this.PutStr(fmt.Sprintf("%s reached at $%.4X.\r\n", "Breakpoint", this.CPU.PC))
	//~ this.PutStr(fmt.Sprintf(
	//~ "  A=$%.2x X=$%.2x Y=$%.2x PC=$%.4x P=$%.2x (%s)\r\n",
	//~ this.CPU.A,
	//~ this.CPU.X,
	//~ this.CPU.Y,
	//~ this.CPU.PC,
	//~ this.CPU.P,
	//~ this.CPU.FlagString(),
	//~ ))
	this.connect()
	this.run()
	MonitorPanel(this.Int, false)
	if !settings.PureBoot(this.Int.GetMemIndex()) {
		TEXT40(this.Int)
	}
}

func (this *Monitor) BreakMemory() {
	this.ScreenOn(1)
	this.PutStr(fmt.Sprintf("%s reached at $%.4X.\r\n", "Memory breakpoint", this.CPU.PC))
	//~ this.PutStr(fmt.Sprintf(
	//~ "  A=$%.2x X=$%.2x Y=$%.2x PC=$%.4x P=$%.2x (%s)\r\n",
	//~ this.CPU.A,
	//~ this.CPU.X,
	//~ this.CPU.Y,
	//~ this.CPU.PC,
	//~ this.CPU.P,
	//~ this.CPU.FlagString(),
	//~ ))
	this.connect()
	this.run()
	MonitorPanel(this.Int, false)
	if !settings.PureBoot(this.Int.GetMemIndex()) {
		TEXT40(this.Int)
	}
}

func (this *Monitor) BreakInterrupt() {
	this.ScreenOn(1)
	this.PutStr(fmt.Sprintf("%s reached at $%.4X.\r\n", "BRK", this.CPU.PC))
	//~ this.PutStr(fmt.Sprintf(
	//~ "  A=$%.2x X=$%.2x Y=$%.2x PC=$%.4x P=$%.2x (%s)\r\n",
	//~ this.CPU.A,
	//~ this.CPU.X,
	//~ this.CPU.Y,
	//~ this.CPU.PC,
	//~ this.CPU.P,
	//~ this.CPU.FlagString(),
	//~ ))
	this.connect()
	this.run()

	MonitorPanel(this.Int, false)
	if !settings.PureBoot(this.Int.GetMemIndex()) {
		TEXT40(this.Int)
	}
}

func (this *Monitor) IllegalOpcode() {
	this.ScreenOn(1)
	this.PutStr(fmt.Sprintf("Illegal opcode ($%.2X) reached at $%.4X.\r\n", this.Int.GetMemory(this.CPU.PC)&0xff, this.CPU.PC))
	this.connect()
	this.run()
	MonitorPanel(this.Int, false)
	if !settings.PureBoot(this.Int.GetMemIndex()) {
		TEXT40(this.Int)
	}
}

func (this *Monitor) MemoryProtect() {
	this.ScreenOn(1)
	this.PutStr(fmt.Sprintf("Memory violation occurred at $%.4X.\r\n", this.CPU.PC))
	//~ this.PutStr(fmt.Sprintf(
	//~ "  A=$%.2x X=$%.2x Y=$%.2x PC=$%.4x P=$%.2x (%s)\r\n",
	//~ this.CPU.A,
	//~ this.CPU.X,
	//~ this.CPU.Y,
	//~ this.CPU.PC,
	//~ this.CPU.P,
	//~ this.CPU.FlagString(),
	//~ ))
	this.run()
	MonitorPanel(this.Int, false)
	if !settings.PureBoot(this.Int.GetMemIndex()) {
		TEXT40(this.Int)
	}
}

func (this *Monitor) Manual(command string) {
	this.ScreenOn(1)
	time.Sleep(1000 * time.Millisecond)
	this.PutStr("Monitor started.\r\n")
	//~ this.PutStr(fmt.Sprintf(
	//~ "  A=$%.2x X=$%.2x Y=$%.2x PC=$%.4x P=$%.2x (%s)\r\n",
	//~ this.CPU.A,
	//~ this.CPU.X,
	//~ this.CPU.Y,
	//~ this.CPU.PC,
	//~ this.CPU.P,
	//~ this.CPU.FlagString(),
	//~ ))
	time.Sleep(1000 * time.Millisecond)
	this.command = command
	this.connect()
	this.run()
	MonitorPanel(this.Int, false)
	if !settings.PureBoot(this.Int.GetMemIndex()) {
		TEXT40(this.Int)
	}
}

func (this *Monitor) attach(msg *debugtypes.WebSocketMessage) {
	this.Msg("Attached to debugger backend...\r\n")
	this.c.SendCommand("pause", nil)
}

func (this *Monitor) pause(msg *debugtypes.WebSocketMessage) {
	this.Msg("CPU has paused.\r\n")
	this.paused = true
}

func (this *Monitor) breakpoint(msg *debugtypes.WebSocketMessage) {
	this.ScreenOn(1)
	this.Msg(msg.Payload.(string) + "\r\n")
	this.paused = true
	go this.run()
}

func (this *Monitor) handleErr(msg *debugtypes.WebSocketMessage) {
	this.Msg("Error: " + msg.Payload.(string) + ".\r\n")
	this.paused = true
}

func (this *Monitor) general(msg *debugtypes.WebSocketMessage) {
	if msg.Ok {
		switch msg.Type {
		case "pause-response":
			this.Msg("CPU is paused.\r\n")
		case "listbp-response":
			this.Msg("Breakpoints updated.\r\n")

			bmap, ok := msg.Payload.(map[string]interface{})
			if ok {
				blist := bmap["Breakpoints"].([]interface{})
				for i, b := range blist {
					bp := b.(map[string]interface{})
					bpid := i + 1
					if bp["ValuePC"] != nil {
						this.Msg(fmt.Sprintf("%3d PC=$%.4x\r\n", bpid, int(bp["ValuePC"].(float64))))
					} else if bp["WriteAddress"] != nil {
						this.Msg(fmt.Sprintf("%3d WA=$%.4x\r\n", bpid, int(bp["WriteAddress"].(float64))))
					} else if bp["ReadAddress"] != nil {
						this.Msg(fmt.Sprintf("%3d RA=$%.4x\r\n", bpid, int(bp["ReadAddress"].(float64))))
					}
				}
			}

		case "attach-response":
			this.Msg("Attached.\r\n")
		}
	} else {
		this.Msg("Not Ok\r\n")
	}
	this.paused = true
}

func (this *Monitor) connect() {
	this.c.Handle("break-response", this.breakpoint)
	// this.c.Handle("attach-response", this.attach)
	// this.c.Handle("pause-response", this.pause)
	// this.c.Handle("error", this.handleErr)
	this.c.Handle("*", this.general)
	this.paused = false
	this.c.Connect()
	for !this.paused {
		time.Sleep(time.Millisecond)
	}
}

func (this *Monitor) sendCommand(cmd string, args []string) {
	this.paused = false
	this.c.SendCommand(cmd, args)
	for !this.paused {
		time.Sleep(time.Millisecond)
	}
}

func (this *Monitor) disconnect() {
	this.c.Disconnect()
}

func (this *Monitor) PutStr(s string) {
	moni := GETHUD(this.Int, "MONI")
	moni.Control.PutStr(s)
}

func (this *Monitor) run() {

	// settings.BlockCSR[this.Int.GetMemIndex()] = true
	// defer func() {
	// 	settings.BlockCSR[this.Int.GetMemIndex()] = false
	// }()

	// save existing vstate
	bus.StartDefault()

	var memptr int = this.CPU.PC
	var tracenum int

	//Clearscreen(this.Int)
	mm := this.Int.GetMemoryMap()
	index := this.Int.GetMemIndex()

	for {

		for len(this.lines) > 0 {
			s := <-this.lines
			this.PutStr(s)
		}

		if mm.IntGetSlotMenu(index) {
			bus.StartDefault()
			this.Int.WaitForWorld()
			bus.StartDefault()
			continue
		}

		if mm.IntGetSlotInterrupt(index) {
			bus.StartDefault()
			this.Int.WaitForWorld()
			bus.StartDefault()
			continue
		}

		if mm.IntGetSlotRestart(index) {
			//debug.PrintStack()
			settings.DebuggerAttachSlot = -1
			this.Int.WaitForWorld()
			bus.StartDefault()
			return // drop out of monitor -- slot restart
		}

		this.PutStr(fmt.Sprintf(
			"  A=$%.2x X=$%.2x Y=$%.2x PC=$%.4x P=$%.2x (%s)\r\n",
			this.CPU.A,
			this.CPU.X,
			this.CPU.Y,
			this.CPU.PC,
			this.CPU.P,
			this.CPU.FlagString(),
		))

		var command string

		if this.command != "" {
			command = this.command
			this.PutStr("\r\n*" + command + "\r\n")
			this.command = ""
		} else {
			command = strings.TrimSpace(strings.ToUpper(this.GetCRTLine("\r\n*")))
			if mm.IntGetSlotRestart(index) {
				log.Printf("Restarting...")
				//debug.PrintStack()
				settings.DebuggerAttachSlot = -1
				return // drop out of monitor -- slot restart
			}
			if command == "" {
				continue
			}
		}

		if command == "Q" {
			this.disconnect()
			this.ScreenOff()

			MODE40(this.Int)
			TEXT40(this.Int)
			this.Int.GetDialect().InitVDU(this.Int, true)

			fmt.Println("EXIT HERE")

			this.Int.Halt6502(0)

			if settings.LaunchQuitCPUExit {
				os.Exit(0)
			}

			return
		}

		if reEntry.MatchString(command) {
			m := reEntry.FindAllStringSubmatch(command, -1)
			//fmt.Printf("%v\n", m)
			//fmt.Printf("%s\n", m[0][2])
			start := m[0][1]
			if start != "" {
				// set memptr
				i, _ := strconv.ParseInt(start, 16, 64)
				if i < 0x20000 {
					memptr = int(i)
				}
			}

			values := strings.Split(strings.Trim(m[0][2], " "), " ")
			for _, s := range values {
				s = strings.Trim(s, " ")
				v, _ := strconv.ParseInt(s, 16, 64)
				this.Int.SetMemory(memptr, uint64(v))
				memptr++
			}
			continue
		}

		if reEntryASM.MatchString(command) {
			m := reEntryASM.FindAllStringSubmatch(command, -1)
			//fmt.Printf("%v\n", m)
			//fmt.Printf("%s\n", m[0][2])
			start := m[0][1]
			if start != "" {
				// set memptr
				i, _ := strconv.ParseInt(start, 16, 64)
				if i < 0x20000 {
					memptr = int(i)
				}
			}

			asmcode := m[0][2]

			if this.Asm == nil {
				this.Asm = asm.NewAsm6502Custom(this.CPU)
			}
			asmlines := []string{asmcode}
			this.Asm.PassCount = 0
			codeblocks, _, _, e := this.Asm.Assemble(asmlines, memptr)
			if e != nil {
				Beep(this.Int)
				this.PutStr(e.Error())
			} else {
				// add code
				this.PutStr(fmt.Sprintf("%.4X:", memptr))
				for _, v := range codeblocks[memptr] {
					this.Int.SetMemory(memptr, uint64(v))
					this.PutStr(fmt.Sprintf(" %.2X", v))
					memptr++
				}
				this.PutStr("\r\n")
			}
			this.PutStr(fmt.Sprintf("%.4X:\r\n", memptr))

			continue
		}

		if reBreak.MatchString(command) {
			m := reBreak.FindAllStringSubmatch(command, -1)
			if m[0][1] != "" {
				start, _ := strconv.ParseInt(m[0][1], 16, 64)
				memptr = int(start)
			}

			// if this.CPU.AreMemFlagsSet(memptr, mos6502.AF_EXEC_BREAK) {
			// 	// clear
			// 	//delete(this.CPU.ExecBreakpoint, memptr)

			// 	this.CPU.SetMemFlags(memptr, mos6502.AF_EXEC_BREAK, false)

			// 	this.PutStr(fmt.Sprintf("Breakpoint removed at $%.4X.\r\n", memptr))
			// } else {
			// 	this.CPU.SetMemFlags(memptr, mos6502.AF_EXEC_BREAK, true)
			// 	this.PutStr(fmt.Sprintf("Breakpoint set at $%.4X.\r\n", memptr))
			// }

			this.sendCommand("setbp", []string{fmt.Sprintf("PC=$%.4x", memptr)})

			continue
		}

		if reBreakW.MatchString(command) {
			m := reBreakW.FindAllStringSubmatch(command, -1)
			if m[0][1] != "" {
				start, _ := strconv.ParseInt(m[0][1], 16, 64)
				memptr = int(start)
			}

			// if this.CPU.AreMemFlagsSet(memptr, mos6502.AF_WRITE_BREAK) {
			// 	// clear
			// 	this.CPU.SetMemFlags(memptr, mos6502.AF_WRITE_BREAK, false)
			// 	this.PutStr(fmt.Sprintf("WRITE Breakpoint removed at $%.4X.\r\n", memptr))
			// } else {
			// 	this.CPU.SetMemFlags(memptr, mos6502.AF_WRITE_BREAK, true)
			// 	this.PutStr(fmt.Sprintf("WRITE Breakpoint set at $%.4X.\r\n", memptr))
			// }

			this.sendCommand("setbp", []string{fmt.Sprintf("WA=$%.4x", memptr)})
			continue
		}

		if reBreakR.MatchString(command) {
			m := reBreakR.FindAllStringSubmatch(command, -1)
			if m[0][1] != "" {
				start, _ := strconv.ParseInt(m[0][1], 16, 64)
				memptr = int(start)
			}

			// if this.CPU.AreMemFlagsSet(memptr, mos6502.AF_READ_BREAK) {
			// 	// clear
			// 	this.CPU.SetMemFlags(memptr, mos6502.AF_READ_BREAK, false)
			// 	this.PutStr(fmt.Sprintf("READ Breakpoint removed at $%.4X.\r\n", memptr))
			// } else {
			// 	this.CPU.SetMemFlags(memptr, mos6502.AF_READ_BREAK, true)
			// 	this.PutStr(fmt.Sprintf("READ Breakpoint set at $%.4X.\r\n", memptr))
			// }
			this.sendCommand("setbp", []string{fmt.Sprintf("RA=$%.4x", memptr)})
			continue
		}

		if reBreakZ.MatchString(command) {
			m := reBreakZ.FindAllStringSubmatch(command, -1)
			if m[0][1] != "" {
				start, _ := strconv.ParseInt(m[0][1], 16, 64)
				memptr = int(start)
			}

			if this.CPU.AreMemFlagsSet(memptr, mos6502.AF_WRITE_LOCK) {
				// clear
				this.CPU.SetMemFlags(memptr, mos6502.AF_WRITE_LOCK, false)
				this.PutStr(fmt.Sprintf("WRITE LOCK removed at $%.4X.\r\n", memptr))
			} else {
				this.CPU.SetMemFlags(memptr, mos6502.AF_WRITE_LOCK, true)
				this.PutStr(fmt.Sprintf("WRITE LOCK set at $%.4X.\r\n", memptr))
			}
			continue
		}

		if reVerify.MatchString(command) {
			m := reVerify.FindAllStringSubmatch(command, -1)

			target, _ := strconv.ParseInt(m[0][1], 16, 64)
			start, _ := strconv.ParseInt(m[0][2], 16, 64)
			end, _ := strconv.ParseInt(m[0][3], 16, 64)

			size := int(end - start + 1)

			ok := true

			for i := 0; i < size; i++ {
				if this.Int.GetMemory(int(target)+i) != this.Int.GetMemory(int(start)+i) {
					ok = false
					memptr = int(target) + i
					break
				}
			}

			if ok {
				this.PutStr("OK.")
			} else {
				this.PutStr(fmt.Sprintf("Verify failed at $%.4x\r\n", memptr))
			}
			continue
		}

		if command == "ST" {
			// dump switches
			this.PutStr("SoftSwitches (Memory):\r\n")
			this.PutStr(this.Int.GetStatusText() + "\r\n")
			this.PutStr("SoftSwitches (Video):\r\n")
			this.PutStr(this.Int.GetVideoStatusText() + "\r\n")
			continue
		}

		if reMove.MatchString(command) {
			m := reMove.FindAllStringSubmatch(command, -1)

			target, _ := strconv.ParseInt(m[0][1], 16, 64)
			start, _ := strconv.ParseInt(m[0][2], 16, 64)
			end, _ := strconv.ParseInt(m[0][3], 16, 64)

			size := int(end - start + 1)
			tmp := make([]uint64, size)
			for i := 0; i < size; i++ {
				tmp[i] = this.Int.GetMemory(int(start) + i)
			}
			for i, v := range tmp {
				this.Int.SetMemory(int(target)+i, v)
			}

			this.PutStr(fmt.Sprintf("Moved %d bytes from $%.4x to $%.4x.\r\n", size, start, target))
			continue
		}

		if reExec.MatchString(command) {

			memptr = this.CPU.PC

			m := reExec.FindAllStringSubmatch(command, -1)
			if m[0][1] != "" {
				start, _ := strconv.ParseInt(m[0][1], 16, 64)
				memptr = int(start)
			}

			//rom := DoCall(memptr, this.Int, false)
			rom := false
			if !settings.PureBoot(this.CPU.MemIndex) {
				rom = DoCall(memptr, this.Int, false)
			}

			if !rom {
				this.CPU.PC = memptr
				this.CPU.Halted = false
				this.ScreenOff()
				this.sendCommand("continue", nil)
				return
			}

			continue
		}

		if reStep.MatchString(command) {
			m := reStep.FindAllStringSubmatch(command, -1)
			if m[0][1] != "" {
				start, _ := strconv.ParseInt(m[0][1], 16, 64)
				memptr = int(start)
			}
			desc, _ := this.CPU.Decode(memptr)
			this.PutStr(desc)
			//this.CPU.PC = memptr
			this.CPU.Halted = false
			this.ScreenOff()
			this.CPU.FetchExecute()
			this.ScreenOn(0)
			this.PutStr(fmt.Sprintf(
				"  A=%.2x X=%.2x Y=%.2x PC=%.4x P=%.2x\r\n",
				this.CPU.A,
				this.CPU.X,
				this.CPU.Y,
				this.CPU.PC,
				this.CPU.P,
			))
			memptr = this.CPU.PC
			continue
		}

		if reTrace.MatchString(command) {
			m := reTrace.FindAllStringSubmatch(command, -1)

			memptr = this.CPU.PC

			if m[0][1] != "" {
				start, _ := strconv.ParseInt(m[0][1], 16, 64)
				//memptr = int(start)
				//this.CPU.ExecBreakpoint[int(start)] = true
				this.CPU.SetMemFlags(int(start), mos6502.AF_EXEC_BREAK, true)
			}

			this.CPU.PC = memptr
			this.CPU.Halted = false

			tracenum++

			filename := fmt.Sprintf("trace_6502_%d_%.4x.txt", tracenum, this.CPU.PC)
			this.PutStr("Trace started to file " + filename + "\r\n")

			f, _ := os.Create(filename)

			var r cpu.FEResponse
			for !this.CPU.Halted && r == cpu.FE_OK {

				txtd, _ := this.CPU.DecodeTrace(this.CPU.PC)
				//fmt.Println(txtd)
				f.WriteString(txtd + "\r\n")

				//desc, _ := this.CPU.Decode(this.CPU.PC)
				//~ this.PutStr(desc)
				r = this.CPU.FetchExecute()

				if r == cpu.FE_BREAKPOINT || r == cpu.FE_ILLEGALOPCODE || r == cpu.FE_BREAKINTERRUPT {
					event := "Breakpoint"
					switch r {
					case cpu.FE_BREAKINTERRUPT:
						event = "BRK"
					case cpu.FE_ILLEGALOPCODE:
						event = fmt.Sprintf("Illegal opcode ($%.2X)", this.Int.GetMemory(this.CPU.PC))
					}

					this.PutStr(fmt.Sprintf("%s reached at $%.4X.\r\n", event, this.CPU.PC))
				}
			}
			f.Close()
			memptr = this.CPU.PC
			continue
		}

		if reDisass.MatchString(command) {
			m := reDisass.FindAllStringSubmatch(command, -1)
			if m[0][1] != "" {
				start, _ := strconv.ParseInt(m[0][1], 16, 64)
				memptr = int(start)
			}
			for i := 0; i < 20; i++ {
				desc, size := this.CPU.Decode(memptr)
				this.PutStr(desc)
				memptr += size
			}
			continue
		}

		if reDisplay.MatchString(command) {
			m := reDisplay.FindAllStringSubmatch(command, -1)
			start := m[0][1]
			end := m[0][3]
			if end == "" {
				end = start
			}

			s, _ := strconv.ParseInt(start, 16, 64)
			e, _ := strconv.ParseInt(end, 16, 64)

			count := int(e) - int(s) + 1
			if count == 0 {
				count = 1
			}
			for i := 0; i < count; i++ {
				if i%8 == 0 {
					if i != 0 {
						this.PutStr("\r\n")
					}
					a := int(s) + i
					this.PutStr(strings.ToUpper(fmt.Sprintf("%.4x-", a)))
				}
				v := this.Int.GetMemory(int(s) + i)
				this.PutStr(strings.ToUpper(fmt.Sprintf(" %.2x", v%256)))
			}
			this.PutStr("\r\n")
			continue
		}

		if strings.Trim(command, " ") != "" {
			parts := strings.Split(command, " ")
			this.sendCommand(strings.ToLower(parts[0]), parts[1:])
		}

	}

}
