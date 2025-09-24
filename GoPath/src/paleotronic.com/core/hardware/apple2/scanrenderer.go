package apple2

import (
	"image"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"time"

	"fmt"
	"paleotronic.com/core/types"

	yaml "gopkg.in/yaml.v2"
)

var startup bool
var startTime = time.Now()

const VisibleScanLines = 192
const SegCycles = 40

type ScreenScanData struct {
	ModeState [VisibleScanLines][SegCycles]ScanMode
	Main      [VisibleScanLines][SegCycles]byte
	Aux       [VisibleScanLines][SegCycles]byte
	Alt       [VisibleScanLines][SegCycles]bool
}

func (sd ScreenScanData) Pack() []byte {
	out := make([]byte, 0, VisibleScanLines*SegCycles*3)
	for y := 0; y < 192; y++ {
		for seg := 0; seg < 40; seg++ {
			out = append(
				out,
				sd.Main[y][seg],
				sd.Aux[y][seg],
				byte(sd.ModeState[y][seg])|byteBool(sd.Alt[y][seg]),
			)
		}
	}
	return out
}

func (sd *ScreenScanData) Unpack(data []byte) bool {
	if len(data) != VisibleScanLines*SegCycles*3 {
		return false
	}
	for y := 0; y < 192; y++ {
		for seg := 0; seg < 40; seg++ {
			idx := y*SegCycles + seg
			sd.Main[y][seg] = data[idx+0]
			sd.Aux[y][seg] = data[idx+1]
			sd.ModeState[y][seg] = ScanMode(data[idx+2] & 0x7f)
			sd.Alt[y][seg] = boolByte(data[idx+2])
		}
	}
	return true
}

type ScanRenderer struct {
	mem                        *memory.MemoryMap
	index                      int
	ScanData, PrevScanData     ScreenScanData //[MaxScanLines][ScanSegments]*ScanlineState
	Clock                      int64
	CyclesPerSecond            int64
	VerticalRetrace            int64
	VBlankLength               int64
	ScanCycles                 int64
	InVBlank                   bool // true if in vblank period
	FramesPerSecond            int64
	FrameCount                 int64
	FrameSkip                  int64
	ScanLine                   int
	ScanLineStart64k           int // memory offset
	LastScanMode               ScanMode
	LastScanPage               int
	ScanSegment                int
	ScanHPOS                   int
	NextFrame                  *image.RGBA
	NextFrameReady             bool
	FloatingBus                byte
	ColorBurst, LastColorBurst bool
	ClockMult                  int64
	ForceMono                  bool
	ForceRedrawAll             bool
	memMain48k                 *memory.MemoryBlock
	memAux48k                  *memory.MemoryBlock
	IO                         *Apple2IOChip
}

func NewScanRenderer(index int, mem *memory.MemoryMap, io *Apple2IOChip) *ScanRenderer {
	sr := &ScanRenderer{
		index:           index,
		mem:             mem,
		memMain48k:      mem.BlockMapper[index].Get("main.all"),
		memAux48k:       mem.BlockMapper[index].Get("aux.all"),
		IO:              io,
		ClockMult:       1,
		CyclesPerSecond: io.Clocks,
		VerticalRetrace: io.VerticalRetrace,
		VBlankLength:    io.VBlankLength,
		ScanCycles:      io.ScanCycles,
		Clock:           0,
		FrameSkip:       4,
	}
	sr.SetFPS(io.FPS)
	//log.Printf("ScanRenderer: Video config: FPS=%d, CPU=%d, Retrace=%d, VBlank=%d", sr.FramesPerSecond, sr.CyclesPerSecond, sr.VerticalRetrace, sr.VBlankLength)
	return sr
}

const (
	hBump = 1024
	vBump = 2048
)

// Returns entire memory offset
func baseOffsetWoz(x, y int) int {
	base, mx, my := baseOffsetWozModXY(x, y)
	// At this point base refers to the interlaced Quadrant base
	// address in memory, mx and my are divd by 2 to yield co-ords
	// ammmenable to the standard memory map Woz calcs
	jump := ((my % 8) * 128) + ((my / 8) * 40) + mx
	return base + jump
}

