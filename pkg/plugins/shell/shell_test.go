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

package shell

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-arcade/arcade/pkg/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewShellPlugin(t *testing.T) {
	plugin := NewShell()

	assert.NotNil(t, plugin)
	assert.Equal(t, "shell", plugin.name)
	assert.Equal(t, "A custom plugin that executes shell scripts and commands", plugin.description)
	assert.Equal(t, "1.0.0", plugin.version)
	assert.Equal(t, "/bin/sh", plugin.cfg.Shell)
	assert.Equal(t, "", plugin.cfg.WorkDir)
	assert.Equal(t, 300, plugin.cfg.Timeout)
	assert.False(t, plugin.cfg.AllowDangerous)
}

func TestShellPlugin_Name(t *testing.T) {
	plugin := NewShell()
	name := plugin.Name()

	assert.Equal(t, "shell", name)
}

func TestShellPlugin_Description(t *testing.T) {
	plugin := NewShell()
	desc := plugin.Description()

	assert.Equal(t, "A custom plugin that executes shell scripts and commands", desc)
}

func TestShellPlugin_Version(t *testing.T) {
	plugin := NewShell()
	version := plugin.Version()

	assert.Equal(t, "1.0.0", version)
}

func TestShellPlugin_Type(t *testing.T) {
	p := NewShell()
	typ := p.Type()

	assert.Equal(t, plugin.TypeCustom, typ)
}

func TestShellPlugin_Init(t *testing.T) {
	tests := []struct {
		name        string
		config      ShellConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "正常配置",
			config: ShellConfig{
				Shell:   "/bin/sh",
				WorkDir: "/tmp",
				Timeout: 60,
				Env: map[string]string{
					"TEST_VAR": "test_value",
				},
			},
			expectError: false,
		},
		{
			name: "空Shell使用默认值",
			config: ShellConfig{
				Shell: "",
			},
			expectError: false,
		},
		{
			name: "不存在的Shell",
			config: ShellConfig{
				Shell: "/nonexistent/shell",
			},
			expectError: true,
			errorMsg:    "shell not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewShell()

			configJSON, err := sonic.Marshal(tt.config)
			require.NoError(t, err)

			err = plugin.Init(configJSON)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.config.Shell != "" {
					assert.Equal(t, tt.config.Shell, plugin.cfg.Shell)
				} else {
					assert.Equal(t, "/bin/sh", plugin.cfg.Shell)
				}
			}
		})
	}
}

func TestShellPlugin_InitWithEmptyConfig(t *testing.T) {
	plugin := NewShell()
	err := plugin.Init(json.RawMessage{})

	assert.NoError(t, err)
	assert.Equal(t, "/bin/sh", plugin.cfg.Shell)
}

func TestShellPlugin_Cleanup(t *testing.T) {
	plugin := NewShell()
	err := plugin.Cleanup()

	assert.NoError(t, err)
}

func TestShellPlugin_ExecuteScript(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		args        []string
		env         map[string]string
		expectError bool
		checkStdout func(t *testing.T, stdout string)
	}{
		{
			name:        "简单echo脚本",
			script:      "#!/bin/sh\necho 'Hello, World!'",
			expectError: false,
			checkStdout: func(t *testing.T, stdout string) {
				assert.Contains(t, stdout, "Hello, World!")
			},
		},
		{
			name:        "带参数的脚本",
			script:      "#!/bin/sh\necho $1 $2",
			args:        []string{"arg1", "arg2"},
			expectError: false,
			checkStdout: func(t *testing.T, stdout string) {
				assert.Contains(t, stdout, "arg1 arg2")
			},
		},
		{
			name:        "带环境变量的脚本",
			script:      "#!/bin/sh\necho $TEST_VAR",
			env:         map[string]string{"TEST_VAR": "test_value"},
			expectError: false,
			checkStdout: func(t *testing.T, stdout string) {
				assert.Contains(t, stdout, "test_value")
			},
		},
		{
			name:        "空脚本",
			script:      "",
			expectError: true,
		},
		{
			name:        "危险操作-不允许",
			script:      "#!/bin/sh\nrm -rf /",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewShell()
			plugin.Init(json.RawMessage{})

			params := map[string]any{
				"script": tt.script,
			}
			if tt.args != nil {
				params["args"] = tt.args
			}
			if tt.env != nil {
				params["env"] = tt.env
			}

			paramsJSON, err := sonic.Marshal(params)
			require.NoError(t, err)

			result, err := plugin.Execute("script", paramsJSON, nil)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				var resultMap map[string]any
				err = sonic.Unmarshal(result, &resultMap)
				require.NoError(t, err)

				if tt.checkStdout != nil {
					stdout, ok := resultMap["stdout"].(string)
					require.True(t, ok)
					tt.checkStdout(t, stdout)
				}
			}
		})
	}
}

