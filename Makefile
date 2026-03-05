MODULE := github.com/SecKatie/serper-mcp
VERSION_PKG := $(MODULE)/cmd
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X $(VERSION_PKG).version=$(VERSION)"

.PHONY: build test lint audit release

build:
	go build $(LDFLAGS) ./...

test:
	go test -race ./...

lint:
	golangci-lint run

audit:
	govulncheck ./...
	gosec ./...

release:
	goreleaser release --clean