// baseOffsetModXY returns the given memory offset address,
func baseOffsetWozModXY(x, y int) (int, int, int) {
	return wozVInterlace(y)*vBump + wozHInterlace(x)*hBump, x / 2, y / 2
}

func wozVInterlace(y int) int {
	return y % 2
}

func wozHInterlace(x int) int {
	return x % 2
}

func HGRXYToOffset(x, y int) int {

	thirdOfScreen := y / 64         // 0,1,2,3 (? - vblank zone)
	textLine := y / 8               // 0 - 23
	textLineOfThird := textLine % 8 // 0 - 7
	scanLine := y % 8               // 0 - 7

	offset := (textLineOfThird * 128) + (40 * thirdOfScreen) + (1024 * scanLine)

	return offset
}

//func init() {
//	fmt.Printf("192 starts at %d\n", 8192+HGRXYToOffset(0, 194))
//	os.Exit(0)
//}

func (io *ScanRenderer) SetForceMono(b bool) {
	io.ForceMono = b
	io.ForceRedrawAll = true
}

func byteComp(a, b byte) byte {
	return a ^ b
}

func byteBool(b bool) byte {
	if b {
		return 0x80
	} else {
		return 0x00
	}
}

func boolByte(b byte) bool {
	if b&0x80 != 0 {
		return true
	}
	return false
}

func (io *ScanRenderer) SaveState() []byte {
	return io.ScanData.Pack()
}

func (io *ScanRenderer) RestoreState(data []byte) bool {
	if io.ScanData.Unpack(data) {
		io.RealSync()
	}
	return false
}

func (io *ScanRenderer) CaptureMem(vm VideoFlag, index int, mem *memory.MemoryMap, line int, xoffs int) {

	// in range check mode
	var mode = SMText40
	var page = 0
	var alt = false
	if vm&VF_SHR_ENABLE != 0 {
		// don't bother for SHR
		return
	} else if vm&VF_TEXT != 0 {
		if vm&VF_80COL != 0 {
			mode = SMText80
		} else {
			mode = SMText40
		}
	} else if vm&VF_MIXED != 0 && line >= 160 {
		if vm&VF_80COL != 0 {
			mode = SMText80
		}
	} else if vm&VF_HIRES != 0 && vm&VF_DHIRES == 0 {
		// graphics\
		mode = SMHiRES
	} else if vm&VF_HIRES != 0 && vm&VF_DHIRES != 0 {
		// graphics\
		mode = SMDoubleHiRES
	} else if vm&VF_DHIRES != 0 {
		mode = SMDoubleLoRES
	} else {
		mode = SMLoRES
	}
	if vm&VF_PAGE2 != 0 {
		page = 1 // page 2 source
	}
	if vm&VF_ALTCHAR != 0 {
		alt = true
	}

	// var hasAux = (mode == SMText80 || mode == SMDoubleLoRES || mode == SMDoubleHiRES)
	var base int
	var hblank = int(io.ScanCycles - SegCycles)
	var y int

	if mode != io.LastScanMode || xoffs == 0 || page != io.LastScanPage {

		//if mode != io.LastScanMode {
		//	log.Printf("%.3d:%.3d: %s -> %s", line, xoffs, io.LastScanMode, mode)
		//}

		if !io.ColorBurst && mode != SMText80 && mode != SMText40 {
			io.ColorBurst = true
		}

		if mode == SMText40 || mode == SMLoRES || mode == SMText80 || mode == SMDoubleLoRES {
			y = line / 8
			base = 0x400
			if page == 1 {
				base = 0x800
			}
			io.ScanLineStart64k = baseOffsetWoz(0, y*2) + base - hblank // pull back 25
		} else if mode == SMDoubleHiRES || mode == SMHiRES {
			base = 0x2000
			if page == 1 {
				base = 0x4000
			}
			io.ScanLineStart64k = HGRXYToOffset(0, line) + base - hblank // pull back 25
		}
	}

	io.LastScanMode = mode
	io.LastScanPage = page

	// now read memory
	var main, aux byte
	var offset = int(io.ScanLineStart64k + xoffs)

	if mode == SMText40 || mode == SMLoRES || mode == SMHiRES {
		main = byte(io.memMain48k.DirectRead(offset))
		aux = 0
	} else if mode == SMText80 || mode == SMDoubleLoRES || mode == SMDoubleHiRES {
		main = byte(io.memMain48k.DirectRead(offset))
		aux = byte(io.memAux48k.DirectRead(offset))
	}

	io.FloatingBus = main

	// captures state based on ScanLine, ScanHPOS
	if line >= VisibleScanLines {
		return
	}

	if xoffs < hblank {
		return
	}

	xoffs -= hblank

	// if we aren't skipping this frame, we should capture the data
	if io.FrameCount%io.FrameSkip == 0 {
		io.ScanData.Main[line][xoffs] = main
		io.ScanData.Aux[line][xoffs] = aux
		io.ScanData.ModeState[line][xoffs] = mode
		io.ScanData.Alt[line][xoffs] = alt

		if servicebus.HasReceiver(io.index, servicebus.UnifiedScanUpdate) {
			// We only do this if there is a recorder registered
			if io.PrevScanData.Main[line][xoffs] != main || io.PrevScanData.Aux[line][xoffs] != aux || io.PrevScanData.ModeState[line][xoffs] != mode || io.PrevScanData.Alt[line][xoffs] != alt {

				clock := io.Clock % (io.VBlankLength + io.VerticalRetrace)

				servicebus.SendServiceBusMessage(
					io.index,
					servicebus.UnifiedScanUpdate,
					[]byte{
						byte(line),
						byte(xoffs),
						io.PrevScanData.Main[line][xoffs],
						main,
						io.PrevScanData.Aux[line][xoffs],
						aux,
						byte(io.PrevScanData.ModeState[line][xoffs]) | byteBool(io.PrevScanData.Alt[line][xoffs]),
						byte(mode) | byteBool(alt),
						byte(clock % 256),
						byte(clock / 256),
					},
				)
			}
		}
	}
}

