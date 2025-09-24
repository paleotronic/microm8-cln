package vduproto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sync"

	"paleotronic.com/log"
	"paleotronic.com/core/types"
)

type ThinScreenEventType int

const (
	ToggleHGR     ThinScreenEventType = 100 + iota
	Plot2D        ThinScreenEventType = 100 + iota
	Plot3D        ThinScreenEventType = 100 + iota
	Line2D        ThinScreenEventType = 100 + iota
	Line3D        ThinScreenEventType = 100 + iota
	Fill2D        ThinScreenEventType = 100 + iota
	TextCell      ThinScreenEventType = 100 + iota
	ClearScreen   ThinScreenEventType = 100 + iota
	ScrollScreen  ThinScreenEventType = 100 + iota
	CurrentPage   ThinScreenEventType = 100 + iota
	DisplayPage   ThinScreenEventType = 100 + iota
	HGRScanLine   ThinScreenEventType = 100 + iota
	BGColorSet    ThinScreenEventType = 100 + iota
	CamReset      ThinScreenEventType = 100 + iota
	CamPos        ThinScreenEventType = 100 + iota
	CamLocRel     ThinScreenEventType = 100 + iota
	CamPivPnt     ThinScreenEventType = 100 + iota
	CamDolly      ThinScreenEventType = 100 + iota
	CamZoom       ThinScreenEventType = 100 + iota
	CamLock       ThinScreenEventType = 100 + iota
	CamMove       ThinScreenEventType = 100 + iota
	CamRotate     ThinScreenEventType = 100 + iota
	CamOrbit      ThinScreenEventType = 100 + iota
	ToggleControl ThinScreenEventType = 100 + iota
	TextPut		  ThinScreenEventType = 100 + iota
	TextClear     ThinScreenEventType = 100 + iota
)

type ThinScreenEvent struct {
	ID         ThinScreenEventType
	X0, Y0, Z0 int
	X1, Y1, Z1 int
	C, W, H    int
	F          float32
	FX, FY, FZ float32
	LayerID    byte
	Data       []int
}

type ThinScreenEventList []ThinScreenEvent

type ThinScreenEventBuffer struct {
	mutex  sync.Mutex
	events ThinScreenEventList
}

// Helper for converting floats to []byte
func float2Slice(f float32) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, f)
	if err != nil {
		return []byte(nil)
	}
	return buf.Bytes()
}

func slice2Float(data []byte) float32 {
	var f float32
	b := bytes.NewBuffer(data)
	_ = binary.Read(b, binary.LittleEndian, &f)
	return f
}

func NewThinScreenEventBuffer() *ThinScreenEventBuffer {
	this := &ThinScreenEventBuffer{events: make([]ThinScreenEvent, 0)}
	return this
}

func (this *ThinScreenEventBuffer) Add(ev ThinScreenEvent) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.events = append(this.events, ev)
}

func (this *ThinScreenEventBuffer) Clear() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.events = make([]ThinScreenEvent, 0)
}

func (this *ThinScreenEventBuffer) GetEvents() ThinScreenEventList {

	this.mutex.Lock()
	defer this.mutex.Unlock()

	d := this.events
	this.events = make([]ThinScreenEvent, 0)

	return d
}

func (this *ThinScreenEventBuffer) ToggleControls() {
	this.Add(ThinScreenEvent{ID: ToggleControl})
}

func (this *ThinScreenEventBuffer) ToggleHGR(v bool) {
	c := 0
	if v {
		c = 1
	}
	this.Add(ThinScreenEvent{ID: ToggleHGR, C: c})
}

func (this *ThinScreenEventBuffer) CamDolly(c float32) {
	this.Add(ThinScreenEvent{F: c, ID: CamDolly})
}

func (this *ThinScreenEventBuffer) CamZoom(c float32) {
	this.Add(ThinScreenEvent{F: c, ID: CamZoom})
}

func (this *ThinScreenEventBuffer) CamMove(x, y, z float32) {
	this.Add(ThinScreenEvent{FX: x, FY: y, FZ: z, ID: CamMove})
}

func (this *ThinScreenEventBuffer) CamOrbit(x, y float32) {
	this.Add(ThinScreenEvent{FX: x, FY: y, FZ: 0, ID: CamOrbit})
}

