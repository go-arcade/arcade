# 插件包结构规范

## 概述

从 v1.0.0 版本开始，Arcade 插件系统要求插件以 **zip 包** 的形式上传。zip 包必须包含插件的 `.so` 文件和 `manifest.json` 清单文件。

## 包结构

### 标准结构

```
my-plugin.zip
├── plugin.so          # 插件二进制文件（必需）
└── manifest.json      # 插件清单文件（必需）
```

### 可选文件

```
my-plugin.zip
├── plugin.so          # 插件二进制文件（必需）
├── manifest.json      # 插件清单文件（必需）
├── README.md          # 说明文档（可选）
├── LICENSE            # 许可证（可选）
└── config/            # 配置示例（可选）
    └── example.json
```

**注意：**
- ✅ 系统会自动提取 `.so` 文件和 `manifest.json`
- ✅ 其他文件会被忽略，不影响安装
- ✅ 支持嵌套目录，系统会递归查找

## 创建插件包

### 方式一：命令行创建

#### 1. 准备文件

```bash
# 创建工作目录
mkdir my-notify-plugin
cd my-notify-plugin

# 编译插件
go build -buildmode=plugin -o plugin.so main.go

# 创建 manifest.json
cat > manifest.json <<EOF
{
  "name": "my-notify",
  "version": "1.0.0",
  "description": "My notification plugin",
  "author": "Your Name",
  "pluginType": "notify",
  "entryPoint": "plugin.so",
  "configSchema": {
    "type": "object",
    "properties": {
      "webhook_url": {
        "type": "string",
        "description": "Webhook URL"
      }
    },
    "required": ["webhook_url"]
  },
  "defaultConfig": {},
  "tags": ["notify"],
  "resources": {
    "cpu": "100m",
    "memory": "128Mi"
  }
}
EOF
```

#### 2. 打包

```bash
# 创建 zip 包
zip my-notify-plugin.zip plugin.so manifest.json

# 验证 zip 包内容
unzip -l my-notify-plugin.zip
```

输出示例：
```
Archive:  my-notify-plugin.zip
  Length      Date    Time    Name
---------  ---------- -----   ----
  2145678  2025-01-16 10:30   plugin.so
     1234  2025-01-16 10:30   manifest.json
---------                     -------
  2146912                     2 files
```

### 方式二：脚本自动化

创建 `build-plugin.sh` 脚本：

```bash
#!/bin/bash
# build-plugin.sh - 自动构建插件包

PLUGIN_NAME="my-notify"
VERSION="1.0.0"
OUTPUT_ZIP="${PLUGIN_NAME}-${VERSION}.zip"

echo "Building plugin: ${PLUGIN_NAME} v${VERSION}"

# 1. 编译插件
echo "Compiling plugin..."
go build -buildmode=plugin -o plugin.so main.go
if [ $? -ne 0 ]; then
    echo "Error: Failed to compile plugin"
    exit 1
fi

# 2. 验证 manifest.json 存在
if [ ! -f "manifest.json" ]; then
    echo "Error: manifest.json not found"
    exit 1
fi

# 3. 验证 manifest 格式
echo "Validating manifest..."
if ! jq empty manifest.json 2>/dev/null; then
    echo "Error: Invalid JSON in manifest.json"
    exit 1
fi

# 4. 创建 zip 包
echo "Creating zip package..."
zip -q "${OUTPUT_ZIP}" plugin.so manifest.json

# 5. 计算校验和
echo "Calculating checksum..."
CHECKSUM=$(sha256sum "${OUTPUT_ZIP}" | awk '{print $1}')

echo "✅ Plugin package created successfully!"
echo "   File: ${OUTPUT_ZIP}"
echo "   Size: $(ls -lh ${OUTPUT_ZIP} | awk '{print $5}')"
echo "   SHA256: ${CHECKSUM}"
```

使用方式：

```bash
chmod +x build-plugin.sh
./build-plugin.sh
```

