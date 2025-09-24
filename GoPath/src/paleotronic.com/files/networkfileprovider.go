package files

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"paleotronic.com/api"
	"paleotronic.com/filerecord"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
	"paleotronic.com/utils"
)

// NetworkUserFileProvider is a provider of local files based off the local filesystem
type NetworkUserFileProvider struct {
	FileProvider
	cwd         string
	writeable   bool
	basedir     string
	priority    int
	shared      bool
	system      bool
	project     bool
	remint      bool
	editable    bool
	projectname string
	remintname  string
	mappings    map[string]FileProvider // virtual mount point
	c           *Cache
}

// Note only uses the basepath
func (lfp *NetworkUserFileProvider) GetMappedFS(filepath string) FileProvider {
	nfp, ok := lfp.mappings[GetPath(filepath)]
	if ok {
		return nfp
	}
	return lfp
}

func (lfp *NetworkUserFileProvider) IsVisible() bool {
	return true
}

// GetPriority returns the priority of this filesystem
func (lfp *NetworkUserFileProvider) GetPriority() int {
	return lfp.priority
}

// GetCurrentPath returns the current working dir of this path
func (lfp *NetworkUserFileProvider) GetCurrentPath() string {
	return lfp.cwd
}

// SetCurrentPath sets the current path, creating it if specified
func (lfp *NetworkUserFileProvider) SetCurrentPath(p string, create bool) error {

	lfp.cwd = p

	return nil

}

// IsReadOnly returns true if the user cannot create directories and files
func (lfp *NetworkUserFileProvider) IsReadOnly() bool {
	return !lfp.writeable
}

func (lfp *NetworkUserFileProvider) GetCacheContext() string {

	switch {
	case lfp.project:
		return "project/" + lfp.projectname
	case lfp.system:
		return "system"
	case lfp.editable:
		return "user"
	case lfp.remint:
		return "remote"
	case lfp.shared:
		return "legacy"
	}

	return "unknown"

}

func (lfp *NetworkUserFileProvider) C() *Cache {
	if lfp.c == nil {
		lfp.c = New(BASEDIR, "FILECACHE")
	}
	return lfp.c
}

// GetFileContent returns the content of filename "f" at path "p"
func (lfp *NetworkUserFileProvider) GetFileContent(p string, f string) (filerecord.FileRecord, error) {

	if lfp.basedir != "" {
		p = strings.Replace(lfp.basedir+"/"+p, "//", "/", -1)
		if strings.HasSuffix(p, "/") {
			p = p[0 : len(p)-1]
		}
	}

	if p == "" {
		p = lfp.cwd
	}

	if len(p) > 1 && rune(p[0]) == '/' {
		p = p[1:]
	}

	fmt.Printf("NFP looking for %s\n", p+"/"+f)

	cacheContext := lfp.GetCacheContext()
	current, err := lfp.C().Get(cacheContext, p+"/"+f)

	if err != nil {
		var b filerecord.FileRecord
		var e error
		if lfp.remint {
			b, e = s8webclient.CONN.FetchRemIntFile(lfp.remintname, p, f)
		} else if lfp.system {
			b, e = s8webclient.CONN.FetchSystemFile(p, f)
		} else if lfp.project {
			b, e = s8webclient.CONN.FetchProjectFile(lfp.projectname, p, f)
		} else if lfp.shared {
			b, e = s8webclient.CONN.FetchLegacyFile(p, f)
		} else {
			b, e = s8webclient.CONN.FetchUserFile(p, f)
		}

		if e == nil {
			lfp.C().Put(cacheContext, &b)
		}

		return filt(b), e
	} else {
		var b *filerecord.FileRecord
		var e error

		if time.Since(current.Modified) > 60*time.Second {

			if lfp.system {
				b, e = s8webclient.CONN.ValidateCacheSystemFile(current)
			} else if lfp.project {
				b, e = s8webclient.CONN.ValidateCacheProjectFile(lfp.projectname, current)
			} else if lfp.shared {
				b, e = s8webclient.CONN.ValidateCacheLegacyFile(current)
			} else {
				b, e = s8webclient.CONN.ValidateCacheUserFile(current)
			}

		} else {

			b = &filerecord.FileRecord{}

			*b = *current

		}

		if e == nil && b.Checksum != current.Checksum {
			lfp.C().Put(cacheContext, b)
			return *b, e
		} else if e == nil {
			// got something but let's recheck cache anyway
			log.Println("Cache match but recheck cache")
			current, err = lfp.C().Get(cacheContext, p+"/"+f)
			if err == nil && current.Checksum == b.Checksum {
				log.Println("On recheck cache identical")
			} else {
				log.Println("On recheck cache different")
				lfp.C().Put(cacheContext, b)
				return *b, errors.New("Possible cache tamper")
			}
		}

		pattern := "(?i)^(.+)#(0x|[$])([a-f0-9]+)[.]([a-z]+)$"
		r := regexp.MustCompile(pattern)
		if r.MatchString(current.FileName) && current.Address == 0 {
			m := r.FindAllStringSubmatch(f, -1)

			a, _ := strconv.ParseInt(m[0][3], 16, 32)
			current.Address = int(a)
		}

		return filt(*current), e
	}

}

