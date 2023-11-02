BINARY_NAME = "chefbrowser"
RELEASE?=dev
DOCKER_NAMESPACE?=drewhammond
DOCKER_TAG?=latest
HOST_GOOS ?= $(shell go env GOOS)
HOST_GOARCH ?= $(shell go env GOARCH)
GIT_SHA=$(shell git rev-parse HEAD)
DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
BUILD_INFO_PATH="github.com/drewhammond/chefbrowser/internal/common/version"
BUILD_INFO=-ldflags "-X $(BUILD_INFO_PATH).version=$(RELEASE) -X $(BUILD_INFO_PATH).commitHash=$(GIT_SHA) -X $(BUILD_INFO_PATH).date=$(DATE)"
GOBUILD=CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -trimpath $(BUILD_INFO)
CURRENT_DIR=$(shell pwd)
DIST_DIR=$(CURDIR)/dist
TARGET_ARCH?=linux/amd64

.PHONY: lint
lint:
	golangci-lint -v run

.PHONY: test
test:
	go test -v ./...

.PHONY: fmt
fmt:
	find . -name '*.go' | grep -v pb.go | grep -v vendor | xargs gofumpt -w

.PHONY: ui-deps
ui-deps:
	cd $(CURDIR)/ui && yarn install

.PHONY: build
build: build-ui build-backend

.PHONY: build-backend
build-backend:
	$(GOBUILD) -o $(DIST_DIR)/$(BINARY_NAME) .

.PHONY: build-ui
build-ui:
	docker build -t chefbrowser-ui --platform=$(TARGET_ARCH) --target ui-builder .
	find $(CURDIR)/ui/dist -type f -not -name gitkeep -delete || true
	docker run --platform=$(TARGET_ARCH) -v $(CURDIR)/ui/dist:/tmp/app --rm -t chefbrowser-ui sh -c 'cp -r ./dist/* /tmp/app/'

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 TARGET_ARCH=linux/amd64 $(MAKE) build

.PHONY: build-docker
build-docker:
	docker build --no-cache -t $(DOCKER_NAMESPACE)/$(BINARY_NAME):$(DOCKER_TAG) -f Dockerfile .
