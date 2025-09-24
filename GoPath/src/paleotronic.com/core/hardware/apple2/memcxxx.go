package apple2

import (
	"runtime"
	"strings"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/debugger/debugtypes"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
)

type MemoryFlag int

const (
	MF_80STORE MemoryFlag = 1 << iota
	MF_ALTZP
	MF_AUXREAD
	MF_AUXWRITE
	MF_BANK2
	MF_HIGHRAM
	MF_HIRES
	MF_PAGE2
	MF_SLOTC3ROM
	MF_SLOTCXROM
	MF_WRITERAM
	MF_INTC8ROM
	MF_EXPC8ROM
	MF_FORCEREFRESH
	// mask pattern
	MF_IMAGEMASK = MF_80STORE | MF_ALTZP | MF_AUXREAD | MF_BANK2 | MF_HIGHRAM | MF_HIRES | MF_PAGE2 | MF_SLOTC3ROM | MF_SLOTCXROM | MF_EXPC8ROM | MF_INTC8ROM
	// default
	MF_DEFAULT = MF_BANK2 | MF_SLOTCXROM | MF_WRITERAM
)

func GetMemoryFlagMask(s string) (MemoryFlag, bool, bool) {
	switch s {
	case "80STORE":
		return MF_80STORE, true, false
	case "ALTZP":
		return MF_ALTZP, true, false
	case "AUXREAD":
		return MF_AUXREAD, true, false
	case "RAMRD":
		return MF_AUXREAD, true, false
	case "AUXWRITE":
		return MF_AUXWRITE, true, false
	case "RAMWRT":
		return MF_AUXWRITE, true, false
	case "BANK2":
		return MF_BANK2, true, false
	case "BSRBANK2":
		return MF_BANK2, true, false
	case "BSRREADRAM":
		return MF_HIGHRAM, true, false
	case "READRAM":
		return MF_HIGHRAM, true, false
	case "HIRES":
		return MF_HIRES, true, false
	case "PAGE2":
		return MF_PAGE2, true, false
	case "SLOTC3ROM":
		return MF_SLOTC3ROM, true, false
	case "SLOTCXROM":
		return MF_SLOTCXROM, true, false
	case "INTCXROM":
		return MF_SLOTCXROM, true, true
	case "WRITERAM":
		return MF_WRITERAM, true, false
	case "INTC8ROM":
		return MF_INTC8ROM, true, false
	case "EXPC8ROM":
		return MF_EXPC8ROM, true, false
	}
	return 0, false, false
}

func (m MemoryFlag) String() string {
	s := []string(nil)
	if m&MF_80STORE != 0 {
		s = append(s, "80STORE")
	}
	if m&MF_ALTZP != 0 {
		s = append(s, "ALTZP")
	}
	if m&MF_AUXREAD != 0 {
		s = append(s, "AUXREAD")
	}
	if m&MF_AUXWRITE != 0 {
		s = append(s, "AUXWRITE")
	}
	if m&MF_BANK2 != 0 {
		s = append(s, "BANK2")
	}
	if m&MF_HIGHRAM != 0 {
		s = append(s, "HIGHRAM")
	}
	if m&MF_HIRES != 0 {
		s = append(s, "HIRES")
	}
	if m&MF_PAGE2 != 0 {
		s = append(s, "PAGE2")
	}
	if m&MF_SLOTC3ROM != 0 {
		s = append(s, "SLOTC3ROM")
	}
	if m&MF_SLOTCXROM == 0 {
		s = append(s, "INTCXROM")
	}
	if m&MF_WRITERAM != 0 {
		s = append(s, "WRITERAM")
	}
	if m&MF_EXPC8ROM != 0 {
		s = append(s, "EXPC8ROM")
	}
	if m&MF_INTC8ROM != 0 {
		s = append(s, "INTC8ROM")
	}
	return strings.Join(s, "|")
}

type C8ROMType int

const (
	c8RomEmpty C8ROMType = iota
	c8RomInternal
	c8RomDevice
)

func (t C8ROMType) String() string {
	switch t {
	case c8RomEmpty:
		return "None"
	case c8RomDevice:
		return "Slot Device ROM"
	case c8RomInternal:
		return "Internal ROM"
	}
	return "Unknown"
}

type VideoFlag int

const (
	VF_TEXT VideoFlag = 1 << iota
	VF_MIXED
	VF_HIRES
	VF_PAGE2
	VF_80COL
	VF_80STORE
	VF_DHIRES
	VF_ALTCHAR
	VF_SHR_ENABLE
	VF_SHR_LINEAR
	// mask
	VF_MASK = VF_TEXT | VF_MIXED | VF_HIRES | VF_PAGE2 | VF_80COL | VF_80STORE | VF_DHIRES | VF_ALTCHAR | VF_SHR_ENABLE | VF_SHR_LINEAR
	// default
	VF_DEFAULT = VF_TEXT
)

func GetVideoFlagMask(s string) (VideoFlag, bool, bool) {
	switch s {
	case "80COL":
		return VF_80COL, true, false
	case "TEXT":
		return VF_TEXT, true, false
	case "MIXED":
		return VF_MIXED, true, false
	case "80STORE":
		return VF_80STORE, true, false
	case "DHIRES":
		return VF_DHIRES, true, false
	case "DBLRES":
		return VF_DHIRES, true, false
	case "ALTCHAR":
		return VF_ALTCHAR, true, false
	case "ALTCHARSET":
		return VF_ALTCHAR, true, false
	case "HIRES":
		return VF_HIRES, true, false
	case "PAGE2":
		return VF_PAGE2, true, false
	}
	return 0, false, false
}

func (m VideoFlag) String() string {
	s := []string(nil)
	if m&VF_80COL != 0 {
		s = append(s, "80COL")
	}
	if m&VF_TEXT != 0 {
		s = append(s, "TEXT")
	} else {
		s = append(s, "GRAPHICS")
	}
	if m&VF_MIXED != 0 {
		s = append(s, "MIXED")
	}
	if m&VF_80STORE != 0 {
		s = append(s, "80STORE")
	}
	if m&VF_DHIRES != 0 {
		s = append(s, "DHIRES")
	}
	if m&VF_ALTCHAR != 0 {
		s = append(s, "ALTCHAR")
	}
	if m&VF_HIRES != 0 {
		s = append(s, "HIRES")
	}
	if m&VF_PAGE2 != 0 {
		s = append(s, "PAGE2")
	}
	if m&VF_SHR_ENABLE != 0 {
		s = append(s, "SHR")
	}
	return strings.Join(s, "|")
}

const (
	SlotStartAddress  = 0xC100
	SlotEndAddress    = 0xC7FF
	C8ROMStartAddress = 0xC800
	C8ROMEndAddress   = 0xCFFF
	C8ROMSize         = 0x800
	MaxSlots          = 7
)

func (mr *Apple2IOChip) ToggleMemorySwitch(name string) {
	mask, ok, _ := GetMemoryFlagMask(name)
	if ok {
		mr.SetMemModeForce(mr.memmode ^ mask)
	}
}

func (mr *Apple2IOChip) ToggleVideoSwitch(name string) {
	mask, ok, _ := GetVideoFlagMask(name)
	if ok {
		mr.vidmode ^= mask
		mr.ConfigureVideo()
	}
}

func (mr *Apple2IOChip) SetExpRomState(v C8ROMType) {
	switch v {
	case c8RomDevice:
		mr.memmode &= (0xffffff ^ MF_INTC8ROM)
		mr.memmode |= MF_EXPC8ROM
	case c8RomEmpty:
		mr.memmode &= (0xffffff ^ (MF_INTC8ROM | MF_EXPC8ROM))
	case c8RomInternal:
		mr.memmode &= (0xffffff ^ MF_EXPC8ROM)
		mr.memmode |= MF_INTC8ROM
	}
}

func (mr *Apple2IOChip) SW_EXPC8ROM() bool {
	return (mr.memmode & MF_EXPC8ROM) != 0
}

func (mr *Apple2IOChip) SW_INTC8ROM() bool {
	return (mr.memmode & MF_INTC8ROM) != 0
}

func (mr *Apple2IOChip) SW_80STORE() bool {
	return !mr.DisableSW80 && (mr.memmode&MF_80STORE) != 0
}

func (mr *Apple2IOChip) SW_ALTZP() bool {
	return !mr.DisableSWAux && (mr.memmode&MF_ALTZP) != 0
}

func (mr *Apple2IOChip) SW_AUXREAD() bool {
	return !mr.DisableSWAux && (mr.memmode&MF_AUXREAD) != 0
}

func (mr *Apple2IOChip) SW_AUXWRITE() bool {
	return !mr.DisableSWAux && (mr.memmode&MF_AUXWRITE) != 0
}

func (mr *Apple2IOChip) SW_BANK2() bool {
	return (mr.memmode & MF_BANK2) != 0
}

func (mr *Apple2IOChip) SW_HIGHRAM() bool {
	return (mr.memmode & MF_HIGHRAM) != 0
}

func (mr *Apple2IOChip) SW_HIRES() bool {
	return (mr.memmode&MF_HIRES) != 0 || (mr.vidmode&VF_HIRES) != 0
}

func (mr *Apple2IOChip) SW_PAGE2() bool {
	return (mr.memmode & MF_PAGE2) != 0
}

func (mr *Apple2IOChip) SW_SLOTC3ROM() bool {
	return !mr.DisableSW80 && (mr.memmode&MF_SLOTC3ROM) != 0
}

func (mr *Apple2IOChip) SW_SLOTCXROM() bool {
	return (mr.memmode & MF_SLOTCXROM) != 0
}

func (mr *Apple2IOChip) SW_WRITERAM() bool {
	return (mr.memmode & MF_WRITERAM) != 0
}

func (mr *Apple2IOChip) SW_80COL() bool {
	return !mr.DisableSW80 && (mr.vidmode&VF_80COL) != 0
}

func (mr *Apple2IOChip) SW_TEXT() bool {
	return (mr.vidmode & VF_TEXT) != 0
}

func (mr *Apple2IOChip) SW_MIXED() bool {
	return (mr.vidmode & VF_MIXED) != 0
}

func (mr *Apple2IOChip) SW_ALTCHAR() bool {
	return !mr.DisableSWAlt && (mr.vidmode&VF_ALTCHAR) != 0
}

func (mr *Apple2IOChip) SW_DHIRES() bool {
	return !mr.DisableSWDbl && (mr.vidmode&VF_DHIRES) != 0
}

