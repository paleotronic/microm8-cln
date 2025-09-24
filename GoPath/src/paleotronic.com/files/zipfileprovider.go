package files

import (
	"archive/zip"
	"bytes"
	"errors"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"paleotronic.com/filerecord"
	"paleotronic.com/fmt"
)

type ZipFileProvider struct {
	content   map[string]*filerecord.FileRecord
	canupdate bool
	source    *filerecord.FileRecord
	fullpath  string
	cwd       string
}

func NewZipProvider(source *filerecord.FileRecord, fullpath string) *ZipFileProvider {

	z := &ZipFileProvider{
		canupdate: true,
		content:   make(map[string]*filerecord.FileRecord),
		source:    source,
		fullpath:  fullpath,
	}

	if len(source.Content) > 0 {
		z.readFiles()
	}

	return z

}

func (z *ZipFileProvider) GetPath() string {
	return z.fullpath
}

func (z *ZipFileProvider) writeFiles() error {

	bb := new(bytes.Buffer)
	w := zip.NewWriter(bb)

	for path, fr := range z.content {
		fh := &zip.FileHeader{
			Comment: fr.Description,
			Name:    path,
			Method:  zip.Deflate,
		}
		//fh.SetMode(os.FileMode)
		fh.SetModTime(fr.Modified)
		if fr.Directory {
			// set dir attribute
			fh.ExternalAttrs = 16
			if !strings.HasSuffix(fh.Name, "/") {
				fh.Name += "/"
			}
		}
		f, err := w.CreateHeader(fh)
		if err != nil {
			return err
		}
		f.Write(fr.Content)
	}

	w.Close()

	z.source.Content = bb.Bytes()

	// re-publish
	return WriteBytesViaProvider(GetPath(z.fullpath), GetFilename(z.fullpath), z.source.Content)

}

func (z *ZipFileProvider) readFiles() error {

	if z.content != nil && len(z.content) > 0 {
		return nil
	}

	bb := bytes.NewReader(z.source.Content)

	r, err := zip.NewReader(bb, int64(bb.Len()))

	if err != nil {
		return err
	}

	m := make(map[string]*filerecord.FileRecord)

	for _, file := range r.File {

		name := strings.ToLower(strings.Replace(file.Name, "\\", "/", -1))
		isDir := strings.HasSuffix(name, "/") || file.ExternalAttrs&16 == 16
		name = strings.Trim(name, "/")

		p := GetPath(name)
		f := GetFilename(name)

		rc, err := file.Open()
		if err != nil {
			return err
		}
		data, err := ioutil.ReadAll(rc)
		if err != nil {
			return err
		}

		fr := &filerecord.FileRecord{
			Description: file.Comment,
			ContentSize: int(file.UncompressedSize),
			Content:     data,
			FileName:    f,
			FilePath:    p,
			Created:     file.ModTime(),
			Modified:    file.ModTime(),
			Creator:     z.source.Creator,
			Owner:       z.source.Owner,
			Directory:   isDir,
		}

		m[name] = fr

	}

	z.content = m

	return nil

}

func (mfp *ZipFileProvider) exists(p, f string) (bool, []string) {

	p = strings.Trim(p, "/")

	if mfp.readFiles() != nil {
		return false, []string(nil)
	}

	if (p == "" || p == "/") && f == "" {
		return true, []string{""}
	}

	pattern := "(?i)^(.+)#(0x|[$])([a-f0-9]+)[.]([a-z0-9]+)$"
	r := regexp.MustCompile(pattern)

	for target, entry := range mfp.content {
		if strings.ToLower(entry.FilePath) == strings.ToLower(p) && strings.ToLower(entry.FileName) == strings.ToLower(f) {
			return true, []string{target}
		} else if strings.ToLower(entry.FilePath) == strings.ToLower(p) && r.MatchString(entry.FileName) {
			ext := GetExt(f)
			base := f[0 : len(f)-len(ext)-1]
			m := r.FindAllStringSubmatch(entry.FileName, -1)
			eExt := m[0][4]
			eBase := m[0][1]
			if strings.ToLower(eBase) == strings.ToLower(base) && strings.ToLower(eExt) == strings.ToLower(ext) {
				return true, []string{target}
			}
		}
	}

	return false, []string(nil)

}

