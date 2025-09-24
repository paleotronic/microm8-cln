package filerecord

import (
	"encoding/json"
	"errors"
	"time"

	"gopkg.in/mgo.v2/bson"
	"paleotronic.com/log"
)

type FileRecord struct {
	ID            bson.ObjectId `bson:"_id,omitempty";json:"-"`
	FileName      string		 `json:"filename"`
	FilePath      string		 `json:"path"`
	Categories    []string		 `json:"-"`
	Owner         string		 `json:"-"`
	UserCanWrite  bool		 `json:"-"`
	Content       []byte		`json:"content"`
	ContentSize   int			`json:"size"`
	Address       int		 `json:"address"`
	Version       int // increment this when we update a file, copy old record into file_versions
	Checksum      [16]byte		 `json:"-"`
	Created       time.Time		 `json:"-"`
	Modified      time.Time		 `json:"-"`
	Creator       string		 `json:"-"`
	Modifier      string		 `json:"-"`
	Description   string		`json:"description"`
	Locked        bool		 `json:"-"`
	LockUser      string		 `json:"-"`
	LockedTime    time.Time		 `json:"-"`
	LocalPath     string		 `json:"-"`
	Directory     bool			`json:"dir"`
	DirIsFile     bool		 `json:"-"`
	Deleted       bool			`json:"deleted"`
	MetaData      map[string]string		 `json:"-"`
	PositionRead  int		 `json:"-"`
	PositionWrite int		 `json:"-"`
	RecordSize    int		 `json:"-"`
}

func (this *FileRecord) AddMeta(name, value string) {
	if this.MetaData == nil {
		this.MetaData = make(map[string]string)
	}

	if value == "" {
		delete(this.MetaData, name)
	}

	this.MetaData[name] = value
}

func (this *FileRecord) GetMeta(name, def string) string {
	if this.MetaData == nil {
		this.MetaData = make(map[string]string)
	}

	value, ok := this.MetaData[name]
	if ok {
		return value
	}

	return def
}

func (this *FileRecord) JSON() []byte {
	b, _ := json.Marshal(this)
	log.Printf("JSON: %s", string(b))
	return b
}

func (this *FileRecord) UnJSON(data []byte) {
	_ = json.Unmarshal(data, this)
}

func NewFileRecord(p, f string) *FileRecord {
	this := &FileRecord{FilePath: p, FileName: f, Content: make([]byte, 0)}

	return this
}

func (this *FileRecord) Write(b byte) {
	this.Content = append(this.Content, b)
}

func (this *FileRecord) WriteBytes(b []byte) {
	this.Content = append(this.Content, b...)
}

func (this *FileRecord) ReadBytes(b []byte) (int, error) {

	avail := len(this.Content) - this.PositionRead

	count := len(b)
	if avail > 0 {
		if avail < count {
			count = avail
		}
		chunk := this.Content[this.PositionRead : this.PositionRead+count]
		for i, v := range chunk {
			b[i] = v
		}
		this.PositionRead += count
		return count, nil
	}

	return 0, errors.New("EOF")

}

func (this *FileRecord) CanWrite(username string) bool {
	return this.UserCanWrite
}
