package files

import (
	"errors"
	"paleotronic.com/fmt"

	"paleotronic.com/filerecord"
	"paleotronic.com/log"
)

var Buffers map[string]*filerecord.FileRecord = make(map[string]*filerecord.FileRecord)

var Writer string
var Reader string
var Dribble string

// DOSOPEN opens a file, reading if necessary from the network
func DOSOPEN(p, f string, rs int) error {

	log.Println(Buffers)

	_, ok := Buffers[p+"/"+f]
	if ok {
		return errors.New("FILE ALREADY OPEN")
	}

	fr, err := ReadBytesViaProvider(p, f)
	if err != nil {
		fr = *filerecord.NewFileRecord(p, f)
	}
	fr.RecordSize = rs
	Buffers[p+"/"+f] = &fr
	log.Println(Buffers)

	return nil
}

func DOSDRIBBLE(p, f string) error {

	log.Println(Buffers)

	_, ok := Buffers[p+"/"+f]
	if ok {
		return errors.New("FILE ALREADY OPEN")
	}

	fr, err := ReadBytesViaProvider(p, f)
	if err != nil {
		fr = *filerecord.NewFileRecord(p, f)
	}
	Buffers[p+"/"+f] = &fr
	log.Println(Buffers)
	Dribble = p + "/" + f
	fr.Locked = true

	return nil
}

func DOSNODRIBBLE() error {

	if Dribble == "" {
		return nil
	}

	p := GetPath(Dribble)
	f := GetFilename(Dribble)

	fr, ok := Buffers[p+"/"+f]
	if !ok {
		return nil
	}

	defer delete(Buffers, p+"/"+f)

	if fr.Locked {

		// Write out file
		e := WriteBytesViaProvider(p, f, fr.Content)
		return e

	}

	log.Println("NODRIBBLE being called")

	return nil
}

func DOSDRIBBLEDATA(b []byte) error {
	if Dribble == "" {
		return nil
	}
	return DOSPRINT(GetPath(Dribble), GetFilename(Dribble), b)
}

// DOSREAD marks an open file for reading
func DOSREAD(p, f string, rec int) error {

	fr, ok := Buffers[p+"/"+f]
	if !ok {
		return errors.New("FILE NOT OPEN")
	}
	fr.ContentSize = len(fr.Content)

	if fr.ContentSize == 0 {
		delete(Buffers, p+"/"+f)
		return errors.New("FILE NOT FOUND")
	}

	if fr.RecordSize != 0 {
		fmt.Printf("Seeking to offset %d\n", rec*fr.RecordSize)
		fr.PositionRead = rec * fr.RecordSize
		fr.PositionWrite = -1
	} else {
		fr.PositionRead = 0
		fr.PositionWrite = -1
	}

	return nil
}

// DOSWRITE marks an open file for writing (erases data)
func DOSSEEK(p, f string, offset int) error {

	fr, ok := Buffers[p+"/"+f]
	if !ok {
		return errors.New("FILE NOT OPEN")
	}

	if offset < 0 || offset > len(fr.Content) {
		return errors.New("FILE POINTER OUT OF RANGE")
	}

	if fr.PositionRead != -1 {
		fr.PositionRead = offset
	}
	if fr.PositionWrite != -1 {
		fr.PositionWrite = offset
	}

	fr.ContentSize = len(fr.Content) - offset

	return nil
}

func DOSREADPOS(p, f string) (int, error) {

	fr, ok := Buffers[p+"/"+f]
	if !ok {
		return 0, errors.New("FILE NOT OPEN")
	}

	return fr.PositionRead, nil
}

func DOSLEN(p, f string) (int, error) {

	fr, ok := Buffers[p+"/"+f]
	if !ok {
		return 0, errors.New("FILE NOT OPEN")
	}

	return len(fr.Content), nil
}

func DOSWRITEPOS(p, f string) (int, error) {

	fr, ok := Buffers[p+"/"+f]
	if !ok {
		return 0, errors.New("FILE NOT OPEN")
	}

	return fr.PositionWrite, nil
}