func (mr *Apple2IOChip) VSW_PAGE2() bool {
	return (mr.vidmode & VF_PAGE2) != 0
}

func (mr *Apple2IOChip) SW_SHR() bool {
	return (mr.vidmode & VF_SHR_ENABLE) != 0
}

func (mr *Apple2IOChip) SW_SHR_LINEAR() bool {
	return (mr.vidmode & VF_SHR_LINEAR) != 0
}

/*
MEMORY MANAGEMENT SOFT SWITCHES
 $C000   W       80STOREOFF      Allow page2 to switch video page1 page2
 $C001   W       80STOREON       Allow page2 to switch main & aux video memory
 $C002   W       RAMRDOFF        Read enable main memory from $0200-$BFFF
 $C003   W       RAMDRON         Read enable aux memory from $0200-$BFFF
 $C004   W       RAMWRTOFF       Write enable main memory from $0200-$BFFF
 $C005   W       RAMWRTON        Write enable aux memory from $0200-$BFFF
 $C006   W       INTCXROMOFF     Enable slot ROM from $C100-$CFFF
 $C007   W       INTCXROMON      Enable main ROM from $C100-$CFFF
 $C008   W       ALTZPOFF        Enable main memory from $0000-$01FF & avl BSR
 $C009   W       ALTZPON         Enable aux memory from $0000-$01FF & avl BSR
 $C00A   W       SLOTC3ROMOFF    Enable main ROM from $C300-$C3FF
 $C00B   W       SLOTC3ROMON     Enable slot ROM from $C300-$C3FF

VIDEO SOFT SWITCHES
 $C00C   W       80COLOFF        Turn off 80 column display
 $C00D   W       80COLON         Turn on 80 column display
 $C00E   W       ALTCHARSETOFF   Turn off alternate characters
 $C00F   W       ALTCHARSETON    Turn on alternate characters
 $C050   R/W     TEXTOFF         Select graphics mode
 $C051   R/W     TEXTON          Select text mode
 $C052   R/W     MIXEDOFF        Use full screen for graphics
 $C053   R/W     MIXEDON         Use graphics with 4 lines of text
 $C054   R/W     PAGE2OFF        Select panel display (or main video memory)
 $C055   R/W     PAGE2ON         Select page2 display (or aux video memory)
 $C056   R/W     HIRESOFF        Select low resolution graphics
 $C057   R/W     HIRESON         Select high resolution graphics

SOFT SWITCH STATUS FLAGS
 $C010   R7      AKD             1=key pressed   0=keys free    (clears strobe)
 $C011   R7      BSRBANK2        1=bank2 available    0=bank1 available
 $C012   R7      BSRREADRAM      1=BSR active for read   0=$D000-$FFFF active
 $C013   R7      RAMRD           0=main $0200-$BFFF active reads  1=aux active
 $C014   R7      RAMWRT          0=main $0200-$BFFF active writes 1=aux writes
 $C015   R7      INTCXROM        1=main $C100-$CFFF ROM active   0=slot active
 $C016   R7      ALTZP           1=aux $0000-$1FF+auxBSR    0=main available
 $C017   R7      SLOTC3ROM       1=slot $C3 ROM active   0=main $C3 ROM active
 $C018   R7      80STORE         1=page2 switches main/aux   0=page2 video
 $C019   R7      VERTBLANK       1=vertical retrace on   0=vertical retrace off
 $C01A   R7      TEXT            1=text mode is active   0=graphics mode active
 $C01B   R7      MIXED           1=mixed graphics & text    0=full screen
 $C01C   R7      PAGE2           1=video page2 selected or aux
 $C01D   R7      HIRES           1=high resolution graphics   0=low resolution
 $C01E   R7      ALTCHARSET      1=alt character set on    0=alt char set off
 $C01F   R7      80COL           1=80 col display on     0=80 col display off
*/

func (mr *Apple2IOChip) EnableMode(args ...string) {
	for _, mode := range args {
		//log2.Printf("mode flag: %s", mode)
		switch mode {
		case "text":
			mr.RelativeWrite(0x51, 0)
		case "graphics":
			mr.RelativeWrite(0x50, 0)
		case "80col":
			mr.RelativeWrite(0x0d, 0)
		case "40col":
			mr.RelativeWrite(0x0c, 0)
		case "splitscreen":
			mr.RelativeWrite(0x53, 0)
		case "fullscreen":
			mr.RelativeWrite(0x52, 0)
		case "page1":
			mr.RelativeWrite(0x54, 0)
		case "page2":
			mr.RelativeWrite(0x55, 0)
		case "lores":
			mr.RelativeWrite(0x56, 0)
		case "hires":
			mr.RelativeWrite(0x57, 0)
		case "double":
			mr.RelativeWrite(0x5e, 0)
		case "single":
			mr.RelativeWrite(0x5f, 0)
		default:
			panic("bad enable mode: " + mode)
		}
	}
}

func (mr *Apple2IOChip) GetMemorySwitchInfo() []debugtypes.SoftSwitchInfo {

	out := make([]debugtypes.SoftSwitchInfo, 0)

	// 80STORE
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "80STORE",
			StatusAddress:  0xc018,
			Enabled:        mr.SW_80STORE(),
			DisableAddress: 0xc000,
			EnableAddress:  0xc001,
			DisabledText:   "OFF",
			EnabledText:    "ON",
		},
	)

	// ALTZP
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "ALTZP",
			StatusAddress:  0xc016,
			Enabled:        mr.SW_ALTZP(),
			DisableAddress: 0xc008,
			EnableAddress:  0xc009,
			DisabledText:   "OFF",
			EnabledText:    "ON",
		},
	)

	// BSRREADRAM
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "BSRREADRAM",
			StatusAddress:  0xc012,
			Enabled:        mr.SW_HIGHRAM(),
			DisableAddress: 0xc008,
			EnableAddress:  0xc009,
			DisabledText:   "OFF",
			EnabledText:    "ON",
		},
	)

	// BSRBANK2
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "BSRBANK2",
			StatusAddress:  0xc011,
			Enabled:        mr.SW_BANK2(),
			DisableAddress: 0xc008,
			EnableAddress:  0xc009,
			DisabledText:   "OFF",
			EnabledText:    "ON",
		},
	)

	// RAMRD
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "RAMRD",
			StatusAddress:  0xc013,
			Enabled:        mr.SW_AUXREAD(),
			DisableAddress: 0xc008,
			EnableAddress:  0xc009,
			DisabledText:   "OFF",
			EnabledText:    "ON",
		},
	)

	// RAMWRT
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "RAMWRT",
			StatusAddress:  0xc014,
			Enabled:        mr.SW_AUXWRITE(),
			DisableAddress: 0xc008,
			EnableAddress:  0xc009,
			DisabledText:   "OFF",
			EnabledText:    "ON",
		},
	)

	// INTCXROM
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "INTCXROM",
			StatusAddress:  0xc015,
			Enabled:        !mr.SW_SLOTCXROM(),
			DisableAddress: 0xc008,
			EnableAddress:  0xc009,
			DisabledText:   "OFF",
			EnabledText:    "ON",
		},
	)

	// SLOTC3ROM
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "SLOTC3ROM",
			StatusAddress:  0xc017,
			Enabled:        mr.SW_SLOTC3ROM(),
			DisableAddress: 0xc008,
			EnableAddress:  0xc009,
			DisabledText:   "OFF",
			EnabledText:    "ON",
		},
	)

	return out
}

func (mr *Apple2IOChip) GetVideoSwitchInfo() []debugtypes.SoftSwitchInfo {

	out := make([]debugtypes.SoftSwitchInfo, 0)

	// AN3/DBLRES
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "DBLRES",
			StatusAddress:  0xc046,
			Enabled:        mr.SW_DHIRES(),
			DisableAddress: 0xc05e,
			EnableAddress:  0xc05f,
			DisabledText:   "OFF",
			EnabledText:    "ON",
		},
	)

	// 80COL
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "80COL",
			StatusAddress:  0xc01f,
			Enabled:        mr.SW_80COL(),
			DisableAddress: 0xc00c,
			EnableAddress:  0xc00d,
			DisabledText:   "OFF",
			EnabledText:    "ON",
		},
	)

	// ALTCHARSET
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "ALTCHARSET",
			StatusAddress:  0xc01e,
			Enabled:        mr.SW_ALTCHAR(),
			DisableAddress: 0xc00e,
			EnableAddress:  0xc00f,
			DisabledText:   "OFF",
			EnabledText:    "ON",
		},
	)

	// PAGE2
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "PAGE2",
			StatusAddress:  0xc01c,
			Enabled:        mr.ss.SoftSwitch_PAGE2,
			DisableAddress: 0xc054,
			EnableAddress:  0xc055,
			DisabledText:   "PAGE1",
			EnabledText:    "PAGE2",
		},
	)

	// MIXED
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "MIXED",
			StatusAddress:  0xc01b,
			Enabled:        mr.SW_MIXED(),
			DisableAddress: 0xc052,
			EnableAddress:  0xc053,
			DisabledText:   "FULLSCREEN",
			EnabledText:    "MIXED",
		},
	)

	// HIRES
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "HIRES",
			StatusAddress:  0xc01d,
			Enabled:        mr.ss.SoftSwitch_HIRES,
			DisableAddress: 0xc056,
			EnableAddress:  0xc057,
			DisabledText:   "LORES",
			EnabledText:    "HIRES",
		},
	)

	// TEXT
	out = append(
		out,
		debugtypes.SoftSwitchInfo{
			Name:           "TEXT",
			StatusAddress:  0xc01a,
			Enabled:        mr.SW_TEXT(),
			DisableAddress: 0xc050,
			EnableAddress:  0xc051,
			DisabledText:   "GRAPHICS",
			EnabledText:    "TEXT",
		},
	)

	return out

}

func (mr *Apple2IOChip) SetMemMode(newmode MemoryFlag) {
	mr.memmode = newmode
	mr.e.SetStatusText(newmode.String())
	mr.e.SetMemMode(int(newmode))
}

func (mr *Apple2IOChip) SetMemModeForce(newmode MemoryFlag) {
	lastmemmode := mr.memmode
	mr.memmode = newmode
	if lastmemmode != mr.memmode {
		mr.ConfigurePaging(false)
	}
}

func (mr *Apple2IOChip) SetVidModeForce(newmode VideoFlag) {
	lastvidmode := mr.vidmode
	mr.vidmode = newmode
	if lastvidmode != mr.vidmode {
		mr.ConfigureVideo()
	}
}

