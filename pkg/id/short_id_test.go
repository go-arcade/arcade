package id

import "testing"

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/20 12:40
 * @file: short_id_test.go
 * @description:
 */

func TestShortId(t *testing.T) {
	id := ShortId()
	if id != "" {
		println(id)
	}
}
