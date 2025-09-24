package files

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io/ioutil"
	log2 "log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	ogdl "gopkg.in/rveen/ogdl.v1"
	s8webclient "paleotronic.com/api"
	"paleotronic.com/core/hardware/apple2/woz"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/disk"
	"paleotronic.com/filerecord"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
	"paleotronic.com/octalyzer/assets"
	"paleotronic.com/octalyzer/bus"
	"paleotronic.com/presentation"
)

var BASEDIR = "Documents/microM8"

const (
	OLDBASEDIR = "microM8"
)

var System bool = true
var Project bool = false
var RemInt bool = false

var overlays map[string]FileProvider

func CopyContent(oldpath, newpath string) error {
	err := os.MkdirAll(newpath, 0755)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(oldpath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			err = CopyContent(oldpath+"/"+file.Name(), newpath+"/"+file.Name())
			if err != nil {
				return err
			}
		} else {
			var modtime = file.ModTime()
			var mode = file.Mode()
			data, err := ioutil.ReadFile(oldpath + "/" + file.Name())
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(newpath+"/"+file.Name(), data, mode)
			if err != nil {
				return err
			}
			os.Chtimes(newpath+"/"+file.Name(), modtime, modtime)
		}
	}

	return nil
}

func MigrateLegacyStructure(oldpath, newpath string) error {
	err := os.MkdirAll(newpath, 0755)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(oldpath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			err := CopyContent(oldpath+"/"+file.Name(), newpath+"/"+file.Name())
			if err != nil {
				return err
			}
		} else {
			var modtime = file.ModTime()
			var mode = file.Mode()
			data, err := ioutil.ReadFile(oldpath + "/" + file.Name())
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(newpath+"/"+file.Name(), data, mode)
			if err != nil {
				return err
			}
			os.Chtimes(newpath+"/"+file.Name(), modtime, modtime)
		}
	}

	log2.Printf("[migrate] Migrated userdata from %s to %s", oldpath, newpath)

	return nil
}

func init() {

	if settings.SystemType == "nox" {
		return
	}

	if runtime.GOOS == "windows" {
		odhome := GetUserDirectory("OneDrive/Documents")
		if Exists(odhome) {
			BASEDIR = "OneDrive/Documents/microM8"
		}
	}

	home := GetUserDirectory(BASEDIR)
	oldhome := GetUserDirectory(OLDBASEDIR)

	if !Exists(home) {
		if Exists(oldhome) {
			// rename it
			e := MigrateLegacyStructure(oldhome, home)
			if e != nil {
				panic(e)
			}
		} else {
			// create it
			_ = os.MkdirAll(home, 0700)
		}
	}

	if !Exists(home + "/logs") {
		_ = os.MkdirAll(home+"/logs", 0700)
	}

	if !Exists(home + "/MyPrints") {
		_ = os.MkdirAll(home+"/MyPrints", 0700)
	}

	if !Exists(home + "/settings") {
		_ = os.MkdirAll(home+"/settings", 0700)
	}

	if !Exists(home+"/MyRecordings") && Exists(home+"/recordings") {
		_ = os.Rename(home+"/recordings", home+"/MyRecordings")
	} else if !Exists(home + "/MyRecordings") {
		_ = os.MkdirAll(home+"/MyRecordings", 0700)
	}

	if !Exists(home+"/MySaves") && Exists(home+"/Saves") {
		_ = os.Rename(home+"/Saves", home+"/MySaves")
	} else if !Exists(home + "/MySaves") {
		_ = os.MkdirAll(home+"/MySaves", 0700)
	}

	if !Exists(home+"/MyScreenshots") && Exists(home+"/Screenshots") {
		_ = os.Rename(home+"/Screenshots", home+"/MyScreenshots")
	} else if !Exists(home + "/MyScreenshots") {
		_ = os.MkdirAll(home+"/MyScreenshots", 0700)
	}

	if !Exists(home+"/MyAudio") && Exists(home+"/Audio") {
		_ = os.Rename(home+"/Audio", home+"/MyAudio")
	} else if !Exists(home + "/MyAudio") {
		_ = os.MkdirAll(home+"/MyAudio", 0700)
	}

	if !Exists(home+"/MyDisks") && Exists(home+"/Disks") {
		_ = os.Rename(home+"/Disks", home+"/MyDisks")
	} else if !Exists(home + "/MyDisks") {
		_ = os.MkdirAll(home+"/MyDisks", 0700)
	}

	overlays = make(map[string]FileProvider)

	CheckPaletteCfg()

}

const MAX_SETTINGS = 16

func GetBinPath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	// macos: base dir should be same dir as .app file, not inside

	userDir := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		userDir = strings.Replace(os.Getenv("USERPROFILE"), "\\", "/", -1)
	}

	// TODO: remove this logging once testing complete
	f, _ := os.Create(userDir + "/Desktop/nox.emu.log")
	defer f.Close()
	f.WriteString("Binary path is " + dir + "\n")

	dir = strings.TrimRight(strings.Replace(dir, "\\", "/", -1), "/")

	if runtime.GOOS == "darwin" && strings.HasSuffix(strings.ToLower(dir), "/contents/macos") {

		// we need to be outside the app dir
		//  aaa.app/Contents/microm8
		parts := strings.Split(dir, "/")
		parts = parts[:len(parts)-3]
		dir = strings.Join(parts, "/")
	}

	f.WriteString("Search path is " + dir + "\n")

	return dir
}

func GetSysRoot() string {
	s := "/"
	if runtime.GOOS == "windows" {
		s = os.Getenv("HOMEDRIVE")
	}
	s = strings.Replace(s, "\\", "/", -1)
	log.Printf("SysRoot is [%s]", s)
	return s
}

