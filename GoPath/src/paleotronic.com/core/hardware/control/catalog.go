package control

import (
	s8webclient "paleotronic.com/api"
	"paleotronic.com/core/editor"
)
import "paleotronic.com/core/interfaces"
import "paleotronic.com/core/settings"
import "paleotronic.com/core/memory"
import "paleotronic.com/files"

var filepanel [memory.OCTALYZER_NUM_INTERPRETERS]*editor.FileCatalog
var lastindex [memory.OCTALYZER_NUM_INTERPRETERS]int

func CatalogPresentDiskPicker(ent interfaces.Interpretable, drive int) {

	//ent.StopTheWorld()
	s, e, p := files.System, settings.EBOOT, files.Project

	//	files.System = false
	//	settings.EBOOT = true
	if filepanel[ent.GetMemIndex()] == nil {
		s := editor.FileCatalogSettings{
			DiskExtensions: files.GetExtensionsDisk(),
			Title:          "microFile File Manager",
			Pattern:        files.GetPatternBootable(),
			Path:           "",
			BootstrapDisk:  false,
			InsertDisk:     true,
			TargetDisk:     drive,
			HidePaths:      []string{"FILECACHE", "system", "fs", "boot", "appleii/media", "appleii/micropaks", "appleii/spectrum"},
		}
		filepanel[ent.GetMemIndex()] = editor.NewFileCatalog(ent, s)
	}
	settings.DisableMetaMode[ent.GetMemIndex()] = true
	lastindex[ent.GetMemIndex()], _ = filepanel[ent.GetMemIndex()].Do(lastindex[ent.GetMemIndex()])
	settings.DisableMetaMode[ent.GetMemIndex()] = false

	files.System = s
	settings.EBOOT = e
	files.Project = p

	//settings.SlotInterrupt[ent.GetMemIndex()] = false
	ent.GetMemoryMap().IntSetSlotInterrupt(ent.GetMemIndex(), false)

	if lastindex[ent.GetMemIndex()] != -1 {
		// launch
		ent.GetMemoryMap().IntSetLayerState(ent.GetMemIndex(), 0)
		ent.GetMemoryMap().IntSetActiveState(ent.GetMemIndex(), 0)
	}

	//ent.ResumeTheWorld()

}

func CatalogPresentO(ent interfaces.Interpretable) {

	//ent.StopTheWorld()
	s, e, p := files.System, settings.EBOOT, files.Project

	//	files.System = false
	//	settings.EBOOT = true
	if filepanel[ent.GetMemIndex()] == nil {
		s := editor.FileCatalogSettings{
			DiskExtensions: files.GetExtensionsDisk(),
			Title:          "microFile File Manager",
			Pattern:        files.GetPatternAll(),
			Path:           "",
			BootstrapDisk:  false,
			HidePaths:      []string{"FILECACHE", "system", "fs", "boot", "appleii/media", "appleii/micropaks", "appleii/spectrum"},
		}
		filepanel[ent.GetMemIndex()] = editor.NewFileCatalog(ent, s)
	}
	settings.DisableMetaMode[ent.GetMemIndex()] = true
	lastindex[ent.GetMemIndex()], _ = filepanel[ent.GetMemIndex()].Do(lastindex[ent.GetMemIndex()])
	settings.DisableMetaMode[ent.GetMemIndex()] = false

	files.System = s
	settings.EBOOT = e
	files.Project = p

	//settings.SlotInterrupt[ent.GetMemIndex()] = false
	ent.GetMemoryMap().IntSetSlotInterrupt(ent.GetMemIndex(), false)

	if lastindex[ent.GetMemIndex()] != -1 {
		// launch
		ent.GetMemoryMap().IntSetLayerState(ent.GetMemIndex(), 0)
		ent.GetMemoryMap().IntSetActiveState(ent.GetMemIndex(), 0)
	}

	//ent.ResumeTheWorld()

}

func CatalogPresent(ent interfaces.Interpretable) int {

	settings.EBOOT = !s8webclient.CONN.IsConnected()

	//ent.StopTheWorld()
	s, e, p := files.System, settings.EBOOT, files.Project

	if settings.EBOOT {
		filepanel[ent.GetMemIndex()] = nil
	}

	//	files.System = false
	//	settings.EBOOT = true
	if filepanel[ent.GetMemIndex()] == nil {
		s := editor.FileCatalogSettings{
			DiskExtensions: files.GetExtensionsDisk(),
			Title:          "microFile File Manager",
			Pattern:        files.GetPatternAll(),
			Path:           "",
			BootstrapDisk:  true,
			HidePaths:      []string{"FILECACHE", "system", "fs", "boot", "appleii/media", "appleii/micropaks", "appleii/spectrum"},
		}
		filepanel[ent.GetMemIndex()] = editor.NewFileCatalog(ent, s)
	}
	filepanel[ent.GetMemIndex()].Int = ent
	settings.DisableMetaMode[ent.GetMemIndex()] = true
	lastindex[ent.GetMemIndex()], _ = filepanel[ent.GetMemIndex()].Do(lastindex[ent.GetMemIndex()])
	settings.DisableMetaMode[ent.GetMemIndex()] = false

	files.System = s
	settings.EBOOT = e
	files.Project = p
	//settings.SlotInterrupt[ent.GetMemIndex()] = false
	ent.GetMemoryMap().IntSetSlotInterrupt(ent.GetMemIndex(), false)

	if lastindex[ent.GetMemIndex()] != -1 {
		// launch
		ent.GetMemoryMap().IntSetLayerState(ent.GetMemIndex(), 0)
		ent.GetMemoryMap().IntSetActiveState(ent.GetMemIndex(), 0)
	}

	//ent.ResumeTheWorld()
	return lastindex[ent.GetMemIndex()]

}
