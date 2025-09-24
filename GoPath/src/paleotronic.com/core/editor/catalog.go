package editor

import (
	"errors"
	"path/filepath"
	"sort"
	"strings"
	"time"

	s8webclient "paleotronic.com/api"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/hardware/spectrum"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/filerecord"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
	"paleotronic.com/octalyzer/bus"
	"paleotronic.com/presentation"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type FileCatalogSettings struct {
	BootstrapDisk  bool
	InsertDisk     bool
	TargetDisk     int
	Pattern        string
	Title          string
	Path           string
	DiskExtensions []string
	HidePaths      []string
}

type FileCatalog struct {
	edit           *CoreEdit
	path           string
	title          string
	filter         string
	recs, frecs    []files.FileDef
	Int            interfaces.Interpretable
	selected       int
	settings       FileCatalogSettings
	CopyList       map[string]*filerecord.FileRecord
	CutList        map[string]*filerecord.FileRecord
	DelTarget      string
	RepeatKey      rune
	RepeatInterval time.Duration
	RepeatStart    time.Duration
	KeyDown        time.Time
	LastRepeat     time.Time
	SelectedDrive  int
	skipReload     bool
	lastLine       int
	lastVoffset    int
	lastTime       time.Time
	keyIndex       int
	searchTerm     runestring.RuneString
}

func NewFileCatalog(ent interfaces.Interpretable, s FileCatalogSettings) *FileCatalog {

	this := &FileCatalog{
		path:           s.Path,
		title:          s.Title,
		filter:         s.Pattern,
		Int:            ent,
		settings:       s,
		RepeatInterval: 100 * time.Millisecond,
		RepeatStart:    1000 * time.Millisecond,
	}

	return this

}

func (this *FileCatalog) GetRepeatKey() (rune, bool) {

	if this.RepeatKey == 0 {
		return 0, false
	}

	if time.Since(this.KeyDown) < this.RepeatStart {
		return 0, false
	}

	if time.Since(this.LastRepeat) < this.RepeatInterval {
		return 0, false
	}

	// can repeat
	this.LastRepeat = time.Now()
	return this.RepeatKey, true

}

func (this *FileCatalog) Exit(edit *CoreEdit) {
	edit.Running = false
}

func (this *FileCatalog) Copy(edit *CoreEdit) {
	if this.CopyList == nil {
		this.CopyList = make(map[string]*filerecord.FileRecord)
	}

	l := this.edit.GetLine()
	if l < len(this.recs) {
		r := this.recs[l]
		if r.Kind == files.DIRECTORY {
			this.CopyList[strings.Trim(this.path, "/")+"/"+r.Name] = &filerecord.FileRecord{
				Directory: true,
			}
			this.edit.Changed = true
			this.edit.Running = false
			this.skipReload = true
			this.lastLine = l
			this.lastVoffset = edit.Voffset
			this.InfoPopup(edit, "Copy", "\r\nAdded dir")
			return
		}

		var e error
		if inList(files.GetExt(this.path), this.settings.DiskExtensions) {
			this.path, e = files.MountDSKImage(files.GetPath(this.path), files.GetFilename(this.path), 0)
			if e != nil {
				this.InfoPopup(edit, "Mount Image", "\r\nFailed.")
				return
			}
		}

		fr, err := files.ReadBytesViaProvider(strings.Trim(this.path, "/"), r.Name+"."+r.Extension)
		if err == nil {

			if fr.FileName == "" {
				fr.FileName = r.Name + "." + r.Extension
				fr.FilePath = strings.Trim(this.path, "/")
			}

			this.CopyList[strings.Trim(this.path, "/")+"/"+r.Name+"."+r.Extension] = &fr
			this.edit.Changed = true
			this.edit.Running = false
			this.skipReload = true
			this.lastLine = l
			this.lastVoffset = edit.Voffset
			this.InfoPopup(edit, "Copy", "\r\nAdded file")

		} else {
			this.InfoPopup(edit, "Copy", "\r\nFailed to add file\r\n"+this.path+"/"+r.Name+"."+r.Extension+"\r\n"+err.Error())
		}
	}
}

func (this *FileCatalog) Cut(edit *CoreEdit) {
	if this.CutList == nil {
		this.CutList = make(map[string]*filerecord.FileRecord)
	}

	l := this.edit.GetLine()
	if l < len(this.recs) {
		r := this.recs[l]
		if r.Kind == files.DIRECTORY {
			this.InfoPopup(edit, "Copy", "\r\nCan't cut directory - yet")
			return
		}

		var e error
		if inList(files.GetExt(this.path), this.settings.DiskExtensions) {
			this.path, e = files.MountDSKImage(files.GetPath(this.path), files.GetFilename(this.path), 0)
			if e != nil {
				this.InfoPopup(edit, "Mount Image", "\r\nFailed.")
				return
			}
		}

		fr, err := files.ReadBytesViaProvider(strings.Trim(this.path, "/"), r.Name+"."+r.Extension)
		if err == nil {

			if fr.FileName == "" {
				fr.FileName = r.Name + "." + r.Extension
				fr.FilePath = strings.Trim(this.path, "/")
			}

			this.CutList[strings.Trim(this.path, "/")+"/"+r.Name+"."+r.Extension] = &fr
			this.edit.Running = false
			this.edit.Changed = true
			this.InfoPopup(edit, "Cut", "\r\nAdded file")

		} else {
			this.InfoPopup(edit, "Cut", "\r\nFailed to add file\r\n"+this.path+"/"+r.Name+"."+r.Extension+"\r\n"+err.Error())
		}
	}
}

func (this *FileCatalog) Paste(edit *CoreEdit) {
	if len(this.CopyList) == 0 && len(this.CutList) == 0 {
		return
	}

	lines := 4
	out := []string{}
	for i := 0; i < lines; i++ {
		out = append(out, "                                                           ")
	}

	this.InfoPopup(edit, "Paste Progress",
		strings.Join(out, "\r\n"),
	)
	txt := apple2helpers.TEXT(this.Int)
	x, y := txt.CX, txt.CY

	// have a file, let's read it
	count := 0
	total := len(this.CopyList) + len(this.CutList)

	for p, catCopySrc := range this.CopyList {

		opc := float64(count) / float64(total)
		DrawProgressBar(txt, x, y, "Overall progress:", opc, 60)

		count++

		catCopyDest := catCopySrc.FileName

		if catCopySrc.Directory {
			err := files.RecursiveCopyViaProvider(p, this.path, func(pc float64) {
				DrawProgressBar(txt, x, y+3, "Item progress:", pc, 60)
			})

			if err == nil {
				this.edit.Running = false
				this.edit.Changed = true
				this.edit.Line = len(this.recs)

				this.CopyList = make(map[string]*filerecord.FileRecord)
				this.CutList = make(map[string]*filerecord.FileRecord)
			}

			DrawProgressBar(txt, x, y+3, "Item progress:", 1.0, 60)
			time.Sleep(50 * time.Millisecond)

			continue
		}

		if files.ExistsViaProvider(this.path, catCopyDest) {
			index := 1
			ext := files.GetExt(catCopyDest)
			base := catCopyDest
			npat := "%s %d"
			if ext != "" {
				base = base[0 : len(base)-len(ext)-1]
				npat = "%s %d.%s"
			}
			for files.ExistsViaProvider(this.path, fmt.Sprintf(npat, base, index, ext)) {
				index++
			}
			catCopyDest = fmt.Sprintf(npat, base, index, ext)
		}

		fmt.Printf("[copy] Copy %d bytes from %s to %s/%s\n", len(catCopySrc.Content), p, this.path, catCopyDest)

		DrawProgressBar(txt, x, y+3, "Item progress:", 0, 60)

		e := files.WriteBytesViaProvider(this.path, catCopyDest, catCopySrc.Content)

		if e != nil {
			this.InfoPopup(edit, "Paste", "\r\nFailed: "+catCopyDest)
			return
		}

		DrawProgressBar(txt, x, y+3, "Item progress:", 0.98, 60)

		if catCopySrc.Address != 0 {
			e = files.SetLoadAddressViaProvider(this.path, catCopyDest, catCopySrc.Address)
			if e != nil {
				fmt.Printf("Blurp: %v\n", e)
				this.InfoPopup(edit, "Paste", "\r\nFailed: "+catCopyDest)
				return
			}
		}

		DrawProgressBar(txt, x, y+3, "Item progress:", 1, 60)
		time.Sleep(50 * time.Millisecond)

	}

	for p, catCopySrc := range this.CutList {

		opc := float64(count) / float64(total)
		DrawProgressBar(txt, x, y, "Overall progress:", opc, 60)

		count++

		catCopyDest := catCopySrc.FileName

		fmt.Printf("Copy %d bytes to %s/%s\n", len(catCopySrc.Content), this.path, catCopyDest)

		if strings.Trim(p, "/") == strings.Trim(this.path, "/")+"/"+catCopyDest {
			continue
		}

		DrawProgressBar(txt, x, y+3, "Item progress:", 0, 60)

		e := files.WriteBytesViaProvider(this.path, catCopyDest, catCopySrc.Content)

		if e != nil {
			this.InfoPopup(edit, "Paste", "\r\nFailed: "+catCopyDest)
			return
		}

		DrawProgressBar(txt, x, y+3, "Item progress:", 0.5, 60)

		e = files.DeleteViaProvider(p)
		if e != nil {
			this.InfoPopup(edit, "Cut/Paste", "\r\nFailed: "+catCopyDest)
		}

		DrawProgressBar(txt, x, y+3, "Item progress:", 0.98, 60)

		if catCopySrc.Address != 0 {
			e = files.SetLoadAddressViaProvider(this.path, catCopyDest, catCopySrc.Address)
			if e != nil {
				this.InfoPopup(edit, "Cut/Paste", "\r\nFailed: "+catCopyDest)
				return
			}
		}

		DrawProgressBar(txt, x, y+3, "Item progress:", 1, 60)
		time.Sleep(50 * time.Millisecond)
	}

	this.edit.Running = false
	this.edit.Changed = true
	this.edit.Line = len(this.recs)

	this.CopyList = make(map[string]*filerecord.FileRecord)
	this.CutList = make(map[string]*filerecord.FileRecord)

	this.InfoPopup(edit, "Paste", "\r\nOK")
}

func (this *FileCatalog) Delete(edit *CoreEdit) {
	l := this.edit.GetLine()
	if l < len(this.recs) {
		r := this.recs[l]
		//		if r.Kind == files.DIRECTORY {
		//			return
		//		}
		ot := this.DelTarget
		this.DelTarget = strings.Trim(this.path, "/") + "/" + r.Name + "." + r.Extension
		if r.Extension == "" {
			this.DelTarget = strings.Trim(this.path, "/") + "/" + r.Name
		}

		if ot == this.DelTarget {
			// delete
			e := files.DeleteViaProvider(this.DelTarget)
			if e != nil {
				this.InfoPopup(edit, "Delete", "\r\nFailed.")
				return
			}
			this.DelTarget = ""
			this.edit.Running = false
			this.edit.Changed = true
			this.InfoPopup(edit, "Delete", "\r\nOK.")
		}

	}
}

func (this *FileCatalog) Reboot(edit *CoreEdit) {

	// if inList(files.GetExt(this.path), this.settings.DiskExtensions) {
	// 	// pre-insert disk to drive 0
	// 	settings.PureBootVolume[this.Int.GetMemIndex()] = this.path
	// }

	var bootPak bool

	wd := strings.Trim(this.Int.GetWorkDir(), "/")
	fmt.Printf("working dir is %s\n", wd)
	if strings.HasSuffix(wd, ".pak") {
		bootPak = true
	}
	if strings.HasSuffix(strings.Trim(this.path, "/"), ".pak") {
		bootPak = true
		wd = strings.Trim(this.path, "/")
	}

	if bootPak {
		settings.PureBootVolume[this.Int.GetMemIndex()] = ""
		fmt.Println("Have got a pakfile")
		// reboot pak
		wd = "/" + wd
		settings.IsPakBoot = true
		settings.Pakfile[0] = wd
		this.Int.GetMemoryMap().IntSetSlotRestart(0, true)
		this.Int.GetMemoryMap().MetaKeySet(this.Int.GetMemIndex(), vduconst.SCTRL1, ' ')
		this.Int.GetProducer().Select(0)

		// fr, err := files.ReadBytesViaProvider(files.GetPath(wd), files.GetFilename(wd))
		// if err != nil {
		// 	return
		// }
		// zp, err := files.NewOctContainer(&fr, wd)
		// if err != nil {
		// 	return
		// }

		// //settings.MicroPakPath = wd
		// this.Int.GetMemoryMap().IntSetSlotInterrupt(this.Int.GetMemIndex(), false)
		// if slot, err := BootStrap(wd, zp, this.Int); err == nil {
		// 	apple2helpers.MonitorPanel(this.Int, false)
		// 	this.Int.GetMemoryMap().IntSetSlotInterrupt(this.Int.GetMemIndex(), false)
		// 	this.Int.GetProducer().Select(slot)
		// 	this.edit.Running = false
		// 	this.Int.GetMemoryMap().IntSetSlotRestart(slot, true)
		// 	this.Int.GetMemoryMap().MetaKeySet(this.Int.GetMemIndex(), vduconst.SCTRL1+rune(slot), ' ')
		// }
		return
	}

	apple2helpers.MonitorPanel(this.Int, false)

	//	settings.SlotRestart[this.Int.GetMemIndex()] = true
	this.Int.GetMemoryMap().IntSetSlotRestart(this.Int.GetMemIndex(), true)

	this.edit.Running = false
}

func (this *FileCatalog) RebootOctamode(edit *CoreEdit) {
	settings.TemporaryMute = false
	time.Sleep(50 * time.Millisecond)
	this.Int.GetMemoryMap().MetaKeySet(this.Int.GetMemIndex(), vduconst.SCTRL1, ' ')
	time.Sleep(50 * time.Millisecond)
	edit.Running = false
	edit.ForceExit = true
	settings.CleanBootRequested = true
	settings.BlueScreen = true
}

func (this *FileCatalog) Disk1(edit *CoreEdit) {
	l := edit.GetLine()
	if l < len(this.recs) {

		switch this.SelectedDrive {
		case 0:
			if settings.PureBootVolume[this.Int.GetMemIndex()] != "" {
				this.EjectDisk1(edit)
			}
		case 1:
			if settings.PureBootVolume2[this.Int.GetMemIndex()] != "" {
				this.EjectDisk1(edit)
			}
		}

		r := this.recs[l]
		if r.Kind == files.DIRECTORY {
			return
		}

		this.Int.SetWorkDir("/")
		diskfile := "/" + strings.Trim(this.path, "/") + "/" + r.Name + "." + r.Extension

		if files.Apple2IsHighCapacity(r.Extension, int(r.Size)) {

			servicebus.SendServiceBusMessage(
				this.Int.GetMemIndex(),
				servicebus.SmartPortInsertFilename,
				servicebus.DiskTargetString{
					Drive:    this.SelectedDrive,
					Filename: diskfile,
				},
			)

		} else {

			servicebus.SendServiceBusMessage(
				this.Int.GetMemIndex(),
				servicebus.DiskIIInsertFilename,
				servicebus.DiskTargetString{
					Drive:    this.SelectedDrive,
					Filename: diskfile,
				},
			)

		}

		// apple2.DiskInsert(this.Int, this.SelectedDrive, diskfile, settings.PureBootVolumeWP[this.Int.GetMemIndex()])

		// fmt.Printf("%d: <- %s\n", this.SelectedDrive, diskfile)

		// // update slot 0 disk
		// switch this.SelectedDrive {
		// case 0:
		// 	settings.PureBootVolume[this.Int.GetMemIndex()] = diskfile
		// case 1:
		// 	settings.PureBootVolume2[this.Int.GetMemIndex()] = diskfile
		// }
		files.MountDSKImage(files.GetPath(diskfile), files.GetFilename(diskfile), this.SelectedDrive)
		this.GetVolumes(edit)
	}
}

func (this *FileCatalog) ToggleDrive(edit *CoreEdit) {

	this.SelectedDrive = (this.SelectedDrive + 1) & 1

	this.InfoPopup(edit, fmt.Sprintf("Drive %d Selected", this.SelectedDrive), "")

}

func (this *FileCatalog) ToggleWP1(edit *CoreEdit) {

	// s := "\r\nOff"
	// switch this.SelectedDrive {
	// case 0:
	// 	settings.PureBootVolumeWP[this.Int.GetMemIndex()] = !settings.PureBootVolumeWP[this.Int.GetMemIndex()]
	// 	if settings.PureBootVolumeWP[this.Int.GetMemIndex()] {
	// 		s = "\r\nEnabled"
	// 	}
	// 	apple2.SetWriteProtect(this.Int, this.SelectedDrive, settings.PureBootVolumeWP[this.Int.GetMemIndex()])
	// case 1:
	// 	settings.PureBootVolumeWP2[this.Int.GetMemIndex()] = !settings.PureBootVolumeWP2[this.Int.GetMemIndex()]
	// 	if settings.PureBootVolumeWP2[this.Int.GetMemIndex()] {
	// 		s = "\r\nEnabled"
	// 	}
	// 	apple2.SetWriteProtect(this.Int, this.SelectedDrive, settings.PureBootVolumeWP2[this.Int.GetMemIndex()])
	// }
	servicebus.SendServiceBusMessage(
		this.Int.GetMemIndex(),
		servicebus.DiskIIToggleWriteProtect,
		this.SelectedDrive,
	)

	resp, _ := servicebus.SendServiceBusMessage(
		this.Int.GetMemIndex(),
		servicebus.DiskIIQueryWriteProtect,
		this.SelectedDrive,
	)
	s := "Disabled"
	if resp[0].Payload.(bool) {
		s = "Enabled"
	}

	this.InfoPopup(edit, fmt.Sprintf("Drive %d WriteProtect is %s", this.SelectedDrive, s), "")

}

func (this *FileCatalog) EjectDisk1(edit *CoreEdit) {

	// check existing disk
	dsk := apple2.GetDisk(this.Int, this.SelectedDrive)
	if dsk != nil {
		if dsk.IsModified() {
			var path string
			switch this.SelectedDrive {
			case 0:
				path = settings.PureBootVolume[this.Int.GetMemIndex()]
			case 1:
				path = settings.PureBootVolume2[this.Int.GetMemIndex()]
			}

			var save bool
			if strings.HasPrefix(path, "local/mydisks") {
				save = true // resave existing
			}
			if !save {
				save = this.YesNoPopup(edit, "DISK STATE CHANGED", "Save content of disk to local environment? (Y/N)")
			}
			if save {
				filename := files.GetFilename(path)
				if !strings.HasSuffix(filename, ".woz") {
					ext := files.GetExt(filename)
					filename = strings.Replace(filename, "."+ext, ".woz", -1)
				}
				files.MkdirViaProvider("/local/mydisks")
				files.WriteBytesViaProvider(
					"/local/mydisks",
					filename,
					dsk.GetData().ByteSlice(0, dsk.GetSize()),
				)
			}
		}
	}

	// diskfile := ""
	// apple2.DiskInsert(this.Int, this.SelectedDrive, diskfile, settings.PureBootVolumeWP[this.Int.GetMemIndex()])

	// // update slot 0 disk
	// switch this.SelectedDrive {
	// case 0:
	// 	settings.PureBootVolume[this.Int.GetMemIndex()] = diskfile
	// case 1:
	// 	settings.PureBootVolume2[this.Int.GetMemIndex()] = diskfile
	// }
	servicebus.SendServiceBusMessage(
		this.Int.GetMemIndex(),
		servicebus.DiskIIEject,
		this.SelectedDrive,
	)

	this.EjectSP(edit)

	this.GetVolumes(edit)

}

func (this *FileCatalog) EjectSP(edit *CoreEdit) {

	settings.PureBootSmartVolume[this.Int.GetMemIndex()] = ""
	servicebus.SendServiceBusMessage(
		this.Int.GetMemIndex(),
		servicebus.SmartPortEject,
		this.SelectedDrive,
	)

}

func (this *FileCatalog) BrowseDisk(edit *CoreEdit) {
	l := edit.GetLine()
	if l < len(this.recs) {

		r := this.recs[l]
		if r.Kind == files.DIRECTORY {
			return
		}

		if inList(r.Extension, this.settings.DiskExtensions) || r.Extension == "pak" {

			fmt.Printf("%s is a disk, pathing into it\n", r.Name)

			this.path = strings.TrimRight(this.path, "/") + "/" + r.Name + "." + r.Extension

			if files.ExistsViaProvider(this.path, "") {
				fmt.Printf("Browse path: %s\n", this.path)
			}
			this.edit.Running = false
			this.edit.Changed = true
		} else {
			this.InfoPopup(edit, "Disk Browse", "\r\nNot a disk!")
		}
	}
}

func (this *FileCatalog) SaveDisk1(edit *CoreEdit) {

	// check existing disk
	dsk := apple2.GetDisk(this.Int, 0)
	var diskfile string
	if dsk != nil {
		path := settings.PureBootVolume[this.Int.GetMemIndex()]
		filename := files.GetFilename(path)
		if !strings.HasSuffix(filename, ".woz") {
			ext := files.GetExt(filename)
			filename = strings.Replace(filename, "."+ext, ".woz", -1)
		}
		files.MkdirViaProvider("/local/mydisks")
		files.WriteBytesViaProvider(
			"/local/mydisks",
			filename,
			dsk.GetData().ByteSlice(0, dsk.GetSize()),
		)

		diskfile = "/local/mydisks/" + filename

		apple2.DiskInsert(this.Int, 0, diskfile, settings.PureBootVolumeWP[this.Int.GetMemIndex()])

		// update slot 0 disk
		settings.PureBootVolume[this.Int.GetMemIndex()] = diskfile

		this.GetVolumes(edit)
	}

}

func (this *FileCatalog) SaveDisk2(edit *CoreEdit) {

	// check existing disk
	dsk := apple2.GetDisk(this.Int, 1)
	var diskfile string
	if dsk != nil {
		path := settings.PureBootVolume2[this.Int.GetMemIndex()]
		filename := files.GetFilename(path)
		if !strings.HasSuffix(filename, ".woz") {
			ext := files.GetExt(filename)
			filename = strings.Replace(filename, "."+ext, ".woz", -1)
		}
		files.MkdirViaProvider("/local/mydisks")
		files.WriteBytesViaProvider(
			"/local/mydisks",
			filename,
			dsk.GetData().ByteSlice(0, dsk.GetSize()),
		)

		diskfile = "/local/mydisks/" + filename

		apple2.DiskInsert(this.Int, 1, diskfile, settings.PureBootVolumeWP2[this.Int.GetMemIndex()])

		// update slot 0 disk
		settings.PureBootVolume2[this.Int.GetMemIndex()] = diskfile

		this.GetVolumes(edit)
	}

}

func (this *FileCatalog) Swap(edit *CoreEdit) {

	//apple2.DiskInsert(this.Int, 1, diskfile)
	// apple2.DiskSwap(this.Int)

	// v0 := settings.PureBootVolume[this.Int.GetMemIndex()]
	// v1 := settings.PureBootVolume2[this.Int.GetMemIndex()]

	// settings.PureBootVolume[this.Int.GetMemIndex()] = v1
	// settings.PureBootVolume2[this.Int.GetMemIndex()] = v0
	e := this.Int

	// wp0 := settings.PureBootVolumeWP[e.GetMemIndex()]
	// wp1 := settings.PureBootVolumeWP2[e.GetMemIndex()]
	// s := apple2.GetDiskVolumes(e)
	// d0 := apple2.GetDisk(e, 0)
	// d1 := apple2.GetDisk(e, 1)
	// apple2.SetDisk(e, 0, d1)
	// apple2.SetDisk(e, 1, d0)
	// settings.PureBootVolume[e.GetMemIndex()] = s[1]
	// settings.PureBootVolume2[e.GetMemIndex()] = s[0]

	servicebus.SendServiceBusMessage(
		e.GetMemIndex(),
		servicebus.DiskIIExchangeDisks,
		"Disk swap",
	)

	this.GetVolumes(edit)

}

func (this *FileCatalog) GetVolumes(edit *CoreEdit) {

	//apple2.DiskInsert(this.Int, 1, diskfile)
	volumes := []string{"", "", ""}

	volumes[0] = settings.PureBootVolume[this.Int.GetMemIndex()]

	volumes[1] = settings.PureBootVolume2[this.Int.GetMemIndex()]

	volumes[2] = settings.PureBootSmartVolume[this.Int.GetMemIndex()]

	this.InfoPopup(edit, "Disk Status", fmt.Sprintf("\r\n\r\nDrive A: %s\r\nDrive B: %s\r\nSmartPort: %s\r\n", volumes[0], volumes[1], volumes[2]))

}

func (this *FileCatalog) InfoPopup(edit *CoreEdit, title string, message string) {

	apple2helpers.TextAddWindow(this.Int, "std", 0, 0, 79, 47)

	if settings.HighContrastUI {
		apple2helpers.SetFGColor(this.Int, 15)
		apple2helpers.SetBGColor(this.Int, 0)
	} else {
		apple2helpers.SetFGColor(this.Int, 4)
		apple2helpers.SetBGColor(this.Int, 15)
	}

	content := []string{strings.Trim(title, "\r\n")}
	lines := strings.Split(message, "\r\n")
	content = append(content, lines...)

	txt := apple2helpers.TEXT(this.Int)
	txt.HideCursor()
	txt.DrawTextBoxAuto(content, true, true)

	apple2helpers.TextUseWindow(this.Int, "std")

	time.Sleep(2 * time.Second)
	txt.ShowCursor()

}

func (this *FileCatalog) InfoPopupLong(edit *CoreEdit, title string, message string) {

	this.InfoPopup(edit, title, message)

	// apple2helpers.TextAddWindow(this.Int, "std", 0, 0, 79, 47)

	// if settings.HighContrastUI {
	// 	apple2helpers.SetFGColor(this.Int, 0)
	// 	apple2helpers.SetBGColor(this.Int, 15)
	// } else {
	// 	apple2helpers.SetFGColor(this.Int, 4)
	// 	apple2helpers.SetBGColor(this.Int, 15)
	// }

	// apple2helpers.TextDrawBox(
	// 	this.Int,
	// 	5,
	// 	5,
	// 	70,
	// 	38,
	// 	title,
	// 	true,
	// 	true,
	// )

	// this.Int.PutStr(message)

	// apple2helpers.TextUseWindow(this.Int, "std")

	// //_ = edit.GetCRTKey()
	// time.Sleep(time.Second)

}

func (this *FileCatalog) YesNoPopup(edit *CoreEdit, title string, message string) bool {

	apple2helpers.TextAddWindow(this.Int, "std", 0, 0, 79, 47)

	if settings.HighContrastUI {
		apple2helpers.SetFGColor(this.Int, 0)
		apple2helpers.SetBGColor(this.Int, 15)
	} else {
		apple2helpers.SetFGColor(this.Int, 4)
		apple2helpers.SetBGColor(this.Int, 15)
	}

	apple2helpers.TextDrawBox(
		this.Int,
		5,
		20,
		70,
		8,
		title,
		true,
		true,
	)

	this.Int.PutStr(title + "\r\n")
	this.Int.PutStr(message)

	apple2helpers.TextUseWindow(this.Int, "std")

	key := edit.GetCRTKey()
	for key != 'y' && key != 'Y' && key != 'n' && key != 'N' {
		key = edit.GetCRTKey()
	}

	return (key == 'y') || (key == 'Y')

}

func (this *FileCatalog) InputPopup(edit *CoreEdit, title string, message string) string {

	apple2helpers.TextAddWindow(this.Int, "std", 0, 0, 79, 47)

	if settings.HighContrastUI {
		apple2helpers.SetFGColor(this.Int, 15)
		apple2helpers.SetBGColor(this.Int, 0)
	} else {
		apple2helpers.SetFGColor(this.Int, 4)
		apple2helpers.SetBGColor(this.Int, 15)
	}

	apple2helpers.TextDrawBox(
		this.Int,
		14,
		20,
		50,
		5,
		title,
		true,
		true,
	)

	this.Int.PutStr(title + "\r\n")

	s := edit.GetCRTLine(message + ": ")

	apple2helpers.TextUseWindow(this.Int, "std")

	return s

}

func (this *FileCatalog) SetupDiskPicker() {
	//	this.path = ""
	//	this.filter = "*.(do,po,nib,dsk)"

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_Q,
		"Quit",
		true,
		this.Exit,
	)

	this.edit.RegisterCommand(
		vduconst.CTRL_TWIDDLE,
		"Quit",
		false,
		this.Exit,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_M,
		"Mount",
		true,
		this.Disk1,
	)

	if !this.settings.InsertDisk {

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_E,
			"Edit",
			true,
			this.EditFile,
		)

	}

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_J,
		"Eject",
		true,
		this.EjectDisk1,
	)

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_W,
		"Write",
		true,
		this.ToggleWP1,
	)

	if !this.settings.InsertDisk {

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_C,
			"Copy",
			true,
			this.Copy,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_X,
			"Cut",
			true,
			this.Cut,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_D,
			"Del",
			true,
			this.Delete,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_V,
			"Paste",
			true,
			this.Paste,
		)
	}

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_A,
		"Swap",
		true,
		this.Swap,
	)

	if !this.settings.InsertDisk {

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_P,
			"Pak",
			true,
			this.CreateOctafile,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_B,
			"Boot",
			true,
			this.Reboot,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_R,
			"Power",
			true,
			this.RebootOctamode,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_O,
			"Open",
			true,
			this.BrowseDisk,
		)
	}

	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_S,
		"Drive",
		true,
		this.ToggleDrive,
	)

	if !this.settings.InsertDisk {

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_N,
			"Name",
			true,
			this.RenameFile,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_F,
			"Folder",
			true,
			this.NewFolder,
		)

		this.edit.RegisterCommand(
			vduconst.SHIFT_CTRL_H,
			"Help",
			true,
			this.Help,
		)

	}

}

