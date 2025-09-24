package tracker

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"paleotronic.com/core/hardware/cpu/mos6502/asm"
	"paleotronic.com/files"
	"paleotronic.com/microtracker/mock"
	"paleotronic.com/octalyzer/assets"
	"paleotronic.com/utils"
)

type NoteMask byte

const (
	NmNote NoteMask = 1 << iota
	NmInst
	NmVolume
	NmCommand
)

// CompressNotes packs a track down to a list of unique notes
// and a list of indexes
func (a *ASMEncoder) CompressNotes(s *TSong, t *TTrack) ([]byte, [][]byte, [16]byte) {
	var m = map[[16]byte]int{}
	var notes = [][]byte{}
	var tracknoteptr = []byte{}
	var trackchecksum [16]byte
	var prevRegs = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	if t == nil {
		return tracknoteptr, notes, trackchecksum
	}

	for _, nn := range t.Notes {
		notedata := a.GetNoteData(s, nn, prevRegs)
		if notedata == nil {
			tracknoteptr = append(tracknoteptr, 0xff) // empty
			continue
		}
		ck := md5.Sum(notedata)
		if index, ok := m[ck]; ok {
			// got it in the cache
			tracknoteptr = append(tracknoteptr, byte(index))
			continue
		}
		// not in cache and not blank so add
		newindex := len(notes)
		notes = append(notes, notedata)
		m[ck] = newindex
		tracknoteptr = append(tracknoteptr, byte(newindex))
		prevRegs = notedata
	}

	// generate a full track checksum...
	d := make([]byte, 0)
	d = append(d, tracknoteptr...)
	for _, nd := range notes {
		d = append(d, nd...)
	}
	trackchecksum = md5.Sum(d)

	return tracknoteptr, notes, trackchecksum
}

func (a *ASMEncoder) GetNoteData(s *TSong, t *TNote, prevRegs []byte) []byte {

	if t == nil {
		return nil
	}

	/*
		0: tone period coarse | 0xff
		1: tone period fine | 0xff
		2: noise period | 0xff
		3: volume | 0xff
		4: env coarse | 0xff
		5: env fine | 0xff
		6: env shape | 0xff
		7: command | 0xff
		8: command val
		9: featureenable
	*/

	data := make([]byte, 10)
	for i, _ := range data {
		data[i] = 0xff
	}

	var squelch bool

	if t.Note != nil {
		notename := fmt.Sprintf("%s%d", *t.Note, *t.Octave)
		f, ok := mock.PSGNoteTable[notename]
		if ok {
			// got note
			period := mock.FreqHzToTonePeriod(f)
			data[0] = byte((period >> 8) & 0xff)
			data[1] = byte(period & 0xff)
		} else if *t.Note == "X" {
			data[0] = 0xfe
			data[1] = 0xff
			squelch = true
		}
	}

	if t.Instrument != nil {
		data[9] = 0x00
		i := s.Instruments[int(*t.Instrument)]
		if i.Voice.UseTone {
			data[9] |= 1
		}
		if i.Voice.UseNoise {
			data[2] = byte(i.Voice.NoisePeriod)
			data[9] |= 2
		}
		if i.Voice.UseEnv {
			data[6] = byte(i.Voice.EnvShape)
			data[4] = byte(i.Voice.EnvCoarse)
			data[5] = byte(i.Voice.EnvFine)
			data[3] = 16
			data[9] |= 4 // env period present...
		} else {
			data[3] = 0
		}
		data[3] = (data[3] & 0xf0) | byte(i.Voice.Amplitude)
	}

	if t.Volume != nil && !squelch {
		if data[3] == 0xff {
			data[3] = prevRegs[3]
			if data[3] == 0xff {
				data[3] = 0x00
			}
		}
		data[3] = (data[3] & 0xf0) | byte(*t.Volume)
	}

	if t.Command != nil {
		// cmd := string(*t.Command)
		// val := *t.CommandValue
		data[7] = byte(*t.Command)
		data[8] = byte(*t.CommandValue)
		// for tempos convert to a handy table index
		if data[7] == byte('f') {
			data[8] = byte(a.GetTempoIndex(data[8]))
		}
	}

	countff := 0
	for _, v := range data {
		if v == 0xff {
			countff++
		}
	}
	if countff == len(data) {
		return nil
	}

	return data
}

