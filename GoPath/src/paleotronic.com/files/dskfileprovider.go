package files

import (
	"errors"
	"strings" //"paleotronic.com/fmt"

	"paleotronic.com/disk"
	"paleotronic.com/filerecord"
	"fmt"
)

type DSKFileProvider struct {
	FileProvider
	cwd       string
	writeable bool
	basedir   string
	pattern   string
	priority  int
	pack      *disk.DSKWrapper // path to disk object mapper
}

func NewDSKFileProvider(basedir string, priority int) *DSKFileProvider {
	pfp := &DSKFileProvider{basedir: basedir, writeable: false, priority: priority, cwd: ""}
	//pfp.pack = make(map[string]*disk.DSKWrapper)
	return pfp
}

func (pfp *DSKFileProvider) Delete(p string, f string) error {
	ext := GetExt(f)
	f = strings.ReplaceAll(f, "."+ext, "")
	if pfp.pack.Format.ID == disk.DF_DOS_SECTORS_16 {
		err :=  pfp.pack.AppleDOSDeleteFile(f)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return WriteBytesViaProvider( GetPath(pfp.pack.Filename), GetFilename(pfp.pack.Filename), pfp.pack.Data )
	} else if pfp.pack.Format.ID == disk.DF_PRODOS {
		err := pfp.pack.PRODOSDeleteFile(p, f)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return WriteBytesViaProvider( GetPath(pfp.pack.Filename), GetFilename(pfp.pack.Filename), pfp.pack.Data )
	}
	return nil
}

func (pfp *DSKFileProvider) RegisterImage(dsk *disk.DSKWrapper) {
	pfp.pack = dsk
}

func (pfp *DSKFileProvider) IsVisible() bool {
	return true
}

func (pfp *DSKFileProvider) IsReadOnly() bool {
	return !pfp.writeable
}

// Lock method does nothing on a local filesystem
func (lfp *DSKFileProvider) Lock(p string, f string) error {
	return nil
}

// DirFromBase gives a list of matching files under path
func (pfp *DSKFileProvider) DirFromBase(p string, filespec string) ([]FileDef, []FileDef, error) {

	if pfp == nil {
		return []FileDef(nil), []FileDef(nil), nil
	}

	fmt.Printf("DSK.DirFromBase(%s, %s)\n", p, filespec)

	if (p == "" || p == "/") && pfp.basedir != "" {
		p = pfp.basedir
	}

	if pfp.Exists(p, "") {
		ocwd := pfp.cwd
		pfp.cwd = p

		d, f, err := pfp.Dir(filespec)
		pfp.cwd = ocwd
		fmt.Println("Path exists")
		return d, f, err
	}

	fmt.Println("Path does not exist")

	return []FileDef(nil), []FileDef(nil), nil
}

func (pfp *DSKFileProvider) Dir(filespec string) ([]FileDef, []FileDef, error) {

	if filespec == "" {
		filespec = "*"
	}

	//	fqpn := GetUserPath(pfp.basedir, strings.Split(pfp.cwd, string('/')))

	dlist := make([]FileDef, 0)
	flist := make([]FileDef, 0)

	if pfp.pack == nil {
		return dlist, flist, nil
	}

	fmt.Printf("checking for [%s] in mapper", pfp.cwd)

	dsk, exists := pfp.pack, true

	fmt.Printf("Format read is %s\n", dsk.Format.String())

	if exists {

		fmt.Println("Exists, getting catalog", pfp.cwd)

		fdlist, err := dsk.GenericGetCatalog(pfp.cwd, filespec)

		if err != nil {
			fmt.Println("Err:", err)
			return dlist, flist, err
		}

		fmt.Println(fdlist, err)

		for _, fd := range fdlist {
			if fd.Directory {
				f := FileDef{
					Name:        fd.Name,
					Description: fd.Type,
					Size:        int64(fd.Size),
					Kind:        DIRECTORY,
				}
				dlist = append(dlist, f)
			} else {
				ff := strings.Trim(pfp.cwd+"/"+fd.Name, "/")
				f := FileDef{
					Name:        fd.Name,
					Path:        ff,
					Description: fd.Type,
					Size:        int64(fd.Size),
					Kind:        FILE,
					// Extension:   fd.Ext,
				}
				f.Extension = GetExt(f.Name)
				z := len(f.Extension) + 1
				f.Name = f.Name[0 : len(f.Name)-z]
				flist = append(flist, f)
			}
		}

	}

	return dlist, flist, nil
}

