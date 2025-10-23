package main

import (
	"github.com/go-arcade/arcade/pkg/version"
	"github.com/spf13/cobra"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/4 19:51
 * @file: main.go
 * @description: cli program
 */

var rootCmd = &cobra.Command{
	Use:   "arcade-cli",
	Short: "arcade cli is a command line tool",
	Long:  "arcade cli is a command line tool",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(version.VersionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