// func init() {
// 	log.Printf("%-3s %-7s %-3s", "BPM", "TIMER1", "CNT")
// 	for i := 1; i < 0x100; i++ {
// 		counter, ticks := BPMToIRQParams(1020484, byte(i))
// 		log.Printf("%.3d $%.6x $%.2x", i, counter, ticks)
// 	}
// 	os.Exit(0)
// }

type TempoParams struct {
	Counter uint16
	Ticks   byte
}

func (a *ASMEncoder) BPMToIRQParams(clockBase int, bpm byte) TempoParams {

	baseTicks := ((clockBase * 60) / int(bpm)) / 4

	ticks := (baseTicks / 65536) + 1
	if ticks < 2 {
		ticks++
	}
	baseTicks /= ticks

	return TempoParams{
		uint16(baseTicks),
		byte(ticks),
	}

}

type CompressedTrack struct {
	Label         string
	TrackNotePtrs []byte
	Notes         [][]byte
}

type CompressedPattern struct {
	Tracks []string
}

type ASMEncoder struct {
	Tracks            map[[16]byte]CompressedTrack
	Patterns          []CompressedPattern
	Song              *TSong
	PatternIndexes    []byte
	NumPatterns       int
	TrackCounter      int
	TempoTable        []TempoParams
	TempoToTableIndex map[byte]int
}

func (a *ASMEncoder) GetTempoIndex(tempo byte) int {
	index, ok := a.TempoToTableIndex[tempo]
	if ok {
		return index
	}
	index = len(a.TempoTable)
	a.TempoToTableIndex[tempo] = index
	a.TempoTable = append(a.TempoTable, a.BPMToIRQParams(1020484, tempo))
	return index
}

func NewASMEncoder() *ASMEncoder {
	a := &ASMEncoder{
		Tracks:            make(map[[16]byte]CompressedTrack),
		PatternIndexes:    make([]byte, 0),
		Patterns:          make([]CompressedPattern, 0),
		TempoTable:        make([]TempoParams, 0),
		TempoToTableIndex: make(map[byte]int),
	}
	return a
}

func (a *ASMEncoder) patternUsed(pattid int) bool {
	for _, v := range a.PatternIndexes {
		if v == byte(pattid) {
			return true
		}
	}
	return false
}

func (a *ASMEncoder) Compress(s *TSong) {
	a.Song = s
	maxPatternIndex := -1
	a.PatternIndexes = []byte{}
	for _, v := range s.PatternList {
		if v == -1 {
			v = 0xff
		} else {
			if v > maxPatternIndex {
				maxPatternIndex = v
			}
		}
		a.PatternIndexes = append(a.PatternIndexes, byte(v))
	}

	// now we know how many patterns to process
	a.NumPatterns = maxPatternIndex + 1
	for pnum := 0; pnum < a.NumPatterns; pnum++ {
		//for _, p := range s.Patterns {
		p := s.Patterns[pnum]
		cp := CompressedPattern{
			Tracks: make([]string, 6),
		}
		// process for each pattern
		if a.patternUsed(pnum) {
			for i, t := range p.Tracks {
				tracknoteptrs, notes, checksum := a.CompressNotes(s, t)
				if len(tracknoteptrs) == 0 {
					cp.Tracks[i] = "EMPTYTRACK"
					continue
				}
				// ok have a track, check cache for it?
				if ct, ok := a.Tracks[checksum]; ok {
					cp.Tracks[i] = ct.Label
					continue
				}
				// not cached
				label := fmt.Sprintf("TRACK%d", a.TrackCounter)
				a.TrackCounter++
				ct := CompressedTrack{
					Label:         label,
					TrackNotePtrs: tracknoteptrs,
					Notes:         notes,
				}
				a.Tracks[checksum] = ct
				cp.Tracks[i] = ct.Label
			}
		}

		a.Patterns = append(a.Patterns, cp)
	}

}

