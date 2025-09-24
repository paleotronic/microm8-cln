package ui

import (
	"strings"
	"time"
	"unicode/utf8"

	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/octalyzer/clientperipherals"
	"paleotronic.com/presentation"
	"paleotronic.com/utils"
)

type UIElement struct {
	X, Y             int
	W, H             int
	VOffset, HOffset int
}

type MenuItem struct {
	Text         string
	Label        string
	handler      func(i int)
	menu         *Menu
	sep          bool
	parent       *Menu
	rx, ry       int
	Checked      bool
	IsCheck      bool
	Hidden       bool
	Hint         string
	IsPercentage bool
}

type Menu struct {
	UIElement
	Items        []*MenuItem
	Fg, Bg       uint64
	SelFg, SelBg uint64
	TitFg, TitBg uint64
	selected     int
	OnBeforeDraw func(m *Menu)
	parent       *Menu
	useTitle     bool
	title        string
	HintW        int
	IsPercentage bool
}

func (m *Menu) Parent() *Menu {
	return m.parent
}

func (m *Menu) Add(label string, text string, f func(i int)) *Menu {
	item := &MenuItem{Label: label, Text: text, handler: f, parent: m}
	m.Items = append(m.Items, item)
	return m
}

func (m *Menu) AddIf(condition bool, label string, text string, f func(i int)) *Menu {
	if !condition {
		return m
	}
	item := &MenuItem{Label: label, Text: text, handler: f, parent: m}
	m.Items = append(m.Items, item)
	return m
}

func (m *Menu) AddCustom(f func(m *Menu)) *Menu {
	f(m)
	return m
}

func (m *Menu) AddCheck(label string, text string, f func(i int)) *Menu {
	item := &MenuItem{Label: label, Text: text, handler: f, parent: m, IsCheck: true}
	m.Items = append(m.Items, item)
	return m
}

func (m *Menu) AddCheckIf(label string, text string, condition bool, f func(i int)) *Menu {
	if !condition {
		return m
	}
	item := &MenuItem{Label: label, Text: text, handler: f, parent: m, IsCheck: true}
	m.Items = append(m.Items, item)
	return m
}

func (m *Menu) AddSep() *Menu {
	item := &MenuItem{sep: true, parent: m}
	m.Items = append(m.Items, item)
	return m
}

func (m *Menu) AddMenu(label string, text string) *Menu {
	item := &MenuItem{Label: label, Text: text, menu: NewMenu(m).Title(text), parent: m}
	m.Items = append(m.Items, item)
	return item.menu
}

func (m *Menu) Title(text string) *Menu {
	m.useTitle = true
	m.title = text
	return m
}

func (m *Menu) Draw(e interfaces.Interpretable) {

	if m.OnBeforeDraw != nil {
		m.OnBeforeDraw(m)
	}

	m.SetOptimumSize()

	txt := apple2helpers.GETHUD(e, "OOSD").Control
	txt.HideCursor()

	txt.FGColor = 0
	txt.BGColor = m.Bg

	txt.DrawTextBox(
		m.X, m.Y,
		m.W+1, m.H+1,
		"",
		true,
		true,
	)

	SP := string(rune(299))

	deadLines := 1
	if m.useTitle {
		aw := m.W - 4
		deadLines = 2
		text := m.title
		if utf8.RuneCountInString(text) > aw {
			text = text[0:aw-3] + "..."
		}
		for utf8.RuneCountInString(text) < aw {
			text += SP
		}
		txt.FGColor = m.TitFg
		txt.BGColor = m.TitBg
		txt.PutStr(SP + text + SP + SP + SP)
	}

	items := m.GetVisible()
	for i := 0; i < m.H-deadLines; i++ {
		if m.VOffset+i < len(items) {
			idx := m.VOffset + i
			item := items[idx]
			item.rx = m.X + m.W
			item.ry = m.Y + (i - m.VOffset)
			selected := (idx == m.selected)
			aw := m.W - 4
			sym := SP
			if item.menu != nil {
				sym = string(rune(258))
			}
			text := strings.Replace(item.Text, " ", SP, -1)
			if item.sep {
				text = ""
				for i := 0; i < aw; i++ {
					text += string(rune(1077))
				}
			}
			if utf8.RuneCountInString(text) > aw && !item.sep {
				text = text[0:aw-3] + "..."
			}
			for utf8.RuneCountInString(text) < aw-utf8.RuneCountInString(item.Hint) {
				text += SP
			}
			text += item.Hint
			for utf8.RuneCountInString(text) < aw {
				text += SP
			}
			if item.IsCheck {
				if m.IsPercentage && (strings.HasSuffix(item.Text, "%") || item.IsPercentage) {
					if item.Checked {
						sym = SymbolSliderMark
					} else {
						sym = SymbolSliderHandle
					}
				} else {
					if item.Checked {
						sym = SymbolOn
					} else {
						sym = SymbolOff
					}
				}
			}
			if selected {
				txt.FGColor = m.SelFg
				txt.BGColor = m.SelBg
			} else {
				txt.FGColor = m.Fg
				txt.BGColor = m.Bg
			}
			txt.PutStr(SP + text + SP + sym + SP)
			txt.FGColor = m.Fg
			txt.BGColor = m.Bg
		}
	}
	txt.ClearToEOLWindow()

}

