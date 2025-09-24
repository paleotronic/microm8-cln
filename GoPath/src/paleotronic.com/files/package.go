package files

import (
    "strings"
    "errors"
    "bytes"
    //"paleotronic.com/fmt"
    "crypto/md5"
	"encoding/hex"
    "regexp"
)

const (
    maxFilename = 64
    maxMetaSize = 1024
)

type PackageDirEntry struct {
    Name string   // File name
    Meta string
    Data []byte
}

func MarshalInteger( size int ) ([]byte, error) {
    data := make([]byte, 4)
    data[0] = byte((size >> 24) & 0xff)
    data[1] = byte((size >> 16) & 0xff)
    data[2] = byte((size >> 8) & 0xff)
    data[3] = byte(size & 0xff)
    return data, nil
}

func UnmarshalInteger( data []byte ) (int, []byte, error) {

    if len(data) < 4 {
        return 0, data, errors.New("Insufficient data")
    }

    s := data[0:4]
    data = data[4:]

    size := (int(s[0]) << 24) | (int(s[1]) << 16) | (int(s[2]) << 8) | int(s[3])

    return size, data, nil

}

// Pack a string
func MarshalString( s string ) ([]byte, error) {

    data := make([]byte, 0)

    size := len(s)

    sizeb, err := MarshalInteger(size)
    if err != nil {
        return []byte(nil), err
    }
    data = append(data, sizeb...)

    b := []byte(s)

    data = append(data, b...)

    return data, nil

}

// Unpack a string
func UnmarshalString( data []byte ) (string, []byte, error) {

    if len(data) < 4 {
        return "", data, errors.New("Insufficient data")
    }

    size, data, err := UnmarshalInteger(data)
    if err != nil {
        return "", data, errors.New("Insufficient data")
    }

    if len(data) < size {
        return "", data, errors.New("Insufficient data")
    }

    s := data[0:size]
    data = data[size:]

    str := string(s)

    return str, data, nil

}

func (cde *PackageDirEntry) MarshalBinary() ([]byte,error) {
    data := make([]byte,0)

    // Name field
    nameBytes, e := MarshalString( cde.Name )
    if e != nil {
        return []byte(nil), e
    }
    data = append(data, nameBytes...)

    // Meta field
    metaBytes, e := MarshalString( cde.Meta )
    if e != nil {
        return []byte(nil), e
    }
    data = append(data, metaBytes...)

    size := len(cde.Data)

    // Size
    sizeb, e := MarshalInteger(size)
    if e != nil {
        return []byte(nil), e
    }
    data = append(data, sizeb...)

    // Data
    data = append(data, cde.Data...)

    return data, nil
}

func (cde *PackageDirEntry) UnmarshalBinary(data []byte) ([]byte, error) {

    name, data, e := UnmarshalString(data)
    if e != nil {
        return data, e
    }
    cde.Name = name

    meta, data, e := UnmarshalString(data)
    if e != nil {
        return data, e
    }
    cde.Meta = meta

    size, data, e := UnmarshalInteger(data)
    if e != nil {
        return data, e
    }

    if len(data) < size {
        return data, errors.New("Insufficient length")
    }

    cde.Data = data[0:size]
    data = data[size:]

    return data, nil

}

type Package struct {
    Name string
    Meta string
    Content []PackageDirEntry
}

func (c *Package) MarshalBinary() ([]byte, error) {
    data := make([]byte, 0)

    id := []byte("S8CT")
    data = append(data, id...)

    // Name field
    nameBytes, e := MarshalString( c.Name )
    if e != nil {
        return []byte(nil), e
    }
    data = append(data, nameBytes...)

    // Meta field
    metaBytes, e := MarshalString( c.Meta )
    if e != nil {
        return []byte(nil), e
    }
    data = append(data, metaBytes...)

    for _, cde := range c.Content {
        b, e := cde.MarshalBinary()
        if e != nil {
            return []byte(nil), e
        }
        data = append(data, b...)
    }

    return data, nil

}

func (c *Package) UnmarshalBinary(data []byte) error {

    if len(data) < 4  {
        return errors.New("insufficient length")
    }

    if !bytes.Equal(data[0:4], []byte("S8CT")) {
        return errors.New("Not S8 Package file")
    }

    data = data[4:]

    name, data, e := UnmarshalString(data)
    if e != nil {
        return e
    }
    c.Name = name

    meta, data, e := UnmarshalString(data)
    if e != nil {
        return e
    }
    c.Meta = meta

    for (e == nil) && len(data) > 0 {
        cde := &PackageDirEntry{}

        data, e = cde.UnmarshalBinary(data)
        if e == nil {
            c.Content = append( c.Content, *cde )
            ////fmt.Printntf("Found file %s, %d bytes\n", string(cde.Name[0:]), len(cde.Data))
        } else {
            ////fmt.Printntf("Error: %v\n", e)
        }
    }

    return e

}

