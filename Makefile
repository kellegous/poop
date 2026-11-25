.DEFAULT_GOAL := default

include tools.mk

GO_MOD := $(shell go list -m)

.PHONY: default test lint fmt validate link clean

default: validate

test:
	go test ./...

lint: bin/golangci-lint
	bin/golangci-lint run

fmt: bin/goimports
	find . -type f -name '*.go' -exec bin/goimports -local $(GO_MOD) -w {} +

validate: test lint fmt

link:
	mkdir -p bin
	go build -o bin/example ./example

clean:
	rm -rf bin
