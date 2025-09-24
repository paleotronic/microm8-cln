package restalgia

import (
	//	"paleotronic.com/fmt"
	"time"
)

type Track struct {
	Instrument string
	Notes      []*Note
	NoteIndex  int
	NoteCount  int
	Orchestra  map[string]*Instrument
	Aliases    map[string]string
	Looping    bool
}

const (
	maxNotes = 64
)

func (this *Track) Add(note string) {
	n := NewNote(note)
	this.Notes = append(this.Notes, n)
	this.NoteCount++
}

func (this *Track) PlayNote(it *InstrumentPack, voice *Voice) {
	// play current note, increment index
	inst := this.Orchestra[this.Instrument]

	// now get the note freq
	idx := this.NoteIndex % this.NoteCount

	////fmt.Printntln("INDEX =", idx)
	n := this.Notes[idx]

	if n.Inst != "" {
		alias, ok := this.Aliases[n.Inst]
		if ok {
			this.Instrument = alias
			inst = this.Orchestra[this.Instrument]
			////fmt.Printntf("Switched to Instrument %s\n", this.Instrument)
		} else {
			in, ok := this.Orchestra[n.Inst]
			if ok {
				this.Instrument = n.Inst
				inst = in
				////fmt.Printntf("Switched to Instrument %s\n", this.Instrument)
			}
		}
	}

	var vol float64 = -1
	var panning float64 = -2
	var freq float64 = -1
	var iii *Instrument = nil

	if n != nil {
		// any volume change?
		if n.GetVolume() >= 0 {
			//voice.SetVolume(n.GetVolume());
			vol = n.GetVolume()
		}

		// any volume change?
		if n.GetPanning() > -2 {
			//voice.SetVolume(n.GetVolume());
			panning = n.GetPanning()
		}

		// any note?
		if n.GetNote() != "" {
			//inst.Apply(voice);
			iii = inst
			//voice.Play(n.GetNote(), n.GetOctave());
			freq = NT.Frequency(n.GetNote(), n.GetOctave())
		}

		// apply it
		//////fmt.Printf("Adding inst to pack: %v %v %v %v\n", voice, iii, freq, vol)
		it.Add(voice, iii, freq, vol, panning)
	} else {
		////fmt.Printntf("Note %d is null!\n", this.NoteIndex)
	}

	this.NoteIndex++

}

func (this *Track) HasNotes() bool {
	return (this.NoteIndex < this.GetNoteCount()) || this.Looping
}

func (this *Track) Delay(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func (this *Track) GetNoteCount() int {
	return this.NoteCount
}

func (this *Track) SetNoteCount(noteCount int) {
	this.NoteCount = noteCount
}

func NewTrack(instrument string, orchestra map[string]*Instrument, aliases map[string]string) *Track {
	this := &Track{}
	this.Notes = make([]*Note, 0)
	// create track
	this.Instrument = instrument
	this.Orchestra = orchestra
	this.Aliases = aliases
	return this
}
