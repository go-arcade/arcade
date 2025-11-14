package event


type Event interface {
	// EventName returns the name of the event
	EventName() string
	// EventType returns the type of the event
	EventType() string
}

type EventHandler interface {
	Handle(event Event)
}
