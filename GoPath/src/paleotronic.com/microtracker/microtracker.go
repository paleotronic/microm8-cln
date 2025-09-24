package microtracker

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/control"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
	"paleotronic.com/microtracker/mock"
	"paleotronic.com/microtracker/tracker"
	"paleotronic.com/octalyzer/bus"
	"paleotronic.com/utils"
)

const displayedTrackRows = 26

var waveforms = []string{
	"SINE",
	"TRIANGLE",
	"TRISAW",
	"SAWTOOTH",
	"SQUARE",
	"PULSE",
	"NOISE",
	"FM",
	"ADD",
}

var lfobindings = []string{
	"NONE",
	"FREQUENCY",
	"PULSEWIDTH",
	"PHASESHIFT",
	"VOLUME",
}

func getWaveformIndex(wf string) int {
	for i, v := range waveforms {
		if v == wf {
			return i
		}
	}
	return 0
}

func getBindingIndex(wf string) int {
	for i, v := range lfobindings {
		if v == wf {
			return i
		}
	}
	return 0
}

type SongParam int

const (
	spName SongParam = iota
	spList
)

type InstParam int

const (
	ipName InstParam = iota
	ipAmplitude
	ipUseTone
	ipUseNoise
	ipNoiseFreq
	ipUseEnv
	ipEnvCoarse
	ipEnvFine
	ipEnvShape
)

type MicroTrackerCtx int

const (
	ctxTrackEdit = iota
	ctxSongEdit
	ctxInstEdit
	ctxInstruct
)

type K struct {
	Note      string
	OctaveMod int
}

var noteKeys = map[rune]K{
	'`': K{"X", 0},
	'z': K{"C", 0},
	's': K{"C#", 0},
	'x': K{"D", 0},
	'd': K{"D#", 0},
	'c': K{"E", 0},
	'v': K{"F", 0},
	'g': K{"F#", 0},
	'b': K{"G", 0},
	'h': K{"G#", 0},
	'n': K{"A", 0},
	'j': K{"A#", 0},
	'm': K{"B", 0},
	',': K{"C", 1},
	'l': K{"C#", 1},
	'.': K{"D", 1},
	';': K{"D#", 1},
	'/': K{"E", 1},
	'q': K{"C", 1},
	'2': K{"C#", 1},
	'w': K{"D", 1},
	'3': K{"D#", 1},
	'e': K{"E", 1},
	'r': K{"F", 1},
	'5': K{"F#", 1},
	't': K{"G", 1},
	'6': K{"G#", 1},
	'y': K{"A", 1},
	'7': K{"A#", 1},
	'u': K{"B", 1},
	'i': K{"C", 2},
	'9': K{"C#", 2},
	'o': K{"D", 2},
	'0': K{"D#", 2},
	'p': K{"E", 2},
	'[': K{"F", 2},
	'=': K{"F#", 2},
	']': K{"G", 2},
}

func getNoteKeys() string {
	keys := ""
	for k, _ := range noteKeys {
		keys += string(k)
	}
	return keys
}

var hexKeys = "0123456789abcdefABCDEF"

var validKeys = []string{
	getNoteKeys(),                 // note keys
	"b#-",                         // sharps or flats
	"0123456789",                  // octave
	"0123456789abcdefABCDEF",      // inst * 16
	"0123456789abcdefABCDEF",      // inst
	"0123456789abcdefABCDEF",      // vol * 16
	"0123456789abcdefABCDEF",      // vol
	"abcdefghijklmnopqrstuvwxyz ", // command
	"0123456789abcdefABCDEF",      // cval * 16
	"0123456789abcdefABCDEF",      // cval
}

type MicroTracker struct {
	sync.Mutex
	events            []*servicebus.ServiceBusRequest
	Int               interfaces.Interpretable
	TrackFG           int
	TrackBG           int
	TrackShade        int
	TrackFGAlt        int
	TrackBGAlt        int
	TrackShadeAlt     int
	TrackOffset       int
	NoteFG            int
	InvalidFG         int
	InvalidBG         int
	NoteInstFG        int
	NoteVolFG         int
	NoteCmdFG         int
	BG, FG            int
	Song              *tracker.TSong
	running           bool
	ctx               MicroTrackerCtx
	pX, pY, pH        int
	sX, sY, sH, sW    int
	cTrack, cTrackSub int
	Octave            int
	Inst              byte
	Volume            byte
	slXOffset         int
	viewPLIndex       int
	viewPLSubIndex    int
	ReadOnlyTrack     bool
	CurrentFile       string
	clipTrack         *tracker.TTrack
	clipPattern       *tracker.TPattern
	InstParam         InstParam
	SongParam         SongParam
	refresh           bool
	Drv               *mock.MockDriver
	//
	lastHCMode bool
}

// C#4iivv...

func NewMicroTracker(e interfaces.Interpretable) *MicroTracker {
	m := &MicroTracker{
		TrackOffset: 10,
		BG:          5,
		FG:          0,
		NoteFG:      15,
		NoteInstFG:  12,
		NoteVolFG:   13,
		NoteCmdFG:   14,
		TrackFG:     15,
		TrackBG:     2,
		TrackFGAlt:  15,
		TrackBGAlt:  6,
		InvalidFG:   9,
		InvalidBG:   8,
		pH:          displayedTrackRows,
		pX:          10,
		pY:          20,
		sW:          78,
		sH:          2,
		sX:          1,
		sY:          17,
		cTrack:      0,
		cTrackSub:   0,
		Octave:      4,
		Volume:      0x80,
		CurrentFile: "untitled.sng",
		InstParam:   ipName,
		SongParam:   spList,
		refresh:     true,
		Int:         e,
		Drv:         mock.New(e, 0xc400),
		events:      make([]*servicebus.ServiceBusRequest, 0, 16),
	}
	if settings.HighContrastUI {
		m.BG = 15
		m.FG = 0
		m.NoteFG = 15
		m.NoteInstFG = 15
		m.NoteVolFG = 15
		m.NoteCmdFG = 15
		m.TrackFG = 15
		m.TrackBG = 0
		m.TrackFGAlt = 0
		m.TrackBGAlt = 15
		m.InvalidFG = 15
		m.InvalidBG = 0
	}
	s := tracker.NewSong(120, m.Drv)
	m.Song = s
	m.Inst = 0
	return m
}

const logoStr = `\u047B\u0479\u042F\u0479\u0468\u0476\u042E\u0479\u042E\u0479\u042E\u0475\u042F\u0475\u042E\u0479\u042E\u0479\u042E\u0475\u046B\u047B\u0475\u042E\u0479\u042E\u0450
\u042F \u042F \u0468\u042F\u042E\u046E\u0479\u046E\u042E\u0468\u042F\u046E\u0479\u042F\u042E\u046B\u044F\u0450\u0471\u0468\u046E\u042E\u042F\u0410\u047B
\u042F \u042E \u0468\u0443\u0443\u0443\u0443\u0443\u0443\u0473\u042F\u046E\u0464\u047B \u046B\u044C\u0470|\u0479\u042F\u0470\u047B \u046E
\u0479 \u0479 \u0479\u0479\u0479\u0479\u0479\u0479\u0479\u0479\u0479\u0479 \u0479 \u0479\u0479\u0479\u0479 \u0479\u0479\u0479\u0479\u0479`

