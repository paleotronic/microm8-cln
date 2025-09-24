package restalgia

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	//	"paleotronic.com/fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"paleotronic.com/utils"
)

type AudioDevice interface {
	NoteChange(it *InstrumentPack)
	Process()
	GetAmplitude() float32
	GetPannedAmplitude() []float32
}

type Song struct {
	Speed              int64
	NextNote           int64
	WaveVoice          []*Voice
	Band               map[string]*Instrument
	Alias              map[string]string
	Patterns           map[string]*Pattern
	NoteIndex          int
	CurrentSongPattern *Pattern
	Tracks             map[string]*Track
	Order              []string
	Playing            bool
	PatternIndex       int
	SamplesForNote     int
	Looping            bool
	PatternLength      int
}

func (this *Song) SetCurrentSongPattern(currentSongPattern *Pattern) {
	this.CurrentSongPattern = currentSongPattern
}

// Use reader here because we can feed bytes or files to it :)
func (this *Song) LoadReader(reader io.Reader) error {

	r := bufio.NewReader(reader)

	text, err := r.ReadString(10)
	////fmt.Printntln(err, text)
	s := ""

	// repeat until all lines is read
	for err == nil {
		s = s + "\n" + text
		text, err = r.ReadString(10)
	}
	if text != "" {
		s = s + "\n" + text
	}

	return this.LoadString(s)

}

func (this *Song) GetTracks() map[string]*Track {
	return this.Tracks
}

func (this *Song) IsPlaying() bool {
	return this.Playing
}

func (this *Song) GetPatterns() map[string]*Pattern {
	return this.Patterns
}

func (this *Song) GetSpeed() int64 {
	return this.Speed
}

func (this *Song) Start() {
	this.PatternIndex = 0
	this.NoteIndex = 0
	pattname := this.Order[this.PatternIndex]
	this.CurrentSongPattern = this.Patterns[pattname]
	this.Playing = true
}

func (this *Song) Play() {
	this.Start()
}

func (this *Song) Pause() {
	this.Playing = false
}

func (this *Song) Resume() {
	this.Playing = true
}

func (this *Song) Stop() {
	this.PatternIndex = 0
	this.NoteIndex = 0
	pattname := this.Order[this.PatternIndex]
	this.CurrentSongPattern = this.Patterns[pattname]
	this.Playing = false
}

func (this *Song) SetOrder(order []string) {
	this.Order = order
}

func (this *Song) SetPatternIndex(patternIndex int) {
	this.PatternIndex = patternIndex
}

func (this *Song) SetNoteIndex(noteIndex int) {
	this.NoteIndex = noteIndex
}

