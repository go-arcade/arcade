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
	Version   string `json:"version"`
	GitBranch string `json:"gitBranch"`
	GitCommit string `json:"gitCommit"`
	BuildTime string `json:"buildTime"`
	GoVersion string `json:"goVersion"`
	Compiler  string `json:"compiler"`
	Platform  string `json:"platform"`
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