func (m *MicroTracker) CheckPalette() {
	if m.lastHCMode != settings.HighContrastUI {
		m.lastHCMode = settings.HighContrastUI
		//
		if settings.HighContrastUI {
			m.BG = 15
			m.FG = 0
			m.NoteFG = 15
			m.NoteInstFG = 15
			m.NoteVolFG = 15
			m.NoteCmdFG = 15
			m.TrackFG = 15
			m.TrackBG = 0
			m.TrackFGAlt = 0
			m.TrackBGAlt = 15
			m.InvalidFG = 15
			m.InvalidBG = 0
		} else {
			m.BG = 5
			m.FG = 0
			m.NoteFG = 15
			m.NoteInstFG = 12
			m.NoteVolFG = 13
			m.NoteCmdFG = 14
			m.TrackFG = 15
			m.TrackBG = 2
			m.TrackFGAlt = 15
			m.TrackBGAlt = 6
			m.InvalidFG = 9
			m.InvalidBG = 8
		}

		m.FullRefresh()
	}
}

func (m *MicroTracker) FullRefresh() {
	ent := m.Int
	m.DrawScreen(ent)
	switch m.ctx {
	case ctxInstEdit:
		m.DrawTrackStatusDisplay(ent)
		m.DrawPatternDisplay(ent)
		m.DrawSongEditDisplay(ent)
		m.DrawInstrumentDisplay(ent)
	case ctxTrackEdit:
		m.DrawPatternDisplay(ent)
		m.DrawSongEditDisplay(ent)
		m.DrawInstrumentDisplay(ent)
		m.DrawTrackStatusDisplay(ent)
	case ctxSongEdit:
		m.DrawPatternDisplay(ent)
		m.DrawInstrumentDisplay(ent)
		m.DrawTrackStatusDisplay(ent)
		m.DrawSongEditDisplay(ent)
	}
	apple2helpers.TEXT(ent).FullRefresh()
	m.refresh = true
}

func (m *MicroTracker) DrawLogo(ent interfaces.Interpretable) {
	txt := apple2helpers.TEXT(ent)
	txt.HideCursor()
	txt.FGColor = uint64(m.FG)
	txt.BGColor = uint64(m.BG)
	lines := strings.Split(utils.Unescape(logoStr), "\n")
	for i, l := range lines {
		txt.GotoXY(1, 1+i)
		txt.PutStr(l)
	}
}

func (m *MicroTracker) ShowHelp() {
	apple2helpers.MonitorPanel(m.Int, false)
	hc := control.NewHelpController(m.Int, "microTracker Help", settings.HelpBase, "microtracker/quickhelp")
	hc.Do(m.Int)
	apple2helpers.MonitorPanel(m.Int, true)
	m.DrawScreen(m.Int)
	m.refresh = true
}

func (m *MicroTracker) NewSong() {
	resp := strings.ToLower(InputPopup(
		m.Int,
		"Erase current song?",
		"Erase current song? (y/n)",
		"y",
	))
	if resp == "y" {
		m.Song.Stop()
		m.Song = tracker.NewSong(120, m.Drv)
		m.Song.Start(tracker.PMBoundPattern)
	}
}

func (m *MicroTracker) DrawTrackStatusDisplay(ent interfaces.Interpretable) {

	if !m.refresh {
		return
	}

	_, pp := m.Song.CurrentPatternPos()

	txt := apple2helpers.TEXT(ent)
	txt.HideCursor()

	txt.FGColor = uint64(m.FG)
	txt.BGColor = uint64(m.BG)
	txt.Font = 0
	txt.Shade = 0
	txt.GotoXY(0, 6)
	wm := "           "
	if m.ReadOnlyTrack {
		wm = "(read-only)"
	}
	txt.PutStr(fmt.Sprintf(" Pattern    : %.2x %s\r\n", m.Song.PatternList[m.Song.PatternListIndex], wm))
	txt.PutStr(fmt.Sprintf(" Pattern Pos: %.2x  \r\n", pp))
	txt.PutStr(fmt.Sprintf(" Volume     : %.2x  \r\n", m.Volume))
	txt.PutStr(fmt.Sprintf(" Instrument : %.2x  \r\n", m.Inst))
	txt.PutStr(fmt.Sprintf(" Octave     : %.2x  \r\n", m.Octave))
	txt.PutStr(fmt.Sprintf(" Tempo      : %.2x  \r\n", m.Song.Tempo))

}

func (m *MicroTracker) DrawInstrumentDisplay(ent interfaces.Interpretable) {

	if !m.refresh {
		return
	}

	i := m.Song.Instruments[int(m.Inst)]

	// //if len(i.OSC) == 0 {
	// return
	// //}

	txt := apple2helpers.TEXT(ent)
	txt.HideCursor()
	txt.FGColor = uint64(m.FG)
	txt.BGColor = uint64(m.BG)
	txt.Font = 0
	txt.Shade = 0
	txt.GotoXY(40, 1)
	txt.PutStr(fmt.Sprintf(" Instrument  : %.2x   \r\n ", m.Inst))
	txt.GotoXY(40, 3)
	txt.PutStr(fmt.Sprintf(" Name        : %-12s  ", i.Name))
	txt.GotoXY(40, 4)
	txt.PutStr(fmt.Sprintf(" Amplitude   : %.2x  ", i.Voice.Amplitude))
	txt.GotoXY(40, 5)
	txt.PutStr(fmt.Sprintf(" Tone On     : %-5v  ", i.Voice.UseTone))
	txt.GotoXY(40, 6)
	txt.PutStr(fmt.Sprintf(" Noise On    : %-5v  ", i.Voice.UseNoise))
	txt.GotoXY(40, 7)
	txt.PutStr(fmt.Sprintf(" Noise Freq. : %.2x  ", i.Voice.NoisePeriod))
	txt.GotoXY(40, 8)
	txt.PutStr(fmt.Sprintf(" Use Env     : %-5v  ", i.Voice.UseEnv))
	txt.GotoXY(40, 9)
	txt.PutStr(fmt.Sprintf(" Env Coarse. : %.2x  ", i.Voice.EnvCoarse))
	txt.GotoXY(40, 10)
	txt.PutStr(fmt.Sprintf(" Env Fine    : %.2x  ", i.Voice.EnvFine))
	txt.GotoXY(40, 11)
	txt.PutStr(fmt.Sprintf(" Env Shape   : %.2x  ", i.Voice.EnvShape))

	if m.ctx == ctxInstEdit {
		txt.GotoXY(55, 3+int(m.InstParam))
		txt.ShowCursor()
	}
}

func (m *MicroTracker) DrawSongEditDisplay(ent interfaces.Interpretable) {

	if !m.refresh {
		return
	}

	//
	//if m.Song.PlayMode == tracker.PMLoopSong {
	m.viewPLIndex = m.Song.PatternListIndex
	//	}

	numSlots := m.sW / 2
	plIndex := m.viewPLIndex
	loPos := m.slXOffset
	hiPos := loPos + numSlots - 1
	for plIndex < m.slXOffset && m.slXOffset > 0 {
		m.slXOffset--
		loPos = m.slXOffset
		hiPos = loPos + numSlots - 1
	}
	for plIndex > hiPos {
		m.slXOffset++
		loPos = m.slXOffset
		hiPos = loPos + numSlots - 1
	}
	cOffset := 2*(plIndex-loPos) + m.viewPLSubIndex

	txt := apple2helpers.TEXT(ent)
	txt.HideCursor()

	txt.GotoXY(m.sX, m.sY)
	txt.FGColor = uint64(m.FG)
	txt.BGColor = uint64(m.BG)
	txt.PutStr("Song: ")
	if m.SongParam == spName {
		txt.FGColor = uint64(m.BG)
		txt.BGColor = uint64(m.FG)
	}
	txt.PutStr(fmt.Sprintf("%-40s", m.Song.Name))
	txt.GotoXY(m.sX, m.sY+1)
	for i := 0; i < numSlots; i++ {
		index := m.slXOffset + i

		if index%2 == 0 {
			txt.FGColor = uint64(m.TrackFG)
			txt.BGColor = uint64(m.TrackBG)
			if settings.HighContrastUI {
				txt.Shade = 0
			}
		} else {
			txt.FGColor = uint64(m.TrackFGAlt)
			txt.BGColor = uint64(m.TrackBGAlt)
			if settings.HighContrastUI {
				txt.Shade = 2
			}
		}

		pnum := m.Song.PatternList[index]
		if index == m.viewPLIndex {
			txt.Shade = 0
		} else {
			txt.Shade = 2
		}
		var label string
		if pnum == -1 {
			txt.Shade += 2
			label = "--"
			if index%2 == 0 {
				txt.FGColor = uint64(m.InvalidFG)
				txt.BGColor = uint64(m.InvalidBG)
			} else {
				txt.FGColor = uint64(m.InvalidBG)
				txt.BGColor = uint64(m.InvalidFG)
			}
		} else {
			label = fmt.Sprintf("%.2x", pnum)
		}
		txt.PutStr(label)
	}

	txt.Shade = 0
	txt.FGColor = uint64(m.FG)
	txt.BGColor = uint64(m.BG)
	txt.PutStr(" ")

	if m.ctx == ctxSongEdit {
		if m.SongParam == spList {
			txt.GotoXY(m.sX+cOffset, m.sY+1)
		} else {
			txt.GotoXY(m.sX+6+len(m.Song.Name), m.sY)
		}
	}

	txt.ShowCursor()

}

