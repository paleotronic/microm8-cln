package files

import (
	"errors"

	"paleotronic.com/fmt"
	//	"os"
	"regexp"
	"strings"

	"paleotronic.com/filerecord"
	"paleotronic.com/octalyzer/assets"
)

type AssetRetriever func(s string) ([]byte, error)

// InternalFileProvider is a provider of local files based off the local filesystem
type InternalFileProvider struct {
	FileProvider
	cwd       string
	writeable bool
	basedir   string
	priority  int
	Visible   bool
}

func (lfp *InternalFileProvider) IsVisible() bool {
	return lfp.Visible
}

// GetPriority returns the priority of this filesystem
func (lfp *InternalFileProvider) GetPriority() int {
	return lfp.priority
}

// GetCurrentPath returns the current working dir of this path
func (lfp *InternalFileProvider) GetCurrentPath() string {
	return lfp.cwd
}

// SetCurrentPath sets the current path, creating it if specified
func (lfp *InternalFileProvider) SetCurrentPath(p string, create bool) error {

	fqpn := GetRelativePath(lfp.basedir, strings.Split(p, string('/')))

	////fmt.Printntf("InternalFileProvider.SetCurrentPath() Checking is %s exists or not\n", fqpn)

	_, err := assets.AssetDir(fqpn)

	if err != nil {
		return errors.New(FPNotExist)
	}

	////fmt.Printntln("OK")

	lfp.cwd = p
	return nil

}

// IsReadOnly returns true if the user cannot create directories and files
func (lfp *InternalFileProvider) IsReadOnly() bool {
	return !lfp.writeable
}

// GetFileContent returns the content of filename "f" at path "p"
func (lfp *InternalFileProvider) GetFileContent(p string, f string) (filerecord.FileRecord, error) {

	if p == "" {
		p = lfp.cwd
	}

	fqpn := GetRelativePath(lfp.basedir, append(strings.Split(p, string('/')), f))

	d, err := assets.Asset(fqpn)

	return filerecord.FileRecord{Content: d}, err
}

// SetFileContent writes a file with the current specified content
func (lfp *InternalFileProvider) SetFileContent(p string, f string, data []byte) error {
	return errors.New(FPAccess)
}

// NewInternalFileProvider creates a new instance of a InternalFileProvider
func NewInternalFileProvider(basedir string, priority int) *InternalFileProvider {
	lfp := &InternalFileProvider{basedir: basedir, writeable: false, priority: priority, cwd: ""}
	return lfp
}

// DirFromBase gives a list of matching files under path
func (lfp *InternalFileProvider) DirFromBase(p string, filespec string) ([]FileDef, []FileDef, error) {
	if lfp.Exists(p, "") {
		ocwd := lfp.cwd
		lfp.cwd = p

		d, f, err := lfp.Dir(filespec)
		lfp.cwd = ocwd
		return d, f, err
	}

	return []FileDef(nil), []FileDef(nil), nil
}

// Dir returns a list of files in the current directory
func (lfp *InternalFileProvider) Dir(filespec string) ([]FileDef, []FileDef, error) {

	if filespec == "" {
		filespec = "*.*"
	}

	regstr := strings.Replace(filespec, ".", "[.]", -1)
	regstr = strings.Replace(regstr, "*", ".*", -1)

	r := regexp.MustCompile(regstr)

	rs := regexp.MustCompile("^[0-9]+[ ]+REM[ ]+(.*)[\r\n]")

	fqpn := GetRelativePath(lfp.basedir, strings.Split(lfp.cwd, string('/')))
	if fqpn[len(fqpn)-1] == '/' {
		fqpn = fqpn[0 : len(fqpn)-1]
	}

	////fmt.Printntf("Internal Dir : [%s]\n", fqpn)

	info, err := assets.AssetDir(fqpn)

	var dlist = make([]FileDef, 0)
	var flist = make([]FileDef, 0)

	if lfp.cwd != "" && lfp.cwd != "\\" && lfp.cwd != "/" {
		dlist = append(dlist, FileDef{Name: "..", Size: 0, Kind: DIRECTORY, Path: lfp.cwd, Owner: lfp})
	}

	if err != nil {
		////fmt.Printntln(err)
		return dlist, flist, errors.New(FPIOError)
	}

	for _, iname := range info {

		////fmt.Printntln(iname)

		tp := GetRelativePath(fqpn, []string{iname})
		if tp[len(tp)-1] == '/' {
			tp = tp[0 : len(tp)-1]
		}

		var dir = false
		i, err := assets.AssetInfo(tp)

		if err != nil {
			// could be a dir?
			_, err = assets.AssetDir(fqpn)

			if err != nil {
				continue
			}

			// is a dir
			dir = true
		}

		f := FileDef{}
		if dir {

			if iname == "." || iname == ".." {
				continue
			}
			f.Kind = DIRECTORY
			f.Name = iname
			f.Path = fqpn + string('/') + iname
			f.Size = 0
			f.Owner = lfp

			dlist = append(dlist, f)
		} else {

			parts := strings.Split(i.Name(), "/")
			iname := parts[len(parts)-1]

			if !r.MatchString(i.Name()) {
				continue
			}
			f.Kind = FILE
			f.Name = iname
			f.Path = fqpn + string('/') + iname
			f.Size = i.Size()
			f.Owner = lfp
			f.Extension = GetExt(f.Name)
			z := len(f.Extension) + 1
			f.Name = f.Name[0 : len(f.Name)-z]
			f.Description = ""

			if (f.Extension == "i") || (f.Extension == "a") {

				data, e := assets.Asset(fqpn + "/" + iname)

				stmp := string(data)

				if e == nil && len(stmp) > 0 {
					pp := rs.FindAllStringSubmatch(stmp, -1)

					if len(pp) > 0 {
						f.Description = pp[0][1]
					}
				} else {
					////fmt.Printntln(e)
				}
			}

			flist = append(flist, f)
		}

	}

	return dlist, flist, nil

}

// ChDir() changes DIRECTORY
func (lfp *InternalFileProvider) ChDir(p string) error {
	if p == ".." {
		if lfp.cwd == "/" || lfp.cwd == "" || lfp.cwd == "\\" {
			return errors.New(FPAccess)
		}
		parts := strings.Split(lfp.cwd, string('/'))
		if len(parts) == 1 {
			lfp.cwd = ""
			return nil
		} else {
			newpath := strings.Join(parts[0:len(parts)-2], string('/'))
			lfp.cwd = newpath
		}
	} else if p[0] == '/' {
		return lfp.SetCurrentPath(p, false)
	}
	return nil
}

func (lfp *InternalFileProvider) Exists(p string, f string) bool {
	fqpn := GetRelativePath(lfp.basedir, append(strings.Split(p, string('/')), f))

	if f == "" {
		for fqpn[len(fqpn)-1] == '/' {
			fqpn = fqpn[0 : len(fqpn)-1]
		}
	}

	fmt.Printf("Internal asset? [%s]\n", fqpn)

	_, e := assets.AssetInfo(fqpn)

	if e != nil {
		// could be a dir?
		_, e = assets.AssetDir(fqpn)
	}

	fmt.Println((e == nil))

	return (e == nil)
}
