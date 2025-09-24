package memory

import (
	"paleotronic.com/core/settings"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
)

type MemoryAction int

const (
	MA_WRITE MemoryAction = iota
	MA_READ
	MA_ANY
)

func (a MemoryAction) String() string {
	if a == MA_READ {
		return "READ"
	} else {
		return "WRITE"
	}
}

type MemoryBlock struct {
	mcb        *MemoryControlBlock
	mm         *MemoryMap
	Index      int
	ReadOk     bool
	WriteOk    bool
	Active     bool
	GlobalBase int
	Base       int
	MuxBase    int
	ForceMux   bool
	Size       int
	//Data []uint64
	Label string
	mr    Mappable
	fr    Firmware
	Type  int
}

func NewMemoryBlockRAM(mm *MemoryMap, index int, globalbase int, addr int, size int, active bool, label string, mux int, forceMux bool, t int) *MemoryBlock {
	b := &MemoryBlock{
		ReadOk:     true,
		WriteOk:    true,
		Active:     active,
		Base:       addr,
		GlobalBase: globalbase,
		Size:       size,
		Label:      label,
		MuxBase:    mux,
		ForceMux:   forceMux,
		mm:         mm,
		Index:      index,
		Type:       t,
	}
	b.InitMemory([]uint64(nil))
	return b
}

func NewMemoryBlockIO(mm *MemoryMap, index int, globalbase int, addr int, size int, active bool, label string, mr Mappable) *MemoryBlock {
	b := &MemoryBlock{
		ReadOk:     true,
		WriteOk:    true,
		Active:     active,
		Base:       addr,
		GlobalBase: globalbase,
		Size:       size,
		Label:      label,
		mr:         mr,
		mm:         mm,
		Index:      index,
	}
	return b
}

func NewMemoryBlockFirmware(mm *MemoryMap, index int, globalbase int, addr int, size int, active bool, label string, fr Firmware) *MemoryBlock {
	b := &MemoryBlock{
		ReadOk:     true,
		WriteOk:    true,
		Active:     active,
		Base:       addr,
		GlobalBase: globalbase,
		Size:       size,
		Label:      label,
		fr:         fr,
		mm:         mm,
		Index:      index,
	}
	return b
}

func NewMemoryBlockROM(mm *MemoryMap, index int, globalbase int, addr int, size int, active bool, label string, romdata []uint64) *MemoryBlock {
	b := &MemoryBlock{
		ReadOk:     true,
		WriteOk:    false,
		Active:     active,
		Base:       addr,
		GlobalBase: globalbase,
		Size:       size,
		Label:      label,
		mm:         mm,
		Index:      index,
	}
	b.InitROM(romdata)
	return b
}

func (b *MemoryBlock) InitMemory(data []uint64) {

	// create the mcb
	s := NewMemoryControlBlock(b.mm, b.Index, false)

	if b.MuxBase != 0 || b.ForceMux {
		s.Add(b.mm.Data[b.Index][b.GlobalBase+b.MuxBase:b.GlobalBase+b.MuxBase+b.Size], b.GlobalBase+b.MuxBase)
	} else {
		s.Add(b.mm.Data[b.Index][b.GlobalBase+b.Base:b.GlobalBase+b.Base+b.Size], b.GlobalBase+b.Base)
	}

	b.mcb = s

	if data != nil {
		var overflow int
		for i, v := range data {
			if i < b.Size {
				b.mcb.Write(i, v)
			} else {
				overflow++
			}
		}
		if overflow > 0 {
			b.Log("Data init overflowed Block by %d bytes (extra ignored)")
		}
	}
}

func (b *MemoryBlock) InitROM(data []uint64) {

	// create the mcb
	s := NewMemoryControlBlock(b.mm, b.Index, false)

	s.Add(data, b.GlobalBase+b.Base)

	b.mcb = s

	if data != nil {
		var overflow int
		for i, v := range data {
			if i < b.Size {
				b.mcb.Write(i, v)
			} else {
				overflow++
			}
		}
		if overflow > 0 {
			b.Log("Data init overflowed Block by %d bytes (extra ignored)")
		}
	}
}

