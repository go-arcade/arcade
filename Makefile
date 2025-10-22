SHELL := /bin/sh
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c

.PHONY: help prebuild build plugins plugins-clean plugins-package proto proto-clean proto-install wire wire-install wire-clean

# git
VERSION    = $(shell git describe --tags --always)
GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
#GIT_COMMIT = $(shell git rev-parse --short=7 HEAD)
GIT_COMMIT = $(shell git rev-parse HEAD)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# plugins
PLUGINS_SRC_DIR := pkg/plugins
PLUGINS_OUT_DIR := plugins

# use go list to only pick out plugin directories with package=main
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

help: ## show help information
	@echo "Arcade CI/CD platform Makefile commands"
	@echo ""
	@echo "Usage: make [command]"
	@echo ""
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make proto-install  # install proto tool"
	@echo "  make proto          # generate proto code"
	@echo "  make all            # full build"

all: deps-sync prebuild plugins build ## full build (frontend+plugins+main program)

JOBS ?= $(shell getconf _NPROCESSORS_ONLN 2>/dev/null || echo 4)

plugins: ## build all RPC plugins (executable files)
	@echo ">> building RPC plugins from $(PLUGINS_SRC_DIR) into $(PLUGINS_OUT_DIR)"
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
	  printf "%s %s/%s_1.0.0\n", dir, "$(PLUGINS_OUT_DIR)", outbase; \
	}' \
	| xargs -P $(JOBS) -n 2 sh -c '\
		dir="$$1"; out="$$2"; \
		echo "   -> $$dir  ==>  $$out"; \
		cd "$$(git rev-parse --show-toplevel)" && \
		go build -o "$$out" "$$dir" \
	' sh
	@echo ">> RPC plugins build done."
	@echo ">> built plugins:"
	@ls -lh $(PLUGINS_OUT_DIR)/ | grep -v "^total" | grep -v "^d" || true

plugins-package: plugins ## package plugins to zip files
	@echo ">> packaging plugins..."
	@for plugin_path in $(PLUGINS_OUT_DIR)/*_*; do \
		if [ -f "$$plugin_path" ] && [ -x "$$plugin_path" ]; then \
			plugin_name=$$(basename $$plugin_path); \
			plugin_base=$$(echo $$plugin_name | sed 's/_[0-9.]*$$//'); \
			src_dir=$$(find $(PLUGINS_SRC_DIR) -type d -name $$plugin_base | head -1); \
			if [ -n "$$src_dir" ] && [ -f "$$src_dir/manifest.json" ]; then \
				zip_file=$(PLUGINS_OUT_DIR)/$${plugin_base}.zip; \
				echo "   -> packaging $$plugin_name to $$zip_file"; \
				cd $$(git rev-parse --show-toplevel) && \
				zip -j $$zip_file $$plugin_path $$src_dir/manifest.json; \
			else \
				echo "   !! manifest.json not found for $$plugin_name, skipping"; \
			fi; \
		fi; \
	done
	@echo ">> plugin packaging done."

plugins-clean: ## clean plugin build artifacts
	@echo ">> cleaning $(PLUGINS_OUT_DIR)/*"
	@rm -f $(PLUGINS_OUT_DIR)/* || true
	@rm -rf $(PLUGINS_OUT_DIR)/.tmp || true

prebuild: ## download and embed the front-end file
	echo "begin download and embed the front-end file..."
	sh dl.sh
	echo "web file download and embedding completed."

build: wire plugins buf ## build main program
	go build -ldflags "${LDFLAGS}" -o arcade ./cmd/arcade/

build-cli: ## build CLI tool
	go build -ldflags "${LDFLAGS}" -o arcade-cli ./cmd/cli/

run: deps-sync wire buf
	go run ./cmd/arcade/

release: ## create release version
	goreleaser --skip-validate --skip-publish --snapshot

# proto code generation
buf-install: ## install buf related plugins
	@echo ">> installing buf..."
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@echo ">> buf installed: $$(which buf)"

buf: buf-check ## generate buf code
	@echo ">> generating buf code from $(PROTO_DIR)"
	@cd $(PROTO_DIR) && buf generate --template buf.gen.yaml
	@echo ">> buf code generation done."

buf-check: ## check if buf tool is installed
	@command -v buf >/dev/null 2>&1 || { \
		echo "error: buf is not installed, please run make buf-install"; \
		exit 1; \
	}
	@echo ">> buf installed: $$(which buf)"

buf-lint: ## check buf code style
	@echo ">> linting buf code..."
	@cd $(PROTO_DIR) && buf lint
	@echo ">> buf code linting done."

buf-breaking: ## check buf code breaking changes
	@echo ">> checking buf code breaking changes..."
	@cd $(PROTO_DIR) && buf breaking
	@echo ">> buf code breaking changes checking done."

buf-push: ## push buf code
	@echo ">> pushing buf code..."
	@cd $(PROTO_DIR) && buf push
	@echo ">> buf code pushing done."

buf-clean: ## clean generated buf code
	@echo ">> cleaning generated protobuf files..."
	@find $(PROTO_DIR) -type f \( -name "*.pb.go" -o -name "*_grpc.pb.go" \) -delete 2>/dev/null || true
	@echo ">> protobuf files cleaned."

wire-install: ## install wire tool
	@echo ">> installing wire..."
	@go install github.com/google/wire/cmd/wire@latest
	@echo ">> wire installed: $$(which wire)"

wire: ## generate wire dependency injection code
	@echo ">> generating wire code..."
	@cd cmd/arcade && wire
	@echo ">> wire code generation done."

wire-clean: ## clean wire generated code
	@echo ">> cleaning wire generated files..."
	@find . -name "wire_gen.go" -type f -delete
	@echo ">> wire files cleaned."