func (this *FileCatalog) SetupBrowser() {
	this.path = this.Int.GetWorkDir()
	this.filter = "*.*"
	this.edit.RegisterCommand(
		vduconst.SHIFT_CTRL_Q,
		"Quit",
		true,
		this.Exit,
	)
}

func (this *FileCatalog) Pad(str string, l int) string {
	for len(str) < l {
		str = str + " "
	}
	return str[0:l]
}

func (this *FileCatalog) EndScreen() {
	apple2helpers.TextRestoreScreen(this.Int)
	apple2helpers.MonitorPanel(this.Int, false)
}

func (this *FileCatalog) StartScreen() {
	apple2helpers.MonitorPanel(this.Int, true)
	apple2helpers.TextSaveScreen(this.Int)
	if settings.HighContrastUI {
		apple2helpers.TEXTMAX(this.Int)
	} else {
		apple2helpers.TEXTMAX(this.Int)
	}
}

func inList(s string, list []string) bool {
	for _, v := range list {
		if strings.ToLower(s) == strings.ToLower(v) {
			return true
		}
	}
	return false
}

func (this *FileCatalog) Do(gotoline int) (int, error) {

	//log2.Printf("launching catalog...")

	//debug.PrintStack()

	settings.TemporaryMute = true
	settings.DisableOverlays = true
	settings.VideoSuspended = false

	gfx, hud := apple2helpers.GetActiveLayers(this.Int)
	delete(hud, "OOSD") // #fix for stuck osd issue
	apple2helpers.SetActiveLayers(this.Int,
		map[string]bool{
			"HGR1": false,
			"HGR2": false,
			"LOGR": false,
			"DLGR": false,
			"XGR1": false,
			"XGR2": false,
			"VCTR": false,
			"DHR1": false,
			"DHR2": false,
			"SHR1": false,
		},
		hud,
	)

	if strings.HasSuffix(this.path, ".pak") {
		this.path = files.GetPath(this.path)
	}

	//r := bus.IsClock()
	//if !r {
	bus.StartDefault()
	//}

	txt := apple2helpers.TEXT(this.Int)
	txt.HideCursor()
	defer func() {
		settings.DisableOverlays = false
		this.CopyList = make(map[string]*filerecord.FileRecord)
		this.CutList = make(map[string]*filerecord.FileRecord)
		apple2helpers.SetActiveLayers(this.Int, gfx, hud)
		settings.TemporaryMute = false
		settings.DisableMetaMode[this.Int.GetMemIndex()] = false
		apple2helpers.TextHideCursor(this.Int)
		//if !r {
		bus.StopClock()
		//}
		cpu := apple2helpers.GetCPU(this.Int)
		cpu.ResetSliced()
	}()

	caller := this.Int

	//this.Int.SetBuffer(runestring.NewRuneString())

	for strings.HasSuffix(this.path, "//") {
		this.path = strings.Replace(this.path, "//", "/", -1)
	}

	if this.settings.InsertDisk {
		this.SelectedDrive = this.settings.TargetDisk
	}

	this.selected = -1

	//var dskp *files.DSKFileProvider
	//dskp = files.NewDSKFileProvider("", 0)

	this.StartScreen()
	gotoVOffset := 0

	for true {

		if this.path != "" && rune(this.path[0]) == '/' {
			this.path = this.path[1:]
		}
		//this.path = strings.TrimRight(this.path, "/")

		var e error

		if !this.skipReload {
			fmt.Printf(">>> READING DIR: path=[%s], pattern=[%s]\n", this.path, this.filter)
			this.recs, this.frecs, e = files.ReadDirViaProvider(this.path, this.filter)
			// }

			if e != nil {
				this.EndScreen()
				return this.selected, e
			}

			sort.Sort(files.ByName(this.recs))
			sort.Sort(files.ByName(this.frecs))

			this.recs = append(this.recs, this.frecs...)

			if len(this.settings.HidePaths) > 0 {
				tmp := make([]files.FileDef, 0)
				for _, fr := range this.recs {
					if fr.Kind != files.DIRECTORY ||
						(!inList(strings.TrimPrefix(strings.TrimSuffix(fr.Name, "/"), "/"), this.settings.HidePaths) &&
							!inList(strings.Trim(this.path, "/")+"/"+strings.TrimPrefix(strings.TrimSuffix(fr.Name, "/"), "/"), this.settings.HidePaths)) {
						tmp = append(tmp, fr)
					}
				}
				this.recs = tmp
			}
		} else {
			gotoline = this.lastLine
			gotoVOffset = this.lastVoffset
		}

		this.skipReload = false

		maxlen := 20
		if settings.HighContrastUI {
			maxlen = 40
		}

		text := ""
		filecount := 0
		for _, r := range this.recs {

			// if r.Extension == "zip" {
			// 	r.Kind = files.DIRECTORY
			// }
			if r.Name != ".." {
				filecount++
			}

			namecol := vduconst.COLOR15
			bgcol := vduconst.BGCOLOR4

			r.Description = strings.Replace(r.Description, string(rune(7)), "", -1)
			if r.Kind == files.DIRECTORY {
				r.Extension = "dir"
				namecol = vduconst.COLOR13
			}
			if len(r.Name) > maxlen && r.Description == "" {
				r.Description = r.Name
			}
			text = text + string(rune(bgcol)) +
				string(rune(namecol)) + this.Pad(
				utils.FlattenAccent(r.Name), maxlen) +
				" " +
				this.Pad(r.Extension+" ", 4) +
				this.Pad(""+utils.IntToStr(int(r.Size)), 10) +
				string(rune(vduconst.BGCOLOR4)) +
				string(rune(vduconst.COLOR14)) +
				utils.FlattenAccent(r.Description) +
				"\r\n"
		}
		text = text + " \r\n" + utils.IntToStr(filecount) + " file(s)"

		//        this.Int.SaveVDUState()
		//        this.Int.SetVideoMode(this.Int.GetVideoModes()[0])

		this.edit = NewCoreEdit(this.Int, this.title, text, false, true)
		this.edit.SetEventHandler(this)
		this.edit.BarBG = 15
		this.edit.BarFG = 4
		this.edit.SelBG = 13
		this.edit.SelFG = 4
		this.edit.BGColor = 4
		this.edit.FGColor = 15
		this.edit.MY = 36

		//this.edit.GotoLine(this.selected)

		if len(this.CopyList) > 0 {
			this.edit.SubTitle = fmt.Sprintf("%d files to copy", len(this.CopyList))
		} else if this.DelTarget != "" {
			this.edit.SubTitle = "Ctrl+D again to delete"
		} else {
			this.edit.SubTitle = ""
		}

		if this.settings.InsertDisk {
			this.edit.Title = "microM8 Disk Selector"
			this.edit.SubTitle = "ENTER selects disk"
		}

		this.SetupDiskPicker()

		if gotoline > 0 && gotoline < len(this.recs) {
			this.edit.GotoLine(gotoline)
			gotoline = 0
		}

		if gotoVOffset > 0 && gotoVOffset < len(this.recs) {
			this.edit.Voffset = gotoVOffset
			gotoVOffset = 0
		}

		var running bool = true

		go func() {

			for running {

				time.Sleep(this.RepeatInterval)
				if ch, ok := this.GetRepeatKey(); ok {
					this.OnEditorKeypress(this.edit, ch)
					this.edit.Display()
				}

			}

		}()

		this.edit.Run()
		if this.Int.VM().IsDying() {
			return -1, nil
		}
		//        this.Int.RestoreVDUState()

		running = false

		if this.edit.ForceExit {
			fmt.Println("Force exit")
			apple2helpers.TextRestoreScreen(caller)
			apple2helpers.MonitorPanel(this.Int, false)
			break
		}

		if this.edit.Changed {
			this.edit.Changed = false
			continue
		}

		if (this.selected >= 0) && (this.selected < len(this.recs)) {

			rr := this.recs[this.selected]

			ext := rr.Extension
			if ext == "" {
				ext = files.GetExt(rr.Name)
			}

			info, metaValid := files.GetInfo(ext)

			fmt.Printf("rr.Extension = [%s]\n", ext)

			if !metaValid && rr.Kind != files.DIRECTORY {

				this.InfoPopup(this.edit, "Unsupported FileType", "\r\nNo handler for this filetype.")

			} else if rr.Kind == files.FILE && rr.Extension == "wav" {
				zp := "/" + strings.Trim(this.path, "/") + "/" + rr.Name + ".wav"
				fp, err := files.ReadBytesViaProvider(files.GetPath(zp), files.GetFilename(zp))
				if err == nil {
					// if zxs.Is128K {
					// 	settings.SpecFile[this.Int.GetMemIndex()] = "zx-spectrum-128k.yaml"
					// } else {
					// 	settings.SpecFile[this.Int.GetMemIndex()] = "zx-spectrum-48k.yaml"
					// }
					// settings.VMLaunch[this.Int.GetMemIndex()] = &settings.VMLauncherConfig{
					// 	ZXState: zp,
					// }
					// this.Int.GetMemoryMap().IntSetSlotRestart(this.Int.GetMemIndex(), true)]
					mr, ok := this.Int.GetMemoryMap().InterpreterMappableAtAddress(this.Int.GetMemIndex(), 0xc000)
					if ok {
						io := mr.(*apple2.Apple2IOChip)
						io.TapeAttach(fp.Content)
					}
					// this.Int.GetMemoryMap().IntSetSlotRestart(this.Int.GetMemIndex(), true)
					//apple2helpers.MonitorPanel(this.Int, false)
					return this.selected, nil
				}

			} else if rr.Kind == files.FILE && rr.Extension == "z80" {

				zp := "/" + strings.Trim(this.path, "/") + "/" + rr.Name + ".z80"
				zxs, err := spectrum.ReadSnapshot(zp)
				if err == nil {
					if zxs.Is128K {
						settings.SpecFile[this.Int.GetMemIndex()] = "zx-spectrum-128k.yaml"
					} else {
						settings.SpecFile[this.Int.GetMemIndex()] = "zx-spectrum-48k.yaml"
					}
					settings.VMLaunch[this.Int.GetMemIndex()] = &settings.VMLauncherConfig{
						ZXState: zp,
					}
					this.Int.GetMemoryMap().IntSetSlotRestart(this.Int.GetMemIndex(), true)
					apple2helpers.MonitorPanel(this.Int, false)
					return this.selected, nil
				}

			} else if rr.Kind == files.FILE && rr.Extension == "pak" {

				fmt.Println("In pak handler")

				// fr, err := files.ReadBytesViaProvider(this.path, rr.Name+".pak")
				// if err != nil {
				// 	this.InfoPopup(this.edit, "Error", "\r\nRead failed.")
				// 	continue
				// }
				// zp, err := files.NewOctContainer(&fr, strings.Trim(this.path, "/")+"/"+rr.Name+".pak")

				pakpath := strings.Trim(this.path, "/") + "/" + rr.Name + ".pak"

				//if err == nil {
				settings.MicroPakPath = pakpath
				apple2helpers.MonitorPanel(this.Int, false)
				apple2helpers.MODE40Preserve(this.Int)
				apple2helpers.Clearscreen(this.Int)
				apple2helpers.TextHideCursor(this.Int)
				this.Int.GetMemoryMap().IntSetBackdrop(
					this.Int.GetMemIndex(),
					"",
					7,
					0,
					1,
					1,
					false,
				)
				gfx = map[string]bool{}
				hud = map[string]bool{}
				this.Int.GetMemoryMap().IntSetLayerState(this.Int.GetMemIndex(), 0)
				this.Int.GetMemoryMap().IntSetActiveState(this.Int.GetMemIndex(), 0)
				this.Int.StopMusic()

				go func() {
					this.Int.GetProducer().Select(0)
					this.Int.GetMemoryMap().MetaKeySet(0, vduconst.SCTRL1+rune(0), ' ')
					this.Int.GetMemoryMap().IntSetSlotRestart(0, true)
				}()
				return this.selected, nil
				//}

				// this.InfoPopup(this.edit, "Error", "\r\n\r\nPakfile read failed.")
				// continue

			} else if rr.Kind == files.FILE && rr.Extension == "zip" {

				fr, err := files.ReadBytesViaProvider(this.path, rr.Name+".zip")
				if err != nil {
					panic(err)
				}
				zp := files.NewZipProvider(&fr, strings.Trim(this.path, "/")+"/"+rr.Name+".zip")
				files.AddMapping(strings.Trim(this.path, "/")+"/"+rr.Name+".zip", "", zp)
				this.path = strings.Trim(this.path, "/") + "/" + rr.Name + ".zip"

			} else if rr.Kind == files.DIRECTORY && (info.Browsable || !metaValid) {

				fmt.Printf("dir=%s, path=%s\n", rr.Name, this.path)

				if rr.Name == ".." {

					fmt.Printf("About to .. from %s\n", this.path)

					// if inList(files.GetExt(this.path), this.settings.DiskExtensions) || inList(files.GetExt(this.path), files.GetExtensionsBrowsable()) {
					// 	this.path += "/"
					// }
					if !strings.HasSuffix(this.path, "/") {
						this.path += "/"
					}

					parts := strings.Split(this.path, string('/'))
					if len(parts) >= 2 {
						parts = parts[0 : len(parts)-2]
					} else {
						parts = []string{}
					}
					this.path = strings.Join(parts, string('/')) + string('/')

					fmt.Println("NEW PATH after .. :", this.path)

				} else {
					if inList(files.GetExt(this.path), this.settings.DiskExtensions) {
						// we need to mount this to slot 0
						opath := this.path
						this.path, e = files.MountDSKImage(files.GetPath(this.path), files.GetFilename(this.path), 0)

						this.Int.SetDiskImage(files.GetPath(this.path) + "/" + files.GetFilename(this.path))

						if e != nil {
							apple2helpers.Beep(caller)
							this.path = opath
							continue
						}
					}

					this.path = this.path + rr.Name + string('/')
				}
			} else if rr.Extension == "cfg" || rr.Extension == "ini" {
				this.EditFile(this.edit)

			} else if files.Apple2IsHighCapacity(rr.Extension, int(rr.Size)) {

				if this.Int.IsRecordingVideo() {
					go this.Int.StopRecordingHard()
				}

				fmt.Println("GOT A HDD / 3.5 image ", rr.Name, this.settings.BootstrapDisk)
				settings.Pakfile[this.Int.GetMemIndex()] = ""
				fmt.Println("booting")
				settings.PureBootSmartVolume[this.Int.GetMemIndex()] = this.path + "/" + rr.Name + "." + rr.Extension
				settings.MicroPakPath = ""
				settings.PureBootVolume[this.Int.GetMemIndex()] = ""
				settings.PureBootVolume2[this.Int.GetMemIndex()] = ""
				// this.Int.GetProducer().StopMicroControls()
				// this.Int.PeripheralReset()
				// this.Int.GetMemoryMap().SlotReset(this.Int.GetMemIndex())
				e := this.Int
				if !strings.HasPrefix(settings.SpecFile[e.GetMemIndex()], "apple2") {
					settings.SpecFile[e.GetMemIndex()] = "apple2e-en.yaml"
				}

				this.Int.GetMemoryMap().IntSetSlotRestart(this.Int.GetMemIndex(), true)

				return this.selected, nil

			} else if inList(rr.Extension, this.settings.DiskExtensions) {

				if this.Int.IsRecordingVideo() {
					go this.Int.StopRecordingHard()
				}

				fmt.Println("GOT A DISK ", rr.Name, this.settings.BootstrapDisk)
				settings.Pakfile[this.Int.GetMemIndex()] = ""
				fmt.Println("booting")
				this.Disk1(this.edit)
				settings.PureBootSmartVolume[this.Int.GetMemIndex()] = ""
				settings.MicroPakPath = ""
				if !this.settings.InsertDisk {
					e := this.Int
					if !strings.HasPrefix(settings.SpecFile[e.GetMemIndex()], "apple2") {
						settings.SpecFile[e.GetMemIndex()] = "apple2e-en.yaml"
					}
					// //this.Int.GetProducer().StopMicroControls()
					// this.Int.GetMemoryMap().IntSetSlotRestart(this.Int.GetMemIndex(), true)

					//this.Int.GetMemoryMap().SlotReset(this.Int.GetMemIndex())
					//this.Int.PeripheralReset()
					cfg := &settings.VMLauncherConfig{
						Disks: []string{
							this.path + "/" + rr.Name + "." + rr.Extension,
						},
					}
					settings.VMLaunch[this.Int.GetMemIndex()] = cfg
					this.Int.GetMemoryMap().IntSetSlotRestart(this.Int.GetMemIndex(), true)
					return this.selected, nil
				}
				return -1, nil

			} else if !info.Binary && info.Dialect == "" {

				if this.Int.IsRecordingVideo() {
					go this.Int.StopRecordingHard()
				}

				data, _ := files.ReadBytesViaProvider(this.path, rr.Name+"."+rr.Extension)
				text = utils.Unescape(string(data.Content))

				apple2helpers.TextSaveScreen(caller)
				if settings.HighContrastUI {
					apple2helpers.TEXTMAX(this.Int)
				} else {
					apple2helpers.TEXTMAX(this.Int)
				}

				viewer := NewViewer(caller, "Reader: "+rr.Name+"."+rr.Extension, text)
				viewer.Do()

				this.edit.Changed = true
			} else {

				if this.Int.IsRecordingVideo() {
					go this.Int.StopRecordingHard()
				}

				if info.Binary && !info.Visual && !info.Audio {
					apple2helpers.TextRestoreScreen(caller)
					apple2helpers.MonitorPanel(this.Int, false)
					fmt.Println("TRYING BRUN (cat)")
					settings.UnifiedRender[caller.GetMemIndex()] = false

					//this.Int.GetProducer().StopMicroControls()
					//apple2helpers.CommandResetOverride(this.Int, "fp", this.path, "? chr$(4); \"BRUN "+strings.ToUpper(rr.Name)+"\"")

					settings.VMLaunch[this.Int.GetMemIndex()] = &settings.VMLauncherConfig{
						WorkingDir: this.path,
						Dialect:    "fp",
						RunCommand: "? chr$(4); \"BRUN " + strings.ToUpper(rr.Name) + "\"",
					}
					this.Int.GetMemoryMap().IntSetSlotRestart(this.Int.GetMemIndex(), true)

					return this.selected, nil
				}

				if info.Dialect != "" {

					if this.Int.IsRecordingVideo() {
						go this.Int.StopRecordingHard()
					}

					if info.Dialect != caller.GetDialect().GetShortName() {
						settings.SlotZPEmu[this.Int.GetMemIndex()] = true
						settings.SetPureBoot(this.Int.GetMemIndex(), false)
						apple2helpers.TrashCPU(this.Int)
						_ = apple2helpers.GetCPU(this.Int)
						apple2helpers.SelectHUD(this.Int, "TEXT")
						apple2helpers.TEXT40(this.Int)
					}

					switch info.Dialect {
					case "fp":

						apple2helpers.TextRestoreScreen(caller)
						apple2helpers.MonitorPanel(this.Int, false)

						rs := runestring.NewRuneString()

						fn := "/" + this.path + this.recs[this.selected].Name + "." + this.recs[this.selected].Extension
						if strings.HasSuffix(fn, ".") {
							fn = fn[0 : len(fn)-1]
						}

						if info.Launcher != "" {
							tmp := strings.Replace(info.Launcher, "%f", fn, -1)
							rs.Append(tmp)
						} else {
							rs.Append("run \"/" + strings.Trim(this.path, "/") + "/" + this.recs[this.selected].Name + "\"\r\n")
						}

						settings.SpecFile[this.Int.GetMemIndex()] = "apple2e-en.yaml"
						settings.VMLaunch[this.Int.GetMemIndex()] = &settings.VMLauncherConfig{
							WorkingDir: "",
							Dialect:    "fp",
							RunCommand: rs.String(),
						}
						this.Int.GetMemoryMap().IntSetSlotRestart(this.Int.GetMemIndex(), true)

						return this.selected, nil
					case "int":
						rs := runestring.NewRuneString()

						if info.Launcher != "" {
							tmp := strings.Replace(info.Launcher, "%f", "/"+this.path+this.recs[this.selected].Name+"."+this.recs[this.selected].Extension, -1)
							rs.Append(tmp)
						} else {
							fmt.Printf("run %s\n", this.path+this.recs[this.selected].Name)
							rs.Append("run \"/" + this.path + this.recs[this.selected].Name + "\"\r\n")
						}

						settings.SpecFile[this.Int.GetMemIndex()] = "apple2e-en.yaml"
						settings.VMLaunch[this.Int.GetMemIndex()] = &settings.VMLauncherConfig{
							WorkingDir: "",
							Dialect:    "int",
							RunCommand: rs.String(),
						}
						this.Int.GetMemoryMap().IntSetSlotRestart(this.Int.GetMemIndex(), true)

						return this.selected, nil
					case "logo":
						rs := runestring.NewRuneString()

						if info.Launcher != "" {
							tmp := strings.Replace(info.Launcher, "%f", "/"+this.path+this.recs[this.selected].Name+"."+this.recs[this.selected].Extension, -1)
							rs.Append(tmp)
						} else {
							rs.Append("load \"/" + this.path + this.recs[this.selected].Name + "\r\n")
						}

						// allow inline load of code
						// if this.Int.GetDialect().GetShortName() == "logo" {

						// 	apple2helpers.TextRestoreScreen(this.Int)
						// 	apple2helpers.MonitorPanel(this.Int, false)

						// 	this.Int.GetMemoryMap().IntSetSlotInterrupt(this.Int.GetMemIndex(), false)
						// 	this.Int.PutStr("\r\n")
						// 	this.Int.Parse(rs.String())
						// 	this.Int.SaveCPOS()
						// 	this.Int.SetNeedsPrompt(true)
						// 	this.edit.Running = false
						// 	this.selected = -1
						// 	return this.selected, nil
						// }

						settings.SpecFile[this.Int.GetMemIndex()] = "apple2e-en.yaml"
						settings.VMLaunch[this.Int.GetMemIndex()] = &settings.VMLauncherConfig{
							WorkingDir: "",
							Dialect:    "logo",
							RunCommand: rs.String(),
						}
						this.Int.GetMemoryMap().IntSetSlotRestart(this.Int.GetMemIndex(), true)

						return this.selected, nil
					}

				}
			}
			this.selected = -1
		} else {
			fmt.Println("Fallout here")
			this.EndScreen()
			return this.selected, nil
		}

	}

	this.EndScreen()

	return -1, nil
}

