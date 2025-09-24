package disk

import (
	"errors"
	"strings"
	"time"

	"paleotronic.com/fmt"
)

type DSKFileMetaData struct {
	Name         string
	Path         string
	Size         int
	Type         string
	Ext          string
	CreatedDate  time.Time
	ModifiedDate time.Time
	Directory    bool
}

func (dsk *DSKWrapper) GenericReadData(path string, name string) (int, []byte, error) {

	fmt.Printf("GenericReadData(%s)\n", name)

	switch {
	case dsk.Format.IsOneOf(DF_DOS_SECTORS_13, DF_DOS_SECTORS_16):
		_, info, err := dsk.AppleDOSGetCatalog(name)
		if err != nil {
			return 0, []byte(nil), err
		}
		if len(info) != 1 {
			return 0, []byte(nil), errors.New("file not found")
		}
		fd := info[0]
		_, addr, data, err := dsk.AppleDOSReadFile(fd)
		fmt.Println(string(data))
		return addr, data, err
	case dsk.Format.IsOneOf(DF_RDOS_3, DF_RDOS_32, DF_RDOS_33):
		info, err := dsk.RDOSGetCatalog(name)
		if err != nil {
			return 0, []byte(nil), err
		}
		fmt.Println("Info returned", info)
		if len(info) != 1 {
			return 0, []byte(nil), errors.New("file not found")
		}
		fd := info[0]
		_, data, err := dsk.RDOSReadFile(fd)
		return 0, data, err
	case dsk.Format.IsOneOf(DF_PASCAL):
		info, err := dsk.PascalGetCatalog(name)
		if err != nil {
			return 0, []byte(nil), err
		}
		fmt.Println("Info returned", info)
		if len(info) != 1 {
			return 0, []byte(nil), errors.New("file not found")
		}
		fd := info[0]
		data, err := dsk.PascalReadFile(fd)
		return 0, data, err
	case dsk.Format.IsOneOf(DF_PRODOS, DF_PRODOS_400KB, DF_PRODOS_800KB, DF_PRODOS_CUSTOM):
		fmt.Println("PRODOS")
		_, info, err := dsk.PRODOSGetCatalogPathed(2, path, "*")
		if err != nil {
			return 0, []byte(nil), err
		}

		for _, fd := range info {
			if strings.ToLower(fd.Name()) == strings.ToLower(name) {
				_, addr, data, err := dsk.PRODOSReadFile(fd)
				return addr, data, err
			}
		}

		return 0, []byte(nil), errors.New("file not found")

	}

	return 0, []byte(nil), errors.New("Unsupported type")

}

func (dsk *DSKWrapper) GenericGetCatalog(path string, pattern string) ([]DSKFileMetaData, error) {

	var out []DSKFileMetaData

	fmt.Printf("GenericGetCatalog(%s, %s)\n", path, pattern)

	switch {
	case dsk.Format.IsOneOf(DF_DOS_SECTORS_13, DF_DOS_SECTORS_16):
		_, info, err := dsk.AppleDOSGetCatalog(pattern)
		if err != nil {
			return out, err
		}
		for _, v := range info {
			f := DSKFileMetaData{
				Name:         v.Name(),
				Path:         "",
				Type:         v.Type().String(),
				Ext:          v.Type().Ext(),
				Size:         v.TotalSectors() * 256,
				CreatedDate:  time.Now(),
				ModifiedDate: time.Now(),
			}
			out = append(out, f)
		}
	case dsk.Format.IsOneOf(DF_RDOS_3, DF_RDOS_32, DF_RDOS_33):
		fmt.Println("RDOS")
		info, err := dsk.RDOSGetCatalog(pattern)
		if err != nil {
			return out, err
		}
		for _, v := range info {
			f := DSKFileMetaData{
				Name:         v.Name(),
				Path:         "",
				Type:         v.Type().String(),
				Ext:          v.Type().Ext(),
				Size:         v.Length(),
				CreatedDate:  time.Now(),
				ModifiedDate: time.Now(),
			}
			out = append(out, f)
		}
	case dsk.Format.IsOneOf(DF_PASCAL):
		info, err := dsk.PascalGetCatalog(pattern)
		if err != nil {
			return out, err
		}
		for _, v := range info {
			f := DSKFileMetaData{
				Name:         v.GetName(),
				Path:         "",
				Type:         v.GetType().String(),
				Ext:          v.GetType().Ext(),
				Size:         v.GetFileSize(),
				CreatedDate:  time.Now(),
				ModifiedDate: time.Now(),
			}
			out = append(out, f)
		}
	case dsk.Format.IsOneOf(DF_PRODOS, DF_PRODOS_400KB, DF_PRODOS_800KB, DF_PRODOS_CUSTOM):
		_, info, err := dsk.PRODOSGetCatalogPathed(2, path, pattern)
		if err != nil {
			return out, err
		}
		for _, v := range info {
			f := DSKFileMetaData{
				Name:         v.Name(),
				Path:         "",
				Type:         v.Type().String(),
				Ext:          v.Type().Ext(),
				Size:         v.Size(),
				CreatedDate:  v.CreateTime(),
				ModifiedDate: v.ModTime(),
			}
			if f.Ext == "DIR" {
				f.Ext = ""
				f.Directory = true
			}
			out = append(out, f)
		}
	}

	return out, nil

}
