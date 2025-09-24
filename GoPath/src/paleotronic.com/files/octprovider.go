package files

import (
	"encoding/hex"
	"math/rand"
	"strings"

	"gopkg.in/rveen/ogdl.v1"
	"paleotronic.com/filerecord"
	"paleotronic.com/fmt"
	"paleotronic.com/presentation"
)

type OctContainer struct {
	ZipFileProvider
	octid string
	p     *presentation.Presentation
}

func NewOctContainer(source *filerecord.FileRecord, fullpath string) (*OctContainer, error) {
	o := &OctContainer{
		ZipFileProvider: *NewZipProvider(source, fullpath),
	}

	return o, o.Init(nil, "apple2e-en")

}

func NewOctContainerFromFile(fr []*filerecord.FileRecord, fullpath string, profile string) (*OctContainer, error) {

	source := &filerecord.FileRecord{
		FileName: GetFilename(fullpath),
		FilePath: GetPath(fullpath),
	}

	o := &OctContainer{
		ZipFileProvider: *NewZipProvider(source, fullpath),
	}

	e := o.Init(fr, profile)

	return o, e

}

func randomId(size int) string {

	data := make([]byte, size)

	for i, _ := range data {
		data[i] = byte(rand.Intn(256))
	}

	return hex.EncodeToString(data)

}

func (o *OctContainer) loadPresentationState() (*presentation.Presentation, error) {

	if o.p != nil {
		return o.p, nil
	}

	var e error
	o.p, e = OpenPresentationState(o.fullpath)

	return o.p, e
}

func (o *OctContainer) GetStartup() string {

	p, e := o.loadPresentationState()
	if e != nil {
		fmt.Printf("error reading p state: %s\n", e.Error())
		return ""
	}

	startup, e := p.ReadString("boot", "init.startfile")
	if e != nil {
		fmt.Printf("error reading startup state: %s\n", e.Error())
		return ""
	}

	return startup

}

func (o *OctContainer) GetAux() string {

	p, e := o.loadPresentationState()
	if e != nil {
		return ""
	}

	startup, _ := p.ReadString("boot", "init.auxdisk1")

	return startup

}

func (o *OctContainer) GetBackdrop() (string, float64, float64, float64, bool) {

	p, e := o.loadPresentationState()
	if e != nil {
		return "", 1, 0, 16, false
	}

	var backdrop = ""
	var opacity float64 = 1
	var zoom float64 = 16
	var zrat float64
	var camtrack bool

	if b, e := p.ReadString("video", "init.backdrop.source"); e == nil {
		backdrop = b
	}
	if f, e := p.ReadFloat("video", "init.backdrop.opacity"); e == nil {
		opacity = f
	}
	if f, e := p.ReadFloat("video", "init.backdrop.zoom"); e == nil {
		zoom = f
	}
	if f, e := p.ReadFloat("video", "init.backdrop.zrat"); e == nil {
		zrat = f
	}
	if i, e := p.ReadInt("video", "init.backdrop.camtrack"); e == nil {
		camtrack = (i != 0)
	}

	return backdrop, opacity, zrat, zoom, camtrack

}

func (o *OctContainer) GetProfile(def string) string {

	p, e := o.loadPresentationState()
	if e != nil {
		return def
	}

	var profile = def
	if i, e := p.ReadString("boot", "init.profile"); e == nil {
		profile = i
	}

	return profile

}

func (o *OctContainer) GetMusicTrack() (string, int, int) {

	p, e := o.loadPresentationState()
	if e != nil {
		return "", 0, 0
	}

	var musicfile string
	var leadin int
	var fadein int

	if i, e := p.ReadInt("audio", "init.music.leadin"); e == nil {
		leadin = int(i)
	}
	if i, e := p.ReadInt("audio", "init.music.fadein"); e == nil {
		fadein = int(i)
	}
	if i, e := p.ReadString("audio", "init.music.source"); e == nil {
		musicfile = i
	}

	return musicfile, leadin, fadein

}

