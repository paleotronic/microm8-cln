package common

import (
	"log"
	"math/bits"
	"strconv"
	"time"

	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"

	"paleotronic.com/fmt"
)

type ImageWriterIIDevice struct {
	w, h               float64 // inches
	pagehdpi, pagevdpi float64
	lineSpacing        float64
	formSpacing        float64
	formLines          float64
	formSkipBottom     float64
	xpos, ypos         float64
	hdpi, vdpi         float64
	hcpi               float64
	flushInterval      time.Duration
	marginLeft         float64
	marginRight        float64
	marginTop          float64
	running            bool
	in                 chan byte
	terminate          chan bool
	autoFF             bool
	lastActivity       time.Time
	style              fxFontStyle
	printColor         fxColor
	ent                interfaces.Interpretable
}

func NewImageWriterIIDevice(driver OutputDriver, ent interfaces.Interpretable) *ImageWriterIIDevice {
	d := &ImageWriterIIDevice{}
	d.in = make(chan byte, 24576)
	d.terminate = make(chan bool)
	d.ent = ent
	d.Reset()

	// start processing command codes with the specified driver
	go d.Process(driver)

	return d
}

func (d *ImageWriterIIDevice) BufferCount() int {
	return len(d.in)
}

func (d *ImageWriterIIDevice) Write(b []byte) (int, error) {
	for _, bb := range b {
		d.in <- bb
	}
	return len(b), nil
}

func (d *ImageWriterIIDevice) Close() error {
	// d.terminate <- true
	d.running = false
	// close(d.in)
	return nil
}

// Reset resets the Epson emulation to page level defaults
func (d *ImageWriterIIDevice) Reset() {
	d.lineSpacing = float64(1) / 6
	d.marginLeft = 0.250
	d.marginTop = 0.0
	d.xpos = d.marginLeft
	d.ypos = d.marginTop
	d.hdpi = 160
	d.vdpi = 72
	d.hcpi = 12
	d.marginRight = float64(1) / d.hcpi * 80
	d.running = true
	d.autoFF = true
	d.lastActivity = time.Now()
	d.w = 8.5
	d.h = 11
	d.pagehdpi = 160
	d.pagevdpi = 72
	d.formSpacing = d.h
	d.formSkipBottom = 0
	d.formLines = 66
	d.style = fxFontStyle{
		imgWriter: true,
	}
	d.printColor = fxColorBlack
}

// Stop terminates the spool action
func (d *ImageWriterIIDevice) Stop() {
	d.running = false
}

func (d *ImageWriterIIDevice) readByte() byte {
	// time.Sleep(2 * time.Millisecond)
	return <-d.in
}

func (d *ImageWriterIIDevice) readWord() int {
	// time.Sleep(4 * time.Millisecond)
	lo := <-d.in
	hi := <-d.in
	return int(lo) + 256*int(hi)
}

func (d *ImageWriterIIDevice) readSlice(b []byte) {
	for i, _ := range b {
		b[i] = <-d.in
	}
	// time.Sleep(time.Duration(len(b)) * 2 * time.Millisecond)
}

func (d *ImageWriterIIDevice) readSliceInverted(b []byte) {
	for i, _ := range b {
		b[i] = bits.Reverse8(<-d.in)
	}
	// time.Sleep(time.Duration(len(b)) * 2 * time.Millisecond)
}

// CheckBounds ensures the print head is within valid margins, and page feeds if needed
func (d *ImageWriterIIDevice) CheckBounds(o OutputDriver, extra float64) {
	if d.ypos+extra >= d.h-d.formSkipBottom {
		o.FinalizePage()
		o.NewPage()
		d.ypos = d.marginTop
	}
}