// DOSWRITE marks an open file for writing (erases data)
func DOSWRITE(p, f string, rec int) error {

	fr, ok := Buffers[p+"/"+f]
	if !ok {
		return errors.New("FILE NOT OPEN")
	}
	// fr.Content = make([]byte, 0) - don't wipe it yet as we may append
	fr.Locked = true
	if fr.RecordSize != 0 {
		fr.PositionRead = -1
		fr.PositionWrite = rec * fr.RecordSize
		needed := (rec + 1) * fr.RecordSize
		if len(fr.Content) < needed {
			fmt.Printf("Extend random access file %s to %d bytes\n", f, needed)
			chunk := make([]byte, needed-len(fr.Content))
			fr.Content = append(fr.Content, chunk...)
		}
	} else {
		fr.PositionRead = -1
		fr.PositionWrite = 0
	}
	return nil
}

// DOSAPPEND marks an open file for append (keeps data)
func DOSAPPEND(p, f string) error {

	fr, ok := Buffers[p+"/"+f]
	if !ok {
		return errors.New("FILE NOT OPEN")
	}

	if fr.RecordSize != 0 {
		return errors.New("FILE NOT SEQUENTIAL")
	}

	fr.Locked = true
	fr.PositionRead = -1
	fr.PositionWrite = len(fr.Content)
	return nil
}

// DOSCLOSE closes an opened file, writing it if necessary
func DOSCLOSE(p, f string) error {

	fr, ok := Buffers[p+"/"+f]
	if !ok {
		return nil
	}

	defer delete(Buffers, p+"/"+f)

	if fr.Locked {

		// Write out file
		e := WriteBytesViaProvider(p, f, fr.Content)
		return e

	}

	log.Println("DOSCLOSE being called")

	return nil
}

func DOSCLOSEALL() error {

	log.Println("DOSCLOSEALL being called")

	for _, fr := range Buffers {
		if fr.Locked {
			// Write out file
			_ = WriteBytesViaProvider(fr.FilePath, fr.FileName, fr.Content)
		}
	}

	Buffers = make(map[string]*filerecord.FileRecord)
	return nil
}

// DOSPRINT writes bytes to an open file
func DOSPRINT(p, f string, b []byte) error {

	fr, ok := Buffers[p+"/"+f]
	if !ok {
		return errors.New("FILE NOT OPEN")
	}

	if !fr.Locked {
		return errors.New("FILE NOT IN WRITE MODE")
	}

	if fr.RecordSize == 0 {
		if len(fr.Content) > fr.PositionWrite {
			fr.Content = fr.Content[0:fr.PositionWrite]
		}

		fr.WriteBytes(b)
		fr.PositionWrite += len(b) // move write pointer
	} else {
		// random access
		for i, v := range b {
			if fr.PositionWrite+i < len(fr.Content) {
				fr.Content[fr.PositionWrite+i] = v
			} else {
				return errors.New("RECORD TOO LONG")
			}
		}
		fr.PositionWrite += len(b)
	}

	return nil
}

// DOSGET reads a single character from the file
func DOSGET(p, f string) (byte, error) {

	fr, ok := Buffers[p+"/"+f]
	if !ok {
		return 0, errors.New("FILE NOT OPEN")
	}

	//ptr := len(fr.Content) - fr.ContentSize // position in file

	ptr := fr.PositionRead

	if ptr == len(fr.Content) {
		return 0, errors.New("EOF")
	}

	b := fr.Content[ptr]
	fr.ContentSize--
	fr.PositionRead++

	return b, nil
}

// DOSINPUT reads until either a newline or a comma
func DOSINPUT(p, f string) ([]byte, error) {

	var complete bool

	data := make([]byte, 0)

	b, e := DOSGET(p, f)

	for e == nil && !complete {

		switch rune(b) {
		case 13: // nothing
			if len(data) > 0 {
				return data, nil
			}
		case 10: // end record
			if len(data) > 0 {
				return data, nil
			}
		case ',': // end record
			return data, nil
		default:
			data = append(data, b)
		}

		b, e = DOSGET(p, f)
	}

	return data, e

}

func DOSBUFFERS() []string {
	out := make([]string, 0)
	for k, _ := range Buffers {
		out = append(out, GetFilename(k))
	}
	return out
}
