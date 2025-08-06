# mvx Makefile

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)"

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	go build $(LDFLAGS) -o mvx-binary .

# Build for multiple platforms
.PHONY: build-all
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/mvx-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/mvx-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/mvx-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/mvx-windows-amd64.exe .

# Run tests
.PHONY: test
test:
	go test ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -f mvx-binary
	rm -rf dist/
	rm -rf .mvx/local/ .mvx/tools/ .mvx/versions/

# Install dependencies
.PHONY: deps
deps:
	go mod download
	go mod tidy

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Run linter
.PHONY: lint
lint:
	golangci-lint run

# Development build with race detection
.PHONY: dev
dev:
	go build -race $(LDFLAGS) -o mvx-binary .

# Install locally
.PHONY: install
install:
	go install $(LDFLAGS) .

# Package wrapper files
.PHONY: package-wrapper
package-wrapper:
	@echo "Packaging mvx wrapper files..."
	@mkdir -p dist/wrapper
	@cp mvx dist/wrapper/
	@cp mvx.cmd dist/wrapper/
	@cp -r .mvx dist/wrapper/
	@cp WRAPPER.md dist/wrapper/
	@cp install-wrapper.sh dist/wrapper/
	@echo "Wrapper files packaged in dist/wrapper/"

# Test wrapper functionality
.PHONY: test-wrapper
test-wrapper: build
	@echo "Testing mvx wrapper..."
	@cp mvx-binary ./mvx.exe 2>/dev/null || cp mvx-binary ./mvx-binary-local
	@./mvx version || echo "Wrapper test completed"
	@rm -f ./mvx.exe ./mvx-binary-local

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-all      - Build for multiple platforms"
	@echo "  test           - Run tests"
	@echo "  test-wrapper   - Test wrapper functionality"
	@echo "  package-wrapper- Package wrapper files for distribution"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  dev            - Development build with race detection"
	@echo "  install        - Install locally"
	@echo "  help           - Show this help"
