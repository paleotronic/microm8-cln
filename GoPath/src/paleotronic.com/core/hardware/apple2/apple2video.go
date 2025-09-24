package apple2

import (
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
	"paleotronic.com/octalyzer/bus"
)

//var startup bool
//var startTime = time.Now()

const ScanSegments = 4

func (io *Apple2IOChip) Increment(n int) {

	io.CheckAudioMode()

	index := io.e.GetMemIndex()

	for i := 0; i < n; i++ {
		if io.UnifiedFrame.Increment(1) {
			// FIXME: account for soft switch delays?
			vm := io.vidmode
			if io.GlobalCycles < io.NextVideoLatch {
				vm = io.LastVidMode
			}
			io.UnifiedFrame.CaptureMem(vm, index, io.e.GetMemoryMap(), io.UnifiedFrame.ScanLine, io.UnifiedFrame.ScanHPOS)
			if settings.DebuggerActiveSlot == index {
				//log.Println("generating unified subframe")
				f, ok := io.UnifiedFrame.GenerateUnifiedFrame(io.UnifiedFrame.NextFrame)
				io.UnifiedFrame.NextFrame = f
				if ok {
					io.UnifiedFrame.NextFrameReady = ok // only set true, never false as we cancel frames
					settings.UnifiedRenderFrame[index] = io.UnifiedFrame.NextFrame
					settings.UnifiedRenderChanged[index] = true
				}
			}
		}
		io.e.IncrementRecording(1)
	}
	io.speaker.Increment(n)
	io.cassette.Increment(n)
	var c SlotCard
	for _, c = range io.cards {
		if c == nil {
			continue
		}
		c.Increment(n)
	}

	// inc cycles
	io.GlobalCycles += int64(n)

	// video cycles
	vb := io.IsVBlank()
	if vb != io.LastBlankState {
		if vb {
			// log2.Printf("vblank triggered")
			io.TriggerVideo()
		}
	}
	io.LastBlankState = vb

}

func (io *Apple2IOChip) Decrement(n int) {

}

func (io *Apple2IOChip) AdjustClock(n int) {
	//log2.Printf("Notifying speaker of speed change (%d)", n)
	io.speaker.AdjustClock(n)
	io.cassette.AdjustClock(n)
	var c SlotCard
	for _, c = range io.cards {
		if c == nil {
			continue
		}
		c.AdjustClock(n)
	}
}

func (io *Apple2IOChip) ImA() string {
	return "VideoBus"
}

func (io *Apple2IOChip) TriggerVideo() {
	//index := io.e.GetMemIndex()
	//if settings.UnifiedRender {
	//	var rc bool
	//	settings.UnifiedRenderFrame[index], rc = io.GenerateUnifiedFrame(settings.UnifiedRenderFrame[index])
	//	if rc {
	//		settings.UnifiedRenderChanged[index] = true
	//	}
	//}

	if !settings.IsRemInt {
		if io.e.GetMemoryMap().IntGetLayerState(io.e.GetMemIndex()) != 0 {
			bus.Sync()
		}
	}
	servicebus.InjectServiceBusMessage(
		io.e.GetMemIndex(),
		servicebus.BeginVBLANK,
		"",
	)
}

func (io *Apple2IOChip) IsVBlank() bool {
	total := io.VerticalRetrace + io.VBlankLength
	return io.GlobalCycles%total > io.VerticalRetrace
	//return io.UnifiedFrame.InVBlank
}

func (io *Apple2IOChip) ScrnScanPos() (hpos int, scansegment int, line int) {
	total := io.VerticalRetrace + io.VBlankLength
	offset := io.GlobalCycles % total
	// if offset >= io.VerticalRetrace {
	// 	return 0, 0
	// }
	if offset < io.VBlankLength {
		return
	}
	offset -= io.VBlankLength
	hpos, line = int(offset%65), int(offset/65)
	scansegment = hpos / (40 / ScanSegments)
	if scansegment >= ScanSegments {
		scansegment = ScanSegments - 1
	}

	return
}