func GetSettingsDirectory() string {
	filepath := GetUserDirectory(BASEDIR) + "/settings"
	return filepath
}

func GetSettingsFiles() []string {
	out := make([]string, 0)
	dir := GetSettingsDirectory()

	for i := 1; i <= MAX_SETTINGS; i++ {
		files, err := filepath.Glob(dir + fmt.Sprintf("/%.2d_*/video.cfg", i))
		if err != nil {
			return out
		}
		if len(files) > 0 {
			files[0] = strings.Replace(files[0], "\\", "/", -1)
			parts := strings.Split(files[0], "/")
			out = append(out, parts[len(parts)-2])
		} else {
			out = append(out, fmt.Sprintf("%.2d_empty", i))
		}
	}
	return out
}

func GetDiskSaveDirectory(index int) string {
	diskname := GetFilename(settings.PureBootVolume[index])
	diskname = strings.Replace(diskname, " ", "_", -1)
	diskname = strings.Replace(diskname, ".", "_", -1)
	diskname = strings.Replace(diskname, "(", "_", -1)
	diskname = strings.Replace(diskname, ")", "_", -1)
	diskname = strings.Replace(diskname, "[", "_", -1)
	diskname = strings.Replace(diskname, "]", "_", -1)
	if diskname == "" {
		diskname = "nodisk"
	}
	filepath := GetUserDirectory(BASEDIR) + "/MySaves/" + diskname
	filepath = strings.Replace(filepath, "\\", "/", -1)
	if !Exists(filepath) {
		os.MkdirAll(filepath, 0755)
	}
	return filepath
}

func GetDiskSaveFiles(index int) []string {
	out := make([]string, 0)
	dir := GetDiskSaveDirectory(index)
	for i := 0; i < 8; i++ {
		if Exists(fmt.Sprintf("%s/microM8%d.frz", dir, i)) {
			out = append(out, fmt.Sprintf("%s/microM8%d.frz", dir, i))
		} else {
			out = append(out, "")
		}
	}
	return out
}

func GetDisk(b byte) *disk.DSKWrapper {
	switch b {
	case 0x00:
		return dskp0.pack
	case 0x01:
		return dskp1.pack
	}
	return nil
}

func providers() []FileProvider {

	authed := s8webclient.CONN.IsAuthenticated()
	connected := s8webclient.CONN.IsConnected()

	if RemInt {
		return r_providers
	} else if settings.EBOOT || !connected {
		//fmt.Println("system")
		log.Println("Using eboot")
		return e_providers
	} else if System {
		//fmt.Println("system")
		log.Println("Using system")
		return s_providers
	} else if Project {
		return p_providers
	} else if authed {
		log.Println("Using user")
		return u_providers
	} else {
		//fmt.Println("system")
		log.Println("Using system")
		return s_providers
	}

}

func SetProject(name string) {
	Project = (name != "")
	p_providers = []FileProvider{
		NewMappedFileProvider(
			map[string]ProviderHolder{"userdata": ProviderHolder{BasePath: "", Provider: NewNetworkUserFileProvider("", true, false, 0)},
				"": ProviderHolder{BasePath: "", Provider: NewProjectProvider("", true, true, true, name, 0)},
			}),
	}
}

func SetRemIntUUID(name string) {
	RemInt = (name != "")
	r_providers = []FileProvider{
		NewMappedFileProvider(
			map[string]ProviderHolder{
				"software": ProviderHolder{BasePath: "", Provider: NewNetworkUserFileProvider("", false, true, 0)},
				"system":   ProviderHolder{BasePath: "", Provider: NewNetworkSystemFileProvider("", true, false, 0)},
				"":         ProviderHolder{BasePath: "", Provider: NewNetworkRemIntFileProvider("", name, true, false, 0)},
			}),
	}
}

func SetProjectRO(name string) {
	Project = (name != "")
	p_providers = []FileProvider{
		NewMappedFileProvider(
			map[string]ProviderHolder{"userdata": ProviderHolder{BasePath: "", Provider: NewNetworkUserFileProvider("", true, false, 0)},
				"": ProviderHolder{BasePath: "", Provider: NewProjectProvider("", false, true, true, name, 0)},
			}),
	}
}

func MkdirViaProvider(p string) error {

	p = strings.ToLower(p)

	p = "/" + strings.Trim(p, "/")

	if exists, rp, fp := hasOverlay(p); exists {
		path := rp
		if path == "" {
			path = fp.GetCurrentPath()
		}

		return fp.MkDir(GetPath(path), GetFilename(path))
	}

	for _, fp := range providers() {

		if fp.IsReadOnly() {
			continue
		}

		return fp.MkDir(GetPath(p), GetFilename(p))
	}

	return errors.New(FPAccess)
}

func CopyFileViaProviders(p string, dp string) error {

	SetLED0(true)
	defer SetLED0(false)

	p = strings.ToLower(p)
	dp = strings.ToLower(dp)

	p = "/" + strings.Trim(p, "/")
	dp = "/" + strings.Trim(dp, "/")

	b, err := ReadBytesViaProvider(GetPath(p), GetFilename(p))

	if err != nil {
		return err
	}

	// Got file, write it
	return WriteBytesViaProvider(GetPath(dp), GetFilename(dp), b.Content)
}

func DeleteViaProvider(p string) error {

	SetLED0(true)
	defer SetLED0(false)

	p = strings.ToLower(p)

	p = "/" + strings.Trim(p, "/")

	if exists, rp, fp := hasOverlay(p); exists {

		fmt.Println("Using overlay with rp =", rp)

		path := rp
		if path == "" {
			path = fp.GetCurrentPath()
		}

		if fp.Exists(GetPath(path), GetFilename(path)) {
			fmt.Printf("Path=%s, f=%s\n", GetPath(path), GetFilename(path))
			return fp.Delete(GetPath(path), GetFilename(path))
		}
	}

	for _, fp := range providers() {
		if fp.IsReadOnly() {
			continue
		}

		return fp.Delete(GetPath(p), GetFilename(p))
	}

	return errors.New(FPAccess)
}