### 方式三：Makefile

在项目中添加 `Makefile`：

```makefile
PLUGIN_NAME := my-notify
VERSION := 1.0.0
OUTPUT := $(PLUGIN_NAME)-$(VERSION).zip

.PHONY: all build package clean

all: package

build:
	@echo "Building plugin..."
	go build -buildmode=plugin -o plugin.so main.go

validate:
	@echo "Validating manifest..."
	@jq empty manifest.json || (echo "Invalid manifest.json" && exit 1)

package: build validate
	@echo "Creating plugin package..."
	@zip -q $(OUTPUT) plugin.so manifest.json
	@echo "✅ Package created: $(OUTPUT)"
	@echo "   Size: $$(ls -lh $(OUTPUT) | awk '{print $$5}')"
	@echo "   SHA256: $$(sha256sum $(OUTPUT) | awk '{print $$1}')"

clean:
	rm -f plugin.so $(OUTPUT)

install: package
	@echo "Installing plugin..."
	curl -X POST http://localhost:8080/api/v1/plugins/install \
	  -F "source=local" \
	  -F "file=@$(OUTPUT)"
```

使用方式：

```bash
# 构建并打包
make package

# 构建、打包并安装
make install
```

## 上传插件包

### 使用 cURL

```bash
curl -X POST http://localhost:8080/api/v1/plugins/install \
  -F "source=local" \
  -F "file=@my-notify-plugin.zip"
```

**参数说明：**
- `source`: 插件来源（`local` 或 `market`）
- `file`: zip 包文件

### 使用 HTTP Client (Go)

```go
package main

import (
    "bytes"
    "io"
    "mime/multipart"
    "net/http"
    "os"
)

func uploadPlugin(zipPath string) error {
    // 打开 zip 文件
    file, err := os.Open(zipPath)
    if err != nil {
        return err
    }
    defer file.Close()

    // 创建 multipart form
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // 添加 source 字段
    writer.WriteField("source", "local")

    // 添加文件
    part, err := writer.CreateFormFile("file", filepath.Base(zipPath))
    if err != nil {
        return err
    }
    io.Copy(part, file)
    writer.Close()

    // 发送请求
    req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/plugins/install", body)
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 处理响应
    // ...
    return nil
}
```

### 使用 Python

```python
import requests

def upload_plugin(zip_path):
    url = 'http://localhost:8080/api/v1/plugins/install'
    
    with open(zip_path, 'rb') as f:
        files = {'file': f}
        data = {'source': 'local'}
        
        response = requests.post(url, files=files, data=data)
        return response.json()

# 使用示例
result = upload_plugin('my-notify-plugin.zip')
print(result)
```

## manifest.json 示例

### 最小化示例

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "My plugin description",
  "author": "Your Name",
  "pluginType": "notify",
  "entryPoint": "plugin.so"
}
```

### 完整示例

```json
{
  "name": "slack-notify",
  "version": "1.0.0",
  "description": "Slack notification plugin",
  "author": "Arcade Team <team@arcade.io>",
  "homepage": "https://example.com/plugins/slack-notify",
  "repository": "https://github.com/arcade/slack-notify",
  "pluginType": "notify",
  "entryPoint": "plugin.so",
  "dependencies": [],
  "configSchema": {
    "type": "object",
    "properties": {
      "webhook_url": {
        "type": "string",
        "description": "Slack Webhook URL"
      },
      "channel": {
        "type": "string",
        "description": "Default channel",
        "default": "#general"
      }
    },
    "required": ["webhook_url"]
  },
  "paramsSchema": {
    "type": "object",
    "properties": {
      "message": {
        "type": "string",
        "description": "Message to send"
      }
    },
    "required": ["message"]
  },
  "defaultConfig": {
    "channel": "#general"
  },
  "icon": "https://cdn.example.com/icons/slack.png",
  "tags": ["notify", "slack"],
  "minVersion": "1.0.0",
  "resources": {
    "cpu": "100m",
    "memory": "128Mi",
    "disk": "10Mi"
  }
}
```

## 验证插件包

### 上传前验证

```bash
# 方式1：验证 zip 包内容
unzip -l my-plugin.zip