func (d *Apple2IOChip) HandleServiceBusRequest(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool) {
	switch r.Type {
	case servicebus.UnifiedPlaybackSync:
		d.UnifiedFrame.ArtificialSync(d.vidmode) // whatever mode in play here
	case servicebus.Clocks6502Update:
		n, ok := r.Payload.(int)
		if ok {
			d.AdjustClock(n)
		}
	case servicebus.Cycles6502Update:
		n, ok := r.Payload.(int)
		if ok {
			d.Increment(n)
		}
	}
	return &servicebus.ServiceBusResponse{
		Payload: "",
	}, false
}

func (d *Apple2IOChip) InjectServiceBusRequest(r *servicebus.ServiceBusRequest) {
	// log.Printf("Injecting ServiceBus request: %+v", r)
	// d.Lock()
	// defer d.Unlock()
	// if d.events == nil {
	// 	d.events = make([]*servicebus.ServiceBusRequest, 0, 16)
	// }
	// d.events = append(d.events, r)
}

func (d *Apple2IOChip) HandleServiceBusInjection(handler func(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool)) {
	// if d.events == nil || len(d.events) == 0 {
	// 	return
	// }
	// d.Lock()
	// defer d.Unlock()
	// for _, r := range d.events {
	// 	if handler != nil {
	// 		handler(r)
	// 	}
	// }
	// d.events = make([]*servicebus.ServiceBusRequest, 0, 16)
}

func (d *Apple2IOChip) ServiceBusProcessPending() {
	// d.HandleServiceBusInjection(d.HandleServiceBusRequest)
}

func (d *Apple2IOChip) FullUpdate() {
	//for y := 0; y < 192; y++ {
	//	d.ScanChanged[y] = true
	//}
}

func (d *Apple2IOChip) RegisterScanUpdates(offset int, value uint64, modes []string) {
	//for _, mode := range modes {
	//	var l *types.LayerSpecMapped
	//	var ok bool
	//	l, ok = d.e.GetGFXLayerByID(mode)
	//	if !ok {
	//		l, ok = d.e.GetHUDLayerByID(mode)
	//		if !ok {
	//			continue
	//		}
	//	}
	//	d.sm.Lock()
	//	defer d.sm.Unlock()
	//	switch mode {
	//	case "HGR1", "HGR2", "DHR1", "DHR2":
	//		y := l.HControl.OffsetToScanline(offset)
	//		if y >= 0 && y <= 191 {
	//			d.ScanChanged[y] = true
	//		}
	//	case "TEXT", "TXT2", "LOGR", "LGR2", "DLGR", "DGR2":
	//		y := l.Control.OffsetToY(offset)
	//		// if mode == "TEXT" {
	//		// 	log2.Printf("TEXT, offset %d (y = %d)", offset, y)
	//		// }
	//		if y >= 0 && y <= 23 {
	//			for yy := y * 8; yy < (y+1)*8; yy++ {
	//				d.ScanChanged[yy] = true
	//			}
	//		}
	//	}
	//}
}

func (io *Apple2IOChip) CaptureScan() {
	// var ols = io.ScanData[io.LastScanLine][io.LastScanSegment]
	//var sls = &ScanlineStateIO{
	//	Line:        io.LastScanLine,
	//	ScanSegment: io.LastScanSegment,
	//	Page:        0,
	//}
	//// force update for previous non existent scan data
	//// if ols == nil {
	//// 	io.sm.Lock()
	//// 	io.ScanChanged[io.LastScanLine] = true
	//// 	io.sm.Unlock()
	//// }
	//var mode = SMText40
	//var vm = io.vidmode
	//if vm&VF_SHR_ENABLE != 0 {
	//	// don't bother for SHR
	//	return
	//} else if vm&VF_TEXT != 0 {
	//	if vm&VF_80COL != 0 {
	//		mode = SMText80
	//	} else {
	//		mode = SMText40
	//	}
	//} else if vm&VF_MIXED != 0 && io.LastScanLine >= 160 {
	//	if vm&VF_80COL != 0 {
	//		mode = SMText80
	//	}
	//} else if vm&VF_HIRES != 0 && vm&VF_80COL == 0 {
	//	// graphics\
	//	mode = SMHiRES
	//} else if vm&VF_HIRES != 0 && vm&VF_80COL != 0 {
	//	// graphics\
	//	mode = SMDoubleHiRES
	//} else if vm&VF_80COL != 0 {
	//	mode = SMDoubleLoRES
	//} else {
	//	mode = SMLoRES
	//}
	//if vm&VF_PAGE2 != 0 {
	//	sls.Page = 1 // page 2 source
	//}
	//sls.Mode = mode
	//sls.DataLo, sls.DataCkSum = io.CaptureMemory(sls.Line, sls.ScanSegment, sls.Mode, sls.Page)
	//io.ScanData[io.LastScanLine][io.LastScanSegment] = sls
	//
	//// if ols != nil && (sls.Mode != ols.Mode || !fastComp(ols.DataLo, sls.DataLo)) {
	//// 	io.sm.Lock()
	//// 	io.ScanChanged[io.LastScanLine] = true
	//// 	io.sm.Unlock()
	//// }
}

