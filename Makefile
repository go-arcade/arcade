SHELL := /bin/sh
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c

.PHONY: help prebuild build plugins plugins-clean proto proto-clean proto-install wire wire-install wire-clean

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
PROTO_FILES := $(shell find $(PROTO_DIR) -path '*/proto/*.proto')
PROTOC := protoc
PROTOC_GEN_GO := $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(shell go env GOPATH)/bin/protoc-gen-go-grpc

LDFLAGS := \
 -X 'github.com/observabil/arcade/pkg/version.Version=$(VERSION)' \
 -X 'github.com/observabil/arcade/pkg/version.GitBranch=$(GIT_BRANCH)' \
 -X 'github.com/observabil/arcade/pkg/version.GitCommit=$(GIT_COMMIT)' \
 -X 'github.com/observabil/arcade/pkg/version.BuildTime=$(BUILD_TIME)'

.DEFAULT_GOAL := help

deps-sync:
	go mod tidy
	go mod verify

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

all: deps-sync prebuild plugins build ## 完整构建（前端+插件+主程序）

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
		cd "$$(git rev-parse --show-toplevel)" && \
		go build -buildmode=plugin -o "$$out" "$$dir" \
	' sh
	@echo ">> plugins build done."

plugins-clean: ## 清理插件构建产物
	@echo ">> cleaning $(PLUGINS_OUT_DIR)/*.so"
	@rm -f $(PLUGINS_OUT_DIR)/*.so || true

prebuild: ## 下载并嵌入前端文件
	echo "begin download and embed the front-end file..."
	sh dl.sh
	echo "web file download and embedding completed."

build: wire plugins ## 构建主程序
	go build -ldflags "${LDFLAGS}" -o arcade ./cmd/arcade/

build-cli: ## 构建CLI工具
	go build -ldflags "${LDFLAGS}" -o arcade-cli ./cmd/cli/

run: deps-sync wire 
	go run ./cmd/arcade/

release: ## 创建发布版本
	goreleaser --skip-validate --skip-publish --snapshot

# proto代码生成
buf-install: ## 安装buf相关插件
	@echo ">> installing buf..."
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@echo ">> buf installed: $$(which buf)"

buf: buf-check ## 生成buf代码
	@echo ">> generating buf code from $(PROTO_DIR)"
	@cd $(PROTO_DIR) && buf generate --template buf.gen.yaml
	@echo ">> buf code generation done."

buf-check: ## 检查buf工具是否已安装
	@command -v buf >/dev/null 2>&1 || { \
		echo "错误: buf 未安装，请先运行 make buf-install"; \
		exit 1; \
	}
	@echo ">> buf installed: $$(which buf)"

buf-lint: ## 检查buf代码风格
	@echo ">> linting buf code..."
	@cd $(PROTO_DIR) && buf lint
	@echo ">> buf code linting done."

buf-breaking: ## 检查buf代码破坏性变更
	@echo ">> checking buf code breaking changes..."
	@cd $(PROTO_DIR) && buf breaking
	@echo ">> buf code breaking changes checking done."

buf-push: ## 推送buf代码
	@echo ">> pushing buf code..."
	@cd $(PROTO_DIR) && buf push
	@echo ">> buf code pushing done."

buf-clean: ## 清理生成的buf代码
	@echo ">> cleaning generated protobuf files..."
	@find $(PROTO_DIR) -type f \( -name "*.pb.go" -o -name "*_grpc.pb.go" \) -delete 2>/dev/null || true
	@echo ">> protobuf files cleaned."

# wire依赖注入代码生成
wire-install: ## 安装wire工具
	@echo ">> installing wire..."
	@go install github.com/google/wire/cmd/wire@latest
	@echo ">> wire installed: $$(which wire)"

wire: ## 生成wire依赖注入代码
	@echo ">> generating wire code..."
	@cd cmd/arcade && wire
	@echo ">> wire code generation done."

wire-clean: ## 清理wire生成的代码
	@echo ">> cleaning wire generated files..."
	@find . -name "wire_gen.go" -type f -delete
	@echo ">> wire files cleaned."