func (a *ASMEncoder) Encode(s *TSong, filename string) error {

	// compress data structure with anger
	a.Compress(s)

	//startAddr := 8192

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	writeblank := func() {
		f.Write([]byte("\r\n\r\n"))
	}

	writeln := func(label string, instruction string, comment string) {
		f.Write([]byte(fmt.Sprintf("%-20s%-60s; %-20s\r\n", label, instruction, comment)))
	}

	writecomment := func(comment string) {
		f.Write([]byte("; " + comment + "\r\n"))
	}

	writebytes := func(label string, data []byte) {
		width := 8
		line := ""
		for i, v := range data {
			if i%width == 0 {
				if line != "" {
					writeln(label, line, "")
					label = ""
				}
				line = ".byte "
			}
			if line != ".byte " {
				line += ","
			}
			line += fmt.Sprintf("$%.2x", v)
		}
		if line != ".byte " {
			writeln(label, line, "")
		}
	}

	// writewords := func(label string, data []uint16) {
	// 	width := 8
	// 	line := ""
	// 	for i, v := range data {
	// 		if i%width == 0 {
	// 			if line != "" {
	// 				writeln(label, line, "")
	// 				label = ""
	// 			}
	// 			line = ".word "
	// 		}
	// 		if line != ".word " {
	// 			line += ","
	// 		}
	// 		line += fmt.Sprintf("$%.4x", v)
	// 	}
	// 	if line != ".word " {
	// 		writeln(label, line, "")
	// 	}
	// }

	writestring := func(label string, str string) {
		writeln(label, ".byte \""+str+"\"", "")
	}

	writecomment("This listing is generated by microTracker")
	//writeln("", fmt.Sprintf("ORG $%.4x", startAddr), "start address (change it)")

	writeblank()

	writecomment("Song header")
	writestring("SNGID", "MSNG")

	writecomment("Song meta data")
	writestring("SNGTITLE", s.Name)
	writebytes("", []byte{0})

	writebytes("SNGVERSION", []byte{1, 0})

	// writeblank()
	// writecomment("lookup table pointers to find data fast")
	// writeln("PATTTABLE", ".word PATTADDR", "pointer to start of pattern lookup")
	// writeln("INSTTABLE", ".word INSTADDR", "pointer to start of instrument lookup")
	// writeln("SONGTABLE", ".word SONGADDR", "pointer to start of song lookup")

	writeblank()
	writecomment("song data begins: $ff == null pattern")
	writebytes("SONGADDR", a.PatternIndexes)

	// // inst data
	// writeblank()
	// writecomment("instrument data begins")
	// writeln("INSTDATA", "", "")
	// numInst := s.NumInstruments()
	// instLabels := make([]string, numInst)
	// for i := 0; i < numInst; i++ {
	// 	log.Printf("--> Encoding instrument %.2x", i)
	// 	inst := s.Instruments[i]
	// 	id := inst.Bytes()
	// 	log.Printf("    Inst data is %d bytes", len(id))
	// 	label := fmt.Sprintf("INST%.2x", i)
	// 	instLabels[i] = label
	// 	writebytes(
	// 		label,
	// 		id,
	// 	)
	// }

	// // write out instrument lookup
	// writeblank()
	// writecomment("lookup table")
	// writeln("INSTADDR", "", "")
	// for _, v := range instLabels {
	// 	writeln("", ".word "+v, "")
	// }

	writeblank()
	writecomment("pattern data begins")
	writeln("PATTDATA", "", "")
	pattLabels := make([]string, len(a.Patterns))
	for i, cp := range a.Patterns {
		writeblank()
		writecomment(fmt.Sprintf("Start of track map for pattern %d", i))
		label := fmt.Sprintf("PATTERN%d", i)
		writeln(label, "", "")

		if a.patternUsed(i) {
			// write track ptrs
			for _, tp := range cp.Tracks {
				writeln("", ".word "+tp, "")
			}
			writecomment("End of track map")
		} else {
			writecomment("this pattern is not used, placeholder only")
		}
		pattLabels[i] = label
	}

	// write out pattern lookup
	writeblank()
	writecomment("lookup table")
	writeln("PATTADDR", "", "")
	for _, v := range pattLabels {
		writeln("", ".word "+v, "")
	}

	writeblank()
	writecomment("track data begins")
	writeln("TRACKDATA", "", "")
	trackLabels := make([]string, len(a.Tracks))
	i := 0
	for _, ct := range a.Tracks {
		writeblank()
		writecomment(fmt.Sprintf("Start of track map for pattern %d", i))
		label := ct.Label
		notetablelabel := fmt.Sprintf("%sNOTETABLE", label)
		writeln(label, "", "")
		writeln(label+"LOOKUP", ".word "+notetablelabel, "")
		writebytes(label+"PTRS", ct.TrackNotePtrs)
		writeln(label+"NOTES", "", "")
		noteLabels := make([]string, len(ct.Notes))
		for j, nd := range ct.Notes {
			nl := fmt.Sprintf("%sNOTE%d", label, j)
			writebytes(nl, nd)
			noteLabels[j] = nl
		}

		writeblank()
		writecomment("note lookup table")
		writeln(notetablelabel, "", "")
		for _, v := range noteLabels {
			writeln("", ".word "+v, "")
		}

		trackLabels[i] = label
		i++
	}

	writeblank()
	writecomment("track lookup table")
	writeln("TRACKADDR", "", "")
	for _, v := range trackLabels {
		writeln("", ".word "+v, "")
	}

	// barf out tempo data
	writeblank()
	writecomment("tempo timer1lo table")
	writeln("TEMPOTIMER1LO", "", "")
	for _, v := range a.TempoTable {
		writebytes("", []byte{byte(v.Counter & 0xff)})
	}

	writeblank()
	writecomment("tempo timer1hi table")
	writeln("TEMPOTIMER1HI", "", "")
	for _, v := range a.TempoTable {
		writebytes("", []byte{byte(v.Counter >> 8)})
	}

	writeblank()
	writecomment("tempo tickcount table")
	writeln("TEMPOTICKCOUNT", "", "")
	for _, v := range a.TempoTable {
		writebytes("", []byte{v.Ticks})
	}

	return nil
}

