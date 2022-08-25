GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOBUILD=CGO_ENABLED=0 go build -trimpath
GIT_SHA=$(shell git rev-parse HEAD)
DATE=$(shell date -u -d @$(shell git show -s --format=%ct) +'%Y-%m-%dT%H:%M:%SZ')