func (m *MicroTracker) DrawPatternDisplay(ent interfaces.Interpretable) {

	if !m.refresh {
		return
	}

	currentPattern, patternIndex := m.Song.CurrentPatternPos()

	middle := m.pH/2 - 1

	txt := apple2helpers.TEXT(ent)
	txt.HideCursor()

	var smap [displayedTrackRows]int
	for i, _ := range smap {
		smap[i] = i + patternIndex - middle
	}

	dr := 0
	for r := 0; r < displayedTrackRows; r++ {
		txt.GotoXY(m.pX-2, m.pY+dr)
		tidx := smap[r]
		txt.FGColor = uint64(m.FG)
		txt.BGColor = uint64(m.BG)
		txt.Shade = 0
		if r == middle {
			txt.Font = 1
		} else {
			txt.Font = 0
		}
		if tidx < 0 || tidx >= tracker.MaxTrackNotes {
			txt.PutStr("  ")
		} else {
			txt.PutStr(fmt.Sprintf("%.2x", tidx))
		}

		for t := 0; t < tracker.MaxPatternTracks; t++ {

			if r == middle {
				switch t % 2 {
				case 1:
					txt.BGColor = 0
				case 0:
					txt.BGColor = 0
				}
				txt.Shade = 0
				txt.Font = 1
			} else {
				switch t % 2 {
				case 0:
					txt.BGColor = uint64(m.TrackBG)
				case 1:
					txt.BGColor = uint64(m.TrackBGAlt)

				}
				txt.Shade = 0
				if settings.HighContrastUI {
					txt.Shade = uint64(r % 2)
				}
				txt.Font = 0
			}

			if m.Song.TrackDisabled[t] {
				txt.Shade = 5
			}

			if currentPattern == nil {
				switch t % 2 {
				case 0:
					txt.BGColor = uint64(m.InvalidBG)
				case 1:
					txt.BGColor = uint64(m.InvalidFG)
				}
				txt.PutStr("          ")
				continue
			}

			nidx := smap[r]

			if nidx < 0 || nidx >= tracker.MaxTrackNotes {
				txt.Shade = 6
				txt.PutStr("          ")
				continue
			}

			if nidx%4 == 0 {
				txt.Shade += 1
			}

			// actual data
			var n *tracker.TNote
			if currentPattern != nil {
				n = currentPattern.Tracks[t].Notes[nidx]
			}
			if n == nil {
				n = &tracker.TNote{}
			}

			if t%2 == 1 && settings.HighContrastUI && r != middle && currentPattern != nil {
				txt.FGColor = uint64(0)
				txt.PutStr(n.StringNote())
				txt.FGColor = uint64(0)
				txt.PutStr(n.StringInstrument())
				txt.FGColor = uint64(0)
				txt.PutStr(n.StringVolume())
				txt.FGColor = uint64(0)
				txt.PutStr(n.StringCommand())
				txt.FGColor = uint64(m.TrackFG)
			} else {
				txt.FGColor = uint64(m.NoteFG)
				txt.PutStr(n.StringNote())
				txt.FGColor = uint64(m.NoteInstFG)
				txt.PutStr(n.StringInstrument())
				txt.FGColor = uint64(m.NoteVolFG)
				txt.PutStr(n.StringVolume())
				txt.FGColor = uint64(m.NoteCmdFG)
				txt.PutStr(n.StringCommand())
				txt.FGColor = uint64(m.TrackFG)
			}
		}

		// position
		dr++
		if r == middle {
			dr++
		}
	}

	if m.ctx == ctxTrackEdit {
		txt.GotoXY(m.pX+10*m.cTrack+m.cTrackSub, m.pY+middle)
	}

	txt.Shade = 0

	txt.ShowCursor()
}

func (m *MicroTracker) DrawScreen(ent interfaces.Interpretable) {
	txt := apple2helpers.TEXT(ent)
	txt.HideCursor()

	txt.FGColor = uint64(m.FG)
	txt.BGColor = uint64(m.BG)

	txt.Shade = 0
	txt.Font = 0
	txt.ClearScreen()
	txt.GotoXY(1, 47)
	txt.PutStr("microTracker (c) 2018 Paleotronic. (Ctrl+Shift+H for help)")

	m.DrawLogo(ent)
}

func (m *MicroTracker) clearTrack() {
	pp, _ := m.Song.CurrentPatternPos()
	pp.Tracks[m.cTrack].Clear()
}

func (m *MicroTracker) clearPattern() {
	pp, _ := m.Song.CurrentPatternPos()
	pp.Clear()
}

func (m *MicroTracker) copyPattern() {
	pp, _ := m.Song.CurrentPatternPos()
	m.clipPattern = pp.Copy()
}

func (m *MicroTracker) copyTrack() {
	pp, _ := m.Song.CurrentPatternPos()
	m.clipTrack = pp.Tracks[m.cTrack].Copy()
}

func (m *MicroTracker) pasteTrack() {
	pp, _ := m.Song.CurrentPatternPos()
	if m.clipTrack != nil {
		pp.Tracks[m.cTrack] = m.clipTrack.Copy()
	}
}

func (m *MicroTracker) pastePattern() {
	ival := m.Song.PatternList[m.Song.PatternListIndex]
	if m.clipPattern != nil && ival != -1 {
		m.Song.Patterns[ival] = m.clipPattern.Copy()
	}
}

func (m *MicroTracker) newPatternAllocate() {
	num := -1
	for i := 0; i < tracker.MaxSongPatternList; i++ {
		if m.Song.Patterns[i] == nil {
			num = i
			break
		}
	}
	if num != -1 {
		m.Song.Patterns[num] = tracker.NewPattern(m.Song.Tempo)
		m.Song.PatternList[m.viewPLIndex] = num
		m.Song.PatternListIndex = m.viewPLIndex
		m.Song.PatternPos = 0
	}
}

func (m *MicroTracker) incPattern() {
	ival := m.Song.PatternList[m.viewPLIndex]
	ival++
	if ival >= tracker.MaxSongPatterns {
		ival = 0
	}
	m.Song.PatternList[m.viewPLIndex] = ival
}

