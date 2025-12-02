package safe

import "testing"

func TestDo(t *testing.T) {
	panicFunc := func() {
		panic("test panic")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Do did not recover from panic: %v", r)
		}
	}()

	Do(panicFunc)
}

func TestGo(t *testing.T) {
	done := make(chan bool)
	panicFunc := func() {
		defer func() {
			done <- true
		}()
		panic("test panic in goroutine")
	}

	Go(panicFunc)
	<-done
}
