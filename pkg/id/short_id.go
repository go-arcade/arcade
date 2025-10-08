package id

import "github.com/teris-io/shortid"

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/20 12:38
 * @file: short_id.go
 * @description:
 */

func ShortId() string {
	id, err := shortid.Generate()
	if err != nil {
		return ""
	}
	return id
}
