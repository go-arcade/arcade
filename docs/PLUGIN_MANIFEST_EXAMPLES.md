# 插件清单（Manifest）样例

本文档提供了各种类型插件的清单样例，帮助开发者快速创建符合规范的插件清单。

## 清单结构说明

```json
{
  "name": "插件名称（必填）",
  "version": "版本号（必填，遵循语义化版本）",
  "description": "插件描述（必填）",
  "author": "作者信息（必填）",
  "homepage": "主页地址（可选）",
  "repository": "代码仓库地址（可选）",
  "pluginType": "插件类型（必填：notify/ci/cd/security/storage/approval/custom）",
  "entryPoint": "插件入口文件（必填，如：plugin.so）",
  "dependencies": ["依赖的其他插件（可选）"],
  "configSchema": {
    "插件配置的 JSON Schema（可选）"
  },
  "paramsSchema": {
    "插件参数的 JSON Schema（可选）"
  },
  "defaultConfig": {
    "默认配置值（可选）"
  },
  "icon": "图标URL（可选）",
  "tags": ["标签数组（可选）"],
  "minVersion": "最低平台版本要求（可选）",
  "resources": {
    "资源需求（可选）"
  }
}
```

## 1. 通知类插件（Notify）

### 1.1 Slack 通知插件

```json
{
  "name": "slack-notify",
  "version": "1.0.0",
  "description": "Slack notification plugin for sending messages to Slack channels",
  "author": "Arcade Team <team@arcade.io>",
  "homepage": "https://example.com/plugins/slack-notify",
  "repository": "https://github.com/arcade/slack-notify-plugin",
  "pluginType": "notify",
  "entryPoint": "slack-notify.so",
  "dependencies": [],
  "configSchema": {
    "type": "object",
    "properties": {
      "webhook_url": {
        "type": "string",
        "description": "Slack Webhook URL",
        "format": "uri"
      },
      "channel": {
        "type": "string",
        "description": "Default Slack channel",
        "default": "#general"
      },
      "username": {
        "type": "string",
        "description": "Bot username",
        "default": "Arcade Bot"
      },
      "icon_emoji": {
        "type": "string",
        "description": "Bot icon emoji",
        "default": ":robot_face:"
      },
      "timeout": {
        "type": "integer",
        "description": "Request timeout in seconds",
        "default": 30,
        "minimum": 1,
        "maximum": 300
      }
    },
    "required": ["webhook_url"]
  },
  "paramsSchema": {
    "type": "object",
    "properties": {
      "message": {
        "type": "string",
        "description": "Message content to send"
      },
      "channel": {
        "type": "string",
        "description": "Override default channel"
      },
      "mentions": {
        "type": "array",
        "items": {
          "type": "string"
        },
        "description": "Users to mention"
      }
    },
    "required": ["message"]
  },
  "defaultConfig": {
    "channel": "#general",
    "username": "Arcade Bot",
    "icon_emoji": ":robot_face:",
    "timeout": 30
  },
  "icon": "https://cdn.example.com/icons/slack.png",
  "tags": ["notify", "slack", "messaging", "communication"],
  "minVersion": "1.0.0",
  "resources": {
    "cpu": "100m",
    "memory": "128Mi",
    "disk": "10Mi"
  }
}
```

### 1.2 企业微信通知插件

```json
{
  "name": "wecom-notify",
  "version": "1.0.0",
  "description": "企业微信群机器人通知插件",
  "author": "Arcade Team",
  "homepage": "https://example.com/plugins/wecom-notify",
  "repository": "https://github.com/arcade/wecom-notify-plugin",
  "pluginType": "notify",
  "entryPoint": "wecom-notify.so",
  "dependencies": [],
  "configSchema": {
    "type": "object",
    "properties": {
      "webhook_url": {
        "type": "string",
        "description": "企业微信机器人 Webhook URL"
      },
      "mentioned_list": {
        "type": "array",
        "items": {
          "type": "string"
        },
        "description": "默认提醒的成员列表（userid）"
      },
      "mentioned_mobile_list": {
        "type": "array",
        "items": {
          "type": "string"
        },
        "description": "默认提醒的成员列表（手机号）"
      }
    },
    "required": ["webhook_url"]
  },
  "paramsSchema": {
    "type": "object",
    "properties": {
      "msgtype": {
        "type": "string",
        "enum": ["text", "markdown"],
        "description": "消息类型"
      },
      "content": {
        "type": "string",
        "description": "消息内容"
      }
    },
    "required": ["content"]
  },
  "defaultConfig": {
    "mentioned_list": [],
    "mentioned_mobile_list": []
  },
  "icon": "https://cdn.example.com/icons/wecom.png",
  "tags": ["notify", "wecom", "wechat", "企业微信"],
  "minVersion": "1.0.0",
  "resources": {
    "cpu": "50m",
    "memory": "64Mi",
    "disk": "5Mi"
  }
}
```