func (mr *Apple2IOChip) AddressSetPaging(addr int, value *uint64) {

	lastmemmode := mr.memmode

	if addr >= 0x80 && addr <= 0x8f {

		var writeram bool = ((addr & 1) == 1)

		mr.SetMemMode(mr.memmode & (0xFFFF ^ (MF_BANK2 | MF_HIGHRAM | MF_WRITERAM)))

		if writeram {
			mr.SetMemMode(mr.memmode | MF_WRITERAM)
		}
		if addr&8 == 0 {
			mr.SetMemMode(mr.memmode | MF_BANK2)
		}
		if ((addr & 2) >> 1) == (addr & 1) {
			mr.SetMemMode(mr.memmode | MF_HIGHRAM)
		}

	} else {
		switch addr {
		case 0x28:
			/*
				$C028 is a toggle for flipping rom banks.  But only for apple2c-* profiles
			*/
			if strings.HasPrefix(settings.SystemID[mr.e.GetMemIndex()], "apple2c-") {
				mr.UseHighROMs = !mr.UseHighROMs
				// we do this so we reconfigure the ram to take paging effect
				lastmemmode = MF_FORCEREFRESH
			}
		case 0x00:
			mr.SetMemMode(mr.memmode & (0xffff ^ MF_80STORE))
		case 0x01:
			mr.SetMemMode(mr.memmode | MF_80STORE)
		case 0x02:
			mr.SetMemMode(mr.memmode & (0xffff ^ MF_AUXREAD))
		case 0x03:
			mr.SetMemMode(mr.memmode | MF_AUXREAD)
		case 0x04:
			mr.SetMemMode(mr.memmode & (0xffff ^ MF_AUXWRITE))
		case 0x05:
			mr.SetMemMode(mr.memmode | MF_AUXWRITE)
		case 0x07:
			mr.SetMemMode(mr.memmode & (0xffff ^ MF_SLOTCXROM))
		case 0x06:
			mr.SetMemMode(mr.memmode | MF_SLOTCXROM)
		case 0x08:
			mr.SetMemMode(mr.memmode & (0xffff ^ MF_ALTZP))
		case 0x09:
			mr.SetMemMode(mr.memmode | MF_ALTZP)
		case 0x0A:
			mr.SetMemMode(mr.memmode & (0xffff ^ MF_SLOTC3ROM))
		case 0x0B:
			mr.SetMemMode(mr.memmode | MF_SLOTC3ROM)
		case 0x54:
			//if mr.SW_80STORE() {
			mr.SetMemMode(mr.memmode & (0xffff ^ MF_PAGE2))
		//	}
		case 0x55:
			if mr.SW_80STORE() {
				mr.SetMemMode(mr.memmode | MF_PAGE2)
			}
		case 0x56:
			mr.SetMemMode(mr.memmode & (0xffff ^ MF_HIRES))
		case 0x57:
			mr.SetMemMode(mr.memmode | MF_HIRES)
		}
	}

	if lastmemmode != mr.memmode {
		//cpu := apple2helpers.GetCPU(mr.e)
		//log2.Printf("Paging change: %s -> %s at or around $%.4x", lastmemmode.String(), mr.memmode.String(), cpu.PC)
		mr.ConfigurePaging(false)
	}

	// do we need to update the video mode?
	if (addr <= 1) || ((addr >= 0x54) && (addr <= 0x57)) {
		mr.AddressSetVideo(addr, value)
	}

}

func (mr *Apple2IOChip) ReadPagingState(addr int, value *uint64) {
	addr &= 0xFF
	var result bool
	switch addr {
	case 0x11:
		result = mr.SW_BANK2()
	case 0x12:
		result = mr.SW_HIGHRAM()
	case 0x13:
		result = mr.SW_AUXREAD()
	case 0x14:
		result = mr.SW_AUXWRITE()
	case 0x15:
		result = !mr.SW_SLOTCXROM()
	case 0x16:
		result = mr.SW_ALTZP()
	case 0x17:
		result = mr.SW_SLOTC3ROM()
	case 0x18:
		result = mr.SW_80STORE()
	case 0x1C:
		result = mr.SW_PAGE2()
	case 0x1D:
		result = mr.SW_HIRES()
	}
	*value = mr.GetRegKeyVal() & 0x7F
	if result {
		*value |= 0x80
	}
}

func (mr *Apple2IOChip) ReadFB() uint64 {
	return 0x00
}

func (mr *Apple2IOChip) MuxValue(condition bool, baseval uint64, trueval uint64, falseval uint64) uint64 {
	v := baseval
	if condition {
		v |= trueval
	} else {
		v |= falseval
	}
	return v
}

func (mr *Apple2IOChip) ReadVideoState(addr int, value *uint64) {
	addr &= 0xff
	if addr == 0x7f {
		*value = mr.MuxValue(mr.SW_DHIRES(), 0x00, 0x80, 0x00)
		*value = mr.FloatingBus(addr, value)
	} else {
		switch addr {
		case 0x1A:
			*value = mr.MuxValue(mr.SW_TEXT(), mr.GetRegKeyVal()&0x7f, 0x80, 0x00)
		case 0x1B:
			*value = mr.MuxValue(mr.SW_MIXED(), mr.GetRegKeyVal()&0x7f, 0x80, 0x00)
		case 0x1D:
			*value = mr.MuxValue(mr.SW_HIRES(), mr.GetRegKeyVal()&0x7f, 0x80, 0x00)
		case 0x1E:
			*value = mr.MuxValue(mr.SW_ALTCHAR(), mr.GetRegKeyVal()&0x7f, 0x80, 0x00)
		case 0x1F:
			*value = mr.MuxValue(mr.SW_80COL(), mr.GetRegKeyVal()&0x7f, 0x80, 0x00)
		}
	}
}

func (mr *Apple2IOChip) VideoCheckVbl(addr int, value *uint64) {
	//*value = mr.ReadVBLANK(mr.MappedRegion, addr)
	if !mr.IsVBlank() {
		*value = 128
	} else {
		*value = 0
	}
	//*value |= (mr.FloatingBus(addr, value) & 127)
}

func (mr *Apple2IOChip) clrVSwitch() {
	mr.VideoSW = ""
	settings.DHGRMode3Detected[mr.e.GetMemIndex()] = false
}

func (mr *Apple2IOChip) logVSwitch(sw string) {
	if mr.VideoSW != "" {
		mr.VideoSW += ","
	}
	mr.VideoSW += sw
	log.Printf("Video seq: %s", mr.VideoSW)
	if strings.HasSuffix(mr.VideoSW, "") {

	} else if strings.HasSuffix(mr.VideoSW, "80COL:on,AN3:off,AN3:on,AN3:off,AN3:on,AN3:off") {
		settings.DHGRMode3Detected[mr.e.GetMemIndex()] = false
	} else if strings.HasSuffix(mr.VideoSW, "80COL:on,AN3:off,AN3:on,AN3:off") {
		//log.Println("Mode3 DHGR Detected :)")
		settings.DHGRMode3Detected[mr.e.GetMemIndex()] = true
	} else if strings.HasSuffix(mr.VideoSW, "80COL:off,AN3:off,AN3:on,AN3:off,AN3:on,80COL:on,AN3:off") {
		settings.DHGRMode3Detected[mr.e.GetMemIndex()] = true
	} else if strings.HasSuffix(mr.VideoSW, "80COL:on,AN3:off,80COL:on,AN3:off,80COL:on,AN3:off,80COL:on,AN3:off,80COL:on") {
		settings.DHGRMode3Detected[mr.e.GetMemIndex()] = false

		v := mr.e.GetMemoryMap().IntGetDHGRRender(mr.e.GetMemIndex())
		switch v {
		case settings.VM_DOTTY:
			v = settings.VM_MONO_DOTTY
		case settings.VM_FLAT:
			v = settings.VM_MONO_FLAT
		case settings.VM_VOXELS:
			v = settings.VM_MONO_VOXELS
		}

		mr.e.GetMemoryMap().IntSetDHGRRender(mr.e.GetMemIndex(), v)
	}
}

func (mr *Apple2IOChip) AddressSetVideo(addr int, value *uint64) {

	mr.LastVidMode = mr.vidmode
	mr.NextVideoLatch = mr.GlobalCycles + 1 // FIXME: figure out correct timings

	addr &= 0xff

	lastvidmode := mr.vidmode

	//log2.Printf("video set mode of %x", addr+0xc000)

	switch addr {
	case 0x29:
		shrenable := *value&0x80 != 0
		linenable := *value&0x40 != 0
		//log2.Printf("Switch changes: shr %v, linear %v", shrenable, linenable)
		if shrenable {
			mr.vidmode |= VF_SHR_ENABLE
		} else {
			mr.vidmode &= (VF_MASK ^ VF_SHR_ENABLE)
		}
		if linenable {
			mr.vidmode |= VF_SHR_LINEAR
		} else {
			mr.vidmode &= (VF_MASK ^ VF_SHR_LINEAR)
		}
	case 0x00:
		mr.vidmode &= (VF_MASK ^ VF_80STORE)
	case 0x01:
		mr.vidmode |= VF_80STORE
	case 0x0C:
		log.Println("80COL Off")
		mr.vidmode &= (VF_MASK ^ VF_80COL)
		mr.COL80OFF(mr.MappedRegion, addr, *value)
		mr.logVSwitch("80COL:off")
	case 0x0D:
		log.Println("80COL On")
		mr.vidmode |= VF_80COL
		mr.COL80ON(mr.MappedRegion, addr, *value)
		mr.logVSwitch("80COL:on")
	case 0x0E:
		mr.vidmode &= (VF_MASK ^ VF_ALTCHAR)
		mr.ALTCHARSETOFF(mr.MappedRegion, addr, *value)
	case 0x0F:
		mr.vidmode |= VF_ALTCHAR
		mr.ALTCHARSETON(mr.MappedRegion, addr, *value)
	case 0x50:
		log.Println("Graphics")
		mr.vidmode &= (VF_MASK ^ VF_TEXT)
	case 0x51:
		log.Println("Text")
		mr.vidmode |= VF_TEXT
		mr.vidmode &= (VF_MASK ^ VF_SHR_ENABLE)
	case 0x52:
		log.Println("Fullscreen")
		mr.vidmode &= (VF_MASK ^ VF_MIXED)
	case 0x53:
		log.Println("Mixed")
		mr.vidmode |= VF_MIXED
	case 0x54:
		if mr.memmode&MF_80STORE == 0 {
			mr.vidmode &= (VF_MASK ^ VF_PAGE2)
		}
	case 0x55:
		if mr.memmode&MF_80STORE == 0 {
			mr.vidmode |= VF_PAGE2
		}
	case 0x56:
		log.Println("Lores")
		mr.vidmode &= (VF_MASK ^ VF_HIRES)
	case 0x57:
		log.Println("Hires")
		mr.vidmode |= VF_HIRES
	case 0x5F:
		log.Println("AN3 On")
		mr.logVSwitch("AN3:on")
		mr.vidmode &= (VF_MASK ^ VF_DHIRES)
	case 0x5E:
		log.Println("AN3 Off")
		mr.logVSwitch("AN3:off")
		mr.vidmode |= VF_DHIRES
	}

	// if mr.SW_80STORE() {
	// 	mr.vidmode &= (VF_MASK ^ VF_PAGE2)
	// }

	if lastvidmode != mr.vidmode {
		mr.ConfigureVideo()
	}

	mr.e.SetVideoStatusText(mr.vidmode.String())
	mr.e.SetVidMode(int(mr.vidmode))
}