type CompileMode int

const (
	CmBINCombined CompileMode = iota
	CmBINSongOnly
	CmASMSongOnly
)

func (a *ASMEncoder) CompileSong(s *TSong, mode CompileMode, address int) (string, error) {

	// tempdir := files.GetUserDirectory(files.BASEDIR + "/temp")
	// err := os.MkdirAll(tempdir, 0755)
	// if err != nil {
	// 	return "", err
	// }
	var main string
	var startAddr int
	var skipASM bool

	switch mode {
	case CmASMSongOnly:
		skipASM = true
		main = "song.asm"
		startAddr = address
	case CmBINCombined:
		skipASM = false
		main = "player.asm"
		startAddr = address
	case CmBINSongOnly:
		skipASM = false
		main = "song.asm"
		startAddr = address
	}

	temp, err := ioutil.TempDir("", "song-asm-")
	if err != nil {
		return "", err
	}

	dir := "bootsystem/boot/templates/tracker"

	filelist, err := assets.AssetDir(dir)
	if err != nil {
		return "", err
	}

	// write files to temp dir
	for _, file := range filelist {
		data, err := assets.Asset(dir + "/" + file)
		if err != nil {
			return "", err
		}
		outfile := fmt.Sprintf("%s/%s", temp, file)
		//log.Printf("Copying template %s to %s", file, outfile)
		err = ioutil.WriteFile(outfile, data, 0755)
		if err != nil {
			return "", err
		}
	}

	outfile := fmt.Sprintf("%s/%s", temp, "song.asm")
	//log.Printf("Encoding to %s", outfile)
	err = a.Encode(s, outfile)
	if err != nil {
		return "", err
	}

	if skipASM {
		data, err := ioutil.ReadFile(outfile)
		if err != nil {
			return "", err
		}
		outname := strings.ToLower(strings.Replace(s.Name, " ", "", -1))
		if len(outname) > 8 {
			outname = outname[:8]
		}
		outname = fmt.Sprintf("/local/song-%s.asm", outname)
		err = files.WriteBytesViaProvider(files.GetPath(outname), files.GetFilename(outname), data)
		if err != nil {
			return "", err
		}
		return outname, nil
	}

	addr := startAddr

	// now try assemble it
	lines, err := utils.ReadTextFile(temp + "/" + main)
	if err != nil {
		return "", err
	}
	asm := asm.NewAsm6502()
	asm.SearchPath = temp
	blocks, _, _, err := asm.AssembleMultipass(lines, addr)
	//log.Printf("Blocks = %v, Addr = %d, Something = %s, err = %v", blocks, saddr, something, err)
	if err != nil {
		return "", err
	}

	outname := strings.ToLower(strings.Replace(s.Name, " ", "", -1))
	if len(outname) > 8 {
		outname = outname[:8]
	}
	outname = fmt.Sprintf("/local/%s#0x%.4x.bin", outname, addr)
	err = files.WriteBytesViaProvider(files.GetPath(outname), files.GetFilename(outname), blocks[addr])

	return outname, err

}
