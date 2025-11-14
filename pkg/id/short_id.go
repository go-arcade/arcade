package id

import "github.com/teris-io/shortid"


func ShortId() string {
	id, err := shortid.Generate()
	if err != nil {
		return ""
	}
	return id
}
