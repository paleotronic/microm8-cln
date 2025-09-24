// +build !remint

package main

import (
	"os"

	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/settings"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/glumby"
	"paleotronic.com/octalyzer/backend"
)

func OnDropFiles(w *glumby.Window, names []string, fake bool) {

	if backend.ProducerMain == nil || settings.BlueScreen {
		return
	}

	fmt.Printf("Dropped %v at %f,%f\n", names, mx, my)

	ww, _ := w.GetGLFWWindow().GetSize()
	drive := (int(mx) / (ww / 2)) + 1
	if fake {
		drive = 1
	}

	n := names[0]

	ext := files.GetExt(n)
	if files.IsBootable(ext) || ext == "pak" || ext == "wav" {
		if !settings.BlueScreen {

			var size int64

			s, err := os.Stat(n)
			if err == nil {
				size = s.Size()
			}

			if !fake {
				if files.Apple2IsHighCapacity(ext, int(size)) {
					go dropAnimation(3) // HDV animation
				} else if ext == "pak" {
					go dropAnimation(4)
				} else if ext == "wav" {
					go dropAnimation(5)
				} else {
					go dropAnimation(drive)
				}
			}
		}
	}

	// if files.GetExt(n) == "wav" {
	// 	servicebus.InjectServiceBusMessage(
	// 		SelectedIndex,
	// 		servicebus.AppleIIAttachTape,
	// 		n,
	// 	)
	// 	return
	// }

	if files.GetExt(n) == "sng" && settings.MicroTrackerEnabled[SelectedIndex] {
		servicebus.InjectServiceBusMessage(
			SelectedIndex,
			servicebus.TrackerLoadSong,
			n,
		)
		return
	}

	// servicebus.InjectServiceBusMessage(
	// 	SelectedIndex,
	// 	servicebus.LaunchEmulator,
	// 	&servicebus.LaunchEmulatorTarget{
	// 		Filename: n,
	// 		Drive:    drive,
	// 	},
	// )

	backend.ProducerMain.VMMediaChange(SelectedIndex, drive, n)

}
