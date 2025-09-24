package files

import (
	"errors"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"paleotronic.com/filerecord"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
	"paleotronic.com/utils"
)

// LocalFileProvider is a provider of local files based off the local filesystem
type LocalFileProvider struct {
	FileProvider
	cwd       string
	writeable bool
	basedir   string
	priority  int
}

func (lfp *LocalFileProvider) IsVisible() bool {
	return true
}

// GetPriority returns the priority of this filesystem
func (lfp *LocalFileProvider) GetPriority() int {
	return lfp.priority
}

// GetCurrentPath returns the current working dir of this path
func (lfp *LocalFileProvider) GetCurrentPath() string {
	return lfp.cwd
}

// SetCurrentPath sets the current path, creating it if specified
func (lfp *LocalFileProvider) SetCurrentPath(p string, create bool) error {

	fqpn := GetExpandedDir(lfp.basedir, strings.Split(p, "/"))

	////fmt.Printntf("LocalFileProvider.SetCurrentPath() Checking is %s exists or not\n", fqpn)

	ok, _ := CheckPathCI(fqpn)

	if !ok {
		if create && !lfp.IsReadOnly() {
			os.MkdirAll(fqpn, 0755)
		} else {
			return errors.New(FPNotExist)
		}
	}

	////fmt.Printntln("OK")

	lfp.cwd = p
	return nil

}

// IsReadOnly returns true if the user cannot create directories and files
func (lfp *LocalFileProvider) IsReadOnly() bool {
	return !lfp.writeable
}

func filt(fr filerecord.FileRecord) filerecord.FileRecord {

	pattern := "(?i)^(.+)#(0x|[$])([a-f0-9]+)[.]([a-z]+)$"
	r := regexp.MustCompile(pattern)
	if r.MatchString(fr.FileName) {
		m := r.FindAllStringSubmatch(fr.FileName, -1)

		a, _ := strconv.ParseInt(m[0][3], 16, 32)
		addr := int(a)

		desc := fmt.Sprintf("%s", fr.FileName)

		if fr.Address == 0 {
			fr.Address = addr
		}
		fr.Description = desc
		fr.FileName = m[0][1] + "." + m[0][4]
	}

	// if strings.HasSuffix(fr.FileName, ".pak") {
	// 	fp, err := NewOctContainer(&fr, fr.FilePath+"/"+fr.FileName)
	// 	if err == nil {
	// 		fr.Description = fp.GetDescription()
	// 	}
	// }

	return fr

}

// GetFileContent returns the content of filename "f" at path "p"
func (lfp *LocalFileProvider) GetFileContent(p string, f string) (filerecord.FileRecord, error) {

	fmt.Printf("LocalFileProvider::GetFileContent() -> p=%s, f=%s\n", p, f)

	fqpn := GetExpandedDir(lfp.basedir, append(strings.Split(p, string('/')), f))

	ok, nfqpn := CheckPathCI(fqpn)

	fmt.Printf("ok=%v, nfqpn=%s\n", ok, nfqpn)

	if !ok {
		return filerecord.FileRecord{}, errors.New(FPNotExist)
	}

	fqpn = nfqpn

	fmt.Printf("local read of %s\n", nfqpn)

	f = GetFilename(fqpn)

	b, e := ReadBytes(fqpn)
	return filt(filerecord.FileRecord{
			Content:      b,
			FileName:     f,
			FilePath:     p,
			Description:  "",
			UserCanWrite: true,
		}),
		e
}

// GetFileContent returns the content of filename "f" at path "p"
func (lfp *LocalFileProvider) Validate(current *filerecord.FileRecord) (*filerecord.FileRecord, error) {

	fr, err := lfp.GetFileContent(current.FilePath, current.FileName)

	return &fr, err

}

// SetFileContent writes a file with the current specified content
func (lfp *LocalFileProvider) SetFileContent(p string, f string, data []byte) error {

	fmt.Printf("LFP.SetFileContent(%s, %s, %d)\n", p, f, len(data))

	if lfp.IsReadOnly() {
		return errors.New(FPAccess)
	}

	//	if p == "" {
	//		p = lfp.cwd
	//	}

	// write file
	fqpn := GetExpandedDir(lfp.basedir, append(strings.Split(p, string('/')), f))
	path := "/" + GetPath(fqpn)

	ex, nfqpn := CheckPathCI(path + "/" + f)
	fmt.Printf("path=%s, ex=%v, nfqpn=%s\n", path, ex, nfqpn)
	if ex {
		fqpn = nfqpn
	} else {
		ex, nfqpn := CheckPathCI(path)
		if ex {
			fqpn = nfqpn + "/" + f
		}
	}

	fmt.Printf("############### WRITEL [%s]\n", fqpn)

	e := WriteBytes(fqpn, data, false)
	return e
	// if e != nil {
	// 	return e
	// }

	// return lfp.SetLoadAddress(p, f, 0x1122)
}

