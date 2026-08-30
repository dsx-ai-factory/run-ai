package displayer

import (
	"time"
)

// Builder
type EventBuilder struct {
	event Event
}

func NewEventBuilder() *EventBuilder {
	return &EventBuilder{event: Event{}}
}

func (b *EventBuilder) CreatedAt(createdAt time.Time) *EventBuilder {
	b.event.CreatedAt = createdAt
	return b
}

func (b *EventBuilder) Id(id string) *EventBuilder {
	b.event.Id = &id
	return b
}

func (b *EventBuilder) Message(message string) *EventBuilder {
	b.event.Message = message
	return b
}

func (b *EventBuilder) Reason(reason string) *EventBuilder {
	b.event.Reason = &reason
	return b
}

func (b *EventBuilder) Source(source string) *EventBuilder {
	b.event.Source = &source
	return b
}

func (b *EventBuilder) InvolvedObject(io InvolvedObject) *EventBuilder {
	b.event.InvolvedObject = &io
	return b
}

func (b *EventBuilder) Build() Event {
	return b.event
}

type InvolvedObjectBuilder struct {
	io InvolvedObject
}

func NewInvolvedObjectBuilder() *InvolvedObjectBuilder {
	return &InvolvedObjectBuilder{io: InvolvedObject{}}
}

func (b *InvolvedObjectBuilder) Uid(uid string) *InvolvedObjectBuilder {
	b.io.Uid = uid
	return b
}

func (b *InvolvedObjectBuilder) Kind(kind string) *InvolvedObjectBuilder {
	b.io.Kind = kind
	return b
}

func (b *InvolvedObjectBuilder) Name(name string) *InvolvedObjectBuilder {
	b.io.Name = name
	return b
}

func (b *InvolvedObjectBuilder) Namespace(namespace string) *InvolvedObjectBuilder {
	b.io.Namespace = namespace
	return b
}

func (b *InvolvedObjectBuilder) Build() InvolvedObject {
	return b.io
}
