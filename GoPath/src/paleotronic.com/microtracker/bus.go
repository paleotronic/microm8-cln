package microtracker

import (
	"bytes"
	"io/ioutil"

	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/files"
	"paleotronic.com/log"
)

func (d *MicroTracker) LoadSong(filename string) {

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	log.Println("read file data")

	err = d.Song.LoadBuffer(bytes.NewBuffer(data))
	if err == nil {
		d.CurrentFile = "/local/" + files.GetFilename(filename)
	}
	log.Printf("Result = %v", err)
	d.refresh = true
}

func (d *MicroTracker) HandleServiceBusRequest(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool) {
	switch r.Type {
	case servicebus.TrackerLoadSong:
		fn := r.Payload.(string)
		log.Printf("Got request to load file: %s", fn)
		d.LoadSong(fn)
	}
	return &servicebus.ServiceBusResponse{
		Payload: "",
	}, true
}

func (d *MicroTracker) InjectServiceBusRequest(r *servicebus.ServiceBusRequest) {
	log.Printf("Injecting ServiceBus request: %+v", r)
	d.Lock()
	defer d.Unlock()
	if d.events == nil {
		d.events = make([]*servicebus.ServiceBusRequest, 0, 16)
	}
	d.events = append(d.events, r)
}

func (d *MicroTracker) HandleServiceBusInjection(handler func(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool)) {
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

func (d *MicroTracker) ServiceBusProcessPending() {
	d.HandleServiceBusInjection(d.HandleServiceBusRequest)
}