// SetFileContent writes a file with the current specified content
func (lfp *LocalFileProvider) Delete(p string, f string) error {
	if lfp.IsReadOnly() {
		return errors.New(FPAccess)
	}

	// write file
	fqpn := GetExpandedDir(lfp.basedir, append(strings.Split(p, string('/')), f))

	ex, nfqpn := CheckPathCI(fqpn)
	if ex {

		st, err := os.Stat(nfqpn)
		if err != nil {
			return err
		}
		if st.IsDir() {
			entries, err := getFilenames(nfqpn)
			if err != nil {
				return err
			}

			if len(entries) > 0 {
				return errors.New("dir not empty")
			}
		}

		return os.RemoveAll(nfqpn)
	}

	return errors.New("File not found")
}

// Lock method does nothing on a local filesystem
func (lfp *LocalFileProvider) Lock(p string, f string) error {
	return nil
}

// NewLocalFileProvider creates a new instance of a LocalFileProvider
func NewLocalFileProvider(basedir string, writeable bool, priority int) *LocalFileProvider {
	lfp := &LocalFileProvider{basedir: basedir, writeable: writeable, priority: priority, cwd: ""}
	return lfp
}

// DirFromBase gives a list of matching files under path
func (lfp *LocalFileProvider) DirFromBase(p string, filespec string) ([]FileDef, []FileDef, error) {
	if lfp.Exists(p, "") {
		fmt.Println("OK")
		ocwd := lfp.cwd
		lfp.cwd = p

		d, f, err := lfp.Dir(filespec)
		lfp.cwd = ocwd
		return d, f, err
	}

	return []FileDef(nil), []FileDef(nil), nil
}

func (lfp *LocalFileProvider) SetLoadAddress(p string, f string, address int) error {

	fqpn := GetExpandedDir(lfp.basedir, append(strings.Split(p, string('/')), f))
	ee, wp := CheckPathCI(fqpn)

	if !ee {
		return errors.New("file not found")
	}

	// wp is the real file path
	fr, err := lfp.GetFileContent(p, f)
	if err != nil {
		return nil
	}

	ext := GetExt(fr.FileName)
	base := fr.FileName[0 : len(fr.FileName)-len(ext)-1]

	nf := fmt.Sprintf("%s#0x%.4x.%s", base, address, ext)
	wpp := GetPath(wp)
	if runtime.GOOS != "windows" {
		wpp = "/" + strings.Trim(wpp, "/")
	}

	ff, err := os.Create(wpp + "/" + nf)
	if err != nil {
		return err
	}
	ff.Write(fr.Content)
	ff.Close()

	return os.RemoveAll(wp)
}