// Claimed returns true if this Memory Block can handle the current action
func (b *MemoryBlock) Claimed(addr int, action MemoryAction) bool {

	return (addr >= b.Base && addr < b.Base+b.Size) &&
		((b.ReadOk == true && action == MA_READ) || (b.WriteOk == true && action == MA_WRITE))

}

func (b *MemoryBlock) GetState() string {

	st := "off"

	if b.Active && (b.ReadOk || b.WriteOk) {
		st = ""
		if b.ReadOk {
			st += "r"
		}
		if b.WriteOk {
			st += "w"
		}
	}

	return st

}

func (b *MemoryBlock) SetState(st string) {

	switch st {
	case "off":
		b.Active = false
		b.ReadOk = false
		b.WriteOk = false
	case "r":
		b.Active = true
		b.ReadOk = true
		b.WriteOk = false
	case "w":
		b.Active = true
		b.ReadOk = false
		b.WriteOk = true
	case "rw":
		b.Active = true
		b.ReadOk = true
		b.WriteOk = true
	default:
		panic("invalid state: " + st + " for " + b.Label)
	}

}

func (b *MemoryBlock) DirectRead(offset int) uint64 {

	return b.mcb.Read(offset)

}

// Do does the read write
func (b *MemoryBlock) Do(addr int, action MemoryAction, value *uint64) bool {

	if !b.Claimed(addr, action) {
		if addr == 0xc800 {
			fmt.Println("not claimed", addr, b.Base, b.Size, b.ReadOk, action.String())
		}
		return false
	}

	saddr := addr
	if b.MuxBase != 0 || b.ForceMux {
		saddr = saddr + (b.MuxBase - b.Base)
	}

	offset := addr - b.Base

	// if addr == 0xfff0 {
	// 	log.Printf("0xfff0: offset %.4x access to range %.4x:%.4x", offset, b.GlobalBase+b.MuxBase, b.GlobalBase+b.MuxBase+b.Size)
	// }

	index := b.GlobalBase / OCTALYZER_INTERPRETER_SIZE
	b.mm.BlockMapper[index].FirmwareLastRead = nil

	switch action {
	case MA_READ:
		if b.mr != nil {
			*value = b.mr.RelativeRead(offset)
			return true
		} else if b.fr != nil {
			*value = b.fr.FirmwareRead(offset)

			b.mm.BlockMapper[index].FirmwareLastRead = b.fr
			return true
		}
		*value = b.mcb.Read(offset)
		return true
	case MA_WRITE:
		if b.mr != nil {
			//oval := b.mr.ReadData(offset)
			b.mr.RelativeWrite(offset, *value)
			return true
		} else if b.fr != nil {
			b.fr.FirmwareWrite(offset, *value)
			return true
		}
		//g, o := b.mcb.GetRef(offset)
		//oval := b.mcb.Data[g][o]
		b.mcb.Write(offset, *value)
		return true
	}

	return true

}

// Absolute returns the "Absolute" slot address read from or written to
func (b *MemoryBlock) Absolute(addr int, action MemoryAction) (int, bool) {

	if !b.Claimed(addr, action) {
		return addr, false
	}

	saddr := addr
	if b.MuxBase != 0 || b.ForceMux {
		saddr = saddr + (b.MuxBase - b.Base)
	}

	//fmt.Printf("%s($%.4x) -> %s (actual $%.4x)\n", action.String(), addr, b.Label, saddr)

	offset := addr - b.Base

	switch action {
	case MA_READ:
		if b.mr != nil {
			return b.GlobalBase + saddr, true
		} else if b.fr != nil {
			return b.GlobalBase + saddr, true
		}
		g, o := b.mcb.GetRef(offset)
		return b.mcb.GStart[g] + o, true
	case MA_WRITE:
		if b.mr != nil {
			return b.GlobalBase + saddr, true
		} else if b.fr != nil {
			return b.GlobalBase + saddr, true
		}
		g, o := b.mcb.GetRef(offset)
		return b.mcb.GStart[g] + o, true
	}

	return 0, false

}

