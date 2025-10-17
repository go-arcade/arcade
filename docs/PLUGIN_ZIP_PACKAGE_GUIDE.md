# 插件 Zip 包上传指南

## 核心变更

从 v1.0.0 开始，插件上传方式改为 **Zip 包格式**，系统会自动解压并解析 manifest.json。

## 包格式要求

### 必需文件

```
my-plugin.zip
├── plugin.so          ✅ 必需：插件二进制文件
└── manifest.json      ✅ 必需：插件清单文件
```

### 系统行为

1. **接收 zip 包** - 用户上传 .zip 文件
2. **自动解压** - 系统解压 zip 包
3. **查找文件** - 查找 `.so` 文件和 `manifest.json`
4. **解析清单** - 自动解析 manifest.json
5. **验证格式** - 验证清单格式和必填字段
6. **保存到数据库** - 将 manifest 保存到 Plugin 表的 manifest 字段
7. **建立关联** - 插件配置和插件 ID 自动关联

## 快速开始

### 1. 创建插件包

```bash
# 步骤1：编译插件
go build -buildmode=plugin -o plugin.so main.go

# 步骤2：创建 manifest.json
cat > manifest.json <<'EOF'
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "My awesome plugin",
  "author": "Your Name",
  "pluginType": "notify",
  "entryPoint": "plugin.so",
  "configSchema": {
    "type": "object",
    "properties": {
      "api_key": {"type": "string"}
    }
  }
}
EOF

# 步骤3：打包
zip my-plugin.zip plugin.so manifest.json

# 步骤4：验证
unzip -l my-plugin.zip
```

### 2. 上传安装

```bash
curl -X POST http://localhost:8080/api/v1/plugins/install \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "source=local" \
  -F "file=@my-plugin.zip"
```

### 3. 查看结果

```bash
# 列出已安装的插件
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/plugins
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "pluginId": "plugin_my-plugin_abc123",
      "name": "my-plugin",
      "version": "1.0.0",
      "pluginType": "notify",
      "manifest": {
        "name": "my-plugin",
        "version": "1.0.0",
        ...
      }
    }
  ]
}
```

## Manifest 与数据库的关系

### 数据存储

```sql
CREATE TABLE t_plugin (
    plugin_id VARCHAR(100),     -- 自动生成
    name VARCHAR(100),          -- 从 manifest.name 提取
    version VARCHAR(20),        -- 从 manifest.version 提取
    plugin_type VARCHAR(50),    -- 从 manifest.pluginType 提取
    config_schema JSON,         -- 从 manifest.configSchema 提取
    params_schema JSON,         -- 从 manifest.paramsSchema 提取
    default_config JSON,        -- 从 manifest.defaultConfig 提取
    manifest JSON,              -- 保存完整的 manifest
    ...
);
```

### 字段映射

| Manifest 字段 | 数据库字段 | 说明 |
|--------------|-----------|------|
| name | name | 插件名称 |
| version | version | 版本号 |
| description | description | 描述 |
| author | author | 作者 |
| pluginType | plugin_type | 插件类型 |
| entryPoint | entry_point | 入口文件 |
| configSchema | config_schema | 配置 Schema |
| paramsSchema | params_schema | 参数 Schema |
| defaultConfig | default_config | 默认配置 |
| icon | icon | 图标 URL |
| repository | repository | 仓库地址 |
| *整个 manifest* | manifest | 完整清单（JSON） |

### 配置关联

插件的配置通过 `plugin_id` 关联：

```sql
-- t_plugin 表
INSERT INTO t_plugin (plugin_id, name, manifest, ...) VALUES (...);

-- t_plugin_config 表（插件配置）
INSERT INTO t_plugin_config (plugin_id, config_items, ...) VALUES (...);

-- 关联关系
SELECT p.*, pc.config_items
FROM t_plugin p
LEFT JOIN t_plugin_config pc ON p.plugin_id = pc.plugin_id
WHERE p.plugin_id = 'plugin_my-plugin_abc123';
```

## 完整示例

### 插件源码 (main.go)

```go
package main

import (
    "context"
    "fmt"
    "github.com/observabil/arcade/pkg/plugin"
)

type MyPlugin struct {
    config map[string]interface{}
}

func (p *MyPlugin) Name() string        { return "my-plugin" }
func (p *MyPlugin) Description() string { return "My awesome plugin" }
func (p *MyPlugin) Version() string     { return "1.0.0" }
func (p *MyPlugin) Type() plugin.PluginType { return plugin.TypeNotify }

func (p *MyPlugin) Init(ctx context.Context, config any) error {
    if cfg, ok := config.(map[string]interface{}); ok {
        p.config = cfg
    }
    return nil
}

func (p *MyPlugin) Cleanup() error {
    return nil
}

func (p *MyPlugin) Send(ctx context.Context, message any, opts ...plugin.Option) error {
    fmt.Printf("Sending message: %v with config: %v\n", message, p.config)
    return nil
}

func (p *MyPlugin) SendTemplate(ctx context.Context, template string, data any, opts ...plugin.Option) error {
    return nil
}

var Plugin plugin.NotifyPlugin = &MyPlugin{}
```

