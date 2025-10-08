#!/bin/bash

# 插件自动加载测试脚本
# 用于演示插件系统的自动监控和热加载功能

set -e

COLOR_GREEN='\033[0;32m'
COLOR_BLUE='\033[0;34m'
COLOR_YELLOW='\033[1;33m'
COLOR_RED='\033[0;31m'
COLOR_RESET='\033[0m'

function info() {
    echo -e "${COLOR_BLUE}[INFO]${COLOR_RESET} $1"
}

function success() {
    echo -e "${COLOR_GREEN}[SUCCESS]${COLOR_RESET} $1"
}

function warn() {
    echo -e "${COLOR_YELLOW}[WARN]${COLOR_RESET} $1"
}

function error() {
    echo -e "${COLOR_RED}[ERROR]${COLOR_RESET} $1"
}

function separator() {
    echo "========================================"
}

# 检查是否在项目根目录
if [ ! -f "go.mod" ]; then
    error "请在项目根目录下运行此脚本"
    exit 1
fi

separator
info "Arcade 插件自动加载测试脚本"
separator
echo ""

# 1. 创建必要的目录
info "创建插件目录..."
mkdir -p plugins
mkdir -p conf.d
mkdir -p examples/plugin_autowatch
success "目录创建完成"
echo ""

# 2. 检查示例插件
info "检查示例插件..."
if [ -f "plugins/stdout.so" ]; then
    success "找到示例插件: plugins/stdout.so"
else
    warn "未找到示例插件，你需要先编译插件"
    info "编译命令示例:"
    echo "  cd pkg/plugins/notify/stdout"
    echo "  go build -buildmode=plugin -o ../../../../plugins/stdout.so stdout.go"
fi
echo ""

# 3. 检查配置文件
info "检查配置文件..."
if [ -f "conf.d/plugins.yaml" ]; then
    success "找到配置文件: conf.d/plugins.yaml"
else
    warn "未找到配置文件，创建示例配置..."
    cat > conf.d/plugins.yaml << 'EOF'
# 插件配置文件
# 支持自动监控和热重载

plugins:
  - path: ./plugins/stdout.so
    name: stdout
    type: notify
    version: "1.0.0"
    config:
      prefix: "[通知]"
EOF
    success "配置文件已创建: conf.d/plugins.yaml"
fi
echo ""

# 4. 检查演示程序
info "检查演示程序..."
if [ -f "examples/plugin_autowatch/main.go" ]; then
    success "找到演示程序"
else
    error "未找到演示程序: examples/plugin_autowatch/main.go"
    exit 1
fi
echo ""

# 5. 显示使用说明
separator
info "使用说明"
separator
echo ""
echo "1. 启动演示程序:"
echo -e "   ${COLOR_GREEN}go run examples/plugin_autowatch/main.go${COLOR_RESET}"
echo ""
echo "2. 在另一个终端窗口，你可以进行以下操作:"
echo ""
echo "   a) 添加新插件:"
echo -e "      ${COLOR_GREEN}cp /path/to/new-plugin.so plugins/${COLOR_RESET}"
echo "      系统会自动检测并加载新插件"
echo ""
echo "   b) 删除插件:"
echo -e "      ${COLOR_GREEN}rm plugins/some-plugin.so${COLOR_RESET}"
echo "      系统会自动卸载该插件"
echo ""
echo "   c) 修改配置:"
echo -e "      ${COLOR_GREEN}vim conf.d/plugins.yaml${COLOR_RESET}"
echo "      保存后系统会自动重新加载配置"
echo ""
echo "3. 查看日志输出，观察插件的加载和卸载过程"
echo ""
echo "4. 按 Ctrl+C 停止演示程序"
echo ""

separator
info "准备工作"
separator
echo ""

# 询问是否编译示例插件
if [ ! -f "plugins/stdout.so" ]; then
    read -p "是否编译示例插件? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        info "编译示例插件..."
        if [ -f "pkg/plugins/notify/stdout/stdout.go" ]; then
            cd pkg/plugins/notify/stdout
            go build -buildmode=plugin -o ../../../../plugins/stdout.so stdout.go
            cd - > /dev/null
            success "插件编译完成: plugins/stdout.so"
        else
            warn "未找到插件源代码: pkg/plugins/notify/stdout/stdout.go"
        fi
    fi
    echo ""
fi

# 询问是否启动演示
separator
read -p "是否立即启动演示程序? (y/n) " -n 1 -r
echo
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    separator
    info "启动演示程序..."
    separator
    echo ""
    sleep 1
    go run examples/plugin_autowatch/main.go
else
    info "你可以稍后手动运行:"
    echo -e "  ${COLOR_GREEN}go run examples/plugin_autowatch/main.go${COLOR_RESET}"
    echo ""
fi

success "测试脚本执行完成"