func BootStrap(path string, zp *files.OctContainer, ent interfaces.Interpretable) (int, error) {

	return ent.GetMemIndex(), ent.GetProducer().BootStrap(path)

}

func (this *FileCatalog) OnEditorExit(edit *CoreEdit) {
	//apple2helpers.Clearscreen(edit.Int)
}

func (this *FileCatalog) OnEditorBegin(edit *CoreEdit) {
	//edit.Int.PutStr(string(rune(7)))
}

func (this *FileCatalog) OnEditorChange(edit *CoreEdit) {

}

func (this *FileCatalog) OnEditorMove(edit *CoreEdit) {

}

func (this *FileCatalog) OnMouseMove(edit *CoreEdit, x, y int) {
	fmt.Printf("Mouse at (%d, %d)\n", x, y)

	if y > 0 && y < 46 {
		p := int(y - 1)
		// if settings.HighContrastUI {
		// 	p = int(y-2) / 2
		// }
		tl := edit.Voffset + p
		if tl < len(edit.Content) {
			edit.GotoLine(tl)
			edit.Display()
		}
	}

	edit.MouseMoved = true
}

func (this *FileCatalog) OnMouseClick(edit *CoreEdit, left, right bool) {

	if left && edit.MY > 0 && edit.MY < 46 {

		p := int(edit.MY - 1)
		// if settings.HighContrastUI {
		// 	p = int(edit.MY-2) / 2
		// }

		tl := edit.Voffset + p
		if tl < len(edit.Content) {

			edit.GotoLine(tl)
			edit.Int.GetMemoryMap().KeyBufferAdd(edit.Int.GetMemIndex(), 13)

		}

	}

}