//func fastComp(a, b []byte) bool {
//	if len(a) != len(b) {
//		return false
//	}
//	for i, av := range a {
//		if av != b[i] {
//			return false
//		}
//	}
//	return true
//}

//func (io *Apple2IOChip) CaptureMemory(line, segment int, mode ScanMode, page int) (lo []byte, cksum byte) {
//	cksum = byte(line) ^ byte(segment)
//	if line >= 192 {
//		return nil, cksum
//	}
//	var l = io.ScanLayerCache[page][int(mode)]
//	if l == nil {
//		// page not in cache
//		var layer = "UNKN"
//		switch page {
//		case 0:
//			switch mode {
//			case SMText40, SMText80:
//				layer = "TEXT"
//			case SMLoRES:
//				layer = "LOGR"
//			case SMHiRES:
//				layer = "HGR1"
//			case SMDoubleHiRES:
//				layer = "DHR1"
//			case SMDoubleLoRES:
//				layer = "DLGR"
//			}
//		case 1:
//			switch mode {
//			case SMText40, SMText80:
//				layer = "TXT2"
//			case SMLoRES:
//				layer = "LGR2"
//			case SMHiRES:
//				layer = "HGR2"
//			case SMDoubleHiRES:
//				layer = "DHR2"
//			case SMDoubleLoRES:
//				layer = "DGR2"
//			}
//		}
//		var ok bool
//		if layer == "TEXT" || layer == "TXT2" {
//			// hud
//			l, ok = io.e.GetHUDLayerByID(layer)
//		} else {
//			// gfx
//			l, ok = io.e.GetGFXLayerByID(layer)
//		}
//		if !ok {
//			panic("layer " + layer + " not found")
//		}
//		io.ScanLayerCache[page][int(mode)] = l
//	}
//
//	// l is the layer accessor
//	switch mode {
//	case SMText40, SMLoRES:
//		chunk := 40 / ScanSegments
//		c := l.Control
//		s := chunk * segment
//		e := s + chunk
//		y := line / 8
//		lo = make([]byte, e-s)
//		for i := 0; i < chunk; i++ {
//			lo[i] = byte(c.GetValueXY((s+i)*2, y*2) & 0xff)
//		}
//	case SMText80, SMDoubleLoRES:
//		chunk := 80 / ScanSegments
//		c := l.Control
//		s := chunk * segment
//		e := s + chunk
//		y := line / 8
//		lo = make([]byte, e-s)
//		var x int
//		for i := 0; i < chunk; i++ {
//			x = (s + i) ^ 1
//			lo[i] = byte(c.GetValueXY(x, y*2) & 0xff)
//		}
//	case SMHiRES:
//		chunk := 40 / ScanSegments
//		h := l.HControl
//		offset := h.XYToOffset(0, line)
//		offset += (segment * chunk)
//		tmp := h.(*hires.HGRScreen).Data.ReadSlice(offset, offset+chunk)
//		lo = make([]byte, chunk)
//		for i, v := range tmp {
//			lo[i] = byte(v & 0xff)
//		}
//	case SMDoubleHiRES:
//		chunk := 40 / ScanSegments
//		h := l.HControl
//		offsetaux := h.XYToOffset(0, line) + (segment * chunk)
//		offsetmain := h.XYToOffset(7, line) + (segment * chunk)
//		tmpaux := h.(*hires.DHGRScreen).Data.ReadSlice(offsetaux, offsetaux+chunk)
//		tmpmain := h.(*hires.DHGRScreen).Data.ReadSlice(offsetmain, offsetmain+chunk)
//		lo = make([]byte, chunk*2)
//		for i, _ := range lo {
//			switch i % 2 {
//			case 0:
//				lo[i] = byte(tmpaux[i/2] & 0xff)
//			case 1:
//				lo[i] = byte(tmpmain[i/2] & 0xff)
//			}
//		}
//	}
//
//	// for _, v := range lo {
//	// 	cksum ^= v
//	// }
//	// cksum ^= byte(len(lo) & 0xff)
//
//	return lo, cksum
//}