func (m *MicroTracker) decPattern() {
	ival := m.Song.PatternList[m.viewPLIndex]
	ival--
	if ival < 0 {
		ival = tracker.MaxSongPatterns - 1
	}
	m.Song.PatternList[m.viewPLIndex] = ival
}

func (m *MicroTracker) showPattern() {
	ival := m.Song.PatternList[m.viewPLIndex]
	if ival != -1 && m.Song.Patterns[ival] == nil {
		m.Song.Patterns[ival] = tracker.NewPattern(m.Song.Tempo)
	}
	if ival != -1 {
		// goto pattern
		m.Song.PatternListIndex = m.viewPLIndex
		m.Song.PatternPos = 0
	}
}

func (m *MicroTracker) deletePattern() {
	for i := m.viewPLIndex + 1; i < tracker.MaxSongPatternList; i++ {
		m.Song.PatternList[i-1] = m.Song.PatternList[i]
	}
}

func (m *MicroTracker) insertBlankPattern() {
	for i := tracker.MaxSongPatternList - 2; i >= m.viewPLIndex; i-- {
		m.Song.PatternList[i+1] = m.Song.PatternList[i]
	}
	m.Song.PatternList[m.viewPLIndex] = -1
}

func (m *MicroTracker) insertPattern() {
	ival := m.Song.PatternList[m.Song.PatternListIndex]
	for i := tracker.MaxSongPatternList - 2; i >= m.viewPLIndex; i-- {
		m.Song.PatternList[i+1] = m.Song.PatternList[i]
	}
	m.Song.PatternList[m.viewPLIndex] = ival
}

func (m *MicroTracker) PlayPattern() {
	switch m.Song.PlayMode {
	case tracker.PMLoopPattern:
		m.Song.SetPlayMode(tracker.PMBoundPattern)
	default:
		m.Song.PatternPos = 0
		m.Song.SetPlayMode(tracker.PMLoopPattern)
	}
}

func (m *MicroTracker) PlaySongFromSOP() {
	switch m.Song.PlayMode {
	case tracker.PMLoopSong:
		m.Song.SetPlayMode(tracker.PMBoundPattern)
	default:
		m.Song.PatternPos = 0
		m.Song.SetPlayMode(tracker.PMLoopSong)
	}
}

func (m *MicroTracker) PlaySong() {
	switch m.Song.PlayMode {
	case tracker.PMLoopSong:
		m.Song.SetPlayMode(tracker.PMBoundPattern)
	default:
		m.Song.SetPlayMode(tracker.PMLoopSong)
	}
}

func GetCRTLine(ent interfaces.Interpretable, promptString string, def string) string {

	txt := apple2helpers.TEXT(ent)

	command := def
	collect := true

	cb := ent.GetProducer().GetMemoryCallback(ent.GetMemIndex())

	txt.PutStr(promptString)
	txt.PutStr(command)

	if cb != nil {
		cb(ent.GetMemIndex())
	}

	for collect {
		if cb != nil {
			cb(ent.GetMemIndex())
		}
		txt.ShowCursor()
		for ent.GetMemory(49152) < 128 {
			time.Sleep(5 * time.Millisecond)
			if cb != nil {
				cb(ent.GetMemIndex())
			}
		}
		txt.HideCursor()
		ch := rune(ent.GetMemory(49152) & 0xff7f)
		ent.SetMemory(49168, 0)
		switch ch {
		case 27:
			return ""
		case 10:
			{
				//txt.SetSuppressFormat(true)
				txt.PutStr("\r\n")
				//txt.SetSuppressFormat(false)
				return command
			}
		case 13:
			{
				//txt.SetSuppressFormat(true)
				txt.PutStr("\r\n")
				//txt.SetSuppressFormat(false)
				return command
			}
		case 127:
			{
				if len(command) > 0 {
					command = utils.Copy(command, 1, len(command)-1)
					txt.Backspace()
					//txt.SetSuppressFormat(true)
					txt.PutStr(" ")
					//txt.SetSuppressFormat(false)
					txt.Backspace()
					if cb != nil {
						cb(ent.GetMemIndex())
					}
				}
				break
			}
		default:
			{

				//txt.SetSuppressFormat(true)
				txt.Put(rune(ch))
				//txt.SetSuppressFormat(false)

				if cb != nil {
					cb(ent.GetMemIndex())
				}

				command = command + string(ch)
				break
			}
		}
	}

	if cb != nil {
		cb(ent.GetMemIndex())
	}

	return command

}

func InputPopup(ent interfaces.Interpretable, title string, message string, def string) string {

	apple2helpers.TextAddWindow(ent, "std", 0, 0, 79, 47)

	if settings.HighContrastUI {
		apple2helpers.SetFGColor(ent, 15)
		apple2helpers.SetBGColor(ent, 0)
	} else {
		apple2helpers.SetFGColor(ent, 15)
		apple2helpers.SetBGColor(ent, 2)
	}

	apple2helpers.TextDrawBox(
		ent,
		14,
		20,
		50,
		5,
		title,
		true,
		true,
	)

	ent.PutStr(title + "\r\n")

	s := GetCRTLine(ent, message+": ", def)

	apple2helpers.TextUseWindow(ent, "std")

	return s

}

func InfoPopup(ent interfaces.Interpretable, title string, message string, delay time.Duration) {

	apple2helpers.TextAddWindow(ent, "std", 0, 0, 79, 47)

	if settings.HighContrastUI {
		apple2helpers.SetFGColor(ent, 15)
		apple2helpers.SetBGColor(ent, 0)
	} else {
		apple2helpers.SetFGColor(ent, 15)
		apple2helpers.SetBGColor(ent, 2)
	}

	content := []string{strings.Trim(title, "\r\n")}
	lines := strings.Split(message, "\r\n")
	content = append(content, lines...)

	txt := apple2helpers.TEXT(ent)
	txt.HideCursor()
	txt.DrawTextBoxAuto(content, true, true)

	apple2helpers.TextUseWindow(ent, "std")

	time.Sleep(delay)
	txt.ShowCursor()

	return

}

var menufunc func(e interfaces.Interpretable)

func SetMenuHook(f func(e interfaces.Interpretable)) {
	menufunc = f
}

