BINARY ?= freshservice_label

.PHONY: all build test fmt vet tidy check preview clean

all: check build

build:
	go build -o bin/$(BINARY) ./cmd/freshservice-label

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...

tidy:
	go mod tidy

check: fmt tidy test vet

preview:
	go run ./cmd/freshservice-label preview < preview.json

clean:
	rm -rf bin dist coverage.out