func (mr *Apple2IOChip) HasC8Rom(slot uint64) bool {
	mm := mr.e.GetMemoryMap() // cheese
	mbm := mm.BlockMapper[mr.e.GetMemIndex()]
	v := mbm.Get(fmt.Sprintf("slotexp%d.rom", slot)) != nil
	//fmt.Printf("HasC8ROM(%d) = %v\n", slot, v)
	return v
}

func (mr *Apple2IOChip) HasCard(slot uint64) bool {
	return mr.cards[int(slot)] != nil
}

func (mr *Apple2IOChip) TestNoSlotClock(addr int) bool {
	ah := addr >> 8
	return (((!mr.SW_SLOTCXROM() || !mr.SW_SLOTC3ROM()) && (ah == 0xC3)) || // Internal ROM at [$C100-CFFF or $C300-C3FF] && AddrHi == $C3
		(!mr.SW_SLOTCXROM() && (ah == 0xC8))) // Internal ROM at [$C100-CFFF]               && AddrHi == $C8
}

func (mr *Apple2IOChip) AddressRead_Cxxx(addr int, value *uint64) {

	mm := mr.e.GetMemoryMap()
	mbm := mm.BlockMapper[mr.e.GetMemIndex()]

	if addr == 0xCFFF {
		mr.SetRegIOSelect(0)
		mr.SetRegIOInternalROM(0)
		mr.SetRegPeripheralROMSlot(0)

		if mr.SW_SLOTCXROM() {
			for bank := 0xc8; bank < 0xd0; bank++ {
				mbm.PageSetREAD(bank, nil)
			}
			mr.SetRegC8ROMType(c8RomEmpty)
		}
	}

	IOStrobe := 0

	if mr.SW_SLOTCXROM() {

		//fmt.Println("SW_SLOTCXROM")

		if addr >= SlotStartAddress && addr <= SlotEndAddress {
			slot := uint64((addr >> 8) & 0x7)
			if slot != 3 {
				if mr.HasC8Rom(slot) {
					mr.SetRegIOSelect(mr.GetRegIOSelect() | 1<<slot)
				}
			} else {
				if mr.SW_SLOTC3ROM() && mr.HasC8Rom(slot) {
					mr.SetRegIOSelect(mr.GetRegIOSelect() | 1<<slot)
				} else if !mr.SW_SLOTC3ROM() {
					mr.SetRegIOInternalROM(1)
				}
			}
		} else if addr >= C8ROMStartAddress && addr <= C8ROMEndAddress {
			IOStrobe = 1
		}

		//log.Printf("IOSelect = %b, IOStrobe = %d", mr.GetRegIOSelect(), IOStrobe)

		if mr.GetRegIOSelect() != 0 && IOStrobe != 0 && mr.GetRegIOInternalROM() == 0 {

			//fmt.Printf("RegIOSelect=%d, IOStrobe=%d\n", mr.GetRegIOSelect(), IOStrobe)

			// Turn on peripheral expansion ROM
			var slot uint64
			for slot = 1; slot < MaxSlots; slot++ {
				if mr.GetRegIOSelect()&(1<<slot) != 0 {
					//remainingSelected := mr.GetRegIOSelect() & (0xff ^ (1 << slot))
					//~ if remainingSelected != 0 {
					//~ panic("IOSelect issue")
					//~ }
					break
				}
			}

			// slot should be set
			if mr.HasC8Rom(slot) && mr.GetRegPeripheralROMSlot() != slot {
				//fmt.Printf("Turning on peripheral ROM in slot %d\n", slot)
				cpu := apple2helpers.GetCPU(mr.e)
				log.Printf("Turning on C8 Device rom at %.4x", cpu.PC)
				exprom := mbm.Get(fmt.Sprintf("slotexp%d.rom", slot))
				exprom.SetState("r")
				for bank := 0xc8; bank < 0xd0; bank++ {
					mbm.PageSetREAD(bank, exprom)
				}
				mr.SetRegC8ROMType(c8RomDevice)
				mr.SetRegPeripheralROMSlot(slot)
			}

		} else if mr.GetRegIOInternalROM() != 0 && IOStrobe != 0 && mr.GetRegC8ROMType() != c8RomInternal {
			rname := "rom.intcxrom"
			if strings.HasPrefix(settings.SystemID[mr.e.GetMemIndex()], "apple2c-") && mr.UseHighROMs {
				rname += "-alt"
			}
			intcxrom := mbm.Get(rname)
			for bank := 0xc8; bank < 0xd0; bank++ {
				mbm.PageSetREAD(bank, intcxrom)
			}
			mr.SetRegC8ROMType(c8RomInternal)
			mr.SetRegPeripheralROMSlot(0)
		}

	}

	// Handle switching internal ROM back on
	if !mr.SW_SLOTCXROM() {

		//fmt.Println("!SW_SLOTCXROM")

		if addr >= SlotStartAddress && addr <= SlotEndAddress {
			mr.SetRegIOInternalROM(1)
		} else if addr >= C8ROMStartAddress && addr <= C8ROMEndAddress {
			IOStrobe = 1
		}
		if !mr.SW_SLOTCXROM() && mr.GetRegIOInternalROM() != 0 && IOStrobe != 0 && mr.GetRegC8ROMType() != c8RomInternal {
			rname := "rom.intcxrom"
			if strings.HasPrefix(settings.SystemID[mr.e.GetMemIndex()], "apple2c") && mr.UseHighROMs && mbm.Get("rom.intcxrom-alt") != nil {
				rname += "-alt"
			}
			intcxrom := mbm.Get(rname)
			for bank := 0xc8; bank < 0xd0; bank++ {
				mbm.PageSetREAD(bank, intcxrom)
			}
			mr.SetRegC8ROMType(c8RomInternal)
			mr.SetRegPeripheralROMSlot(0)
			//fmt.Println("Turning internal C8 rom back on")
		}
	}

	if addr >= SlotStartAddress && addr <= SlotEndAddress {
		slot := uint64((addr >> 8) & 0xf)
		if mr.SW_SLOTCXROM() && !(!mr.SW_SLOTC3ROM() && slot == 3) && !mr.HasCard(slot) {
			*value = 0xA0
			*value = mr.FloatingBus(addr, value)
			return
		}
	}

	if mr.GetRegC8ROMType() == c8RomEmpty && addr >= C8ROMStartAddress {
		*value = 0x00
		return
	}

	// read from memory
	if addr == 0xc800 {
		//fmt.Printf("Block read for %d\n", addr)
	}
	mbm.Do(addr, memory.MA_READ, value)

}

func (mr *Apple2IOChip) AddressRead_C00x(addr int, value *uint64) {
	mr.ReadKeyData(addr&0xff, value)
}

func (mr *Apple2IOChip) AddressWrite_C00x(addr int, value *uint64) {
	if (addr & 0xf) <= 0xb {
		mr.AddressSetPaging(addr, value)
	} else {
		mr.AddressSetVideo(addr, value)
	}
}

func (mr *Apple2IOChip) ReadKeyData(addr int, value *uint64) {
	if settings.PasteWarp {
		cpu := apple2helpers.GetCPU(mr.e)
		if len(mr.e.GetPasteBuffer().Runes) > 0 && !settings.Pasting[mr.e.GetMemIndex()] {
			//log2.Printf("*** start pasting warp")
			settings.Pasting[mr.e.GetMemIndex()] = true
			cpu.SetWarp(8)
		} else if len(mr.e.GetPasteBuffer().Runes) == 0 && settings.Pasting[mr.e.GetMemIndex()] {
			//log2.Printf("*** end pasting warp")
			settings.Pasting[mr.e.GetMemIndex()] = false
			cpu.SetWarp(1)
		}
	}
	if x := mr.MappedRegion.Global.KeyBufferGetLatestNoRedirect(mr.e.GetMemIndex()); x != 0 {
		switch x {
		case vduconst.DELETE:
			x = 127
		case vduconst.CSR_DOWN:
			x = 10
		case vduconst.CSR_UP:
			x = 11
		case vduconst.CSR_LEFT:
			x = 8
		case vduconst.CSR_RIGHT:
			x = 21
		}
		mr.SetRegKeyVal(x | 128)
	}
	*value = mr.GetRegKeyVal()
}

func (mr *Apple2IOChip) ClearKeyStrobe(addr int, value *uint64) {
	//mr.SetRegKeyVal(mr.GetRegKeyVal() & 0x7f)
	mr.ClearKeyStrobeR(addr, value)
}

func (mr *Apple2IOChip) ClearKeyStrobeR(addr int, value *uint64) {

	if x := mr.MappedRegion.Global.KeyBufferGetLatest(mr.e.GetMemIndex()); x != 0 {
		switch x {
		case vduconst.CSR_DOWN:
			x = 10
		case vduconst.CSR_UP:
			x = 11
		case vduconst.CSR_LEFT:
			x = 8
		case vduconst.CSR_RIGHT:
			x = 21
		}
		mr.SetRegKeyVal(x | 128)
	}

	*value = mr.GetRegKeyVal()
	mr.SetRegKeyVal(mr.GetRegKeyVal() & 0x7f)
}

func (mr *Apple2IOChip) AddressAnnunciator(addr int, value *uint64) {
	// TODO: Implement this
}

