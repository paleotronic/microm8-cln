package plus

import (
	"strings"

	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
)

func GetList(dia interfaces.Dialecter) []string {
	out := []string(nil)
	f := dia.GetPlusFunctions()
	for k, v := range f {

		if v.IsHidden() {
			continue
		}

		params := []string(nil)
		if v.GetRaw() {
			for i, dt := range v.GetNamedDefaults() {

				params = append(params, v.GetNamedParams()[i]+": "+dt.Type.String())
			}
		} else {
			for _, dt := range v.FunctionParams() {
				params = append(params, dt.String())
			}
		}

		out = append(out, k+strings.Join(params, ", ")+"}")
	}
	return out
}

func RegisterFunctions(dia interfaces.Dialecter) {

	dia.AddPlusFunction("@render", "shr{", NewPlusRenderSHR(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@builtin", "forum{", NewPlusBuiltinForum(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@builtin", "chat{", NewPlusBuiltinChat(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@text", "font{", NewPlusTextFont(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@music", "edit{", NewPlusMicroTracker(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@music", "play{", NewPlusMicroTrackerPlayer(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@music", "pattern{", NewPlusMicroTrackerPattern(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@music", "pause{", NewPlusMicroTrackerPause(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@system", "log{", NewPlusLog(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "launch{", NewPlusLaunch(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@key", "type{", NewPlusKeyType(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@disk", "info{", NewPlusGetDisk(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@system", "monitor{", NewPlusMonitor(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "catalog{", NewPlusCatalog(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "hasnetwork{", NewPlusNetworkOK(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@mem", "range{", NewPlusMemRange(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@trigger", "def{", NewPlusTriggerDef(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@mem", "peek{", NewPlusMemPeek(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@mem", "poke{", NewPlusMemPoke(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@imagemap", "add{", NewPlusImageMapAdd(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@imagemap", "clear{", NewPlusImageMapClear(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@render", "hgr{", NewPlusRenderHGR(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@render", "dhgr{", NewPlusRenderDHGR(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@disk", "insert{", NewPlusDiskIIInsert(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@disk", "swap{", NewPlusDiskIISwap(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@color", "tint{", NewPlusPaletteTint(0, 0, *types.NewTokenList()))

	// key valyue
	dia.AddPlusFunction("@system", "setkey{", NewPlusSetSKV(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "getkey{", NewPlusGetSKV(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@user", "setkey{", NewPlusSetUKV(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@user", "getkey{", NewPlusGetUKV(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@image", "draw{", NewPlusPNG2HGR(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@image", "rect{", NewPlusPNGRect(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@draw", "image{", NewPlusPNG2HGRH(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@draw", "rect{", NewPlusPNGRectH(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@image", "colorspace{", NewPlusColorSpace(0, 0, *types.NewTokenList()))
	//dia.AddPlusFunction("@image", "splash{", NewPlusSplash(0, 0, *types.NewTokenList()))
	//dia.AddPlusFunction("@image", "backdrop{", NewPlusBackdrop(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@audio", "play{", NewPlusPlay(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@music", "notes{", NewPlusMusic(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@music", "instrument{", NewPlusInstrument(0, 0, *types.NewTokenList()))
	//dia.AddPlusFunction("@music", "sound{", NewPlusCustomSound(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@music", "noise{", NewPlusNoise(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@music", "tone{", NewPlusSound(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@music", "stop{", NewPlusAudioStop(0, 0, *types.NewTokenList()))
	//	dia.AddPlusFunction("@music", "pause{", NewPlusAudioPause(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@music", "chanselect{", NewPlusTrackerChanSelect(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@music", "speed{", NewPlusTrackerTempo(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@music", "resume{", NewPlusAudioResume(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@audio", "stream{", NewPlusBGMusic(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@audio", "stop{", NewPlusStreamStop(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@audio", "pause{", NewPlusStreamPause(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@audio", "resume{", NewPlusStreamResume(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@restalgia", "play{", NewPlusRestalgiaPlay(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@color", "bg{", NewPlusBGColor(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@color", "fg{", NewPlusFGColor(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@color", "background{", NewPlusCGColor(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "reset{", NewPlusCamReset(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "dolly{", NewPlusCamDolly(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "zoom{", NewPlusCamZoom(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "location{", NewPlusCamLoc(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "move{", NewPlusCamMove(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "pivpnt{", NewPlusCamPivPnt(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "rotate{", NewPlusCamAng(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "orbit{", NewPlusCamOrbit(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "select{", NewPlusCamSelect(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "view{", NewPlusCamView(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "shake{", NewPlusCamShake(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "params{", NewPlusCameraSave(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "aspect{", NewPlusCamAspect(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "updir{", NewPlusCamUpDir(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "rightdir{", NewPlusCamRightDir(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "viewdir{", NewPlusCamViewDir(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "noaspect{", NewPlusBlockPC(0, 0, *types.NewTokenList()))
	//dia.AddPlusFunction("@camera", "load{", NewPlusCameraLoad(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@camera", "pan{", NewPlusCamPan(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "spawn{", NewPlusSpawn(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@text", "echo{", NewPlusEcho(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@vm", "interpreter{", NewPlusVMLaunch(0, 0, *types.NewTokenList()))
	echo := NewPlusEcho(0, 0, *types.NewTokenList())
	echo.Hidden = true
	dia.AddPlusFunction("@system", "echo{", echo)
	dia.AddPlusFunction("@system", "nobreak{", NewPlusNoBreak(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "norestore{", NewPlusNoRestore(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "exit{", NewPlusExit(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "mousekeys{", NewPlusAltCase(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@mode", "video{", NewPlusVideoMode(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@mode", "font{", NewPlusTextMode(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@mode", "hgr{", NewPlusSwitchHGR(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "disableslot{", NewPlusDisableSlot(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "ignoreaudio{", NewPlusNoAudio(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "thaw{", NewPlusThaw(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@input", "uppercase{", NewPlusUpperCase(0, 0, *types.NewTokenList()))

	/* text drawing */
	dia.AddHiddenPlusFunction("@textdraw", "color{", NewPlusTextDrawColor(0, 0, *types.NewTokenList()))
	dia.AddHiddenPlusFunction("@textdraw", "pos{", NewPlusTextDrawPos(0, 0, *types.NewTokenList()))
	dia.AddHiddenPlusFunction("@textdraw", "width{", NewPlusTextDrawWidth(0, 0, *types.NewTokenList()))
	dia.AddHiddenPlusFunction("@textdraw", "height{", NewPlusTextDrawHeight(0, 0, *types.NewTokenList()))
	dia.AddHiddenPlusFunction("@textdraw", "inverse{", NewPlusTextDrawInverse(0, 0, *types.NewTokenList()))
	dia.AddHiddenPlusFunction("@textdraw", "print{", NewPlusTextDrawPrint(0, 0, *types.NewTokenList()))
	dia.AddHiddenPlusFunction("@textdraw", "font{", NewPlusTextDrawFont(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@typeset", "color{", NewPlusTextDrawColor(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@typeset", "pos{", NewPlusTextDrawPos(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@typeset", "width{", NewPlusTextDrawWidth(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@typeset", "height{", NewPlusTextDrawHeight(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@typeset", "inverse{", NewPlusTextDrawInverse(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@typeset", "print{", NewPlusTextDrawPrint(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@typeset", "font{", NewPlusTextDrawFont(0, 0, *types.NewTokenList()))

	/* added functions for fun */
	//dia.AddFunction( "time$", NewStandardFunctionTIMEDollar(0,0,nil) );
	dia.AddPlusFunction("@cpu", "throttle{", NewPlusCPUThrottle(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "uptime{", NewPlusUpTime(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "boottime{", NewPlusBootTime(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "pause{", NewPlusPause(0, 0, *types.NewTokenList()))
	//	dia.AddPlusFunction("system", "stack{", NewPlusStack(0, 0, *types.NewTokenList()))
	//	dia.AddPlusFunction("system", "publish{", NewPlusPublish(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "shim{", NewPlusShim(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "new{", NewPlusActivateSlot(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "switch{", NewPlusInputSwitch(0, 0, *types.NewTokenList()))
	paddle := NewPlusPaddleValue(0, 0, *types.NewTokenList())
	paddle.Hidden = true
	dia.AddPlusFunction("@system", "paddle{", paddle)
	dia.AddPlusFunction("@paddle", "value{", NewPlusPaddleValue(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@paddle", "button{", NewPlusPaddleButton(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "textbox{", NewPlusDrawBox(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@system", "feedback{", NewPlusFeedback(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@cpu", "switch{", NewPlusSwitchCPU(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@cpu", "zeropage{", NewPlusZeroPage(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@cpu", "prodos{", NewPlusProDOS(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@share", "sendmessage{", NewPlusSendMsg(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@share", "getmessages{", NewPlusGetMsgs(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@share", "connect{", NewPlusConnect(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@share", "getcontrol{", NewPlusGetControls(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@share", "getremotes{", NewPlusGetRemotes(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@share", "endremotes{", NewPlusEndRemotes(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@share", "controls{", NewPlusAllocControl(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@share", "transfer{", NewPlusTransfer(0, 0, *types.NewTokenList()))

	// project stuff
	dia.AddPlusFunction("@project", "new{", NewPlusProjectCreate(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@project", "use{", NewPlusProjectUse(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@project", "close{", NewPlusProjectClose(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@paddle", "swap{", NewPlusJoyToggle(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@system", "showmotd{", NewPlusDisplayMOTD(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "setmotd{", NewPlusSetMOTD(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@dos", "mkdir{", NewPlusMkDir(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "cd{", NewPlusCd(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "dir{", NewPlusDir(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "ls{", NewPlusDir(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "del{", NewPlusDelete(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "rm{", NewPlusDelete(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "copy{", NewPlusCopy(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "cp{", NewPlusCopy(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "grant{", NewPlusGrant(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "revoke{", NewPlusRevoke(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "meta{", NewPlusMetaMod(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "mount{", NewPlusMount(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@dos", "open{", NewPlusDOSOPEN(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "close{", NewPlusDOSCLOSE(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "read{", NewPlusDOSREAD(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "write{", NewPlusDOSWRITE(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "append{", NewPlusDOSAPPEND(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "chain{", NewPlusDOSCHAIN(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@dos", "paramcount{", NewPlusParamCount(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "param{", NewPlusParam(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@dos", "programdir{", NewPlusProgramDir(0, 0, *types.NewTokenList()))

	//dia.AddPlusFunction("@system", "debugger{", NewPlusDebugger(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@log", "debug{", NewPlusDebug(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@log", "mode{", NewPlusLogging(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@system", "nbi{", NewPlusInput(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@layer", "move{", NewPlusLayerPos(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@layer", "pos{", NewPlusLayerGlobalPos(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@window", "add{", NewPlusAddWindow(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@window", "use{", NewPlusUseWindow(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@cursor", "push{", NewPlusPushCursor(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@cursor", "pop{", NewPlusPopCursor(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@cursor", "hide{", NewPlusHideCursor(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@cursor", "show{", NewPlusShowCursor(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@color", "offset{", NewPlusPaletteOffset(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@color", "depth{", NewPlusPaletteDepth(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@color", "rgba{", NewPlusPaletteRGBA(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@color", "palette{", NewPlusPaletteSelect(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@color", "reset{", NewPlusPaletteReset(0, 0, *types.NewTokenList()))

	if !settings.VMRedirectDisable {
		dia.AddPlusFunction("@vm", "redirect{", NewPlusSlotSelect(0, 0, *types.NewTokenList()))
		dia.AddPlusFunction("@vm", "default{", NewPlusSlotDefault(0, 0, *types.NewTokenList()))
		dia.AddPlusFunction("@vm", "target{", NewPlusSlotTarget(0, 0, *types.NewTokenList()))
		dia.AddPlusFunction("@vm", "restart{", NewPlusReboot(0, 0, *types.NewTokenList()))
		dia.AddPlusFunction("@vm", "var{", NewPlusVarSlot(0, 0, *types.NewTokenList()))
		dia.AddPlusFunction("@vm", "enable{", NewPlusEnableSlot(0, 0, *types.NewTokenList()))
		dia.AddPlusFunction("@vm", "force{", NewPlusForceRenderSlot(0, 0, *types.NewTokenList()))
		dia.AddPlusFunction("@vm", "disable{", NewPlusDisableSlotV(0, 0, *types.NewTokenList()))
		dia.AddPlusFunction("@vm", "select{", NewPlusSelectSlot(0, 0, *types.NewTokenList()))
		dia.AddPlusFunction("@vm", "exec{", NewPlusSlotDo(0, 0, *types.NewTokenList()))
		dia.AddPlusFunction("@vm", "id{", NewPlusSlotID(0, 0, *types.NewTokenList()))
		dia.AddPlusFunction("@vm", "pause{", NewPlusSlotPause(0, 0, *types.NewTokenList()))
		dia.AddPlusFunction("@vm", "speaker{", NewPlusVMSpeaker(0, 0, *types.NewTokenList()))
		dia.AddPlusFunction("@vm", "restore{", NewPlusVMRestore(0, 0, *types.NewTokenList()))
	}

	dia.AddPlusFunction("@vm", "launchpak{", NewPlusVMLaunchPAK(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@gfx", "hgrpixel{", NewPlusPixelSize(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@user", "firstname{", NewPlusUserFirstname(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@user", "lastname{", NewPlusUserLastname(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@user", "gender{", NewPlusUserGender(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@user", "dob{", NewPlusUserDOB(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@user", "changepassword{", NewPlusUserDOB(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@user", "name{", NewPlusUserName(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@text", "color{", NewPlusTextColor(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@text", "size{", NewPlusTextMode(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@text", "reset{", NewPlusFontReset(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@text", "shade{", NewPlusShade(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@bug", "list{", NewPlusBugList(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@bug", "create{", NewPlusBugCreate(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@bug", "show{", NewPlusBugShow(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@bug", "close{", NewPlusBugClose(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@bug", "comment{", NewPlusBugComment(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@bug", "load{", NewPlusBugLoad(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@feature", "list{", NewPlusBugList(1, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@feature", "create{", NewPlusBugCreate(1, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@feature", "show{", NewPlusBugShow(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@feature", "close{", NewPlusBugClose(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@feature", "comment{", NewPlusBugComment(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@feature", "load{", NewPlusBugLoad(0, 0, *types.NewTokenList()))

	//dia.AddPlusFunction("@gfx", "cubegr{", NewPlusCubeGR(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@cube", "plot{", NewPlusCubePlot(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@cube", "line{", NewPlusCubeLine(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@video", "record{", NewPlusRecordStart(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@video", "stop{", NewPlusRecordStop(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@video", "play{", NewPlusRecordPlay(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@video", "slice{", NewPlusRecordSlice(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@system", "launchdebug{", NewPlusLaunchDebug(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@asm", "build{", NewPlusAssemble(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@light", "ambient{", NewPlusLightAmbient(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@light", "diffuse{", NewPlusLightDiffuse(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@db", "connect{", NewPlusAppDBConnect(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@db", "do{", NewPlusAppDBQuery(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@db", "count{", NewPlusAppDBResultCount(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@db", "fetch{", NewPlusAppDBFetch(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@db", "dosub{", NewPlusAppDBQuerySub(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@db", "prepare{", NewPlusAppDBPrepare(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@db", "unsub{", NewPlusAppDBUnsub(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@auth", "login{", NewPlusAuthLogin(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@auth", "register{", NewPlusAuthRegister(0, 0, *types.NewTokenList()))
	//~ dia.AddPlusFunction("auth", "login{", NewPlusAuthLogin(0, 0, *types.NewTokenList()))
	//~ dia.AddPlusFunction("auth", "register{", NewPlusAuthRegister(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@share", "command{", NewPlusShareCommand(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@share", "input{", NewPlusShareInput(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@color", "rotate{", NewPlusRotPal(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@string", "wrap{", NewPlusTextWrap(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@cpu", "txs{", NewPlusTXS(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@scan", "sequence{", NewPlusAnalyzeFindSeq(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@scan", "jumps{", NewPlusAnalyzeFindJump(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@mem", "lock{", NewPlusMemLock(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@screen", "contains{", NewPlusScreenScrape(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@screen", "read{", NewPlusScreenTextCapture(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@counter", "value{", NewPlusMemCounterVal(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@counter", "bump{", NewPlusMemCounterBump(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@counter", "set{", NewPlusMemCounterBumpSet(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@overlay", "filename{", NewPlusOverlayFilename(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@backdrop", "filename{", NewPlusBackdropFilename(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@backdrop", "opacity{", NewPlusBackdropOpacity(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@backdrop", "zoom{", NewPlusBackdropZoom(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@backdrop", "zrat{", NewPlusBackdropZoomFactor(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@backdrop", "camtrack{", NewPlusBackdropCamTrack(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@backdrop", "pos{", NewPlusBackdropPos(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@backdrop", "move{", NewPlusBackdropMove(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@backdrop", "reset{", NewPlusBackdropReset(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@color", "text{", NewPlusTextRGBA(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@draw", "line{", NewPlusDrawLine(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@draw", "arc{", NewPlusDrawArc(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@draw", "poly{", NewPlusDrawPoly(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@draw", "box{", NewPlusDrawRect(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@draw", "circle{", NewPlusDrawCircle(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@fill", "poly{", NewPlusDrawPolyF(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@fill", "box{", NewPlusDrawRectF(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@fill", "circle{", NewPlusDrawCircleF(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@render", "suspend{", NewPlusRenderSuspend(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@fill", "point{", NewPlusFillPoint(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@fill", "screen{", NewPlusFillScreen(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@mixer", "master{", NewPlusVolumeMaster(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@mixer", "speaker{", NewPlusVolumeSpeaker(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@mixer", "mute{", NewPlusVolumeMute(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@sprite", "reset{", NewPlusSpriteReset(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "on{", NewPlusSpriteOn(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "off{", NewPlusSpriteOff(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "xpos{", NewPlusSpriteXPos(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "ypos{", NewPlusSpriteYPos(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "rotate{", NewPlusSpriteRotate(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "flip{", NewPlusSpriteFlip(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "scale{", NewPlusSpriteScale(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "define{", NewPlusSpriteDefine(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "place{", NewPlusSpritePlace(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "test{", NewPlusSpriteTest(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "collision{", NewPlusSpriteCollision(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "bounds{", NewPlusSpriteBounds(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "stamp{", NewPlusSpriteStamp(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "unstamp{", NewPlusSpriteUnstamp(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "color{", NewPlusSpriteColor(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@sprite", "copy{", NewPlusSpriteCopy(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@mouse", "pos{", NewPlusMousePos(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@mouse", "select{", NewPlusMouseSelect(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@mouse", "button{", NewPlusMouseButton(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@text", "copy{", NewPlusTextCopy(0, 0, *types.NewTokenList()))
	//dia.AddPlusFunction("@text", "paste{", NewPlusTextPaste(0, 0, *types.NewTokenList()))

	for r := 'a'; r <= 'z'; r++ {
		dia.AddPlusFunction("@key", string(r)+"{", NewPlusKey(0, 0, vduconst.SHIFT_CTRL_A+(r-'a'), *types.NewTokenList()))
	}

	dia.AddPlusFunction("@key", "redirect{", NewPlusKeyRedirect(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@cpu", "speed{", NewPlusCPUSpeed(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@turtle", "exec{", NewPlusTurtleExec(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@turtle", "load{", NewPlusTurtleLoad(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@turtle", "last{", NewPlusTurtleLast(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@turtle", "location{", NewPlusTurtleLocation(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@turtle", "heading{", NewPlusTurtleHeading(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@turtle", "procedure{", NewPlusTurtleProc(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@turtle", "loadstate{", NewPlusTurtleLoadState(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@turtle", "savestate{", NewPlusTurtleSaveState(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@turtle", "semaphore{", NewPlusTurtleSemaphore(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@turtle", "pots{", NewPlusTurtlePots(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@turtle", "camera{", NewPlusTurtleCamera(0, 0, *types.NewTokenList()))

	// special control verbs
	dia.AddPlusFunction("@turtle", "stop{", NewPlusTurtleSendCommand(0, 0, *types.NewTokenList(), "stop"))
	dia.AddPlusFunction("@turtle", "pause{", NewPlusTurtleSendCommand(0, 0, *types.NewTokenList(), "suspend"))
	dia.AddPlusFunction("@turtle", "cont{", NewPlusTurtleSendCommand(0, 0, *types.NewTokenList(), "resume"))

	dia.AddPlusFunction("@string", "uppercase{", NewPlusStringUpper(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@string", "lowercase{", NewPlusStringLower(0, 0, *types.NewTokenList()))

	// zone commands
	dia.AddPlusFunction("@zone", "create{", NewPlusZoneCreate(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@zone", "rgba{", NewPlusZoneRGBA(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@zone", "depth{", NewPlusZoneDepth(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@zone", "offset{", NewPlusZoneOffset(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@zone", "reset{", NewPlusZoneReset(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@zone", "delete{", NewPlusZoneDelete(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@zone", "init{", NewPlusZoneInit(0, 0, *types.NewTokenList()))

	dia.AddPlusFunction("@cursor", "position{", NewPlusCursorPosition(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@text", "posfg{", NewPlusPositionFG(0, 0, *types.NewTokenList()))
	dia.AddPlusFunction("@text", "posbg{", NewPlusPositionBG(0, 0, *types.NewTokenList()))

}