func (m *MicroTracker) Run(ent interfaces.Interpretable) {

	settings.MicroTrackerEnabled[ent.GetMemIndex()] = true

	ent.StopMusic()
	ent.ParseImm("@music.stop{}")
	for i := 0; i < 6; i++ {
		m.Song.Squelch(i)
	}

	wd := ent.GetWorkDir()
	if wd != "" {
		m.CurrentFile = "/" + strings.Trim(wd, "/") + "/untitled.sng"
	} else {
		m.CurrentFile = "/local/untitled.sng"
	}

	servicebus.Subscribe(
		ent.GetMemIndex(),
		servicebus.TrackerLoadSong,
		m,
	)

	apple2helpers.TextSaveScreen(ent)
	settings.DisableMetaMode[ent.GetMemIndex()] = true

	defer func() {
		apple2helpers.MonitorPanel(ent, false)
		apple2helpers.TextRestoreScreen(ent)
		settings.DisableMetaMode[ent.GetMemIndex()] = false
		servicebus.Unsubscribe(ent.GetMemIndex(), m)
		for i := 0; i < 6; i++ {
			m.Song.Squelch(i)
		}
		settings.MicroTrackerEnabled[ent.GetMemIndex()] = false
	}()

	apple2helpers.MonitorPanel(ent, true)
	apple2helpers.TEXTMAX(ent)
	apple2helpers.GFXDisable(ent)

	m.running = true
	m.Song.Start(tracker.PMBoundPattern)

	m.DrawScreen(ent)

	defer m.Song.Stop()

	var pi = m.Song.PatternPos + 64*m.Song.PatternListIndex

	for m.running {

		m.CheckPalette()

		m.ServiceBusProcessPending()

		if pi != m.Song.PatternPos+64*m.Song.PatternListIndex {
			m.refresh = true
		}

		pi = m.Song.PatternPos + 64*m.Song.PatternListIndex

		if ent.GetMemoryMap().IntGetSlotMenu(ent.GetMemIndex()) {
			if menufunc != nil {
				menufunc(ent)
			}
			ent.GetMemoryMap().IntSetSlotMenu(ent.GetMemIndex(), false)
			m.refresh = true
		}

		if m.Int.GetMemoryMap().IntGetSlotInterrupt(m.Int.GetMemIndex()) {
			apple2helpers.TextSaveScreen(m.Int)
			m.Int.DoCatalog()
			apple2helpers.MonitorPanel(m.Int, true)
			m.Int.GetMemoryMap().IntSetSlotInterrupt(m.Int.GetMemIndex(), false)
			apple2helpers.TextRestoreScreen(m.Int)
			bus.StartDefault()
		}

		if m.Int.GetMemoryMap().IntGetSlotRestart(m.Int.GetMemIndex()) {
			break
		}

		switch m.ctx {
		case ctxInstEdit:
			m.DrawTrackStatusDisplay(ent)
			m.DrawPatternDisplay(ent)
			m.DrawSongEditDisplay(ent)
			m.DrawInstrumentDisplay(ent)
			m.refresh = false
			if ch := ent.GetMemoryMap().KeyBufferGet(ent.GetMemIndex()); ch != 0 {
				m.refresh = true
				switch {
				case ch == vduconst.SHIFT_CSR_LEFT:
					m.Inst--
					for m.Song.Instruments[int(m.Inst)] == nil && m.Inst > 0 {
						m.Inst--
					}
				case ch == vduconst.SHIFT_CSR_RIGHT:
					m.Inst++
					for m.Song.Instruments[int(m.Inst)] == nil {
						m.Inst++
					}
				case ch == vduconst.SHIFT_CTRL_N:
					m.NewSong()
				case ch == vduconst.SHIFT_CTRL_H:
					m.ShowHelp()
				case ch == vduconst.SHIFT_CTRL_E:
					m.ExportSongCombinedWLA()
				case ch == vduconst.SHIFT_CTRL_F:
					m.ExportSongWLA()
				case ch == vduconst.SHIFT_CTRL_A:
					m.ExportSongASM()
				case ch == vduconst.SHIFT_CTRL_S:
					inst := m.Song.Instruments[m.Inst]
					name := InputPopup(ent, "Save instrument", "\r\nEnter filename", inst.Name+".snd")
					if name != "" {
						if !strings.HasPrefix(name, "/") {
							name = "/" + strings.Trim(ent.GetWorkDir(), "/") + "/" + name
						}
						inst.Save(name)
					}
				case ch == vduconst.SHIFT_CTRL_L:
					inst := m.Song.Instruments[m.Inst]
					name := InputPopup(ent, "Load instrument", "\r\nEnter filename", inst.Name+".snd")
					if name != "" {
						if !strings.HasPrefix(name, "/") {
							name = "/" + strings.Trim(ent.GetWorkDir(), "/") + "/" + name
						}
						inst.Load(name)
					}
				case ch == 9:
					m.ctx = ctxTrackEdit
				case ch == vduconst.CSR_UP:
					i := m.Song.Instruments[m.Inst]
					log.Printf("instrument after loading... %v", i)
					m.InstParam--
					if m.InstParam < ipName {
						m.InstParam = ipEnvShape
					}
				case ch == vduconst.CSR_DOWN:
					m.InstParam++
					if m.InstParam > ipEnvShape {
						m.InstParam = ipName
					}
				case ch == vduconst.CSR_RIGHT:
					inst := m.Song.Instruments[int(m.Inst)]
					switch m.InstParam {
					case ipUseTone:
						inst.Voice.UseTone = !inst.Voice.UseTone
					case ipUseNoise:
						inst.Voice.UseNoise = !inst.Voice.UseNoise
					case ipUseEnv:
						inst.Voice.UseEnv = !inst.Voice.UseEnv
					case ipNoiseFreq:
						i := inst.Voice.NoisePeriod
						i += 1
						if i > 31 {
							i = 31
						}
						inst.Voice.NoisePeriod = i
					case ipAmplitude:
						i := inst.Voice.Amplitude
						i += 1
						if i > 15 {
							i = 15
						}
						inst.Voice.Amplitude = i
					case ipEnvCoarse:
						i := inst.Voice.EnvCoarse
						i += 1
						if i > 0xff {
							i = 0x00
						}
						inst.Voice.EnvCoarse = i
					case ipEnvFine:
						i := inst.Voice.EnvFine
						i += 1
						if i > 0xff {
							i = 0x00
						}
						inst.Voice.EnvFine = i
					case ipEnvShape:
						i := inst.Voice.EnvShape
						i++
						if i > 15 {
							i = 0
						}
						inst.Voice.EnvShape = i
					}
					m.Song.SendNote(
						m.cTrack,
						m.Song.Instruments[int(m.Inst)],
						"C",
						4,
						byte(m.Song.Instruments[m.Inst].Voice.Amplitude),
					)
				case ch == vduconst.CSR_LEFT:
					inst := m.Song.Instruments[int(m.Inst)]
					switch m.InstParam {
					case ipUseTone:
						inst.Voice.UseTone = !inst.Voice.UseTone
					case ipUseNoise:
						inst.Voice.UseNoise = !inst.Voice.UseNoise
					case ipUseEnv:
						inst.Voice.UseEnv = !inst.Voice.UseEnv
					case ipNoiseFreq:
						i := inst.Voice.NoisePeriod
						i -= 1
						if i < 0 {
							i = 0
						}
						inst.Voice.NoisePeriod = i
					case ipAmplitude:
						i := inst.Voice.Amplitude
						i -= 1
						if i < 0 {
							i = 0
						}
						inst.Voice.Amplitude = i
					case ipEnvCoarse:
						i := inst.Voice.EnvCoarse
						i -= 1
						if i < 0 {
							i = 0xff
						}
						inst.Voice.EnvCoarse = i
					case ipEnvFine:
						i := inst.Voice.EnvFine
						i -= 1
						if i < 0 {
							i = 0xff
						}
						inst.Voice.EnvFine = i
					case ipEnvShape:
						i := inst.Voice.EnvShape
						i--
						if i < 0 {
							i = 0
						}
						inst.Voice.EnvShape = i
					}
					m.Song.SendNote(
						m.cTrack,
						m.Song.Instruments[int(m.Inst)],
						"C",
						4,
						byte(m.Song.Instruments[m.Inst].Voice.Amplitude),
					)
				case ch == 13:
					i := m.Song.Instruments[m.Inst]
					/*if m.InstParam == ipHiPass {
						def := fmt.Sprintf("%.0f", i.OSC[0].HiPass)
						val := InputPopup(ent, "Enter hi-pass freq", "\r\n\r\nFrequency(Hz)", def)
						if f, err := strconv.ParseFloat(val, 64); err == nil && f >= 0 && f < 20000 {
							i.OSC[0].HiPass = f
							ok = true
						}
					} else if m.InstParam == ipLoPass {
						def := fmt.Sprintf("%.0f", i.OSC[0].LoPass)
						val := InputPopup(ent, "Enter lo-pass freq", "\r\n\r\nFrequency(Hz)", def)
						if f, err := strconv.ParseFloat(val, 64); err == nil && f >= 0 && f < 20000 {
							i.OSC[0].LoPass = f
							ok = true
						}
					} else */
					if m.InstParam == ipName {
						def := i.Name
						val := InputPopup(ent, "Enter Instrument name", "\r\n\r\nName", def)
						if len(val) > 20 {
							val = val[:20]
						}
						if val != "" {
							i.Name = val
						}
					} else {
						m.Song.SendNote(
							m.cTrack,
							m.Song.Instruments[int(m.Inst)],
							"C",
							4,
							1.0,
						)
					}
				}
			} else {
				time.Sleep(50 * time.Millisecond)
			}
		case ctxSongEdit:
			m.DrawTrackStatusDisplay(ent)
			m.DrawPatternDisplay(ent)
			m.DrawInstrumentDisplay(ent)
			m.DrawSongEditDisplay(ent)
			m.refresh = false
			if ch := ent.GetMemoryMap().KeyBufferGet(ent.GetMemIndex()); ch != 0 {
				// do something here
			gotkeysong:
				m.refresh = true

				if ch >= vduconst.CTRL1 && ch <= vduconst.CTRL6 {
					i := int(ch - vduconst.CTRL1)
					m.Song.TrackDisabled[i] = !m.Song.TrackDisabled[i]
				}
				if m.SongParam == spName {

					switch {
					case ch == vduconst.PAGE_UP:
						m.SongParam--
						if m.SongParam < spName {
							m.SongParam = spList
						}
					case ch == vduconst.PAGE_DOWN:
						m.SongParam++
						if m.SongParam > spList {
							m.SongParam = spName
						}
					case ch == 127 && len(m.Song.Name) > 0:
						m.Song.Name = m.Song.Name[:len(m.Song.Name)-1]
					case ch >= 32 && ch <= 126 && len(m.Song.Name) < 40:
						m.Song.Name += string(ch)
					case ch == 13:
						m.SongParam = spList
					case ch == 9:
						m.SongParam = spList
						m.ctx = ctxInstEdit
					}

				} else {

					switch {
					case ch == vduconst.SHIFT_CTRL_N:
						m.NewSong()
					case ch == vduconst.SHIFT_CTRL_H:
						m.ShowHelp()
					case ch == vduconst.PAGE_UP:
						m.SongParam--
						if m.SongParam < spName {
							m.SongParam = spList
						}
					case ch == vduconst.PAGE_DOWN:
						m.SongParam++
						if m.SongParam > spList {
							m.SongParam = spName
						}
					case ch == ' ':
						m.PlaySong()
					case ch == vduconst.CTRL_SPACE:
						m.PlayPattern()
					case ch == vduconst.SHIFT_SPACE:
						m.PlaySongFromSOP()
					case ch == vduconst.SHIFT_CTRL_E:
						m.ExportSongCombinedWLA()
					case ch == vduconst.SHIFT_CTRL_F:
						m.ExportSongWLA()
					case ch == vduconst.SHIFT_CTRL_A:
						m.ExportSongASM()
					case ch == vduconst.SHIFT_CTRL_C:
						m.copyPattern()
					case ch == vduconst.SHIFT_CTRL_V:
						m.pastePattern()
					case ch == vduconst.SHIFT_CTRL_I:
						m.insertBlankPattern()
					case ch == vduconst.SHIFT_CTRL_P:
						m.PlaySong()
					case ch == vduconst.SHIFT_CTRL_S:
						name := InputPopup(ent, "Save song", "\r\nEnter filename", m.CurrentFile)
						fmt.Printf("Save: %v\n", m.Song.Save(name))
						if name != "" {
							if !strings.HasPrefix(name, "/") {
								name = "/" + strings.Trim(ent.GetWorkDir(), "/") + "/" + name
							}
							if m.Song.Save(name) == nil {
								m.CurrentFile = name
							}
						}
					case ch == vduconst.SHIFT_CTRL_L:
						name := InputPopup(ent, "Load song", "\r\nEnter filename", m.CurrentFile)
						if name != "" {
							if !strings.HasPrefix(name, "/") {
								name = "/" + strings.Trim(ent.GetWorkDir(), "/") + "/" + name
							}
							if m.Song.Load(name) == nil {
								m.CurrentFile = name
								m.Song.PatternPos = 0
								m.Song.PatternListIndex = 0
								m.refresh = true
								m.DrawSongEditDisplay(m.Int)
							}
						}
					case ch == 27:
						m.running = false
					case ch == vduconst.CSR_UP:
						m.incPattern()
					case ch == vduconst.CSR_DOWN:
						m.decPattern()
					case ch == 13:
						m.showPattern()
					case ch == 127:
						m.deletePattern()
					case ch == 9:
						m.ctx = ctxInstEdit
					case ch == vduconst.CSR_LEFT:
						m.viewPLSubIndex--
						if m.viewPLSubIndex < 0 {
							if m.viewPLIndex > 0 {
								m.viewPLIndex--
								m.viewPLSubIndex = 1
							} else {
								m.viewPLSubIndex = 0
							}
						}
						if m.Song.PlayMode != tracker.PMLoopSong {
							m.Song.PatternListIndex = m.viewPLIndex
						}
					case ch == vduconst.CSR_RIGHT:
						m.viewPLSubIndex++
						if m.viewPLSubIndex > 1 {
							if m.viewPLIndex < 0xff {
								m.viewPLIndex++
								m.viewPLSubIndex = 0
							} else {
								m.viewPLSubIndex = 1
							}
						}
						if m.Song.PlayMode != tracker.PMLoopSong {
							m.Song.PatternListIndex = m.viewPLIndex
						}
					case strings.Contains(hexKeys, string(rune(ch))):
						// entry key
						ival := m.Song.PatternList[m.viewPLIndex]
						if ival == -1 {
							ival = 0
						}
						switch m.viewPLSubIndex {
						case 0:
							istr := fmt.Sprintf("%.2x", ival)
							istr = string(rune(ch)) + istr[1:]
							if ival, err := strconv.ParseUint(istr, 16, 8); err == nil {
								m.Song.PatternList[m.viewPLIndex] = int(ival)
							}
							ch = vduconst.CSR_RIGHT
							goto gotkeysong
						case 1:
							istr := fmt.Sprintf("%.2x", ival)
							istr = istr[:1] + string(rune(ch))
							if ival, err := strconv.ParseUint(istr, 16, 8); err == nil {
								m.Song.PatternList[m.viewPLIndex] = int(ival)
							}
							ch = vduconst.CSR_RIGHT
							goto gotkeysong
						}
					}
				}
			} else {
				time.Sleep(50 * time.Millisecond)
			}
		case ctxTrackEdit:
			m.DrawTrackStatusDisplay(ent)
			m.DrawSongEditDisplay(ent)
			m.DrawInstrumentDisplay(ent)
			m.DrawPatternDisplay(ent)
			m.refresh = false
			if ch := ent.GetMemoryMap().KeyBufferGet(ent.GetMemIndex()); ch != 0 {
				m.refresh = true
				// do something here
				if ch >= vduconst.CTRL1 && ch <= vduconst.CTRL6 {
					i := int(ch - vduconst.CTRL1)
					m.Song.TrackDisabled[i] = !m.Song.TrackDisabled[i]
					m.Song.Squelch(i)
				}

				ro := m.Song.TrackDisabled[m.cTrack] || m.ReadOnlyTrack

				switch ch {
				case '`':
					p, idx := m.Song.CurrentPatternPos()
					if p != nil && !ro {
						var str = "X"
						var oct = byte(1)
						nn := p.Tracks[m.cTrack].Notes[idx]
						if nn == nil {
							nn = &tracker.TNote{}
							p.Tracks[m.cTrack].Notes[idx] = nn
						}
						nn.Note = &str
						nn.Instrument = nil
						nn.Volume = nil
						nn.Octave = &oct
						nn.Command = nil
						nn.CommandValue = nil
						m.Song.Squelch(m.cTrack)
						if m.Song.PlayMode == tracker.PMBoundPattern && !ro {
							m.Song.PatternAdvance() // advance
						}
					}
				case vduconst.SHIFT_CTRL_P:
					name := InputPopup(ent, "Enter notes", "\r\nEnter notes", "")
					if name != "" {
						m.Song.ResetEntry(m.cTrack)
						m.Song.EnterNotes(name)
					}
				case vduconst.SHIFT_CTRL_N:
					m.NewSong()
				case vduconst.SHIFT_CTRL_E:
					m.ExportSongCombinedWLA()
				case vduconst.SHIFT_CTRL_F:
					m.ExportSongWLA()
				case vduconst.SHIFT_CTRL_A:
					m.ExportSongASM()
				case vduconst.SHIFT_CTRL_H:
					m.ShowHelp()
				case vduconst.CTRL_C:
					m.copyTrack()
				case vduconst.CTRL_V:
					m.Song.RememberPattern()
					m.pasteTrack()
				case vduconst.SHIFT_CTRL_C:
					m.copyPattern()
				case vduconst.SHIFT_CTRL_V:
					m.Song.RememberPattern()
					m.pastePattern()
				case ' ':
					m.PlaySong()
				case vduconst.CTRL_SPACE:
					m.PlayPattern()
				case vduconst.SHIFT_SPACE:
					m.PlaySongFromSOP()
				case vduconst.SHIFT_CTRL_D:
					m.Song.RememberPattern()
					m.clearPattern()
				case vduconst.CTRL_D:
					m.Song.RememberPattern()
					m.clearTrack()
				case vduconst.CTRL_X:
					m.Song.RememberPattern()
					m.copyTrack()
					m.clearTrack()
				case vduconst.SHIFT_CTRL_X:
					m.Song.RememberPattern()
					m.copyPattern()
					m.clearPattern()
				case vduconst.SHIFT_CTRL_Z:
					m.Song.UndoPattern()
				case 9:
					m.ctx = ctxSongEdit
				case vduconst.SHIFT_CTRL_S:
					name := InputPopup(ent, "Save song", "\r\nEnter filename", m.CurrentFile)
					if name != "" {
						fmt.Printf("Save: %v\n", m.Song.Save(name))
						if !strings.HasPrefix(name, "/") {
							name = "/" + strings.Trim(ent.GetWorkDir(), "/") + "/" + name
						}
						if m.Song.Save(name) == nil {
							m.CurrentFile = name
						}
					}
				case vduconst.SHIFT_CTRL_L:
					name := InputPopup(ent, "Load song", "\r\nEnter filename", m.CurrentFile)
					if name != "" {
						if !strings.HasPrefix(name, "/") {
							name = "/" + strings.Trim(ent.GetWorkDir(), "/") + "/" + name
						}
						if m.Song.Load(name) == nil {
							m.CurrentFile = name
							m.Song.PatternListIndex = 0
							m.Song.PatternPos = 0
							m.refresh = true
							m.DrawSongEditDisplay(m.Int)
						}
					}
				case 27:
					resp := InputPopup(ent, "Quit microTracker?", "\r\nQuit microTracker (y/n)", "n")
					if resp != "" {
						if strings.ToLower(resp) == "y" {
							m.running = false
						} else {
							m.refresh = true
						}
					}
				case vduconst.PAGE_UP:
					m.Octave++
					if m.Octave > 6 {
						m.Octave = 6
					}
				case vduconst.PAGE_DOWN:
					m.Octave--
					if m.Octave < 0 {
						m.Octave = 0
					}
				case vduconst.SHIFT_CSR_LEFT:
					m.Inst--
					for m.Song.Instruments[int(m.Inst)] == nil && m.Inst > 0 {
						m.Inst--
					}
				case vduconst.SHIFT_CSR_RIGHT:
					m.Inst++
					for m.Song.Instruments[int(m.Inst)] == nil {
						m.Inst++
					}
				case '\\':
					m.ReadOnlyTrack = !m.ReadOnlyTrack
				// case vduconst.PAGE_UP:
				// 	m.Song.Tempo++
				// case vduconst.PAGE_DOWN:
				// 	m.Song.Tempo--
				case vduconst.CSR_DOWN:
					m.Song.TrackStep(1)
				case vduconst.CSR_UP:
					m.Song.TrackStep(-1)
				case vduconst.CSR_LEFT:
					m.cTrackSub--
					if m.cTrackSub < 0 {
						m.cTrack--
						if m.cTrack < 0 {
							m.cTrack = tracker.MaxPatternTracks - 1
						}
						m.cTrackSub = 9
					}
				case vduconst.CSR_RIGHT:
					m.cTrackSub++
					if m.cTrackSub > 9 {
						m.cTrack++
						if m.cTrack >= tracker.MaxPatternTracks {
							m.cTrack = 0
						}
						m.cTrackSub = 0
					}
				case 13:
					p, _ := m.Song.CurrentPatternPos()
					if p != nil {
						m.Song.PlayLine(-1)
						m.Song.TrackStep(1)
					}
				case 127:
					if !ro {
						p, idx := m.Song.CurrentPatternPos()
						if p != nil {
							nn := p.Tracks[m.cTrack].Notes[idx]
							if nn != nil {
								switch {
								case m.cTrackSub < 3:
									//if nn.Note != nil {
									nn.Note = nil
									nn.Octave = nil
									nn.Instrument = nil
									nn.Volume = nil
									//} else {
									m.Song.TrackStep(1)
									//}
								case m.cTrackSub < 5:
									nn.Instrument = nil
								case m.cTrackSub < 7:
									nn.Volume = nil
								default:
									nn.Command = nil
									nn.CommandValue = nil
								}
								p.Tracks[m.cTrack].Notes[idx] = nn
							}
						}
					}
				default:
					pattern := validKeys[m.cTrackSub]
					tmp := ch
					if tmp >= 'A' && tmp <= 'Z' {
						tmp += 32
					}
					if strings.Contains(pattern, string(rune(tmp))) {
						// valid key!
						switch m.cTrackSub {
						case 0, 1:
							// Note key
							n, ok := noteKeys[rune(tmp)]
							if ok {
								//
								fmt.Printf("Got note %s%d\n", n.Note, n.OctaveMod)
								p, idx := m.Song.CurrentPatternPos()
								if p != nil {
									if !ro {
										nn := p.Tracks[m.cTrack].Notes[idx]
										p.Tracks[m.cTrack].Notes[idx] =
											nn.SetNote(n.Note).
												SetOctave(byte(m.Octave + n.OctaveMod)).
												SetInstrument(m.Inst).
												SetVolume(byte(m.Song.Instruments[m.Inst].Voice.Amplitude))

										if *p.Tracks[m.cTrack].Notes[idx].Note == "X" {
											p.Tracks[m.cTrack].Notes[idx].Volume = nil
										}
									}
									m.Song.SendNote(m.cTrack, m.Song.Instruments[m.Inst], n.Note, m.Octave+n.OctaveMod, byte(m.Song.Instruments[m.Inst].Voice.Amplitude))
									if m.Song.PlayMode == tracker.PMBoundPattern && !ro {
										m.Song.PatternAdvance() // advance
									}
								}
							}
						case 2:
							// octave
							if !ro {
								p, idx := m.Song.CurrentPatternPos()
								if p != nil {
									nn := p.Tracks[m.cTrack].Notes[idx]
									if nn != nil && nn.Note != nil {
										newOctave, err := strconv.ParseUint(string(rune(ch)), 10, 32)
										if err == nil && newOctave >= 0 && newOctave <= 6 {
											nn.SetOctave(byte(newOctave))
											m.Song.PlayLine(m.cTrack)
										}
									}
								}
							}
						case 3:
							if !ro {
								// inst hi
								p, idx := m.Song.CurrentPatternPos()
								if p != nil {
									nn := p.Tracks[m.cTrack].Notes[idx]
									if nn == nil {
										nn = &tracker.TNote{}
										p.Tracks[m.cTrack].Notes[idx] = nn
									}
									istr := fmt.Sprintf("%.2x", nn.GetInstrument())
									istr = string(rune(ch)) + istr[1:]
									if ival, err := strconv.ParseUint(istr, 16, 8); err == nil {
										nn.SetInstrument(byte(ival))
										m.Inst = byte(ival)
										m.Song.PlayLine(m.cTrack)
									}
								}
							}
						case 4:
							// inst lo
							if !ro {
								p, idx := m.Song.CurrentPatternPos()
								if p != nil {
									nn := p.Tracks[m.cTrack].Notes[idx]
									if nn == nil {
										nn = &tracker.TNote{}
										p.Tracks[m.cTrack].Notes[idx] = nn
									}
									istr := fmt.Sprintf("%.2x", nn.GetInstrument())
									istr = istr[:1] + string(rune(ch))
									if ival, err := strconv.ParseUint(istr, 16, 8); err == nil {
										nn.SetInstrument(byte(ival))
										m.Inst = byte(ival)
										m.Song.PlayLine(m.cTrack)
									}
								}
							}
						case 5:
							// inst hi
							if !ro {
								p, idx := m.Song.CurrentPatternPos()
								if p != nil {
									nn := p.Tracks[m.cTrack].Notes[idx]
									if nn == nil {
										nn = &tracker.TNote{}
										p.Tracks[m.cTrack].Notes[idx] = nn
									}
									istr := fmt.Sprintf("%.2x", nn.GetVolume())
									istr = string(rune(ch)) + istr[1:]
									if ival, err := strconv.ParseUint(istr, 16, 8); err == nil {
										nn.SetVolume(byte(ival))
										m.Volume = byte(ival)
										m.Song.PlayLine(m.cTrack)
									}
								}
							}
						case 6:
							if !ro {
								// inst lo
								p, idx := m.Song.CurrentPatternPos()
								if p != nil {
									nn := p.Tracks[m.cTrack].Notes[idx]
									if nn == nil {
										nn = &tracker.TNote{}
										p.Tracks[m.cTrack].Notes[idx] = nn
									}
									istr := fmt.Sprintf("%.2x", nn.GetVolume())
									istr = istr[:1] + string(rune(ch))
									if ival, err := strconv.ParseUint(istr, 16, 8); err == nil {
										nn.SetVolume(byte(ival))
										m.Volume = byte(ival)
										m.Song.PlayLine(m.cTrack)
									}
								}
							}
						case 7:
							if !ro {
								p, idx := m.Song.CurrentPatternPos()
								if p != nil {
									nn := p.Tracks[m.cTrack].Notes[idx]
									if nn == nil {
										nn = &tracker.TNote{}
										p.Tracks[m.cTrack].Notes[idx] = nn
									}
									if ch >= 'A' && ch <= 'Z' {
										ch += 32
									}
									nn.SetCommand(byte(ch))
									if nn.CommandValue == nil {
										nn.SetCommandValue(0)
									}
								}
							}
						case 8:
							// inst hi
							if !ro {
								p, idx := m.Song.CurrentPatternPos()
								if p != nil {
									nn := p.Tracks[m.cTrack].Notes[idx]
									if nn == nil {
										nn = &tracker.TNote{}
										p.Tracks[m.cTrack].Notes[idx] = nn
									}
									istr := fmt.Sprintf("%.2x", nn.GetCommandValue())
									istr = string(rune(ch)) + istr[1:]
									if ival, err := strconv.ParseUint(istr, 16, 8); err == nil {
										nn.SetCommandValue(byte(ival))
									}
								}
							}
						case 9:
							if !ro {
								// inst lo
								p, idx := m.Song.CurrentPatternPos()
								if p != nil {
									nn := p.Tracks[m.cTrack].Notes[idx]
									if nn == nil {
										nn = &tracker.TNote{}
										p.Tracks[m.cTrack].Notes[idx] = nn
									}
									istr := fmt.Sprintf("%.2x", nn.GetCommandValue())
									istr = istr[:1] + string(rune(ch))
									if ival, err := strconv.ParseUint(istr, 16, 8); err == nil {
										nn.SetCommandValue(byte(ival))
									}
								}
							}
						}
						if m.cTrackSub != 0 && !ro {
							m.cTrackSub++
							if m.cTrackSub > 9 {
								m.cTrack++
								if m.cTrack >= tracker.MaxPatternTracks {
									m.cTrack = 0
								}
								m.cTrackSub = 0
							}
						}
					}
				}
			} else {
				//tempogap := time.Minute / time.Duration(m.Song.Tempo)
				time.Sleep(50 * time.Millisecond)
				m.CheckPalette()
			}
		default:
			// stuff
		}

	}

}

