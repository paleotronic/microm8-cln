package apple2

import (
	"bytes"
	"errors"
	"fmt"
	log2 "log"
	"os"
	"strings"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"

	"paleotronic.com/log"
	//"log"

	"paleotronic.com/core/settings"
	"paleotronic.com/files"
)

/*
Prefix Format

0000-0003: 32 49 4D 47   "2IMG"    ID for 2MG format (ASCII Text)
0004-0007: 58 47 53 21   "XGS!"    Creator ID (ASCII Text) **
0008-0009: 40 00                   Header size ($0040= 64 bytes)
000A-000B: 01 00                   Version number
000C-000F: 01 00 00 00             Image Format
                                    00= DOS 3.3 sector order
                                    01= ProDOS sector order
                                    02= NIB data

**Note: "Creator" refers to application creating the image.
Here are ID's in use by various applications:

ASIMOV2-              "!nfc"
Bernie ][ the Rescue- "B2TR"
Catakig-              "CTKG"
Sheppy's ImageMaker-  "ShIm"
Sweet 16-             "WOOF"
XGS-                  "XGS!"



0010-0013: 00 00 00 00  (Flags & DOS 3.3 Volume Number)

The four-byte flags field contains bit flags and data relating to
the disk image. Bits not defined should be zero.

Bit   Description

31    Locked? If Bit 31 is 1 (set), the disk image is locked. The
      emulator should allow no changes of disk data-- i.e. the disk
      should be viewed as write-protected.

8     DOS 3.3 Volume Number? If Bit 8 is 1 (set), then Bits 0-7
      specify the DOS 3.3 Volume Number. If Bit 8 is 0 and the
      image is in DOS 3.3 order (Image Format = 0), then Volume
      Number will be taken as 254.

7-0   The DOS 3.3 Volume Number, usually 1 through 254,
      if Bit 8 is 1 (set). Otherwise, these bits should be 0.


0014-0017: 18 01 00 00  (ProDOS Blocks = 280 for 5.25")

The number of 512-byte blocks in the disk image- this value
should be zero unless the image format is 1 (ProDOS order).
Note: ASIMOV2 sets to $118 whether or not format is ProDOS.


0018-001B: 40 00 00 00  (Offset to disk data = 64 bytes)

Offset to the first byte of disk data in the image file
from the beginning of the file- disk data must come before
any Comment and Creator-specific chunks.


001C-001F: 00 30 02 00  (Bytes of disk data = 143,360 for 5.25")

Length of the disk data in bytes. (For ProDOS should be
512 x Number of blocks)


0020-0023: 00 00 00 00  (Offset to optional Comment)

Offset to the first byte of the image Comment- zero if there
is no Comment. The Comment must come after the data chunk,
but before the creator-specific chunk. The Comment, if it
exists, should be raw text; no length byte or C-style null
terminator byte is required (that's what the next field is for).


0024-0027: 00 00 00 00  (Length of optional Comment)

Length of the Comment chunk- zero if there's no Comment.


0028-002B: 00 00 00 00  (Offset to optional Creator data)

Offset to the first byte of the Creator-specific data chunk-
zero if there is none.


002C-002F: 00 00 00 00  (Length of optional Creator data)

Length of the Creator-specific data chunk- zero if there is no
creator-specific data.


0030-003F: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
*/

var magic2MG = []byte{0x32, 0x49, 0x4D, 0x47}

type SmartPortBlockDevice struct {
	Data []byte
	//w        *disk.DSKWrapper
	pdboot     []byte
	err        error
	NoHeader   bool
	Filename   string
	Updated    bool
	e          interfaces.Interpretable
	BlockCount int
}

func NewSmartPortBlockDeviceNoHeader(data []byte, filename string) (*SmartPortBlockDevice, error) {
	d := &SmartPortBlockDevice{
		Data:     data,
		NoHeader: true,
		Filename: filename,
	}
	d.BlockCount = len(d.Data) / 512
	d.BlockCount = d.GetTotalBlocks()
	d.DumpInfo()
	log.Printf("")
	return d, nil
}

