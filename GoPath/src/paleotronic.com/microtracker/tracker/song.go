package tracker

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
	"paleotronic.com/files"
	"paleotronic.com/log"
	"paleotronic.com/microtracker/mock"
	"paleotronic.com/octalyzer/assets"
	"paleotronic.com/restalgia"
	"paleotronic.com/utils"
)

const MaxTrackNotes = 64
const MaxPatternTracks = 6
const MaxSongPatterns = 256
const MaxSongPatternList = 1024
const MaxSongInstruments = 256
const MaxOperators = 8

type TPatternMode int

const (
	PMBoundPattern TPatternMode = iota
	PMLoopPattern
	PMLoopSong
	PMSingleSong
)

type TNote struct {
	Note         *string `bson:"N,omitempty"`
	Octave       *byte   `bson:"O,omitempty"`
	Instrument   *byte   `bson:"I,omitempty"`
	Volume       *byte   `bson:"V,omitempty"`
	Command      *byte   `bson:"C,omitempty"`
	CommandValue *byte   `bson:"c,omitempty"`
}

func (t *TNote) Copy() *TNote {
	c := &TNote{}
	if t.Note != nil {
		c.SetNote(t.GetNote())
	}
	if t.Octave != nil {
		c.SetOctave(t.GetOctave())
	}
	if t.Instrument != nil {
		c.SetInstrument(t.GetInstrument())
	}
	if t.Volume != nil {
		c.SetVolume(t.GetVolume())
	}
	if t.Command != nil {
		c.SetCommand(t.GetCommand())
	}
	if t.CommandValue != nil {
		c.SetCommandValue(t.GetCommandValue())
	}
	return c
}

func (t *TNote) StringNote() string {
	if t.Note == nil || t.Octave == nil {
		return "---"
	}
	n := *t.Note

	if n == "X" {
		return "xxx"
	}

	if len(n) == 1 {
		n += "-"
	}
	return fmt.Sprintf("%2s%d", n, *t.Octave)
}

func (t *TNote) StringInstrument() string {
	if t.Instrument == nil {
		return ".."
	}
	if t.Note != nil && *t.Note == "X" {
		return ".."
	}
	return fmt.Sprintf("%.2X", *t.Instrument)
}

func (t *TNote) StringVolume() string {
	if t.Volume == nil {
		return "--"
	}
	return fmt.Sprintf("%.2X", *t.Volume)
}

func (t *TNote) StringCommand() string {
	if t.Command == nil {
		return "..."
	}
	return fmt.Sprintf("%s%.2x", string(rune(*t.Command)-32), *t.CommandValue)
}

func (t *TNote) SetNote(note string) *TNote {
	if t == nil {
		t = &TNote{}
	}
	t.Note = &note
	return t
}

func (t *TNote) SetOctave(octave byte) *TNote {
	if t == nil {
		t = &TNote{}
	}
	t.Octave = &octave
	return t
}

func (t *TNote) SetVolume(volume byte) *TNote {
	if t == nil {
		t = &TNote{}
	}
	t.Volume = &volume
	return t
}

func (t *TNote) SetInstrument(v byte) *TNote {
	if t == nil {
		t = &TNote{}
	}
	t.Instrument = &v
	return t
}

func (t *TNote) SetCommand(v byte) *TNote {
	if t == nil {
		t = &TNote{}
	}
	t.Command = &v
	return t
}

func (t *TNote) SetCommandValue(v byte) *TNote {
	if t == nil {
		t = &TNote{}
	}
	t.CommandValue = &v
	return t
}

func (t *TNote) GetNote() string {
	if t.Note == nil {
		return ""
	}
	return *t.Note
}

func (t *TNote) GetOctave() byte {
	if t.Octave == nil {
		return 0
	}
	return *t.Octave
}

func (t *TNote) GetInstrument() byte {
	if t.Instrument == nil {
		return 0
	}
	return *t.Instrument
}

func (t *TNote) GetVolume() byte {
	if t.Volume == nil {
		return 0
	}
	return *t.Volume
}

func (t *TNote) GetCommand() byte {
	if t.Command == nil {
		return 0
	}
	return *t.Command
}

func (t *TNote) GetCommandValue() byte {
	if t.CommandValue == nil {
		return 0
	}
	return *t.CommandValue
}

