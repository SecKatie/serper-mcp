MODULE := github.com/SecKatie/serper-mcp
VERSION_PKG := $(MODULE)/cmd
LDFLAGS := -ldflags "-X $(VERSION_PKG).version={{.Version}}"

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
