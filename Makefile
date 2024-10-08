.PHONY: prebuild build

# git
VERSION    = $(shell git describe --tags --always)
GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
#GIT_COMMIT = $(shell git rev-parse --short=7 HEAD)
GIT_COMMIT = $(shell git rev-parse HEAD)
BUILD_TIME = $(shell date +"%Y-%m-%d %H:%M:%S")

define ldflags
"-X 'github.com/arcade/arcade/pkg/version.Version=${VERSION}' \
 -X 'github.com/arcade/arcade/pkg/version.GitBranch=${GIT_BRANCH}' \
 -X 'github.com/arcade/arcade/pkg/version.GitCommit=${GIT_COMMIT}' \
 -X 'github.com/arcade/arcade/pkg/version.BuildTime=${BUILD_TIME}'"
endef

all: prebuild build

prebuild:
	echo "begin download and embed the front-end file..."
	sh dl_web.sh
	echo "web file download and embedding completed."

build:
	go build -ldflags ${ldflags} -o build-distribution ./cmd/main/main.go

build-cli:
	go build -ldflags ${ldflags} -o build-distribution-cli ./cmd/cli/

run:
	nohup ./build-distribution > build-distribution.log 2>&1 &

release:
	goreleaser --skip-validate --skip-publish --snapshot
