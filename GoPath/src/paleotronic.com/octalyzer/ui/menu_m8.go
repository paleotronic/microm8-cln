//go:build !nox
// +build !nox

package ui

import (
	"log"
	"math"
	"os"
	"paleotronic.com/core/hardware/common"
	"runtime"
	"strings"
	"time"

	"paleotronic.com/core/hardware"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/control"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/freeze"
	"paleotronic.com/octalyzer/bus"
	"paleotronic.com/octalyzer/clientperipherals"
	"paleotronic.com/octalyzer/video/font"
	"paleotronic.com/presentation"
	"paleotronic.com/utils"
)

func cleanupWAVE(in []float32, normalize bool, scale float32, smooth bool, smoothpc float32, sampleRate int) []float32 {
	out := make([]float32, len(in))
	var min, max, ctr float32
	for i, v := range in {
		if min > v {
			min = v
		}
		if max < v {
			max = v
		}
		out[i] = v
	}

	// center wave if needed
	ctr = (min + max) / 2
	if ctr != 0 {
		for i, v := range out {
			out[i] = v - ctr
		}
	}

	// normallize
	if normalize {
		var max float32
		for _, v := range out {
			if math.Abs(float64(v)) > float64(max) {
				max = float32(math.Abs(float64(v)))
			}
		}
		for i, v := range out {
			out[i] = (v / max) * scale
		}
	}

	// scale
	for i, v := range out {
		out[i] = v * scale
	}

	// smooth
	if smooth {
		var minSmooth = int(float32(sampleRate) * smoothpc)
		var sameCount int
		var lastV float32 = -99999476 // junk value
		for i, v := range out {
			if v == lastV {
				sameCount++
			} else {
				if sameCount >= minSmooth {
					end := i - 1 // last value
					start := end - sameCount + 1
					for j := start; j <= end; j++ {
						out[j] = 0 // smooth value
					}
				}
				sameCount = 0
			}
			lastV = v
		}
		if sameCount >= minSmooth {
			end := len(out) - 1 // last value
			start := end - sameCount + 1
			for j := start; j <= end; j++ {
				out[j] = 0 // smooth value
			}
		}
	}

	return out
}

func checkExecDebugger() {
	if settings.DebuggerAttachSlot == 0 || !settings.DebuggerOn {
		return
	}
	settings.PureBootCheck(settings.DebuggerAttachSlot - 1)
	if settings.PureBoot(settings.DebuggerAttachSlot - 1) {
		utils.OpenURL(fmt.Sprintf("http://localhost:%d/?attach=%d", settings.DebuggerPort, settings.DebuggerAttachSlot))
	}
}

const (
	SymbolCTRL         = string(rune(259))
	SymbolSHIFT        = string(rune(260))
	SymbolOption       = string(rune(261))
	SymbolAlt          = string(rune(262))
	SymbolENTER        = string(rune(1120))
	SymbolOff          = string(rune(256))
	SymbolOn           = string(rune(257))
	SymbolBackspace    = string(rune(263))
	SymbolSliderHandle = string(rune(1154))
	SymbolSliderMark   = string(rune(1105))
)

func alt() string {
	if runtime.GOOS == "darwin" {
		return SymbolOption
	}
	return SymbolAlt
}

