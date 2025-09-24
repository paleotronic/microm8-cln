package core

import (
	"bytes"
	"strings"

	"paleotronic.com/core/hardware/spectrum/snapshot"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/freeze"
	"paleotronic.com/presentation"
)

type VMLaunchType int

type VMLauncherConfig struct {
	WorkingDir     string
	Disks          []string
	Pakfile        *files.OctContainer
	SmartPort      string
	RunFile        string
	RunCommand     string
	Dialect        string
	ResumeCPU      bool
	VideoFrames    []*bytes.Buffer
	VideoBackwards bool
	Freeze         *freeze.FreezeState
	Controls       []string
	VideoOff       bool
	Pre            *presentation.Presentation
	ZXState        *snapshot.Z80
	Forceboot      bool
}

var defaultShellInit = `
cls
display "Welcome to microShell..."
display ""
display "Type 'fp' for floating point basic, 'int' for integer basic',"
display "or 'logo' for 3D logo!"
display ""
cd ~
`

func (vm *VM) ApplyLaunchConfig(config *VMLauncherConfig) error {

	if config.ZXState != nil || strings.Contains(settings.SpecFile[vm.Index], "spectrum") {
		settings.UnifiedRender[vm.Index] = false
	} else {
		settings.UnifiedRender[vm.Index] = settings.UnifiedRenderGlobal
	}

	vm.GetInterpreter().SetState(types.STOPPED)

	if config == nil {

		// fallback to shell
		_, err := vm.ExecuteRequest("vm.hud.select", "TEXT")
		if err != nil {
			return err
		}
		_, err = vm.ExecuteRequest("vm.hud.mode", "mode40")
		if err != nil {
			return err
		}
		_, err = vm.ExecuteRequest("vm.interpreter.bootstrap", "shell")
		if err != nil {
			return err
		}

		return nil
	}

	defer func() {
		if config.VideoOff {
			vm.Logf("Turning video off..")
			vm.RAM.IntSetActiveState(vm.Index, 0)
			vm.RAM.IntSetLayerState(vm.Index, 0)
			// vm.DisableHUDLayers()
			// vm.DisableGFXLayers()
		}
	}()

	//log.Printf("Got launch config: %+v", *config)

	defer func() {
		// Apply presentation state
		if config.Pre != nil {
			vm.GetInterpreter().SetWorkDir(config.Pre.Filepath)
			config.Pre.Apply("init", vm.GetInterpreter())
		}
		// handle controls
		if len(config.Controls) > 0 {
			for i, cp := range config.Controls {
				slot := 1 + vm.Index + i
				// log.Printf("Begetting a child vm#%d for %s, (p==%v)", slot, cp, vm.p)
				err := vm.p.CreateVM(
					slot,
					&VMLauncherConfig{
						Dialect:  "fp",
						RunFile:  cp,
						VideoOff: true,
					},
					nil, // this must be nil or bad things happen with a subslot reboot in a pak
				)
				if err != nil {
					return
				}
				// log.Printf("Post beget (p==%v)", vm.p)
				nvm := vm.p.VM[slot]
				vm.AddDependant(nvm)
			}
		}
	}()

	config.Forceboot = settings.ForcePureBoot[vm.Index]

	// Freeze state
	if config.ZXState != nil {
		_, err := vm.ExecuteRequest("vm.zxstate.restore", config.ZXState)
		if err != nil {
			return err
		}
		return nil
	}

	// Freeze state
	if config.Freeze != nil {
		_, err := vm.ExecuteRequest("vm.freeze.restore", config.Freeze)
		if err != nil {
			return err
		}
		return nil
	}

	// video
	if config.VideoFrames != nil {
		_, err := vm.ExecuteRequest("vm.recording.play", config.VideoFrames, config.VideoBackwards)
		if err != nil {
			return err
		}
		return nil
	}

	// resume cpu
	if config.ResumeCPU {
		_, err := vm.ExecuteRequest("vm.cpu.resume")
		if err != nil {
			return err
		}
		return nil
	}

	// Bootable image support
	if len(config.Disks) > 0 || config.SmartPort != "" || config.Forceboot {
		// Floppy images
		if len(config.Disks) > 0 {
			for i, diskname := range config.Disks {
				if diskname == "" {
					continue
				}
				if i > 1 {
					break
				}
				_, err := vm.ExecuteRequest("vm.hardware.floppyinsert", i, diskname)
				if err != nil {
					return err
				}
			}
		}

		// Smartport
		if config.SmartPort != "" {
			_, err := vm.ExecuteRequest("vm.hardware.smartportinsert", 0, config.SmartPort)
			if err != nil {
				return err
			}
		}

		// Set modes and start CPU
		if strings.HasPrefix(settings.CPUModel[vm.Index], "65") {
			_, err := vm.ExecuteRequest("vm.hud.mode", "mode40.preserve")
			if err != nil {
				return err
			}
			_, err = vm.ExecuteRequest("vm.hardware.start6502")
			if err != nil {
				return err
			}
			if settings.AutoLiveRecording() && vm.Index == 0 {
				vm.GetInterpreter().StopRecording()
				vm.GetInterpreter().StartRecording("", false) // enable live record
			}
		} else if settings.CPUModel[vm.Index] == "Z80" {
			_, err := vm.ExecuteRequest("vm.hardware.startz80")
			if err != nil {
				return err
			}
		}

		return nil
	}

	// force UR off for anything that does not need it
	settings.UnifiedRender[vm.Index] = false
	settings.PureBootVolume[vm.Index] = ""
	settings.PureBootVolume2[vm.Index] = ""
	settings.PureBootSmartVolume[vm.Index] = ""

	if config.RunCommand != "" {
		// bootstrap dialect
		//if !config.VideoOff {
		_, err := vm.ExecuteRequest("vm.hud.select", "TEXT")
		if err != nil {
			return err
		}
		_, err = vm.ExecuteRequest("vm.hud.mode", "mode40")
		if err != nil {
			return err
		}
		//}
		if strings.HasPrefix(config.Dialect, "!") {
			_, err = vm.ExecuteRequest("vm.interpreter.bootstrap", strings.TrimLeft(config.Dialect, "!"))
			if err != nil {
				return err
			}
		} else {
			_, err = vm.ExecuteRequest("vm.interpreter.bootstrapsilent", config.Dialect)
			if err != nil {
				return err
			}
		}
		_, err = vm.ExecuteRequest("vm.files.setworkdir", config.WorkingDir)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(config.Dialect, "!") {
			_, err = vm.ExecuteRequest("vm.hud.clearscreen")
			if err != nil {
				return err
			}
		}
		if strings.Trim(config.RunCommand, " ") != "" {
			_, err = vm.ExecuteRequest("vm.interpreter.command", []string{config.RunCommand})
			if err != nil {
				return err
			}
		}
		//vm.GetInterpreter().SetNeedsPrompt(true)
		return nil
	}

	if config.RunFile != "" {
		// bootstrap dialect
		//if !config.VideoOff {
		_, err := vm.ExecuteRequest("vm.hud.select", "TEXT")
		if err != nil {
			return err
		}
		_, err = vm.ExecuteRequest("vm.hud.mode", "mode40")
		if err != nil {
			return err
		}
		//}
		_, err = vm.ExecuteRequest("vm.interpreter.bootstrapsilent", config.Dialect)
		if err != nil {
			return err
		}
		_, err = vm.ExecuteRequest("vm.files.setworkdir", config.WorkingDir)
		if err != nil {
			return err
		}
		_, err = vm.ExecuteRequest("vm.hud.clearscreen")
		if err != nil {
			return err
		}
		switch config.Dialect {
		case "logo":
			_, err = vm.ExecuteRequest("vm.interpreter.command", []string{"LOAD \"" + config.RunFile})
			if err != nil {
				return err
			}
		default:
			_, err = vm.ExecuteRequest("vm.interpreter.command", []string{"RUN \"" + config.RunFile + "\""})
			if err != nil {
				return err
			}
		}

	}

	if config.Pakfile != nil {

		vm.Logf("Starting pak load...")
		// we are going to load a pak...
		p, err := files.OpenPresentationState(config.Pakfile.GetPath())
		if err != nil {
			vm.Logf("Error reading pakfile: %v", err)
			return err
		}
		vm.Logf("Read/apply presentation state...")
		vm.GetInterpreter().SetWorkDir(p.Filepath)
		p.Apply("init", vm.GetInterpreter())

		startfile := config.Pakfile.GetStartup()
		if startfile != "" {
			cfg := &VMLauncherConfig{}

			music, leadin, fadein := config.Pakfile.GetMusicTrack()
			if music != "" {
				if !strings.HasPrefix(music, "/") {
					music = "/" + strings.Trim(config.Pakfile.GetPath(), "/") + "/" + music
				}
				_, err = vm.ExecuteRequest("vm.music.play", music, leadin, fadein)
				if err != nil {
					return err
				}
			}

			clist := config.Pakfile.GetControlFiles()
			if len(clist) > 0 {
				cfg.Controls = []string{}
				for _, cp := range clist {
					if files.ExistsViaProvider(config.Pakfile.GetPath(), cp+".apl") {
						cfg.Controls = append(cfg.Controls, "/"+strings.Trim(config.Pakfile.GetPath(), "/")+"/"+cp+".apl")
					}
				}
				vm.Logf("micro controls = %v", cfg.Controls)
			}

			ext := files.GetExt(startfile)
			vm.Logf("startfile = %s (%s)", startfile, ext)

			if ext == "frz" {
				cfg.WorkingDir = config.Pakfile.GetPath()
				data, _ := files.ReadBytesViaProvider("/"+strings.Trim(config.Pakfile.GetPath(), "/"), startfile)
				f := freeze.NewEmptyState(vm.GetInterpreter())
				cfg.Freeze = f
				cfg.Freeze.LoadFromBytes(data.Content)
				return vm.ApplyLaunchConfig(
					cfg,
				)
			} else if files.IsBootable(ext) {
				data, _ := files.ReadBytesViaProvider("/"+strings.Trim(config.Pakfile.GetPath(), "/"), startfile)
				if files.Apple2IsHighCapacity(ext, data.ContentSize) {
					cfg.WorkingDir = config.Pakfile.GetPath()
					cfg.SmartPort = config.Pakfile.GetPath() + "/" + startfile
					return vm.ApplyLaunchConfig(
						cfg,
					)
				} else {
					cfg.WorkingDir = config.Pakfile.GetPath()
					cfg.Disks = []string{config.Pakfile.GetPath() + "/" + startfile}
					return vm.ApplyLaunchConfig(
						cfg,
					)
				}
			} else if ok, dia, command := files.IsLaunchable(ext); ok {
				cfg.WorkingDir = config.Pakfile.GetPath()
				cfg.RunCommand = strings.Replace(command, "%f", startfile, -1)
				cfg.Dialect = dia
				return vm.ApplyLaunchConfig(
					cfg,
				)
			} else if ok, dia := files.IsRunnable(ext); ok {
				//log2.Printf("File is runnable (dialect %s): %s", dia, "/"+strings.Trim(config.Pakfile.GetPath(), "/")+"/"+startfile)
				cfg.WorkingDir = config.Pakfile.GetPath()
				cfg.RunFile = "/" + strings.Trim(config.Pakfile.GetPath(), "/") + "/" + startfile
				cfg.Dialect = dia
				return vm.ApplyLaunchConfig(
					cfg,
				)
			}
		}
	}

	return nil

}
