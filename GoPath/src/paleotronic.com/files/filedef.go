package files

// FileType represents a filetype 
type FileType int

const (
	// FILE represents a file
	FILE FileType = 1 + iota
	// DIRECTORY represents a folder
	DIRECTORY FileType = 1 + iota
)

// FileDef contains information about a file
type FileDef struct {
	Name        string
	Path        string
	Kind        FileType
	Size        int64
	Writable    bool
	Owner       FileProvider
	Extension   string
	Description string
}