func TestMenu(e interfaces.Interpretable) {

	settings.TemporaryMute = true
	settings.VideoSuspended = false
	defer func() {
		settings.TemporaryMute = false
	}()

	settingsFiles := files.GetSettingsFiles()

	cfg := NewDefaultSettings(e)

	profiles := GetMachineList()

	serialports, _ := common.EnumerateSerialPorts()

	m := NewMenu(nil).Title("microM8 Menu").
		AddMenu("profiles", "Profile").Title("Choose Profile").
		Before(func(m *Menu) {
			m.Items = make([]*MenuItem, 0)
			for i, prf := range profiles {
				m.AddCheck(prf.Filename, prf.Name, func(i int) {
					// Set profile and restart
					settings.SpecFile[e.GetMemIndex()] = profiles[i].Filename
					settings.ForcePureBoot[e.GetMemIndex()] = false
					if settings.MicroPakPath != "" {
						settings.MicroPakPath = ""
						settings.PureBootVolume[e.GetMemIndex()] = ""
						settings.PureBootVolume2[e.GetMemIndex()] = ""
						settings.PureBootSmartVolume[e.GetMemIndex()] = ""
					}
					settings.DiskIIUse13Sectors[e.GetMemIndex()] = false
					if ms, err := hardware.LoadSpec(profiles[i].Filename); err == nil {
						settings.ForcePureBoot[e.GetMemIndex()] = ms.AllowDisklessBoot
					}
					e.GetMemoryMap().IntSetSlotRestart(e.GetMemIndex(), true)
				})
				m.Items[i].Checked = settings.SpecFile[e.GetMemIndex()] == prf.Filename
			}
		}).
		Parent().
		Before(
			func(m *Menu) {
				for _, item := range m.Items {
					if item.Label == "menudisable" {
						item.Checked = !settings.ShowHamburger
					}
					if item.Label == "menuhover" {
						item.Checked = settings.HamburgerOnHover
					}
					if item.Label == "contrast" {
						item.Checked = settings.HighContrastUI
					}
					if item.Label == "reset" {
						item.Hint = SymbolCTRL + SymbolSHIFT + SymbolBackspace
					}
					if item.Label == "help" {
						item.Hint = SymbolCTRL + SymbolSHIFT + "H"
					}
					if item.Label == "cat" {
						item.Hint = SymbolCTRL + SymbolSHIFT + "~"
					}
				}
			},
		).
		// disks
		AddMenu("int", "Interpreter").
		Add("fp", "Floating Point Basic", func(i int) {
			settings.SpecFile[e.GetMemIndex()] = "apple2e-en.yaml"
			settings.VMLaunch[e.GetMemIndex()] = &settings.VMLauncherConfig{}
			apple2helpers.SwitchToDialect(e, "fp")
		}).
		Add("int", "Integer Basic", func(i int) {
			settings.SpecFile[e.GetMemIndex()] = "apple2e-en.yaml"
			settings.VMLaunch[e.GetMemIndex()] = &settings.VMLauncherConfig{}
			apple2helpers.SwitchToDialect(e, "int")
		}).
		Add("logo", "Logo", func(i int) {
			settings.SpecFile[e.GetMemIndex()] = "apple2e-en.yaml"
			settings.VMLaunch[e.GetMemIndex()] = &settings.VMLauncherConfig{}
			apple2helpers.SwitchToDialect(e, "logo")
		}).
		Parent().
		AddMenu("vcr", "VCR").
		Before(
			func(m *Menu) {
				for _, item := range m.Items {
					if item.Label == "rwen" {
						item.Checked = e.IsRecordingVideo()
						item.Hint = SymbolCTRL + SymbolSHIFT + "R,R"
					}
					if item.Label == "back5" {
						item.Hint = SymbolSHIFT + SymbolBackspace
					}
					if item.Label == "liverec" {
						item.Checked = settings.AutoLiveRecording()
					}
					if item.Label == "fracrew" {
						item.Checked = settings.DisableFractionalRewindSpeeds
					}
					if item.Label == "fullcpu" {
						item.Checked = settings.FileFullCPURecord
					}
					if item.Label == "noaudio" {
						item.Checked = !settings.RecordIgnoreAudio[e.GetMemIndex()]
					}
					if item.Label == "lstart" {
						item.Hint = SymbolCTRL + SymbolSHIFT + "R,F"
					}
					if item.Label == "resume" {
						item.Hint = "Space"
					}
					if item.Label == "lrewind" {
						item.Hint = SymbolCTRL + SymbolSHIFT + "["
					}
					if item.Label == "lplay" {
						item.Hint = SymbolCTRL + SymbolSHIFT + "]"
					}
					if item.Label == "lstop" {
						item.Hint = SymbolCTRL + SymbolSHIFT + "R,S"
					}
				}
			},
		).
		Add("lstart", "Start File Recording", func(i int) {
			e.StopRecording()
			e.RecordToggle(settings.FileFullCPURecord)
		}).
		Add("lstop", "Stop Recording", func(i int) { e.StopRecording() }).
		AddCheck("rwen", "Rewind Enable", func(i int) {
			if !e.IsRecordingVideo() {
				e.StartRecording("", false)
			}
		}).
		AddSep().
		Add("back5", "Jump back 5 seconds", func(i int) { go e.BackstepVideo(5000) }).
		Add("lrewind", "Rewind", func(i int) { e.BackVideo() }).
		Add("lplay", "Play", func(i int) { e.ForwardVideo() }).
		Add("resume", "Resume", func(i int) { e.BreakIntoVideo() }).
		AddSep().
		AddCheck("liverec", "Auto Enable Rewind", func(i int) {
			settings.SetAutoLiveRecording(!settings.AutoLiveRecording())
			v := 0
			if settings.AutoLiveRecording() {
				v = 1
			}
			cfg.SetI("hardware", "init.liverecording", v)
		}).
		AddCheck("fracrew", "Disable Slow Rewind", func(i int) {
			settings.DisableFractionalRewindSpeeds = !settings.DisableFractionalRewindSpeeds
			v := 0
			if settings.DisableFractionalRewindSpeeds {
				v = 1
			}
			cfg.SetI("hardware", "init.disablefractionalrewind", v)
		}).
		AddCheck("fullcpu", "All CPU states (file)", func(i int) {
			settings.FileFullCPURecord = !settings.FileFullCPURecord
		}).
		AddCheck("noaudio", "Record audio", func(i int) {
			settings.RecordIgnoreAudio[e.GetMemIndex()] = !settings.RecordIgnoreAudio[e.GetMemIndex()]
			if settings.RecordIgnoreAudio[e.GetMemIndex()] {
				cfg.SetI("audio", "init.recording.noaudio", 1)
			} else {
				cfg.SetI("audio", "init.recording.noaudio", 0)
			}
		}).
		Parent().
		AddMenu("freezer", "Freeze").
		AddMenu("load", "Load States").
		Before(
			func(m *Menu) {
				saves := files.GetDiskSaveFiles(e.GetMemIndex())
				for i, item := range m.Items {
					item.Text = fmt.Sprintf("%d: ", i+1) + files.GetFilename(saves[i])
					item.Hint = fmt.Sprintf(SymbolCTRL+SymbolSHIFT+"L,%d", i+1)
				}
			},
		).
		Add("0", "", func(i int) {
			idx := 0
			saves := files.GetDiskSaveFiles(e.GetMemIndex())
			if saves[idx] != "" {
				f := freeze.NewEmptyState(e)
				_ = f.LoadFromFile(saves[idx])
				f.Apply(e)
			}
		}).
		Add("1", "", func(i int) {
			idx := 1
			saves := files.GetDiskSaveFiles(e.GetMemIndex())
			if saves[idx] != "" {
				f := freeze.NewEmptyState(e)
				_ = f.LoadFromFile(saves[idx])
				f.Apply(e)
			}
		}).
		Add("2", "", func(i int) {
			idx := 2
			saves := files.GetDiskSaveFiles(e.GetMemIndex())
			if saves[idx] != "" {
				f := freeze.NewEmptyState(e)
				err := f.LoadFromFile(saves[idx])
				f.Apply(e)
				fmt.Println(err)
			}
		}).
		Add("3", "", func(i int) {
			idx := 3
			saves := files.GetDiskSaveFiles(e.GetMemIndex())
			if saves[idx] != "" {
				f := freeze.NewEmptyState(e)
				_ = f.LoadFromFile(saves[idx])
				f.Apply(e)
			}
		}).
		Add("4", "", func(i int) {
			idx := 4
			saves := files.GetDiskSaveFiles(e.GetMemIndex())
			if saves[idx] != "" {
				f := freeze.NewEmptyState(e)
				_ = f.LoadFromFile(saves[idx])
				f.Apply(e)
			}
		}).
		Add("5", "", func(i int) {
			idx := 5
			saves := files.GetDiskSaveFiles(e.GetMemIndex())
			if saves[idx] != "" {
				f := freeze.NewEmptyState(e)
				_ = f.LoadFromFile(saves[idx])
				f.Apply(e)
			}
		}).
		Add("6", "", func(i int) {
			idx := 6
			saves := files.GetDiskSaveFiles(e.GetMemIndex())
			if saves[idx] != "" {
				f := freeze.NewEmptyState(e)
				_ = f.LoadFromFile(saves[idx])
				f.Apply(e)
			}
		}).
		Add("7", "", func(i int) {
			idx := 7
			saves := files.GetDiskSaveFiles(e.GetMemIndex())
			if saves[idx] != "" {
				f := freeze.NewEmptyState(e)
				_ = f.LoadFromFile(saves[idx])
				f.Apply(e)
			}
		}).
		Parent().
		AddMenu("save", "Save States").
		Before(
			func(m *Menu) {
				saves := files.GetDiskSaveFiles(e.GetMemIndex())
				for i, item := range m.Items {
					item.Text = fmt.Sprintf("%d: ", i+1) + files.GetFilename(saves[i])
					item.Hint = fmt.Sprintf(SymbolCTRL+SymbolSHIFT+"S,%d", i+1)
				}
			},
		).
		Add("0", "", func(i int) {
			idx := 0
			f := freeze.NewFreezeState(e, false)
			p := files.GetDiskSaveDirectory(e.GetMemIndex()) + fmt.Sprintf("/microM8%d.frz", idx)
			err := f.SaveToFile(p)
			fmt.Println(err)
		}).
		Add("1", "", func(i int) {
			idx := 1
			f := freeze.NewFreezeState(e, false)
			p := files.GetDiskSaveDirectory(e.GetMemIndex()) + fmt.Sprintf("/microM8%d.frz", idx)
			f.SaveToFile(p)
		}).
		Add("2", "", func(i int) {
			idx := 2
			f := freeze.NewFreezeState(e, false)
			p := files.GetDiskSaveDirectory(e.GetMemIndex()) + fmt.Sprintf("/microM8%d.frz", idx)
			f.SaveToFile(p)
		}).
		Add("3", "", func(i int) {
			idx := 3
			f := freeze.NewFreezeState(e, false)
			p := files.GetDiskSaveDirectory(e.GetMemIndex()) + fmt.Sprintf("/microM8%d.frz", idx)
			f.SaveToFile(p)
		}).
		Add("4", "", func(i int) {
			idx := 4
			f := freeze.NewFreezeState(e, false)
			p := files.GetDiskSaveDirectory(e.GetMemIndex()) + fmt.Sprintf("/microM8%d.frz", idx)
			f.SaveToFile(p)
		}).
		Add("5", "", func(i int) {
			idx := 5
			f := freeze.NewFreezeState(e, false)
			p := files.GetDiskSaveDirectory(e.GetMemIndex()) + fmt.Sprintf("/microM8%d.frz", idx)
			f.SaveToFile(p)
		}).
		Add("6", "", func(i int) {
			idx := 6
			f := freeze.NewFreezeState(e, false)
			p := files.GetDiskSaveDirectory(e.GetMemIndex()) + fmt.Sprintf("/microM8%d.frz", idx)
			f.SaveToFile(p)
		}).
		Add("7", "", func(i int) {
			idx := 7
			f := freeze.NewFreezeState(e, false)
			p := files.GetDiskSaveDirectory(e.GetMemIndex()) + fmt.Sprintf("/microM8%d.frz", idx)
			f.SaveToFile(p)
		}).
		Parent().
		Parent().
		AddMenu("hardware", "Hardware").
		AddMenu("serial", "Super Serial Card").
		Before(
			func(m *Menu) {
				for _, item := range m.Items {
					if item.Label == "modem" {
						item.Checked = settings.SSCCardMode[e.GetMemIndex()] == settings.SSCModeVirtualModem
					}
					if item.Label == "server" {
						item.Checked = settings.SSCCardMode[e.GetMemIndex()] == settings.SSCModeTelnetServer
					}
					if item.Label == "imgwrt" {
						item.Checked = settings.SSCCardMode[e.GetMemIndex()] == settings.SSCModeEmulatedImageWriter
					}
					if item.Label == "escp" {
						item.Checked = settings.SSCCardMode[e.GetMemIndex()] == settings.SSCModeEmulatedESCP
					}
					if item.Label == "rawserial" {
						item.Checked = settings.SSCCardMode[e.GetMemIndex()] == settings.SSCModeSerialRaw
					}
					if item.Label == "serialports" {
						item.Hidden = len(serialports) == 0
					}
				}
			},
		).
		AddCheck("modem", "Virtual Modem", func(i int) {
			settings.SSCCardMode[e.GetMemIndex()] = settings.SSCModeVirtualModem
			cfg.SetI("hardware", "init.serial.mode", int(settings.SSCCardMode[e.GetMemIndex()]))
		}).
		AddCheckIf("server", "Telnet Server (ADTPro etc)", !settings.IsSetBoolOverride(e.GetMemIndex(), "ssc.disable.telnetserver"), func(i int) {
			settings.SSCCardMode[e.GetMemIndex()] = settings.SSCModeTelnetServer
			cfg.SetI("hardware", "init.serial.mode", int(settings.SSCCardMode[e.GetMemIndex()]))
		}).
		AddCheck("imgwrt", "Emulated ImageWriter II", func(i int) {
			settings.SSCCardMode[e.GetMemIndex()] = settings.SSCModeEmulatedImageWriter
			cfg.SetI("hardware", "init.serial.mode", int(settings.SSCCardMode[e.GetMemIndex()]))
		}).
		AddCheck("escp", "Emulated Epson Printer", func(i int) {
			settings.SSCCardMode[e.GetMemIndex()] = settings.SSCModeEmulatedESCP
			cfg.SetI("hardware", "init.serial.mode", int(settings.SSCCardMode[e.GetMemIndex()]))
		}).
		AddCheckIf("rawserial", "Raw Serial Port", len(serialports) > 0, func(i int) {
			settings.SSCCardMode[e.GetMemIndex()] = settings.SSCModeSerialRaw
			cfg.SetI("hardware", "init.serial.mode", int(settings.SSCCardMode[e.GetMemIndex()]))
		}).
		AddMenu("serialports", "Serial Ports").
		Before(
			func(m *Menu) {
				m.Items = []*MenuItem{}
				serialports, _ = common.EnumerateSerialPorts()
				for i, port := range serialports {
					m.AddCheck(port, port, func(i int) {
						settings.SSCHardwarePort = port
						settings.SSCCardMode[e.GetMemIndex()] = settings.SSCModeSerialRaw // auto set raw mode
					})
					m.Items[i].Checked = settings.SSCHardwarePort == port
				}
			},
		).
		Parent().
		AddMenu("sscsw1", "SW1").
		Before(func(m *Menu) {
			for i := 0; i < 8; i++ {
				sw := m.Items[i]
				bitmask := 1 << (7 - i)
				sw.Checked = settings.SSCDipSwitch1&bitmask == 0
			}
		}).
		AddCheck("sw1-1", "SW1-1", func(i int) {
			settings.SSCDipSwitch1 = settings.SSCDipSwitch1 ^ 0x80
			cfg.SetI("hardware", "init.serial.dipsw1", int(settings.SSCDipSwitch1))
		}).
		AddCheck("sw1-2", "SW1-2", func(i int) {
			settings.SSCDipSwitch1 = settings.SSCDipSwitch1 ^ 0x40
			cfg.SetI("hardware", "init.serial.dipsw1", int(settings.SSCDipSwitch1))
		}).
		AddCheck("sw1-3", "SW1-3", func(i int) {
			settings.SSCDipSwitch1 = settings.SSCDipSwitch1 ^ 0x20
			cfg.SetI("hardware", "init.serial.dipsw1", int(settings.SSCDipSwitch1))
		}).
		AddCheck("sw1-4", "SW1-4", func(i int) {
			settings.SSCDipSwitch1 = settings.SSCDipSwitch1 ^ 0x10
			cfg.SetI("hardware", "init.serial.dipsw1", int(settings.SSCDipSwitch1))
		}).
		AddCheck("sw1-5", "SW1-5", func(i int) {
			settings.SSCDipSwitch1 = settings.SSCDipSwitch1 ^ 0x08
			cfg.SetI("hardware", "init.serial.dipsw1", int(settings.SSCDipSwitch1))
		}).
		AddCheck("sw1-6", "SW1-6", func(i int) {
			settings.SSCDipSwitch1 = settings.SSCDipSwitch1 ^ 0x04
			cfg.SetI("hardware", "init.serial.dipsw1", int(settings.SSCDipSwitch1))
		}).
		AddCheck("sw1-7", "SW1-7", func(i int) {
			settings.SSCDipSwitch1 = settings.SSCDipSwitch1 ^ 0x02
			cfg.SetI("hardware", "init.serial.dipsw1", int(settings.SSCDipSwitch1))
		}).
		AddCheck("sw1-8", "SW1-8", func(i int) {
			settings.SSCDipSwitch1 = settings.SSCDipSwitch1 ^ 0x01
			cfg.SetI("hardware", "init.serial.dipsw1", int(settings.SSCDipSwitch1))
		}).
		Parent().
		AddMenu("sscsw2", "SW2").
		Before(func(m *Menu) {
			for i := 0; i < 8; i++ {
				sw := m.Items[i]
				bitmask := 1 << (7 - i)
				sw.Checked = settings.SSCDipSwitch2&bitmask == 0
			}
		}).
		AddCheck("sw2-1", "SW2-1", func(i int) {
			settings.SSCDipSwitch2 = settings.SSCDipSwitch2 ^ 0x80
			cfg.SetI("hardware", "init.serial.dipsw2", int(settings.SSCDipSwitch2))
		}).
		AddCheck("sw2-2", "SW2-2", func(i int) {
			settings.SSCDipSwitch2 = settings.SSCDipSwitch2 ^ 0x40
			cfg.SetI("hardware", "init.serial.dipsw2", int(settings.SSCDipSwitch2))
		}).
		AddCheck("sw2-3", "SW2-3", func(i int) {
			settings.SSCDipSwitch2 = settings.SSCDipSwitch2 ^ 0x20
			cfg.SetI("hardware", "init.serial.dipsw2", int(settings.SSCDipSwitch2))
		}).
		AddCheck("sw2-4", "SW2-4", func(i int) {
			settings.SSCDipSwitch2 = settings.SSCDipSwitch2 ^ 0x10
			cfg.SetI("hardware", "init.serial.dipsw2", int(settings.SSCDipSwitch2))
		}).
		AddCheck("sw2-5", "SW2-5", func(i int) {
			settings.SSCDipSwitch2 = settings.SSCDipSwitch2 ^ 0x08
			cfg.SetI("hardware", "init.serial.dipsw2", int(settings.SSCDipSwitch2))
		}).
		AddCheck("sw2-6", "SW2-6", func(i int) {
			settings.SSCDipSwitch2 = settings.SSCDipSwitch2 ^ 0x04
			cfg.SetI("hardware", "init.serial.dipsw2", int(settings.SSCDipSwitch2))
		}).
		AddCheck("sw2-7", "SW2-7", func(i int) {
			settings.SSCDipSwitch2 = settings.SSCDipSwitch2 ^ 0x02
			cfg.SetI("hardware", "init.serial.dipsw2", int(settings.SSCDipSwitch2))
		}).
		AddCheck("sw2-8", "SW2-8", func(i int) {
			settings.SSCDipSwitch2 = settings.SSCDipSwitch2 ^ 0x01
			cfg.SetI("hardware", "init.serial.dipsw2", int(settings.SSCDipSwitch2))
		}).
		Parent().
		Parent().
		AddMenu("cpu", "CPU").
		Before(
			func(m *Menu) {
				for _, item := range m.Items {
					if item.Label == "diskwarp" {
						item.Checked = settings.NoDiskWarp[e.GetMemIndex()]
					}
				}
			},
		).
		Add("mon", "Monitor", func(i int) {
			e.GetMemoryMap().IntSetCPUHalt(e.GetMemIndex(), true)
		}).
		AddMenu("type", "Type").
		Before(
			func(m *Menu) {
				cpu := apple2helpers.GetCPU(e)
				for _, item := range m.Items {
					item.Checked = strings.ToLower(item.Label) == strings.ToLower(cpu.Model)
				}
			},
		).
		AddCheck("6502", "6502", func(i int) { apple2helpers.SwitchCPU(e, apple2helpers.New6502(e)) }).
		AddCheck("65C02", "65C02", func(i int) { apple2helpers.SwitchCPU(e, apple2helpers.New65C02(e)) }).
		Parent().
		AddMenu("cpuspeed", "Warp").
		Before(
			func(m *Menu) {
				sstr := fmt.Sprintf("%.2f", apple2helpers.GetCPU(e).GetWarp())
				for i, item := range m.Items {
					item.Checked = item.Label == sstr
					item.IsPercentage = true
					item.Hint = fmt.Sprintf(SymbolCTRL+SymbolSHIFT+"W,%d", i+1)
				}
				m.IsPercentage = true
			},
		).
		AddCheck("0.25", "25%", func(i int) {
			apple2helpers.GetCPU(e).SetWarpUser(0.25)
			apple2helpers.GetZ80CPU(e).SetWarpUser(0.25)
		}).
		AddCheck("0.50", "50%", func(i int) {
			apple2helpers.GetCPU(e).SetWarpUser(0.5)
			apple2helpers.GetZ80CPU(e).SetWarpUser(0.5)
		}).
		AddCheck("1.00", "100%", func(i int) {
			apple2helpers.GetCPU(e).SetWarpUser(1)
			apple2helpers.GetZ80CPU(e).SetWarpUser(1)
		}).
		AddCheck("2.00", "200%", func(i int) {
			apple2helpers.GetCPU(e).SetWarpUser(2)
			apple2helpers.GetZ80CPU(e).SetWarpUser(2)
		}).
		AddCheck("4.00", "400%", func(i int) {
			apple2helpers.GetCPU(e).SetWarpUser(4)
			apple2helpers.GetZ80CPU(e).SetWarpUser(4)
		}).
		AddCheck("8.00", "800%", func(i int) {
			apple2helpers.GetCPU(e).SetWarpUser(8)
			apple2helpers.GetZ80CPU(e).SetWarpUser(8)
		}).
		Parent().
		Parent().
		AddMenu("printer", "Printer").
		AddMenu("timeout", "PDF Timeout").
		Before(
			func(m *Menu) {
				sstr := fmt.Sprintf("%d", settings.PrintToPDFTimeoutSec)
				for _, item := range m.Items {
					item.Checked = item.Label == sstr
					if item.Label == "osprinter" {
						item.Checked = settings.PDFSpool
					}
				}
			},
		).
		AddCheck("5", "5 seconds", func(i int) {
			settings.PrintToPDFTimeoutSec = 5
			cfg.SetI("hardware", "init.printer.timeout", settings.PrintToPDFTimeoutSec)
		}).
		AddCheck("15", "15 seconds", func(i int) {
			settings.PrintToPDFTimeoutSec = 15
			cfg.SetI("hardware", "init.printer.timeout", settings.PrintToPDFTimeoutSec)
		}).
		AddCheck("30", "30 seconds", func(i int) {
			settings.PrintToPDFTimeoutSec = 30
			cfg.SetI("hardware", "init.printer.timeout", settings.PrintToPDFTimeoutSec)
		}).
		AddCheck("45", "45 seconds", func(i int) {
			settings.PrintToPDFTimeoutSec = 45
			cfg.SetI("hardware", "init.printer.timeout", settings.PrintToPDFTimeoutSec)
		}).
		AddCheck("60", "60 seconds", func(i int) {
			settings.PrintToPDFTimeoutSec = 60
			cfg.SetI("hardware", "init.printer.timeout", settings.PrintToPDFTimeoutSec)
		}).
		AddCheck("86400", "Manual (use Flush PDF)", func(i int) {
			settings.PrintToPDFTimeoutSec = 86400
			cfg.SetI("hardware", "init.printer.timeout", settings.PrintToPDFTimeoutSec)
		}).
		AddSep().
		Add("savepdf", "Flush PDF", func(i int) {
			settings.FlushPDF[e.GetMemIndex()] = true
		}).
		AddCheck("osprinter", "Spool to def OS printer", func(i int) {
			settings.PDFSpool = !settings.PDFSpool
			v := 0
			if settings.PDFSpool {
				v = 1
			}
			cfg.SetI("hardware", "init.printer.autospool", v)
		}).
		Parent().
		Parent().
		// AddMenu("tape", "Tape").
		// Before(
		// 	func(m *Menu) {
		// 		for _, item := range m.Items {
		// 			if item.Label == "record" {
		// 				item.Checked = settings.RecordC020[e.GetMemIndex()]
		// 			}
		// 		}
		// 	},
		// ).
		// Add("attach", "Attach Tape", func(i int) {
		// 	mr, ok := e.GetMemoryMap().InterpreterMappableAtAddress(e.GetMemIndex(), 0xc000)
		// 	if ok {
		// 		//state := e.GetState()
		// 		io := mr.(*apple2.Apple2IOChip)
		// 		log.Printf("io=%v", io)
		// 		a := bus.IsClock()
		// 		//apple2helpers.OSDPanel(e, false)
		// 		go CatalogPresentTapePicker(e, 0)
		// 		//apple2helpers.OSDPanel(e, true)
		// 		if a {
		// 			bus.StartDefault()
		// 		}
		// 		//e.SetState(state)
		// 	}
		// }).
		// Add("eject", "Eject Tape", func(i int) {
		// 	mr, ok := e.GetMemoryMap().InterpreterMappableAtAddress(e.GetMemIndex(), 0xc000)
		// 	if ok {
		// 		io := mr.(*apple2.Apple2IOChip)
		// 		io.TapeDetach()
		// 	}
		// }).
		// AddSep().
		// Add("rewind", "Rewind Tape", func(i int) {
		// 	mr, ok := e.GetMemoryMap().InterpreterMappableAtAddress(e.GetMemIndex(), 0xc000)
		// 	if ok {
		// 		io := mr.(*apple2.Apple2IOChip)
		// 		io.TapeBegin()
		// 	}
		// }).
		// AddSep().
		// AddCheck("record", "Record to Tape", func(i int) {
		// 	index := e.GetMemIndex()
		// 	settings.RecordC020[index] = !settings.RecordC020[index]
		// 	if settings.RecordC020[index] {
		// 		// start recording
		// 		settings.RecordC020Rate[index] = settings.SampleRate
		// 		settings.RecordC020Buffer[index] = []float32{}
		// 	} else {
		// 		if len(settings.RecordC020Buffer[index]) > 0 {

		// 			settings.RecordC020Buffer[index] = cleanupWAVE(
		// 				settings.RecordC020Buffer[index],
		// 				true,
		// 				1.2,
		// 				true,
		// 				0.05,
		// 				settings.RecordC020Rate[index],
		// 			)

		// 			// encode a wave here
		// 			log.Printf("recorded %d samples", len(settings.RecordC020Buffer[index]))
		// 			b := bytes.NewBuffer([]byte{})
		// 			wr := wav.NewWriter(b, uint32(len(settings.RecordC020Buffer[index])), 1, uint32(settings.RecordC020Rate[index]), 16)
		// 			for _, v := range settings.RecordC020Buffer[index] {
		// 				wr.WriteSamples([]wav.Sample{wav.Sample{Values: [2]int{int(16384 * v), int(16384 * v)}}})
		// 			}
		// 			basePath := "/local/MyTapes"
		// 			baseName := fmt.Sprintf("cassette-recording-%d.wav", time.Now().UnixNano())
		// 			_ = files.MkdirViaProvider(basePath)
		// 			_ = files.WriteBytesViaProvider(basePath, baseName, b.Bytes())
		// 		}
		// 	}
		// }).
		// Parent().
		AddMenu("disk", "Disks").
		Before(
			func(m *Menu) {
				sstr := apple2helpers.GetCPU(e).Model
				for _, item := range m.Items {
					item.Checked = item.Label == sstr
					if item.Label == "diskwarp" {
						item.Checked = settings.NoDiskWarp[e.GetMemIndex()]
					}
					if item.Label == "dsk2woz" {
						item.Checked = !settings.PreserveDSK
					}
					if item.Label == "13sec" {
						item.Checked = settings.DiskIIUse13Sectors[e.GetMemIndex()]
					}
				}
			},
		).
		AddCheck("13sec", "Use 13 Sectors", func(i int) {
			settings.DiskIIUse13Sectors[e.GetMemIndex()] = !settings.DiskIIUse13Sectors[e.GetMemIndex()]
			e.GetMemoryMap().IntSetSlotRestart(e.GetMemIndex(), true)
		}).
		AddCheck("dsk2woz", "Convert DSK->WOZ", func(i int) {
			settings.PreserveDSK = !settings.PreserveDSK
			v := 0
			if settings.PreserveDSK {
				v = 1
			}
			cfg.SetI("hardware", "init.apple2.disk.nodskwoz", v)
		}).
		AddCheck("diskwarp", "Disable warp", func(i int) {
			settings.NoDiskWarp[e.GetMemIndex()] = !settings.NoDiskWarp[e.GetMemIndex()]
			v := 0
			if settings.NoDiskWarp[e.GetMemIndex()] {
				v = 1
			}
			cfg.SetI("hardware", "init.apple2.disk.nowarp", v)
		}).
		Add("swap", "Swap disks", func(i int) {
			servicebus.SendServiceBusMessage(e.GetMemIndex(), servicebus.DiskIIExchangeDisks, "Disk swap")
		}).
		AddMenu("disk0", "Drive 1").
		Before(
			func(m *Menu) {
				resp, _ := servicebus.SendServiceBusMessage(
					e.GetMemIndex(),
					servicebus.DiskIIQueryWriteProtect,
					0,
				)
				m.Items[1].Checked = resp[0].Payload.(bool)
			},
		).
		Add("eject", "Eject", func(i int) {
			//hardware.DiskInsert(e, 0, "", false)
			servicebus.SendServiceBusMessage(e.GetMemIndex(), servicebus.DiskIIEject, 0)
		}).
		AddCheck("wp", "Write Protect", func(i int) {
			servicebus.SendServiceBusMessage(
				e.GetMemIndex(),
				servicebus.DiskIIToggleWriteProtect,
				0,
			)
		}).
		Add("iblank", "Insert blank", func(i int) {
			now := time.Now()
			servicebus.SendServiceBusMessage(e.GetMemIndex(), servicebus.DiskIIInsertBlank, servicebus.DiskTargetString{
				Drive: 0,
				Filename: fmt.Sprintf(
					"/local/MyDisks/blank_%.4d%.2d%.2d_%.2d%.2d%.2d.woz",
					now.Year(),
					now.Month(),
					now.Day(),
					now.Hour(),
					now.Minute(),
					now.Second(),
				),
			})
		}).
		Add("insert", "Insert disk", func(i int) {
			a := bus.IsClock()
			apple2helpers.OSDPanel(e, false)
			CatalogPresentDiskPicker(e, 0)
			apple2helpers.OSDPanel(e, true)
			if a {
				bus.StartDefault()
			}
		}).
		Parent().
		AddMenu("disk1", "Drive 2").
		Before(
			func(m *Menu) {
				resp, _ := servicebus.SendServiceBusMessage(
					e.GetMemIndex(),
					servicebus.DiskIIQueryWriteProtect,
					1,
				)
				m.Items[1].Checked = resp[0].Payload.(bool)
			},
		).
		Add("eject", "Eject", func(i int) {
			//hardware.DiskInsert(e, 1, "", false)
			servicebus.SendServiceBusMessage(e.GetMemIndex(), servicebus.DiskIIEject, 1)
		}).
		AddCheck("wp", "Write Protect", func(i int) {
			// d := hardware.GetDisk(e, 1)
			// if d != nil {
			// 	d.INFO.SetWriteProtected(!d.INFO.WriteProtected())
			// }
			servicebus.SendServiceBusMessage(
				e.GetMemIndex(),
				servicebus.DiskIIToggleWriteProtect,
				1,
			)
		}).
		Add("iblank", "Insert blank", func(i int) {
			now := time.Now()
			servicebus.SendServiceBusMessage(e.GetMemIndex(), servicebus.DiskIIInsertBlank, servicebus.DiskTargetString{
				Drive: 1,
				Filename: fmt.Sprintf(
					"/local/MyDisks/blank_%.4d%.2d%.2d_%.2d%.2d%.2d.woz",
					now.Year(),
					now.Month(),
					now.Day(),
					now.Hour(),
					now.Minute(),
					now.Second(),
				),
			})
		}).
		Add("insert", "Insert disk", func(i int) {
			a := bus.IsClock()
			apple2helpers.OSDPanel(e, false)
			CatalogPresentDiskPicker(e, 1)
			apple2helpers.OSDPanel(e, true)
			if a {
				bus.StartDefault()
			}
		}).
		Parent().
		Parent().
		Parent().
		// video
		AddMenu("video", "Video").
		Before(func(m *Menu) {
			for _, item := range m.Items {
				switch item.Label {
				case "ss":
					item.Hint = SymbolCTRL + SymbolSHIFT + "\\"
				case "fs":
					item.Checked = (!settings.Windowed)
					item.Hint = alt() + SymbolENTER
				case "aspect", "voxel", "tintmenu", "rendermenu", "rendermenugr", "font":
					item.Hidden = settings.UnifiedRender[e.GetMemIndex()]
				case "demomode":
					isSpectrum := strings.Contains(settings.SpecFile[e.GetMemIndex()], "spectrum")
					item.Checked = settings.UnifiedRender[e.GetMemIndex()]
					item.Hidden = isSpectrum
				case "vertBlendU":
					item.Hidden = !settings.UnifiedRender[e.GetMemIndex()]
					item.Checked = settings.UseVerticalBlend[e.GetMemIndex()]
				}
			}
		}).
		Add("ss", "Take Screenshot", func(i int) {
			time.AfterFunc(time.Second, func() {
				settings.TakeScreenshot = true
			})
		}).
		AddCheck("fs", "Fullscreen", func(i int) { settings.Windowed = !settings.Windowed }).
		AddSep().
		AddCheck("demomode", "Cycle Accurate Render", func(i int) {
			go bus.SyncDo(func() {
				settings.UnifiedRender[e.GetMemIndex()] = !settings.UnifiedRender[e.GetMemIndex()]
				v := 0
				if settings.UnifiedRender[e.GetMemIndex()] {
					v = 1
				}
				cfg.SetI("video", "init.unified", v)
			})
		}).
		AddCheck("vertBlendU", "Vertical Blending", func(i int) {
			settings.UseVerticalBlend[e.GetMemIndex()] = !settings.UseVerticalBlend[e.GetMemIndex()]
			//h1, _ := e.GetGFXLayerByID("HGR1")
			//h1.SetDirty(true)
			//h2, _ := e.GetGFXLayerByID("HGR2")
			//h2.SetDirty(true)
			go bus.SyncDo(func() {
				if settings.UseVerticalBlend[e.GetMemIndex()] {
					cfg.SetI("video", "init.video.vertblend", 1)
				} else {
					cfg.SetI("video", "init.video.vertblend", 0)
				}
			})
		}).
		AddMenu("aspect", "Aspect Ratio").
		Before(
			func(m *Menu) {
				mm := e.GetMemoryMap()
				idx := e.GetMemIndex()
				control := types.NewOrbitController(mm, idx, -1)
				astr := fmt.Sprintf("%.2f", control.GetAspect())
				for i, item := range m.Items {
					item.Checked = astr == item.Label
					item.Hint = fmt.Sprintf(SymbolCTRL+SymbolSHIFT+"A,%d", i+1)
				}
			},
		).
		AddCheck("1.00", "1.00", func(i int) {
			setSlotAspect(e.GetMemoryMap(), e.GetMemIndex(), 1)
			cfg.SetF("video", "init.aspect", 1.00)
		}).
		AddCheck("1.33", "1.33", func(i int) {
			setSlotAspect(e.GetMemoryMap(), e.GetMemIndex(), 1.33)
			cfg.SetF("video", "init.aspect", 1.33)
		}).
		AddCheck("1.46", "1.46", func(i int) {
			setSlotAspect(e.GetMemoryMap(), e.GetMemIndex(), 1.46)
			cfg.SetF("video", "init.aspect", 1.46)
		}).
		AddCheck("1.62", "1.62", func(i int) {
			setSlotAspect(e.GetMemoryMap(), e.GetMemIndex(), 1.62)
			cfg.SetF("video", "init.aspect", 1.62)
		}).
		AddCheck("1.79", "1.78", func(i int) {
			setSlotAspect(e.GetMemoryMap(), e.GetMemIndex(), 1.78889)
			cfg.SetF("video", "init.aspect", 1.78889)
		}).
		Parent().
		AddMenu("voxel", "Voxel Depth").
		Before(
			func(m *Menu) {
				mm := e.GetMemoryMap()
				idx := e.GetMemIndex()
				vd := mm.IntGetVoxelDepth(idx)
				for i, item := range m.Items {
					item.Checked = int(vd) == i
					item.Hint = fmt.Sprintf(SymbolCTRL+SymbolSHIFT+"D,%d", i+1)
				}
			},
		).
		AddCheck("0", "1x depth", func(i int) {
			e.GetMemoryMap().IntSetVoxelDepth(e.GetMemIndex(), 0)
			cfg.SetI("video", "init.video.voxeldepth", 0)
		}).
		AddCheck("1", "2x depth", func(i int) {
			e.GetMemoryMap().IntSetVoxelDepth(e.GetMemIndex(), 1)
			cfg.SetI("video", "init.video.voxeldepth", 1)
		}).
		AddCheck("2", "3x depth", func(i int) {
			e.GetMemoryMap().IntSetVoxelDepth(e.GetMemIndex(), 2)
			cfg.SetI("video", "init.video.voxeldepth", 2)
		}).
		AddCheck("3", "4x depth", func(i int) {
			e.GetMemoryMap().IntSetVoxelDepth(e.GetMemIndex(), 3)
			cfg.SetI("video", "init.video.voxeldepth", 3)
		}).
		AddCheck("4", "5x depth", func(i int) {
			e.GetMemoryMap().IntSetVoxelDepth(e.GetMemIndex(), 4)
			cfg.SetI("video", "init.video.voxeldepth", 4)
		}).
		AddCheck("5", "6x depth", func(i int) {
			e.GetMemoryMap().IntSetVoxelDepth(e.GetMemIndex(), 5)
			cfg.SetI("video", "init.video.voxeldepth", 5)
		}).
		AddCheck("6", "7x depth", func(i int) {
			e.GetMemoryMap().IntSetVoxelDepth(e.GetMemIndex(), 6)
			cfg.SetI("video", "init.video.voxeldepth", 6)
		}).
		AddCheck("7", "8x depth", func(i int) {
			e.GetMemoryMap().IntSetVoxelDepth(e.GetMemIndex(), 7)
			cfg.SetI("video", "init.video.voxeldepth", 7)
		}).
		AddCheck("8", "9x depth", func(i int) {
			e.GetMemoryMap().IntSetVoxelDepth(e.GetMemIndex(), 8)
			cfg.SetI("video", "init.video.voxeldepth", 8)
		}).
		Parent().
		// hgr
		AddMenu("scanlines", "Scanline intensity").
		Before(
			func(m *Menu) {
				// prepare menu
				//m.IsPercentage = true
				for _, item := range m.Items {
					//item.IsPercentage = true
					item.Hint = fmt.Sprintf(SymbolCTRL+SymbolSHIFT+"I,%s", item.Label)
					switch item.Label {
					case "0":
						item.Checked = settings.ScanLineIntensity == 1
					case "1":
						item.Checked = settings.ScanLineIntensity == 0.88
					case "2":
						item.Checked = settings.ScanLineIntensity == 0.77
					case "3":
						item.Checked = settings.ScanLineIntensity == 0.66
					case "4":
						item.Checked = settings.ScanLineIntensity == 0.55
					case "5":
						item.Checked = settings.ScanLineIntensity == 0.44
					case "6":
						item.Checked = settings.ScanLineIntensity == 0.33
					case "7":
						item.Checked = settings.ScanLineIntensity == 0.22
					case "8":
						item.Checked = settings.ScanLineIntensity == 0.11
					case "9":
						item.Checked = settings.ScanLineIntensity == 0
					}
				}
			},
		).
		AddCheck("0", "Off", func(i int) {
			settings.ScanLineIntensity = 1
			cfg.SetF("video", "init.video.scanline", 1)
		}).
		AddCheck("1", "Intensity 1", func(i int) {
			settings.ScanLineIntensity = 0.88
			cfg.SetF("video", "init.video.scanline", 0.88)
		}).
		AddCheck("2", "Intensity 2", func(i int) {
			settings.ScanLineIntensity = 0.77
			cfg.SetF("video", "init.video.scanline", 0.77)
		}).
		AddCheck("3", "Intensity 3", func(i int) {
			settings.ScanLineIntensity = 0.66
			cfg.SetF("video", "init.video.scanline", 0.66)
		}).
		AddCheck("4", "Intensity 4", func(i int) {
			settings.ScanLineIntensity = 0.55
			cfg.SetF("video", "init.video.scanline", 0.55)
		}).
		AddCheck("5", "Intensity 5", func(i int) {
			settings.ScanLineIntensity = 0.44
			cfg.SetF("video", "init.video.scanline", 0.44)
		}).
		AddCheck("6", "Intensity 6", func(i int) {
			settings.ScanLineIntensity = 0.33
			cfg.SetF("video", "init.video.scanline", 0.33)
		}).
		AddCheck("7", "Intensity 7", func(i int) {
			settings.ScanLineIntensity = 0.22
			cfg.SetF("video", "init.video.scanline", 0.22)
		}).
		AddCheck("8", "Intensity 8", func(i int) {
			settings.ScanLineIntensity = 0.11
			cfg.SetF("video", "init.video.scanline", 0.11)
		}).
		AddCheck("9", "Intensity 9", func(i int) {
			settings.ScanLineIntensity = 0
			cfg.SetF("video", "init.video.scanline", 0)
		}).
		Parent().
		// tint mode
		AddMenu("tintmenu", "Tint mode").
		Before(
			func(m *Menu) {
				// prepare menu
				for _, item := range m.Items {
					switch item.Label {
					case "tintNormal":
						item.Checked = e.GetMemoryMap().IntGetVideoTint(e.GetMemIndex()) == settings.VPT_NONE
						item.Hint = SymbolCTRL + SymbolSHIFT + "T,1"
					case "tintGrey":
						item.Checked = e.GetMemoryMap().IntGetVideoTint(e.GetMemIndex()) == settings.VPT_GREY
						item.Hint = SymbolCTRL + SymbolSHIFT + "T,2"
					case "tintGreen":
						item.Checked = e.GetMemoryMap().IntGetVideoTint(e.GetMemIndex()) == settings.VPT_GREEN
						item.Hint = SymbolCTRL + SymbolSHIFT + "T,3"
					case "tintAmber":
						item.Checked = e.GetMemoryMap().IntGetVideoTint(e.GetMemIndex()) == settings.VPT_AMBER
						item.Hint = SymbolCTRL + SymbolSHIFT + "T,4"
					}
				}
			},
		).
		AddCheck("tintNormal", "No tint", func(i int) {
			e.GetMemoryMap().IntSetVideoTint(e.GetMemIndex(), settings.VPT_NONE)
			cfg.SetI("video", "init.video.tintmode", int(settings.VPT_NONE))
		}).
		AddCheck("tintGrey", "Black and white", func(i int) {
			e.GetMemoryMap().IntSetVideoTint(e.GetMemIndex(), settings.VPT_GREY)
			cfg.SetI("video", "init.video.tintmode", int(settings.VPT_GREY))
		}).
		AddCheck("tintGreen", "Phosphor", func(i int) {
			e.GetMemoryMap().IntSetVideoTint(e.GetMemIndex(), settings.VPT_GREEN)
			cfg.SetI("video", "init.video.tintmode", int(settings.VPT_GREEN))
		}).
		AddCheck("tintAmber", "Amber", func(i int) {
			e.GetMemoryMap().IntSetVideoTint(e.GetMemIndex(), settings.VPT_AMBER)
			cfg.SetI("video", "init.video.tintmode", int(settings.VPT_AMBER))
		}).Parent().Parent().
		// Audio menu
		AddMenu("audio", "Audio").Title("Audio").
		Before(
			func(m *Menu) {
				for _, item := range m.Items {
					if item.Label == "mute" {
						item.Checked = clientperipherals.SPEAKER.Mixer.IsMuted()
					}
					if item.Label == "hq" {
						item.Checked = settings.UseHQAudio
					}
				}
			},
		).
		AddCheck("mute", "Mute Audio", func(i int) { clientperipherals.SPEAKER.Mixer.SetMute(!clientperipherals.SPEAKER.Mixer.IsMuted()) }).
		AddCheck("hq", "HQ Audio", func(i int) {
			settings.UseHQAudio = !settings.UseHQAudio
		}).
		// volume
		AddMenu("volume", "Master").Title("Vol").
		Before(
			func(m *Menu) {
				for i, item := range m.Items {
					if item.Label == "mute" {
						item.Checked = clientperipherals.SPEAKER.Mixer.IsMuted()
					} else {
						item.Checked = i*10 == int(settings.MixerVolume*100)
						item.IsPercentage = true
					}
				}
				m.IsPercentage = true
			},
		).
		AddCheck("0", "0%", func(i int) { settings.MixerVolume = 0 }).
		AddCheck("10", "10%", func(i int) { settings.MixerVolume = 0.1 }).
		AddCheck("20", "20%", func(i int) { settings.MixerVolume = 0.2 }).
		AddCheck("30", "30%", func(i int) { settings.MixerVolume = 0.3 }).
		AddCheck("40", "40%", func(i int) { settings.MixerVolume = 0.4 }).
		AddCheck("50", "50%", func(i int) { settings.MixerVolume = 0.5 }).
		AddCheck("60", "60%", func(i int) { settings.MixerVolume = 0.6 }).
		AddCheck("70", "70%", func(i int) { settings.MixerVolume = 0.7 }).
		AddCheck("80", "80%", func(i int) { settings.MixerVolume = 0.8 }).
		AddCheck("90", "90%", func(i int) { settings.MixerVolume = 0.9 }).
		AddCheck("100", "100%", func(i int) { settings.MixerVolume = 1.0 }).
		Parent().
		AddMenu("speaker", "Speaker").Title("Vol").
		Before(
			func(m *Menu) {
				for i, item := range m.Items {
					item.Checked = i*10 == int(settings.SpeakerVolume[e.GetMemIndex()]*100)
				}
				m.IsPercentage = true
			},
		).
		AddCheck("0", "0%", func(i int) {
			settings.SpeakerVolume[e.GetMemIndex()] = 0
			cfg.SetF("audio", "init.speaker.volume", 0)
		}).
		AddCheck("10", "10%", func(i int) {
			settings.SpeakerVolume[e.GetMemIndex()] = 0.1
			cfg.SetF("audio", "init.speaker.volume", 0.1)

		}).
		AddCheck("20", "20%", func(i int) {
			settings.SpeakerVolume[e.GetMemIndex()] = 0.2
			cfg.SetF("audio", "init.speaker.volume", 0.2)
		}).
		AddCheck("30", "30%", func(i int) {
			settings.SpeakerVolume[e.GetMemIndex()] = 0.3
			cfg.SetF("audio", "init.speaker.volume", 0.3)
		}).
		AddCheck("40", "40%", func(i int) {
			settings.SpeakerVolume[e.GetMemIndex()] = 0.4
			cfg.SetF("audio", "init.speaker.volume", 0.4)
		}).
		AddCheck("50", "50%", func(i int) {
			settings.SpeakerVolume[e.GetMemIndex()] = 0.5
			cfg.SetF("audio", "init.speaker.volume", 0.5)
		}).
		AddCheck("60", "60%", func(i int) {
			settings.SpeakerVolume[e.GetMemIndex()] = 0.6
			cfg.SetF("audio", "init.speaker.volume", 0.6)
		}).
		AddCheck("70", "70%", func(i int) {
			settings.SpeakerVolume[e.GetMemIndex()] = 0.7
			cfg.SetF("audio", "init.speaker.volume", 0.7)
		}).
		AddCheck("80", "80%", func(i int) {
			settings.SpeakerVolume[e.GetMemIndex()] = 0.8
			cfg.SetF("audio", "init.speaker.volume", 0.8)
		}).
		AddCheck("90", "90%", func(i int) {
			settings.SpeakerVolume[e.GetMemIndex()] = 0.9
			cfg.SetF("audio", "init.speaker.volume", 0.9)
		}).
		AddCheck("100", "100%", func(i int) {
			settings.SpeakerVolume[e.GetMemIndex()] = 1.0
			cfg.SetF("audio", "init.speaker.volume", 1)
		}).
		Parent().
		AddMenu("mock", "Mockingboard").Title("MB").
		AddMenu("psg0", "Balance PSG0").Title("Bal").
		Before(
			func(m *Menu) {
				for _, item := range m.Items {
					item.Checked = item.Label == fmt.Sprintf("%d", int(settings.MockingBoardPSG0Bal*100))
					item.IsPercentage = true
				}
				m.IsPercentage = true
			},
		).
		AddCheck("-100", "Left", func(i int) {
			settings.MockingBoardPSG0Bal = -1
			cfg.SetF("audio", "init..mockingboard.psg0balance", -1)
		}).
		AddCheck("-75", "...", func(i int) {
			settings.MockingBoardPSG0Bal = -0.75
			cfg.SetF("audio", "init..mockingboard.psg0balance", -0.75)
		}).
		AddCheck("-50", "...", func(i int) {
			settings.MockingBoardPSG0Bal = -0.5
			cfg.SetF("audio", "init..mockingboard.psg0balance", -0.5)
		}).
		AddCheck("-25", "...", func(i int) {
			settings.MockingBoardPSG0Bal = -0.25
			cfg.SetF("audio", "init..mockingboard.psg0balance", -0.25)
		}).
		AddCheck("0", "Center", func(i int) {
			settings.MockingBoardPSG0Bal = 0
			cfg.SetF("audio", "init..mockingboard.psg0balance", 0)
		}).
		AddCheck("25", "...", func(i int) {
			settings.MockingBoardPSG0Bal = 0.25
			cfg.SetF("audio", "init..mockingboard.psg0balance", 0.25)
		}).
		AddCheck("50", "...", func(i int) {
			settings.MockingBoardPSG0Bal = 0.5
			cfg.SetF("audio", "init..mockingboard.psg0balance", 0.5)
		}).
		AddCheck("75", "...", func(i int) {
			settings.MockingBoardPSG0Bal = 0.75
			cfg.SetF("audio", "init..mockingboard.psg0balance", 0.75)
		}).
		AddCheck("100", "Right", func(i int) {
			settings.MockingBoardPSG0Bal = 1
			cfg.SetF("audio", "init..mockingboard.psg0balance", 1)
		}).
		Parent().
		AddMenu("psg1", "Balance PSG1").Title("Bal").
		Before(
			func(m *Menu) {
				for _, item := range m.Items {
					item.Checked = item.Label == fmt.Sprintf("%d", int(settings.MockingBoardPSG1Bal*100))
					item.IsPercentage = true
				}
				m.IsPercentage = true
			},
		).
		AddCheck("-100", "Left", func(i int) {
			settings.MockingBoardPSG1Bal = -1
			cfg.SetF("audio", "init..mockingboard.psg1balance", -1)
		}).
		AddCheck("-75", "...", func(i int) {
			settings.MockingBoardPSG1Bal = -0.75
			cfg.SetF("audio", "init..mockingboard.psg1balance", -0.75)
		}).
		AddCheck("-50", "...", func(i int) {
			settings.MockingBoardPSG1Bal = -0.5
			cfg.SetF("audio", "init..mockingboard.psg1balance", -0.5)
		}).
		AddCheck("-25", "...", func(i int) {
			settings.MockingBoardPSG1Bal = -0.25
			cfg.SetF("audio", "init..mockingboard.psg1balance", -0.25)
		}).
		AddCheck("0", "Center", func(i int) {
			settings.MockingBoardPSG1Bal = 0
			cfg.SetF("audio", "init..mockingboard.psg1balance", 0)
		}).
		AddCheck("25", "...", func(i int) {
			settings.MockingBoardPSG1Bal = 0.25
			cfg.SetF("audio", "init..mockingboard.psg1balance", 0.25)
		}).
		AddCheck("50", "...", func(i int) {
			settings.MockingBoardPSG1Bal = 0.5
			cfg.SetF("audio", "init..mockingboard.psg1balance", 0.5)
		}).
		AddCheck("75", "...", func(i int) {
			settings.MockingBoardPSG1Bal = 0.75
			cfg.SetF("audio", "init..mockingboard.psg1balance", 0.75)
		}).
		AddCheck("100", "Right", func(i int) {
			settings.MockingBoardPSG1Bal = 1
			cfg.SetF("audio", "init..mockingboard.psg1balance", 1)
		}).
		Parent().
		Parent().
		AddMenu("record", "Recording").Title("Record").
		Before(
			func(m *Menu) {
				for _, item := range m.Items {
					switch item.Label {
					case "pcm":
						item.Checked = clientperipherals.SPEAKER.Mixer.IsRecording()
					case "synth":
						item.Checked = clientperipherals.SPEAKER.Mixer.Slots[clientperipherals.SPEAKER.Mixer.SlotSelect].IsRecording()
					}
				}
			},
		).
		AddCheck("pcm", "Record PCM", func(i int) {
			clientperipherals.SPEAKER.Mixer.Slots[clientperipherals.SPEAKER.Mixer.SlotSelect].StopRecording()
			path := files.GetUserDirectory(files.BASEDIR + "/MyAudio")
			os.MkdirAll(path, 0755)
			filename := path + "/" + fmt.Sprintf("audio-%d.raw", time.Now().Unix())
			clientperipherals.SPEAKER.Mixer.StartRecording(filename)
		}).
		AddCheck("synth", "Record Synth", func(i int) {
			clientperipherals.SPEAKER.Mixer.StopRecording()
			rm := clientperipherals.SPEAKER.Mixer.Slots[clientperipherals.SPEAKER.Mixer.SlotSelect]
			path := files.GetUserDirectory(files.BASEDIR + "/MyAudio")
			os.MkdirAll(path, 0755)
			filename := path + "/" + fmt.Sprintf("restalgia-%d.rst", time.Now().Unix())
			rm.StartRecording(filename)
		}).
		AddSep().
		Add("stop", "Stop", func(i int) {
			clientperipherals.SPEAKER.Mixer.StopRecording()
			clientperipherals.SPEAKER.Mixer.Slots[clientperipherals.SPEAKER.Mixer.SlotSelect].StopRecording()
		}).
		Parent().
		Parent().
		AddMenu("input", "Input").
		Before(
			func(m *Menu) {
				for _, item := range m.Items {
					switch item.Label {
					case "allcaps":
						item.Checked = e.GetMemoryMap().IntGetUppercaseOnly(e.GetMemIndex())
					case "paddlekeys":
						item.Checked = settings.ArrowKeyPaddles
					case "disablemeta":
						item.Checked = settings.DisableMetaMode[e.GetMemIndex()]
					}
				}
			},
		).
		AddMenu("speed", "Paste Speed").
		Before(
			func(m *Menu) {
				for _, item := range m.Items {
					switch item.Label {
					case "slow":
						item.Checked = settings.PasteCPS == 10
					case "medium":
						item.Checked = settings.PasteCPS == 50
					case "fast":
						item.Checked = settings.PasteCPS == 100
					case "vfast":
						item.Checked = settings.PasteCPS == 200
					case "warp":
						item.Checked = settings.PasteWarp
					}
				}
			},
		).
		AddCheck("slow", "10cps", func(i int) {
			settings.PasteCPS = 10
			cfg.SetI("input", "init.paste.cps", settings.PasteCPS)
		}).
		AddCheck("medium", "50cps", func(i int) {
			settings.PasteCPS = 50
			cfg.SetI("input", "init.paste.cps", settings.PasteCPS)
		}).
		AddCheck("fast", "100cps", func(i int) {
			settings.PasteCPS = 100
			cfg.SetI("input", "init.paste.cps", settings.PasteCPS)
		}).
		AddCheck("vfast", "200cps", func(i int) {
			settings.PasteCPS = 200
			cfg.SetI("input", "init.paste.cps", settings.PasteCPS)
		}).
		AddSep().
		AddCheck("warp", "Warp during paste", func(i int) {
			settings.PasteWarp = !settings.PasteWarp
			if settings.PasteWarp {
				cfg.SetI("input", "init.paste.warp", 1)
			} else {
				cfg.SetI("input", "init.paste.warp", 0)
			}
		}).
		Parent().
		AddCheck("allcaps", "All Caps", func(i int) {
			e.GetMemoryMap().IntSetUppercaseOnly(e.GetMemIndex(), !e.GetMemoryMap().IntGetUppercaseOnly(e.GetMemIndex()))
			if e.GetMemoryMap().IntGetUppercaseOnly(e.GetMemIndex()) {
				cfg.SetI("input", "init.uppercase", 1)
			} else {
				cfg.SetI("input", "init.uppercase", 0)
			}
		}).
		AddMenu("joystick", "Joystick").
		Before(
			func(m *Menu) {
				for _, item := range m.Items {
					switch item.Label {
					case "jsxreverse":
						item.Checked = settings.JoystickReverseX[e.GetMemIndex()]
					case "jsyreverse":
						item.Checked = settings.JoystickReverseY[e.GetMemIndex()]
					case "jsreverse":
						item.Checked = e.GetMemoryMap().PaddleMap[e.GetMemIndex()][0] == 1
					}
				}
			},
		).
		AddCheck("jsreverse", "Axis swap", func(i int) {
			mm := e.GetMemoryMap()
			idx := e.GetMemIndex()
			p0 := mm.PaddleMap[idx][0]
			p1 := mm.PaddleMap[idx][1]
			mm.PaddleMap[idx][0] = p1
			mm.PaddleMap[idx][1] = p0
			cfg.SetI("input", "init.joystick.axis0", p1)
			cfg.SetI("input", "init.joystick.axis1", p0)
		}).
		AddCheck("jsxreverse", "X Axis Reverse", func(i int) {
			settings.JoystickReverseX[e.GetMemIndex()] = !settings.JoystickReverseX[e.GetMemIndex()]
			if settings.JoystickReverseX[e.GetMemIndex()] {
				cfg.SetI("input", "init.joystick.reversex", 1)
			} else {
				cfg.SetI("input", "init.joystick.reversex", 0)
			}
		}).
		AddCheck("jsyreverse", "Y Axis Reverse", func(i int) {
			settings.JoystickReverseY[e.GetMemIndex()] = !settings.JoystickReverseY[e.GetMemIndex()]
			if settings.JoystickReverseY[e.GetMemIndex()] {
				cfg.SetI("input", "init.joystick.reversey", 1)
			} else {
				cfg.SetI("input", "init.joystick.reversey", 0)
			}
		}).
		Parent().
		AddCheck("paddlekeys", "Arrow key paddles", func(i int) {
			settings.ArrowKeyPaddles = !settings.ArrowKeyPaddles
		}).
		AddMenu("mouse", "Mouse mode").
		Before(
			func(m *Menu) {
				for i, item := range m.Items {
					item.Hint = fmt.Sprintf(SymbolCTRL+SymbolSHIFT+"M,%d", i+1)
					switch item.Label {
					case "joystick":
						item.Checked = settings.GetMouseMode() == settings.MM_MOUSE_JOYSTICK
					case "dpad":
						item.Checked = settings.GetMouseMode() == settings.MM_MOUSE_DPAD
					case "geos":
						item.Checked = settings.GetMouseMode() == settings.MM_MOUSE_GEOS
					case "dazzle":
						item.Checked = settings.GetMouseMode() == settings.MM_MOUSE_DDRAW
					case "camera":
						item.Checked = settings.GetMouseMode() == settings.MM_MOUSE_CAMERA
					case "off":
						item.Checked = settings.GetMouseMode() == settings.MM_MOUSE_OFF
					}
				}
			},
		).
		AddCheck("joystick", "Joystick", func(i int) {
			settings.SetMouseMode(settings.MM_MOUSE_JOYSTICK)
			cfg.SetI("input", "init.mouse", int(settings.MM_MOUSE_JOYSTICK))
		}).
		AddCheck("dpad", "D-Pad", func(i int) {
			settings.SetMouseMode(settings.MM_MOUSE_DPAD)
			cfg.SetI("input", "init.mouse", int(settings.MM_MOUSE_DPAD))
		}).
		AddCheck("geos", "GEOS", func(i int) {
			settings.SetMouseMode(settings.MM_MOUSE_GEOS)
			//cfg.SetI("input", "init.mouse", int(settings.MM_MOUSE_GEOS))
		}).
		AddCheck("dazzle", "Dazzle Draw", func(i int) {
			settings.SetMouseMode(settings.MM_MOUSE_DDRAW)
			//cfg.SetI("input", "init.mouse", int(settings.MM_MOUSE_DDRAW))
		}).
		AddCheck("camera", "Camera", func(i int) {
			settings.SetMouseMode(settings.MM_MOUSE_CAMERA)
			//cfg.SetI("input", "init.mouse", int(settings.MM_MOUSE_DDRAW))
		}).
		AddCheck("off", "Off", func(i int) {
			settings.SetMouseMode(settings.MM_MOUSE_OFF)
			cfg.SetI("input", "init.mouse", int(settings.MM_MOUSE_OFF))
		}).
		Parent().
		AddCheck("disablemeta", "Disable Meta Keys", func(i int) {
			settings.DisableMetaMode[e.GetMemIndex()] = !settings.DisableMetaMode[e.GetMemIndex()]
		}).
		Parent().
		AddMenu("presets", "Presets").
		AddMenu("load", "Load Preset").
		AddCustom(func(m *Menu) {
			for i, name := range settingsFiles {
				m.Add(fmt.Sprintf("%d", i), fmt.Sprintf("%.2d: %s", i+1, strings.Replace(name[3:], "empty", "", -1)), func(j int) {
					fmt.Printf("Option %d selected\n", j)
					if j < len(settingsFiles) && !strings.HasSuffix(settingsFiles[j], "_empty") {
						filepath := "/local/settings/" + settingsFiles[j]
						p, err := files.OpenPresentationStateSoft(e, filepath)
						if err == nil {
							e.GetProducer().SetPState(e.GetMemIndex(), p, filepath)
						} else {
							fmt.Println(err)
						}
					}
				})
			}
		}).
		AddSep().
		Add("default", "Default", func(j int) {
			filepath := "/local/settings/default"
			p, err := files.OpenPresentationStateSoft(e, filepath)
			if err == nil {
				e.GetProducer().SetPState(e.GetMemIndex(), p, filepath)
			} else {
				fmt.Println(err)
			}
		}).
		Parent().
		AddMenu("save", "Save Preset").
		AddCustom(func(m *Menu) {
			for i, name := range settingsFiles {
				m.Add(fmt.Sprintf("%d", i), fmt.Sprintf("%.2d: %s", i+1, strings.Replace(name[3:], "empty", "", -1)), func(j int) {
					fmt.Printf("Option %d selected\n", j)
					if j < len(settingsFiles) {

						oldpath := files.GetUserDirectory(files.BASEDIR + "/settings/" + settingsFiles[j])
						name := InputPopup(e, "Enter name", "\r\nName", strings.Replace(settingsFiles[j][3:], "empty", "", -1))
						name = strings.Replace(name, ".", "", -1)
						filepath := fmt.Sprintf("/local/settings/%.2d_%s", j+1, name)
						os.RemoveAll(oldpath)
						_ = files.MkdirViaProvider(filepath)
						p, err := presentation.NewPresentationState(e, filepath)
						if err == nil {
							files.SavePresentationStateToFolder(p, filepath)
						}

					}
				})
			}
		}).
		AddSep().
		Add("default", "Default", func(j int) {
			settings.SkipCameraOnSave = true
			filepath := "/local/settings/default"
			_ = files.MkdirViaProvider(filepath)
			p, err := presentation.NewPresentationState(e, filepath)
			if err == nil {
				files.SavePresentationStateToFolder(p, filepath)
			}
			settings.SkipCameraOnSave = false
		}).
		Parent().
		AddSep().
		Add("resdef", "Reset Default", func(j int) {
			filepath := "/local/settings/default"
			_ = files.MkdirViaProvider(filepath)
			p, err := files.NewPresentationStateDefault(e, filepath)
			if err == nil {
				files.SavePresentationStateToFolder(p, filepath)
				e.GetProducer().SetPState(e.GetMemIndex(), p, filepath)
			}
		}).
		// AddSep().
		// AddIf(len(settingsFiles) < 16, "save", "Save custom...", func(i int) {
		// 	// prompt
		// 	name := InputPopup(e, "Preset Name", "\r\nPreset name:")
		// 	name = strings.Replace(name, ".", "", -1)
		// 	filepath := "/local/settings/" + name
		// 	_ = files.MkdirViaProvider(filepath)
		// 	p, err := presentation.NewPresentationState(e, filepath)
		// 	if err == nil {
		// 		files.SavePresentationStateToFolder(p, filepath)
		// 	}
		// }).
		// Add("savedefault", "Save default", func(i int) {
		// 	filepath := "/local/settings/default"
		// 	_ = files.MkdirViaProvider(filepath)
		// 	p, err := presentation.NewPresentationState(e, filepath)
		// 	if err == nil {
		// 		files.SavePresentationStateToFolder(p, filepath)
		// 	}
		// }).
		Parent().
		// main
		AddMenu("apps", "Applications").Title("Apps").
		AddMenu("pt", "PLATOTerm").Title("PLATOTerm").
		Add("irata", "Irata Online", func(i int) {
			settings.MicroPakPath = "/micropaks/comms/platoterm-irataonline.pak"
			//settings.ModemInitString[e.GetMemIndex()] = "atx telnet irata.online 8005"
			e.GetMemoryMap().IntSetSlotRestart(0, true)
		}).
		Add("cyber1", "Cyber1", func(i int) {
			settings.MicroPakPath = "/micropaks/comms/platoterm-cyberserv.pak"
			//settings.ModemInitString[e.GetMemIndex()] = "atx telnet irata.online 8005"
			e.GetMemoryMap().IntSetSlotRestart(0, true)
		}).
		Add("pthelp", "PLATOTerm Help", func(i int) {
			h := control.NewHelpController(e, "PLATOTerm Help", "/boot/help/platoterm", "platoterm")
			h.Do(e)
			return
		}).
		Parent().
		AddSep().
		Add("term", "Dial BBSes", func(i int) {
			settings.MicroPakPath = "/micropaks/comms/proterm.pak"
			//settings.ModemInitString[e.GetMemIndex()] = "atx telnet irata.online 8005"
			e.GetMemoryMap().IntSetSlotRestart(0, true)
		}).
		Add("ps", "Print Shop", func(i int) {
			c := "/appleii/disk images/applications/print shop/the print shop (color version).nib"
			settings.PureBootVolume[e.GetMemIndex()] = c
			settings.MicroPakPath = ""
			settings.PureBootSmartVolume[e.GetMemIndex()] = ""
			e.GetMemoryMap().IntSetSlotRestart(e.GetMemIndex(), true)
		}).
		Add("816", "816 Paint", func(i int) {
			c := "/appleii/disk images/2mg_hdv/816paint.po"
			settings.PureBootSmartVolume[e.GetMemIndex()] = c
			settings.MicroPakPath = ""
			settings.PureBootVolume[e.GetMemIndex()] = ""
			e.GetMemoryMap().IntSetSlotRestart(e.GetMemIndex(), true)
		}).
		Add("bw", "Beagle Write", func(i int) {
			c := "/appleii/disk images/2mg_hdv/beagle_write_v3.2_1989.2mg"
			settings.PureBootSmartVolume[e.GetMemIndex()] = c
			settings.MicroPakPath = ""
			settings.PureBootVolume[e.GetMemIndex()] = ""
			e.GetMemoryMap().IntSetSlotRestart(e.GetMemIndex(), true)
		}).
		Parent().
		AddMenu("tools", "Tools").Title("Tools").
		Add("tracker", "microTracker", func(i int) {
			apple2helpers.CommandResetOverride(e, "fp", "/local", "@music.edit{}")
		}).
		Add("spr", "Sprite Editor", func(i int) {
			apple2helpers.CommandResetOverride(e, "fp", "/", "RUN /boot/apps/spreditor.apl")
		}).
		AddMenu("debug", "Web Debugger").
		Before(func(m *Menu) {
			for _, item := range m.Items {
				item.Checked = item.Label == fmt.Sprintf("%d", settings.DebuggerAttachSlot)
			}
		}).
		AddCheck("0", "Disabled", func(i int) {
			settings.DebuggerAttachSlot = 0
			//settings.DebuggerOn = false
		}).
		AddSep().
		AddCheck("1", "Attach VM #1", func(i int) {
			settings.DebuggerAttachSlot = 1
			//settings.DebuggerOn = true
			checkExecDebugger()
		}).
		AddCheck("2", "Attach VM #2", func(i int) {
			settings.DebuggerAttachSlot = 2
			//settings.DebuggerOn = true
			checkExecDebugger()
		}).
		AddCheck("3", "Attach VM #3", func(i int) {
			settings.DebuggerAttachSlot = 3
			//settings.DebuggerOn = true
			checkExecDebugger()
		}).
		AddCheck("4", "Attach VM #4", func(i int) {
			settings.DebuggerAttachSlot = 4
			//settings.DebuggerOn = true
			checkExecDebugger()
		}).
		AddCheck("5", "Attach VM #5", func(i int) {
			settings.DebuggerAttachSlot = 5
			//settings.DebuggerOn = true
			checkExecDebugger()
		}).
		AddCheck("6", "Attach VM #6", func(i int) {
			settings.DebuggerAttachSlot = 6
			//settings.DebuggerOn = true
			checkExecDebugger()
		}).
		AddCheck("7", "Attach VM #7", func(i int) {
			settings.DebuggerAttachSlot = 7
			//settings.DebuggerOn = true
			checkExecDebugger()
		}).
		AddCheck("8", "Attach VM #8", func(i int) {
			settings.DebuggerAttachSlot = 8
			//settings.DebuggerOn = true
			checkExecDebugger()
		}).
		Parent().
		Parent().
		Add("cat", "File Catalog", func(i int) { e.GetMemoryMap().IntSetSlotInterrupt(e.GetMemIndex(), true) }).
		AddSep().
		AddCheck("contrast", "High Contrast", func(i int) {
			settings.HighContrastUI = !settings.HighContrastUI
			v := 0
			if settings.HighContrastUI {
				v = 1
			}
			cfg.SetI("video", "init.highcontrast", v)
			settings.LastScanLineIntensity = -0.1
		}).
		AddCheck("menudisable", "Disable icon", func(i int) {
			settings.ShowHamburger = !settings.ShowHamburger
			if !settings.ShowHamburger {
				time.AfterFunc(1*time.Second, func() {
					apple2helpers.OSDShow(e, "Shift-Control-. activates menu")
				})
			}
		}).
		AddCheck("menuhover", "Auto hide icon", func(i int) {
			settings.HamburgerOnHover = !settings.HamburgerOnHover
			v := 0
			if settings.HamburgerOnHover {
				v = 1
			}
			cfg.SetI("video", "init.menuhover", v)
		}).
		AddSep().
		Add("reset", "Reset this VM", func(i int) {
			e.GetMemoryMap().IntSetSlotRestart(e.GetMemIndex(), true)
		}).
		AddSep().
		Add("about", "About microM8", func(i int) {
			h := control.NewHelpController(e, "About microM8", "/boot/help", "about")
			h.Do(e)
		}).
		Add("notes", "Release Notes", func(i int) {
			control.CheckNewReleaseNotes(e, true)
			e.GetMemoryMap().IntSetSlotMenu(e.GetMemIndex(), false)
		}).
		Add("help", "Help", func(i int) {
			control.HelpPresent(e)
		}).
		AddSep().
		Add("quit", "Quit microM8", func(i int) { os.Exit(0) })

	zz := m.Find("video").menu

	if strings.HasPrefix(settings.SystemID[e.GetMemIndex()], "spectrum") {
		zz.AddMenu("rendermenu", "Spectrum Render mode").
			Before(
				func(m *Menu) {
					// prepare menu
					for _, item := range m.Items {
						switch item.Label {
						case "renderCD":
							item.Checked = e.GetMemoryMap().IntGetSpectrumRender(e.GetMemIndex()) == settings.VM_DOTTY
						case "renderCR":
							item.Checked = e.GetMemoryMap().IntGetSpectrumRender(e.GetMemIndex()) == settings.VM_FLAT
						case "renderCV":
							item.Checked = e.GetMemoryMap().IntGetSpectrumRender(e.GetMemIndex()) == settings.VM_VOXELS
						}
					}
				},
			).
			AddCheck("renderCD", "Color Dots", func(i int) {
				e.GetMemoryMap().IntSetSpectrumRender(e.GetMemIndex(), settings.VM_DOTTY)
				cfg.SetI("video", "init.video.zxmode", int(settings.VM_DOTTY))
			}).
			AddCheck("renderCR", "Color Raster", func(i int) {
				e.GetMemoryMap().IntSetSpectrumRender(e.GetMemIndex(), settings.VM_FLAT)
				cfg.SetI("video", "init.video.zxmode", int(settings.VM_FLAT))
			}).
			AddCheck("renderCV", "Color Voxels", func(i int) {
				e.GetMemoryMap().IntSetSpectrumRender(e.GetMemIndex(), settings.VM_VOXELS)
				cfg.SetI("video", "init.video.zxmode", int(settings.VM_VOXELS))
			})
	}

	if strings.HasPrefix(settings.SystemID[e.GetMemIndex()], "apple2") {
		zz.AddMenu("rendermenugr", "GR Render mode").Title("GR Render").
			Before(
				func(m *Menu) {
					// prepare menu
					for _, item := range m.Items {
						switch item.Label {
						case "renderCR":
							item.Checked = e.GetMemoryMap().IntGetGRRender(e.GetMemIndex()) == settings.VM_FLAT
						case "renderCV":
							item.Checked = e.GetMemoryMap().IntGetGRRender(e.GetMemIndex()) == settings.VM_VOXELS
						}
					}
				},
			).
			AddCheck("renderCR", "Color Raster", func(i int) {
				e.GetMemoryMap().IntSetGRRender(e.GetMemIndex(), settings.VM_FLAT)
				cfg.SetI("video", "init.video.grmode", int(settings.VM_FLAT))
			}).
			AddCheck("renderCV", "Color Voxels", func(i int) {
				e.GetMemoryMap().IntSetGRRender(e.GetMemIndex(), settings.VM_VOXELS)
				cfg.SetI("video", "init.video.grmode", int(settings.VM_VOXELS))
			}).
			Parent().
			// hgr
			AddMenu("rendermenu", "HGR Render mode").
			Before(
				func(m *Menu) {
					// prepare menu
					for _, item := range m.Items {
						switch item.Label {
						case "renderCD":
							item.Checked = e.GetMemoryMap().IntGetHGRRender(e.GetMemIndex()) == settings.VM_DOTTY
						case "renderCR":
							item.Checked = e.GetMemoryMap().IntGetHGRRender(e.GetMemIndex()) == settings.VM_FLAT
						case "renderCV":
							item.Checked = e.GetMemoryMap().IntGetHGRRender(e.GetMemIndex()) == settings.VM_VOXELS
						case "renderMD":
							item.Checked = e.GetMemoryMap().IntGetHGRRender(e.GetMemIndex()) == settings.VM_MONO_DOTTY
						case "renderMR":
							item.Checked = e.GetMemoryMap().IntGetHGRRender(e.GetMemIndex()) == settings.VM_MONO_FLAT
						case "renderMV":
							item.Checked = e.GetMemoryMap().IntGetHGRRender(e.GetMemIndex()) == settings.VM_MONO_VOXELS
						case "renderNTSC":
							item.Checked = settings.UseDHGRForHGR[e.GetMemIndex()]
						case "vertBlend":
							item.Checked = settings.UseVerticalBlend[e.GetMemIndex()]
						}
					}
				},
			).
			AddCheck("renderCD", "Color Dots", func(i int) {
				e.GetMemoryMap().IntSetHGRRender(e.GetMemIndex(), settings.VM_DOTTY)
				cfg.SetI("video", "init.video.hgrmode", int(settings.VM_DOTTY))
			}).
			AddCheck("renderCR", "Color Raster", func(i int) {
				e.GetMemoryMap().IntSetHGRRender(e.GetMemIndex(), settings.VM_FLAT)
				cfg.SetI("video", "init.video.hgrmode", int(settings.VM_FLAT))
			}).
			AddCheck("renderCV", "Color Voxels", func(i int) {
				e.GetMemoryMap().IntSetHGRRender(e.GetMemIndex(), settings.VM_VOXELS)
				cfg.SetI("video", "init.video.hgrmode", int(settings.VM_VOXELS))
			}).
			AddSep().
			AddCheck("renderMD", "Mono Dots", func(i int) {
				e.GetMemoryMap().IntSetHGRRender(e.GetMemIndex(), settings.VM_MONO_DOTTY)
				cfg.SetI("video", "init.video.hgrmode", int(settings.VM_MONO_DOTTY))
			}).
			AddCheck("renderMR", "Mono Raster", func(i int) {
				e.GetMemoryMap().IntSetHGRRender(e.GetMemIndex(), settings.VM_MONO_FLAT)
				cfg.SetI("video", "init.video.hgrmode", int(settings.VM_MONO_FLAT))
			}).
			AddCheck("renderMV", "Mono Voxels", func(i int) {
				e.GetMemoryMap().IntSetHGRRender(e.GetMemIndex(), settings.VM_MONO_VOXELS)
				cfg.SetI("video", "init.video.hgrmode", int(settings.VM_MONO_VOXELS))
			}).
			AddSep().
			AddCheck("renderNTSC", "NTSC Mode", func(i int) {
				settings.UseDHGRForHGR[e.GetMemIndex()] = !settings.UseDHGRForHGR[e.GetMemIndex()]
				h1, _ := e.GetGFXLayerByID("HGR1")
				h1.SetDirty(true)
				h2, _ := e.GetGFXLayerByID("HGR2")
				h2.SetDirty(true)
				if settings.UseDHGRForHGR[e.GetMemIndex()] {
					cfg.SetI("video", "init.video.hgrntsc", 1)
				} else {
					cfg.SetI("video", "init.video.hgrntsc", 0)
				}
			}).
			AddCheck("vertBlend", "Vertical Blending", func(i int) {
				settings.UseVerticalBlend[e.GetMemIndex()] = !settings.UseVerticalBlend[e.GetMemIndex()]
				h1, _ := e.GetGFXLayerByID("HGR1")
				h1.SetDirty(true)
				h2, _ := e.GetGFXLayerByID("HGR2")
				h2.SetDirty(true)
				if settings.UseVerticalBlend[e.GetMemIndex()] {
					cfg.SetI("video", "init.video.vertblend", 1)
				} else {
					cfg.SetI("video", "init.video.vertblend", 0)
				}
			}).
			Parent().
			// dhgr
			AddMenu("rendermenu", "DHGR Render mode").
			Before(
				func(m *Menu) {
					// prepare menu
					for _, item := range m.Items {
						switch item.Label {
						case "renderCD":
							item.Checked = e.GetMemoryMap().IntGetDHGRRender(e.GetMemIndex()) == settings.VM_DOTTY
						case "renderCR":
							item.Checked = e.GetMemoryMap().IntGetDHGRRender(e.GetMemIndex()) == settings.VM_FLAT
						case "renderCV":
							item.Checked = e.GetMemoryMap().IntGetDHGRRender(e.GetMemIndex()) == settings.VM_VOXELS
						case "renderMD":
							item.Checked = e.GetMemoryMap().IntGetDHGRRender(e.GetMemIndex()) == settings.VM_MONO_DOTTY
						case "renderMR":
							item.Checked = e.GetMemoryMap().IntGetDHGRRender(e.GetMemIndex()) == settings.VM_MONO_FLAT
						case "renderMV":
							item.Checked = e.GetMemoryMap().IntGetDHGRRender(e.GetMemIndex()) == settings.VM_MONO_VOXELS
						}
					}
				},
			).
			AddCheck("renderCD", "Color Dots", func(i int) {
				e.GetMemoryMap().IntSetDHGRRender(e.GetMemIndex(), settings.VM_DOTTY)
				cfg.SetI("video", "init.video.dhgrmode", int(settings.VM_DOTTY))
			}).
			AddCheck("renderCR", "Color Raster", func(i int) {
				e.GetMemoryMap().IntSetDHGRRender(e.GetMemIndex(), settings.VM_FLAT)
				cfg.SetI("video", "init.video.dhgrmode", int(settings.VM_FLAT))
			}).
			AddCheck("renderCV", "Color Voxels", func(i int) {
				e.GetMemoryMap().IntSetDHGRRender(e.GetMemIndex(), settings.VM_VOXELS)
				cfg.SetI("video", "init.video.dhgrmode", int(settings.VM_VOXELS))
			}).
			AddSep().
			AddCheck("renderMD", "Mono Dots", func(i int) {
				e.GetMemoryMap().IntSetDHGRRender(e.GetMemIndex(), settings.VM_MONO_DOTTY)
				cfg.SetI("video", "init.video.dhgrmode", int(settings.VM_MONO_DOTTY))
			}).
			AddCheck("renderMR", "Mono Raster", func(i int) {
				e.GetMemoryMap().IntSetDHGRRender(e.GetMemIndex(), settings.VM_MONO_FLAT)
				cfg.SetI("video", "init.video.dhgrmode", int(settings.VM_MONO_FLAT))
			}).
			AddCheck("renderMV", "Mono Voxels", func(i int) {
				e.GetMemoryMap().IntSetDHGRRender(e.GetMemIndex(), settings.VM_MONO_VOXELS)
				cfg.SetI("video", "init.video.dhgrmode", int(settings.VM_MONO_VOXELS))
			}).
			AddSep().
			AddMenu("highbit", "Enhanced mode").Title("Enhanced").
			Before(func(m *Menu) {
				for _, item := range m.Items {
					switch item.Label {
					case "auto":
						item.Checked = settings.DHGRHighBit[e.GetMemIndex()] == settings.DHB_MIXED_AUTO
					case "off":
						item.Checked = settings.DHGRHighBit[e.GetMemIndex()] == settings.DHB_MIXED_OFF
					case "on":
						item.Checked = settings.DHGRHighBit[e.GetMemIndex()] == settings.DHB_MIXED_ON
					}
				}
			}).
			AddCheck("auto", "Autodetect", func(i int) {
				settings.DHGRHighBit[e.GetMemIndex()] = settings.DHB_MIXED_AUTO
				cfg.SetI("video", "init.video.dhgrhighbit", int(settings.DHB_MIXED_AUTO))
			}).
			AddCheck("off", "Off", func(i int) {
				settings.DHGRHighBit[e.GetMemIndex()] = settings.DHB_MIXED_OFF
				cfg.SetI("video", "init.video.dhgrhighbit", int(settings.DHB_MIXED_OFF))
			}).
			AddCheck("on", "On", func(i int) {
				settings.DHGRHighBit[e.GetMemIndex()] = settings.DHB_MIXED_ON
				cfg.SetI("video", "init.video.dhgrhighbit", int(settings.DHB_MIXED_ON))
			}).
			Parent().
			Parent().
			AddMenu("rendermenu", "SHR Render mode").
			Before(
				func(m *Menu) {
					// prepare menu
					for _, item := range m.Items {
						switch item.Label {
						case "renderCD":
							item.Checked = e.GetMemoryMap().IntGetSHRRender(e.GetMemIndex()) == settings.VM_DOTTY
						case "renderCR":
							item.Checked = e.GetMemoryMap().IntGetSHRRender(e.GetMemIndex()) == settings.VM_FLAT
						case "renderCV":
							item.Checked = e.GetMemoryMap().IntGetSHRRender(e.GetMemIndex()) == settings.VM_VOXELS
						}
					}
				},
			).
			AddCheck("renderCD", "Color Dots", func(i int) {
				e.GetMemoryMap().IntSetSHRRender(e.GetMemIndex(), settings.VM_DOTTY)
				cfg.SetI("video", "init.video.shrmode", int(settings.VM_DOTTY))
			}).
			AddCheck("renderCR", "Color Raster", func(i int) {
				e.GetMemoryMap().IntSetSHRRender(e.GetMemIndex(), settings.VM_FLAT)
				cfg.SetI("video", "init.video.shrmode", int(settings.VM_FLAT))
			}).
			AddCheck("renderCV", "Color Voxels", func(i int) {
				e.GetMemoryMap().IntSetSHRRender(e.GetMemIndex(), settings.VM_VOXELS)
				cfg.SetI("video", "init.video.shrmode", int(settings.VM_VOXELS))
			}).
			Parent()

	}
	fontList := settings.AuxFonts[e.GetMemIndex()]
	if len(fontList) > 0 {
		fm := zz.AddMenu("font", "Fonts").Title("Fonts")
		for i, fontName := range fontList {
			fm.Add(fmt.Sprintf("font%d", i), font.GetFontName(fontName), func(i int) {
				log.Printf("Loading from file: %s", fontList[i])
				f, err := font.LoadFromFile(fontList[i])
				if err == nil {
					settings.DefaultFont[e.GetMemIndex()] = f
					settings.ForceTextVideoRefresh = true
				}
			},
			)
		}
	}

	m.X = 1
	m.Y = 1

	apple2helpers.OSDPanel(e, true)
	settings.MenuActive = true
	defer func() {
		settings.MenuActive = false
		apple2helpers.OSDPanel(e, false)
	}()

	txt := apple2helpers.GETHUD(e, "OOSD").Control
	txt.Font = 0
	txt.FGColor = 15
	txt.BGColor = 0
	txt.ClearScreen()

	m.Run(e)

	fmt.Printf("save = %v", cfg.Finalize())
}
