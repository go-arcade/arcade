package event

import "fmt"


type EventBus struct {
	handlers map[string][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (eb *EventBus) RegisterHandler(eventName string, handler EventHandler) {
	if _, ok := eb.handlers[eventName]; !ok {
		eb.handlers[eventName] = make([]EventHandler, 0)
	}
	eb.handlers[eventName] = append(eb.handlers[eventName], handler)
}

func (eb *EventBus) Publish(event Event) {
	eventName := event.EventName()
	if handlers, ok := eb.handlers[eventName]; ok {
		fmt.Println("event:", eb)
		for _, handler := range handlers {
			handler.Handle(event)
		}
	}
}

func (eb *EventBus) Consume(event Event) {
	eb.Publish(event)
}
