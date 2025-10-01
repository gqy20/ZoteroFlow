.PHONY: build test lint fmt clean run

# Variables
BINARY_NAME=zoteroflow2
PKG_NAME=zoteroflow2-server

# Build targets
build:
	go build -o bin/$(BINARY_NAME) .

test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')"

fmt:
	go fmt ./...

lint:
	golangci-lint run

check: fmt lint test

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

deps:
	go mod download
	go mod tidy

run: build
	./bin/$(BINARY_NAME)

dev:
	go run .

# Development helpers
mod-upgrade:
	go get -u ./...
	go mod tidy

vet:
	go vet ./...

# Quick checks
quick: fmt vet