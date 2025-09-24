package s8webclient

import (
	"time"

	"github.com/golang/protobuf/proto"
	"paleotronic.com/server/remoteapi"
)

// ReportStatus reports a remint status to the server
func (c *Client) ReportStatus(req *remoteapi.RemoteStatusRequest) (*remoteapi.RemoteStatusResponse, error) {

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	msg, err := c.c.SendMessageAndCatchResponses(
		"RSR",
		buffer,
		true,
		"RSR",
		"ERR",
		time.Second*10,
	)

	// get response
	if msg != nil {
		resp := &remoteapi.RemoteStatusResponse{}
		err = proto.Unmarshal(msg.Payload, resp)
		return resp, err
	}

	return nil, err
}

// ReportStatus reports a remint status to the server
func (c *Client) RequestRemoteList() (*remoteapi.RemoteListResponse, error) {

	req := &remoteapi.RemoteListRequest{}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	msg, err := c.c.SendMessageAndCatchResponses(
		"RLR",
		buffer,
		true,
		"RLR",
		"ERR",
		time.Second*10,
	)

	// get response
	if msg != nil {
		resp := &remoteapi.RemoteListResponse{}
		err = proto.Unmarshal(msg.Payload, resp)
		return resp, err
	}

	return nil, err
}
