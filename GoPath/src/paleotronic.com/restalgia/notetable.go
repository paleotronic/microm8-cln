package restalgia

import (
	"strings"
)

/*
 * This class encapsulates the music scale of reference frequencies
 * Source: http://www.seventhstring.com/resources/notefrequencies.html
 */

var nSharps = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
var nFlats = []string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}
var nOctaves = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"}
var nFrequencies = []float64{
	16.35, 17.32, 18.35, 19.45, 20.6, 21.83, 23.12, 24.5, 25.96, 27.5, 29.14, 30.87,
	32.7, 34.65, 36.71, 38.89, 41.2, 43.65, 46.25, 49, 51.91, 55, 58.27, 61.74,
	65.41, 69.3, 73.42, 77.78, 82.41, 87.31, 92.5, 98, 103.8, 110, 116.5, 123.5,
	130.8, 138.6, 146.8, 155.6, 164.8, 174.6, 185, 196, 207.7, 220, 233.1, 246.9,
	261.6, 277.2, 293.7, 311.1, 329.6, 349.2, 370, 392, 415.3, 440, 466.2, 493.9,
	523.3, 554.4, 587.3, 622.3, 659.3, 698.5, 740, 784, 830.6, 880, 932.3, 987.8,
	1047, 1109, 1175, 1245, 1319, 1397, 1480, 1568, 1661, 1760, 1865, 1976,
	2093, 2217, 2349, 2489, 2637, 2794, 2960, 3136, 3322, 3520, 3729, 3951,
	4186, 4435, 4699, 4978, 5274, 5588, 5920, 6272, 6645, 7040, 7459, 7902,
}

type NoteTable struct {
}

var NT NoteTable

func (n NoteTable) NoteIndex(note string) int {
	r := -1

	for i := 0; i < len(nSharps); i++ {
		flat := nFlats[i]
		sharp := nSharps[i]
		if note == flat || note == sharp {
			r = i
			break
		}
	}

	return r
}

func (n NoteTable) GetValue(octave int, note int) float64 {
	return nFrequencies[(octave*12)+note]
}

func (n NoteTable) Frequency(note string, octave int) float64 {
	noteIndex := n.NoteIndex(note)
	if noteIndex == -1 {
		return -1
	}
	if octave < 0 {
		octave = 0
	}
	if octave > 8 {
		octave = 8
	}
	return n.GetValue(octave, noteIndex)
}

func (n NoteTable) IsSemiTone(noteIndex int) bool {
	return strings.HasSuffix(nSharps[noteIndex], "#")
}

func (n NoteTable) IsFullTone(noteIndex int) bool {
	return !n.IsSemiTone(noteIndex)
}

func (n NoteTable) Higher(octave, noteIndex int) (int, int) {
	noteIndex++
	if noteIndex >= 12 {
		noteIndex = 0
		octave++
		if octave > 8 {
			return -1, -1
		}
	}
	return octave, noteIndex
}

func (n NoteTable) Lower(octave, noteIndex int) (int, int) {
	noteIndex--
	if noteIndex < 0 {
		noteIndex = 11
		octave--
		if octave < 0 {
			return -1, -1
		}
	}
	return octave, noteIndex
}

func (n NoteTable) SemiToneUp(octave int, noteIndex int) (int, int) {

	oo, nn := n.Higher(octave, noteIndex)
	if oo == -1 {
		return -1, -1
	}
	// is note semi tone
	if n.IsFullTone(nn) && n.IsFullTone(noteIndex) {
		return -1, -1
	}
	return oo, nn
}

func (n NoteTable) SemiToneDown(octave int, noteIndex int) (int, int) {

	oo, nn := n.Lower(octave, noteIndex)
	if oo == -1 {
		return -1, -1
	}
	// is note semi tone
	if n.IsFullTone(nn) && n.IsFullTone(noteIndex) {
		return -1, -1
	}
	return oo, nn
}