func GetExt(s string) string {
	parts := strings.Split(s, ".")
	if len(parts) == 1 {
		return ""
	}
	return parts[len(parts)-1]
}

func GetPath(s string) string {
	parts := strings.Split(s, string('/'))
	if parts[0] == "" && len(parts) > 1 {
		parts = parts[1:]
	}
	return strings.Join(parts[0:len(parts)-1], string('/'))
}

func GetFilename(s string) string {
	s = strings.Replace(s, "\\", "/", -1)
	parts := strings.Split(s, string('/'))
	return parts[len(parts)-1]
}

func AppendBytesViaProvider(p, f string, data []byte) error {

	SetLED0(true)
	defer SetLED0(false)

	p = strings.ToLower(p)
	f = strings.ToLower(f)

	p = "/" + strings.Trim(p, "/")

	exdata, err := ReadBytesViaProvider(p, f)
	if err != nil {
		return err
	}

	exdata.Content = append(exdata.Content, data...)

	return WriteBytesViaProvider(p, f, exdata.Content)
}

func WriteBytesViaProvider(p, f string, data []byte) error {

	SetLED0(true)
	defer SetLED0(false)

	p = strings.ToLower(p)
	f = strings.ToLower(f)

	p = "/" + strings.Trim(p, "/")

	fmt.Printf("WriteBytesViaProvider() path = %s, file = %s\n", p, f)

	if exists, rp, fp := hasOverlay(p); exists {
		path := rp
		if path == "" {
			path = fp.GetCurrentPath()
		}

		err := fp.SetFileContent(path, f, data)

		if err == nil {
			fmt.Printf("Saved %s/%s via provider\n", rp, f)
			s8webclient.CONN.LogMessage("UFW", p+"/"+f)
		} else {
			fmt.Printf("Failed to save %s/%s via overlaid %s provider: %v\n", p, f, reflect.TypeOf(fp).String(), err)
		}

		return err

	}

	var err error

	for i, fp := range providers() {

		fmt.Printf("FP#%d: %v\n", i, fp)

		if fp.IsReadOnly() {
			continue
		}

		fmt.Printf("Gonna try %s\n", reflect.TypeOf(fp).Name())

		fmt.Printf("WriteBytesViaProvider() path = %s, file = %s, provider index %d is WRITABLE :D\n", p, f, i)

		//err := fp.SetCurrentPath(p, true)
		//if err != nil {
		//	return err
		//}

		err = fp.SetFileContent(p, f, data)

		if err == nil {
			log.Printf("Saved %s/%s via provider %d\n", p, f, i)
			s8webclient.CONN.LogMessage("UFW", p+"/"+f)
		} else {
			log.Printf("Failed to save %s/%s via provider %d: %v\n", p, f, i, err)
		}

		return err
	}

	fmt.Println("Got here sadly...")

	return errors.New(FPAccess)
}

type DirCacheEntry struct {
	Created time.Time
	dirs    []FileDef
	files   []FileDef
	err     error
}

var DirCache = make(map[string]*DirCacheEntry)

func IsCached(p string, filespec string) (bool, []FileDef, []FileDef, error) {

	key := strings.ToLower(strings.Trim(p, "/")) + ":" + strings.ToLower(filespec)

	info, ok := DirCache[key]
	if ok && time.Since(info.Created) < 5*time.Second {
		return true, info.dirs, info.files, info.err
	}

	// not cached
	return false, []FileDef(nil), []FileDef(nil), nil

}

func CacheDir(p, filespec string, dirs []FileDef, files []FileDef, err error) {

	key := strings.ToLower(strings.Trim(p, "/")) + ":" + strings.ToLower(filespec)

	DirCache[key] = &DirCacheEntry{
		Created: time.Now(),
		dirs:    dirs,
		files:   files,
		err:     err,
	}

}

func ReadDirViaProvider(p string, filespec string) ([]FileDef, []FileDef, error) {

	SetLED0(true)
	defer SetLED0(false)

	p = strings.ToLower(p)

	p = "/" + strings.Trim(p, "/")

	filespec = strings.ToLower(filespec)

	mf := make(map[string]FileDef)
	md := make(map[string]FileDef)

	var allitems []FileDef
	var alldirs []FileDef

	var notroot bool = (p != "" && p != "/")

	if exists, rp, fp := hasOverlay(p); exists && notroot {

		fmt.Printf("ReadDir: p=%s, exists=%v, rp=%s\n", p, exists, rp)

		path := rp
		if path == "" {
			path = fp.GetCurrentPath()
		}

		dirs, items, _ := fp.DirFromBase(path, filespec)

		for _, i := range items {
			_, ex := mf[i.Name+i.Extension]
			if !ex {
				//allitems = append(allitems, i)
				mf[i.Name+i.Extension] = i
			}
		}

		for _, i := range dirs {
			_, ex := md[i.Name]
			if !ex {
				//alldirs = append(alldirs, i)
				md[i.Name] = i
			}
		}

	} else {

		for _, fp := range providers() {

			path := p
			if path == "" {
				path = fp.GetCurrentPath()
			}

			fmt.Printf("Call fp.DirFromBase(%s, %s)\n", path, filespec)

			dirs, items, _ := fp.DirFromBase(path, filespec)

			for _, i := range items {
				_, ex := mf[i.Name+i.Extension]
				if !ex {
					//allitems = append(allitems, i)
					mf[i.Name+i.Extension] = i
				}
			}

			for _, i := range dirs {
				_, ex := md[i.Name]
				if !ex {
					//alldirs = append(alldirs, i)
					md[i.Name] = i
				}
			}

		}

	}

	dkeys := make([]string, 0)
	for k, _ := range md {
		dkeys = append(dkeys, k)
	}
	sort.Strings(dkeys)

	fkeys := make([]string, 0)
	for k, _ := range mf {
		fkeys = append(fkeys, k)
	}
	sort.Strings(fkeys)

	for _, k := range dkeys {
		alldirs = append(alldirs, md[k])
	}

	if notroot && (len(alldirs) == 0 || alldirs[0].Name != "..") {
		alldirs = append(alldirs,
			FileDef{
				Description: "<dir>",
				Name:        "..",
				Extension:   "",
				Kind:        DIRECTORY,
				Path:        "",
			},
		)
	}

	for _, k := range fkeys {
		allitems = append(allitems, mf[k])
	}

	CacheDir(p, filespec, alldirs, allitems, nil)

	return alldirs, allitems, nil

}