func NewSmartPortBlockDevice(data []byte, filename string) (*SmartPortBlockDevice, error) {
	d := &SmartPortBlockDevice{
		Data:     data,
		Filename: filename,
	}
	if bytes.Compare(magic2MG, d.GetMagic()) != 0 {
		return nil, errors.New("Bad magic: not a 2MG file")
	}
	log.Println("Mounting volume")
	d.BlockCount = d.GetBlockCount()
	d.DumpInfo()
	return d, nil
}

func (s *SmartPortBlockDevice) GetMagic() []byte {
	if s.NoHeader {
		return []byte("DISK")
	}
	return s.Data[:4]
}

func (s *SmartPortBlockDevice) GetCreator() []byte {
	if s.NoHeader {
		return []byte("None")
	}
	return s.Data[4:8]
}

func (s *SmartPortBlockDevice) getInt16(offs int) int {
	return int(s.Data[offs]) | (int(s.Data[offs+1]) << 8)
}

func (s *SmartPortBlockDevice) getInt32(offs int) int {
	return int(s.Data[offs]) | (int(s.Data[offs+1]) << 8) | (int(s.Data[offs+2]) << 16) | (int(s.Data[offs+3]) << 24)
}

func (s *SmartPortBlockDevice) GetHeaderSize() int {
	if s.NoHeader {
		return 0
	}
	return s.getInt16(8)
}

func (s *SmartPortBlockDevice) GetVersionNumber() int {
	if s.NoHeader {
		return 1
	}
	return s.getInt16(0xa)
}

func (s *SmartPortBlockDevice) GetImageFormat() int {
	if s.NoHeader {
		return 1
	}
	return s.getInt32(0xc)
}

func (s *SmartPortBlockDevice) GetFlags() (locked bool, hasVolume bool, vnum int) {
	if s.NoHeader {
		return false, false, 0
	}
	f := s.getInt32(0x10)
	locked = (f >> 31) != 0
	hasVolume = f&(1<<8) != 0
	vnum = f & 0xff
	return locked, hasVolume, vnum
}

func (s *SmartPortBlockDevice) GetBlockCount() int {
	if s.NoHeader {
		return len(s.Data) / 512
	}
	return s.getInt32(0x14)
}

func (s *SmartPortBlockDevice) GetDataOffset() int {
	if s.NoHeader {
		return 0
	}
	return s.getInt32(0x18)
}

func (s *SmartPortBlockDevice) GetDataSize() int {
	if s.NoHeader {
		return len(s.Data)
	}
	return s.getInt32(0x1C)
}

func (s *SmartPortBlockDevice) GetCommentOffset() int {
	if s.NoHeader {
		return 0
	}
	return s.getInt32(0x20)
}

func (s *SmartPortBlockDevice) GetCommentSize() int {
	if s.NoHeader {
		return 0
	}
	return s.getInt32(0x24)
}

func (s *SmartPortBlockDevice) GetCreatorOffset() int {
	if s.NoHeader {
		return 0
	}
	return s.getInt32(0x28)
}

func (s *SmartPortBlockDevice) GetCreatorSize() int {
	if s.NoHeader {
		return 0
	}
	return s.getInt32(0x2C)
}

func (s *SmartPortBlockDevice) GetReserved() []byte {
	if s.NoHeader {
		return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	}
	return s.Data[0x30:0x40]
}

func (s *SmartPortBlockDevice) GetTotalBlocks() int {
	data, _ := s.GetRawBlock(2)
	return int(data[4+37]) + 256*int(data[4+38])
}