# 方式2：使用 API 验证
curl -X POST http://localhost:8080/api/v1/plugins/validate-manifest \
  -F "file=@my-plugin.zip"

# 方式3：验证 JSON 格式
unzip -p my-plugin.zip manifest.json | jq .
```

### 验证清单内容

```bash
# 提取 manifest.json 并验证
unzip -p my-plugin.zip manifest.json | jq . > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ Manifest is valid JSON"
else
    echo "❌ Manifest is invalid JSON"
fi
```

## 常见错误

### 1. 缺少 manifest.json

```
Error: manifest.json not found in zip package
```

**解决：** 确保 zip 包根目录包含 `manifest.json` 文件。

### 2. 缺少 .so 文件

```
Error: plugin .so file not found in zip package
```

**解决：** 确保 zip 包包含编译好的 `.so` 文件。

### 3. manifest.json 格式错误

```
Error: failed to parse manifest.json: invalid character...
```

**解决：** 使用 `jq` 或其他工具验证 JSON 格式。

### 4. 文件类型错误

```
Error: invalid file type, expected .zip file
```

**解决：** 确保上传的是 `.zip` 格式的压缩包。

## 最佳实践

### 1. 版本管理

```bash
# 文件名包含版本号
slack-notify-1.0.0.zip
slack-notify-1.1.0.zip
slack-notify-2.0.0.zip
```

### 2. 目录结构

```
my-plugin/
├── main.go              # 插件源代码
├── manifest.json        # 清单文件
├── README.md           # 说明文档
├── LICENSE             # 许可证
├── examples/           # 使用示例
│   └── config.json
└── Makefile           # 构建脚本
```

### 3. 自动化构建

使用 CI/CD 自动构建和打包：

```yaml
# .github/workflows/build.yml
name: Build Plugin

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21
      
      - name: Build plugin
        run: |
          go build -buildmode=plugin -o plugin.so main.go
      
      - name: Create package
        run: |
          zip my-plugin-${{ github.ref_name }}.zip plugin.so manifest.json
      
      - name: Upload artifact
        uses: actions/upload-artifact@v2
        with:
          name: plugin-package
          path: my-plugin-*.zip
```

### 4. 版本号规范

在 `manifest.json` 中使用语义化版本：

```json
{
  "version": "1.2.3"
  // 主版本.次版本.修订号
}
```

### 5. 清单完整性

确保 `manifest.json` 包含所有必需字段：

```bash
#!/bin/bash
# validate-manifest.sh

REQUIRED_FIELDS="name version description author pluginType entryPoint"

for field in $REQUIRED_FIELDS; do
    value=$(jq -r ".$field" manifest.json)
    if [ "$value" = "null" ] || [ -z "$value" ]; then
        echo "❌ Missing required field: $field"
        exit 1
    fi
done

echo "✅ All required fields present"
```

## 示例：完整的构建流程

### 项目结构

```
slack-notify-plugin/
├── main.go
├── manifest.json
├── README.md
├── Makefile
└── build-plugin.sh
```

### main.go

```go
package main

import (
    "context"
    "github.com/observabil/arcade/pkg/plugin"
)

type SlackNotifyPlugin struct{}

func (p *SlackNotifyPlugin) Name() string        { return "slack-notify" }
func (p *SlackNotifyPlugin) Description() string { return "Slack notification plugin" }
func (p *SlackNotifyPlugin) Version() string     { return "1.0.0" }
func (p *SlackNotifyPlugin) Type() plugin.PluginType { return plugin.TypeNotify }

func (p *SlackNotifyPlugin) Init(ctx context.Context, config any) error {
    return nil
}

func (p *SlackNotifyPlugin) Cleanup() error {
    return nil
}

