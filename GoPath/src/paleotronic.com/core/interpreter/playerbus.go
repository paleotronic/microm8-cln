package interpreter

import (
	"time"

	"paleotronic.com/debugger/debugtypes"
	"paleotronic.com/log"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/servicebus"
)

func (c *Player) HandleServiceBusRequest(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool) {
	log.Printf("Recieved ServiceBus request: %+v", r)

	switch r.Type {
	case servicebus.PlayerBackstep:
		c.AddSeekDelta(5000)
	case servicebus.PlayerPause:
		c.timeFactor = 0
		time.AfterFunc(time.Millisecond, func() {
			cpu := apple2helpers.GetCPU(c.Destination)
			servicebus.SendServiceBusMessage(
				c.Destination.GetMemIndex(),
				servicebus.CPUState,
				&debugtypes.CPUState{
					A:           cpu.A,
					X:           cpu.X,
					Y:           cpu.Y,
					PC:          cpu.PC,
					P:           cpu.P,
					SP:          cpu.SP,
					IsRecording: true, // only set this so we know the lookback/ahead decodes need to come from the player
					ForceUpdate: true,
				},
			)
		})
	case servicebus.PlayerFaster:
		c.Faster()
	case servicebus.PlayerSlower:
		c.Slower()
	case servicebus.PlayerReset:
		c.ResetToStart()
	case servicebus.PlayerResume:
		if c.IsNearEnd() {
			c.timeFactor = 4 // go forwards fast
			c.backwards = false
			log.Printf("resuming 4x rolloff")
		} else {
			//if settings.UnifiedRender[c.Destination.MemIndex] {
			//	c.stopNextVSync = true
			//} else {
			c.rtNextSync = true
			//}
			if c.timeFactor == 0 {
				c.timeFactor = 1
			}
		}
	case servicebus.PlayerJump:
		pa := r.Payload.(*servicebus.PlayerJumpCommand)
		c.Jump(pa.SyncCount)
	}

	c.ShowProgress()

	return &servicebus.ServiceBusResponse{
		Payload: "",
	}, true
}

func (c *Player) InjectServiceBusRequest(r *servicebus.ServiceBusRequest) {
	log.Printf("Injecting ServiceBus request: %+v", r)
	c.m.Lock()
	defer c.m.Unlock()
	if c.injectedBusRequests == nil {
		c.injectedBusRequests = make([]*servicebus.ServiceBusRequest, 0, 16)
	}
	c.injectedBusRequests = append(c.injectedBusRequests, r)
}

func (c *Player) HandleServiceBusInjection(handler func(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool)) {
	if c.injectedBusRequests == nil || len(c.injectedBusRequests) == 0 {
		return
	}
	c.m.Lock()
	defer c.m.Unlock()
	for _, r := range c.injectedBusRequests {
		if handler != nil {
			handler(r)
		}
	}
	c.injectedBusRequests = make([]*servicebus.ServiceBusRequest, 0, 16)
}

func (c *Player) ServiceBusProcessPending() {
	c.HandleServiceBusInjection(c.HandleServiceBusRequest)
}