func (pfp *DSKFileProvider) ChDir(p string) error {
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

func (pfp *DSKFileProvider) Exists(p string, f string) bool {

	if pfp.pack == nil {
		return false
	}

	fmt.Printf("*** DSK::Exists(%s, %s)\n", p, f)

	if p == "" && f == "" {
		return true
	}

	if (p == "" || p == "/") && pfp.basedir != "" {
		p = pfp.basedir
	}

	dsk, exists := pfp.pack, true
	if !exists || dsk == nil {
		fmt.Printf("unable to find dsk %s in %v", p, pfp.pack)
		return false
	}

	fmt.Printf("Searching DSK catalog for entry [%s/%s]", p, strings.ToUpper(f))

	list, err := dsk.GenericGetCatalog(p, strings.ToUpper(f))
	fmt.Println(list, err)
	if err == nil && len(list) > 0 {
		return true
	}

	return false

}

func (pfp *DSKFileProvider) GetFileContent(p string, f string) (filerecord.FileRecord, error) {

	fmt.Printf("DSKFileProvider.GetFileContent(%s, %s)\n", p, f)

	if (p == "" || p == "/") && pfp.basedir != "" {
		p = pfp.basedir
	}

	//	if (p != "/" && p != "") && f != "" {
	dsk, exists := pfp.pack, true
	if !exists {
		return filerecord.FileRecord{}, errors.New("FILE NOT FOUND")
	}
	//

	addr, data, err := dsk.GenericReadData(p, f)
	if err == nil {
		return filerecord.FileRecord{
			FileName:    f,
			FilePath:    p,
			Content:     data,
			ContentSize: len(data),
			Address:     addr,
		}, nil
	}

	//	}

	return filerecord.FileRecord{}, errors.New("FILE NOT FOUND")
}

func (pfp *DSKFileProvider) SetFileContent(p string, f string, data []byte) error {

	dsk, exists := pfp.pack, true
	if !exists {
		return errors.New("FILE NOT FOUND")
	}
	//
	
	ext := GetExt(f)
	f = strings.ReplaceAll(f, "."+ext, "")

	if dsk.Format.ID == disk.DF_DOS_SECTORS_16 {
		err := dsk.AppleDOSWriteFile(f, disk.AppleDOSFileTypeFromExt(ext), data, 0x4000)
		if err != nil {
			fmt.Println(err)
		}
		return WriteBytesViaProvider( GetPath(dsk.Filename), GetFilename(dsk.Filename), dsk.Data )
	} else if dsk.Format.ID == disk.DF_PRODOS {

		err := dsk.PRODOSWriteFile(p, f, disk.ProDOSFileTypeFromExt(ext), data, 0x4000)
		if err != nil {
			fmt.Println(err)
		}
		return WriteBytesViaProvider( GetPath(dsk.Filename), GetFilename(dsk.Filename), dsk.Data )
	}

	return errors.New(FPAccess)
}

// GetPriority returns the priority of this filesystem
func (pfp *DSKFileProvider) GetPriority() int {
	return pfp.priority
}

// GetCurrentPath returns the current working dir of this path
func (pfp *DSKFileProvider) GetCurrentPath() string {
	return pfp.cwd
}

// SetCurrentPath sets the current path, creating it if specified
func (pfp *DSKFileProvider) SetCurrentPath(p string, create bool) error {

	//	fqpn := GetRelativePath(pfp.basedir, strings.Split(p, string('/')))

	////fmt.Printntf("InternalFileProvider.SetCurrentPath() Checking is %s exists or not\n", fqpn)
	if (p == "" || p == "/") && pfp.basedir != "" {
		p = pfp.basedir
	}

	_, exists := pfp.pack, true

	if !exists {
		return errors.New(FPNotExist)
	}

	////fmt.Println("OK")

	pfp.cwd = p
	return nil

}

// ---------------------------------------
// MountDSKImage mounts a disk to the /disk/ path
func MountDSKImage(p, f string, disknum int) (string, error) {
	data, e := ReadBytesViaProvider(p, f)
	if e != nil {
		return "", e
	}
	dsk, e := disk.NewDSKWrapperBin(nil, data.Content, f)
	if e != nil {
		return "", e
	}

	fp := NewDSKFileProvider("", 0)
	fp.RegisterImage(dsk)
	AddMapping(strings.Trim(p, "/")+"/"+f, "", fp)

	switch disknum {
	case 0:
		dskp0.RegisterImage(dsk)
		return "/disk0/", nil
	case 1:
		dskp1.RegisterImage(dsk)
		return "/disk1/", nil
	}

	// return "", errors.New("Bad disk number")

	return p + "/" + f, nil
}