func ReadBytesViaProvider(p, f string) (filerecord.FileRecord, error) {

	r := bus.IsClock()
	bus.StartDefault()
	defer func() {
		if !r {
			bus.StopClock()
		}
	}()

	SetLED0(true)
	defer SetLED0(false)

	p = strings.ToLower(p)
	f = strings.ToLower(f)

	p = "/" + strings.Trim(p, "/")

	for len(p) > 1 && rune(p[0]) == '/' {
		p = p[1:]
	}

	//_ = ExistsViaProvider(p, f)
	ResolveBrowseable(p)

	fmt.Printf("ReadBytesViaProvider() path = %s, file = %s\n", p, f)

	if exists, rp, fp := hasOverlay(p); exists {

		fmt.Println("Using overlay with rp =", rp)

		path := rp
		if path == "" {
			path = fp.GetCurrentPath()
		}

		if fp.Exists(path, f) {
			fmt.Printf("Path=%s, f=%s\n", path, f)
			b, err := fp.GetFileContent(path, f)
			fmt.Println(err)
			if err == nil {
				s8webclient.CONN.LogMessage("UFR", p+"/"+f)
			}
			return b, err
		}
	}

	SetLED0(true)
	defer SetLED0(false)

	for _, fp := range providers() {

		path := p
		if path == "" {
			path = fp.GetCurrentPath()
		}

		if fp.Exists(path, f) {
			fmt.Printf("Path=%s, f=%s\n", path, f)
			b, err := fp.GetFileContent(path, f)

			fmt.Printf("GetFileContent(%s,%s) -> %d, %v\n", p, f, len(b.Content), err)

			if err == nil {
				s8webclient.CONN.LogMessage("UFR", p+"/"+f)
				return b, err
			}
		}

	}

	return filerecord.FileRecord{}, errors.New(FPNotExist)
}

func hasOverlay(p string) (bool, string, FileProvider) {

	tp := strings.TrimLeft(p, "/")

	log.Printf("Check overlay for [%s]\n", p)

	for path, fp := range overlays {
		log.Printf("Check %s vs %s\n", path, p)
		if strings.HasPrefix(strings.ToLower(p), strings.ToLower(path)) {
			//s := strings.Trim(strings.Replace(p, path, "", -1), "/")
			s := strings.Trim(p[len(path):], "/")
			log.Printf("[ok] Overlay path rel: %s\n", s)
			return true, s, fp
		}
		if strings.HasPrefix(strings.ToLower(tp), strings.ToLower(path)) {
			//s := strings.Trim(strings.Replace(p, path, "", -1), "/")
			s := strings.Trim(tp[len(path):], "/")
			log.Printf("[ok] Overlay path rel: %s\n", s)
			return true, s, fp
		}
	}

	return false, "", nil

}

func SetLoadAddressViaProvider(p, f string, address int) error {

	SetLED0(true)
	defer SetLED0(false)

	p = strings.ToLower(p)
	f = strings.ToLower(f)

	p = "/" + strings.Trim(p, "/")

	fmt.Printf("SetLoadAddressViaProvider() path = %s, file = %s, address = %d\n", p, f, address)

	if exists, rp, fp := hasOverlay(p); exists {

		fmt.Println("Using overlay with rp =", rp)

		path := rp
		if path == "" {
			path = fp.GetCurrentPath()
		}

		if fp.Exists(path, f) {
			fmt.Printf("Path=%s, f=%s\n", path, f)
			return fp.SetLoadAddress(path, f, address)
		}
	}

	for _, fp := range providers() {

		path := p
		if path == "" {
			path = fp.GetCurrentPath()
		}

		if fp.Exists(path, f) {
			fmt.Printf("Path=%s, f=%s\n", path, f)
			return fp.SetLoadAddress(path, f, address)
		}

	}

	return nil

}