func (this *ThinScreenEventBuffer) CamRotate(x, y, z float32) {
	this.Add(ThinScreenEvent{FX: x, FY: y, FZ: z, ID: CamRotate})
}

func (this *ThinScreenEventBuffer) CamPivPnt(x, y, z float32) {
	this.Add(ThinScreenEvent{FX: x, FY: y, FZ: z, ID: CamPivPnt})
}

func (this *ThinScreenEventBuffer) CamPos(x, y, z float32) {
	this.Add(ThinScreenEvent{FX: x, FY: y, FZ: z, ID: CamPos})
}

func (this *ThinScreenEventBuffer) CamReset(c int) {
	this.Add(ThinScreenEvent{C: c, ID: CamReset})
}

func (this *ThinScreenEventBuffer) CamLock(c int) {
	this.Add(ThinScreenEvent{C: c, ID: CamLock})
}

func (this *ThinScreenEventBuffer) DisplayPage(c int) {
	this.Add(ThinScreenEvent{C: c, ID: DisplayPage})
}

func (this *ThinScreenEventBuffer) CurrentPage(c int) {
	this.Add(ThinScreenEvent{C: c, ID: CurrentPage})
}

func (this *ThinScreenEventBuffer) SetBGColor(r int) {
	this.Add(ThinScreenEvent{X0: r, ID: BGColorSet})
}

func (this *ThinScreenEventBuffer) Fill2DFull(l byte, c int) {
	this.Add(ThinScreenEvent{C: c, ID: Fill2D, LayerID: l})
}

func (this *ThinScreenEventBuffer) Plot2D(l byte, x, y, c int) {
	this.Add(ThinScreenEvent{X0: x, Y0: y, C: c, ID: Plot2D, LayerID: l})
}

func (this *ThinScreenEventBuffer) Line2D(l byte, x0, y0, x1, y1, c int) {
	this.Add(ThinScreenEvent{X0: x0, Y0: y0, X1: x1, Y1: y1, C: c, ID: Line2D, LayerID: l})
}

func (this *ThinScreenEventBuffer) Fill2D(l byte, x0, y0, x1, y1, c int) {
	this.Add(ThinScreenEvent{X0: x0, Y0: y0, X1: x1, Y1: y1, C: c, ID: Fill2D, LayerID: l})
}

func (this *ThinScreenEventBuffer) TextCell(x, y, w, h int, data []int) {
	this.Add(ThinScreenEvent{X0: x, Y0: y, W: w, H: h, Data: data, ID: TextCell})
}

func (this *ThinScreenEventBuffer) TextPut( layer, x, y int, ch rune, fg, bg int, attr types.VideoAttribute, w, h int ) {
	this.Add( ThinScreenEvent{
		ID: TextPut,
		X0: x,
		Y0: y,
		Z0: layer,
		C: int(ch),
		X1: fg,
		Y1: bg,
		Z1: int(attr),
		W: w,
		H: h,
	} )
}

func (this *ThinScreenEventBuffer) ScanLine(y int, data []int) {
	this.Add(ThinScreenEvent{Y0: y, Data: data, ID: HGRScanLine})
}

func (this *ThinScreenEventBuffer) ClearScreen(c int) {
	this.Add(ThinScreenEvent{C: c, ID: ClearScreen})
}