//type ScanMode int

//const (
//	SMText40 ScanMode = iota
//	SMText80
//	SMLoRES
//	SMHiRES
//	SMDoubleLoRES
//	SMDoubleHiRES
//	SMMaxMode
//)

//func (sm ScanMode) String() string {
//	var out string
//	switch sm {
//	case SMText40:
//		out = "TEXT40"
//	case SMText80:
//		out = "TEXT80"
//	case SMLoRES:
//		out = "LORES"
//	case SMHiRES:
//		out = "HIRES"
//	case SMDoubleLoRES:
//		out = "DLORES"
//	case SMDoubleHiRES:
//		out = "DHIRES"
//	}
//	return out
//}

type ScanlineStateIO struct {
	Line        int // scanline 0-191 for valid data
	ScanSegment int // halfline 0-1 for valid data
	Page        int
	Mode        ScanMode // mode = scan mode in use
	DataLo      []byte
	DataCkSum   byte
	// DataHi   []byte // data hi is used by 80-column card modes
}

func (sls *ScanlineStateIO) String() string {
	return fmt.Sprintf("line: %d, half-line: %d, %s, page %d, data %d bytes", sls.Line, sls.ScanSegment, sls.Mode, sls.Page+1, len(sls.DataLo))
}

//const RenderSurfaceH = 192
//const RenderSurfaceW = 560

