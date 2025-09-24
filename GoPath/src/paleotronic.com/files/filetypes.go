package files

import (
	"sort"
	"strings"

	"paleotronic.com/disk"
)

type FileMeta struct {
	Ext            string
	Description    string
	Dialect        string
	Launcher       string
	Bootable       bool
	System         string
	SaveExt        string
	Plain          bool
	EditPlain      bool
	Binary         bool
	Audio          bool
	Visual         bool
	Browsable      bool
	IsHighCapacity bool
}

var FileInfo map[string]FileMeta = map[string]FileMeta{
	"a":   FileMeta{Ext: "a", SaveExt: "apl", Description: "Applesoft", Dialect: "fp", System: "apple2", Plain: true},
	"apl": FileMeta{Ext: "apl", SaveExt: "apl", Description: "Applesoft", Dialect: "fp", System: "apple2", Plain: true},
	"app": FileMeta{Ext: "app", SaveExt: "apl", Description: "Applesoft", Dialect: "fp", System: "apple2", Plain: true},
	"bas": FileMeta{Ext: "bas", SaveExt: "apl", Description: "Applesoft", Dialect: "fp", System: "apple2", Plain: false},
	"i":   FileMeta{Ext: "i", SaveExt: "int", Description: "Integer Basic", Dialect: "int", System: "apple2", Plain: true},
	"int": FileMeta{Ext: "int", SaveExt: "int", Description: "Integer Basic", Dialect: "int", System: "apple2", Plain: true},
	"l":   FileMeta{Ext: "l", SaveExt: "lgo", Description: "Logo", Dialect: "logo", System: "apple2", Plain: true},
	"lgo": FileMeta{Ext: "lgo", SaveExt: "lgo", Description: "Logo", Dialect: "logo", System: "apple2", Plain: true},
	"z":   FileMeta{Ext: "z", SaveExt: "bat", Description: "Shell", Dialect: "shell", System: "apple2", Plain: true},
	"bat": FileMeta{Ext: "bat", SaveExt: "bat", Description: "Shell", Dialect: "shell", System: "apple2", Plain: true},
	"s":   FileMeta{Ext: "s", SaveExt: "bin", Description: "Binary data", System: "apple2", Binary: true},
	"bin": FileMeta{Ext: "bin", SaveExt: "bin", Description: "Binary data", Dialect: "fp", Binary: true, Launcher: "? chr$(4);\"BRUN %f\"\r\n"},
	"d":   FileMeta{Ext: "d", SaveExt: "dat", Description: "Data", Plain: true, System: "apple2"},
	"dat": FileMeta{Ext: "dat", SaveExt: "dat", Description: "Data", Plain: true, System: "apple2"},
	"t":   FileMeta{Ext: "t", SaveExt: "txt", Description: "Text", Plain: true},
	"txt": FileMeta{Ext: "txt", SaveExt: "txt", Description: "Text", Plain: true},
	"dsk": FileMeta{Ext: "dsk", SaveExt: "dsk", Description: "Disk Image", Browsable: true, System: "apple2", Binary: true, Bootable: true},
	"d13": FileMeta{Ext: "d13", SaveExt: "d13", Description: "Disk Image", Browsable: true, System: "apple2", Binary: true, Bootable: true},
	"nib": FileMeta{Ext: "nib", SaveExt: "nib", Description: "Disk Image", System: "apple2", Binary: true, Bootable: true, Browsable: true},
	"do":  FileMeta{Ext: "do", SaveExt: "do", Description: "Disk Image", Browsable: true, System: "apple2", Binary: true, Bootable: true},
	"po":  FileMeta{Ext: "po", SaveExt: "po", Description: "Disk Image", Browsable: true, System: "apple2", Binary: true, Bootable: true},
	"png": FileMeta{Ext: "png", SaveExt: "png", Description: "Image", Binary: true, Visual: true},
	"jpg": FileMeta{Ext: "jpg", SaveExt: "jpg", Description: "Image", Binary: true, Visual: true},
	"wav": FileMeta{Ext: "wav", SaveExt: "wav", Description: "Audio", Binary: true, Audio: true},
	"ogg": FileMeta{Ext: "ogg", SaveExt: "ogg", Description: "Audio", Binary: true, Audio: true},
	"asm": FileMeta{Ext: "asm", SaveExt: "asm", Description: "6502 Assembly", Dialect: "fp", Launcher: "@asm.build{\"%f\"}\r\n", System: "apple2", Plain: true},
	"frz": FileMeta{Ext: "frz", SaveExt: "frz", Description: "Freeze State", Dialect: "fp", Launcher: "@system.thaw{\"%f\"}\r\n", System: "apple2"},
	"rec": FileMeta{Ext: "rec", SaveExt: "rec", Description: "Reality Record", Dialect: "fp", Launcher: "@video.play{\"%f\"}\r\n", System: "apple2"},
	"zip": FileMeta{Ext: "zip", SaveExt: "zip", Description: "Compressed filesystem", Browsable: true},
	"dir": FileMeta{Ext: "dir", SaveExt: "dir", Description: "Filesystem", Browsable: true},
	"pak": FileMeta{Ext: "pak", SaveExt: "pak", Description: "Packfile Compressed filesystem", Browsable: true},
	"cfg": FileMeta{Ext: "cfg", SaveExt: "cfg", Description: "Packfile config file", Plain: true, EditPlain: true},
	"ini": FileMeta{Ext: "ini", SaveExt: "ini", Description: "Packfile boot file", Plain: true},
	"rst": FileMeta{Ext: "rst", SaveExt: "rst", Description: "Audio recording (synth)", Dialect: "fp", Launcher: "@audio.play{\"%f\", 1}\r\n", System: "apple2"},
	"woz": FileMeta{Ext: "woz", SaveExt: "woz", Description: "Applesauce Disk Image", System: "apple2", Binary: true, Bootable: true, Browsable: true},
	"2mg": FileMeta{Ext: "2mg", SaveExt: "2mg", Description: "2MG Disk Image", Browsable: true, System: "apple2", Binary: true, Bootable: true, IsHighCapacity: true},
	"hdv": FileMeta{Ext: "hdv", SaveExt: "hdv", Description: "HDV Disk Image", Browsable: true, System: "apple2", Binary: true, Bootable: true, IsHighCapacity: true},
	"dbz": FileMeta{Ext: "dbz", SaveExt: "dbz", Description: "Debugger State", Dialect: "fp", Launcher: "@system.launchdebug{\"%f\"}\r\n", System: "apple2"},
	"sng": FileMeta{Ext: "sng", SaveExt: "sng", Description: "microTracker file", Dialect: "fp", Launcher: "@music.edit{\"%f\"}\r\n", System: "apple2", Plain: true},
	"tsf": FileMeta{Ext: "tsf", SaveExt: "tsf", Description: "microm8 turtle state", Dialect: "logo", Launcher: "st setcursor [0 20] loadpic \"%f\r\n", System: "apple2", Plain: true},
	"z80": FileMeta{Ext: "z80", SaveExt: "z80", Description: "Spectrum Freeze State", System: "zxspectrum", Binary: false, Bootable: false},
}

