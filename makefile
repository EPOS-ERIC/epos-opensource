BIN=epos-opensource
VERSION?=makefile
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
EXT?=

LDFLAGS=-s -w -X github.com/epos-eu/epos-opensource/cmd.Version=$(VERSION)
BUILDFLAGS=-trimpath

.DEFAULT_GOAL := build

build: generate
	CGO_ENABLED=0 go build $(BUILDFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) .

# Build for specific platform (used by CI)
build-release:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILDFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN)-$(GOOS)-$(GOARCH)$(EXT) .

clean:
	rm -f $(BIN) $(BIN)-* || true

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

.PHONY: build build-release clean generate fmt lint vet test test-race test-integration test-all
