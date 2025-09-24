package main

import (
	"strings"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/octalyzer/backend"
	"paleotronic.com/octalyzer/clientperipherals"
	"paleotronic.com/octalyzer/ui"
	"paleotronic.com/utils"
)

func bToStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func GetConfig(index int, name string) (string, string, bool) {

	var value, kind string
	parts := strings.Split(name, "/")
	scope := parts[0]
	key := parts[1]

	e := backend.ProducerMain.GetInterpreter(index)
	mm := e.GetMemoryMap()

	if scope == "override" {
		if settings.IsSetBoolOverride(index, key) {
			value = "1"
		} else {
			value = "0"
		}
	} else {
		switch strings.ToLower(name) {
		case "video/current.mousemovecamera.enabled":
			value = bToStr(mouseMoveCamera)
		case "video/current.mousemovecamera.alternate":
			value = bToStr(mouseMoveCameraAlt)
		case "hardware/init.liverecording":
			value = bToStr(settings.AutoLiveRecording())
		case "hardware/init.disablefractionalrewind":
			value = bToStr(settings.DisableFractionalRewindSpeeds)
		case "system/current.suppressmenu":
			value = bToStr(settings.SuppressWindowedMenu)
		case "system/current.freezedir":
			value = files.GetDiskSaveDirectory(SelectedIndex)
		case "video/current.fullscreen":
			value = bToStr(!settings.Windowed)
		case "hardware/init.apple2.disk.nodskwoz":
			value = bToStr(settings.PreserveDSK)
		case "hardware/init.apple2.disk.nowarp":
			value = bToStr(settings.NoDiskWarp[index])
		case "hardware/init.printer.timeout":
			value = utils.IntToStr(settings.PrintToPDFTimeoutSec)
		case "hardware/current.cpu.model":
			value = apple2helpers.GetCPU(e).GetModel()
			kind = "float"
		case "hardware/current.cpu.warp":
			value = fmt.Sprintf("%.2f", apple2helpers.GetCPU(e).GetWarp())
			kind = "float"
		case "hardware/init.serial.mode":
			value = utils.IntToStr(int(settings.SSCCardMode[e.GetMemIndex()]))
		case "audio/init.master.volume":
			value = fmt.Sprintf("%.2f", settings.MixerVolume)
			kind = "float"
		case "audio/init.speaker.volume":
			value = fmt.Sprintf("%.2f", settings.SpeakerVolume[index])
			kind = "float"
		case "audio/init.mockingboard.psg0balance":
			value = fmt.Sprintf("%.2f", settings.MockingBoardPSG0Bal)
			kind = "float"
		case "audio/init.mockingboard.psg1balance":
			value = fmt.Sprintf("%.2f", settings.MockingBoardPSG1Bal)
			kind = "float"
		case "input/init.joystick.reversex":
			if settings.JoystickReverseX[index] {
				value = "1"
			} else {
				value = "0"
			}
		case "input/init.joystick.reversey":
			if settings.JoystickReverseY[index] {
				value = "1"
			} else {
				value = "0"
			}
		case "input/init.joystick.axis0":
			value = utils.IntToStr(mm.PaddleMap[index][0])
		case "input/init.joystick.axis1":
			value = utils.IntToStr(mm.PaddleMap[index][1])
		case "input/init.uppercase":
			value = bToStr(e.GetMemoryMap().IntGetUppercaseOnly(index))
		case "video/init.video.voxeldepth":
			value = utils.IntToStr(int(mm.IntGetVoxelDepth(index)))
		case "video/init.aspect":
			control := types.NewOrbitController(mm, index, -1)
			value = fmt.Sprintf("%.2f", control.GetAspect())
			kind = "float"
		case "video/current.rendermode":
			mode := GetVideoMode(SelectedIndex)
			value = "2"
			switch {
			case mode == "LOGR" || mode == "DLGR":
				v := RAM.IntGetGRRender(index)
				value = utils.IntToStr(int(v))
				kind = "int"
			case mode == "HGR":
				v := RAM.IntGetHGRRender(index)
				value = utils.IntToStr(int(v))
				kind = "int"
			case mode == "DHGR":
				v := RAM.IntGetDHGRRender(index)
				value = utils.IntToStr(int(v))
				kind = "int"
			case mode == "SHR1":
				v := RAM.IntGetSHRRender(index)
				value = utils.IntToStr(int(v))
				kind = "int"
			}
		case "video/init.video.tintmode":
			v := RAM.IntGetVideoTint(index)
			value = utils.IntToStr(int(v))
			kind = "int"
		case "input/init.mouse":
			v := settings.GetMouseMode()
			value = utils.IntToStr(int(v))
			kind = "int"
		case "video/init.video.scanlinedisable":
			value = bToStr(settings.DisableScanlines)
			kind = "int"
		case "video/init.video.scanline":
			value = fmt.Sprintf("%.2f", settings.ScanLineIntensity)
			kind = "float"
		case "video/init.video.dhgrhighbit":
			value = utils.IntToStr(int(settings.DHGRHighBit[index]))
			kind = "int"
		case "video/init.video.hgrmode":
			v := RAM.IntGetHGRRender(index)
			value = utils.IntToStr(int(v))
			kind = "int"
		case "video/init.video.grmode":
			v := RAM.IntGetGRRender(index)
			value = utils.IntToStr(int(v))
			kind = "int"
		case "video/init.video.dhgrmode":
			v := RAM.IntGetDHGRRender(index)
			value = utils.IntToStr(int(v))
			kind = "int"
		case "video/init.video.shrmode":
			v := RAM.IntGetSHRRender(index)
			value = utils.IntToStr(int(v))
			kind = "int"
		case "audio/init.master.mute":
			if clientperipherals.SPEAKER.Mixer.IsMuted() {
				value = "1"
			} else {
				value = "0"
			}
		default:
			return "", "", false
		}

	}

	return value, kind, true
}