### 清单文件 (manifest.json)

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "My awesome notification plugin",
  "author": "Your Name <your.email@example.com>",
  "homepage": "https://github.com/yourusername/my-plugin",
  "repository": "https://github.com/yourusername/my-plugin",
  "pluginType": "notify",
  "entryPoint": "plugin.so",
  "dependencies": [],
  "configSchema": {
    "type": "object",
    "properties": {
      "api_key": {
        "type": "string",
        "description": "API Key for authentication"
      },
      "timeout": {
        "type": "integer",
        "description": "Request timeout in seconds",
        "default": 30
      }
    },
    "required": ["api_key"]
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
    "timeout": 30
  },
  "icon": "https://example.com/icon.png",
  "tags": ["notify", "custom"],
  "minVersion": "1.0.0",
  "resources": {
    "cpu": "100m",
    "memory": "128Mi",
    "disk": "10Mi"
  }
}
```

### 构建和打包

```bash
# 编译
go build -buildmode=plugin -o plugin.so main.go

# 打包
zip my-plugin-1.0.0.zip plugin.so manifest.json

# 上传
curl -X POST http://localhost:8080/api/v1/plugins/install \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "source=local" \
  -F "file=@my-plugin-1.0.0.zip"
```

## 使用 Makefile 自动化

完整的 `Makefile` 示例：

```makefile
PLUGIN_NAME := my-plugin
VERSION := 1.0.0
OUTPUT_ZIP := $(PLUGIN_NAME)-$(VERSION).zip
SERVER_URL := http://localhost:8080

.PHONY: all build package install test clean

all: package

# 编译插件
build:
	@echo "Building plugin..."
	@go build -buildmode=plugin -o plugin.so main.go
	@echo "✅ Build completed"

# 验证 manifest
validate:
	@echo "Validating manifest..."
	@jq empty manifest.json 2>/dev/null || (echo "❌ Invalid manifest.json" && exit 1)
	@echo "✅ Manifest is valid"

# 打包
package: build validate
	@echo "Creating package..."
	@zip -q $(OUTPUT_ZIP) plugin.so manifest.json
	@echo "✅ Package created: $(OUTPUT_ZIP)"
	@echo "   Size: $$(ls -lh $(OUTPUT_ZIP) | awk '{print $$5}')"
	@echo "   SHA256: $$(sha256sum $(OUTPUT_ZIP) | awk '{print $$1}')"

# 安装到服务器
install: package
	@echo "Installing plugin..."
	@curl -X POST $(SERVER_URL)/api/v1/plugins/install \
	  -H "Authorization: Bearer $(TOKEN)" \
	  -F "source=local" \
	  -F "file=@$(OUTPUT_ZIP)"

# 测试插件功能
test:
	@echo "Testing plugin..."
	@go test -v ./...

# 清理
clean:
	@rm -f plugin.so $(OUTPUT_ZIP)
	@echo "✅ Cleaned"

# 显示帮助
help:
	@echo "Available targets:"
	@echo "  make build    - Compile plugin"
	@echo "  make package  - Create zip package"
	@echo "  make install  - Install to server (requires TOKEN env var)"
	@echo "  make test     - Run tests"
	@echo "  make clean    - Remove build artifacts"
```

使用方式：

```bash
# 设置访问令牌
export TOKEN="your_access_token"

# 构建并安装
make install
```

## 验证插件包

### 在上传前验证

```bash
# 1. 检查 zip 包结构
unzip -l my-plugin.zip

# 2. 验证 manifest.json 格式
unzip -p my-plugin.zip manifest.json | jq .

# 3. 检查 .so 文件存在
unzip -l my-plugin.zip | grep '\.so$'

# 4. 使用 API 验证
curl -X POST http://localhost:8080/api/v1/plugins/validate-manifest \
  -F "file=@my-plugin.zip"
```

## 常见问题

### Q1: zip 包可以包含其他文件吗？

A: 可以，系统只会提取 `.so` 文件和 `manifest.json`，其他文件会被忽略。

### Q2: manifest.json 必须在根目录吗？

A: 是的，系统会在 zip 包中递归查找，但建议放在根目录以提高处理速度。

### Q3: 可以包含多个 .so 文件吗？

A: 不建议。系统会使用找到的第一个 .so 文件，多个文件可能导致混淆。

### Q4: entryPoint 字段如何填写？

A: 填写 .so 文件的文件名即可，如 `plugin.so`。如果留空，系统会自动使用找到的 .so 文件名。

### Q5: 上传失败了怎么办？

A: 检查以下几点：
1. zip 包是否包含必需文件
2. manifest.json 格式是否正确
3. .so 文件是否有效
4. 服务器日志中的详细错误信息

## 最佳实践

1. **使用版本号命名**
   ```
   my-plugin-1.0.0.zip
   my-plugin-1.1.0.zip
   ```

2. **自动化构建**
   - 使用 Makefile 或脚本
   - 集成到 CI/CD 流程

3. **验证后上传**
   - 先本地验证 zip 包内容
   - 使用验证 API 检查 manifest

4. **保留构建记录**
   - 记录构建时间和版本
   - 保存校验和

5. **测试后发布**
   - 先在测试环境安装
   - 验证功能正常后再发布

## 相关文档

- [插件包结构规范](./PLUGIN_PACKAGE_STRUCTURE.md)
- [插件清单示例](./PLUGIN_MANIFEST_EXAMPLES.md)
- [快速入门指南](./PLUGIN_QUICKSTART_CN.md)

---

**文档版本：** v1.0.0  
**最后更新：** 2025-01-16  
**维护者：** Arcade Team

