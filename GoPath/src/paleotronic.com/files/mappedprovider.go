package files

import (
	"errors"
	"reflect"

	"paleotronic.com/fmt"
	//"paleotronic.com/fmt"
	"strings"

	"paleotronic.com/log"

	"paleotronic.com/filerecord"
)

type ProviderHolder struct {
	Provider FileProvider
	BasePath string
}

type MapBuilder func() map[string]ProviderHolder

type MappedFileProvider struct {
	//	FileProvider
	cwd        string
	writeable  bool
	basedir    string
	priority   int
	shared     bool
	mappings   map[string]ProviderHolder // virtual mount point
	mapbuilder MapBuilder
}

func (mfp *MappedFileProvider) CheckMap() {

	if mfp.mapbuilder != nil {

		mfp.mappings = mfp.mapbuilder()

		log.Println(mfp.mappings)

	}

}

func NewMappedFileProvider(mappings map[string]ProviderHolder) *MappedFileProvider {
	this := &MappedFileProvider{}
	if mappings != nil {
		this.mappings = mappings
	} else {
		this.mappings = make(map[string]ProviderHolder)
	}
	return this
}

func NewMappedFileProviderFunc(mapfunc MapBuilder) *MappedFileProvider {
	this := &MappedFileProvider{}
	this.mapbuilder = mapfunc
	this.mappings = make(map[string]ProviderHolder)
	return this
}

func (mfp *MappedFileProvider) Add(basepath string, actualpath string, fp FileProvider) {
	mfp.mappings[basepath] = ProviderHolder{Provider: fp, BasePath: actualpath}
}

func (mfp *MappedFileProvider) Remove(basepath string) {
	delete(mfp.mappings, basepath)
}

func (mfp *MappedFileProvider) ResolveProviderPathFile(p, f string) (FileProvider, string, string) {

	p = strings.Trim(p, "/ ")
	f = strings.Trim(f, "/ ")

	if exists, np, fp := hasOverlay(p); exists {
		fmt.Printf("Over exists for path %s/%s\n", p, f)
		return fp, np, f
	}

	fmt.Printf("Resolving path=[%s], file=[%s]\n", p, f)

	if p != "" {

		//
		parts := strings.Split(p, "/")

		fmt.Printf("parts=%s\n", strings.Join(parts, ","))
		fmt.Printf("mappings=%v\n", mfp.mappings)

		if m, ok := mfp.mappings[strings.ToLower(parts[0])]; ok {
			return m.Provider, strings.Join(parts[1:], "/"), f
		} else {
			// not defined
			if m, ok := mfp.mappings[""]; ok {
				return m.Provider, p, f
			} else {
				return nil, p, f
			}
		}
	} else {
		// p == ""
		if m, ok := mfp.mappings[strings.ToLower(f)]; ok {
			return m.Provider, "", ""
		} else {
			if m, ok := mfp.mappings[""]; ok {
				return m.Provider, p, f
			} else {
				return nil, p, f
			}
		}
	}

	return nil, p, f
}

func (mfp *MappedFileProvider) Resolve(basepath string) (FileProvider, string) {
	fp, ok := mfp.mappings[strings.ToLower(basepath)]
	if ok {
		return fp.Provider, fp.BasePath
	}

	// mapping for ""
	fp, ok = mfp.mappings[""]
	if ok {
		return fp.Provider, basepath
	}

	return nil, ""
}

func (mfp *MappedFileProvider) ChDir(p string) error {
	mfp.cwd = p
	return nil
}

func (mfp *MappedFileProvider) Dir(filespec string) ([]FileDef, []FileDef, error) {

	if mfp.cwd == "/" {
		mfp.cwd = ""
	}

	mfp.cwd = strings.Replace(mfp.cwd, "//", "/", -1)

	if mfp.cwd == "" {
		mfp.CheckMap()
	}

	fmt.Printf("MappedFileProvider::Dir(%s), cwd = %s\n", filespec, mfp.cwd)

	p := mfp.cwd
	f := GetFilename(mfp.cwd)

	fp, dp, df := mfp.ResolveProviderPathFile(p, "")

	//	fmt.Printf("MappedFileProvider::Dir(%s), fp = %v, p=%s, f=%s\n", filespec, fp, p, f)

	if fp == nil {

		log.Printf("===> NIL resolve of %s, full cwd = %s\n", p, mfp.cwd)

		dd := make([]FileDef, 0)

		for k, _ := range mfp.mappings {
			if strings.Trim(k, " ") != "" && p == "" {
				dd = append(dd, FileDef{Name: k, Size: 0, Kind: DIRECTORY, Path: mfp.cwd + "/" + k, Owner: mfp, Description: "<dir>"})
			}
		}

		return dd, []FileDef(nil), nil
	}

	p = dp
	f = df

	// Do dir and return
	if p == "/" {
		p = ""
	}

	if p == "" && f != "" {
		p = f
		f = ""
	}

	fp.SetCurrentPath(dp, false)

	log.Printf("=====> %s/%s\n", p, filespec)
	dd, ff, ee := fp.DirFromBase(p, filespec)

	if mfp.cwd != "" && mfp.cwd != "/" {
		dd = append(dd, FileDef{Name: "..", Size: 0, Kind: DIRECTORY, Path: mfp.cwd + "/..", Owner: mfp, Description: "<dir>"})
	}

	// Add in level 0 keys for mounted systems

	for k, _ := range mfp.mappings {
		if strings.Trim(k, " ") != "" && (mfp.cwd == "" || mfp.cwd == "/") {
			dd = append(dd, FileDef{Name: k, Size: 0, Kind: DIRECTORY, Path: mfp.cwd + "/" + k, Owner: mfp, Description: "<dir>"})
		}
	}
	return dd, ff, ee
}