func (mfp *NetworkUserFileProvider) zexists(p, f string) (bool, []string) {

	p = strings.Trim(p, "/")

	_, files, err := mfp.DirFromBase(p, f)
	if err != nil {
		return false, []string(nil)
	}

	if (p == "" || p == "/") && f == "" {
		return true, []string{""}
	}

	pattern := "(?i)^(.+)#(0x|[$])([a-f0-9]+)[.]([a-z0-9]+)$"
	r := regexp.MustCompile(pattern)

	for _, entry := range files {
		wname := entry.Name + "." + entry.Extension
		if strings.ToLower(wname) == strings.ToLower(f) {
			return true, []string{wname}
		} else if r.MatchString(wname) {
			ext := GetExt(f)
			base := f[0 : len(f)-len(ext)-1]
			m := r.FindAllStringSubmatch(wname, -1)
			eExt := m[0][4]
			eBase := m[0][1]
			if strings.ToLower(eBase) == strings.ToLower(base) && strings.ToLower(eExt) == strings.ToLower(ext) {
				return true, []string{wname}
			}
		}
	}

	return false, []string(nil)

}

func (lfp *NetworkUserFileProvider) SetLoadAddress(p string, f string, address int) error {

	var e error
	if lfp.shared && !lfp.writeable {
		e = s8webclient.CONN.SetLegacyLoadAddress(p, f, address)
	} else if lfp.system && lfp.writeable {
		e = s8webclient.CONN.SetSystemLoadAddress(p, f, address)
	} else if lfp.project && lfp.editable {
		e = s8webclient.CONN.SetProjectLoadAddress(lfp.projectname, p, f, address)
	} else {
		e = s8webclient.CONN.SetUserLoadAddress(p, f, address)
	}

	fmt.Printf("NFP::SetLoadAddress(%s, %s) -> %v\n", p, f, e)

	if e == nil {
		cacheContext := lfp.GetCacheContext()
		fr, e := lfp.C().Get(cacheContext, p+"/"+f)
		if e == nil {
			fr.Address = address
			lfp.C().Upsert(cacheContext, fr)
		}
	}

	return e

}

// SetFileContent writes a file with the current specified content
func (lfp *NetworkUserFileProvider) SetFileContent(p string, f string, data []byte) error {
	//if lfp.IsReadOnly() {
	//	return errors.New(FPAccess)
	//}

	if lfp.basedir != "" {
		p = strings.Replace(lfp.basedir+"/"+p, "//", "/", -1)
		if strings.HasSuffix(p, "/") {
			p = p[0 : len(p)-1]
		}
	}

	if p == "" {
		p = lfp.cwd
	}

	if len(p) > 1 && rune(p[0]) == '/' {
		p = p[1:]
	}

	cacheContext := lfp.GetCacheContext()

	var e error

	chunks := make([][]byte, 0)
	ptr := 0
	sz := 16384
	for ptr < len(data) {
		end := ptr + sz
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[ptr:end])
		ptr += sz
	}

	for i, block := range chunks {

		//fmt.Printf("Writing chunk %d of %s (%d bytes)\n", i+1, f, len(block))

		if i == 0 {
			if lfp.shared && !lfp.writeable {
				e = s8webclient.CONN.StoreLegacyFile(p, f, block)
			} else if lfp.system && lfp.writeable {
				e = s8webclient.CONN.StoreSystemFile(p, f, block)
			} else if lfp.project && lfp.editable {
				e = s8webclient.CONN.StoreProjectFile(lfp.projectname, p, f, block)
			} else {
				e = s8webclient.CONN.StoreUserFile(p, f, block)
			}
		} else {
			if lfp.shared && !lfp.writeable {
				e = s8webclient.CONN.AppendLegacyFile(p, f, block)
			} else if lfp.system && lfp.writeable {
				e = s8webclient.CONN.AppendSystemFile(p, f, block)
			} else if lfp.project && lfp.editable {
				e = s8webclient.CONN.AppendProjectFile(lfp.projectname, p, f, block)
			} else {
				e = s8webclient.CONN.AppendUserFile(p, f, block)
			}
		}

		if e != nil {
			break
		}

	}

	if e == nil {
		lfp.C().Upsert(
			cacheContext,
			&filerecord.FileRecord{
				FileName: f,
				FilePath: p,
				Content:  data,
			},
		)
	}

	return e

}

