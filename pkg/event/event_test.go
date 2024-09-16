package event

import (
	"fmt"
	"testing"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 12:33
 * @file: event_test.go
 * @description:
 */

type TestEvent struct {
	Name   string
	Detail Detail
}

type Detail struct {
	Type string
	Data string
}

func (e TestEvent) EventName() string {
	return e.Name
}

func (e TestEvent) EventType() string {
	return e.Detail.Type
}

type TestHandler struct{}

func (h TestHandler) Handle(event Event) {
	if customEvent, ok := event.(TestEvent); ok {
		fmt.Printf("Handled event: %s with detail: %s\n", customEvent.Name, customEvent.Detail.Type)
	}
}

func TestNewEventBus(t *testing.T) {
	bus := NewEventBus()
	if bus == nil {
		t.Error("NewEventBus failed")
	}

	bus.RegisterHandler("test", TestHandler{})

	event := TestEvent{
		Name: "test",
		Detail: Detail{
			Type: "test",
			Data: "test",
		},
	}

	bus.Publish(event)
}
