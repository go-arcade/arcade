package id

import "testing"


func TestShortId(t *testing.T) {
	id := ShortId()
	if id != "" {
		println(id)
	}
}