// GenerateUnifiedFrame builds a 560x192 representation of a screen :D
//func (io *Apple2IOChip) GenerateUnifiedFrame(frame *image.RGBA) (*image.RGBA, bool) {
//
//	layer, _ := io.e.GetGFXLayerByID("DHR1")
//	palette := layer.GetPalette()
//
//	// We use a fixed resolution here... Everything maps to this for A2 modes
//	if frame == nil {
//		frame = image.NewRGBA(image.Rect(0, 0, RenderSurfaceW, RenderSurfaceH))
//	}
//
//	// process each scan/half line
//	var col *types.VideoColor
//
//	// log2.Printf("format of scanline 0 is %s", io.ScanData[0][0])
//
//	// var mono bool
//	var segment, osegment *ScanlineStateIO
//	// var bits []byte
//	// var emptysegment [80 / ScanSegments]byte // used to backfill missing data
//	var emptysegment [80 / ScanSegments]byte // used to
//	var tmpbits []byte
//	var tmp []byte // colors for the scanline
//	var offset int
//	var buffer = make([]byte, 0, 80)
//	var changeCount = 0
//	io.sm.Lock()
//	defer io.sm.Unlock()
//	var changedSegments, changedLineSegments int
//	var bmupdate bool
//	for y := 0; y < 192; y++ {
//
//		// var changed = true
//
//		// if !io.ScanChanged[y] {
//		// 	continue
//		// }
//
//		// mono = false
//		var buffer = buffer[:] // buffer holds accumulated bits for the segment
//		// var lastMode ScanMode
//
//		var hgrSegs int
//		var txtSegs int
//
//		var lastMode ScanMode
//
//		tmp = []byte{}
//		changedLineSegments = 0
//		for seg := 0; seg < ScanSegments; seg++ {
//			segment = io.ScanData[y][seg]
//			osegment = io.PrevScanData[y][seg]
//			io.PrevScanData[y][seg] = segment
//
//			// if segment empty, we skip it and zero-fill
//			if segment == nil {
//				tmp = append(tmp, emptysegment[:]...)
//				// changed = true // clear it
//				changedSegments++
//				changedLineSegments++
//			} else {
//
//				if osegment == nil {
//					changedSegments++
//					changedLineSegments++
//				} else {
//					// only test for differences if no differences found
//					if osegment.Mode != segment.Mode || !fastComp(osegment.DataLo, segment.DataLo) {
//						changedSegments++
//						changedLineSegments++
//					}
//				}
//
//				// log2.Printf("line: %d, segment: %d, data-size: %d", y, seg, len(segment.DataLo))
//				if segment.Mode == SMText40 || segment.Mode == SMText80 {
//					txtSegs++
//				} else if segment.Mode == SMHiRES {
//					hgrSegs++
//				}
//
//				if segment.Mode != lastMode && len(tmp) > 0 {
//					tmpbits = nil
//					switch segment.Mode {
//					case SMText40:
//						tmpbits = io.ScanBitsText(y, tmp, 2, false)
//						// txtSegs++
//						// mono = true
//					case SMText80:
//						// txtSegs++
//						tmpbits = io.ScanBitsText(y, tmp, 1, false)
//						// mono = true
//					case SMDoubleHiRES:
//						tmpbits = tmp
//					case SMHiRES:
//						// hgrSegs++
//						tmpbits = ConvertHGRToDHGRPattern(tmp, false)
//					case SMLoRES:
//						tmpbits = io.ScanBitsLoRES(y, tmp, 2, false)
//					case SMDoubleLoRES:
//						tmpbits = io.ScanBitsLoRES(y, tmp, 1, false)
//					}
//
//					buffer = append(buffer, tmpbits...)
//					tmp = tmp[:]
//				} else {
//					tmp = append(tmp, segment.DataLo...)
//				}
//
//				// we have data..
//
//				lastMode = segment.Mode
//
//			}
//		}
//
//		if len(tmp) > 0 {
//			tmpbits = nil
//			switch segment.Mode {
//			case SMText40:
//				tmpbits = io.ScanBitsText(y, tmp, 2, false)
//				// txtSegs++
//				// mono = true
//			case SMText80:
//				// txtSegs++
//				tmpbits = io.ScanBitsText(y, tmp, 1, false)
//				// mono = true
//			case SMDoubleHiRES:
//				tmpbits = tmp
//			case SMHiRES:
//				// hgrSegs++
//				tmpbits = ConvertHGRToDHGRPattern(tmp, false)
//			case SMLoRES:
//				tmpbits = io.ScanBitsLoRES(y, tmp, 2, false)
//			case SMDoubleLoRES:
//				tmpbits = io.ScanBitsLoRES(y, tmp, 1, false)
//			}
//
//			buffer = append(buffer, tmpbits...)
//		}
//
//		if changedLineSegments > 0 {
//			var cfs []int
//			if len(buffer) > 0 {
//				if hgrSegs == ScanSegments {
//					cfs = ColorFlip(ColorsForScanLineDHGR(buffer[:80], false))
//				} else {
//					cfs = ColorsForScanLineDHGR(buffer[:80], (txtSegs == ScanSegments))
//				}
//			}
//
//			// log2.Printf("cfs returns %d values for line %d", len(cfs), y)
//
//			for x, c := range cfs {
//				col = palette.Get(c)
//				offset = 4*y*560 + 4*x
//				frame.Pix[offset+0] = col.Red
//				frame.Pix[offset+1] = col.Green
//				frame.Pix[offset+2] = col.Blue
//				frame.Pix[offset+3] = 255
//			}
//
//			// io.ScanChanged[y] = false // mark clean
//			changeCount++
//			bmupdate = true
//		}
//
//	}
//
//	// log changes
//	// if changedSegments > 0 {
//	// 	log2.Printf("frame has %d changed scansegments", changedSegments)
//	// }
//
//	// return the completed bitmap
//	return frame, bmupdate
//}

// func (io *Apple2IOChip) GenerateUnifiedDataColors() [192][560]byte {

// 	// layer, _ := io.e.GetGFXLayerByID("DHR1")
// 	// palette := layer.GetPalette()

// 	// We use a fixed resolution here... Everything maps to this for A2 modes
// 	// frame := image.NewRGBA(image.Rect(0, 0, RenderSurfaceW, RenderSurfaceH))
// 	var colors [192][560]byte

// 	// process each scan/half line
// 	var left, right *ScanlineStateIO
// 	// var col *types.VideoColor

// 	// log2.Printf("format of scanline 0 is %s", io.ScanData[0][0])

