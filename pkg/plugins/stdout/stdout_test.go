package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStdoutNotify(t *testing.T) {
	plugin := NewStdoutNotify()

	assert.NotNil(t, plugin)
	assert.Equal(t, "stdout", plugin.name)
	assert.Equal(t, "A simple notify plugin that prints messages to stdout", plugin.description)
	assert.Equal(t, "1.0.0", plugin.version)
}

func TestStdoutNotify_Name(t *testing.T) {
	plugin := NewStdoutNotify()
	name, err := plugin.Name()

	assert.NoError(t, err)
	assert.Equal(t, "stdout", name)
}

func TestStdoutNotify_Description(t *testing.T) {
	plugin := NewStdoutNotify()
	desc, err := plugin.Description()

	assert.NoError(t, err)
	assert.Equal(t, "A simple notify plugin that prints messages to stdout", desc)
}

func TestStdoutNotify_Version(t *testing.T) {
	plugin := NewStdoutNotify()
	version, err := plugin.Version()

	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", version)
}

func TestStdoutNotify_Type(t *testing.T) {
	plugin := NewStdoutNotify()
	typ, err := plugin.Type()

	assert.NoError(t, err)
	assert.Equal(t, "custom", typ)
}

func TestStdoutNotify_Init(t *testing.T) {
	tests := []struct {
		name   string
		config StdoutNotifyConfig
	}{
		{
			name: "正常配置",
			config: StdoutNotifyConfig{
				Prefix: "TEST",
				JSON:   true,
			},
		},
		{
			name: "空配置",
			config: StdoutNotifyConfig{
				Prefix: "",
				JSON:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewStdoutNotify()

			configJSON, err := sonic.Marshal(tt.config)
			require.NoError(t, err)

			err = plugin.Init(configJSON)
			assert.NoError(t, err)
			assert.Equal(t, tt.config.Prefix, plugin.cfg.Prefix)
			assert.Equal(t, tt.config.JSON, plugin.cfg.JSON)
		})
	}
}

func TestStdoutNotify_InitWithEmptyConfig(t *testing.T) {
	plugin := NewStdoutNotify()
	err := plugin.Init(json.RawMessage{})

	assert.NoError(t, err)
}

func TestStdoutNotify_InitWithInvalidJSON(t *testing.T) {
	plugin := NewStdoutNotify()

	invalidJSON := json.RawMessage(`{"invalid": json`)
	err := plugin.Init(invalidJSON)

	// 应该不返回错误，而是使用默认配置
	assert.NoError(t, err)
}

func TestStdoutNotify_Cleanup(t *testing.T) {
	plugin := NewStdoutNotify()
	err := plugin.Cleanup()

	assert.NoError(t, err)
}