func (this *FileCatalog) OnEditorKeypress(edit *CoreEdit, ch rune) bool {
	//System.Err.Println("Got editor keypress: "+ch);

	if ch == 27 || ch == vduconst.DPAD_B {
		if this.path == "" || this.path == "/" {
			this.edit.Running = false
			this.selected = -1
			return true
		}
		if strings.HasSuffix(this.path, ".dsk") {
			this.path += "/"
		}

		parts := strings.Split(this.path, string('/'))
		fmt.Printf("parts = [%s]\n", strings.Join(parts, ","))
		if len(parts) >= 2 {
			parts = parts[0 : len(parts)-2]
		} else {
			parts = []string{}
		}
		this.path = strings.Join(parts, string('/')) + string('/')
		edit.Changed = true
		edit.Running = false
		return true
	}

	if ch == vduconst.CTRL_TWIDDLE {
		this.edit.Running = false
		this.edit.Done()
		return true
	}

	if ch == 13 || ch == vduconst.DPAD_A {
		l := edit.GetLine()
		if l >= 0 && l < len(this.recs) {
			this.selected = l
			edit.Done()
			return true
		}
	}

	// if ch == 32 || ch == vduconst.DPAD_SELECT {
	// 	this.Disk1(this.edit)
	// 	return true
	// }

	if ch == vduconst.DPAD_UP_PRESS {
		ch = vduconst.DPAD_UP
		this.RepeatKey = vduconst.DPAD_UP
		this.LastRepeat = time.Now()
		this.KeyDown = time.Now()
	}

	if ch == vduconst.DPAD_DOWN_PRESS {
		ch = vduconst.DPAD_DOWN
		this.RepeatKey = vduconst.DPAD_DOWN
		this.LastRepeat = time.Now()
		this.KeyDown = time.Now()
	}

	if ch == vduconst.DPAD_UP_RELEASE || ch == vduconst.DPAD_DOWN_RELEASE {
		this.RepeatKey = 0
	}

	if ch == vduconst.DPAD_UP {
		this.edit.CursorUp()
		return true
	}

	if ch == vduconst.DPAD_DOWN {
		this.edit.CursorDown()
		return true
	}

	//if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
	if ch >= ' ' && ch < 127 {

		if time.Since(this.lastTime) > 1*time.Second {
			this.searchTerm = runestring.Cast("")
		}

		this.searchTerm.AppendSlice([]rune{ch})

		ol := this.edit.Line
		ovo := this.edit.Voffset

		if !this.edit.FindNextPrefix(this.searchTerm) {
			this.edit.Line = 0
			this.edit.Voffset = 0
			if !this.edit.FindNextPrefix(this.searchTerm) {
				this.edit.Line = ol
				this.edit.Voffset = ovo
			}
		}

		this.lastTime = time.Now()

	}

	return false
}

