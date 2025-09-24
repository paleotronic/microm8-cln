package main

import (
	"github.com/gordonklaus/portaudio"
	"paleotronic.com/restalgia"
)

var SPEAKER *SoundPod

type SoundPod struct {
	Output        *portaudio.Stream
	buzzer        *restalgia.Voice
	tone          *restalgia.Voice
	instTone      *restalgia.Instrument
	Song          *restalgia.Song
	sam           float32
	lastSoundTick int64
}

func (this *SoundPod) StopAudio() {
	this.Output.Stop()
}

func (this *SoundPod) RedefineInst(s string) {
	this.tone.AddInstrumentChange(s)
}

func (this *SoundPod) RedefineInstImmediate(s string) {
	this.instTone = restalgia.NewInstrument(s)
	this.instTone.Apply(this.tone)
}

func (this *SoundPod) PlayNotes(s string) {
	this.StopSong()
	this.tone.AddNoteStream(s)
	this.tone.SetVolume(1.0)
}

func (this *SoundPod) PlaySound(freq int) {
	this.tone.SetFrequency(float64(freq))
	this.tone.SetVolume(1.0)
}

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

func (this *SoundPod) StartAudio() {

	output, err := portaudio.OpenDefaultStream(0, 1, 44100, )
	if err != nil {
		panic(err)
	}

	this.Output = output

	// Now lets add a tone generator
	this.instTone = restalgia.NewInstrument("WAVE=SQUARE:VOLUME=1.0:ADSR=0,0,100,0")
	this.tone = restalgia.NewVoice(44100, restalgia.PULSE, 1.0)
	this.instTone.Apply(this.tone)
	this.tone.SetVolume(1.0)
	this.tone.SetFrequency(40)
	this.tone.AddNoteStream("O5;G;O3")

	// Now lets add a custom wave generator
	fred := restalgia.NewInstrument("WAVE=CUSTOM:VOLUME=1.0:ADSR=0,0,1000,0")
	this.buzzer = restalgia.NewVoice(44100, restalgia.CUSTOM, 1.0)
	fred.Apply(this.buzzer)
	this.buzzer.SetVolume(1.0)
	this.buzzer.SetFrequency(1000)

	this.Output.Start()

	go func() {
		// feed tones
		size := 8

		fragment := make([][]float32, size)
		for i := 0; i < size; i++ {
			fragment[i] = make([]float32, 4096)
		}

		r := 0
		w := 1

		for {
			this.CheckToneLevel()

			//fragment := make([]float32, 2048)

			for i := 0; i < len(fragment[w]); i++ {

				vl := (this.tone.GetAmplitude() + this.buzzer.GetAmplitude()) / 2
				vr := vl

				if this.Song != nil {
					p := this.Song.PullSampleStereo()
					vl = (vl + p[0]) / 2
					vr = (vr + p[1]) / 2
				}

				fragment[w][i] = vl
				fragment[w][i+1] = vr
				i++
			}

			this.Output.Push(fragment[r])
			w = (w + 1) % size
			r = (r + 1) % size

		}
	}()

}

func (this *SoundPod) LoadSong(data []byte) {
	song, err := restalgia.NewSong("")
	buffer := bytes.NewBuffer(data)
	err = song.LoadReader(buffer)
	if err == nil {
		this.Song = song
	}
}

func (this *SoundPod) PassWaveBuffer(data []float32, loop bool) {
	this.buzzer.GetOSC(0).GetWfCUSTOM().Stimulate(data, loop)
	this.buzzer.SetVolume(1)
	this.buzzer.GetOSC(0).Trigger()
}

func (this *SoundPod) CheckToneLevel() {
	now := time.Now().UnixNano()
	duration := int64(math.Abs(float64(now - this.lastSoundTick)))

	if duration == 0 {
		duration = 1000
	}

	freq := 1000000000 / (duration * 2)
	if (freq < 2) && (this.tone.GetVolume() > 0) {
		this.tone.SetVolume(0)
		//System.err.println("Silence voice");
	} else if (freq >=2) && (this.tone.GetVolume() == 0) {
		this.tone.SetVolume(1)
	}
}

func (this *SoundPod) ClickFreq(usefreq int64) {
	// click the speaker
	now := time.Now().UnixNano()

	//System.out.println("Approx frequency is "+freq+"Hz, but using median "+usefreq+"Hz");

	if this.instTone == nil {
		this.instTone = restalgia.NewInstrument("WAVE=PULSE:VOLUME=1.0:ADSR=0,0,100,0")
		this.instTone.Apply(this.tone)
		this.tone.GetOSC(0).Trigger()
	}

	if (usefreq > 10) && (float64(usefreq) != this.tone.OSC[0].GetFrequency()) {
		this.tone.OSC[0].SetFrequency(float64(usefreq))
		if this.tone.GetVolume() == 0 {
			this.tone.SetVolume(1)
			this.tone.GetOSC(0).Trigger()
		}
	}

	this.lastSoundTick = now

}

func (this *SoundPod) Click(usefreq int64) {

	if this.sam == 0 {
		this.sam = -1
	}

	this.sam = -this.sam

	f := make([]float32, 1)
	for x := 0; x < len(f); x++ {
		f[x] = this.sam
	}

	this.PassWaveBuffer(f, true) // pass and loop
}

func (this *SoundPod) LoadWAVE(data []byte) {

	var b bytes.Buffer

	_, _ = b.Write(data)

	wr, err := wav.New(&b)
	if err != nil {
		return
	}

	////fmt.Printntf("WAVE %d channels,  sample rate %d, %d samples\n", wr.NumChannels, wr.SampleRate, wr.Samples)

	var raw []float32

	raw, _ = wr.ReadFloats(wr.Samples)
	for len(raw) > 0 {
		var fl = make([]float32, len(raw)/int(wr.NumChannels))
		for i := 0; i < len(raw); i++ {
			if i%int(wr.NumChannels) == 0 {
				fl[i/int(wr.NumChannels)] = raw[i]
			}
		}

		this.PassWaveBuffer(fl, false)

		////fmt.Printntf("WAVE decoded %d samples\n", len(fl))

		raw, _ = wr.ReadFloats(44100 * int(wr.NumChannels))
	}

}