func (b *MemoryBlock) Log(format string, args ...interface{}) {

	prefix := fmt.Sprintf("%s@$%.4x:$%.4x", b.Label, b.Base, b.Base+b.Size-1)

	log.Printf(prefix+": "+format, args...)
}

func (b *MemoryBlock) String() string {
	return fmt.Sprintf("%s@$%.4x:$%.4x (read: %v, write %v, active: %v)", b.Label, b.Base, b.Base+b.Size-1, b.ReadOk, b.WriteOk, b.Active)
}

type MemoryManagementUnit struct {
	m         []*MemoryBlock
	cache     [256][]*MemoryBlock // only useful below 64Kb but speed is speed
	listeners [256][]*MemoryListener

	MinListenerRange int
	MaxListenerRange int

	PageREAD, PageWRITE [256]*MemoryBlock

	LastMem *MemoryBlock

	DefaultMap map[string]string

	NoListeners bool

	ResetFunc func(skip bool)

	FirmwareLastRead Firmware // for IO mounted firmwares

	HeatMap     [256]byte // represents a 256 bank heatmap
	HeatMapMode settings.HeatMapMode
	HeatMapBank int
	CPURead     bool
}

func NewMemoryManagementUnit() *MemoryManagementUnit {
	m := &MemoryManagementUnit{
		m: make([]*MemoryBlock, 0),
		//listeners:        make([]*MemoryListener, 0),
		MinListenerRange: 999999,
		MaxListenerRange: -999999,
		// PageREAD:  make([]*MemoryBlock, 256),
		// PageWRITE: make([]*MemoryBlock, 256),
	}
	return m
}

func (mmu *MemoryManagementUnit) SetBankREAD(start, end int, target *MemoryBlock) {
	for i := start; i < end; i++ {
		mmu.PageSetREAD(i, target)
	}
}

func (mmu *MemoryManagementUnit) SetBankWRITE(start, end int, target *MemoryBlock) {
	for i := start; i < end; i++ {
		mmu.PageSetWRITE(i, target)
	}
}

func (mmu *MemoryManagementUnit) FillBanksREAD(condition bool, start int, end int, tbank *MemoryBlock, fbank *MemoryBlock) {
	if condition {
		mmu.SetBankREAD(start, end, tbank)
	} else {
		mmu.SetBankREAD(start, end, fbank)
	}
}

func (mmu *MemoryManagementUnit) FillBanksWRITE(condition bool, start int, end int, tbank *MemoryBlock, fbank *MemoryBlock) {
	if condition {
		mmu.SetBankWRITE(start, end, tbank)
	} else {
		mmu.SetBankWRITE(start, end, fbank)
	}
}

func (m *MemoryManagementUnit) ClearCache() {
	for i, _ := range m.cache {
		m.cache[i] = nil
	}
}

func (m *MemoryManagementUnit) IndexOf(name string) int {
	for i, b := range m.m {
		if b.Label == name {
			return i
		}
	}
	return -1
}

func (m *MemoryManagementUnit) SetHeatMapMode(slotid int, mode settings.HeatMapMode, bank int) {
	settings.HeatMap[slotid] = mode
	settings.HeatMapBank[slotid] = bank
	m.HeatMapBank = bank
	m.HeatMapMode = mode
	m.ClearHeatmap()
}