func (mr *Apple2IOChip) AddressRead_C01x(addr int, value *uint64) {
	switch addr & 0xf {
	case 0x0:
		mr.ClearKeyStrobeR(addr, value)
	case 0x1:
		mr.ReadPagingState(addr, value)
	case 0x2:
		mr.ReadPagingState(addr, value)
	case 0x3:
		mr.ReadPagingState(addr, value)
	case 0x4:
		mr.ReadPagingState(addr, value)
	case 0x5:
		mr.ReadPagingState(addr, value)
	case 0x6:
		mr.ReadPagingState(addr, value)
	case 0x7:
		mr.ReadPagingState(addr, value)
	case 0x8:
		mr.ReadPagingState(addr, value)
	case 0x9:
		mr.VideoCheckVbl(addr, value)
	case 0xA:
		mr.ReadVideoState(addr, value)
	case 0xB:
		mr.ReadVideoState(addr, value)
	case 0xC:
		mr.ReadPagingState(addr, value)
	case 0xD:
		mr.ReadPagingState(addr, value)
	case 0xE:
		mr.ReadVideoState(addr, value)
	case 0xF:
		mr.ReadVideoState(addr, value)
	}
}

func (mr *Apple2IOChip) AddressWrite_C01x(addr int, value *uint64) {
	mr.ClearKeyStrobe(addr, value)
}

func (mr *Apple2IOChip) AddressRead_C02x(addr int, value *uint64) {
	switch addr & 0xf {
	case 0x0:
		mr.cassette.ToggleSpeaker(false)
	}
	*value = mr.FloatingBus(addr, value)
}

func (mr *Apple2IOChip) AddressWrite_C02x(addr int, value *uint64) {
	//
	switch addr & 0xf {
	case 0x8:
		mr.AddressSetPaging(addr, value)
	case 0x9:
		mr.AddressSetVideo(addr, value)
	}
}

func (mr *Apple2IOChip) AddressRead_C03x(addr int, value *uint64) {
	// speaker
	mr.speaker.ToggleSpeaker(false)

	*value = mr.FloatingBus(addr, value)
}

func (mr *Apple2IOChip) SetFBRange(addr int, size int) {
	mr.FBAddr = addr
	mr.FBSize = size
}

func (t *Apple2IOChip) Between(v, lo, hi uint64) bool {
	return ((v >= lo) && (v <= hi))
}

func (t *Apple2IOChip) PokeToAsciiApple(v uint64, usealt bool) int {
	highbit := v & 1024

	v = v & 1023

	if t.Between(v, 0, 31) {
		return int((64 + (v % 32)) | highbit)
	}

	if t.Between(v, 32, 63) {
		return int((32 + (v % 32)) | highbit)
	}

	if t.Between(v, 64, 95) {
		if usealt {
			return int((128 + (v % 32)) | highbit)
		} else {
			return int((64 + (v % 32)) | highbit)
		}
	}

	if t.Between(v, 96, 127) {
		if usealt {
			return int((96 + (v % 32)) | highbit)
		} else {
			return int((32 + (v % 32)) | highbit)
		}
	}

	if t.Between(v, 128, 159) {
		return int((64 + (v % 32)) | highbit)
	}

	if t.Between(v, 160, 191) {
		return int((32 + (v % 32)) | highbit)
	}

	if t.Between(v, 192, 223) {
		return int((64 + (v % 32)) | highbit)
	}

	if t.Between(v, 224, 255) {
		return int((96 + (v % 32)) | highbit)
	}

	return int(v | highbit)
}

func (t *Apple2IOChip) PokeToAttribute(v uint64, usealt bool) types.VideoAttribute {

	v = v & 1023

	va := types.VA_INVERSE
	if (v & 64) > 0 {
		if usealt {
			if v < 96 {
				va = types.VA_NORMAL
			} else {
				va = types.VA_INVERSE
			}
		} else {
			va = types.VA_BLINK
		}
	}
	if (v & 128) > 0 {
		va = types.VA_NORMAL
	}
	if (v & 256) > 0 {
		va = types.VA_NORMAL
	}
	return va
}

func (mr *Apple2IOChip) FloatingBus(addr int, value *uint64) uint64 {
	// Simulate floatingbus reads
	// if mr.FBSize == 0 {
	// 	return *value
	// }
	// idx := rand.Intn(mr.FBSize)
	//
	//modCycles := float64(mr.LastScanHPOS)
	//sd := mr.ScanData[mr.LastScanLine][mr.LastScanSegment]
	//if sd == nil || len(sd.DataLo) == 0 {
	//	return /*(*value & 0x80)*/ *value
	//}
	//
	//if modCycles >= 20 {
	//	modCycles -= 20
	//}
	//
	//idx := int((modCycles / 20) * float64(len(sd.DataLo)))
	//if idx >= len(sd.DataLo) {
	//	idx = len(sd.DataLo) - 1
	//}
	//
	////log.Printf("FB: Addr: %.4x, Index: %.4x", mr.FBAddr, mr.FBSize)
	//
	//// if mr.FBAddr != 0x400 && mr.FBAddr != 0x800 {
	//// 	return (*value & 0x80) | (mr.e.GetMemoryMap().ReadInterpreterMemory(
	//// 		mr.e.GetMemIndex(),
	//// 		mr.FBAddr+idx,
	//// 	) & 0x7f)
	//// }
	//
	//// y := scanLine / 8
	//// x := modCycles
	//// if modCycles >= 40 {
	//// 	x = 0
	//// }
	//// idx = ((y % 8) * 128) + ((y / 8) * 40) + x
	//
	//// cell := mr.e.GetMemoryMap().ReadInterpreterMemory(
	//// 	mr.e.GetMemIndex(),
	//// 	mr.FBAddr+idx,
	//// ) & 0xff
	//
	//return /*(*value & 0x80) |*/ uint64(sd.DataLo[idx])

	return uint64(mr.UnifiedFrame.FloatingBus)

}

func (mr *Apple2IOChip) AddressWrite_C03x(addr int, value *uint64) {
	// speaker
	mr.speaker.ToggleSpeaker(true)
}

func (mr *Apple2IOChip) AddressRead_C04x(addr int, value *uint64) {
	//
	*value = mr.FloatingBus(addr, value)
	if addr == 0x4f {
		*value = 0x2e
	}
}

func (mr *Apple2IOChip) AddressWrite_C04x(addr int, value *uint64) {
	//
}

func (mr *Apple2IOChip) AddressRead_C05x(addr int, value *uint64) {
	//
	switch addr & 0xf {
	case 0x0:
		mr.AddressSetVideo(addr, value)
		*value = 0x80
		return
	case 0x1:
		mr.AddressSetVideo(addr, value)
	case 0x2:
		mr.AddressSetVideo(addr, value)
	case 0x3:
		mr.AddressSetVideo(addr, value)
	case 0x4:
		mr.AddressSetPaging(addr, value)
	case 0x5:
		mr.AddressSetPaging(addr, value)
	case 0x6:
		mr.AddressSetPaging(addr, value)
	case 0x7:
		mr.AddressSetPaging(addr, value)
	case 0x8:
		mr.AddressAnnunciator(addr, value)
	case 0x9:
		mr.AddressAnnunciator(addr, value)
	case 0xA:
		mr.AddressAnnunciator(addr, value)
	case 0xB:
		mr.AddressAnnunciator(addr, value)
	case 0xC:
		mr.AddressAnnunciator(addr, value)
	case 0xD:
		mr.AddressAnnunciator(addr, value)
	case 0xE:
		mr.AddressSetVideo(addr, value)
	case 0xF:
		mr.AddressSetVideo(addr, value)
	}

	*value = mr.FloatingBus(addr, value)
}

func (mr *Apple2IOChip) AddressWrite_C05x(addr int, value *uint64) {
	//
	switch addr & 0xf {
	case 0x0:
		mr.AddressSetVideo(addr, value)
	case 0x1:
		mr.AddressSetVideo(addr, value)
	case 0x2:
		mr.AddressSetVideo(addr, value)
	case 0x3:
		mr.AddressSetVideo(addr, value)
	case 0x4:
		mr.AddressSetPaging(addr, value)
	case 0x5:
		mr.AddressSetPaging(addr, value)
	case 0x6:
		mr.AddressSetPaging(addr, value)
		mr.AddressSetVideo(addr, value)
	case 0x7:
		mr.AddressSetPaging(addr, value)
		mr.AddressSetVideo(addr, value)
	case 0x8:
		mr.AddressAnnunciator(addr, value)
	case 0x9:
		mr.AddressAnnunciator(addr, value)
	case 0xA:
		mr.AddressAnnunciator(addr, value)
	case 0xB:
		mr.AddressAnnunciator(addr, value)
	case 0xC:
		mr.AddressAnnunciator(addr, value)
	case 0xD:
		mr.AddressAnnunciator(addr, value)
	case 0xE:
		mr.AddressSetVideo(addr, value)
	case 0xF:
		mr.AddressSetVideo(addr, value)
	}
}

func (mr *Apple2IOChip) TapeRead(addr int, value *uint64) {
	// stub

	*value = mr.FloatingBus(addr, value)
	if mr.IsTapeAttached() {
		bit := mr.Tape.PullSample()
		if bit {
			*value = *value | 128
		} else {
			*value = *value & 127
		}
	}

}

func (mr *Apple2IOChip) AddressRead_C06x(addr int, value *uint64) {
	//
	switch addr & 0xf {
	case 0x0:
		mr.TapeRead(addr, value)
	case 0x1:
		*value = mr.ReadPaddleButton0(mr.MappedRegion, addr)
	case 0x2:
		*value = mr.ReadPaddleButton1(mr.MappedRegion, addr)
	case 0x3:
		*value = mr.ReadPaddleButton3(mr.MappedRegion, addr)
	case 0x4:
		*value = mr.PaddleXRead(mr.MappedRegion, addr)
	case 0x5:
		*value = mr.PaddleXRead(mr.MappedRegion, addr)
	case 0x6:
		*value = mr.PaddleXRead(mr.MappedRegion, addr)
	case 0x7:
		*value = mr.PaddleXRead(mr.MappedRegion, addr)
	case 0xb:
		cpu := apple2helpers.GetCPU(mr.e)
		ok, _ := cpu.HasUserWarp()
		if ok {
			*value = 128
		} else {
			*value = 0
		}
		//log2.Printf("returning fastchip enable status code = %d", *value)
	case 0xd:
		cpu := apple2helpers.GetCPU(mr.e)
		ok, warp := cpu.HasUserWarp()
		if ok {
			*value = uint64(FC2EWarpCodesReverse[warp])
		} else {
			*value = 0x00
		}
		//log2.Printf("returning fastchip speed mode code = %d", *value)
	default:
		*value = mr.FloatingBus(addr, value)
	}
}