func (this *FileCatalog) CreateOctafile(edit *CoreEdit) {
	ent := edit.Int
	l := this.edit.GetLine()
	if l < len(this.recs) {
		r := this.recs[l]
		if r.Kind == files.DIRECTORY {
			return
		}
		if r.Extension == "zip" || r.Extension == "pak" {
			return
		}

		octaname := r.Name + ".pak"
		octafull := "/local/" + octaname
		if s8webclient.CONN.IsAuthenticated() {
			octafull = "/" + octaname
		}

		//prename := r.Name + ".cfg"
		//prefull := octafull + "/" + prename

		if fr, err := files.ReadBytesViaProvider(strings.Trim(this.path, "/"), r.Name+"."+r.Extension); err == nil {

			o, e := files.NewOctContainerFromFile([]*filerecord.FileRecord{&fr}, octafull, "apple2e-en")
			if e != nil {
				this.InfoPopup(edit, "Failed to create MICROPAK\n\n", "Error: "+e.Error())
			} else {

				p, e := files.NewPresentationStateDefault(ent, octafull)
				if e == nil {
					// b := p.GetBytes()
					// o.SetFileContent("", prename, b)
					// fmt.Printf("Added presentation state: %s\n", prename)
					files.SavePresentationState(p, o)
					ent.GetProducer().SetPresentationSource(ent.GetMemIndex(), octafull)
				}

				this.InfoPopup(edit, "Created MICROPAK\r\n\r\n", "\r\n\r\n"+octafull)
				fmt.Printf("%v", o)
			}

			edit.Running = false
			edit.Changed = true

		} else {
			this.InfoPopup(edit, "Failed to create MICROPAK (read)\n\n", "Error: "+err.Error())
		}

	}
}

