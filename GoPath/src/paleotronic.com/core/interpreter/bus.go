package interpreter

import (
	"os"
	"runtime"
	"strings"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
)

func (c *Interpreter) HandleServiceBusRequest(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool) {
	log.Printf("Recieved ServiceBus request: %+v", r)

	switch r.Type {
	case servicebus.LaunchEmulator:
		settings.Pakfile[c.MemIndex] = ""
		p := r.Payload.(*servicebus.LaunchEmulatorTarget)
		n := p.Filename
		drive := p.Drive
		ext := files.GetExt(n)
		if files.IsBootable(ext) {
			if !settings.BlueScreen {
				log.Printf("Got file named %s", n)
				data, err := files.ReadBytes(n)
				if files.Apple2IsHighCapacity(ext, len(data)) {
					log.Println("Smartport handling")
					if err == nil {
						log.Println("No read error")
						disk := "local:" + n
						e := c
						//go dropAnimation(drive)
						settings.PureBootSmartVolume[e.MemIndex] = disk
						//hardware.DiskInsert(e, 0, settings.PureBootVolume[e.MemIndex], settings.PureBootVolumeWP[e.MemIndex])
						servicebus.SendServiceBusMessage(
							e.MemIndex,
							servicebus.SmartPortInsertFilename,
							servicebus.DiskTargetString{
								Drive:    0,
								Filename: disk,
							},
						)
					} else {
						panic(err)
					}
				} else {
					disk := "local:" + n
					e := c
					//go dropAnimation(drive)
					switch drive {
					case 1:
						settings.PureBootVolume[e.MemIndex] = disk
						settings.PureBootSmartVolume[e.MemIndex] = ""
					case 2:
						settings.PureBootVolume2[e.MemIndex] = disk
					}
					//hardware.DiskInsert(e, 0, settings.PureBootVolume[e.MemIndex], settings.PureBootVolumeWP[e.MemIndex])

					servicebus.SendServiceBusMessage(
						e.MemIndex,
						servicebus.DiskIIInsertFilename,
						servicebus.DiskTargetString{
							Drive:    drive - 1,
							Filename: disk,
						},
					)
				}
			} else {
				settings.SplashDisk = "local:" + n
			}
		} else {
			// handle other types
			if runtime.GOOS == "windows" && strings.HasPrefix(n, os.Getenv("HOMEDRIVE")) {
				n = n[2:]
			}
			n = strings.Replace(n, "\\", "/", -1)
			n = "/fs/" + strings.Trim(n, "/")
			log.Printf("Bootstrap path: %s", n)
			ent := c
			if files.IsBinary(ext) {
				fp, err := files.ReadBytesViaProvider(files.GetPath(n), files.GetFilename(n))
				if err == nil {
					startfile := fp.FileName
					path := files.GetPath(n)
					log.Printf("Going to try BRUN of file %s", startfile)
					apple2helpers.CommandResetOverride(
						ent,
						"fp",
						"/"+strings.Trim(path, "/")+"/",
						"PRINT CHR$(4);\"BRUN "+startfile[0:len(startfile)-len(ext)-1]+"\"",
					)
				}
			} else if launchable, dialect, command := files.IsLaunchable(ext); launchable {
				startfile := n
				path := files.GetPath(n)

				ent.Halt()
				ent.Halt6502(0)
				ent.Bootstrap(dialect, true)
				ent.SetWorkDir(files.GetPath(n))

				command = strings.Replace(command, "%f", startfile, -1)
				fmt.Printf("Launching command: %s\n", command)
				s := &settings.RState{
					Dialect:   dialect,
					Command:   command,
					WorkDir:   path,
					IsControl: p.IsControl,
				}
				settings.ResetState[ent.GetMemIndex()] = s
				ent.GetMemoryMap().IntSetSlotRestart(ent.GetMemIndex(), true)
				fmt.Printf("Requesting controlled restart of slot %d\n", ent.GetMemIndex())
			} else if runnable, d := files.IsRunnable(ext); runnable {
				log.Printf("Going to try running command: %s", n)
				switch d {
				case "fp":
					apple2helpers.CommandResetOverride(
						ent,
						d,
						files.GetPath(n),
						"run "+n,
					)
					settings.ResetState[ent.GetMemIndex()].IsControl = p.IsControl
				case "int":
					apple2helpers.CommandResetOverride(
						ent,
						d,
						files.GetPath(n),
						"run "+n,
					)
					settings.ResetState[ent.GetMemIndex()].IsControl = p.IsControl
				case "logo":
					apple2helpers.CommandResetOverride(
						ent,
						d,
						files.GetPath(n),
						"load \""+n,
					)
					settings.ResetState[ent.GetMemIndex()].IsControl = p.IsControl
				}
			} else if ext == "pak" {
				if err := ent.GetProducer().BootStrap(n); err == nil {
					apple2helpers.MonitorPanel(ent, false)
					ent.GetProducer().Select(ent.GetMemIndex())
					ent.GetMemoryMap().MetaKeySet(ent.GetMemIndex(), vduconst.SCTRL1+rune(0), ' ')
				}
			}
		}
	}

	return &servicebus.ServiceBusResponse{
		Payload: "",
	}, true
}

func (c *Interpreter) InjectServiceBusRequest(r *servicebus.ServiceBusRequest) {
	log.Printf("Injecting ServiceBus request: %+v", r)
	c.m.Lock()
	defer c.m.Unlock()
	if c.injectedBusRequests == nil {
		c.injectedBusRequests = make([]*servicebus.ServiceBusRequest, 0, 16)
	}
	c.injectedBusRequests = append(c.injectedBusRequests, r)
}

func (c *Interpreter) HandleServiceBusInjection(handler func(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool)) {
	if c.injectedBusRequests == nil || len(c.injectedBusRequests) == 0 {
		return
	}
	c.m.Lock()
	defer c.m.Unlock()
	for _, r := range c.injectedBusRequests {
		if handler != nil {
			handler(r)
		}
	}
	c.injectedBusRequests = make([]*servicebus.ServiceBusRequest, 0, 16)
}

func (c *Interpreter) ServiceBusProcessPending() {
	c.HandleServiceBusInjection(c.HandleServiceBusRequest)
}