func (t *TNote) IsEmpty() bool {
	if t == nil {
		return true
	}
	if t.Note == nil && t.Octave == nil && t.Instrument == nil && t.Volume == nil && t.Command == nil {
		return true
	}
	return false
}

type TTrack struct {
	Notes [MaxTrackNotes]*TNote `bson:"N,omitempty"`
}

func NewTrack() *TTrack {
	t := &TTrack{}
	// for i := range t.Notes {
	// 	t.Notes[i] = &TNote{}
	// }
	return t
}

func (t *TTrack) Clear() {
	for i, _ := range t.Notes {
		t.Notes[i] = nil
	}
}

func (t *TTrack) Copy() *TTrack {

	c := &TTrack{}
	for i, v := range t.Notes {
		if v != nil {
			c.Notes[i] = v.Copy()
		}
	}
	return c

}

func (t *TTrack) NextNote(i int) int {
	i++
	for i < MaxTrackNotes {
		if t.Notes[i] != nil && t.Notes[i].Note != nil {
			return i
		}
		i++
	}
	return -1
}

type TPattern struct {
	Speed  int                       `bson:"S,omitempty"`
	Tracks [MaxPatternTracks]*TTrack `bson:"T,omitempty"`
}

func NewPattern(speed int) *TPattern {
	p := &TPattern{
		Speed: speed,
	}
	for t := 0; t < MaxPatternTracks; t++ {
		p.Tracks[t] = NewTrack()
	}
	return p
}

func (t *TPattern) Copy() *TPattern {
	c := &TPattern{
		Speed: t.Speed,
	}
	for i, tr := range t.Tracks {
		if tr != nil {
			c.Tracks[i] = tr.Copy()
		}
	}
	return c
}

func (t *TPattern) Clear() {
	for _, tr := range t.Tracks {
		if tr != nil {
			tr.Clear()
		}
	}
}

type TVoiceConfig struct {
	UseTone     bool `bson:"UT,omitempty" json:"useTone,omitempty"`
	UseNoise    bool `bson:"UN,omitempty" json:"useNoise,omitempty"`
	NoisePeriod int  `bson:"NP,omitempty" json:"noiseFreq,omitempty"`
	UseEnv      bool `bson:"UE,omitempty" json:"useEnv,omitempty"`
	EnvCoarse   int  `bson:"EC,omitempty" json:"envCoarse,omitempty"`
	EnvFine     int  `bson:"EF,omitempty" json:"envFine,omitempty"`
	EnvShape    int  `bson:"ES,omitempty" json:"envShape,omitempty"`
	Amplitude   int  `bson:"AM,omitempty" json:"amp,omitempty"`
}

func NewVoiceConfig(
	useTone bool,
	useNoise bool,
	noiseFreq int,
	useEnv bool,
	envCoarse int,
	envFine int,
	envShape int,
	amp int,
) *TVoiceConfig {
	return &TVoiceConfig{
		useTone,
		useNoise,
		noiseFreq & 31,
		useEnv,
		envCoarse & 0xff,
		envFine & 0xff,
		envShape & 0x0f,
		amp & 0x0f,
	}
}

func (v *TVoiceConfig) String() string {
	return fmt.Sprintf(
		"useTone=%v,useNoise=%v,noisePeriod=$%.2x,useEnv=%v,envCoarse=$%.2x,envFine=$%.2x,envShape=$%x,volume=$%x",
		v.UseTone,
		v.UseNoise,
		v.NoisePeriod,
		v.UseEnv,
		v.EnvCoarse,
		v.EnvFine,
		v.EnvShape,
		v.Amplitude,
	)
}

type TInstrument struct {
	Name  string       `bson:"N,omitempty" json:"instrumentName,omitempty"`
	Voice TVoiceConfig `bson:"O,omitempty" json:"Oscillators,omitempty"`
}

func NewInstrument(name string, voice *TVoiceConfig) *TInstrument {
	if voice != nil {
		i := &TInstrument{
			Name:  name,
			Voice: *voice,
		}
		return i
	}
	i := &TInstrument{
		Name: name,
	}
	return i
}

func (i *TInstrument) String() string {
	if i == nil {
		return ""
	}
	return i.Voice.String()
}

func (i *TInstrument) Save(filename string) error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}

	p := files.GetPath(filename)
	f := files.GetFilename(filename)

	return files.WriteBytesViaProvider(p, f, data)
}

