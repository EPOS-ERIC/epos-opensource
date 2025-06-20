BIN=epos-opensource
VERSION?=make

run: build
	./$(BIN)
	@make clean

build:
	go build -ldflags "-X github.com/epos-eu/epos-opensource/cmd.Version=$(VERSION)" -o $(BIN) .

clean:
	rm -f $(BIN)

vet:
	go vet ./...

test:
	go test ./... -v
