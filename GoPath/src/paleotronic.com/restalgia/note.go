package restalgia

import (
	"paleotronic.com/utils"
	"regexp"
	"strings"
)

type Note struct {
	Note    string
	Octave  int
	Volume  float64
	Panning float64
	Inst    string
}

var Hex string = "0123456789ABCDEF"

func NewNote(params string) *Note {
	pr := regexp.MustCompile("[\t ]+")
	this := &Note{Volume: -1, Panning: 0}
	parts := pr.Split(params, -1)
	idx := 0
	for _, p := range parts {

		p = strings.ToUpper(p)

		if idx == 0 {
			if p != "" && p != "--" {
				// parse note
				//System.Out.Println( "Note = "+p );

				oct := string(p[len(p)-1])
				octave := utils.StrToInt(oct)
				note := p[0 : len(p)-1]

				//System.Out.Println("Oct = "+oct+", note = "+note);

				f := NT.Frequency(note, octave)

				if f > 0 {
					this.Note = note
					this.Octave = octave
				}

			}
		} else {
			if len(p) > 0 && p[0] == 'V' {
				// following is hex
				hi := p[1]
				lo := p[2]

				v := (16 * strings.IndexRune(Hex, rune(hi))) + strings.IndexRune(Hex, rune(lo))

				vol := float64(v) / 255.0
				this.Volume = vol

			} else if len(p) > 0 && p[0] == 'I' {
				// following is hex representation of instrument

				alias := strings.ToLower(p[1:])
				this.Inst = alias // instrument alias

			} else if len(p) > 0 && p[0] == 'P' {
				// following is hex
				hi := p[1]
				lo := p[2]

				v := (16 * strings.IndexRune(Hex, rune(hi))) + strings.IndexRune(Hex, rune(lo))

				pan := float64(v-128) / 255.0
				this.Panning = pan

			}
		}

		//System.Out.Println("*** "+p);

		idx++
	}
	return this
}

func (this *Note) GetNote() string {
	return this.Note
}

func (this *Note) SetNote(note string) {
	this.Note = note
}

func (this *Note) GetOctave() int {
	return this.Octave
}

func (this *Note) SetOctave(octave int) {
	this.Octave = octave
}

func (this *Note) GetVolume() float64 {
	return this.Volume
}

func (this *Note) SetVolume(volume float64) {
	this.Volume = volume
}

func (this *Note) GetPanning() float64 {
	return this.Panning
}

func (this *Note) SetPanning(panning float64) {
	this.Panning = panning
}