func (m *Menu) Before(f func(m *Menu)) *Menu {
	m.OnBeforeDraw = f
	m.SetOptimumSize()
	return m
}

func (m *Menu) SetOptimumSize() {
	w := 5
	h := 1
	idx := 0
	hintw := 0
	for _, item := range m.Items {
		if item.Hidden {
			continue
		}

		if len(item.Hint) > hintw {
			hintw = len(item.Hint)
		}

		if len(item.Text)+4 > w {
			w = len(item.Text) + 4
		}

		if idx+3 > h {
			h = idx + 3
		}

		idx++
	}
	if w > 40 {
		w = 40
	}
	if h > 40 {
		h = 40
	}
	// m.UIElement.X = 0
	// m.UIElement.Y = 0
	m.UIElement.W = w + hintw
	m.UIElement.H = h
	if m.X+m.W+1 >= 80 {
		m.X = 80 - (m.W + 1)
	}
	m.HintW = hintw
}

const autoPop = false

func (m *Menu) Find(name string) *MenuItem {
	for _, v := range m.Items {
		if v.Label == name {
			return v
		}
	}
	return nil
}

func (m *Menu) GetVisible() []*MenuItem {
	m.SetOptimumSize()
	out := make([]*MenuItem, 0, len(m.Items))
	for _, item := range m.Items {
		if !item.Hidden {
			out = append(out, item)
		}
	}
	return out
}

func (m *Menu) Run(e interfaces.Interpretable) int {

	mm := e.GetMemoryMap()
	idx := e.GetMemIndex()

	// save screen
	txt := apple2helpers.GETHUD(e, "OOSD").Control
	txt.SaveState()
	defer txt.RestoreState()

	mouseEntered := false

	running := true
	key := rune(0)
	var lx, ly uint64 = mm.IntGetMousePos(idx)
	var lastB0 bool

	clientperipherals.SPEAKER.MakeTone(967, 50)

	if m.OnBeforeDraw != nil {
		m.OnBeforeDraw(m)
	}
	items := m.GetVisible()

	for running {

		m.Draw(e)

		for {
			key = rune(mm.KeyBufferGet(idx))
			if key != 0 {
				break
			}

			x, y := mm.IntGetMousePos(idx)
			if x != lx || y != ly {
				// movement
				fmt.Printf("Mouse(x=%d, y=%d)\n", x, y)
				lx, ly = x, y

				sx, sy := uint64(m.X), uint64(m.Y)
				ex, ey := uint64(m.X+m.W-1), uint64(m.Y+m.H-1)

				deadLines := 0
				if m.useTitle {
					deadLines = 1
				}

				if x >= sx && x <= ex && y >= sy && y <= ey {
					// mouse in bounds
					nsel := int(y) - m.Y - deadLines
					if nsel >= len(items) {
						nsel = len(items) - 1
					}
					if nsel < 0 {
						nsel = 0
					}
					if nsel != m.selected {
						//clientperipherals.SPEAKER.MakeTone(967, 33)
						//log2.Printf("new sel=%d, old sel=%d, label = %s", nsel, m.selected, items[nsel])
						m.selected = nsel
					}
					if autoPop && items[m.selected].menu != nil {
						key = 13
						break
					}
					mouseEntered = true
					break
				} else if mouseEntered {
					//return 0
				}
			}

			b0, _ := mm.IntGetMouseButtons(idx)
			if b0 != lastB0 {
				sx, sy := uint64(m.X), uint64(m.Y)
				ex, ey := uint64(m.X+m.W-1), uint64(m.Y+m.H-1)
				lastB0 = b0
				if b0 == false && lx >= sx && lx <= ex && ly >= sy && ly <= ey {
					key = 13
					break
				} else if b0 == false {
					key = 27
					break
				}
			}

			time.Sleep(50 * time.Millisecond)
		}

		//fmt.Printf("Got key == %d\n", key)

		switch {
		case key == vduconst.CSR_UP:
			m.selected--
			if m.selected < 0 {
				m.selected = 0
			}
			//clientperipherals.SPEAKER.MakeTone(967, 33)
		case key == vduconst.CSR_DOWN:
			m.selected++
			if m.selected >= len(items) {
				m.selected = len(items) - 1
			}
			//clientperipherals.SPEAKER.MakeTone(967, 33)
		case key == 32:
			if items[m.selected].handler != nil {
				clientperipherals.SPEAKER.MakeTone(967, 50)
				time.Sleep(50 * time.Millisecond)
				clientperipherals.SPEAKER.MakeTone(488, 50)
				items[m.selected].handler(m.selected)
				if m.OnBeforeDraw != nil {
					m.OnBeforeDraw(m)
				}
			}
		case key == 13 || key == vduconst.CSR_RIGHT:
			if items[m.selected].menu != nil {
				menu := items[m.selected].menu
				menu.X = items[m.selected].rx + 1
				menu.Y = items[m.selected].ry
				mouseEntered = false
				lx = 0
				clientperipherals.SPEAKER.MakeTone(967, 50)
				if menu.Run(e) == -1 {
					return -1
				}
			} else if items[m.selected].handler != nil {
				// TODO: sound here
				clientperipherals.SPEAKER.MakeTone(967, 50)
				time.Sleep(50 * time.Millisecond)
				clientperipherals.SPEAKER.MakeTone(488, 50)
				items[m.selected].handler(m.selected)
				return -1
			}
		case key == vduconst.CSR_LEFT || key == 27:
			return 0
		}

	}

	return 0

}

