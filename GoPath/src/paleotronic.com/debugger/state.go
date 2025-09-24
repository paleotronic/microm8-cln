package debugger

import (
	"encoding/json"
	"time"

	"paleotronic.com/core/settings"
	"paleotronic.com/octalyzer/backend"

	"paleotronic.com/freeze"

	"paleotronic.com/files"

	"paleotronic.com/fmt"
)

const PathPrefix = "/local/MyDebug"

func (d *Debugger) getStatePath() string {
	files.MkdirViaProvider(PathPrefix)
	s := time.Now().Unix()
	path := fmt.Sprintf("%s/debug-%d.dbz", PathPrefix, s)
	files.MkdirViaProvider(path)
	return path
}

func (d *Debugger) SaveState() error {
	path := d.getStatePath()
	d.PauseCPU()
	f := freeze.NewFreezeState(d.ent(), false)
	err := files.WriteBytesViaProvider(path, "vm.frz", f.SaveToBytes())
	if err != nil {
		return err
	}
	st := d.DebuggerState
	sdata, err := json.Marshal(&st)
	if err != nil {
		return err
	}
	return files.WriteBytesViaProvider(path, "state.json", sdata)
}

func (d *Debugger) LoadState(path string) error {
	j, err := files.ReadBytesViaProvider(path, "state.json")
	if err != nil {
		return err
	}
	var s DebuggerState
	err = json.Unmarshal(j.Content, &s)
	if err != nil {
		return err
	}
	d.DebuggerState = s
	fr, err := files.ReadBytesViaProvider(path, "vm.frz")
	if err != nil {
		return err
	}
	settings.PureBootRestoreStateBin[0] = fr.Content
	settings.DebuggerActiveSlot = 1
	settings.DebuggerAttachSlot = 1
	backend.ProducerMain.AddressSpace.IntSetSlotRestart(0, true)

	return nil
}
