.PHONY: fmt lint build

fmt:
	gofmt -s -w .

lint:
	golangci-lint run

build:
	go build