func (io *ScanRenderer) SetFPS(n int64) {
	if n > io.IO.FPS {
		n = io.IO.FPS
	}
	if n < 1 {
		n = 1
	}
	io.FramesPerSecond = n
	io.FrameSkip = io.IO.FPS / io.FramesPerSecond
	//log.Printf("FPS set to %d, will render every %d virtual frames", io.FramesPerSecond, io.FrameSkip)
}

// func (io *ScanRenderer) CaptureFrame(mode VideoFlag, mem *MemoryApple2) {
// 	for line := 0; line < 192; line++ {
// 		for seg := 0; seg < SegCycles; seg++ {
// 			io.CaptureMem(mode, mem, line, seg)
// 		}
// 	}
// }

func (io *ScanRenderer) Increment(n int64) bool {
	var captureNeeded bool
	defer func() {
		io.Clock += n
	}()
	modframe := io.Clock % (io.VerticalRetrace + io.VBlankLength)
	inVBlank := (modframe >= io.VerticalRetrace)
	if !inVBlank && io.InVBlank {
		if servicebus.HasReceiver(io.index, servicebus.UnifiedVBLANK) {
			servicebus.SendServiceBusMessage(
				io.index,
				servicebus.UnifiedVBLANK,
				io.FrameCount+1,
			)
		}
		// end vblank
		io.FrameCount++ // advance a frame
		if io.FrameCount%(io.FrameSkip*io.ClockMult) == 0 {
			// should generate frame
			f, ok := io.GenerateUnifiedFrame(io.NextFrame)
			//log.Printf("Generate frame @ clock == %d cycles", io.Clock)
			io.NextFrame = f
			if ok {
				io.NextFrameReady = ok // only set true, never false as we cancel frames
				settings.UnifiedRenderFrame[io.index] = io.NextFrame
				settings.UnifiedRenderChanged[io.index] = true
			}
		}
	}
	io.InVBlank = inVBlank

	//if !inVBlank {
	scanline := int((modframe) / io.ScanCycles)
	modline := int((modframe) % io.ScanCycles)
	//	if scanline >= 0 && scanline < 192 {
	if modline != io.ScanHPOS || scanline != io.ScanLine {
		captureNeeded = true
	}
	io.ScanLine = scanline
	io.ScanHPOS = modline
	//}
	//}

	return captureNeeded
}