func CreateOctafileFromState(ent interfaces.Interpretable, targetpath string) (string, error) {

	index := ent.GetMemIndex()

	if !settings.PureBoot(index) {
		return "", errors.New("Only pureboot right now ;)")
	}

	diskname := filepath.Base(settings.PureBootVolume[index])
	ext := filepath.Ext(diskname)

	octaname := strings.Replace(diskname, ext, ".pak", -1)
	prename := strings.Replace(diskname, ext, ".cfg", -1)
	octafull := strings.Trim(targetpath, "/") + "/" + octaname
	prefull := octafull + "/" + prename
	//control := "control.apl"

	filelist := make([]*filerecord.FileRecord, 0)

	path := settings.PureBootVolume[index]
	if strings.HasPrefix(path, "local:") {
		path = strings.Replace(path, "local:", "/fs", -1)
	}

	if fr, err := files.ReadBytesViaProvider(files.GetPath(path), files.GetFilename(path)); err == nil {

		filelist = append(filelist, &fr)

	}

	path = settings.PureBootVolume2[index]
	if strings.HasPrefix(path, "local:") {
		path = strings.Replace(path, "local:", "/fs", -1)
	}
	if path != "" {
		if fr, err := files.ReadBytesViaProvider(files.GetPath(path), files.GetFilename(path)); err == nil {

			filelist = append(filelist, &fr)

		}
	}

	if len(filelist) > 0 {

		o, e := files.NewOctContainerFromFile(filelist, octafull, strings.Replace(settings.SpecFile[ent.GetMemIndex()], ".yaml", "", -1))
		if e != nil {
			return octafull, e
		}

		// at this point we should create a pre-file with the current video / camera settings etc.
		p, e := presentation.NewPresentationState(ent, octafull)
		if e == nil {
			// b := p.GetBytes()
			// o.SetFileContent("", prename, b)
			// fmt.Printf("Added presentation state: %s\n", prename)
			files.SavePresentationState(p, o)
			ent.GetProducer().SetPresentationSource(ent.GetMemIndex(), prefull)
		}

		//o.SetFileContent("", control, []byte("10 REM CONTROL PROGRAM"))

		return octafull, nil

	}

	return "", nil
}

