package safe

import (
	"fmt"
	"runtime/debug"
)

// Go starts a new goroutine to run the given function f safely.
func Go(f func()) {
	go Do(f)
}

// Do runs the given function f and recovers from any panic, printing the stack trace.
func Do(f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic: %v\n", r)
			debug.PrintStack()
		}
	}()
	f()
}