### 1.3 邮件通知插件

```json
{
  "name": "email-notify",
  "version": "1.2.0",
  "description": "Email notification plugin with SMTP support",
  "author": "Arcade Team",
  "pluginType": "notify",
  "entryPoint": "email-notify.so",
  "configSchema": {
    "type": "object",
    "properties": {
      "smtp_host": {
        "type": "string",
        "description": "SMTP server host"
      },
      "smtp_port": {
        "type": "integer",
        "description": "SMTP server port",
        "default": 587
      },
      "smtp_user": {
        "type": "string",
        "description": "SMTP username"
      },
      "smtp_password": {
        "type": "string",
        "description": "SMTP password"
      },
      "from_address": {
        "type": "string",
        "format": "email",
        "description": "Sender email address"
      },
      "from_name": {
        "type": "string",
        "description": "Sender name",
        "default": "Arcade Notification"
      },
      "use_tls": {
        "type": "boolean",
        "description": "Use TLS encryption",
        "default": true
      }
    },
    "required": ["smtp_host", "smtp_user", "smtp_password", "from_address"]
  },
  "paramsSchema": {
    "type": "object",
    "properties": {
      "to": {
        "type": "array",
        "items": {
          "type": "string",
          "format": "email"
        },
        "description": "Recipient email addresses"
      },
      "subject": {
        "type": "string",
        "description": "Email subject"
      },
      "body": {
        "type": "string",
        "description": "Email body"
      },
      "html": {
        "type": "boolean",
        "description": "Use HTML format",
        "default": false
      }
    },
    "required": ["to", "subject", "body"]
  },
  "defaultConfig": {
    "smtp_port": 587,
    "from_name": "Arcade Notification",
    "use_tls": true
  },
  "icon": "https://cdn.example.com/icons/email.png",
  "tags": ["notify", "email", "smtp"],
  "minVersion": "1.0.0",
  "resources": {
    "cpu": "100m",
    "memory": "128Mi",
    "disk": "20Mi"
  }
}
```

## 2. CI 类插件

### 2.1 代码检查插件

```json
{
  "name": "code-linter",
  "version": "2.0.0",
  "description": "Multi-language code linter plugin",
  "author": "DevOps Team",
  "repository": "https://github.com/arcade/code-linter-plugin",
  "pluginType": "ci",
  "entryPoint": "code-linter.so",
  "dependencies": [],
  "configSchema": {
    "type": "object",
    "properties": {
      "languages": {
        "type": "array",
        "items": {
          "type": "string",
          "enum": ["go", "python", "javascript", "typescript", "java"]
        },
        "description": "Enabled languages"
      },
      "severity_threshold": {
        "type": "string",
        "enum": ["error", "warning", "info"],
        "description": "Minimum severity to report",
        "default": "warning"
      },
      "fail_on_error": {
        "type": "boolean",
        "description": "Fail build if errors found",
        "default": true
      }
    },
    "required": ["languages"]
  },
  "paramsSchema": {
    "type": "object",
    "properties": {
      "files": {
        "type": "array",
        "items": {
          "type": "string"
        },
        "description": "Files or directories to lint"
      },
      "exclude": {
        "type": "array",
        "items": {
          "type": "string"
        },
        "description": "Patterns to exclude"
      }
    }
  },
  "defaultConfig": {
    "languages": ["go", "javascript"],
    "severity_threshold": "warning",
    "fail_on_error": true
  },
  "icon": "https://cdn.example.com/icons/linter.png",
  "tags": ["ci", "lint", "code-quality"],
  "minVersion": "1.0.0",
  "resources": {
    "cpu": "500m",
    "memory": "512Mi",
    "disk": "100Mi"
  }
}
```

## 3. CD 类插件

### 3.1 Kubernetes 部署插件