func (i *TInstrument) Load(filename string) error {
	pp := files.GetPath(filename)
	ff := files.GetFilename(filename)

	data, err := files.ReadBytesViaProvider(pp, ff)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data.Content, i)

	if i.Voice.Amplitude == 0 {
		i.Voice.Amplitude = 15
	}

	return err
}

type TSong struct {
	m                sync.Mutex
	Name             string                           `bson:"N,omitempty"`
	Tempo            int                              `bson:"T,omitempty"`
	Patterns         [MaxSongPatterns]*TPattern       `bson:"P,omitempty"`
	PatternList      [MaxSongPatternList]int          `bson:"p,omitempty"`
	Instruments      [MaxSongInstruments]*TInstrument `bson:"i,omitempty"`
	PatternListIndex int                              `bson:"l,omitempty"`
	PatternPos       int                              `bson:"m,omitempty"`
	lastInstrument   [MaxPatternTracks]string         `bson:"-"`
	Drv              *mock.MockDriver                 `bson:"-"`
	playing          bool                             `bson:"-"`
	running          bool                             `bson:"-"`
	PlayMode         TPatternMode                     `bson:"-"`
	TrackDisabled    [MaxPatternTracks]bool           `bson:"D,omitempty"`
	semiAdjustF      [MaxPatternTracks]float64
	semiAdjustC      [MaxPatternTracks]int
	lastNoteIndex    [MaxPatternTracks]int
	lastNoteOctave   [MaxPatternTracks]int
	//
	TimeCircuitsActivated   bool
	DestinationPatternIndex int
	DestinationPatternRow   int
	//
	undoPatterns [MaxSongPatterns]*TPattern

	targetTrack      int
	targetPattern    int
	targetPos        int
	targetNoteLength int
	targetVolume     int
	lastTargetEntry  time.Time
	targetIsPaused   bool
	targetHasEnding  bool
}

func NewSong(speed int, drv *mock.MockDriver) *TSong {
	s := &TSong{
		Tempo: speed,
		Drv:   drv,
		Name:  "untitled",
	}
	s.ClearPatternList()
	//s.InitInstruments()
	s.DefaultInstruments()
	s.InitPatterns(s.Tempo)
	s.PatternList[0] = 0
	return s
}

func (s *TSong) InitInstruments() {
	for i := range s.Instruments {
		s.Instruments[i] = NewInstrument("empty", nil)
	}
}

func (s *TSong) InitPatterns(speed int) {
	for i := range s.Patterns {
		s.Patterns[i] = NewPattern(speed)
	}
}

func (s *TSong) Stop() {
	if s.playing {
		s.playing = false
		time.Sleep(50 * time.Millisecond)
	}
	for i := 0; i < 6; i++ {
		s.Squelch(i)
	}
}

func (s *TSong) Start(mode TPatternMode) {
	for i := 0; i < MaxPatternTracks; i++ {
		s.Squelch(i)
		s.SetTrackVolume(i, 0)
	}
	s.Stop()
	s.SetPlayMode(mode)

	for i := 0; i < MaxPatternTracks; i++ {

		s.SendCommands(
			fmt.Sprintf(`
use mixer.voices.trk%d
set frequency 1
set volume 1
`, i),
		)

	}

	go s.player()
}

func (s *TSong) RememberPattern() {
	pnum := s.PatternList[s.PatternListIndex]
	s.undoPatterns[pnum] = s.Patterns[pnum].Copy()
}

func (s *TSong) UndoPattern() {
	pnum := s.PatternList[s.PatternListIndex]
	oldp := s.undoPatterns[pnum]
	if oldp == nil {
		return
	}
	s.undoPatterns[pnum] = s.Patterns[pnum]
	s.Patterns[pnum] = oldp
}

func (s *TSong) SetPlayMode(mode TPatternMode) {
	s.PlayMode = mode
	if s.PlayMode == PMBoundPattern {
		for i := 0; i < MaxPatternTracks; i++ {
			s.Squelch(i)
		}
	} else {
		s.ResetChannels()
	}
}

func (s *TSong) ResetChannels() {
	s.Drv.ResetRegs(0)
	s.Drv.ResetRegs(1)
}

func (s *TSong) ClearInstruments() {
	for i, _ := range s.Instruments {
		s.Instruments[i] = nil
	}
}