func fastComp(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, av := range a {
		if av != b[i] {
			return false
		}
	}
	return true
}

type ScanMode int

const (
	SMText40 ScanMode = iota
	SMText80
	SMLoRES
	SMHiRES
	SMDoubleLoRES
	SMDoubleHiRES
	SMMaxMode
)

func (sm ScanMode) String() string {
	var out string
	switch sm {
	case SMText40:
		out = "TEXT40"
	case SMText80:
		out = "TEXT80"
	case SMLoRES:
		out = "LORES"
	case SMHiRES:
		out = "HIRES"
	case SMDoubleLoRES:
		out = "DLORES"
	case SMDoubleHiRES:
		out = "DHIRES"
	}
	return out
}

type ScanlineState struct {
	Line        int // scanline 0-191 for valid data
	ScanSegment int // halfline 0-1 for valid data
	Page        int
	Mode        ScanMode // mode = scan mode in use
	DataLo      []byte
	DataCkSum   byte
	// DataHi   []byte // data hi is used by 80-column card modes
}

func (sls *ScanlineState) String() string {
	return fmt.Sprintf("line: %d, half-line: %d, %s, page %d, data %d bytes", sls.Line, sls.ScanSegment, sls.Mode, sls.Page+1, len(sls.DataLo))
}

const RenderSurfaceH = 192
const RenderSurfaceW = 560

func (io *ScanRenderer) ArtificialSync(vm VideoFlag) {
	for y := 0; y < 192; y++ {
		for x := 0; x < 65; x++ {
			io.CaptureMem(vm, io.index, io.mem, y, x)
		}
	}
	f, _ := io.GenerateUnifiedFrame(io.NextFrame)
	io.NextFrame = f
}

func (io *ScanRenderer) RealSync() {
	io.ForceRedrawAll = true
	f, _ := io.GenerateUnifiedFrame(io.NextFrame)
	io.NextFrame = f
	settings.UnifiedRenderFrame[io.index] = f
	settings.UnifiedRenderChanged[io.index] = true
}

