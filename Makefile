.PHONY: build test

BIN := build/miro

build:
	mkdir -p build
	go build -o $(BIN) ./cmd/miro

test:
	go test ./...