func (mfp *MappedFileProvider) DirFromBase(p string, filespec string) ([]FileDef, []FileDef, error) {

	fmt.Printf("mfp.DirFromBase(%s)\n", p)

	if mfp.Exists(p, "") {
		ocwd := mfp.cwd
		mfp.cwd = p

		d, f, err := mfp.Dir(filespec)
		mfp.cwd = ocwd

		fmt.Println("Yes")

		return d, f, err
	} else {
		fmt.Println("No")
	}

	return []FileDef(nil), []FileDef(nil), nil
}

func (mfp *MappedFileProvider) Exists(p string, f string) bool {

	fmt.Printf("MFP::Exists(%s, %s)\n", p, f)

	fp, p, f := mfp.ResolveProviderPathFile(p, f)

	if fp != nil {
		fmt.Printf("MappedFileProvider: fp=%s, p=%v, f=%v\n", reflect.TypeOf(fp).Elem().Name(), p, f)
		return fp.Exists(p, f)
	} else {
		if p == "" && f == "" {
			return true
		}
		return false
	}
}

func (mfp *MappedFileProvider) MkDir(p string, f string) error {

	fp, p, f := mfp.ResolveProviderPathFile(p, f)

	if fp == nil {
		return errors.New("File Not Found")
	}

	return fp.MkDir(p, f)
}

func (mfp *MappedFileProvider) Delete(p string, f string) error {

	fp, p, f := mfp.ResolveProviderPathFile(p, f)

	if fp == nil {
		return errors.New("File Not Found")
	}

	return fp.Delete(p, f)
}

func (mfp *MappedFileProvider) Rename(p string, f string, nf string) error {

	fp, p, f := mfp.ResolveProviderPathFile(p, f)

	if fp == nil {
		return errors.New("File Not Found")
	}

	return fp.Rename(p, f, nf)

}

func (mfp *MappedFileProvider) Lock(p string, f string) error {

	fp, p, f := mfp.ResolveProviderPathFile(p, f)

	if fp == nil {
		return errors.New("File Not Found")
	}

	return fp.Lock(p, f)
}

func (mfp *MappedFileProvider) Meta(p string, f string, meta map[string]string) error {

	fp, p, f := mfp.ResolveProviderPathFile(p, f)

	if fp == nil {
		return errors.New("File Not Found")
	}

	return fp.Meta(p, f, meta)
}

func (mfp *MappedFileProvider) Share(p string, f string) (string, string, bool, error) {

	fp, p, f := mfp.ResolveProviderPathFile(p, f)

	if fp == nil {
		return "", "", false, errors.New("File Not Found")
	}

	return fp.Share(p, f)
}

func (mfp *MappedFileProvider) IsVisible() bool {
	return true
}

func (mfp *MappedFileProvider) IsReadOnly() bool {
	return false
}

// GetPriority returns the priority of this filesystem
func (mfp *MappedFileProvider) GetPriority() int {
	return mfp.priority
}

// GetCurrentPath returns the current working dir of this path
func (mfp *MappedFileProvider) GetCurrentPath() string {
	return mfp.cwd
}

// SetCurrentPath sets the current path, creating it if specified
func (mfp *MappedFileProvider) SetCurrentPath(p string, create bool) error {

	mfp.cwd = p

	return nil

}

func (mfp *MappedFileProvider) SetLoadAddress(p string, f string, address int) error {

	if p == "" {
		p = mfp.cwd
	}

	fp, p, f := mfp.ResolveProviderPathFile(p, f)

	//	fmt.Printf("MappedFileProvider::GetFileContent() fp=%v, p=%v, f=%v\n", fp, p, f)

	if fp == nil {
		return errors.New(FPNotExist)
	}

	e := fp.SetLoadAddress(p, f, address)

	return e
}

// GetFileContent returns the content of filename "f" at path "p"
func (mfp *MappedFileProvider) GetFileContent(p string, f string) (filerecord.FileRecord, error) {

	if p == "" {
		p = mfp.cwd
	}

	fp, p, f := mfp.ResolveProviderPathFile(p, f)

	//	fmt.Printf("MappedFileProvider::GetFileContent() fp=%v, p=%v, f=%v\n", fp, p, f)

	if fp == nil {
		return filerecord.FileRecord{}, errors.New(FPNotExist)
	}

	b, e := fp.GetFileContent(p, f)

	return b, e
}

// GetFileContent returns the content of filename "f" at path "p"
func (mfp *MappedFileProvider) Validate(current *filerecord.FileRecord) (*filerecord.FileRecord, error) {

	p := current.FilePath
	f := current.FileName

	if p == "" {
		p = mfp.cwd
	}

	fp, p, f := mfp.ResolveProviderPathFile(p, f)

	//	fmt.Printf("MappedFileProvider::GetFileContent() fp=%v, p=%v, f=%v\n", fp, p, f)

	if fp == nil {
		return &filerecord.FileRecord{}, errors.New(FPNotExist)
	}

	current.FileName = f
	current.FilePath = p

	b, e := fp.Validate(current)

	return b, e
}

func (mfp *MappedFileProvider) SetFileContent(p string, f string, data []byte) error {

	if p == "" {
		p = mfp.cwd
	}

	log.Printf("In SetFileContent with p = [%s], f = [%s]\n", p, f)

	fp, p, f := mfp.ResolveProviderPathFile(p, f)

	if fp == nil {
		return errors.New("No handler for path [" + p + "]")
	}

	return fp.SetFileContent(p, f, data)

}
