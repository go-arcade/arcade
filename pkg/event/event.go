package event

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 12:19
 * @file: event.go
 * @description:
 */

type Event interface {
	// EventName returns the name of the event
	EventName() string
	// EventType returns the type of the event
	EventType() string
}

type EventHandler interface {
	Handle(event Event)
}
