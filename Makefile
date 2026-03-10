VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: build install clean test lint

build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o grove .

install:
	go install -ldflags "-s -w -X main.version=$(VERSION)" .

test:
	go test ./... -race

lint:
	golangci-lint run

clean:
	rm -f grove
