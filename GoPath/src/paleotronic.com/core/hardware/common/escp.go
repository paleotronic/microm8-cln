package common

import (
	"time"

	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"

	"paleotronic.com/fmt"
)

type OutputDriver interface {
	LetterAt(x, y float64, ch rune, style fxFontStyle, color fxColor)
	PlotDots(x, y float64, chunk []byte, hdpi, vdpi float64, color fxColor)
	SetPageSize(w, h float64, hdpi, vdpi float64)
	FinalizePage()
	NewPage()
	Flush(ent interfaces.Interpretable)
}

type fxColor int

const (
	fxColorBlack   fxColor = 0
	fxColorMagenta fxColor = 1
	fxColorCyan    fxColor = 2
	fxColorViolet  fxColor = 3
	fxColorYellow  fxColor = 4
	fxColorRed     fxColor = 5
	fxColorGreen   fxColor = 6
	fxColorBlue    fxColor = 7
	fxColorOrange  fxColor = 8
)

type ESCPDevice struct {
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

func NewESCPDevice(driver OutputDriver, ent interfaces.Interpretable) *ESCPDevice {
	d := &ESCPDevice{}
	d.in = make(chan byte, 24576)
	d.terminate = make(chan bool)
	d.ent = ent
	d.Reset()

	// start processing command codes with the specified driver
	go d.Process(driver)

	return d
}

func (d *ESCPDevice) BufferCount() int {
	return len(d.in)
}

func (d *ESCPDevice) Write(b []byte) (int, error) {
	for _, bb := range b {
		d.in <- bb
	}
	return len(b), nil
}

func (d *ESCPDevice) Close() error {
	// d.terminate <- true
	d.running = false
	// close(d.in)
	return nil
}

// Reset resets the Epson emulation to page level defaults
func (d *ESCPDevice) Reset() {
	d.lineSpacing = float64(1) / 6
	d.marginLeft = 0.250
	d.marginTop = 0.0
	d.xpos = d.marginLeft
	d.ypos = d.marginTop
	d.hdpi = 120
	d.vdpi = 72
	d.hcpi = 10
	d.marginRight = float64(1) / d.hcpi * 80
	d.running = true
	d.autoFF = true
	d.lastActivity = time.Now()
	d.w = 8.5
	d.h = 11
	d.pagehdpi = 240
	d.pagevdpi = 144
	d.formSpacing = d.h
	d.formSkipBottom = 0
	d.formLines = 66
	d.style = fxFontStyle{}
	d.printColor = fxColorBlack
}

// Stop terminates the spool action
func (d *ESCPDevice) Stop() {
	d.running = false
}

func (d *ESCPDevice) readByte() byte {
	return <-d.in
}

func (d *ESCPDevice) readWord() int {
	lo := <-d.in
	hi := <-d.in
	return int(lo) + 256*int(hi)
}

func (d *ESCPDevice) readSlice(b []byte) {
	for i, _ := range b {
		b[i] = <-d.in
	}
}

// CheckBounds ensures the print head is within valid margins, and page feeds if needed
func (d *ESCPDevice) CheckBounds(o OutputDriver, extra float64) {
	if d.ypos+extra >= d.h-d.formSkipBottom {
		o.FinalizePage()
		o.NewPage()
		d.ypos = d.marginTop
	}
}

// Process receives command bytes and controls the print driver
func (d *ESCPDevice) Process(o OutputDriver) {

	o.SetPageSize(d.w, d.h, d.pagehdpi, d.pagevdpi)
	o.NewPage()

	d.xpos = d.marginLeft
	d.ypos = d.marginTop

	for d.running {

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

				fmt.Printf("Command code %.2x %s\n", command, string(rune(command)))

				switch command & 0x7f {
				case 0x40:
					d.Reset() // reset printer
				case 0x30:
					d.lineSpacing = float64(9) / 72
				case 0x31:
					d.lineSpacing = float64(7) / 72
				case 0x32:
					d.lineSpacing = float64(12) / 72 // 1/6 spacing
				case 0x33:
					// line spacing
					d.lineSpacing = float64(d.readByte()) / 216
				case 0x41:
					// line spacing
					d.lineSpacing = float64(d.readByte()&0x7f) / 72
				case 0x43:
					// form size
					v := d.readByte()
					if v == 0 {
						d.formSpacing = float64(d.readByte()) // inches
					} else {
						d.formSpacing = float64(v) * d.lineSpacing // lines
					}
				case 0x4a:
					// line spacing
					tmp := float64(d.readByte()) / 216
					d.ypos += tmp
					d.CheckBounds(o, d.lineSpacing)
				case 0x4e:
					// set line skip at bottom of form
					d.formSkipBottom = float64(d.readByte()) * d.lineSpacing // lines
					d.CheckBounds(o, d.lineSpacing)
				case 0x4b:
					/* 60x72 dpi graphics mode */
					d.hdpi = 60
					d.vdpi = 72

					count := d.readWord()

					chunk := make([]byte, count)
					d.readSlice(chunk)

					d.CheckBounds(o, d.lineSpacing)

					o.PlotDots(d.xpos, d.ypos, chunk, d.hdpi, d.vdpi, d.printColor)

					continue

				case 0x4c:
					/* 120x72 dpi graphics mode */
					d.hdpi = 120
					d.vdpi = 72

					count := d.readWord()

					chunk := make([]byte, count)
					d.readSlice(chunk)

					d.CheckBounds(o, d.lineSpacing)

					o.PlotDots(d.xpos, d.ypos, chunk, d.hdpi, d.vdpi, d.printColor)

					continue

				case 0x5a:
					/* 240x72 dpi graphics mode */
					d.hdpi = 240
					d.vdpi = 72

					count := d.readWord()

					chunk := make([]byte, count)
					d.readSlice(chunk)

					d.CheckBounds(o, d.lineSpacing)

					o.PlotDots(d.xpos, d.ypos, chunk, d.hdpi, d.vdpi, d.printColor)

					continue

				case 0x3d:

				case 0x70:
					_ = d.readByte()

				case 0x51:
					// right margin
					tmp := float64(1) / d.hcpi * float64(d.readByte())
					if tmp-d.marginLeft >= d.hcpi {
						d.marginRight = tmp
					}

				case 0x6a:
					// line spacing
					tmp := float64(d.readByte()) / 216
					if d.ypos >= tmp {
						d.ypos -= tmp
					}
					d.CheckBounds(o, d.lineSpacing)

				case 0x6c:
					// left margin
					d.marginLeft = float64(1) / d.hcpi * float64(d.readByte())

				case byte('M'):
					d.style.pitch = fxPitchElite

				case byte('P'):
					d.style.pitch = fxPitchPica

				case byte('G'):
					d.style.doubleStrike = true

				case byte('H'):
					d.style.doubleStrike = false

				case byte('-'):
					mode := d.readByte() & 0x7f
					d.style.underline = (mode == byte('1'))

				case byte('T'):
					d.style.submode = fxSubModeNone

				case byte('S'):
					fmt.Println("subscript")
					mode := d.readByte() & 0x7f
					if mode == byte('0') {
						d.style.submode = fxSubModeSuperscript
					} else {
						d.style.submode = fxSubModeSubscript
					}

				case byte('4'):
					switch d.style.style {
					case fxStyleRegular:
						d.style.style = fxStyleItalic
					case fxStyleBold:
						d.style.style = fxStyleBoldItalic
					}

				case byte('5'):
					switch d.style.style {
					case fxStyleBoldItalic:
						d.style.style = fxStyleBold
					case fxStyleItalic:
						d.style.style = fxStyleRegular
					}

				case byte('r'):
					cc := rune(d.readByte() & 0x7f)
					fmt.Printf("color code = %d\n", cc)
					switch cc {
					case 0:
						d.printColor = fxColorBlack
					case 1:
						d.printColor = fxColorMagenta
					case 2:
						d.printColor = fxColorCyan
					case 3:
						d.printColor = fxColorViolet
					case 4:
						d.printColor = fxColorYellow
					case 5:
						d.printColor = fxColorRed
					case 6:
						d.printColor = fxColorGreen
					}

				case byte('!'):
					// master select
					x := d.readByte() & 0x7f
					switch x {
					case 0:
						d.style = fxFontStyle{}
					case 1:
						d.style = fxFontStyle{pitch: fxPitchElite}
					case 4:
						d.style = fxFontStyle{pitch: fxPitchCondensed}
					case 8:
						d.style = fxFontStyle{}
					case 16:
						d.style = fxFontStyle{doubleStrike: true}
					case 17:
						d.style = fxFontStyle{doubleStrike: true, pitch: fxPitchElite}
					case 20:
						d.style = fxFontStyle{doubleStrike: true, pitch: fxPitchCondensed}
					case 24:
						d.style = fxFontStyle{doubleStrike: true}
					case 32:
						d.style = fxFontStyle{expanded: true}
					case 33:
						d.style = fxFontStyle{expanded: true, pitch: fxPitchElite}
					case 36:
						d.style = fxFontStyle{expanded: true, pitch: fxPitchCondensed}
					case 40:
						d.style = fxFontStyle{expanded: true}
					case 48:
						d.style = fxFontStyle{doubleStrike: true, expanded: true}
					case 49:
						d.style = fxFontStyle{doubleStrike: true, expanded: true, pitch: fxPitchElite}
					case 52:
						d.style = fxFontStyle{doubleStrike: true, expanded: true, pitch: fxPitchCondensed}
					case 56:
						d.style = fxFontStyle{doubleStrike: true, expanded: true}
					}

				default:
					fmt.Printf("unknown escape sequence %.2x %.2x\n", 0x1b, command)

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

			time.Sleep(1 * time.Second)

		}

	}

	// flush
	o.FinalizePage()
	o.Flush(d.ent)

	d.xpos = d.marginLeft
	d.ypos = d.marginTop

}
