package clientperipherals

/*
Simple PC Speaker style implementation
*/

import (
	"bytes"
	"math"
	"strings"
	"time"

	"github.com/mjibson/go-dsp/wav"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
	"paleotronic.com/log" //"runtime"
	"paleotronic.com/restalgia"
	"paleotronic.com/restalgia/control"
	"paleotronic.com/restalgia/driver"
	"paleotronic.com/utils"

	log2 "log"
)

var SPEAKER *SoundPod
var Context int

const TONE_VOLUME = 0.5
const WAVE_VOLUME = 0.5

type SoundPod struct {
	Output driver.Output
	Mixer  *restalgia.Mixer
	//buzzer, musicl, musicr, track [memory.OCTALYZER_NUM_INTERPRETERS]*restalgia.Voice
	//tone, beep                    *restalgia.Voice
	instTone, instBeep *restalgia.Instrument
	Song               *restalgia.Song
	sam                float32
	lastSoundTick      int64
	buffersize         int
	lastbuffer         int
	BuzzerChannel      *int
	//output             *os.File
}

func (this *SoundPod) StopAudio() {
	// if this.output != nil {
	// 	this.output.Close()
	// }
	this.Output.Stop()
}

func (this *SoundPod) RedefineInst(s string) {
	tone := this.Mixer.FindVoice(Context, "tone")
	tone.AddInstrumentChange(s)
}

func (this *SoundPod) RedefineInstImmediate(s string) {
	tone := this.Mixer.FindVoice(Context, "tone")
	this.instTone = restalgia.NewInstrument(s)
	this.instTone.Apply(tone)
}

func (this *SoundPod) PlayNotes(s string) {
	tone := this.Mixer.FindVoice(Context, "tone")
	this.StopSong()
	tone.AddNoteStream(s)
	tone.SetVolume(TONE_VOLUME)
}

func (this *SoundPod) PlaySound(freq int) {
	tone := this.Mixer.FindVoice(Context, "tone")
	tone.SetFrequency(float64(freq))
	tone.SetVolume(TONE_VOLUME)
}

func (this *SoundPod) MakeCustomTone(freq int, duration int) {
	beep := this.Mixer.FindVoice(Context, "beep")
	beep.SetFrequency(float64(freq))
	beep.SetVolume(TONE_VOLUME)
}

func (this *SoundPod) GetAudioPort(name string) int {
	return this.Mixer.FindVoicePort(0, name)
}

func (this *SoundPod) RestalgiaCommand(index int, voice int, opcode int, value uint64) uint64 {
	this.Mixer.Slots[0].ExecuteOpcode(voice, opcode, value)
	return this.Mixer.Slots[0].BusValue
}

func (this *SoundPod) MakeTone(freq int, duration int) {
	beep := this.Mixer.FindVoice(Context, "beep")
	if beep == nil {
		log2.Printf("Could not find voice 'beep' needed for tone.")
		return
	}
	inst := restalgia.NewInstrument("WAVE=SQUARE:VOLUME=1.0:ADSR=0,0," + utils.IntToStr(duration) + ",1")
	inst.Apply(beep)
	beep.SetFrequency(float64(freq))
	beep.SetVolume(TONE_VOLUME)
	time.AfterFunc(time.Duration(duration)*time.Millisecond,
		func() {
			beep.SetVolume(0.0)
		},
	)
}

// func (this *SoundPod) TrackInit() {
// 	// for i := 0; i < 8; i++ {
// 	// 	if this.track[i] != nil {
// 	// 		this.Mixer.RemoveVoice(i, this.track[i])
// 	// 	}
// 	// 	x := restalgia.NewInstrument("WAVE=TRIANGLE:VOLUME=1.0:ADSR=1,0,1000,1")
// 	// 	this.track[i] = restalgia.NewVoice(fmt.Sprintf("trk%d", i), int(settings.SampleRate), restalgia.TRIANGLE, 1.0)
// 	// 	x.Apply(this.track[i])
// 	// 	this.track[i].SetVolume(0)
// 	// 	this.Mixer.AddVoice(this.track[i])
// 	// }
// }

// func (this *SoundPod) ToneInit() {
// 	if this.tone != nil {
// 		this.Mixer.RemoveVoice(this.tone)
// 	}
// 	this.instTone = restalgia.NewInstrument("WAVE=SAWTOOTH:VOLUME=1.0:ADSR=1,0,1000,1")
// 	this.tone = restalgia.NewVoice("tone", int(settings.SampleRate), restalgia.PULSE, 1.0)
// 	this.instTone.Apply(this.tone)
// 	this.tone.SetVolume(0)
// 	this.Mixer.AddVoice(this.tone)
// }

// func (this *SoundPod) BeepInit() {
// 	if this.beep != nil {
// 		this.Mixer.RemoveVoice(this.beep)
// 	}
// 	this.instBeep = restalgia.NewInstrument("WAVE=SINE:VOLUME=1.0:ADSR=1,0,1000,1")
// 	this.beep = restalgia.NewVoice("beep", int(settings.SampleRate), restalgia.SINE, 1.0)
// 	this.instBeep.Apply(this.beep)
// 	this.beep.SetVolume(0)
// 	this.Mixer.AddVoice(this.beep)
// }