func (s *TSong) DefaultInstruments() {
	s.ClearInstruments()
	for i := 0; i < MaxSongInstruments; i++ {
		s.Instruments[i] = NewInstrument(
			"beep",
			NewVoiceConfig(
				true,
				false,
				0,
				false,
				0x02,
				0xff,
				0x00,
				15,
			),
		)
	}

	// try load from assets
	s.LoadDefaultInstruments()
}

func (s *TSong) ClearPatternList() {
	for i, _ := range s.PatternList {
		s.PatternList[i] = -1
	}
}

func (s *TSong) CurrentPatternPos() (*TPattern, int) {
	if s.PatternListIndex < 0 || s.PatternListIndex >= MaxSongPatternList {
		return nil, -1
	}
	pi := s.PatternList[s.PatternListIndex]
	if pi < 0 || pi > MaxSongPatterns {
		return nil, -1
	}
	p := s.Patterns[pi]
	if p == nil {
		return nil, -1
	}
	return p, s.PatternPos
}

func (s *TSong) TrackStep(i int) {
	p, pp := s.CurrentPatternPos()
	if p == nil {
		return
	}
	pp += i
	if pp >= MaxTrackNotes {
		pp = 0
	}
	if pp < 0 {
		pp = MaxTrackNotes - 1
	}
	s.PatternPos = pp
}

func (s *TSong) PatternAdvance() {
	p, pp := s.CurrentPatternPos()
	if p == nil {
		return
	}
	pp += 1
	switch s.PlayMode {
	case PMBoundPattern:
		if pp >= MaxTrackNotes {
			pp = MaxTrackNotes - 1
		}
		s.PatternPos = pp
	case PMLoopPattern:
		if pp >= MaxTrackNotes {
			pp = 0
		}
		s.PatternPos = pp
	case PMSingleSong:
		if pp >= MaxTrackNotes {
			s.PatternListIndex++
			s.PatternPos = 0
			p, pp = s.CurrentPatternPos()
			if p == nil {
				s.Stop() // nothing to play?
				inst0 := s.Instruments[0]
				*s = *NewSong(120, s.Drv)
				s.Instruments[0] = inst0
				s.Start(PMBoundPattern)
			}
		}
		s.PatternPos = pp
	case PMLoopSong:
		if pp >= MaxTrackNotes {
			s.PatternListIndex++
			s.PatternPos = 0
			p, pp = s.CurrentPatternPos()
			if p == nil {
				s.PatternListIndex = 0 // reset
				p, pp = s.CurrentPatternPos()
				if p == nil {
					s.Stop() // nothing to play?
				}
			} else {
				log.Printf("Moving to pattern list entry %.2x", s.PatternListIndex)
			}
		}
		s.PatternPos = pp
	}

}

func (s *TSong) JumpNext(index int) {
	s.PatternListIndex++
	s.PatternPos = index
	p, _ := s.CurrentPatternPos()
	if p == nil {
		s.PatternListIndex = 0 // reset
		p, _ = s.CurrentPatternPos()
		if p == nil {
			s.Stop() // nothing to play?
		}
	} else {
		log.Printf("Moving to pattern list entry %.2x", s.PatternListIndex)
	}
}

func (s *TSong) GetNotes() ([MaxPatternTracks]*TNote, bool) {
	s.m.Lock()
	defer s.m.Unlock()

	var notes [MaxPatternTracks]*TNote
	p, pp := s.CurrentPatternPos()
	if p == nil {
		return notes, false
	}
	for i, t := range p.Tracks {
		if !s.TrackDisabled[i] {
			notes[i] = t.Notes[pp]
		}
	}
	return notes, true
}

func (s *TSong) ResetEnv(track int) {
	s.Drv.ResetEnv(track)
}

