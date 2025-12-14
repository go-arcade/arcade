// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package version

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

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
