package mqclient

/*
This is the client library for the mqpiggy queue system.

mqc, err := mqclient.New( "hostname", port )

*/

import (
	"errors"
	"net/rpc"
	"paleotronic.com/postoffice/core"
)

type MQClient struct {
	Host   string
	Port   string
	Client *rpc.Client
}

func New(hostname, port string) (*MQClient, error) {
	this := &MQClient{Host: hostname, Port: port}

	c, err := rpc.Dial("tcp", this.Host+":"+this.Port)
	if err != nil {
		return nil, errors.New("Unable to connect")
	}

	this.Client = c

	return this, nil
}

func (this *MQClient) Put(queuename string, payload []byte, tag, sid string) error {

	var request postoffice.PutArgs
	var response postoffice.MQResponse

	request.Queue = queuename
	request.Payload = payload
	request.Tag = tag
	request.SecretID = sid

	err := this.Client.Call("MQ.Put", request, &response)

	return err
}

func (this *MQClient) Lease(queuename, tag string, duration int, max int) ([]postoffice.QueueEntry, error) {
	var request postoffice.LeaseArgs
	var response postoffice.MQResponse

	request.Queue = queuename
	request.Tag = tag
	request.Time = duration
	request.Max = max

	response = postoffice.MQResponse{}

	err := this.Client.Call("MQ.Lease", request, &response)

	return response.Data, err
}

func (this *MQClient) Remove(queuename string, entries []postoffice.QueueEntry) error {
	for _, qe := range entries {
		var request postoffice.DeleteArgs
		var response postoffice.MQResponse

		request.Queue = queuename
		request.RequestID = qe.ID
		request.SecretID = qe.SecretID

		err := this.Client.Call("MQ.Delete", request, &response)

		if err != nil {
			return err
		}
	}
	return nil
}

func (this *MQClient) Close() {
	this.Client.Close()
}

type Bytable interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
}

// JSON byte encode thing
func Wrap(in Bytable) []byte {
	bb, _ := in.MarshalBinary()
	return bb
}

func Unwrap(in []byte, out Bytable) error {
	err := out.UnmarshalBinary(in)
	return err
}
