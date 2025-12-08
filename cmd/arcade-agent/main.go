package main

import (
	"fmt"
	"os"

	"github.com/go-arcade/arcade/internal/agent/bootstrap"
	"github.com/spf13/cobra"
)

var configFile string

var rootCmd = &cobra.Command{
	Use:   "arcade-agent",
	Short: "Arcade Agent - Agent node that can connect to the Arcade Server to receive and execute tasks",
	Long:  "Arcade Agent is an Agent node that can connect to the Arcade Server to receive and execute tasks.",
	Run: func(cmd *cobra.Command, args []string) {
		// Bootstrap initialize application
		app, cleanup, _, err := bootstrap.Bootstrap(configFile, initAgent)
		if err != nil {
			panic(err)
		}

		// Start application and wait for exit signal
		bootstrap.Run(app, cleanup)
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Agent",
	Long:  "Start Agent and connect to Server",
	Run: func(cmd *cobra.Command, args []string) {
		// Bootstrap initialize application
		app, cleanup, _, err := bootstrap.Bootstrap(configFile, initAgent)
		if err != nil {
			panic(err)
		}

		// Start application and wait for exit signal
		bootstrap.Run(app, cleanup)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "conf", "c", "conf.d/agent.toml", "configuration file path, e.g. -conf ./conf.d/agent.toml")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(RegisterCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, err = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}
