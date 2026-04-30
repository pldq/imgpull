package display

import (
	"context"
	"strings"
)

// EventStatus indicates the status of an action
type EventStatus int

const (
	// Working means that the current task is working
	Working EventStatus = iota
	// Done means that the current task is done
	Done
	// Warning means that the current task has warning
	Warning
	// Error means that the current task has errored
	Error
)

// Event represents status change and progress for a compose resource.
type Event struct {
	ID          string
	Text        string
	Details     string
	Status      EventStatus
	Current     int64
	Percent     int
	Total       int64
	detailParts []string
}

func (e *Event) AppendDetailPart(part string) {
	e.detailParts = append(e.detailParts, part)
}

func (e *Event) ContactDetail() {
	e.Details = strings.Join(e.detailParts, "\r\n")
}

func (e *Event) StatusText() string {
	switch e.Status {
	case Working:
		return "Working"
	case Warning:
		return "Warning"
	case Done:
		return "Done"
	default:
		return "Error"
	}
}

func (e *Event) WithError(err error) *Event {
	e.Status = Error
	e.Details = err.Error()
	return e
}

func (e *Event) ClearProgress() *Event {
	e.Total = 0
	e.Percent = 0
	e.Current = 0
	return e
}

// EventProcessor is notified about Compose operations and tasks
type EventProcessor interface {
	// Start is triggered as a Compose operation is starting with context
	Start(ctx context.Context, event Event)
	// Done is triggered as a Compose operation completed
	Done()
	On(event Event)
	OnText(event *Event, text string)
	OnDetails(event *Event, details string)
}