```json
{
  "name": "k8s-deploy",
  "version": "1.5.0",
  "description": "Kubernetes deployment plugin",
  "author": "Platform Team",
  "repository": "https://github.com/arcade/k8s-deploy-plugin",
  "pluginType": "cd",
  "entryPoint": "k8s-deploy.so",
  "dependencies": [],
  "configSchema": {
    "type": "object",
    "properties": {
      "kubeconfig": {
        "type": "string",
        "description": "Kubeconfig file content or path"
      },
      "context": {
        "type": "string",
        "description": "Kubernetes context to use"
      },
      "namespace": {
        "type": "string",
        "description": "Default namespace",
        "default": "default"
      }
    },
    "required": ["kubeconfig"]
  },
  "paramsSchema": {
    "type": "object",
    "properties": {
      "manifest": {
        "type": "string",
        "description": "Kubernetes manifest file path or content"
      },
      "namespace": {
        "type": "string",
        "description": "Override namespace"
      },
      "wait": {
        "type": "boolean",
        "description": "Wait for deployment to complete",
        "default": true
      },
      "timeout": {
        "type": "integer",
        "description": "Deployment timeout in seconds",
        "default": 300
      }
    },
    "required": ["manifest"]
  },
  "defaultConfig": {
    "namespace": "default"
  },
  "icon": "https://cdn.example.com/icons/kubernetes.png",
  "tags": ["cd", "kubernetes", "k8s", "deployment"],
  "minVersion": "1.0.0",
  "resources": {
    "cpu": "200m",
    "memory": "256Mi",
    "disk": "50Mi"
  }
}
```

## 4. Security 类插件

### 4.1 漏洞扫描插件

```json
{
  "name": "security-scanner",
  "version": "1.3.0",
  "description": "Container image and code vulnerability scanner",
  "author": "Security Team",
  "pluginType": "security",
  "entryPoint": "security-scanner.so",
  "configSchema": {
    "type": "object",
    "properties": {
      "scan_type": {
        "type": "array",
        "items": {
          "type": "string",
          "enum": ["image", "code", "dependencies"]
        },
        "description": "Types of scans to perform"
      },
      "severity_threshold": {
        "type": "string",
        "enum": ["critical", "high", "medium", "low"],
        "description": "Minimum severity to report",
        "default": "medium"
      },
      "fail_on_critical": {
        "type": "boolean",
        "description": "Fail if critical vulnerabilities found",
        "default": true
      }
    },
    "required": ["scan_type"]
  },
  "paramsSchema": {
    "type": "object",
    "properties": {
      "target": {
        "type": "string",
        "description": "Target to scan (image name, directory, etc.)"
      }
    },
    "required": ["target"]
  },
  "defaultConfig": {
    "scan_type": ["image", "dependencies"],
    "severity_threshold": "medium",
    "fail_on_critical": true
  },
  "icon": "https://cdn.example.com/icons/security.png",
  "tags": ["security", "vulnerability", "scanner"],
  "minVersion": "1.0.0",
  "resources": {
    "cpu": "1000m",
    "memory": "1Gi",
    "disk": "500Mi"
  }
}
```

## 5. Storage 类插件

### 5.1 S3 存储插件

```json
{
  "name": "s3-storage",
  "version": "1.0.0",
  "description": "Amazon S3 compatible storage plugin",
  "author": "Storage Team",
  "pluginType": "storage",
  "entryPoint": "s3-storage.so",
  "configSchema": {
    "type": "object",
    "properties": {
      "endpoint": {
        "type": "string",
        "description": "S3 endpoint URL"
      },
      "region": {
        "type": "string",
        "description": "AWS region",
        "default": "us-east-1"
      },
      "bucket": {
        "type": "string",
        "description": "Default bucket name"
      },
      "access_key": {
        "type": "string",
        "description": "AWS access key ID"
      },
      "secret_key": {
        "type": "string",
        "description": "AWS secret access key"
      },
      "use_ssl": {
        "type": "boolean",
        "description": "Use SSL/TLS",
        "default": true
      }
    },
    "required": ["bucket", "access_key", "secret_key"]
  },
  "paramsSchema": {
    "type": "object",
    "properties": {
      "key": {
        "type": "string",
        "description": "Object key/path"
      },
      "bucket": {
        "type": "string",
        "description": "Override bucket"
      }
    },
    "required": ["key"]
  },
  "defaultConfig": {
    "region": "us-east-1",
    "use_ssl": true
  },
  "icon": "https://cdn.example.com/icons/s3.png",
  "tags": ["storage", "s3", "aws"],
  "minVersion": "1.0.0",
  "resources": {
    "cpu": "200m",
    "memory": "256Mi",
    "disk": "100Mi"
  }
}
```