func (this *SoundPod) StopSong() {
	if this.Song != nil {
		this.Song.Playing = false
	}
}

func (this *SoundPod) StartSong() {
	if this.Song != nil {
		this.Song.Playing = true
		this.Song.Start()
	}
}

func (this *SoundPod) SelectChannel(i int) {
	// for n := 0; n < memory.OCTALYZER_NUM_INTERPRETERS; n++ {
	// 	speaker := this.Mixer.FindVoice(n, "speaker")
	// 	lc := this.Mixer.FindVoice(n, "musicl")
	// 	rc := this.Mixer.FindVoice(n, "musicr")
	// 	tone := this.Mixer.FindVoice(n, "tone")
	// 	beep := this.Mixer.FindVoice(n, "beep")

	// 	speaker.Active = (n == i)
	// 	lc.Active = (n == i)
	// 	rc.Active = (n == i)
	// 	tone.Active = (n == i)
	// 	beep.Active = true
	// }
	this.Mixer.SlotSelect = i
}

func (this *SoundPod) StartAudio(bc *int) {
	output, err := driver.Get(44100, 2)
	if err != nil {
		panic(err)
	}

	this.Output = output
	this.buffersize = 384
	this.lastbuffer = 0
	this.BuzzerChannel = bc
	this.Mixer = restalgia.NewMixer()
	this.Output.SetPullSource(this.Mixer)
	this.Output.Start()
	log2.Printf("RESTALGIA has started at %dHz", settings.SampleRate)
}

func (this *SoundPod) Starved() {
	this.buffersize += 64
	fmt.Printf("* Growing audio buffer to %d because openal says so...\n", this.buffersize)
}

func (this *SoundPod) LoadSong(data []byte) {
	song, err := restalgia.NewSong("")
	buffer := bytes.NewBuffer(data)
	err = song.LoadReader(buffer)
	if err == nil {
		this.Song = song
		return
	}
	panic(err)
}

// func (this *SoundPod) GetAudioLag(index int) int64 {
// 	speaker := this.Mixer.FindVoice(index, "speaker")
// 	samples := speaker.GetOSC(0).GetWfCUSTOM().LateSamples
// 	return int64((float64(samples) / float64(settings.SampleRate)) * 1000000000)
// }

func (this *SoundPod) PassWaveBuffer(index int, channel int, data []float32, loop bool, rate int) {

	switch channel {
	case 0:
		speaker := this.Mixer.FindVoice(index, "speaker")

		//log2.Printf("Sending audio to speaker... %d samples", len(data))

		if speaker != nil {
			if speaker.GetOSC(0).GetWfCUSTOM().Waiting() >= 2 {
				speaker.GetOSC(0).GetWfCUSTOM().Drop()
			}

			speaker.GetOSC(0).GetWfCUSTOM().Stimulate(data, loop, true, rate)
			speaker.SetVolume(settings.SpeakerVolume[index])
			speaker.GetOSC(0).Trigger()
			//log2.Printf("sent")
		}
	case 1:
		log.Printf("Got cassette audio, length %d", len(data))
		cassette := this.Mixer.FindVoice(index, "cassette")

		if cassette != nil {
			if cassette.GetOSC(0).GetWfCUSTOM().Waiting() > 2 {
				cassette.GetOSC(0).GetWfCUSTOM().Drop()
			}

			cassette.GetOSC(0).GetWfCUSTOM().Stimulate(data, loop, true, rate)
			cassette.SetVolume(settings.SpeakerVolume[index])
			cassette.GetOSC(0).Trigger()
		}
	}

}

func (this *SoundPod) PassMusicBuffer(index int, data []float32, loop bool, rate int, channels int) {

	lc := this.Mixer.FindVoice(index, "musicl")
	rc := this.Mixer.FindVoice(index, "musicr")

	if lc == nil || rc == nil {
		return
	}

	for lc.GetOSC(0).GetWfCUSTOM().Waiting() > 2 {
		time.Sleep(5 * time.Millisecond)
	}

	alen := len(data) / channels

	if channels == 1 {
		lc.GetOSC(0).GetWfCUSTOM().Stimulate(data, loop, true, rate)
		lc.SetVolume(WAVE_VOLUME)
		lc.GetOSC(0).Trigger()
		rc.GetOSC(0).GetWfCUSTOM().Stimulate(data, loop, true, rate)
		rc.SetVolume(WAVE_VOLUME)
		rc.GetOSC(0).Trigger()
	} else if channels == 2 {

		l := make([]float32, alen)
		r := make([]float32, alen)

		for i, _ := range l {
			l[i] = data[i*2+0]
			r[i] = data[i*2+1]
		}

		lc.GetOSC(0).GetWfCUSTOM().Stimulate(l, loop, true, rate)
		lc.SetVolume(WAVE_VOLUME)
		lc.GetOSC(0).Trigger()

		rc.GetOSC(0).GetWfCUSTOM().Stimulate(r, loop, true, rate)
		rc.SetVolume(WAVE_VOLUME)
		rc.GetOSC(0).Trigger()

	}
}

