GOPATH     ?= $(shell go env GOPATH)
GORELEASER ?= $(GOPATH)/bin/goreleaser
VERSION    := v$(shell cat VERSION)

.PHONY: setup lint fmt test

setup:
	go get golang.org/x/tools/cmd/goimports

lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint golangci-lint run -E gofmt,goimports

fmt:
	go fmt ./...
	goimports -l -w .

test:
	@go test ./...