func (m *MemoryManagementUnit) HeatMapWrite(addr int, block int, t int) {
	if m.HeatMapMode == settings.HMOff {
		return
	}
	switch m.HeatMapMode {
	case settings.HMMain:
		if t == 1 {
			return
		}
		v := m.HeatMap[block]
		vv := (v & 0xf0) >> 4
		vv++
		if vv > 15 {
			vv = 15
		}
		m.HeatMap[block] = v&0xf | (vv << 4)
	case settings.HMAux:
		if t == 0 {
			return
		}
		v := m.HeatMap[block]
		vv := (v & 0xf0) >> 4
		vv++
		if vv > 15 {
			vv = 15
		}
		m.HeatMap[block] = v&0xf | (vv << 4)
	case settings.HMMainBank:
		if t == 1 || block != m.HeatMapBank {
			return
		}
		v := m.HeatMap[addr&255]
		vv := (v & 0xf0) >> 4
		vv++
		if vv > 15 {
			vv = 15
		}
		m.HeatMap[addr&255] = v&0xf | (vv << 4)
	case settings.HMAuxBank:
		if t == 0 || block != m.HeatMapBank {
			return
		}
		v := m.HeatMap[addr&255]
		vv := (v & 0xf0) >> 4
		vv++
		if vv > 15 {
			vv = 15
		}
		m.HeatMap[addr&255] = v&0xf | (vv << 4)
	}
}

func (m *MemoryManagementUnit) HeatMapRead(addr int, block int, t int) {
	if m.HeatMapMode == settings.HMOff {
		return
	}
	switch m.HeatMapMode {
	case settings.HMExecCombined:
		if t == 2 || !m.CPURead {
			return
		}
		switch t {
		case 0:
			v := m.HeatMap[block]
			vv := v & 0xf
			vv++
			if vv > 15 {
				vv = 15
			}
			m.HeatMap[block] = v&0xf0 | vv
		case 1:
			v := m.HeatMap[block]
			vv := (v & 0xf0) >> 4
			vv++
			if vv > 15 {
				vv = 15
			}
			m.HeatMap[block] = v&0xf | (vv << 4)
		}
	case settings.HMMain:
		if t == 1 {
			return
		}
		v := m.HeatMap[block]
		vv := v & 0xf
		vv++
		if vv > 15 {
			vv = 15
		}
		m.HeatMap[block] = v&0xf0 | vv
	case settings.HMAux:
		if t == 0 {
			return
		}
		v := m.HeatMap[block]
		vv := v & 0xf
		vv++
		if vv > 15 {
			vv = 15
		}
		m.HeatMap[block] = v&0xf0 | vv
	case settings.HMMainBank:
		if t == 1 || block != m.HeatMapBank {
			return
		}
		log.Printf("main bank read addr = %.4x", addr)
		v := m.HeatMap[addr&255]
		vv := v & 0xf
		vv++
		if vv > 15 {
			vv = 15
		}
		m.HeatMap[addr&255] = v&0xf0 | vv
	case settings.HMAuxBank:
		if t == 0 || block != m.HeatMapBank {
			return
		}
		v := m.HeatMap[addr&255]
		vv := v & 0xf
		vv++
		if vv > 15 {
			vv = 15
		}
		m.HeatMap[addr&255] = v&0xf0 | vv
	}
}

func (m *MemoryManagementUnit) Done() {
	for _, b := range m.m {
		if b.mr != nil {
			b.mr.Done()
		}
	}
}

func (m *MemoryManagementUnit) Register(mb *MemoryBlock) {
	current := m.IndexOf(mb.Label)

	if current != -1 {

		// we are replacing something
		old := m.m[current]
		if old.mr != nil {
			fmt.Printf("Replacing %s\n", old.mr.GetLabel())
			old.mr.Done()
		}

		m.m[current] = mb
	} else {
		m.m = append(m.m, mb)
	}

	// We add references to the cache
	startBlock := (mb.Base / 256) % 256
	endBlock := ((mb.Base + mb.Size - 1) / 256) % 256

	for b := startBlock; b <= endBlock; b++ {
		m.CacheAdd(b, mb)
	}
}