func (s *TSong) ExecCommand(track int, cmd rune, param byte) bool {
	jump := false

	// t = tone coarse, u = tone fine adjust
	//

	switch cmd {
	case 'p':
		if s.PlayMode == PMSingleSong {
			s.Stop()
			inst0 := s.Instruments[0]
			*s = *NewSong(120, s.Drv)
			s.Instruments[0] = inst0
			s.Start(PMBoundPattern)
		} else {
			s.SetPlayMode(PMBoundPattern)
		}

	case 'w':
		s.SetTrackEnvPeriodCoarse(track, int(param)&0xff)
		s.ResetEnv(track)
	case 'x':
		s.SetTrackEnvPeriodFine(track, int(param)&0xff)
		s.ResetEnv(track)
	case 'n':
		s.SetTrackNoisePeriod(track, int(param)&0x1f)
	case 'e':
		v := int(param) & 1
		s.SetTrackEnvEnable(track, v == 1)
	case 'c':
		v := int(param)
		up := (v >> 4) & 0xf
		down := v & 0xf
		diff := up - down
		s.AdjustTrackVolume(track, diff)
	case 'f':
		v := int(param)
		if v > 0 {
			s.Tempo = v
		} else if v == 0 {
			s.SetPlayMode(PMBoundPattern)
		}
	case 'v':
		// set env shape --
		s.Drv.SetEnvelopeShape(byte(track/3), param)
	case 'j':
		v := int(param)
		if v < MaxTrackNotes {
			if s.PlayMode != PMBoundPattern {
				// s.JumpNext(v)
				// jump = true
				s.TimeCircuitsActivated = true
				if s.PlayMode == PMLoopSong || s.PlayMode == PMSingleSong {
					s.DestinationPatternIndex = s.PatternListIndex + 1
				} else {
					s.DestinationPatternIndex = s.PatternListIndex
				}
				s.DestinationPatternRow = v
			}

		}
	case 'b':
		v := int(param)
		if v < MaxSongPatternList {
			if s.PlayMode != PMBoundPattern {
				// s.PatternListIndex = v
				// s.PatternPos = 0
				// jump = true
				s.TimeCircuitsActivated = true
				if s.PlayMode == PMLoopSong || s.PlayMode == PMSingleSong {
					s.DestinationPatternIndex = v
				} else {
					s.DestinationPatternIndex = s.PatternListIndex
				}
				s.DestinationPatternRow = 0
			}
		}
	case 'y':
		// env period coarse - rel		v := int(param)
		v := int(param)
		up := (v >> 4) & 0xf
		down := v & 0xf
		diff := up - down
		s.AdjustEnvPeriodCoarse(track, diff)
	case 'z':
		// env period fine - rel
		v := int(param)
		up := (v >> 4) & 0xf
		down := v & 0xf
		diff := up - down
		s.AdjustEnvPeriodFine(track, diff)
	case 't':
		// tone period coarse - rel
		v := int(param)
		up := (v >> 4) & 0xf
		down := v & 0xf
		diff := up - down
		s.AdjustTonePeriodCoarse(track, diff)
	case 'u':
		// tone period fine - rel
		v := int(param)
		up := (v >> 4) & 0xf
		down := v & 0xf
		diff := up - down
		s.AdjustTonePeriodFine(track, diff)
	case 's':
		// toner, noise enable 00 = both off 10 = tone on, 01 = noise on, 11 = both on.
		s.SetVoiceEnableState(track, param)
	case 'r':
		// env period coarse - rel
		v := int(param)
		up := (v >> 4) & 0xf
		down := v & 0xf
		diff := up - down
		s.AdjustTonePeriodCoarse(track, diff)
		if diff < 0 {
			s.SetTrackTonePeriodFine(track, 0xff)
		} else {
			s.SetTrackTonePeriodFine(track, 0x00)
		}
	case 'q':
		// env period coarse - rel
		v := int(param)
		up := (v >> 4) & 0xf
		down := v & 0xf
		diff := up - down
		s.AdjustEnvPeriodCoarse(track, diff)
		if diff < 0 {
			s.SetTrackEnvPeriodFine(track, 0xff)
		} else {
			s.SetTrackEnvPeriodFine(track, 0x00)
		}
	}

	return jump
}

func (s *TSong) ProcessEffects(track int) {
	if s.semiAdjustC[track] > 0 {
		s.semiAdjustC[track]--
		cmd := fmt.Sprintf("use mixer.voices.trk%d\nadjust frequencyslide %f", track, s.semiAdjustF[track])
		s.SendCommands(cmd)
	}
}

