package files

import (
	"errors"
	"strings"

	//"paleotronic.com/fmt"
	"paleotronic.com/filerecord"
)

type PackageFileProvider struct {
	FileProvider
	cwd       string
	writeable bool
	basedir   string
	pattern   string
	priority  int
	pack      map[string]*Package
}

func NewPackageFileProvider(basedir string, priority int) *PackageFileProvider {
	pfp := &PackageFileProvider{basedir: basedir, writeable: false, priority: priority, cwd: ""}
	return pfp
}

func (pfp *PackageFileProvider) IsVisible() bool {
	return true
}

func (pfp *PackageFileProvider) IsReadOnly() bool {
	return !pfp.writeable
}

// DirFromBase gives a list of matching files under path
func (pfp *PackageFileProvider) DirFromBase(p string, filespec string) ([]FileDef, []FileDef, error) {
	if pfp.Exists(p, "") {
		ocwd := pfp.cwd
		pfp.cwd = p

		d, f, err := pfp.Dir(filespec)
		pfp.cwd = ocwd
		return d, f, err
	}

	return []FileDef(nil), []FileDef(nil), nil
}

func (pfp *PackageFileProvider) Dir(filespec string) ([]FileDef, []FileDef, error) {

	if filespec == "" {
		filespec = "*.*"
	}

	fqpn := GetUserPath(pfp.basedir, strings.Split(pfp.cwd, string('/')))

	dlist := make([]FileDef, 0)
	flist := make([]FileDef, 0)

	if pfp.cwd == "" || pfp.cwd == "/" {

		for n, _ := range pfp.pack {

			f := FileDef{}
			f.Kind = DIRECTORY
			f.Name = n
			f.Path = fqpn + string('/') + n
			f.Size = 0
			f.Owner = pfp

			dlist = append(dlist, f)
		}

	} else {
		p, exists := pfp.pack[pfp.cwd]

		if exists {
			_, f, e := p.Dir(filespec)
			if e == nil {
				flist = append(flist, f...)
			}
		}
	}

	return dlist, flist, nil
}

func (pfp *PackageFileProvider) ChDir(p string) error {
	if p == ".." {
		if pfp.cwd == "/" || pfp.cwd == "" {
			return errors.New(FPAccess)
		}
		parts := strings.Split(pfp.cwd, string('/'))
		if len(parts) == 1 {
			pfp.cwd = ""
			return nil
		} else {
			newpath := strings.Join(parts[0:len(parts)-2], string('/'))
			pfp.cwd = newpath
		}
	} else if p[0] == '/' {
		return pfp.SetCurrentPath(p, false)
	}
	return nil
}

func (pfp *PackageFileProvider) Exists(p string, f string) bool {

	if (p == "/" || p == "") && f == "" {
		return true
	} else if (p != "/" && p != "") && f == "" {
		return (pfp.pack[p] != nil)
	} else if (p != "/" && p != "") && f == "" {
		p, exists := pfp.pack[p]
		if !exists {
			return false
		}
		//
		i := p.IndexOf(f)
		if i > -1 {
			return true
		}
	}

	return false

}

func (pfp *PackageFileProvider) GetFileContent(p string, f string) (filerecord.FileRecord, error) {

	if p == "" {
		p = pfp.cwd
	}

	if (p != "/" && p != "") && f == "" {
		p, exists := pfp.pack[p]
		if !exists {
			return filerecord.FileRecord{}, errors.New("FILE NOT FOUND")
		}
		//
		i := p.IndexOf(f)
		if i > -1 {
			return filerecord.FileRecord{Content: p.Content[i].Data}, nil
		}
	}

	return filerecord.FileRecord{}, errors.New("FILE NOT FOUND")
}

func (pfp *PackageFileProvider) SetFileContent(p string, f string, data []byte) error {
	return errors.New(FPAccess)
}

// GetPriority returns the priority of this filesystem
func (pfp *PackageFileProvider) GetPriority() int {
	return pfp.priority
}

// GetCurrentPath returns the current working dir of this path
func (pfp *PackageFileProvider) GetCurrentPath() string {
	return pfp.cwd
}

// SetCurrentPath sets the current path, creating it if specified
func (pfp *PackageFileProvider) SetCurrentPath(p string, create bool) error {

	//	fqpn := GetRelativePath(pfp.basedir, strings.Split(p, string('/')))

	////fmt.Printntf("InternalFileProvider.SetCurrentPath() Checking is %s exists or not\n", fqpn)

	_, exists := pfp.pack[p]

	if !exists {
		return errors.New(FPNotExist)
	}

	////fmt.Println("OK")

	pfp.cwd = p
	return nil

}