var filepanel [memory.OCTALYZER_NUM_INTERPRETERS]*editor.FileCatalog
var lastindex [memory.OCTALYZER_NUM_INTERPRETERS]int

func CatalogPresentDiskPicker(ent interfaces.Interpretable, drive int) {

	//ent.StopTheWorld()
	s, e, p := files.System, settings.EBOOT, files.Project

	//	files.System = false
	//	settings.EBOOT = true
	if filepanel[ent.GetMemIndex()] == nil {
		s := editor.FileCatalogSettings{
			DiskExtensions: files.GetExtensionsDisk(),
			Title:          "microFile File Manager",
			Pattern:        files.GetPatternBootable(),
			Path:           "",
			BootstrapDisk:  false,
			InsertDisk:     true,
			TargetDisk:     drive,
			HidePaths:      []string{"FILECACHE", "system"},
		}
		filepanel[ent.GetMemIndex()] = editor.NewFileCatalog(ent, s)
	}
	settings.DisableMetaMode[ent.GetMemIndex()] = true
	lastindex[ent.GetMemIndex()], _ = filepanel[ent.GetMemIndex()].Do(lastindex[ent.GetMemIndex()])
	settings.DisableMetaMode[ent.GetMemIndex()] = false

	files.System = s
	settings.EBOOT = e
	files.Project = p

	//ent.ResumeTheWorld()

}

func CatalogPresentTapePicker(ent interfaces.Interpretable, drive int) {

	//ent.StopTheWorld()
	s, e, p := files.System, settings.EBOOT, files.Project

	//	files.System = false
	//	settings.EBOOT = true
	if filepanel[ent.GetMemIndex()] == nil {
		s := editor.FileCatalogSettings{
			DiskExtensions: []string{"wav"},
			Title:          "microFile File Manager",
			Pattern:        files.GetPatternTape(),
			Path:           "",
			BootstrapDisk:  false,
			InsertDisk:     true,
			TargetDisk:     drive,
			HidePaths:      []string{"FILECACHE", "system"},
		}
		filepanel[ent.GetMemIndex()] = editor.NewFileCatalog(ent, s)
	}
	settings.DisableMetaMode[ent.GetMemIndex()] = true
	lastindex[ent.GetMemIndex()], _ = filepanel[ent.GetMemIndex()].Do(lastindex[ent.GetMemIndex()])
	settings.DisableMetaMode[ent.GetMemIndex()] = false

	files.System = s
	settings.EBOOT = e
	files.Project = p

	//ent.ResumeTheWorld()

}