// Process receives command bytes and controls the print driver
func (d *ImageWriterIIDevice) Process(o OutputDriver) {

	o.SetPageSize(d.w, d.h, d.pagehdpi, d.pagevdpi)
	o.NewPage()

	d.xpos = d.marginLeft
	d.ypos = d.marginTop

	for d.running {

		time.Sleep(5 * time.Millisecond)

		select {

		case _ = <-d.terminate:

			d.running = false
			return

		case ch := <-d.in:

			//fmt.Printf("%.2x/", ch)

			switch {

			case ch == 0x0d || ch == 0x8d:
				d.xpos = d.marginLeft
			case ch == 0x0a || ch == 0x8a:
				d.ypos += d.lineSpacing
				d.CheckBounds(o, d.lineSpacing)
			case ch == 0x0c || ch == 0x8c:
				o.FinalizePage()
				o.NewPage()
				d.xpos = d.marginLeft
				d.ypos = d.marginTop
			case ch == 0x1b || ch == 0x9b:
				command := d.readByte()

				log.Printf("ImageWriter: Command code %.2x %s\n", command, string(rune(command)))

				switch command & 0x7f {
				case 99:
					d.Reset() // reset printer
				case 0x42:
					d.lineSpacing = float64(9) / 72 // 1/8 spacing
					d.CheckBounds(o, d.lineSpacing)
				case 0x41:
					d.lineSpacing = float64(12) / 72 // 1/6 spacing
					d.CheckBounds(o, d.lineSpacing)
				case 70:
					// ESC F nnnn - begin printing at dot position nnnn
					str := string(rune(d.readByte())) + string(rune(d.readByte())) + string(rune(d.readByte())) + string(rune(d.readByte()))
					count, err := strconv.ParseInt(str, 10, 32)
					if err == nil {
						chunk := make([]byte, count)
						d.readSliceInverted(chunk)
						d.CheckBounds(o, d.lineSpacing)
						o.PlotDots(d.xpos, d.ypos, chunk, d.hdpi, d.vdpi, d.printColor)
					}
				case 71:
					// ESC G nnnn - next nnnn bytes as bit image graphics
					str := string(rune(d.readByte())) + string(rune(d.readByte())) + string(rune(d.readByte())) + string(rune(d.readByte()))
					count, err := strconv.ParseInt(str, 10, 32)
					if err == nil {
						chunk := make([]byte, count)
						d.readSliceInverted(chunk)
						d.CheckBounds(o, d.lineSpacing)
						o.PlotDots(d.xpos, d.ypos, chunk, d.hdpi, d.vdpi, d.printColor)
					}
				case 103:
					// ESC g nnn - next nnn x 8 bytes as bit image graphics
					str := string(rune(d.readByte())) + string(rune(d.readByte())) + string(rune(d.readByte()))
					count, err := strconv.ParseInt(str, 10, 32)
					if err == nil {
						chunk := make([]byte, count*8)
						d.readSliceInverted(chunk)
						d.CheckBounds(o, d.lineSpacing)
						o.PlotDots(d.xpos, d.ypos, chunk, d.hdpi, d.vdpi, d.printColor)
					}
				case 72:
					// form size
					str := string(rune(d.readByte())) + string(rune(d.readByte())) + string(rune(d.readByte())) + string(rune(d.readByte()))
					i, err := strconv.ParseInt(str, 10, 32)
					if err == nil {
						d.formSpacing = float64(i) / 144
					}
				case 76:
					// ESC L nnn - set left margin as n characters
					str := string(rune(d.readByte())) + string(rune(d.readByte())) + string(rune(d.readByte()))
					count, err := strconv.ParseInt(str, 10, 32)
					if err == nil {
						d.marginLeft = float64(1) / d.hcpi * float64(count)
						log.Printf("ImageWriter: Setting left-margin: %d chars at %f cpi", count, d.hcpi)
					}
				case 84:
					// line spacing dd
					str := string(rune(d.readByte())) + string(rune(d.readByte()))
					i, err := strconv.ParseInt(str, 10, 32)
					if err == nil {
						d.lineSpacing = float64(i) / 144
					}
					d.CheckBounds(o, d.lineSpacing)

				case 118:
					o.FinalizePage() // set top of form, treat as new page

				case byte('q'):
					d.style.pitch = fxPitchCondensed

				case byte('E'):
					d.style.pitch = fxPitchElite

				case byte('N'):
					d.style.pitch = fxPitchPica

				case 88:
					d.style.underline = true

				case 89:
					d.style.underline = false

				case 33:
					d.style.doubleStrike = true

				case 120:
					d.style.submode = fxSubModeSuperscript

				case 121:
					d.style.submode = fxSubModeSubscript

				case 122:
					d.style.submode = fxSubModeNone

				case 34:
					d.style.doubleStrike = false

				case byte('S'):
					fmt.Println("subscript")
					mode := d.readByte() & 0x7f
					if mode == byte('0') {
						d.style.submode = fxSubModeSuperscript
					} else {
						d.style.submode = fxSubModeSubscript
					}

				case byte('K'):
					cc := rune(d.readByte() & 0x7f)
					fmt.Printf("color code = %d\n", cc)
					switch cc {
					case '0':
						d.printColor = fxColorBlack
					case '1':
						d.printColor = fxColorYellow
					case '2':
						d.printColor = fxColorRed
					case '3':
						d.printColor = fxColorBlue
					case '4':
						d.printColor = fxColorOrange
					case '5':
						d.printColor = fxColorGreen
					case '6':
						d.printColor = fxColorMagenta
					}

				default:
					log.Printf("ImageWriter: unknown escape sequence %.2x %.2x\n", 0x1b, command)

				}

			case ch&0x7f == 15:
				d.style.pitch = fxPitchCondensed

			case ch&0x7f == 18:
				d.style.pitch = fxPitchPica

			case ch&0x7f == 14:
				d.style.expanded = true

			case ch&0x7f == 20:
				d.style.expanded = false

			case (ch & 127) >= 32:

				d.CheckBounds(o, d.lineSpacing)
				o.LetterAt(d.xpos, d.ypos, rune(ch&127), d.style, d.printColor)
				d.xpos += float64(1) / d.style.Cpi()

			}

			d.lastActivity = time.Now()

		default:

			if settings.FlushPDF[d.ent.GetMemIndex()] || (d.autoFF && time.Since(d.lastActivity) > time.Duration(settings.PrintToPDFTimeoutSec)*time.Second) {
				settings.FlushPDF[d.ent.GetMemIndex()] = false
				o.FinalizePage()
				o.Flush(d.ent)
				o.SetPageSize(d.w, d.h, d.pagehdpi, d.pagevdpi)
				o.NewPage()
				d.xpos = d.marginLeft
				d.ypos = d.marginTop
				d.lastActivity = time.Now()
			}

			time.Sleep(10 * time.Millisecond)

		}

	}

	// flush
	o.FinalizePage()
	o.Flush(d.ent)

	d.xpos = d.marginLeft
	d.ypos = d.marginTop

}
