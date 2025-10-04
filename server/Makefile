.PHONY: build test lint fmt clean run

# Variables
BINARY_NAME=zoteroflow2
PKG_NAME=zoteroflow2-server

# Build targets
build:
	@echo "开始构建..."
	go build -o bin/$(BINARY_NAME) .
	upx --best --lzma bin/$(BINARY_NAME)
	@echo "构建完成，文件已UPX优化："
	@ls -lh bin/$(BINARY_NAME)

# 生产构建 - 功能完整且最小
build-prod:  # 生产版本（完整功能 + UPX压缩）
	CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/$(BINARY_NAME)-prod .
	upx --best --lzma bin/$(BINARY_NAME)-prod
	@echo "生产版本构建完成，大小："
	@ls -lh bin/$(BINARY_NAME)-prod

# 开发构建 - 调试友好
build-dev:   # 开发版本（保留调试信息）
	go build -o bin/$(BINARY_NAME)-dev .

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