func (s *TSong) PlayLine(trackLimit int) bool {
	jump := false
	if notes, ok := s.GetNotes(); ok {
		// can play back
		for i, n := range notes {
			if trackLimit != -1 && i != trackLimit {
				continue
			}
			if n != nil {

				skip := false //n.StringCommand() == "x00"

				if n.Note != nil && !skip {
					//str := fmt.Sprintf("%s%d", *n.Note, *n.Octave)

					no := int(*n.Octave)
					ni := restalgia.NT.NoteIndex(*n.Note)

					if no != -1 {
						// str := fmt.Sprintf("%s%d", *n.Note, *n.Octave)
						// ncmd = append(ncmd, fmt.Sprintf("set notes \"%s\"", str))
						if *n.Note == "X" {
							// silence voice
							s.Squelch(i)
						} else {
							if n.Instrument != nil {
								inst := s.Instruments[int(*n.Instrument)]
								if inst != nil {
									vol := byte(15)
									if n.Volume != nil {
										vol = *n.Volume
									}
									s.SendNote(i, inst, *n.Note, int(*n.Octave), vol)
								}
							}
						}

						s.lastNoteIndex[i] = ni
						s.lastNoteOctave[i] = no
					}
				} else if n.Volume != nil {
					// turn envelope off in this case since no note
					s.Drv.AppleVoiceEnvelope(i, false)
					s.SetTrackVolume(i, *n.Volume)
				}

				if n.Command != nil {
					jump = s.ExecCommand(i, rune(*n.Command), *n.CommandValue)
				}

				time.Sleep(1 * time.Millisecond)

			}
		}

		if s.TimeCircuitsActivated {
			s.TimeCircuitsActivated = false
			s.PatternListIndex = s.DestinationPatternIndex
			s.PatternPos = s.DestinationPatternRow
			jump = true
			// fix for empty pattern jump - return to start
			if s.PatternList[s.PatternListIndex] == -1 {
				s.PatternListIndex = 0
			}
		}

	}
	return jump
}

func (s *TSong) WipeInstMemory(t int) {
	s.lastInstrument[t] = ""
}

func (s *TSong) SendCommands(cmdbuffer string) {
	// if s.!= nil {
	// 	s.cmdbuffer)
	// }
}

func (s *TSong) player() {

	if s.running {
		s.playing = false
		for s.running {
			time.Sleep(time.Millisecond)
		}
	}

	var subcount int
	var doLine, movePattern bool

	tempogap := time.Minute / time.Duration(s.Tempo*4*4)

	s.playing = true
	s.running = true
	for s.playing {

		subcount = (subcount + 1) % 4
		doLine = (subcount == 0)
		movePattern = false

		switch s.PlayMode {
		case PMLoopPattern:
			if doLine {
				movePattern = !s.PlayLine(-1)
			}
		case PMLoopSong, PMSingleSong:
			if doLine {
				movePattern = !s.PlayLine(-1)
			}
		}

		tempogap = time.Minute / time.Duration(s.Tempo*4*4)
		time.Sleep(tempogap)
		if movePattern {
			s.PatternAdvance()
		}

		//log2.Printf("pattern pos: %d, tempogap: %d, subcount: %d", s.PatternPos, tempogap, subcount)

		for i := 0; i < MaxPatternTracks; i++ {
			s.ProcessEffects(i)
		}

	}

	s.running = false
}

func (s *TSong) Save(filename string) error {
	data, err := bson.Marshal(s)
	if err != nil {
		return err
	}

	data = utils.GZIPBytes(data)

	p := files.GetPath(filename)
	f := files.GetFilename(filename)

	return files.WriteBytesViaProvider(p, f, data)
}

func (s *TSong) Load(filename string) error {
	s.Stop()
	f := s.Drv
	defer func() {
		s.Drv = f
		s.Start(s.PlayMode)
	}()

	pp := files.GetPath(filename)
	ff := files.GetFilename(filename)

	data, err := files.ReadBytesViaProvider(pp, ff)
	if err != nil {
		return err
	}
	b := utils.UnGZIPBytes(data.Content)
	err = bson.Unmarshal(b, s)
	for _, inst := range s.Instruments {
		if inst != nil && inst.Voice.Amplitude == 0 {
			inst.Voice.Amplitude = 15
		}
	}
	s.PatternListIndex = 0
	s.PatternPos = 0

	// tmp := s.EncodeBinary()
	// log.Printf("Song encoded to memory is %d bytes", len(tmp))

	// s.EncodeASM("test.asm")

	return err
}

func (s *TSong) JumpPattern(pindex int) {
	s.DestinationPatternIndex = pindex
	s.DestinationPatternRow = 0
	s.TimeCircuitsActivated = true
}

func (s *TSong) IsPaused() bool {
	return s.PlayMode == PMBoundPattern
}

