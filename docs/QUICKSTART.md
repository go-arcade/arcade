# Arcade CI/CD 平台 快速开始

本文档帮助您快速搭建和运行Arcade CI/CD平台。

## 前置要求

### 1. 基础环境

- Go 1.19+
- Protocol Buffers 编译器 (protoc)
- Make
- Git

### 2. 安装Protocol Buffers

根据您的操作系统选择安装方式：

#### macOS

```bash
brew install protobuf
```

#### Ubuntu/Debian

```bash
sudo apt-get update
sudo apt-get install -y protobuf-compiler
```

#### CentOS/RHEL

```bash
sudo yum install -y protobuf-compiler
```

#### 验证安装

```bash
protoc --version
# 应该显示: libprotoc 3.x.x 或更高版本
```

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/observabil/arcade.git
cd arcade
```

### 2. 安装protoc插件

首次使用需要安装Go的protoc插件：

```bash
make proto-install
```

这将安装：
- `protoc-gen-go`: Protocol Buffers Go插件
- `protoc-gen-go-grpc`: gRPC Go插件

### 3. 生成proto代码

```bash
make proto
```

这将生成以下文件：
- `api/agent/v1/agent.pb.go` - Agent服务消息定义
- `api/agent/v1/agent_grpc.pb.go` - Agent服务gRPC接口
- `api/job/v1/job.pb.go` - Job服务消息定义
- `api/job/v1/job_grpc.pb.go` - Job服务gRPC接口
- `api/stream/v1/stream.pb.go` - Stream服务消息定义
- `api/stream/v1/stream_grpc.pb.go` - Stream服务gRPC接口

### 4. 构建插件

```bash
make plugins
```

### 5. 构建主程序

```bash
make build
```

或者一次性完成所有构建（推荐）：

```bash
make all
```

### 6. 运行程序

```bash
# 前台运行
./arcade

# 或后台运行
make run
```

## Makefile命令参考

运行 `make` 或 `make help` 查看所有可用命令：

```bash
make help
```

### 常用命令

| 命令 | 说明 |
|------|------|
| `make help` | 显示帮助信息 |
| `make proto-install` | 安装protoc插件（首次使用） |
| `make proto` | 生成proto代码 |
| `make proto-clean` | 清理生成的proto代码 |
| `make plugins` | 构建所有插件 |
| `make plugins-clean` | 清理插件构建产物 |
| `make build` | 构建主程序 |
| `make build-cli` | 构建CLI工具 |
| `make all` | 完整构建（前端+插件+主程序） |
| `make run` | 后台运行主程序 |
| `make release` | 创建发布版本 |

## 开发工作流

### 修改proto文件后

```bash
# 1. 修改 api/**/*.proto 文件
vim api/agent/v1/agent.proto

# 2. 重新生成代码
make proto

# 3. 重新构建
make build
```

### 开发新插件

```bash
# 1. 在 pkg/plugins/ 下创建新插件目录
mkdir -p pkg/plugins/myplugin

# 2. 编写插件代码
vim pkg/plugins/myplugin/main.go

# 3. 构建插件
make plugins

# 4. 插件将生成到 plugins/myplugin.so
```

### 清理重建

```bash
# 清理proto生成的代码
make proto-clean

# 清理插件
make plugins-clean

# 重新生成和构建
make proto
make all
```

## 配置

### 1. 服务器配置

编辑 `conf.d/config.toml`：

```toml
[server]
http_port = 8080
grpc_port = 9090

[database]
type = "mysql"
host = "localhost"
port = 3306
```

### 2. 插件配置

编辑 `conf.d/plugins.yaml`：

```yaml
plugins:
  - name: stdout
    type: notify
    enabled: true
```

### 3. 权限配置

编辑 `conf.d/model.conf` 配置Casbin权限模型。

## 项目结构

```
arcade/
├── api/                    # Proto定义
│   ├── agent/v1/          # Agent服务
│   ├── job/v1/            # Job和Stream服务
│   ├── README.md          # API文档
│   └── LABEL_EXAMPLES.md  # Label使用示例
├── cmd/
│   ├── arcade/            # 主程序入口
│   └── cli/               # CLI工具
├── internal/
│   └── engine/            # 核心引擎
├── pkg/                   # 公共包
│   ├── plugins/           # 插件源码
│   └── ...
├── plugins/               # 编译后的插件
├── conf.d/                # 配置文件
├── Makefile               # 构建脚本
└── QUICKSTART.md          # 本文档
```

## 常见问题

### Q1: protoc命令未找到

**错误信息**：
```
错误: protoc 未安装，请先安装 Protocol Buffers 编译器
```

**解决方法**：
参考前置要求章节安装protoc。

### Q2: protoc-gen-go未找到

**错误信息**：
```
错误: protoc-gen-go 未安装，请运行: make proto-install
```

**解决方法**：
```bash
make proto-install
```

### Q3: proto生成失败

**可能原因**：
1. protoc版本过低
2. go环境变量配置问题

**解决方法**：
```bash
# 检查protoc版本
protoc --version

# 检查GOPATH
echo $GOPATH

# 确保$GOPATH/bin在PATH中
export PATH=$PATH:$(go env GOPATH)/bin

# 重新安装插件
make proto-install

# 重新生成
make proto
```

### Q4: 构建插件失败

**错误信息**：
```
plugin was built with a different version of package...
```

**解决方法**：
插件和主程序必须使用相同版本的Go编译。清理后重新构建：
```bash
make plugins-clean
make plugins
make build
```

## 下一步

1. 阅读 [API文档](api/README.md) 了解服务接口
2. 阅读 [Label示例](api/LABEL_EXAMPLES.md) 了解标签系统
3. 实现Server端服务逻辑
4. 实现Agent端服务逻辑
5. 开发自定义插件

## 获取帮助

- 查看文档：`api/README.md`
- 查看示例：`api/LABEL_EXAMPLES.md`
- 运行帮助：`make help`
- 提交Issue：https://github.com/observabil/arcade/issues

## 许可证

本项目采用开源许可证，详见 LICENSE 文件。