// GenerateUnifiedFrame builds a 560x192 representation of a screen :D
func (io *ScanRenderer) GenerateUnifiedFrame(frame *image.RGBA) (*image.RGBA, bool) {

	defer func() {
		io.ForceRedrawAll = false
	}()

	//if settings. {
	//	io.ColorBurst = false
	//}

	// layer, _ := io.e.GetGFXLayerByID("DHR1")
	palette := ps.Palette //layer.GetPalette()

	// We use a fixed resolution here... Everything maps to this for A2 modes
	if frame == nil {
		frame = image.NewRGBA(image.Rect(0, 0, RenderSurfaceW, RenderSurfaceH))
	}

	// process each scan/half line
	var col Color

	// log2.Printf("format of scanline 0 is %s", io.ScanData[0][0])

	// var mono bool
	// var segment *ScanlineState
	// var bits []byte
	// var emptysegment [80 / ScanSegments]byte // used to backfill missing data
	var tmpbits []byte
	var tmp []byte // colors for the scanline
	var offset int
	var buffer = make([]byte, 0, 80)
	var changeCount = 0
	var bmupdate bool
	var colorBurstChanged = (io.ColorBurst != io.LastColorBurst)

	for y := 0; y < 192; y++ {

		//log.Printf("scanline %d...", y)

		// var changed = true

		// if !io.ScanChanged[y] {
		// 	continue
		// }

		// mono = false
		var buffer = buffer[:] // buffer holds accumulated bits for the segment
		// var lastMode ScanMode

		var hgrSegs, dhgrSegs int
		var txtSegs int

		tmp = []byte{}
		var main, aux, omain, oaux byte
		var mode, omode, lastMode, lastMidSplit ScanMode
		var alt, oalt bool
		var lineRedraw bool = io.ForceRedrawAll
		var auxActive bool
		for xoffs := 0; xoffs < SegCycles; xoffs++ {
			// new values
			main = io.ScanData.Main[y][xoffs]
			aux = io.ScanData.Aux[y][xoffs]
			mode = io.ScanData.ModeState[y][xoffs]
			alt = io.ScanData.Alt[y][xoffs]
			// prev values
			omain = io.PrevScanData.Main[y][xoffs]
			oaux = io.PrevScanData.Aux[y][xoffs]
			omode = io.PrevScanData.ModeState[y][xoffs]
			oalt = io.PrevScanData.Alt[y][xoffs]

			// if omode != mode {
			// 	log.Printf("line = %d, seg = %d mode change (omode = %s, nmode = %s)", y, xoffs, omode, mode)
			// }

			// update prev frame for next time
			io.PrevScanData.Main[y][xoffs] = main
			io.PrevScanData.Aux[y][xoffs] = aux
			io.PrevScanData.ModeState[y][xoffs] = mode
			io.PrevScanData.Alt[y][xoffs] = alt

			auxActive = (mode == SMDoubleHiRES || mode == SMText80 || mode == SMDoubleLoRES)

			// detect changed line - only if we haven't already figured it has changed
			if !lineRedraw && (omode != mode || main != omain || (auxActive && aux != oaux) || (oalt != alt)) {
				lineRedraw = true
			}
			if colorBurstChanged && !lineRedraw && (mode == SMText40 || mode == SMText80) {
				lineRedraw = true
			}

			if mode == SMText40 || mode == SMText80 {
				txtSegs++
			} else if mode == SMHiRES {
				hgrSegs++
			} else if mode == SMDoubleHiRES || mode == SMDoubleLoRES {
				dhgrSegs++
			}

			if xoffs == 0 {
				lastMidSplit = mode
				tmp = tmp[:0]
			}

			if xoffs > 0 && lastMode != mode {
				//log.Printf("render mode change: y = %d, x = %d, old = %s, new = %s", y, xoffs, lastMode, mode)
				lastMidSplit = mode // save this
				// mode has changed - generate bits
				tmpbits = nil
				switch lastMode {
				case SMText40:
					tmpbits = io.ScanBitsText(y, tmp, 2, io.ColorBurst || io.ForceMono, alt)
					// txtSegs++
					// mono = true
				case SMText80:
					// txtSegs++
					tmpbits = io.ScanBitsText(y, tmp, 1, io.ColorBurst || io.ForceMono, alt)
					// mono = true
				case SMDoubleHiRES:
					tmpbits = tmp
				case SMHiRES:
					// hgrSegs++
					tmpbits = ConvertHGRToDHGRPattern(tmp, false)
				case SMLoRES:
					tmpbits = io.ScanBitsLoRES(y, tmp, 2, false)
				case SMDoubleLoRES:
					tmpbits = io.ScanBitsLoRES(y, tmp, 1, false)
				}

				buffer = append(buffer, tmpbits...)
				tmp = tmp[:0]
			}

			if auxActive {
				tmp = append(tmp, aux)
			}
			tmp = append(tmp, main)

			lastMode = mode
		}

		if len(tmp) > 0 {
			tmpbits = nil
			//if len(tmp) < 40 {
			//	log.Printf("line %d, remainder (%d) is %s: %+v", y, len(tmp), lastMidSplit, tmp)
			//}
			switch lastMidSplit {
			case SMText40:
				tmpbits = io.ScanBitsText(y, tmp, 2, io.ColorBurst || io.ForceMono, alt)
				// txtSegs++
				// mono = true
			case SMText80:
				// txtSegs++
				tmpbits = io.ScanBitsText(y, tmp, 1, io.ColorBurst || io.ForceMono, alt)
				// mono = true
			case SMDoubleHiRES:
				tmpbits = tmp
			case SMHiRES:
				// hgrSegs++
				tmpbits = ConvertHGRToDHGRPattern(tmp, false)
			case SMLoRES:
				tmpbits = io.ScanBitsLoRES(y, tmp, 2, false)
			case SMDoubleLoRES:
				tmpbits = io.ScanBitsLoRES(y, tmp, 1, false)
			}

			buffer = append(buffer, tmpbits...)

			//if len(tmp) < 40 {
			//	log.Printf("got %d tmpbits, total length = %d", len(tmpbits), len(buffer))
			//}

			tmp = tmp[:0]
		}

		if lineRedraw || io.ForceRedrawAll {
			var cfs []int
			if len(buffer) > 0 {
				if dhgrSegs == SegCycles {
					cfs = ColorsForScanLineDHGR(buffer[:80], io.ForceMono)
				} else if hgrSegs == SegCycles {
					cfs = ColorsForScanLineDHGR(buffer[:80], io.ForceMono)
					cfs = ColorFlip(cfs)
				} else if io.ColorBurst && txtSegs == SegCycles {
					cfs = ColorsForScanLineDHGR(buffer[:80], (txtSegs == SegCycles) || io.ForceMono)
					cfs = ColorFlip(cfs)
				} else {
					cfs = ColorsForScanLineDHGR(buffer[:80], (txtSegs == SegCycles) || io.ForceMono)
					cfs = ColorFlip(cfs)
				}
			}

			// log2.Printf("cfs returns %d values for line %d", len(cfs), y)

			for x, c := range cfs {
				col = palette[c]
				offset = 4*y*560 + 4*x
				frame.Pix[offset+0] = col.Red
				frame.Pix[offset+1] = col.Green
				frame.Pix[offset+2] = col.Blue
				frame.Pix[offset+3] = 255
			}

			// io.ScanChanged[y] = false // mark clean
			changeCount++
			bmupdate = true
		}

	}

	// log.Printf("colorburst = %v", io.ColorBurst)
	//
	io.LastColorBurst = io.ColorBurst
	io.ColorBurst = false // reset for next frame

	// return the completed bitmap
	return frame, bmupdate || io.ForceRedrawAll
}

