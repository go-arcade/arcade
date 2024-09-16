package event

import "fmt"

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 12:32
 * @file: event_bus.go
 * @description:
 */

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
