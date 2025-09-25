// Package event contains the ports or interefaces for the event module
package event

type EventConsumer interface {
	JobCreatedEvent()
}