func (c *Package) Size() uint32 {
    if c.Content == nil {
        c.Content = make([]PackageDirEntry, 0)
    }
    var size uint32
    for _, e := range c.Content {
        size += 32 // name
        size += 4  // Size
        size += uint32(len(e.Data))
    }
    return size
}

func (c *Package) Length() int {
    if c.Content == nil {
        c.Content = make([]PackageDirEntry, 0)
    }
    return len(c.Content)
}

func slice2ByteArray( n []byte ) [maxFilename]byte {
    var nn [maxFilename]byte
    for i := 0; i<maxFilename; i++ {
        if i < len(n) {
            nn[i] = n[i]
        } else {
            nn[i] = 32
        }
    }
    return nn
}

func byteArrayToSlice( n [maxFilename]byte ) []byte {
    var nn = make([]byte, maxFilename)
    for i := 0; i<maxFilename; i++ {
            nn[i] = n[i]
    }

    nn = []byte(strings.Trim(string(nn), " "))

    return nn
}

func (c *Package) Remove(name string) {
    if c.Content == nil {
        c.Content = make([]PackageDirEntry, 0)
    }

    i := c.IndexOf(name)
    if i > -1 {
        a := c.Content[0:i]
        b := c.Content[i+1:]
        c.Content = append(a, b...)
    }
}

func (c *Package) AddFile( filepath string, metadata string ) error {

    basename := GetFilename(filepath)
    d, e := ReadBytes(filepath)

    if e != nil {
        return e
    }

    c.Add(basename, metadata, d)

    return nil

}

func (c *Package) Add( name, meta string, data []byte ) {
    if c.Content == nil {
        c.Content = make([]PackageDirEntry, 0)
    }

    cde := PackageDirEntry{ Name: name, Meta: meta, Data: data }

    i := c.IndexOf(name)
    if i > -1 {
        c.Content[i] = cde
    } else {
        c.Content = append(c.Content, cde)
    }
}

func (c *Package) IndexOf(name string) int {
    if c.Content == nil {
        c.Content = make([]PackageDirEntry, 0)
    }

    for i, e := range c.Content {
        ////fmt.Printntf("Comparing %s and %s\n", e.Name, name)
        if e.Name == name {
            return i
        }
    }

    return -1
}

/*
 * 	Name        string
	Path        string
	Kind        FileType
	Size        int64
	Writable    bool
	Owner       FileProvider
	Extension   string
	Description string
 */

func (c *Package) Dir(filespec string) ([]FileDef, []FileDef, error) {

	if filespec == "" {
		filespec = "*.*"
	}

	regstr := strings.Replace(filespec, ".", "[.]", -1)
	regstr = strings.Replace(regstr, "*", ".*", -1)

    r := regexp.MustCompile(regstr)

    d := make([]FileDef, 0)
    f := make([]FileDef, 0)

    for _, cde := range c.Content {
        name := strings.Trim(cde.Name, " ")

        if !r.MatchString(name) {
            continue
        }

        fd := FileDef{}
        fd.Extension = GetExt(name)
        fd.Name = name[0:len(name) - len(fd.Extension) - 1]
        fd.Size = int64(len(cde.Data))
        fd.Kind = FILE
        fd.Description = "Package file"
        fd.Path = name
        fd.Writable = false
        f = append(f, fd)
    }

    return d, f, nil

}

func (c *Package) Checksum() string {
    data := make([]byte, 0)
    for _, cde := range c.Content {
        data = append(data, cde.Data...)
    }
    enc := md5.Sum(data)
    senc := hex.EncodeToString(enc[0:16])
    return senc
}

func (c *Package) ChecksumFile( name string ) string {
    data := make([]byte, 0)
    i := c.IndexOf(name)
    if i >-1 {
        enc := md5.Sum(data)
        senc := hex.EncodeToString(enc[0:16])
        return senc
    }
    return ""
}

func StringToMap( s string ) map[string]string {
    m := make(map[string]string)
    parts := strings.Split(s, ",")
    for _, p := range parts {
        nv := strings.Split(p, "=")
        if len(nv) == 2 {
            m[nv[0]] = nv[1]
        }
    }
    return m
}

func MapToString( m map[string]string ) string {
    s := ""
    for k, v := range m {
        if v != "" {
            ss := k+"="+v
            if s != "" {
                s = s + ","
            }
            s = s + ss
        }
    }
    return s
}

func (c *Package) SetMetadata( n,  v string) {
    m := StringToMap( c.Meta )
    m[n] = v
    c.Meta = MapToString(m)
}

func (c *Package) SetFileMetadata( file, n,  v string) error {

    i := c.IndexOf(file)

    if i > -1 {
        cde := c.Content[i]
        m := StringToMap( cde.Meta )
        m[n] = v
        cde.Meta = MapToString(m)
        return nil
    }

    return errors.New("No such file: "+file)
}