func IsPlain(ext string) bool {
	fm, ok := FileInfo[strings.ToLower(ext)]
	if !ok {
		return false
	}
	return fm.Plain
}

func IsEditPlain(ext string) bool {
	fm, ok := FileInfo[strings.ToLower(ext)]
	if !ok {
		return false
	}
	return fm.EditPlain
}

func IsBootable(ext string) bool {
	fm, ok := FileInfo[strings.ToLower(ext)]
	if !ok {
		return false
	}
	return fm.Bootable
}

func IsBrowsable(ext string) bool {
	fm, ok := FileInfo[strings.ToLower(ext)]
	if !ok {
		return false
	}
	return fm.Browsable
}

func IsVisual(ext string) bool {
	fm, ok := FileInfo[strings.ToLower(ext)]
	if !ok {
		return false
	}
	return fm.Visual
}

func IsAudio(ext string) bool {
	fm, ok := FileInfo[strings.ToLower(ext)]
	if !ok {
		return false
	}
	return fm.Audio
}

func IsBinary(ext string) bool {
	fm, ok := FileInfo[strings.ToLower(ext)]
	if !ok {
		return false
	}
	return fm.Binary
}

func IsLaunchable(ext string) (bool, string, string) {
	fm, ok := FileInfo[strings.ToLower(ext)]
	if !ok {
		return false, "", ""
	}
	return fm.Launcher != "", fm.Dialect, fm.Launcher
}

