BIN=epos-opensource
VERSION?=makefile
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

build: generate
	CGO_ENABLED=0 go build -ldflags "-X github.com/epos-eu/epos-opensource/cmd.Version=$(VERSION)" -o $(BIN) .

# Build for specific platform (used by CI)
build-release: generate
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-X github.com/epos-eu/epos-opensource/cmd.Version=$(VERSION)" -o $(BIN)-$(GOOS)-$(GOARCH)$(EXT) .

clean:
	rm $(BIN)*

generate:
	go generate ./...

lint:
	golangci-lint run ./...

test: vet
	go test ./...

test-integration: vet
	go test -tags=integration ./...

vet:
	go vet ./...
	go tool sqlc vet ./...

.PHONY: build build-release clean generate lint test vet