func (m *MicroTracker) SendRest(s string) {
	//fmt.Printf("Sending:\n%s\n", s)
	// dummy use tone channel
	//clientperipherals.SPEAKER.SendCommands(
	//	s,
	//)
	m.refresh = true
}

func (m *MicroTracker) ExportSongCombinedWLA() {

	def := "4000"

addrBad:

	addrStr := InputPopup(
		m.Int,
		"Load Address",
		"Enter address for song+player code: $",
		def,
	)

	if addrStr == "" {
		return
	}

	addrBin, err := strconv.ParseUint(addrStr, 16, 32)
	if err != nil {
		goto addrBad
	}

	a := tracker.NewASMEncoder()
	path, err := a.CompileSong(m.Song, tracker.CmBINCombined, int(addrBin))
	if err != nil {
		// handle err
		log.Printf("Failed to compile song: %v", err)
		InfoPopup(m.Int, "Compile song", "Export failed.", 5*time.Second)
		return
	}
	// handle success
	log.Printf("Created file %s", path)
	InfoPopup(m.Int, "Compile Song", "Export succeeded to: "+path, 5*time.Second)
}

func (m *MicroTracker) ExportSongASM() {
	a := tracker.NewASMEncoder()
	path, err := a.CompileSong(m.Song, tracker.CmASMSongOnly, 0x4000)
	if err != nil {
		// handle err
		log.Printf("Failed to Generate song: %v", err)
		InfoPopup(m.Int, "Generate song", "Export failed.", 5*time.Second)
		return
	}
	// handle success
	log.Printf("Created file %s", path)
	InfoPopup(m.Int, "Generate Song", "Export succeeded to: "+path, 5*time.Second)
}

func (m *MicroTracker) ExportSongWLA() {

	def := "4000"

addrBad:

	addrStr := InputPopup(
		m.Int,
		"Load Address",
		"Enter address for song data: $",
		def,
	)

	if addrStr == "" {
		return
	}

	addrBin, err := strconv.ParseUint(addrStr, 16, 32)
	if err != nil {
		goto addrBad
	}

	a := tracker.NewASMEncoder()
	path, err := a.CompileSong(m.Song, tracker.CmBINSongOnly, int(addrBin))
	if err != nil {
		// handle err
		log.Printf("Failed to compile song: %v", err)
		InfoPopup(m.Int, "Compile song", "Export failed.", 5*time.Second)
		return
	}
	// handle success
	log.Printf("Created file %s", path)
	InfoPopup(m.Int, "Compile Song", "Export succeeded to: "+path, 5*time.Second)
}
