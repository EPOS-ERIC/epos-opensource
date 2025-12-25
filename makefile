BIN=epos-opensource
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
EXT?=
PREFIX ?= /usr/local

LDFLAGS=-s -w -X github.com/epos-eu/epos-opensource/common.Version=$(VERSION)
BUILDFLAGS=-trimpath

.DEFAULT_GOAL := build

build: generate
	CGO_ENABLED=0 go build $(BUILDFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) .

# Build for specific platform (used by CI)
build-release:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILDFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN)-$(GOOS)-$(GOARCH)$(EXT) .

install: build
	install -d $(PREFIX)/bin
	install -m 755 $(BIN) $(PREFIX)/bin/$(BIN)

clean:
	rm -f $(BIN) $(BIN)-* || true
	rm log.log || true

generate:
	go generate ./...

fmt:
	go tool gofumpt -w .

lint: fmt
	golangci-lint run ./...

vet:
	go vet ./...
	go tool sqlc vet ./...

test: vet
	go test ./...

test-race:
	go test ./... -race -count=1

test-integration:
	go test -tags=integration ./...

test-all: test test-race test-integration

debug: generate
	go build -gcflags "all=-N -l" -o $(BIN) .
	dlv exec --headless=true --listen=127.0.0.1:2345 --api-version=2 --accept-multiclient ./$(BIN)

run: clean build
	./$(BIN)

.PHONY: build build-release clean generate fmt lint vet test test-race test-integration test-all install
