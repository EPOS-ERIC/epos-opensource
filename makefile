BIN=epos-opensource
VERSION?=makefile
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

build: generate
	go build -ldflags "-X github.com/epos-eu/epos-opensource/cmd.Version=$(VERSION)" -o $(BIN) .

# Build for specific platform (used by CI)
build-release:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-X github.com/epos-eu/epos-opensource/cmd.Version=$(VERSION)" -o $(BIN)-$(GOOS)-$(GOARCH)$(EXT) .

clean:
	rm -f $(BIN)*

generate:
	go generate ./...

lint:
	golangci-lint run ./...

test: build
	go test ./...

test-integration: build
	go test -tags=integration -parallel 2 ./integration

# run the integration tests and update the fixtures/golden files with the new outputs of the commands.
# ONLY RUN THIS ON FUNCTIONAL CHANGES
update-fixtures: build
	go test -tags=integration -parallel 2 ./integration -update

.PHONY: build build-release clean generate lint test