func (m *MemoryManagementUnit) CacheAdd(b int, mb *MemoryBlock) {
	list := m.cache[b]
	if list == nil {
		list = []*MemoryBlock{mb}
		m.cache[b] = list
		return
	}
	i := -1
	for idx, ob := range list {
		if ob.Label == mb.Label {
			i = idx
		}
	}

	if i != -1 {
		list[i] = mb
		return
	}

	list = append(list, mb)
	m.cache[b] = list
}

func (m *MemoryManagementUnit) CacheRemove(b int, mb *MemoryBlock) {
	list := m.cache[b]
	if list == nil {
		return
	}

	i := -1
	for idx, ob := range list {
		if ob.Label == mb.Label {
			i = idx
		}
	}

	if i != -1 {
		list = append(list[:i], list[i+1:]...)
		m.cache[b] = list
		return
	}
}

func (m *MemoryManagementUnit) ClearHeatmap() {
	for i, _ := range m.HeatMap {
		m.HeatMap[i] = 0x00
	}
}

func (m *MemoryManagementUnit) CoolHeatmap() {
	var w, r byte
	for i, v := range m.HeatMap {
		w = (v & 0xf0 >> 4)
		r = (v & 0x0f)
		if w > 1 {
			w--
		}
		if r > 1 {
			r--
		}
		m.HeatMap[i] = (w << 4) | r
	}
}

func (m *MemoryManagementUnit) CacheGet(b int) []*MemoryBlock {
	return m.cache[b]
}

func (m *MemoryManagementUnit) Reset(skip bool) {
	if m.ResetFunc != nil {
		m.ResetFunc(skip)
	}
	m.ClearHeatmap()
}

func (m *MemoryManagementUnit) BuildDefaultMap() {

	d := make(map[string]string)

	for _, mb := range m.m {
		d[mb.Label] = mb.GetState()

		fmt.Printf("Default: %s ~~> [%s]\n", mb.Label, d[mb.Label])
	}

	m.DefaultMap = d
}

func (m *MemoryManagementUnit) GetDefaultMap() map[string]string {
	d := make(map[string]string)
	for k, v := range m.DefaultMap {
		d[k] = v
	}
	return d
}

func (m *MemoryManagementUnit) SetMap(d map[string]string) {

	for name, st := range d {

		mb := m.Get(name)
		if mb != nil {

			mb.SetState(st)

		}

	}

}

func (m *MemoryManagementUnit) Unregister(name string) {

	var mb *MemoryBlock

	current := m.IndexOf(name)
	if current != -1 {
		mb = m.m[current]
		m.m = append(m.m[0:current], m.m[current+1:]...)
	}
	// We add references to the cache
	if mb != nil {
		startBlock := (mb.Base / 256) % 256
		endBlock := ((mb.Base + mb.Size - 1) / 256) % 256

		for b := startBlock; b <= endBlock; b++ {
			m.CacheRemove(b, mb)
		}
	}
}

func (m *MemoryManagementUnit) Do(addr int, action MemoryAction, value *uint64) bool {

	// check listeners
	if !m.NoListeners {
		cont := m.ProcessListeners(addr, value, action)
		if !cont {
			return true // handled... no further update
		}
	}

	block := (addr / 256)

	if block < 256 {

		switch action {
		case MA_READ:
			mb := m.PageREAD[block]
			if mb == nil {
				// fmt.Printf("Request for block %d is nil\n", addr)
				if block >= 0xc1 && block <= 0xc7 {
					*value = 0xa0 // junk value
					return false
				}
				return false
			}
			if settings.DebuggerOn {
				m.HeatMapRead(addr, block, mb.Type)
			}
			// if addr == 0xc800 {
			// 	fmt.Printf("%s\n", mb.Label)
			// }
			z := mb.Do(addr, action, value)
			// this is to trick certain copy protections
			// TODO: need to put this fix into the apple side!!!
			if block >= 0xc1 && block <= 0xc7 && *value == 0x00 && addr == block*256 && m.IndexOf("apple2iochip") >= 0 {
				*value = 0x60
			}
			return z
		case MA_WRITE:
			mb := m.PageWRITE[block]
			if mb == nil {
				return false
			}
			if settings.DebuggerOn {
				m.HeatMapWrite(addr, block, mb.Type)
			}
			return mb.Do(addr, action, value)
		}

	}

	return false

}