func (s *SmartPortBlockDevice) DumpInfo() {
	log.Println("SmartPort Device Info:")
	log.Printf("Creator      : %s", string(s.GetCreator()))
	log.Printf("Version      : %d", s.GetVersionNumber())
	log.Printf("Image format : %d", s.GetImageFormat())
	log.Printf("ProDOS Blocks: %d", s.GetTotalBlocks())
	log.Printf("ProDOS Start : %d", s.GetDataOffset())
	log.Printf("ProDOS Bytes : %d", s.GetDataSize())
	locked, hasVolume, vnum := s.GetFlags()
	log.Printf("Locked?      : %v", locked)
	log.Printf("Has DOS Vol  : %v", hasVolume)
	if hasVolume {
		log.Printf("DOS Volume   : %d", vnum)
	}
	log.Printf("PDBoot length: %d", len(s.pdboot))
	log.Printf("PD status    : %v", s.err)
	//os.Exit(1)
}

func (s *SmartPortBlockDevice) GetRawBlock(b int) ([]byte, error) {
	if b < 0 || b >= s.BlockCount {
		return nil, fmt.Errorf("Invalid block reference %d", b)
	}
	start := s.GetDataOffset() + b*512
	end := start + 512
	return s.Data[start:end], nil
	//return s.w.PRODOS800GetBlock(b)
}

func (s *SmartPortBlockDevice) grow(b int) {
	neededSize := (b + 1) * 512
	for len(s.Data) < neededSize {
		log.Printf("Growing the disk as requested..")
		tmp := make([]byte, 512)
		s.Data = append(s.Data, tmp...)
	}
}

func (s *SmartPortBlockDevice) SetRawBlock(b int, data []byte) error {
	if b < 0 || b >= s.BlockCount {
		return fmt.Errorf("Invalid block reference %d", b)
	}
	locked, _, _ := s.GetFlags()
	if locked {
		return errors.New("Device is read-only")
	}
	if len(data) != 512 {
		return errors.New("Invalid block size")
	}
	s.grow(b)
	start := s.GetDataOffset() + b*512
	for i, v := range data {
		s.Data[start+i] = v
	}
	log.Printf("--> Wrote block %d", b)
	s.Updated = true
	return nil
}

// func (s *SmartPortBlockDevice) MountVolume() error {

// 	offset := s.GetDataOffset()
// 	size := s.GetDataSize()
// 	s.w, s.err = disk.NewDSKWrapperBin(nil, s.Data[offset:offset+size], "SmartPort")

// 	return s.err
// }

// func (s *SmartPortBlockDevice) FindPRODOS() ([]byte, error) {
// 	_, info, err := s.w.PRODOSGetCatalog(2, "*.*")
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, fd := range info {
// 		//log.Printf("Found %s, %s", v.NameUnadorned(), v.Name())
// 		if fd.Name() == "prodos.SYS" {
// 			_, _, data, err := s.w.PRODOSReadFile(fd)
// 			return data, err
// 		}
// 	}
// 	return nil, errors.New("Could not find PRODOS")
// }

// func (s *SmartPortBlockDevice) FoundPRODOS() bool {
// 	return s.err == nil && len(s.pdboot) > 0
// }

// func (s *SmartPortBlockDevice) GetPRODOSBootstrap() []byte {
// 	return s.pdboot
// }

func (s *SmartPortBlockDevice) CheckUpdate() {
	if !s.Updated {
		return
	}

	if strings.HasPrefix(s.Filename, "/appleii") {
		s.Filename = "/local/mydisks/" + files.GetFilename(s.Filename)
	}

	log.Printf("Updating volume to %s...", s.Filename)
	settings.PureBootSmartVolume[s.e.GetMemIndex()] = s.Filename

	if strings.HasPrefix(s.Filename, "local:") {
		fn := s.Filename[6:]
		f, err := os.Create(fn)
		if err == nil {
			defer f.Close()
			f.Write(s.Data)
		}
	} else {
		err := files.WriteBytesViaProvider(files.GetPath(s.Filename), files.GetFilename(s.Filename), s.Data)
		if err != nil {
			apple2helpers.OSDShow(s.e, err.Error())
			log2.Printf("Write of 2mg/hdv failed with: %v", err)
		}
	}
	log.Printf("Updated disk %s", s.Filename)
	log.Printf("Buffer is %d bytes", len(s.Data))

	s.Updated = false
}