var FC2EWarpCodes = map[int]float64{
	0:  1.0,
	1:  0.2,
	2:  0.3,
	3:  0.4,
	4:  0.5,
	5:  0.6,
	6:  0.7,
	7:  0.8,
	8:  0.9,
	9:  1.0,
	10: 1.1,
	11: 1.2,
	12: 1.3,
	13: 1.4,
	14: 1.5,
	15: 1.6,
	16: 1.7,
	17: 1.8,
	18: 1.9,
	19: 2.0,
	20: 2.1,
	21: 2.2,
	22: 2.3,
	23: 2.5,
	24: 2.6,
	25: 2.7,
	26: 2.9,
	27: 3.1,
	28: 3.3,
	29: 3.5,
	30: 3.8,
	31: 4.1,
	32: 4.55,
	33: 5.0,
	34: 5.5,
	35: 6.2,
	36: 7.1,
	37: 8.3,
	38: 10.0,
	39: 12.5,
	40: 16.6,
}

var FC2EWarpCodesReverse = map[float64]int{
	1.0:  0,
	0.2:  1,
	0.3:  2,
	0.4:  3,
	0.5:  4,
	0.6:  5,
	0.7:  6,
	0.8:  7,
	0.9:  8,
	1.1:  10,
	1.2:  11,
	1.3:  12,
	1.4:  13,
	1.5:  14,
	1.6:  15,
	1.7:  16,
	1.8:  17,
	1.9:  18,
	2.0:  19,
	2.1:  20,
	2.2:  21,
	2.3:  22,
	2.5:  23,
	2.6:  24,
	2.7:  25,
	2.9:  26,
	3.1:  27,
	3.3:  28,
	3.5:  29,
	3.8:  30,
	4.1:  31,
	4.55: 32,
	5.0:  33,
	5.5:  34,
	6.2:  35,
	7.1:  36,
	8.3:  37,
	10.0: 38,
	12.5: 39,
	16.6: 40,
}

func (mr *Apple2IOChip) AddressWrite_C06x(addr int, value *uint64) {
	// Do nothing - we don't care about the Pravets8A

	switch addr & 0xff {
	case 0x6b:
		// enable fast chip
		//log2.Printf("fast chip //e fake mode enable")
	case 0x6d:
		// set speed
		cpu := apple2helpers.GetCPU(mr.e)
		i := int(*value % 41)
		//log2.Printf("fast chip speed requested code = %d, speed = %.2f", i, FC2EWarpCodes[i])
		if !settings.UserWarpOverride[mr.e.GetMemIndex()] {
			cpu.SetWarpUser(FC2EWarpCodes[i])
		} else {
			//log2.Printf("ignoring fastchip request due to user requested speed")
		}

	}
}

func (mr *Apple2IOChip) AddressRead_C07x(addr int, value *uint64) {
	//
	switch addr & 0xf {
	case 0x0:
		*value = mr.TriggerPaddles(mr.MappedRegion, addr)
	case 0xf:
		mr.ReadVideoState(addr, value)
	default:
		*value = mr.FloatingBus(addr, value)
	}
}

func (mr *Apple2IOChip) AddressWrite_C07x(addr int, value *uint64) {
	//
	switch addr & 0xf {
	case 0x0:
		*value = mr.TriggerPaddles(mr.MappedRegion, addr)
	}
}

func (mr *Apple2IOChip) RelativeWrite(offset int, value uint64) {
	if offset >= mr.Size {
		return // ignore write outside our bounds
	}

	//mr.Data.Write(offset, value)

	//log2.Printf("Custom handler for offset 0x%.2x -> $c0%.2x", value, offset)

	custom, ok := mr.writehandlers[offset]
	if ok {
		if !mr.ExecuteActions(offset, &value, custom) {
			return
		}
	}

	// switch here
	switch (offset & 0xff) / 16 {
	case 0x0:
		mr.AddressWrite_C00x(offset, &value)
	case 0x1:
		mr.AddressWrite_C01x(offset, &value)
	case 0x2:
		mr.AddressWrite_C02x(offset, &value)
	case 0x3:
		mr.AddressWrite_C03x(offset, &value)
	case 0x4:
		mr.AddressWrite_C04x(offset, &value)
	case 0x5:
		mr.AddressWrite_C05x(offset, &value)
	case 0x6:
		mr.AddressWrite_C06x(offset, &value)
	case 0x7:
		mr.AddressWrite_C07x(offset, &value)
	case 0x8:
		mr.AddressSetPaging(offset, &value)
	default:
		if offset >= 0x80 && offset <= 0xff {
			slot := (offset - 0x80) / 16
			card := mr.cards[slot]
			if card != nil {
				card.HandleIO(offset&0xf, &value, IOT_WRITE)
			}
		}
	}
}

/* RelativeRead handles a read within this regions address space */
func (mr *Apple2IOChip) RelativeRead(offset int) uint64 {
	if offset >= mr.Size {
		return 0 // ignore read outside our bounds
	}

	custom, ok := mr.readhandlers[offset]
	if ok {
		var value uint64
		//log.Printf("Custom handler for offset 0x%.2x -> %s", offset, custom.Name)
		if !mr.ExecuteActions(offset, &value, custom) {
			return value
		}
	}

	//cpu := apple2helpers.GetCPU(mr.e)
	var value uint64

	// switch here
	switch (offset & 0xff) / 16 {
	case 0x0:
		mr.AddressRead_C00x(offset, &value)
		return value
	case 0x1:
		mr.AddressRead_C01x(offset, &value)
		return value
	case 0x2:
		mr.AddressRead_C02x(offset, &value)
		return value
	case 0x3:
		mr.AddressRead_C03x(offset, &value)
		return value
	case 0x4:
		mr.AddressRead_C04x(offset, &value)
		return value
	case 0x5:
		mr.AddressRead_C05x(offset, &value)
		return value
	case 0x6:
		mr.AddressRead_C06x(offset, &value)
		return value
	case 0x7:
		mr.AddressRead_C07x(offset, &value)
		return value
	case 0x8:
		mr.AddressSetPaging(offset, &value)
		return value
	default:
		if offset >= 0x80 && offset <= 0xff {
			slot := (offset - 0x80) / 16
			value := uint64(0)
			card := mr.cards[slot]

			//log.Printf("Apple2IO slot access slot = %d, register = 0x%x\n", slot, offset & 0xf )

			if card == nil {
				return mr.Data.Read(offset)
			} else {
				card.HandleIO(offset&0xf, &value, IOT_READ)
				return value
			}
		}
	}

	return mr.Data.Read(offset)
}