func (o *OctContainer) GetControlFiles() []string {

	def := []string{"control", "control2", "control3", "control4"}

	p, e := o.loadPresentationState()
	if e != nil {
		return def
	}

	if i, e := p.ReadString("control", "init.controlprograms"); e == nil {
		return strings.Split(i, ",")
	}

	return def

}

func (o *OctContainer) GetDescription() string {

	p, e := o.loadPresentationState()
	if e != nil {
		return ""
	}

	startup, _ := p.ReadString("boot", "init.description")

	return startup

}

func (o *OctContainer) Init(fr []*filerecord.FileRecord, profile string) error {

	if o.Exists("", "OCTID") {

		data, e := o.GetFileContent("", "OCTID")
		if e == nil {
			o.octid = string(data.Content)
			return e
		}

	}

	o.octid = randomId(16)

	e := o.SetFileContent("", "OCTID", []byte(o.octid))
	if e != nil {
		return e
	}

	// Create a pak.cfg
	bootfiles := o.createIni(fr, profile)

	if bootfiles != nil && len(bootfiles) > 0 {

		for _, bootfile := range bootfiles {
			fmt.Printf("bootfile.Filename = [%s]", bootfile.FileName)
			ext := GetExt(bootfile.FileName)
			base := bootfile.FileName[0 : len(bootfile.FileName)-len(ext)-1]
			if ext == "bin" {
				bootfile.FileName = fmt.Sprintf("%s#0x%.4x.%s", base, bootfile.Address, ext)
			}

			e = o.SetFileContent("", bootfile.FileName, bootfile.Content)
			if e != nil {
				return e
			}
		}
	}

	return nil

}

type dummyNibbler struct{}

func (d *dummyNibbler) GetNibble(offset int) byte {
	return 0
}

func (d *dummyNibbler) SetNibble(offset int, value byte) {
	//
}

func (o *OctContainer) createIni(files []*filerecord.FileRecord, profile string) []*filerecord.FileRecord {

	//fmt.Printf("FR=%s\n", fr.FileName)
	ini := &presentation.Presentation{
		Filepath: o.fullpath,
		G:        make(map[string]*ogdl.Graph),
	}

	if len(files) == 0 {
		return []*filerecord.FileRecord{}
	}

	ini.WriteString("boot", "init.description", "pakfile of "+files[0].FileName)

	for i, fr := range files {

		//dsk, err := disk.NewDSKWrapperBin(&dummyNibbler{}, fr.Content, fr.FileName)
		// if err == nil && len(dsk.Data) == disk.STD_DISK_BYTES {
		// 	//if dsk.Format.ID != disk.DF_NONE && len(dsk.Data) == disk.STD_DISK_BYTES {
		// 	nb := woz.CreateWOZFromDSK(dsk, memory.NewMemByteSlice(233216))
		// 	ext := GetExt(fr.FileName)
		// 	fr.FileName = strings.Replace(fr.FileName, "."+ext, ".woz", -1)
		// 	fr.Content = nb.Data.Bytes()
		// 	//}
		// } else if strings.HasSuffix(fr.FileName, ".nib") {
		// 	nb := woz.CreateWOZFromNIB(fr.Content, memory.NewMemByteSlice(233216))
		// 	ext := GetExt(fr.FileName)
		// 	fr.FileName = strings.Replace(fr.FileName, "."+ext, ".woz", -1)
		// 	fr.Content = nb.Data.Bytes()
		// }

		if i == 0 {
			ini.WriteString("boot", "init.startfile", fr.FileName)
		} else {
			ini.WriteString("boot", fmt.Sprintf("init.auxdisk%d", i), fr.FileName)
		}

	}

	if len(files) == 1 {
		ini.WriteString("boot", "init.auxdisk1", "_.nib")
	}

	ini.WriteString("boot", "init.profile", profile)

	// Defaults
	ini.WriteString("control", "init.controlprograms", "control,control2,control3,control4")
	SavePresentationState(ini, o)

	return files

}
