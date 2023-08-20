package message_service

import (
	"errors"
	"fmt"
)

type EventType uint32

const (
	ChatMessage = 0
)

type Response struct {
	Data string
}

type Request interface {
	event() EventType
}

type MessageService interface {
	HandleEvent(req Request) (*Response, error)
}

type UserInfo struct{}

type SimpleRequest struct {
	Sender  UserInfo
	Content string
}

func (r *SimpleRequest) event() EventType {
	return ChatMessage
}

type SampleMessageService struct{}

func (s *SampleMessageService) HandleEvent(req Request) (*Response, error) {
	simpleRequest, ok := req.(*SimpleRequest)

	if req.event() != ChatMessage {
		return nil, errors.New("invalid event type")
	}

	if !ok {
		return nil, errors.New("invalid request type")
	}

	return &Response{
		Data: fmt.Sprintf("response from: %s", simpleRequest.Content),
	}, nil

}