func (this *ThinScreenEvent) MarshalBinary() ([]byte, error) {
	switch this.ID {
	case TextPut:
		s := []byte{
						byte(this.ID),
						byte(this.X0),
						byte(this.Y0),
						byte(this.Z0),
						byte(this.X1),
						byte(this.Y1),
						byte(this.Z1),
						byte(this.C),
						byte(this.W),
						byte(this.H),
		}
		return s, nil
	case ToggleControl:
		return []byte{byte(this.ID)}, nil
	case ToggleHGR:
		return []byte{byte(this.ID), byte(this.C)}, nil
	case CamRotate:
		x := float2Slice(this.FX)
		y := float2Slice(this.FY)
		z := float2Slice(this.FZ)
		s := append([]byte{byte(this.ID)}, x...)
		s = append(s, y...)
		s = append(s, z...)
		return s, nil
	case CamMove:
		x := float2Slice(this.FX)
		y := float2Slice(this.FY)
		z := float2Slice(this.FZ)
		s := append([]byte{byte(this.ID)}, x...)
		s = append(s, y...)
		s = append(s, z...)
		return s, nil
	case CamOrbit:
		x := float2Slice(this.FX)
		y := float2Slice(this.FY)
		z := float2Slice(this.FZ)
		s := append([]byte{byte(this.ID)}, x...)
		s = append(s, y...)
		s = append(s, z...)
		return s, nil
	case CamPivPnt:
		x := float2Slice(this.FX)
		y := float2Slice(this.FY)
		z := float2Slice(this.FZ)
		s := append([]byte{byte(this.ID)}, x...)
		s = append(s, y...)
		s = append(s, z...)
		return s, nil
	case CamPos:
		x := float2Slice(this.FX)
		y := float2Slice(this.FY)
		z := float2Slice(this.FZ)
		s := append([]byte{byte(this.ID)}, x...)
		s = append(s, y...)
		s = append(s, z...)
		return s, nil
	case CamZoom:
		b := float2Slice(this.F)
		return append([]byte{byte(this.ID)}, b...), nil
	case CamDolly:
		b := float2Slice(this.F)
		return append([]byte{byte(this.ID)}, b...), nil
	case CamReset:
		return []byte{byte(this.ID), byte(this.C)}, nil
	case CamLock:
		return []byte{byte(this.ID), byte(this.C)}, nil
	case BGColorSet:
		return []byte{byte(this.ID), byte(this.X0)}, nil
	case CurrentPage:
		return []byte{byte(this.ID), byte(this.C)}, nil
	case DisplayPage:
		return []byte{byte(this.ID), byte(this.C)}, nil
	case Fill2D:
		return []byte{byte(this.ID), byte(this.C), this.LayerID}, nil
	case Plot2D:
		return []byte{byte(this.ID), byte(this.X0 % 256), byte(this.X0 / 256), byte(this.Y0), byte(this.C), this.LayerID}, nil
	case Line2D:
		return []byte{byte(this.ID), byte(this.X0 % 256), byte(this.X0 / 256), byte(this.Y0), byte(this.X1 % 256), byte(this.X1 / 256), byte(this.Y1), byte(this.C), this.LayerID}, nil
	}
	return make([]byte, 0), errors.New("Unable to pack ThinScreenEvent (unrecognized)")
}

