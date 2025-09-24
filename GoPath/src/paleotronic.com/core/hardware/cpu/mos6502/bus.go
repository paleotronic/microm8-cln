package mos6502

import (
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/log"
)

func (d *Core6502) HandleServiceBusRequest(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool) {
	switch r.Type {
	case servicebus.CPUControl:
		data := r.Payload.(*servicebus.CPUControlData)
		switch data.Action {
		case "breakpoint":
			d.RunState = CrsPaused
		case "pause":
			d.RunState = CrsPaused
		case "step":
			d.RunState = CrsSingleStep
		case "step-over":
			d.RunState = CrsStepOver
			d.StepOverSP = data.Data["sp-level"].(int)
			d.StepOverAddr = data.Data["pc"].(int)
		case "continue-out":
			d.RunState = CrsFreeRun
			d.PauseNextRTS = true
		case "continue":
			d.RunState = CrsFreeRun
		}
	}
	return &servicebus.ServiceBusResponse{
		Payload: "",
	}, true
}

func (d *Core6502) InjectServiceBusRequest(r *servicebus.ServiceBusRequest) {
	log.Printf("Injecting ServiceBus request: %+v", r)
	d.Lock()
	defer d.Unlock()
	if d.events == nil {
		d.events = make([]*servicebus.ServiceBusRequest, 0, 16)
	}
	d.events = append(d.events, r)
}

func (d *Core6502) HandleServiceBusInjection(handler func(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool)) {
	if d.events == nil || len(d.events) == 0 {
		return
	}
	d.Lock()
	defer d.Unlock()
	for _, r := range d.events {
		if handler != nil {
			handler(r)
		}
	}
	d.events = make([]*servicebus.ServiceBusRequest, 0, 16)
}

func (d *Core6502) ServiceBusProcessPending() {
	d.HandleServiceBusInjection(d.HandleServiceBusRequest)
}
