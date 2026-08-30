package displayer

import (
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/field"
)

// Type definitions
type InvolvedObject struct {
	Uid       string `json:"uid"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Event struct {
	CreatedAt      time.Time       `json:"createdAt"`
	Id             *string         `json:"id,omitempty"`
	Message        string          `json:"message"`
	Reason         *string         `json:"reason,omitempty"`
	Source         *string         `json:"source,omitempty"`
	InvolvedObject *InvolvedObject `json:"involvedObject,omitempty"`
}

// Displayer
type EventsDisplayer struct {
	events []Event
}

func NewEventsSection(events []Event) *Section {
	return &Section{
		Displayer: &EventsDisplayer{
			events: events,
		},
	}
}

func (e *EventsDisplayer) Subject() string {
	return TitleEvents
}

func (e *EventsDisplayer) Columns() []table.Column {
	return []table.Column{
		{Title: FieldDateTime, Width: EventDateColumnWidth},
		{Title: FieldEvent, Width: EventDefaultColumnWidth},
		{Title: FieldEventIssuer, Width: EventDefaultColumnWidth},
		{Title: FieldComponent, Width: EventDefaultColumnWidth},
		{Title: FieldDetails, Width: EventDetailsColumnWidth},
	}
}

func (e *EventsDisplayer) Data() [][]string {
	var data [][]string
	for _, event := range e.events {
		involvedObjName := ""
		if event.InvolvedObject != nil {
			involvedObjName = event.InvolvedObject.Name
		}
		row := []string{
			event.CreatedAt.Format(DateFormat),
			field.StringOrEmpty(event.Reason),
			field.StringOrEmpty(event.Source),
			involvedObjName,
			event.Message,
		}
		data = append(data, row)
	}
	return data
}
