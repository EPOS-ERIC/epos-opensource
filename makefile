BIN=epos-cli
VERSION?=make

run: build
	./epos-cli
	@make clean

build:
	go build -ldflags "-X epos-cli/cmd.Version=$(VERSION)" -o $(BIN) .

clean:
	rm -f $(BIN)

vet:
	go vet ./...
