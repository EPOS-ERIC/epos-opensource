BIN=epos-opensource
VERSION?=make

run: build
	./$(BIN)
	@make clean

build:
	go build -ldflags "-X $(BIN)/cmd.Version=$(VERSION)" -o $(BIN) .

clean:
	rm -f $(BIN)

vet:
	go vet ./...

test:
	go test ./... -v