// 	for y := 0; y < 192; y++ {
// 		left = io.ScanData[y][0]
// 		right = io.ScanData[y][1]
// 		if left == nil || right == nil {
// 			continue
// 		}
// 		if left.Mode != right.Mode {
// 			continue // TODO: just for now till i figure this out
// 		}
// 		tmpleft := left.DataLo
// 		if len(left.DataHi) > 0 {
// 			tmpleft = make([]byte, len(left.DataLo)+len(left.DataHi))
// 			for i, _ := range tmpleft {
// 				switch i % 2 {
// 				case 0:
// 					tmpleft[i] = left.DataHi[i/2]
// 				case 1:
// 					tmpleft[i] = left.DataLo[i/2]
// 				}
// 			}
// 		}
// 		tmpright := right.DataLo
// 		if len(left.DataHi) > 0 {
// 			tmpright = make([]byte, len(left.DataLo)+len(left.DataHi))
// 			for i, _ := range tmpright {
// 				switch i % 2 {
// 				case 0:
// 					tmpright[i] = right.DataHi[i/2]
// 				case 1:
// 					tmpright[i] = right.DataLo[i/2]
// 				}
// 			}
// 		}
// 		var bits []byte
// 		tmp := append(tmpleft, tmpright...)
// 		switch left.Mode {
// 		case SMText40:
// 			bits = io.ScanBitsText(y, tmp, 2, false)
// 		case SMText80:
// 			bits = io.ScanBitsText(y, tmp, 1, false)
// 		case SMDoubleHiRES:
// 			bits = tmp
// 		case SMHiRES:
// 			bits = ConvertHGRToDHGRPattern(bits, false)
// 		case SMLoRES:
// 			bits = io.ScanBitsLoRES(y, tmp, 2, false)
// 		case SMDoubleLoRES:
// 			bits = io.ScanBitsLoRES(y, tmp, 1, false)
// 		}

// 		if bits == nil {
// 			continue
// 		}

// 		if len(bits) > 0 {
// 			// got bitpattern
// 			var cfs []int

// 			if left.Mode == SMHiRES {
// 				cfs = ColorsForScanLineDHGR(bits, false)
// 				for i, v := range cfs {
// 					cfs[i] = 8 + v
// 				}
// 			} else {
// 				cfs = ColorsForScanLineDHGR(bits, false)
// 			}
// 			// var offset int
// 			// log2.Printf("scanline %d length = %d", y, len(cfs))

// 			for x, c := range cfs {
// 				colors[y][x] = byte(c)
// 				// col = palette.Get(c)
// 				// offset = 4*y*560 + 4*x
// 				// frame.Pix[offset+0] = col.Red
// 				// frame.Pix[offset+1] = col.Green
// 				// frame.Pix[offset+2] = col.Blue
// 				// frame.Pix[offset+3] = 255
// 			}
// 		}
// 	}

// 	// return the completed bitmap
// 	return colors
// }

func (io *Apple2IOChip) ScanBitsText(line int, raw []byte, mult int, mono bool) []byte {

	// log2.Printf("sbt: recv %d bytes", len(raw))

	linemod := line % 8
	fontn, fonti := GetDefaultFont(io.e.GetMemIndex())
	var asc rune
	var attr types.VideoAttribute
	var cell []byte
	out := make([]byte, len(raw))
	ptr := 0
	// var b, bit6, bit5, bit4, bit3, bit2, bit1, bit0 byte
	var txt = ""
	for _, v := range raw {
		asc = rune(io.PokeToAsciiApple(uint64(v), false))
		txt += string(asc)
		attr = io.PokeToAttribute(uint64(v), false)
		switch attr {
		case types.VA_NORMAL:
			cell = fontn[asc]
		default:
			cell = fonti[asc]
		}
		if cell == nil {
			ptr += mult
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

	return out
}

func (io *Apple2IOChip) ScanBitsLoRES(line int, raw []byte, mult int, mono bool) []byte {
	linemod := line % 8
	out := make([]byte, len(raw)*mult)
	ptr := 0
	// var b, bit6, bit5, bit4, bit3, bit2, bit1, bit0 byte
	// var txt = ""
	var c0, c1, c, dc, x0, x1 int
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
			x0 = (i*2 + 0) % 4
			x1 = (i*2 + 1) % 4
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
