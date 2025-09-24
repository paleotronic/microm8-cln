package tracker

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (s *TSong) ResetEntry(t int) {
	s.targetTrack = t
	s.targetPattern = 0
	s.targetPos = 0
	s.targetNoteLength = 1
	s.targetVolume = 0
	s.targetIsPaused = false
}

func (s *TSong) SelectEntryTrack(t int) {
	s.ResetEntry(t)
}

func (s *TSong) ClearTrack(t int) {
	for _, p := range s.Patterns {
		if p != nil {
			tr := p.Tracks[t]
			if tr != nil {
				for i, _ := range tr.Notes {
					tr.Notes[i] = nil
				}
			}
		}
	}
}

func (s *TSong) EnterNotes(notes string) {

	s.m.Lock()
	defer s.m.Unlock()

	// if time.Since(s.lastTargetEntry) > 10*time.Second {
	// 	s.Stop()
	// 	*s = *NewSong(120, s.Drv) // wipe song
	// 	s.Start(PMBoundPattern)
	// }

	defer func() {
		s.lastTargetEntry = time.Now()
	}()

	// start the current track empty
	s.targetPos = 0
	s.ClearTrack(s.targetTrack)
	s.targetNoteLength = 1
	var chunk string
	for _, ch := range notes {
		if ch >= 'A' && ch <= 'Z' {
			if chunk != "" {
				s.EnterSingle(chunk)
				chunk = ""
			}
			chunk += string(ch)
		} else if ch == '#' {
			chunk += string(ch)
		} else if ch >= '0' && ch <= '9' {
			chunk += string(ch)
		}
	}
	if chunk != "" {
		s.EnterSingle(chunk)
	}

	if !s.targetIsPaused {
		s.Stop()
		time.Sleep(5 * time.Millisecond)
		s.PatternListIndex = 0
		s.PatternPos = 0
		s.Start(PMSingleSong)
	}

	if !s.targetHasEnding {
		s.targetHasEnding = true
		c := byte('p')
		cv := byte(1)
		nn := &TNote{
			Command:      &c,
			CommandValue: &cv,
		}
		if s.targetPos > 0 {
			s.targetPos--
		}
		s.MergeTargetNote(nn)
	}
}

var noteLengths = []int{
	1, 2, 3, 4, 6, 8, 12, 16, 24, 32,
}
var reNoteVal = regexp.MustCompile("^([A-G][#]?)([0-9])$")

func (s *TSong) EnterSingle(note string) {
	note = strings.ToUpper(note)
	noteletter := note[:1]
	switch noteletter {
	case "A", "B", "C", "D", "E", "F", "G":
		// musical note string
		if reNoteVal.MatchString(note) {
			m := reNoteVal.FindAllStringSubmatch(note, -1)
			n := m[0][1]
			o := m[0][2]
			oct, err := strconv.ParseInt(o, 10, 8)
			if err != nil {
				break
			}
			ob := byte(oct)
			if ob < 0 || ob > 6 {
				break
			}
			inst := byte(0)
			v := byte(s.Instruments[0].Voice.Amplitude)
			if s.targetVolume != 0 {
				v = byte(s.targetVolume)
			}
			nn := &TNote{
				Note:       &n,
				Octave:     &ob,
				Instrument: &inst,
				Volume:     &v,
			}
			s.PutTargetNote(nn)
		}
	case "X":
		n := "X"
		o := byte(1)
		nn := &TNote{
			Note:   &n,
			Octave: &o,
		}
		temp := s.targetPos
		s.PutTargetNote(nn)
		s.targetPos = temp
	case "V":
		// volume
		i, err := strconv.ParseInt(note[1:], 10, 32)
		if err != nil {
			return
		}
		if i >= 0 && int(i) < 16 {
			s.targetVolume = int(i)
		}
	case "L":
		// note length
		i, err := strconv.ParseInt(note[1:], 10, 32)
		if err != nil {
			return
		}
		if i >= 0 && int(i) < len(noteLengths) {
			s.targetNoteLength = noteLengths[int(i)]
		}
	case "R":
		// do a rest
		i, err := strconv.ParseInt(note[1:], 10, 32)
		if err != nil {
			return
		}
		if i >= 0 && int(i) < len(noteLengths) {
			s.targetPos += noteLengths[int(i)]
		}
	case "O":
		if s.targetHasEnding {
			break
		}
		// loop
		c := byte('b')
		cv := byte(0)
		nn := &TNote{
			Command:      &c,
			CommandValue: &cv,
		}
		if s.targetPos > 0 {
			s.targetPos--
		}
		s.MergeTargetNote(nn)
		s.targetHasEnding = true
	case "P":
		s.SetPlayMode(PMBoundPattern)
		s.targetIsPaused = true
	case "U":
		s.targetIsPaused = false
	}

}

func (s *TSong) PutTargetNote(nn *TNote) {
	s.targetPattern = s.targetPos / MaxTrackNotes
	subpos := s.targetPos % MaxTrackNotes

	if s.PatternList[s.targetPattern] == -1 {
		s.PatternList[s.targetPattern] = s.targetPattern
		s.Patterns[s.targetPattern] = NewPattern(120)
	}

	t := s.Patterns[s.targetPattern].Tracks[s.targetTrack]
	if t == nil {
		t = NewTrack()
		s.Patterns[s.targetPattern].Tracks[s.targetTrack] = t
	}

	t.Notes[subpos] = nn
	s.targetPos += s.targetNoteLength
}

func (s *TSong) MergeTargetNote(nn *TNote) {
	s.targetPattern = s.targetPos / MaxTrackNotes
	subpos := s.targetPos % MaxTrackNotes

	if s.PatternList[s.targetPattern] == -1 {
		s.PatternList[s.targetPattern] = s.targetPattern
		s.Patterns[s.targetPattern] = NewPattern(120)
	}

	t := s.Patterns[s.targetPattern].Tracks[s.targetTrack]
	if t == nil {
		t = NewTrack()
		s.Patterns[s.targetPattern].Tracks[s.targetTrack] = t
	}

	current := t.Notes[subpos]
	if current == nil {
		current = nn
	} else {
		current.Command = nn.Command
		current.CommandValue = nn.CommandValue
	}

	t.Notes[subpos] = current
	s.targetPos += s.targetNoteLength
}
