// +build nox

package ui

import (
	"os"
	"time"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
	"paleotronic.com/octalyzer/clientperipherals"
	"paleotronic.com/utils"
)

func checkExecDebugger() {
	if settings.DebuggerAttachSlot == 0 || !settings.DebuggerOn {
		return
	}
	settings.PureBootCheck(settings.DebuggerAttachSlot - 1)
	if settings.PureBoot(settings.DebuggerAttachSlot - 1) {
		utils.OpenURL(fmt.Sprintf("http://localhost:%d/?attach=%d", settings.DebuggerPort, settings.DebuggerAttachSlot))
	}
}

func TestMenu(e interfaces.Interpretable) {

	settings.TemporaryMute = true
	settings.VideoSuspended = false
	defer func() {
		settings.TemporaryMute = false
	}()

	//settingsFiles := files.GetSettingsFiles()

	cfg := NewDefaultSettings(e)

	m := NewMenu(nil).Title("Nox-Archaist").
		// video
		AddMenu("video", "Video").
		Before(func(m *Menu) {
			for _, item := range m.Items {
				if item.Label == "fs" {
					item.Checked = (!settings.Windowed)
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
		AddMenu("aspect", "Aspect Ratio").
		Before(
			func(m *Menu) {
				mm := e.GetMemoryMap()
				idx := e.GetMemIndex()
				control := types.NewOrbitController(mm, idx, -1)
				astr := fmt.Sprintf("%.2f", control.GetAspect())
				for _, item := range m.Items {
					item.Checked = astr == item.Label
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
		// hgr
		AddMenu("scanlines", "Scanline intensity").
		Before(
			func(m *Menu) {
				// prepare menu
				for _, item := range m.Items {
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
		Parent().
		// Audio menu
		AddMenu("audio", "Audio").Title("Audio").
		Before(
			func(m *Menu) {
				for _, item := range m.Items {
					if item.Label == "mute" {
						item.Checked = clientperipherals.SPEAKER.Mixer.IsMuted()
					}
				}
			},
		).
		AddCheck("mute", "Mute Audio", func(i int) { clientperipherals.SPEAKER.Mixer.SetMute(!clientperipherals.SPEAKER.Mixer.IsMuted()) }).
		// volume
		AddMenu("volume", "Master").Title("Vol").
		Before(
			func(m *Menu) {
				for i, item := range m.Items {
					if item.Label == "mute" {
						item.Checked = clientperipherals.SPEAKER.Mixer.IsMuted()
					} else {
						item.Checked = i*10 == int(settings.MixerVolume*100)
					}
				}
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
				}
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
				}
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
		// AddMenu("record", "Recording").Title("Record").
		// Before(
		// 	func(m *Menu) {
		// 		for _, item := range m.Items {
		// 			switch item.Label {
		// 			case "pcm":
		// 				item.Checked = clientperipherals.SPEAKER.Mixer.IsRecording()
		// 			case "synth":
		// 				item.Checked = clientperipherals.SPEAKER.Mixer.Slots[clientperipherals.SPEAKER.Mixer.SlotSelect].IsRecording()
		// 			}
		// 		}
		// 	},
		// ).
		// AddCheck("pcm", "Record PCM", func(i int) {
		// 	clientperipherals.SPEAKER.Mixer.Slots[clientperipherals.SPEAKER.Mixer.SlotSelect].StopRecording()
		// 	path := files.GetUserDirectory(files.BASEDIR + "/MyAudio")
		// 	os.MkdirAll(path, 0755)
		// 	filename := path + "/" + fmt.Sprintf("audio-%d.raw", time.Now().Unix())
		// 	clientperipherals.SPEAKER.Mixer.StartRecording(filename)
		// }).
		// AddCheck("synth", "Record Synth", func(i int) {
		// 	clientperipherals.SPEAKER.Mixer.StopRecording()
		// 	rm := clientperipherals.SPEAKER.Mixer.Slots[clientperipherals.SPEAKER.Mixer.SlotSelect]
		// 	path := files.GetUserDirectory(files.BASEDIR + "/MyAudio")
		// 	os.MkdirAll(path, 0755)
		// 	filename := path + "/" + fmt.Sprintf("restalgia-%d.rst", time.Now().Unix())
		// 	rm.StartRecording(filename)
		// }).
		// AddSep().
		// Add("stop", "Stop", func(i int) {
		// 	clientperipherals.SPEAKER.Mixer.StopRecording()
		// 	clientperipherals.SPEAKER.Mixer.Slots[clientperipherals.SPEAKER.Mixer.SlotSelect].StopRecording()
		// }).
		// Parent().
		Parent().
		// AddMenu("input", "Input").
		// Before(
		// 	func(m *Menu) {
		// 		for _, item := range m.Items {
		// 			switch item.Label {
		// 			case "allcaps":
		// 				item.Checked = e.GetMemoryMap().IntGetUppercaseOnly(e.GetMemIndex())
		// 			}
		// 		}
		// 	},
		// ).
		// AddCheck("allcaps", "All Caps", func(i int) {
		// 	e.GetMemoryMap().IntSetUppercaseOnly(e.GetMemIndex(), !e.GetMemoryMap().IntGetUppercaseOnly(e.GetMemIndex()))
		// 	if e.GetMemoryMap().IntGetUppercaseOnly(e.GetMemIndex()) {
		// 		cfg.SetI("input", "init.uppercase", 1)
		// 	} else {
		// 		cfg.SetI("input", "init.uppercase", 0)
		// 	}
		// }).
		// AddMenu("joystick", "Joystick").
		// Before(
		// 	func(m *Menu) {
		// 		for _, item := range m.Items {
		// 			switch item.Label {
		// 			case "jsxreverse":
		// 				item.Checked = settings.JoystickReverseX[e.GetMemIndex()]
		// 			case "jsyreverse":
		// 				item.Checked = settings.JoystickReverseY[e.GetMemIndex()]
		// 			case "jsreverse":
		// 				item.Checked = e.GetMemoryMap().PaddleMap[e.GetMemIndex()][0] == 1
		// 			}
		// 		}
		// 	},
		// ).
		// AddCheck("jsreverse", "Axis swap", func(i int) {
		// 	mm := e.GetMemoryMap()
		// 	idx := e.GetMemIndex()
		// 	p0 := mm.PaddleMap[idx][0]
		// 	p1 := mm.PaddleMap[idx][1]
		// 	mm.PaddleMap[idx][0] = p1
		// 	mm.PaddleMap[idx][1] = p0
		// 	cfg.SetI("input", "init.joystick.axis0", p1)
		// 	cfg.SetI("input", "init.joystick.axis1", p0)
		// }).
		// AddCheck("jsxreverse", "X Axis Reverse", func(i int) {
		// 	settings.JoystickReverseX[e.GetMemIndex()] = !settings.JoystickReverseX[e.GetMemIndex()]
		// 	if settings.JoystickReverseX[e.GetMemIndex()] {
		// 		cfg.SetI("input", "init.joystick.reversex", 1)
		// 	} else {
		// 		cfg.SetI("input", "init.joystick.reversex", 0)
		// 	}
		// }).
		// AddCheck("jsyreverse", "Y Axis Reverse", func(i int) {
		// 	settings.JoystickReverseY[e.GetMemIndex()] = !settings.JoystickReverseY[e.GetMemIndex()]
		// 	if settings.JoystickReverseY[e.GetMemIndex()] {
		// 		cfg.SetI("input", "init.joystick.reversey", 1)
		// 	} else {
		// 		cfg.SetI("input", "init.joystick.reversey", 0)
		// 	}
		// }).
		// Parent().
		// AddMenu("mouse", "Mouse mode").
		// Before(
		// 	func(m *Menu) {
		// 		for _, item := range m.Items {
		// 			switch item.Label {
		// 			case "joystick":
		// 				item.Checked = settings.GetMouseMode() == settings.MM_MOUSE_JOYSTICK
		// 			case "dpad":
		// 				item.Checked = settings.GetMouseMode() == settings.MM_MOUSE_DPAD
		// 			case "geos":
		// 				item.Checked = settings.GetMouseMode() == settings.MM_MOUSE_GEOS
		// 			case "dazzle":
		// 				item.Checked = settings.GetMouseMode() == settings.MM_MOUSE_DDRAW
		// 			case "off":
		// 				item.Checked = settings.GetMouseMode() == settings.MM_MOUSE_OFF
		// 			}
		// 		}
		// 	},
		// ).
		// AddCheck("joystick", "Joystick", func(i int) {
		// 	settings.SetMouseMode(settings.MM_MOUSE_JOYSTICK)
		// 	cfg.SetI("input", "init.mouse", int(settings.MM_MOUSE_JOYSTICK))
		// }).
		// AddCheck("dpad", "D-Pad", func(i int) {
		// 	settings.SetMouseMode(settings.MM_MOUSE_DPAD)
		// 	cfg.SetI("input", "init.mouse", int(settings.MM_MOUSE_DPAD))
		// }).
		// AddCheck("geos", "GEOS", func(i int) {
		// 	settings.SetMouseMode(settings.MM_MOUSE_GEOS)
		// 	//cfg.SetI("input", "init.mouse", int(settings.MM_MOUSE_GEOS))
		// }).
		// AddCheck("dazzle", "Dazzle Draw", func(i int) {
		// 	settings.SetMouseMode(settings.MM_MOUSE_DDRAW)
		// 	//cfg.SetI("input", "init.mouse", int(settings.MM_MOUSE_DDRAW))
		// }).
		// AddCheck("off", "Off", func(i int) {
		// 	settings.SetMouseMode(settings.MM_MOUSE_OFF)
		// 	cfg.SetI("input", "init.mouse", int(settings.MM_MOUSE_OFF))
		// }).
		// Parent().
		// Parent().
		// AddMenu("presets", "Presets").
		// AddMenu("load", "Load Preset").
		// AddCustom(func(m *Menu) {
		// 	for i, name := range settingsFiles {
		// 		m.Add(fmt.Sprintf("%d", i), fmt.Sprintf("%.2d: %s", i+1, strings.Replace(name[3:], "empty", "", -1)), func(j int) {
		// 			fmt.Printf("Option %d selected\n", j)
		// 			if j < len(settingsFiles) && !strings.HasSuffix(settingsFiles[j], "_empty") {
		// 				filepath := "/local/settings/" + settingsFiles[j]
		// 				p, err := files.OpenPresentationStateSoft(e, filepath)
		// 				if err == nil {
		// 					e.GetProducer().SetPState(e.GetMemIndex(), p, filepath)
		// 				} else {
		// 					fmt.Println(err)
		// 				}
		// 			}
		// 		})
		// 	}
		// }).
		// AddSep().
		// Add("default", "Default", func(j int) {
		// 	filepath := "/local/settings/default"
		// 	p, err := files.OpenPresentationStateSoft(e, filepath)
		// 	if err == nil {
		// 		e.GetProducer().SetPState(e.GetMemIndex(), p, filepath)
		// 	} else {
		// 		fmt.Println(err)
		// 	}
		// }).
		// Parent().
		// AddMenu("save", "Save Preset").
		// AddCustom(func(m *Menu) {
		// 	for i, name := range settingsFiles {
		// 		m.Add(fmt.Sprintf("%d", i), fmt.Sprintf("%.2d: %s", i+1, strings.Replace(name[3:], "empty", "", -1)), func(j int) {
		// 			fmt.Printf("Option %d selected\n", j)
		// 			if j < len(settingsFiles) {

		// 				oldpath := files.GetUserDirectory(files.BASEDIR + "/settings/" + settingsFiles[j])
		// 				name := InputPopup(e, "Enter name", "\r\nName", strings.Replace(settingsFiles[j][3:], "empty", "", -1))
		// 				name = strings.Replace(name, ".", "", -1)
		// 				filepath := fmt.Sprintf("/local/settings/%.2d_%s", j+1, name)
		// 				os.RemoveAll(oldpath)
		// 				_ = files.MkdirViaProvider(filepath)
		// 				p, err := presentation.NewPresentationState(e, filepath)
		// 				if err == nil {
		// 					files.SavePresentationStateToFolder(p, filepath)
		// 				}

		// 			}
		// 		})
		// 	}
		// }).
		// AddSep().
		// Add("default", "Default", func(j int) {
		// 	settings.SkipCameraOnSave = true
		// 	filepath := "/local/settings/default"
		// 	_ = files.MkdirViaProvider(filepath)
		// 	p, err := presentation.NewPresentationState(e, filepath)
		// 	if err == nil {
		// 		files.SavePresentationStateToFolder(p, filepath)
		// 	}
		// 	settings.SkipCameraOnSave = false
		// }).
		// Parent().
		// AddSep().
		// Add("resdef", "Reset Default", func(j int) {
		// 	filepath := "/local/settings/default"
		// 	_ = files.MkdirViaProvider(filepath)
		// 	p, err := files.NewPresentationStateDefault(e, filepath)
		// 	if err == nil {
		// 		files.SavePresentationStateToFolder(p, filepath)
		// 		e.GetProducer().SetPState(e.GetMemIndex(), p, filepath)
		// 	}
		// }).
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
		// Parent().
		// main
		// AddMenu("apps", "Applications").Title("Apps").
		// AddMenu("pt", "PLATOTerm").Title("PLATOTerm").
		// Add("irata", "Irata Online", func(i int) {
		// 	settings.MicroPakPath = "/micropaks/comms/platoterm-irataonline.pak"
		// 	//settings.ModemInitString[e.GetMemIndex()] = "atx telnet irata.online 8005"
		// 	e.GetMemoryMap().IntSetSlotRestart(0, true)
		// }).
		// Add("cyber1", "Cyber1", func(i int) {
		// 	settings.MicroPakPath = "/micropaks/comms/platoterm-cyberserv.pak"
		// 	//settings.ModemInitString[e.GetMemIndex()] = "atx telnet irata.online 8005"
		// 	e.GetMemoryMap().IntSetSlotRestart(0, true)
		// }).
		// Add("pthelp", "PLATOTerm Help", func(i int) {
		// 	h := control.NewHelpController(e, "PLATOTerm Help", "/boot/help/platoterm", "platoterm")
		// 	h.Do(e)
		// 	return
		// }).
		// Parent().
		// AddSep().
		// Add("ml", "microLink", func(i int) {
		// 	apple2helpers.CommandResetOverride(e, "fp", "/", "RUN /system/autoexec")
		// }).
		// Add("term", "Dial BBSes", func(i int) {
		// 	settings.MicroPakPath = "/micropaks/comms/proterm.pak"
		// 	//settings.ModemInitString[e.GetMemIndex()] = "atx telnet irata.online 8005"
		// 	e.GetMemoryMap().IntSetSlotRestart(0, true)
		// }).
		// Add("ps", "Print Shop", func(i int) {
		// 	c := "/appleii/disk images/applications/print shop/the print shop (color version).nib"
		// 	settings.PureBootVolume[e.GetMemIndex()] = c
		// 	settings.MicroPakPath = ""
		// 	settings.PureBootSmartVolume[e.GetMemIndex()] = ""
		// 	e.GetMemoryMap().IntSetSlotRestart(e.GetMemIndex(), true)
		// }).
		// Add("816", "816 Paint", func(i int) {
		// 	c := "/appleii/disk images/2mg_hdv/816paint.po"
		// 	settings.PureBootSmartVolume[e.GetMemIndex()] = c
		// 	settings.MicroPakPath = ""
		// 	settings.PureBootVolume[e.GetMemIndex()] = ""
		// 	e.GetMemoryMap().IntSetSlotRestart(e.GetMemIndex(), true)
		// }).
		// Parent().
		AddMenu("tools", "Tools").Title("Tools").
		// Add("tracker", "microTracker", func(i int) {
		// 	apple2helpers.CommandResetOverride(e, "fp", "/local", "@music.edit{}")
		// }).
		// Add("spr", "Sprite Editor", func(i int) {
		// 	apple2helpers.CommandResetOverride(e, "fp", "/", "RUN /boot/apps/spreditor.apl")
		// }).
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
		// Add("cat", "File Catalog", func(i int) { e.GetMemoryMap().IntSetSlotInterrupt(e.GetMemIndex(), true) }).
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
		// Add("menu", "Disable icon", func(i int) {
		// 	settings.ShowHamburger = !settings.ShowHamburger
		// 	if !settings.ShowHamburger {
		// 		time.AfterFunc(1*time.Second, func() {
		// 			apple2helpers.OSDShow(e, "Shift-Control-. activates menu")
		// 		})
		// 	}
		// }).
		AddSep().
		Add("reset", "Reset", func(i int) {
			e.GetMemoryMap().IntSetSlotRestart(e.GetMemIndex(), true)
		}).
		AddSep().
		Add("quit", "Quit Nox", func(i int) {
			e.VM().Teardown()
			os.Exit(0)
		})

	//zz := m.Find("video").menu

	// if strings.HasPrefix(settings.SystemID[e.GetMemIndex()], "spectrum") {
	// 	zz.AddMenu("rendermenu", "Spectrum Render mode").
	// 		Before(
	// 			func(m *Menu) {
	// 				// prepare menu
	// 				for _, item := range m.Items {
	// 					switch item.Label {
	// 					case "renderCD":
	// 						item.Checked = e.GetMemoryMap().IntGetSpectrumRender(e.GetMemIndex()) == settings.VM_DOTTY
	// 					case "renderCR":
	// 						item.Checked = e.GetMemoryMap().IntGetSpectrumRender(e.GetMemIndex()) == settings.VM_FLAT
	// 					case "renderCV":
	// 						item.Checked = e.GetMemoryMap().IntGetSpectrumRender(e.GetMemIndex()) == settings.VM_VOXELS
	// 					}
	// 				}
	// 			},
	// 		).
	// 		AddCheck("renderCD", "Color Dots", func(i int) {
	// 			e.GetMemoryMap().IntSetSpectrumRender(e.GetMemIndex(), settings.VM_DOTTY)
	// 			cfg.SetI("video", "init.video.zxmode", int(settings.VM_DOTTY))
	// 		}).
	// 		AddCheck("renderCR", "Color Raster", func(i int) {
	// 			e.GetMemoryMap().IntSetSpectrumRender(e.GetMemIndex(), settings.VM_FLAT)
	// 			cfg.SetI("video", "init.video.zxmode", int(settings.VM_FLAT))
	// 		}).
	// 		AddCheck("renderCV", "Color Voxels", func(i int) {
	// 			e.GetMemoryMap().IntSetSpectrumRender(e.GetMemIndex(), settings.VM_VOXELS)
	// 			cfg.SetI("video", "init.video.zxmode", int(settings.VM_VOXELS))
	// 		})
	// }

	// if strings.HasPrefix(settings.SystemID[e.GetMemIndex()], "apple2") {
	// 	zz.AddMenu("rendermenugr", "GR Render mode").Title("GR Render").
	// 		Before(
	// 			func(m *Menu) {
	// 				// prepare menu
	// 				for _, item := range m.Items {
	// 					switch item.Label {
	// 					case "renderCR":
	// 						item.Checked = e.GetMemoryMap().IntGetGRRender(e.GetMemIndex()) == settings.VM_FLAT
	// 					case "renderCV":
	// 						item.Checked = e.GetMemoryMap().IntGetGRRender(e.GetMemIndex()) == settings.VM_VOXELS
	// 					}
	// 				}
	// 			},
	// 		).
	// 		AddCheck("renderCR", "Color Raster", func(i int) {
	// 			e.GetMemoryMap().IntSetGRRender(e.GetMemIndex(), settings.VM_FLAT)
	// 			cfg.SetI("video", "init.video.grmode", int(settings.VM_FLAT))
	// 		}).
	// 		AddCheck("renderCV", "Color Voxels", func(i int) {
	// 			e.GetMemoryMap().IntSetGRRender(e.GetMemIndex(), settings.VM_VOXELS)
	// 			cfg.SetI("video", "init.video.grmode", int(settings.VM_VOXELS))
	// 		}).
	// 		Parent().
	// 		// hgr
	// 		AddMenu("rendermenu", "HGR Render mode").
	// 		Before(
	// 			func(m *Menu) {
	// 				// prepare menu
	// 				for _, item := range m.Items {
	// 					switch item.Label {
	// 					case "renderCD":
	// 						item.Checked = e.GetMemoryMap().IntGetHGRRender(e.GetMemIndex()) == settings.VM_DOTTY
	// 					case "renderCR":
	// 						item.Checked = e.GetMemoryMap().IntGetHGRRender(e.GetMemIndex()) == settings.VM_FLAT
	// 					case "renderCV":
	// 						item.Checked = e.GetMemoryMap().IntGetHGRRender(e.GetMemIndex()) == settings.VM_VOXELS
	// 					case "renderMD":
	// 						item.Checked = e.GetMemoryMap().IntGetHGRRender(e.GetMemIndex()) == settings.VM_MONO_DOTTY
	// 					case "renderMR":
	// 						item.Checked = e.GetMemoryMap().IntGetHGRRender(e.GetMemIndex()) == settings.VM_MONO_FLAT
	// 					case "renderMV":
	// 						item.Checked = e.GetMemoryMap().IntGetHGRRender(e.GetMemIndex()) == settings.VM_MONO_VOXELS
	// 					case "renderNTSC":
	// 						item.Checked = settings.UseDHGRForHGR[e.GetMemIndex()]
	// 					case "vertBlend":
	// 						item.Checked = settings.UseVerticalBlend[e.GetMemIndex()]
	// 					}
	// 				}
	// 			},
	// 		).
	// 		AddCheck("renderCD", "Color Dots", func(i int) {
	// 			e.GetMemoryMap().IntSetHGRRender(e.GetMemIndex(), settings.VM_DOTTY)
	// 			cfg.SetI("video", "init.video.hgrmode", int(settings.VM_DOTTY))
	// 		}).
	// 		AddCheck("renderCR", "Color Raster", func(i int) {
	// 			e.GetMemoryMap().IntSetHGRRender(e.GetMemIndex(), settings.VM_FLAT)
	// 			cfg.SetI("video", "init.video.hgrmode", int(settings.VM_FLAT))
	// 		}).
	// 		AddCheck("renderCV", "Color Voxels", func(i int) {
	// 			e.GetMemoryMap().IntSetHGRRender(e.GetMemIndex(), settings.VM_VOXELS)
	// 			cfg.SetI("video", "init.video.hgrmode", int(settings.VM_VOXELS))
	// 		}).
	// 		AddSep().
	// 		AddCheck("renderMD", "Mono Dots", func(i int) {
	// 			e.GetMemoryMap().IntSetHGRRender(e.GetMemIndex(), settings.VM_MONO_DOTTY)
	// 			cfg.SetI("video", "init.video.hgrmode", int(settings.VM_MONO_DOTTY))
	// 		}).
	// 		AddCheck("renderMR", "Mono Raster", func(i int) {
	// 			e.GetMemoryMap().IntSetHGRRender(e.GetMemIndex(), settings.VM_MONO_FLAT)
	// 			cfg.SetI("video", "init.video.hgrmode", int(settings.VM_MONO_FLAT))
	// 		}).
	// 		AddCheck("renderMV", "Mono Voxels", func(i int) {
	// 			e.GetMemoryMap().IntSetHGRRender(e.GetMemIndex(), settings.VM_MONO_VOXELS)
	// 			cfg.SetI("video", "init.video.hgrmode", int(settings.VM_MONO_VOXELS))
	// 		}).
	// 		AddSep().
	// 		AddCheck("renderNTSC", "NTSC Mode", func(i int) {
	// 			settings.UseDHGRForHGR[e.GetMemIndex()] = !settings.UseDHGRForHGR[e.GetMemIndex()]
	// 			h1, _ := e.GetGFXLayerByID("HGR1")
	// 			h1.SetDirty(true)
	// 			h2, _ := e.GetGFXLayerByID("HGR2")
	// 			h2.SetDirty(true)
	// 			if settings.UseDHGRForHGR[e.GetMemIndex()] {
	// 				cfg.SetI("video", "init.video.hgrntsc", 1)
	// 			} else {
	// 				cfg.SetI("video", "init.video.hgrntsc", 0)
	// 			}
	// 		}).
	// 		AddCheck("vertBlend", "Vertical Blending", func(i int) {
	// 			settings.UseVerticalBlend[e.GetMemIndex()] = !settings.UseVerticalBlend[e.GetMemIndex()]
	// 			h1, _ := e.GetGFXLayerByID("HGR1")
	// 			h1.SetDirty(true)
	// 			h2, _ := e.GetGFXLayerByID("HGR2")
	// 			h2.SetDirty(true)
	// 			if settings.UseVerticalBlend[e.GetMemIndex()] {
	// 				cfg.SetI("video", "init.video.vertblend", 1)
	// 			} else {
	// 				cfg.SetI("video", "init.video.vertblend", 0)
	// 			}
	// 		}).
	// 		Parent().
	// 		// dhgr
	// 		AddMenu("rendermenu", "DHGR Render mode").
	// 		Before(
	// 			func(m *Menu) {
	// 				// prepare menu
	// 				for _, item := range m.Items {
	// 					switch item.Label {
	// 					case "renderCD":
	// 						item.Checked = e.GetMemoryMap().IntGetDHGRRender(e.GetMemIndex()) == settings.VM_DOTTY
	// 					case "renderCR":
	// 						item.Checked = e.GetMemoryMap().IntGetDHGRRender(e.GetMemIndex()) == settings.VM_FLAT
	// 					case "renderCV":
	// 						item.Checked = e.GetMemoryMap().IntGetDHGRRender(e.GetMemIndex()) == settings.VM_VOXELS
	// 					case "renderMD":
	// 						item.Checked = e.GetMemoryMap().IntGetDHGRRender(e.GetMemIndex()) == settings.VM_MONO_DOTTY
	// 					case "renderMR":
	// 						item.Checked = e.GetMemoryMap().IntGetDHGRRender(e.GetMemIndex()) == settings.VM_MONO_FLAT
	// 					case "renderMV":
	// 						item.Checked = e.GetMemoryMap().IntGetDHGRRender(e.GetMemIndex()) == settings.VM_MONO_VOXELS
	// 					}
	// 				}
	// 			},
	// 		).
	// 		AddCheck("renderCD", "Color Dots", func(i int) {
	// 			e.GetMemoryMap().IntSetDHGRRender(e.GetMemIndex(), settings.VM_DOTTY)
	// 			cfg.SetI("video", "init.video.dhgrmode", int(settings.VM_DOTTY))
	// 		}).
	// 		AddCheck("renderCR", "Color Raster", func(i int) {
	// 			e.GetMemoryMap().IntSetDHGRRender(e.GetMemIndex(), settings.VM_FLAT)
	// 			cfg.SetI("video", "init.video.dhgrmode", int(settings.VM_FLAT))
	// 		}).
	// 		AddCheck("renderCV", "Color Voxels", func(i int) {
	// 			e.GetMemoryMap().IntSetDHGRRender(e.GetMemIndex(), settings.VM_VOXELS)
	// 			cfg.SetI("video", "init.video.dhgrmode", int(settings.VM_VOXELS))
	// 		}).
	// 		AddSep().
	// 		AddCheck("renderMD", "Mono Dots", func(i int) {
	// 			e.GetMemoryMap().IntSetDHGRRender(e.GetMemIndex(), settings.VM_MONO_DOTTY)
	// 			cfg.SetI("video", "init.video.dhgrmode", int(settings.VM_MONO_DOTTY))
	// 		}).
	// 		AddCheck("renderMR", "Mono Raster", func(i int) {
	// 			e.GetMemoryMap().IntSetDHGRRender(e.GetMemIndex(), settings.VM_MONO_FLAT)
	// 			cfg.SetI("video", "init.video.dhgrmode", int(settings.VM_MONO_FLAT))
	// 		}).
	// 		AddCheck("renderMV", "Mono Voxels", func(i int) {
	// 			e.GetMemoryMap().IntSetDHGRRender(e.GetMemIndex(), settings.VM_MONO_VOXELS)
	// 			cfg.SetI("video", "init.video.dhgrmode", int(settings.VM_MONO_VOXELS))
	// 		}).
	// 		AddSep().
	// 		AddMenu("highbit", "Enhanced mode").Title("Enhanced").
	// 		Before(func(m *Menu) {
	// 			for _, item := range m.Items {
	// 				switch item.Label {
	// 				case "auto":
	// 					item.Checked = settings.DHGRHighBit[e.GetMemIndex()] == settings.DHB_MIXED_AUTO
	// 				case "off":
	// 					item.Checked = settings.DHGRHighBit[e.GetMemIndex()] == settings.DHB_MIXED_OFF
	// 				case "on":
	// 					item.Checked = settings.DHGRHighBit[e.GetMemIndex()] == settings.DHB_MIXED_ON
	// 				}
	// 			}
	// 		}).
	// 		AddCheck("auto", "Autodetect", func(i int) {
	// 			settings.DHGRHighBit[e.GetMemIndex()] = settings.DHB_MIXED_AUTO
	// 			cfg.SetI("video", "init.video.dhgrhighbit", int(settings.DHB_MIXED_AUTO))
	// 		}).
	// 		AddCheck("off", "Off", func(i int) {
	// 			settings.DHGRHighBit[e.GetMemIndex()] = settings.DHB_MIXED_OFF
	// 			cfg.SetI("video", "init.video.dhgrhighbit", int(settings.DHB_MIXED_OFF))
	// 		}).
	// 		AddCheck("on", "On", func(i int) {
	// 			settings.DHGRHighBit[e.GetMemIndex()] = settings.DHB_MIXED_ON
	// 			cfg.SetI("video", "init.video.dhgrhighbit", int(settings.DHB_MIXED_ON))
	// 		}).
	// 		Parent().
	// 		Parent().
	// 		AddMenu("rendermenu", "SHR Render mode").
	// 		Before(
	// 			func(m *Menu) {
	// 				// prepare menu
	// 				for _, item := range m.Items {
	// 					switch item.Label {
	// 					case "renderCD":
	// 						item.Checked = e.GetMemoryMap().IntGetSHRRender(e.GetMemIndex()) == settings.VM_DOTTY
	// 					case "renderCR":
	// 						item.Checked = e.GetMemoryMap().IntGetSHRRender(e.GetMemIndex()) == settings.VM_FLAT
	// 					case "renderCV":
	// 						item.Checked = e.GetMemoryMap().IntGetSHRRender(e.GetMemIndex()) == settings.VM_VOXELS
	// 					}
	// 				}
	// 			},
	// 		).
	// 		AddCheck("renderCD", "Color Dots", func(i int) {
	// 			e.GetMemoryMap().IntSetSHRRender(e.GetMemIndex(), settings.VM_DOTTY)
	// 			cfg.SetI("video", "init.video.shrmode", int(settings.VM_DOTTY))
	// 		}).
	// 		AddCheck("renderCR", "Color Raster", func(i int) {
	// 			e.GetMemoryMap().IntSetSHRRender(e.GetMemIndex(), settings.VM_FLAT)
	// 			cfg.SetI("video", "init.video.shrmode", int(settings.VM_FLAT))
	// 		}).
	// 		AddCheck("renderCV", "Color Voxels", func(i int) {
	// 			e.GetMemoryMap().IntSetSHRRender(e.GetMemIndex(), settings.VM_VOXELS)
	// 			cfg.SetI("video", "init.video.shrmode", int(settings.VM_VOXELS))
	// 		}).
	// 		Parent()

	// }
	// fontList := settings.AuxFonts[e.GetMemIndex()]
	// if len(fontList) > 0 {
	// 	fm := zz.AddMenu("font", "Fonts").Title("Fonts")
	// 	for i, fontName := range fontList {
	// 		fm.Add(fmt.Sprintf("font%d", i), font.GetFontName(fontName), func(i int) {
	// 			log.Printf("Loading from file: %s", fontList[i])
	// 			f, err := font.LoadFromFile(fontList[i])
	// 			if err == nil {
	// 				settings.DefaultFont[e.GetMemIndex()] = f
	// 				settings.ForceTextVideoRefresh = true
	// 			}
	// 		},
	// 		)
	// 	}
	// }

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
