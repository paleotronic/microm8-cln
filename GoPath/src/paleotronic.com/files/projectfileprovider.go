package files

// ProjectFileProvider holds mappings for the project space.
type ProjectFileProvider struct {
	//	FileProvider
	cwd       string
	writeable bool
	basedir   string
	priority  int
	shared    bool
    projects  []string
}

func (lfp *ProjectFileProvider) IsVisible() bool {
	return true
}

// GetPriority returns the priority of this filesystem
func (lfp *ProjectFileProvider) GetPriority() int {
	return lfp.priority
}

// GetCurrentPath returns the current working dir of this path
func (lfp *ProjectFileProvider) GetCurrentPath() string {
	return lfp.cwd
}

// SetCurrentPath sets the current path, creating it if specified
func (lfp *ProjectFileProvider) SetCurrentPath(p string, create bool) error {

	lfp.cwd = p

	return nil

}