func (io *ScanRenderer) ScanBitsText(line int, raw []byte, mult int, mono bool, useAlt bool) []byte {

	linemod := line % 8
	fontn, fonti := GetDefaultFont(0)
	var asc rune
	var attr types.VideoAttribute
	var cell []byte
	out := make([]byte, len(raw))
	ptr := 0
	// var b, bit6, bit5, bit4, bit3, bit2, bit1, bit0 byte
	var txt = ""
	for _, v := range raw {
		asc = rune(io.PokeToAsciiApple(uint64(v), useAlt))
		txt += string(asc)
		attr = io.PokeToAttribute(uint64(v), useAlt)
		switch attr {
		case types.VA_NORMAL:
			cell = fontn[asc]
		default:
			cell = fonti[asc]
		}
		if cell == nil {
			ptr++
			continue
		}
		out[ptr] = cell[linemod]
		if mult == 2 {
			out[ptr] |= 128
		}
		ptr++
	}

	if mult == 2 {
		out = ConvertHGRToDHGRPattern(out, mono)
	}

	//if strings.Trim(txt, " ") != "" {
	//	log.Printf("ScanBitsText: line = %d, txt = '%s', raw = %+v", line, txt, raw)
	//}

	return out
}

func (io *ScanRenderer) ScanBitsLoRES(line int, raw []byte, mult int, mono bool) []byte {
	linemod := line % 8
	out := make([]byte, len(raw)*mult)
	ptr := 0
	// var b, bit6, bit5, bit4, bit3, bit2, bit1, bit0 byte
	// var txt = ""
	var c0, c1, c, x0, x1, dc int
	for i, v := range raw {

		c0 = int(v & 0xf)
		c1 = int((v & 0xf0) >> 4)

		if mult == 1 && (i%2) == 0 {
			c0 = rol4bit(c0)
			c1 = rol4bit(c1)
		}

		c = c0
		if linemod >= 4 {
			c = c1
		}

		dc = DHGRPaletteToLores[c]
		bptable := DHGRBytePatterns[dc]

		// if c != 0 {
		// 	log2.Printf("byte patterns for %d are %+v", c, bptable)
		// }

		if mult == 2 {
			// lores
			x0 = (i*2 + 1) % 4
			x1 = (i*2 + 2) % 4
			out[ptr] = byte(bptable[x0])
			out[ptr+1] = byte(bptable[x1])
			ptr += 2
		} else {
			// x0 = (i - (i % 2) + ((i + 1) % 2)) %
			out[ptr] = byte(bptable[i%4])
			ptr++
		}
	}

	return out
}

