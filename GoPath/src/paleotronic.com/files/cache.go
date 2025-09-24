package files

import (
	"strings"

	"paleotronic.com/api"

	"os"

	"paleotronic.com/log"

	"errors"

	"crypto/md5"

	"encoding/hex"

	"gopkg.in/mgo.v2/bson"
	"paleotronic.com/filerecord"
)

type Cache struct {
	BaseDir string
}

func ExtractFromCache(pathname string) error {
	c := New(BASEDIR, "FILECACHE")
	if data, e := c.Extract(pathname); e == nil {
		fname := strings.TrimSuffix(GetFilename(pathname), ".obj")
		return WriteBytes(fname, data, false)
	}
	return errors.New("Extract failed")
}

func New(base, cachedir string) *Cache {
	return &Cache{
		BaseDir: base + "/" + cachedir,
	}
}

func (c *Cache) getPath(scope string, filename string) (string, string) {

	return c.BaseDir + "/" + s8webclient.CONN.Username + "/" + scope + "/" + sanitize(GetPath(filename)), sanitize(GetFilename(filename))

}

func (c *Cache) IsCached(scope, filename string) bool {
	p, f := c.getPath(scope, filename)
	fname := GetUserDirectory(p + "/" + f)
	return Exists(fname + ".obj")
}

func (c *Cache) Upsert(scope string, data *filerecord.FileRecord) error {

	filename := strings.TrimPrefix(strings.TrimSuffix(data.FilePath+"/"+data.FileName, "/"), "/")
	fr, err := c.Get(scope, filename)
	if err != nil {
		fr.Content = data.Content
		fr.Checksum = md5.Sum(fr.Content)
		fr.ContentSize = len(fr.Content)
		if data.Address != 0 {
			fr.Address = data.Address
		}
		return c.Put(scope, fr)
	}
	return c.Put(scope, data)

}

func (c *Cache) Put(scope string, data *filerecord.FileRecord) error {

	filename := strings.TrimPrefix(strings.TrimSuffix(data.FilePath+"/"+data.FileName, "/"), "/")

	p, f := c.getPath(scope, filename)
	fname := GetUserDirectory(p)
	os.MkdirAll(fname, 0755)

	//if data.Checksum == [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0} {
	// data.Checksum = md5.Sum(data.Content)
	// log.Printf("Setting data checksum %v", data.Checksum)
	//}
	data.Checksum = md5.Sum(data.Content)
	data.ContentSize = len(data.Content)

	log.Printf("NetCache putting file: %s (len %d, ck %v)\n", fname, len(data.Content), data.Checksum)

	b, _ := bson.Marshal(data)

	//_ = WriteBytes(fname+"/"+f, data.Content, false)

	return WriteBytes(fname+"/"+f+".obj", b, false)

}

func (c *Cache) Extract(pathname string) ([]byte, error) {
	var data []byte
	var err error
	var fr *filerecord.FileRecord = &filerecord.FileRecord{}
	data, err = ReadBytes(pathname)
	if err != nil {
		return []byte(nil), err
	}
	//fr.UnJSON(data)
	bson.Unmarshal(data, fr)

	log.Printf("- Found with signature %s\n", hex.EncodeToString(fr.Checksum[:]))
	log.Println("- Verifying signature")
	log.Printf("- data size is %d", len(fr.Content))
	contentSum := md5.Sum(fr.Content)
	log.Printf("- Content checksum is %s\n", hex.EncodeToString(contentSum[:]))
	if contentSum != fr.Checksum {
		log.Println("Bad signature!")
		//return []byte(nil), errors.New("Signature mismatch")
	} else {
		log.Println("Ok!")
	}
	return fr.Content, nil
}

func (c *Cache) Get(scope, filename string) (*filerecord.FileRecord, error) {

	p, f := c.getPath(scope, strings.TrimPrefix(strings.TrimSuffix(filename, "/"), "/"))

	var fr *filerecord.FileRecord = &filerecord.FileRecord{}
	var err error

	fname := GetUserDirectory(p + "/" + f)

	log.Printf("NetCache checking file: %s\n", fname)

	if c.IsCached(scope, filename) {
		var data []byte
		data, err = ReadBytes(fname + ".obj")
		s, _ := os.Stat(fname + ".obj")
		if err != nil {
			return fr, err
		}
		//fr.UnJSON(data)
		bson.Unmarshal(data, fr)
		fr.Modified = s.ModTime()

		log.Printf("- Found with signature %s\n", hex.EncodeToString(fr.Checksum[:]))
		log.Println("- Verifying signature")
		contentSum := md5.Sum(fr.Content)
		log.Printf("- Content checksum is %s %v\n", hex.EncodeToString(contentSum[:]), contentSum)
		if contentSum != fr.Checksum {
			log.Println("Bad signature!")
			return fr, errors.New("Signature mismatch")
		} else {
			log.Println("Ok!")
		}

		return fr, nil
	}

	log.Println("- Not found")

	return fr, errors.New("not found")

}

func sanitize(s string) string {
	s = strings.Replace(s, " ", "_", -1)
	s = strings.Replace(s, "(", "", -1)
	s = strings.Replace(s, ")", "", -1)
	s = strings.Replace(s, "[", "", -1)
	s = strings.Replace(s, "]", "", -1)
	s = strings.Replace(s, "'", "", -1)
	return s
}