func (m *MemoryManagementUnit) Absolute(addr int, action MemoryAction) (int, bool) {

	block := (addr / 256)

	if block < 256 {

		switch action {
		case MA_READ:
			mb := m.PageREAD[block]
			if mb == nil {
				return addr, false
			}
			return mb.Absolute(addr, action)
		case MA_WRITE:
			mb := m.PageWRITE[block]
			if mb == nil {
				return addr, false
			}
			return mb.Absolute(addr, action)
		}

	}

	return addr, false

}

func (m *MemoryManagementUnit) Log(format string, args ...interface{}) {

	prefix := "MemoryManagementUnit"

	log.Printf(prefix+": "+format, args...)
}

func (m *MemoryManagementUnit) Get(name string) *MemoryBlock {
	c := m.IndexOf(name)
	if c == -1 {
		return nil
	}
	return m.m[c]
}

func (m *MemoryManagementUnit) Enable(name string) {
	b := m.Get(name)
	if b != nil {
		b.Active = true
	}
}

func (m *MemoryManagementUnit) Disable(name string) {
	b := m.Get(name)
	if b != nil {
		b.Active = false
	}
}

func (m *MemoryManagementUnit) IsEnabled(name string) bool {
	b := m.Get(name)
	if b != nil {
		return b.Active
	}
	return false
}

func (m *MemoryManagementUnit) IsReadable(name string) bool {
	b := m.Get(name)
	if b != nil {
		return b.Active && b.ReadOk
	}
	return false
}

func (m *MemoryManagementUnit) IsWritable(name string) bool {
	b := m.Get(name)
	if b != nil {
		return b.Active && b.WriteOk
	}
	return false
}

func (m *MemoryManagementUnit) SetMode(name string, r, w bool) {
	b := m.Get(name)
	if b != nil {
		b.ReadOk = r
		b.WriteOk = w
	}
}

func (m *MemoryManagementUnit) SetWriteMode(name string, w bool) {
	b := m.Get(name)
	if b != nil {
		b.WriteOk = w
	}
}

func (m *MemoryManagementUnit) SetReadMode(name string, r bool) {
	b := m.Get(name)
	if b != nil {
		b.ReadOk = r
	}
}

// GetActiveBlocks return a list of active blocks for a context
func (m *MemoryManagementUnit) GetActiveBlocks(context MemoryAction) []*MemoryBlock {
	blocks := make([]*MemoryBlock, 0)
	for _, b := range m.m {
		switch context {
		case MA_READ:
			if b.Active && b.ReadOk {
				blocks = append(blocks, b)
			}
		case MA_WRITE:
			if b.Active && b.WriteOk {
				blocks = append(blocks, b)
			}
		}
	}
	return blocks
}

func (m *MemoryManagementUnit) GetMappedBlocks(context MemoryAction) []*MemoryBlock {
	// blocks := make([]*MemoryBlock, 256)
	// for i, _ := range blocks {
	// 	switch context {
	// 	case MA_READ:
	// 		blocks[i] = m.PageREAD[i]
	// 	case MA_WRITE:
	// 		blocks[i] = m.PageWRITE[i]
	// 	}
	// }
	// return blocks
	switch context {
	case MA_READ:
		return m.PageREAD[:]
	case MA_WRITE:
		return m.PageWRITE[:]
	}
	return []*MemoryBlock{}
}

func (m *MemoryManagementUnit) EnableBlocks(list []string) {
	for _, name := range list {
		m.Enable(name)
	}
}

func (m *MemoryManagementUnit) DisableBlocks(list []string) {
	for _, name := range list {
		m.Disable(name)
	}
}

func (m *MemoryManagementUnit) EnableRead(list []string) {
	for _, name := range list {
		m.SetReadMode(name, true)
	}
}

