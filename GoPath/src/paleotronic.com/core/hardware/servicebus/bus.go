package servicebus

import (
	"sync"

	"paleotronic.com/core/settings"
	"paleotronic.com/fmt"
)

type ServiceBusRequestType int

const (
	DiskIIEject ServiceBusRequestType = iota
	DiskIIExchangeDisks
	DiskIIInsertBlank
	DiskIIInsertFilename
	DiskIIInsertBytes
	DiskIIToggleWriteProtect
	DiskIIQueryWriteProtect
	DiskIIFlush
	BBCKeyPressed
	LaunchEmulator
	CPUState
	CPUControl
	VideoSoftSwitchState
	MemorySoftSwitchState
	SmartPortInsertFilename
	SmartPortInsertBytes
	SmartPortEject
	SmartPortFlush
	CPUWriteMem
	CPUReadMem
	LiveRewindStateUpdate
	MousePosition
	MouseButton
	BeginVBLANK
	LaunchControl
	// Playback
	PlayerPause
	PlayerFaster
	PlayerSlower
	PlayerResume
	PlayerJump
	PlayerReset
	PlayerBackstep
	// tracker
	TrackerLoadSong
	// microM8
	ExecutorShutdown
	//
	SpectrumLoadSnapshot
	KeyEvent
	JoyEvent
	Z80SpeedChange
	SpectrumSaveSnapshot
	// CPU Report cycles
	Apple2AttachTape
	Cycles6502Update
	Clocks6502Update
	UnifiedPlaybackSync
	UnifiedScanUpdate
	UnifiedVBLANK
	RecorderCPUState
	RecorderSuspendEmulation
	RecorderTerminate
	// always last
	LastSBRT
)

var registry [settings.NUMSLOTS][LastSBRT][]ServiceBusSubscriber
var m [settings.NUMSLOTS]sync.RWMutex

type ServiceBusRequest struct {
	VM      int
	SlotID  int
	Type    ServiceBusRequestType
	Payload interface{} // dynamic payload
}

func (r ServiceBusRequest) String() string {
	return fmt.Sprintf("Type: %.2x, Payload: %v", r.Type, r.Payload)
}

type ServiceBusResponse struct {
	Type    ServiceBusRequestType
	Payload interface{}
}

type ServiceBusSubscriber interface {
	HandleServiceBusRequest(r *ServiceBusRequest) (*ServiceBusResponse, bool)
	InjectServiceBusRequest(r *ServiceBusRequest)
	HandleServiceBusInjection(handler func(r *ServiceBusRequest) (*ServiceBusResponse, bool))
}

func ClearSubscriptions(slotid int) {
	m[slotid].Lock()
	defer m[slotid].Unlock()
	for i, _ := range registry[slotid] {
		registry[slotid][i] = nil
	}
	//registry[slotid] = make(map[ServiceBusRequestType][]ServiceBusSubscriber)
}

func HasReceiver(slotid int, messageType ServiceBusRequestType) bool {
	list := registry[slotid][messageType]
	return len(list) > 0
}

func Subscribe(slotid int, messageType ServiceBusRequestType, receiver ServiceBusSubscriber) {

	m[slotid].Lock()
	defer m[slotid].Unlock()

	list := registry[slotid][messageType]
	if list == nil {
		list = make([]ServiceBusSubscriber, 0)
	}
	for _, v := range list {
		if v == receiver {
			return // already subscribed
		}
	}
	list = append(list, receiver)
	registry[slotid][messageType] = list

	// log.Printf("list for vm %d event type %d is %+v", slotid, messageType, list)

}

func UnsubscribeAll(slotid int) {
	m[slotid].Lock()
	defer m[slotid].Unlock()

	//registry[slotid] = make(map[ServiceBusRequestType][]ServiceBusSubscriber)
	for i, _ := range registry[slotid] {
		registry[slotid][i] = nil
	}
}

func UnsubscribeType(slotid int, t ServiceBusRequestType) {
	m[slotid].Lock()
	defer m[slotid].Unlock()

	//registry[slotid] = make(map[ServiceBusRequestType][]ServiceBusSubscriber)
	registry[slotid][t] = nil
}

func Unsubscribe(slotid int, receiver ServiceBusSubscriber) {
	if slotid > settings.NUMSLOTS || slotid < -1 {
		return
	}
	m[slotid].Lock()
	defer m[slotid].Unlock()

	for messageType, list := range registry[slotid] {
		idx := -1
		for i, v := range list {
			if v == receiver {
				idx = i
				break
			}
		}
		if idx != -1 {
			list = append(list[0:idx], list[idx+1:]...)
			registry[slotid][messageType] = list
		}
	}
}

func SendServiceBusMessage(slotid int, messageType ServiceBusRequestType, payload interface{}) ([]*ServiceBusResponse, bool) {
	if slotid > settings.NUMSLOTS || slotid < 0 {
		return []*ServiceBusResponse{}, false
	}

	m[slotid].RLock()
	defer m[slotid].RUnlock()

	var resp []*ServiceBusResponse

	list := registry[slotid][messageType]
	if list != nil {
		for _, subscriber := range list {
			r, done := subscriber.HandleServiceBusRequest(&ServiceBusRequest{
				Payload: payload,
				SlotID:  slotid,
				Type:    messageType,
			})
			resp = append(resp, r)
			if done {
				return resp, true
			}
		}
	}

	return resp, len(resp) != 0
}

func InjectServiceBusMessage(slotid int, messageType ServiceBusRequestType, payload interface{}) {
	if slotid > settings.NUMSLOTS || slotid < 0 {
		return
	}

	m[slotid].RLock()
	defer m[slotid].RUnlock()

	list := registry[slotid][messageType]
	if list != nil {
		for _, subscriber := range list {
			// if messageType == LaunchEmulator {
			// 	log.Printf("injecting event type %d to vm %d (%+v)", messageType, slotid, subscriber)
			// }
			subscriber.InjectServiceBusRequest(&ServiceBusRequest{
				Payload: payload,
				SlotID:  slotid,
				Type:    messageType,
			})
		}
	}

	return
}
