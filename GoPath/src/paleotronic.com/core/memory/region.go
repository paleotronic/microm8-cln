package memory

import "paleotronic.com/log"

type MRAccessType int

const (
	MR_ACCESS_GLOBAL      MRAccessType = 0
	MR_ACCESS_INTERPRETER MRAccessType = 1
	MR_ACCESS_LOCAL       MRAccessType = 2
)

/* IntelligentAccessHandler is a ondemand realtime assistant for memory mapping
   where an immediate processing is required due to for example a softswitch */
type WriteSubscriptionHandler func(mm *MappedRegion, offset int, value uint64)
type ReadSubscriptionHandler func(mm *MappedRegion, offset int) uint64
type ExecSubscriptionHandler func(mm *MappedRegion, offset int)

/* MappedRegion represents a block of memory or a device and can include special
access handlers for softswitches */
type MappedRegion struct {
	Global             *MemoryMap
	Label              string // human label for reference (in logging etc)
	Data               *MemoryControlBlock
	Base               int
	GlobalBase         int
	Size               int
	Dirty              bool
	SpecialAccessExec  [256]ExecSubscriptionHandler
	SpecialAccessRead  [256]ReadSubscriptionHandler
	SpecialAccessWrite [256]WriteSubscriptionHandler
	Disabled           bool
	MaskEnabled        bool
	Owner              interface{}
}

/* NewMapppedRegion returns a new memory-mapped region, which is a window
on the global address space */
func NewMappedRegion(m *MemoryMap, index int, globalbase int, base int, size int, label string, r_map [256]ReadSubscriptionHandler, e_map [256]ExecSubscriptionHandler, w_map [256]WriteSubscriptionHandler) *MappedRegion {

	s := NewMemoryControlBlock(m, index, false)
	s.Add(m.Data[index][globalbase+base:globalbase+base+size], globalbase+base)

	this := &MappedRegion{
		Global:             m,
		GlobalBase:         globalbase,
		Data:               s,
		Base:               base,
		Size:               size,
		Dirty:              true,
		SpecialAccessRead:  r_map,
		SpecialAccessExec:  e_map,
		SpecialAccessWrite: w_map,
		Label:              label,
		Disabled:           false,
	}

	return this

}

func NewMappedRegionFromMCB(m *MemoryMap, globalbase int, base int, size int, label string, r_map [256]ReadSubscriptionHandler, e_map [256]ExecSubscriptionHandler, w_map [256]WriteSubscriptionHandler, s *MemoryControlBlock) *MappedRegion {

	this := &MappedRegion{
		Global:             m,
		GlobalBase:         globalbase,
		Data:               s,
		Base:               base,
		Size:               s.Size,
		Dirty:              true,
		SpecialAccessRead:  r_map,
		SpecialAccessExec:  e_map,
		SpecialAccessWrite: w_map,
		Label:              label,
		Disabled:           false,
	}

	return this

}

func NewMappedRegionFromHint(m *MemoryMap, globalbase int, base int, size int, label string, hint string, r_map [256]ReadSubscriptionHandler, e_map [256]ExecSubscriptionHandler, w_map [256]WriteSubscriptionHandler) *MappedRegion {

	this := &MappedRegion{
		Global:             m,
		GlobalBase:         globalbase,
		Data:               m.GetHintedMemorySlice((globalbase / OCTALYZER_INTERPRETER_SIZE), hint),
		Base:               base,
		Size:               size,
		Dirty:              true,
		SpecialAccessRead:  r_map,
		SpecialAccessExec:  e_map,
		SpecialAccessWrite: w_map,
		Label:              label,
		Disabled:           false,
	}

	return this

}

func (mr *MappedRegion) ProcessEvent(name string, addr int, action MemoryAction) {
	panic("this should be overridden")
}

func (mr *MappedRegion) IsEnabled() bool {
	return !mr.Disabled
}

func (mr *MappedRegion) SetEnabled(v bool) {
	mr.Disabled = !v
}

func (mr *MappedRegion) IsMaskEnabled() bool {
	return mr.MaskEnabled
}

func (mr *MappedRegion) SetMaskEnabled(v bool) {
	mr.MaskEnabled = v
}

func (mr *MappedRegion) GetBase() int {
	return mr.Base
}

func (mr *MappedRegion) GetSize() int {
	return mr.Size
}

func (mr *MappedRegion) GetLabel() string {
	return mr.Label
}

/* Returns true if Global address is claimed by this mapper */
func (mr *MappedRegion) ClaimsAddress(address int) bool {
	return (address >= mr.Base && address < mr.Base+mr.Size)
}

/* Returns true if the address range has been updated by a write-op */
func (mr *MappedRegion) IsDirty() bool {
	return mr.Dirty
}

func (mr *MappedRegion) SetDirty(b bool) {
	mr.Dirty = b
}

func (mr *MappedRegion) ReadData(offset int) uint64 {
	return mr.Data.Read(offset)
}

func (mr *MappedRegion) WriteData(offset int, value uint64) {
	mr.Data.Write(offset, value)
}

/*  RelativeWrite handles a write request within this regions address space */
func (mr *MappedRegion) RelativeWrite(offset int, value uint64) {
	if offset >= mr.Size {
		return // ignore write outside our bounds
	}

	mr.Data.Write(offset, value)
	mr.Dirty = true

	f := mr.SpecialAccessWrite[offset]
	if f != nil {
		f(mr, offset, value)
	}
}

/* RelativeRead handles a read within this regions address space */
func (mr *MappedRegion) RelativeRead(offset int) uint64 {
	if offset >= mr.Size {
		return 0 // ignore read outside our bounds
	}

	f := mr.SpecialAccessRead[offset]
	if f != nil {
		return f(mr, offset)
	}
	return mr.Data.Read(offset)
}

/* RelativeExec handles an exec request within this regions address space */
func (mr *MappedRegion) RelativeExec(offset int) {
	if offset >= mr.Size {
		return // ignore exec outside our bounds
	}

	f := mr.SpecialAccessExec[offset]
	if f != nil {
		f(mr, offset)
	} else {
		log.Printf("%s: Undefined exec request at offset %d (%d)\n", mr.Label, offset, mr.Base+offset)
	}
	return
}

func (mr *MappedRegion) SubscribeReadHandler(localaddress int, handler ReadSubscriptionHandler) {
	mr.SpecialAccessRead[localaddress] = handler
}

func (mr *MappedRegion) SubscribeExecHandler(localaddress int, handler ExecSubscriptionHandler) {
	mr.SpecialAccessExec[localaddress] = handler
}

func (mr *MappedRegion) SubscribeWriteHandler(localaddress int, handler WriteSubscriptionHandler) {
	mr.SpecialAccessWrite[localaddress] = handler
}

func (mr *MappedRegion) Read(address int) uint64 {
	return mr.RelativeRead(address - mr.Base)
}

func (mr *MappedRegion) Exec(address int) {
	mr.RelativeExec(address - mr.Base)
}

func (mr *MappedRegion) Write(address int, value uint64) {
	mr.RelativeWrite(address-mr.Base, value)
}