// NewNetworkUserFileProvider creates a new instance of a NetworkUserFileProvider
func NewNetworkUserFileProvider(basedir string, writeable bool, shared bool, priority int) *NetworkUserFileProvider {
	lfp := &NetworkUserFileProvider{basedir: basedir, writeable: writeable, shared: shared, priority: priority, cwd: ""}
	lfp.mappings = make(map[string]FileProvider)
	return lfp
}

func NewNetworkRemIntFileProvider(basedir string, rin string, writeable bool, shared bool, priority int) *NetworkUserFileProvider {
	lfp := &NetworkUserFileProvider{basedir: basedir, remintname: rin, writeable: writeable, shared: shared, remint: true, priority: priority, cwd: ""}
	lfp.mappings = make(map[string]FileProvider)
	return lfp
}

func NewNetworkSystemFileProvider(basedir string, writeable bool, shared bool, priority int) *NetworkUserFileProvider {
	lfp := &NetworkUserFileProvider{basedir: basedir, writeable: writeable, shared: shared, system: true, priority: priority, cwd: ""}
	lfp.mappings = make(map[string]FileProvider)
	return lfp
}

func NewProjectProvider(basedir string, writeable bool, shared bool, editable bool, projectname string, priority int) *NetworkUserFileProvider {
	lfp := &NetworkUserFileProvider{basedir: basedir, writeable: writeable, shared: shared, priority: priority, cwd: "", projectname: projectname, project: true, editable: editable}
	lfp.mappings = make(map[string]FileProvider)
	return lfp
}

// DirFromBase gives a list of matching files under path
func (lfp *NetworkUserFileProvider) DirFromBase(p string, filespec string) ([]FileDef, []FileDef, error) {

	fmt.Printf("Nfp.DirFromBase(%s, %s)\n", p, filespec)

	if lfp.basedir != "" {
		p = strings.Replace(lfp.basedir+"/"+p, "//", "/", -1)
		if strings.HasSuffix(p, "/") {
			p = p[0 : len(p)-1]
		}
	}

	if lfp.Exists(p, "") {

		fmt.Printf("Exists: %s\n", p)

		ocwd := lfp.cwd
		lfp.cwd = p

		d, f, err := lfp.Dir(filespec)
		lfp.cwd = ocwd
		return d, f, err
	} else {
		fmt.Printf("NotExists: %s\n", p)
	}

	return []FileDef(nil), []FileDef(nil), nil
}