func (mr *Apple2IOChip) ConfigureVideo() {

	// store softswitches
	// TODO: Add toggles for 80COL / ALTCHAR
	ent := mr.e
	//	of := mr.ss
	nf := apple2helpers.SoftSwitchConfig{
		SoftSwitch_80COL:      mr.SW_80COL(),
		SoftSwitch_ALTCHARSET: mr.SW_ALTCHAR(),
		SoftSwitch_GRAPHICS:   !mr.SW_TEXT(),
		SoftSwitch_HIRES:      mr.SW_HIRES(),
		SoftSwitch_PAGE2:      mr.VSW_PAGE2(),
		SoftSwitch_DoubleRes:  mr.SW_DHIRES() && mr.SW_80COL(),
		SoftSwitch_MIXED:      mr.SW_MIXED(),
		SoftSwitch_SHR:        mr.SW_SHR(),
		SoftSwitch_SHR_LINEAR: mr.SW_SHR_LINEAR(),
	}

	if !nf.SoftSwitch_SHR && mr.ss.SoftSwitch_SHR {
		runtime.GC() // hacky force free for SHR
	}
	//log2.Printf("Softswitches: %+v", nf)
	//log2.Printf("Using config: %+v", nf)

	// 1 - TEXT Layer
	txt := apple2helpers.GETHUD(ent, "TEXT")
	if txt == nil {
		panic("Expected layer id TEXT not found")
	}
	txt2 := apple2helpers.GETHUD(ent, "TXT2")
	if txt2 == nil {
		panic("Expected layer id TXT2 not found")
	}
	//log2.Printf("mixed text = %v", nf.SoftSwitch_MIXED)
	switch {
	case nf.SoftSwitch_SHR:
		txt.SetActive(false)
		txt2.SetActive(false)
	case nf.SoftSwitch_PAGE2 == false && nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false:
		txt.SetActive(false)
		txt.SetBounds(80, 48, 80, 48)
		txt2.SetActive(false)
		txt2.SetBounds(80, 48, 80, 48)
	case nf.SoftSwitch_PAGE2 == false && nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true:
		txt.SetBounds(0, 40, 79, 47)
		txt.SetActive(true)
		txt.SetDirty(true)
		mr.SetFBRange(0x400, 0x3ff)
		//log2.Printf("txt active split")
		txt2.SetActive(false)
		txt2.SetBounds(80, 48, 80, 48)
	case nf.SoftSwitch_PAGE2 == false && nf.SoftSwitch_GRAPHICS == false:
		txt.SetActive(true)
		mr.SetFBRange(0x400, 0x3ff)
		txt.SetBounds(0, 0, 79, 47)
		txt2.SetActive(false)
		txt2.SetBounds(80, 48, 80, 48)
	// ---
	case nf.SoftSwitch_PAGE2 == true && nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false:
		txt2.SetActive(false)
		txt2.SetBounds(80, 48, 80, 48)
		txt.SetActive(false)
		txt.SetBounds(80, 48, 80, 48)
		//log2.Printf("txt inactive 2")
	case nf.SoftSwitch_PAGE2 == true && nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true:
		txt2.SetActive(true)
		mr.SetFBRange(0x800, 0x3ff)
		txt2.SetBounds(0, 40, 79, 47)
		txt.SetActive(false)
		txt.SetBounds(80, 48, 80, 48)
		//log2.Printf("txt inactive 3")
	case nf.SoftSwitch_PAGE2 == true && nf.SoftSwitch_GRAPHICS == false:
		txt2.SetActive(true)
		mr.SetFBRange(0x800, 0x3ff)
		txt2.SetBounds(0, 0, 79, 47)
		txt.SetActive(false)
		txt.SetBounds(80, 48, 80, 48)
		//log2.Printf("txt inactive 4")
	}

	txt.SetDirty(true)
	txt.Control.FullRefresh()
	txt2.Control.FullRefresh()

	// 2 - LOGR Layer
	gr := apple2helpers.GETGFX(ent, "LOGR")
	if gr == nil {
		panic("Expected layer id LOGR not found")
	}
	gr2 := apple2helpers.GETGFX(ent, "LGR2")
	if !nf.SoftSwitch_DoubleRes {
		switch {
		case nf.SoftSwitch_SHR && nf.SoftSwitch_GRAPHICS:
			gr.SetActive(false)
			gr2.SetActive(false)
		case nf.SoftSwitch_PAGE2 == false && nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == false:
			//txt.SetActive(false)
			//log2.Printf("txt inactive 5")
			gr.SetActive(true)
			mr.SetFBRange(0x400, 0x3ff)
			gr.SetBounds(0, 0, 39, 47)
			if gr2 != nil {
				gr2.SetActive(false)
			}
		case nf.SoftSwitch_PAGE2 == false && nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == false:
			gr.SetActive(true)
			mr.SetFBRange(0x400, 0x3ff)
			gr.SetBounds(0, 0, 39, 39)
			if gr2 != nil {
				gr2.SetActive(false)
			}
		case nf.SoftSwitch_PAGE2 == true && nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == false:
			// page2, full
			if gr2 != nil {
				gr2.SetActive(true)
			}
			mr.SetFBRange(0x800, 0x3ff)
			if gr2 != nil {
				gr2.SetBounds(0, 0, 39, 47)
			}
			gr.SetActive(false)
		case nf.SoftSwitch_PAGE2 == true && nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == false:
			// page2, mixed
			if gr2 != nil {
				gr2.SetActive(true)
			}
			mr.SetFBRange(0x800, 0x3ff)
			if gr2 != nil {
				gr2.SetBounds(0, 0, 39, 39)
			}
			gr.SetActive(false)
		default:
			gr.SetActive(false)
			if gr2 != nil {
				gr2.SetActive(false)
			}
		}
	} else {
		gr.SetActive(false)
		if gr2 != nil {
			gr2.SetActive(false)
		}
	}

	// 2 - DLGR Layer
	gr = apple2helpers.GETGFX(ent, "DLGR")
	if gr == nil {
		panic("Expected layer id DLGR not found")
	}
	gr2 = apple2helpers.GETGFX(ent, "DLG2")
	// if gr2 == nil {
	// 	panic("Expected layer id DLG2 not found")
	// }
	if nf.SoftSwitch_DoubleRes {
		switch {
		case nf.SoftSwitch_SHR:
			gr.SetActive(false)
			if gr2 != nil {
				gr2.SetActive(false)
			}
		case nf.SoftSwitch_PAGE2 == false && nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == false:
			//txt.SetActive(false)
			//log2.Printf("txt inactive 6")
			//log2.Println("dlgr page 1")
			gr.SetActive(true)
			mr.SetFBRange(0x400, 0x3ff)
			gr.SetBounds(0, 0, 79, 47)
			if gr2 != nil {
				gr2.SetActive(false)
			}
		case nf.SoftSwitch_PAGE2 == false && nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == false:
			//log2.Println("dlgr page 1")
			gr.SetActive(true)
			mr.SetFBRange(0x400, 0x3ff)
			gr.SetBounds(0, 0, 79, 39)
			if gr2 != nil {
				gr2.SetActive(false)
			}
		case nf.SoftSwitch_PAGE2 == true && nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == false:
			//txt.SetActive(false)
			//log2.Printf("txt inactive 6")
			//log2.Println("dlgr page 2")
			if gr2 != nil {
				gr2.SetActive(true)
			}
			mr.SetFBRange(0x800, 0x3ff)
			if gr2 != nil {
				gr2.SetBounds(0, 0, 79, 47)
			}
			gr.SetActive(false)
		case nf.SoftSwitch_PAGE2 == true && nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == false:
			//log2.Println("dlgr page 2")
			if gr2 != nil {
				gr2.SetActive(true)
			}
			mr.SetFBRange(0x800, 0x3ff)
			if gr2 != nil {
				gr2.SetBounds(0, 0, 79, 39)
			}
			gr.SetActive(false)
		default:
			gr.SetActive(false)
			if gr2 != nil {
				gr2.SetActive(false)
			}
		}
	} else {
		gr.SetActive(false)
		if gr2 != nil {
			gr2.SetActive(false)
		}
	}

	// 3 - HGR1 Layer
	gr = apple2helpers.GETGFX(ent, "HGR1")
	if gr == nil {
		panic("Expected layer id HGR1 not found")
	}
	if !nf.SoftSwitch_DoubleRes {
		switch {
		case nf.SoftSwitch_SHR:
			gr.SetActive(false)
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == false:
			//txt.SetActive(false)
			//log2.Printf("txt inactive 7")
			gr.SetActive(true)
			mr.SetFBRange(0x2000, 0x2000)
			mr.e.SetCurrentPage("HGR1")
			gr.SetBounds(0, 0, 279, 191)
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == false:
			gr.SetActive(true)
			mr.SetFBRange(0x2000, 0x2000)
			mr.e.SetCurrentPage("HGR1")
			gr.SetBounds(0, 0, 279, 159)
		default:
			gr.SetActive(false)
		}
		//		fmt.Println(gr.GetBoundsRect())
	} else {
		gr.SetActive(false)
	}

	// 4 - HGR2 Layer
	gr = apple2helpers.GETGFX(ent, "HGR2")
	if gr == nil {
		panic("Expected layer id HGR2 not found")
	}
	if !nf.SoftSwitch_DoubleRes {
		switch {
		case nf.SoftSwitch_SHR:
			gr.SetActive(false)
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == true:
			//txt.SetActive(false)
			//log2.Printf("txt inactive 8")
			gr.SetActive(true)
			mr.SetFBRange(0x4000, 0x2000)
			mr.e.SetCurrentPage("HGR2")
			gr.SetBounds(0, 0, 279, 191)
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == true:
			gr.SetActive(true)
			mr.SetFBRange(0x4000, 0x2000)
			mr.e.SetCurrentPage("HGR2")
			gr.SetBounds(0, 0, 279, 159)
		default:
			gr.SetActive(false)
		}
	} else {
		gr.SetActive(false)
	}

	// 3 - DHR1 Layer
	gr = apple2helpers.GETGFX(ent, "DHR1")
	if gr == nil {
		panic("Expected layer id DHR1 not found")
	}
	if nf.SoftSwitch_DoubleRes {
		switch {
		case nf.SoftSwitch_SHR:
			gr.SetActive(false)
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == false:
			//txt.SetActive(false)
			//log2.Printf("txt inactive 9")
			gr.SetActive(true)
			mr.SetFBRange(0x2000, 0x2000)
			gr.SetBounds(0, 0, 559, 191)
			mr.e.SetCurrentPage("DHR1")
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == false:
			//txt.SetActive(true)
			//txt.SetBounds(0,40,)
			gr.SetActive(true)
			mr.SetFBRange(0x2000, 0x2000)
			mr.e.SetCurrentPage("DHR1")
			gr.SetBounds(0, 0, 559, 159)
		default:
			gr.SetActive(false)
		}
	} else {
		gr.SetActive(false)
	}

	// 3 - DHR1 Layer
	gr = apple2helpers.GETGFX(ent, "DHR2")
	if gr == nil {
		panic("Expected layer id DHR2 not found")
	}
	if nf.SoftSwitch_DoubleRes {
		switch {
		case nf.SoftSwitch_SHR:
			gr.SetActive(false)
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == true:
			//txt.SetActive(false)
			//log2.Printf("txt inactive 10")
			gr.SetActive(true)
			mr.e.SetCurrentPage("DHR2")
			mr.SetFBRange(0x4000, 0x2000)
			gr.SetBounds(0, 0, 559, 191)
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == true:
			gr.SetActive(true)
			mr.SetFBRange(0x4000, 0x2000)
			mr.e.SetCurrentPage("DHR2")
			gr.SetBounds(0, 0, 559, 159)
		default:
			gr.SetActive(false)
		}
	} else {
		gr.SetActive(false)
	}

	gr = apple2helpers.GETGFX(ent, "SHR1")
	if gr != nil {
		gr.SetActive(nf.SoftSwitch_SHR)
	}

	mr.ss = nf

	servicebus.SendServiceBusMessage(
		mr.e.GetMemIndex(),
		servicebus.VideoSoftSwitchState,
		mr.GetVideoSwitchInfo(),
	)

}

func (mr *Apple2IOChip) MemReset() {
	mm := mr.e.GetMemoryMap()
	index := mr.e.GetMemIndex()
	membase := mm.MEMBASE(index)
	auxbase := mm.MEMBASE(index) + 0x10000
	for i := 0; i < 0x10000; i++ {
		mm.Data[index][membase+i] = 0x0
		mm.Data[index][auxbase+i] = 0x0
	}

	// fix upper memory space
	mr.SetRegIOSelect(0)
	mr.SetRegIOInternalROM(0)
	mr.SetRegC8ROMType(c8RomEmpty)
	mr.SetRegPeripheralROMSlot(0)
	mr.clrVSwitch() // clear state

	mr.memmode = MF_DEFAULT
	mr.ConfigurePaging(true)

	// for i := 0; i < 0x10000; i++ {
	// 	if i%4 < 2 {
	// 		mm.Data[membase+i] = mm.Data[membase+i]&0xffffffffffffff00 | 0xff
	// 		mm.Data[auxbase+i] = mm.Data[auxbase+i]&0xffffffffffffff00 | 0xff
	// 	} else {
	// 		mm.Data[membase+i] = mm.Data[membase+i] & 0xffffffffffffff00
	// 		mm.Data[auxbase+i] = mm.Data[auxbase+i] & 0xffffffffffffff00
	// 	}
	// }

	// time.Sleep(50 * time.Millisecond)

}