func (this *Song) LoadString(s string) error {

	var currentTrack *Track = nil
	var trackLength int = 0
	var trackName string = ""

	sr := regexp.MustCompile("[ \t]+")

	//////fmt.Println("FILE =", s)

	for _, line := range strings.Split(s, "\n") {
		line = strings.Trim(line, " \t\r")
		////fmt.Printntln("LINE =", line)

		if line == "" {
			continue
		}

		// get rid of leading and trailing spaces

		// instrument def
		if strings.Index(line, "INSTRUMENT") == 0 {

			line = strings.Trim(strings.Replace(line, "INSTRUMENT", "", -1), " \t")

			i := strings.Index(line, " ")

			name := line[0:i]
			line = strings.Trim(strings.Replace(line, name, "", -1), " \t")
			params := line

			n := byte(len(this.Band) + 1)
			alias := strings.ToLower(hex.EncodeToString([]byte{n}))
			if len(alias) < 2 {
				alias = "0" + alias
			}

			if strings.IndexRune(name, ':') > 0 {
				p := strings.SplitN(name, ":", 2)
				name = p[1]
				alias = p[0]
			}

			//System.Out.Println("Got instrument name = "+name+", params = "+params);
			////fmt.Printntf("Instrument %s with alias %s\n", name, alias)

			this.Band[name] = NewInstrument(params)
			this.Alias[alias] = name
			continue
		}

		// track def
		if strings.Index(line, "TRACK") == 0 {

			// handle event where we were in a previous track
			if currentTrack != nil {
				if currentTrack.GetNoteCount() != trackLength {
					return errors.New("Invalid tracklength for " + trackName + " expected " + utils.IntToStr(trackLength) + " got " + utils.IntToStr(currentTrack.GetNoteCount()))
				}
				currentTrack = nil
			}

			line = strings.Trim(strings.Replace(line, "TRACK", "", -1), " \t")
			i := strings.Index(line, " ")
			name := line[0:i]
			line = strings.Trim(strings.Replace(line, name, "", -1), " \t")

			parts := sr.Split(line, -1)

			instname := ""

			loop := false

			for _, p := range parts {
				//////fmt.Println(p)
				nv := strings.Split(p, "=")
				//////fmt.Println(nv[0] + " == " + nv[1])
				if nv[0] == "INSTRUMENT" {
					instname = nv[1]

					mapped, ok := this.Alias[instname]
					if ok {
						instname = mapped
					}

					continue
				}
				if nv[0] == "LENGTH" {
					trackLength = utils.StrToInt(nv[1])
					continue
				}
				if nv[0] == "LOOP" {
					n := utils.StrToInt(nv[1])

					loop = (n != 0)

					continue
				}
			}

			//System.Out.Println("Got track = "+name+" ("+instname+")");
			t := NewTrack(instname, this.Band, this.Alias)
			t.Looping = loop
			this.Tracks[name] = t
			currentTrack = t
			trackName = name

			//////fmt.Println("TRACK LEN =", trackLength)

			continue
		}

		if strings.Index(line, "PATTERNLENGTH") == 0 {

			// handle event where we were in a previous track
			if currentTrack != nil {
				if currentTrack.GetNoteCount() != trackLength {
					return errors.New("Invalid tracklength for " + trackName + " expected " + utils.IntToStr(trackLength) + " got " + utils.IntToStr(currentTrack.GetNoteCount()))
				}
				currentTrack = nil
			}

			line = strings.Trim(strings.Replace(line, "PATTERNLENGTH", "", -1), " \t")
			//////fmt.Println("ZZZZZZZZZZZZZZZZZZZZZ " + line)
			this.PatternLength = utils.StrToInt(line)

			continue

		}

		if strings.Index(line, "PATTERN") == 0 {

			// handle event where we were in a previous track
			if currentTrack != nil {
				if currentTrack.GetNoteCount() != trackLength {
					return errors.New("Invalid tracklength for " + trackName + " expected " + utils.IntToStr(trackLength) + " got " + utils.IntToStr(currentTrack.GetNoteCount()))
				}
				currentTrack = nil
			}

			line = strings.Trim(strings.Replace(line, "PATTERN", "", -1), " \t")
			i := strings.Index(line, " ")
			name := line[0:i]
			line = strings.Trim(strings.Replace(line, name, "", -1), " \t")

			parts := sr.Split(line, -1)

			pattern := NewPattern(name)
			this.Patterns[name] = pattern

			for _, trackname := range parts {
				// does it exist
				t, ok := this.Tracks[trackname]

				if !ok {
					return errors.New("Trackname " + name + " does not exist...")
				}

				// okay all good lets bake this shit
				pattern.Add(t)
			}

			continue

		}

		if strings.Index(line, "SPEED") == 0 {

			// handle event where we were in a previous track
			if currentTrack != nil {
				if currentTrack.GetNoteCount() != trackLength {
					return errors.New("Invalid tracklength for " + trackName + " expected " + utils.IntToStr(trackLength) + " got " + utils.IntToStr(currentTrack.GetNoteCount()))
				}
				currentTrack = nil
			}

			line = strings.Trim(strings.Replace(line, "SPEED", "", -1), " \t")
			//////fmt.Println("ZZZZZZZZZZZZZZZZZZZZZ " + line)
			this.Speed = int64(utils.StrToInt(line))

			continue

		}

		if strings.Index(line, "SONG") == 0 {

			// handle event where we were in a previous track
			if currentTrack != nil {
				if currentTrack.GetNoteCount() != trackLength {
					return errors.New("Invalid tracklength for " + trackName + " expected " + utils.IntToStr(trackLength) + " got " + utils.IntToStr(currentTrack.GetNoteCount()))
				}
				currentTrack = nil
			}

			line = strings.Trim(strings.Replace(line, "SONG", "", -1), " \t")

			parts := sr.Split(line, -1)

			this.Order = parts

			for _, pattname := range parts {
				// does it exist
				_, ok := this.Patterns[pattname]

				if !ok {
					return errors.New("Error: Song specifies a pattern " + pattname + " that does not exist!")
				}
			}

			continue

		}

		// fallback is note data
		if line != "" && currentTrack != nil {
			//////fmt.Println("---> NOTE = " + line)
			currentTrack.Add(line)
			continue
		}

		// error
		if line == "" {
			if currentTrack != nil {
				if currentTrack.GetNoteCount() != trackLength {
					return errors.New("Invalid tracklength for " + trackName + " expected " + utils.IntToStr(trackLength) + " got " + utils.IntToStr(currentTrack.GetNoteCount()))
				}
				currentTrack = nil
			}
		}
	}

	if len(this.Order) == 0 {
		return errors.New("Missing SONG directive?")
	}

	////fmt.Printntln("Loaded song: OK")

	this.Start()

	return nil

}

func (this *Song) SetBand(band map[string]*Instrument) {
	this.Band = band
}

func (this *Song) GetPannedAmplitude() []float32 {

	var amp []float32 = make([]float32, 2)
	amp[0] = 0.0
	amp[1] = 0.0
	var count int = 0
	var left = make([]float32, numVOICES)
	var right = make([]float32, numVOICES)

	for _, v := range this.WaveVoice {
		vamp := v.GetPannedAmplitudes()

		left[count] = vamp[0]
		right[count] = vamp[1]

		count++
	}

	amp = mixChannelsStereo(left, right)

	return amp
}