func IsRunnable(ext string) (bool, string) {
	fm, ok := FileInfo[strings.ToLower(ext)]
	if !ok {
		return false, ""
	}
	return (fm.Dialect != "" && !fm.Binary), fm.Dialect
}

func GetPatternAll() string {
	out := "*.("
	for k, _ := range FileInfo {
		if out != "*.(" {
			out += "|"
		}
		out += k
	}
	out += ")"
	return out
}

func GetPatternTape() string {
	return "*.(wav)"
}

func GetPatternBootable() string {
	out := "*.("
	for k, v := range FileInfo {
		if !v.Bootable {
			continue
		}
		if out != "*.(" {
			out += "|"
		}
		out += k
	}
	out += ")"
	return out
}

func GetExtensionsDisk() []string {
	out := []string(nil)
	for k, v := range FileInfo {
		if !v.Bootable {
			continue
		}
		out = append(out, k)
	}
	return out
}

func GetExtensionsText() []string {
	out := []string(nil)
	for k, v := range FileInfo {
		if !v.Plain {
			continue
		}
		out = append(out, k)
	}
	return out
}

func GetExtensionsBrowsable() []string {
	out := []string(nil)
	for k, v := range FileInfo {
		if !v.Browsable {
			continue
		}
		out = append(out, k)
	}
	return out
}

func GetPatternCode() string {
	out := "*.("
	for k, v := range FileInfo {
		if v.Dialect == "" {
			continue
		}
		if out != "*.(" {
			out += "|"
		}
		out += k
	}
	out += ")"
	return out
}

func GetPatternText() string {
	out := "*.("
	for k, v := range FileInfo {
		if v.Binary {
			continue
		}
		if out != "*.(" {
			out += "|"
		}
		out += k
	}
	out += ")"
	return out
}

func GetPatternAudio() string {
	out := "*.("
	for k, v := range FileInfo {
		if !v.Audio {
			continue
		}
		if out != "*.(" {
			out += "|"
		}
		out += k
	}
	out += ")"
	return out
}

func GetPatternVisual() string {
	out := "*.("
	for k, v := range FileInfo {
		if !v.Visual {
			continue
		}
		if out != "*.(" {
			out += "|"
		}
		out += k
	}
	out += ")"
	return out
}

func GetInfo(ext string) (FileMeta, bool) {
	fi, ok := FileInfo[strings.ToLower(ext)]
	return fi, ok
}

func GetTypeCode() []FileMeta {
	out := FileMetas(nil)
	for _, v := range FileInfo {
		if v.Dialect == "" {
			continue
		}
		out = append(out, v)
	}
	sort.Sort(out)
	//	sort.Reverse(out)
	return out
}

func GetTypeBin() []FileMeta {
	out := FileMetas(nil)
	for _, v := range FileInfo {
		if !v.Binary || v.Audio || v.Visual || v.Bootable {
			continue
		}
		out = append(out, v)
	}
	sort.Sort(out)
	sort.Reverse(out)
	return out
}

func GetTypeAll() []FileMeta {
	out := FileMetas(nil)
	for _, v := range FileInfo {
		if v.Audio || v.Visual || v.Bootable {
			continue
		}
		out = append(out, v)
	}
	sort.Sort(out)
	sort.Reverse(out)
	return out
}

func GetPreferredExt(ext string) string {
	i, ok := GetInfo(ext)
	if ok {
		return i.SaveExt
	}
	return ext
}

type FileMetas []FileMeta

func (slice FileMetas) Len() int {
	return len(slice)
}

func (slice FileMetas) Less(i, j int) bool {

	if len(slice[i].Ext) == len(slice[j].Ext) {
		return slice[i].Ext < slice[j].Ext
	}

	return len(slice[i].Ext) > len(slice[j].Ext)
}

func (slice FileMetas) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func Apple2IsHighCapacity(ext string, size int) bool {

	if ext == "nib" || ext == "NIB" || ext == "woz" || ext == "WOZ" {
		return false
	}

	info, ok := GetInfo(ext)
	if !ok {
		return false
	}

	if info.System != "apple2" {
		return false
	}

	if !info.Bootable {
		return false
	}

	if info.IsHighCapacity {
		return true
	}

	if (size/256)*256 > disk.STD_DISK_BYTES {
		return true
	}

	return false

}