func (this *ThinScreenEvent) UnmarshalBinary(data []byte) error {
	switch ThinScreenEventType(data[0]) {
	case TextPut:
		this.ID = TextPut
		if len(data) < 10 {
			return errors.New("Not enough data")
		}
		this.X0 = int(data[1])
		this.Y0 = int(data[2])
		this.Z0 = int(data[3])
		this.X1 = int(data[4])
		this.Y1 = int(data[5])
		this.Z1 = int(data[6])
		this.C  = int(data[7])
		this.W  = int(data[8])
		this.H  = int(data[9])
		return nil
	case ToggleControl:
		this.ID = ToggleControl
		return nil
	case ToggleHGR:
		if len(data) < 2 {
			return errors.New("not enough data")
		}
		this.ID = ToggleHGR
		this.C = int(data[1])
		return nil
	case CamRotate:
		{
			if len(data) < 13 {
				return errors.New("not enough data")
			}
			this.ID = CamRotate
			this.FX = slice2Float(data[1:5])
			this.FY = slice2Float(data[5:9])
			this.FZ = slice2Float(data[9:13])

			return nil
		}
	case CamOrbit:
		{
			if len(data) < 13 {
				return errors.New("not enough data")
			}
			this.ID = CamOrbit
			this.FX = slice2Float(data[1:5])
			this.FY = slice2Float(data[5:9])
			this.FZ = slice2Float(data[9:13])

			return nil
		}
	case CamMove:
		{
			if len(data) < 13 {
				return errors.New("not enough data")
			}
			this.ID = CamMove
			this.FX = slice2Float(data[1:5])
			this.FY = slice2Float(data[5:9])
			this.FZ = slice2Float(data[9:13])

			return nil
		}
	case CamPivPnt:
		{
			if len(data) < 13 {
				return errors.New("not enough data")
			}
			this.ID = CamPivPnt
			this.FX = slice2Float(data[1:5])
			this.FY = slice2Float(data[5:9])
			this.FZ = slice2Float(data[9:13])

			return nil
		}
	case CamPos:
		{
			if len(data) < 13 {
				return errors.New("not enough data")
			}
			this.ID = CamPos
			this.FX = slice2Float(data[1:5])
			this.FY = slice2Float(data[5:9])
			this.FZ = slice2Float(data[9:13])

			return nil
		}
	case CamZoom:
		{
			if len(data) < 5 {
				return errors.New("not enough data")
			}
			this.ID = CamZoom
			this.F = slice2Float(data[1:])

			return nil
		}
	case CamDolly:
		{
			if len(data) < 5 {
				return errors.New("not enough data")
			}
			this.ID = CamDolly
			this.F = slice2Float(data[1:])

			return nil
		}
	case BGColorSet:
		{
			if len(data) < 2 {
				return errors.New("not enough data")
			}
			this.ID = BGColorSet
			this.X0 = int(data[1])

			return nil
		}
	case DisplayPage:
		{
			if len(data) < 2 {
				return errors.New("not enough data")
			}
			this.ID = DisplayPage
			this.C = int(data[1])
			return nil
		}
	case CamReset:
		{
			if len(data) < 2 {
				return errors.New("not enough data")
			}
			this.ID = CamReset
			this.C = int(data[1])
			return nil
		}
	case CamLock:
		{
			if len(data) < 2 {
				return errors.New("not enough data")
			}
			this.ID = CamLock
			this.C = int(data[1])
			return nil
		}
	case CurrentPage:
		{
			if len(data) < 2 {
				return errors.New("not enough data")
			}
			this.ID = CurrentPage
			this.C = int(data[1])
			return nil
		}
	case Fill2D:
		{
			if len(data) < 3 {
				return errors.New("not enough data")
			}
			this.ID = Fill2D
			this.C = int(data[1])
			this.LayerID = data[2]
			return nil
		}
	case Plot2D:
		{
			if len(data) < 6 {
				return errors.New("not enough data")
			}
			this.ID = Plot2D
			this.X0 = int(data[1]) + 256*int(data[2])
			this.Y0 = int(data[3])
			this.C = int(data[4])
			this.LayerID = data[5]
			return nil
		}
	case Line2D:
		{
			if len(data) < 9 {
				return errors.New("not enough data")
			}
			this.ID = Line2D
			this.X0 = int(data[1]) + 256*int(data[2])
			this.Y0 = int(data[3])
			this.X1 = int(data[4]) + 256*int(data[5])
			this.Y1 = int(data[6])
			this.C = int(data[7])
			this.LayerID = data[8]
			return nil
		}
	}
	return nil
}

func (this ThinScreenEventList) MarshalBinary() ([]byte, error) {

	data := []byte{byte(types.MtThinScreen)}

	data = append(data, byte(len(this)%256), byte(len(this)/256))

	for _, d := range this {
		chunk, _ := d.MarshalBinary()
		data = append(data, byte(len(chunk)))
		data = append(data, chunk...)
	}

	return data, nil

}

func (this *ThinScreenEventList) UnmarshalBinary(data []byte) error {
	/*
		Offset	Length 				Desc
		======	======				====
		0		types.MtThinScreen	ID for this chunk
		1		2					16 bit int #messages (lo, hi)
		------  ------              ----
		3		1					Length of next chunk
		4...    ?				    chunk data
		?		1					Length of next chunk
		?...	?					chunk
	*/

	if data[0] != types.MtThinScreen {
		return errors.New("wrong type")
	}

	log.Printf("Got thinscreenlist type")

	dataptr := 1

	// number of chunks
	numchunks := int(data[dataptr]) + 256*int(data[dataptr+1])
	dataptr += 2

	log.Printf("This package contains %d messages", numchunks)

	count := 0

	for count < numchunks {
		if dataptr >= len(data) {
			return errors.New("out of data")
		}
		chunklen := int(data[dataptr])
		dataptr++
		count++

		// Process chunk
		s := data[dataptr : dataptr+chunklen]
		dataptr += chunklen

		t := &ThinScreenEvent{}
		err := t.UnmarshalBinary(s)
		if err != nil {
			return err
		}
		*this = append(*this, *t)

	}

	return nil
}