func (t *ScanRenderer) Between(v, lo, hi uint64) bool {
	return ((v >= lo) && (v <= hi))
}

func (t *ScanRenderer) PokeToAsciiApple(v uint64, usealt bool) int {

	return t.IO.PokeToAsciiApple(v, usealt)

	highbit := v & 1024

	v = v & 1023

	if t.Between(v, 0, 31) {
		return int((64 + (v % 32)) | highbit)
	}

	if t.Between(v, 32, 63) {
		return int((32 + (v % 32)) | highbit)
	}

	if t.Between(v, 64, 95) {
		if usealt {
			return int((128 + (v % 32)) | highbit)
		} else {
			return int((64 + (v % 32)) | highbit)
		}
	}

	if t.Between(v, 96, 127) {
		if usealt {
			return int((96 + (v % 32)) | highbit)
		} else {
			return int((32 + (v % 32)) | highbit)
		}
	}

	if t.Between(v, 128, 159) {
		return int((64 + (v % 32)) | highbit)
	}

	if t.Between(v, 160, 191) {
		return int((32 + (v % 32)) | highbit)
	}

	if t.Between(v, 192, 223) {
		return int((64 + (v % 32)) | highbit)
	}

	if t.Between(v, 224, 255) {
		return int((96 + (v % 32)) | highbit)
	}

	return int(v | highbit)
}

func (t *ScanRenderer) PokeToAttribute(v uint64, usealt bool) types.VideoAttribute {

	v = v & 1023

	va := types.VA_INVERSE
	if (v & 64) > 0 {
		if usealt {
			if v < 96 {
				va = types.VA_NORMAL
			} else {
				va = types.VA_INVERSE
			}
		} else {
			va = types.VA_BLINK
		}
	}
	if (v & 128) > 0 {
		va = types.VA_NORMAL
	}
	if (v & 256) > 0 {
		va = types.VA_NORMAL
	}
	return va
}

func (io *ScanRenderer) SetForceUpdate(b bool) {
	io.ForceRedrawAll = b
}

func (io *ScanRenderer) ApplyScanDelta(line int, xoffs int, mainC byte, auxC byte, modeC byte) {
	io.ScanData.Main[line][xoffs] = mainC
	io.ScanData.Aux[line][xoffs] = auxC
	io.ScanData.Alt[line][xoffs] = boolByte(modeC)
	io.ScanData.ModeState[line][xoffs] = ScanMode(modeC & 0x7f)
}

var pdef = `
palette:
- red: 0
  green: 0
  blue: 0
  alpha: 0
- red: 224
  green: 0
  blue: 48
  alpha: 255
- red: 0
  green: 0
  blue: 128
  alpha: 255
- red: 255
  green: 0
  blue: 255
  alpha: 255
- red: 0
  green: 128
  blue: 0
  alpha: 255
- red: 128
  green: 128
  blue: 128
  alpha: 255
- red: 47
  green: 149
  blue: 229
  alpha: 255
- red: 171
  green: 171
  blue: 255
  alpha: 255
- red: 128
  green: 80
  blue: 0
  alpha: 255
- red: 255
  green: 80
  blue: 0
  alpha: 255
- red: 192
  green: 192
  blue: 192
  alpha: 255
- red: 255
  green: 161
  blue: 234
  alpha: 255
- red: 0
  green: 255
  blue: 0
  alpha: 255
- red: 255
  green: 255
  blue: 0
  alpha: 255
- red: 64
  green: 255
  blue: 144
  alpha: 255
- red: 255
  green: 255
  blue: 255
  alpha: 255
`

type Color struct {
	Red   byte
	Green byte
	Blue  byte
	Alpha byte
}

type Palette []Color

type PaletteStruct struct {
	Palette Palette
}

var ps = &PaletteStruct{}

func init() {
	err := yaml.Unmarshal([]byte(pdef), ps)
	if err != nil {
		panic(err)
	}
}
