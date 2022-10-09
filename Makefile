BINARY_NAME = "chefbrowser"
GO_VERSION_MIN=1.19
RELEASE?=dev
GIN_MODE?=release
DOCKER_NAMESPACE?=drewhammond
DOCKER_TAG?=latest
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOBUILD=CGO_ENABLED=0 go build -trimpath
GIT_SHA=$(shell git rev-parse HEAD)
DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
BUILD_INFO_PATH="github.com/drewhammond/chefbrowser/internal/common/version"
BUILD_INFO=-ldflags "-X $(BUILD_INFO_PATH).version=$(RELEASE) -X $(BUILD_INFO_PATH).commitHash=$(GIT_SHA) -X $(BUILD_INFO_PATH).date=$(DATE)"

.PHONY: lint
lint:
	golangci-lint -v run

.PHONY: test
test:
	go test -v ./...

.PHONY: fmt
fmt:
	find . -name '*.go' | grep -v pb.go | grep -v vendor | xargs gofumpt -w

.PHONY: build
build:
	$(GOBUILD) -o bin/${BINARY_NAME}-$(GOOS)-$(GOARCH) $(BUILD_INFO) main.go

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 $(MAKE) build

.PHONY: build-docker
build-docker: build-linux
	docker build --no-cache -t $(DOCKER_NAMESPACE)/$(BINARY_NAME):$(DOCKER_TAG) -f Dockerfile .