func (this *Song) GetMonoAmplitude() float32 {
	var amp float32
	amp = 0.0
	var count int = 0
	var left = make([]float32, numVOICES)

	for _, v := range this.WaveVoice {
		left[count] = v.GetAmplitude()
		count++
	}

	amp = mixChannelsMono(left)

	return amp
}

func (this *Song) PlayLine() {

	if !this.Playing {
		return
	}

	// get current pattern
	if this.NoteIndex >= this.PatternLength {
		// get pattern
		this.PatternIndex++

		if this.PatternIndex >= len(this.Order) {
			this.PatternIndex = 0
			this.Playing = this.Looping

			if !this.Playing {
				return
			}
		}

		pattname := this.Order[this.PatternIndex]
		this.CurrentSongPattern = this.Patterns[pattname]

		x := 0
		for _, t := range this.CurrentSongPattern.Tracks {
			this.WaveVoice[x].Pan = 0
			t.NoteIndex = 0
			///audio.Process()
			x++
		}

		this.NoteIndex = 0
	}

	////fmt.Printntf("patternIndex = %d, noteIndex = %d\n", this.PatternIndex, this.NoteIndex)
	// okay good now let's play it
	it := NewInstrumentPack()
	for v := 0; v < this.CurrentSongPattern.Size(); v++ {
		track := this.CurrentSongPattern.Get(v)
		track.NoteIndex = this.NoteIndex
		voice := this.WaveVoice[v]

		if track.HasNotes() {
			//track.PlayNote(voice);
			////fmt.Printntln("Paying note on track " + utils.IntToStr(v))
			track.PlayNote(it, voice)
		}

		//		audio.Process()
	}

	// attach the data to the audio device
	it.Apply()

	// set countdown in samples for next note
	////fmt.Printntf("Add %d to this.NextNote\n", this.Speed)
	this.NextNote = (time.Now().UnixNano() / 1000000) + this.Speed

	// advance the play head
	this.NoteIndex++

}

func (this *Song) SetTracks(tracks map[string]*Track) {
	this.Tracks = tracks
}

func (this *Song) SetPlaying(playing bool) {
	this.Playing = playing
}

func (this *Song) GetCurrentSongPattern() *Pattern {
	return this.CurrentSongPattern
}

func (this *Song) oldPause(audio AudioDevice) {

	now := time.Now().UnixNano()
	when := now + int64(this.Speed*1000000)

	inst := time.Now().UnixNano()
	for inst < when {
		inst = time.Now().UnixNano()
		audio.Process()
	}

	////fmt.Printntf("Song.Pause() waited for %d ns\n", (inst - now))
}

func (this *Song) Load(filename string) error {

	b, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	buffer := bytes.NewBuffer(b)
	return this.LoadReader(buffer)

}

func NewSong(filename string) (*Song, error) {
	this := &Song{}
	this.WaveVoice = make([]*Voice, numVOICES)
	this.Speed = 250
	this.PatternIndex = -1
	this.NoteIndex = 0
	this.NextNote = time.Now().UnixNano() / 1000000 // Ms
	this.Playing = true
	this.Looping = false
	this.PatternLength = 64 // default pattern length

	this.Band = make(map[string]*Instrument)
	this.Alias = make(map[string]string)
	this.Patterns = make(map[string]*Pattern)
	this.Tracks = make(map[string]*Track)
	this.Order = make([]string, 0)

	for x := 0; x < numVOICES; x++ {
		this.WaveVoice[x] = NewVoice(fmt.Sprintf("voice%d", x), 44100, SINE, 1.0)
	}

	// load a song from a file
	var err error
	if filename != "" {
		err = this.Load(filename)
	}
	return this, err
}

func (this *Song) SetPatterns(patterns map[string]*Pattern) {
	this.Patterns = patterns
}

func (this *Song) SetSpeed(speed int64) {
	this.Speed = speed
}

func (this *Song) GetOrder() []string {
	return this.Order
}

func (this *Song) GetPatternIndex() int {
	return this.PatternIndex
}

func (this *Song) GetNoteIndex() int {
	return this.NoteIndex
}

func (this *Song) GetBand() map[string]*Instrument {
	return this.Band
}

func (this *Song) PullSampleMono() float32 {
	if !this.Playing {
		return 0
	}

	now := time.Now().UnixNano() / 1000000

	if now < this.NextNote {
		return this.GetMonoAmplitude()
	}

	// advance note
	this.PlayLine()

	return this.GetMonoAmplitude()
}

func (this *Song) PullSampleStereo() []float32 {
	if !this.Playing {
		return []float32{0, 0}
	}

	now := time.Now().UnixNano() / 1000000
	if now < this.NextNote {
		return this.GetPannedAmplitude()
	}

	// advance note
	this.PlayLine()

	return this.GetPannedAmplitude()
}