func ResolveBrowseable(p string) {

	ext := GetExt(p)
	if IsBrowsable(ext) {
		if exists, _, _ := hasOverlay(p); !exists {
			fmt.Printf("*** AUTOMOUNT: %s\n", p)
			switch {
			case ext == "nib":
				fr, e := ReadBytesViaProvider(GetPath(p), GetFilename(p))
				if e == nil {
					d, err := woz.DeNibblizeImage(fr.Content, nil)
					if err == nil {
						dsk, err := disk.NewDSKWrapperBin(nil, d, GetFilename(p))
						if err == nil {
							fp := NewDSKFileProvider("", 0)
							dsk.WriteProtected = true
							fp.RegisterImage(dsk)
							AddMapping(p, "", fp)
						}
					}
				}
			case ext == "woz":
				fr, e := ReadBytesViaProvider(GetPath(p), GetFilename(p))
				if e == nil {
					w, err := woz.NewWOZImage(bytes.NewBuffer(fr.Content), memory.NewMemByteSlice(len(fr.Content)))
					if err == nil {
						fp := NewDSKFileProvider("", 0)
						dsk, err := w.ConvertToDSK()
						if err == nil {
							dsk.WriteProtected = true
							fp.RegisterImage(dsk)
							AddMapping(p, "", fp)
						}
					}
				}
			case ext == "pak":
				fr, e := ReadBytesViaProvider(GetPath(p), GetFilename(p))
				if e == nil {
					fp, e := NewOctContainer(&fr, p)
					if e == nil {
						AddMapping(p, "", fp)
					}
				}
			case ext == "zip":
				fr, e := ReadBytesViaProvider(GetPath(p), GetFilename(p))
				if e == nil {
					fp := NewZipProvider(&fr, p)
					AddMapping(p, "", fp)
				}
			case IsBootable(ext):
				fr, e := ReadBytesViaProvider(GetPath(p), GetFilename(p))
				if e == nil {
					fp := NewDSKFileProvider("", 0)
					dsk, err := disk.NewDSKWrapperBin(nil, fr.Content, p)
					if err == nil {
						fp.RegisterImage(dsk)
						AddMapping(p, "", fp)
					}
				}
			}
		}
	}
}

func ExistsViaProvider(p, f string) bool {

	SetLED0(true)
	defer SetLED0(false)

	// if f != "" {
	// 	_, _ = ResolveFileViaProvider(p, f)
	// }

	p = strings.ToLower(p)
	f = strings.ToLower(f)

	p = "/" + strings.Trim(p, "/")

	//fmt.Printf(">---------------------------------------------------------ExistsViaProvider(%s, %s)\n", p, f)

	ResolveBrowseable(p)

	p = strings.Trim(p, "/")

	// if f != "" {
	// 	log.Printf("Checking for overlay for: %s", p+"/"+f)
	// 	exists, _, _ := hasOverlay(p + "/" + f)
	// 	if exists {
	// 		return true
	// 	}
	// }

	if exists, rp, fp := hasOverlay(p); exists {
		path := rp
		if path == "" {
			path = fp.GetCurrentPath()
		}

		return fp.Exists(path, f)
	}

	for _, fp := range providers() {

		path := p
		if path == "" {
			path = fp.GetCurrentPath()
		}

		if fp.Exists(path, f) {
			return true
		}

		//fmt.Printf("ExistsViaProvider(%v) path = %s, file = %s - NOT EXIST\n", fp, path, f)
	}

	return false

}

func AddMapping(p string, base string, fp FileProvider) {

	p = strings.Trim(p, "/")

	fmt.Printf("OVERLAY ADDED: %s", p)

	// if strings.HasPrefix(strings.Trim(p, "/"), "local/") {
	// 	parts := strings.Split(strings.Trim(p, "/"), "/")
	// 	altp := GetUserPath(BASEDIR, parts[1:])
	// 	overlays[altp] = fp
	// 	fmt.Printf("ADDED LOCALMAPPER FOR PATH: %s\n", altp)
	// }

	overlays[p] = fp
}

func RemoveMapping(p string) {
	delete(overlays, p)
}

func ShareViaProvider(p, f string) (string, string, bool, error) {

	SetLED0(true)
	defer SetLED0(false)

	p = strings.ToLower(p)
	f = strings.ToLower(f)

	//fmt.Printf("ShareViaProvider() path = %s, file = %s\n", p, f)

	p = "/" + strings.Trim(p, "/")

	for _, fp := range providers() {

		path := p
		if path == "" {
			path = fp.GetCurrentPath()
		}

		if h, p, c, l := fp.Share(path, f); l == nil {
			return h, p, c, nil
		}

		//fmt.Printf("ShareViaProvider(%v) path = %s, file = %s - FAILED", fp, path, f)
	}

	return "", "", false, errors.New("Failed")

}

func LockViaProvider(p, f string) error {

	SetLED0(true)
	defer SetLED0(false)

	p = strings.ToLower(p)
	f = strings.ToLower(f)

	//fmt.Printf("LockViaProvider() path = %s, file = %s\n", p, f)

	p = "/" + strings.Trim(p, "/")

	for _, fp := range providers() {

		path := p
		if path == "" {
			path = fp.GetCurrentPath()
		}

		if l := fp.Lock(path, f); l == nil {
			return nil
		}

		//fmt.Printf("LockViaProvider(%v) path = %s, file = %s - FAILED", fp, path, f)
	}

	return errors.New("Failed")

}

func UnlockViaProvider(p, f string) error {

	SetLED0(true)
	defer SetLED0(false)

	//fmt.Printf("UnlockViaProvider() path = %s, file = %s\n", p, f)

	p = strings.ToLower(p)
	f = strings.ToLower(f)

	p = "/" + strings.Trim(p, "/")

	for _, fp := range providers() {

		path := p
		if path == "" {
			path = fp.GetCurrentPath()
		}

		if l := fp.Lock(path, f); l == nil {
			return nil
		}

		//fmt.Printf("UnlockViaProvider(%v) path = %s, file = %s - FAILED", fp, path, f)
	}

	return errors.New("Failed")

}

func RenameViaProvider(p, f, nf string) error {

	SetLED0(true)
	defer SetLED0(false)

	//fmt.Printf("UnlockViaProvider() path = %s, file = %s\n", p, f)

	p = strings.ToLower(p)
	f = strings.ToLower(f)

	p = "/" + strings.Trim(p, "/")

	for _, fp := range providers() {

		path := p
		if path == "" {
			path = fp.GetCurrentPath()
		}

		if l := fp.Rename(path, f, nf); l == nil {
			return nil
		}

		//fmt.Printf("UnlockViaProvider(%v) path = %s, file = %s - FAILED", fp, path, f)
	}

	return errors.New("Failed")

}