// Dir returns a list of files in the current directory
func (lfp *LocalFileProvider) Dir(filespec string) ([]FileDef, []FileDef, error) {

	if filespec == "" {
		filespec = "*.*"
	}

	regstr := strings.Replace(filespec, ".", "[.]", -1)
	regstr = strings.Replace(regstr, "*", ".*", -1)

	r := regexp.MustCompile("(?i)" + regstr)

	rs := regexp.MustCompile("^[0-9]+[ ]+REM[ ]+(.*)")

	fqpn := GetExpandedDir(lfp.basedir, strings.Split(lfp.cwd, string('/')))

	_, fqpn = CheckPathCI(fqpn)

	fmt.Printf("LocalFileProvider: %s for %s\n", fqpn, filespec)

	info, err := getFilenames(fqpn)

	var dlist = make([]FileDef, 0)
	var flist = make([]FileDef, 0)

	if lfp.cwd != "" && lfp.cwd != "\\" && lfp.cwd != "/" {
		dlist = append(dlist, FileDef{Name: "..", Size: 0, Kind: DIRECTORY, Path: lfp.cwd, Owner: lfp})
	}

	if err != nil {
		return dlist, flist, errors.New(FPIOError)
	}

	apattern := "(?i)^(.+)#(0x|[$])([a-f0-9]+)$"
	ra := regexp.MustCompile(apattern)

	for _, i := range info {
		s, err := os.Stat(fqpn + "/" + i)
		if err != nil {
			continue
		}
		f := FileDef{}
		if s.IsDir() {
			if s.Name() == "." || s.Name() == ".." {
				continue
			}
			f.Kind = DIRECTORY
			f.Name = s.Name()
			f.Path = fqpn + string('/') + s.Name()
			f.Size = 0
			f.Owner = lfp
			f.Description = "<dir>"

			//f.Extension = GetExt(i.Name())
			//z := len(f.Extension) + 1
			//f.Name = f.Name[0 : len(f.Name)-z]

			dlist = append(dlist, f)
		} else {
			if strings.HasPrefix(s.Name(), ".") {
				continue
			}
			if !r.MatchString(s.Name()) {
				continue
			}
			f.Kind = FILE
			f.Name = s.Name()
			f.Path = fqpn + string('/') + s.Name()
			f.Size = s.Size()
			f.Owner = lfp
			f.Extension = GetExt(f.Name)
			z := len(f.Extension) + 1
			f.Name = f.Name[0 : len(f.Name)-z]
			f.Description = ""

			if (f.Extension == "i") || (f.Extension == "a") {
				sl, e := utils.ReadTextFile(fqpn + string('/') + s.Name())
				if e == nil && len(sl) > 0 {
					pp := rs.FindAllStringSubmatch(sl[0], -1)

					if len(pp) > 0 {
						f.Description = pp[0][1]
					}
				}
			}

			if ra.MatchString(f.Name) {
				m := ra.FindAllStringSubmatch(f.Name, -1)
				f.Name = m[0][1]
			}

			// if f.Extension == "pak" {
			// 	fmt.Printf("%s is a pakfile\n", f.Name)
			// 	data, err := lfp.GetFileContent(lfp.cwd, f.Name+".pak")
			// 	if err == nil {
			// 		fmt.Println("read pakfile")
			// 		zp, err := NewOctContainer(&data, lfp.cwd+"/"+f.Name+".pak")
			// 		if err == nil {
			// 			fmt.Println("get extended description")
			// 			f.Description = zp.GetDescription()
			// 		}
			// 	}
			// }

			flist = append(flist, f)
		}

	}

	return dlist, flist, nil

}

// ChDir() changes DIRECTORY
func (lfp *LocalFileProvider) ChDir(p string) error {
	if p == ".." {
		if lfp.cwd == "/" || lfp.cwd == "" {
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

func (lfp *LocalFileProvider) Rename(p string, f string, nf string) error {

	fqpn := GetExpandedDir(lfp.basedir, append(strings.Split(p, string('/')), f))

	fqpn = strings.Replace(fqpn, "//", "/", -1)

	ee, wp := CheckPathCI(fqpn)

	if !ee {
		return errors.New("file not found")
	}

	fqpn = wp

	nfqpn := GetPath(fqpn) + "/" + nf

	if !strings.HasPrefix(nfqpn, "/") {
		nfqpn = "/" + nfqpn
	}

	log.Printf("Renaming %s to %s\n", fqpn, nfqpn)

	return os.Rename(fqpn, nfqpn)

}

func (lfp *LocalFileProvider) Exists(p string, f string) bool {
	fqpn := GetExpandedDir(lfp.basedir, append(strings.Split(p, string('/')), f))

	fqpn = strings.Replace(fqpn, "//", "/", -1)

	//ok, m := ExistsCP(fqpn)
	//fmt.Printf("Matches found: %v, %v\n", ok, m)

	fmt.Println("lfp check existence for", fqpn)

	ee, wp := CheckPathCI(fqpn)

	fmt.Printf("LFP::Exists() -> %v (as %s)\n", ee, wp)

	return ee
}

func (lfp *LocalFileProvider) MkDir(p string, f string) error {

	var lock error = nil

	//log.Printf("LocalFileProvider.MkDir(%s, %s)\n", p, f)

	if f != "" {
		fqpn := GetExpandedDir(lfp.basedir, append(strings.Split(p, string('/')), f))
		oparts := strings.Split(strings.Trim(fqpn, "/"), "/")
		//rlog.Printf("initial fqpn = %s", fqpn)

		_, fqpn = CheckPathCI(fqpn)

		nparts := strings.Split(strings.Trim(fqpn, "/"), "/")
		for len(nparts) < len(oparts) {
			nparts = append(nparts, oparts[len(nparts)])
		}
		fqpn = strings.Join(nparts, "/")
		if runtime.GOOS != "windows" {
			fqpn = "/" + fqpn
		}

		//rlog.Printf("Will make directory as follows: %s", fqpn)

		lock = os.MkdirAll(fqpn, 0755)

	}

	return lock
}