func TestStdoutNotify_Send(t *testing.T) {
	tests := []struct {
		name           string
		config         StdoutNotifyConfig
		message        any
		expectInOutput []string
	}{
		{
			name: "发送简单字符串消息",
			config: StdoutNotifyConfig{
				Prefix: "INFO",
				JSON:   false,
			},
			message:        "Hello, World!",
			expectInOutput: []string{"[INFO]", "Hello, World!"},
		},
		{
			name: "发送JSON对象消息",
			config: StdoutNotifyConfig{
				Prefix: "BUILD",
				JSON:   true,
			},
			message: map[string]any{
				"project": "myapp",
				"status":  "success",
			},
			expectInOutput: []string{"[BUILD]", "project", "myapp", "status", "success"},
		},
		{
			name: "无前缀发送消息",
			config: StdoutNotifyConfig{
				Prefix: "",
				JSON:   false,
			},
			message:        "Test message",
			expectInOutput: []string{"Test message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 捕获标准输出
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			plugin := NewStdoutNotify()
			configJSON, _ := sonic.Marshal(tt.config)
			plugin.Init(configJSON)

			messageJSON, err := sonic.Marshal(tt.message)
			require.NoError(t, err)

			err = plugin.Send(messageJSON, nil)

			// 恢复标准输出
			w.Close()
			os.Stdout = oldStdout

			// 读取捕获的输出
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			assert.NoError(t, err)
			for _, expected := range tt.expectInOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestStdoutNotify_SendTemplate(t *testing.T) {
	tests := []struct {
		name           string
		config         StdoutNotifyConfig
		template       string
		data           map[string]any
		expectInOutput []string
		expectError    bool
	}{
		{
			name: "简单模板渲染",
			config: StdoutNotifyConfig{
				Prefix: "DEPLOY",
			},
			template: "Deploy {{.ProjectName}} to {{.Environment}}",
			data: map[string]any{
				"ProjectName": "myapp",
				"Environment": "production",
			},
			expectInOutput: []string{"[DEPLOY]", "Deploy myapp to production"},
			expectError:    false,
		},
		{
			name: "带条件的模板",
			config: StdoutNotifyConfig{
				Prefix: "STATUS",
			},
			template: "Build {{.ProjectName}}: {{if .Success}}SUCCESS{{else}}FAILED{{end}}",
			data: map[string]any{
				"ProjectName": "testapp",
				"Success":     true,
			},
			expectInOutput: []string{"[STATUS]", "Build testapp: SUCCESS"},
			expectError:    false,
		},
		{
			name: "无效的模板语法",
			config: StdoutNotifyConfig{
				Prefix: "ERROR",
			},
			template:    "Invalid {{.Missing",
			data:        map[string]any{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 捕获标准输出
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			plugin := NewStdoutNotify()
			configJSON, _ := sonic.Marshal(tt.config)
			plugin.Init(configJSON)

			dataJSON, err := sonic.Marshal(tt.data)
			require.NoError(t, err)

			err = plugin.SendTemplate(tt.template, dataJSON, nil)

			// 恢复标准输出
			w.Close()
			os.Stdout = oldStdout

			// 读取捕获的输出
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, expected := range tt.expectInOutput {
					assert.Contains(t, output, expected)
				}
			}
		})
	}
}

func TestStdoutNotify_SendTemplateWithEmptyData(t *testing.T) {
	// 捕获标准输出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	plugin := NewStdoutNotify()
	plugin.Init(json.RawMessage{})

	err := plugin.SendTemplate("Static message", json.RawMessage{}, nil)

	// 恢复标准输出
	w.Close()
	os.Stdout = oldStdout

	// 读取捕获的输出
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "Static message")
}

func TestStdoutNotify_Execute(t *testing.T) {
	tests := []struct {
		name        string
		action      string
		params      map[string]any
		expectError bool
	}{
		{
			name:   "执行send动作",
			action: "send",
			params: map[string]any{
				"message": "Test message",
			},
			expectError: false,
		},
		{
			name:   "执行template动作",
			action: "template",
			params: map[string]any{
				"template": "Test {{.Value}}",
				"data": map[string]any{
					"Value": "123",
				},
			},
			expectError: false,
		},
		{
			name:        "未知动作",
			action:      "unknown",
			params:      map[string]any{},
			expectError: true,
		},
		{
			name:   "send动作带空消息",
			action: "send",
			params: map[string]any{
				"message": "",
			},
			expectError: false,
		},
		{
			name:   "template动作带空数据",
			action: "template",
			params: map[string]any{
				"template": "Test static message",
				"data":     map[string]any{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 捕获标准输出
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			plugin := NewStdoutNotify()
			plugin.Init(json.RawMessage{})

			paramsJSON, err := sonic.Marshal(tt.params)
			require.NoError(t, err)

			result, err := plugin.Execute(tt.action, paramsJSON, nil)

			// 恢复标准输出
			w.Close()
			os.Stdout = oldStdout

			// 清空管道
			var buf bytes.Buffer
			io.Copy(&buf, r)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				var resultMap map[string]any
				err = sonic.Unmarshal(result, &resultMap)
				require.NoError(t, err)
				assert.Equal(t, "sent", resultMap["status"])
			}
		})
	}
}

func TestStdoutNotify_ExecuteWithInvalidParams(t *testing.T) {
	plugin := NewStdoutNotify()
	plugin.Init(json.RawMessage{})

	invalidJSON := json.RawMessage(`{"invalid": json`)
	_, err := plugin.Execute("send", invalidJSON, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestStdoutNotify_SendJSONFormatting(t *testing.T) {
	// 捕获标准输出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	plugin := NewStdoutNotify()
	config := StdoutNotifyConfig{
		Prefix: "JSON",
		JSON:   true,
	}
	configJSON, _ := sonic.Marshal(config)
	plugin.Init(configJSON)

	message := map[string]any{
		"level":   "info",
		"message": "test",
		"count":   42,
	}
	messageJSON, _ := sonic.Marshal(message)

	err := plugin.Send(messageJSON, nil)

	// 恢复标准输出
	w.Close()
	os.Stdout = oldStdout

	// 读取捕获的输出
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "[JSON]")
	assert.Contains(t, output, "level")
	assert.Contains(t, output, "info")
	assert.Contains(t, output, "message")
	assert.Contains(t, output, "test")
}

func TestStdoutNotify_SendStringMessage(t *testing.T) {
	// 捕获标准输出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	plugin := NewStdoutNotify()
	plugin.Init(json.RawMessage{})

	// 发送纯字符串消息（JSON编码）
	messageJSON, _ := sonic.Marshal("Simple string message")

	err := plugin.Send(messageJSON, nil)

	// 恢复标准输出
	w.Close()
	os.Stdout = oldStdout

	// 读取捕获的输出
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "Simple string message")
}

func TestStdoutNotify_OutputFormat(t *testing.T) {
	// 捕获标准输出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	plugin := NewStdoutNotify()
	config := StdoutNotifyConfig{
		Prefix: "TEST",
	}
	configJSON, _ := sonic.Marshal(config)
	plugin.Init(configJSON)

	messageJSON, _ := sonic.Marshal("test message")
	err := plugin.Send(messageJSON, nil)

	// 恢复标准输出
	w.Close()
	os.Stdout = oldStdout

	// 读取捕获的输出
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)

	// 验证输出格式：[PREFIX] TIMESTAMP | MESSAGE
	assert.Contains(t, output, "[TEST]")
	assert.Contains(t, output, "|")
	assert.Contains(t, output, "test message")

	// 验证包含时间戳（RFC3339格式包含T和Z）
	assert.True(t, strings.Contains(output, "T") || strings.Contains(output, ":"))
}

func TestStdoutNotify_TemplateWithComplexData(t *testing.T) {
	// 捕获标准输出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	plugin := NewStdoutNotify()
	plugin.Init(json.RawMessage{})

	template := "Project: {{.Project}}, Users: {{range .Users}}{{.}}, {{end}}Status: {{.Status}}"
	data := map[string]any{
		"Project": "myapp",
		"Users":   []string{"user1", "user2", "user3"},
		"Status":  "active",
	}
	dataJSON, _ := sonic.Marshal(data)

	err := plugin.SendTemplate(template, dataJSON, nil)

	// 恢复标准输出
	w.Close()
	os.Stdout = oldStdout

	// 读取捕获的输出
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "Project: myapp")
	assert.Contains(t, output, "user1")
	assert.Contains(t, output, "user2")
	assert.Contains(t, output, "user3")
	assert.Contains(t, output, "Status: active")
}

func TestStdoutNotify_MultipleMessages(t *testing.T) {
	// 捕获标准输出
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	plugin := NewStdoutNotify()
	plugin.Init(json.RawMessage{})

	messages := []string{"message1", "message2", "message3"}

	for _, msg := range messages {
		msgJSON, _ := sonic.Marshal(msg)
		err := plugin.Send(msgJSON, nil)
		assert.NoError(t, err)
	}

	// 恢复标准输出
	w.Close()
	os.Stdout = oldStdout

	// 读取捕获的输出
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// 验证所有消息都被输出
	for _, msg := range messages {
		assert.Contains(t, output, msg)
	}

	// 验证输出了多行（每条消息一行）
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, len(messages), len(lines))
}
