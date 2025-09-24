package restalgia

import ()

type Pattern struct {
	Name   string
	Tracks []*Track
}

func NewPattern(name string) *Pattern {
	this := &Pattern{Name: name, Tracks: make([]*Track, 0)}
	return this
}

func (this *Pattern) Add(track *Track) {
	this.Tracks = append(this.Tracks, track)
}

func (this *Pattern) Get(index int) *Track {
	if index < 0 || index >= len(this.Tracks) {
		return nil
	}
	return this.Tracks[index]
}

func (this *Pattern) Size() int {
	return len(this.Tracks)
}
