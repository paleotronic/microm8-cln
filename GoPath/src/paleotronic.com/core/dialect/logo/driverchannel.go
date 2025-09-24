package logo

import (
	"errors"

	"paleotronic.com/core/types"
)

func (d *LogoDriver) ChannelCreate(name string) {
	_, ok := d.Channels[name]
	if ok {
		return
	}
	d.Channels[name] = make(chan *types.Token, 1024)
}

func (d *LogoDriver) ChannelSend(name string, value *types.Token) error {
	ch, ok := d.Channels[name]
	if !ok {
		return errors.New("channel " + name + " does not exist")
	}
	ch <- value
	return nil
}

func (d *LogoDriver) ChannelRecv(name string) (*types.Token, error) {
	ch, ok := d.Channels[name]
	if !ok {
		return nil, errors.New("channel " + name + " does not exist")
	}
	v := <-ch
	return v, nil
}
