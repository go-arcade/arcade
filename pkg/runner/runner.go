package runner

import (
	"os"
)


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
