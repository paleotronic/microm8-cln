package s8webclient

import (
	"errors"
	"time"

	"github.com/golang/protobuf/proto"
	"paleotronic.com/log"
	"paleotronic.com/server/chatapi"
)

func (c *Client) FetchChatMessages(chat_id int32, count int) (*chatapi.FetchMessagesResponse, error) {

	req := &chatapi.FetchMessagesRequest{
		ChatId: chat_id,
		Count:  int32(count),
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	msg, err := c.c.SendMessageAndCatchResponses(
		"FKM",
		buffer,
		true,
		"FKM",
		"ERR",
		time.Second*10,
	)

	// get response
	if msg != nil {
		resp := &chatapi.FetchMessagesResponse{}
		err = proto.Unmarshal(msg.Payload, resp)
		return resp, err
	}

	return nil, err
}

func (c *Client) FetchChats() (*chatapi.FetchChatsResponse, error) {

	var err error

	buffer := []byte(c.Session)

	msg, err := c.c.SendMessageAndCatchResponses(
		"FKL",
		buffer,
		true,
		"FKL",
		"ERR",
		time.Second*10,
	)

	// get response
	if msg != nil {
		resp := &chatapi.FetchChatsResponse{}
		err = proto.Unmarshal(msg.Payload, resp)
		return resp, err
	}

	return nil, err
}

func (c *Client) PostChatMessage(chat_id int32, body string, isAction bool) (*chatapi.PostChatMessageResponse, error) {

	req := &chatapi.PostChatMessageRequest{
		ChatId:   chat_id,
		Message:  body,
		IsAction: isAction,
	}

	log.Printf("Sending post request: %+v", req)

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	msg, err := c.c.SendMessageAndCatchResponses(
		"PKM",
		buffer,
		true,
		"PKM",
		"ERR",
		time.Second*10,
	)

	// get response
	if msg != nil {
		resp := &chatapi.PostChatMessageResponse{}
		err = proto.Unmarshal(msg.Payload, resp)
		return resp, err
	}

	return nil, err
}

func (c *Client) JoinChat(chat_id int32) (*chatapi.JoinChatResponse, error) {

	req := &chatapi.JoinChatRequest{
		ChatId: chat_id,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	msg, err := c.c.SendMessageAndCatchResponses(
		"JKR",
		buffer,
		true,
		"JKR",
		"ERR",
		time.Second*10,
	)

	// get response
	if msg != nil {
		resp := &chatapi.JoinChatResponse{}
		err = proto.Unmarshal(msg.Payload, resp)
		return resp, err
	}

	return nil, err
}

func (c *Client) LeaveChat(chat_id int32) (*chatapi.LeaveChatResponse, error) {

	req := &chatapi.LeaveChatRequest{
		ChatId: chat_id,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	msg, err := c.c.SendMessageAndCatchResponses(
		"LKR",
		buffer,
		true,
		"LKR",
		"ERR",
		time.Second*10,
	)

	// get response
	if msg != nil {
		resp := &chatapi.LeaveChatResponse{}
		err = proto.Unmarshal(msg.Payload, resp)
		return resp, err
	}

	return nil, err
}

func (c *Client) AwayChat(chat_id int32) (*chatapi.LeaveChatResponse, error) {

	req := &chatapi.LeaveChatRequest{
		ChatId: chat_id,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	msg, err := c.c.SendMessageAndCatchResponses(
		"SKA",
		buffer,
		true,
		"SKA",
		"ERR",
		time.Second*10,
	)

	// get response
	if msg != nil {
		resp := &chatapi.LeaveChatResponse{}
		err = proto.Unmarshal(msg.Payload, resp)
		return resp, err
	}

	return nil, err
}

func (c *Client) ChangeChatTopic(chat_id int32, topic string) (*chatapi.UpdateTopicResponse, error) {

	req := &chatapi.UpdateTopicRequest{
		ChatId: chat_id,
		Topic:  topic,
	}

	log.Printf("Sending topic update request: %+v", req)

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	buffer := []byte(c.Session)
	buffer = append(buffer, data...)

	c.c.SendMessage("CKT", buffer, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		if msg.ID == "CKT" {
			resp := &chatapi.UpdateTopicResponse{}
			err = proto.Unmarshal(msg.Payload, resp)
			log.Printf("Got topic response: %+v", resp)
			return resp, err
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return nil, err
}
