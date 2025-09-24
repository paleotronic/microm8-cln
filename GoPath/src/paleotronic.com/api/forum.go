package s8webclient

import (
	"errors"
	"time"

	"github.com/golang/protobuf/proto"
	"paleotronic.com/server/forumapi"
)

// FFM -> FMR
func (c *Client) FetchForumMessages(forum_id int32, parent_id int32) (*forumapi.FetchMessagesResponse, error) {

	req := &forumapi.FetchMessagesRequest{
		ForumId:  forum_id,
		ParentId: parent_id,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	c.c.SendMessage("FFM", buffer, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		if msg.ID == "FMR" {
			resp := &forumapi.FetchMessagesResponse{}
			err = proto.Unmarshal(msg.Payload, resp)
			return resp, err
		} else if msg.ID == "ERR" {
			err = errors.New("i/o error")
		}
	}

	return nil, err
}

// FFF -> FFR
func (c *Client) FetchForums() (*forumapi.FetchForumsResponse, error) {

	var err error

	buffer := []byte(c.Session)

	c.c.SendMessage("FFF", buffer, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		if msg.ID == "FFR" {
			resp := &forumapi.FetchForumsResponse{}
			err = proto.Unmarshal(msg.Payload, resp)
			return resp, err
		} else if msg.ID == "ERR" {
			err = errors.New("i/o error")
		}
	}

	return nil, err
}

// FFU -> FUR
func (c *Client) FetchForumUnread(forum_id int32) (*forumapi.FetchUnreadMessagesResponse, error) {

	req := &forumapi.FetchUnreadMessagesRequest{
		ForumId: forum_id,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	c.c.SendMessage("FFU", buffer, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		if msg.ID == "FUR" {
			resp := &forumapi.FetchUnreadMessagesResponse{}
			err = proto.Unmarshal(msg.Payload, resp)
			return resp, err
		} else if msg.ID == "ERR" {
			err = errors.New("i/o error")
		}
	}

	return nil, err
}

// FPM -> FPR
func (c *Client) PostMessage(forum_id int32, parent_id int32, subject, body string) (*forumapi.PostMessageResponse, error) {

	req := &forumapi.PostMessageRequest{
		ForumId:  forum_id,
		ParentId: parent_id,
		Subject:  subject,
		Body:     body,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	c.c.SendMessage("FPM", buffer, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		if msg.ID == "FPR" {
			resp := &forumapi.PostMessageResponse{}
			err = proto.Unmarshal(msg.Payload, resp)
			return resp, err
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return nil, err
}

// FMS -> FSR
func (c *Client) FetchMessage(forum_id int32, message_id int32) (*forumapi.FetchMessageResponse, error) {

	req := &forumapi.FetchMessageRequest{
		ForumId:   forum_id,
		MessageId: message_id,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	c.c.SendMessage("FMS", buffer, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		if msg.ID == "FSR" {
			resp := &forumapi.FetchMessageResponse{}
			err = proto.Unmarshal(msg.Payload, resp)
			return resp, err
		} else if msg.ID == "ERR" {
			err = errors.New("i/o error")
		}
	}

	return nil, err
}

// FMS -> FSR
func (c *Client) MarkMessageRead(forum_id int32, message_id int32) (*forumapi.MarkMessageResponse, error) {

	req := &forumapi.MarkMessageRequest{
		ForumId:   forum_id,
		MessageId: message_id,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	c.c.SendMessage("FMR", buffer, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		if msg.ID == "FRR" {
			resp := &forumapi.MarkMessageResponse{}
			err = proto.Unmarshal(msg.Payload, resp)
			return resp, err
		} else if msg.ID == "ERR" {
			err = errors.New("i/o error")
		}
	}

	return nil, err
}

func (c *Client) SearchForumMessages(forum_id int32, searchTerm string) (*forumapi.SearchForumResponse, error) {

	req := &forumapi.SearchForumRequest{
		ForumId:    forum_id,
		SearchTerm: searchTerm,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	c.c.SendMessage("FSQ", buffer, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		if msg.ID == "FSR" {
			resp := &forumapi.SearchForumResponse{}
			err = proto.Unmarshal(msg.Payload, resp)
			return resp, err
		} else if msg.ID == "ERR" {
			err = errors.New("i/o error")
		}
	}

	return nil, err
}

func (c *Client) FetchForumsWithNewActivity() (*forumapi.ForumsWithNewActivityResponse, error) {

	var err error

	buffer := []byte(c.Session)

	msg, err := c.c.SendMessageAndCatchResponses(
		"FNA",
		buffer,
		true,
		"FNA",
		"ERR",
		time.Second*10,
	)

	// get response
	if msg != nil {
		resp := &forumapi.ForumsWithNewActivityResponse{}
		err = proto.Unmarshal(msg.Payload, resp)
		return resp, err
	}

	return nil, err
}