func (p *SlackNotifyPlugin) Send(ctx context.Context, message any, opts ...plugin.Option) error {
    // 实现发送逻辑
    return nil
}

func (p *SlackNotifyPlugin) SendTemplate(ctx context.Context, template string, data any, opts ...plugin.Option) error {
    return nil
}

var Plugin plugin.NotifyPlugin = &SlackNotifyPlugin{}
```

### manifest.json

```json
{
  "name": "slack-notify",
  "version": "1.0.0",
  "description": "Send notifications to Slack",
  "author": "Arcade Team",
  "pluginType": "notify",
  "entryPoint": "plugin.so",
  "configSchema": {
    "type": "object",
    "properties": {
      "webhook_url": {
        "type": "string"
      }
    },
    "required": ["webhook_url"]
  }
}
```

### Makefile

```makefile
VERSION := 1.0.0
OUTPUT := slack-notify-$(VERSION).zip

.PHONY: all build package install clean

all: package

build:
	go build -buildmode=plugin -o plugin.so main.go

package: build
	zip $(OUTPUT) plugin.so manifest.json
	@echo "Package created: $(OUTPUT)"

install: package
	curl -X POST http://localhost:8080/api/v1/plugins/install \
	  -F "source=local" \
	  -F "file=@$(OUTPUT)"

clean:
	rm -f plugin.so $(OUTPUT)
```

### 使用流程

```bash
# 1. 构建并打包
make package

# 2. 安装到 Arcade
make install

# 或者分步执行
make build           # 编译
make package         # 打包
curl ... # 手动上传
```

## 完整示例下载

我们提供了一些官方插件示例供参考：

- [Slack Notify Plugin](https://github.com/arcade/plugin-slack-notify)
- [Email Notify Plugin](https://github.com/arcade/plugin-email-notify)
- [K8s Deploy Plugin](https://github.com/arcade/plugin-k8s-deploy)

## API 使用

### 安装插件

```bash
curl -X POST http://localhost:8080/api/v1/plugins/install \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "source=local" \
  -F "file=@my-plugin-1.0.0.zip"
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "success": true,
    "message": "plugin installed successfully",
    "pluginId": "plugin_my-plugin_abc123",
    "version": "1.0.0"
  }
}
```

### 验证插件包

```bash
curl -X POST http://localhost:8080/api/v1/plugins/validate-manifest \
  -F "file=@my-plugin-1.0.0.zip"
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "manifest is valid",
    "valid": true,
    "manifest": {
      "name": "my-plugin",
      "version": "1.0.0",
      ...
    }
  }
}
```

## 故障排查

### 检查 zip 包内容

```bash
# 列出文件
unzip -l my-plugin.zip

# 提取并查看 manifest
unzip -p my-plugin.zip manifest.json | jq .

# 验证 .so 文件存在
unzip -l my-plugin.zip | grep '\.so$'
```

### 检查文件大小

```bash
# zip 包大小
ls -lh my-plugin.zip

# 解压后文件大小
unzip -l my-plugin.zip
```

### 检查权限

```bash
# 确保文件可读
chmod 644 my-plugin.zip

# 验证文件完整性
sha256sum my-plugin.zip
```

## 安全建议

1. **验证来源**
   - 只安装可信来源的插件
   - 检查作者信息

2. **检查内容**
   - 解压查看 zip 包内容
   - 验证 manifest 合法性

3. **校验和**
   - 记录并验证插件校验和
   - 防止文件被篡改

4. **测试环境**
   - 先在测试环境安装
   - 验证功能正常后再在生产环境部署

## 相关文档

- [插件清单示例](./PLUGIN_MANIFEST_EXAMPLES.md)
- [插件开发指南](./PLUGIN_DEVELOPMENT.md)
- [快速入门](./PLUGIN_QUICKSTART_CN.md)

---

**文档版本：** v1.0.0  
**最后更新：** 2025-01-16  
**维护者：** Arcade Team