func (this *SoundPod) SetBlockingMode(index int, blocking bool) {
	speaker := this.Mixer.FindVoice(index, "speaker")
	speaker.GetOSC(0).GetWfCUSTOM().Blocking = blocking
	//this.buzzer[index].GetOSC(0).GetWfCUSTOM().Drop()
	//fmt.Printf("--> Set speaker #%d source blocking = %v\n", index, blocking)
}

func (this *SoundPod) CheckToneLevel() {

	tone := this.Mixer.FindVoice(Context, "tone")

	now := time.Now().UnixNano()
	duration := int64(math.Abs(float64(now - this.lastSoundTick)))

	if duration == 0 {
		duration = 1000
	}

	freq := 1000000000 / (duration * 2)
	if (freq < 2) && (tone.GetVolume() > 0) {
		tone.SetVolume(0)
		//System.err.println("Silence voice");
	} else if (freq >= 2) && (tone.GetVolume() == 0) {
		tone.SetVolume(TONE_VOLUME)
	}
}

func (this *SoundPod) ClickFreq(usefreq int64) {
	// click the speaker
	tone := this.Mixer.FindVoice(Context, "tone")
	now := time.Now().UnixNano()

	if this.instTone == nil {
		this.instTone = restalgia.NewInstrument("WAVE=PULSE:VOLUME=1.0:ADSR=0,0,100,0")
		this.instTone.Apply(tone)
		tone.GetOSC(0).Trigger()
	}

	if (usefreq > 10) && (float64(usefreq) != tone.OSC[0].GetFrequency()) {
		tone.OSC[0].SetFrequency(float64(usefreq))
		if tone.GetVolume() == 0 {
			tone.SetVolume(TONE_VOLUME)
			tone.GetOSC(0).Trigger()
		}
	}

	this.lastSoundTick = now

}

func (this *SoundPod) Click(usefreq int64) {

	if this.sam == 0 {
		this.sam = -1
	}

	this.sam = -this.sam

	f := make([]float32, 4)
	for x := 0; x < len(f); x++ {
		f[x] = this.sam
		this.sam = -this.sam
	}

	this.PassWaveBuffer(0, 0, f, false, 48000) // pass and loop
}

func (this *SoundPod) LoadWAVE(index int, data []byte) {

	var b bytes.Buffer

	_, _ = b.Write(data)

	wr, err := wav.New(&b)
	if err != nil {
		return
	}

	var raw []float32

	raw, _ = wr.ReadFloats(wr.Samples)
	for len(raw) > 0 {
		var fl = make([]float32, len(raw)/int(wr.NumChannels))
		for i := 0; i < len(raw); i++ {
			if i%int(wr.NumChannels) == 0 {
				fl[i/int(wr.NumChannels)] = raw[i]
			}
		}

		this.PassWaveBuffer(index, 0, fl, false, int(wr.SampleRate))

		////fmt.Printntf("WAVE decoded %d samples\n", len(fl))

		raw, _ = wr.ReadFloats(48000 * int(wr.NumChannels))
	}

}

func (this *SoundPod) SendCommands(s string) {
	lines := utils.SplitLines([]byte(s))

	for _, l := range lines {
		control.ShellProcess(this.Mixer, strings.Trim(l, " "))
	}
}

// Command is used to process in-memory commands
func (this *SoundPod) Command(data []uint64) {

	tone := this.Mixer.FindVoice(Context, "tone")

	var rc types.RestalgiaCommand

	rc = types.RestalgiaCommand(data[0])

	switch rc {
	case types.RS_Instrument:
		// 1 = length
		// 2 = data start
		l := data[1]
		instdata := data[2 : 2+l]
		inststr := ""
		for _, v := range instdata {
			inststr = inststr + string(rune(v))
		}
		iii := restalgia.NewInstrument(inststr)
		iii.Apply(tone)

	case types.RS_Sound:
		// 1 = length
		// 2 = freq
		// 3 = ms duration
		//			 l := data[1]
		freq := int(data[2])
		ms := int(data[3])
		this.MakeCustomTone(freq, ms)

	case types.RS_PlayNotes:
		// 1 = length
		// 2 = data start
		l := data[1]
		notedata := data[2 : 2+l]
		notes := ""
		for _, v := range notedata {
			notes = notes + string(rune(v))
		}
		this.PlayNotes(notes)

	case types.RS_StopSong:
		if this.Song != nil {
			this.Song.Stop()
			this.Song = nil
		}

	case types.RS_PauseSong:
		if this.Song != nil {
			this.Song.Pause()
		}

	case types.RS_ResumeSong:
		if this.Song != nil {
			this.Song.Resume()
		}

	case types.RS_PlaySong:
		// 1 = length
		// 2 = data start
		l := data[1]
		log.Printf("In memory song size is %d bytes\n", l)
		rdata := data[2 : 2+l]
		bdata := make([]byte, l)
		for i, v := range rdata {
			bdata[i] = byte(v)
		}
		this.LoadSong(bdata)
		this.Song.Start()

	default:
		// nothing
	}

	data[0] = 0 // cancel it

}