func (m *MemoryManagementUnit) DisableRead(list []string) {
	for _, name := range list {
		m.SetReadMode(name, false)
	}
}

func (m *MemoryManagementUnit) EnableWrite(list []string) {
	for _, name := range list {
		m.SetWriteMode(name, true)
	}
}

func (m *MemoryManagementUnit) DisableWrite(list []string) {
	for _, name := range list {
		m.SetWriteMode(name, false)
	}
}

func (m *MemoryManagementUnit) DumpMap() {

	rblocks := m.GetActiveBlocks(MA_READ)
	wblocks := m.GetActiveBlocks(MA_WRITE)

	maxwidth := 30

	// READ MAP

	fmt.Print("+")
	for i := 0; i < maxwidth; i++ {
		fmt.Print("-")
	}
	fmt.Println("+")

	fmt.Printf("|%-30s|\n", "MEMORY READ")

	fmt.Print("+")
	for i := 0; i < maxwidth; i++ {
		fmt.Print("-")
	}
	fmt.Println("+")

	for _, pagename := range rblocks {
		fmt.Printf("|%-30s|\n", " "+pagename.Label)
	}

	fmt.Print("+")
	for i := 0; i < maxwidth; i++ {
		fmt.Print("-")
	}
	fmt.Println("+")

	// WRITE MAP

	fmt.Println()

	fmt.Print("+")
	for i := 0; i < maxwidth; i++ {
		fmt.Print("-")
	}
	fmt.Println("+")

	fmt.Printf("|%-30s|\n", " MEMORY WRITE")

	fmt.Print("+")
	for i := 0; i < maxwidth; i++ {
		fmt.Print("-")
	}
	fmt.Println("+")

	for _, pagename := range wblocks {
		fmt.Printf("|%-30s|\n", " "+pagename.Label)
	}

	fmt.Print("+")
	for i := 0; i < maxwidth; i++ {
		fmt.Print("-")
	}
	fmt.Println("+")

}

func (m *MemoryManagementUnit) RegisterListener(l *MemoryListener) {

	// log2.Printf("Register listener: %.4x - %.4x", l.Start, l.End)

	startbank := l.Start / 256
	endbank := l.End / 256

	for bank := startbank; bank <= endbank; bank++ {
		list := m.listeners[bank]
		if list == nil {
			list = []*MemoryListener{}
		}
		list = append(list, l)
		m.listeners[bank] = list
	}
}

func (m *MemoryManagementUnit) ProcessListeners(addr int, value *uint64, action MemoryAction) bool {

	// if addr < m.MinListenerRange || addr > m.MaxListenerRange {
	// 	return true
	// }

	contlisteners := true
	contmemoryupdate := true

	m.NoListeners = true

	bank := (addr % 65536) / 256
	list := m.listeners[bank]
	var l *MemoryListener
	if list != nil {
		for _, l = range list {
			if (action == l.Type || l.Type == MA_ANY) && l.Start <= addr && l.End >= addr {
				// if l.Label == "0x400" {
				// 	log2.Printf("listener %s triggering for addr %.4x", l.Label, addr)
				// }
				contlisteners, contmemoryupdate = l.Target.ProcessEvent(l.Label, addr, value, action)
				if !contlisteners {
					m.NoListeners = false
					return contmemoryupdate
				}
			}
		}
	}

	m.NoListeners = false

	return contmemoryupdate

}

func (m *MemoryManagementUnit) PageSetREAD(bank int, mb *MemoryBlock) {
	m.PageREAD[bank] = mb
	m.LastMem = mb
}

func (m *MemoryManagementUnit) PageSetWRITE(bank int, mb *MemoryBlock) {
	m.PageWRITE[bank] = mb
}

type MemoryListener struct {
	Type   MemoryAction
	Start  int
	End    int
	Label  string
	Target MemoryEventProcessor // must define ProcessMemoryEvent(name, addr, action)
}
