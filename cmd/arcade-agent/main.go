package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-arcade/arcade/internal/agent/bootstrap"
)

func main() {
	// Parse command line flags
	configFile := flag.String("conf", "conf.d/agent.toml", "configuration file path, e.g. -conf ./conf.d/agent.toml")
	flag.Parse()

	// Bootstrap initialize application
	app, cleanup, _, err := bootstrap.Bootstrap(*configFile, initAgent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Start application and wait for exit signal
	bootstrap.Run(app, cleanup)
}