// Dir returns a list of files in the current directory
func (lfp *NetworkUserFileProvider) Dir(filespec string) ([]FileDef, []FileDef, error) {

	apattern := "(?i)^(.+)#(0x|[$])([a-f0-9]+)$"
	ra := regexp.MustCompile(apattern)

	var dlist = make([]FileDef, 0)
	var flist = make([]FileDef, 0)

	lfp.cwd = strings.Trim(lfp.cwd, "/")

	fmt.Printf("Calling dir for %s\n", lfp.cwd)

	if lfp.cwd != "" && lfp.cwd != "\\" && lfp.cwd != "/" {
		dlist = append(dlist, FileDef{Name: "..", Size: 0, Kind: DIRECTORY, Path: lfp.cwd, Owner: lfp})
	}

	var raw []byte
	var err error

	if lfp.remint {
		raw, err = s8webclient.CONN.FetchRemIntDir(lfp.remintname, lfp.cwd, filespec)
	} else if lfp.system {
		raw, err = s8webclient.CONN.FetchSystemDir(lfp.cwd, filespec)
	} else if lfp.project {
		raw, err = s8webclient.CONN.FetchProjectDir(lfp.projectname, lfp.cwd, filespec)
	} else if lfp.shared {
		raw, err = s8webclient.CONN.FetchLegacyDir(lfp.cwd, filespec)
	} else {
		raw, err = s8webclient.CONN.FetchUserDir(lfp.cwd, filespec)
	}

	if err != nil {
		return dlist, flist, err
	}

	rows := strings.Split(string(raw), "\r\n")
	for _, r := range rows {

		if r == "" {
			continue
		}

		parts := strings.SplitN(r, ":", 4)
		typesym := parts[0]
		name := parts[1]
		size := utils.StrToInt(parts[2])
		desc := parts[3]

		if typesym == "D" {
			if name != "" {
				f := FileDef{Name: name, Description: desc, Size: 0, Kind: DIRECTORY, Path: lfp.cwd + "/" + name, Owner: lfp, Extension: ""}
				dlist = append(dlist, f)
			}
		} else {
			f := FileDef{}
			f.Kind = FILE
			f.Name = name
			f.Path = lfp.cwd + string('/') + name
			f.Size = int64(size)
			f.Owner = lfp
			f.Extension = GetExt(name)
			z := len(f.Extension) + 1
			f.Name = name[0 : len(name)-z]
			f.Description = desc

			if ra.MatchString(f.Name) {
				m := ra.FindAllStringSubmatch(f.Name, -1)
				f.Name = m[0][1]
			} else {
				fmt.Printf("%s.%s not match\n", f.Name, f.Extension)
			}

			flist = append(flist, f)
		}
	}

	return dlist, flist, nil

}