func (this *FileCatalog) Help(edit *CoreEdit) {

	message := strings.Replace(`

Ctrl+Shift+O   Open (browse) into disk image, zip or pakfile.
Ctrl+Shift+N   Rename file.

Ctrl+Shift+C   Copy (select file for copy)
Ctrl+Shift+X   Cut (select file for cut/move)
Ctrl+Shift+V   Paste (action file copies/moves)
Ctrl+Shift+D   Delete file (do twice)

Ctrl+Shift+E   Edit file.
Ctrl+Shift+F   Create folder

Ctrl+Shift+S   Select active drive.
Ctrl+Shift+M   Mount disk in selected drive.
Ctrl+Shift+J   Eject disk from selected drive.
Ctrl+Shift+W   Toggle Write Protect on selected drive.
Ctrl+Shift+A   Swap disks in drive 0 and 1

Ctrl+Shift+B   Reboot current VM.
Ctrl+Shift+R   Reboot current VM to octamode.

`, "\n", "\r\n", -1)

	this.InfoPopupLong(edit, "HELP\r\n", message)
	edit.GetCRTKey()

}

func (this *FileCatalog) EditFile(edit *CoreEdit) {
	l := this.edit.GetLine()
	if l < len(this.recs) {
		r := this.recs[l]
		if r.Kind == files.DIRECTORY {
			this.InfoPopup(edit, "Copy", "\r\nCan't edit directory - yet")
			return
		}

		if inList(r.Extension, files.GetExtensionsText()) {
			this.Int.SetWorkDir("")
			textpath := strings.Trim(this.path, "/") + "/" + r.Name + "." + r.Extension
			tl := *types.NewTokenList()
			log.Printf("textpath = %s", textpath)
			tl.Push(types.NewToken(types.STRING, textpath))
			a := this.Int.GetCode()
			mstate := settings.DisableMetaMode[this.Int.GetMemIndex()]

			this.Int.GetDialect().GetCommands()["edit"].Execute(nil, this.Int, tl, a, *this.Int.GetPC())

			this.Int.GetMemoryMap().IntSetSlotInterrupt(this.Int.GetMemIndex(), false)

			// e := this.Int.NewChild("editor")
			// e.Bootstrap("fp", true)
			// e.SetWorkDir("/")
			// e.ParseImm("edit " + textpath)

			apple2helpers.MonitorPanel(this.Int, true)
			apple2helpers.SetBGColor(this.Int, uint64(edit.BGColor))
			apple2helpers.SetFGColor(this.Int, uint64(edit.FGColor))
			if settings.HighContrastUI {
				apple2helpers.TEXTMAX(this.Int)
			} else {
				apple2helpers.TEXTMAX(this.Int)
			}
			settings.DisableMetaMode[this.Int.GetMemIndex()] = mstate
		} else {
			fmt.Println("Not editable")
		}
	}
}