func (s *TSong) TogglePause() {
	log.Printf("playmode = %d", s.PlayMode)
	switch s.PlayMode {
	case PMBoundPattern:
		s.SetPlayMode(PMLoopSong)
	case PMLoopSong:
		s.SetPlayMode(PMBoundPattern)
	case PMLoopPattern:
		s.SetPlayMode(PMBoundPattern)
	}
	log.Printf("playmode new = %d", s.PlayMode)
}

func (s *TSong) LoadBuffer(r io.Reader) error {

	s.Stop()
	f := s.Drv
	defer func() {
		s.Drv = f
		s.Start(s.PlayMode)
	}()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	b := utils.UnGZIPBytes(data)
	err = bson.Unmarshal(b, s)
	for _, inst := range s.Instruments {
		if inst != nil && inst.Voice.Amplitude == 0 {
			inst.Voice.Amplitude = 15
		}
	}
	s.PatternListIndex = 0
	s.PatternPos = 0

	return err
}

func (m *TSong) SetTrackNoisePeriod(track int, period int) {
	m.Drv.SetVoiceNoisePeriod(track, byte(period))
}

func (m *TSong) SetTrackEnvPeriodCoarse(track int, period int) {
	m.Drv.SetVoiceEnvPeriodCoarse(track, byte(period))
}

func (m *TSong) SetTrackEnvPeriodFine(track int, period int) {
	m.Drv.SetVoiceEnvPeriodFine(track, byte(period))
}

func (m *TSong) SetTrackEnvEnable(track int, enabled bool) {
	m.Drv.AppleVoiceEnvelope(track, enabled)
}

func (m *TSong) SetTrackVolume(track int, volume byte) {
	m.Drv.ApplyVoiceVolume(track, volume)
}

func (m *TSong) Squelch(track int) {
	m.Drv.Squelch(track)
}

func (m *TSong) AdjustTrackVolume(track int, diff int) {
	m.Drv.AdjustVoiceVolume(track, diff)
}

func (m *TSong) AdjustEnvPeriodCoarse(track int, diff int) {
	m.Drv.AdjustEnvPeriodCoarse(track, diff)
}

func (m *TSong) SetTrackTonePeriodFine(track int, param byte) {
	m.Drv.SetTrackTonePeriodFine(track, param)
}

func (m *TSong) AdjustEnvPeriodFine(track int, diff int) {
	m.Drv.AdjustEnvPeriodFine(track, diff)
}

func (m *TSong) AdjustTonePeriodCoarse(track int, diff int) {
	m.Drv.AdjustTonePeriodCoarse(track, diff)
}

func (m *TSong) AdjustTonePeriodFine(track int, diff int) {
	m.Drv.AdjustTonePeriodFine(track, diff)
}

func (m *TSong) SetVoiceEnableState(track int, param byte) {
	m.Drv.SetVoiceEnableState(track, param)
}

func (m *TSong) SendNote(track int, inst *TInstrument, note string, octave int, volume byte) {
	// dummy use tone channel

	log.Printf("note = %s", note)

	if note == "X" {
		m.Drv.Squelch(track)
		return
	}

	log.Printf("Got note: %s", note)
	f, ok := mock.PSGNoteTable[note+utils.IntToStr(octave)]
	if ok {
		log.Printf("Applying freq of %f hz", f)
		m.Drv.ApplyVoice(
			track,
			inst.Voice.UseTone,
			mock.FreqHzToTonePeriod(f),
			inst.Voice.UseNoise,
			byte(inst.Voice.NoisePeriod),
			int(volume),
			inst.Voice.UseEnv,
			uint16(inst.Voice.EnvCoarse*256+inst.Voice.EnvFine),
			inst.Voice.EnvShape,
		)
	}
	m.WipeInstMemory(track)
}

func (s *TSong) LoadDefaultInstruments() {
	dir := "bootsystem/boot/templates/instruments"

	filelist, err := assets.AssetDir(dir)
	if err != nil {
		return
	}

	sort.Strings(filelist)

	for i, file := range filelist {

		s.Instruments[i].Voice = *NewVoiceConfig(false, false, 0, false, 0, 0, 0, 0)
		s.Instruments[i].Name = ""
		_ = s.Instruments[i].Load("/boot/templates/instruments/" + file)
		// if err != nil {
		// 	return
		// }
	}
}