func setSlotAspect(mm *memory.MemoryMap, index int, aspect float64) {

	for i := -1; i < 9; i++ {
		control := types.NewOrbitController(mm, index, i)
		control.SetAspect(aspect)
	}

}

func GetCRTLine(ent interfaces.Interpretable, promptString string, def string) string {

	txt := apple2helpers.GETHUD(ent, "OOSD").Control

	command := def
	collect := true

	cb := ent.GetProducer().GetMemoryCallback(ent.GetMemIndex())

	txt.PutStr(promptString)
	txt.PutStr(command)

	if cb != nil {
		cb(ent.GetMemIndex())
	}

	for collect {
		if cb != nil {
			cb(ent.GetMemIndex())
		}
		txt.ShowCursor()
		for ent.GetMemory(49152) < 128 {
			time.Sleep(5 * time.Millisecond)
			if cb != nil {
				cb(ent.GetMemIndex())
			}
		}
		txt.HideCursor()
		ch := rune(ent.GetMemory(49152) & 0xff7f)
		ent.SetMemory(49168, 0)
		switch ch {
		case 10:
			{
				//txt.SetSuppressFormat(true)
				txt.PutStr("\r\n")
				//txt.SetSuppressFormat(false)
				return command
			}
		case 13:
			{
				//txt.SetSuppressFormat(true)
				txt.PutStr("\r\n")
				//txt.SetSuppressFormat(false)
				return command
			}
		case 127:
			{
				if len(command) > 0 {
					command = utils.Copy(command, 1, len(command)-1)
					txt.Backspace()
					//txt.SetSuppressFormat(true)
					txt.PutStr(" ")
					//txt.SetSuppressFormat(false)
					txt.Backspace()
					if cb != nil {
						cb(ent.GetMemIndex())
					}
				}
				break
			}
		default:
			{

				//txt.SetSuppressFormat(true)
				txt.Put(rune(ch))
				//txt.SetSuppressFormat(false)

				if cb != nil {
					cb(ent.GetMemIndex())
				}

				command = command + string(ch)
				break
			}
		}
	}

	if cb != nil {
		cb(ent.GetMemIndex())
	}

	return command

}

func InputPopup(ent interfaces.Interpretable, title string, message string, def string) string {

	txt := apple2helpers.GETHUD(ent, "OOSD").Control

	txt.AddNamedWindow("std", 0, 0, 79, 47)

	txt.FGColor = 15
	txt.BGColor = 2

	txt.DrawTextBox(
		5,
		20,
		70,
		8,
		title,
		true,
		true,
	)

	txt.PutStr(title + "\r\n")
	defer txt.HideCursor()

	txt.ShowCursor()
	s := GetCRTLine(ent, message+": ", def)

	txt.SetNamedWindow("std")

	return s

}

type DefaultSettings struct {
	P       *presentation.Presentation
	updated bool
	path    string
}

func NewDefaultSettings(e interfaces.Interpretable) *DefaultSettings {
	filepath := "/local/settings/default"
	p, err := files.OpenPresentationStateSoft(e, filepath)
	if err == nil {
		return &DefaultSettings{P: p, path: filepath}
	}

	p, _ = files.NewPresentationStateDefault(e, filepath)

	return &DefaultSettings{P: p, path: filepath}
}

func (d *DefaultSettings) SetF(section string, path string, v float64) {
	d.P.WriteFloat(section, path, v)
	d.updated = true
}

func (d *DefaultSettings) SetI(section string, path string, v int) {
	d.P.WriteInt(section, path, v)
	d.updated = true
}

func (d *DefaultSettings) SetS(section string, path string, v string) {
	d.P.WriteString(section, path, v)
	d.updated = true
}

func (d *DefaultSettings) SetSL(section string, path string, v []string) {
	d.P.WriteStringlist(section, path, v)
	d.updated = true
}

func (d *DefaultSettings) Finalize() error {
	fmt.Printf("Was settings changed? %v\n", d.updated)
	if !d.updated {
		return nil
	}

	d.path = strings.Replace(d.path, "\\", "/", -1)
	if !files.ExistsViaProvider(d.path, "") {
		fmt.Printf("mkdir %s\n", d.path)
		files.MkdirViaProvider(d.path)
	}

	return files.SavePresentationStateToFolder(d.P, d.path)
}