func (mfp *ZipFileProvider) SetLoadAddress(p string, f string, address int) error {

	match, names := mfp.exists(p, f)

	if !match {
		return errors.New("file not found")
	}

	fr := mfp.content[names[0]]
	ext := GetExt(fr.FileName)
	base := fr.FileName[0 : len(fr.FileName)-len(ext)-1]
	pattern := "(?i)^(.+)#(0x|[$])([a-f0-9]+)[.]([a-z0-9])$"
	ra := regexp.MustCompile(pattern)
	if ra.MatchString(fr.FileName) {
		m := ra.FindAllStringSubmatch(fr.FileName, -1)
		base = m[0][1]
		ext = m[0][4]
	}
	fr.FileName = fmt.Sprintf("%s#0x%.4x.%s", base, address, ext)

	return mfp.writeFiles()

}

func (z *ZipFileProvider) IsReadOnly() bool {
	return !z.canupdate
}

func (z *ZipFileProvider) IsVisible() bool {
	return true
}

func (z *ZipFileProvider) GetPriority() int {
	return 0
}

func (z *ZipFileProvider) ChDir(p string) error {

	var target string
	if strings.HasPrefix(p, "/") {
		target = strings.ToLower(strings.Trim(p, "/"))
	} else {
		target = strings.ToLower(strings.Trim(z.cwd+"/"+p, "/"))
	}

	_, ok := z.content[target]
	if !ok && target != "" {
		return errors.New("Path not found")
	}
	z.cwd = target

	return nil
}

func (z *ZipFileProvider) Delete(p, f string) error {

	target := strings.ToLower(strings.Trim(p+"/"+f, "/"))

	// make sure not a path with files
	for key, _ := range z.content {
		if strings.HasPrefix(key, target+"/") {
			return errors.New("directory not empty")
		}
	}

	if _, ok := z.content[target]; ok {

		delete(z.content, target)
		return z.writeFiles()

	}

	return nil
}

func (z *ZipFileProvider) Exists(p, f string) bool {

	match, _ := z.exists(p, f)
	return match
}

func (z *ZipFileProvider) Lock(p, f string) error {

	return nil
}

func (z *ZipFileProvider) MkDir(p string, f string) error {

	p = strings.ToLower(p)
	f = strings.ToLower(f)

	if f != "" {
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
		p += f
	}

	if p != "" && p != "/" {
		// make sure path exists
		path := ""
		for _, s := range strings.Split(p, "/") {

			ppath := path
			path += s + "/"
			_, ok := z.content[strings.Trim(path, "/")]
			if !ok {
				z.content[strings.Trim(path, "/")] = &filerecord.FileRecord{
					FileName:  s,
					FilePath:  strings.Trim(ppath, "/"),
					Directory: true,
					Created:   time.Now().Local(),
					Modified:  time.Now().Local(),
				}
				err := z.writeFiles()
				if err != nil {
					z.readFiles()
					return err
				}
			}
		}
	}

	return nil
}

func (z *ZipFileProvider) Validate(f *filerecord.FileRecord) (*filerecord.FileRecord, error) {
	return nil, nil
}

func (z *ZipFileProvider) Meta(p, f string, meta map[string]string) error {
	return nil
}

func (z *ZipFileProvider) Share(p, f string) (string, string, bool, error) {
	return p, f, false, nil
}

func (z *ZipFileProvider) DirFromBase(p string, filespec string) ([]FileDef, []FileDef, error) {

	cwd := z.cwd
	z.cwd = ""

	defer func() {
		z.cwd = cwd
	}()

	err := z.ChDir(p)
	if err != nil {
		return []FileDef(nil), []FileDef(nil), err
	}

	return z.Dir(filespec)

}

func (z *ZipFileProvider) GetCurrentPath() string {
	return "/"
}

func (z *ZipFileProvider) SetCurrentPath(p string, create bool) error {
	return nil
}