## 6. Custom 类插件

### 6.1 自定义脚本执行插件

```json
{
  "name": "custom-script-runner",
  "version": "1.0.0",
  "description": "Execute custom scripts with environment variables",
  "author": "Your Name",
  "pluginType": "custom",
  "entryPoint": "script-runner.so",
  "configSchema": {
    "type": "object",
    "properties": {
      "shell": {
        "type": "string",
        "enum": ["bash", "sh", "python", "node"],
        "description": "Script interpreter",
        "default": "bash"
      },
      "timeout": {
        "type": "integer",
        "description": "Execution timeout in seconds",
        "default": 600
      },
      "working_directory": {
        "type": "string",
        "description": "Working directory for script execution"
      }
    }
  },
  "paramsSchema": {
    "type": "object",
    "properties": {
      "script": {
        "type": "string",
        "description": "Script content to execute"
      },
      "args": {
        "type": "array",
        "items": {
          "type": "string"
        },
        "description": "Script arguments"
      },
      "env": {
        "type": "object",
        "description": "Environment variables"
      }
    },
    "required": ["script"]
  },
  "defaultConfig": {
    "shell": "bash",
    "timeout": 600
  },
  "icon": "https://cdn.example.com/icons/script.png",
  "tags": ["custom", "script", "automation"],
  "minVersion": "1.0.0",
  "resources": {
    "cpu": "500m",
    "memory": "512Mi",
    "disk": "200Mi"
  }
}
```

## 最小化示例

如果你只需要最基本的字段，这是一个最小化的示例：

```json
{
  "name": "simple-plugin",
  "version": "1.0.0",
  "description": "A simple plugin example",
  "author": "Your Name",
  "pluginType": "custom",
  "entryPoint": "plugin.so"
}
```

## 字段说明

### 必填字段

- `name`: 插件名称，建议使用小写字母和连字符
- `version`: 版本号，遵循语义化版本规范（如 1.0.0）
- `description`: 插件描述
- `author`: 作者信息
- `pluginType`: 插件类型（notify/ci/cd/security/storage/approval/custom）
- `entryPoint`: 插件入口文件名（通常是 .so 文件）

### 可选字段

- `homepage`: 插件主页 URL
- `repository`: 代码仓库 URL
- `dependencies`: 依赖的其他插件列表
- `configSchema`: 插件配置的 JSON Schema 定义
- `paramsSchema`: 插件参数的 JSON Schema 定义
- `defaultConfig`: 默认配置值
- `icon`: 插件图标 URL
- `tags`: 标签数组，用于分类和搜索
- `minVersion`: 最低平台版本要求
- `resources`: 资源需求
  - `cpu`: CPU 需求（如 "100m", "1000m"）
  - `memory`: 内存需求（如 "128Mi", "1Gi"）
  - `disk`: 磁盘需求（如 "50Mi", "1Gi"）

## 使用方法

### 1. 创建清单文件

将上述任一示例保存为 `manifest.json` 文件。

### 2. 上传插件

```bash
curl -X POST http://localhost:8080/api/v1/plugins/install \
  -F "source=local" \
  -F "file=@your-plugin.so" \
  -F "manifest=$(cat manifest.json)"
```

### 3. 验证清单

在上传前验证清单格式：

```bash
curl -X POST http://localhost:8080/api/v1/plugins/validate-manifest \
  -H "Content-Type: application/json" \
  -d @manifest.json
```

## 最佳实践

1. **版本管理**: 严格遵循语义化版本规范
2. **文档完善**: 在 `description` 中提供清晰的说明
3. **Schema 定义**: 为配置和参数提供完整的 JSON Schema
4. **默认值**: 为所有可选配置提供合理的默认值
5. **资源限制**: 准确评估并设置资源需求
6. **标签使用**: 添加相关标签便于分类和搜索
7. **依赖声明**: 明确列出所有依赖项

## 相关文档

- [插件开发指南](./PLUGIN_DEVELOPMENT.md)
- [插件系统架构](./PLUGIN_SYSTEM_REFACTOR.md)
- [快速入门](./PLUGIN_QUICKSTART_CN.md)

---

**文档版本：** v1.0.0  
**最后更新：** 2025-01-16  
**维护者：** Arcade Team