func ValidateViaProvider(current *filerecord.FileRecord) (*filerecord.FileRecord, error) {

	SetLED0(true)
	defer SetLED0(false)

	p := strings.ToLower(current.FilePath)

	p = "/" + strings.Trim(p, "/")

	for _, fp := range providers() {

		path := p
		if path == "" {
			path = fp.GetCurrentPath()
		}

		if nf, l := fp.Validate(current); l == nil {
			return nf, l
		}

	}

	return current, errors.New("Failed")

}

func MetaUpdateViaProvider(p, f string, meta map[string]string) error {

	SetLED0(true)
	defer SetLED0(false)

	p = strings.ToLower(p)
	f = strings.ToLower(f)

	//fmt.Printf("MetaUpdateViaProvider() path = %s, file = %s\n", p, f)

	p = "/" + strings.Trim(p, "/")

	for _, fp := range providers() {

		path := p
		if path == "" {
			path = fp.GetCurrentPath()
		}

		if l := fp.Meta(path, f, meta); l == nil {
			return nil
		}

		//fmt.Printf("LockViaProvider(%v) path = %s, file = %s - FAILED", fp, path, f)
	}

	return errors.New("Failed")

}

func getUserDirectoryUNIX(name string) string {
	return os.Getenv("HOME") + string('/') + name
}

func getUserDirectoryWin(name string) string {
	v := os.Getenv("USERPROFILE") + string('/') + name

	v = strings.Replace(v, "\\", "/", -1)

	return v
}

func GetUserDirectory(name string) string {

	if runtime.GOOS == "windows" {
		return getUserDirectoryWin(name)
	} else {
		return getUserDirectoryUNIX(name)
	}

}

func GetRelativePath(name string, subpath []string) string {
	v := name + string('/') + strings.Join(subpath, string('/'))

	v = strings.Replace(v, "\\", "/", -1)
	v = strings.Replace(v, "//", "/", -1)
	return v
}

func GetUserPath(name string, subpath []string) string {
	return GetUserDirectory(name) + string('/') + strings.Join(subpath, string('/'))
}

func GetExpandedDir(name string, subpath []string) string {
	s := name + string('/') + strings.Join(subpath, string('/'))
	log.Printf("Expanded path is [%s]", s)
	return s
}

func Exists(path string) bool {
	_, err := os.Stat(path)

	fmt.Printf("Checking existence of path: %s %v\n", path, err)

	return (err == nil)
}

func ExistsCP(path string) (bool, []string) {

	SetLED0(true)
	defer SetLED0(false)

	path = strings.Replace(path, "//", "/", -1)

	//fmt.Printf("ExistsCP for %s\n", path)

	pattern := "(?i)^(.+)#(0x|[$])([a-z0-9]+)[.]([a-z0-9]+)$"
	r := regexp.MustCompile(pattern)
	if r.MatchString(path) {
		if Exists(path) {
			return true, []string{path}
		}
		return false, []string(nil)
	}

	f := GetFilename(path)

	if f == "" {
		return Exists(path), []string{path}
	}

	if Exists(path) {
		return true, []string{path}
	}

	p := GetPath(path)

	if runtime.GOOS != "windows" && !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	e := GetExt(f)
	if e != "" {
		f = f[0 : len(f)-len(e)-1]
		pattern = p + "/" + f + "#*." + e
	} else {
		pattern = p + "/" + f + "#*"
	}

	fmt.Printf("ExistsCP looking for %s\n", pattern)

	results, _ := filepath.Glob(pattern)
	for i, v := range results {
		results[i] = strings.Replace(v, "\\", "/", -1)
	}

	return len(results) > 0, results

}

func ResolveFileViaProvider(p string, f string) ([]string, error) {

	SetLED0(true)
	defer SetLED0(false)

	ResolveBrowseable(p)

	needsExt := GetExt(f) != ""

	pattern := "(?i)^([^#]+)(#(0x|[$])([a-z0-9]+))?([.]([a-z0-9]+))$"
	r := regexp.MustCompile(pattern)

	spec := "*.*"

	if GetExt(f) != "" {
		spec = "*." + GetExt(f)
	}

	_, filelist, err := ReadDirViaProvider(p, spec)
	if err != nil {
		return []string(nil), err
	}

	fmt.Printf("Looking for: %s\n", f)

	for _, file := range filelist {

		fmt.Println("PATH =", file.Path, " vs ", f)
		name := GetFilename(file.Path)
		if strings.ToLower(f) == strings.ToLower(name) {
			return []string{name}, nil
		}

		if r.MatchString(name) {
			m := r.FindAllStringSubmatch(name, -1)
			base := m[0][1]
			ext := m[0][6]
			cname := base
			if ext != "" {
				cname += "." + ext
			}
			if strings.ToLower(f) == strings.ToLower(cname) || (!needsExt && strings.ToLower(f) == strings.ToLower(base)) {
				fmt.Printf("-------------------> Found: %s\n", name)
				return []string{name}, nil
			}
		}

	}

	return []string(nil), nil

}

func getFilenames(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return []string{}, err
	}
	defer f.Close()
	return f.Readdirnames(-1)
}

func getFileSize(path string) (int64, error) {
	s, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return s.Size(), nil
}