func (z *ZipFileProvider) Dir(filespec string) ([]FileDef, []FileDef, error) {
	dlist, flist := []FileDef(nil), []FileDef(nil)

	if filespec == "" {
		filespec = "*.*"
	}

	regstr := strings.Replace(filespec, ".", "[.]", -1)
	regstr = strings.Replace(regstr, "*", ".*", -1)

	r := regexp.MustCompile("(?i)" + regstr)

	if z.cwd != "" && z.cwd != "/" {
		dlist = append(dlist, FileDef{Name: "..", Size: 0, Kind: DIRECTORY, Path: z.cwd, Owner: z})
	}

	apattern := "(?i)^(.+)#(0x|[$])([a-f0-9]+)$"
	ra := regexp.MustCompile(apattern)

	for _, v := range z.content {
		if v.FilePath == z.cwd {
			f := FileDef{}
			// File / Dir in path
			if v.Directory {
				f.Kind = DIRECTORY
				f.Name = v.FileName
				f.Path = z.cwd + string('/') + v.FileName
				f.Size = 0
				f.Owner = z
				f.Description = "<dir>"
				dlist = append(dlist, f)
			} else {
				f.Kind = FILE
				f.Name = v.FileName
				f.Path = z.cwd + string('/') + v.FileName
				f.Size = int64(v.ContentSize)
				f.Owner = z
				f.Extension = GetExt(f.Name)
				z := len(f.Extension) + 1
				f.Name = f.Name[0 : len(f.Name)-z]
				f.Description = ""

				if ra.MatchString(f.Name) {
					m := ra.FindAllStringSubmatch(f.Name, -1)
					f.Name = m[0][1]
				}

				if r.MatchString(v.FileName) {
					flist = append(flist, f)
				}
			}
		}
	}

	return dlist, flist, nil
}

func (z *ZipFileProvider) GetFileContent(p string, f string) (filerecord.FileRecord, error) {

	if f == "" {
		return filerecord.FileRecord{}, errors.New("Not found")
	}

	ok, list := z.exists(p, f)

	// target := strings.ToLower(strings.Trim(p+"/"+f, "/"))

	// fmt.Printf("Getfilecontent:%s\n", target)

	// fr, ok := z.content[target]
	// if ok {
	// 	fmt.Println("Found")
	// 	return filt(*fr), nil
	// }

	if ok {
		target := list[0]
		fr := z.content[target]
		fr.UserCanWrite = z.source.UserCanWrite
		return filt(*fr), nil
	}

	return filerecord.FileRecord{}, errors.New("Not found")
}

func (z *ZipFileProvider) SetFileContent(p string, f string, data []byte) error {

	target := strings.ToLower(strings.Trim(p+"/"+f, "/"))

	if z.content == nil {
		z.readFiles()
	}

	fr, ok := z.content[target]
	if ok {
		fr.Content = data
		fr.Modified = time.Now()
		z.content[target] = fr
		err := z.writeFiles()
		if err != nil {
			z.readFiles()
			return err
		}
		return nil
	}

	if p != "" && p != "/" {
		err := z.MkDir(p, "")
		if err != nil {
			return err
		}
	}

	z.content[target] = &filerecord.FileRecord{
		FileName:    GetFilename(target),
		FilePath:    strings.Trim(GetPath(target), "/"),
		Content:     data,
		ContentSize: len(data),
		Created:     time.Now().Local(),
		Modified:    time.Now().Local(),
	}

	err := z.writeFiles()
	if err != nil {
		z.readFiles()
		return err
	}

	return nil
}

func (zfp *ZipFileProvider) Rename(p string, f string, nf string) error {

	leader := strings.ToLower(strings.Trim(p+"/"+f, "/"))
	nleader := strings.ToLower(strings.Trim(p+"/"+nf, "/"))

	newcontent := make(map[string]*filerecord.FileRecord)

	for path, f := range zfp.content {
		if strings.HasPrefix(path, leader) {
			path = strings.Replace(path, leader, nleader, -1)
			f.FileName = GetFilename(path)
			f.FilePath = GetPath(path)
		}
		newcontent[path] = f
	}

	zfp.content = newcontent
	err := zfp.writeFiles()
	if err != nil {
		zfp.readFiles()
		return err
	}

	return nil
}
