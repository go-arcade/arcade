SHELL := /bin/sh
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c

.PHONY: help prebuild build plugins plugins-clean proto proto-clean proto-install

# git
VERSION    = $(shell git describe --tags --always)
GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
#GIT_COMMIT = $(shell git rev-parse --short=7 HEAD)
GIT_COMMIT = $(shell git rev-parse HEAD)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# plugins
PLUGINS_SRC_DIR := pkg/plugins
PLUGINS_OUT_DIR := plugins

# 用 go list 只挑出 package=main 的插件目录
PLUGINS_MAIN_DIRS := $(shell go list -f '{{if eq .Name "main"}}{{.Dir}}{{end}}' ./$(PLUGINS_SRC_DIR)/... | sed '/^$$/d')

# proto
PROTO_DIR := api
PROTO_OUT_DIR := proto
PROTO_FILES := $(shell find $(PROTO_DIR) -name '*.proto')
PROTOC := protoc
PROTOC_GEN_GO := $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(shell go env GOPATH)/bin/protoc-gen-go-grpc

LDFLAGS := \
 -X 'github.com/observabil/arcade/pkg/version.Version=$(VERSION)' \
 -X 'github.com/observabil/arcade/pkg/version.GitBranch=$(GIT_BRANCH)' \
 -X 'github.com/observabil/arcade/pkg/version.GitCommit=$(GIT_COMMIT)' \
 -X 'github.com/observabil/arcade/pkg/version.BuildTime=$(BUILD_TIME)'

.DEFAULT_GOAL := help

help: ## 显示帮助信息
	@echo "Arcade CI/CD 平台 Makefile 命令"
	@echo ""
	@echo "使用方法: make [命令]"
	@echo ""
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "示例:"
	@echo "  make proto-install  # 安装proto工具"
	@echo "  make proto          # 生成proto代码"
	@echo "  make all            # 完整构建"

all: prebuild plugins build ## 完整构建（前端+插件+主程序）

JOBS ?= $(shell getconf _NPROCESSORS_ONLN 2>/dev/null || echo 4)

plugins: ## 构建所有插件
	@echo ">> building plugins from $(PLUGINS_SRC_DIR) into $(PLUGINS_OUT_DIR)"
	@mkdir -p "$(PLUGINS_OUT_DIR)"
	@go list -f '{{if eq .Name "main"}}{{.Dir}} {{range .GoFiles}}{{.}} {{end}}{{end}}' ./$(PLUGINS_SRC_DIR)/... \
	| sed -e '/^[[:space:]]*$$/d' \
	| awk '\
	function basename(p){sub(".*/","",p); return p} \
	{ \
	  dir=$$1; outbase=""; \
	  for(i=2;i<=NF;i++){ if($$i=="main.go"){ outbase="main"; break } } \
	  if(outbase==""){ for(i=2;i<=NF;i++){ f=$$i; sub(/\.go$$/,"",f); if(f ~ /^main.*/){ outbase=f; break } } } \
	  if(outbase==""){ outbase=basename(dir) } \
	  printf "%s %s/%s.so\n", dir, "$(PLUGINS_OUT_DIR)", outbase; \
	}' \
	| xargs -P $(JOBS) -n 2 sh -c '\
		dir="$$1"; out="$$2"; \
		echo "   -> $$dir  ==>  $$out"; \
		GO111MODULE=on go build -buildmode=plugin -o "$$out" "$$dir" \
	' sh
	@echo ">> plugins build done."

plugins-clean: ## 清理插件构建产物
	@echo ">> cleaning $(PLUGINS_OUT_DIR)/*.so"
	@rm -f $(PLUGINS_OUT_DIR)/*.so || true

prebuild: ## 下载并嵌入前端文件
	echo "begin download and embed the front-end file..."
	sh dl.sh
	echo "web file download and embedding completed."

build: ## 构建主程序
	go build -ldflags "${LDFLAGS}" -o arcade ./cmd/arcade/main.go

build-cli: ## 构建CLI工具
	go build -ldflags "${LDFLAGS}" -o arcade-cli ./cmd/cli/

run: ## 后台运行主程序
	nohup ./arcade > arcade.log 2>&1 &

release: ## 创建发布版本
	goreleaser --skip-validate --skip-publish --snapshot

# proto代码生成
proto-install: ## 安装protoc相关插件
	@echo ">> installing protoc plugins..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo ">> protoc plugins installed."
	@echo "   protoc-gen-go: $(PROTOC_GEN_GO)"
	@echo "   protoc-gen-go-grpc: $(PROTOC_GEN_GO_GRPC)"

proto: proto-check ## 生成proto代码
	@echo ">> generating proto code from $(PROTO_DIR)"
	@for proto in $(PROTO_FILES); do \
		dir=$$(dirname $$proto); \
		proto_dir=$$dir/proto; \
		echo "   -> generating $$proto"; \
		mkdir -p $$proto_dir; \
		$(PROTOC) --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			-I. $$proto; \
		mkdir -p $$proto_dir; \
		mv $$dir/*.pb.go $$proto_dir/ 2>/dev/null || true; \
	done
	@echo ">> proto code generation done."

proto-check: ## 检查proto工具是否已安装
	@command -v $(PROTOC) >/dev/null 2>&1 || { \
		echo "错误: protoc 未安装，请先安装 Protocol Buffers 编译器"; \
		echo ""; \
		echo "macOS 安装方法:"; \
		echo "  brew install protobuf"; \
		echo ""; \
		echo "Linux 安装方法:"; \
		echo "  apt-get install -y protobuf-compiler  # Debian/Ubuntu"; \
		echo "  yum install -y protobuf-compiler      # CentOS/RHEL"; \
		echo ""; \
		echo "或从官网下载: https://github.com/protocolbuffers/protobuf/releases"; \
		exit 1; \
	}
	@test -f $(PROTOC_GEN_GO) || { \
		echo "错误: protoc-gen-go 未安装，请运行: make proto-install"; \
		exit 1; \
	}
	@test -f $(PROTOC_GEN_GO_GRPC) || { \
		echo "错误: protoc-gen-go-grpc 未安装，请运行: make proto-install"; \
		exit 1; \
	}

proto-clean: ## 清理生成的proto代码
	@echo ">> cleaning generated proto files..."
	@find $(PROTO_DIR) -type d -name "proto" -exec rm -rf {} + 2>/dev/null || true
	@echo ">> proto files cleaned."