func (this *FileCatalog) ViewFile(edit *CoreEdit) {
	l := this.edit.GetLine()
	if l < len(this.recs) {
		r := this.recs[l]
		if r.Kind == files.DIRECTORY {
			this.InfoPopup(edit, "View", "\r\nCan't view directory as text - yet")
			return
		}

		if inList(r.Extension, files.GetExtensionsText()) {
			mstate := settings.DisableMetaMode[this.Int.GetMemIndex()]

			// edit in here
			data, err := files.ReadBytesViaProvider(this.path, r.Name+"."+r.Extension)
			if err == nil {
				text := string(data.Content)
				view := NewCoreEdit(this.Int, "Viewer: "+r.Name+"."+r.Extension, text, false, false)
				view.BGColor = 4
				view.FGColor = 14
				view.BarBG = 14
				view.BarFG = 4
				view.RegisterCommand(
					vduconst.SHIFT_CTRL_Q,
					"Quit",
					true,
					this.Exit,
				)
				view.Run()
			}

			apple2helpers.MonitorPanel(this.Int, true)
			if settings.HighContrastUI {
				apple2helpers.TEXTMAX(this.Int)
			} else {
				apple2helpers.TEXTMAX(this.Int)
			}
			settings.DisableMetaMode[this.Int.GetMemIndex()] = mstate
		} else {
			fmt.Println("Not editable")
		}
	}
}

func (this *FileCatalog) RenameFile(edit *CoreEdit) {
	l := this.edit.GetLine()
	if l < len(this.recs) {
		r := this.recs[l]
		if r.Kind == files.DIRECTORY {
			this.InfoPopup(edit, "Copy", "\r\nCan't rename directory - yet")
			return
		}

		filepath := strings.Trim(this.path, "/")
		filename := r.Name + "." + r.Extension
		newname := this.InputPopup(edit, "RENAME FILE\r\n", "New name")

		// if extension not supplied use the current one
		if files.GetExt(newname) == "" {
			newname += "." + r.Extension
		}

		if newname != "" {
			err := files.RenameViaProvider(filepath, filename, newname)
			if err != nil {
				this.InfoPopup(edit, "RENAME FAILED\r\n", err.Error())
			} else {
				this.InfoPopup(edit, "RENAME OK\r\n", "Rename succeeded")
			}
			edit.Changed = true
			edit.Running = false
		}
	}
}

func (this *FileCatalog) NewFolder(edit *CoreEdit) {

	newname := this.InputPopup(edit, "CREATE FOLDER\r\n", "New name")

	if newname != "" {
		err := files.MkdirViaProvider(this.path + "/" + newname)
		if err != nil {
			this.InfoPopup(edit, "MKDIR FAILED\r\n", err.Error())
		} else {
			this.InfoPopup(edit, "MKDIR OK\r\n", "Create succeeded")
		}
		edit.Changed = true
		edit.Running = false
	}

}

func LoadPresentationFile(path string, ent interfaces.Interpretable) error {

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if !files.ExistsViaProvider(files.GetPath(path), files.GetFilename(path)) {
		return errors.New("Not found")
	}

	// pre, err := files.ReadBytesViaProvider(files.GetPath(path), files.GetFilename(path))
	// if err != nil {
	// 	return err
	// }

	p, _ := files.OpenPresentationState(path)
	//p.LoadBytes(pre.Content)
	//ent.GetProducer().SetPState(ent.GetMemIndex(), p, path)
	p.Apply("init", ent)

	return nil
}
