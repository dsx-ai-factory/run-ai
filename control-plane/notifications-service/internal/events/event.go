package events

import "context"

type UID string

type Event struct {
	RawEvent   interface{}
	UID        UID
	ListenerId string
	Properties map[string]interface{}
}

func NewEvent() *Event {
	return &Event{
		Properties: make(map[string]interface{}),
	}
}

type EventHandler interface {
	Handle(context.Context, *Event) error
}
