.PHONY: build test lint vet tidy install clean snapshot release-dry-run help

GO ?= go
BINARY := agents-toc
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/noamsiegel/agents-toc/internal/buildinfo.Version=$(VERSION) \
	-X github.com/noamsiegel/agents-toc/internal/buildinfo.Commit=$(COMMIT) \
	-X github.com/noamsiegel/agents-toc/internal/buildinfo.Date=$(DATE)

help:           ## list targets
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk -F':.*?## ' '{printf "  %-18s %s\n", $$1, $$2}'

build:          ## build local binary into ./bin
	@mkdir -p bin
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/$(BINARY) ./

test:           ## run unit + golden tests
	$(GO) test -race -count=1 ./...

vet:            ## go vet
	$(GO) vet ./...

lint:           ## golangci-lint (requires local install)
	golangci-lint run --timeout=5m

tidy:           ## go mod tidy
	$(GO) mod tidy

install:        ## install binary to GOPATH/bin
	$(GO) install -ldflags '$(LDFLAGS)' ./

clean:          ## remove build artifacts
	rm -rf bin dist

snapshot:       ## goreleaser snapshot build (requires goreleaser installed)
	goreleaser release --snapshot --clean

release-dry-run:## goreleaser dry-run (no publish)
	goreleaser release --skip=publish --clean