// CheckPathCI checks for a path and verifies its existence independant of case
func CheckPathCI(path string) (bool, string) {
	path = strings.Replace(path, "//", "/", -1)

	if runtime.GOOS == "windows" {
		ok, m := ExistsCP(path)
		p := path
		if ok {
			p = m[0]
		}
		return ok, p
	}

	wp := ""
	parts := strings.Split(path, "/")
	if parts[0] == "" {
		parts = parts[1:]
		wp = "/"
	}
	for parts[len(parts)-1] == "" {
		parts = parts[0 : len(parts)-1]
	}

	for idx, chunk := range parts {
		//fmt.Printf("Path chunk [%s]\n", chunk)
		files, err := getFilenames(wp)
		if err != nil {
			return false, wp
		}
		// we got some stuff, check for a match
		match := false
		for _, info := range files {
			if strings.ToLower(info) == strings.ToLower(chunk) {
				wp += info

				//fmt.Printf("wp: [%s]\n", wp)

				if idx < len(parts)-1 {
					wp += "/"
				}
				match = true
				break
			} else if idx == len(parts)-1 {
				// last element
				ok, m := ExistsCP(wp + chunk)
				if ok && m[0] == wp+info {
					return true, wp + info
				}
			}
		}
		if !match {
			return false, wp
		}
	}

	return true, wp

}

func ReadBytes(path string) ([]byte, error) {
	if !Exists(path) {
		return []byte(nil), errors.New("FILE NOT FOUND")
	}
	return ioutil.ReadFile(path)
}

func WriteString(path, data string, app bool) error {
	var f *os.File
	var e error

	if app {
		f, e = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	} else {
		f, e = os.Create(path)
	}

	if e != nil {
		return e
	}

	defer f.Close()

	_, e = f.WriteString(data)
	return e
}

func WriteBytes(path string, data []byte, app bool) error {
	var f *os.File
	var e error

	if app {
		f, e = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	} else {
		f, e = os.Create(path)
	}

	if e != nil {
		return e
	}

	defer f.Close()

	_, e = f.Write(data)
	if e == nil {
		e = f.Sync()
	}
	return e
}

var CompletionCacheFiles map[string][]FileDef
var CompletionCacheDirs map[string][]FileDef
var CompletionExpiry map[string]time.Time

// GetCompletions returns a list of likely file names based on a supplied prefix
func GetCompletions(basepath string) ([]string, []string) {

	if CompletionExpiry == nil {
		CompletionExpiry = make(map[string]time.Time)
	}

	if CompletionCacheFiles == nil {
		CompletionCacheFiles = make(map[string][]FileDef)
	}

	if CompletionCacheDirs == nil {
		CompletionCacheDirs = make(map[string][]FileDef)
	}

	p := basepath
	pmatch := ""
	dmatches, fmatches := []string(nil), []string(nil)

	if !strings.HasSuffix(basepath, "/") {
		parts := strings.Split(p, "/")
		pmatch = parts[len(parts)-1]
		p = strings.Join(parts[0:len(parts)-1], "/")
		p += "/"
	}

	fmt.Printf("Base: [%s], Prefix: [%s]\n", p, pmatch)

	// check if we have a cache we can use
	mf, okf := CompletionCacheFiles[strings.ToLower(p)]
	md, okd := CompletionCacheDirs[strings.ToLower(p)]
	me, oke := CompletionExpiry[strings.ToLower(p)]
	if oke && time.Now().After(me) {
		okf, okd = false, false
	}

	var f, d []FileDef
	var e error

	if okf && okd {
		f = mf
		d = md
	} else {
		d, f, e = ReadDirViaProvider(p, "*.*")
		CompletionCacheFiles[strings.ToLower(p)] = f
		CompletionCacheDirs[strings.ToLower(p)] = d
		CompletionExpiry[strings.ToLower(p)] = time.Now().Add(60 * time.Second)
	}

	if e != nil {
		return dmatches, fmatches
	}

	for _, v := range d {
		if pmatch == "" || strings.HasPrefix(strings.ToLower(v.Name), strings.ToLower(pmatch)) {
			n := strings.Replace(p+"/"+v.Name, "//", "/", -1)
			dmatches = append(dmatches, n) // add trailing slash to dir completions
		}
	}

	for _, v := range f {
		if pmatch == "" || strings.HasPrefix(strings.ToLower(v.Name), strings.ToLower(pmatch)) {
			n := strings.Replace(p+"/"+v.Name, "//", "/", -1)
			fmatches = append(fmatches, n)
		}
	}

	sort.Strings(dmatches)
	sort.Strings(fmatches)

	return dmatches, fmatches
}

var configSections = []string{"boot", "video", "palette", "control", "camera", "audio", "input", "hardware"}

func OpenPresentationState(filepath string) (*presentation.Presentation, error) {

	fmt.Printf("Opening pstate: %s\n", filepath)

	p := &presentation.Presentation{
		G:        make(map[string]*ogdl.Graph),
		Filepath: filepath,
	}

	for _, v := range configSections {
		spath := "/" + strings.Trim(filepath, "/") + "/" + v + ".cfg"
		if ExistsViaProvider(GetPath(spath), GetFilename(spath)) {
			fmt.Printf("presentation state exists: %s\n", spath)
			data, err := ReadBytesViaProvider(GetPath(spath), GetFilename(spath))
			if err == nil {
				p.LoadBytes(v, data.Content)
			}
		} // } else {
		// 	//p.LoadString(v, "")
		// 	return p, errors.New("Invalid pak format")
		// }
	}

	return p, nil
}

