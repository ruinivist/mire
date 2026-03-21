.PHONY: build test

BIN := build/mire

build:
	mkdir -p build
	go build -o $(BIN) ./cmd/mire

test:
	go test ./...
