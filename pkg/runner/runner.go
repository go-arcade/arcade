package runner

import (
	"os"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 18:05
 * @file: runner.go
 * @description:
 */

var (
	Hostname string
	Pwd      string
)

func Noop(string, ...interface{}) {

}

func init() {

	var err error
	Hostname, err = os.Hostname()
	if err != nil {
		Hostname = "unknown"
	}

	Pwd, _ = os.Getwd()
}
