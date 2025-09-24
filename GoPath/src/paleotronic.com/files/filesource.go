package files

import "paleotronic.com/filerecord"

// FileProvider is an interface which represents a filesystem
// The filesystem could be a network drive, ftp, local path or other system
type FileProvider interface {
	GetPriority() int
	GetCurrentPath() string
	SetCurrentPath(p string, create bool) error
	GetFileContent(p string, f string) (filerecord.FileRecord, error)
	SetFileContent(p string, f string, data []byte) error
	Exists(p string, f string) bool
	Lock(p string, f string) error
	Meta(p string, f string, meta map[string]string) error
	Share(p string, f string) (string, string, bool, error)
	IsReadOnly() bool
	IsVisible() bool
	Dir(filespec string) ([]FileDef, []FileDef, error)
	DirFromBase(p string, filespec string) ([]FileDef, []FileDef, error)
	ChDir(p string) error
	MkDir(p string, f string) error
	Delete(p string, f string) error
	Validate(current *filerecord.FileRecord) (*filerecord.FileRecord, error)
	SetLoadAddress(p string, f string, address int) error
	Rename(p string, f string, nf string) error
}

// File error types
var (
	FPNotExist = "File Not Found"
	FPAccess   = "Access is denied"
	FPIOError  = "I/O Error"
)