// ChDir() changes DIRECTORY
func (lfp *NetworkUserFileProvider) ChDir(p string) error {
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

func (lfp *NetworkUserFileProvider) Exists(p string, f string) bool {

	var exists bool = true

	// if len(p) > 0 && rune(p[len(p)-1]) == '/' {
	// 	p = p[0 : len(p)-1]
	// }

	if lfp.basedir != "" {
		p = strings.Replace(lfp.basedir+"/"+p, "//", "/", -1)
		if strings.HasSuffix(p, "/") {
			p = p[0 : len(p)-1]
		}
	}

	p = strings.TrimRight(p, "/")

	log.Printf("NetworkUserFileProvider.Exists(%s, %s)\n", p, f)

	p = strings.ToLower(p)
	f = strings.ToLower(f)

	if f != "" {

		if lfp.remint {
			exists, _ = s8webclient.CONN.ExistsRemIntFile(lfp.remintname, p, f)
		} else if lfp.system {
			exists, _ = s8webclient.CONN.ExistsSystemFile(p, f)
		} else if lfp.project {
			exists, _ = s8webclient.CONN.ExistsProjectFile(lfp.projectname, p, f)
		} else if lfp.shared {
			exists, _ = s8webclient.CONN.ExistsLegacyFile(p, f)
		} else {
			fmt.Printf("user file check %s, %s\n", p, f)
			exists, _ = s8webclient.CONN.ExistsUserFile(p, f)
		}

	}

	if exists {
		fmt.Println("Yes")
	} else {
		fmt.Println("No")
	}

	return exists
}

func (lfp *NetworkUserFileProvider) Lock(p string, f string) error {

	var lock error = nil

	if lfp.basedir != "" {
		p = strings.Replace(lfp.basedir+"/"+p, "//", "/", -1)
		if strings.HasSuffix(p, "/") {
			p = p[0 : len(p)-1]
		}
	}

	log.Printf("NetworkUserFileProvider.Lock(%s, %s)\n", p, f)

	if f != "" {
		if lfp.remint {
			lock = s8webclient.CONN.LockRemIntFile(lfp.remintname, p, f)
		} else if lfp.system {
			lock = s8webclient.CONN.LockSystemFile(p, f)
		} else if lfp.project {
			lock = s8webclient.CONN.LockProjectFile(lfp.projectname, p, f)
		} else if lfp.shared {
			lock = s8webclient.CONN.LockLegacyFile(p, f)
		} else {
			lock = s8webclient.CONN.LockUserFile(p, f)
		}

	}

	return lock
}

func (lfp *NetworkUserFileProvider) Meta(p string, f string, meta map[string]string) error {

	var lock error = nil

	if lfp.basedir != "" {
		p = strings.Replace(lfp.basedir+"/"+p, "//", "/", -1)
		if strings.HasSuffix(p, "/") {
			p = p[0 : len(p)-1]
		}
	}

	log.Printf("NetworkUserFileProvider.Meta(%s, %s)\n", p, f)

	if f != "" {
		if lfp.remint {
			lock = s8webclient.CONN.MetaDataRemIntFile(lfp.remintname, p, f, meta)
		} else if lfp.system {
			lock = s8webclient.CONN.MetaDataSystemFile(p, f, meta)
		} else if lfp.project {
			lock = s8webclient.CONN.MetaDataProjectFile(lfp.projectname, p, f, meta)
		} else if lfp.shared {
			lock = s8webclient.CONN.MetaDataLegacyFile(p, f, meta)
		} else {
			lock = s8webclient.CONN.MetaDataUserFile(p, f, meta)
		}

	}

	return lock
}

func (lfp *NetworkUserFileProvider) MkDir(p string, f string) error {

	var lock error = nil

	if lfp.basedir != "" {
		p = strings.Replace(lfp.basedir+"/"+p, "//", "/", -1)
		if strings.HasSuffix(p, "/") {
			p = p[0 : len(p)-1]
		}
	}

	log.Printf("NetworkUserFileProvider.MkDir(%s, %s)\n", p, f)

	if f != "" {
		if lfp.remint {
			lock = s8webclient.CONN.CreateRemIntDir(lfp.remintname, p, f)
		} else if lfp.system {
			lock = s8webclient.CONN.CreateSystemDir(p, f)
		} else if lfp.project {
			lock = s8webclient.CONN.CreateAProjectDir(lfp.projectname, p, f)
		} else if lfp.shared {
			lock = s8webclient.CONN.CreateLegacyDir(p, f)
		} else {
			lock = s8webclient.CONN.CreateUserDir(p, f)
		}

	}

	return lock
}

func (lfp *NetworkUserFileProvider) Delete(p string, f string) error {

	var lock error = nil

	if lfp.basedir != "" {
		p = strings.Replace(lfp.basedir+"/"+p, "//", "/", -1)
		if strings.HasSuffix(p, "/") {
			p = p[0 : len(p)-1]
		}
	}

	log.Printf("NetworkUserFileProvider.Delete(%s, %s)\n", p, f)

	if f != "" {
		if lfp.system {
			lock = s8webclient.CONN.DeleteSystemFile(p, f)
		} else if lfp.project {
			lock = s8webclient.CONN.DeleteProjectFile(lfp.projectname, p, f)
		} else if lfp.shared {
			lock = s8webclient.CONN.DeleteLegacyFile(p, f)
		} else {
			lock = s8webclient.CONN.DeleteUserFile(p, f)
		}

	}

	return lock
}

func (lfp *NetworkUserFileProvider) Share(p string, f string) (string, string, bool, error) {

	var err error = nil
	var host string
	var port string
	var c bool

	if lfp.basedir != "" {
		p = strings.Replace(lfp.basedir+"/"+p, "//", "/", -1)
		if strings.HasSuffix(p, "/") {
			p = p[0 : len(p)-1]
		}
	}

	log.Printf("NetworkUserFileProvider.Share(%s, %s)\n", p, f)

	if f != "" {
		if lfp.system {
			host, port, c, err = s8webclient.CONN.ShareSystemFile(p, f)
		} else if lfp.project {
			host, port, c, err = s8webclient.CONN.ShareProjectFile(lfp.projectname, p, f)
		} else if lfp.shared {
			host, port, c, err = s8webclient.CONN.ShareLegacyFile(p, f)
		} else {
			host, port, c, err = s8webclient.CONN.ShareUserFile(p, f)
		}

	}

	return host, port, c, err
}

func (lfp *NetworkUserFileProvider) Rename(p string, f string, nf string) error {

	var lock error = nil

	if lfp.basedir != "" {
		p = strings.Replace(lfp.basedir+"/"+p, "//", "/", -1)
		if strings.HasSuffix(p, "/") {
			p = p[0 : len(p)-1]
		}
	}

	log.Printf("NetworkUserFileProvider.Rename(%s, %s)\n", p, f)

	if f != "" {
		if lfp.system {
			lock = s8webclient.CONN.RenameSystemFile(p, f, nf)
		} else if lfp.project {
			lock = s8webclient.CONN.RenameProjectFile(lfp.projectname, p, f, nf)
		} else if lfp.shared {
			lock = s8webclient.CONN.RenameLegacyFile(p, f, nf)
		} else {
			lock = s8webclient.CONN.RenameUserFile(p, f, nf)
		}

	}

	return lock

}
