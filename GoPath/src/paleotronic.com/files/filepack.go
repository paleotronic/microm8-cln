package files

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"paleotronic.com/core/vduproto"
)

type FilePack struct {
	Data []byte
}

// GetMD5Sum returns the 16 byte MD5 of the file
func (fp *FilePack) GetMD5Sum() [16]byte {
	return md5.Sum(fp.Data)
}

// CompressGZIP returns a compressed version of the file data using the GZIP codec
func (fp *FilePack) CompressGZIP() []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(fp.Data)
	w.Close()

	return b.Bytes()
}

func NewFilePackFromBytes(b []byte) *FilePack {
	return &FilePack{Data: b}
}

func NewFilePackFromReader(r io.Reader) *FilePack {

	fp := &FilePack{Data: make([]byte, 0)}

	var n int
	var e error
	var b = make([]byte, 4096)

	for {
		n, e = r.Read(b)
		if n > 0 {
			fp.Data = append(fp.Data, b[0:n]...)
		}
		if e != nil {
			break
		}
	}

	return fp
}

func (fp *FilePack) Query() *vduproto.AssetQuery {

	aq := &vduproto.AssetQuery{MD5: fp.GetMD5Sum()}

	return aq

}

func (fp *FilePack) GetBlock(seq int) *vduproto.AssetBlock {

	blocks := (len(fp.Data) / 8192) + 1

	if seq >= blocks {
		return nil
	}

	start := 8192 * seq
	end := len(fp.Data)

	////fmt.Printntf("Slice of data %d:%d\n", start, end)

	if end-start > 8192 {
		end = start + 8192
	}

	ab := &vduproto.AssetBlock{MD5: fp.GetMD5Sum(), Sequence: uint16(seq), Total: uint16(blocks), Data: fp.Data[start:end]}

	return ab
}

func (fp *FilePack) CacheData() {
	// store to disk here if does not already exist

	cksum := fp.GetMD5Sum()

	name := hex.EncodeToString(cksum[0:16])

	filepath := GetUserDirectory(BASEDIR + "/cache")

	//////fmt.Printf("os.MkdirAll(\"%s\", 0755)\n", filepath)

	os.MkdirAll(filepath, 0755)

	filepath = filepath + "/" + name

	if _, err := os.Stat(filepath); err == nil {
		return
	}

	// Create it
	if f, err := os.Create(filepath); err == nil {
		_, err = f.Write(fp.Data)
		_ = f.Close()
		////fmt.Printntf("Cached resource locally %s\n", filepath)
	} else {
		////fmt.Printntln(err)
	}
}

// We recieved an MD5, is it cached?  if so return newfilepack with its data
func NewFilePackFromMD5(cksum [16]byte) (*FilePack, error) {

	name := hex.EncodeToString(cksum[0:16])

	filepath := GetUserDirectory(BASEDIR + "/cache")

	os.MkdirAll(filepath, 0755)

	filepath = filepath + "/" + name

	if _, err := os.Stat(filepath); err != nil {
		return nil, errors.New("Not found")
	}

	// Found in cache - awesome
	f, _ := os.Open(filepath)
	fp := NewFilePackFromReader(f)

	f.Close()
	return fp, nil
}
