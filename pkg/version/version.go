package version

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"runtime"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/5 21:30
 * @file: version.go
 * @description: version
 */

var (
	Version   = ""
	GitBranch = ""
	GitCommit = ""
	BuildTime = ""
	GoVersion = ""
	Compiler  = ""
	Platform  = ""
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the application version information",
	Run: func(cmd *cobra.Command, args []string) {
		v := GetVersion()
		fmt.Println(string(v.Json()))
	},
}

type Info struct {
	Version   string `Json:"Version"`
	GitBranch string `Json:"GitBranch"`
	GitCommit string `Json:"GitCommit"`
	BuildTime string `Json:"BuildTime"`
	GoVersion string `Json:"GoVersion"`
	Compiler  string `Json:"Compiler"`
	Platform  string `Json:"Platform"`
}

func (v *Info) string() string {
	return v.GitCommit
}

func GetVersion() *Info {
	return &Info{
		Version:   Version,
		GitBranch: GitBranch,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}

func (v *Info) Json() json.RawMessage {
	j, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return j
}