func TestShellPlugin_ExecuteScriptWithDangerousAllowed(t *testing.T) {
	plugin := NewShell()
	config := ShellConfig{
		Shell:          "/bin/sh",
		AllowDangerous: true,
	}
	configJSON, _ := sonic.Marshal(config)
	plugin.Init(configJSON)

	// 即使包含危险模式，如果允许危险操作，也应该能执行
	// 但为了安全，我们只测试一个无害的命令
	params := map[string]any{
		"script": "#!/bin/sh\necho 'test'",
	}
	paramsJSON, _ := sonic.Marshal(params)

	result, err := plugin.Execute("script", paramsJSON, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestShellPlugin_ExecuteCommand(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		env         map[string]string
		expectError bool
		checkStdout func(t *testing.T, stdout string)
	}{
		{
			name:        "简单echo命令",
			command:     "echo 'Hello from command'",
			expectError: false,
			checkStdout: func(t *testing.T, stdout string) {
				assert.Contains(t, stdout, "Hello from command")
			},
		},
		{
			name:        "pwd命令",
			command:     "pwd",
			expectError: false,
			checkStdout: func(t *testing.T, stdout string) {
				assert.NotEmpty(t, stdout)
			},
		},
		{
			name:        "带环境变量的命令",
			command:     "echo $MY_VAR",
			env:         map[string]string{"MY_VAR": "my_value"},
			expectError: false,
			checkStdout: func(t *testing.T, stdout string) {
				assert.Contains(t, stdout, "my_value")
			},
		},
		{
			name:        "空命令",
			command:     "",
			expectError: true,
		},
		{
			name:        "危险命令",
			command:     "rm -rf /",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewShell()
			plugin.Init(json.RawMessage{})

			params := map[string]any{
				"command": tt.command,
			}
			if tt.env != nil {
				params["env"] = tt.env
			}

			paramsJSON, err := sonic.Marshal(params)
			require.NoError(t, err)

			result, err := plugin.Execute("command", paramsJSON, nil)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				var resultMap map[string]any
				err = sonic.Unmarshal(result, &resultMap)
				require.NoError(t, err)

				if tt.checkStdout != nil {
					stdout, ok := resultMap["stdout"].(string)
					require.True(t, ok)
					tt.checkStdout(t, stdout)
				}
			}
		})
	}
}

func TestShellPlugin_ExecuteUnknownAction(t *testing.T) {
	plugin := NewShell()
	plugin.Init(json.RawMessage{})

	_, err := plugin.Execute("unknown_action", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown action")
}

func TestShellPlugin_ExecuteWithTimeout(t *testing.T) {
	plugin := NewShell()
	config := ShellConfig{
		Shell:   "/bin/sh",
		Timeout: 1, // 1秒超时
	}
	configJSON, _ := sonic.Marshal(config)
	plugin.Init(configJSON)

	// 执行一个需要较长时间的命令
	params := map[string]any{
		"command": "sleep 5",
	}
	paramsJSON, _ := sonic.Marshal(params)

	start := time.Now()
	result, err := plugin.Execute("command", paramsJSON, nil)
	duration := time.Since(start)

	// 应该在超时时间内返回
	assert.True(t, duration < 3*time.Second, "command should timeout within 3 seconds")

	if err == nil {
		// 如果没有错误，检查结果
		var resultMap map[string]any
		err = sonic.Unmarshal(result, &resultMap)
		require.NoError(t, err)

		// 检查是否标记为失败
		success, ok := resultMap["success"].(bool)
		if ok {
			assert.False(t, success, "command should fail due to timeout")
		}
	}
}

func TestShellPlugin_ExecuteWithInvalidJSON(t *testing.T) {
	plugin := NewShell()
	plugin.Init(json.RawMessage{})

	invalidJSON := json.RawMessage(`{"invalid": json`)
	_, err := plugin.Execute("script", invalidJSON, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestShellPlugin_SecurityCheck(t *testing.T) {
	plugin := NewShell()
	plugin.Init(json.RawMessage{})

	dangerousPatterns := []string{
		"rm -rf",
		":(){ :|:& };:",
		"mkfs",
		"dd if=",
		"> /dev/",
	}

	for _, pattern := range dangerousPatterns {
		t.Run("dangerous_pattern_"+pattern, func(t *testing.T) {
			params := map[string]any{
				"script": "#!/bin/sh\n" + pattern + " test",
			}
			paramsJSON, _ := sonic.Marshal(params)

			_, err := plugin.Execute("script", paramsJSON, nil)
			assert.Error(t, err)
			assert.Contains(t, strings.ToLower(err.Error()), "dangerous")
		})
	}
}

func TestShellPlugin_ResultStructure(t *testing.T) {
	plugin := NewShell()
	plugin.Init(json.RawMessage{})

	params := map[string]any{
		"command": "echo 'test'",
	}
	paramsJSON, _ := sonic.Marshal(params)

	result, err := plugin.Execute("command", paramsJSON, nil)
	require.NoError(t, err)

	var resultMap map[string]any
	err = sonic.Unmarshal(result, &resultMap)
	require.NoError(t, err)

	// 验证结果结构
	assert.Contains(t, resultMap, "stdout")
	assert.Contains(t, resultMap, "stderr")
	assert.Contains(t, resultMap, "exit_code")
	assert.Contains(t, resultMap, "duration_ms")
	assert.Contains(t, resultMap, "success")

	// 验证类型
	_, ok := resultMap["stdout"].(string)
	assert.True(t, ok)
	_, ok = resultMap["stderr"].(string)
	assert.True(t, ok)
	_, ok = resultMap["success"].(bool)
	assert.True(t, ok)
}

func TestShellPlugin_EnvironmentVariables(t *testing.T) {
	plugin := NewShell()

	// 配置插件级别的环境变量
	config := ShellConfig{
		Shell: "/bin/sh",
		Env: map[string]string{
			"PLUGIN_VAR": "plugin_value",
		},
	}
	configJSON, _ := sonic.Marshal(config)
	plugin.Init(configJSON)

	// 执行时添加额外的环境变量
	params := map[string]any{
		"command": "echo $PLUGIN_VAR $COMMAND_VAR",
		"env": map[string]string{
			"COMMAND_VAR": "command_value",
		},
	}
	paramsJSON, _ := sonic.Marshal(params)

	result, err := plugin.Execute("command", paramsJSON, nil)
	require.NoError(t, err)

	var resultMap map[string]any
	err = sonic.Unmarshal(result, &resultMap)
	require.NoError(t, err)

	stdout := resultMap["stdout"].(string)
	assert.Contains(t, stdout, "plugin_value")
	assert.Contains(t, stdout, "command_value")
}