func (mr *Apple2IOChip) ConfigurePaging(initialize bool) {

	mm := mr.e.GetMemoryMap()
	mbm := mm.BlockMapper[mr.e.GetMemIndex()]

	memmain := mbm.Get("main.all")
	memaux := mbm.Get("aux.all")
	zpemu := mbm.Get("apple2iozeropage")
	//	mockingboard := mbm.Get("apple2mockingboard")
	ioregion := mbm.Get("apple2iochip")

	languagecard := mbm.Get("main.languagecard")
	lcbank2 := mbm.Get("main.lcbank2")
	auxlanguagecard := mbm.Get("aux.languagecard")
	auxlcbank2 := mbm.Get("aux.lcbank2")

	var intcxrom, applerom, monitorrom *memory.MemoryBlock

	if strings.HasPrefix(settings.SystemID[mr.e.GetMemIndex()], "apple2c") && mr.UseHighROMs && mbm.Get("rom.intcxrom-alt") != nil {
		intcxrom = mbm.Get("rom.intcxrom-alt")
		applerom = mbm.Get("rom.applesoft-alt")
		monitorrom = mbm.Get("rom.monitor-alt")
	} else {
		intcxrom = mbm.Get("rom.intcxrom")
		applerom = mbm.Get("rom.applesoft")
		monitorrom = mbm.Get("rom.monitor")
	}

	var slotroms [8]*memory.MemoryBlock
	var slotfw [8]*memory.MemoryBlock
	for slot := 1; slot < 8; slot++ {
		slotroms[slot] = mbm.Get(fmt.Sprintf("slot%d.rom", slot))
		slotfw[slot] = mbm.Get(fmt.Sprintf("slot%d.firmware", slot))
	}

	// Initial state
	if initialize {
		for bank := 0; bank < 0xC0; bank++ {
			mbm.PageSetREAD(bank, memmain)
			mbm.PageSetWRITE(bank, memmain)
		}
		for bank := 0xC0; bank < 0x100; bank++ {
			mbm.PageSetREAD(bank, nil)
			mbm.PageSetWRITE(bank, nil)
		}
		if strings.HasPrefix(settings.SystemID[mr.e.GetMemIndex()], "apple2c") {
			// IIc start with C8 internal mapped in
			mr.SetRegC8ROMType(c8RomInternal)
		} else {
			mr.SetRegC8ROMType(c8RomEmpty)
		}
	}

	// zeropage and stack 0x0000 - 0x01ff
	for bank := 0; bank < 0x2; bank++ {
		if mr.SW_ALTZP() {
			mbm.PageSetREAD(bank, memaux)
			mbm.PageSetWRITE(bank, memaux)
		} else {
			if bank == 0 {
				mbm.PageSetREAD(bank, zpemu)
				mbm.PageSetWRITE(bank, zpemu)
			} else {
				mbm.PageSetREAD(bank, memmain)
				mbm.PageSetWRITE(bank, memmain)
			}
		}
	}

	// main memory - 0x200 - 0xbfff
	for bank := 0x2; bank < 0xC0; bank++ {
		// reads
		if mr.SW_AUXREAD() {
			mbm.PageSetREAD(bank, memaux)
		} else {
			mbm.PageSetREAD(bank, memmain)
		}
		// writes
		if mr.SW_AUXWRITE() {
			mbm.PageSetWRITE(bank, memaux)
		} else {
			mbm.PageSetWRITE(bank, memmain)
		}
	}

	// map io - always 0xC000 - 0xC0FF
	mbm.PageSetREAD(0xC0, ioregion)
	mbm.PageSetWRITE(0xC0, ioregion)

	// peripheral slot roms 0xc100 - 0xc7ff -- or intcxrom
	for bank := 0xC1; bank < 0xC8; bank++ {
		if bank == 0xC3 {
			if mr.SW_SLOTC3ROM() && mr.SW_SLOTCXROM() {
				mbm.PageSetREAD(bank, slotroms[3])
			} else {
				mbm.PageSetREAD(bank, intcxrom)
			}
		} else {
			if mr.SW_SLOTCXROM() {
				if slotfw[bank-0xc0] != nil {
					mbm.PageSetREAD(bank, slotfw[bank-0xc0])
					mbm.PageSetWRITE(bank, slotfw[bank-0xc0])
				} else {
					mbm.PageSetREAD(bank, slotroms[bank-0xC0])
				}
			} else {
				mbm.PageSetREAD(bank, intcxrom)
				if slotroms[bank-0xC0] != nil {
					mbm.PageSetWRITE(bank, slotroms[bank-0xC0])
				}
			}
		}
	}

	// upper c8 rom
	switch mr.GetRegC8ROMType() {
	case c8RomEmpty:
		for bank := 0xC8; bank < 0xD0; bank++ {
			mbm.PageSetREAD(bank, nil)
		}
	case c8RomDevice:
		slot := mr.GetRegPeripheralROMSlot()
		exprom := mbm.Get(fmt.Sprintf("slotexp%d.rom", slot))
		for bank := 0xc8; bank < 0xd0; bank++ {
			mbm.PageSetREAD(bank, exprom)
		}
	case c8RomInternal:
		for bank := 0xC8; bank < 0xD0; bank++ {
			mbm.PageSetREAD(bank, intcxrom)
		}
	}

	// language / RAM card 0xd000 - 0xdfff
	for bank := 0xD0; bank < 0xE0; bank++ {
		// READ of memory

		if mr.SW_HIGHRAM() {
			if mr.SW_ALTZP() {
				if mr.SW_BANK2() {
					mbm.PageSetREAD(bank, auxlcbank2)
				} else {
					mbm.PageSetREAD(bank, auxlanguagecard)
				}
			} else {
				if mr.SW_BANK2() {
					mbm.PageSetREAD(bank, lcbank2)
				} else {
					mbm.PageSetREAD(bank, languagecard)
				}
			}
		} else {
			mbm.PageSetREAD(bank, applerom)
		}
		// WRITE of memory
		if mr.SW_WRITERAM() {
			if mr.SW_ALTZP() {
				if mr.SW_BANK2() {
					mbm.PageSetWRITE(bank, auxlcbank2)
				} else {
					mbm.PageSetWRITE(bank, auxlanguagecard)
				}
			} else {
				if mr.SW_BANK2() {
					mbm.PageSetWRITE(bank, lcbank2)
				} else {
					mbm.PageSetWRITE(bank, languagecard)
				}
			}
			//}
		} else {
			mbm.PageSetWRITE(bank, nil) // not writeable
		}
	}

	// upper memory 0xE000 - 0xFFFF
	for bank := 0xE0; bank < 0x100; bank++ {
		// reads
		if mr.SW_HIGHRAM() {
			if mr.SW_ALTZP() {
				mbm.PageSetREAD(bank, auxlanguagecard)
			} else {
				mbm.PageSetREAD(bank, languagecard)
			}
		} else {
			// map in ROM
			if bank < 0xF8 {
				mbm.PageSetREAD(bank, applerom)
			} else {
				mbm.PageSetREAD(bank, monitorrom)
			}
		}
		// writes
		if mr.SW_WRITERAM() {
			//if mr.SW_HIGHRAM() {
			//	mbm.PageSetWRITE(bank, mbm.LastMem) // TODO:
			// leave last setting on
			//} else {
			if mr.SW_ALTZP() {
				mbm.PageSetWRITE(bank, auxlanguagecard)
			} else {
				mbm.PageSetWRITE(bank, languagecard)
			}
			//}
		} else {
			mbm.PageSetWRITE(bank, nil) // no write
		}
	}

	// Modify based on 80STORE
	if mr.SW_80STORE() {
		for bank := 0x04; bank < 0x08; bank++ {
			// READ
			if mr.SW_PAGE2() {
				mbm.PageSetREAD(bank, memaux)
				mbm.PageSetWRITE(bank, memaux)
			} else {
				mbm.PageSetREAD(bank, memmain)
				mbm.PageSetWRITE(bank, memmain)
			}
			// WRITE
			//mbm.PageSetWRITE(bank, mbm.LastMem)
		}

		// Now for hires
		if mr.SW_HIRES() {
			for bank := 0x20; bank < 0x40; bank++ {
				if mr.SW_PAGE2() {
					mbm.PageSetREAD(bank, memaux)
					mbm.PageSetWRITE(bank, memaux)
				} else {
					mbm.PageSetREAD(bank, memmain)
					mbm.PageSetWRITE(bank, memmain)
				}
				// WRITE
				//mbm.PageSetWRITE(bank, mbm.LastMem)
			}
		}
	}

	// cpu := apple2helpers.GetCPU(mr.e)
	// fmt.Printf("======= %s - Configuration @ PC = 0x%.4x\n", cpu.GetModel(), cpu.PC)
	// mr.DumpMapR()
	// mr.DumpMapW()

	servicebus.SendServiceBusMessage(
		mr.e.GetMemIndex(),
		servicebus.MemorySoftSwitchState,
		mr.GetMemorySwitchInfo(),
	)

}

// func (mr *Apple2IOChip) DumpMapR() {
// 	mm := mr.e.GetMemoryMap()
// 	mbm := mm.BlockMapper[mr.e.GetMemIndex()]
// 	begin := 0x00
// 	end := 0x00
// 	last := ""
// 	for bank := 0; bank < 0x100; bank++ {
// 		rs := "NULL"
// 		r := mbm.PageREAD[bank]
// 		if r != nil {
// 			rs = r.Label
// 		}
// 		if rs != last {
// 			if last != "" {
// 				//fmt.Printf("R: %.4x:%.4x: %s\n", begin*0x100, end*0x100+0xff, last)
// 			}
// 			last = rs
// 			begin = bank
// 			end = bank
// 		} else {
// 			end = bank
// 		}
// 	}
// 	if last != "" {
// 		//fmt.Printf("R: %.4x:%.4x: %s\n", begin*0x100, end*0x100+0xff, last)
// 	}
// }

// func (mr *Apple2IOChip) DumpMapW() {
// 	mm := mr.e.GetMemoryMap()
// 	mbm := mm.BlockMapper[mr.e.GetMemIndex()]
// 	begin := 0x00
// 	end := 0x00
// 	last := ""
// 	for bank := 0; bank < 0x100; bank++ {
// 		rs := "NULL"
// 		r := mbm.PageWRITE[bank]
// 		if r != nil {
// 			rs = r.Label
// 		}
// 		if rs != last {
// 			if last != "" {
// 				//fmt.Printf("W: %.4x:%.4x: %s\n", begin*0x100, end*0x100+0xff, last)
// 			}
// 			last = rs
// 			begin = bank
// 			end = bank
// 		} else {
// 			end = bank
// 		}
// 	}
// 	if last != "" {
// 		//fmt.Printf("W: %.4x:%.4x: %s\n", begin*0x100, end*0x100+0xff, last)
// 	}
// }