func SetConfig(index int, name string, value string, persist bool) bool {

	e := backend.ProducerMain.GetInterpreter(index)
	mm := e.GetMemoryMap()
	cfg := ui.NewDefaultSettings(e)
	parts := strings.Split(name, "/")
	scope := parts[0]
	key := parts[1]
	kind := "int"

	switch strings.ToLower(name) {
	case "video/current.mousemovecamera.enabled":
		mouseMoveCamera = (value == "1")
	case "video/current.mousemovecamera.alternate":
		mouseMoveCameraAlt = (value == "1")
	case "video/init.video.scanlinedisable":
		settings.DisableScanlines = (value == "1")
		kind = "int"
	case "hardware/init.liverecording":
		settings.SetAutoLiveRecording((value == "1"))
	case "hardware/init.disablefractionalrewind":
		settings.DisableFractionalRewindSpeeds = (value == "1")
	case "system/current.suppressmenu":
		settings.SuppressWindowedMenu = (value == "1")
	case "video/current.fullscreen":
		settings.Windowed = (value == "0")
		if !settings.Windowed {
			win.GetGLFWWindow().Focus()
		}
	case "hardware/init.apple2.disk.nodskwoz":
		settings.PreserveDSK = (value == "1")
	case "hardware/init.apple2.disk.nowarp":
		settings.NoDiskWarp[index] = (value == "1")
	case "hardware/init.printer.timeout":
		settings.PrintToPDFTimeoutSec = utils.StrToInt(value)
	case "hardware/current.cpu.model":
		switch value {
		case "6502":
			apple2helpers.SwitchCPU(e, apple2helpers.New6502(e))
		case "65C02":
			apple2helpers.SwitchCPU(e, apple2helpers.New65C02(e))
		}
		kind = "float"
	case "hardware/current.cpu.warp":
		apple2helpers.GetCPU(e).SetWarpUser(utils.StrToFloat64(value))
		kind = "float"
	case "hardware/init.serial.mode":
		settings.SSCCardMode[index] = settings.SSCMode(utils.StrToInt(value))
	case "audio/init.master.volume":
		settings.MixerVolume = utils.StrToFloat64(value)
		kind = "float"
	case "audio/init.speaker.volume":
		settings.SpeakerVolume[index] = utils.StrToFloat64(value)
		kind = "float"
	case "audio/init.mockingboard.psg0balance":
		settings.MockingBoardPSG0Bal = utils.StrToFloat64(value)
		kind = "float"
	case "audio/init.mockingboard.psg1balance":
		settings.MockingBoardPSG1Bal = utils.StrToFloat64(value)
		kind = "float"
	case "audio/init.master.mute":
		clientperipherals.SPEAKER.Mixer.SetMute(value == "1")
	case "input/init.joystick.reversex":
		settings.JoystickReverseX[index] = (value == "1")
	case "input/init.joystick.reversey":
		settings.JoystickReverseY[index] = (value == "1")
	case "input/init.joystick.axis0":
		mm.PaddleMap[index][0] = utils.StrToInt(value)
	case "input/init.joystick.axis1":
		mm.PaddleMap[index][1] = utils.StrToInt(value)
	case "input/init.uppercase":
		e.GetMemoryMap().IntSetUppercaseOnly(index, value != "0")
	case "video/init.video.voxeldepth":
		mm.IntSetVoxelDepth(index, settings.VoxelDepth(utils.StrToInt(value)))
	case "video/init.aspect":
		aspect := utils.StrToFloat64(value)
		for i := -1; i < 9; i++ {
			control := types.NewOrbitController(mm, index, i)
			control.SetAspect(aspect)
		}
		kind = "float"
	case "video/current.rendermode": // live video mode
		mode := GetVideoMode(SelectedIndex)
		switch {
		case mode == "LOGR" || mode == "DLGR":
			RAM.IntSetGRRender(index, settings.VideoMode(utils.StrToInt(value)))
		case mode == "HGR":
			RAM.IntSetHGRRender(index, settings.VideoMode(utils.StrToInt(value)))
		case mode == "DHGR":
			RAM.IntSetDHGRRender(index, settings.VideoMode(utils.StrToInt(value)))
		case mode == "SHR1":
			RAM.IntSetSHRRender(index, settings.VideoMode(utils.StrToInt(value)))
		}
	case "input/init.mouse":
		settings.SetMouseMode(settings.MouseMode(utils.StrToInt(value)))
	case "video/init.video.dhgrhighbit":
		settings.DHGRHighBit[index] = settings.DHGRHighBitMode(utils.StrToInt(value))
	case "video/init.video.scanline":
		settings.ScanLineIntensity = utils.StrToFloat(value)
		kind = "float"
	case "video/init.video.tintmode":
		RAM.IntSetVideoTint(index, settings.VideoPaletteTint(utils.StrToInt(value)))
	case "video/init.video.hgrmode":
		RAM.IntSetHGRRender(index, settings.VideoMode(utils.StrToInt(value)))
	case "video/init.video.dhgrmode":
		RAM.IntSetDHGRRender(index, settings.VideoMode(utils.StrToInt(value)))
	case "video/init.video.shrmode":
		RAM.IntSetSHRRender(index, settings.VideoMode(utils.StrToInt(value)))
	case "video/init.video.grmode":
		RAM.IntSetGRRender(index, settings.VideoMode(utils.StrToInt(value)))
	default:
		return false
	}

	if persist {
		switch kind {
		case "int":
			cfg.SetI(scope, key, utils.StrToInt(value))
		case "float":
			cfg.SetF(scope, key, utils.StrToFloat64(value))
		case "string":
			cfg.SetS(scope, key, value)
		default:
			panic("unsupported config type: " + kind)
		}
		cfg.Finalize()
	}

	return true
}