func OpenPresentationStateSoft(ent interfaces.Interpretable, filepath string) (*presentation.Presentation, error) {

	fmt.Println(filepath)

	p, e := NewPresentationStateDefault(ent, filepath)

	if e != nil {
		return nil, e
	}

	var found bool

	for _, v := range configSections {
		spath := "/" + strings.Trim(filepath, "/") + "/" + v + ".cfg"
		fmt.Printf("Looking for %s\n", spath)
		if ExistsViaProvider(GetPath(spath), GetFilename(spath)) {
			data, err := ReadBytesViaProvider(GetPath(spath), GetFilename(spath))
			if err == nil {
				p.LoadBytes(v, data.Content)
				found = true
			}
		}
	}

	if !found {
		return nil, errors.New("Not found")
	}

	return p, nil
}

func NewPresentationStateDefault(ent interfaces.Interpretable, path string) (*presentation.Presentation, error) {

	g := make(map[string]*ogdl.Graph)

	p := &presentation.Presentation{G: g, Filepath: path}

	for _, cfg := range configSections {
		fr, err := ReadBytesViaProvider("/boot/defaults", cfg+".cfg")
		if err == nil {
			p.LoadBytes(cfg, fr.Content)
		}
	}

	return p, nil

}

func SavePresentationState(p *presentation.Presentation, o *OctContainer) error {

	for section, _ := range p.G {
		data := p.GetBytes(section)
		spath := section + ".cfg"
		fmt.Printf("Saving state %s (%d bytes)\n", spath, len(data))
		err := o.SetFileContent("", spath, data)
		fmt.Printf("Result: %v\n", err)
		if err != nil {
			return err
		}
	}

	return nil

}

func SavePresentationStateToFolder(p *presentation.Presentation, path string) error {

	path = strings.Replace(path, "\\", "/", -1)
	if !ExistsViaProvider(path, "") {
		MkdirViaProvider(path)
	}

	for section, _ := range p.G {
		data := p.GetBytes(section)
		fname := section + ".cfg"
		fmt.Printf("Saving state %s (%d bytes)\n", fname, len(data))
		err := WriteBytesViaProvider(path, fname, data)
		fmt.Printf("Result: %v\n", err)
		if err != nil {
			return err
		}
	}

	return nil

}

func CheckPaletteCfg() {
	lfp := NewLocalFileProvider(GetUserDirectory(BASEDIR), true, 0)
	fr, err := lfp.GetFileContent("settings/default", "palette.cfg")
	if err != nil {
		return
	}
	// exists, so check it
	dat := md5.Sum(fr.Content)
	str := hex.EncodeToString(dat[:])
	if str != "4e7e98b41bfcb47422aceb877261d126" {
		return
	}
	// old and busted palette is in settings
	data, err := assets.Asset("bootsystem/boot/defaults/palette.cfg")
	if err != nil {
		panic(err)
	}
	lfp.SetFileContent("settings/default", "palette.cfg", data)
	//log2.Printf("UPDATED PALETTE CFG")
}

func SaveDefaultState(ent interfaces.Interpretable) error {
	path := "/local/settings/default"
	p, _ := presentation.NewPresentationState(ent, path)
	return SavePresentationStateToFolder(p, path)
}

func LoadDefaultState(ent interfaces.Interpretable) (*presentation.Presentation, error) {
	path := "/local/settings/default"
	if settings.SystemType == "nox" {
		path = "/boot/defaultsnox"
	}
	p, e := OpenPresentationStateSoft(ent, path)
	if e != nil {
		p, _ = NewPresentationStateDefault(ent, path)
		_ = SavePresentationStateToFolder(p, path)
		return p, nil
	}
	return p, e
}

func PurgeCache() {
	path := GetUserDirectory(BASEDIR + "/FILECACHE")
	os.RemoveAll(path)
}

func RecursiveDirViaProvider(p, filespec string) ([]FileDef, []FileDef, error) {

	fmt.Printf("RDir: %s/%s\n", p, filespec)

	var dirs, files []FileDef
	d, f, err := ReadDirViaProvider(p, filespec)
	if err != nil {
		return dirs, files, err
	}

	for _, file := range f {
		file.Path = p
		files = append(files, file)
	}

	for _, item := range d {
		if !strings.HasPrefix(item.Name, ".") && item.Extension == "" {
			// valid dir
			_, f, err = RecursiveDirViaProvider(p+"/"+item.Name, filespec)

			for _, file := range f {
				//file.Path = p
				files = append(files, file)
			}
		}
	}
	return dirs, files, nil
}

func RecursiveCopyViaProvider(srcdir, destdir string, ff func(pc float64)) error {

	srcdir = "/" + strings.Trim(srcdir, "/")

	base := "/" + strings.Trim(GetPath(srcdir), "/")

	_, files, err := RecursiveDirViaProvider(srcdir, "*.*")
	if err != nil {
		return err
	}

	pathmap := make(map[string]bool)
	for _, f := range files {
		pathmap[f.Path[len(base):]] = true
	}

	fmt.Println(pathmap)

	// mkdir yea?
	for path := range pathmap {
		fmt.Printf("mkdir for path %s\n", path)
		tp := "/" + strings.Trim(destdir, "/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		for _, p := range parts {
			if !ExistsViaProvider(tp, p) {
				e := MkdirViaProvider(tp + "/" + p)
				if e != nil {
					return e
				}
			}
			tp += "/" + p
		}
	}

	// Copy, yea?
	for i, f := range files {
		if ff != nil {
			pc := float64(i) / float64(len(files))
			ff(pc)
		}
		name := f.Name
		if f.Extension != "" {
			name += "." + f.Extension
		}
		fr, e := ReadBytesViaProvider(f.Path, name)
		if e != nil {
			return e
		}
		tp := "/" + strings.Trim(destdir, "/") + f.Path[len(base):]
		e = WriteBytesViaProvider(tp, name, fr.Content)
		if e != nil {
			return e
		}
		fmt.Printf("Copied %s/%s to %s\n", f.Path, name, tp)
	}

	return nil
}
